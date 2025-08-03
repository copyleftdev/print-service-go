package golden

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"print-service/internal/core/domain"
)

// TestRunner executes golden test cases and validates results
type TestRunner struct {
	printService PrintService // Interface to the actual print service
	validator    *ResultValidator
}

// PrintService interface for testing (to be implemented by actual service)
type PrintService interface {
	ProcessDocument(ctx context.Context, doc domain.Document, opts domain.PrintOptions) (*domain.RenderResult, error)
	GetJobStatus(ctx context.Context, jobID string) (*domain.PrintJob, error)
}

// NewTestRunner creates a new test runner
func NewTestRunner(printService PrintService) *TestRunner {
	return &TestRunner{
		printService: printService,
		validator:    NewResultValidator(),
	}
}

// RunTestSuite executes all test cases in a test suite
func (r *TestRunner) RunTestSuite(ctx context.Context, suite TestSuite) (*TestSuiteResult, error) {
	result := &TestSuiteResult{
		SuiteName:   suite.Name,
		StartTime:   time.Now(),
		TestResults: make([]TestCaseResult, 0, len(suite.TestCases)),
	}

	for _, testCase := range suite.TestCases {
		caseResult := r.runTestCase(ctx, testCase)
		result.TestResults = append(result.TestResults, caseResult)
		
		if caseResult.Status == TestStatusFailed {
			result.FailedCount++
		} else if caseResult.Status == TestStatusPassed {
			result.PassedCount++
		} else {
			result.SkippedCount++
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TotalCount = len(suite.TestCases)

	return result, nil
}

// runTestCase executes a single test case
func (r *TestRunner) runTestCase(ctx context.Context, testCase TestCase) TestCaseResult {
	result := TestCaseResult{
		TestCaseID: testCase.ID,
		TestName:   testCase.Name,
		StartTime:  time.Now(),
		Status:     TestStatusRunning,
	}

	// Execute the test
	renderResult, err := r.printService.ProcessDocument(ctx, testCase.Input.Document, testCase.Input.Options)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Status = TestStatusFailed
		result.Error = err.Error()
		result.Details = map[string]interface{}{
			"error_type": "execution_error",
			"message":    err.Error(),
		}
		return result
	}

	// Validate the result against expected output
	validation := r.validator.ValidateResult(renderResult, testCase.Expected)
	result.Status = validation.Status
	result.ValidationResult = validation
	result.ActualOutput = &ActualOutput{
		PageCount:  renderResult.PageCount,
		OutputSize: renderResult.OutputSize,
		RenderTime: renderResult.RenderTime,
		Warnings:   renderResult.Warnings,
		CacheHit:   renderResult.CacheHit,
	}

	return result
}

// LoadTestSuite loads a test suite from a JSON file
func LoadTestSuite(filepath string) (*TestSuite, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite file: %w", err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse test suite JSON: %w", err)
	}

	return &suite, nil
}

// SaveTestResults saves test results to a JSON file
func SaveTestResults(result *TestSuiteResult, outputPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test results: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write test results: %w", err)
	}

	return nil
}

// TestSuiteResult represents the result of running a test suite
type TestSuiteResult struct {
	SuiteName    string           `json:"suite_name"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Duration     time.Duration    `json:"duration"`
	TotalCount   int              `json:"total_count"`
	PassedCount  int              `json:"passed_count"`
	FailedCount  int              `json:"failed_count"`
	SkippedCount int              `json:"skipped_count"`
	TestResults  []TestCaseResult `json:"test_results"`
}

// TestCaseResult represents the result of running a single test case
type TestCaseResult struct {
	TestCaseID       string                 `json:"test_case_id"`
	TestName         string                 `json:"test_name"`
	Status           TestStatus             `json:"status"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	Error            string                 `json:"error,omitempty"`
	ValidationResult *ValidationResult      `json:"validation_result,omitempty"`
	ActualOutput     *ActualOutput          `json:"actual_output,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// ActualOutput represents the actual output from a test execution
type ActualOutput struct {
	PageCount  int           `json:"page_count"`
	OutputSize int64         `json:"output_size"`
	RenderTime time.Duration `json:"render_time"`
	Warnings   []string      `json:"warnings,omitempty"`
	CacheHit   bool          `json:"cache_hit"`
}

// TestStatus represents the status of a test case
type TestStatus string

const (
	TestStatusPending  TestStatus = "pending"
	TestStatusRunning  TestStatus = "running"
	TestStatusPassed   TestStatus = "passed"
	TestStatusFailed   TestStatus = "failed"
	TestStatusSkipped  TestStatus = "skipped"
	TestStatusTimeout  TestStatus = "timeout"
)

// GetSummary returns a summary of the test suite results
func (r *TestSuiteResult) GetSummary() string {
	successRate := float64(r.PassedCount) / float64(r.TotalCount) * 100
	return fmt.Sprintf(
		"Test Suite: %s\nTotal: %d, Passed: %d, Failed: %d, Skipped: %d\nSuccess Rate: %.2f%%\nDuration: %v",
		r.SuiteName,
		r.TotalCount,
		r.PassedCount,
		r.FailedCount,
		r.SkippedCount,
		successRate,
		r.Duration,
	)
}

// GetFailedTests returns all failed test cases
func (r *TestSuiteResult) GetFailedTests() []TestCaseResult {
	var failed []TestCaseResult
	for _, result := range r.TestResults {
		if result.Status == TestStatusFailed {
			failed = append(failed, result)
		}
	}
	return failed
}
