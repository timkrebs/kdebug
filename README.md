![kdebug logo](/assets/images/kdebug-logo.png)

# kdebug

[![Go Version](https://img.shields.io/github/go-mod/go-version/timkrebs/kdebug)](https://golang.org/)
[![License](https://img.shields.io/github/license/timkrebs/kdebug)](LICENSE)
[![Release](https://img.shields.io/github/v/release/timkrebs/kdebug)](https://github.com/timkrebs/kdebug/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/timkrebs/kdebug)](https://goreportcard.com/report/github.com/timkrebs/kdebug)
[![Contributors](https://img.shields.io/github/contributors/timkrebs/kdebug)](https://github.com/timkrebs/kdebug/graphs/contributors)
[![CI](https://github.com/timkrebs/kdebug/actions/workflows/ci.yml/badge.svg)](https://github.com/timkrebs/kdebug/actions/workflows/ci.yml)
[![CodeQL](https://github.com/timkrebs/kdebug/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/timkrebs/kdebug/actions/workflows/github-code-scanning/codeql)
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
- **CLI Framework** - Complete command structure with global options
- **Cluster Health Checks**
  - Node conditions (DiskPressure, MemoryPressure, NotReady)
  - Control plane availability and responsiveness
  - DNS (CoreDNS) health validation
  - Basic resource and connectivity diagnostics
- **Pod Diagnostics** ⭐ 
  - Pending pods → scheduling constraints, resource limits, node taints
  - Image pull errors and registry connectivity analysis
  - CrashLoopBackOff detection with intelligent log analysis
  - RBAC permission validation for pods and service accounts
  - Init container failures and dependency issues
  - Resource constraint analysis and QoS validation
  - Network connectivity and DNS configuration checks
- **Service Diagnostics** ⭐ 
  - Service configuration validation (ports, selectors, service types)
  - Endpoint health and backend pod availability
  - Service selector matching with available pods
  - DNS resolution testing for service names within the cluster
  - Load balancing and traffic distribution issues
  - Connectivity validation between services and pods
- **Ingress Diagnostics** ⭐ *New in v1.0.1!*
  - Ingress resource existence and configuration validation
  - Backend service availability and endpoint health checks
  - SSL/TLS certificate validation and expiration monitoring
  - Ingress controller discovery and status verification
  - Host routing and path configuration analysis
  - LoadBalancer IP assignment and connectivity testing

### 🔄 In Development

- **Advanced Networking**
  - Network policy validation and connectivity testing
  - LoadBalancer and NodePort service troubleshooting

- **DNS & CoreDNS**
  - Deploys ephemeral test pods for DNS validation
  - CoreDNS configuration and health checks
  - Upstream DNS connectivity testing

## Installation

### Prerequisites
- Kubernetes cluster access (local or remote)
- `kubectl` configured and working

### Install via Homebrew (Recommended for macOS)
```bash
# Add the kdebug tap
brew tap timkrebs/kdebug

# Install kdebug
brew install kdebug

# Or install directly in one command
brew install timkrebs/kdebug/kdebug

# Verify installation
kdebug --version
```

### Install from Release (Linux/Windows)
```bash
# Download the latest release for your platform
curl -LO https://github.com/timkrebs/kdebug/releases/latest/download/kdebug-linux-amd64
chmod +x kdebug-linux-amd64
sudo mv kdebug-linux-amd64 /usr/local/bin/kdebug

# Verify installation
kdebug version
```

### Install via Go
```bash
go install github.com/timkrebs/kdebug@latest
```

### Install via Docker
```bash
# Pull the latest image
docker pull timkrebs/kdebug:latest

# Run with your kubeconfig
docker run --rm -v ~/.kube:/root/.kube timkrebs/kdebug:latest cluster

# Or create an alias for easier use
alias kdebug='docker run --rm -v ~/.kube:/root/.kube timkrebs/kdebug:latest'
kdebug cluster --help
```

### Build from Source
```bash
git clone https://github.com/timkrebs/kdebug.git
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

5. **Diagnose ingress routing issues:**
   ```bash
   kdebug ingress <ingress-name> --namespace <namespace>
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
kdebug pod myapp-deployment-7d4b8c6f9-x8k2l --namespace production

# Debug all pods in a namespace
kdebug pod --all --namespace default

# Focus on specific diagnostic areas
kdebug pod myapp-pod --checks=basic,scheduling,images,rbac

# Include detailed log analysis for crashed pods
kdebug pod myapp-pod --include-logs --log-lines 50

# Watch pod status and re-run diagnostics on changes
kdebug pod myapp-pod --watch

# Export pod diagnostics to JSON
kdebug pod myapp-pod --output json
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

#### Ingress Diagnostics
```bash
# Debug a specific ingress resource
kdebug ingress myapp-ingress --namespace production

# Diagnose all ingress resources in a namespace
kdebug ingress --all --namespace default

# Diagnose ingress resources across all namespaces
kdebug ingress --all --all-namespaces

# Run specific checks only (existence, config, backends, endpoints, ssl)
kdebug ingress myapp-ingress --checks config,backends,ssl

# Focus on SSL/TLS certificate validation
kdebug ingress myapp-ingress --checks ssl

# Export ingress diagnostics to JSON for further analysis
kdebug ingress myapp-ingress --output json

# Use command aliases for convenience
kdebug ing myapp-ingress          # Short form
kdebug ingresses --all           # Plural form
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

### Ingress Diagnostics
```bash
$ kdebug ingress myapp-ingress --namespace production

🔍 Analyzing ingress: myapp-ingress (production)

┌─────────────────────────────────────────────────────────────┐
│                        KDEBUG REPORT                        │
├─────────────────────────────────────────────────────────────┤
│ Ingress: myapp-ingress                                      │
│ Namespace: production                                        │
│ Class: nginx                                                │
│ Created: 2h ago                                             │
└─────────────────────────────────────────────────────────────┘

✅ PASSED: Ingress resource exists and is accessible
✅ PASSED: Ingress configuration is valid

   📄 Details: 2 rules configured for hosts:
              • api.myapp.com → myapp-api-service:80
              • web.myapp.com → myapp-web-service:80

❌ FAILED: Backend service health

   📍 Issue: Service myapp-api-service has no ready endpoints
   📄 Details: Backend service exists but 0/3 pods are ready
              • myapp-api-deployment-abc123: CrashLoopBackOff
              • myapp-api-deployment-def456: ImagePullBackOff  
              • myapp-api-deployment-ghi789: Pending
   
   💡 Suggestions:
      1. Check backend pod status: kdebug pod --all -n production
      2. Verify service selector matches pod labels
      3. Ensure pods are healthy and ready

⚠️  WARNING: SSL certificate expires soon

   📍 Issue: TLS certificate expires in 7 days
   📄 Details: Certificate for *.myapp.com expires 2024-09-25
              Issued by: Let's Encrypt Authority X3
   
   💡 Suggestions:
      1. Renew certificate before expiration
      2. Check cert-manager auto-renewal configuration
      3. Verify ACME challenge configuration

📊 Summary: 2/4 checks passed, 1 warning
⏱️  Analysis completed in 4.1s
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
│   ├── ingress.go         # Ingress diagnostics and validation
│   └── dns.go             # DNS resolution testing
├── pkg/                   # Core business logic
│   ├── cluster/           # Cluster-level diagnostic checks
│   ├── pod/               # Pod-level diagnostic checks
│   ├── service/           # Service and endpoint checks
│   ├── ingress/           # Ingress resource diagnostics
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

# Local CI - Run all checks before pushing
make local-ci-quick      # Fast: tests + lint only
make local-ci            # Full: all GitHub Actions checks
make local-ci-verbose    # Full with detailed output
```

### Local CI Workflow

Run comprehensive CI checks locally in Docker to catch issues before pushing to GitHub:

```bash
# Quick validation during development (recommended)
make local-ci-quick

# Full CI validation before push
make local-ci

# Build CI Docker image
make local-ci-build
```

The local CI system provides:
- **Docker-based environment** that mirrors GitHub Actions
- **Fast feedback loop** to catch issues early
- **All CI checks**: tests, linting, security scanning, vulnerability checks
- **Flexible execution** with various flags and options

For detailed usage, see [Local CI Documentation](docs/local-ci.md).

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
- ✅ Comprehensive pod diagnostics (Pending, CrashLoopBackOff, scheduling, images, RBAC)
- ✅ Image pull error detection and registry connectivity analysis
- ✅ RBAC permission validation for pods and service accounts
- ✅ Basic DNS resolution testing and configuration validation
- ✅ Resource constraint analysis and QoS validation
- ✅ Cluster health checks (nodes, control plane, DNS)
- ✅ Service health and endpoint validation
- ✅ Ingress diagnostics and SSL/TLS validation (v1.0.1)

### 🎯 Version 0.2.0 - Core Features
- Enhanced cluster node condition monitoring
- Improved control plane health checks
- Enhanced output formatting and reporting
- Performance optimizations

### 🎯 Version 0.3.0 - Advanced Diagnostics
- Resource quota and limit analysis
- Network policy validation
- ConfigMap and Secret validation
- Advanced SSL certificate management

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

**Q: What ingress controllers are supported?**
A: kdebug works with any Kubernetes-compliant ingress controller including NGINX, Traefik, HAProxy, Istio Gateway, and cloud provider load balancers.

**Q: Can kdebug validate SSL certificates from external Certificate Authorities?**
A: Yes! kdebug can validate SSL/TLS certificates from any CA including Let's Encrypt, DigiCert, and internal certificate authorities.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**[⭐ Star this project](https://github.com/username/kdebug)** if you find it useful!

Made with ❤️ by the kdebug community

</div>

