---
layout: docs
title: Getting Started
description: Learn how to install and use kdebug to debug your Kubernetes clusters and pods.
permalink: /docs/getting-started/
order: 1
---

# Getting Started with kdebug

kdebug is a powerful command-line tool designed to help you debug Kubernetes pods and clusters quickly and efficiently. This guide will walk you through the basics of getting started with kdebug.

## What is kdebug?

kdebug simplifies Kubernetes debugging by providing:

- **Pod Analysis**: Comprehensive pod debugging with resource analysis, logs, and health checks
- **Cluster Insights**: Detailed cluster information including nodes, namespaces, and resources
- **Easy to Use**: Simple command-line interface with intuitive commands
- **Cross Platform**: Works on Linux, macOS, and Windows

## Prerequisites

Before using kdebug, ensure you have:

- A Kubernetes cluster (local or remote)
- `kubectl` installed and configured
- Appropriate permissions to access the cluster resources you want to debug

## Quick Start

### 1. Install kdebug

Download the latest release for your platform:

```bash
# Linux
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-linux-amd64 -o kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/

# macOS
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-darwin-amd64 -o kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/your-username/kdebug/releases/latest/download/kdebug-windows-amd64.exe" -OutFile "kdebug.exe"
```

### 2. Verify Installation

```bash
kdebug --version
```

### 3. Basic Usage

#### Debug a Pod

```bash
# Debug a specific pod
kdebug pod nginx-pod

# Debug pods in a specific namespace
kdebug pod --namespace=production

# Debug all pods matching a label selector
kdebug pod --selector=app=nginx
```

#### Analyze Your Cluster

```bash
# Get cluster overview
kdebug cluster

# Detailed cluster analysis
kdebug cluster --verbose

# Check specific node
kdebug cluster --node=worker-node-1
```

## Next Steps

Now that you have kdebug installed and running, explore these guides:

- [Installation Guide]({{ '/docs/installation/' | relative_url }}) - Detailed installation instructions for all platforms
- [Commands Reference]({{ '/docs/commands/' | relative_url }}) - Complete command reference
- [Examples]({{ '/docs/examples/' | relative_url }}) - Real-world usage examples
- [Contributing]({{ '/docs/contributing/' | relative_url }}) - How to contribute to the project

## Need Help?

- Check out our [Examples]({{ '/docs/examples/' | relative_url }}) for common use cases
- Report issues on [GitHub](https://github.com/your-username/kdebug/issues)
- Read the [Commands Reference]({{ '/docs/commands/' | relative_url }}) for detailed documentation