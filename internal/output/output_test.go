package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestNewOutputManager(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		verbose bool
		want    OutputFormat
	}{
		{"table format", "table", false, FormatTable},
		{"json format", "json", true, FormatJSON},
		{"yaml format", "yaml", false, FormatYAML},
		{"uppercase format", "JSON", false, FormatJSON},
		{"invalid format defaults to table", "invalid", false, FormatTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			om := NewOutputManager(tt.format, tt.verbose)
			if om.Format != tt.want {
				t.Errorf("NewOutputManager() format = %v, want %v", om.Format, tt.want)
			}
			if om.Verbose != tt.verbose {
				t.Errorf("NewOutputManager() verbose = %v, want %v", om.Verbose, tt.verbose)
			}
		})
	}
}

func TestDiagnosticReport_JSON(t *testing.T) {
	report := createTestReport()
	om := NewOutputManager("json", false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := om.PrintReport(report)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("PrintReport() error = %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Validate JSON structure
	var parsed DiagnosticReport
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Invalid JSON output: %v\nOutput: %s", err, output)
	}

	// Validate content
	if parsed.Target != report.Target {
		t.Errorf("JSON target = %v, want %v", parsed.Target, report.Target)
	}
	if len(parsed.Checks) != len(report.Checks) {
		t.Errorf("JSON checks count = %v, want %v", len(parsed.Checks), len(report.Checks))
	}
}

func TestDiagnosticReport_YAML(t *testing.T) {
	report := createTestReport()
	om := NewOutputManager("yaml", false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := om.PrintReport(report)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("PrintReport() error = %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Validate YAML structure
	var parsed DiagnosticReport
	if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Invalid YAML output: %v\nOutput: %s", err, output)
	}

	// Validate content
	if parsed.Target != report.Target {
		t.Errorf("YAML target = %v, want %v", parsed.Target, report.Target)
	}
}

func TestDiagnosticReport_Table(t *testing.T) {
	report := createTestReport()
	om := NewOutputManager("table", false)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := om.PrintReport(report)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("PrintReport() error = %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Validate table output contains expected elements
	expectedElements := []string{
		"Analyzing", // Header
		"✅",         // Passed status
		"❌",         // Failed status
		"Summary",   // Summary section
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Table output missing expected element: %s\nOutput: %s", element, output)
		}
	}
}

func TestDiagnosticReport_TableVerbose(t *testing.T) {
	report := createTestReport()
	om := NewOutputManager("table", true)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := om.PrintReport(report)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("PrintReport() error = %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Verbose output should contain cluster info and suggestions
	expectedElements := []string{
		"Cluster Information",
		"💡", // Suggestion indicator
		"test-context",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Verbose table output missing expected element: %s", element)
		}
	}
}

func TestFormatStatus(t *testing.T) {
	om := NewOutputManager("table", false)

	tests := []struct {
		status CheckStatus
		want   string
	}{
		{StatusPassed, "✅"},
		{StatusFailed, "❌"},
		{StatusWarning, "⚠️"},
		{StatusSkipped, "⏭️"},
		{CheckStatus("unknown"), "❓"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := om.formatStatus(tt.status)
			if got != tt.want {
				t.Errorf("formatStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputManager_PrintMessages(t *testing.T) {
	om := NewOutputManager("table", false)

	// Test each message type
	tests := []struct {
		name        string
		fn          func()
		expectEmpty bool
	}{
		{"PrintError", func() { om.PrintError("test error", nil) }, false},
		{"PrintWarning", func() { om.PrintWarning("test warning") }, false},
		{"PrintInfo", func() { om.PrintInfo("test info") }, false},
		{"PrintSuccess", func() { om.PrintSuccess("test success") }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For PrintError, we need to capture stderr, not stdout
			if tt.name == "PrintError" {
				// Capture stderr
				old := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				tt.fn()

				w.Close()
				os.Stderr = old

				buf := new(bytes.Buffer)
				buf.ReadFrom(r)
				output := buf.String()

				if !tt.expectEmpty && len(output) == 0 {
					t.Errorf("%s produced no output", tt.name)
				}
			} else {
				// Capture stdout for other messages
				old := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				tt.fn()

				w.Close()
				os.Stdout = old

				buf := new(bytes.Buffer)
				buf.ReadFrom(r)
				output := buf.String()

				if !tt.expectEmpty && len(output) == 0 {
					t.Errorf("%s produced no output", tt.name)
				}
			}
		})
	}
}

func TestSummaryCalculation(t *testing.T) {
	checks := []CheckResult{
		{Status: StatusPassed},
		{Status: StatusPassed},
		{Status: StatusFailed},
		{Status: StatusWarning},
		{Status: StatusSkipped},
	}

	summary := Summary{}
	for _, check := range checks {
		summary.Total++
		switch check.Status {
		case StatusPassed:
			summary.Passed++
		case StatusFailed:
			summary.Failed++
		case StatusWarning:
			summary.Warnings++
		case StatusSkipped:
			summary.Skipped++
		}
	}

	expected := Summary{
		Total:    5,
		Passed:   2,
		Failed:   1,
		Warnings: 1,
		Skipped:  1,
	}

	if summary != expected {
		t.Errorf("Summary calculation incorrect: got %+v, want %+v", summary, expected)
	}
}

// Helper function to create a test report
func createTestReport() *DiagnosticReport {
	return &DiagnosticReport{
		Target:    "test-cluster",
		Timestamp: time.Now().Format(time.RFC3339),
		ClusterInfo: map[string]string{
			"context": "test-context",
			"version": "v1.28.0",
		},
		Summary: Summary{
			Total:    3,
			Passed:   1,
			Failed:   1,
			Warnings: 1,
		},
		Checks: []CheckResult{
			{
				Name:    "API Server Connectivity",
				Status:  StatusPassed,
				Message: "Connected successfully",
			},
			{
				Name:       "Node Health",
				Status:     StatusFailed,
				Message:    "Node not ready",
				Suggestion: "Check node status",
				Details: map[string]string{
					"node": "test-node",
				},
			},
			{
				Name:    "DNS Health",
				Status:  StatusWarning,
				Message: "DNS partially functional",
			},
		},
	}
}
