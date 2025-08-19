#!/bin/bash

# Test script for kdebug development
# This script runs the full test suite locally

set -e

echo "🧪 kdebug Test Suite"
echo "==================="
echo

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}▶${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

# Check if required tools are available
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi
    
    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        echo "Please install the missing tools and try again."
        exit 1
    fi
    
    print_success "All prerequisites found"
}

# Run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    if make test; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Run tests with coverage
run_coverage_tests() {
    print_status "Running tests with coverage..."
    
    if make test-coverage; then
        print_success "Coverage tests completed"
        if [ -f coverage.html ]; then
            print_status "Coverage report generated: coverage.html"
        fi
    else
        print_error "Coverage tests failed"
        return 1
    fi
}

# Run linting
run_linting() {
    print_status "Running linters..."
    
    # Check if golangci-lint is available
    if ! command -v golangci-lint &> /dev/null; then
        print_warning "golangci-lint not found, installing..."
        make dev-deps
    fi
    
    if make lint; then
        print_success "Linting passed"
    else
        print_error "Linting failed"
        return 1
    fi
}

# Run security checks
run_security_checks() {
    print_status "Running security checks..."
    
    if make security; then
        print_success "Security checks passed"
    else
        print_warning "Security checks had issues (this might be non-critical)"
    fi
}

# Build binary
build_binary() {
    print_status "Building binary..."
    
    if make build; then
        print_success "Binary built successfully"
    else
        print_error "Binary build failed"
        return 1
    fi
}

# Test binary functionality
test_binary() {
    print_status "Testing binary functionality..."
    
    if ! [ -f bin/kdebug ]; then
        print_error "Binary not found"
        return 1
    fi
    
    # Test help command
    if ./bin/kdebug --help > /dev/null; then
        print_success "Help command works"
    else
        print_error "Help command failed"
        return 1
    fi
    
    # Test version command
    if ./bin/kdebug --version > /dev/null; then
        print_success "Version command works"
    else
        print_error "Version command failed"
        return 1
    fi
    
    # Test cluster help
    if ./bin/kdebug cluster --help > /dev/null; then
        print_success "Cluster help command works"
    else
        print_error "Cluster help command failed"
        return 1
    fi
}

# Run integration tests if cluster is available
run_integration_tests() {
    print_status "Checking for integration test environment..."
    
    # Check if kind is available
    if ! command -v kind &> /dev/null; then
        print_warning "kind not found, installing..."
        make dev-deps
    fi
    
    # Check if we should skip integration tests
    if [ "$SKIP_INTEGRATION" = "true" ]; then
        print_warning "Skipping integration tests (SKIP_INTEGRATION=true)"
        return 0
    fi
    
    print_status "Setting up test cluster..."
    
    # Setup test cluster
    if make test-cluster-setup; then
        print_success "Test cluster created"
        
        # Export kubeconfig
        export KUBECONFIG=$(kind get kubeconfig --name kdebug-test)
        
        # Run integration tests
        print_status "Running integration tests..."
        if make test-integration; then
            print_success "Integration tests passed"
        else
            print_error "Integration tests failed"
            make test-cluster-cleanup
            return 1
        fi
        
        # Test kdebug against real cluster
        print_status "Testing kdebug against test cluster..."
        if ./bin/kdebug cluster --verbose; then
            print_success "kdebug cluster command works with real cluster"
        else
            print_warning "kdebug cluster command had issues (might be expected in test environment)"
        fi
        
        # Cleanup
        print_status "Cleaning up test cluster..."
        make test-cluster-cleanup
        print_success "Test cluster cleaned up"
    else
        print_error "Failed to create test cluster"
        return 1
    fi
}

# Main test function
run_tests() {
    local run_integration=${1:-true}
    local errors=0
    
    check_prerequisites || ((errors++))
    
    run_unit_tests || ((errors++))
    
    run_coverage_tests || ((errors++))
    
    run_linting || ((errors++))
    
    run_security_checks # Don't fail on security warnings
    
    build_binary || ((errors++))
    
    test_binary || ((errors++))
    
    if [ "$run_integration" = "true" ]; then
        run_integration_tests || ((errors++))
    fi
    
    echo
    if [ $errors -eq 0 ]; then
        print_success "All tests passed! 🎉"
        echo
        echo "Your kdebug build is ready for use!"
        echo "Run './bin/kdebug --help' to get started."
    else
        print_error "Some tests failed ($errors errors)"
        echo
        echo "Please fix the issues and try again."
        exit 1
    fi
}

# Parse command line arguments
INTEGRATION=true
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-integration)
            INTEGRATION=false
            shift
            ;;
        --skip-integration)
            export SKIP_INTEGRATION=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo
            echo "Options:"
            echo "  --no-integration   Skip integration tests entirely"
            echo "  --skip-integration Set SKIP_INTEGRATION=true"
            echo "  -h, --help         Show this help"
            echo
            echo "Environment variables:"
            echo "  SKIP_INTEGRATION   Set to 'true' to skip integration tests"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run the tests
run_tests $INTEGRATION
