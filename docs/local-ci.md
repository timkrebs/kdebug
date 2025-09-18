# Local CI Workflow

This document describes how to run CI checks locally using Docker to mirror the GitHub Actions workflow. This helps catch issues before pushing to GitHub and avoid failing CI pipelines.

## Overview

The local CI system provides:

- **Docker-based CI environment** that mirrors GitHub Actions
- **Comprehensive checks** including tests, linting, security, and vulnerability scanning
- **Fast feedback loop** to catch issues early
- **Multiple execution modes** for different use cases

## Prerequisites

- Docker installed and running
- Make (for Makefile targets)
- Bash shell

## Quick Start

### 1. Run Quick CI Checks (Recommended)

For the fastest feedback during development:

```bash
make local-ci-quick
```

This runs:
- Unit tests
- Code linting with golangci-lint
- Basic validation

### 2. Run Full CI Checks

To run all checks that mirror the GitHub Actions workflow:

```bash
make local-ci
```

This runs:
- Unit tests with coverage
- Code linting
- Build verification
- Security scanning with gosec
- Vulnerability checking with nancy

### 3. Run CI with Verbose Output

For debugging or detailed output:

```bash
make local-ci-verbose
```

## Usage Options

### Direct Script Usage

You can also run the CI script directly with various options:

```bash
# Run all checks
./scripts/local-ci.sh --full

# Run quick checks only
./scripts/local-ci.sh --quick

# Run with verbose output
./scripts/local-ci.sh --verbose

# Skip specific checks
./scripts/local-ci.sh --no-tests
./scripts/local-ci.sh --no-security
./scripts/local-ci.sh --no-lint

# Get help
./scripts/local-ci.sh --help
```

### Available Flags

| Flag | Description |
|------|-------------|
| `--quick` | Run only tests and linting (fastest) |
| `--full` | Run all CI checks including security scans |
| `--no-tests` | Skip unit tests |
| `--no-lint` | Skip linting checks |
| `--no-build` | Skip build verification |
| `--no-security` | Skip security scanning |
| `--no-vulnerability` | Skip vulnerability checking |
| `--integration` | Include integration tests (experimental) |
| `--verbose` | Show detailed output |
| `--no-cleanup` | Don't cleanup Docker containers |

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make local-ci` | Run complete local CI (full checks) |
| `make local-ci-quick` | Run quick CI checks (tests + lint) |
| `make local-ci-build` | Build the Docker CI image |
| `make local-ci-no-tests` | Run CI without tests |
| `make local-ci-verbose` | Run CI with verbose output |

## CI Checks Overview

### 1. Unit Tests 📋
- **Purpose**: Verify code functionality
- **Tools**: Go test framework
- **Commands**: `go test ./...`, `make test-coverage`
- **Files**: All `*_test.go` files

### 2. Code Linting 🔍
- **Purpose**: Enforce code quality and style
- **Tools**: golangci-lint, gofumpt, go vet
- **Configuration**: `.golangci.yml`
- **Checks**: Style, complexity, best practices

### 3. Build Verification 🔨
- **Purpose**: Ensure code compiles correctly
- **Commands**: `make build`
- **Verification**: Binary execution test

### 4. Security Scanning 🔒
- **Purpose**: Detect security vulnerabilities
- **Tools**: gosec
- **Scope**: Source code analysis

### 5. Vulnerability Checking 🛡️
- **Purpose**: Check dependencies for known vulnerabilities
- **Tools**: nancy
- **Scope**: Go module dependencies

## Docker Environment

### CI Docker Image

The `Dockerfile.ci` creates an environment with:

- Go 1.24 (matching CI)
- golangci-lint
- gosec (security scanner)
- nancy (vulnerability checker)
- gofumpt (formatter)
- Optional: kind and kubectl for integration tests

### Building the Image

```bash
# Build CI image
make local-ci-build

# Or manually
docker build -f Dockerfile.ci -t kdebug-ci:latest .
```

### Running Interactive Container

For debugging or manual testing:

```bash
docker run -it --rm \
  -v "$(pwd):/workspace" \
  -w "/workspace" \
  kdebug-ci:latest bash
```

## Integration with Development Workflow

### Pre-Push Hook

Add to your git pre-push hook (`.git/hooks/pre-push`):

```bash
#!/bin/bash
echo "Running local CI checks..."
make local-ci-quick
if [ $? -ne 0 ]; then
    echo "❌ Local CI checks failed. Push blocked."
    exit 1
fi
```

### VS Code Integration

Add to `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Local CI Quick",
            "type": "shell",
            "command": "make local-ci-quick",
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ]
}
```

### IDE Integration

Most IDEs can run Makefile targets directly:
- **VS Code**: Use the "Tasks" feature
- **GoLand/IntelliJ**: Run Makefile targets
- **Vim/Neovim**: Use `:make local-ci-quick`

## Troubleshooting

### Common Issues

1. **Docker not running**
   ```
   Error: Docker daemon is not running
   ```
   **Solution**: Start Docker Desktop or Docker daemon

2. **Permission denied**
   ```
   Error: Permission denied
   ```
   **Solution**: Make script executable: `chmod +x scripts/local-ci.sh`

3. **Go module issues**
   ```
   Error: go mod download failed
   ```
   **Solution**: Clear module cache: `go clean -modcache`

4. **Linting failures**
   ```
   Error: golangci-lint run failed
   ```
   **Solution**: Check `.golangci.yml` configuration

### Debug Mode

Run with verbose output to see detailed logs:

```bash
./scripts/local-ci.sh --verbose --full
```

### Manual Container Inspection

```bash
# Run container interactively
docker run -it --rm \
  -v "$(pwd):/workspace" \
  -w "/workspace" \
  kdebug-ci:latest bash

# Inside container, run individual commands
go mod download
go test ./...
golangci-lint run
```

## Performance Tips

1. **Use Quick Mode for Development**
   - Run `make local-ci-quick` during development
   - Save full CI for pre-push validation

2. **Docker Layer Caching**
   - The Dockerfile is optimized for layer caching
   - Dependencies are downloaded before copying source code

3. **Parallel Execution**
   - Multiple checks run in sequence but optimized
   - Consider running specific checks: `./scripts/local-ci.sh --no-security`

## Comparison with GitHub Actions

| Check | GitHub Actions | Local CI | Notes |
|-------|---------------|----------|-------|
| Unit Tests | ✅ | ✅ | Identical |
| Coverage | ✅ | ✅ | Generated locally |
| Linting | ✅ | ✅ | Same golangci-lint config |
| Build | ✅ | ✅ | Same Go version |
| Security | ✅ | ✅ | Same gosec rules |
| Vulnerability | ✅ | ✅ | Same nancy checks |
| Integration | ✅ | ⚠️ | Optional (requires setup) |
| Cross-compile | ✅ | ❌ | GitHub only |
| Docker build | ✅ | ❌ | GitHub only |

## Best Practices

1. **Run Before Every Push**
   ```bash
   make local-ci-quick && git push
   ```

2. **Use in CI/CD Pipeline**
   - Add to pre-commit hooks
   - Integrate with IDE
   - Use in code review process

3. **Keep Docker Image Updated**
   ```bash
   make local-ci-build  # Rebuild when tools change
   ```

4. **Monitor Performance**
   - Track execution times
   - Optimize for development workflow

## Files Overview

```
.
├── Dockerfile.ci              # CI Docker environment
├── docker-compose.ci.yml      # Docker Compose for CI
├── scripts/
│   └── local-ci.sh           # Main CI script
├── .golangci.yml             # Linting configuration
└── docs/
    └── local-ci.md           # This documentation
```

## Support

For issues or improvements:

1. Check the troubleshooting section
2. Run with `--verbose` for detailed output
3. Review Docker logs: `docker logs kdebug-ci`
4. Open an issue in the repository

---

**Happy coding! 🚀**

Remember: Local CI helps catch issues early and keeps your GitHub Actions green! ✅