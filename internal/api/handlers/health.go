package handlers

import (
	"net/http"

	"print-service/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger.With("handler", "health"),
	}
}

// Health returns the health status of the service
func (hh *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "print-service",
		"version": "1.0.0",
	})
}

// Ready returns the readiness status of the service
func (hh *HealthHandler) Ready(c *gin.Context) {
	// In a real implementation, this would check dependencies
	// like database connections, external services, etc.

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"database": "ok",
			"cache":    "ok",
			"storage":  "ok",
		},
	})
}
