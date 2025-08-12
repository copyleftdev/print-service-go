#!/bin/bash

# Print Service Test Runner
# Comprehensive test execution script for Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.test.yml"
TEST_TIMEOUT=300  # 5 minutes
LOG_FILE="test-results.log"

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Print Service Test Suite${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

print_step() {
    echo -e "${YELLOW}➤ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

cleanup() {
    print_step "Cleaning up test environment..."
    docker compose -f $COMPOSE_FILE down -v --remove-orphans 2>/dev/null || true
    print_success "Cleanup completed"
}

wait_for_services() {
    print_step "Waiting for services to be healthy..."
    local timeout=$TEST_TIMEOUT
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        if docker compose -f $COMPOSE_FILE ps --format json | jq -r '.[].Health' | grep -q "healthy"; then
            print_success "Services are healthy!"
            return 0
        fi
        sleep 5
        elapsed=$((elapsed + 5))
        echo -n "."
    done
    
    print_error "Services failed to become healthy within $timeout seconds"
    return 1
}

run_test_suite() {
    local test_type=$1
    print_step "Running $test_type tests..."
    
    case $test_type in
        "unit")
            docker compose -f $COMPOSE_FILE up --build unit-tests --abort-on-container-exit
            ;;
        "e2e")
            docker compose -f $COMPOSE_FILE up -d --build print-server-test print-worker-test
            wait_for_services
            docker compose -f $COMPOSE_FILE up --build e2e-tests --abort-on-container-exit
            ;;
        "integration")
            docker compose -f $COMPOSE_FILE up --build integration-tests --abort-on-container-exit
            ;;
        "all")
            docker compose -f $COMPOSE_FILE --profile test up --build --abort-on-container-exit
            ;;
        *)
            print_error "Unknown test type: $test_type"
            return 1
            ;;
    esac
}

show_usage() {
    echo "Usage: $0 [OPTIONS] [TEST_TYPE]"
    echo ""
    echo "TEST_TYPE:"
    echo "  unit         Run unit tests only"
    echo "  e2e          Run end-to-end tests only"
    echo "  integration  Run integration tests only"
    echo "  all          Run all tests (default)"
    echo ""
    echo "OPTIONS:"
    echo "  -c, --clean     Clean up before running tests"
    echo "  -k, --keep      Keep services running after tests"
    echo "  -l, --logs      Show logs after test completion"
    echo "  -s, --services  Start services only (no tests)"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 all              # Run all tests"
    echo "  $0 e2e --clean      # Clean and run E2E tests"
    echo "  $0 --services       # Start test services only"
}

# Parse command line arguments
CLEAN_BEFORE=false
KEEP_SERVICES=false
SHOW_LOGS=false
SERVICES_ONLY=false
TEST_TYPE="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--clean)
            CLEAN_BEFORE=true
            shift
            ;;
        -k|--keep)
            KEEP_SERVICES=true
            shift
            ;;
        -l|--logs)
            SHOW_LOGS=true
            shift
            ;;
        -s|--services)
            SERVICES_ONLY=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        unit|e2e|integration|all)
            TEST_TYPE=$1
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_header
    
    # Trap to ensure cleanup on exit
    trap cleanup EXIT
    
    # Clean up if requested
    if [ "$CLEAN_BEFORE" = true ]; then
        cleanup
    fi
    
    # Start logging
    exec > >(tee -a $LOG_FILE)
    exec 2>&1
    
    print_step "Starting test execution at $(date)"
    
    if [ "$SERVICES_ONLY" = true ]; then
        print_step "Starting test services only..."
        docker compose -f $COMPOSE_FILE up -d --build
        print_success "Test services started!"
        print_step "📊 Health check details:"
        docker compose exec print-server wget -qO- http://localhost:8080/health 2>/dev/null || echo "❌ Server health check failed"
        if [ "$KEEP_SERVICES" = true ]; then
            trap - EXIT  # Remove cleanup trap
            print_step "Services will keep running. Use 'make test-services-down' to stop."
        fi
        return 0
    fi
    
    # Run tests
    print_step "Building and starting test environment..."
    
    if run_test_suite $TEST_TYPE; then
        print_success "Test suite completed successfully!"
    else
        print_error "Test suite failed!"
        if [ "$SHOW_LOGS" = true ]; then
            print_step "Showing service logs..."
            docker compose -f $COMPOSE_FILE logs
        fi
        exit 1
    fi
    
    # Show logs if requested
    if [ "$SHOW_LOGS" = true ]; then
        print_step "Showing service logs..."
        docker compose -f $COMPOSE_FILE logs
    fi
    
    # Keep services running if requested
    if [ "$KEEP_SERVICES" = true ]; then
        trap - EXIT  # Remove cleanup trap
        print_step "Services will keep running. Use 'make test-services-down' to stop."
    fi
    
    print_success "Test execution completed at $(date)"
}

# Check if Docker Compose is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not available"
    exit 1
fi

# Run main function
main "$@"
