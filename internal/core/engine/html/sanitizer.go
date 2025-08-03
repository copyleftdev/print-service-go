package html

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"print-service/internal/core/domain"
)

// Sanitizer handles HTML sanitization for security
type Sanitizer struct {
	allowedTags       map[string]bool
	allowedAttributes map[string]map[string]bool
	urlValidator      *URLValidator
}

// NewSanitizer creates a new HTML sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		allowedTags:       getDefaultAllowedTags(),
		allowedAttributes: getDefaultAllowedAttributes(),
		urlValidator:      NewURLValidator(),
	}
}

// Sanitize sanitizes HTML content according to security options
func (s *Sanitizer) Sanitize(content string, options domain.SecurityOptions) (string, error) {
	// Parse the HTML
	parser := &Parser{}
	domNode, err := parser.Parse(content, domain.SecurityOptions{SanitizeHTML: false})
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML for sanitization: %w", err)
	}

	// Sanitize the DOM tree
	sanitized := s.sanitizeNode(domNode, options)

	// Render back to HTML
	var builder strings.Builder
	if err := sanitized.Render(&builder); err != nil {
		return "", fmt.Errorf("failed to render sanitized HTML: %w", err)
	}

	return builder.String(), nil
}

// sanitizeNode recursively sanitizes a DOM node
func (s *Sanitizer) sanitizeNode(node *DOMNode, options domain.SecurityOptions) *DOMNode {
	if node == nil {
		return nil
	}

	switch node.Type {
	case TextNode:
		// Text nodes are generally safe, but we can filter content if needed
		return &DOMNode{
			Type: TextNode,
			Data: s.sanitizeText(node.Data),
		}

	case ElementNode:
		// Check if tag is allowed
		if !s.isTagAllowed(node.Data) {
			// Convert disallowed tags to text or remove them
			return &DOMNode{
				Type: TextNode,
				Data: "", // Remove disallowed tags
			}
		}

		sanitizedNode := &DOMNode{
			Type:       ElementNode,
			Data:       strings.ToLower(node.Data),
			Attributes: make(map[string]string),
		}

		// Sanitize attributes
		for key, value := range node.Attributes {
			if s.isAttributeAllowed(node.Data, key) {
				sanitizedValue, err := s.sanitizeAttributeValue(key, value, options)
				if err == nil && sanitizedValue != "" {
					sanitizedNode.Attributes[strings.ToLower(key)] = sanitizedValue
				}
			}
		}

		// Sanitize children
		for _, child := range node.Children {
			sanitizedChild := s.sanitizeNode(child, options)
			if sanitizedChild != nil {
				sanitizedChild.Parent = sanitizedNode
				sanitizedNode.Children = append(sanitizedNode.Children, sanitizedChild)
			}
		}

		return sanitizedNode

	case CommentNode:
		// Remove comments for security
		return nil

	default:
		// Keep other node types as-is
		return node
	}
}

// isTagAllowed checks if an HTML tag is allowed
func (s *Sanitizer) isTagAllowed(tag string) bool {
	return s.allowedTags[strings.ToLower(tag)]
}

// isAttributeAllowed checks if an attribute is allowed for a given tag
func (s *Sanitizer) isAttributeAllowed(tag, attr string) bool {
	tag = strings.ToLower(tag)
	attr = strings.ToLower(attr)

	// Check tag-specific attributes
	if tagAttrs, exists := s.allowedAttributes[tag]; exists {
		if tagAttrs[attr] {
			return true
		}
	}

	// Check global attributes
	if globalAttrs, exists := s.allowedAttributes["*"]; exists {
		return globalAttrs[attr]
	}

	return false
}

// sanitizeAttributeValue sanitizes an attribute value
func (s *Sanitizer) sanitizeAttributeValue(attr, value string, options domain.SecurityOptions) (string, error) {
	attr = strings.ToLower(attr)

	switch attr {
	case "href", "src", "action":
		return s.sanitizeURL(value, options)
	case "style":
		return s.sanitizeStyle(value), nil
	case "class", "id":
		return s.sanitizeIdentifier(value), nil
	default:
		return s.sanitizeGenericAttribute(value), nil
	}
}

// sanitizeURL sanitizes URL attributes
func (s *Sanitizer) sanitizeURL(urlStr string, options domain.SecurityOptions) (string, error) {
	if urlStr == "" {
		return "", nil
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Check for javascript: and data: schemes
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme == "javascript" || scheme == "vbscript" {
		return "", fmt.Errorf("unsafe URL scheme: %s", scheme)
	}

	// Allow data: URLs for images only in some cases
	if scheme == "data" {
		if !strings.HasPrefix(strings.ToLower(urlStr), "data:image/") {
			return "", fmt.Errorf("unsafe data URL")
		}
	}

	// Check domain restrictions
	if parsedURL.Host != "" {
		if err := s.urlValidator.ValidateDomain(parsedURL.Host, options); err != nil {
			return "", err
		}
	}

	return urlStr, nil
}

// sanitizeStyle sanitizes CSS style attributes
func (s *Sanitizer) sanitizeStyle(style string) string {
	// Remove potentially dangerous CSS
	dangerous := []string{
		"expression", "javascript:", "vbscript:", "data:",
		"@import", "behavior", "-moz-binding",
	}

	style = strings.ToLower(style)
	for _, danger := range dangerous {
		style = strings.ReplaceAll(style, danger, "")
	}

	return style
}

// sanitizeIdentifier sanitizes class and ID attributes
func (s *Sanitizer) sanitizeIdentifier(value string) string {
	// Remove potentially dangerous characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_\s]`)
	return re.ReplaceAllString(value, "")
}

// sanitizeGenericAttribute sanitizes generic attributes
func (s *Sanitizer) sanitizeGenericAttribute(value string) string {
	// Remove control characters and potentially dangerous content
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	value = re.ReplaceAllString(value, "")

	// Remove javascript: and similar
	dangerous := []string{"javascript:", "vbscript:", "data:", "about:"}
	lowerValue := strings.ToLower(value)
	for _, danger := range dangerous {
		if strings.Contains(lowerValue, danger) {
			return ""
		}
	}

	return value
}

// sanitizeText sanitizes text content
func (s *Sanitizer) sanitizeText(text string) string {
	// Remove control characters
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	return re.ReplaceAllString(text, "")
}

// getDefaultAllowedTags returns the default set of allowed HTML tags
func getDefaultAllowedTags() map[string]bool {
	return map[string]bool{
		// Document structure
		"html": true, "head": true, "body": true, "title": true,
		"meta": true, "link": true, "style": true,

		// Sections
		"header": true, "nav": true, "main": true, "section": true,
		"article": true, "aside": true, "footer": true,
		"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,

		// Grouping content
		"div": true, "p": true, "hr": true, "pre": true, "blockquote": true,
		"ol": true, "ul": true, "li": true, "dl": true, "dt": true, "dd": true,

		// Text-level semantics
		"a": true, "em": true, "strong": true, "small": true, "s": true,
		"cite": true, "q": true, "dfn": true, "abbr": true, "time": true,
		"code": true, "var": true, "samp": true, "kbd": true, "sub": true,
		"sup": true, "i": true, "b": true, "u": true, "mark": true,
		"ruby": true, "rt": true, "rp": true, "bdi": true, "bdo": true,
		"span": true, "br": true, "wbr": true,

		// Embedded content
		"img": true, "picture": true, "source": true,

		// Tabular data
		"table": true, "caption": true, "colgroup": true, "col": true,
		"tbody": true, "thead": true, "tfoot": true, "tr": true,
		"td": true, "th": true,
	}
}

// getDefaultAllowedAttributes returns the default set of allowed attributes
func getDefaultAllowedAttributes() map[string]map[string]bool {
	return map[string]map[string]bool{
		"*": {
			"class": true, "id": true, "style": true, "title": true,
			"lang": true, "dir": true, "data-*": true,
		},
		"a": {
			"href": true, "target": true, "rel": true,
		},
		"img": {
			"src": true, "alt": true, "width": true, "height": true,
			"loading": true, "decoding": true,
		},
		"table": {
			"border": true, "cellpadding": true, "cellspacing": true,
		},
		"th": {
			"scope": true, "colspan": true, "rowspan": true,
		},
		"td": {
			"colspan": true, "rowspan": true,
		},
		"ol": {
			"start": true, "type": true,
		},
		"li": {
			"value": true,
		},
	}
}

// URLValidator validates URLs against security policies
type URLValidator struct{}

// NewURLValidator creates a new URL validator
func NewURLValidator() *URLValidator {
	return &URLValidator{}
}

// ValidateDomain validates a domain against security options
func (v *URLValidator) ValidateDomain(domain string, options domain.SecurityOptions) error {
	domain = strings.ToLower(domain)

	// Check blocked domains
	for _, blocked := range options.BlockedDomains {
		if strings.Contains(domain, strings.ToLower(blocked)) {
			return fmt.Errorf("domain %s is blocked", domain)
		}
	}

	// Check allowed domains (if specified)
	if len(options.AllowedDomains) > 0 {
		allowed := false
		for _, allowedDomain := range options.AllowedDomains {
			if strings.Contains(domain, strings.ToLower(allowedDomain)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("domain %s is not in allowed list", domain)
		}
	}

	return nil
}
