package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// UltimateE2ETest - Essential end-to-end validation of the print service
type UltimateE2ETest struct {
	serviceURL string
	httpClient *http.Client
}

// TestResult captures essential E2E metrics
type TestResult struct {
	TestName     string
	Concurrent   int
	Total        int
	Duration     time.Duration
	Throughput   float64
	AvgLatency   time.Duration
	SuccessRate  float64
	AvgPDFSize   float64
	TotalPDFData float64
	HTTPErrors   map[int]int
}

// PrintRequest represents the API request
type PrintRequest struct {
	Content string                 `json:"content"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// PrintResponse represents the API response
type PrintResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// NewUltimateE2ETest creates the ultimate E2E test
func NewUltimateE2ETest(serviceURL string) *UltimateE2ETest {
	return &UltimateE2ETest{
		serviceURL: serviceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// RunUltimateE2ETest executes essential E2E validation
func (t *UltimateE2ETest) RunUltimateE2ETest() error {
	fmt.Println("🚀 ULTIMATE E2E TEST SUITE")
	fmt.Println("==========================")
	fmt.Printf("🎯 Testing REAL Print Service Pipeline\n")
	fmt.Printf("📊 Service URL: %s\n", t.serviceURL)
	fmt.Printf("💻 System: %d CPU cores, Go %s\n\n", runtime.NumCPU(), runtime.Version())

	// Essential test scenarios
	tests := []struct {
		name       string
		concurrent int
		total      int
		document   string
	}{
		{"Essential Baseline", 10, 50, "Simple Invoice"},
		{"Production Load", 25, 100, "Simple Invoice"},
		{"High Concurrency", 50, 200, "Complex Report"},
		{"Ultimate Performance", 100, 300, "Complex Report"},
	}

	var results []TestResult

	for i, test := range tests {
		fmt.Printf("🔥 TEST %d/%d: %s\n", i+1, len(tests), test.name)
		fmt.Printf("   Concurrent: %d | Total: %d | Document: %s\n", test.concurrent, test.total, test.document)

		result, err := t.runSingleTest(test.name, test.concurrent, test.total, test.document)
		if err != nil {
			fmt.Printf("   ❌ ERROR: %v\n", err)
			continue
		}

		results = append(results, *result)
		t.printResult(result)

		if i < len(tests)-1 {
			fmt.Println("   ⏳ Cooling down...")
			time.Sleep(2 * time.Second)
		}
		fmt.Println()
	}

	t.printSummary(results)
	return nil
}

// runSingleTest executes a single E2E test scenario
func (t *UltimateE2ETest) runSingleTest(testName string, concurrent, total int, docType string) (*TestResult, error) {
	var (
		successCount  int64
		totalLatency  int64
		totalPDFSize  int64
		pdfCount      int64
		httpErrors    = make(map[int]int)
		httpErrorsMux sync.Mutex
		wg            sync.WaitGroup
		semaphore     = make(chan struct{}, concurrent)
	)

	// Create test document content
	content := t.createTestDocument(docType)

	startTime := time.Now()

	// Execute concurrent requests
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			reqStart := time.Now()
			pdfSize, statusCode, err := t.makeRequest(content, reqID)
			latency := time.Since(reqStart)

			atomic.AddInt64(&totalLatency, int64(latency))

			if err != nil || statusCode != 200 {
				httpErrorsMux.Lock()
				httpErrors[statusCode]++
				httpErrorsMux.Unlock()
				return
			}

			atomic.AddInt64(&successCount, 1)
			if pdfSize > 0 {
				atomic.AddInt64(&totalPDFSize, int64(pdfSize))
				atomic.AddInt64(&pdfCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Calculate metrics
	successRate := float64(successCount) / float64(total) * 100
	throughput := float64(total) / duration.Seconds()
	avgLatency := time.Duration(totalLatency / int64(total))
	avgPDFSize := float64(totalPDFSize) / float64(pdfCount) / 1024 // KB
	totalPDFData := float64(totalPDFSize) / 1024 / 1024            // MB

	return &TestResult{
		TestName:     testName,
		Concurrent:   concurrent,
		Total:        total,
		Duration:     duration,
		Throughput:   throughput,
		AvgLatency:   avgLatency,
		SuccessRate:  successRate,
		AvgPDFSize:   avgPDFSize,
		TotalPDFData: totalPDFData,
		HTTPErrors:   httpErrors,
	}, nil
}

// makeRequest makes a complete E2E request (submit + poll + download)
func (t *UltimateE2ETest) makeRequest(content string, reqID int) (int, int, error) {
	// 1. Submit print job
	request := PrintRequest{
		Content: content,
		Options: map[string]interface{}{"format": "A4"},
	}

	reqBody, _ := json.Marshal(request)
	resp, err := t.httpClient.Post(t.serviceURL+"/api/v1/print", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		return 0, resp.StatusCode, fmt.Errorf("submit failed: %d", resp.StatusCode)
	}

	var printResp PrintResponse
	if err := json.NewDecoder(resp.Body).Decode(&printResp); err != nil {
		return 0, resp.StatusCode, err
	}

	// 2. Poll for completion (max 10 seconds)
	jobID := printResp.JobID
	for attempts := 0; attempts < 20; attempts++ {
		time.Sleep(500 * time.Millisecond)

		statusResp, err := t.httpClient.Get(fmt.Sprintf("%s/api/v1/print/%s", t.serviceURL, jobID))
		if err != nil {
			continue
		}
		statusResp.Body.Close()

		if statusResp.StatusCode == 200 {
			break
		}
	}

	// 3. Download PDF
	downloadResp, err := t.httpClient.Get(fmt.Sprintf("%s/api/v1/print/%s/download", t.serviceURL, jobID))
	if err != nil {
		return 0, 0, err
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != 200 {
		return 0, downloadResp.StatusCode, fmt.Errorf("download failed: %d", downloadResp.StatusCode)
	}

	// Read PDF data to get size
	pdfData, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return 0, downloadResp.StatusCode, err
	}

	return len(pdfData), 200, nil
}

// createTestDocument creates test content based on document type
func (t *UltimateE2ETest) createTestDocument(docType string) string {
	switch docType {
	case "Simple Invoice":
		return `<html><body><h1>Invoice #12345</h1><p>Amount: $1,234.56</p><p>Date: 2025-08-12</p></body></html>`
	case "Complex Report":
		return `<html><body><h1>Performance Report</h1><div style="margin:20px;"><h2>Executive Summary</h2><p>This report analyzes system performance metrics and provides recommendations for optimization.</p><table border="1"><tr><th>Metric</th><th>Value</th></tr><tr><td>Throughput</td><td>2,942 req/sec</td></tr><tr><td>Latency</td><td>4.03ms</td></tr></table></div></body></html>`
	default:
		return `<html><body><h1>Test Document</h1><p>Content for testing.</p></body></html>`
	}
}

// printResult prints test result
func (t *UltimateE2ETest) printResult(result *TestResult) {
	fmt.Printf("   📊 E2E RESULTS:\n")
	fmt.Printf("      Duration:        %v\n", result.Duration)
	fmt.Printf("      Throughput:      %.0f req/sec\n", result.Throughput)
	fmt.Printf("      Avg Latency:     %v\n", result.AvgLatency)
	fmt.Printf("      Success Rate:    %.1f%% (%d/%d)\n", result.SuccessRate, int(result.SuccessRate*float64(result.Total)/100), result.Total)
	fmt.Printf("      Avg PDF Size:    %.1f KB\n", result.AvgPDFSize)
	fmt.Printf("      Total PDF Data:  %.1f MB\n", result.TotalPDFData)

	if len(result.HTTPErrors) > 0 {
		fmt.Printf("      HTTP Errors:\n")
		for code, count := range result.HTTPErrors {
			fmt.Printf("        %d: %d requests\n", code, count)
		}
	}

	status := "✅ EXCELLENT"
	if result.SuccessRate < 90 {
		status = "⚠️  NEEDS OPTIMIZATION"
	}
	if result.SuccessRate < 50 {
		status = "❌ FAILED"
	}
	fmt.Printf("      Status:          %s\n", status)
}

// printSummary prints overall test summary
func (t *UltimateE2ETest) printSummary(results []TestResult) {
	fmt.Println("🏆 ULTIMATE E2E TEST SUMMARY")
	fmt.Println("=============================")
	fmt.Printf("📈 E2E PERFORMANCE ACHIEVEMENTS:\n")

	var totalThroughput, totalLatency float64
	passedTests := 0

	for _, result := range results {
		status := "❌ FAILED"
		if result.SuccessRate >= 90 {
			status = "✅ PASSED"
			passedTests++
		} else if result.SuccessRate >= 50 {
			status = "⚠️  PARTIAL"
		}

		fmt.Printf("   %s: %.0f req/sec, %v latency, %.1f%% success, %.1fKB avg PDF %s\n",
			result.TestName, result.Throughput, result.AvgLatency, result.SuccessRate, result.AvgPDFSize, status)

		totalThroughput += result.Throughput
		totalLatency += float64(result.AvgLatency)
	}

	fmt.Printf("\n🎯 ULTIMATE E2E FINAL VERDICT:\n")
	fmt.Printf("   Tests Passed:         %d/%d\n", passedTests, len(results))
	fmt.Printf("   Average Throughput:   %.0f requests/second\n", totalThroughput/float64(len(results)))
	fmt.Printf("   Average Latency:      %v\n", time.Duration(totalLatency/float64(len(results))))

	if passedTests == len(results) {
		fmt.Printf("   Status:               🎉 QUANTUM PERFORMANCE ACHIEVED!\n")
	} else if passedTests > len(results)/2 {
		fmt.Printf("   Status:               ⚠️  Good performance, optimization opportunities exist\n")
	} else {
		fmt.Printf("   Status:               ❌ Performance targets missed\n")
	}

	fmt.Printf("\n🚀 ULTIMATE E2E TEST COMPLETE!\n")
}

// checkServiceHealth verifies the service is running
func (t *UltimateE2ETest) checkServiceHealth() error {
	fmt.Printf("🔍 Checking if print service is running at %s...\n", t.serviceURL)

	resp, err := t.httpClient.Get(t.serviceURL + "/health")
	if err != nil {
		return fmt.Errorf("service not accessible: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("service health check failed: %d", resp.StatusCode)
	}

	fmt.Println("✅ Print service is running!")
	return nil
}

func main() {
	test := NewUltimateE2ETest("http://localhost:8080")

	if err := test.checkServiceHealth(); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Println()
	if err := test.RunUltimateE2ETest(); err != nil {
		fmt.Printf("❌ Test failed: %v\n", err)
		return
	}

	fmt.Println("🎉 Ultimate E2E test completed successfully!")
}
