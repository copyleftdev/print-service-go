package layout

import (
	"print-service/internal/core/domain"
)

// FlowEngine handles document flow calculations
type FlowEngine struct{}

// NewFlowEngine creates a new flow engine
func NewFlowEngine() *FlowEngine {
	return &FlowEngine{}
}

// Calculate calculates the document flow for a layout node
func (fe *FlowEngine) Calculate(node *domain.LayoutNode, ctx *LayoutContext) error {
	if node == nil {
		return nil
	}

	switch node.Style.Display {
	case domain.DisplayBlock:
		return fe.calculateBlockFlow(node, ctx)
	case domain.DisplayInline:
		return fe.calculateInlineFlow(node, ctx)
	case domain.DisplayInlineBlock:
		return fe.calculateInlineBlockFlow(node, ctx)
	case domain.DisplayFlex:
		return fe.calculateFlexFlow(node, ctx)
	default:
		return fe.calculateBlockFlow(node, ctx)
	}
}

// calculateBlockFlow calculates block-level element flow
func (fe *FlowEngine) calculateBlockFlow(node *domain.LayoutNode, ctx *LayoutContext) error {
	currentY := node.Box.Y + node.Style.Padding.Top

	for _, child := range node.Children {
		// Position child at current Y
		child.Box.X = node.Box.X + node.Style.Padding.Left + child.Style.Margin.Left
		child.Box.Y = currentY + child.Style.Margin.Top

		// Move Y position down by child's total height
		currentY += child.Box.Height + child.Style.Margin.Top + child.Style.Margin.Bottom
	}

	// Update parent height if needed
	totalContentHeight := currentY - (node.Box.Y + node.Style.Padding.Top)
	if node.Style.Height == "auto" {
		node.Box.Height = totalContentHeight + node.Style.Padding.Top + node.Style.Padding.Bottom
	}

	return nil
}

// calculateInlineFlow calculates inline element flow
func (fe *FlowEngine) calculateInlineFlow(node *domain.LayoutNode, ctx *LayoutContext) error {
	// Inline elements flow horizontally
	currentX := node.Box.X + node.Style.Padding.Left
	lineHeight := node.Style.Font.Size * node.Style.Text.LineHeight

	for _, child := range node.Children {
		// Check if child fits on current line
		availableWidth := node.Box.Width - (currentX - node.Box.X) - node.Style.Padding.Right

		if child.Box.Width <= availableWidth {
			// Child fits on current line
			child.Box.X = currentX
			child.Box.Y = node.Box.Y + node.Style.Padding.Top
			currentX += child.Box.Width
		} else {
			// Child doesn't fit, wrap to next line
			currentX = node.Box.X + node.Style.Padding.Left
			child.Box.X = currentX
			child.Box.Y = node.Box.Y + node.Style.Padding.Top + lineHeight
			currentX += child.Box.Width
		}
	}

	return nil
}

// calculateInlineBlockFlow calculates inline-block element flow
func (fe *FlowEngine) calculateInlineBlockFlow(node *domain.LayoutNode, ctx *LayoutContext) error {
	// Inline-block elements flow like inline but maintain block characteristics
	return fe.calculateInlineFlow(node, ctx)
}

// calculateFlexFlow calculates flexbox layout
func (fe *FlowEngine) calculateFlexFlow(node *domain.LayoutNode, ctx *LayoutContext) error {
	// Simple flex layout - distribute children evenly
	if len(node.Children) == 0 {
		return nil
	}

	availableWidth := node.Box.Width - node.Style.Padding.Left - node.Style.Padding.Right
	childWidth := availableWidth / float64(len(node.Children))

	currentX := node.Box.X + node.Style.Padding.Left

	for _, child := range node.Children {
		child.Box.X = currentX
		child.Box.Y = node.Box.Y + node.Style.Padding.Top
		child.Box.Width = childWidth
		currentX += childWidth
	}

	return nil
}

// CalculateAbsolutePosition calculates absolute positioning
func (fe *FlowEngine) CalculateAbsolutePosition(node *domain.LayoutNode, ctx *LayoutContext) error {
	if node.Style.Position != domain.PositionAbsolute {
		return nil
	}

	// For absolute positioning, position relative to containing block
	containingBlock := fe.findContainingBlock(node)
	if containingBlock == nil {
		// Use viewport as containing block
		node.Box.X = 0
		node.Box.Y = 0
	} else {
		node.Box.X = containingBlock.Box.X
		node.Box.Y = containingBlock.Box.Y
	}

	return nil
}

// findContainingBlock finds the containing block for positioned elements
func (fe *FlowEngine) findContainingBlock(node *domain.LayoutNode) *domain.LayoutNode {
	current := node.Parent
	for current != nil {
		if current.Style.Position == domain.PositionRelative ||
			current.Style.Position == domain.PositionAbsolute ||
			current.Style.Position == domain.PositionFixed {
			return current
		}
		current = current.Parent
	}
	return nil
}

// CalculateFloatPosition calculates float positioning
func (fe *FlowEngine) CalculateFloatPosition(node *domain.LayoutNode, ctx *LayoutContext) error {
	// Simplified float implementation
	// In a real implementation, this would be much more complex
	return nil
}

// CalculateStackingContext calculates z-index stacking
func (fe *FlowEngine) CalculateStackingContext(node *domain.LayoutNode, ctx *LayoutContext) error {
	// Sort children by z-index
	if len(node.Children) <= 1 {
		return nil
	}

	// Simple bubble sort by z-index
	for i := 0; i < len(node.Children)-1; i++ {
		for j := 0; j < len(node.Children)-i-1; j++ {
			if node.Children[j].Style.ZIndex > node.Children[j+1].Style.ZIndex {
				node.Children[j], node.Children[j+1] = node.Children[j+1], node.Children[j]
			}
		}
	}

	return nil
}
