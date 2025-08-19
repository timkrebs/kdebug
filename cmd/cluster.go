package cmd

import (
	"context"
	"fmt"
	"time"

	"kdebug/internal/client"
	"kdebug/internal/output"
	"kdebug/pkg/cluster"

	"github.com/spf13/cobra"
)

var (
	// Cluster command specific flags
	nodesOnly bool
	timeout   time.Duration
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Run comprehensive cluster health checks",
	Long: `Analyze the overall health of your Kubernetes cluster by checking:

• API Server connectivity and response time
• Node health and conditions (ready, memory pressure, disk pressure)
• Control plane components (etcd, scheduler, controller manager)
• DNS functionality and CoreDNS health
• Basic cluster configuration

This command provides a quick overview of cluster-wide issues that might
affect workload deployment and operation.`,
	Example: `  # Run all cluster health checks
  kdebug cluster

  # Check only node health
  kdebug cluster --nodes-only

  # Output results as JSON
  kdebug cluster --output json

  # Verbose output with detailed information
  kdebug cluster --verbose

  # Set custom timeout for checks
  kdebug cluster --timeout 30s`,
	RunE: runClusterDiagnostics,
}

func init() {
	// Add cluster command to root
	rootCmd.AddCommand(clusterCmd)

	// Cluster-specific flags
	clusterCmd.Flags().BoolVar(&nodesOnly, "nodes-only", false, "check only node health (skip control plane and DNS checks)")
	clusterCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "timeout for cluster checks")
}

// runClusterDiagnostics executes the cluster diagnostic checks
func runClusterDiagnostics(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Initialize output manager
	outputMgr := output.NewOutputManager(outputFormat, verbose)

	// Print initial info
	if verbose {
		outputMgr.PrintInfo("Starting cluster diagnostics...")
		outputMgr.PrintInfo(fmt.Sprintf("Timeout: %v", timeout))
		if kubeconfig != "" {
			outputMgr.PrintInfo(fmt.Sprintf("Using kubeconfig: %s", kubeconfig))
		}
		fmt.Println()
	}

	// Initialize Kubernetes client
	k8sClient, err := client.NewKubernetesClient(kubeconfig)
	if err != nil {
		outputMgr.PrintError("Failed to initialize Kubernetes client", err)
		return err
	}

	// Test basic connectivity first
	if err := k8sClient.TestConnection(ctx); err != nil {
		outputMgr.PrintError("Failed to connect to Kubernetes cluster", err)
		outputMgr.PrintInfo("Please check your kubeconfig and cluster connectivity")
		return err
	}

	// Initialize cluster diagnostic
	clusterDiag := cluster.NewClusterDiagnostic(k8sClient, outputMgr)

	// Run diagnostics
	report, err := clusterDiag.RunDiagnostics(ctx)
	if err != nil {
		outputMgr.PrintError("Failed to run cluster diagnostics", err)
		return err
	}

	// Filter results if nodes-only flag is set
	if nodesOnly {
		report = filterNodeChecksOnly(report)
	}

	// Print results
	if err := outputMgr.PrintReport(report); err != nil {
		outputMgr.PrintError("Failed to print diagnostic report", err)
		return err
	}

	// Print additional information based on results
	if report.Summary.Failed > 0 {
		fmt.Println()
		outputMgr.PrintWarning("Some critical issues were found that may affect cluster functionality")
		outputMgr.PrintInfo("Review the failed checks above and follow the suggested actions")

		// Exit with non-zero code if there are critical failures
		return fmt.Errorf("cluster health check failed: %d critical issues found", report.Summary.Failed)
	}

	if report.Summary.Warnings > 0 {
		fmt.Println()
		outputMgr.PrintWarning("Some warnings were found that should be addressed")
		outputMgr.PrintInfo("These issues may not immediately affect functionality but should be monitored")
	}

	if report.Summary.Failed == 0 && report.Summary.Warnings == 0 {
		fmt.Println()
		outputMgr.PrintSuccess("Cluster appears to be healthy! 🎉")
	}

	return nil
}

// filterNodeChecksOnly filters the report to include only node-related checks
func filterNodeChecksOnly(report *output.DiagnosticReport) *output.DiagnosticReport {
	var filteredChecks []output.CheckResult

	for _, check := range report.Checks {
		// Include connectivity check and node-related checks
		if check.Name == "API Server Connectivity" ||
			containsAny(check.Name, []string{"Node", "node"}) {
			filteredChecks = append(filteredChecks, check)
		}
	}

	// Update the report
	newReport := *report
	newReport.Checks = filteredChecks
	newReport.Target = "cluster (nodes only)"

	// Recalculate summary
	newSummary := output.Summary{}
	for _, check := range filteredChecks {
		newSummary.Total++
		switch check.Status {
		case output.StatusPassed:
			newSummary.Passed++
		case output.StatusFailed:
			newSummary.Failed++
		case output.StatusWarning:
			newSummary.Warnings++
		case output.StatusSkipped:
			newSummary.Skipped++
		}
	}
	newReport.Summary = newSummary

	return &newReport
}

// containsAny checks if the string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
