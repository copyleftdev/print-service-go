package render

import (
	"fmt"
	"strings"

	"print-service/internal/core/domain"

	"github.com/jung-kurt/gofpdf"
)

// PDFRenderer handles PDF generation with advanced rendering capabilities
type PDFRenderer struct {
	fontManager *FontManager
	imageCache  *ImageCache
	options     PDFRenderOptions
}

// PDFRenderOptions configures PDF rendering behavior and output quality
type PDFRenderOptions struct {
	Compression    bool         // Enable PDF compression
	EmbedFonts     bool         // Embed fonts in PDF
	OptimizeImages bool         // Optimize embedded images
	ColorProfile   ColorProfile // Color profile for output
	OutputIntent   OutputIntent // PDF output intent
	PDFVersion     string       // PDF version (e.g., "1.4", "1.7")
}

// ColorProfile represents color profiles for PDF output
type ColorProfile string

const (
	ColorProfileRGB  ColorProfile = "RGB"  // RGB color space
	ColorProfileCMYK ColorProfile = "CMYK" // CMYK color space for print
	ColorProfileGray ColorProfile = "Gray" // Grayscale color space
)

// OutputIntent represents PDF output intent for color management
type OutputIntent struct {
	Type         string // Output intent type
	Identifier   string // Unique identifier
	Condition    string // Human-readable condition
	Info         string // Additional information
	RegistryName string // Registry name
}

// RenderContext provides rendering context for PDF generation
type RenderContext struct {
	PDF         *gofpdf.Fpdf // PDF document instance
	CurrentPage int          // Current page number
	PageWidth   float64      // Page width in mm
	PageHeight  float64      // Page height in mm
	DPI         float64      // Dots per inch
	Scale       float64      // Scaling factor
}

// NewPDFRenderer creates a new PDF renderer with specified options
func NewPDFRenderer(opts PDFRenderOptions) *PDFRenderer {
	return &PDFRenderer{
		fontManager: NewFontManager(),
		imageCache:  NewImageCache(),
		options:     opts,
	}
}

// Render renders a layout tree to PDF format with high-quality output
func (r *PDFRenderer) Render(layout *domain.LayoutNode, options domain.PrintOptions) ([]byte, error) {
	// Initialize PDF document with specified orientation and page size
	pdf := gofpdf.New(
		string(options.Page.Orientation), // Portrait or Landscape
		"mm",                             // Unit of measurement
		string(options.Page.Size.Name),   // Page size (A4, Letter, etc.)
		"",                               // Font directory (empty for built-in)
	)

	// Configure PDF metadata for document properties
	pdf.SetTitle("Generated Document", false)
	pdf.SetAuthor("Print Service", false)
	pdf.SetCreator("Pure Go Print Service", false)

	// Create rendering context with document parameters
	ctx := RenderContext{
		PDF:         pdf,                         // PDF document instance
		CurrentPage: 1,                           // Start with page 1
		PageWidth:   options.Page.Size.Width,     // Page width in mm
		PageHeight:  options.Page.Size.Height,    // Page height in mm
		DPI:         float64(options.Layout.DPI), // Resolution
		Scale:       options.Page.Scale,          // Scaling factor
	}

	// Add the first page to the document
	pdf.AddPage()

	// Render the complete layout tree recursively
	if err := r.renderLayoutNode(layout, ctx); err != nil {
		return nil, fmt.Errorf("failed to render layout: %w", err)
	}

	// Generate final PDF as byte array
	var buf strings.Builder
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return []byte(buf.String()), nil
}

// renderLayoutNode renders a layout node and its children recursively
func (r *PDFRenderer) renderLayoutNode(node *domain.LayoutNode, ctx RenderContext) error {
	if node == nil {
		return nil
	}

	// Render content based on node type
	switch node.Type {
	case "text":
		// Render text content with styling
		if err := r.RenderText(node.Content, node.Style, ctx); err != nil {
			return fmt.Errorf("failed to render text: %w", err)
		}
	case "element":
		// Render element with background and borders
		if err := r.RenderElement(node, ctx); err != nil {
			return fmt.Errorf("failed to render element: %w", err)
		}
	}

	// Recursively render all child nodes
	for _, child := range node.Children {
		if err := r.renderLayoutNode(child, ctx); err != nil {
			return fmt.Errorf("failed to render child node: %w", err)
		}
	}

	return nil
}

// RenderElement renders a layout element with background and border styling
func (r *PDFRenderer) RenderElement(elem *domain.LayoutNode, ctx RenderContext) error {
	// Render element background if present
	if err := r.renderBackground(elem.Style.Background, elem.Box, ctx); err != nil {
		return fmt.Errorf("failed to render background: %w", err)
	}

	// Render element border if present
	if err := r.renderBorder(elem.Style.Border, elem.Box, ctx); err != nil {
		return fmt.Errorf("failed to render border: %w", err)
	}

	return nil
}

// RenderText renders text content with proper font styling and positioning
func (r *PDFRenderer) RenderText(content string, style domain.ComputedStyle, ctx RenderContext) error {
	if content == "" {
		return nil // Skip empty content
	}

	// Configure font properties
	fontFamily := r.mapFontFamily(style.Font.Family)                 // Map CSS font to PDF font
	fontSize := style.Font.Size                                      // Font size in points
	fontStyle := r.mapFontStyle(style.Font.Weight, style.Font.Style) // Bold/italic styling

	// Apply font settings to PDF context
	ctx.PDF.SetFont(fontFamily, fontStyle, fontSize)

	// Configure text color from RGBA values
	red := float64(style.Color.R) / 255.0   // Normalize red component
	green := float64(style.Color.G) / 255.0 // Normalize green component
	blue := float64(style.Color.B) / 255.0  // Normalize blue component
	ctx.PDF.SetTextColor(int(red*255), int(green*255), int(blue*255))

	// Calculate text position (simplified positioning)
	x := 10.0 // Left margin in mm
	y := 20.0 // Top margin in mm

	// Render text at calculated position
	ctx.PDF.Text(x, y, content)

	return nil
}

// renderBackground renders element background with color and transparency support
func (r *PDFRenderer) renderBackground(bg domain.Background, bounds domain.Box, ctx RenderContext) error {
	if bg.Color.A == 0 {
		return nil // Skip transparent backgrounds
	}

	// Configure background fill color from RGBA values
	red := float64(bg.Color.R) / 255.0   // Normalize red component
	green := float64(bg.Color.G) / 255.0 // Normalize green component
	blue := float64(bg.Color.B) / 255.0  // Normalize blue component
	ctx.PDF.SetFillColor(int(red*255), int(green*255), int(blue*255))

	// Draw filled rectangle for background
	ctx.PDF.Rect(bounds.X, bounds.Y, bounds.Width, bounds.Height, "F")

	return nil
}

// renderBorder renders element border with support for different styles and colors
func (r *PDFRenderer) renderBorder(border domain.BorderStyle, bounds domain.Box, ctx RenderContext) error {
	if border.Width <= 0 {
		return nil // Skip borders with zero width
	}

	// Configure border line width
	ctx.PDF.SetLineWidth(border.Width)

	// Configure border color from RGBA values
	red := float64(border.Color.R) / 255.0   // Normalize red component
	green := float64(border.Color.G) / 255.0 // Normalize green component
	blue := float64(border.Color.B) / 255.0  // Normalize blue component
	ctx.PDF.SetDrawColor(int(red*255), int(green*255), int(blue*255))

	// Render border based on specified style
	switch border.Style {
	case domain.BorderSolid:
		// Draw solid border rectangle
		ctx.PDF.Rect(bounds.X, bounds.Y, bounds.Width, bounds.Height, "D")
	case domain.BorderDashed:
		// Configure and draw dashed border
		ctx.PDF.SetDashPattern([]float64{5, 3}, 0) // 5mm dash, 3mm gap
		ctx.PDF.Rect(bounds.X, bounds.Y, bounds.Width, bounds.Height, "D")
		ctx.PDF.SetDashPattern([]float64{}, 0) // Reset to solid
	case domain.BorderDotted:
		// Configure and draw dotted border
		ctx.PDF.SetDashPattern([]float64{1, 2}, 0) // 1mm dot, 2mm gap
		ctx.PDF.Rect(bounds.X, bounds.Y, bounds.Width, bounds.Height, "D")
		ctx.PDF.SetDashPattern([]float64{}, 0) // Reset to solid
	}

	return nil
}

// mapFontFamily maps CSS font family names to PDF-compatible font families
func (r *PDFRenderer) mapFontFamily(family string) string {
	family = strings.ToLower(family)
	switch {
	case strings.Contains(family, "serif"):
		return "Times" // Times New Roman for serif fonts
	case strings.Contains(family, "sans-serif"):
		return "Arial" // Arial for sans-serif fonts
	case strings.Contains(family, "monospace"):
		return "Courier" // Courier for monospace fonts
	default:
		return "Arial" // Default fallback font
	}
}

// mapFontStyle maps CSS font weight and style to PDF font style codes
func (r *PDFRenderer) mapFontStyle(weight int, style string) string {
	bold := weight >= 700                        // Bold if weight >= 700
	italic := strings.ToLower(style) == "italic" // Italic if style is italic

	// Combine bold and italic styles
	switch {
	case bold && italic:
		return "BI" // Bold + Italic
	case bold:
		return "B" // Bold only
	case italic:
		return "I" // Italic only
	default:
		return "" // Regular style
	}
}

// FontManager manages font resources and font loading for PDF rendering
type FontManager struct {
	fonts map[string]FontInfo // Map of font name to font information
}

// FontInfo represents detailed information about a font resource
type FontInfo struct {
	Family string // Font family name (e.g., "Arial", "Times")
	Style  string // Font style (e.g., "B", "I", "BI")
	Path   string // Path to font file (for custom fonts)
}

// NewFontManager creates a new font manager with initialized font registry
func NewFontManager() *FontManager {
	return &FontManager{
		fonts: make(map[string]FontInfo),
	}
}

// ImageCache manages caching of images for efficient PDF rendering
type ImageCache struct {
	cache map[string][]byte // Map of image URL/hash to image data
}

// NewImageCache creates a new image cache with initialized storage
func NewImageCache() *ImageCache {
	return &ImageCache{
		cache: make(map[string][]byte),
	}
}
