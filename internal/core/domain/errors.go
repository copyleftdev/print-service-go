package domain

import (
	"errors"
	"fmt"
)

// Domain-specific error types
var (
	// Document errors
	ErrDocumentNotFound  = errors.New("document not found")
	ErrInvalidDocument   = errors.New("invalid document")
	ErrDocumentTooLarge  = errors.New("document too large")
	ErrUnsupportedFormat = errors.New("unsupported document format")

	// Job errors
	ErrJobNotFound          = errors.New("job not found")
	ErrJobAlreadyProcessing = errors.New("job already processing")
	ErrJobCancelled         = errors.New("job was cancelled")
	ErrJobTimeout           = errors.New("job timed out")

	// Rendering errors
	ErrRenderFailed  = errors.New("rendering failed")
	ErrInvalidHTML   = errors.New("invalid HTML content")
	ErrInvalidCSS    = errors.New("invalid CSS content")
	ErrFontNotFound  = errors.New("font not found")
	ErrImageNotFound = errors.New("image not found")

	// Layout errors
	ErrLayoutFailed      = errors.New("layout calculation failed")
	ErrInvalidDimensions = errors.New("invalid dimensions")
	ErrPageBreakFailed   = errors.New("page break calculation failed")

	// Resource errors
	ErrResourceNotFound = errors.New("resource not found")
	ErrResourceTooLarge = errors.New("resource too large")
	ErrResourceTimeout  = errors.New("resource timeout")
	ErrInvalidURL       = errors.New("invalid URL")

	// Cache errors
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheExpired  = errors.New("cache entry expired")
	ErrCacheFull     = errors.New("cache is full")

	// Storage errors
	ErrStorageNotFound   = errors.New("storage item not found")
	ErrStorageFull       = errors.New("storage is full")
	ErrStoragePermission = errors.New("storage permission denied")

	// Security errors
	ErrSecurityViolation = errors.New("security violation")
	ErrUnsafeContent     = errors.New("unsafe content detected")
	ErrBlockedDomain     = errors.New("domain is blocked")
)

// PrintError represents a structured error with context
type PrintError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

func (e *PrintError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *PrintError) Unwrap() error {
	return e.Cause
}

// NewPrintError creates a new PrintError
func NewPrintError(code, message string, cause error) *PrintError {
	return &PrintError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error
func (e *PrintError) WithDetail(key string, value interface{}) *PrintError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error codes
const (
	ErrCodeInvalidInput  = "INVALID_INPUT"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeTimeout       = "TIMEOUT"
	ErrCodeResourceLimit = "RESOURCE_LIMIT"
	ErrCodeSecurity      = "SECURITY"
	ErrCodeInternal      = "INTERNAL"
	ErrCodeUnavailable   = "UNAVAILABLE"
	ErrCodeConflict      = "CONFLICT"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(e))
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var printErr *PrintError
	if errors.As(err, &printErr) {
		return printErr.Code == ErrCodeNotFound
	}

	return errors.Is(err, ErrDocumentNotFound) ||
		errors.Is(err, ErrJobNotFound) ||
		errors.Is(err, ErrResourceNotFound) ||
		errors.Is(err, ErrCacheNotFound) ||
		errors.Is(err, ErrStorageNotFound)
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	var printErr *PrintError
	if errors.As(err, &printErr) {
		return printErr.Code == ErrCodeTimeout
	}

	return errors.Is(err, ErrJobTimeout) ||
		errors.Is(err, ErrResourceTimeout)
}

// IsSecurityError checks if an error is a security error
func IsSecurityError(err error) bool {
	if err == nil {
		return false
	}

	var printErr *PrintError
	if errors.As(err, &printErr) {
		return printErr.Code == ErrCodeSecurity
	}

	return errors.Is(err, ErrSecurityViolation) ||
		errors.Is(err, ErrUnsafeContent) ||
		errors.Is(err, ErrBlockedDomain)
}
