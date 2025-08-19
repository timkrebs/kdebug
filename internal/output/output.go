package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
)

// CheckResult represents a single diagnostic check result
type CheckResult struct {
	Name       string            `json:"name" yaml:"name"`
	Status     CheckStatus       `json:"status" yaml:"status"`
	Message    string            `json:"message" yaml:"message"`
	Suggestion string            `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
	Details    map[string]string `json:"details,omitempty" yaml:"details,omitempty"`
	Error      string            `json:"error,omitempty" yaml:"error,omitempty"`
}

// CheckStatus represents the status of a check
type CheckStatus string

const (
	StatusPassed  CheckStatus = "PASSED"
	StatusFailed  CheckStatus = "FAILED"
	StatusWarning CheckStatus = "WARNING"
	StatusSkipped CheckStatus = "SKIPPED"
)

// DiagnosticReport represents a complete diagnostic report
type DiagnosticReport struct {
	ClusterInfo map[string]string      `json:"cluster_info" yaml:"cluster_info"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Checks      []CheckResult          `json:"checks" yaml:"checks"`
	Target      string                 `json:"target" yaml:"target"`
	Timestamp   string                 `json:"timestamp" yaml:"timestamp"`
	Summary     Summary                `json:"summary" yaml:"summary"`
}

// Summary provides a summary of check results
type Summary struct {
	Total    int `json:"total" yaml:"total"`
	Passed   int `json:"passed" yaml:"passed"`
	Failed   int `json:"failed" yaml:"failed"`
	Warnings int `json:"warnings" yaml:"warnings"`
	Skipped  int `json:"skipped" yaml:"skipped"`
}

// OutputManager handles formatting and outputting results
type OutputManager struct {
	Format  OutputFormat
	Verbose bool
}

// NewOutputManager creates a new output manager
func NewOutputManager(format string, verbose bool) *OutputManager {
	outputFormat := OutputFormat(strings.ToLower(format))

	// Validate format and default to table if invalid
	switch outputFormat {
	case FormatTable, FormatJSON, FormatYAML:
		// Valid format
	default:
		outputFormat = FormatTable
	}

	return &OutputManager{
		Format:  outputFormat,
		Verbose: verbose,
	}
}

// PrintReport prints the diagnostic report in the specified format
func (o *OutputManager) PrintReport(report *DiagnosticReport) error {
	switch o.Format {
	case FormatJSON:
		return o.printJSON(report)
	case FormatYAML:
		return o.printYAML(report)
	case FormatTable:
		return o.printTable(report)
	default:
		return o.printTable(report)
	}
}

// printJSON prints the report as JSON
func (o *OutputManager) printJSON(report *DiagnosticReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// printYAML prints the report as YAML
func (o *OutputManager) printYAML(report *DiagnosticReport) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer func() {
		if err := encoder.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing YAML encoder: %v\n", err)
		}
	}()
	return encoder.Encode(report)
}

// printTable prints the report as a formatted table
func (o *OutputManager) printTable(report *DiagnosticReport) error {
	// Print header
	fmt.Printf("🔍 Analyzing %s\n\n", report.Target)

	// Print cluster info if verbose
	if o.Verbose && len(report.ClusterInfo) > 0 {
		fmt.Println("📋 Cluster Information:")

		for key, value := range report.ClusterInfo {
			fmt.Printf("   %s: %s\n", key, value)
		}

		fmt.Println()
	}

	// Create table for checks
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Check", "Status", "Message"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// Add check results to table
	for _, check := range report.Checks {
		status := o.formatStatus(check.Status)
		message := check.Message

		// Truncate long messages for table display
		if len(message) > 80 && !o.Verbose {
			message = message[:77] + "..."
		}

		table.Append([]string{check.Name, status, message})

		// Print suggestion and details if failed and verbose
		if (check.Status == StatusFailed || check.Status == StatusWarning) && o.Verbose {
			if check.Suggestion != "" {
				table.Append([]string{"", "💡", "Suggestion: " + check.Suggestion})
			}

			if check.Error != "" {
				table.Append([]string{"", "❌", "Error: " + check.Error})
			}

			for key, value := range check.Details {
				table.Append([]string{"", "📄", fmt.Sprintf("%s: %s", key, value)})
			}
		}
	}

	table.Render()

	// Print summary
	fmt.Printf("\n📊 Summary: %d/%d checks passed", report.Summary.Passed, report.Summary.Total)

	if report.Summary.Failed > 0 {
		fmt.Printf(", %d failed", report.Summary.Failed)
	}

	if report.Summary.Warnings > 0 {
		fmt.Printf(", %d warnings", report.Summary.Warnings)
	}

	if report.Summary.Skipped > 0 {
		fmt.Printf(", %d skipped", report.Summary.Skipped)
	}

	fmt.Println()

	// Print failed checks summary
	if !o.Verbose && (report.Summary.Failed > 0 || report.Summary.Warnings > 0) {
		fmt.Println("\n🎯 Issues Found:")

		for _, check := range report.Checks {
			if check.Status == StatusFailed || check.Status == StatusWarning {
				fmt.Printf("   %s %s\n", o.formatStatus(check.Status), check.Name)

				if check.Suggestion != "" {
					fmt.Printf("      💡 %s\n", check.Suggestion)
				}
			}
		}

		fmt.Println("\nRun with --verbose for detailed information")
	}

	return nil
}

// formatStatus returns a colored status indicator
func (o *OutputManager) formatStatus(status CheckStatus) string {
	switch status {
	case StatusPassed:
		return "✅"
	case StatusFailed:
		return "❌"
	case StatusWarning:
		return "⚠️"
	case StatusSkipped:
		return "⏭️"
	default:
		return "❓"
	}
}

// PrintError prints an error message
func (o *OutputManager) PrintError(message string, err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", message)

	if err != nil {
		if o.Verbose {
			fmt.Fprintf(os.Stderr, "   Details: %v\n", err)
		}
	}
}

// PrintWarning prints a warning message
func (o *OutputManager) PrintWarning(message string) {
	fmt.Printf("⚠️  Warning: %s\n", message)
}

// PrintInfo prints an informational message
func (o *OutputManager) PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// PrintSuccess prints a success message
func (o *OutputManager) PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}
