package golden

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"print-service/internal/core/domain"
)

// EnhancedRigorGenerator provides additional rigorous test scenarios
type EnhancedRigorGenerator struct {
	rand *rand.Rand
}

// NewEnhancedRigorGenerator creates a new enhanced rigor generator
func NewEnhancedRigorGenerator() *EnhancedRigorGenerator {
	return &EnhancedRigorGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateUnicodeTestVariants generates Unicode and internationalization tests
func (g *EnhancedRigorGenerator) GenerateUnicodeTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "unicode_i18n",
		Description: "Unicode and internationalization test variants",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// Unicode test cases
	unicodeTests := []struct {
		name     string
		content  string
		language string
		expected int
	}{
		{
			name:     "Arabic RTL Text",
			content:  "<html dir='rtl'><body><p>مرحبا بالعالم - هذا نص تجريبي باللغة العربية</p></body></html>",
			language: "ar",
			expected: 1,
		},
		{
			name:     "Chinese Characters",
			content:  "<html><body><h1>你好世界</h1><p>这是一个中文测试文档，包含各种中文字符。</p></body></html>",
			language: "zh",
			expected: 1,
		},
		{
			name:     "Emoji and Special Characters",
			content:  "<html><body><h1>🚀 Test Document 📄</h1><p>Special chars: ©®™€£¥§¶†‡•…‰‹›\"\"''</p></body></html>",
			language: "en",
			expected: 1,
		},
		{
			name:     "Mixed Scripts",
			content:  "<html><body><p>English, العربية, 中文, Русский, हिन्दी, 日本語</p></body></html>",
			language: "multi",
			expected: 1,
		},
	}

	for i, test := range unicodeTests {
		suite.TestCases = append(suite.TestCases, TestCase{
			ID:          fmt.Sprintf("unicode_%d", i),
			Name:        test.name,
			Description: fmt.Sprintf("Unicode handling test for %s", test.language),
			Tags:        []string{"unicode", "i18n", test.language},
			Input: TestInput{
				Document: domain.Document{
					ID:          fmt.Sprintf("doc_unicode_%d", i),
					Content:     test.content,
					ContentType: domain.ContentTypeHTML,
					Metadata: domain.DocumentMetadata{
						Title:    test.name,
						Subject:  "Unicode Testing",
						Keywords: []string{"unicode", "i18n", test.language},
					},
					Options:   domain.DefaultPrintOptions(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Options: domain.DefaultPrintOptions(),
			},
			Expected: ExpectedOutput{
				Status:    domain.JobStatusCompleted,
				PageCount: test.expected,
			},
		})
	}

	return suite
}

// GeneratePropertyBasedTestVariants generates property-based test cases
func (g *EnhancedRigorGenerator) GeneratePropertyBasedTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "property_based",
		Description: "Property-based test variants with random valid inputs",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// Generate 50 random but valid HTML documents
	for i := 0; i < 50; i++ {
		content := g.generateRandomValidHTML()

		suite.TestCases = append(suite.TestCases, TestCase{
			ID:          fmt.Sprintf("property_html_%d", i),
			Name:        fmt.Sprintf("Property-based HTML Test %d", i),
			Description: "Randomly generated valid HTML document",
			Tags:        []string{"property", "random", "html"},
			Input: TestInput{
				Document: domain.Document{
					ID:          fmt.Sprintf("doc_property_%d", i),
					Content:     content,
					ContentType: domain.ContentTypeHTML,
					Metadata: domain.DocumentMetadata{
						Title:   fmt.Sprintf("Property Test %d", i),
						Subject: "Property-based Testing",
					},
					Options:   domain.DefaultPrintOptions(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Options: g.generateRandomValidPrintOptions(),
			},
			Expected: ExpectedOutput{
				Status: domain.JobStatusCompleted,
				// Don't specify exact page count for property-based tests
			},
		})
	}

	return suite
}

// GenerateMemoryStressTestVariants generates memory and resource stress tests
func (g *EnhancedRigorGenerator) GenerateMemoryStressTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "memory_stress",
		Description: "Memory and resource stress test variants",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// Extremely large document
	suite.TestCases = append(suite.TestCases, TestCase{
		ID:          "memory_large_document",
		Name:        "Extremely Large Document",
		Description: "Document with 10,000 paragraphs to test memory handling",
		Tags:        []string{"memory", "stress", "large"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_memory_large",
				Content:     g.generateExtremelyLargeHTML(),
				ContentType: domain.ContentTypeHTML,
				Metadata: domain.DocumentMetadata{
					Title:   "Memory Stress Test",
					Subject: "Large Document Testing",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: func() domain.PrintOptions {
				opts := domain.DefaultPrintOptions()
				opts.Performance.MaxMemory = 1024 * 1024 * 1024 // 1GB limit
				return opts
			}(),
		},
		Expected: ExpectedOutput{
			Status:   domain.JobStatusCompleted,
			Warnings: []string{"Large document may impact performance"},
		},
	})

	// Deep nesting test
	suite.TestCases = append(suite.TestCases, TestCase{
		ID:          "memory_deep_nesting",
		Name:        "Deeply Nested HTML",
		Description: "HTML with 1000 levels of nesting to test stack limits",
		Tags:        []string{"memory", "nesting", "stack"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_memory_nesting",
				Content:     g.generateDeeplyNestedHTML(1000),
				ContentType: domain.ContentTypeHTML,
				Metadata: domain.DocumentMetadata{
					Title:   "Deep Nesting Test",
					Subject: "Stack Limit Testing",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status: domain.JobStatusCompleted,
		},
	})

	return suite
}

// GenerateRegressionTestVariants generates regression detection tests
func (g *EnhancedRigorGenerator) GenerateRegressionTestVariants() TestSuite {
	suite := TestSuite{
		Name:        "regression",
		Description: "Regression detection test variants with baseline comparisons",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		TestCases:   []TestCase{},
	}

	// Baseline performance tests
	baselineTests := []struct {
		name          string
		content       string
		maxRenderTime time.Duration
		maxMemory     int64
	}{
		{
			name:          "Baseline Simple Document",
			content:       "<html><body><h1>Baseline Test</h1><p>This is a baseline performance test.</p></body></html>",
			maxRenderTime: 100 * time.Millisecond,
			maxMemory:     10 * 1024 * 1024, // 10MB
		},
		{
			name:          "Baseline Complex Table",
			content:       g.generateComplexTableHTML(100), // 100 rows
			maxRenderTime: 500 * time.Millisecond,
			maxMemory:     50 * 1024 * 1024, // 50MB
		},
	}

	for i, test := range baselineTests {
		suite.TestCases = append(suite.TestCases, TestCase{
			ID:          fmt.Sprintf("regression_%d", i),
			Name:        test.name,
			Description: "Regression test to detect performance degradation",
			Tags:        []string{"regression", "performance", "baseline"},
			Input: TestInput{
				Document: domain.Document{
					ID:          fmt.Sprintf("doc_regression_%d", i),
					Content:     test.content,
					ContentType: domain.ContentTypeHTML,
					Metadata: domain.DocumentMetadata{
						Title:   test.name,
						Subject: "Regression Testing",
					},
					Options:   domain.DefaultPrintOptions(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Options: domain.DefaultPrintOptions(),
			},
			Expected: ExpectedOutput{
				Status:     domain.JobStatusCompleted,
				RenderTime: test.maxRenderTime,
			},
			Metadata: map[string]interface{}{
				"baseline":    true,
				"max_memory":  test.maxMemory,
				"performance": "critical",
			},
		})
	}

	return suite
}

// Helper functions for enhanced rigor

func (g *EnhancedRigorGenerator) generateRandomValidHTML() string {
	elements := []string{"h1", "h2", "h3", "p", "div", "span", "ul", "ol", "li", "strong", "em"}
	content := "<html><head><title>Random Test</title></head><body>"

	numElements := g.rand.Intn(20) + 5
	for i := 0; i < numElements; i++ {
		element := elements[g.rand.Intn(len(elements))]
		text := g.generateRandomText(g.rand.Intn(100) + 10)

		if element == "ul" || element == "ol" {
			content += fmt.Sprintf("<%s>", element)
			for j := 0; j < g.rand.Intn(5)+2; j++ {
				content += fmt.Sprintf("<li>%s</li>", g.generateRandomText(20))
			}
			content += fmt.Sprintf("</%s>", element)
		} else if element == "li" {
			continue // Skip standalone li elements
		} else {
			content += fmt.Sprintf("<%s>%s</%s>", element, text, element)
		}
	}

	content += "</body></html>"
	return content
}

func (g *EnhancedRigorGenerator) generateRandomText(length int) string {
	words := []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit", "sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua"}

	var result []string
	for i := 0; i < length/5; i++ { // Approximate word count
		result = append(result, words[g.rand.Intn(len(words))])
	}

	return strings.Join(result, " ")
}

func (g *EnhancedRigorGenerator) generateRandomValidPrintOptions() domain.PrintOptions {
	opts := domain.DefaultPrintOptions()

	// Randomize within valid ranges
	qualities := []domain.RenderQuality{domain.QualityDraft, domain.QualityNormal, domain.QualityHigh}
	opts.Render.Quality = qualities[g.rand.Intn(len(qualities))]

	formats := []domain.OutputFormat{domain.FormatPDF, domain.FormatPNG, domain.FormatJPEG}
	opts.Output.Format = formats[g.rand.Intn(len(formats))]

	opts.Page.Scale = 0.5 + g.rand.Float64()*1.5 // 0.5 to 2.0
	opts.Layout.DPI = 72 + g.rand.Intn(228)      // 72 to 300 DPI

	return opts
}

func (g *EnhancedRigorGenerator) generateExtremelyLargeHTML() string {
	content := "<html><body><h1>Extremely Large Document</h1>"

	// Generate 10,000 paragraphs
	for i := 0; i < 10000; i++ {
		content += fmt.Sprintf("<p>Paragraph %d: %s</p>", i, g.generateRandomText(50))

		// Add some variety
		if i%100 == 0 {
			content += fmt.Sprintf("<h2>Section %d</h2>", i/100)
		}
		if i%50 == 0 {
			content += "<hr>"
		}
	}

	content += "</body></html>"
	return content
}

func (g *EnhancedRigorGenerator) generateDeeplyNestedHTML(depth int) string {
	content := "<html><body>"

	// Create deep nesting
	for i := 0; i < depth; i++ {
		content += "<div>"
	}

	content += "<p>Deeply nested content</p>"

	// Close all divs
	for i := 0; i < depth; i++ {
		content += "</div>"
	}

	content += "</body></html>"
	return content
}

func (g *EnhancedRigorGenerator) generateComplexTableHTML(rows int) string {
	content := `<html><body><h1>Complex Table</h1>
	<table border="1" style="border-collapse: collapse; width: 100%;">
	<thead>
		<tr>
			<th>ID</th><th>Name</th><th>Email</th><th>Department</th>
			<th>Salary</th><th>Start Date</th><th>Status</th><th>Notes</th>
		</tr>
	</thead>
	<tbody>`

	for i := 1; i <= rows; i++ {
		content += fmt.Sprintf(`
		<tr>
			<td>%d</td>
			<td>Employee %d</td>
			<td>emp%d@company.com</td>
			<td>Dept %d</td>
			<td>$%d,000</td>
			<td>2023-01-%02d</td>
			<td>Active</td>
			<td>%s</td>
		</tr>`, i, i, i, (i%10)+1, 40+(i%60), (i%28)+1, g.generateRandomText(20))
	}

	content += "</tbody></table></body></html>"
	return content
}

// GenerateAllEnhancedVariants generates all enhanced rigor test variants
func (g *EnhancedRigorGenerator) GenerateAllEnhancedVariants() []TestSuite {
	return []TestSuite{
		g.GenerateUnicodeTestVariants(),
		g.GeneratePropertyBasedTestVariants(),
		g.GenerateMemoryStressTestVariants(),
		g.GenerateRegressionTestVariants(),
	}
}
