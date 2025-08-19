# Testing Guide

This document describes the testing strategy and procedures for kdebug.

## Testing Strategy

kdebug uses a multi-layered testing approach:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test against real Kubernetes clusters
3. **End-to-End Tests** - Test complete user workflows
4. **Security Tests** - Scan for vulnerabilities and security issues
5. **Performance Tests** - Ensure acceptable performance characteristics

## Test Structure

```
test/
├── integration/           # Integration tests with real clusters
│   └── cluster_test.go   # Cluster command integration tests
├── fixtures/             # Test data and fixtures
├── e2e/                  # End-to-end tests
└── performance/          # Performance benchmarks

pkg/*/                    # Unit tests alongside source code
├── *_test.go            # Unit tests for each package
```

## Running Tests

### Prerequisites

- Go 1.23+
- Docker
- kubectl
- kind (for integration tests)

### Quick Test Run

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run all quality checks
make quality
```

### Full Test Suite

```bash
# Run comprehensive test suite
./scripts/test.sh

# Run without integration tests
./scripts/test.sh --no-integration
```

### Specific Test Types

#### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run tests for specific package
go test ./pkg/cluster/

# Run with verbose output
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Integration Tests

Integration tests require a real Kubernetes cluster. They use build tags to separate from unit tests.

```bash
# Setup test cluster
make test-cluster-setup

# Run integration tests
make test-integration

# Run specific integration test
go test -tags integration -v ./test/integration/ -run TestClusterDiagnostics

# Cleanup test cluster
make test-cluster-cleanup
```

#### Linting and Code Quality

```bash
# Install development dependencies
make dev-deps

# Run linters
make lint

# Format code
make fmt

# Run go vet
make vet

# Security scan
make security
```

## Integration Test Environment

### Using kind (Kubernetes in Docker)

The preferred method for integration testing is using kind:

```bash
# Create test cluster
kind create cluster --name kdebug-test

# Configure kubectl
export KUBECONFIG=$(kind get kubeconfig --name kdebug-test)

# Verify cluster
kubectl cluster-info
kubectl get nodes

# Run tests
go test -tags integration -v ./test/integration/

# Cleanup
kind delete cluster --name kdebug-test
```

### Using Existing Cluster

You can run integration tests against any Kubernetes cluster:

```bash
# Ensure kubectl is configured
kubectl cluster-info

# Run integration tests
export KUBECONFIG=/path/to/your/kubeconfig
go test -tags integration -v ./test/integration/
```

**⚠️ Warning**: Integration tests may modify cluster state. Use dedicated test clusters.

## Test Scenarios

### Normal Operation Tests

1. **Healthy Cluster**: Test against a fully functional cluster
2. **API Server Connectivity**: Various network conditions
3. **Output Formats**: JSON, YAML, table formats
4. **Command Line Options**: All flags and combinations

### Error Condition Tests

1. **No Cluster Access**: Test behavior when cluster is unreachable
2. **Insufficient Permissions**: Test RBAC limitations
3. **Partial Cluster Issues**: Node problems, DNS issues, etc.
4. **Network Issues**: Timeouts, intermittent connectivity

### Edge Cases

1. **Large Clusters**: Performance with many nodes
2. **Old Kubernetes Versions**: Compatibility testing
3. **Custom Resources**: Behavior with CRDs
4. **Mixed Architectures**: ARM and x86 nodes

## Writing Tests

### Unit Test Guidelines

```go
func TestComponentName(t *testing.T) {
    // Setup
    testData := createTestData()
    
    // Execute
    result := componentFunction(testData)
    
    // Verify
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Integration Test Guidelines

```go
//go:build integration
// +build integration

func TestIntegrationScenario(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Ensure test cluster
    ensureTestCluster(t)
    defer cleanupTestCluster(t)
    
    // Test logic here
}
```

### Test Data and Fixtures

Use the `test/fixtures/` directory for test data:

```go
func loadTestFixture(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("../../test/fixtures", filename))
    if err != nil {
        t.Fatalf("Failed to load fixture %s: %v", filename, err)
    }
    return data
}
```

## Continuous Integration

### GitHub Actions Workflows

The project uses GitHub Actions for CI/CD:

- **CI Workflow** (`.github/workflows/ci.yml`)
  - Runs on every push and PR
  - Unit tests, linting, security scans
  - Integration tests with kind cluster
  - Cross-platform builds

- **Release Workflow** (`.github/workflows/release.yml`)
  - Triggered on version tags
  - Full test suite
  - Multi-platform builds
  - Container image publishing
  - GitHub release creation

### Test Matrix

CI tests across:
- Go versions: 1.19, 1.20, 1.21
- Operating systems: Linux, macOS, Windows
- Kubernetes versions: 1.25, 1.26, 1.27, 1.28

## Performance Testing

### Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkClusterDiagnostics ./pkg/cluster/

# Memory profiling
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Performance Criteria

- Cluster diagnostics should complete in < 30 seconds for typical clusters
- Memory usage should remain < 100MB for normal operations
- Binary size should be < 50MB for release builds

## Test Coverage

### Coverage Targets

- **Unit Tests**: > 80% line coverage
- **Integration Tests**: Cover all major user workflows
- **Critical Paths**: 100% coverage for error handling

### Generating Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View in browser
open coverage.html

# Coverage by package
go tool cover -func=coverage.out
```

## Mock and Fake Objects

### Kubernetes Fake Client

Use Kubernetes fake client for unit tests:

```go
import (
    "k8s.io/client-go/kubernetes/fake"
    "k8s.io/apimachinery/pkg/runtime"
)

func TestWithFakeClient(t *testing.T) {
    // Create fake objects
    nodes := &corev1.NodeList{...}
    
    // Create fake client
    fakeClient := fake.NewSimpleClientset(nodes)
    
    // Use in tests
    diagnostic := NewDiagnostic(fakeClient)
    result := diagnostic.CheckNodes()
}
```

## Debugging Tests

### Verbose Output

```bash
# Verbose test output
go test -v ./...

# Show test names only
go test -v ./... | grep "^=== RUN"

# Failed tests only
go test ./... | grep FAIL
```

### Test-Specific Debugging

```bash
# Run single test
go test -run TestSpecificFunction ./pkg/cluster/

# Debug with delve
dlv test -- -test.run TestSpecificFunction

# Print debugging info
go test -v ./... -args -debug
```

### Environment Variables

Control test behavior with environment variables:

```bash
# Skip integration tests
export SKIP_INTEGRATION=true

# Keep test cluster after tests
export KEEP_TEST_CLUSTER=true

# Debug mode
export DEBUG=true

# Custom kubeconfig
export KUBECONFIG=/path/to/test/kubeconfig
```

## Best Practices

### Test Organization

1. **Group Related Tests**: Use subtests for related scenarios
2. **Clear Test Names**: Describe what is being tested
3. **Independent Tests**: Each test should be self-contained
4. **Fast Unit Tests**: Keep unit tests fast (< 1 second each)
5. **Comprehensive Integration Tests**: Cover real-world scenarios

### Test Data Management

1. **Use Fixtures**: Store test data in files
2. **Generate Data**: Create test data programmatically when possible
3. **Clean State**: Each test should start with clean state
4. **Realistic Data**: Use data that resembles production

### Error Testing

1. **Test Error Paths**: Every error condition should be tested
2. **Meaningful Errors**: Verify error messages are helpful
3. **Error Types**: Test different types of failures
4. **Recovery**: Test error recovery mechanisms

## Troubleshooting Tests

### Common Issues

1. **Cluster Not Ready**: Wait for cluster to be fully ready
2. **Permission Denied**: Check RBAC permissions
3. **Resource Not Found**: Ensure test resources exist
4. **Timeout Issues**: Increase timeouts for slow environments

### Debug Commands

```bash
# Check cluster status
kubectl cluster-info
kubectl get nodes
kubectl get pods -A

# Check kdebug output
./bin/kdebug cluster --verbose --output json

# Check test logs
go test -v ./test/integration/ 2>&1 | tee test.log
```

## Contributing Tests

When contributing to kdebug:

1. **Add Tests**: New features must include tests
2. **Update Existing Tests**: Modify tests when changing behavior
3. **Integration Tests**: Add integration tests for new commands
4. **Documentation**: Update test documentation
5. **CI Compatibility**: Ensure tests work in CI environment

### Test Checklist

Before submitting a PR:

- [ ] Unit tests added/updated
- [ ] Integration tests added for new features
- [ ] All tests pass locally
- [ ] Coverage doesn't decrease
- [ ] Tests are documented
- [ ] CI pipeline passes
