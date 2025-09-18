# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2024-09-18

### Added
- **Ingress Diagnostics**: Comprehensive ingress resource analysis and troubleshooting
  - Ingress existence and configuration validation
  - Backend service availability and endpoint health checks
  - SSL/TLS certificate validation and expiration monitoring
  - Ingress controller discovery and status verification
  - Host routing and path configuration analysis
  - LoadBalancer IP assignment and connectivity testing
- New command `kdebug ingress` with full CLI support
  - Support for single ingress analysis: `kdebug ingress my-ingress`
  - Bulk analysis: `kdebug ingress --all`
  - Cross-namespace analysis: `kdebug ingress --all --all-namespaces`
  - Selective checks: `kdebug ingress my-ingress --checks config,backends,ssl`
  - Multiple output formats: table, JSON, YAML
  - Command aliases: `ingress`, `ing`, `ingresses`
- Comprehensive test suite for ingress diagnostics
- Updated documentation with ingress examples and usage patterns

### Enhanced
- Service diagnostics improvements and stability fixes
- Better error handling and user feedback
- Improved output formatting consistency

### Documentation
- Updated README.md with ingress diagnostics features and examples
- Enhanced development guide with ingress testing scenarios
- Updated testing documentation with ingress-specific test cases
- Added comprehensive ingress diagnostic examples and use cases

## [1.0.0] - 2024-09-15

### Added
- Initial release of kdebug
- **Cluster Health Checks**
  - Node condition monitoring (DiskPressure, MemoryPressure, NotReady)
  - Control plane availability and responsiveness testing
  - DNS (CoreDNS) health validation
  - Basic resource and connectivity diagnostics
- **Pod Diagnostics**
  - Pending pod analysis (scheduling constraints, resource limits, node taints)
  - Image pull error detection and registry connectivity analysis
  - CrashLoopBackOff detection with intelligent log analysis
  - RBAC permission validation for pods and service accounts
  - Init container failure detection and dependency analysis
  - Resource constraint analysis and QoS validation
  - Network connectivity and DNS configuration checks
- **Service Diagnostics** 
  - Service configuration validation (ports, selectors, service types)
  - Endpoint health and backend pod availability checks
  - Service selector matching with available pods
  - DNS resolution testing for service names within clusters
  - Load balancing and traffic distribution issue detection
  - Connectivity validation between services and pods
- **CLI Framework**
  - Complete command structure with global options
  - Multiple output formats: table (default), JSON, YAML
  - Verbose mode for detailed debugging
  - Comprehensive help system and examples
- **Cross-platform Support**
  - Linux (AMD64, ARM64)
  - macOS (Intel, Apple Silicon)
  - Windows (AMD64)
- **Testing Infrastructure**
  - Comprehensive unit test suite
  - Integration testing with kind clusters
  - CI/CD pipeline with GitHub Actions
  - Code quality checks and security scanning

### Technical Features
- Built with Go 1.23+ for optimal performance
- Uses official Kubernetes client-go library
- Follows Kubernetes API conventions
- Implements least-privilege security model
- Zero external dependencies for core functionality

---

## Version History

- **v1.0.1** - Ingress diagnostics and enhanced service validation
- **v1.0.0** - Initial release with cluster, pod, and service diagnostics