# Development Guide

## Getting Started

### Prerequisites

- Go 1.19 or later
- Docker (for local testing with kind)
- kubectl
- make

### Building kdebug

```bash
# Clone the repository
git clone https://github.com/username/kdebug.git
cd kdebug

# Download dependencies
make deps

# Build the binary
make build

# Run tests
make test
```

### Testing with a Local Cluster

For development and testing, you can use [kind](https://kind.sigs.k8s.io/) to create a local Kubernetes cluster:

#### 1. Install kind

```bash
# On macOS
brew install kind

# On Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

#### 2. Create a test cluster

```bash
# Create a simple cluster
kind create cluster --name kdebug-test

# Or create a multi-node cluster for more realistic testing
cat <<EOF | kind create cluster --name kdebug-test --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
EOF
```

#### 3. Test kdebug

```bash
# Verify cluster is running
kubectl cluster-info --context kind-kdebug-test

# Test kdebug
make run-cluster

# Or run directly
./bin/kdebug cluster --verbose
```

#### 4. Test different scenarios

```bash
# Test with JSON output
./bin/kdebug cluster --output json

# Test with YAML output
./bin/kdebug cluster --output yaml

# Test nodes-only mode
./bin/kdebug cluster --nodes-only

# Test with different timeout
./bin/kdebug cluster --timeout 10s
```

#### 5. Simulate issues for testing

```bash
# Drain a node to test node issues
kubectl drain kind-kdebug-test-worker --ignore-daemonsets --delete-emptydir-data

# Scale down CoreDNS to test DNS issues
kubectl scale deployment coredns --replicas=0 -n kube-system

# Test kdebug with these issues
./bin/kdebug cluster --verbose

# Restore cluster
kubectl scale deployment coredns --replicas=2 -n kube-system
kubectl uncordon kind-kdebug-test-worker
```

#### 6. Cleanup

```bash
# Delete the test cluster
kind delete cluster --name kdebug-test
```

## Development Workflow

### Adding New Checks

1. Create a new check function in the appropriate package (e.g., `pkg/cluster/`)
2. Add the check to the diagnostic runner
3. Write tests for the new check
4. Update documentation

### Code Style

- Follow Go conventions
- Use `gofmt` for formatting: `make fmt`
- Run linters: `make lint`
- Ensure tests pass: `make test`

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linting
make lint

# Run go vet
make vet
```

### Building for Different Platforms

```bash
# Build for all platforms
make build-all

# This creates binaries for:
# - Linux AMD64/ARM64
# - macOS AMD64/ARM64 (Apple Silicon)
# - Windows AMD64
```

## Project Structure

```
kdebug/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and global flags
│   ├── cluster.go         # Cluster health checks
│   ├── pod.go             # Pod diagnostics (future)
│   └── service.go         # Service checks (future)
├── pkg/                   # Core business logic
│   ├── cluster/           # Cluster-level checks
│   ├── pod/               # Pod-level checks (future)
│   └── service/           # Service checks (future)
├── internal/              # Private application code
│   ├── client/            # Kubernetes client utilities
│   └── output/            # Output formatting
├── docs/                  # Documentation
├── test/                  # Test files and fixtures
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── main.go               # Application entrypoint
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.
