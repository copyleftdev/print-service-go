# Pure Go Print Service - Design Specification

## Project Structure

```
print-service/
├── cmd/
│   ├── server/
│   │   └── main.go                 # HTTP server entry point
│   └── worker/
│       └── main.go                 # Background worker entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── print.go            # Print request handlers
│   │   │   ├── health.go           # Health check handlers
│   │   │   └── metrics.go          # Metrics handlers
│   │   ├── middleware/
│   │   │   ├── auth.go             # Authentication middleware
│   │   │   ├── cors.go             # CORS middleware
│   │   │   ├── logging.go          # Request logging
│   │   │   └── ratelimit.go        # Rate limiting
│   │   ├── router.go               # Route definitions
│   │   └── server.go               # HTTP server setup
│   ├── core/
│   │   ├── domain/
│   │   │   ├── document.go         # Document domain types
│   │   │   ├── layout.go           # Layout domain types
│   │   │   ├── options.go          # Configuration types
│   │   │   └── errors.go           # Domain-specific errors
│   │   ├── engine/
│   │   │   ├── html/
│   │   │   │   ├── parser.go       # HTML parsing engine
│   │   │   │   ├── sanitizer.go    # HTML sanitization
│   │   │   │   └── validator.go    # HTML validation
│   │   │   ├── css/
│   │   │   │   ├── parser.go       # CSS parsing engine
│   │   │   │   ├── selector.go     # CSS selector engine
│   │   │   │   ├── cascade.go      # CSS cascade logic
│   │   │   │   └── computed.go     # Computed styles
│   │   │   ├── layout/
│   │   │   │   ├── engine.go       # Layout calculation engine
│   │   │   │   ├── box.go          # Box model calculations
│   │   │   │   ├── text.go         # Text layout engine
│   │   │   │   ├── flow.go         # Document flow engine
│   │   │   │   └── page.go         # Page breaking logic
│   │   │   └── render/
│   │   │       ├── pdf.go          # PDF rendering engine
│   │   │       ├── image.go        # Image rendering engine
│   │   │       ├── canvas.go       # Canvas abstraction
│   │   │       └── fonts.go        # Font rendering
│   │   └── services/
│   │       ├── print.go            # Print service orchestrator
│   │       ├── cache.go            # Caching service
│   │       ├── queue.go            # Job queue service
│   │       └── storage.go          # File storage service
│   ├── infrastructure/
│   │   ├── cache/
│   │   │   ├── redis.go            # Redis cache implementation
│   │   │   ├── memory.go           # In-memory cache
│   │   │   └── interface.go        # Cache interface
│   │   ├── storage/
│   │   │   ├── local.go            # Local file storage
│   │   │   ├── s3.go               # S3 storage implementation
│   │   │   ├── gcs.go              # Google Cloud Storage
│   │   │   └── interface.go        # Storage interface
│   │   ├── metrics/
│   │   │   ├── prometheus.go       # Prometheus metrics
│   │   │   ├── datadog.go          # DataDog metrics
│   │   │   └── interface.go        # Metrics interface
│   │   └── logger/
│   │       ├── structured.go       # Structured logging
│   │       ├── correlation.go      # Correlation ID handling
│   │       └── interface.go        # Logger interface
│   ├── pkg/
│   │   ├── pool/
│   │   │   ├── worker.go           # Worker pool implementation
│   │   │   ├── buffer.go           # Buffer pool
│   │   │   └── resource.go         # Resource pooling
│   │   ├── config/
│   │   │   ├── loader.go           # Configuration loader
│   │   │   ├── validator.go        # Config validation
│   │   │   └── types.go            # Config types
│   │   └── utils/
│   │       ├── hash.go             # Hashing utilities
│   │       ├── units.go            # Unit conversion
│   │       └── validation.go       # Input validation
│   └── tests/
│       ├── integration/
│       ├── unit/
│       └── fixtures/
├── assets/
│   ├── fonts/                      # System fonts
│   ├── templates/                  # HTML templates
│   └── styles/                     # CSS stylesheets
├── configs/
│   ├── development.yaml
│   ├── production.yaml
│   └── docker.yaml
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── docker-compose.yml
│   │   └── docker-compose.prod.yml
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   └── monitoring/
│       ├── prometheus.yml
│       └── grafana-dashboard.json
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Core Domain Types

### `internal/core/domain/document.go`
```go
package domain

import (
    "time"
    "golang.org/x/net/html"
)

// Document represents a parsed HTML document with metadata
type Document struct {
    ID          string
    Root        *html.Node
    Styles      *Stylesheet
    Meta        *Metadata
    Images      []ImageRef
    Fonts       []FontRef
    CreatedAt   time.Time
    Hash        string
}

// Stylesheet contains parsed CSS rules
type Stylesheet struct {
    Rules       []Rule
    MediaQueries []MediaQuery
    FontFaces   []FontFace
    Variables   map[string]string
}

// Metadata contains document metadata
type Metadata struct {
    Title       string
    Author      string
    Subject     string
    Keywords    []string
    Language    string
    Charset     string
}

// ImageRef references an image in the document
type ImageRef struct {
    URL         string
    AltText     string
    Width       int
    Height      int
    MimeType    string
    Data        []byte
}

// FontRef references a font in the document
type FontRef struct {
    Family      string
    Weight      int
    Style       string
    URL         string
    Data        []byte
}

func NewDocument(html string) (*Document, error)
func (d *Document) Validate() error
func (d *Document) GetElement(id string) *html.Node
func (d *Document) ExtractText() string
func (d *Document) ComputeHash() string
```

### `internal/core/domain/layout.go`
```go
package domain

// Layout represents the computed layout of a document
type Layout struct {
    Document    *Document
    Pages       []*Page
    PageSize    PageSize
    Margins     Margins
    Fonts       []ComputedFont
    Images      []ComputedImage
    Metadata    *LayoutMetadata
}

// Page represents a single page in the layout
type Page struct {
    Number      int
    Elements    []LayoutElement
    Width       float64
    Height      float64
    Margins     Margins
    Background  *Background
}

// LayoutElement represents an element positioned on a page
type LayoutElement struct {
    Type        ElementType
    Node        *html.Node
    BoundingBox Rectangle
    Styles      ComputedStyles
    Content     *Content
    Children    []LayoutElement
}

// Rectangle defines a positioned rectangle
type Rectangle struct {
    X, Y        float64
    Width       float64
    Height      float64
}

// ComputedStyles contains all computed CSS properties
type ComputedStyles struct {
    Display     DisplayType
    Position    PositionType
    Float       FloatType
    Clear       ClearType
    Overflow    OverflowType
    Typography  Typography
    BoxModel    BoxModel
    Background  Background
    Border      Border
    Transform   Transform
}

// Content represents element content
type Content struct {
    Text        string
    Image       *ImageContent
    Background  *BackgroundContent
}

func NewLayout(doc *Document, options LayoutOptions) (*Layout, error)
func (l *Layout) CalculatePages() error
func (l *Layout) GetPageCount() int
func (l *Layout) ValidateLayout() error
func (l *Layout) OptimizeLayout() error
```

### `internal/core/domain/options.go`
```go
package domain

// PrintOptions defines configuration for print operations
type PrintOptions struct {
    Type        PrintType
    PDF         *PDFOptions
    Image       *ImageOptions
    Template    *TemplateOptions
    Cache       *CacheOptions
    Quality     Quality
}

// PDFOptions configures PDF generation
type PDFOptions struct {
    PageSize    PageSize
    Orientation Orientation
    Margins     Margins
    DPI         int
    Compression bool
    Encryption  *PDFEncryption
    Metadata    *PDFMetadata
    Fonts       FontOptions
}

// ImageOptions configures image generation
type ImageOptions struct {
    Width       int
    Height      int
    Format      ImageFormat
    Quality     int
    DPI         int
    Background  Color
    Scale       float64
    Antialias   bool
}

// PageSize defines page dimensions
type PageSize struct {
    Width       float64  // mm
    Height      float64  // mm
    Name        string   // A4, Letter, etc.
}

// Standard page sizes
var (
    A4     = PageSize{Width: 210, Height: 297, Name: "A4"}
    Letter = PageSize{Width: 216, Height: 279, Name: "Letter"}
    Legal  = PageSize{Width: 216, Height: 356, Name: "Legal"}
)

func (o *PrintOptions) Validate() error
func (o *PrintOptions) SetDefaults()
func (p *PDFOptions) ToGofpdfOptions() interface{}
func (i *ImageOptions) ToCanvasOptions() interface{}
```

## Core Engine Components

### `internal/core/engine/html/parser.go`
```go
package html

import (
    "golang.org/x/net/html"
    "github.com/PuerkitoBio/goquery"
)

// Parser handles HTML parsing and DOM manipulation
type Parser struct {
    sanitizer   *Sanitizer
    validator   *Validator
    options     ParserOptions
}

// ParserOptions configures HTML parsing
type ParserOptions struct {
    StrictMode      bool
    SanitizeHTML    bool
    ValidateHTML    bool
    PreserveSpace   bool
    ErrorHandler    ErrorHandler
}

// ParseResult contains parsing results and metadata
type ParseResult struct {
    Document    *html.Node
    Errors      []ParseError
    Warnings    []ParseWarning
    Stats       ParseStats
}

func NewParser(opts ParserOptions) *Parser
func (p *Parser) Parse(htmlContent string) (*ParseResult, error)
func (p *Parser) ParseFromReader(r io.Reader) (*ParseResult, error)
func (p *Parser) ExtractStyles(doc *html.Node) ([]string, error)
func (p *Parser) ExtractImages(doc *html.Node) ([]ImageRef, error)
func (p *Parser) ExtractFonts(doc *html.Node) ([]FontRef, error)
func (p *Parser) ExtractMeta(doc *html.Node) (*Metadata, error)
func (p *Parser) BuildDocumentTree(node *html.Node) (*Document, error)
func (p *Parser) NormalizeWhitespace(doc *html.Node)
func (p *Parser) ResolveRelativeURLs(doc *html.Node, baseURL string)
```

### `internal/core/engine/css/parser.go`
```go
package css

// Parser handles CSS parsing and processing
type Parser struct {
    selector    *SelectorEngine
    cascade     *CascadeEngine
    computer    *ComputedStyleEngine
    options     ParserOptions
}

// ParserOptions configures CSS parsing
type ParserOptions struct {
    StrictMode      bool
    SupportLevel    CSSLevel
    VendorPrefixes  []string
    ErrorHandler    ErrorHandler
}

// Rule represents a CSS rule
type Rule struct {
    Selector    Selector
    Properties  []Property
    Specificity Specificity
    Important   bool
    SourceMap   SourceLocation
}

// Selector represents a CSS selector
type Selector struct {
    Type        SelectorType
    Value       string
    Combinator  Combinator
    Pseudo      []PseudoSelector
    Specificity Specificity
}

func NewParser(opts ParserOptions) *Parser
func (p *Parser) Parse(cssContent string) (*Stylesheet, error)
func (p *Parser) ParseRule(rule string) (*Rule, error)
func (p *Parser) ParseProperty(prop string) (*Property, error)
func (p *Parser) ResolveImports(css string, baseURL string) (string, error)
func (p *Parser) ValidateCSS(css string) []ValidationError
func (p *Parser) OptimizeCSS(css string) string
func (p *Parser) ExtractFontFaces(css string) ([]FontFace, error)
func (p *Parser) ExtractMediaQueries(css string) ([]MediaQuery, error)
```

### `internal/core/engine/layout/engine.go`
```go
package layout

// Engine handles layout calculations and positioning
type Engine struct {
    boxCalculator   *BoxCalculator
    textEngine      *TextEngine
    flowEngine      *FlowEngine
    pageBreaker     *PageBreaker
    options         EngineOptions
}

// EngineOptions configures layout engine behavior
type EngineOptions struct {
    DefaultFontSize float64
    DefaultMargins  Margins
    TextDirection   TextDirection
    WritingMode     WritingMode
    LineHeight      float64
    WordSpacing     float64
    LetterSpacing   float64
}

// LayoutContext provides context for layout calculations
type LayoutContext struct {
    ContainingBlock Rectangle
    AvailableWidth  float64
    AvailableHeight float64
    CurrentPage     int
    LineHeight      float64
    FontMetrics     FontMetrics
}

func NewEngine(opts EngineOptions) *Engine
func (e *Engine) CalculateLayout(doc *Document, pageSize PageSize) (*Layout, error)
func (e *Engine) CalculateElementLayout(elem *html.Node, ctx LayoutContext) (*LayoutElement, error)
func (e *Engine) PositionElements(layout *Layout) error
func (e *Engine) HandlePageBreaks(layout *Layout) error
func (e *Engine) OptimizeLayout(layout *Layout) error
func (e *Engine) ValidateLayout(layout *Layout) []LayoutError
func (e *Engine) GetIntrinsicDimensions(elem *html.Node) (float64, float64, error)
func (e *Engine) CalculateContentSize(elem *html.Node) (float64, float64, error)
```

### `internal/core/engine/render/pdf.go`
```go
package render

import "github.com/jung-kurt/gofpdf"

// PDFRenderer handles PDF generation
type PDFRenderer struct {
    fontManager *FontManager
    imageCache  *ImageCache
    options     PDFRenderOptions
}

// PDFRenderOptions configures PDF rendering
type PDFRenderOptions struct {
    Compression     bool
    EmbedFonts      bool
    OptimizeImages  bool
    ColorProfile    ColorProfile
    OutputIntent    OutputIntent
    PDFVersion      string
}

// RenderContext provides context for PDF rendering
type RenderContext struct {
    PDF         *gofpdf.Fpdf
    CurrentPage int
    PageWidth   float64
    PageHeight  float64
    DPI         float64
    Scale       float64
}

func NewPDFRenderer(opts PDFRenderOptions) *PDFRenderer
func (r *PDFRenderer) Render(layout *Layout, options PDFOptions) ([]byte, error)
func (r *PDFRenderer) RenderPage(page *Page, ctx RenderContext) error
func (r *PDFRenderer) RenderElement(elem *LayoutElement, ctx RenderContext) error
func (r *PDFRenderer) RenderText(content string, styles ComputedStyles, ctx RenderContext) error
func (r *PDFRenderer) RenderImage(img *ImageContent, bounds Rectangle, ctx RenderContext) error
func (r *PDFRenderer) RenderBackground(bg *Background, bounds Rectangle, ctx RenderContext) error
func (r *PDFRenderer) RenderBorder(border *Border, bounds Rectangle, ctx RenderContext) error
func (r *PDFRenderer) EmbedFonts(fonts []FontRef, pdf *gofpdf.Fpdf) error
func (r *PDFRenderer) OptimizeOutput(pdf *gofpdf.Fpdf) error
```

### `internal/core/engine/render/image.go`
```go
package render

import "github.com/fogleman/gg"

// ImageRenderer handles image generation
type ImageRenderer struct {
    fontManager *FontManager
    options     ImageRenderOptions
}

// ImageRenderOptions configures image rendering
type ImageRenderOptions struct {
    Antialias       bool
    Interpolation   InterpolationType
    ColorSpace      ColorSpace
    Quality         int
    Optimization    bool
}

// ImageRenderContext provides context for image rendering
type ImageRenderContext struct {
    Canvas      *gg.Context
    Width       int
    Height      int
    DPI         float64
    Scale       float64
    Background  Color
}

func NewImageRenderer(opts ImageRenderOptions) *ImageRenderer
func (r *ImageRenderer) Render(layout *Layout, options ImageOptions) ([]byte, error)
func (r *ImageRenderer) RenderToCanvas(layout *Layout, ctx ImageRenderContext) error
func (r *ImageRenderer) RenderElement(elem *LayoutElement, ctx ImageRenderContext) error
func (r *ImageRenderer) RenderText(content string, styles ComputedStyles, ctx ImageRenderContext) error
func (r *ImageRenderer) RenderImage(img *ImageContent, bounds Rectangle, ctx ImageRenderContext) error
func (r *ImageRenderer) RenderBackground(bg *Background, bounds Rectangle, ctx ImageRenderContext) error
func (r *ImageRenderer) ExportPNG(canvas *gg.Context) ([]byte, error)
func (r *ImageRenderer) ExportJPEG(canvas *gg.Context, quality int) ([]byte, error)
func (r *ImageRenderer) ExportWebP(canvas *gg.Context, quality int) ([]byte, error)
```

## Service Layer

### `internal/core/services/print.go`
```go
package services

// PrintService orchestrates the print pipeline
type PrintService struct {
    htmlEngine    *html.Parser
    cssEngine     *css.Parser
    layoutEngine  *layout.Engine
    pdfRenderer   *render.PDFRenderer
    imageRenderer *render.ImageRenderer
    cache         cache.Interface
    storage       storage.Interface
    metrics       metrics.Interface
    logger        logger.Interface
}

// PrintRequest represents a print job request
type PrintRequest struct {
    ID            string
    HTML          string
    CSS           string
    Type          PrintType
    Options       PrintOptions
    Priority      Priority
    CorrelationID string
    CreatedAt     time.Time
    Timeout       time.Duration
}

// PrintResult represents the result of a print operation
type PrintResult struct {
    ID           string
    Data         []byte
    ContentType  string
    FileSize     int64
    Duration     time.Duration
    Error        error
    Metrics      PrintMetrics
    CacheHit     bool
    CreatedAt    time.Time
}

func NewPrintService(deps ServiceDependencies) *PrintService
func (s *PrintService) ProcessSync(ctx context.Context, req PrintRequest) (*PrintResult, error)
func (s *PrintService) ProcessAsync(ctx context.Context, req PrintRequest) (string, error)
func (s *PrintService) GetJobStatus(jobID string) (*JobStatus, error)
func (s *PrintService) CancelJob(jobID string) error
func (s *PrintService) ValidateRequest(req PrintRequest) error
func (s *PrintService) EstimateProcessingTime(req PrintRequest) time.Duration
func (s *PrintService) OptimizeRequest(req PrintRequest) PrintRequest
```

### `internal/core/services/cache.go`
```go
package services

// CacheService handles template and result caching
type CacheService struct {
    backend     cache.Interface
    compressor  Compressor
    serializer  Serializer
    options     CacheOptions
    metrics     metrics.Interface
}

// CacheOptions configures caching behavior
type CacheOptions struct {
    TTL             time.Duration
    MaxSize         int64
    EvictionPolicy  EvictionPolicy
    Compression     bool
    Serialization   SerializationType
    Sharding        int
}

// CacheKey represents a cache key with metadata
type CacheKey struct {
    Type        string
    Hash        string
    Version     string
    Namespace   string
    TTL         time.Duration
}

func NewCacheService(backend cache.Interface, opts CacheOptions) *CacheService
func (s *CacheService) Get(key CacheKey) ([]byte, bool, error)
func (s *CacheService) Set(key CacheKey, value []byte) error
func (s *CacheService) Delete(key CacheKey) error
func (s *CacheService) Clear(namespace string) error
func (s *CacheService) GetStats() CacheStats
func (s *CacheService) GetTemplateCache() *TemplateCache
func (s *CacheService) InvalidatePattern(pattern string) error
func (s *CacheService) WarmupCache(templates []string) error
```

## Worker Pool System

### `internal/pkg/pool/worker.go`
```go
package pool

// WorkerPool manages a pool of print workers
type WorkerPool struct {
    workers     []*Worker
    jobQueue    chan Job
    resultQueue chan Result
    options     PoolOptions
    metrics     *PoolMetrics
    logger      logger.Interface
    mu          sync.RWMutex
    running     bool
}

// PoolOptions configures worker pool behavior
type PoolOptions struct {
    WorkerCount     int
    QueueSize       int
    MaxConcurrency  int
    IdleTimeout     time.Duration
    GracefulTimeout time.Duration
    HealthCheck     time.Duration
}

// Worker represents a single print worker
type Worker struct {
    ID          int
    printSvc    *services.PrintService
    jobChan     <-chan Job
    resultChan  chan<- Result
    quitChan    chan bool
    status      WorkerStatus
    currentJob  *Job
    stats       WorkerStats
    mu          sync.RWMutex
}

// Job represents a print job
type Job struct {
    ID          string
    Request     services.PrintRequest
    Priority    Priority
    CreatedAt   time.Time
    StartedAt   *time.Time
    Deadline    time.Time
    Retries     int
    Context     context.Context
}

func NewWorkerPool(opts PoolOptions, printSvc *services.PrintService) *WorkerPool
func (p *WorkerPool) Start() error
func (p *WorkerPool) Stop() error
func (p *WorkerPool) Submit(job Job) error
func (p *WorkerPool) GetResult(jobID string) (*Result, error)
func (p *WorkerPool) GetStats() PoolStats
func (p *WorkerPool) GetWorkerStats() []WorkerStats
func (p *WorkerPool) ScaleWorkers(count int) error
func (p *WorkerPool) HealthCheck() error
```

## API Layer

### `internal/api/handlers/print.go`
```go
package handlers

// PrintHandler handles print API requests
type PrintHandler struct {
    printSvc    *services.PrintService
    workerPool  *pool.WorkerPool
    validator   *validator.Validator
    metrics     metrics.Interface
    logger      logger.Interface
}

// PrintSyncRequest represents a synchronous print request
type PrintSyncRequest struct {
    HTML         string       `json:"html" validate:"required,max=10485760"`
    CSS          string       `json:"css,omitempty" validate:"max=1048576"`
    Type         PrintType    `json:"type" validate:"required,oneof=PDF PNG JPEG"`
    Options      interface{}  `json:"options,omitempty"`
    CorrelationID string      `json:"correlationId,omitempty"`
}

// PrintAsyncRequest represents an asynchronous print request
type PrintAsyncRequest struct {
    PrintSyncRequest
    Priority    Priority `json:"priority,omitempty" validate:"oneof=normal high urgent"`
    CallbackURL string   `json:"callbackUrl,omitempty" validate:"omitempty,url"`
    Webhook     *Webhook `json:"webhook,omitempty"`
}

func NewPrintHandler(deps HandlerDependencies) *PrintHandler
func (h *PrintHandler) PrintPDFSync(c *gin.Context)
func (h *PrintHandler) PrintImageSync(c *gin.Context)
func (h *PrintHandler) PrintAsync(c *gin.Context)
func (h *PrintHandler) GetJobStatus(c *gin.Context)
func (h *PrintHandler) DownloadResult(c *gin.Context)
func (h *PrintHandler) CancelJob(c *gin.Context)
func (h *PrintHandler) ListJobs(c *gin.Context)
func (h *PrintHandler) validateRequest(req interface{}) error
func (h *PrintHandler) buildPrintRequest(req PrintSyncRequest) services.PrintRequest
```

### `internal/api/middleware/ratelimit.go`
```go
package middleware

// RateLimiter implements request rate limiting
type RateLimiter struct {
    store       RateLimitStore
    algorithms  []Algorithm
    keyGen      KeyGenerator
    options     RateLimitOptions
    metrics     metrics.Interface
}

// RateLimitOptions configures rate limiting behavior
type RateLimitOptions struct {
    Requests      int
    Window        time.Duration
    BurstSize     int
    SkipPaths     []string
    SkipIPs       []string
    ErrorMessage  string
    RetryAfter    bool
}

func NewRateLimiter(store RateLimitStore, opts RateLimitOptions) *RateLimiter
func (r *RateLimiter) Middleware() gin.HandlerFunc
func (r *RateLimiter) Allow(key string) (bool, time.Duration, error)
func (r *RateLimiter) Reset(key string) error
func (r *RateLimiter) GetStats(key string) (*RateLimitStats, error)
```

## Configuration Management

### `internal/pkg/config/types.go`
```go
package config

// Config represents the application configuration
type Config struct {
    Server      ServerConfig      `yaml:"server"`
    WorkerPool  WorkerPoolConfig  `yaml:"workerPool"`
    Cache       CacheConfig       `yaml:"cache"`
    Storage     StorageConfig     `yaml:"storage"`
    Metrics     MetricsConfig     `yaml:"metrics"`
    Logging     LoggingConfig     `yaml:"logging"`
    Security    SecurityConfig    `yaml:"security"`
    Performance PerformanceConfig `yaml:"performance"`
}

// ServerConfig configures the HTTP server
type ServerConfig struct {
    Host            string        `yaml:"host" default:"0.0.0.0"`
    Port            int           `yaml:"port" default:"8080"`
    ReadTimeout     time.Duration `yaml:"readTimeout" default:"30s"`
    WriteTimeout    time.Duration `yaml:"writeTimeout" default:"300s"`
    MaxHeaderSize   int           `yaml:"maxHeaderSize" default:"1048576"`
    TLSConfig       *TLSConfig    `yaml:"tls,omitempty"`
}

// WorkerPoolConfig configures the worker pool
type WorkerPoolConfig struct {
    WorkerCount     int           `yaml:"workerCount" default:"100"`
    QueueSize       int           `yaml:"queueSize" default:"10000"`
    MaxConcurrency  int           `yaml:"maxConcurrency" default:"1000"`
    IdleTimeout     time.Duration `yaml:"idleTimeout" default:"60s"`
    GracefulTimeout time.Duration `yaml:"gracefulTimeout" default:"30s"`
}

func LoadConfig(path string) (*Config, error)
func (c *Config) Validate() error
func (c *Config) SetDefaults()
func (c *Config) GetEnvironmentOverrides() map[string]interface{}
```

## Performance Specifications

### Target Metrics
```go
// Performance targets for different document types
const (
    SmallDocumentTarget  = 5 * time.Millisecond   // 1 page
    MediumDocumentTarget = 20 * time.Millisecond  // 5 pages  
    LargeDocumentTarget  = 200 * time.Millisecond // 50 pages
    
    SmallDocumentMemory  = 2 * 1024 * 1024   // 2MB
    MediumDocumentMemory = 5 * 1024 * 1024   // 5MB
    LargeDocumentMemory  = 20 * 1024 * 1024  // 20MB
    
    TargetThroughputSync  = 1000 // req/sec
    TargetThroughputAsync = 5000 // req/sec
    TargetConcurrency    = 10000 // simultaneous requests
)
```

## Testing Structure

### `internal/tests/unit/`
- Engine unit tests (HTML, CSS, Layout, Render)
- Service unit tests
- Utility unit tests
- Mock implementations

### `internal/tests/integration/`
- End-to-end API tests
- Performance benchmarks
- Cache integration tests
- Storage integration tests

### `internal/tests/fixtures/`
- Sample HTML documents
- Test CSS files
- Expected output files
- Performance test data

This design provides a clean, modular architecture following Go best practices with clear separation of concerns, comprehensive error handling, and high-performance characteristics suitable for enterprise deployment.