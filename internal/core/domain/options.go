package domain

import "time"

// PrintOptions represents options for printing a document
type PrintOptions struct {
	Page        PageOptions        `json:"page"`
	Layout      LayoutOptions      `json:"layout"`
	Render      RenderOptions      `json:"render"`
	Output      OutputOptions      `json:"output"`
	Performance PerformanceOptions `json:"performance"`
	Security    SecurityOptions    `json:"security"`
}

// PageOptions represents page-specific options
type PageOptions struct {
	Size        PageSize    `json:"size"`
	Orientation Orientation `json:"orientation"`
	Margins     Margins     `json:"margins"`
	Scale       float64     `json:"scale"`
	Background  bool        `json:"background"`
}

// LayoutOptions represents layout-specific options
type LayoutOptions struct {
	WaitForFonts   bool          `json:"wait_for_fonts"`
	WaitForImages  bool          `json:"wait_for_images"`
	WaitTimeout    time.Duration `json:"wait_timeout"`
	ViewportWidth  int           `json:"viewport_width"`
	ViewportHeight int           `json:"viewport_height"`
	DPI            int           `json:"dpi"`
	PrintMediaType bool          `json:"print_media_type"`
	EmulateMedia   string        `json:"emulate_media"`
}

// RenderOptions represents rendering-specific options
type RenderOptions struct {
	Quality         RenderQuality    `json:"quality"`
	ColorProfile    ColorProfile     `json:"color_profile"`
	Compression     CompressionLevel `json:"compression"`
	EmbedFonts      bool             `json:"embed_fonts"`
	OptimizeImages  bool             `json:"optimize_images"`
	GenerateOutline bool             `json:"generate_outline"`
	Accessibility   bool             `json:"accessibility"`
}

// OutputOptions represents output-specific options
type OutputOptions struct {
	Format      OutputFormat `json:"format"`
	Filename    string       `json:"filename"`
	Destination string       `json:"destination"`
	Metadata    bool         `json:"metadata"`
	Watermark   *Watermark   `json:"watermark,omitempty"`
}

// PerformanceOptions represents performance-specific options
type PerformanceOptions struct {
	EnableCache    bool           `json:"enable_cache"`
	CacheTTL       time.Duration  `json:"cache_ttl"`
	MaxMemory      int64          `json:"max_memory"`
	Timeout        time.Duration  `json:"timeout"`
	ConcurrentJobs int            `json:"concurrent_jobs"`
	ResourceLimits ResourceLimits `json:"resource_limits"`
}

// SecurityOptions represents security-specific options
type SecurityOptions struct {
	SanitizeHTML     bool     `json:"sanitize_html"`
	AllowedDomains   []string `json:"allowed_domains"`
	BlockedDomains   []string `json:"blocked_domains"`
	MaxFileSize      int64    `json:"max_file_size"`
	AllowJavaScript  bool     `json:"allow_javascript"`
	AllowExternalCSS bool     `json:"allow_external_css"`
}

// RenderQuality represents rendering quality levels
type RenderQuality string

const (
	QualityDraft  RenderQuality = "draft"
	QualityNormal RenderQuality = "normal"
	QualityHigh   RenderQuality = "high"
	QualityPrint  RenderQuality = "print"
)

// ColorProfile represents color profiles
type ColorProfile string

const (
	ColorProfileRGB  ColorProfile = "rgb"
	ColorProfileCMYK ColorProfile = "cmyk"
	ColorProfileGray ColorProfile = "gray"
	ColorProfilesRGB ColorProfile = "srgb"
)

// CompressionLevel represents compression levels
type CompressionLevel string

const (
	CompressionNone   CompressionLevel = "none"
	CompressionLow    CompressionLevel = "low"
	CompressionMedium CompressionLevel = "medium"
	CompressionHigh   CompressionLevel = "high"
)

// OutputFormat represents output formats
type OutputFormat string

const (
	FormatPDF  OutputFormat = "pdf"
	FormatPNG  OutputFormat = "png"
	FormatJPEG OutputFormat = "jpeg"
	FormatSVG  OutputFormat = "svg"
)

// Watermark represents watermark options
type Watermark struct {
	Text     string  `json:"text"`
	Image    string  `json:"image"`
	Opacity  float64 `json:"opacity"`
	Position string  `json:"position"`
	Scale    float64 `json:"scale"`
}

// ResourceLimits represents resource usage limits
type ResourceLimits struct {
	MaxCPU    float64       `json:"max_cpu"`
	MaxMemory int64         `json:"max_memory"`
	MaxDisk   int64         `json:"max_disk"`
	MaxTime   time.Duration `json:"max_time"`
}

// DefaultPrintOptions returns default print options
func DefaultPrintOptions() PrintOptions {
	return PrintOptions{
		Page: PageOptions{
			Size:        A4,
			Orientation: OrientationPortrait,
			Margins:     Margins{Top: 20, Right: 20, Bottom: 20, Left: 20},
			Scale:       1.0,
			Background:  true,
		},
		Layout: LayoutOptions{
			WaitForFonts:   true,
			WaitForImages:  true,
			WaitTimeout:    30 * time.Second,
			ViewportWidth:  1024,
			ViewportHeight: 768,
			DPI:            96,
			PrintMediaType: true,
			EmulateMedia:   "print",
		},
		Render: RenderOptions{
			Quality:         QualityNormal,
			ColorProfile:    ColorProfilesRGB,
			Compression:     CompressionMedium,
			EmbedFonts:      true,
			OptimizeImages:  true,
			GenerateOutline: false,
			Accessibility:   false,
		},
		Output: OutputOptions{
			Format:      FormatPDF,
			Destination: "local",
			Metadata:    true,
		},
		Performance: PerformanceOptions{
			EnableCache:    true,
			CacheTTL:       1 * time.Hour,
			MaxMemory:      512 * 1024 * 1024, // 512MB
			Timeout:        60 * time.Second,
			ConcurrentJobs: 4,
			ResourceLimits: ResourceLimits{
				MaxCPU:    0.8,
				MaxMemory: 1024 * 1024 * 1024,      // 1GB
				MaxDisk:   10 * 1024 * 1024 * 1024, // 10GB
				MaxTime:   5 * time.Minute,
			},
		},
		Security: SecurityOptions{
			SanitizeHTML:     true,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
			AllowJavaScript:  false,
			AllowExternalCSS: true,
		},
	}
}
