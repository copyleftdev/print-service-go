package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"print-service/internal/infrastructure/logger"
)

// Logging returns a middleware that logs HTTP requests
func Logging(logger logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log fields
		fields := []interface{}{
			"method", c.Request.Method,
			"path", path,
			"status", statusCode,
			"latency", latency,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}

		if raw != "" {
			fields = append(fields, "query", raw)
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("HTTP request", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}
	})
}
