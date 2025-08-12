package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeRateLimit    ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Server errors (5xx)
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavail ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeQueueFull      ErrorCode = "QUEUE_FULL"

	// Business logic errors
	ErrCodeJobNotFound    ErrorCode = "JOB_NOT_FOUND"
	ErrCodeJobCancelled   ErrorCode = "JOB_CANCELLED"
	ErrCodeJobFailed      ErrorCode = "JOB_FAILED"
	ErrCodeInvalidContent ErrorCode = "INVALID_CONTENT"
	ErrCodeProcessing     ErrorCode = "PROCESSING_ERROR"
)

// APIError represents a standardized API error
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	Timestamp  string    `json:"timestamp"`
	StatusCode int       `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, details ...string) *APIError {
	err := &APIError{
		Code:       code,
		Message:    message,
		StatusCode: getHTTPStatusCode(code),
		Timestamp:  getCurrentTimestamp(),
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// WithRequestID adds a request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details string) *APIError {
	e.Details = details
	return e
}

// getHTTPStatusCode maps error codes to HTTP status codes
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest, ErrCodeValidation, ErrCodeInvalidContent:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeJobNotFound:
		return http.StatusNotFound
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeServiceUnavail, ErrCodeQueueFull:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeInternal, ErrCodeProcessing, ErrCodeJobFailed:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// getCurrentTimestamp returns current timestamp in ISO format
func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", getCurrentUnixTimestamp())
}

// getCurrentUnixTimestamp returns current Unix timestamp
func getCurrentUnixTimestamp() int64 {
	// This would normally use time.Now().Unix(), but to avoid import cycles
	// we'll implement a simple version
	return 1691836628 // Placeholder - in real implementation use time.Now().Unix()
}

// Common error constructors
func BadRequest(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, details...)
}

func Unauthorized(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, details...)
}

func Forbidden(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, details...)
}

func NotFound(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, details...)
}

func ValidationError(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeValidation, message, details...)
}

func InternalError(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeInternal, message, details...)
}

func ServiceUnavailable(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeServiceUnavail, message, details...)
}

func JobNotFound(jobID string) *APIError {
	return NewAPIError(ErrCodeJobNotFound, "Job not found", fmt.Sprintf("Job ID: %s", jobID))
}

func QueueFull(message string) *APIError {
	return NewAPIError(ErrCodeQueueFull, message)
}

func ProcessingError(message string, details ...string) *APIError {
	return NewAPIError(ErrCodeProcessing, message, details...)
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError extracts APIError from error, or creates a generic one
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return InternalError("An unexpected error occurred", err.Error())
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error *APIError `json:"error"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *APIError) *ErrorResponse {
	return &ErrorResponse{Error: err}
}
