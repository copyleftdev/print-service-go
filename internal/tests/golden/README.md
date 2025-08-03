# Golden Test Data Generator

A comprehensive tool for generating golden test data variants for rigorous testing of the print service.

## Overview

The Golden Test Data Generator creates standardized test datasets that cover:

- **Basic Variants**: Common use cases and standard scenarios
- **Edge Cases**: Boundary conditions and unusual inputs
- **Stress Tests**: High-load and concurrent processing scenarios
- **Security Tests**: XSS, injection, and sanitization testing
- **Performance Tests**: Benchmarking different quality levels and configurations

## Features

### Test Data Types

1. **HTML Documents**
   - Simple HTML with basic elements
   - HTML with CSS styling
   - Complex HTML with tables and forms
   - Invalid/malformed HTML

2. **Markdown Documents**
   - Basic Markdown formatting
   - Code blocks and syntax highlighting
   - Complex nested structures

3. **Plain Text Documents**
   - Simple text content
   - Large text files
   - Special characters and encoding

### Test Variants

#### Basic Variants
- Standard document types
- Common print options
- Typical use cases

#### Edge Cases
- Empty content
- Very large documents (1000+ paragraphs)
- Invalid HTML/Markdown
- Extreme option values

#### Stress Tests
- 100 concurrent document processing
- Random content generation
- Variable print options

#### Security Tests
- XSS payload injection
- HTML sanitization testing
- Content validation

#### Performance Tests
- Different quality levels (draft, normal, high)
- Various output formats
- Compression and optimization testing

## Usage

### Command Line Tool

Generate test data using the command line tool:

```bash
# Generate all test variants
go run cmd/testgen/main.go

# Generate specific variants
go run cmd/testgen/main.go -variants=basic
go run cmd/testgen/main.go -variants=edge
go run cmd/testgen/main.go -variants=stress

# Specify output directory
go run cmd/testgen/main.go -output=./custom/testdata

# Enable verbose output
go run cmd/testgen/main.go -verbose
```

### Programmatic Usage

```go
package main

import (
    "github.com/copyleftdev/print-service-go/internal/tests/golden"
)

func main() {
    generator := golden.NewTestDataGenerator()
    
    // Generate basic test variants
    basicSuite := generator.GenerateBasicVariants()
    
    // Generate edge case variants
    edgeSuite := generator.GenerateEdgeCaseVariants()
    
    // Generate stress test variants
    stressSuite := generator.GenerateStressTestVariants()
}
```

### Test Runner

Execute golden tests against your print service:

```go
// Implement the PrintService interface
type MyPrintService struct {
    // Your implementation
}

func (s *MyPrintService) ProcessDocument(ctx context.Context, doc domain.Document, opts domain.PrintOptions) (*domain.RenderResult, error) {
    // Your implementation
    return nil, nil
}

func (s *MyPrintService) GetJobStatus(ctx context.Context, jobID string) (*domain.PrintJob, error) {
    // Your implementation
    return nil, nil
}

// Run tests
func runGoldenTests() {
    service := &MyPrintService{}
    runner := golden.NewTestRunner(service)
    
    // Load test suite
    suite, err := golden.LoadTestSuite("./testdata/golden/basic_golden_data.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Run tests
    ctx := context.Background()
    result, err := runner.RunTestSuite(ctx, *suite)
    if err != nil {
        log.Fatal(err)
    }
    
    // Save results
    err = golden.SaveTestResults(result, "./results/basic_test_results.json")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.GetSummary())
}
```

## Generated Test Data Structure

Each test suite contains:

```json
{
  "name": "basic",
  "description": "Basic test variants for common print scenarios",
  "version": "1.0.0",
  "created_at": "2024-01-01T00:00:00Z",
  "test_cases": [
    {
      "id": "html_simple",
      "name": "Simple HTML Document",
      "description": "Basic HTML document with standard elements",
      "tags": ["html", "basic", "standard"],
      "input": {
        "document": {
          "id": "doc_html_simple",
          "content": "<html>...</html>",
          "content_type": "html",
          "metadata": {...},
          "options": {...}
        },
        "options": {...}
      },
      "expected": {
        "status": "completed",
        "page_count": 1,
        "warnings": []
      }
    }
  ]
}
```

## Validation and Tolerances

The test runner includes configurable validation with tolerances:

- **Page Count**: ±1 page variance
- **Output Size**: ±10% variance
- **Render Time**: ±500ms variance
- **Warnings**: Exact match required

### Custom Tolerances

```go
validator := golden.NewResultValidator()
validator.SetTolerances(golden.Tolerances{
    PageCountVariance:  2,                // ±2 pages
    OutputSizeVariance: 0.15,             // ±15%
    RenderTimeVariance: 1 * time.Second,  // ±1 second
})
```

## Integration with CI/CD

Add to your GitHub Actions workflow:

```yaml
- name: Generate Golden Test Data
  run: |
    go run cmd/testgen/main.go -output=./testdata/golden -verbose

- name: Run Golden Tests
  run: |
    go test ./internal/tests/golden/... -v
```

## File Structure

```
internal/tests/golden/
├── README.md           # This documentation
├── generator.go        # Main test data generator
├── helpers.go          # Helper functions for test case creation
├── runner.go           # Test execution and validation
├── validator.go        # Result validation logic
└── testdata/           # Generated test data files
    ├── basic_golden_data.json
    ├── edge_cases_golden_data.json
    ├── stress_tests_golden_data.json
    ├── security_tests_golden_data.json
    └── performance_tests_golden_data.json
```

## Best Practices

1. **Version Control**: Include generated test data in version control for consistency
2. **Regular Updates**: Regenerate test data when domain models change
3. **Baseline Testing**: Use golden tests to detect regressions
4. **Performance Monitoring**: Track performance metrics over time
5. **Security Validation**: Regularly update security test payloads

## Contributing

When adding new test variants:

1. Add new test case generators in `helpers.go`
2. Update the main generator functions in `generator.go`
3. Add corresponding validation logic in `validator.go`
4. Update this documentation

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure all dependencies are properly imported
2. **File Permissions**: Check write permissions for output directory
3. **Memory Issues**: Large stress tests may require increased memory limits
4. **Timeout Issues**: Adjust context timeouts for long-running tests

### Debug Mode

Enable verbose logging:

```bash
go run cmd/testgen/main.go -verbose
```

This will provide detailed information about test generation and execution.
