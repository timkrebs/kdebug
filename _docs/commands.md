---
layout: docs
title: Commands Reference
description: Complete reference for all kdebug commands and their options.
permalink: /docs/commands/
order: 3
---

# Commands Reference

kdebug provides several commands to help you debug Kubernetes clusters and pods. Each command is designed to identify specific types of issues and provide actionable suggestions.

## Global Flags

These flags are available for all commands:

| Flag | Description | Default |
|------|-------------|---------|
| `--kubeconfig` | Path to kubeconfig file | `$HOME/.kube/config` |
| `--namespace, -n` | Kubernetes namespace | `default` |
| `--output, -o` | Output format (table, json, yaml) | `table` |
| `--verbose, -v` | Enable verbose output | `false` |
| `--help, -h` | Show help for command | - |
| `--version` | Show version information | - |

## Commands

### `kdebug cluster`

Run cluster-wide health checks and diagnose common infrastructure issues.

#### Usage

```bash
kdebug cluster [flags]
```

#### Description

The cluster command analyzes your Kubernetes cluster infrastructure including:

- Node health and resource availability
- System pod status (DNS, networking, storage)
- Cluster networking and connectivity
- Resource quotas and limits
- Critical system services

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--node` | Analyze specific node | All nodes |
| `--check-networking` | Perform network connectivity tests | `true` |
| `--check-dns` | Test DNS resolution | `true` |
| `--timeout` | Timeout for cluster checks | `30s` |

#### Examples

```bash
# Run all cluster health checks
kdebug cluster

# Check specific node
kdebug cluster --node worker-node-1

# Verbose output with detailed analysis
kdebug cluster --verbose

# Quick check without network tests
kdebug cluster --check-networking=false
```

### `kdebug pod`

Diagnose pod-level issues and provide remediation suggestions.

#### Usage

```bash
kdebug pod [pod-name] [flags]
```

#### Description

The pod command analyzes pod-specific issues including:

- Pending pods (scheduling constraints, resource limits, node taints)
- Image pull errors and registry connectivity problems
- CrashLoopBackOff detection with log analysis and hints
- RBAC permission validation for pods and service accounts
- Init container failures and misconfigurations
- Resource constraints and quality of service issues

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--selector, -l` | Label selector to filter pods | - |
| `--all-namespaces, -A` | Check pods across all namespaces | `false` |
| `--since` | Show logs since this time (e.g., 1h, 30m) | `1h` |
| `--tail` | Number of log lines to show | `100` |
| `--follow, -f` | Follow log output | `false` |
| `--previous` | Show logs from previous container instance | `false` |

#### Examples

```bash
# Debug a specific pod
kdebug pod nginx-deployment-7d4b8c6f9-x8k2l

# Debug all pods in current namespace
kdebug pod

# Debug pods in specific namespace
kdebug pod --namespace production

# Debug pods matching label selector
kdebug pod --selector app=nginx

# Debug across all namespaces
kdebug pod --all-namespaces

# Show more log history
kdebug pod myapp --since 2h --tail 500

# Follow logs in real-time
kdebug pod myapp --follow
```

### `kdebug service` (Coming Soon)

Diagnose service and endpoint issues.

#### Usage

```bash
kdebug service [service-name] [flags]
```

#### Description

The service command will analyze service-related issues including:

- Service discovery and endpoint availability
- Load balancer configuration
- Network policies affecting service traffic
- Service mesh integration issues

### `kdebug ingress` (Coming Soon)

Diagnose ingress routing and TLS issues.

#### Usage

```bash
kdebug ingress [ingress-name] [flags]
```

#### Description

The ingress command will analyze ingress-related issues including:

- Ingress controller health
- TLS certificate validation
- Routing rule conflicts
- Backend service connectivity

### `kdebug dns` (Coming Soon)

Test DNS resolution and diagnose DNS issues.

#### Usage

```bash
kdebug dns [flags]
```

#### Description

The dns command will test DNS resolution including:

- CoreDNS pod health
- DNS policy configuration
- Service discovery resolution
- External DNS resolution

## Output Formats

kdebug supports multiple output formats:

### Table Format (Default)

Human-readable table format with color coding:

```bash
kdebug pod myapp
```

### JSON Format

Machine-readable JSON output:

```bash
kdebug pod myapp --output json
```

### YAML Format

YAML output for configuration review:

```bash
kdebug pod myapp --output yaml
```

## Exit Codes

kdebug uses standard exit codes to indicate the result of operations:

| Exit Code | Description |
|-----------|-------------|
| `0` | Success - No issues found |
| `1` | General error or issues found |
| `2` | Warning - Minor issues found |
| `3` | Critical - Severe issues found |
| `4` | Configuration error |
| `5` | Kubernetes API error |

## Environment Variables

kdebug recognizes these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to kubeconfig file | `$HOME/.kube/config` |
| `KDEBUG_NAMESPACE` | Default namespace | `default` |
| `KDEBUG_OUTPUT` | Default output format | `table` |
| `KDEBUG_TIMEOUT` | Default timeout for operations | `30s` |

## Configuration File

kdebug can use a configuration file for default settings. Create `~/.kdebug.yaml`:

```yaml
# Default configuration for kdebug
namespace: default
output: table
verbose: false
timeout: 30s

# Cluster-specific settings
cluster:
  check-networking: true
  check-dns: true

# Pod-specific settings
pod:
  tail: 100
  since: 1h
```

## Common Usage Patterns

### Quick Health Check

```bash
# Check entire cluster health
kdebug cluster

# Check all pods in current namespace
kdebug pod
```

### Troubleshooting Specific Issues

```bash
# Debug failing deployment
kdebug pod --selector app=myapp --verbose

# Analyze node issues
kdebug cluster --node worker-node-1

# Check system components
kdebug pod --namespace kube-system
```

### Continuous Monitoring

```bash
# Follow pod logs
kdebug pod myapp --follow

# Monitor specific application
watch -n 30 kdebug pod --selector app=myapp
```

## Best Practices

1. **Start with cluster-wide checks**: Use `kdebug cluster` to identify infrastructure issues first
2. **Use label selectors**: Filter pods by labels to focus on specific applications
3. **Enable verbose output**: Use `--verbose` for detailed analysis and troubleshooting steps
4. **Check logs**: Use `--since` and `--tail` flags to get relevant log context
5. **Use appropriate namespaces**: Always specify the correct namespace or use `--all-namespaces`

## Next Steps

- [Examples]({{ '/docs/examples/' | relative_url }}) - Real-world usage scenarios
- [Contributing]({{ '/docs/contributing/' | relative_url }}) - How to contribute to kdebug
- [Getting Started]({{ '/docs/getting-started/' | relative_url }}) - Basic usage guide