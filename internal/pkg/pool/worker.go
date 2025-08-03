package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"print-service/internal/infrastructure/logger"
)

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	size    int
	jobs    chan interface{}
	results chan Result
	workers []*Worker
	wg      sync.WaitGroup
	logger  logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

// Worker represents a single worker
type Worker struct {
	id     int
	pool   *WorkerPool
	logger logger.Logger
}

// Result represents the result of a job
type Result struct {
	Job   interface{}
	Error error
}

// JobHandler defines the function signature for job processing
type JobHandler func(job interface{}) error

// NewWorkerPool creates a new worker pool
func NewWorkerPool(size int, logger logger.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		size:    size,
		jobs:    make(chan interface{}, size*2), // Buffer for jobs
		results: make(chan Result, size*2),     // Buffer for results
		workers: make([]*Worker, 0, size),
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context, handler JobHandler) {
	wp.logger.Info("Starting worker pool", "size", wp.size)

	// Create and start workers
	for i := 0; i < wp.size; i++ {
		worker := &Worker{
			id:     i + 1,
			pool:   wp,
			logger: wp.logger.With("worker_id", i+1),
		}
		wp.workers = append(wp.workers, worker)
		
		wp.wg.Add(1)
		go worker.start(ctx, handler)
	}

	// Start result processor
	go wp.processResults()
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job interface{}) error {
	select {
	case wp.jobs <- job:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	default:
		wp.logger.Warn("Job queue is full, job rejected")
		return ErrQueueFull
	}
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop(ctx context.Context) {
	wp.logger.Info("Stopping worker pool")
	
	// Close job channel to signal workers to stop
	close(wp.jobs)
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		wp.logger.Info("All workers stopped gracefully")
	case <-ctx.Done():
		wp.logger.Warn("Worker pool stop timeout, forcing shutdown")
		wp.cancel()
		wp.wg.Wait()
	}
	
	// Close results channel
	close(wp.results)
}

// GetStats returns worker pool statistics
func (wp *WorkerPool) GetStats() PoolStats {
	return PoolStats{
		Size:        wp.size,
		ActiveJobs:  len(wp.jobs),
		QueuedJobs:  cap(wp.jobs) - len(wp.jobs),
		WorkerCount: len(wp.workers),
	}
}

// start starts a worker
func (w *Worker) start(ctx context.Context, handler JobHandler) {
	defer w.pool.wg.Done()
	
	w.logger.Debug("Worker started")
	defer w.logger.Debug("Worker stopped")
	
	for {
		select {
		case job, ok := <-w.pool.jobs:
			if !ok {
				// Job channel closed, worker should exit
				return
			}
			
			// Process the job
			w.logger.Debug("Processing job", "job_type", getJobType(job))
			start := time.Now()
			
			err := handler(job)
			
			duration := time.Since(start)
			if err != nil {
				w.logger.Error("Job failed", "error", err, "duration", duration)
			} else {
				w.logger.Debug("Job completed", "duration", duration)
			}
			
			// Send result
			select {
			case w.pool.results <- Result{Job: job, Error: err}:
			case <-ctx.Done():
				return
			}
			
		case <-ctx.Done():
			w.logger.Debug("Worker context cancelled")
			return
		}
	}
}

// processResults processes job results
func (wp *WorkerPool) processResults() {
	for result := range wp.results {
		if result.Error != nil {
			wp.logger.Error("Job result error", "error", result.Error)
		}
		// Additional result processing can be added here
	}
}

// getJobType returns a string representation of the job type
func getJobType(job interface{}) string {
	if job == nil {
		return "nil"
	}
	return "unknown" // In a real implementation, this would use reflection or type assertions
}

// PoolStats represents worker pool statistics
type PoolStats struct {
	Size        int `json:"size"`
	ActiveJobs  int `json:"active_jobs"`
	QueuedJobs  int `json:"queued_jobs"`
	WorkerCount int `json:"worker_count"`
}

// Custom errors
var (
	ErrQueueFull = fmt.Errorf("job queue is full")
	ErrPoolStopped = fmt.Errorf("worker pool is stopped")
)
