# kdebug Makefile

.PHONY: build clean test lint fmt vet install dev-deps help

# Binary name and version
BINARY_NAME=kdebug
VERSION=0.1.0-dev
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-X kdebug/cmd.Version=$(VERSION)"

# Default target
all: clean build

## Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Cross-platform binaries built in $(BUILD_DIR)/"

## Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run integration tests (requires cluster access)
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration -timeout=10m ./test/integration/...

## Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	$(GOTEST) -v -tags=e2e ./...

## Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

## Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## Run linting (requires golangci-lint)
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it with 'make dev-deps'" && exit 1)
	golangci-lint run

## Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest
	$(GOGET) mvdan.cc/gofumpt@latest
	@echo "Installing kind for integration tests..."
	$(GOGET) sigs.k8s.io/kind@latest

## Install the binary to $GOPATH/bin or /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ -n "$$GOPATH" ]; then \
		echo "Installing to $$GOPATH/bin/"; \
		mkdir -p "$$GOPATH/bin"; \
		cp $(BUILD_DIR)/$(BINARY_NAME) "$$GOPATH/bin/"; \
	elif [ -w "/usr/local/bin" ]; then \
		echo "Installing to /usr/local/bin/"; \
		cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/; \
	else \
		echo "Cannot install to system directories. Try:"; \
		echo "  sudo make install-system"; \
		echo "  or"; \
		echo "  make install-user"; \
		exit 1; \
	fi
	@echo "$(BINARY_NAME) installed successfully!"

## Install the binary to /usr/local/bin (requires sudo)
install-system: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin/ (requires sudo)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(BINARY_NAME) installed successfully!"

## Install the binary to ~/.local/bin (user-local)
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin/..."
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@echo "$(BINARY_NAME) installed to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your PATH. Add this to your shell profile:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

## Run the binary (for quick testing)
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) . && $(BUILD_DIR)/$(BINARY_NAME)

## Run cluster command (for testing)
run-cluster:
	@echo "Running $(BINARY_NAME) cluster..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) . && $(BUILD_DIR)/$(BINARY_NAME) cluster --verbose

## Setup test cluster with kind
test-cluster-setup:
	@echo "Setting up test cluster with kind..."
	@which kind > /dev/null || (echo "kind not found. Install it with 'make dev-deps'" && exit 1)
	kind create cluster --name kdebug-test --wait 60s
	@echo "Test cluster ready. Use: export KUBECONFIG=$$(kind get kubeconfig --name kdebug-test)"

## Cleanup test cluster
test-cluster-cleanup:
	@echo "Cleaning up test cluster..."
	kind delete cluster --name kdebug-test

## Run full test suite (unit + integration)
test-all: test test-integration

## Run security checks
security:
	@echo "Running security checks..."
	@which gosec > /dev/null || $(GOGET) github.com/securego/gosec/v2/cmd/gosec@latest
	$(shell go env GOPATH)/bin/gosec ./...
	@echo "Running vulnerability check..."
	@which govulncheck > /dev/null || $(GOGET) golang.org/x/vuln/cmd/govulncheck@latest
	$(shell go env GOPATH)/bin/govulncheck ./...

## Run all quality checks
quality: fmt vet lint security test-coverage

## Create release build
release: clean quality build-all
	@echo "Release build completed"

## Format code with gofumpt
fmt-gofumpt:
	@echo "Formatting code with gofumpt..."
	@which gofumpt > /dev/null || $(GOGET) mvdan.cc/gofumpt@latest
	$(shell go env GOPATH)/bin/gofumpt -w .

## Run pre-push validation
pre-push:
	@echo "Running pre-push validation..."
	./scripts/pre-push.sh

## Local integration testing
test-integration-local:
	@echo "Running local integration tests (full suite)..."
	./scripts/test-integration-local.sh

## Local integration testing (skip cluster tests - faster)
test-integration-local-skip:
	@echo "Running local integration tests (skipping cluster creation)..."
	SKIP_INTEGRATION_TESTS=true ./scripts/test-integration-local.sh

## Quick local testing
test-quick:
	@echo "Running quick local tests..."
	./scripts/test-quick-local.sh

## Test everything locally (quick + integration)
test-local-all: test-quick test-integration-local

## Check if local environment is ready for integration tests
check-integration-env:
	@echo "Checking integration test environment..."
	@command -v kind >/dev/null || (echo "❌ kind not found. Install with: go install sigs.k8s.io/kind@latest" && exit 1)
	@command -v kubectl >/dev/null || (echo "❌ kubectl not found. Please install kubectl" && exit 1)
	@docker info >/dev/null || (echo "❌ Docker not running. Please start Docker" && exit 1)
	@echo "✅ Integration test environment is ready"

## Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  build-all          - Build for all platforms"
	@echo "  test               - Run unit tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-e2e           - Run end-to-end tests"
	@echo "  test-all           - Run all tests (unit + integration)"
	@echo "  test-cluster-setup - Setup test cluster with kind"
	@echo "  test-cluster-cleanup - Cleanup test cluster"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run linters"
	@echo "  security           - Run security checks"
	@echo "  quality            - Run all quality checks"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Download dependencies"
	@echo "  dev-deps           - Install development dependencies"
	@echo "  install            - Install binary (auto-detect location)"
	@echo "  install-system     - Install binary to /usr/local/bin (requires sudo)"
	@echo "  install-user       - Install binary to ~/.local/bin"
	@echo "  run                - Build and run the binary"
	@echo "  run-cluster        - Build and run cluster command"
	@echo "  fmt-gofumpt        - Format code with gofumpt"
	@echo "  pre-push           - Run pre-push validation"
	@echo "  test-quick         - Run quick local tests (build, format, unit, lint)"
	@echo "  test-integration-local - Run full local integration tests with Kind"
	@echo "  test-local-all     - Run all local tests (quick + integration)"
	@echo "  check-integration-env - Check if integration test environment is ready"
	@echo "  release            - Create release build"
	@echo "  help               - Show this help"
