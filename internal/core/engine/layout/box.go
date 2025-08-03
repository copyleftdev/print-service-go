package layout

import (
	"strconv"

	"print-service/internal/core/domain"
)

// BoxCalculator handles box model calculations
type BoxCalculator struct{}

// NewBoxCalculator creates a new box calculator
func NewBoxCalculator() *BoxCalculator {
	return &BoxCalculator{}
}

// Calculate calculates the box model for a layout node
func (bc *BoxCalculator) Calculate(node *domain.LayoutNode, ctx *LayoutContext) error {
	if node == nil {
		return nil
	}

	// Calculate content box
	contentBox := bc.calculateContentBox(node, ctx)

	// Calculate padding box
	paddingBox := bc.calculatePaddingBox(contentBox, node.Style.Padding)

	// Calculate border box
	borderBox := bc.calculateBorderBox(paddingBox, node.Style.Border)

	// Calculate margin box
	marginBox := bc.calculateMarginBox(borderBox, node.Style.Margin)

	// Set the final box
	node.Box = marginBox

	return nil
}

// calculateContentBox calculates the content box dimensions
func (bc *BoxCalculator) calculateContentBox(node *domain.LayoutNode, ctx *LayoutContext) domain.Box {
	box := domain.Box{}

	// Calculate width
	if node.Style.Width == "auto" {
		// Auto width - use available space
		if node.Parent != nil {
			box.Width = node.Parent.Box.Width
		} else {
			box.Width = ctx.Viewport.Width
		}
	} else {
		box.Width = bc.parseLength(node.Style.Width, ctx.Viewport.Width)
	}

	// Calculate height
	if node.Style.Height == "auto" {
		// Auto height - calculate based on content
		box.Height = bc.calculateAutoHeight(node, ctx)
	} else {
		box.Height = bc.parseLength(node.Style.Height, ctx.Viewport.Height)
	}

	return box
}

// calculatePaddingBox calculates the padding box
func (bc *BoxCalculator) calculatePaddingBox(contentBox domain.Box, padding domain.Margins) domain.Box {
	return domain.Box{
		X:      contentBox.X - padding.Left,
		Y:      contentBox.Y - padding.Top,
		Width:  contentBox.Width + padding.Left + padding.Right,
		Height: contentBox.Height + padding.Top + padding.Bottom,
	}
}

// calculateBorderBox calculates the border box
func (bc *BoxCalculator) calculateBorderBox(paddingBox domain.Box, border domain.BorderStyle) domain.Box {
	borderWidth := border.Width
	return domain.Box{
		X:      paddingBox.X - borderWidth,
		Y:      paddingBox.Y - borderWidth,
		Width:  paddingBox.Width + 2*borderWidth,
		Height: paddingBox.Height + 2*borderWidth,
	}
}

// calculateMarginBox calculates the margin box
func (bc *BoxCalculator) calculateMarginBox(borderBox domain.Box, margin domain.Margins) domain.Box {
	return domain.Box{
		X:      borderBox.X - margin.Left,
		Y:      borderBox.Y - margin.Top,
		Width:  borderBox.Width + margin.Left + margin.Right,
		Height: borderBox.Height + margin.Top + margin.Bottom,
	}
}

// calculateAutoHeight calculates automatic height based on content
func (bc *BoxCalculator) calculateAutoHeight(node *domain.LayoutNode, ctx *LayoutContext) float64 {
	if node.Content != "" {
		// Text content - calculate based on font metrics
		lineHeight := node.Style.Text.LineHeight * node.Style.Font.Size
		lines := bc.estimateLineCount(node.Content, node.Box.Width, node.Style.Font)
		return float64(lines) * lineHeight
	}

	// Block element - sum of children heights
	totalHeight := 0.0
	for _, child := range node.Children {
		totalHeight += child.Box.Height
	}

	return totalHeight
}

// estimateLineCount estimates the number of lines for text content
func (bc *BoxCalculator) estimateLineCount(text string, width float64, font domain.FontStyle) int {
	// Simple estimation - assume average character width
	avgCharWidth := font.Size * 0.6 // Rough approximation
	charsPerLine := int(width / avgCharWidth)

	if charsPerLine <= 0 {
		charsPerLine = 1
	}

	lines := (len(text) + charsPerLine - 1) / charsPerLine
	if lines == 0 {
		lines = 1
	}

	return lines
}

// parseLength parses a CSS length value
func (bc *BoxCalculator) parseLength(value string, containerSize float64) float64 {
	if value == "auto" {
		return 0
	}

	// Handle percentage
	if len(value) > 1 && value[len(value)-1] == '%' {
		if percent := parseFloat(value[:len(value)-1]); percent >= 0 {
			return containerSize * percent / 100
		}
	}

	// Handle pixels
	if len(value) > 2 && value[len(value)-2:] == "px" {
		if pixels := parseFloat(value[:len(value)-2]); pixels >= 0 {
			return pixels
		}
	}

	// Handle other units (simplified)
	if len(value) > 2 {
		unit := value[len(value)-2:]
		numValue := parseFloat(value[:len(value)-2])

		switch unit {
		case "em":
			return numValue * 16 // Assume 16px base font size
		case "pt":
			return numValue * 1.33 // 1pt = 1.33px approximately
		case "in":
			return numValue * 96 // 1in = 96px
		case "cm":
			return numValue * 37.8 // 1cm = 37.8px approximately
		case "mm":
			return numValue * 3.78 // 1mm = 3.78px approximately
		}
	}

	// Try to parse as plain number (assume pixels)
	if num := parseFloat(value); num >= 0 {
		return num
	}

	return 0
}

// parseFloat safely parses a float value
func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}
