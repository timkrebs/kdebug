#!/bin/bash

# Enhanced pre-push validation script with optional integration tests
# This script provides options for different levels of testing before push

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Help function
show_help() {
    cat << EOF
Enhanced Pre-Push Validation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -q, --quick             Run only quick tests (default)
    -i, --integration       Include integration tests with Kind cluster
    -f, --full              Run all tests including integration
    --skip-integration      Skip integration tests even if Docker/Kind available
    --interactive           Ask user whether to run integration tests

EXAMPLES:
    $0                      # Quick tests only
    $0 --quick              # Quick tests only
    $0 --integration        # Quick tests + integration tests
    $0 --full               # All tests
    $0 --interactive        # Ask user about integration tests

QUICK TESTS INCLUDE:
    - Code formatting check
    - Unit tests
    - Linting
    - Security scan
    - Build test

INTEGRATION TESTS INCLUDE:
    - Kind cluster creation
    - Integration test suite
    - Manual command testing
    - Cluster cleanup
EOF
}

# Parse command line arguments
RUN_INTEGRATION=false
INTERACTIVE=false
SKIP_INTEGRATION=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -q|--quick)
            RUN_INTEGRATION=false
            shift
            ;;
        -i|--integration)
            RUN_INTEGRATION=true
            shift
            ;;
        -f|--full)
            RUN_INTEGRATION=true
            shift
            ;;
        --skip-integration)
            SKIP_INTEGRATION=true
            shift
            ;;
        --interactive)
            INTERACTIVE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if integration tests are available
check_integration_available() {
    if [ "$SKIP_INTEGRATION" = true ]; then
        return 1
    fi
    
    command -v kind >/dev/null 2>&1 && \
    command -v kubectl >/dev/null 2>&1 && \
    docker info >/dev/null 2>&1
}

# Interactive prompt for integration tests
ask_for_integration() {
    if [ "$INTERACTIVE" = true ] && check_integration_available; then
        echo
        log_info "Integration test environment is available"
        echo -e "${YELLOW}Do you want to run integration tests? This will:${NC}"
        echo "  • Create a Kind cluster"
        echo "  • Run integration test suite"
        echo "  • Test kdebug commands"
        echo "  • Clean up cluster"
        echo "  • Takes ~5-10 minutes"
        echo
        read -p "Run integration tests? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            RUN_INTEGRATION=true
        fi
    fi
}

# Main execution
main() {
    echo "=================================================="
    echo "🚀 ENHANCED PRE-PUSH VALIDATION FOR KDEBUG"
    echo "=================================================="
    echo
    
    cd "$PROJECT_ROOT"
    
    # Show configuration
    log_info "Configuration:"
    echo "  • Quick tests: ✅ Always enabled"
    if [ "$RUN_INTEGRATION" = true ]; then
        echo "  • Integration tests: ✅ Enabled"
    elif [ "$SKIP_INTEGRATION" = true ]; then
        echo "  • Integration tests: ❌ Explicitly skipped"
    elif check_integration_available; then
        echo "  • Integration tests: ⏸️  Available but not selected"
    else
        echo "  • Integration tests: ❌ Environment not ready"
    fi
    echo
    
    # Ask interactively if needed
    ask_for_integration
    
    # Run quick tests first
    log_info "=== PHASE 1: QUICK VALIDATION ==="
    if ! ./scripts/test-quick-local.sh; then
        log_error "Quick tests failed! Fix these issues before proceeding."
        exit 1
    fi
    
    # Run integration tests if requested and available
    if [ "$RUN_INTEGRATION" = true ]; then
        if check_integration_available; then
            log_info "=== PHASE 2: INTEGRATION TESTING ==="
            if ! ./scripts/test-integration-local.sh; then
                log_error "Integration tests failed!"
                exit 1
            fi
        else
            log_error "Integration tests requested but environment not ready"
            log_info "Install requirements:"
            log_info "  • kind: go install sigs.k8s.io/kind@latest"
            log_info "  • kubectl: https://kubernetes.io/docs/tasks/tools/"
            log_info "  • Docker: https://docs.docker.com/get-docker/"
            exit 1
        fi
    fi
    
    # Final success message
    echo
    echo "=================================================="
    log_success "🎉 ALL VALIDATION CHECKS PASSED!"
    echo "=================================================="
    echo
    
    if [ "$RUN_INTEGRATION" = true ]; then
        log_success "✅ Quick tests passed"
        log_success "✅ Integration tests passed"
        log_info "Your changes have been thoroughly tested and are ready for GitHub!"
    else
        log_success "✅ Quick tests passed"
        if check_integration_available && [ "$SKIP_INTEGRATION" != true ]; then
            log_info "💡 Consider running integration tests for extra confidence:"
            log_info "   $0 --integration"
        fi
        log_info "Your changes are ready for GitHub!"
    fi
    
    echo
    log_info "Next steps:"
    echo "  git push origin <branch-name>"
    echo
}

# Execute main function
main "$@"
