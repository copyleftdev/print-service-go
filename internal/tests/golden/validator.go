package golden

import (
	"fmt"
	"math"
	"time"

	"print-service/internal/core/domain"
)

// ResultValidator validates test results against expected outcomes
type ResultValidator struct {
	tolerances Tolerances
}

// Tolerances defines acceptable variances for test validation
type Tolerances struct {
	PageCountVariance  int           // Acceptable page count difference
	OutputSizeVariance float64       // Acceptable output size variance (percentage)
	RenderTimeVariance time.Duration // Acceptable render time variance
}

// NewResultValidator creates a new result validator with default tolerances
func NewResultValidator() *ResultValidator {
	return &ResultValidator{
		tolerances: Tolerances{
			PageCountVariance:  1,                // ±1 page
			OutputSizeVariance: 0.1,              // ±10%
			RenderTimeVariance: 500 * time.Millisecond, // ±500ms
		},
	}
}

// ValidationResult represents the result of validating a test case
type ValidationResult struct {
	Status      TestStatus               `json:"status"`
	Score       float64                  `json:"score"`       // 0.0 to 1.0
	Violations  []ValidationViolation    `json:"violations"`
	Matches     []ValidationMatch        `json:"matches"`
	Details     map[string]interface{}   `json:"details"`
}

// ValidationViolation represents a validation failure
type ValidationViolation struct {
	Field    string      `json:"field"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Severity Severity    `json:"severity"`
	Message  string      `json:"message"`
}

// ValidationMatch represents a successful validation
type ValidationMatch struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// Severity represents the severity of a validation violation
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityMajor    Severity = "major"
	SeverityMinor    Severity = "minor"
	SeverityWarning  Severity = "warning"
)

// ValidateResult validates a render result against expected output
func (v *ResultValidator) ValidateResult(actual *domain.RenderResult, expected ExpectedOutput) *ValidationResult {
	result := &ValidationResult{
		Status:     TestStatusPassed,
		Score:      1.0,
		Violations: []ValidationViolation{},
		Matches:    []ValidationMatch{},
		Details:    make(map[string]interface{}),
	}

	var totalChecks, passedChecks int

	// Validate page count
	if expected.PageCount > 0 {
		totalChecks++
		if v.validatePageCount(actual.PageCount, expected.PageCount, result) {
			passedChecks++
		}
	}

	// Validate output size
	if expected.OutputSize > 0 {
		totalChecks++
		if v.validateOutputSize(actual.OutputSize, expected.OutputSize, result) {
			passedChecks++
		}
	}

	// Validate render time
	if expected.RenderTime > 0 {
		totalChecks++
		if v.validateRenderTime(actual.RenderTime, expected.RenderTime, result) {
			passedChecks++
		}
	}

	// Validate warnings
	if len(expected.Warnings) > 0 {
		totalChecks++
		if v.validateWarnings(actual.Warnings, expected.Warnings, result) {
			passedChecks++
		}
	}

	// Validate errors (if any expected)
	if len(expected.Errors) > 0 {
		totalChecks++
		// For this implementation, we assume no errors should occur in successful cases
		// This would need to be adapted based on specific test requirements
		result.Violations = append(result.Violations, ValidationViolation{
			Field:    "errors",
			Expected: expected.Errors,
			Actual:   []string{}, // Assuming no errors in successful execution
			Severity: SeverityMajor,
			Message:  "Expected errors but none occurred",
		})
	}

	// Calculate overall score and status
	if totalChecks > 0 {
		result.Score = float64(passedChecks) / float64(totalChecks)
	}

	// Determine overall status based on violations
	criticalViolations := v.countViolationsBySeverity(result.Violations, SeverityCritical)
	majorViolations := v.countViolationsBySeverity(result.Violations, SeverityMajor)

	if criticalViolations > 0 {
		result.Status = TestStatusFailed
	} else if majorViolations > 0 {
		result.Status = TestStatusFailed
	} else if result.Score < 0.8 {
		result.Status = TestStatusFailed
	}

	result.Details["total_checks"] = totalChecks
	result.Details["passed_checks"] = passedChecks
	result.Details["critical_violations"] = criticalViolations
	result.Details["major_violations"] = majorViolations

	return result
}

// validatePageCount validates the page count
func (v *ResultValidator) validatePageCount(actual, expected int, result *ValidationResult) bool {
	diff := int(math.Abs(float64(actual - expected)))
	
	if diff <= v.tolerances.PageCountVariance {
		result.Matches = append(result.Matches, ValidationMatch{
			Field:   "page_count",
			Value:   actual,
			Message: fmt.Sprintf("Page count matches expected value (±%d tolerance)", v.tolerances.PageCountVariance),
		})
		return true
	}

	severity := SeverityMinor
	if diff > v.tolerances.PageCountVariance*2 {
		severity = SeverityMajor
	}
	if diff > v.tolerances.PageCountVariance*5 {
		severity = SeverityCritical
	}

	result.Violations = append(result.Violations, ValidationViolation{
		Field:    "page_count",
		Expected: expected,
		Actual:   actual,
		Severity: severity,
		Message:  fmt.Sprintf("Page count differs by %d pages (tolerance: ±%d)", diff, v.tolerances.PageCountVariance),
	})
	return false
}

// validateOutputSize validates the output file size
func (v *ResultValidator) validateOutputSize(actual, expected int64, result *ValidationResult) bool {
	if expected == 0 {
		return true // Skip validation if no expected size
	}

	variance := math.Abs(float64(actual-expected)) / float64(expected)
	
	if variance <= v.tolerances.OutputSizeVariance {
		result.Matches = append(result.Matches, ValidationMatch{
			Field:   "output_size",
			Value:   actual,
			Message: fmt.Sprintf("Output size within acceptable variance (%.1f%%)", v.tolerances.OutputSizeVariance*100),
		})
		return true
	}

	severity := SeverityMinor
	if variance > v.tolerances.OutputSizeVariance*2 {
		severity = SeverityMajor
	}
	if variance > v.tolerances.OutputSizeVariance*5 {
		severity = SeverityCritical
	}

	result.Violations = append(result.Violations, ValidationViolation{
		Field:    "output_size",
		Expected: expected,
		Actual:   actual,
		Severity: severity,
		Message:  fmt.Sprintf("Output size variance %.1f%% exceeds tolerance %.1f%%", variance*100, v.tolerances.OutputSizeVariance*100),
	})
	return false
}

// validateRenderTime validates the rendering time
func (v *ResultValidator) validateRenderTime(actual, expected time.Duration, result *ValidationResult) bool {
	if expected == 0 {
		return true // Skip validation if no expected time
	}

	diff := time.Duration(math.Abs(float64(actual - expected)))
	
	if diff <= v.tolerances.RenderTimeVariance {
		result.Matches = append(result.Matches, ValidationMatch{
			Field:   "render_time",
			Value:   actual,
			Message: fmt.Sprintf("Render time within acceptable variance (±%v)", v.tolerances.RenderTimeVariance),
		})
		return true
	}

	severity := SeverityMinor
	if diff > v.tolerances.RenderTimeVariance*2 {
		severity = SeverityMajor
	}
	if diff > v.tolerances.RenderTimeVariance*5 {
		severity = SeverityCritical
	}

	result.Violations = append(result.Violations, ValidationViolation{
		Field:    "render_time",
		Expected: expected,
		Actual:   actual,
		Severity: severity,
		Message:  fmt.Sprintf("Render time differs by %v (tolerance: ±%v)", diff, v.tolerances.RenderTimeVariance),
	})
	return false
}

// validateWarnings validates expected warnings
func (v *ResultValidator) validateWarnings(actual, expected []string, result *ValidationResult) bool {
	// Check if all expected warnings are present
	missingWarnings := []string{}
	for _, expectedWarning := range expected {
		found := false
		for _, actualWarning := range actual {
			if actualWarning == expectedWarning {
				found = true
				break
			}
		}
		if !found {
			missingWarnings = append(missingWarnings, expectedWarning)
		}
	}

	if len(missingWarnings) == 0 {
		result.Matches = append(result.Matches, ValidationMatch{
			Field:   "warnings",
			Value:   actual,
			Message: "All expected warnings are present",
		})
		return true
	}

	result.Violations = append(result.Violations, ValidationViolation{
		Field:    "warnings",
		Expected: expected,
		Actual:   actual,
		Severity: SeverityMinor,
		Message:  fmt.Sprintf("Missing expected warnings: %v", missingWarnings),
	})
	return false
}

// countViolationsBySeverity counts violations by severity level
func (v *ResultValidator) countViolationsBySeverity(violations []ValidationViolation, severity Severity) int {
	count := 0
	for _, violation := range violations {
		if violation.Severity == severity {
			count++
		}
	}
	return count
}

// SetTolerances allows customizing validation tolerances
func (v *ResultValidator) SetTolerances(tolerances Tolerances) {
	v.tolerances = tolerances
}

// GetTolerances returns current validation tolerances
func (v *ResultValidator) GetTolerances() Tolerances {
	return v.tolerances
}
