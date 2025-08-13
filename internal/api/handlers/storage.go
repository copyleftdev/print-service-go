package handlers

import (
	"net/http"
	"time"

	"print-service/internal/infrastructure/logger"
	"print-service/internal/infrastructure/storage"

	"github.com/gin-gonic/gin"
)

// StorageHandler handles storage-related HTTP requests
type StorageHandler struct {
	logger  logger.Logger
	storage storage.Storage
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(logger logger.Logger, storage storage.Storage) *StorageHandler {
	return &StorageHandler{
		logger:  logger.With("handler", "storage"),
		storage: storage,
	}
}

// HealthCheck checks MinIO storage connectivity
func (sh *StorageHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	
	err := sh.storage.HealthCheck(ctx)
	if err != nil {
		sh.logger.Error("Storage health check failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"type":   "minio",
	})
}

// GetStats returns storage statistics
func (sh *StorageHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()
	
	stats, err := sh.storage.GetStats(ctx)
	if err != nil {
		sh.logger.Error("Failed to get storage stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get storage statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetPDFURL generates a presigned URL for PDF access
func (sh *StorageHandler) GetPDFURL(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Get expiry from query parameter (default 1 hour)
	expiryStr := c.DefaultQuery("expiry", "1h")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid expiry duration format",
			"hint":  "Use format like '1h', '30m', '24h'",
		})
		return
	}

	// For this endpoint, we need to find the storage key for the job
	// This would typically require looking up the job in storage
	// For now, we'll assume the storage key format
	storageKey := c.Query("storage_key")
	if storageKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Storage key is required",
			"hint":  "Provide storage_key query parameter",
		})
		return
	}

	ctx := c.Request.Context()
	url, err := sh.storage.GetPDFURL(ctx, storageKey, expiry)
	if err != nil {
		sh.logger.Error("Failed to generate presigned URL", "storage_key", storageKey, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate download URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":      jobID,
		"storage_key": storageKey,
		"url":         url,
		"expires_in":  expiry.String(),
		"expires_at":  time.Now().Add(expiry).Format(time.RFC3339),
	})
}

// ListPDFs lists stored PDFs with optional prefix filter
func (sh *StorageHandler) ListPDFs(c *gin.Context) {
	prefix := c.DefaultQuery("prefix", "pdfs/")
	
	ctx := c.Request.Context()
	objects, err := sh.storage.ListPDFs(ctx, prefix)
	if err != nil {
		sh.logger.Error("Failed to list PDFs", "prefix", prefix, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list PDFs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prefix":     prefix,
		"total":      len(objects),
		"pdf_files":  objects,
	})
}
