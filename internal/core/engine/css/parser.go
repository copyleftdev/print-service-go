package css

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"print-service/internal/core/domain"
)

// Parser handles CSS parsing and stylesheet construction
type Parser struct {
	strict bool
}

// NewParser creates a new CSS parser
func NewParser(strict bool) *Parser {
	return &Parser{strict: strict}
}

// Parse parses CSS content and returns a stylesheet
func (p *Parser) Parse(content string) (*Stylesheet, error) {
	if content == "" {
		return &Stylesheet{Rules: make([]*Rule, 0)}, nil
	}

	// Remove comments
	content = p.removeComments(content)

	// Parse rules
	rules, err := p.parseRules(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSS rules: %w", err)
	}

	return &Stylesheet{
		Rules: rules,
	}, nil
}

// parseRules parses CSS rules from content
func (p *Parser) parseRules(content string) ([]*Rule, error) {
	var rules []*Rule

	// Split by closing braces to get individual rules
	ruleRegex := regexp.MustCompile(`([^{}]+)\{([^{}]*)\}`)
	matches := ruleRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		selectorText := strings.TrimSpace(match[1])
		declarationsText := strings.TrimSpace(match[2])

		// Parse selectors
		selectors, err := p.parseSelectors(selectorText)
		if err != nil {
			if p.strict {
				return nil, fmt.Errorf("failed to parse selectors '%s': %w", selectorText, err)
			}
			continue // Skip invalid selectors in non-strict mode
		}

		// Parse declarations
		declarations, err := p.parseDeclarations(declarationsText)
		if err != nil {
			if p.strict {
				return nil, fmt.Errorf("failed to parse declarations '%s': %w", declarationsText, err)
			}
			continue // Skip invalid declarations in non-strict mode
		}

		rule := &Rule{
			Selectors:    selectors,
			Declarations: declarations,
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// parseSelectors parses CSS selectors
func (p *Parser) parseSelectors(selectorText string) ([]*Selector, error) {
	var selectors []*Selector

	// Split by commas for multiple selectors
	selectorParts := strings.Split(selectorText, ",")

	for _, part := range selectorParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		selector, err := p.parseSelector(part)
		if err != nil {
			return nil, err
		}

		selectors = append(selectors, selector)
	}

	return selectors, nil
}

// parseSelector parses a single CSS selector
func (p *Parser) parseSelector(selectorText string) (*Selector, error) {
	selector := &Selector{
		Text:       selectorText,
		Specificity: p.calculateSpecificity(selectorText),
	}

	// Parse selector components
	components, err := p.parseSelectorComponents(selectorText)
	if err != nil {
		return nil, err
	}

	selector.Components = components
	return selector, nil
}

// parseSelectorComponents parses selector components
func (p *Parser) parseSelectorComponents(selectorText string) ([]*SelectorComponent, error) {
	var components []*SelectorComponent

	// Simple parsing - split by spaces for descendant selectors
	parts := strings.Fields(selectorText)

	for _, part := range parts {
		component := &SelectorComponent{}

		// Check for ID selector
		if strings.HasPrefix(part, "#") {
			component.Type = SelectorTypeID
			component.Value = part[1:]
		} else if strings.HasPrefix(part, ".") {
			// Class selector
			component.Type = SelectorTypeClass
			component.Value = part[1:]
		} else if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Attribute selector
			component.Type = SelectorTypeAttribute
			component.Value = part[1 : len(part)-1]
		} else if part == "*" {
			// Universal selector
			component.Type = SelectorTypeUniversal
			component.Value = "*"
		} else {
			// Element selector
			component.Type = SelectorTypeElement
			component.Value = part
		}

		components = append(components, component)
	}

	return components, nil
}

// parseDeclarations parses CSS declarations
func (p *Parser) parseDeclarations(declarationsText string) ([]*Declaration, error) {
	var declarations []*Declaration

	// Split by semicolons
	declarationParts := strings.Split(declarationsText, ";")

	for _, part := range declarationParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		declaration, err := p.parseDeclaration(part)
		if err != nil {
			if p.strict {
				return nil, err
			}
			continue // Skip invalid declarations in non-strict mode
		}

		declarations = append(declarations, declaration)
	}

	return declarations, nil
}

// parseDeclaration parses a single CSS declaration
func (p *Parser) parseDeclaration(declarationText string) (*Declaration, error) {
	// Split by colon
	parts := strings.SplitN(declarationText, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid declaration format: %s", declarationText)
	}

	property := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Check for !important
	important := false
	if strings.HasSuffix(value, "!important") {
		important = true
		value = strings.TrimSpace(strings.TrimSuffix(value, "!important"))
	}

	return &Declaration{
		Property:  property,
		Value:     value,
		Important: important,
	}, nil
}

// calculateSpecificity calculates CSS selector specificity
func (p *Parser) calculateSpecificity(selector string) int {
	// Simple specificity calculation
	// IDs: 100, Classes/Attributes: 10, Elements: 1
	specificity := 0

	// Count IDs
	idCount := strings.Count(selector, "#")
	specificity += idCount * 100

	// Count classes and attributes
	classCount := strings.Count(selector, ".")
	attrCount := strings.Count(selector, "[")
	specificity += (classCount + attrCount) * 10

	// Count elements (rough approximation)
	elementRegex := regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9]*\b`)
	elements := elementRegex.FindAllString(selector, -1)
	elementCount := 0
	for _, element := range elements {
		// Filter out CSS keywords
		if !p.isCSSKeyword(element) {
			elementCount++
		}
	}
	specificity += elementCount

	return specificity
}

// isCSSKeyword checks if a string is a CSS keyword
func (p *Parser) isCSSKeyword(word string) bool {
	keywords := map[string]bool{
		"important": true, "inherit": true, "initial": true, "unset": true,
		"auto": true, "none": true, "normal": true, "bold": true,
	}
	return keywords[strings.ToLower(word)]
}

// removeComments removes CSS comments
func (p *Parser) removeComments(content string) string {
	commentRegex := regexp.MustCompile(`/\*.*?\*/`)
	return commentRegex.ReplaceAllString(content, "")
}

// Stylesheet represents a CSS stylesheet
type Stylesheet struct {
	Rules []*Rule `json:"rules"`
}

// Rule represents a CSS rule
type Rule struct {
	Selectors    []*Selector    `json:"selectors"`
	Declarations []*Declaration `json:"declarations"`
}

// Selector represents a CSS selector
type Selector struct {
	Text        string               `json:"text"`
	Specificity int                  `json:"specificity"`
	Components  []*SelectorComponent `json:"components"`
}

// SelectorComponent represents a component of a CSS selector
type SelectorComponent struct {
	Type  SelectorType `json:"type"`
	Value string       `json:"value"`
}

// SelectorType represents the type of a CSS selector
type SelectorType string

const (
	SelectorTypeElement   SelectorType = "element"
	SelectorTypeClass     SelectorType = "class"
	SelectorTypeID        SelectorType = "id"
	SelectorTypeAttribute SelectorType = "attribute"
	SelectorTypeUniversal SelectorType = "universal"
)

// Declaration represents a CSS declaration
type Declaration struct {
	Property  string `json:"property"`
	Value     string `json:"value"`
	Important bool   `json:"important"`
}

// GetDeclaration returns the first declaration with the given property
func (r *Rule) GetDeclaration(property string) *Declaration {
	for _, decl := range r.Declarations {
		if strings.EqualFold(decl.Property, property) {
			return decl
		}
	}
	return nil
}

// HasDeclaration checks if the rule has a declaration with the given property
func (r *Rule) HasDeclaration(property string) bool {
	return r.GetDeclaration(property) != nil
}

// AddDeclaration adds a declaration to the rule
func (r *Rule) AddDeclaration(property, value string, important bool) {
	r.Declarations = append(r.Declarations, &Declaration{
		Property:  property,
		Value:     value,
		Important: important,
	})
}

// ParseValue parses a CSS value into appropriate type
func ParseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// Try to parse as number
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}

	// Try to parse as number with unit
	unitRegex := regexp.MustCompile(`^([\d.]+)(px|em|rem|%|pt|pc|in|cm|mm|ex|ch|vw|vh|vmin|vmax)$`)
	if matches := unitRegex.FindStringSubmatch(value); len(matches) == 3 {
		if num, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return map[string]interface{}{
				"value": num,
				"unit":  matches[2],
			}
		}
	}

	// Try to parse as color
	if color := parseColor(value); color != nil {
		return color
	}

	// Return as string
	return value
}

// parseColor parses CSS color values
func parseColor(value string) *domain.Color {
	value = strings.TrimSpace(strings.ToLower(value))

	// Named colors
	namedColors := map[string]domain.Color{
		"black":   {R: 0, G: 0, B: 0, A: 255},
		"white":   {R: 255, G: 255, B: 255, A: 255},
		"red":     {R: 255, G: 0, B: 0, A: 255},
		"green":   {R: 0, G: 128, B: 0, A: 255},
		"blue":    {R: 0, G: 0, B: 255, A: 255},
		"yellow":  {R: 255, G: 255, B: 0, A: 255},
		"cyan":    {R: 0, G: 255, B: 255, A: 255},
		"magenta": {R: 255, G: 0, B: 255, A: 255},
	}

	if color, exists := namedColors[value]; exists {
		return &color
	}

	// Hex colors
	if strings.HasPrefix(value, "#") {
		hex := value[1:]
		if len(hex) == 3 {
			// Short hex: #rgb -> #rrggbb
			hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
		}
		if len(hex) == 6 {
			if r, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
				if g, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
					if b, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
						return &domain.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
					}
				}
			}
		}
	}

	// RGB/RGBA functions
	rgbRegex := regexp.MustCompile(`rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([\d.]+))?\s*\)`)
	if matches := rgbRegex.FindStringSubmatch(value); len(matches) >= 4 {
		if r, err := strconv.ParseUint(matches[1], 10, 8); err == nil {
			if g, err := strconv.ParseUint(matches[2], 10, 8); err == nil {
				if b, err := strconv.ParseUint(matches[3], 10, 8); err == nil {
					a := uint8(255)
					if len(matches) > 4 && matches[4] != "" {
						if alpha, err := strconv.ParseFloat(matches[4], 64); err == nil {
							a = uint8(alpha * 255)
						}
					}
					return &domain.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: a}
				}
			}
		}
	}

	return nil
}
