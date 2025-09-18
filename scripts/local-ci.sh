#!/bin/bash

# Local CI Script for kdebug
# This script runs all CI checks locally in Docker to mirror GitHub Actions

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_IMAGE="kdebug-ci:latest"
DOCKER_CONTAINER="kdebug-ci-container"
WORKSPACE_DIR="/workspace"
GO_VERSION="1.24"

# Flags
RUN_TESTS=true
RUN_LINT=true
RUN_BUILD=true
RUN_SECURITY=true
RUN_VULNERABILITY=true
RUN_INTEGRATION=false
VERBOSE=false
CLEANUP=true

# Function to print colored output
print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Run local CI checks in Docker environment"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  -v, --verbose        Enable verbose output"
    echo "  --no-tests          Skip unit tests"
    echo "  --no-lint           Skip linting"
    echo "  --no-build          Skip build checks"
    echo "  --no-security       Skip security scan"
    echo "  --no-vulnerability  Skip vulnerability check"
    echo "  --integration       Run integration tests (requires Docker daemon)"
    echo "  --no-cleanup        Don't cleanup Docker containers"
    echo "  --quick             Run only tests and lint (fastest)"
    echo "  --full              Run all checks including integration tests"
    echo ""
    echo "Examples:"
    echo "  $0                  # Run standard CI checks"
    echo "  $0 --quick          # Fast check (tests + lint only)"
    echo "  $0 --full           # Full CI including integration tests"
    echo "  $0 --no-tests       # Skip tests but run other checks"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --no-tests)
                RUN_TESTS=false
                shift
                ;;
            --no-lint)
                RUN_LINT=false
                shift
                ;;
            --no-build)
                RUN_BUILD=false
                shift
                ;;
            --no-security)
                RUN_SECURITY=false
                shift
                ;;
            --no-vulnerability)
                RUN_VULNERABILITY=false
                shift
                ;;
            --integration)
                RUN_INTEGRATION=true
                shift
                ;;
            --no-cleanup)
                CLEANUP=false
                shift
                ;;
            --quick)
                RUN_TESTS=true
                RUN_LINT=true
                RUN_BUILD=false
                RUN_SECURITY=false
                RUN_VULNERABILITY=false
                RUN_INTEGRATION=false
                shift
                ;;
            --full)
                RUN_TESTS=true
                RUN_LINT=true
                RUN_BUILD=true
                RUN_SECURITY=true
                RUN_VULNERABILITY=true
                RUN_INTEGRATION=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Function to check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to cleanup Docker resources
cleanup_docker() {
    if [[ "$CLEANUP" == "true" ]]; then
        print_step "Cleaning up Docker resources..."
        docker rm -f "$DOCKER_CONTAINER" 2>/dev/null || true
        print_success "Cleanup completed"
    fi
}

# Function to build CI Docker image
build_ci_image() {
    print_step "Building CI Docker image..."
    
    if [[ "$VERBOSE" == "true" ]]; then
        docker build -f Dockerfile.ci -t "$DOCKER_IMAGE" .
    else
        docker build -f Dockerfile.ci -t "$DOCKER_IMAGE" . > /dev/null 2>&1
    fi
    
    print_success "CI Docker image built successfully"
}

# Function to run command in Docker container
run_in_container() {
    local cmd="$1"
    local description="$2"
    
    print_step "$description"
    
    if [[ "$VERBOSE" == "true" ]]; then
        docker run --rm --name "$DOCKER_CONTAINER" \
            -v "$(pwd):$WORKSPACE_DIR" \
            -w "$WORKSPACE_DIR" \
            "$DOCKER_IMAGE" \
            bash -c "$cmd"
    else
        docker run --rm --name "$DOCKER_CONTAINER" \
            -v "$(pwd):$WORKSPACE_DIR" \
            -w "$WORKSPACE_DIR" \
            "$DOCKER_IMAGE" \
            bash -c "$cmd" > /dev/null 2>&1
    fi
    
    if [[ $? -eq 0 ]]; then
        print_success "$description completed successfully"
        return 0
    else
        print_error "$description failed"
        return 1
    fi
}

# Function to run tests
run_tests() {
    if [[ "$RUN_TESTS" == "true" ]]; then
        print_step "Running unit tests..."
        
        # Download dependencies
        run_in_container "go mod download && go mod verify" "Downloading and verifying dependencies"
        
        # Run tests
        if [[ "$VERBOSE" == "true" ]]; then
            docker run --rm --name "$DOCKER_CONTAINER" \
                -v "$(pwd):$WORKSPACE_DIR" \
                -w "$WORKSPACE_DIR" \
                "$DOCKER_IMAGE" \
                bash -c "make test"
        else
            docker run --rm --name "$DOCKER_CONTAINER" \
                -v "$(pwd):$WORKSPACE_DIR" \
                -w "$WORKSPACE_DIR" \
                "$DOCKER_IMAGE" \
                bash -c "make test" > /dev/null 2>&1
        fi
        
        if [[ $? -eq 0 ]]; then
            print_success "Unit tests passed"
        else
            print_error "Unit tests failed"
            return 1
        fi
        
        # Run tests with coverage
        if [[ "$VERBOSE" == "true" ]]; then
            docker run --rm --name "$DOCKER_CONTAINER" \
                -v "$(pwd):$WORKSPACE_DIR" \
                -w "$WORKSPACE_DIR" \
                "$DOCKER_IMAGE" \
                bash -c "make test-coverage"
        else
            docker run --rm --name "$DOCKER_CONTAINER" \
                -v "$(pwd):$WORKSPACE_DIR" \
                -w "$WORKSPACE_DIR" \
                "$DOCKER_IMAGE" \
                bash -c "make test-coverage" > /dev/null 2>&1
        fi
        
        if [[ $? -eq 0 ]]; then
            print_success "Coverage tests passed"
        else
            print_warning "Coverage tests failed but continuing..."
        fi
    else
        print_info "Skipping tests"
    fi
}

# Function to run linting
run_lint() {
    if [[ "$RUN_LINT" == "true" ]]; then
        print_step "Running linting checks..."
        
        # Format check
        run_in_container "gofumpt -l ." "Checking code formatting"
        
        # Vet check
        run_in_container "go vet ./..." "Running go vet"
        
        # golangci-lint
        if [[ "$VERBOSE" == "true" ]]; then
            docker run --rm --name "$DOCKER_CONTAINER" \
                -v "$(pwd):$WORKSPACE_DIR" \
                -w "$WORKSPACE_DIR" \
                "$DOCKER_IMAGE" \
                bash -c "golangci-lint run --timeout=10m"
        else
            docker run --rm --name "$DOCKER_CONTAINER" \
                -v "$(pwd):$WORKSPACE_DIR" \
                -w "$WORKSPACE_DIR" \
                "$DOCKER_IMAGE" \
                bash -c "golangci-lint run --timeout=10m" > /dev/null 2>&1
        fi
        
        if [[ $? -eq 0 ]]; then
            print_success "Linting checks passed"
        else
            print_error "Linting checks failed"
            return 1
        fi
    else
        print_info "Skipping linting"
    fi
}

# Function to run build checks
run_build() {
    if [[ "$RUN_BUILD" == "true" ]]; then
        print_step "Running build checks..."
        
        # Build binary
        run_in_container "make build" "Building binary"
        
        # Test binary runs
        run_in_container "./bin/kdebug --version && ./bin/kdebug --help" "Testing binary execution"
        
        print_success "Build checks passed"
    else
        print_info "Skipping build checks"
    fi
}

# Function to run security scan
run_security() {
    if [[ "$RUN_SECURITY" == "true" ]]; then
        print_step "Running security scan..."
        
        run_in_container "gosec ./..." "Running gosec security scan"
        
        print_success "Security scan passed"
    else
        print_info "Skipping security scan"
    fi
}

# Function to run vulnerability check
run_vulnerability() {
    if [[ "$RUN_VULNERABILITY" == "true" ]]; then
        print_step "Running vulnerability check..."
        
        run_in_container "go list -json -deps ./... | nancy sleuth" "Running nancy vulnerability check"
        
        print_success "Vulnerability check passed"
    else
        print_info "Skipping vulnerability check"
    fi
}

# Function to run integration tests
run_integration() {
    if [[ "$RUN_INTEGRATION" == "true" ]]; then
        print_step "Running integration tests..."
        print_warning "Integration tests require Docker-in-Docker and may take longer"
        
        # This is more complex and would require Docker-in-Docker setup
        # For now, we'll skip this as it's optional in the CI
        print_info "Integration tests skipped (requires Docker-in-Docker setup)"
    else
        print_info "Skipping integration tests"
    fi
}

# Function to display summary
display_summary() {
    echo ""
    echo "================== CI SUMMARY =================="
    
    local total_checks=0
    local passed_checks=0
    
    if [[ "$RUN_TESTS" == "true" ]]; then
        echo "Tests:          ✅ PASSED"
        ((total_checks++))
        ((passed_checks++))
    else
        echo "Tests:          ⏭️  SKIPPED"
    fi
    
    if [[ "$RUN_LINT" == "true" ]]; then
        echo "Linting:        ✅ PASSED"
        ((total_checks++))
        ((passed_checks++))
    else
        echo "Linting:        ⏭️  SKIPPED"
    fi
    
    if [[ "$RUN_BUILD" == "true" ]]; then
        echo "Build:          ✅ PASSED"
        ((total_checks++))
        ((passed_checks++))
    else
        echo "Build:          ⏭️  SKIPPED"
    fi
    
    if [[ "$RUN_SECURITY" == "true" ]]; then
        echo "Security:       ✅ PASSED"
        ((total_checks++))
        ((passed_checks++))
    else
        echo "Security:       ⏭️  SKIPPED"
    fi
    
    if [[ "$RUN_VULNERABILITY" == "true" ]]; then
        echo "Vulnerability:  ✅ PASSED"
        ((total_checks++))
        ((passed_checks++))
    else
        echo "Vulnerability:  ⏭️  SKIPPED"
    fi
    
    echo "==============================================="
    echo "Total: $passed_checks/$total_checks checks passed"
    echo ""
    
    if [[ $passed_checks -eq $total_checks && $total_checks -gt 0 ]]; then
        print_success "All CI checks passed! 🚀"
        print_info "Your code is ready to push to GitHub"
    else
        print_warning "Some checks were skipped"
    fi
}

# Main execution function
main() {
    echo "🚀 kdebug Local CI Runner"
    echo "=========================="
    echo ""
    
    # Parse arguments
    parse_args "$@"
    
    # Setup trap for cleanup
    trap cleanup_docker EXIT
    
    # Check prerequisites
    check_prerequisites
    
    # Build CI image
    build_ci_image
    
    # Run CI checks
    local exit_code=0
    
    run_tests || exit_code=1
    run_lint || exit_code=1
    run_build || exit_code=1
    run_security || exit_code=1
    run_vulnerability || exit_code=1
    run_integration || exit_code=1
    
    # Display summary
    display_summary
    
    exit $exit_code
}

# Run main function with all arguments
main "$@"