package layout

import (
	"print-service/internal/core/domain"
)

// PageBreaker handles page breaking logic for print layouts
type PageBreaker struct{}

// NewPageBreaker creates a new page breaker
func NewPageBreaker() *PageBreaker {
	return &PageBreaker{}
}

// CalculatePageBreaks calculates where page breaks should occur
func (pb *PageBreaker) CalculatePageBreaks(node *domain.LayoutNode, pageHeight float64) ([]*PageBreak, error) {
	var pageBreaks []*PageBreak
	currentPage := &PageBreak{
		PageNumber: 1,
		StartY:     0,
		EndY:       pageHeight,
		Nodes:      make([]*domain.LayoutNode, 0),
	}

	err := pb.processNode(node, currentPage, pageHeight, &pageBreaks)
	if err != nil {
		return nil, err
	}

	// Add the last page if it has content
	if len(currentPage.Nodes) > 0 {
		pageBreaks = append(pageBreaks, currentPage)
	}

	return pageBreaks, nil
}

// processNode processes a node for page breaking
func (pb *PageBreaker) processNode(node *domain.LayoutNode, currentPage *PageBreak, pageHeight float64, pageBreaks *[]*PageBreak) error {
	if node == nil {
		return nil
	}

	// Check if node fits on current page
	nodeBottom := node.Box.Y + node.Box.Height

	if nodeBottom > currentPage.EndY {
		// Node doesn't fit, need to break
		if pb.canBreakBefore(node) {
			// Start a new page
			*pageBreaks = append(*pageBreaks, currentPage)
			currentPage = &PageBreak{
				PageNumber: len(*pageBreaks) + 1,
				StartY:     currentPage.EndY,
				EndY:       currentPage.EndY + pageHeight,
				Nodes:      make([]*domain.LayoutNode, 0),
			}

			// Adjust node position for new page
			node.Box.Y = currentPage.StartY + (node.Box.Y - currentPage.StartY)
		} else {
			// Try to break within the node
			if err := pb.breakWithinNode(node, currentPage, pageHeight, pageBreaks); err != nil {
				return err
			}
		}
	}

	// Add node to current page
	currentPage.Nodes = append(currentPage.Nodes, node)

	// Process children
	for _, child := range node.Children {
		if err := pb.processNode(child, currentPage, pageHeight, pageBreaks); err != nil {
			return err
		}
	}

	return nil
}

// canBreakBefore checks if a page break can occur before this node
func (pb *PageBreaker) canBreakBefore(node *domain.LayoutNode) bool {
	// Don't break before inline elements
	if node.Style.Display == domain.DisplayInline {
		return false
	}

	// Don't break in the middle of a line of text
	if node.Content != "" && node.Parent != nil {
		return false
	}

	// Allow breaks before block elements
	return node.Style.Display == domain.DisplayBlock
}

// breakWithinNode attempts to break content within a node
func (pb *PageBreaker) breakWithinNode(node *domain.LayoutNode, currentPage *PageBreak, pageHeight float64, pageBreaks *[]*PageBreak) error {
	if node.Content != "" {
		// Break text content
		return pb.breakTextNode(node, currentPage, pageHeight, pageBreaks)
	}

	// For other nodes, try to break between children
	return pb.breakBetweenChildren(node, currentPage, pageHeight, pageBreaks)
}

// breakTextNode breaks a text node across pages
func (pb *PageBreaker) breakTextNode(node *domain.LayoutNode, currentPage *PageBreak, pageHeight float64, pageBreaks *[]*PageBreak) error {
	// Calculate how much text fits on current page
	availableHeight := currentPage.EndY - node.Box.Y
	lineHeight := node.Style.Font.Size * node.Style.Text.LineHeight

	if lineHeight <= 0 {
		lineHeight = node.Style.Font.Size * 1.2
	}

	linesOnCurrentPage := int(availableHeight / lineHeight)
	if linesOnCurrentPage < 1 {
		linesOnCurrentPage = 1
	}

	// Split text content
	textEngine := NewTextEngine()
	lines := textEngine.SplitTextIntoLines(node.Content, node.Style.Font, node.Box.Width)

	if len(lines) <= linesOnCurrentPage {
		// All text fits on current page
		return nil
	}

	// Split the node
	firstPartLines := lines[:linesOnCurrentPage]
	remainingLines := lines[linesOnCurrentPage:]

	// Update current node with first part
	node.Content = joinLines(firstPartLines)
	node.Box.Height = float64(len(firstPartLines)) * lineHeight

	// Create new node for remaining content
	remainingNode := &domain.LayoutNode{
		ID:      node.ID + "_continued",
		Type:    node.Type,
		Content: joinLines(remainingLines),
		Style:   node.Style,
		Box: domain.Box{
			X:      node.Box.X,
			Y:      currentPage.EndY,
			Width:  node.Box.Width,
			Height: float64(len(remainingLines)) * lineHeight,
		},
		Parent: node.Parent,
	}

	// Add remaining node to parent
	if node.Parent != nil {
		node.Parent.Children = append(node.Parent.Children, remainingNode)
	}

	return nil
}

// breakBetweenChildren breaks between child nodes
func (pb *PageBreaker) breakBetweenChildren(node *domain.LayoutNode, currentPage *PageBreak, pageHeight float64, pageBreaks *[]*PageBreak) error {
	// Find the best break point between children
	for i, child := range node.Children {
		childBottom := child.Box.Y + child.Box.Height
		if childBottom > currentPage.EndY {
			// Break before this child
			if i > 0 {
				// Move this child and subsequent children to next page
				*pageBreaks = append(*pageBreaks, currentPage)
				newPage := &PageBreak{
					PageNumber: len(*pageBreaks) + 1,
					StartY:     currentPage.EndY,
					EndY:       currentPage.EndY + pageHeight,
					Nodes:      make([]*domain.LayoutNode, 0),
				}

				// Adjust positions of moved children
				yOffset := newPage.StartY - child.Box.Y
				for j := i; j < len(node.Children); j++ {
					pb.adjustNodePosition(node.Children[j], 0, yOffset)
				}
			}
			break
		}
	}

	return nil
}

// adjustNodePosition recursively adjusts node positions
func (pb *PageBreaker) adjustNodePosition(node *domain.LayoutNode, xOffset, yOffset float64) {
	node.Box.X += xOffset
	node.Box.Y += yOffset

	for _, child := range node.Children {
		pb.adjustNodePosition(child, xOffset, yOffset)
	}
}

// joinLines joins text lines back into a single string
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += " "
		}
		result += line
	}
	return result
}

// PageBreak represents a page break in the document
type PageBreak struct {
	PageNumber int                  `json:"page_number"`
	StartY     float64              `json:"start_y"`
	EndY       float64              `json:"end_y"`
	Nodes      []*domain.LayoutNode `json:"nodes"`
}

// GetPageCount returns the total number of pages
func (pb *PageBreaker) GetPageCount(pageBreaks []*PageBreak) int {
	if len(pageBreaks) == 0 {
		return 1
	}
	return pageBreaks[len(pageBreaks)-1].PageNumber
}

// GetNodesForPage returns all nodes that appear on a specific page
func (pb *PageBreaker) GetNodesForPage(pageBreaks []*PageBreak, pageNumber int) []*domain.LayoutNode {
	for _, pageBreak := range pageBreaks {
		if pageBreak.PageNumber == pageNumber {
			return pageBreak.Nodes
		}
	}
	return nil
}

// CalculatePageMargins calculates margins for a page
func (pb *PageBreaker) CalculatePageMargins(pageOptions domain.PageOptions) domain.Margins {
	return pageOptions.Margins
}

// ShouldBreakBefore checks if a page break should occur before an element
func (pb *PageBreaker) ShouldBreakBefore(node *domain.LayoutNode) bool {
	// Check for CSS page-break-before property (simplified)
	// In a real implementation, this would check computed styles
	return false
}

// ShouldBreakAfter checks if a page break should occur after an element
func (pb *PageBreaker) ShouldBreakAfter(node *domain.LayoutNode) bool {
	// Check for CSS page-break-after property (simplified)
	// In a real implementation, this would check computed styles
	return false
}

// AvoidBreakInside checks if breaks should be avoided inside an element
func (pb *PageBreaker) AvoidBreakInside(node *domain.LayoutNode) bool {
	// Check for CSS page-break-inside property (simplified)
	// In a real implementation, this would check computed styles
	return false
}
