package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"kdebug/internal/client"
	"kdebug/internal/output"
	"kdebug/pkg/pod"
)

var podCmd = &cobra.Command{
	Use:   "pod [pod-name] [flags]",
	Short: "Diagnose pod-level issues and provide remediation suggestions",
	Long: `Diagnose common pod-level issues in Kubernetes clusters including:

• Pending pods (scheduling constraints, resource limits, node taints)
• Image pull errors and registry connectivity problems  
• CrashLoopBackOff detection with log analysis and hints
• RBAC permission validation for pods and service accounts
• Init container failures and misconfigurations
• Resource constraints and quality of service issues

This command analyzes pod status, events, logs, and related resources to identify
root causes and provide actionable remediation suggestions.`,
	Example: `  # Diagnose a specific pod
  kdebug pod myapp-deployment-7d4b8c6f9-x8k2l

  # Diagnose pod in specific namespace
  kdebug pod myapp-pod --namespace production

  # Diagnose all pods in a namespace
  kdebug pod --all --namespace default

  # Export detailed analysis to JSON
  kdebug pod myapp-pod --output json --verbose

  # Focus on specific diagnostic areas
  kdebug pod myapp-pod --checks=scheduling,images,rbac

  # Include detailed log analysis for crashed pods
  kdebug pod myapp-pod --include-logs --log-lines 50`,
	RunE: runPodDiagnostics,
}

func init() {
	rootCmd.AddCommand(podCmd)

	// Pod-specific flags
	podCmd.Flags().BoolP("all", "a", false, "Diagnose all pods in the specified namespace")
	podCmd.Flags().StringSlice("checks", []string{}, "Comma-separated list of checks to run (scheduling,images,rbac,logs,init-containers)")
	podCmd.Flags().Bool("include-logs", false, "Include container log analysis for failed pods")
	podCmd.Flags().Int("log-lines", 20, "Number of recent log lines to analyze (when --include-logs is enabled)")
	podCmd.Flags().Duration("timeout", 30*time.Second, "Timeout for pod diagnostics")
	podCmd.Flags().Bool("watch", false, "Watch pod status and re-run diagnostics on changes")
	podCmd.Flags().StringSlice("containers", []string{}, "Specific containers to analyze (default: all containers)")
}

func runPodDiagnostics(cmd *cobra.Command, args []string) error {
	// Parse flags
	allPods, _ := cmd.Flags().GetBool("all")
	checks, _ := cmd.Flags().GetStringSlice("checks")
	includeLogs, _ := cmd.Flags().GetBool("include-logs")
	logLines, _ := cmd.Flags().GetInt("log-lines")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	watch, _ := cmd.Flags().GetBool("watch")
	containers, _ := cmd.Flags().GetStringSlice("containers")

	// Get global flags
	outputFormat, _ := cmd.Flags().GetString("outputFormat")
	verbose, _ := cmd.Flags().GetBool("verbose")
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	namespace, _ := cmd.Flags().GetString("namespace")

	// Validate arguments
	if !allPods && len(args) == 0 {
		return fmt.Errorf("pod name is required when --all is not specified")
	}

	if allPods && len(args) > 0 {
		return fmt.Errorf("cannot specify pod name when using --all flag")
	}

	// Initialize dependencies
	outputManager := output.NewOutputManager(outputFormat, verbose)
	k8sClient, err := client.NewKubernetesClient(kubeconfig)
	if err != nil {
		outputManager.PrintError("Failed to initialize Kubernetes client", err)
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Test connectivity
	ctx := context.Background()
	if err := k8sClient.TestConnection(ctx); err != nil {
		outputManager.PrintError("Kubernetes connectivity check failed", err)
		return err
	}

	outputManager.PrintInfo("Initializing pod diagnostics...")

	// Create diagnostic configuration
	config := pod.DiagnosticConfig{
		Namespace:     namespace,
		Checks:        checks,
		IncludeLogs:   includeLogs,
		LogLines:      logLines,
		Timeout:       timeout,
		Containers:    containers,
	}

	// Initialize pod diagnostic
	diagnostic := pod.NewPodDiagnostic(k8sClient, outputManager)

	var report *output.DiagnosticReport
	var podName string

	if allPods {
		outputManager.PrintInfo(fmt.Sprintf("Analyzing all pods in namespace '%s'...", namespace))
		report, err = diagnostic.DiagnoseAllPods(config)
		if err != nil {
			return fmt.Errorf("failed to diagnose pods: %w", err)
		}
	} else {
		podName = args[0]
		outputManager.PrintInfo(fmt.Sprintf("Analyzing pod '%s' in namespace '%s'...", podName, namespace))
		
		if watch {
			return diagnostic.WatchPod(podName, config)
		}
		
		report, err = diagnostic.DiagnosePod(podName, config)
		if err != nil {
			return fmt.Errorf("failed to diagnose pod '%s': %w", podName, err)
		}
	}

	// Output results
	outputManager.PrintReport(report)

	// Print summary
	if verbose {
		outputManager.PrintInfo(fmt.Sprintf("Diagnostic completed in %s", time.Since(time.Now())))
	}

	// Return error if critical issues found
	if report.Summary.Failed > 0 {
		return fmt.Errorf("found %d critical issue(s) requiring attention", report.Summary.Failed)
	}

	return nil
}
