---
layout: docs
title: Installation
description: Detailed installation instructions for kdebug on all supported platforms.
permalink: /docs/installation/
order: 2
---

# Installation

kdebug is available for Linux, macOS, and Windows. Choose the installation method that works best for your environment.

## Download Pre-built Binaries

### Latest Release

Download the latest release from GitHub:

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux | amd64 | [kdebug-linux-amd64](https://github.com/your-username/kdebug/releases/latest/download/kdebug-linux-amd64) |
| Linux | arm64 | [kdebug-linux-arm64](https://github.com/your-username/kdebug/releases/latest/download/kdebug-linux-arm64) |
| macOS | amd64 | [kdebug-darwin-amd64](https://github.com/your-username/kdebug/releases/latest/download/kdebug-darwin-amd64) |
| macOS | arm64 | [kdebug-darwin-arm64](https://github.com/your-username/kdebug/releases/latest/download/kdebug-darwin-arm64) |
| Windows | amd64 | [kdebug-windows-amd64.exe](https://github.com/your-username/kdebug/releases/latest/download/kdebug-windows-amd64.exe) |

## Platform-Specific Instructions

### Linux

#### Using curl

```bash
# For x86_64 systems
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-linux-amd64 -o kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/

# For ARM64 systems
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-linux-arm64 -o kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/
```

#### Using wget

```bash
# For x86_64 systems
wget https://github.com/your-username/kdebug/releases/latest/download/kdebug-linux-amd64 -O kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/
```

### macOS

#### Using curl

```bash
# For Intel Macs
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-darwin-amd64 -o kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/

# For Apple Silicon Macs
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-darwin-arm64 -o kdebug
chmod +x kdebug
sudo mv kdebug /usr/local/bin/
```

#### Using Homebrew (Coming Soon)

```bash
# This will be available in a future release
brew install kdebug
```

### Windows

#### Using PowerShell

```powershell
# Download the binary
Invoke-WebRequest -Uri "https://github.com/your-username/kdebug/releases/latest/download/kdebug-windows-amd64.exe" -OutFile "kdebug.exe"

# Move to a directory in your PATH
Move-Item kdebug.exe C:\Windows\System32\kdebug.exe
```

#### Using Command Prompt

```cmd
# Download using curl (if available)
curl -L https://github.com/your-username/kdebug/releases/latest/download/kdebug-windows-amd64.exe -o kdebug.exe

# Or download manually and add to PATH
```

## Build from Source

If you prefer to build kdebug from source or want to contribute to development:

### Prerequisites

- Go 1.19 or later
- Git

### Build Steps

```bash
# Clone the repository
git clone https://github.com/your-username/kdebug.git
cd kdebug

# Build the binary
make build

# Install to /usr/local/bin (Linux/macOS)
sudo make install

# Or just run from the current directory
./bin/kdebug --version
```

### Development Build

For development purposes, you can use:

```bash
# Install development dependencies
make dev-deps

# Run tests
make test

# Build and test
make all
```

## Verify Installation

After installation, verify that kdebug is working correctly:

```bash
# Check version
kdebug --version

# Show help
kdebug --help

# Test with your cluster (requires kubectl to be configured)
kdebug cluster
```

## Updating kdebug

To update kdebug to the latest version, simply download and replace the binary using the same installation method you used initially.

For automated updates, you can create a simple script:

```bash
#!/bin/bash
# update-kdebug.sh

LATEST_VERSION=$(curl -s https://api.github.com/repos/your-username/kdebug/releases/latest | grep tag_name | cut -d '"' -f 4)
CURRENT_VERSION=$(kdebug --version 2>/dev/null | cut -d ' ' -f 3)

if [ "$LATEST_VERSION" != "$CURRENT_VERSION" ]; then
    echo "Updating kdebug from $CURRENT_VERSION to $LATEST_VERSION"
    curl -L "https://github.com/your-username/kdebug/releases/latest/download/kdebug-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64" -o /tmp/kdebug
    chmod +x /tmp/kdebug
    sudo mv /tmp/kdebug /usr/local/bin/kdebug
    echo "Update complete!"
else
    echo "kdebug is already up to date ($CURRENT_VERSION)"
fi
```

## Troubleshooting

### Permission Denied

If you get a "permission denied" error on Linux/macOS:

```bash
chmod +x kdebug
```

### Command Not Found

If you get "command not found" after installation:

1. Ensure the binary is in your PATH
2. Check that `/usr/local/bin` is in your PATH: `echo $PATH`
3. Try running with the full path: `/usr/local/bin/kdebug`

### Cluster Access Issues

If kdebug can't access your cluster:

1. Ensure `kubectl` is installed and configured
2. Test cluster access: `kubectl get nodes`
3. Check your kubeconfig: `kubectl config current-context`

## Next Steps

- [Getting Started]({{ '/docs/getting-started/' | relative_url }}) - Basic usage guide
- [Commands Reference]({{ '/docs/commands/' | relative_url }}) - Complete command documentation
- [Examples]({{ '/docs/examples/' | relative_url }}) - Real-world usage examples