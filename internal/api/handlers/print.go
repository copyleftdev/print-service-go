package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"print-service/internal/core/domain"
	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/config"
)

// PrintHandler handles print-related HTTP requests
type PrintHandler struct {
	config *config.Config
	logger logger.Logger
}

// NewPrintHandler creates a new print handler
func NewPrintHandler(cfg *config.Config, logger logger.Logger) *PrintHandler {
	return &PrintHandler{
		config: cfg,
		logger: logger.With("handler", "print"),
	}
}

// PrintRequest represents a print request
type PrintRequest struct {
	Content     string                `json:"content" binding:"required"`
	ContentType domain.ContentType    `json:"content_type"`
	Options     domain.PrintOptions   `json:"options"`
	Metadata    domain.DocumentMetadata `json:"metadata"`
}

// Print handles print requests
func (ph *PrintHandler) Print(c *gin.Context) {
	var req PrintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ph.logger.Error("Invalid print request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Create document
	doc := &domain.Document{
		ID:          generateID(),
		Content:     req.Content,
		ContentType: req.ContentType,
		Metadata:    req.Metadata,
		Options:     req.Options,
	}

	// Create print job
	job := &domain.PrintJob{
		ID:       generateID(),
		Document: *doc,
		Status:   domain.JobStatusPending,
		Priority: domain.PriorityNormal,
	}

	ph.logger.Info("Print job created", "job_id", job.ID, "document_id", doc.ID)

	// Return job information
	c.JSON(http.StatusAccepted, gin.H{
		"job_id": job.ID,
		"status": job.Status,
		"message": "Print job submitted successfully",
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
		"job_id": jobID,
		"status": "cancelled",
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

	// Mock file download for now
	c.Header("Content-Disposition", "attachment; filename=output.pdf")
	c.Header("Content-Type", "application/pdf")
	c.String(http.StatusOK, "PDF content placeholder")
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

	// Mock job for now
	job := &domain.PrintJob{
		ID:     jobID,
		Status: domain.JobStatusCompleted,
	}

	c.JSON(http.StatusOK, job)
}

// generateID generates a unique ID (simplified)
func generateID() string {
	return "id-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
