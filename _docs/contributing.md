---
layout: docs
title: Contributing
description: Learn how to contribute to kdebug development and help improve the project.
permalink: /docs/contributing/
order: 5
---

# Contributing to kdebug

Thank you for your interest in contributing to kdebug! We welcome contributions from developers of all skill levels. This guide will help you get started with contributing to the project.

## Code of Conduct

Be kind and respectful to the members of the community. Take time to educate others who are seeking help. Harassment of any kind will not be tolerated.

## Ways to Contribute

There are many ways to contribute to kdebug:

- **Report bugs** - Help us identify and fix issues
- **Suggest features** - Propose new functionality or improvements
- **Write code** - Implement bug fixes or new features
- **Improve documentation** - Help make the docs clearer and more comprehensive
- **Test and review** - Test new features and review pull requests
- **Share knowledge** - Write blog posts, tutorials, or speak at conferences

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.19 or later installed
- Git installed and configured
- A GitHub account
- Basic knowledge of Kubernetes and kubectl

### Setting Up the Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/kdebug.git
   cd kdebug
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/original-owner/kdebug.git
   ```

4. **Install development dependencies**:
   ```bash
   make dev-deps
   ```

5. **Build and test**:
   ```bash
   make all
   ```

## Development Workflow

### Local Development

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** and ensure they follow our coding standards

3. **Run tests** to verify your changes:
   ```bash
   make test
   ```

4. **Run linting** and formatting:
   ```bash
   make lint
   make fmt
   ```

5. **Test the binary**:
   ```bash
   make build
   ./bin/kdebug --help
   ```

### Testing Your Changes

#### Unit Tests

Run the test suite to ensure your changes don't break existing functionality:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/pod/...
```

#### Integration Tests

Test with a real Kubernetes cluster:

```bash
# Start a local cluster (minikube, kind, etc.)
make test-integration

# Or test manually
./bin/kdebug cluster
./bin/kdebug pod --namespace kube-system
```

#### Local CI Testing

Use our Docker-based local CI to test all CI steps:

```bash
# Quick local CI check
make local-ci-quick

# Full local CI (same as GitHub Actions)
make local-ci
```

### Code Style and Standards

#### Go Code Standards

- Follow standard Go formatting (use `gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors appropriately

#### Example Code Style

```go
// AnalyzePod examines a pod and returns diagnostic information
func AnalyzePod(ctx context.Context, podName, namespace string) (*PodAnalysis, error) {
    if podName == "" {
        return nil, fmt.Errorf("pod name cannot be empty")
    }

    pod, err := client.GetPod(ctx, podName, namespace)
    if err != nil {
        return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
    }

    analysis := &PodAnalysis{
        Name:      pod.Name,
        Namespace: pod.Namespace,
        Status:    string(pod.Status.Phase),
    }

    return analysis, nil
}
```

#### Commit Message Format

Use clear, descriptive commit messages:

```
feat: add support for network policy analysis

- Implement network policy checker
- Add tests for policy evaluation
- Update documentation

Closes #123
```

Types of commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test changes
- `refactor:` - Code refactoring
- `style:` - Formatting changes
- `ci:` - CI/CD changes

## Filing Issues

### Bug Reports

When reporting bugs, please include:

1. **Clear title** describing the issue
2. **Steps to reproduce** the problem
3. **Expected behavior** vs **actual behavior**
4. **Environment details**:
   - kdebug version (`kdebug --version`)
   - Kubernetes version (`kubectl version`)
   - Operating system
   - Any relevant configuration

#### Bug Report Template

```markdown
**Bug Description**
A clear description of the bug.

**Steps to Reproduce**
1. Run command: `kdebug pod my-pod`
2. See error message: ...

**Expected Behavior**
The command should analyze the pod and show results.

**Actual Behavior**
Error: connection refused

**Environment**
- kdebug version: v1.0.1
- Kubernetes version: v1.25.0
- OS: macOS 13.0
- kubectl context: minikube
```

### Feature Requests

For feature requests, please include:

1. **Clear description** of the feature
2. **Use case** - why is this feature needed?
3. **Proposed solution** - how should it work?
4. **Alternatives considered** - other approaches you've thought about

#### Feature Request Template

```markdown
**Feature Description**
Add support for analyzing StatefulSets.

**Use Case**
As a DevOps engineer, I want to analyze StatefulSet issues like persistent volume problems and ordered deployment failures.

**Proposed Solution**
Add a new command `kdebug statefulset` that checks:
- PVC binding issues
- Pod ordering problems
- Rolling update status

**Alternatives**
Could extend the existing `kdebug pod` command, but a dedicated command would be clearer.
```

## Pull Request Process

### Before Submitting

1. **Ensure tests pass**: `make test`
2. **Run linting**: `make lint`
3. **Update documentation** if needed
4. **Add tests** for new functionality
5. **Update CHANGELOG.md** if applicable

### Pull Request Template

```markdown
**Description**
Brief description of changes.

**Type of Change**
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

**Testing**
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

**Checklist**
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. **Automated checks** will run (CI, linting, tests)
2. **Code review** by maintainers
3. **Address feedback** and make requested changes
4. **Approval and merge** by maintainers

## Project Structure

Understanding the codebase structure helps with contributions:

```
kdebug/
├── cmd/                 # CLI commands (cobra)
│   ├── root.go         # Root command and global flags
│   ├── cluster.go      # Cluster analysis command
│   └── pod.go          # Pod analysis command
├── internal/           # Internal packages
│   ├── client/         # Kubernetes client wrapper
│   └── output/         # Output formatting
├── pkg/                # Public packages
│   ├── cluster/        # Cluster analysis logic
│   └── pod/            # Pod analysis logic
├── test/               # Integration tests
├── docs/               # Documentation
└── scripts/            # Build and development scripts
```

## Development Guidelines

### Adding New Commands

1. **Create command file** in `cmd/` directory
2. **Implement analysis logic** in appropriate `pkg/` package
3. **Add tests** for both command and logic
4. **Update documentation** and help text
5. **Add examples** to docs

### Adding New Checks

1. **Identify the issue** you want to detect
2. **Implement check function** with clear error messages
3. **Add remediation suggestions**
4. **Write comprehensive tests**
5. **Document the check** in relevant docs

### Error Handling

- Use structured errors with context
- Provide actionable error messages
- Include relevant Kubernetes resource information
- Use appropriate log levels

## Building and Releasing

### Local Builds

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release artifacts
make release
```

### Release Process

Releases are handled by maintainers:

1. Version bump in `cmd/root.go`
2. Update `CHANGELOG.md`
3. Create git tag
4. GitHub Actions builds and publishes binaries

## Getting Help

If you need help while contributing:

- **Check existing issues** and documentation
- **Ask questions** in GitHub issues
- **Join discussions** in GitHub Discussions
- **Review existing code** for patterns and examples

## Recognition

Contributors are recognized in:

- `CONTRIBUTORS.md` file
- Release notes
- GitHub contributors section

## Thank You

Every contribution, no matter how small, helps make kdebug better for the entire Kubernetes community. Thank you for taking the time to contribute!

## Next Steps

Ready to contribute? Here are some good first issues:

1. **Good first issues** - Look for issues labeled `good first issue`
2. **Documentation improvements** - Help improve existing docs
3. **Test coverage** - Add tests for existing functionality
4. **Bug fixes** - Fix reported bugs

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Testing in Go](https://golang.org/doc/tutorial/add-a-test)