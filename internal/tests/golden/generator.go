package golden

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"print-service/internal/core/domain"
)

// TestDataGenerator generates golden test data variants
type TestDataGenerator struct {
	rand *rand.Rand
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	TestCases   []TestCase `json:"test_cases"`
}

// TestCase represents a single test case with input and expected output
type TestCase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Input       TestInput              `json:"input"`
	Expected    ExpectedOutput         `json:"expected"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TestInput represents the input for a test case
type TestInput struct {
	Document domain.Document    `json:"document"`
	Options  domain.PrintOptions `json:"options"`
}

// ExpectedOutput represents the expected output for a test case
type ExpectedOutput struct {
	Status      domain.JobStatus       `json:"status"`
	PageCount   int                    `json:"page_count,omitempty"`
	OutputSize  int64                  `json:"output_size,omitempty"`
	RenderTime  time.Duration          `json:"render_time,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Errors      []string               `json:"errors,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateBasicVariants generates basic test variants for common use cases
func (g *TestDataGenerator) GenerateBasicVariants() TestSuite {
	suite := TestSuite{
		Name:        "basic",
		Description: "Basic test variants for common print scenarios",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// HTML Document variants
	suite.TestCases = append(suite.TestCases, g.createHTMLTestCases()...)
	
	// Markdown Document variants
	suite.TestCases = append(suite.TestCases, g.createMarkdownTestCases()...)
	
	// Text Document variants
	suite.TestCases = append(suite.TestCases, g.createTextTestCases()...)

	return suite
}

// GenerateEdgeCaseVariants generates edge case test variants
func (g *TestDataGenerator) GenerateEdgeCaseVariants() TestSuite {
	suite := TestSuite{
		Name:        "edge_cases",
		Description: "Edge case test variants for boundary conditions",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// Empty content
	suite.TestCases = append(suite.TestCases, TestCase{
		ID:          "edge_empty_content",
		Name:        "Empty Content",
		Description: "Test handling of empty document content",
		Tags:        []string{"edge", "empty", "boundary"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_empty",
				Content:     "",
				ContentType: domain.ContentTypeHTML,
				Metadata:    domain.DocumentMetadata{Title: "Empty Document"},
				Options:     domain.DefaultPrintOptions(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
			Warnings:  []string{"Document content is empty"},
		},
	})

	// Very large content
	largeContent := strings.Repeat("<p>This is a very long paragraph that will be repeated many times to create a large document for testing memory and performance limits.</p>", 1000)
	suite.TestCases = append(suite.TestCases, TestCase{
		ID:          "edge_large_content",
		Name:        "Large Content Document",
		Description: "Test handling of very large document content",
		Tags:        []string{"edge", "large", "performance"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_large",
				Content:     largeContent,
				ContentType: domain.ContentTypeHTML,
				Metadata:    domain.DocumentMetadata{Title: "Large Document"},
				Options:     domain.DefaultPrintOptions(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 50, // Estimated
		},
	})

	// Invalid HTML
	suite.TestCases = append(suite.TestCases, TestCase{
		ID:          "edge_invalid_html",
		Name:        "Invalid HTML Content",
		Description: "Test handling of malformed HTML content",
		Tags:        []string{"edge", "invalid", "html"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_invalid_html",
				Content:     "<html><body><p>Unclosed paragraph<div>Nested without closing</body>",
				ContentType: domain.ContentTypeHTML,
				Metadata:    domain.DocumentMetadata{Title: "Invalid HTML"},
				Options:     domain.DefaultPrintOptions(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:   domain.JobStatusCompleted,
			Warnings: []string{"HTML validation warnings detected"},
		},
	})

	return suite
}

// GenerateStressTestVariants generates stress test variants
func (g *TestDataGenerator) GenerateStressTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "stress_tests",
		Description: "Stress test variants for high load scenarios",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// High concurrent jobs
	for i := 0; i < 100; i++ {
		suite.TestCases = append(suite.TestCases, TestCase{
			ID:          fmt.Sprintf("stress_concurrent_%d", i),
			Name:        fmt.Sprintf("Concurrent Job %d", i),
			Description: "Stress test with multiple concurrent print jobs",
			Tags:        []string{"stress", "concurrent", "performance"},
			Input: TestInput{
				Document: domain.Document{
					ID:          fmt.Sprintf("doc_stress_%d", i),
					Content:     g.generateRandomHTML(),
					ContentType: domain.ContentTypeHTML,
					Metadata:    domain.DocumentMetadata{Title: fmt.Sprintf("Stress Test Document %d", i)},
					Options:     domain.DefaultPrintOptions(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Options: g.generateRandomPrintOptions(),
			},
			Expected: ExpectedOutput{
				Status: domain.JobStatusCompleted,
			},
		})
	}

	return suite
}

// GenerateSecurityTestVariants generates security test variants
func (g *TestDataGenerator) GenerateSecurityTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "security_tests",
		Description: "Security test variants for vulnerability testing",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// XSS attempts
	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
		"<svg onload=alert('xss')>",
	}

	for i, payload := range xssPayloads {
		suite.TestCases = append(suite.TestCases, TestCase{
			ID:          fmt.Sprintf("security_xss_%d", i),
			Name:        fmt.Sprintf("XSS Test %d", i),
			Description: "Test XSS payload handling",
			Tags:        []string{"security", "xss", "sanitization"},
			Input: TestInput{
				Document: domain.Document{
					ID:          fmt.Sprintf("doc_xss_%d", i),
					Content:     fmt.Sprintf("<html><body>%s</body></html>", payload),
					ContentType: domain.ContentTypeHTML,
					Metadata:    domain.DocumentMetadata{Title: "XSS Test"},
					Options:     domain.DefaultPrintOptions(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Options: func() domain.PrintOptions {
					opts := domain.DefaultPrintOptions()
					opts.Security.SanitizeHTML = true
					return opts
				}(),
			},
			Expected: ExpectedOutput{
				Status:   domain.JobStatusCompleted,
				Warnings: []string{"Potentially malicious content sanitized"},
			},
		})
	}

	return suite
}

// GeneratePerformanceTestVariants generates performance test variants
func (g *TestDataGenerator) GeneratePerformanceTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "performance_tests",
		Description: "Performance test variants for benchmarking",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// Different quality levels
	qualities := []domain.RenderQuality{
		domain.QualityDraft,
		domain.QualityNormal,
		domain.QualityHigh,
	}

	for i, quality := range qualities {
		suite.TestCases = append(suite.TestCases, TestCase{
			ID:          fmt.Sprintf("perf_quality_%s", string(quality)),
			Name:        fmt.Sprintf("Performance Test - %s Quality", strings.Title(string(quality))),
			Description: fmt.Sprintf("Performance test with %s quality rendering", quality),
			Tags:        []string{"performance", "quality", string(quality)},
			Input: TestInput{
				Document: domain.Document{
					ID:          fmt.Sprintf("doc_perf_%d", i),
					Content:     g.generateComplexHTML(),
					ContentType: domain.ContentTypeHTML,
					Metadata:    domain.DocumentMetadata{Title: "Performance Test Document"},
					Options:     domain.DefaultPrintOptions(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Options: func() domain.PrintOptions {
					opts := domain.DefaultPrintOptions()
					opts.Render.Quality = quality
					return opts
				}(),
			},
			Expected: ExpectedOutput{
				Status: domain.JobStatusCompleted,
				Metadata: map[string]interface{}{
					"benchmark": true,
					"quality":   string(quality),
				},
			},
		})
	}

	return suite
}
