package golden

import (
	"fmt"
	"time"

	"print-service/internal/core/domain"
)

// createHTMLTestCases generates HTML document test cases
func (g *TestDataGenerator) createHTMLTestCases() []TestCase {
	var testCases []TestCase

	// Simple HTML document
	testCases = append(testCases, TestCase{
		ID:          "html_simple",
		Name:        "Simple HTML Document",
		Description: "Basic HTML document with standard elements",
		Tags:        []string{"html", "basic", "standard"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_html_simple",
				Content:     "<html><head><title>Test Document</title></head><body><h1>Hello World</h1><p>This is a test document.</p></body></html>",
				ContentType: domain.ContentTypeHTML,
				Metadata: domain.DocumentMetadata{
					Title:    "Simple HTML Test",
					Author:   "Test Generator",
					Subject:  "Basic HTML Testing",
					Keywords: []string{"html", "test", "basic"},
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
		},
	})

	// HTML with CSS styling
	testCases = append(testCases, TestCase{
		ID:          "html_with_css",
		Name:        "HTML Document with CSS",
		Description: "HTML document with embedded CSS styling",
		Tags:        []string{"html", "css", "styling"},
		Input: TestInput{
			Document: domain.Document{
				ID:      "doc_html_css",
				Content: `<html><head><style>body{font-family:Arial;margin:20px;}h1{color:blue;}</style></head><body><h1>Styled Document</h1><p>This document has CSS styling.</p></body></html>`,
				ContentType: domain.ContentTypeHTML,
				Metadata: domain.DocumentMetadata{
					Title:   "HTML with CSS",
					Author:  "Test Generator",
					Subject: "CSS Styling Test",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
		},
	})

	// HTML with tables
	testCases = append(testCases, TestCase{
		ID:          "html_table",
		Name:        "HTML Document with Table",
		Description: "HTML document containing a data table",
		Tags:        []string{"html", "table", "data"},
		Input: TestInput{
			Document: domain.Document{
				ID: "doc_html_table",
				Content: `<html><body><h1>Data Table</h1><table border="1"><tr><th>Name</th><th>Age</th><th>City</th></tr><tr><td>John</td><td>30</td><td>New York</td></tr><tr><td>Jane</td><td>25</td><td>London</td></tr></table></body></html>`,
				ContentType: domain.ContentTypeHTML,
				Metadata: domain.DocumentMetadata{
					Title:   "Table Test",
					Subject: "HTML Table Testing",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
		},
	})

	return testCases
}

// createMarkdownTestCases generates Markdown document test cases
func (g *TestDataGenerator) createMarkdownTestCases() []TestCase {
	var testCases []TestCase

	// Simple Markdown
	testCases = append(testCases, TestCase{
		ID:          "markdown_simple",
		Name:        "Simple Markdown Document",
		Description: "Basic Markdown document with common elements",
		Tags:        []string{"markdown", "basic"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_md_simple",
				Content:     "# Hello World\n\nThis is a **bold** text and this is *italic*.\n\n- List item 1\n- List item 2\n- List item 3",
				ContentType: domain.ContentTypeMarkdown,
				Metadata: domain.DocumentMetadata{
					Title:   "Simple Markdown",
					Author:  "Test Generator",
					Subject: "Markdown Testing",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
		},
	})

	// Markdown with code blocks
	testCases = append(testCases, TestCase{
		ID:          "markdown_code",
		Name:        "Markdown with Code Blocks",
		Description: "Markdown document containing code blocks",
		Tags:        []string{"markdown", "code", "syntax"},
		Input: TestInput{
			Document: domain.Document{
				ID: "doc_md_code",
				Content: "# Code Example\n\nHere's some Go code:\n\n```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n\nAnd some inline `code` too.",
				ContentType: domain.ContentTypeMarkdown,
				Metadata: domain.DocumentMetadata{
					Title:   "Markdown Code",
					Subject: "Code Block Testing",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
		},
	})

	return testCases
}

// createTextTestCases generates plain text document test cases
func (g *TestDataGenerator) createTextTestCases() []TestCase {
	var testCases []TestCase

	// Simple text
	testCases = append(testCases, TestCase{
		ID:          "text_simple",
		Name:        "Simple Text Document",
		Description: "Basic plain text document",
		Tags:        []string{"text", "plain", "basic"},
		Input: TestInput{
			Document: domain.Document{
				ID:          "doc_text_simple",
				Content:     "This is a simple plain text document.\n\nIt contains multiple paragraphs and should be rendered as-is without any special formatting.",
				ContentType: domain.ContentTypeText,
				Metadata: domain.DocumentMetadata{
					Title:   "Plain Text",
					Subject: "Text Document Testing",
				},
				Options:   domain.DefaultPrintOptions(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Options: domain.DefaultPrintOptions(),
		},
		Expected: ExpectedOutput{
			Status:    domain.JobStatusCompleted,
			PageCount: 1,
		},
	})

	return testCases
}

// generateRandomHTML creates random HTML content for stress testing
func (g *TestDataGenerator) generateRandomHTML() string {
	elements := []string{
		"<h1>Random Heading %d</h1>",
		"<p>Random paragraph with some text content %d.</p>",
		"<div>Random div element %d</div>",
		"<span>Random span %d</span>",
		"<ul><li>List item %d</li><li>Another item</li></ul>",
	}

	content := "<html><body>"
	for i := 0; i < g.rand.Intn(20)+5; i++ {
		element := elements[g.rand.Intn(len(elements))]
		content += fmt.Sprintf(element, i)
	}
	content += "</body></html>"

	return content
}

// generateComplexHTML creates complex HTML for performance testing
func (g *TestDataGenerator) generateComplexHTML() string {
	return `<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; }
        .content { margin: 20px 0; }
        .table { border-collapse: collapse; width: 100%; }
        .table th, .table td { border: 1px solid #ddd; padding: 8px; }
        .footer { margin-top: 50px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Complex Document Performance Test</h1>
        <p>This document contains various complex elements for performance testing.</p>
    </div>
    
    <div class="content">
        <h2>Data Table</h2>
        <table class="table">
            <thead>
                <tr><th>ID</th><th>Name</th><th>Email</th><th>Department</th><th>Salary</th></tr>
            </thead>
            <tbody>` +
		func() string {
			rows := ""
			for i := 1; i <= 100; i++ {
				rows += fmt.Sprintf(`
                <tr>
                    <td>%d</td>
                    <td>Employee %d</td>
                    <td>emp%d@company.com</td>
                    <td>Department %d</td>
                    <td>$%d,000</td>
                </tr>`, i, i, i, (i%5)+1, 50+(i%50))
			}
			return rows
		}() + `
            </tbody>
        </table>
        
        <h2>Long Text Content</h2>
        <p>` + func() string {
		text := ""
		for i := 0; i < 50; i++ {
			text += "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "
		}
		return text
	}() + `</p>
    </div>
    
    <div class="footer">
        <p>Generated by Print Service Test Data Generator</p>
    </div>
</body>
</html>`
}

// generateRandomPrintOptions creates random print options for testing
func (g *TestDataGenerator) generateRandomPrintOptions() domain.PrintOptions {
	opts := domain.DefaultPrintOptions()

	// Randomize some options
	qualities := []domain.RenderQuality{domain.QualityDraft, domain.QualityNormal, domain.QualityHigh}
	opts.Render.Quality = qualities[g.rand.Intn(len(qualities))]

	formats := []domain.OutputFormat{domain.FormatPDF, domain.FormatPNG, domain.FormatJPEG}
	opts.Output.Format = formats[g.rand.Intn(len(formats))]

	opts.Page.Scale = 0.5 + g.rand.Float64()
	opts.Performance.ConcurrentJobs = g.rand.Intn(10) + 1

	return opts
}
