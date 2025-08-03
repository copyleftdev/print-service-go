package render

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"

	"print-service/internal/core/domain"

	"github.com/fogleman/gg"
)

// ImageRenderer handles image generation
type ImageRenderer struct {
	fontManager *FontManager
	options     ImageRenderOptions
}

// ImageRenderOptions configures image rendering
type ImageRenderOptions struct {
	Antialias     bool
	Interpolation InterpolationType
	ColorSpace    ColorSpace
	Quality       int
	Optimization  bool
}

// InterpolationType represents image interpolation types
type InterpolationType int

const (
	InterpolationNone InterpolationType = iota
	InterpolationLinear
	InterpolationBilinear
	InterpolationBicubic
)

// ColorSpace represents color space for image rendering
type ColorSpace string

const (
	ColorSpaceRGB  ColorSpace = "RGB"
	ColorSpaceCMYK ColorSpace = "CMYK"
	ColorSpaceGray ColorSpace = "Gray"
)

// ImageRenderContext provides context for image rendering
type ImageRenderContext struct {
	Canvas     *gg.Context
	Width      int
	Height     int
	DPI        float64
	Scale      float64
	Background domain.Color
}

// NewImageRenderer creates a new image renderer
func NewImageRenderer(opts ImageRenderOptions) *ImageRenderer {
	return &ImageRenderer{
		fontManager: NewFontManager(),
		options:     opts,
	}
}

// Render renders a layout to image
func (r *ImageRenderer) Render(layout *domain.LayoutNode, options domain.PrintOptions) ([]byte, error) {
	// Calculate image dimensions
	width := int(options.Page.Size.Width * options.Page.Scale * float64(options.Layout.DPI) / 25.4) // Convert mm to pixels
	height := int(options.Page.Size.Height * options.Page.Scale * float64(options.Layout.DPI) / 25.4)

	// Create canvas
	canvas := gg.NewContext(width, height)

	// Set background
	if options.Page.Background {
		canvas.SetRGB(1, 1, 1) // White background
		canvas.Clear()
	}

	// Create render context
	ctx := ImageRenderContext{
		Canvas:     canvas,
		Width:      width,
		Height:     height,
		DPI:        float64(options.Layout.DPI),
		Scale:      options.Page.Scale,
		Background: domain.Color{R: 255, G: 255, B: 255, A: 255},
	}

	// Enable antialiasing if configured
	if r.options.Antialias {
		canvas.SetLineCapRound()
		canvas.SetLineJoinRound()
	}

	// Render the layout tree
	if err := r.renderLayoutNode(layout, ctx); err != nil {
		return nil, fmt.Errorf("failed to render layout: %w", err)
	}

	// Export based on output format
	switch options.Output.Format {
	case domain.FormatPNG:
		return r.ExportPNG(canvas)
	case domain.FormatJPEG:
		return r.ExportJPEG(canvas, r.options.Quality)
	default:
		return r.ExportPNG(canvas)
	}
}

// renderLayoutNode renders a layout node and its children
func (r *ImageRenderer) renderLayoutNode(node *domain.LayoutNode, ctx ImageRenderContext) error {
	if node == nil {
		return nil
	}

	// Render based on node type
	switch node.Type {
	case "text":
		if err := r.RenderText(node.Content, node.Style, ctx); err != nil {
			return err
		}
	case "element":
		if err := r.RenderElement(node, ctx); err != nil {
			return err
		}
	}

	// Render children
	for _, child := range node.Children {
		if err := r.renderLayoutNode(child, ctx); err != nil {
			return err
		}
	}

	return nil
}

// RenderElement renders a layout element
func (r *ImageRenderer) RenderElement(elem *domain.LayoutNode, ctx ImageRenderContext) error {
	// Render background
	if err := r.RenderBackground(elem.Style.Background, elem.Box, ctx); err != nil {
		return err
	}

	// Render border
	if err := r.renderBorder(elem.Style.Border, elem.Box, ctx); err != nil {
		return err
	}

	return nil
}

// RenderText renders text content
func (r *ImageRenderer) RenderText(content string, style domain.ComputedStyle, ctx ImageRenderContext) error {
	if content == "" {
		return nil
	}

	// Set font (simplified - would need proper font loading)
	fontSize := style.Font.Size * ctx.Scale
	if err := ctx.Canvas.LoadFontFace("", fontSize); err != nil {
		// Fallback to default font
		ctx.Canvas.SetFontFace(nil)
	}

	// Set text color
	red := float64(style.Color.R) / 255.0
	green := float64(style.Color.G) / 255.0
	blue := float64(style.Color.B) / 255.0
	alpha := float64(style.Color.A) / 255.0
	ctx.Canvas.SetRGBA(red, green, blue, alpha)

	// Calculate position (simplified)
	x := 10.0 * ctx.Scale // Default margin
	y := 20.0 * ctx.Scale // Default margin

	// Draw text
	ctx.Canvas.DrawString(content, x, y)

	return nil
}

// RenderBackground renders background styling
func (r *ImageRenderer) RenderBackground(bg domain.Background, bounds domain.Box, ctx ImageRenderContext) error {
	if bg.Color.A == 0 {
		return nil // Transparent background
	}

	// Set fill color
	red := float64(bg.Color.R) / 255.0
	green := float64(bg.Color.G) / 255.0
	blue := float64(bg.Color.B) / 255.0
	alpha := float64(bg.Color.A) / 255.0
	ctx.Canvas.SetRGBA(red, green, blue, alpha)

	// Draw rectangle
	x := bounds.X * ctx.Scale
	y := bounds.Y * ctx.Scale
	width := bounds.Width * ctx.Scale
	height := bounds.Height * ctx.Scale

	ctx.Canvas.DrawRectangle(x, y, width, height)
	ctx.Canvas.Fill()

	return nil
}

// renderBorder renders border styling
func (r *ImageRenderer) renderBorder(border domain.BorderStyle, bounds domain.Box, ctx ImageRenderContext) error {
	if border.Width <= 0 {
		return nil
	}

	// Set line width
	ctx.Canvas.SetLineWidth(border.Width * ctx.Scale)

	// Set border color
	red := float64(border.Color.R) / 255.0
	green := float64(border.Color.G) / 255.0
	blue := float64(border.Color.B) / 255.0
	alpha := float64(border.Color.A) / 255.0
	ctx.Canvas.SetRGBA(red, green, blue, alpha)

	// Calculate bounds
	x := bounds.X * ctx.Scale
	y := bounds.Y * ctx.Scale
	width := bounds.Width * ctx.Scale
	height := bounds.Height * ctx.Scale

	// Draw border based on style
	switch border.Style {
	case domain.BorderSolid:
		ctx.Canvas.DrawRectangle(x, y, width, height)
		ctx.Canvas.Stroke()
	case domain.BorderDashed:
		r.drawDashedRectangle(ctx.Canvas, x, y, width, height, []float64{10, 5})
	case domain.BorderDotted:
		r.drawDashedRectangle(ctx.Canvas, x, y, width, height, []float64{2, 3})
	}

	return nil
}

// drawDashedRectangle draws a dashed rectangle
func (r *ImageRenderer) drawDashedRectangle(canvas *gg.Context, x, y, width, height float64, pattern []float64) {
	// Top edge
	r.drawDashedLine(canvas, x, y, x+width, y, pattern)
	// Right edge
	r.drawDashedLine(canvas, x+width, y, x+width, y+height, pattern)
	// Bottom edge
	r.drawDashedLine(canvas, x+width, y+height, x, y+height, pattern)
	// Left edge
	r.drawDashedLine(canvas, x, y+height, x, y, pattern)
}

// drawDashedLine draws a dashed line
func (r *ImageRenderer) drawDashedLine(canvas *gg.Context, x1, y1, x2, y2 float64, pattern []float64) {
	if len(pattern) == 0 {
		canvas.DrawLine(x1, y1, x2, y2)
		canvas.Stroke()
		return
	}

	// Simplified dashed line implementation
	totalLength := ((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	if totalLength <= 0 {
		return
	}

	// For simplicity, just draw a regular line
	// In a full implementation, this would properly handle dash patterns
	canvas.DrawLine(x1, y1, x2, y2)
	canvas.Stroke()
}

// ExportPNG exports the canvas as PNG
func (r *ImageRenderer) ExportPNG(canvas *gg.Context) ([]byte, error) {
	img := canvas.Image()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportJPEG exports the canvas as JPEG
func (r *ImageRenderer) ExportJPEG(canvas *gg.Context, quality int) ([]byte, error) {
	img := canvas.Image()

	var buf bytes.Buffer
	options := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, options); err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// RenderImage renders an image element
func (r *ImageRenderer) RenderImage(img *ImageContent, bounds domain.Box, ctx ImageRenderContext) error {
	// Simplified image rendering - would need proper image loading and scaling
	return nil
}

// ImageContent represents image content
type ImageContent struct {
	Data   []byte
	Format string
	Width  int
	Height int
}
