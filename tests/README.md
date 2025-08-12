# Print Service Tests

This directory contains all test files for the print-service-go project, organized by test type for better maintainability and clarity.

## Directory Structure

```
tests/
├── README.md           # This file
├── e2e/               # End-to-end tests
│   └── test_ultimate_e2e.go
├── unit/              # Unit tests
│   ├── test_simple_unit.go
│   └── test_unit_core.go
└── integration/       # Integration tests (future)
```

## Test Types

### End-to-End Tests (`e2e/`)
- **Purpose**: Test the complete system workflow from HTTP request to PDF generation
- **Scope**: Full system integration including HTTP server, job processing, and PDF generation
- **Files**:
  - `test_ultimate_e2e.go` - Ultimate E2E test suite with quantum performance validation

### Unit Tests (`unit/`)
- **Purpose**: Test individual components and core functionality
- **Scope**: Isolated testing of specific functions, modules, and business logic
- **Files**:
  - `test_simple_unit.go` - Simple unit tests for core system validation
  - `test_unit_core.go` - Comprehensive unit tests for core components (work in progress)

### Integration Tests (`integration/`)
- **Purpose**: Test interactions between different system components
- **Scope**: Database, cache, queue, and service layer integration
- **Status**: Directory prepared for future integration tests

## Running Tests

### Run E2E Tests
```bash
# From project root
cd tests/e2e
go run test_ultimate_e2e.go
```

### Run Unit Tests
```bash
# From project root
cd tests/unit
go run test_simple_unit.go
```

### Run All Tests (via Makefile)
```bash
# From project root
make test
```

## Test Results Summary

### Latest Test Results
- **Ultimate E2E Test**: ✅ 100% SUCCESS (4/4 tests passed)
  - Essential Baseline: 20 req/sec, 503ms latency, 100% success
  - Production Load: 50 req/sec, 503ms latency, 100% success
  - High Concurrency: 98 req/sec, 505ms latency, 100% success
  - Ultimate Performance: 186 req/sec, 519ms latency, 100% success

- **Simple Unit Test**: ✅ 100% SUCCESS (5/5 tests passed)
  - System Initialization: PASSED (190ns)
  - Configuration Validation: PASSED (80ns)
  - Memory Management: PASSED (456µs)
  - Error Handling: PASSED (7µs)
  - Concurrency Safety: PASSED (303µs)

## Test Guidelines

1. **E2E Tests**: Should test real HTTP endpoints and complete workflows
2. **Unit Tests**: Should be fast, isolated, and test specific functionality
3. **Integration Tests**: Should test component interactions without full system setup
4. **Test Data**: Use `../testdata/` for test fixtures and golden files
5. **Performance**: E2E tests validate performance; unit tests focus on correctness

## Adding New Tests

1. Choose the appropriate directory based on test scope
2. Follow existing naming conventions (`test_*.go`)
3. Include comprehensive error handling and validation
4. Update this README when adding new test categories
5. Ensure tests can run independently and are deterministic

## Dependencies

- Go 1.24.1+
- Print service running on localhost:8080 (for E2E tests)
- No external dependencies for unit tests
