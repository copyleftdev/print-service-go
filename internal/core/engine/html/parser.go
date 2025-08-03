package html

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
	"print-service/internal/core/domain"
)

// Parser handles HTML parsing and DOM tree construction
type Parser struct {
	sanitizer *Sanitizer
	validator *Validator
}

// NewParser creates a new HTML parser
func NewParser(sanitizer *Sanitizer, validator *Validator) *Parser {
	return &Parser{
		sanitizer: sanitizer,
		validator: validator,
	}
}

// Parse parses HTML content and returns a DOM tree
func (p *Parser) Parse(content string, options domain.SecurityOptions) (*DOMNode, error) {
	// Sanitize HTML if required
	if options.SanitizeHTML {
		sanitized, err := p.sanitizer.Sanitize(content, options)
		if err != nil {
			return nil, fmt.Errorf("failed to sanitize HTML: %w", err)
		}
		content = sanitized
	}

	// Validate HTML
	if err := p.validator.Validate(content); err != nil {
		return nil, fmt.Errorf("HTML validation failed: %w", err)
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Convert to our DOM structure
	domNode := p.convertNode(doc)
	return domNode, nil
}

// ParseFragment parses an HTML fragment
func (p *Parser) ParseFragment(content string, context *html.Node) ([]*DOMNode, error) {
	nodes, err := html.ParseFragment(strings.NewReader(content), context)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML fragment: %w", err)
	}

	var domNodes []*DOMNode
	for _, node := range nodes {
		domNode := p.convertNode(node)
		domNodes = append(domNodes, domNode)
	}

	return domNodes, nil
}

// convertNode converts an html.Node to our DOMNode structure
func (p *Parser) convertNode(node *html.Node) *DOMNode {
	domNode := &DOMNode{
		Type:       NodeType(node.Type),
		Data:       node.Data,
		Namespace:  node.Namespace,
		Attributes: make(map[string]string),
	}

	// Convert attributes
	for _, attr := range node.Attr {
		domNode.Attributes[attr.Key] = attr.Val
	}

	// Convert children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childNode := p.convertNode(child)
		childNode.Parent = domNode
		domNode.Children = append(domNode.Children, childNode)
	}

	return domNode
}

// DOMNode represents a node in the DOM tree
type DOMNode struct {
	Type       NodeType          `json:"type"`
	Data       string            `json:"data"`
	Namespace  string            `json:"namespace"`
	Attributes map[string]string `json:"attributes"`
	Children   []*DOMNode        `json:"children"`
	Parent     *DOMNode          `json:"-"`
}

// NodeType represents the type of a DOM node
type NodeType int

const (
	ErrorNode NodeType = iota
	TextNode
	DocumentNode
	ElementNode
	CommentNode
	DoctypeNode
)

// GetAttribute returns the value of an attribute
func (n *DOMNode) GetAttribute(key string) (string, bool) {
	value, exists := n.Attributes[key]
	return value, exists
}

// SetAttribute sets the value of an attribute
func (n *DOMNode) SetAttribute(key, value string) {
	if n.Attributes == nil {
		n.Attributes = make(map[string]string)
	}
	n.Attributes[key] = value
}

// HasAttribute checks if an attribute exists
func (n *DOMNode) HasAttribute(key string) bool {
	_, exists := n.Attributes[key]
	return exists
}

// GetElementsByTagName returns all descendant elements with the given tag name
func (n *DOMNode) GetElementsByTagName(tagName string) []*DOMNode {
	var elements []*DOMNode
	n.walkElements(func(node *DOMNode) bool {
		if node.Type == ElementNode && strings.EqualFold(node.Data, tagName) {
			elements = append(elements, node)
		}
		return true
	})
	return elements
}

// GetElementsByClassName returns all descendant elements with the given class name
func (n *DOMNode) GetElementsByClassName(className string) []*DOMNode {
	var elements []*DOMNode
	n.walkElements(func(node *DOMNode) bool {
		if node.Type == ElementNode {
			if class, exists := node.GetAttribute("class"); exists {
				classes := strings.Fields(class)
				for _, c := range classes {
					if c == className {
						elements = append(elements, node)
						break
					}
				}
			}
		}
		return true
	})
	return elements
}

// GetElementByID returns the first descendant element with the given ID
func (n *DOMNode) GetElementByID(id string) *DOMNode {
	var result *DOMNode
	n.walkElements(func(node *DOMNode) bool {
		if node.Type == ElementNode {
			if nodeID, exists := node.GetAttribute("id"); exists && nodeID == id {
				result = node
				return false // Stop walking
			}
		}
		return true
	})
	return result
}

// walkElements walks through all descendant elements
func (n *DOMNode) walkElements(fn func(*DOMNode) bool) {
	if !fn(n) {
		return
	}
	for _, child := range n.Children {
		child.walkElements(fn)
	}
}

// String returns a string representation of the node
func (n *DOMNode) String() string {
	switch n.Type {
	case TextNode:
		return n.Data
	case ElementNode:
		return fmt.Sprintf("<%s>", n.Data)
	case CommentNode:
		return fmt.Sprintf("<!-- %s -->", n.Data)
	case DoctypeNode:
		return fmt.Sprintf("<!DOCTYPE %s>", n.Data)
	default:
		return fmt.Sprintf("Node(%d: %s)", n.Type, n.Data)
	}
}

// Render renders the DOM node back to HTML
func (n *DOMNode) Render(w io.Writer) error {
	switch n.Type {
	case TextNode:
		_, err := w.Write([]byte(html.EscapeString(n.Data)))
		return err
	case ElementNode:
		// Opening tag
		fmt.Fprintf(w, "<%s", n.Data)
		for key, value := range n.Attributes {
			fmt.Fprintf(w, ` %s="%s"`, key, html.EscapeString(value))
		}
		
		// Self-closing tags
		if n.isSelfClosing() {
			_, err := w.Write([]byte(" />"))
			return err
		}
		
		_, err := w.Write([]byte(">"))
		if err != nil {
			return err
		}
		
		// Children
		for _, child := range n.Children {
			if err := child.Render(w); err != nil {
				return err
			}
		}
		
		// Closing tag
		fmt.Fprintf(w, "</%s>", n.Data)
		return nil
	case CommentNode:
		fmt.Fprintf(w, "<!-- %s -->", n.Data)
		return nil
	case DoctypeNode:
		fmt.Fprintf(w, "<!DOCTYPE %s>", n.Data)
		return nil
	default:
		return nil
	}
}

// isSelfClosing checks if the element is self-closing
func (n *DOMNode) isSelfClosing() bool {
	selfClosingTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}
	return selfClosingTags[strings.ToLower(n.Data)]
}
