#!/bin/bash

# Pre-push validation script for kdebug
# This script runs all the checks that GitHub Actions will run
# to ensure changes pass CI before pushing.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Running pre-push validation for kdebug...${NC}\n"

# Function to print step headers
print_step() {
    echo -e "\n${BLUE}==== $1 ====${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
    print_error "Must be run from kdebug project root directory"
    exit 1
fi

# Step 1: Go version check
print_step "Checking Go version"
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MOD_VERSION=$(grep "^go " go.mod | awk '{print $2}')
echo "Local Go version: $GO_VERSION"
echo "go.mod requires: $GO_MOD_VERSION"

if [[ "$GO_VERSION" < "$GO_MOD_VERSION" ]]; then
    print_error "Local Go version ($GO_VERSION) is older than required ($GO_MOD_VERSION)"
    exit 1
fi
print_success "Go version check passed"

# Step 2: Clean and download dependencies
print_step "Cleaning and downloading dependencies"
go mod tidy
go mod download
print_success "Dependencies updated"

# Step 3: Format check
print_step "Checking code formatting"
if command -v gofumpt >/dev/null 2>&1; then
    if ! gofumpt -l . | grep -q .; then
        print_success "Code formatting is correct"
    else
        print_warning "Code formatting issues found, fixing..."
        gofumpt -w .
        print_success "Code formatting fixed"
    fi
else
    print_warning "gofumpt not found, installing..."
    go install mvdan.cc/gofumpt@latest
    $(go env GOPATH)/bin/gofumpt -w .
    print_success "Code formatting applied"
fi

# Step 4: Build check
print_step "Building project"
if go build -o bin/kdebug .; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi

# Step 5: Unit tests
print_step "Running unit tests"
if go test ./... -v; then
    print_success "All unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Step 6: Test coverage
print_step "Checking test coverage"
if go test ./... -coverprofile=coverage.out; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | sed 's/%//')
    echo "Coverage: ${COVERAGE}%"
    if (( $(echo "$COVERAGE >= 70" | bc -l) )); then
        print_success "Test coverage is adequate (${COVERAGE}%)"
    else
        print_warning "Test coverage is low (${COVERAGE}%), consider adding more tests"
    fi
    rm -f coverage.out
else
    print_error "Coverage check failed"
    exit 1
fi

# Step 7: Linting
print_step "Running linter"
# Always use the latest version from GOPATH
if [ -f "$(go env GOPATH)/bin/golangci-lint" ]; then
    if $(go env GOPATH)/bin/golangci-lint run --timeout=5m; then
        print_success "Linting passed"
    else
        print_error "Linting failed - please fix the issues above"
        exit 1
    fi
else
    print_warning "golangci-lint not found, installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    if $(go env GOPATH)/bin/golangci-lint run --timeout=5m; then
        print_success "Linting passed"
    else
        print_error "Linting failed - please fix the issues above"
        exit 1
    fi
fi

# Step 8: Security scan
print_step "Running security scan"
if [ -f "$(go env GOPATH)/bin/gosec" ]; then
    if $(go env GOPATH)/bin/gosec ./...; then
        print_success "Security scan passed"
    else
        print_warning "Security scan found issues (non-blocking)"
    fi
else
    print_warning "gosec not found, installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    if $(go env GOPATH)/bin/gosec ./...; then
        print_success "Security scan passed"
    else
        print_warning "Security scan found issues (non-blocking)"
    fi
fi

# Step 9: Vulnerability check
print_step "Running vulnerability check"
if [ -f "$(go env GOPATH)/bin/govulncheck" ]; then
    if $(go env GOPATH)/bin/govulncheck ./...; then
        print_success "Vulnerability check passed"
    else
        print_error "Vulnerability check failed"
        exit 1
    fi
else
    print_warning "govulncheck not found, installing..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
    if $(go env GOPATH)/bin/govulncheck ./...; then
        print_success "Vulnerability check passed"
    else
        print_error "Vulnerability check failed"
        exit 1
    fi
fi

# Step 10: Integration test readiness
print_step "Checking integration test readiness"
if command -v kind >/dev/null 2>&1; then
    print_success "kind is available for integration tests"
else
    print_warning "kind not found - integration tests may fail in CI"
    echo "Install with: go install sigs.k8s.io/kind@latest"
fi

if command -v kubectl >/dev/null 2>&1; then
    print_success "kubectl is available"
else
    print_warning "kubectl not found - integration tests may fail"
fi

# Step 11: Check for uncommitted changes
print_step "Checking for uncommitted changes"
if git diff --quiet && git diff --staged --quiet; then
    print_success "No uncommitted changes"
else
    print_warning "You have uncommitted changes"
    echo "Modified files:"
    git status --porcelain
    echo ""
    read -p "Continue with push? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Push cancelled"
        exit 1
    fi
fi

# Final summary
echo -e "\n${GREEN}🎉 All pre-push checks passed!${NC}"
echo -e "${GREEN}Your changes are ready to be pushed to GitHub.${NC}\n"

echo -e "${BLUE}To push your changes:${NC}"
echo -e "  git push origin <branch-name>"
echo -e "\n${BLUE}To run specific checks:${NC}"
echo -e "  make test          # Unit tests only"
echo -e "  make lint          # Linting only"  
echo -e "  make security      # Security checks only"
echo -e "  make quality       # All quality checks"
echo -e "  make test-all      # All tests including integration"
