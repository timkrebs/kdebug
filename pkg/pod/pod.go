// Package pod provides diagnostic capabilities for Kubernetes pod-level issues.
//
// This package implements comprehensive checks for common pod problems including:
//   - Scheduling issues: insufficient resources, node taints, affinity constraints
//   - Image problems: pull errors, registry connectivity, authentication failures
//   - Runtime issues: CrashLoopBackOff, container startup failures, resource limits
//   - RBAC problems: permission validation for pods and service accounts
//   - Init container failures: startup errors, dependency issues, misconfigurations
//   - Network issues: DNS resolution, service connectivity, port accessibility
//
// The diagnostics help users identify root causes and provide actionable recommendations
// for resolving pod-related issues in Kubernetes clusters.
package pod

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

// PodDiagnostic performs diagnostic checks for pod-level issues.
// It analyzes pod status, events, logs, and related resources to identify
// root causes and provide actionable recommendations for resolution.
type PodDiagnostic struct {
	// client provides access to the Kubernetes API
	client *client.KubernetesClient

	// output handles formatting and display of diagnostic results
	output *output.OutputManager
}

// DiagnosticConfig contains configuration options for pod diagnostics.
type DiagnosticConfig struct {
	// Namespace specifies the target namespace for diagnostics
	Namespace string

	// Checks specifies which diagnostic checks to run (empty = all checks)
	Checks []string

	// IncludeLogs enables log analysis for failed containers
	IncludeLogs bool

	// LogLines specifies number of recent log lines to analyze
	LogLines int

	// Timeout specifies maximum time for diagnostic operations
	Timeout time.Duration

	// Containers specifies which containers to analyze (empty = all containers)
	Containers []string
}

// PodInfo contains comprehensive information about a pod for diagnostics.
type PodInfo struct {
	Pod               *corev1.Pod
	Events            []corev1.Event
	ContainerLogs     map[string]string
	ServiceAccount    *corev1.ServiceAccount
	Secrets           []corev1.Secret
	ConfigMaps        []corev1.ConfigMap
	PersistentVolumes []corev1.PersistentVolume
	Node              *corev1.Node
}

// NewPodDiagnostic creates a new pod diagnostic instance.
func NewPodDiagnostic(client *client.KubernetesClient, output *output.OutputManager) *PodDiagnostic {
	return &PodDiagnostic{
		client: client,
		output: output,
	}
}

// DiagnosePod performs comprehensive diagnostics on a single pod.
func (d *PodDiagnostic) DiagnosePod(podName string, config DiagnosticConfig) (*output.DiagnosticReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Gather pod information
	podInfo, err := d.gatherPodInfo(ctx, podName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to gather pod information: %w", err)
	}

	// Run diagnostic checks
	checks := d.runDiagnosticChecks(ctx, podInfo, config)

	// Calculate summary
	summary := d.calculateSummary(checks)

	// Create report
	report := &output.DiagnosticReport{
		Target:    fmt.Sprintf("pod/%s", podName),
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    checks,
		Summary:   summary,
	}

	return report, nil
}

// DiagnoseAllPods performs diagnostics on all pods in the specified namespace.
func (d *PodDiagnostic) DiagnoseAllPods(config DiagnosticConfig) (*output.DiagnosticReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// List all pods in namespace
	pods, err := d.client.Clientset.CoreV1().Pods(config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return &output.DiagnosticReport{
			Target:    fmt.Sprintf("pods/namespace=%s", config.Namespace),
			Timestamp: time.Now().Format(time.RFC3339),
			Checks: []output.CheckResult{
				{
					Name:    "Pod Discovery",
					Status:  "SKIPPED",
					Message: fmt.Sprintf("No pods found in namespace '%s'", config.Namespace),
				},
			},
			Summary: output.Summary{Total: 1, Skipped: 1},
		}, nil
	}

	var allChecks []output.CheckResult

	// Diagnose each pod
	for i := range pods.Items {
		pod := &pods.Items[i]

		d.output.PrintInfo(fmt.Sprintf("Executing diagnostic analysis for pod '%s' in namespace '%s'", pod.Name, config.Namespace))

		podInfo, err := d.gatherPodInfoFromPod(ctx, pod, config)
		if err != nil {
			d.output.PrintWarning(fmt.Sprintf("Failed to analyze pod %s: %v", pod.Name, err))
			continue
		}

		podChecks := d.runDiagnosticChecks(ctx, podInfo, config)

		// Prefix check names with pod name for clarity
		for j := range podChecks {
			podChecks[j].Name = fmt.Sprintf("Pod %s: %s", pod.Name, podChecks[j].Name)
		}

		allChecks = append(allChecks, podChecks...)
	}

	// Calculate summary
	summary := d.calculateSummary(allChecks)

	// Create report
	report := &output.DiagnosticReport{
		Target:    fmt.Sprintf("pods/namespace=%s", config.Namespace),
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    allChecks,
		Summary:   summary,
	}

	return report, nil
}

// WatchPod watches a pod and re-runs diagnostics when changes occur.
func (d *PodDiagnostic) WatchPod(podName string, config DiagnosticConfig) error {
	d.output.PrintInfo(fmt.Sprintf("Watching pod '%s' for changes...", podName))

	watchlist := &metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", podName).String(),
	}

	watcher, err := d.client.Clientset.CoreV1().Pods(config.Namespace).Watch(context.Background(), *watchlist)
	if err != nil {
		return fmt.Errorf("failed to watch pod: %w", err)
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified, watch.Added:
			d.output.PrintInfo("Pod changed, re-running diagnostics...")

			report, err := d.DiagnosePod(podName, config)
			if err != nil {
				d.output.PrintError(fmt.Sprintf("Failed to diagnose pod: %v", err), err)
				continue
			}

			d.output.PrintReport(report)
			d.output.PrintInfo("Waiting for next change...")

		case watch.Deleted:
			d.output.PrintWarning("Pod was deleted")
			return nil

		case watch.Error:
			d.output.PrintError("Watch error occurred", fmt.Errorf("watch error"))
			return fmt.Errorf("watch error")
		}
	}

	return nil
}

// gatherPodInfo collects comprehensive information about a pod and related resources.
func (d *PodDiagnostic) gatherPodInfo(ctx context.Context, podName string, config DiagnosticConfig) (*PodInfo, error) {
	// Get the pod
	pod, err := d.client.Clientset.CoreV1().Pods(config.Namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return d.gatherPodInfoFromPod(ctx, pod, config)
}

// gatherPodInfoFromPod collects information about a pod from an existing pod object.
func (d *PodDiagnostic) gatherPodInfoFromPod(ctx context.Context, pod *corev1.Pod, config DiagnosticConfig) (*PodInfo, error) {
	info := &PodInfo{
		Pod:           pod,
		ContainerLogs: make(map[string]string),
	}

	// Get pod events
	events, err := d.client.Clientset.CoreV1().Events(config.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", pod.Name),
	})
	if err != nil {
		d.output.PrintWarning(fmt.Sprintf("Failed to get events for pod %s: %v", pod.Name, err))
	} else {
		info.Events = events.Items
	}

	// Get service account if specified
	if pod.Spec.ServiceAccountName != "" {
		sa, err := d.client.Clientset.CoreV1().ServiceAccounts(config.Namespace).Get(ctx, pod.Spec.ServiceAccountName, metav1.GetOptions{})
		if err != nil {
			d.output.PrintWarning(fmt.Sprintf("Failed to get service account %s: %v", pod.Spec.ServiceAccountName, err))
		} else {
			info.ServiceAccount = sa
		}
	}

	// Get node information
	if pod.Spec.NodeName != "" {
		node, err := d.client.Clientset.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			d.output.PrintWarning(fmt.Sprintf("Failed to get node %s: %v", pod.Spec.NodeName, err))
		} else {
			info.Node = node
		}
	}

	// Get container logs if requested and pod is failing
	if config.IncludeLogs && d.isPodFailing(pod) {
		d.gatherContainerLogs(ctx, info, config)
	}

	return info, nil
}

// gatherContainerLogs collects logs from containers in the pod.
func (d *PodDiagnostic) gatherContainerLogs(ctx context.Context, info *PodInfo, config DiagnosticConfig) {
	pod := info.Pod

	containers := config.Containers
	if len(containers) == 0 {
		// Get all containers if none specified
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Name)
		}
		for _, container := range pod.Spec.InitContainers {
			containers = append(containers, container.Name)
		}
	}

	logOptions := &corev1.PodLogOptions{
		TailLines: int64ptr(config.LogLines),
		Previous:  false, // Get current logs first
	}

	for _, containerName := range containers {
		logOptions.Container = containerName

		// Try current logs first
		logs, err := d.client.Clientset.CoreV1().Pods(config.Namespace).GetLogs(pod.Name, logOptions).Stream(ctx)
		if err != nil {
			// Try previous logs if current logs fail
			logOptions.Previous = true
			logs, err = d.client.Clientset.CoreV1().Pods(config.Namespace).GetLogs(pod.Name, logOptions).Stream(ctx)
			if err != nil {
				d.output.PrintWarning(fmt.Sprintf("Failed to get logs for container %s: %v", containerName, err))
				continue
			}
		}

		defer logs.Close()

		buf := make([]byte, 2048)
		var logContent strings.Builder
		for {
			n, err := logs.Read(buf)
			if n > 0 {
				logContent.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}

		info.ContainerLogs[containerName] = logContent.String()

		// Reset for next container
		logOptions.Previous = false
	}
}

// isPodFailing determines if a pod is in a failing state.
func (d *PodDiagnostic) isPodFailing(pod *corev1.Pod) bool {
	if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodPending {
		return true
	}

	// Check container statuses
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil {
			reason := status.State.Waiting.Reason
			if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" ||
				reason == "ErrImagePull" || reason == "InvalidImageName" {
				return true
			}
		}
		if status.RestartCount > 0 {
			return true
		}
	}

	return false
}

// Helper function to create an int64 pointer
func int64ptr(i int) *int64 {
	val := int64(i)
	return &val
}

// calculateSummary computes summary statistics from diagnostic check results.
func (d *PodDiagnostic) calculateSummary(checks []output.CheckResult) output.Summary {
	summary := output.Summary{}

	for _, check := range checks {
		switch check.Status {
		case "PASSED":
			summary.Passed++
		case "FAILED":
			summary.Failed++
		case "WARNING":
			summary.Warnings++
		case "SKIPPED":
			summary.Skipped++
		}
		summary.Total++
	}

	return summary
}
