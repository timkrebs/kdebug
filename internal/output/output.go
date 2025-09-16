package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// Terminal formatting helpers
func colorize(text, color string) string {
	return color + text + ColorReset
}

func bold(text string) string {
	return ColorBold + text + ColorReset
}

func dim(text string) string {
	return ColorDim + text + ColorReset
}

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

// printTable prints the report as a pytest-style formatted output
func (o *OutputManager) printTable(report *DiagnosticReport) error {
	// Print clean header
	fmt.Printf("%s\n", bold("KDEBUG KUBERNETES DIAGNOSTIC REPORT"))
	fmt.Printf("Target: %s | Timestamp: %s\n", report.Target, report.Timestamp)

	// Print metadata if available
	if len(report.Metadata) > 0 {
		fmt.Printf("Resource: %v | Status: %v\n",
			report.Metadata["pod_name"],
			report.Metadata["status"])
	}

	fmt.Println()

	// Print checks in pytest style
	fmt.Printf("%s\n", bold("Diagnostic Checks:"))
	fmt.Println()

	for _, check := range report.Checks {
		status := o.formatStatusClean(check.Status)
		fmt.Printf("%-50s %s\n", check.Name, status)

		// Print detailed information if verbose and there are issues
		if o.Verbose && (check.Status == StatusFailed || check.Status == StatusWarning) {
			if check.Message != "" {
				fmt.Printf("    %s\n", dim("Message: "+check.Message))
			}
			if check.Suggestion != "" {
				fmt.Printf("    %s\n", dim("Suggestion: "+check.Suggestion))
			}
			if len(check.Details) > 0 {
				for key, value := range check.Details {
					fmt.Printf("    %s\n", dim(fmt.Sprintf("%s: %s", key, value)))
				}
			}
			fmt.Println()
		}
	}

	// Print clean summary like pytest
	fmt.Println()
	fmt.Printf("%s\n", bold("Summary:"))

	// Count and categorize results
	passed := report.Summary.Passed
	failed := report.Summary.Failed
	warnings := report.Summary.Warnings
	skipped := report.Summary.Skipped

	// Print summary line with colors
	summaryParts := []string{}
	if passed > 0 {
		summaryParts = append(summaryParts, colorize(fmt.Sprintf("%d passed", passed), ColorGreen))
	}
	if failed > 0 {
		summaryParts = append(summaryParts, colorize(fmt.Sprintf("%d failed", failed), ColorRed))
	}
	if warnings > 0 {
		summaryParts = append(summaryParts, colorize(fmt.Sprintf("%d warnings", warnings), ColorYellow))
	}
	if skipped > 0 {
		summaryParts = append(summaryParts, colorize(fmt.Sprintf("%d skipped", skipped), ColorCyan))
	}

	fmt.Printf("%s in total\n", strings.Join(summaryParts, ", "))

	// Print failed and warning details if not verbose
	if !o.Verbose && (failed > 0 || warnings > 0) {
		fmt.Println()
		fmt.Printf("%s\n", bold("Issues found:"))

		for _, check := range report.Checks {
			if check.Status == StatusFailed {
				fmt.Printf("  %s %s\n", colorize("FAILED", ColorRed), check.Name)
				if check.Message != "" {
					fmt.Printf("    %s\n", dim(check.Message))
				}
			}
		}

		for _, check := range report.Checks {
			if check.Status == StatusWarning {
				fmt.Printf("  %s %s\n", colorize("WARNING", ColorYellow), check.Name)
				if check.Message != "" {
					fmt.Printf("    %s\n", dim(check.Message))
				}
			}
		}

		fmt.Println()
		fmt.Printf("%s\n", dim("Run with --verbose for detailed information"))
	}

	return nil
}

// formatStatusClean returns a clean pytest-style status indicator
func (o *OutputManager) formatStatusClean(status CheckStatus) string {
	switch status {
	case StatusPassed:
		return colorize("PASSED", ColorGreen)
	case StatusFailed:
		return colorize("FAILED", ColorRed)
	case StatusWarning:
		return colorize("WARNING", ColorYellow)
	case StatusSkipped:
		return colorize("SKIPPED", ColorCyan)
	default:
		return colorize("UNKNOWN", ColorWhite)
	}
}

// formatStatus returns a colored status indicator (legacy function for compatibility)
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
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize("ERROR:", ColorRed), message)

	if err != nil {
		if o.Verbose {
			fmt.Fprintf(os.Stderr, "   Details: %v\n", err)
		}
	}
}

// PrintWarning prints a warning message
func (o *OutputManager) PrintWarning(message string) {
	// For structured output formats, write to stderr to avoid contaminating the output
	if o.Format == FormatJSON || o.Format == FormatYAML {
		fmt.Fprintf(os.Stderr, "%s %s\n", colorize("WARNING:", ColorYellow), message)
	} else {
		fmt.Printf("%s %s\n", colorize("WARNING:", ColorYellow), message)
	}
}

// PrintInfo prints an informational message
func (o *OutputManager) PrintInfo(message string) {
	// For structured output formats, write to stderr to avoid contaminating the output
	if o.Format == FormatJSON || o.Format == FormatYAML {
		fmt.Fprintf(os.Stderr, "%s %s\n", colorize("INFO:", ColorCyan), message)
	} else {
		fmt.Printf("%s %s\n", colorize("INFO:", ColorCyan), message)
	}
}

// PrintSuccess prints a success message
func (o *OutputManager) PrintSuccess(message string) {
	// For structured output formats, write to stderr to avoid contaminating the output
	if o.Format == FormatJSON || o.Format == FormatYAML {
		fmt.Fprintf(os.Stderr, "%s %s\n", colorize("SUCCESS:", ColorGreen), message)
	} else {
		fmt.Printf("%s %s\n", colorize("SUCCESS:", ColorGreen), message)
	}
}
