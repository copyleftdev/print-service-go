package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"print-service/internal/tests/golden"
)

func main() {
	var (
		outputDir = flag.String("output", "./testdata/golden", "Output directory for generated test data")
		format    = flag.String("format", "json", "Output format (json, yaml)")
		variants  = flag.String("variants", "all", "Test variants to generate (all, basic, edge, stress)")
		verbose   = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	generator := golden.NewTestDataGenerator()

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate test data based on variants
	var testSuites []golden.TestSuite
	switch *variants {
	case "basic":
		testSuites = append(testSuites, generator.GenerateBasicVariants())
	case "edge":
		testSuites = append(testSuites, generator.GenerateEdgeCaseVariants())
	case "stress":
		testSuites = append(testSuites, generator.GenerateStressTestVariants())
	case "security":
		testSuites = append(testSuites, generator.GenerateSecurityTestVariants())
	case "performance":
		testSuites = append(testSuites, generator.GeneratePerformanceTestVariants())
	case "rigor":
		// Enhanced rigor - generate comprehensive test suites
		enhancedGen := golden.NewEnhancedRigorGenerator()
		testSuites = enhancedGen.GenerateAllEnhancedVariants()
	case "true-rigor":
		// True rigor - generate production-grade test suites
		trueRigorGen := golden.NewTrueRigorGenerator()
		testSuites = trueRigorGen.GenerateAllRigorousVariants()
	case "ultra-rigor":
		// Ultra rigor - generate next-generation ultra-rigorous test suites
		ultraRigorGen := golden.NewUltraRigorGenerator()
		testSuites = ultraRigorGen.GenerateAllUltraRigorousVariants()
	case "all":
		testSuites = append(testSuites,
			generator.GenerateBasicVariants(),
			generator.GenerateEdgeCaseVariants(),
			generator.GenerateStressTestVariants(),
			generator.GenerateSecurityTestVariants(),
			generator.GeneratePerformanceTestVariants(),
		)
	default:
		log.Fatalf("Unknown variant type: %s (available: basic, edge, stress, security, performance, rigor, true-rigor, ultra-rigor, all)", *variants)
	}

	// Write test suites to files
	for _, suite := range testSuites {
		filename := fmt.Sprintf("%s_golden_data.%s", suite.Name, *format)
		outputPath := filepath.Join(*outputDir, filename)

		var data []byte
		var err error

		switch *format {
		case "json":
			data, err = json.MarshalIndent(suite, "", "  ")
		case "yaml":
			// Would need yaml package for this
			log.Fatalf("YAML format not yet implemented")
		default:
			log.Fatalf("Unsupported format: %s", *format)
		}

		if err != nil {
			log.Fatalf("Failed to marshal test suite %s: %v", suite.Name, err)
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			log.Fatalf("Failed to write test suite %s: %v", suite.Name, err)
		}

		if *verbose {
			log.Printf("Generated %s with %d test cases", outputPath, len(suite.TestCases))
		}
	}

	fmt.Printf("Successfully generated golden test data in %s\n", *outputDir)
}
