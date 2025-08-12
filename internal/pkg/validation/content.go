package validation

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"print-service/internal/pkg/errors"
)

const (
	// Based on rigor test findings
	MaxContentSize     = 10 * 1024 * 1024 // 10MB max content size
	MinContentSize     = 1                // Minimum 1 character
	MaxTitleLength     = 500              // Maximum title length
	MaxMetadataEntries = 50               // Maximum metadata entries
)

// ContentValidator validates document content and metadata
type ContentValidator struct{}

// NewContentValidator creates a new content validator
func NewContentValidator() *ContentValidator {
	return &ContentValidator{}
}

// ValidateContent validates document content based on rigor test findings
func (v *ContentValidator) ValidateContent(content, contentType string) error {
	// Check for empty content (rigor test revealed this edge case)
	if strings.TrimSpace(content) == "" {
		return errors.ValidationError(
			"Content cannot be empty",
			"Document content must contain at least one non-whitespace character",
		)
	}

	// Check content size limits (based on large content test results)
	contentSize := len(content)
	if contentSize < MinContentSize {
		return errors.ValidationError(
			"Content too small",
			fmt.Sprintf("Content must be at least %d characters", MinContentSize),
		)
	}

	if contentSize > MaxContentSize {
		return errors.ValidationError(
			"Content too large",
			fmt.Sprintf("Content exceeds maximum size of %d bytes (%d MB)",
				MaxContentSize, MaxContentSize/(1024*1024)),
		)
	}

	// Validate UTF-8 encoding
	if !utf8.ValidString(content) {
		return errors.ValidationError(
			"Invalid content encoding",
			"Content must be valid UTF-8 encoded text",
		)
	}

	// Content type specific validation
	switch strings.ToLower(contentType) {
	case "html":
		return v.validateHTML(content)
	case "markdown", "md":
		return v.validateMarkdown(content)
	case "text", "plain":
		return v.validatePlainText(content)
	default:
		return errors.ValidationError(
			"Unsupported content type",
			fmt.Sprintf("Content type '%s' is not supported. Use 'html', 'markdown', or 'text'", contentType),
		)
	}
}

// validateHTML performs HTML-specific validation
func (v *ContentValidator) validateHTML(content string) error {
	// Check for basic HTML structure
	content = strings.ToLower(content)

	// Allow fragments, but warn about missing structure
	hasHTML := strings.Contains(content, "<html")
	hasBody := strings.Contains(content, "<body")

	if !hasHTML && !hasBody && len(content) > 1000 {
		// For large content without proper structure, suggest adding HTML tags
		return errors.ValidationError(
			"Large HTML content missing structure",
			"Large HTML documents should include <html> and <body> tags for better processing",
		)
	}

	// Check for potentially problematic patterns (based on security test insights)
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"data:text/html",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(content, pattern) {
			return errors.ValidationError(
				"Potentially unsafe HTML content",
				fmt.Sprintf("Content contains potentially unsafe pattern: %s", pattern),
			)
		}
	}

	return nil
}

// validateMarkdown performs Markdown-specific validation
func (v *ContentValidator) validateMarkdown(content string) error {
	// Basic Markdown validation
	if strings.Count(content, "```")%2 != 0 {
		return errors.ValidationError(
			"Invalid Markdown code blocks",
			"Markdown code blocks (```) must be properly closed",
		)
	}

	return nil
}

// validatePlainText performs plain text validation
func (v *ContentValidator) validatePlainText(content string) error {
	// Check for excessive line breaks or whitespace
	lines := strings.Split(content, "\n")
	emptyLines := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLines++
		}
	}

	// If more than 50% empty lines, might be formatting issue
	if len(lines) > 10 && float64(emptyLines)/float64(len(lines)) > 0.5 {
		return errors.ValidationError(
			"Excessive empty lines in text content",
			"Text content contains too many empty lines, which may indicate formatting issues",
		)
	}

	return nil
}

// ValidateMetadata validates document metadata
func (v *ContentValidator) ValidateMetadata(metadata map[string]interface{}) error {
	if len(metadata) > MaxMetadataEntries {
		return errors.ValidationError(
			"Too many metadata entries",
			fmt.Sprintf("Metadata cannot exceed %d entries", MaxMetadataEntries),
		)
	}

	// Validate title if present
	if title, exists := metadata["title"]; exists {
		if titleStr, ok := title.(string); ok {
			if len(titleStr) > MaxTitleLength {
				return errors.ValidationError(
					"Title too long",
					fmt.Sprintf("Title cannot exceed %d characters", MaxTitleLength),
				)
			}
		}
	}

	return nil
}

// GetContentStats returns statistics about the content
func (v *ContentValidator) GetContentStats(content string) map[string]interface{} {
	lines := strings.Split(content, "\n")
	words := strings.Fields(content)

	return map[string]interface{}{
		"size_bytes":    len(content),
		"size_mb":       float64(len(content)) / (1024 * 1024),
		"lines":         len(lines),
		"words":         len(words),
		"characters":    utf8.RuneCountInString(content),
		"is_valid_utf8": utf8.ValidString(content),
		"usage_percent": float64(len(content)) / float64(MaxContentSize) * 100,
	}
}
