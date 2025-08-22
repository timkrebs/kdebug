# Local Testing Guide

This guide explains how to run comprehensive local tests that mirror the GitHub CI pipeline, helping you catch issues before pushing code.

## Overview

We provide multiple testing levels to suit different development workflows:

1. **Quick Tests** - Fast validation (1-2 minutes)
2. **Integration Tests** - Full Kind cluster testing (5-10 minutes)
3. **Enhanced Pre-Push** - Configurable validation with options

## Prerequisites

### Required for All Tests
- Go 1.23+ 
- Git

### Required for Integration Tests
- [Docker](https://docs.docker.com/get-docker/) (running)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) - Install with: `go install sigs.k8s.io/kind@latest`

Check if your environment is ready:
```bash
make check-integration-env
```

## Testing Options

### 1. Quick Tests (Recommended for Development)

Fast validation that covers the core quality checks:

```bash
# Using make
make test-quick

# Or direct script
./scripts/test-quick-local.sh
```

**What it includes:**
- ✅ Build verification
- ✅ Code formatting check
- ✅ Unit tests
- ✅ Linting (golangci-lint)
- ✅ Security scan (gosec)

**Time:** 1-2 minutes

### 2. Integration Tests (Before Important Pushes)

Full integration testing with real Kubernetes cluster:

```bash
# Using make
make test-integration-local

# Or direct script
./scripts/test-integration-local.sh
```

**What it includes:**
- ✅ All quick tests
- ✅ Kind cluster creation
- ✅ Integration test suite
- ✅ Manual kdebug command testing
- ✅ Cluster cleanup
- ✅ Test report generation

**Time:** 5-10 minutes

### 3. Complete Local Testing

Run everything:

```bash
make test-local-all
```

### 4. Enhanced Pre-Push Hook

Configurable pre-push validation with multiple options:

```bash
# Quick tests only (default)
./scripts/pre-push-with-integration.sh

# Include integration tests
./scripts/pre-push-with-integration.sh --integration

# Interactive mode (asks if you want integration tests)
./scripts/pre-push-with-integration.sh --interactive

# Full testing
./scripts/pre-push-with-integration.sh --full

# Show all options
./scripts/pre-push-with-integration.sh --help
```

## Recommended Workflow

### For Daily Development
```bash
# Before each commit
make test-quick

# Before pushing small changes
./scripts/pre-push-with-integration.sh --interactive
```

### For Feature Completion
```bash
# Before pushing feature branches or important changes
make test-local-all
```

### For Release Preparation
```bash
# Full validation
./scripts/pre-push-with-integration.sh --full
```

## Integration Test Details

The integration test suite creates a temporary Kind cluster and runs:

### Test Categories
1. **Cluster Tests** - Basic connectivity and cluster diagnostics
2. **Pod Tests** - All pod diagnostic scenarios
3. **Output Format Tests** - JSON, YAML, table outputs
4. **Error Handling Tests** - Failure scenarios and edge cases

### Test Scenarios Covered
- ✅ Healthy pods
- ✅ Image pull failures
- ✅ Crash loop scenarios
- ✅ Resource constraint issues
- ✅ RBAC problems
- ✅ Network issues
- ✅ Init container failures

### Cluster Configuration
- **Type:** Kind (Kubernetes in Docker)
- **Version:** Latest stable
- **Nodes:** Single control-plane
- **Network:** Default CNI
- **Cleanup:** Automatic after tests

## Troubleshooting

### Common Issues

**"kind not found"**
```bash
go install sigs.k8s.io/kind@latest
```

**"kubectl not found"**
- Install kubectl: https://kubernetes.io/docs/tasks/tools/

**"Docker not running"**
- Start Docker Desktop or Docker daemon

**"Cluster creation timeout"**
- Ensure Docker has sufficient resources (4GB+ RAM recommended)
- Check Docker is running and healthy

**Integration tests fail locally but pass in CI**
- Check Go version matches go.mod requirements
- Ensure clean environment (no conflicting clusters)
- Run with verbose output for debugging

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Quick tests with verbose output
./scripts/test-quick-local.sh 2>&1 | tee test-debug.log

# Integration tests with verbose output  
./scripts/test-integration-local.sh 2>&1 | tee integration-debug.log
```

### Manual Integration Test Run

For detailed debugging:

```bash
# Build test binary
go test -tags=integration -c ./test/integration -o integration.test

# Create Kind cluster manually
kind create cluster --name debug-cluster --wait 5m

# Export kubeconfig
kind export kubeconfig --name debug-cluster

# Run specific test
./integration.test -test.v -test.run TestPodDiagnosticsIntegration

# Cleanup
kind delete cluster --name debug-cluster
rm integration.test
```

## CI Pipeline Alignment

These local tests closely mirror the GitHub Actions CI pipeline:

| CI Stage | Local Equivalent |
|----------|------------------|
| Lint | `make test-quick` |
| Unit Tests | `make test-quick` |
| Integration Tests | `make test-integration-local` |
| Security Scan | `make test-quick` |
| Build Verification | All test scripts |

## Performance Tips

1. **Use Quick Tests During Development** - Save integration tests for important milestones
2. **Parallel Development** - Use `--interactive` mode to choose when to run full tests
3. **Docker Resources** - Ensure Docker has adequate CPU/memory for Kind
4. **Clean Environment** - Regularly clean up Docker containers and Kind clusters

## Advanced Usage

### Custom Test Configuration

You can modify the test scripts for specific needs:

- **Custom cluster config**: Edit `scripts/test-integration-local.sh`
- **Additional checks**: Extend `scripts/test-quick-local.sh`
- **Different timeouts**: Modify timeout values in scripts

### Integration with Git Hooks

Replace the default pre-push hook:

```bash
# Install enhanced pre-push hook
ln -sf ../../scripts/pre-push-with-integration.sh .git/hooks/pre-push
chmod +x .git/hooks/pre-push
```

This will run the enhanced validation automatically on `git push`.

## Next Steps

1. **Start with Quick Tests** - Run `make test-quick` to validate your current environment
2. **Try Integration Tests** - Run `make check-integration-env` then `make test-integration-local`
3. **Set Up Your Workflow** - Choose the testing approach that fits your development style
4. **Configure Pre-Push** - Set up automatic validation that matches your preferences

The goal is to catch issues early and ensure your changes will pass the GitHub CI pipeline, making development smoother and more reliable.
