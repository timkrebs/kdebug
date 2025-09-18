#!/bin/bash

# Simple CI test to verify the local CI setup works
# This script demonstrates the local CI system

set -e

echo "🧪 Testing Local CI Setup"
echo "========================"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not available. Please install Docker to use local CI."
    exit 1
fi

echo "✅ Docker is available"

# Check if the CI script exists and is executable
if [[ ! -x "./scripts/local-ci.sh" ]]; then
    echo "❌ Local CI script not found or not executable"
    exit 1
fi

echo "✅ Local CI script is ready"

# Check Makefile targets
if ! make help | grep -q "local-ci"; then
    echo "❌ Local CI Makefile targets not found"
    exit 1
fi

echo "✅ Makefile targets are available"

# Show available commands
echo ""
echo "🚀 Available Local CI Commands:"
echo "-------------------------------"
make help | grep local-ci

echo ""
echo "📚 Quick Start:"
echo "  make local-ci-quick    # Fast checks (recommended for development)"
echo "  make local-ci          # Full CI checks"
echo "  make local-ci-verbose  # Full checks with detailed output"

echo ""
echo "✅ Local CI setup is ready!"
echo "   Run 'make local-ci-quick' to start testing your code locally."