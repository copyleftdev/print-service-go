package services

import (
	"context"
	"fmt"
	"time"

	"print-service/internal/core/domain"
	"print-service/internal/core/engine/css"
	"print-service/internal/core/engine/html"
	"print-service/internal/core/engine/layout"
	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/config"
)

// PrintService orchestrates the document printing process
type PrintService struct {
	htmlParser     *html.Parser
	cssParser      *css.Parser
	layoutEngine   *layout.Engine
	cacheService   *CacheService
	storageService *StorageService
	logger         logger.Logger
	config         config.PrintConfig
}

// NewPrintService creates a new print service
func NewPrintService(cfg config.PrintConfig, logger logger.Logger) (*PrintService, error) {
	// Initialize HTML components
	sanitizer := html.NewSanitizer()
	validator := html.NewValidator(false)
	htmlParser := html.NewParser(sanitizer, validator)

	// Initialize CSS parser
	cssParser := css.NewParser(false)

	// Initialize layout engine
	layoutEngine := layout.NewEngine()

	// Initialize cache and storage services (simplified for now)
	cacheService := NewCacheService()
	storageService := NewStorageService(cfg.OutputDirectory)

	return &PrintService{
		htmlParser:     htmlParser,
		cssParser:      cssParser,
		layoutEngine:   layoutEngine,
		cacheService:   cacheService,
		storageService: storageService,
		logger:         logger.With("service", "print"),
		config:         cfg,
	}, nil
}

// ProcessDocument processes a document and generates output
func (ps *PrintService) ProcessDocument(ctx context.Context, doc *domain.Document) (*domain.RenderResult, error) {
	ps.logger.Info("Processing document", "document_id", doc.ID, "content_type", doc.ContentType)

	startTime := time.Now()

	// Validate document
	if err := ps.validateDocument(doc); err != nil {
		return nil, fmt.Errorf("document validation failed: %w", err)
	}

	// Check cache first
	cacheKey := ps.generateCacheKey(doc)
	if cached, err := ps.cacheService.Get(cacheKey); err == nil && cached != nil {
		ps.logger.Info("Document found in cache", "document_id", doc.ID)
		if result, ok := cached.(*domain.RenderResult); ok {
			result.CacheHit = true
			return result, nil
		}
	}

	// Parse HTML content
	domTree, err := ps.parseHTML(doc.Content, doc.Options.Security)
	if err != nil {
		return nil, fmt.Errorf("HTML parsing failed: %w", err)
	}

	// Parse CSS (if any)
	stylesheet, err := ps.parseCSS(doc.Content, doc.Options.Security)
	if err != nil {
		return nil, fmt.Errorf("CSS parsing failed: %w", err)
	}

	// Calculate layout
	layoutTree, err := ps.layoutEngine.CalculateLayout(domTree, stylesheet, doc.Options.Layout)
	if err != nil {
		return nil, fmt.Errorf("layout calculation failed: %w", err)
	}

	// Generate output
	outputPath, err := ps.generateOutput(ctx, layoutTree, doc.Options)
	if err != nil {
		return nil, fmt.Errorf("output generation failed: %w", err)
	}

	// Create result
	result := &domain.RenderResult{
		OutputPath: outputPath,
		OutputSize: ps.getFileSize(outputPath),
		PageCount:  ps.calculatePageCount(layoutTree, doc.Options.Page),
		RenderTime: time.Since(startTime),
		CacheHit:   false,
		Warnings:   make([]string, 0),
	}

	// Cache the result
	if doc.Options.Performance.EnableCache {
		_ = ps.cacheService.Set(cacheKey, result, doc.Options.Performance.CacheTTL)
	}

	ps.logger.Info("Document processed successfully",
		"document_id", doc.ID,
		"output_path", outputPath,
		"render_time", result.RenderTime,
		"page_count", result.PageCount)

	return result, nil
}

// ProcessJob processes a print job
func (ps *PrintService) ProcessJob(ctx context.Context, job interface{}) error {
	printJob, ok := job.(*domain.PrintJob)
	if !ok {
		return fmt.Errorf("invalid job type: expected *domain.PrintJob")
	}

	ps.logger.Info("Processing print job", "job_id", printJob.ID)

	// Update job status
	printJob.Status = domain.JobStatusProcessing
	now := time.Now()
	printJob.StartedAt = &now

	// Process the document
	result, err := ps.ProcessDocument(ctx, &printJob.Document)
	if err != nil {
		printJob.Status = domain.JobStatusFailed
		printJob.Error = err.Error()
		return err
	}

	// Update job with results
	printJob.Status = domain.JobStatusCompleted
	printJob.OutputPath = result.OutputPath
	completed := time.Now()
	printJob.CompletedAt = &completed

	return nil
}

// validateDocument validates a document before processing
func (ps *PrintService) validateDocument(doc *domain.Document) error {
	if doc == nil {
		return domain.ErrInvalidDocument
	}

	if doc.Content == "" {
		return domain.NewPrintError(domain.ErrCodeInvalidInput, "document content is empty", domain.ErrInvalidDocument)
	}

	if len(doc.Content) > int(ps.config.MaxFileSize) {
		return domain.NewPrintError(domain.ErrCodeResourceLimit, "document too large", domain.ErrDocumentTooLarge).
			WithDetail("size", len(doc.Content)).
			WithDetail("max_size", ps.config.MaxFileSize)
	}

	return nil
}

// parseHTML parses HTML content
func (ps *PrintService) parseHTML(content string, securityOptions domain.SecurityOptions) (*html.DOMNode, error) {
	return ps.htmlParser.Parse(content, securityOptions)
}

// parseCSS parses CSS content from HTML
func (ps *PrintService) parseCSS(content string, _ domain.SecurityOptions) (*css.Stylesheet, error) {
	// Extract CSS from HTML (simplified - would need proper extraction)
	cssContent := ps.extractCSS(content)
	return ps.cssParser.Parse(cssContent)
}

// extractCSS extracts CSS from HTML content
func (ps *PrintService) extractCSS(_ string) string {
	// Simplified CSS extraction - in a real implementation this would
	// properly extract CSS from <style> tags and external stylesheets
	return ""
}

// generateOutput generates the final output file
func (ps *PrintService) generateOutput(ctx context.Context, layoutTree *domain.LayoutNode, options domain.PrintOptions) (string, error) {
	// Generate unique filename
	filename := fmt.Sprintf("output_%d.%s", time.Now().UnixNano(), options.Output.Format)
	outputPath := ps.storageService.GetPath(filename)

	// For now, create a placeholder file
	// In a real implementation, this would render the layout tree to the specified format
	if err := ps.storageService.WriteFile(outputPath, []byte("PDF content placeholder")); err != nil {
		return "", err
	}

	return outputPath, nil
}

// generateCacheKey generates a cache key for a document
func (ps *PrintService) generateCacheKey(doc *domain.Document) string {
	// Simple cache key generation - in production would use proper hashing
	return fmt.Sprintf("doc_%s_%d", doc.ID, len(doc.Content))
}

// getFileSize gets the size of a file
func (ps *PrintService) getFileSize(_ string) int64 {
	// Simplified - would use actual file stat
	return 1024 // placeholder
}

// calculatePageCount calculates the number of pages
func (ps *PrintService) calculatePageCount(layoutTree *domain.LayoutNode, pageOptions domain.PageOptions) int {
	// Simplified page count calculation
	if layoutTree == nil {
		return 1
	}

	pageHeight := pageOptions.Size.Height
	if pageHeight <= 0 {
		return 1
	}

	totalHeight := layoutTree.Box.Height
	pages := int(totalHeight/pageHeight) + 1

	if pages < 1 {
		pages = 1
	}

	return pages
}
