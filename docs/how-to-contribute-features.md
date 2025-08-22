# How to Contribute New Features to kdebug

This guide walks you through the complete process of implementing a new feature for kdebug, from initial idea to merged pull request.

## Table of Contents

1. [Before You Start](#before-you-start)
2. [Planning Your Feature](#planning-your-feature)
3. [Setting Up Development](#setting-up-development)
4. [Implementation Guidelines](#implementation-guidelines)
5. [Testing Your Feature](#testing-your-feature)
6. [Documentation](#documentation)
7. [Submitting Your Feature](#submitting-your-feature)
8. [Review Process](#review-process)
9. [Common Patterns](#common-patterns)
10. [Examples](#examples)

## Before You Start

### 🔍 Research and Planning

1. **Search existing issues** - Check if your feature is already requested or planned
2. **Review the roadmap** - Ensure your feature aligns with project direction
3. **Study kdebug architecture** - Understand how existing features work
4. **Consider the scope** - Start with a focused, well-defined feature

### 📝 Create a Feature Request

Before implementing, create a detailed feature request using our [new feature template](.github/ISSUE_TEMPLATE/03_new_feature.yml):

```bash
# Go to GitHub Issues
# Click "New Issue" → "🚀 New Feature Request"
# Fill out all required sections thoroughly
```

**Key sections to focus on:**
- **Problem Statement** - What pain point does this solve?
- **Use Cases** - Specific scenarios where this helps
- **Proposed Solution** - How should it work?
- **Command Examples** - Show the CLI interface
- **Acceptance Criteria** - Definition of done

### 💬 Discuss First

Comment on your feature request or join discussions to:
- Get feedback from maintainers
- Refine the scope and approach
- Understand implementation preferences
- Identify potential challenges

## Planning Your Feature

### 🎯 Define Your Feature Category

Choose the primary category for your feature:

| Category | Examples | Implementation Focus |
|----------|----------|---------------------|
| **Workload Diagnostics** | Pod health, deployment issues | Container inspection, resource analysis |
| **Networking & Connectivity** | Service mesh, DNS resolution | Network policies, connectivity tests |
| **Cluster Health & Resources** | Node capacity, storage issues | Resource utilization, performance metrics |
| **Security & RBAC** | Permission checks, policy validation | Security scanning, compliance checks |
| **Configuration Validation** | Manifest validation, best practices | Static analysis, rule engines |

### 🏗️ Architecture Planning

Consider these architectural aspects:

1. **CLI Interface Design**
   ```bash
   # Command structure
   kdebug <resource> [subcommand] [flags]
   
   # Examples
   kdebug network                    # Network diagnostics
   kdebug network dns               # DNS-specific checks
   kdebug network --namespace kube-system
   ```

2. **Data Flow**
   ```
   CLI Command → Kubernetes Client → Diagnostic Logic → Output Formatter → User
   ```

3. **Error Handling Strategy**
   - Graceful degradation when permissions are limited
   - Actionable error messages with suggestions
   - Fallback mechanisms for different K8s versions

## Setting Up Development

### 🚀 Initial Setup

```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/kdebug.git
cd kdebug

# 3. Add upstream remote
git remote add upstream https://github.com/timkrebs/kdebug.git

# 4. Create feature branch
git checkout -b feature/your-feature-name

# 5. Install dependencies
make deps
make dev-deps

# 6. Verify setup
make test
make lint
```

### 🛠️ Development Environment

**Prerequisites:**
- Go 1.23+
- kubectl configured with a test cluster
- kind (for integration testing)
- Docker (for container testing)

**Recommended tools:**
- golangci-lint
- gosec
- gofumpt

## Implementation Guidelines

### 📁 Project Structure

Follow kdebug's established patterns:

```
kdebug/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command setup
│   ├── cluster.go         # Cluster diagnostics
│   └── your_feature.go    # Your new command
├── pkg/                   # Core business logic
│   ├── cluster/           # Cluster diagnostic logic
│   └── your_feature/      # Your feature logic
├── internal/              # Internal utilities
│   ├── client/            # Kubernetes client wrapper
│   └── output/            # Output formatting
└── test/                  # Tests
    ├── integration/       # Integration tests
    └── unit/             # Unit tests
```

### 🎨 Implementation Patterns

#### 1. CLI Command Structure

```go
// cmd/your_feature.go
package cmd

import (
    "github.com/spf13/cobra"
    "kdebug/pkg/your_feature"
)

var yourFeatureCmd = &cobra.Command{
    Use:   "your-feature [flags]",
    Short: "Brief description of your feature",
    Long: `Detailed description explaining:
- What the feature does
- When to use it
- What it checks or analyzes`,
    Example: `  # Basic usage
  kdebug your-feature

  # With specific options
  kdebug your-feature --namespace production --output json`,
    RunE: runYourFeature,
}

func init() {
    rootCmd.AddCommand(yourFeatureCmd)
    
    // Add feature-specific flags
    yourFeatureCmd.Flags().StringP("namespace", "n", "", "Target namespace")
    yourFeatureCmd.Flags().BoolP("detailed", "d", false, "Show detailed analysis")
}

func runYourFeature(cmd *cobra.Command, args []string) error {
    // 1. Parse flags and validate input
    namespace, _ := cmd.Flags().GetString("namespace")
    detailed, _ := cmd.Flags().GetBool("detailed")
    
    // 2. Get global configuration
    outputFormat, _ := cmd.Flags().GetString("outputFormat")
    verbose, _ := cmd.Flags().GetBool("verbose")
    kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
    
    // 3. Initialize dependencies
    outputManager := output.NewOutputManager(outputFormat, verbose)
    k8sClient, err := client.NewKubernetesClient(kubeconfig, namespace)
    if err != nil {
        return fmt.Errorf("failed to create Kubernetes client: %w", err)
    }
    
    // 4. Test connectivity
    if err := k8sClient.TestConnection(); err != nil {
        outputManager.PrintError("Kubernetes connectivity check failed")
        return err
    }
    
    // 5. Run diagnostics
    diagnostic := your_feature.NewYourFeatureDiagnostic(k8sClient, outputManager)
    report, err := diagnostic.RunDiagnostics(namespace, detailed)
    if err != nil {
        return fmt.Errorf("diagnostic failed: %w", err)
    }
    
    // 6. Output results
    outputManager.PrintReport(report)
    
    // 7. Return appropriate exit code
    if report.Summary.Failed > 0 {
        return fmt.Errorf("found %d issues", report.Summary.Failed)
    }
    
    return nil
}
```

#### 2. Diagnostic Logic Structure

```go
// pkg/your_feature/your_feature.go
package your_feature

import (
    "context"
    "fmt"
    
    "kdebug/internal/client"
    "kdebug/internal/output"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type YourFeatureDiagnostic struct {
    client *client.KubernetesClient
    output *output.OutputManager
}

func NewYourFeatureDiagnostic(client *client.KubernetesClient, output *output.OutputManager) *YourFeatureDiagnostic {
    return &YourFeatureDiagnostic{
        client: client,
        output: output,
    }
}

func (d *YourFeatureDiagnostic) RunDiagnostics(namespace string, detailed bool) (*output.DiagnosticReport, error) {
    checks := make([]output.CheckResult, 0, 5) // Pre-allocate with expected capacity
    
    // Run individual checks
    connectivityCheck := d.checkConnectivity()
    checks = append(checks, connectivityCheck)
    
    if detailed {
        detailedChecks := d.runDetailedChecks(namespace)
        checks = append(checks, detailedChecks...)
    }
    
    // Calculate summary
    summary := calculateSummary(checks)
    
    // Create report
    report := &output.DiagnosticReport{
        Target:    fmt.Sprintf("your-feature/%s", namespace),
        Timestamp: time.Now().Format(time.RFC3339),
        Checks:    checks,
        Summary:   summary,
    }
    
    return report, nil
}

func (d *YourFeatureDiagnostic) checkConnectivity() output.CheckResult {
    // Test basic connectivity to relevant K8s resources
    if err := d.client.TestConnection(); err != nil {
        return output.CheckResult{
            Name:        "Kubernetes Connectivity",
            Status:      "FAILED",
            Message:     "Cannot connect to Kubernetes API",
            Suggestion:  "Check kubeconfig and cluster availability",
            Details:     err.Error(),
        }
    }
    
    return output.CheckResult{
        Name:    "Kubernetes Connectivity",
        Status:  "PASSED",
        Message: "Successfully connected to Kubernetes API",
    }
}

func (d *YourFeatureDiagnostic) runDetailedChecks(namespace string) []output.CheckResult {
    var checks []output.CheckResult
    
    // Example: Check specific resources
    pods, err := d.client.Clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        checks = append(checks, output.CheckResult{
            Name:       "Pod Discovery",
            Status:     "FAILED",
            Message:    "Failed to list pods",
            Suggestion: "Verify namespace exists and you have list permissions",
            Details:    err.Error(),
        })
        return checks
    }
    
    // Analyze pods
    for i := range pods.Items {
        pod := &pods.Items[i]
        checkResult := d.analyzePod(pod)
        checks = append(checks, checkResult)
    }
    
    return checks
}

func (d *YourFeatureDiagnostic) analyzePod(pod *v1.Pod) output.CheckResult {
    // Implement your analysis logic here
    if pod.Status.Phase != v1.PodRunning {
        return output.CheckResult{
            Name:       fmt.Sprintf("Pod %s", pod.Name),
            Status:     "FAILED",
            Message:    fmt.Sprintf("Pod is in %s state", pod.Status.Phase),
            Suggestion: d.getPodSuggestion(pod),
            Details:    fmt.Sprintf("Namespace: %s, Phase: %s", pod.Namespace, pod.Status.Phase),
        }
    }
    
    return output.CheckResult{
        Name:    fmt.Sprintf("Pod %s", pod.Name),
        Status:  "PASSED",
        Message: "Pod is running normally",
    }
}

func (d *YourFeatureDiagnostic) getPodSuggestion(pod *v1.Pod) string {
    switch pod.Status.Phase {
    case v1.PodPending:
        return "Check pod events: kubectl describe pod " + pod.Name + " -n " + pod.Namespace
    case v1.PodFailed:
        return "Check pod logs: kubectl logs " + pod.Name + " -n " + pod.Namespace
    default:
        return "Investigate pod status and events"
    }
}

func calculateSummary(checks []output.CheckResult) output.Summary {
    summary := output.Summary{}
    
    for _, check := range checks {
        switch check.Status {
        case "PASSED":
            summary.Passed++
        case "FAILED":
            summary.Failed++
        case "WARNING":
            summary.Warning++
        case "SKIPPED":
            summary.Skipped++
        }
        summary.Total++
    }
    
    return summary
}
```

### 🎯 Key Implementation Principles

1. **Follow Existing Patterns**
   - Use the same error handling approach as other commands
   - Maintain consistent CLI flag naming
   - Follow the established project structure

2. **Kubernetes Best Practices**
   - Use proper context handling for API calls
   - Implement proper RBAC permission checks
   - Handle different Kubernetes versions gracefully

3. **User Experience**
   - Provide actionable suggestions in error messages
   - Support multiple output formats (table, JSON, YAML)
   - Include relevant kubectl commands in suggestions

4. **Performance Considerations**
   - Pre-allocate slices when possible
   - Use efficient Kubernetes API calls
   - Implement timeouts for long-running operations

## Testing Your Feature

### 🧪 Unit Tests

Create comprehensive unit tests for your feature:

```go
// pkg/your_feature/your_feature_test.go
package your_feature

import (
    "testing"
    
    "kdebug/internal/client"
    "kdebug/internal/output"
)

func TestNewYourFeatureDiagnostic(t *testing.T) {
    // Test constructor
    client := &client.KubernetesClient{}
    outputManager := output.NewOutputManager("table", false)
    
    diagnostic := NewYourFeatureDiagnostic(client, outputManager)
    
    if diagnostic.client != client {
        t.Error("Client not set correctly")
    }
    if diagnostic.output != outputManager {
        t.Error("Output manager not set correctly")
    }
}

func TestCalculateSummary(t *testing.T) {
    tests := []struct {
        name     string
        checks   []output.CheckResult
        expected output.Summary
    }{
        {
            name: "all passed",
            checks: []output.CheckResult{
                {Status: "PASSED"},
                {Status: "PASSED"},
            },
            expected: output.Summary{Total: 2, Passed: 2},
        },
        {
            name: "mixed results",
            checks: []output.CheckResult{
                {Status: "PASSED"},
                {Status: "FAILED"},
                {Status: "WARNING"},
            },
            expected: output.Summary{Total: 3, Passed: 1, Failed: 1, Warning: 1},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := calculateSummary(tt.checks)
            if result != tt.expected {
                t.Errorf("Expected %+v, got %+v", tt.expected, result)
            }
        })
    }
}
```

### 🔧 Integration Tests

Add integration tests that run against a real Kubernetes cluster:

```go
// test/integration/your_feature_test.go
package integration

import (
    "testing"
    "os/exec"
    "strings"
)

func TestYourFeatureIntegration(t *testing.T) {
    // Ensure test cluster is available
    ensureTestCluster(t)
    defer cleanupTestCluster(t)
    
    tests := []struct {
        name          string
        args          []string
        expectSuccess bool
        expectOutput  string
    }{
        {
            name:          "basic feature test",
            args:          []string{"your-feature"},
            expectSuccess: true,
            expectOutput:  "PASSED",
        },
        {
            name:          "feature with namespace",
            args:          []string{"your-feature", "--namespace", "kube-system"},
            expectSuccess: true,
            expectOutput:  "kube-system",
        },
        {
            name:          "json output",
            args:          []string{"your-feature", "--output", "json"},
            expectSuccess: true,
            expectOutput:  `"target":`,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command(getBinaryPath(), tt.args...)
            output, err := cmd.CombinedOutput()
            
            if tt.expectSuccess && err != nil {
                t.Errorf("Expected success but got error: %v\nOutput: %s", err, output)
            }
            
            if !strings.Contains(string(output), tt.expectOutput) {
                t.Errorf("Expected output to contain %q, got: %s", tt.expectOutput, output)
            }
        })
    }
}
```

### 🏃 Run Tests

```bash
# Unit tests
make test

# Integration tests (requires kind cluster)
make test-integration

# All tests
make test-all

# Test coverage
make test-coverage
```

## Documentation

### 📚 Update Documentation

1. **Command Help**
   - Ensure your command has clear `Short` and `Long` descriptions
   - Provide practical examples in the `Example` field
   - Document all flags and their purposes

2. **README Updates**
   - Add your feature to the "Currently Available" section
   - Include usage examples
   - Update the feature matrix if applicable

3. **User Guide** (if needed)
   - Create detailed usage documentation
   - Include troubleshooting guidance
   - Provide real-world scenarios

### 📝 Code Documentation

```go
// Package your_feature provides diagnostic capabilities for [describe your feature area].
//
// This package implements checks for [specific functionality] including:
//   - [Check type 1]: Description of what it validates
//   - [Check type 2]: Description of what it validates
//   - [Check type 3]: Description of what it validates
//
// The diagnostics help users identify and resolve issues with [specific area].
package your_feature

// YourFeatureDiagnostic performs diagnostic checks for [specific functionality].
// It analyzes [what it analyzes] and provides actionable recommendations
// for resolving identified issues.
type YourFeatureDiagnostic struct {
    // client provides access to the Kubernetes API
    client *client.KubernetesClient
    
    // output handles formatting and display of diagnostic results
    output *output.OutputManager
}
```

## Submitting Your Feature

### ✅ Pre-submission Checklist

Before creating a pull request:

- [ ] **Code Quality**
  - [ ] All tests pass (`make test-all`)
  - [ ] Linting passes (`make lint`)
  - [ ] Security scan passes (`make security`)
  - [ ] Code is properly formatted (`make fmt`)

- [ ] **Functionality**
  - [ ] Feature works with different output formats
  - [ ] Error handling is comprehensive
  - [ ] Suggestions are actionable
  - [ ] Integration tests pass

- [ ] **Documentation**
  - [ ] Command help is clear and complete
  - [ ] Code is well-commented
  - [ ] README is updated if needed

### 🚀 Creating the Pull Request

```bash
# 1. Ensure your branch is up to date
git fetch upstream
git rebase upstream/main

# 2. Run pre-push validation
make pre-push

# 3. Push your feature branch
git push origin feature/your-feature-name

# 4. Create pull request on GitHub
```

**Pull Request Guidelines:**

1. **Title**: Use a clear, descriptive title
   ```
   feat: add network policy diagnostics command
   ```

2. **Description**: Include:
   - What the feature does
   - How to test it
   - Any breaking changes
   - Link to the original feature request

3. **Testing**: Provide test commands for reviewers
   ```bash
   # Test the new feature
   kdebug your-feature --namespace test-ns
   kdebug your-feature --output json
   ```

## Review Process

### 🔍 What Reviewers Look For

1. **Code Quality**
   - Follows project conventions
   - Proper error handling
   - Efficient resource usage
   - Clean, readable code

2. **User Experience**
   - Intuitive CLI interface
   - Clear, actionable output
   - Helpful error messages
   - Consistent with existing commands

3. **Kubernetes Integration**
   - Proper API usage
   - RBAC considerations
   - Version compatibility
   - Resource efficiency

### 🔄 Iteration Process

1. **Address Feedback**: Respond to review comments promptly
2. **Update Documentation**: Keep docs in sync with code changes
3. **Re-test**: Verify all tests still pass after changes
4. **Communicate**: Ask questions if feedback is unclear

## Common Patterns

### 🔧 Error Handling

```go
// Always provide context and suggestions
if err != nil {
    return output.CheckResult{
        Name:       "Check Name",
        Status:     "FAILED",
        Message:    "Brief description of what failed",
        Suggestion: "Specific action user can take",
        Details:    fmt.Sprintf("Technical details: %v", err),
    }
}
```

### 🎯 Resource Access Patterns

```go
// List resources with proper error handling
resources, err := d.client.Clientset.
    AppsV1().
    Deployments(namespace).
    List(context.TODO(), metav1.ListOptions{})
    
if err != nil {
    // Handle different error types
    if errors.IsForbidden(err) {
        return output.CheckResult{
            Status:     "SKIPPED",
            Message:    "Insufficient permissions to list deployments",
            Suggestion: "Ensure you have 'list' permissions for deployments",
        }
    }
    // ... handle other error types
}
```

### 📊 Multi-Check Patterns

```go
func (d *Diagnostic) runChecks() []output.CheckResult {
    checks := make([]output.CheckResult, 0, 10) // Pre-allocate
    
    // Use a slice of check functions for maintainability
    checkFunctions := []func() output.CheckResult{
        d.checkConnectivity,
        d.checkPermissions,
        d.checkResources,
    }
    
    for _, checkFunc := range checkFunctions {
        result := checkFunc()
        checks = append(checks, result)
    }
    
    return checks
}
```

## Examples

### 🌟 Simple Feature: Pod Restart Diagnostics

A feature that analyzes pod restart patterns:

```bash
# Usage
kdebug pod-restarts                           # Check all namespaces
kdebug pod-restarts --namespace production    # Specific namespace
kdebug pod-restarts --threshold 5             # Only pods with 5+ restarts
```

**Implementation approach:**
1. List pods in specified namespace(s)
2. Analyze restart counts and reasons
3. Identify patterns (image pull failures, OOMKilled, etc.)
4. Provide specific recommendations

### 🌟 Complex Feature: Network Policy Validation

A feature that validates network policies:

```bash
# Usage
kdebug network-policy                         # Validate all policies
kdebug network-policy --policy my-policy      # Specific policy
kdebug network-policy --test-connectivity     # Test actual connectivity
```

**Implementation approach:**
1. Parse network policies
2. Validate syntax and logic
3. Check for conflicts
4. Test actual connectivity (optional)
5. Suggest improvements

---

## Getting Help

- **Questions**: Comment on your feature request issue
- **Technical Issues**: Ask in pull request discussions
- **General Guidance**: Review existing feature implementations
- **Community**: Follow project communication channels

Remember: Start small, iterate based on feedback, and focus on solving real user problems! 🚀
