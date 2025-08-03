# Pure Go Print Service - Comprehensive Makefile
# Author: Pure Go Print Service Team
# Description: Build, test, and deployment automation for HTML-to-PDF print service

# =============================================================================
# CONFIGURATION VARIABLES
# =============================================================================

# Application Configuration
APP_NAME := print-service
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go Configuration
GO_VERSION := 1.21
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0

# Build Configuration
BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin
DIST_DIR := $(BUILD_DIR)/dist
DOCKER_DIR := $(BUILD_DIR)/docker

# Server and Worker Binaries
SERVER_BINARY := $(BIN_DIR)/server
WORKER_BINARY := $(BIN_DIR)/worker

# Docker Configuration
DOCKER_REGISTRY := docker.io
DOCKER_NAMESPACE := printservice
SERVER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/server
WORKER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/worker
DOCKER_TAG := $(VERSION)

# Test Configuration
TEST_TIMEOUT := 300s
COVERAGE_DIR := $(BUILD_DIR)/coverage
COVERAGE_PROFILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

# Lint Configuration
GOLANGCI_LINT_VERSION := v1.54.2

# Build Flags
LDFLAGS := -w -s \
	-X main.version=$(VERSION) \
	-X main.buildTime=$(BUILD_TIME) \
	-X main.gitCommit=$(GIT_COMMIT)

BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

# =============================================================================
# DEFAULT TARGET
# =============================================================================

.DEFAULT_GOAL := help
.PHONY: help
help: ## Display this help message
	@echo "Pure Go Print Service - Makefile Help"
	@echo "====================================="
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Configuration:"
	@echo "  APP_NAME:     $(APP_NAME)"
	@echo "  VERSION:      $(VERSION)"
	@echo "  GO_VERSION:   $(GO_VERSION)"
	@echo "  GOOS:         $(GOOS)"
	@echo "  GOARCH:       $(GOARCH)"

# =============================================================================
# DEVELOPMENT TARGETS
# =============================================================================

.PHONY: dev
dev: ## Start development environment with hot reload
	@echo "Starting development environment..."
	@go run cmd/server/main.go

.PHONY: dev-worker
dev-worker: ## Start worker in development mode
	@echo "Starting worker in development mode..."
	@go run cmd/worker/main.go

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@go mod tidy

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# =============================================================================
# BUILD TARGETS
# =============================================================================

.PHONY: build
build: clean deps build-server build-worker ## Build all binaries

.PHONY: build-server
build-server: $(SERVER_BINARY) ## Build server binary

.PHONY: build-worker
build-worker: $(WORKER_BINARY) ## Build worker binary

$(SERVER_BINARY): $(shell find . -name "*.go" -type f)
	@echo "Building server binary..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(BUILD_FLAGS) -o $(SERVER_BINARY) ./cmd/server

$(WORKER_BINARY): $(shell find . -name "*.go" -type f)
	@echo "Building worker binary..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(BUILD_FLAGS) -o $(WORKER_BINARY) ./cmd/worker

.PHONY: build-all-platforms
build-all-platforms: clean deps ## Build binaries for all supported platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "Building for $$os/$$arch..."; \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build $(BUILD_FLAGS) \
				-o $(DIST_DIR)/$(APP_NAME)-server-$$os-$$arch$$ext ./cmd/server; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build $(BUILD_FLAGS) \
				-o $(DIST_DIR)/$(APP_NAME)-worker-$$os-$$arch$$ext ./cmd/worker; \
		done; \
	done

# =============================================================================
# TEST TARGETS
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-short
test-short: ## Run short tests only
	@echo "Running short tests..."
	@go test -v -short -race -timeout 60s ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -race -timeout $(TEST_TIMEOUT) -tags=integration ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -v -bench=. -benchmem ./...

# =============================================================================
# GOLDEN TEST DATA TARGETS
# =============================================================================

.PHONY: generate-golden-data
generate-golden-data: ## Generate golden test data variants
	@echo "Generating golden test data..."
	@go run cmd/testgen/main.go -output=./testdata/golden -verbose

.PHONY: generate-golden-basic
generate-golden-basic: ## Generate basic golden test variants
	@echo "Generating basic golden test data..."
	@go run cmd/testgen/main.go -variants=basic -output=./testdata/golden -verbose

.PHONY: generate-golden-rigor
generate-golden-rigor: ## Generate enhanced rigor golden test data
	@echo "Generating enhanced rigor golden test data..."
	./testgen -variant=rigor -output=testdata/golden/rigor_golden_data.json -verbose

.PHONY: generate-golden-true-rigor
generate-golden-true-rigor: ## Generate true rigor golden test data
	@echo "Generating true rigor golden test data..."
	./testgen -variant=true-rigor -output=testdata/golden/true_rigor_golden_data.json -verbose

.PHONY: generate-golden-edge
generate-golden-edge: ## Generate edge case golden test variants
	@echo "Generating edge case golden test data..."
	@go run cmd/testgen/main.go -variants=edge -output=./testdata/golden -verbose

.PHONY: generate-golden-stress
generate-golden-stress: ## Generate stress test golden variants
	@echo "Generating stress test golden data..."
	@go run cmd/testgen/main.go -variants=stress -output=./testdata/golden -verbose

.PHONY: test-golden
test-golden: generate-golden-data ## Generate and run golden tests
	@echo "Running golden tests..."
	@go test -v -tags=golden ./internal/tests/golden/...

# =============================================================================
# CODE QUALITY TARGETS
# =============================================================================

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))
	@golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: check
check: fmt vet lint test ## Run all code quality checks

# =============================================================================
# DOCKER TARGETS
# =============================================================================

.PHONY: docker-build
docker-build: docker-build-server docker-build-worker ## Build all Docker images

.PHONY: docker-build-server
docker-build-server: ## Build server Docker image
	@echo "Building server Docker image..."
	@mkdir -p $(DOCKER_DIR)
	@docker build -f deployments/docker/Dockerfile.server \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(SERVER_IMAGE):$(DOCKER_TAG) \
		-t $(SERVER_IMAGE):latest .

.PHONY: docker-build-worker
docker-build-worker: ## Build worker Docker image
	@echo "Building worker Docker image..."
	@mkdir -p $(DOCKER_DIR)
	@docker build -f deployments/docker/Dockerfile.worker \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(WORKER_IMAGE):$(DOCKER_TAG) \
		-t $(WORKER_IMAGE):latest .

.PHONY: docker-push
docker-push: ## Push Docker images to registry
	@echo "Pushing Docker images..."
	@docker push $(SERVER_IMAGE):$(DOCKER_TAG)
	@docker push $(SERVER_IMAGE):latest
	@docker push $(WORKER_IMAGE):$(DOCKER_TAG)
	@docker push $(WORKER_IMAGE):latest

.PHONY: docker-run-server
docker-run-server: ## Run server in Docker container
	@echo "Running server container..."
	@docker run -p 8080:8080 --rm $(SERVER_IMAGE):latest

.PHONY: docker-run-worker
docker-run-worker: ## Run worker in Docker container
	@echo "Running worker container..."
	@docker run --rm $(WORKER_IMAGE):latest

# =============================================================================
# DEPLOYMENT TARGETS
# =============================================================================

.PHONY: deploy-dev
deploy-dev: ## Deploy to development environment
	@echo "Deploying to development environment..."
	@kubectl apply -f deployments/k8s/development/

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging environment
	@echo "Deploying to staging environment..."
	@kubectl apply -f deployments/k8s/staging/

.PHONY: deploy-prod
deploy-prod: ## Deploy to production environment
	@echo "Deploying to production environment..."
	@kubectl apply -f deployments/k8s/production/

# =============================================================================
# DATABASE TARGETS
# =============================================================================

.PHONY: db-up
db-up: ## Start database services
	@echo "Starting database services..."
	@docker-compose -f deployments/docker-compose.dev.yml up -d redis postgres

.PHONY: db-down
db-down: ## Stop database services
	@echo "Stopping database services..."
	@docker-compose -f deployments/docker-compose.dev.yml down

.PHONY: db-reset
db-reset: db-down db-up ## Reset database services

# =============================================================================
# MONITORING TARGETS
# =============================================================================

.PHONY: metrics
metrics: ## Start metrics collection
	@echo "Starting metrics collection..."
	@docker-compose -f deployments/docker-compose.monitoring.yml up -d

.PHONY: logs
logs: ## View application logs
	@echo "Viewing application logs..."
	@docker-compose logs -f

# =============================================================================
# UTILITY TARGETS
# =============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean -cache
	@go clean -testcache
	@go clean -modcache

.PHONY: install
install: build ## Install binaries to GOPATH/bin
	@echo "Installing binaries..."
	@cp $(SERVER_BINARY) $(shell go env GOPATH)/bin/$(APP_NAME)-server
	@cp $(WORKER_BINARY) $(shell go env GOPATH)/bin/$(APP_NAME)-worker

.PHONY: uninstall
uninstall: ## Remove installed binaries
	@echo "Removing installed binaries..."
	@rm -f $(shell go env GOPATH)/bin/$(APP_NAME)-server
	@rm -f $(shell go env GOPATH)/bin/$(APP_NAME)-worker

.PHONY: version
version: ## Display version information
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(shell go version)"

.PHONY: info
info: ## Display build information
	@echo "Build Information:"
	@echo "=================="
	@echo "App Name:     $(APP_NAME)"
	@echo "Version:      $(VERSION)"
	@echo "Build Time:   $(BUILD_TIME)"
	@echo "Git Commit:   $(GIT_COMMIT)"
	@echo "Go Version:   $(GO_VERSION)"
	@echo "GOOS:         $(GOOS)"
	@echo "GOARCH:       $(GOARCH)"
	@echo "CGO Enabled:  $(CGO_ENABLED)"

# =============================================================================
# RELEASE TARGETS
# =============================================================================

.PHONY: release
release: clean check build-all-platforms docker-build ## Create a complete release

.PHONY: release-notes
release-notes: ## Generate release notes
	@echo "Generating release notes for $(VERSION)..."
	@git log --pretty=format:"- %s" $(shell git describe --tags --abbrev=0)..HEAD

# =============================================================================
# DEVELOPMENT WORKFLOW TARGETS
# =============================================================================

.PHONY: setup
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: pre-commit
pre-commit: fmt vet lint test-short ## Run pre-commit checks

.PHONY: ci
ci: deps check test-coverage build ## Run CI pipeline

# =============================================================================
# SPECIAL TARGETS
# =============================================================================

# Ensure build directories exist
$(BUILD_DIR) $(BIN_DIR) $(DIST_DIR) $(DOCKER_DIR) $(COVERAGE_DIR):
	@mkdir -p $@

# Mark phony targets
.PHONY: all
all: clean setup check build docker-build ## Run complete build pipeline
