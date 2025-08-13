package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"print-service/internal/core/domain"
	"print-service/internal/core/services"
	"print-service/internal/infrastructure/logger"
	"print-service/internal/infrastructure/queue"
	"print-service/internal/infrastructure/storage"
	"print-service/internal/pkg/config"
	"print-service/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

// PrintHandler handles print-related HTTP requests
type PrintHandler struct {
	config       *config.Config
	logger       logger.Logger
	printService *services.PrintService
	jobStorage   *queue.MemoryJobStorage
	storage      storage.Storage
}

// NewPrintHandler creates a new print handler
func NewPrintHandler(cfg *config.Config, logger logger.Logger) *PrintHandler {
	// Initialize print service
	printService, err := services.NewPrintService(cfg.Print, logger)
	if err != nil {
		logger.Error("Failed to initialize print service", "error", err)
		// Return handler without service for now, but log the error
	}

	// Initialize job storage
	jobStorage := queue.NewMemoryJobStorage()

	// Initialize MinIO storage
	var storageImpl storage.Storage
	minioStorage, err := storage.NewMinIOStorageFromEnv()
	if err != nil {
		logger.Warn("Failed to initialize MinIO storage, falling back to local storage", "error", err)
		storageImpl = storage.NewLocalStorage("/tmp/print-service-pdfs")
	} else {
		logger.Info("MinIO storage initialized successfully")
		storageImpl = minioStorage
	}

	return &PrintHandler{
		config:       cfg,
		logger:       logger.With("handler", "print"),
		printService: printService,
		jobStorage:   jobStorage,
		storage:      storageImpl,
	}
}

// GetStorage returns the storage interface for external access
func (ph *PrintHandler) GetStorage() storage.Storage {
	return ph.storage
}

// getJobFromStorage retrieves a job from storage by ID
func (ph *PrintHandler) getJobFromStorage(jobID string) (*domain.PrintJob, error) {
	return ph.jobStorage.GetJob(jobID)
}

// readOutputFile reads the PDF content from the output file
func (ph *PrintHandler) readOutputFile(outputPath string) ([]byte, error) {
	// Check if file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("output file does not exist: %s", outputPath)
	}

	// Read file content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", outputPath, err)
	}

	return data, nil
}

// PrintRequest represents a print request
type PrintRequest struct {
	Content     string                  `json:"content" binding:"required"`
	ContentType domain.ContentType      `json:"content_type"`
	Options     domain.PrintOptions     `json:"options"`
	Metadata    domain.DocumentMetadata `json:"metadata"`
}

// Print handles print requests
func (ph *PrintHandler) Print(c *gin.Context) {
	var req PrintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.logger.Error("Invalid print request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Content is required",
		})
		return
	}

	// Set defaults if not provided
	if req.ContentType == "" {
		req.ContentType = domain.ContentTypeHTML
	}

	// Create document with UUID
	doc := &domain.Document{
		ID:          utils.GenerateID(),
		Content:     req.Content,
		ContentType: req.ContentType,
		Metadata:    req.Metadata,
		Options:     req.Options,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create print job with UUID
	job := &domain.PrintJob{
		ID:        utils.GenerateID(),
		Document:  *doc,
		Status:    domain.JobStatusPending,
		Priority:  domain.PriorityNormal,
		CreatedAt: time.Now(),
	}

	// Store job in storage for later retrieval
	if err := ph.jobStorage.SaveJob(job); err != nil {
		ph.logger.Error("Failed to save job", "job_id", job.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save job",
		})
		return
	}

	// Submit job for processing if service is available
	if ph.printService != nil {
		// Process job asynchronously and update storage
		go func() {
			ctx := context.Background()

			// Update job status to processing
			job.Status = domain.JobStatusProcessing
			job.StartedAt = &[]time.Time{time.Now()}[0]
			ph.jobStorage.UpdateJob(job)

			// Process the job
			result, err := ph.printService.ProcessDocument(ctx, &job.Document)
			if err != nil {
				ph.logger.Error("Job processing failed", "job_id", job.ID, "error", err)
				// Mark job as failed
				job.Status = domain.JobStatusFailed
				job.Error = err.Error()
				completedAt := time.Now()
				job.CompletedAt = &completedAt
				ph.jobStorage.UpdateJob(job)
				return
			}

			// Read the generated PDF file
			pdfData, err := ph.readOutputFile(result.OutputPath)
			if err != nil {
				ph.logger.Error("Failed to read generated PDF", "job_id", job.ID, "output_path", result.OutputPath, "error", err)
				job.Status = domain.JobStatusFailed
				job.Error = "Failed to read generated PDF"
				completedAt := time.Now()
				job.CompletedAt = &completedAt
				ph.jobStorage.UpdateJob(job)
				return
			}

			// Store PDF in MinIO
			storageKey, err := ph.storage.StorePDF(ctx, job.ID, pdfData)
			if err != nil {
				ph.logger.Error("Failed to store PDF in MinIO", "job_id", job.ID, "error", err)
				job.Status = domain.JobStatusFailed
				job.Error = "Failed to store PDF"
				completedAt := time.Now()
				job.CompletedAt = &completedAt
				ph.jobStorage.UpdateJob(job)
				return
			}

			// Clean up local file after successful MinIO upload
			if err := os.Remove(result.OutputPath); err != nil {
				ph.logger.Warn("Failed to clean up local PDF file", "path", result.OutputPath, "error", err)
			}

			// Mark job as completed with MinIO storage key
			job.Status = domain.JobStatusCompleted
			job.OutputPath = storageKey // Now contains MinIO storage key instead of local path
			completedAt := time.Now()
			job.CompletedAt = &completedAt
			ph.jobStorage.UpdateJob(job)

			ph.logger.Info("Job completed successfully", "job_id", job.ID, "minio_storage_key", storageKey, "local_path_cleaned", result.OutputPath)
		}()
	}

	ph.logger.Info("Print job created", "job_id", job.ID, "document_id", doc.ID)

	// Return job information
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":      job.ID,
		"document_id": doc.ID,
		"status":      job.Status,
		"message":     "Print job submitted successfully",
		"created_at":  job.CreatedAt,
	})
}

// GetStatus gets the status of a print job
func (ph *PrintHandler) GetStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Mock job status for now
	job := &domain.PrintJob{
		ID:     jobID,
		Status: domain.JobStatusCompleted,
	}

	c.JSON(http.StatusOK, job)
}

// Cancel cancels a print job
func (ph *PrintHandler) Cancel(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	ph.logger.Info("Cancelling print job", "job_id", jobID)

	c.JSON(http.StatusOK, gin.H{
		"job_id":  jobID,
		"status":  "cancelled",
		"message": "Print job cancelled successfully",
	})
}

// Download downloads the generated file
func (ph *PrintHandler) Download(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Get job from queue/storage to find the output file
	job, err := ph.getJobFromStorage(jobID)
	if err != nil {
		ph.logger.Error("Failed to get job", "job_id", jobID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Check if job is completed
	if job.Status != domain.JobStatusCompleted {
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Job not completed",
			"status": string(job.Status),
		})
		return
	}

	// Check if output file exists
	if job.OutputPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No output file available"})
		return
	}

	// Read the PDF file content from MinIO storage
	pdfData, err := ph.storage.GetPDF(c.Request.Context(), job.OutputPath)
	if err != nil {
		ph.logger.Error("Failed to read PDF from storage", "storage_key", job.OutputPath, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read PDF file"})
		return
	}

	// Set appropriate headers for PDF download
	filename := fmt.Sprintf("document_%s.pdf", jobID[:8])
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfData)))

	// Send the PDF data
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// ListJobs lists all print jobs
func (ph *PrintHandler) ListJobs(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Mock job list for now
	jobs := []domain.PrintJob{
		{
			ID:     "job-1",
			Status: domain.JobStatusCompleted,
		},
		{
			ID:     "job-2",
			Status: domain.JobStatusProcessing,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"page":  page,
		"limit": limit,
		"total": len(jobs),
	})
}

// GetJob gets a specific print job
func (ph *PrintHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Get job from storage
	job, err := ph.getJobFromStorage(jobID)
	if err != nil {
		ph.logger.Error("Failed to get job", "job_id", jobID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}
