#!/bin/bash

# kdebug Demo Script
# This script demonstrates the basic functionality of kdebug

set -e

echo "🚀 kdebug Demo Script"
echo "===================="
echo

# Build kdebug
echo "📦 Building kdebug..."
make build
echo

# Show version
echo "📋 Version information:"
./bin/kdebug --version
echo

# Show help
echo "❓ Help information:"
./bin/kdebug --help
echo

# Show cluster command help
echo "🏥 Cluster command help:"
./bin/kdebug cluster --help
echo

# Try to run cluster diagnostics
echo "🔍 Running cluster diagnostics..."
echo "(This will likely fail if no cluster is available, which is expected)"
echo

if ./bin/kdebug cluster --verbose 2>/dev/null; then
    echo "✅ Cluster diagnostics completed successfully!"
else
    echo "⚠️  Cluster diagnostics failed (expected if no cluster is available)"
    echo "   To test with a real cluster:"
    echo "   1. Install kind: https://kind.sigs.k8s.io/"
    echo "   2. Create cluster: kind create cluster --name kdebug-test"
    echo "   3. Run: ./bin/kdebug cluster --verbose"
fi

echo
echo "🎉 Demo completed!"
echo
echo "Next steps:"
echo "- Set up a Kubernetes cluster (kind, minikube, or real cluster)"
echo "- Run: ./bin/kdebug cluster --verbose"
echo "- Try different output formats: --output json, --output yaml"
echo "- Check docs/development.md for more testing scenarios"
