package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"kdebug/internal/client"
	"kdebug/internal/output"
	"kdebug/pkg/service"
)

var serviceCmd = &cobra.Command{
	Use:   "service [service-name] [flags]",
	Short: "Diagnose service-level issues and validate connectivity",
	Long: `Diagnose common service-level issues in Kubernetes clusters including:

• Service configuration validation (ports, selectors, service types)
• Endpoint health and backend pod availability
• Service selector matching with available pods
• DNS resolution for service names within the cluster
• Load balancing and traffic distribution issues
• Connectivity validation between services and pods

This command analyzes service configuration, endpoints, DNS resolution, and related
resources to identify root causes and provide actionable remediation suggestions.`,
	Example: `  # Diagnose a specific service
  kdebug service frontend --namespace production

  # Check all services in namespace
  kdebug service --all --namespace default

  # Include DNS resolution testing
  kdebug service api-gateway --test-dns

  # Check services across all namespaces
  kdebug service --all-namespaces`,
	RunE: runServiceDiagnostics,
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	// Service-specific flags
	serviceCmd.Flags().BoolP("all", "a", false, "Diagnose all services in the specified namespace")
	serviceCmd.Flags().StringSlice("checks", []string{}, "Comma-separated list of checks to run (config,selector,endpoints,ports)")
	serviceCmd.Flags().Bool("test-dns", false, "Include DNS resolution testing for the service")
	serviceCmd.Flags().Bool("all-namespaces", false, "Check services across all namespaces")
	serviceCmd.Flags().Duration("timeout", 30*time.Second, "Timeout for service diagnostics")
}

func runServiceDiagnostics(cmd *cobra.Command, args []string) error {
	// Parse flags
	allServices, _ := cmd.Flags().GetBool("all")
	checks, _ := cmd.Flags().GetStringSlice("checks")
	testDNS, _ := cmd.Flags().GetBool("test-dns")
	allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	// Get global flags
	outputFormat, _ := cmd.Flags().GetString("outputFormat")
	verbose, _ := cmd.Flags().GetBool("verbose")
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	namespace, _ := cmd.Flags().GetString("namespace")

	// Validate arguments
	if !allServices && !allNamespaces && len(args) == 0 {
		return fmt.Errorf("service name is required when --all and --all-namespaces are not specified")
	}

	if len(args) > 1 {
		return fmt.Errorf("only one service name is supported")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Initialize Kubernetes client
	kubeClient, err := client.NewKubernetesClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Initialize output manager
	outputMgr := output.NewOutputManager(outputFormat, verbose)

	// Initialize service diagnostic
	serviceDiag := service.NewServiceDiagnostic(kubeClient, outputMgr)

	// Create diagnostic configuration
	config := service.DiagnosticConfig{
		Namespace:     namespace,
		Checks:        checks,
		TestDNS:       testDNS,
		AllNamespaces: allNamespaces,
		Timeout:       timeout,
		Verbose:       verbose,
	}

	// Run diagnostics
	if allServices || allNamespaces {
		// Diagnose all services
		reports, err := serviceDiag.DiagnoseAllServices(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to diagnose services: %w", err)
		}

		// Print results
		for _, report := range reports {
			if err := outputMgr.PrintReport(report); err != nil {
				return fmt.Errorf("failed to print report: %w", err)
			}
			fmt.Println() // Add spacing between reports
		}

		// Print summary
		if len(reports) > 0 {
			printServicesSummary(outputMgr, reports)
		}

	} else {
		// Diagnose specific service
		serviceName := args[0]
		report, err := serviceDiag.DiagnoseService(ctx, serviceName, config)
		if err != nil {
			return fmt.Errorf("failed to diagnose service %s: %w", serviceName, err)
		}

		// Print results
		if err := outputMgr.PrintReport(report); err != nil {
			return fmt.Errorf("failed to print report: %w", err)
		}
	}

	return nil
}

// printServicesSummary prints a summary of all service diagnostic results.
func printServicesSummary(outputMgr *output.OutputManager, reports []*output.DiagnosticReport) {
	if outputMgr.Format != output.FormatTable {
		return // Only print summary for table format
	}

	totalServices := len(reports)
	healthyServices := 0
	unhealthyServices := 0
	warningServices := 0

	for _, report := range reports {
		if report.Summary.Failed > 0 {
			unhealthyServices++
		} else if report.Summary.Warnings > 0 {
			warningServices++
		} else {
			healthyServices++
		}
	}

	fmt.Printf("\n📊 Service Health Summary:\n")
	fmt.Printf("   Total Services: %d\n", totalServices)
	fmt.Printf("   ✅ Healthy: %d\n", healthyServices)
	fmt.Printf("   ⚠️  Warnings: %d\n", warningServices)
	fmt.Printf("   ❌ Unhealthy: %d\n", unhealthyServices)

	if unhealthyServices > 0 {
		fmt.Printf("\n🎯 Priority Actions:\n")
		fmt.Printf("   1. Investigate services with failed checks\n")
		fmt.Printf("   2. Verify pod health and readiness for services with endpoint issues\n")
		fmt.Printf("   3. Check service selectors match pod labels\n")
	}
}
