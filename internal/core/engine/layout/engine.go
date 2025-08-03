package layout

import (
	"fmt"
	"strconv"
	"strings"

	"print-service/internal/core/domain"
	"print-service/internal/core/engine/css"
	"print-service/internal/core/engine/html"
)

// Engine handles layout calculations for documents
type Engine struct {
	boxCalculator *BoxCalculator
	textEngine    *TextEngine
	flowEngine    *FlowEngine
	pageBreaker   *PageBreaker
}

// NewEngine creates a new layout engine
func NewEngine() *Engine {
	return &Engine{
		boxCalculator: NewBoxCalculator(),
		textEngine:    NewTextEngine(),
		flowEngine:    NewFlowEngine(),
		pageBreaker:   NewPageBreaker(),
	}
}

// CalculateLayout calculates the layout for a document
func (e *Engine) CalculateLayout(domTree *html.DOMNode, stylesheet *css.Stylesheet, options domain.LayoutOptions) (*domain.LayoutNode, error) {
	if domTree == nil {
		return nil, fmt.Errorf("DOM tree is nil")
	}

	// Create layout context
	ctx := &LayoutContext{
		Viewport: domain.Box{
			Width:  float64(options.ViewportWidth),
			Height: float64(options.ViewportHeight),
		},
		DPI:     float64(options.DPI),
		Options: options,
	}

	// Build layout tree from DOM
	layoutTree, err := e.buildLayoutTree(domTree, stylesheet, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build layout tree: %w", err)
	}

	// Calculate layout
	if err := e.calculateLayout(layoutTree, ctx); err != nil {
		return nil, fmt.Errorf("failed to calculate layout: %w", err)
	}

	return layoutTree, nil
}

// buildLayoutTree builds a layout tree from DOM and CSS
func (e *Engine) buildLayoutTree(domNode *html.DOMNode, stylesheet *css.Stylesheet, ctx *LayoutContext) (*domain.LayoutNode, error) {
	if domNode == nil {
		return nil, nil
	}

	// Create layout node
	layoutNode := &domain.LayoutNode{
		ID:   fmt.Sprintf("node_%p", domNode),
		Type: getNodeTypeName(domNode.Type),
	}

	// Calculate computed styles
	computedStyle, err := e.computeStyle(domNode, stylesheet, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compute style: %w", err)
	}
	layoutNode.Style = *computedStyle

	// Skip nodes with display: none
	if computedStyle.Display == domain.DisplayNone {
		return nil, nil
	}

	// Set content for text nodes
	if domNode.Type == html.TextNode {
		layoutNode.Content = domNode.Data
	}

	// Process children
	for _, child := range domNode.Children {
		childLayout, err := e.buildLayoutTree(child, stylesheet, ctx)
		if err != nil {
			return nil, err
		}
		if childLayout != nil {
			childLayout.Parent = layoutNode
			layoutNode.Children = append(layoutNode.Children, childLayout)
		}
	}

	return layoutNode, nil
}

// computeStyle computes the final styles for a DOM node
func (e *Engine) computeStyle(domNode *html.DOMNode, stylesheet *css.Stylesheet, ctx *LayoutContext) (*domain.ComputedStyle, error) {
	// Start with default styles
	style := getDefaultComputedStyle()

	// Apply matching CSS rules
	for _, rule := range stylesheet.Rules {
		if e.selectorMatches(rule.Selectors, domNode) {
			e.applyDeclarations(rule.Declarations, style)
		}
	}

	// Apply inline styles
	if inlineStyle, exists := domNode.GetAttribute("style"); exists {
		if err := e.applyInlineStyle(inlineStyle, style); err != nil {
			return nil, fmt.Errorf("failed to apply inline style: %w", err)
		}
	}

	return style, nil
}

// selectorMatches checks if any selector matches the DOM node
func (e *Engine) selectorMatches(selectors []*css.Selector, domNode *html.DOMNode) bool {
	for _, selector := range selectors {
		if e.singleSelectorMatches(selector, domNode) {
			return true
		}
	}
	return false
}

// singleSelectorMatches checks if a single selector matches the DOM node
func (e *Engine) singleSelectorMatches(selector *css.Selector, domNode *html.DOMNode) bool {
	// Simple matching - check the last component
	if len(selector.Components) == 0 {
		return false
	}

	lastComponent := selector.Components[len(selector.Components)-1]

	switch lastComponent.Type {
	case css.SelectorTypeElement:
		return domNode.Type == html.ElementNode && domNode.Data == lastComponent.Value
	case css.SelectorTypeClass:
		if class, exists := domNode.GetAttribute("class"); exists {
			classes := splitClasses(class)
			for _, c := range classes {
				if c == lastComponent.Value {
					return true
				}
			}
		}
	case css.SelectorTypeID:
		if id, exists := domNode.GetAttribute("id"); exists {
			return id == lastComponent.Value
		}
	case css.SelectorTypeUniversal:
		return true
	}

	return false
}

// applyDeclarations applies CSS declarations to computed style
func (e *Engine) applyDeclarations(declarations []*css.Declaration, style *domain.ComputedStyle) {
	for _, decl := range declarations {
		e.applyDeclaration(decl, style)
	}
}

// applyDeclaration applies a single CSS declaration
func (e *Engine) applyDeclaration(decl *css.Declaration, style *domain.ComputedStyle) {
	switch decl.Property {
	case "display":
		style.Display = domain.Display(decl.Value)
	case "position":
		style.Position = domain.Position(decl.Value)
	case "width":
		style.Width = decl.Value
	case "height":
		style.Height = decl.Value
	case "color":
		if color := css.ParseValue(decl.Value); color != nil {
			if c, ok := color.(*domain.Color); ok {
				style.Color = *c
			}
		}
	case "font-family":
		style.Font.Family = decl.Value
	case "font-size":
		if size := parseSize(decl.Value); size > 0 {
			style.Font.Size = size
		}
	case "font-weight":
		if weight := parseFontWeight(decl.Value); weight > 0 {
			style.Font.Weight = weight
		}
	case "text-align":
		style.Text.Align = domain.TextAlign(decl.Value)
	case "line-height":
		if height := parseSize(decl.Value); height > 0 {
			style.Text.LineHeight = height
		}
	}
}

// applyInlineStyle applies inline CSS styles
func (e *Engine) applyInlineStyle(inlineStyle string, style *domain.ComputedStyle) error {
	parser := css.NewParser(false)

	// Parse as a single rule
	ruleContent := fmt.Sprintf("dummy { %s }", inlineStyle)
	stylesheet, err := parser.Parse(ruleContent)
	if err != nil {
		return err
	}

	if len(stylesheet.Rules) > 0 {
		e.applyDeclarations(stylesheet.Rules[0].Declarations, style)
	}

	return nil
}

// calculateLayout calculates the actual layout positions and sizes
func (e *Engine) calculateLayout(layoutNode *domain.LayoutNode, ctx *LayoutContext) error {
	if layoutNode == nil {
		return nil
	}

	// Calculate box model
	if err := e.boxCalculator.Calculate(layoutNode, ctx); err != nil {
		return fmt.Errorf("box calculation failed: %w", err)
	}

	// Handle text layout
	if layoutNode.Content != "" {
		if err := e.textEngine.Layout(layoutNode, ctx); err != nil {
			return fmt.Errorf("text layout failed: %w", err)
		}
	}

	// Calculate children layout
	for _, child := range layoutNode.Children {
		if err := e.calculateLayout(child, ctx); err != nil {
			return err
		}
	}

	// Handle document flow
	if err := e.flowEngine.Calculate(layoutNode, ctx); err != nil {
		return fmt.Errorf("flow calculation failed: %w", err)
	}

	return nil
}

// LayoutContext provides context for layout calculations
type LayoutContext struct {
	Viewport domain.Box
	DPI      float64
	Options  domain.LayoutOptions
}

// Helper functions

func getDefaultComputedStyle() *domain.ComputedStyle {
	return &domain.ComputedStyle{
		Display:  domain.DisplayBlock,
		Position: domain.PositionStatic,
		Width:    "auto",
		Height:   "auto",
		Font: domain.FontStyle{
			Family: "serif",
			Size:   16,
			Weight: 400,
			Style:  "normal",
		},
		Text: domain.TextStyle{
			Align:      domain.TextAlignLeft,
			LineHeight: 1.2,
		},
		Color: domain.Color{R: 0, G: 0, B: 0, A: 255},
	}
}

func splitClasses(class string) []string {
	var classes []string
	for _, c := range strings.Fields(class) {
		if c != "" {
			classes = append(classes, c)
		}
	}
	return classes
}

func parseSize(value string) float64 {
	// Simple size parsing - just handle px for now
	if strings.HasSuffix(value, "px") {
		if size, err := strconv.ParseFloat(value[:len(value)-2], 64); err == nil {
			return size
		}
	}
	if size, err := strconv.ParseFloat(value, 64); err == nil {
		return size
	}
	return 0
}

func parseFontWeight(value string) int {
	switch value {
	case "normal":
		return 400
	case "bold":
		return 700
	case "lighter":
		return 300
	case "bolder":
		return 600
	default:
		if weight, err := strconv.Atoi(value); err == nil {
			return weight
		}
		return 400
	}
}

// getNodeTypeName converts NodeType to string
func getNodeTypeName(nodeType html.NodeType) string {
	switch nodeType {
	case html.ErrorNode:
		return "error"
	case html.TextNode:
		return "text"
	case html.DocumentNode:
		return "document"
	case html.ElementNode:
		return "element"
	case html.CommentNode:
		return "comment"
	case html.DoctypeNode:
		return "doctype"
	default:
		return "unknown"
	}
}
