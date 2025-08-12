# Print Service - Simple Makefile
# Clean, focused build automation for our refactored quantum performance print service

# =============================================================================
# CONFIGURATION
# =============================================================================

APP_NAME := print-service
BINARY_NAME := print-service
MAIN_PATH := ./cmd/server/main.go
BUILD_DIR := ./bin

# Go settings
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0

# =============================================================================
# DEFAULT TARGET
# =============================================================================

.DEFAULT_GOAL := help

# =============================================================================
# HELP
# =============================================================================

.PHONY: help
help: ## Show this help message
	@echo "Print Service - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

# =============================================================================
# BUILD COMMANDS
# =============================================================================

.PHONY: build
build: ## Build the print service binary
	@echo "🔨 Building print service..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "🔨 Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)
	@echo "✅ Linux build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf output/
	@rm -rf temp/
	@echo "✅ Clean complete"

# =============================================================================
# RUN COMMANDS
# =============================================================================

.PHONY: run
run: build ## Build and run the print service
	@echo "🚀 Starting print service..."
	@$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: dev
dev: ## Run in development mode (rebuild on changes)
	@echo "🔄 Running in development mode..."
	@go run $(MAIN_PATH)

# =============================================================================
# TEST COMMANDS
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "🧪 Running all tests..."
	@echo ""
	@echo "📋 Running unit tests..."
	@cd tests/unit && go run test_simple_unit.go
	@echo ""
	@echo "🌐 Running E2E tests..."
	@cd tests/e2e && go run test_ultimate_e2e.go
	@echo ""
	@echo "✅ All tests complete!"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@cd tests/unit && go run test_simple_unit.go

.PHONY: test-e2e
test-e2e: ## Run E2E tests only (requires service to be running)
	@echo "🌐 Running E2E tests..."
	@cd tests/e2e && go run test_ultimate_e2e.go

.PHONY: test-go
test-go: ## Run Go standard tests
	@echo "🧪 Running Go tests..."
	@go test -v -race ./...

# =============================================================================
# QUALITY COMMANDS
# =============================================================================

.PHONY: lint
lint: ## Run linting tools
	@echo "🔍 Running linting tools..."
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running staticcheck..."
	@staticcheck ./...
	@echo "Running ineffassign..."
	@ineffassign ./...
	@echo "Checking formatting..."
	@gofmt -l . | grep -v '^$$' && echo "Files need formatting!" && exit 1 || echo "All files properly formatted"
	@echo "✅ Linting complete!"

.PHONY: fmt
fmt: ## Format Go code
	@echo "🎨 Formatting Go code..."
	@gofmt -w .
	@echo "✅ Formatting complete!"

.PHONY: tidy
tidy: ## Tidy Go modules
	@echo "📦 Tidying Go modules..."
	@go mod tidy
	@echo "✅ Module tidy complete!"

# =============================================================================
# DEVELOPMENT COMMANDS
# =============================================================================

.PHONY: deps
deps: ## Install development dependencies
	@echo "📦 Installing development dependencies..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/gordonklaus/ineffassign@latest
	@echo "✅ Dependencies installed!"

.PHONY: check
check: tidy fmt lint test-unit ## Run all quality checks
	@echo "✅ All quality checks passed!"

# =============================================================================
# SERVICE MANAGEMENT
# =============================================================================

.PHONY: start
start: build ## Start the print service in background
	@echo "🚀 Starting print service in background..."
	@$(BUILD_DIR)/$(BINARY_NAME) &
	@echo "✅ Print service started!"

.PHONY: stop
stop: ## Stop the print service
	@echo "🛑 Stopping print service..."
	@pkill -f $(BINARY_NAME) || echo "Service not running"
	@echo "✅ Print service stopped!"

.PHONY: restart
restart: stop start ## Restart the print service
	@echo "🔄 Print service restarted!"

.PHONY: status
status: ## Check print service status
	@echo "📊 Checking print service status..."
	@pgrep -f $(BINARY_NAME) > /dev/null && echo "✅ Print service is running" || echo "❌ Print service is not running"

# =============================================================================
# UTILITY COMMANDS
# =============================================================================

.PHONY: version
version: ## Show version information
	@echo "Print Service Version Information:"
	@echo "  Go Version: $(shell go version)"
	@echo "  OS/Arch: $(GOOS)/$(GOARCH)"
	@echo "  CGO: $(CGO_ENABLED)"

.PHONY: info
info: ## Show project information
	@echo "Print Service Project Information:"
	@echo "  Name: $(APP_NAME)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Main: $(MAIN_PATH)"
	@echo "  Build Dir: $(BUILD_DIR)"
	@echo "  Go Version: $(shell go version)"
	@echo ""
	@echo "Project Structure:"
	@tree -d -L 2 --charset ascii

# =============================================================================
# QUICK COMMANDS
# =============================================================================

.PHONY: all
all: clean build test ## Clean, build, and test everything
	@echo "🎉 All tasks completed successfully!"

.PHONY: quick
quick: build test-unit ## Quick build and unit test
	@echo "⚡ Quick validation complete!"

.PHONY: full
full: clean deps check build test ## Full development workflow
	@echo "🎉 Full development workflow completed!"

# =============================================================================
# ALIASES
# =============================================================================

.PHONY: b
b: build ## Alias for build

.PHONY: r
r: run ## Alias for run

.PHONY: t
t: test-unit ## Alias for test-unit

.PHONY: c
c: clean ## Alias for clean
