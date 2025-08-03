package layout

import (
	"strings"

	"print-service/internal/core/domain"
)

// TextEngine handles text layout calculations
type TextEngine struct{}

// NewTextEngine creates a new text engine
func NewTextEngine() *TextEngine {
	return &TextEngine{}
}

// Layout calculates text layout for a node
func (te *TextEngine) Layout(node *domain.LayoutNode, ctx *LayoutContext) error {
	if node.Content == "" {
		return nil
	}

	// Calculate text metrics
	metrics := te.calculateTextMetrics(node.Content, node.Style.Font, node.Box.Width)
	
	// Update node box height based on text
	node.Box.Height = metrics.TotalHeight
	
	return nil
}

// TextMetrics represents calculated text metrics
type TextMetrics struct {
	LineCount   int     `json:"line_count"`
	LineHeight  float64 `json:"line_height"`
	TotalHeight float64 `json:"total_height"`
	MaxWidth    float64 `json:"max_width"`
}

// calculateTextMetrics calculates metrics for text content
func (te *TextEngine) calculateTextMetrics(text string, font domain.FontStyle, containerWidth float64) *TextMetrics {
	// Calculate line height
	lineHeight := font.Size * 1.2 // Default line height
	if font.Size > 0 {
		lineHeight = font.Size * 1.2
	}

	// Estimate character width based on font
	avgCharWidth := te.estimateCharWidth(font)
	
	// Calculate how many characters fit per line
	charsPerLine := int(containerWidth / avgCharWidth)
	if charsPerLine <= 0 {
		charsPerLine = 1
	}

	// Split text into words and calculate line breaks
	words := strings.Fields(text)
	lines := te.wrapWords(words, charsPerLine, avgCharWidth, containerWidth)
	
	lineCount := len(lines)
	if lineCount == 0 {
		lineCount = 1
	}

	// Calculate total height
	totalHeight := float64(lineCount) * lineHeight

	// Calculate maximum line width
	maxWidth := 0.0
	for _, line := range lines {
		lineWidth := float64(len(line)) * avgCharWidth
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return &TextMetrics{
		LineCount:   lineCount,
		LineHeight:  lineHeight,
		TotalHeight: totalHeight,
		MaxWidth:    maxWidth,
	}
}

// estimateCharWidth estimates the average character width for a font
func (te *TextEngine) estimateCharWidth(font domain.FontStyle) float64 {
	// Base character width estimation
	baseWidth := font.Size * 0.6

	// Adjust for font weight
	weightMultiplier := 1.0
	if font.Weight >= 700 {
		weightMultiplier = 1.1 // Bold fonts are slightly wider
	} else if font.Weight <= 300 {
		weightMultiplier = 0.9 // Light fonts are slightly narrower
	}

	// Adjust for font family (rough approximation)
	familyMultiplier := 1.0
	lowerFamily := strings.ToLower(font.Family)
	if strings.Contains(lowerFamily, "monospace") || strings.Contains(lowerFamily, "courier") {
		familyMultiplier = 1.2 // Monospace fonts are wider
	} else if strings.Contains(lowerFamily, "condensed") {
		familyMultiplier = 0.8 // Condensed fonts are narrower
	}

	return baseWidth * weightMultiplier * familyMultiplier
}

// wrapWords wraps words to fit within the specified constraints
func (te *TextEngine) wrapWords(words []string, charsPerLine int, avgCharWidth, containerWidth float64) []string {
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		// Check if adding this word would exceed the line width
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		// Estimate line width
		lineWidth := float64(len(testLine)) * avgCharWidth

		if lineWidth <= containerWidth || currentLine == "" {
			// Word fits on current line or it's the first word
			currentLine = testLine
		} else {
			// Word doesn't fit, start a new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// CalculateTextPosition calculates the position of text within a box
func (te *TextEngine) CalculateTextPosition(node *domain.LayoutNode) (x, y float64) {
	// Start position is top-left of content box
	x = node.Box.X + node.Style.Padding.Left
	y = node.Box.Y + node.Style.Padding.Top

	// Adjust for text alignment
	switch node.Style.Text.Align {
	case domain.TextAlignCenter:
		// Center horizontally
		contentWidth := node.Box.Width - node.Style.Padding.Left - node.Style.Padding.Right
		textWidth := te.estimateTextWidth(node.Content, node.Style.Font)
		x += (contentWidth - textWidth) / 2
	case domain.TextAlignRight:
		// Align right
		contentWidth := node.Box.Width - node.Style.Padding.Left - node.Style.Padding.Right
		textWidth := te.estimateTextWidth(node.Content, node.Style.Font)
		x += contentWidth - textWidth
	}

	// Add baseline offset for text
	y += node.Style.Font.Size

	return x, y
}

// estimateTextWidth estimates the total width of text
func (te *TextEngine) estimateTextWidth(text string, font domain.FontStyle) float64 {
	avgCharWidth := te.estimateCharWidth(font)
	return float64(len(text)) * avgCharWidth
}

// SplitTextIntoLines splits text into lines that fit within the given width
func (te *TextEngine) SplitTextIntoLines(text string, font domain.FontStyle, maxWidth float64) []string {
	if text == "" {
		return []string{""}
	}

	avgCharWidth := te.estimateCharWidth(font)
	charsPerLine := int(maxWidth / avgCharWidth)
	if charsPerLine <= 0 {
		charsPerLine = 1
	}

	words := strings.Fields(text)
	return te.wrapWords(words, charsPerLine, avgCharWidth, maxWidth)
}

// CalculateLineHeight calculates the line height for text
func (te *TextEngine) CalculateLineHeight(font domain.FontStyle, lineHeightStyle float64) float64 {
	if lineHeightStyle > 0 {
		if lineHeightStyle < 3 {
			// Relative value (e.g., 1.5)
			return font.Size * lineHeightStyle
		} else {
			// Absolute value in pixels
			return lineHeightStyle
		}
	}
	
	// Default line height
	return font.Size * 1.2
}
