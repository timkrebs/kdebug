package cluster

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

// ClusterDiagnostic handles cluster-level health checks
type ClusterDiagnostic struct {
	client *client.KubernetesClient
	output *output.OutputManager
}

// NewClusterDiagnostic creates a new cluster diagnostic
func NewClusterDiagnostic(k8sClient *client.KubernetesClient, outputMgr *output.OutputManager) *ClusterDiagnostic {
	return &ClusterDiagnostic{
		client: k8sClient,
		output: outputMgr,
	}
}

// RunDiagnostics runs all cluster-level diagnostic checks
func (c *ClusterDiagnostic) RunDiagnostics(ctx context.Context) (*output.DiagnosticReport, error) {
	// Get cluster info
	clusterInfo, err := c.client.GetClusterInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}

	// Initialize report
	report := &output.DiagnosticReport{
		Target:      "cluster",
		Timestamp:   time.Now().Format(time.RFC3339),
		ClusterInfo: clusterInfo,
		Checks:      []output.CheckResult{},
		Metadata:    make(map[string]interface{}),
	}

	// Run connectivity check
	connectivityResult := c.checkConnectivity(ctx)
	report.Checks = append(report.Checks, connectivityResult)

	// Run node health checks
	nodeResults := c.checkNodeHealth(ctx)
	report.Checks = append(report.Checks, nodeResults...)

	// Run control plane checks
	controlPlaneResults := c.checkControlPlane(ctx)
	report.Checks = append(report.Checks, controlPlaneResults...)

	// Run DNS checks
	dnsResult := c.checkDNS(ctx)
	report.Checks = append(report.Checks, dnsResult)

	// Calculate summary
	report.Summary = c.calculateSummary(report.Checks)

	return report, nil
}

// checkConnectivity tests basic connectivity to the Kubernetes API server
func (c *ClusterDiagnostic) checkConnectivity(ctx context.Context) output.CheckResult {
	start := time.Now()
	err := c.client.TestConnection(ctx)
	duration := time.Since(start)

	if err != nil {
		return output.CheckResult{
			Name:       "API Server Connectivity",
			Status:     output.StatusFailed,
			Message:    "Failed to connect to Kubernetes API server",
			Error:      err.Error(),
			Suggestion: "Check your kubeconfig file and ensure the cluster is accessible",
		}
	}

	details := map[string]string{
		"response_time": duration.String(),
		"server":        c.client.Config.Host,
	}

	message := fmt.Sprintf("Successfully connected to API server (response time: %v)", duration)
	if duration > 5*time.Second {
		return output.CheckResult{
			Name:       "API Server Connectivity",
			Status:     output.StatusWarning,
			Message:    message + " - slow response",
			Details:    details,
			Suggestion: "API server response is slow, check network connectivity",
		}
	}

	return output.CheckResult{
		Name:    "API Server Connectivity",
		Status:  output.StatusPassed,
		Message: message,
		Details: details,
	}
}

// checkNodeHealth checks the health status of all nodes
func (c *ClusterDiagnostic) checkNodeHealth(ctx context.Context) []output.CheckResult {
	var results []output.CheckResult

	nodes, err := c.client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []output.CheckResult{{
			Name:       "Node Health",
			Status:     output.StatusFailed,
			Message:    "Failed to list cluster nodes",
			Error:      err.Error(),
			Suggestion: "Check RBAC permissions for node access",
		}}
	}

	if len(nodes.Items) == 0 {
		return []output.CheckResult{{
			Name:       "Node Health",
			Status:     output.StatusFailed,
			Message:    "No nodes found in cluster",
			Suggestion: "Ensure cluster has at least one node",
		}}
	}

	// Overall node summary
	totalNodes := len(nodes.Items)
	readyNodes := 0

	var problematicNodes []string

	// Check each node
	for i := range nodes.Items {
		node := &nodes.Items[i]
		nodeReady := false

		var nodeIssues []string

		// Check node conditions
		for _, condition := range node.Status.Conditions {
			switch condition.Type {
			case corev1.NodeReady:
				if condition.Status == corev1.ConditionTrue {
					nodeReady = true
					readyNodes++
				} else {
					nodeIssues = append(nodeIssues, "NotReady")
				}
			case corev1.NodeMemoryPressure:
				if condition.Status == corev1.ConditionTrue {
					nodeIssues = append(nodeIssues, "MemoryPressure")
				}
			case corev1.NodeDiskPressure:
				if condition.Status == corev1.ConditionTrue {
					nodeIssues = append(nodeIssues, "DiskPressure")
				}
			case corev1.NodePIDPressure:
				if condition.Status == corev1.ConditionTrue {
					nodeIssues = append(nodeIssues, "PIDPressure")
				}
			case corev1.NodeNetworkUnavailable:
				if condition.Status == corev1.ConditionTrue {
					nodeIssues = append(nodeIssues, "NetworkUnavailable")
				}
			}
		}

		// Create individual node check if there are issues
		if len(nodeIssues) > 0 {
			problematicNodes = append(problematicNodes, node.Name)

			status := output.StatusWarning
			if !nodeReady {
				status = output.StatusFailed
			}

			results = append(results, output.CheckResult{
				Name:    fmt.Sprintf("Node: %s", node.Name),
				Status:  status,
				Message: fmt.Sprintf("Node has issues: %v", nodeIssues),
				Details: map[string]string{
					"node_name": node.Name,
					"issues":    fmt.Sprintf("%v", nodeIssues),
					"ready":     fmt.Sprintf("%t", nodeReady),
				},
				Suggestion: c.getNodeSuggestion(nodeIssues),
			})
		}
	}

	// Overall node health summary
	if readyNodes == totalNodes && len(problematicNodes) == 0 {
		results = append([]output.CheckResult{{
			Name:    "Node Health Overview",
			Status:  output.StatusPassed,
			Message: fmt.Sprintf("All %d nodes are healthy and ready", totalNodes),
			Details: map[string]string{
				"total_nodes": fmt.Sprintf("%d", totalNodes),
				"ready_nodes": fmt.Sprintf("%d", readyNodes),
			},
		}}, results...)
	} else {
		status := output.StatusWarning
		if readyNodes == 0 {
			status = output.StatusFailed
		}

		results = append([]output.CheckResult{{
			Name:    "Node Health Overview",
			Status:  status,
			Message: fmt.Sprintf("%d/%d nodes ready, %d nodes with issues", readyNodes, totalNodes, len(problematicNodes)),
			Details: map[string]string{
				"total_nodes":       fmt.Sprintf("%d", totalNodes),
				"ready_nodes":       fmt.Sprintf("%d", readyNodes),
				"problematic_nodes": fmt.Sprintf("%v", problematicNodes),
			},
			Suggestion: "Check individual node issues below and address node problems",
		}}, results...)
	}

	return results
}

// checkControlPlane checks the health of control plane components
func (c *ClusterDiagnostic) checkControlPlane(ctx context.Context) []output.CheckResult {
	results := make([]output.CheckResult, 0, 8) // pre-allocate with capacity

	// Check if we can access system namespaces (indicates control plane access)
	systemPods, err := c.client.Clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component in (etcd,kube-apiserver,kube-controller-manager,kube-scheduler)",
	})
	if err != nil {
		return []output.CheckResult{{
			Name:       "Control Plane Health",
			Status:     output.StatusWarning,
			Message:    "Unable to access control plane components",
			Error:      err.Error(),
			Suggestion: "Check RBAC permissions for kube-system namespace access",
		}}
	}

	if len(systemPods.Items) == 0 {
		return []output.CheckResult{{
			Name:       "Control Plane Health",
			Status:     output.StatusWarning,
			Message:    "No control plane components found (might be managed cluster)",
			Suggestion: "This might be a managed cluster (EKS, GKE, AKS) where control plane is managed",
		}}
	}

	// Group pods by component
	components := make(map[string][]*corev1.Pod)
	for i := range systemPods.Items {
		pod := &systemPods.Items[i]
		if component, exists := pod.Labels["component"]; exists {
			components[component] = append(components[component], pod)
		}
	}

	// Check each component
	allHealthy := true
	for component, pods := range components {
		runningPods := 0
		for _, pod := range pods {
			if pod.Status.Phase == corev1.PodRunning {
				runningPods++
			}
		}

		status := output.StatusPassed
		message := fmt.Sprintf("%s: %d/%d pods running", component, runningPods, len(pods))

		if runningPods == 0 {
			status = output.StatusFailed
			allHealthy = false
			message = fmt.Sprintf("%s: no pods running", component)
		} else if runningPods < len(pods) {
			status = output.StatusWarning
			allHealthy = false
		}

		results = append(results, output.CheckResult{
			Name:    fmt.Sprintf("Control Plane: %s", component),
			Status:  status,
			Message: message,
			Details: map[string]string{
				"component":    component,
				"running_pods": fmt.Sprintf("%d", runningPods),
				"total_pods":   fmt.Sprintf("%d", len(pods)),
			},
			Suggestion: c.getControlPlaneSuggestion(component, runningPods, len(pods)),
		})
	}

	// Overall control plane summary
	overallStatus := output.StatusPassed
	summaryMessage := "Control plane components are healthy"

	if !allHealthy {
		overallStatus = output.StatusWarning
		summaryMessage = "Some control plane components have issues"
	}

	results = append([]output.CheckResult{{
		Name:    "Control Plane Overview",
		Status:  overallStatus,
		Message: summaryMessage,
		Details: map[string]string{
			"components_found": fmt.Sprintf("%d", len(components)),
		},
	}}, results...)

	return results
}

// checkDNS performs basic DNS functionality check
func (c *ClusterDiagnostic) checkDNS(ctx context.Context) output.CheckResult {
	// Check CoreDNS/kube-dns pods
	dnsPods, err := c.client.Clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app in (kube-dns,coredns)",
	})
	if err != nil {
		return output.CheckResult{
			Name:       "DNS Health",
			Status:     output.StatusFailed,
			Message:    "Failed to check DNS pods",
			Error:      err.Error(),
			Suggestion: "Check RBAC permissions for kube-system namespace",
		}
	}

	if len(dnsPods.Items) == 0 {
		return output.CheckResult{
			Name:       "DNS Health",
			Status:     output.StatusFailed,
			Message:    "No DNS pods found in kube-system namespace",
			Suggestion: "Install CoreDNS or kube-dns for cluster DNS resolution",
		}
	}

	runningDNSPods := 0

	for i := range dnsPods.Items {
		pod := &dnsPods.Items[i]
		if pod.Status.Phase == corev1.PodRunning {
			runningDNSPods++
		}
	}

	details := map[string]string{
		"dns_pods_running": fmt.Sprintf("%d", runningDNSPods),
		"dns_pods_total":   fmt.Sprintf("%d", len(dnsPods.Items)),
	}

	if runningDNSPods == 0 {
		return output.CheckResult{
			Name:       "DNS Health",
			Status:     output.StatusFailed,
			Message:    "No DNS pods are running",
			Details:    details,
			Suggestion: "Check DNS pod logs and restart DNS deployment",
		}
	}

	if runningDNSPods < len(dnsPods.Items) {
		return output.CheckResult{
			Name:       "DNS Health",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("DNS partially functional: %d/%d pods running", runningDNSPods, len(dnsPods.Items)),
			Details:    details,
			Suggestion: "Some DNS pods are not running, check pod status and logs",
		}
	}

	return output.CheckResult{
		Name:    "DNS Health",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("DNS is healthy: %d/%d pods running", runningDNSPods, len(dnsPods.Items)),
		Details: details,
	}
}

// getNodeSuggestion returns appropriate suggestions for node issues
func (c *ClusterDiagnostic) getNodeSuggestion(issues []string) string {
	suggestions := []string{}

	for _, issue := range issues {
		switch issue {
		case "NotReady":
			suggestions = append(suggestions, "Check node status with 'kubectl describe node'")
		case "MemoryPressure":
			suggestions = append(suggestions, "Free up memory or add more nodes")
		case "DiskPressure":
			suggestions = append(suggestions, "Clean up disk space or add storage")
		case "PIDPressure":
			suggestions = append(suggestions, "Reduce running processes or increase PID limits")
		case "NetworkUnavailable":
			suggestions = append(suggestions, "Check network configuration and CNI")
		}
	}

	if len(suggestions) > 0 {
		return suggestions[0] // Return the first suggestion to avoid clutter
	}
	return "Check node logs and status for more details"
}

// getControlPlaneSuggestion returns suggestions for control plane issues
func (c *ClusterDiagnostic) getControlPlaneSuggestion(component string, running, total int) string {
	if running == 0 {
		return fmt.Sprintf("Restart %s component or check its configuration", component)
	}

	if running < total {
		return fmt.Sprintf("Check %s pod logs for issues", component)
	}

	return ""
}

// calculateSummary calculates the summary statistics for the checks
func (c *ClusterDiagnostic) calculateSummary(checks []output.CheckResult) output.Summary {
	summary := output.Summary{}

	for _, check := range checks {
		summary.Total++

		switch check.Status {
		case output.StatusPassed:
			summary.Passed++
		case output.StatusFailed:
			summary.Failed++
		case output.StatusWarning:
			summary.Warnings++
		case output.StatusSkipped:
			summary.Skipped++
		}
	}

	return summary
}
