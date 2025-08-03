package domain

// PageSize represents page dimensions
type PageSize struct {
	Width  float64 `json:"width"`  // in mm
	Height float64 `json:"height"` // in mm
	Name   string  `json:"name"`
}

// Predefined page sizes
var (
	A4     = PageSize{Width: 210, Height: 297, Name: "A4"}
	Letter = PageSize{Width: 216, Height: 279, Name: "Letter"}
	Legal  = PageSize{Width: 216, Height: 356, Name: "Legal"}
	A3     = PageSize{Width: 297, Height: 420, Name: "A3"}
	A5     = PageSize{Width: 148, Height: 210, Name: "A5"}
)

// Margins represents page margins
type Margins struct {
	Top    float64 `json:"top"`    // in mm
	Right  float64 `json:"right"`  // in mm
	Bottom float64 `json:"bottom"` // in mm
	Left   float64 `json:"left"`   // in mm
}

// Orientation represents page orientation
type Orientation string

const (
	OrientationPortrait  Orientation = "portrait"
	OrientationLandscape Orientation = "landscape"
)

// Box represents a layout box with position and dimensions
type Box struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// TextAlign represents text alignment
type TextAlign string

const (
	TextAlignLeft    TextAlign = "left"
	TextAlignCenter  TextAlign = "center"
	TextAlignRight   TextAlign = "right"
	TextAlignJustify TextAlign = "justify"
)

// VerticalAlign represents vertical alignment
type VerticalAlign string

const (
	VerticalAlignTop      VerticalAlign = "top"
	VerticalAlignMiddle   VerticalAlign = "middle"
	VerticalAlignBottom   VerticalAlign = "bottom"
	VerticalAlignBaseline VerticalAlign = "baseline"
)

// Display represents CSS display property
type Display string

const (
	DisplayBlock       Display = "block"
	DisplayInline      Display = "inline"
	DisplayInlineBlock Display = "inline-block"
	DisplayFlex        Display = "flex"
	DisplayGrid        Display = "grid"
	DisplayNone        Display = "none"
)

// Position represents CSS position property
type Position string

const (
	PositionStatic   Position = "static"
	PositionRelative Position = "relative"
	PositionAbsolute Position = "absolute"
	PositionFixed    Position = "fixed"
	PositionSticky   Position = "sticky"
)

// LayoutNode represents a node in the layout tree
type LayoutNode struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Box      Box           `json:"box"`
	Style    ComputedStyle `json:"style"`
	Children []*LayoutNode `json:"children"`
	Parent   *LayoutNode   `json:"-"`
	Content  string        `json:"content,omitempty"`
}

// ComputedStyle represents computed CSS styles
type ComputedStyle struct {
	Display    Display     `json:"display"`
	Position   Position    `json:"position"`
	Width      string      `json:"width"`
	Height     string      `json:"height"`
	Margin     Margins     `json:"margin"`
	Padding    Margins     `json:"padding"`
	Border     BorderStyle `json:"border"`
	Background Background  `json:"background"`
	Font       FontStyle   `json:"font"`
	Text       TextStyle   `json:"text"`
	Color      Color       `json:"color"`
	ZIndex     int         `json:"z_index"`
}

// BorderStyle represents border styling
type BorderStyle struct {
	Width float64    `json:"width"`
	Style BorderType `json:"style"`
	Color Color      `json:"color"`
}

// BorderType represents border line style
type BorderType string

const (
	BorderSolid  BorderType = "solid"
	BorderDashed BorderType = "dashed"
	BorderDotted BorderType = "dotted"
	BorderDouble BorderType = "double"
	BorderNone   BorderType = "none"
)

// Background represents background styling
type Background struct {
	Color  Color  `json:"color"`
	Image  string `json:"image"`
	Repeat string `json:"repeat"`
	Size   string `json:"size"`
}

// FontStyle represents font styling
type FontStyle struct {
	Family string  `json:"family"`
	Size   float64 `json:"size"`
	Weight int     `json:"weight"`
	Style  string  `json:"style"`
}

// TextStyle represents text styling
type TextStyle struct {
	Align       TextAlign `json:"align"`
	Decoration  string    `json:"decoration"`
	Transform   string    `json:"transform"`
	LineHeight  float64   `json:"line_height"`
	LetterSpace float64   `json:"letter_spacing"`
	WordSpace   float64   `json:"word_spacing"`
}

// Color represents a color value
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}
