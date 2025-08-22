#!/bin/bash

# Local Integration Test Script
# Mirrors the GitHub CI pipeline integration tests for local development
# Run this before pushing to catch issues early

set -euo pipefail

# Configuration
CLUSTER_NAME="kdebug-local-test"
KIND_CONFIG_FILE=$(mktemp)
KUBECONFIG_FILE=$(mktemp)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY_PATH="$PROJECT_ROOT/kdebug"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Cleanup function
cleanup() {
    local exit_code=$?
    log_info "Cleaning up test environment..."
    
    # Delete kind cluster
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        log_info "Deleting Kind cluster: $CLUSTER_NAME"
        kind delete cluster --name "$CLUSTER_NAME" || true
    fi
    
    # Clean up temporary files
    rm -f "$KIND_CONFIG_FILE" "$KUBECONFIG_FILE"
    
    # Remove test binary
    rm -f "$BINARY_PATH"
    
    if [ $exit_code -eq 0 ]; then
        log_success "Integration tests completed successfully!"
    else
        log_error "Integration tests failed with exit code: $exit_code"
    fi
    
    exit $exit_code
}

# Set up cleanup trap
trap cleanup EXIT INT TERM

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kind is installed
    if ! command -v kind &> /dev/null; then
        log_error "kind is not installed. Install with: go install sigs.k8s.io/kind@latest"
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl"
        exit 1
    fi
    
    # Check if docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker"
        exit 1
    fi
    
    log_success "All prerequisites are met"
}

# Create Kind cluster configuration
create_kind_config() {
    log_info "Creating Kind cluster configuration..."
    
    cat > "$KIND_CONFIG_FILE" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: $CLUSTER_NAME
nodes:
- role: control-plane
  image: kindest/node:v1.31.0
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
    
    log_success "Kind configuration created"
}

# Create and setup Kind cluster
setup_cluster() {
    log_info "Setting up Kind cluster: $CLUSTER_NAME"
    
    # Delete existing cluster if it exists
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        log_warning "Cluster $CLUSTER_NAME already exists, deleting..."
        kind delete cluster --name "$CLUSTER_NAME"
    fi
    
    # Create new cluster
    log_info "Creating Kind cluster..."
    kind create cluster --config "$KIND_CONFIG_FILE" --wait 5m
    
    # Export kubeconfig
    log_info "Exporting kubeconfig..."
    kind export kubeconfig --name "$CLUSTER_NAME" --kubeconfig "$KUBECONFIG_FILE"
    export KUBECONFIG="$KUBECONFIG_FILE"
    
    # Wait for cluster to be ready
    log_info "Waiting for cluster to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
    kubectl wait --for=condition=Ready pods --all -n kube-system --timeout=300s
    
    log_success "Kind cluster is ready"
}

# Build the binary
build_binary() {
    log_info "Building kdebug binary..."
    
    cd "$PROJECT_ROOT"
    go build -o "$BINARY_PATH" .
    
    if [ ! -f "$BINARY_PATH" ]; then
        log_error "Failed to build binary"
        exit 1
    fi
    
    log_success "Binary built successfully: $BINARY_PATH"
}

# Run unit tests first
run_unit_tests() {
    log_info "Running unit tests..."
    
    cd "$PROJECT_ROOT"
    go test -v ./... | tee /tmp/unit-test-results.log
    
    if [ ${PIPESTATUS[0]} -ne 0 ]; then
        log_error "Unit tests failed"
        exit 1
    fi
    
    log_success "Unit tests passed"
}

# Run linting
run_linting() {
    log_info "Running linting checks..."
    
    cd "$PROJECT_ROOT"
    
    # Run golangci-lint if available
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
        log_success "Linting passed"
    else
        log_warning "golangci-lint not found, skipping linting"
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    cd "$PROJECT_ROOT"
    export KUBECONFIG="$KUBECONFIG_FILE"
    
    # Build integration test binary
    go test -tags=integration -c ./test/integration -o integration.test
    
    if [ ! -f "integration.test" ]; then
        log_error "Failed to build integration test binary"
        exit 1
    fi
    
    # Run the integration tests
    log_info "Executing integration tests against Kind cluster..."
    ./integration.test -test.v -test.timeout=10m | tee /tmp/integration-test-results.log
    
    local test_exit_code=${PIPESTATUS[0]}
    
    # Clean up test binary
    rm -f integration.test
    
    if [ $test_exit_code -ne 0 ]; then
        log_error "Integration tests failed"
        cat /tmp/integration-test-results.log
        exit 1
    fi
    
    log_success "Integration tests passed"
}

# Test kdebug commands manually
test_kdebug_commands() {
    log_info "Testing kdebug commands manually..."
    
    export KUBECONFIG="$KUBECONFIG_FILE"
    
    # Test cluster command
    log_info "Testing 'kdebug cluster' command..."
    "$BINARY_PATH" cluster --output table || true
    
    # Create a test pod for pod diagnostics
    log_info "Creating test pod..."
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:1.21
    ports:
    - containerPort: 80
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
      limits:
        cpu: 100m
        memory: 128Mi
  restartPolicy: Never
EOF
    
    # Wait for pod to be ready
    kubectl wait --for=condition=Ready pod/test-pod --timeout=60s
    
    # Test pod command
    log_info "Testing 'kdebug pod' command..."
    "$BINARY_PATH" pod test-pod --output table
    "$BINARY_PATH" pod test-pod --output json | jq . > /dev/null
    "$BINARY_PATH" pod test-pod --output yaml | head -20
    
    # Test all pods
    log_info "Testing 'kdebug pod --all' command..."
    "$BINARY_PATH" pod --all --output table
    
    # Clean up test pod
    kubectl delete pod test-pod --ignore-not-found=true
    
    log_success "Manual kdebug commands testing completed"
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="$PROJECT_ROOT/test-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" <<EOF
KDEBUG LOCAL INTEGRATION TEST REPORT
====================================
Date: $(date)
Cluster: $CLUSTER_NAME
Binary: $BINARY_PATH

UNIT TEST RESULTS:
$(cat /tmp/unit-test-results.log | tail -20)

INTEGRATION TEST RESULTS:
$(cat /tmp/integration-test-results.log | tail -20)

CLUSTER INFO:
$(kubectl cluster-info)

NODE STATUS:
$(kubectl get nodes -o wide)

POD STATUS:
$(kubectl get pods --all-namespaces)
EOF
    
    log_success "Test report generated: $report_file"
}

# Main execution
main() {
    echo "========================================"
    echo "🧪 KDEBUG LOCAL INTEGRATION TEST SUITE"
    echo "========================================"
    echo
    
    log_info "Starting local integration test pipeline..."
    
    check_prerequisites
    create_kind_config
    setup_cluster
    build_binary
    run_unit_tests
    run_linting
    run_integration_tests
    test_kdebug_commands
    generate_report
    
    log_success "🎉 All tests completed successfully!"
    log_info "You can now safely push your changes to GitHub"
}

# Execute main function
main "$@"
