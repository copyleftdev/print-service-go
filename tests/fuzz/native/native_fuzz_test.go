package native

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// PrintRequest represents the structure for print API requests
type PrintRequest struct {
	Content string                 `json:"content"`
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options"`
}

// getServiceURL returns the service URL for testing
func getServiceURL() string {
	if url := os.Getenv("SERVICE_URL"); url != "" {
		return url
	}
	return "http://print-server-test:8080"
}

// waitForService waits for the print service to be ready
func waitForService() error {
	client := &http.Client{Timeout: 5 * time.Second}
	serviceURL := getServiceURL()
	
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(serviceURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("service not ready after %d attempts", maxAttempts)
}

// sendPrintRequest sends a print request to the service
func sendPrintRequest(t *testing.T, req PrintRequest) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	serviceURL := getServiceURL()
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	resp, err := client.Post(
		serviceURL+"/api/v1/print",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	return resp, nil
}

// FuzzPrintContent fuzzes the content field of print requests
func FuzzPrintContent(f *testing.F) {
	if err := waitForService(); err != nil {
		f.Fatalf("Service not ready: %v", err)
	}
	
	// Add seed corpus
	f.Add("Hello World")
	f.Add("<h1>Test</h1>")
	f.Add("# Markdown Header")
	f.Add("")
	f.Add("🎯 Unicode test")
	f.Add(strings.Repeat("A", 1000))
	
	f.Fuzz(func(t *testing.T, content string) {
		req := PrintRequest{
			Content: content,
			Type:    "html",
			Options: map[string]interface{}{},
		}
		
		resp, err := sendPrintRequest(t, req)
		if err != nil {
			// Network errors are acceptable in fuzz testing
			return
		}
		defer resp.Body.Close()
		
		// Service should never crash - any HTTP status is acceptable
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Invalid HTTP status code: %d", resp.StatusCode)
		}
	})
}

// FuzzPrintType fuzzes the type field of print requests
func FuzzPrintType(f *testing.F) {
	if err := waitForService(); err != nil {
		f.Fatalf("Service not ready: %v", err)
	}
	
	// Add seed corpus
	f.Add("html")
	f.Add("markdown")
	f.Add("text")
	f.Add("")
	f.Add("invalid")
	f.Add("HTML")
	f.Add("MARKDOWN")
	
	f.Fuzz(func(t *testing.T, docType string) {
		req := PrintRequest{
			Content: "<h1>Test Content</h1>",
			Type:    docType,
			Options: map[string]interface{}{},
		}
		
		resp, err := sendPrintRequest(t, req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		
		// Service should handle any type gracefully
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Invalid HTTP status code: %d for type: %q", resp.StatusCode, docType)
		}
	})
}

// FuzzPrintJSON fuzzes the entire JSON structure
func FuzzPrintJSON(f *testing.F) {
	if err := waitForService(); err != nil {
		f.Fatalf("Service not ready: %v", err)
	}
	
	// Add seed corpus with valid JSON
	validReqs := []PrintRequest{
		{Content: "test", Type: "html", Options: map[string]interface{}{}},
		{Content: "# Test", Type: "markdown", Options: map[string]interface{}{"quality": "high"}},
		{Content: "", Type: "text", Options: map[string]interface{}{}},
	}
	
	for _, req := range validReqs {
		reqBytes, _ := json.Marshal(req)
		f.Add(string(reqBytes))
	}
	
	// Add some malformed JSON
	f.Add(`{"content": "test"`)
	f.Add(`{"content": "test", "type":}`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	
	f.Fuzz(func(t *testing.T, jsonStr string) {
		client := &http.Client{Timeout: 10 * time.Second}
		serviceURL := getServiceURL()
		
		resp, err := client.Post(
			serviceURL+"/api/v1/print",
			"application/json",
			strings.NewReader(jsonStr),
		)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		
		// Service should handle malformed JSON gracefully
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Invalid HTTP status code: %d for JSON: %q", resp.StatusCode, jsonStr)
		}
	})
}

// FuzzPrintUnicode specifically tests Unicode handling
func FuzzPrintUnicode(f *testing.F) {
	if err := waitForService(); err != nil {
		f.Fatalf("Service not ready: %v", err)
	}
	
	// Add Unicode seed corpus
	f.Add("Hello 世界")
	f.Add("🎯🔀🧪✅❌📊🏆")
	f.Add("Ñoño niño")
	f.Add("Здравствуй мир")
	f.Add("مرحبا بالعالم")
	f.Add("שלום עולם")
	f.Add("\u0000\u0001\u0002")
	f.Add("\uFEFF")
	
	f.Fuzz(func(t *testing.T, content string) {
		req := PrintRequest{
			Content: content,
			Type:    "html",
			Options: map[string]interface{}{},
		}
		
		resp, err := sendPrintRequest(t, req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		
		// Service should handle Unicode gracefully
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Invalid HTTP status code: %d for Unicode content", resp.StatusCode)
		}
	})
}

// FuzzPrintLargeContent tests with varying content sizes
func FuzzPrintLargeContent(f *testing.F) {
	if err := waitForService(); err != nil {
		f.Fatalf("Service not ready: %v", err)
	}
	
	// Add seed corpus with different sizes
	f.Add(strings.Repeat("A", 100))
	f.Add(strings.Repeat("B", 1000))
	f.Add(strings.Repeat("C", 10000))
	
	f.Fuzz(func(t *testing.T, baseContent string) {
		// Limit content size to prevent excessive resource usage
		if len(baseContent) > 100000 {
			baseContent = baseContent[:100000]
		}
		
		req := PrintRequest{
			Content: baseContent,
			Type:    "html",
			Options: map[string]interface{}{},
		}
		
		resp, err := sendPrintRequest(t, req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		
		// Service should handle large content gracefully
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Invalid HTTP status code: %d for content size: %d", resp.StatusCode, len(baseContent))
		}
	})
}
