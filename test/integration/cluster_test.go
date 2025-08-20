//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kdebug/internal/output"
)

const (
	testClusterName = "kdebug-integration-test"
	testTimeout     = 60 * time.Second
)

// TestClusterDiagnostics tests the cluster command against a real Kubernetes cluster
func TestClusterDiagnostics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Ensure kdebug binary exists
	binaryPath := getBinaryPath(t)

	// Setup test cluster if needed
	ensureTestCluster(t)
	defer func() {
		if os.Getenv("KEEP_TEST_CLUSTER") != "true" {
			cleanupTestCluster(t)
		}
	}()

	tests := []struct {
		name          string
		args          []string
		allowFailures bool // Allow non-zero exit codes for health check failures
	}{
		{
			name:          "basic cluster check",
			args:          []string{"cluster"},
			allowFailures: true, // Health checks may fail on new cluster
		},
		{
			name:          "verbose cluster check",
			args:          []string{"cluster", "--verbose"},
			allowFailures: true, // Health checks may fail on new cluster
		},
		{
			name:          "json output",
			args:          []string{"cluster", "--output", "json"},
			allowFailures: true, // Health checks may fail on new cluster
		},
		{
			name:          "yaml output",
			args:          []string{"cluster", "--output", "yaml"},
			allowFailures: true, // Health checks may fail on new cluster
		},
		{
			name:          "nodes only check",
			args:          []string{"cluster", "--nodes-only"},
			allowFailures: true, // Health checks may fail on new cluster
		},
		{
			name:          "custom timeout",
			args:          []string{"cluster", "--timeout", "30s"},
			allowFailures: true, // Health checks may fail on new cluster
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", getKubeconfig(t)))

			output, err := cmd.CombinedOutput()

			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Failed to run command: %v", err)
				}
			}

			// Validate exit code behavior
			if !tt.allowFailures && exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d. Output:\n%s", exitCode, string(output))
			}

			// For tests that allow failures, we accept any exit code but validate output
			if tt.allowFailures && exitCode != 0 {
				t.Logf("Command exited with code %d (allowed for health check failures). Output:\n%s", exitCode, string(output))
			}

			// Basic output validation
			outputStr := string(output)
			if len(outputStr) == 0 {
				t.Error("Expected non-empty output")
			}

			// Test-specific validations
			switch tt.name {
			case "json output":
				validateJSONOutput(t, outputStr)
			case "yaml output":
				validateYAMLOutput(t, outputStr)
			case "verbose cluster check":
				validateVerboseOutput(t, outputStr)
			}
		})
	}
}

// TestClusterDiagnosticsWithIssues tests kdebug behavior when cluster has known issues
func TestClusterDiagnosticsWithIssues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binaryPath := getBinaryPath(t)
	ensureTestCluster(t)
	defer cleanupTestCluster(t)

	// Test with DNS issues
	t.Run("with DNS issues", func(t *testing.T) {
		// Scale down CoreDNS
		scaleDeployment(t, "coredns", "kube-system", 0)
		defer scaleDeployment(t, "coredns", "kube-system", 2)

		// Wait a moment for the change to take effect
		time.Sleep(5 * time.Second)

		cmd := exec.Command(binaryPath, "cluster", "--verbose")
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", getKubeconfig(t)))

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should still run but might have warnings/failures
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// Exit code 1 is expected when issues are found
				if exitError.ExitCode() != 1 {
					t.Errorf("Expected exit code 1 for cluster with issues, got %d", exitError.ExitCode())
				}
			}
		}

		// Should contain DNS-related information
		if !strings.Contains(outputStr, "DNS") && !strings.Contains(outputStr, "dns") {
			t.Error("Expected DNS-related output when CoreDNS is scaled down")
		}
	})
}

// TestClusterCommandHelp tests the help functionality
func TestClusterCommandHelp(t *testing.T) {
	binaryPath := getBinaryPath(t)

	tests := []struct {
		name string
		args []string
	}{
		{"root help", []string{"--help"}},
		{"cluster help", []string{"cluster", "--help"}},
		{"version", []string{"--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Help command failed: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)
			if len(outputStr) == 0 {
				t.Error("Expected non-empty help output")
			}

			// Basic validation based on command
			switch tt.name {
			case "root help":
				if !strings.Contains(outputStr, "kdebug") {
					t.Error("Expected 'kdebug' in root help output")
				}
			case "cluster help":
				if !strings.Contains(outputStr, "cluster") {
					t.Error("Expected 'cluster' in cluster help output")
				}
			case "version":
				if !strings.Contains(outputStr, "version") {
					t.Error("Expected version information")
				}
			}
		})
	}
}

// Helper functions

func getBinaryPath(t *testing.T) string {
	// Try to find the binary in common locations
	candidates := []string{
		"../../bin/kdebug",
		"./bin/kdebug",
		"kdebug", // Assume it's in PATH
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}

	// Build the binary if it doesn't exist
	t.Log("Binary not found, building kdebug...")
	projectRoot := getProjectRoot(t)
	cmd := exec.Command("make", "build")
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build kdebug: %v", err)
	}

	binaryPath := filepath.Join(projectRoot, "bin", "kdebug")
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not found after build: %s", binaryPath)
	}

	return binaryPath
}

func getProjectRoot(t *testing.T) string {
	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

func ensureTestCluster(t *testing.T) {
	// Check if kind is available
	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kind not available, skipping integration tests")
	}

	// Check if cluster already exists
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), testClusterName) {
		t.Logf("Using existing test cluster: %s", testClusterName)
		return
	}

	// Create test cluster
	t.Logf("Creating test cluster: %s", testClusterName)
	createCmd := exec.Command("kind", "create", "cluster", "--name", testClusterName, "--wait", "60s")
	if err := createCmd.Run(); err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	// Wait for cluster to be ready
	waitForCluster(t)
}

func cleanupTestCluster(t *testing.T) {
	t.Logf("Cleaning up test cluster: %s", testClusterName)
	cmd := exec.Command("kind", "delete", "cluster", "--name", testClusterName)
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: Failed to delete test cluster: %v", err)
	}
}

func getKubeconfig(t *testing.T) string {
	// Get kubeconfig for kind cluster
	cmd := exec.Command("kind", "get", "kubeconfig", "--name", testClusterName)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get kubeconfig: %v", err)
	}

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp kubeconfig: %v", err)
	}

	if _, err := tmpFile.Write(output); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}
	tmpFile.Close()

	return tmpFile.Name()
}

func waitForCluster(t *testing.T) {
	kubeconfig := getKubeconfig(t)

	// Wait for nodes to be ready
	t.Log("Waiting for cluster nodes to be ready...")
	for i := 0; i < 60; i++ { // Increased timeout
		cmd := exec.Command("kubectl", "--kubeconfig", kubeconfig, "get", "nodes", "--no-headers")
		output, err := cmd.Output()
		if err == nil {
			// Check if all nodes are Ready
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			allReady := true
			for _, line := range lines {
				if !strings.Contains(line, "Ready") {
					allReady = false
					break
				}
			}
			if allReady {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	// Wait for system pods to be ready
	t.Log("Waiting for system pods to be ready...")
	for i := 0; i < 60; i++ {
		cmd := exec.Command("kubectl", "--kubeconfig", kubeconfig, "get", "pods", "-n", "kube-system", "--no-headers")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			allRunning := true
			for _, line := range lines {
				if line != "" && !strings.Contains(line, "Running") && !strings.Contains(line, "Completed") {
					allRunning = false
					break
				}
			}
			if allRunning && len(lines) > 0 {
				t.Log("Cluster is ready!")
				return
			}
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatal("Cluster did not become ready in time")
}

func scaleDeployment(t *testing.T, name, namespace string, replicas int) {
	kubeconfig := getKubeconfig(t)
	cmd := exec.Command("kubectl", "--kubeconfig", kubeconfig, "scale", "deployment", name,
		"--namespace", namespace, "--replicas", fmt.Sprintf("%d", replicas))
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to scale deployment %s: %v", name, err)
	}
}

func validateJSONOutput(t *testing.T, outputStr string) {
	// Extract JSON from the output (might be mixed with other messages)
	jsonStr := extractJSON(outputStr)
	if jsonStr == "" {
		t.Errorf("No JSON found in output: %s", outputStr)
		return
	}

	var report output.DiagnosticReport
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		t.Errorf("Invalid JSON output: %v\nExtracted JSON: %s\nFull Output: %s", err, jsonStr, outputStr)
		return
	}

	// Basic structure validation
	if report.Target == "" {
		t.Error("JSON output missing target field")
	}
	if report.Timestamp == "" {
		t.Error("JSON output missing timestamp field")
	}
	if len(report.Checks) == 0 {
		t.Error("JSON output missing checks")
	}
}

// extractJSON attempts to extract JSON from mixed output
func extractJSON(output string) string {
	// Look for JSON starting with { and ending with }
	start := strings.Index(output, "{")
	if start == -1 {
		return ""
	}
	
	// Find the matching closing brace
	braceCount := 0
	end := -1
	for i := start; i < len(output); i++ {
		if output[i] == '{' {
			braceCount++
		} else if output[i] == '}' {
			braceCount--
			if braceCount == 0 {
				end = i + 1
				break
			}
		}
	}
	
	if end == -1 {
		return ""
	}
	
	return output[start:end]
}

func validateYAMLOutput(t *testing.T, output string) {
	// Basic YAML validation - should contain expected fields
	expectedFields := []string{"target:", "timestamp:", "checks:", "summary:"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("YAML output missing expected field: %s", field)
		}
	}
}

func validateVerboseOutput(t *testing.T, output string) {
	// Verbose output should contain additional information
	expectedPatterns := []string{
		"Starting cluster diagnostics",
		"Timeout:",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Verbose output missing expected pattern: %s", pattern)
		}
	}
}
