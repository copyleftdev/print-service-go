package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"print-service/internal/core/domain"
	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/pool"
)

// JobQueue manages print job processing with priority queuing
type JobQueue struct {
	jobs        chan *domain.PrintJob
	results     chan *JobResult
	workers     *pool.WorkerPool
	storage     JobStorage
	logger      logger.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	jobStatus   map[string]*domain.PrintJob
	statusMutex sync.RWMutex
	retryQueue  chan *domain.PrintJob
	maxRetries  int
	retryDelay  time.Duration
}

// JobResult represents the result of a processed job
type JobResult struct {
	Job    *domain.PrintJob
	Result *domain.RenderResult
	Error  error
}

// JobStorage interface for persisting job data
type JobStorage interface {
	SaveJob(job *domain.PrintJob) error
	GetJob(id string) (*domain.PrintJob, error)
	UpdateJob(job *domain.PrintJob) error
	ListJobs(offset, limit int) ([]*domain.PrintJob, error)
	DeleteJob(id string) error
}

// Config holds job queue configuration
type Config struct {
	WorkerCount    int           `yaml:"worker_count"`
	QueueSize      int           `yaml:"queue_size"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	ProcessTimeout time.Duration `yaml:"process_timeout"`
}

// NewJobQueue creates a new job queue
func NewJobQueue(config Config, storage JobStorage, logger logger.Logger) *JobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	jq := &JobQueue{
		jobs:       make(chan *domain.PrintJob, config.QueueSize),
		results:    make(chan *JobResult, config.QueueSize),
		storage:    storage,
		logger:     logger.With("component", "job_queue"),
		ctx:        ctx,
		cancel:     cancel,
		jobStatus:  make(map[string]*domain.PrintJob),
		retryQueue: make(chan *domain.PrintJob, config.QueueSize/2),
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
	}

	// Initialize worker pool
	jq.workers = pool.NewWorkerPool(config.WorkerCount, logger)

	return jq
}

// Start starts the job queue processing
func (jq *JobQueue) Start(ctx context.Context, processor JobProcessor) error {
	jq.logger.Info("Starting job queue", "worker_count", jq.workers.GetStats().Size)

	// Start worker pool
	jq.workers.Start(ctx, func(job interface{}) error {
		printJob, ok := job.(*domain.PrintJob)
		if !ok {
			return fmt.Errorf("invalid job type: %T", job)
		}
		return jq.processJob(ctx, printJob, processor)
	})

	// Start result processor
	jq.wg.Add(1)
	go jq.processResults()

	// Start retry processor
	jq.wg.Add(1)
	go jq.processRetries(ctx, processor)

	return nil
}

// Stop stops the job queue gracefully
func (jq *JobQueue) Stop(ctx context.Context) error {
	jq.logger.Info("Stopping job queue")

	// Cancel context
	jq.cancel()

	// Close job channels
	close(jq.jobs)
	close(jq.retryQueue)

	// Stop worker pool
	jq.workers.Stop(ctx)

	// Wait for processors to finish
	done := make(chan struct{})
	go func() {
		jq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		jq.logger.Info("Job queue stopped successfully")
		return nil
	case <-ctx.Done():
		jq.logger.Warn("Job queue stop timeout")
		return ctx.Err()
	}
}

// Submit submits a job for processing
func (jq *JobQueue) Submit(job *domain.PrintJob) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}

	// Save job to storage
	if err := jq.storage.SaveJob(job); err != nil {
		jq.logger.Error("Failed to save job", "job_id", job.ID, "error", err)
		return fmt.Errorf("failed to save job: %w", err)
	}

	// Update in-memory status
	jq.statusMutex.Lock()
	jq.jobStatus[job.ID] = job
	jq.statusMutex.Unlock()

	// Submit to worker pool
	if err := jq.workers.Submit(job); err != nil {
		jq.logger.Error("Failed to submit job to worker pool", "job_id", job.ID, "error", err)

		// Update job status to failed
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		jq.updateJobStatus(job)

		return fmt.Errorf("failed to submit job: %w", err)
	}

	jq.logger.Info("Job submitted successfully", "job_id", job.ID)
	return nil
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(id string) (*domain.PrintJob, error) {
	// Check in-memory cache first
	jq.statusMutex.RLock()
	if job, exists := jq.jobStatus[id]; exists {
		jq.statusMutex.RUnlock()
		return job, nil
	}
	jq.statusMutex.RUnlock()

	// Fallback to storage
	return jq.storage.GetJob(id)
}

// ListJobs lists jobs with pagination
func (jq *JobQueue) ListJobs(offset, limit int) ([]*domain.PrintJob, error) {
	return jq.storage.ListJobs(offset, limit)
}

// CancelJob cancels a job
func (jq *JobQueue) CancelJob(id string) error {
	jq.statusMutex.Lock()
	defer jq.statusMutex.Unlock()

	job, exists := jq.jobStatus[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	if job.Status == domain.JobStatusCompleted || job.Status == domain.JobStatusFailed {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	job.Status = domain.JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	jq.updateJobStatus(job)
	jq.logger.Info("Job cancelled", "job_id", id)
	return nil
}

// GetStats returns queue statistics
func (jq *JobQueue) GetStats() QueueStats {
	jq.statusMutex.RLock()
	defer jq.statusMutex.RUnlock()

	stats := QueueStats{
		TotalJobs:   len(jq.jobStatus),
		QueuedJobs:  len(jq.jobs),
		WorkerStats: jq.workers.GetStats(),
	}

	// Count jobs by status
	for _, job := range jq.jobStatus {
		switch job.Status {
		case domain.JobStatusPending:
			stats.PendingJobs++
		case domain.JobStatusProcessing:
			stats.ProcessingJobs++
		case domain.JobStatusCompleted:
			stats.CompletedJobs++
		case domain.JobStatusFailed:
			stats.FailedJobs++
		case domain.JobStatusCancelled:
			stats.CancelledJobs++
		}
	}

	return stats
}

// processJob processes a single job
func (jq *JobQueue) processJob(ctx context.Context, job *domain.PrintJob, processor JobProcessor) error {
	jq.logger.Info("Processing job", "job_id", job.ID)

	// Update job status to processing
	job.Status = domain.JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now
	jq.updateJobStatus(job)

	// Process the job
	result, err := processor.ProcessJob(ctx, job)

	// Create job result
	jobResult := &JobResult{
		Job:    job,
		Result: result,
		Error:  err,
	}

	// Send result for processing
	select {
	case jq.results <- jobResult:
	case <-ctx.Done():
		return ctx.Err()
	}

	return err
}

// processResults processes job results
func (jq *JobQueue) processResults() {
	defer jq.wg.Done()

	for result := range jq.results {
		if result.Error != nil {
			jq.handleJobError(result.Job, result.Error)
		} else {
			jq.handleJobSuccess(result.Job, result.Result)
		}
	}
}

// processRetries processes retry queue
func (jq *JobQueue) processRetries(ctx context.Context, processor JobProcessor) {
	defer jq.wg.Done()

	ticker := time.NewTicker(jq.retryDelay)
	defer ticker.Stop()

	for {
		select {
		case job := <-jq.retryQueue:
			if job != nil {
				jq.logger.Info("Retrying job", "job_id", job.ID, "retry_count", job.RetryCount)
				if err := jq.workers.Submit(job); err != nil {
					jq.logger.Error("Failed to resubmit job for retry", "job_id", job.ID, "error", err)
				}
			}
		case <-ticker.C:
			// Periodic cleanup or maintenance could go here
		case <-ctx.Done():
			return
		}
	}
}

// handleJobSuccess handles successful job completion
func (jq *JobQueue) handleJobSuccess(job *domain.PrintJob, result *domain.RenderResult) {
	job.Status = domain.JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now
	job.Error = ""

	if result != nil {
		job.OutputPath = result.OutputPath
	}

	jq.updateJobStatus(job)
	jq.logger.Info("Job completed successfully", "job_id", job.ID)
}

// handleJobError handles job processing errors
func (jq *JobQueue) handleJobError(job *domain.PrintJob, err error) {
	job.RetryCount++
	job.Error = err.Error()

	if job.RetryCount < jq.maxRetries {
		jq.logger.Warn("Job failed, scheduling retry", "job_id", job.ID, "retry_count", job.RetryCount, "error", err)

		// Schedule for retry
		select {
		case jq.retryQueue <- job:
		default:
			jq.logger.Error("Retry queue full, marking job as failed", "job_id", job.ID)
			jq.markJobFailed(job)
		}
	} else {
		jq.logger.Error("Job failed after max retries", "job_id", job.ID, "retry_count", job.RetryCount, "error", err)
		jq.markJobFailed(job)
	}
}

// markJobFailed marks a job as permanently failed
func (jq *JobQueue) markJobFailed(job *domain.PrintJob) {
	job.Status = domain.JobStatusFailed
	now := time.Now()
	job.CompletedAt = &now
	jq.updateJobStatus(job)
}

// updateJobStatus updates job status in memory and storage
func (jq *JobQueue) updateJobStatus(job *domain.PrintJob) {
	jq.statusMutex.Lock()
	jq.jobStatus[job.ID] = job
	jq.statusMutex.Unlock()

	if err := jq.storage.UpdateJob(job); err != nil {
		jq.logger.Error("Failed to update job in storage", "job_id", job.ID, "error", err)
	}
}

// JobProcessor interface for processing jobs
type JobProcessor interface {
	ProcessJob(ctx context.Context, job *domain.PrintJob) (*domain.RenderResult, error)
}

// QueueStats represents job queue statistics
type QueueStats struct {
	TotalJobs      int            `json:"total_jobs"`
	PendingJobs    int            `json:"pending_jobs"`
	ProcessingJobs int            `json:"processing_jobs"`
	CompletedJobs  int            `json:"completed_jobs"`
	FailedJobs     int            `json:"failed_jobs"`
	CancelledJobs  int            `json:"cancelled_jobs"`
	QueuedJobs     int            `json:"queued_jobs"`
	WorkerStats    pool.PoolStats `json:"worker_stats"`
}
