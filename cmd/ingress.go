package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"kdebug/internal/client"
	"kdebug/internal/output"
	"kdebug/pkg/ingress"
)

var ingressCmd = &cobra.Command{
	Use:   "ingress [INGRESS_NAME]",
	Short: "Diagnose Kubernetes ingress resources",
	Long: `Diagnose Kubernetes ingress resources for common issues and misconfigurations.

This command analyzes ingress resources and their dependencies including:
- Ingress existence and configuration
- Backend service availability and configuration
- Service endpoint health
- SSL/TLS certificate validation
- Controller discovery and status

Examples:
  # Diagnose a specific ingress
  kdebug ingress my-ingress

  # Diagnose all ingress resources in current namespace
  kdebug ingress --all

  # Diagnose all ingress resources across all namespaces
  kdebug ingress --all --all-namespaces

  # Run specific checks only
  kdebug ingress my-ingress --checks config,backends,ssl

  # Output in JSON format
  kdebug ingress my-ingress --output json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIngressDiagnosis,
}

var (
	ingressAll           bool
	ingressAllNamespaces bool
	ingressChecks        []string
	ingressOutputFormat  string
	ingressVerbose       bool
	ingressTimeout       time.Duration
)

func init() {
	rootCmd.AddCommand(ingressCmd)

	// Flags
	ingressCmd.Flags().BoolVar(&ingressAll, "all", false, "Diagnose all ingress resources in namespace(s)")
	ingressCmd.Flags().BoolVar(&ingressAllNamespaces, "all-namespaces", false, "Analyze ingress resources across all namespaces")
	ingressCmd.Flags().StringSliceVar(&ingressChecks, "checks", []string{}, "Comma-separated list of checks to run (existence,config,backends,endpoints,ssl)")
	ingressCmd.Flags().StringVarP(&ingressOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	ingressCmd.Flags().BoolVarP(&ingressVerbose, "verbose", "v", false, "Enable verbose output")
	ingressCmd.Flags().DurationVar(&ingressTimeout, "timeout", 30*time.Second, "Timeout for diagnosis operations")

	// Add aliases for convenience
	ingressCmd.Aliases = []string{"ing", "ingresses"}
}

func runIngressDiagnosis(cmd *cobra.Command, args []string) error {
	// Get global flags
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	namespace, _ := cmd.Flags().GetString("namespace")

	// Set defaults
	if namespace == "" {
		namespace = "default"
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ingressTimeout)
	defer cancel()

	// Initialize Kubernetes client
	kubeClient, err := client.NewKubernetesClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Initialize output manager
	outputMgr := output.NewOutputManager(ingressOutputFormat, ingressVerbose)

	// Initialize ingress diagnostic
	ingressDiag := ingress.NewIngressDiagnostic(kubeClient, outputMgr)

	// Prepare diagnostic configuration
	config := ingress.DiagnosticConfig{
		Namespace:     namespace,
		AllNamespaces: ingressAllNamespaces,
		All:           ingressAll,
		Checks:        ingressChecks,
		Timeout:       ingressTimeout,
	}

	// Handle specific ingress vs. all ingresses
	if len(args) == 1 {
		// Diagnose specific ingress
		ingressName := args[0]
		config.IngressName = ingressName

		report, err := ingressDiag.DiagnoseIngress(ctx, ingressName, config)
		if err != nil {
			return fmt.Errorf("failed to diagnose ingress %s: %w", ingressName, err)
		}

		// Print the report
		if err := outputMgr.PrintReport(report); err != nil {
			return fmt.Errorf("failed to print report: %w", err)
		}

	} else if ingressAll {
		// Diagnose all ingresses
		reports, err := ingressDiag.DiagnoseAllIngresses(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to diagnose ingresses: %w", err)
		}

		// Print each report
		for _, report := range reports {
			if err := outputMgr.PrintReport(report); err != nil {
				return fmt.Errorf("failed to print report: %w", err)
			}
			// Add separator between reports for table format
			if ingressOutputFormat == "table" && len(reports) > 1 {
				fmt.Println(output.ColorDim + strings.Repeat("─", 80) + output.ColorReset)
			}
		}

		// Print summary if multiple ingresses were analyzed
		if len(reports) > 1 {
			printIngressSummary(reports)
		}

	} else {
		return fmt.Errorf("please specify an ingress name or use --all flag")
	}

	return nil
}

// printIngressSummary prints a summary of multiple ingress diagnoses
func printIngressSummary(reports []*output.DiagnosticReport) {
	total := len(reports)
	var totalChecks, passed, failed, warnings, skipped int

	for _, report := range reports {
		totalChecks += report.Summary.Total
		passed += report.Summary.Passed
		failed += report.Summary.Failed
		warnings += report.Summary.Warnings
		skipped += report.Summary.Skipped
	}

	fmt.Printf("\n📊 Ingress Diagnostics Summary:\n")
	fmt.Printf("   Total Ingresses: %d\n", total)
	fmt.Printf("   Total Checks: %d\n", totalChecks)
	fmt.Printf("   ✅ Passed: %d\n", passed)
	fmt.Printf("   ❌ Failed: %d\n", failed)
	fmt.Printf("   ⚠️  Warnings: %d\n", warnings)
	fmt.Printf("   ⏭️  Skipped: %d\n", skipped)

	// Calculate health percentage
	if totalChecks > 0 {
		healthPercent := float64(passed) / float64(totalChecks) * 100
		fmt.Printf("   Health Score: %.1f%%\n", healthPercent)
	}
}
