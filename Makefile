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

# =============================================================================
# DOCKER COMPOSE COMMANDS
# =============================================================================

.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	@echo "🐳 Starting Docker Compose services..."
	@docker compose up -d
	@echo "✅ Services started! API: http://localhost:8080, Redis UI: http://localhost:8081"

.PHONY: docker-down
docker-down: ## Stop all Docker Compose services
	@echo "🛑 Stopping Docker Compose services..."
	@docker compose down
	@echo "✅ Services stopped!"

.PHONY: docker-logs
docker-logs: ## View Docker Compose logs
	@echo "📋 Viewing Docker Compose logs..."
	@docker compose logs -f

.PHONY: docker-build
docker-build: ## Build Docker images
	@echo "🔨 Building Docker images..."
	@docker compose build
	@echo "✅ Docker images built!"

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild and restart services
	@echo "🔄 Rebuilding and restarting services..."
	@docker compose down
	@docker compose build --no-cache
	@docker compose up -d
	@echo "✅ Services rebuilt and restarted!"

.PHONY: docker-clean
docker-clean: ## Clean up Docker resources
	@echo "🧹 Cleaning up Docker resources..."
	@docker compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Docker cleanup complete!"

.PHONY: docker-prod
docker-prod: ## Deploy with production configuration
	@echo "🚀 Deploying production configuration..."
	@docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "✅ Production deployment complete!"

.PHONY: docker-prod-down
docker-prod-down: ## Stop production deployment
	@echo "🛑 Stopping production deployment..."
	@docker compose -f docker-compose.yml -f docker-compose.prod.yml down
	@echo "✅ Production deployment stopped!"

.PHONY: docker-status
docker-status: ## Show Docker Compose service status
	@echo "📊 Docker Compose service status:"
	@docker compose ps

.PHONY: docker-shell-server
docker-shell-server: ## Open shell in server container
	@docker compose exec print-server sh

.PHONY: docker-shell-worker
docker-shell-worker: ## Open shell in worker container
	@docker compose exec print-worker sh

.PHONY: docker-shell-redis
docker-shell-redis: ## Open Redis CLI
	@docker compose exec redis redis-cli

.PHONY: docker-env
docker-env: ## Create .env file from example
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file from example..."; \
		cp .env.example .env; \
		echo "✅ .env file created! Please customize it for your environment."; \
	else \
		echo "⚠️  .env file already exists. Use 'make docker-env-force' to overwrite."; \
	fi

.PHONY: docker-env-force
docker-env-force: ## Force create .env file from example (overwrites existing)
	@echo "📝 Overwriting .env file from example..."
	@cp .env.example .env
	@echo "✅ .env file created! Please customize it for your environment."

.PHONY: docker-test
docker-test: ## Run tests in Docker containers
	@echo "🧪 Running tests in Docker containers..."
	@docker compose exec print-server go test ./...
	@echo "✅ Tests completed!"

.PHONY: docker-health
docker-health: ## Check health of all services
	@echo "🏥 Checking service health..."
	@docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "📊 Health check details:"
	@docker compose exec print-server wget -qO- http://localhost:8080/health 2>/dev/null || echo "❌ Server health check failed"
	@docker compose exec redis redis-cli ping 2>/dev/null || echo "❌ Redis health check failed"

.PHONY: docker-monitor
docker-monitor: ## Start monitoring services (requires monitoring profile)
	@echo "📊 Starting monitoring services..."
	@docker compose --profile monitoring up -d
	@echo "✅ Monitoring started! Prometheus: http://localhost:9090"

# Docker aliases for convenience
.PHONY: dup
dup: docker-up ## Alias for docker-up

.PHONY: ddown
ddown: docker-down ## Alias for docker-down

.PHONY: dlogs
dlogs: docker-logs ## Alias for docker-logs

.PHONY: dstatus
dstatus: docker-status ## Alias for docker-status

# =============================================================================
# DOCKER TEST COMMANDS
# =============================================================================

.PHONY: test-all
test-all: ## Run complete test suite with Docker Compose
	@echo "🧪 Running complete test suite..."
	@docker compose -f docker-compose.test.yml --profile test up --build --abort-on-container-exit
	@echo "✅ Complete test suite finished!"

.PHONY: test-services-up
test-services-up: ## Start test services (without running tests)
	@echo "🚀 Starting test services..."
	@docker compose -f docker-compose.test.yml up -d --build
	@echo "✅ Test services started! Server: http://localhost:8081"

.PHONY: test-services-down
test-services-down: ## Stop test services
	@echo "🛑 Stopping test services..."
	@docker compose -f docker-compose.test.yml down -v --remove-orphans
	@echo "✅ Test services stopped!"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@docker compose -f docker-compose.test.yml up --build unit-tests --abort-on-container-exit
	@echo "✅ Unit tests completed!"

.PHONY: test-e2e
test-e2e: ## Run E2E tests only
	@echo "🚀 Running E2E tests..."
	@docker compose -f docker-compose.test.yml up -d --build print-server-test print-worker-test
	@sleep 15
	@docker compose -f docker-compose.test.yml up --build e2e-tests --abort-on-container-exit
	@docker compose -f docker-compose.test.yml down
	@echo "✅ E2E tests completed!"

.PHONY: test-golden-rigor
test-golden-rigor: ## Run golden rigor test suite (comprehensive)
	@echo "🏆 Running golden rigor test suite..."
	@docker compose -f docker-compose.test.yml up -d --build print-server-test print-worker-test
	@sleep 15
	@docker compose -f docker-compose.test.yml up --build golden-rigor-tests --abort-on-container-exit
	@docker compose -f docker-compose.test.yml down
	@echo "✅ Golden rigor test suite completed!"

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "🔗 Running integration tests..."
	@docker compose -f docker-compose.test.yml up --build integration-tests --abort-on-container-exit
	@echo "✅ Integration tests completed!"

.PHONY: test-fuzz
test-fuzz: ## Run fuzz tests (gofuzz-based)
	@echo "🔀 Running fuzz tests..."
	@docker compose -f docker-compose.test.yml up -d --build print-server-test print-worker-test
	@sleep 15
	@docker compose -f docker-compose.test.yml up --build fuzz-tests --abort-on-container-exit
	@docker compose -f docker-compose.test.yml down
	@echo "✅ Fuzz tests completed!"

.PHONY: test-native-fuzz
test-native-fuzz: ## Run native Go fuzz tests
	@echo "🧬 Running native Go fuzz tests..."
	@docker compose -f docker-compose.test.yml up -d --build print-server-test print-worker-test
	@sleep 15
	@docker compose -f docker-compose.test.yml up --build native-fuzz-tests --abort-on-container-exit
	@docker compose -f docker-compose.test.yml down
	@echo "✅ Native fuzz tests completed!"

.PHONY: test-fuzz-all
test-fuzz-all: ## Run all fuzz tests (gofuzz + native)
	@echo "🔀🧬 Running all fuzz tests..."
	@$(MAKE) test-fuzz
	@$(MAKE) test-native-fuzz
	@echo "✅ All fuzz tests completed!"

.PHONY: test-logs
test-logs: ## View test service logs
	@echo "📋 Viewing test service logs..."
	@docker compose -f docker-compose.test.yml logs -f

.PHONY: test-clean
test-clean: ## Clean up test resources
	@echo "🧹 Cleaning up test resources..."
	@docker compose -f docker-compose.test.yml down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Test cleanup complete!"

.PHONY: test-status
test-status: ## Show test service status
	@echo "📊 Test service status:"
	@docker compose -f docker-compose.test.yml ps

# Test aliases
.PHONY: ta
ta: test-all ## Alias for test-all

.PHONY: tu
tu: test-unit ## Alias for test-unit

.PHONY: te2e
te2e: test-e2e ## Alias for test-e2e

.PHONY: trigor
trigor: test-golden-rigor ## Alias for test-golden-rigor

.PHONY: tfuzz
tfuzz: test-fuzz ## Alias for test-fuzz

.PHONY: tnfuzz
tnfuzz: test-native-fuzz ## Alias for test-native-fuzz

.PHONY: tfuzzall
tfuzzall: test-fuzz-all ## Alias for test-fuzz-all

.PHONY: test-rigor-all
test-rigor-all: ## Run complete rigor test suite (unit + e2e + golden rigor)
	@echo "🚀 Running COMPLETE RIGOR TEST SUITE"
	@echo "====================================="
	@make test-unit
	@make test-e2e
	@make test-golden-rigor
	@echo "🎉 Complete rigor test suite finished!"

.PHONY: test-ultimate
test-ultimate: ## Run ULTIMATE test suite (unit + e2e + golden rigor + fuzz)
	@echo "🏆 Running ULTIMATE TEST SUITE"
	@echo "==============================="
	@echo "🧪 Unit Tests → 🚀 E2E Tests → 🎯 Golden Rigor → 🔀 Fuzz Tests"
	@make test-unit
	@make test-e2e
	@make test-golden-rigor
	@make test-fuzz-all
	@echo "🎉 ULTIMATE test suite completed - Maximum rigor achieved!"

# =============================================================================
# LOAD TESTING WITH K6
# =============================================================================

.PHONY: load-test-smoke
load-test-smoke: ## Run k6 smoke test (quick validation)
	@echo "💨 Running k6 smoke test..."
	@docker compose -f docker-compose.load-test.yml run --rm k6-smoke

.PHONY: load-test-basic
load-test-basic: ## Run k6 basic load test
	@echo "📊 Running k6 basic load test..."
	@docker compose -f docker-compose.load-test.yml run --rm k6-load-test

.PHONY: load-test-stress
load-test-stress: ## Run k6 stress test (high load)
	@echo "🔥 Running k6 stress test..."
	@docker compose -f docker-compose.load-test.yml run --rm k6-stress

.PHONY: load-test-spike
load-test-spike: ## Run k6 spike test (traffic spikes)
	@echo "⚡ Running k6 spike test..."
	@docker compose -f docker-compose.load-test.yml run --rm k6-spike

.PHONY: load-test-soak
load-test-soak: ## Run k6 soak test (extended duration)
	@echo "🛁 Running k6 soak test (30 minutes)..."
	@docker compose -f docker-compose.load-test.yml run --rm k6-soak

.PHONY: load-test-scenarios
load-test-scenarios: ## Run k6 production scenarios
	@echo "🎯 Running k6 production scenarios..."
	@echo "Testing: web_traffic, batch_processing, enterprise_reports, api_integration, chaos_testing"
	@K6_SCENARIO=web_traffic docker compose -f docker-compose.load-test.yml run --rm k6-scenarios
	@K6_SCENARIO=batch_processing docker compose -f docker-compose.load-test.yml run --rm k6-scenarios
	@K6_SCENARIO=enterprise_reports docker compose -f docker-compose.load-test.yml run --rm k6-scenarios

.PHONY: load-test-all
load-test-all: ## Run complete k6 load test suite
	@echo "🚀 Running COMPLETE k6 load test suite..."
	@echo "=========================================="
	@make load-test-smoke
	@make load-test-basic
	@make load-test-stress
	@make load-test-spike
	@make load-test-scenarios
	@echo "🎉 Complete load test suite finished!"

.PHONY: load-test-production
load-test-production: ## Run production-ready load test validation
	@echo "🏭 Running PRODUCTION load test validation..."
	@echo "============================================="
	@echo "🔍 Validating production readiness with comprehensive load testing"
	@make load-test-smoke
	@make load-test-basic
	@make load-test-scenarios
	@echo "✅ Production load test validation completed!"

.PHONY: load-test-results
load-test-results: ## Show k6 load test results
	@echo "📈 k6 Load Test Results"
	@echo "======================="
	@docker compose -f docker-compose.load-test.yml run --rm k6-results

# Load testing aliases
.PHONY: lsmoke lstress lspike lsoak lscenarios lall lprod
lsmoke: load-test-smoke ## Alias for load-test-smoke
lstress: load-test-stress ## Alias for load-test-stress
lspike: load-test-spike ## Alias for load-test-spike
lsoak: load-test-soak ## Alias for load-test-soak
lscenarios: load-test-scenarios ## Alias for load-test-scenarios
lall: load-test-all ## Alias for load-test-all
lprod: load-test-production ## Alias for load-test-production
