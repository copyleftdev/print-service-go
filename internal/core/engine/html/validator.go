package html

import (
	"fmt"
	"regexp"
	"strings"

	"print-service/internal/core/domain"
)

// Validator handles HTML validation
type Validator struct {
	strictMode bool
}

// NewValidator creates a new HTML validator
func NewValidator(strictMode bool) *Validator {
	return &Validator{
		strictMode: strictMode,
	}
}

// Validate validates HTML content
func (v *Validator) Validate(content string) error {
	if content == "" {
		return domain.NewPrintError(domain.ErrCodeInvalidInput, "empty HTML content", domain.ErrInvalidHTML)
	}

	// Basic structure validation
	if err := v.validateBasicStructure(content); err != nil {
		return err
	}

	// Tag validation
	if err := v.validateTags(content); err != nil {
		return err
	}

	// Attribute validation
	if err := v.validateAttributes(content); err != nil {
		return err
	}

	return nil
}

// validateBasicStructure validates basic HTML structure
func (v *Validator) validateBasicStructure(content string) error {
	content = strings.TrimSpace(content)

	// Check for basic HTML structure if in strict mode
	if v.strictMode {
		if !strings.Contains(strings.ToLower(content), "<html") {
			return domain.NewPrintError(domain.ErrCodeInvalidInput, "missing <html> tag", domain.ErrInvalidHTML)
		}
		if !strings.Contains(strings.ToLower(content), "<head") {
			return domain.NewPrintError(domain.ErrCodeInvalidInput, "missing <head> tag", domain.ErrInvalidHTML)
		}
		if !strings.Contains(strings.ToLower(content), "<body") {
			return domain.NewPrintError(domain.ErrCodeInvalidInput, "missing <body> tag", domain.ErrInvalidHTML)
		}
	}

	return nil
}

// validateTags validates HTML tags
func (v *Validator) validateTags(content string) error {
	// Find all tags
	tagRegex := regexp.MustCompile(`<\s*/?([a-zA-Z][a-zA-Z0-9]*)\s*[^>]*>`)
	matches := tagRegex.FindAllStringSubmatch(content, -1)

	tagStack := make([]string, 0)
	selfClosingTags := getSelfClosingTags()

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullTag := match[0]
		tagName := strings.ToLower(match[1])

		// Check if it's a closing tag
		if strings.HasPrefix(fullTag, "</") {
			// Closing tag
			if len(tagStack) == 0 {
				return domain.NewPrintError(domain.ErrCodeInvalidInput, 
					fmt.Sprintf("unexpected closing tag: %s", tagName), domain.ErrInvalidHTML)
			}

			// Check if it matches the last opened tag
			lastTag := tagStack[len(tagStack)-1]
			if lastTag != tagName {
				return domain.NewPrintError(domain.ErrCodeInvalidInput, 
					fmt.Sprintf("mismatched closing tag: expected %s, got %s", lastTag, tagName), domain.ErrInvalidHTML)
			}

			// Remove from stack
			tagStack = tagStack[:len(tagStack)-1]
		} else {
			// Opening tag or self-closing tag
			if selfClosingTags[tagName] || strings.HasSuffix(fullTag, "/>") {
				// Self-closing tag, don't add to stack
				continue
			}

			// Regular opening tag
			tagStack = append(tagStack, tagName)
		}
	}

	// Check for unclosed tags
	if len(tagStack) > 0 {
		return domain.NewPrintError(domain.ErrCodeInvalidInput, 
			fmt.Sprintf("unclosed tags: %v", tagStack), domain.ErrInvalidHTML)
	}

	return nil
}

// validateAttributes validates HTML attributes
func (v *Validator) validateAttributes(content string) error {
	// Find all tags with attributes
	attrRegex := regexp.MustCompile(`<[a-zA-Z][a-zA-Z0-9]*\s+([^>]+)>`)
	matches := attrRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		attributes := match[1]
		if err := v.validateAttributeString(attributes); err != nil {
			return err
		}
	}

	return nil
}

// validateAttributeString validates a string of attributes
func (v *Validator) validateAttributeString(attributes string) error {
	// Simple attribute validation
	attrRegex := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9\-]*)\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	matches := attrRegex.FindAllStringSubmatch(attributes, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		attrName := strings.ToLower(match[1])
		attrValue := match[2]

		// Remove quotes from value
		if (strings.HasPrefix(attrValue, `"`) && strings.HasSuffix(attrValue, `"`)) ||
		   (strings.HasPrefix(attrValue, `'`) && strings.HasSuffix(attrValue, `'`)) {
			attrValue = attrValue[1 : len(attrValue)-1]
		}

		// Validate specific attributes
		if err := v.validateSpecificAttribute(attrName, attrValue); err != nil {
			return err
		}
	}

	return nil
}

// validateSpecificAttribute validates specific attribute types
func (v *Validator) validateSpecificAttribute(name, value string) error {
	switch name {
	case "id":
		if !isValidIdentifier(value) {
			return domain.NewPrintError(domain.ErrCodeInvalidInput, 
				fmt.Sprintf("invalid id attribute: %s", value), domain.ErrInvalidHTML)
		}
	case "class":
		classes := strings.Fields(value)
		for _, class := range classes {
			if !isValidIdentifier(class) {
				return domain.NewPrintError(domain.ErrCodeInvalidInput, 
					fmt.Sprintf("invalid class name: %s", class), domain.ErrInvalidHTML)
			}
		}
	case "href", "src", "action":
		if err := v.validateURL(value); err != nil {
			return err
		}
	}

	return nil
}

// validateURL validates URL attributes
func (v *Validator) validateURL(url string) error {
	if url == "" {
		return nil
	}

	// Check for dangerous protocols
	dangerousProtocols := []string{"javascript:", "vbscript:", "data:text/html"}
	lowerURL := strings.ToLower(url)
	
	for _, protocol := range dangerousProtocols {
		if strings.HasPrefix(lowerURL, protocol) {
			return domain.NewPrintError(domain.ErrCodeSecurity, 
				fmt.Sprintf("dangerous URL protocol: %s", protocol), domain.ErrUnsafeContent)
		}
	}

	return nil
}

// isValidIdentifier checks if a string is a valid CSS identifier
func isValidIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}

	// CSS identifier pattern: starts with letter, underscore, or hyphen,
	// followed by letters, digits, hyphens, or underscores
	pattern := regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]*$`)
	return pattern.MatchString(identifier)
}

// getSelfClosingTags returns a map of self-closing HTML tags
func getSelfClosingTags() map[string]bool {
	return map[string]bool{
		"area":   true,
		"base":   true,
		"br":     true,
		"col":    true,
		"embed":  true,
		"hr":     true,
		"img":    true,
		"input":  true,
		"link":   true,
		"meta":   true,
		"param":  true,
		"source": true,
		"track":  true,
		"wbr":    true,
	}
}

// ValidationResult represents the result of HTML validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ValidateWithResult validates HTML and returns detailed results
func (v *Validator) ValidateWithResult(content string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	if err := v.Validate(content); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Add warnings for common issues
	v.addWarnings(content, result)

	return result
}

// addWarnings adds warnings for common HTML issues
func (v *Validator) addWarnings(content string, result *ValidationResult) {
	// Check for missing alt attributes on images
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	images := imgRegex.FindAllString(content, -1)
	for _, img := range images {
		if !strings.Contains(strings.ToLower(img), "alt=") {
			result.Warnings = append(result.Warnings, "Image missing alt attribute for accessibility")
		}
	}

	// Check for missing title
	if !strings.Contains(strings.ToLower(content), "<title>") {
		result.Warnings = append(result.Warnings, "Document missing title element")
	}

	// Check for inline styles
	if strings.Contains(content, "style=") {
		result.Warnings = append(result.Warnings, "Inline styles detected, consider using external CSS")
	}
}
