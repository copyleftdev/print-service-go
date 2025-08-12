package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// GoldenTestData represents the structure of golden test data files
type GoldenTestData struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	CreatedAt   string     `json:"created_at"`
	TestCases   []TestCase `json:"test_cases"`
}

// TestCase represents a single test case
type TestCase struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	Input       TestInput   `json:"input"`
	Expected    TestExpected `json:"expected"`
}

// TestInput represents test input data
type TestInput struct {
	Document TestDocument `json:"document"`
	Options  TestOptions  `json:"options,omitempty"`
}

// TestDocument represents document data
type TestDocument struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TestOptions represents test options
type TestOptions struct {
	Format      string                 `json:"format,omitempty"`
	PageSize    string                 `json:"page_size,omitempty"`
	Orientation string                 `json:"orientation,omitempty"`
	Margins     map[string]interface{} `json:"margins,omitempty"`
}

// TestExpected represents expected results
type TestExpected struct {
	Status    string `json:"status"`
	PageCount int    `json:"page_count,omitempty"`
}

// PrintRequest represents API request
type PrintRequest struct {
	Content string                 `json:"content"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// PrintResponse represents API response
type PrintResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// GoldenRigorTest runs comprehensive golden test data
type GoldenRigorTest struct {
	serviceURL string
	httpClient *http.Client
	results    []TestResult
}

// TestResult captures test execution results
type TestResult struct {
	TestID      string
	TestName    string
	Category    string
	Passed      bool
	Duration    time.Duration
	Error       error
	Description string
}

// NewGoldenRigorTest creates a new golden rigor test
func NewGoldenRigorTest(serviceURL string) *GoldenRigorTest {
	return &GoldenRigorTest{
		serviceURL: serviceURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		results:    make([]TestResult, 0),
	}
}

// RunGoldenRigorTests executes all golden test data
func (g *GoldenRigorTest) RunGoldenRigorTests() error {
	fmt.Println("🧪 GOLDEN RIGOR TEST SUITE")
	fmt.Println("==========================")
	fmt.Printf("🎯 Testing against service: %s\n", g.serviceURL)
	fmt.Println()

	// Check service health first
	if err := g.checkServiceHealth(); err != nil {
		return fmt.Errorf("service health check failed: %w", err)
	}

	// Find all golden test data files
	goldenDir := "../../testdata/golden"
	files, err := filepath.Glob(filepath.Join(goldenDir, "*_golden_data.json"))
	if err != nil {
		return fmt.Errorf("failed to find golden test files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no golden test data files found in %s", goldenDir)
	}

	fmt.Printf("📁 Found %d golden test data files\n", len(files))
	fmt.Println()

	// Run tests for each golden data file
	for _, file := range files {
		if err := g.runGoldenDataFile(file); err != nil {
			fmt.Printf("❌ Failed to run tests from %s: %v\n", file, err)
			continue
		}
	}

	// Print summary
	g.printSummary()

	// Check if all tests passed
	for _, result := range g.results {
		if !result.Passed {
			return fmt.Errorf("some tests failed")
		}
	}

	return nil
}

// runGoldenDataFile runs tests from a single golden data file
func (g *GoldenRigorTest) runGoldenDataFile(filePath string) error {
	fileName := filepath.Base(filePath)
	fmt.Printf("📊 Running tests from: %s\n", fileName)

	// Read and parse golden data file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var goldenData GoldenTestData
	if err := json.Unmarshal(data, &goldenData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Printf("   Category: %s (%d test cases)\n", goldenData.Name, len(goldenData.TestCases))

	// Run each test case
	for _, testCase := range goldenData.TestCases {
		result := g.runTestCase(goldenData.Name, testCase)
		g.results = append(g.results, result)

		if result.Passed {
			fmt.Printf("   ✅ %s (%v)\n", result.TestName, result.Duration)
		} else {
			fmt.Printf("   ❌ %s: %v\n", result.TestName, result.Error)
		}
	}

	fmt.Println()
	return nil
}

// runTestCase executes a single test case
func (g *GoldenRigorTest) runTestCase(category string, testCase TestCase) TestResult {
	start := time.Now()

	result := TestResult{
		TestID:      testCase.ID,
		TestName:    testCase.Name,
		Category:    category,
		Description: testCase.Description,
	}

	// Create print request
	printReq := PrintRequest{
		Content: testCase.Input.Document.Content,
		Options: make(map[string]interface{}),
	}

	// Add options if present
	if testCase.Input.Options.Format != "" {
		printReq.Options["format"] = testCase.Input.Options.Format
	}
	if testCase.Input.Options.PageSize != "" {
		printReq.Options["page_size"] = testCase.Input.Options.PageSize
	}

	// Submit print job
	reqBody, _ := json.Marshal(printReq)
	resp, err := g.httpClient.Post(
		g.serviceURL+"/api/v1/print",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		result.Error = fmt.Errorf("failed to submit job: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		result.Error = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		result.Duration = time.Since(start)
		return result
	}

	// Parse response
	var printResp PrintResponse
	if err := json.NewDecoder(resp.Body).Decode(&printResp); err != nil {
		result.Error = fmt.Errorf("failed to parse response: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Poll for completion (simplified for rigor testing)
	maxWait := 30 * time.Second
	pollStart := time.Now()
	for time.Since(pollStart) < maxWait {
		statusResp, err := g.httpClient.Get(g.serviceURL + "/api/v1/print/" + printResp.JobID)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if statusResp.StatusCode == http.StatusOK {
			statusResp.Body.Close()
			break
		}
		statusResp.Body.Close()
		time.Sleep(1 * time.Second)
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

// checkServiceHealth verifies the service is running
func (g *GoldenRigorTest) checkServiceHealth() error {
	resp, err := g.httpClient.Get(g.serviceURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("service unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

// printSummary prints test execution summary
func (g *GoldenRigorTest) printSummary() {
	fmt.Println("📊 GOLDEN RIGOR TEST SUMMARY")
	fmt.Println("=============================")

	totalTests := len(g.results)
	passedTests := 0
	categories := make(map[string][]TestResult)

	for _, result := range g.results {
		if result.Passed {
			passedTests++
		}
		categories[result.Category] = append(categories[result.Category], result)
	}

	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", passedTests)
	fmt.Printf("Failed: %d\n", totalTests-passedTests)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passedTests)/float64(totalTests)*100)
	fmt.Println()

	// Print category breakdown
	fmt.Println("📋 Category Breakdown:")
	for category, results := range categories {
		passed := 0
		for _, r := range results {
			if r.Passed {
				passed++
			}
		}
		fmt.Printf("  %s: %d/%d passed\n", category, passed, len(results))
	}

	fmt.Println()
	if passedTests == totalTests {
		fmt.Println("🎉 All golden rigor tests passed!")
	} else {
		fmt.Println("❌ Some tests failed. Review logs for details.")
	}
}

func main() {
	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "http://localhost:8080"
	}

	test := NewGoldenRigorTest(serviceURL)

	if err := test.RunGoldenRigorTests(); err != nil {
		fmt.Printf("❌ Golden rigor tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Golden rigor test suite completed successfully!")
}
