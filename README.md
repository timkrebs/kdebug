![kdebug logo](/assets/images/kdebug-logo.png)

# kdebug

[![Go Version](https://img.shields.io/github/go-mod/go-version/username/kdebug)](https://golang.org/)
[![License](https://img.shields.io/github/license/username/kdebug)](LICENSE)
[![Release](https://img.shields.io/github/v/release/username/kdebug)](https://github.com/username/kdebug/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/username/kdebug)](https://goreportcard.com/report/github.com/username/kdebug)
[![Contributors](https://img.shields.io/github/contributors/username/kdebug)](https://github.com/username/kdebug/graphs/contributors)

> 🩺 A CLI tool that automatically diagnoses common Kubernetes issues and provides actionable suggestions.

Think of it as a "doctor" for Kubernetes clusters (like `brew doctor`, but for K8s). Get instant insights into what's wrong with your workloads and how to fix them.

## Table of Contents

- [Vision](#vision)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Example Output](#example-output)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Development](#development)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)

## Vision

`kdebug` helps developers and operators quickly identify **why their workloads are not running as expected**.  
Instead of manually digging through `kubectl describe` outputs and events, `kdebug` runs a series of checks and gives **clear guidance**.


## Features

### 🚀 Currently Available
- Basic project structure and CLI framework

### 🔄 In Development
- **Cluster Health Checks**
  - Node conditions (DiskPressure, MemoryPressure, NotReady)
  - Control plane availability and responsiveness
  - Resource quotas and limits analysis

- **Pod Diagnostics**
  - Pending pods → insufficient resources, taints, affinity issues
  - Image pull errors and registry connectivity
  - CrashLoopBackOff detection with log analysis
  - RBAC permission validation
  - Init container failures

- **Service & Networking**
  - Selector mismatches (no pods backing the service)
  - Endpoints creation and health
  - DNS resolution testing inside the cluster
  - Ingress configuration validation

- **DNS & CoreDNS**
  - Deploys ephemeral test pods for DNS validation
  - CoreDNS configuration and health checks
  - Upstream DNS connectivity testing

## Installation

### Prerequisites
- Kubernetes cluster access (local or remote)
- `kubectl` configured and working
- Go 1.23+ (for building from source)

### Install from Release (Recommended)
```bash
# Download the latest release for your platform
curl -LO https://github.com/username/kdebug/releases/latest/download/kdebug-linux-amd64
chmod +x kdebug-linux-amd64
sudo mv kdebug-linux-amd64 /usr/local/bin/kdebug

# Verify installation
kdebug version
```

### Install via Go
```bash
go install github.com/username/kdebug@latest
```

### Build from Source
```bash
git clone https://github.com/username/kdebug.git
cd kdebug
make build
sudo cp bin/kdebug /usr/local/bin/
```

## Quick Start

1. **Verify your cluster connection:**
   ```bash
   kubectl cluster-info
   ```

2. **Run a quick cluster health check:**
   ```bash
   kdebug cluster
   ```

3. **Debug a specific pod:**
   ```bash
   kdebug pod <pod-name> --namespace <namespace>
   ```

4. **Check service connectivity:**
   ```bash
   kdebug service <service-name> --namespace <namespace>
   ```

## Usage

### Global Options
```bash
kdebug [command] [options]

Global Flags:
  -h, --help                Help for kdebug
  -n, --namespace string    Kubernetes namespace (default "default")
  -o, --output string       Output format: table, json, yaml (default "table")
  -v, --verbose             Verbose output for debugging
      --kubeconfig string   Path to kubeconfig file
```

### Commands

#### Cluster Diagnostics
```bash
# Run comprehensive cluster health checks
kdebug cluster

# Check only node conditions
kdebug cluster --nodes-only

# Export results to JSON
kdebug cluster --output json > cluster-report.json
```

#### Pod Diagnostics
```bash
# Debug a specific pod
kdebug pod myapp-123 --namespace production

# Debug all pods in a namespace
kdebug pod --all --namespace default

# Focus on specific checks
kdebug pod myapp-123 --checks=resources,rbac,dns
```

#### Service Diagnostics
```bash
# Debug a service and its endpoints
kdebug service myservice --namespace default

# Include DNS resolution testing
kdebug service myservice --test-dns

# Check service across all namespaces
kdebug service --all-namespaces
```

#### DNS Diagnostics
```bash
# Test DNS resolution in the cluster
kdebug dns

# Test specific DNS queries
kdebug dns --query kubernetes.default.svc.cluster.local

# Test external DNS
kdebug dns --external google.com
```

## Example Output

### Pod Diagnostics
```bash
$ kdebug pod myapp-deployment-7d4b8c6f9-x8k2l --namespace production

🔍 Analyzing pod: myapp-deployment-7d4b8c6f9-x8k2l (production)

┌─────────────────────────────────────────────────────────────┐
│                        KDEBUG REPORT                        │
├─────────────────────────────────────────────────────────────┤
│ Pod: myapp-deployment-7d4b8c6f9-x8k2l                      │
│ Namespace: production                                        │
│ Status: Pending                                             │
│ Created: 5m ago                                             │
└─────────────────────────────────────────────────────────────┘

✅ PASSED: Pod exists and is accessible
✅ PASSED: RBAC permissions are correctly configured
❌ FAILED: Pod scheduling

   📍 Issue: Insufficient CPU resources
   📄 Details: Pod requires 2 CPU cores, but nodes have:
              • node-1: 0.5 CPU available
              • node-2: 1.2 CPU available  
              • node-3: 0.8 CPU available
   
   💡 Suggestions:
      1. Reduce CPU requests in deployment spec
      2. Scale up node pool or add larger nodes
      3. Check if other pods can be optimized

❌ FAILED: DNS resolution test

   📍 Issue: CoreDNS pods are not healthy
   📄 Details: 2/2 CoreDNS pods in CrashLoopBackOff state
   
   💡 Suggestions:
      1. Check CoreDNS logs: kubectl logs -n kube-system coredns-xxx
      2. Verify upstream DNS configuration
      3. Restart CoreDNS: kubectl rollout restart -n kube-system deployment/coredns

📊 Summary: 2/4 checks passed
⏱️  Analysis completed in 3.2s
```

### Cluster Health Overview
```bash
$ kdebug cluster --output table

┌─────────────────┬────────┬─────────────────────────────────┐
│ COMPONENT       │ STATUS │ DETAILS                         │
├─────────────────┼────────┼─────────────────────────────────┤
│ API Server      │   ✅   │ Healthy, response time: 45ms    │
│ etcd            │   ✅   │ 3/3 members healthy             │
│ Scheduler       │   ✅   │ Active and responsive           │
│ Controller Mgr  │   ✅   │ Active and responsive           │
│ Node Pool       │   ⚠️   │ 2/5 nodes under memory pressure│
│ DNS             │   ❌   │ CoreDNS pods failing            │
│ Ingress         │   ✅   │ NGINX controller healthy        │
└─────────────────┴────────┴─────────────────────────────────┘

🎯 Priority Actions:
   1. Investigate CoreDNS failures (blocking pod DNS)
   2. Add memory to nodes or evict low-priority workloads
```

## Tech Stack

- **Core Language:** Go 1.23+
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra) - Used by Kubernetes, Helm, and other CNCF projects
- **Kubernetes Client:** [client-go](https://github.com/kubernetes/client-go) - Official Kubernetes API client
- **Output Formats:** Table (default), JSON, YAML
- **Testing:** Go testing framework with Kubernetes test utilities
- **Build:** Make and GoReleaser for cross-platform builds

## Project Structure

```
kdebug/
├── cmd/                    # CLI command definitions (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── cluster.go         # Cluster health checks
│   ├── pod.go             # Pod diagnostics
│   ├── service.go         # Service and networking checks
│   └── dns.go             # DNS resolution testing
├── pkg/                   # Core business logic
│   ├── cluster/           # Cluster-level diagnostic checks
│   ├── pod/               # Pod-level diagnostic checks
│   ├── service/           # Service and endpoint checks
│   ├── dns/               # DNS resolution testing
│   └── types/             # Shared types and interfaces
├── internal/              # Private application code
│   ├── client/            # Kubernetes client initialization
│   ├── output/            # Output formatting (table, JSON, YAML)
│   ├── logger/            # Structured logging
│   └── config/            # Configuration management
├── test/                  # Integration and e2e tests
├── docs/                  # Documentation and examples
├── scripts/               # Build and development scripts
├── .github/               # GitHub workflows and templates
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build and development tasks
└── main.go               # Application entrypoint
```

## Development

### Prerequisites
- Go 1.23 or later
- Access to a Kubernetes cluster (local or remote)
- `kubectl` configured and working

### Getting Started
```bash
# Clone the repository
git clone https://github.com/username/kdebug.git
cd kdebug

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run with your local cluster
./bin/kdebug cluster
```

### Development Commands
```bash
# Build for all platforms
make build-all

# Run linting
make lint

# Run tests with coverage
make test-coverage

# Clean build artifacts
make clean

# Install development dependencies
make dev-deps
```

### Running Tests
```bash
# Unit tests
go test ./...

# Integration tests (requires cluster access)
make test-integration

# End-to-end tests
make test-e2e
```

## Roadmap

### 🎯 Version 0.1.0 - MVP (Current Sprint)
- ✅ Project structure and CLI framework
- 🔄 Basic pod health checks (Pending, CrashLoopBackOff)
- 🔄 Image pull error detection
- 🔄 RBAC permission validation
- 🔄 Basic DNS resolution testing

### 🎯 Version 0.2.0 - Core Features
- Service health and endpoint validation
- Cluster node condition monitoring
- Control plane health checks
- Enhanced output formatting and reporting

### 🎯 Version 0.3.0 - Advanced Diagnostics
- Resource quota and limit analysis
- Network policy validation
- Ingress controller checks
- ConfigMap and Secret validation

### 🎯 Version 1.0.0 - Production Ready
- Comprehensive test coverage
- Performance optimizations
- Pluggable check system for community contributions
- Multi-cluster support
- Web UI dashboard (optional)

### 🚀 Future Enhancements
- Integration with popular monitoring tools (Prometheus, Grafana)
- AI-powered issue prediction and resolution suggestions
- Custom check definitions via YAML
- Operator for continuous cluster monitoring

## Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, improving documentation, or sharing ideas, your help is appreciated.

### Ways to Contribute

- 🐛 **Report Bugs:** Use our [bug report template](.github/ISSUE_TEMPLATE/00_bug_report.yml)
- 📖 **Improve Documentation:** Use our [documentation template](.github/ISSUE_TEMPLATE/01_documentation.yml)
- ✨ **Suggest Enhancements:** Use our [enhancement template](.github/ISSUE_TEMPLATE/02_enhancement.yml)
- 🚀 **Request Features:** Use our [feature request template](.github/ISSUE_TEMPLATE/03_new_feature.yml)
- 💻 **Submit Code:** Fork, develop, and submit a pull request

### Development Workflow

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass** locally
6. **Submit a pull request** with a clear description

### Code Standards
- Follow Go conventions and `gofmt` formatting
- Write clear, self-documenting code with helpful comments
- Include unit tests for new features
- Update documentation for user-facing changes

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## Support

### Getting Help

- 📖 **Documentation:** Check this README and the [docs/](docs/) directory
- 🐛 **Bug Reports:** Use our [issue templates](.github/ISSUE_TEMPLATE/)
- 💬 **Discussions:** Start a [GitHub Discussion](https://github.com/username/kdebug/discussions)
- 📧 **Security Issues:** Email security@kdebug.dev for sensitive reports

### Community

- Follow us on [Twitter](https://twitter.com/kdebug_tool) for updates
- Join our [Slack community](https://join.slack.com/kdebug) for real-time chat
- Read our [blog](https://blog.kdebug.dev) for tutorials and insights

### FAQ

**Q: Does kdebug require special permissions in my cluster?**
A: kdebug only requires read permissions for the resources it analyzes. It follows the principle of least privilege.

**Q: Can I use kdebug with managed Kubernetes services?**
A: Yes! kdebug works with any Kubernetes cluster including EKS, GKE, AKS, and others.

**Q: How is kdebug different from kubectl?**
A: While kubectl provides raw data, kdebug analyzes that data to identify problems and suggest solutions automatically.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**[⭐ Star this project](https://github.com/username/kdebug)** if you find it useful!

Made with ❤️ by the kdebug community

</div>

