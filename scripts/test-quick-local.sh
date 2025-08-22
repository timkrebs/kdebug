#!/bin/bash

# Quick Local Test Script
# Runs a subset of tests for faster development iteration
# Use this for quick validation before running the full integration suite

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Quick checks
check_build() {
    log_info "Testing build..."
    cd "$PROJECT_ROOT"
    go build -o kdebug-test . || { log_error "Build failed"; exit 1; }
    rm -f kdebug-test
    log_success "Build successful"
}

check_unit_tests() {
    log_info "Running unit tests..."
    cd "$PROJECT_ROOT"
    go test ./... || { log_error "Unit tests failed"; exit 1; }
    log_success "Unit tests passed"
}

check_linting() {
    log_info "Running linting..."
    cd "$PROJECT_ROOT"
    if command -v golangci-lint &> /dev/null; then
        # Try to run golangci-lint, but don't fail completely if config issues
        if golangci-lint run 2>&1; then
            log_success "Linting passed"
        else
            log_warning "Linting had issues, but continuing..."
            log_info "Consider updating golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
            # Fall back to basic go vet
            if go vet ./...; then
                log_success "Basic go vet passed"
            else
                log_error "go vet failed"
                exit 1
            fi
        fi
    else
        log_warning "golangci-lint not found, running go vet instead"
        if go vet ./...; then
            log_success "go vet passed"
        else
            log_error "go vet failed"
            exit 1
        fi
    fi
}

check_formatting() {
    log_info "Checking code formatting..."
    cd "$PROJECT_ROOT"
    
    # Check if gofumpt is available and use it
    if command -v gofumpt &> /dev/null; then
        if ! gofumpt -d . | grep -q .; then
            log_success "Code formatting is correct"
        else
            log_warning "Code formatting issues found. Run: gofumpt -w ."
            gofumpt -d .
        fi
    else
        # Fall back to gofmt
        if ! gofmt -d . | grep -q .; then
            log_success "Code formatting is correct"
        else
            log_warning "Code formatting issues found. Run: gofmt -w ."
            gofmt -d .
        fi
    fi
}

check_security() {
    log_info "Running security checks..."
    cd "$PROJECT_ROOT"
    if command -v gosec &> /dev/null; then
        gosec ./... || { log_error "Security issues found"; exit 1; }
        log_success "Security checks passed"
    else
        log_warning "gosec not found, skipping security scan"
    fi
}

# Main execution
main() {
    echo "================================"
    echo "🚀 KDEBUG QUICK LOCAL TEST"
    echo "================================"
    echo
    
    log_info "Running quick validation checks..."
    
    check_build
    check_formatting
    check_unit_tests
    check_linting
    check_security
    
    log_success "🎉 Quick tests completed successfully!"
    log_info "Run './scripts/test-integration-local.sh' for full integration tests"
}

main "$@"
