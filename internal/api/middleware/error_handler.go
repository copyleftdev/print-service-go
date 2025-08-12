package middleware

import (
	"time"

	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/errors"
	"print-service/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns error handling middleware
func ErrorHandler(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateShortID()
			c.Header("X-Request-ID", requestID)
		}

		// Set request ID in context
		c.Set("request_id", requestID)

		// Process request
		c.Next()

		// Handle any errors that occurred during processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err.Err, requestID, logger)
		}
	}
}

// handleError handles different types of errors and returns appropriate responses
func handleError(c *gin.Context, err error, requestID string, logger logger.Logger) {
	// Check if it's already an APIError
	if apiErr, ok := err.(*errors.APIError); ok {
		apiErr.RequestID = requestID
		apiErr.Timestamp = time.Now().Format(time.RFC3339)

		logger.Error("API error occurred",
			"request_id", requestID,
			"error_code", apiErr.Code,
			"message", apiErr.Message,
			"details", apiErr.Details,
			"status_code", apiErr.StatusCode,
		)

		c.JSON(apiErr.StatusCode, errors.NewErrorResponse(apiErr))
		return
	}

	// Convert regular errors to APIError
	apiErr := errors.GetAPIError(err)
	apiErr.RequestID = requestID
	apiErr.Timestamp = time.Now().Format(time.RFC3339)

	logger.Error("Unexpected error occurred",
		"request_id", requestID,
		"error", err.Error(),
		"status_code", apiErr.StatusCode,
	)

	c.JSON(apiErr.StatusCode, errors.NewErrorResponse(apiErr))
}

// AbortWithError aborts the request with a standardized error
func AbortWithError(c *gin.Context, err *errors.APIError) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = utils.GenerateShortID()
	}

	err.RequestID = requestID
	err.Timestamp = time.Now().Format(time.RFC3339)

	c.JSON(err.StatusCode, errors.NewErrorResponse(err))
	c.Abort()
}

// AbortWithBadRequest is a convenience function for bad request errors
func AbortWithBadRequest(c *gin.Context, message string, details ...string) {
	AbortWithError(c, errors.BadRequest(message, details...))
}

// AbortWithUnauthorized is a convenience function for unauthorized errors
func AbortWithUnauthorized(c *gin.Context, message string, details ...string) {
	AbortWithError(c, errors.Unauthorized(message, details...))
}

// AbortWithNotFound is a convenience function for not found errors
func AbortWithNotFound(c *gin.Context, message string, details ...string) {
	AbortWithError(c, errors.NotFound(message, details...))
}

// AbortWithInternalError is a convenience function for internal errors
func AbortWithInternalError(c *gin.Context, message string, details ...string) {
	AbortWithError(c, errors.InternalError(message, details...))
}
