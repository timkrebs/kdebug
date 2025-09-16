package pod

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"kdebug/internal/output"
)

// runDiagnosticChecks executes all diagnostic checks for a pod.
func (d *PodDiagnostic) runDiagnosticChecks(ctx context.Context, info *PodInfo, config DiagnosticConfig) []output.CheckResult {
	// Determine which checks to run
	checkTypes := config.Checks
	if len(checkTypes) == 0 {
		// Run all checks if none specified
		checkTypes = []string{"basic", "scheduling", "images", "rbac", "logs", "init-containers", "resources", "network"}
	}

	// Pre-allocate slice with estimated capacity
	checks := make([]output.CheckResult, 0, len(checkTypes)*3)

	for _, checkType := range checkTypes {
		switch checkType {
		case "basic":
			checks = append(checks, d.checkPodBasicStatus(info))
		case "scheduling":
			checks = append(checks, d.checkPodScheduling(ctx, info)...)
		case "images":
			checks = append(checks, d.checkImageIssues(info)...)
		case "rbac":
			checks = append(checks, d.checkRBACPermissions(ctx, info)...)
		case "logs":
			if config.IncludeLogs {
				checks = append(checks, d.checkContainerLogs(info)...)
			}
		case "init-containers":
			checks = append(checks, d.checkInitContainers(info)...)
		case "resources":
			checks = append(checks, d.checkResourceConstraints(info)...)
		case "network":
			checks = append(checks, d.checkNetworkIssues(info)...)
		}
	}

	return checks
}

// checkPodBasicStatus performs basic pod status checks.
func (d *PodDiagnostic) checkPodBasicStatus(info *PodInfo) output.CheckResult {
	pod := info.Pod

	switch pod.Status.Phase {
	case corev1.PodRunning:
		// Check if all containers are ready
		readyContainers := 0
		totalContainers := len(pod.Status.ContainerStatuses)

		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				readyContainers++
			}
		}

		if readyContainers == totalContainers {
			return output.CheckResult{
				Name:    "Pod Status",
				Status:  output.StatusPassed,
				Message: fmt.Sprintf("Pod is running with %d/%d containers ready", readyContainers, totalContainers),
				Details: map[string]string{
					"phase": string(pod.Status.Phase),
					"ready": formatConditionStatus(getPodCondition(pod, corev1.PodReady)),
				},
			}
		}

		return output.CheckResult{
			Name:       "Pod Status",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Pod is running but only %d/%d containers are ready", readyContainers, totalContainers),
			Suggestion: "Check individual container statuses and logs",
			Details: map[string]string{
				"phase": string(pod.Status.Phase),
			},
		}

	case corev1.PodPending:
		return output.CheckResult{
			Name:       "Pod Status",
			Status:     output.StatusFailed,
			Message:    "Pod is stuck in Pending state",
			Suggestion: "Check scheduling constraints, resource availability, and node conditions",
			Details: map[string]string{
				"phase": string(pod.Status.Phase),
				"age":   time.Since(pod.CreationTimestamp.Time).Truncate(time.Second).String(),
			},
		}

	case corev1.PodFailed:
		return output.CheckResult{
			Name:       "Pod Status",
			Status:     output.StatusFailed,
			Message:    "Pod has failed",
			Suggestion: "Check container exit codes and logs for failure reasons",
			Details: map[string]string{
				"phase":  string(pod.Status.Phase),
				"reason": pod.Status.Reason,
			},
		}

	case corev1.PodSucceeded:
		return output.CheckResult{
			Name:    "Pod Status",
			Status:  output.StatusPassed,
			Message: "Pod completed successfully",
			Details: map[string]string{
				"phase": string(pod.Status.Phase),
			},
		}

	default:
		return output.CheckResult{
			Name:       "Pod Status",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Pod is in unknown phase: %s", pod.Status.Phase),
			Suggestion: "Investigate pod events and cluster status",
			Details: map[string]string{
				"phase": string(pod.Status.Phase),
			},
		}
	}
}

// checkPodScheduling checks for pod scheduling issues.
func (d *PodDiagnostic) checkPodScheduling(_ context.Context, info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, 3) // Pre-allocate for expected number of checks
	pod := info.Pod

	// Check if pod is scheduled
	if pod.Spec.NodeName == "" {
		checks = append(checks, output.CheckResult{
			Name:       "Pod Scheduling",
			Status:     output.StatusFailed,
			Message:    "Pod is not scheduled to any node",
			Suggestion: d.getSchedulingSuggestion(info),
			Details: map[string]string{
				"scheduled": "false",
				"message":   "Pod remains unscheduled - check resource requirements and node availability",
			},
		})
	} else {
		checks = append(checks, output.CheckResult{
			Name:    "Pod Scheduling",
			Status:  output.StatusPassed,
			Message: fmt.Sprintf("Pod is scheduled on node '%s'", pod.Spec.NodeName),
			Details: map[string]string{
				"node":      pod.Spec.NodeName,
				"scheduled": "true",
			},
		})

		// If we have node info, check node conditions
		if info.Node != nil {
			checks = append(checks, d.checkNodeConditions(info.Node))
		}
	}

	// Check resource requests vs node capacity
	if pod.Spec.NodeName != "" && info.Node != nil {
		checks = append(checks, d.checkResourceFit(pod, info.Node))
	}

	return checks
}

// checkImageIssues checks for image pull problems.
func (d *PodDiagnostic) checkImageIssues(info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, len(info.Pod.Status.ContainerStatuses)+len(info.Pod.Status.InitContainerStatuses)) // Pre-allocate based on containers
	pod := info.Pod

	// Check container image statuses
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil {
			reason := status.State.Waiting.Reason
			message := status.State.Waiting.Message

			switch reason {
			case "ErrImagePull", "ImagePullBackOff":
				checks = append(checks, output.CheckResult{
					Name:       fmt.Sprintf("Container %s - Image Pull", status.Name),
					Status:     output.StatusFailed,
					Message:    fmt.Sprintf("Failed to pull image: %s", message),
					Suggestion: d.getImagePullSuggestion(status.Image, message),
					Details: map[string]string{
						"image":  status.Image,
						"reason": reason,
					},
				})

			case "InvalidImageName":
				checks = append(checks, output.CheckResult{
					Name:       fmt.Sprintf("Container %s - Image Name", status.Name),
					Status:     output.StatusFailed,
					Message:    fmt.Sprintf("Invalid image name: %s", status.Image),
					Suggestion: "Verify the image name format and registry URL",
					Details: map[string]string{
						"image": status.Image,
					},
				})

			default:
				if strings.Contains(strings.ToLower(reason), "image") {
					checks = append(checks, output.CheckResult{
						Name:       fmt.Sprintf("Container %s - Image Issue", status.Name),
						Status:     output.StatusWarning,
						Message:    fmt.Sprintf("Image-related issue: %s", reason),
						Suggestion: "Check image availability and registry connectivity",
						Details: map[string]string{
							"image":  status.Image,
							"reason": reason,
						},
					})
				}
			}
		} else {
			switch {
			case status.State.Running != nil, status.State.Terminated != nil:
				// Image pulled successfully
				checks = append(checks, output.CheckResult{
					Name:    fmt.Sprintf("Container %s - Image Pull", status.Name),
					Status:  output.StatusPassed,
					Message: "Image pulled successfully",
					Details: map[string]string{
						"image": status.Image,
					},
				})
			}
		}
	}

	// If no container statuses yet, check init containers
	if len(pod.Status.ContainerStatuses) == 0 {
		for _, status := range pod.Status.InitContainerStatuses {
			if status.State.Waiting != nil && strings.Contains(status.State.Waiting.Reason, "Image") {
				checks = append(checks, output.CheckResult{
					Name:       fmt.Sprintf("Init Container %s - Image Pull", status.Name),
					Status:     output.StatusFailed,
					Message:    fmt.Sprintf("Failed to pull image: %s", status.State.Waiting.Message),
					Suggestion: d.getImagePullSuggestion(status.Image, status.State.Waiting.Message),
					Details: map[string]string{
						"image":  status.Image,
						"reason": status.State.Waiting.Reason,
					},
				})
			}
		}
	}

	return checks
}

// checkRBACPermissions validates RBAC permissions for the pod.
func (d *PodDiagnostic) checkRBACPermissions(_ context.Context, info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, 2) // Pre-allocate for expected number of checks
	pod := info.Pod

	// Check if service account exists
	switch {
	case pod.Spec.ServiceAccountName == "":
		checks = append(checks, output.CheckResult{
			Name:    "RBAC - Service Account",
			Status:  output.StatusPassed,
			Message: "Using default service account",
			Details: map[string]string{
				"serviceAccount": "default",
			},
		})
	case info.ServiceAccount == nil:
		checks = append(checks, output.CheckResult{
			Name:       "RBAC - Service Account",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Service account '%s' not found", pod.Spec.ServiceAccountName),
			Suggestion: fmt.Sprintf("Create service account: kubectl create serviceaccount %s -n %s", pod.Spec.ServiceAccountName, pod.Namespace),
			Details: map[string]string{
				"serviceAccount": pod.Spec.ServiceAccountName,
			},
		})
	default:
		checks = append(checks, output.CheckResult{
			Name:    "RBAC - Service Account",
			Status:  output.StatusPassed,
			Message: fmt.Sprintf("Service account '%s' exists", pod.Spec.ServiceAccountName),
			Details: map[string]string{
				"serviceAccount": pod.Spec.ServiceAccountName,
			},
		})

		// Check for common RBAC issues in events
		checks = append(checks, d.checkRBACEvents(info)...)
	}

	return checks
}

// checkContainerLogs analyzes container logs for common issues.
func (d *PodDiagnostic) checkContainerLogs(info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, len(info.Pod.Status.ContainerStatuses)) // Pre-allocate based on container count

	if len(info.ContainerLogs) == 0 {
		checks = append(checks, output.CheckResult{
			Name:    "Container Logs",
			Status:  output.StatusSkipped,
			Message: "No container logs available",
		})
		return checks
	}

	for containerName, logs := range info.ContainerLogs {
		if logs == "" {
			checks = append(checks, output.CheckResult{
				Name:    fmt.Sprintf("Container %s - Logs", containerName),
				Status:  output.StatusWarning,
				Message: "No log output available",
				Details: map[string]string{
					"message": "Container may not have started or produced any logs",
				},
			})
			continue
		}

		// Analyze logs for common patterns
		logCheck := d.analyzeContainerLogs(containerName, logs)
		checks = append(checks, logCheck)

		// Check for crash loop indicators
		if d.isContainerCrashLooping(info.Pod, containerName) {
			crashCheck := d.analyzeCrashLoopBackOff(containerName, logs, info.Pod)
			checks = append(checks, crashCheck)
		}
	}

	return checks
}

// checkInitContainers checks init container status and issues.
func (d *PodDiagnostic) checkInitContainers(info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, len(info.Pod.Spec.InitContainers)+1) // Pre-allocate based on init containers
	pod := info.Pod

	if len(pod.Spec.InitContainers) == 0 {
		checks = append(checks, output.CheckResult{
			Name:    "Init Containers",
			Status:  output.StatusSkipped,
			Message: "No init containers defined",
		})
		return checks
	}

	// Check each init container status
	for i, initContainer := range pod.Spec.InitContainers {
		if i < len(pod.Status.InitContainerStatuses) {
			status := pod.Status.InitContainerStatuses[i]
			checks = append(checks, d.checkInitContainerStatus(initContainer.Name, status))
		} else {
			checks = append(checks, output.CheckResult{
				Name:    fmt.Sprintf("Init Container %s", initContainer.Name),
				Status:  output.StatusWarning,
				Message: "Init container status not available",
			})
		}
	}

	return checks
}

// checkResourceConstraints analyzes resource requests, limits, and QoS.
func (d *PodDiagnostic) checkResourceConstraints(info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, 3) // Pre-allocate for expected resource checks
	pod := info.Pod

	// Check QoS class
	checks = append(checks, d.checkQoSClass(pod))

	// Check resource requests and limits
	checks = append(checks, d.checkResourceRequests(pod))

	// Check for resource-related events
	checks = append(checks, d.checkResourceEvents(info))

	return checks
}

// checkNetworkIssues analyzes network-related problems.
func (d *PodDiagnostic) checkNetworkIssues(info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, 5) // Pre-allocate for expected network checks
	pod := info.Pod

	// Check if pod has IP address
	if pod.Status.PodIP == "" {
		checks = append(checks, output.CheckResult{
			Name:       "Network - Pod IP",
			Status:     output.StatusFailed,
			Message:    "Pod has no IP address assigned",
			Suggestion: "Check CNI plugin status and network configuration",
		})
	} else {
		checks = append(checks, output.CheckResult{
			Name:    "Network - Pod IP",
			Status:  output.StatusPassed,
			Message: fmt.Sprintf("Pod IP assigned: %s", pod.Status.PodIP),
			Details: map[string]string{
				"podIP": pod.Status.PodIP,
			},
		})
	}

	// Check DNS configuration
	checks = append(checks, d.checkDNSConfiguration(pod))

	// Check for network-related events
	checks = append(checks, d.checkNetworkEvents(info))

	return checks
}

// Helper functions for specific checks

func (d *PodDiagnostic) getSchedulingSuggestion(info *PodInfo) string {
	suggestions := []string{}

	// Check events for scheduling clues
	for _, event := range info.Events {
		if event.Reason == "FailedScheduling" {
			if strings.Contains(event.Message, "Insufficient") {
				suggestions = append(suggestions, "Insufficient resources - scale cluster or reduce resource requests")
			}
			if strings.Contains(event.Message, "node(s) had taint") {
				suggestions = append(suggestions, "Node taints prevent scheduling - add tolerations or remove taints")
			}
			if strings.Contains(event.Message, "didn't match node selector") {
				suggestions = append(suggestions, "Node selector mismatch - verify node labels")
			}
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Check node availability, resource requirements, and scheduling constraints")
	}

	return strings.Join(suggestions, "; ")
}

func (d *PodDiagnostic) getImagePullSuggestion(image, message string) string {
	if strings.Contains(message, "not found") || strings.Contains(message, "manifest unknown") {
		return fmt.Sprintf("Image not found - verify image name and tag: %s", image)
	}
	if strings.Contains(message, "unauthorized") || strings.Contains(message, "authentication") {
		return "Check registry credentials and image pull secrets"
	}
	if strings.Contains(message, "timeout") || strings.Contains(message, "connection") {
		return "Check network connectivity to registry and DNS resolution"
	}
	if strings.Contains(message, "pull rate limit") {
		return "Registry rate limit exceeded - configure pull secrets or use mirror registry"
	}

	return "Check image name, registry accessibility, and authentication"
}

func (d *PodDiagnostic) checkNodeConditions(node *corev1.Node) output.CheckResult {
	var issues []string

	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				issues = append(issues, "Node is not ready")
			}
		case corev1.NodeMemoryPressure:
			if condition.Status == corev1.ConditionTrue {
				issues = append(issues, "Node under memory pressure")
			}
		case corev1.NodeDiskPressure:
			if condition.Status == corev1.ConditionTrue {
				issues = append(issues, "Node under disk pressure")
			}
		case corev1.NodePIDPressure:
			if condition.Status == corev1.ConditionTrue {
				issues = append(issues, "Node under PID pressure")
			}
		case corev1.NodeNetworkUnavailable:
			if condition.Status == corev1.ConditionTrue {
				issues = append(issues, "Node network unavailable")
			}
		}
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       fmt.Sprintf("Node %s - Conditions", node.Name),
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Node has issues: %s", strings.Join(issues, ", ")),
			Suggestion: "Check node resources and system health",
			Details: map[string]string{
				"node": node.Name,
			},
		}
	}

	return output.CheckResult{
		Name:    fmt.Sprintf("Node %s - Conditions", node.Name),
		Status:  output.StatusPassed,
		Message: "Node conditions are healthy",
		Details: map[string]string{
			"node": node.Name,
		},
	}
}

func (d *PodDiagnostic) checkResourceFit(pod *corev1.Pod, node *corev1.Node) output.CheckResult {
	// Calculate total resource requests
	totalCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalMemory := resource.NewQuantity(0, resource.BinarySI)

	for _, container := range pod.Spec.Containers {
		if cpu := container.Resources.Requests[corev1.ResourceCPU]; !cpu.IsZero() {
			totalCPU.Add(cpu)
		}
		if memory := container.Resources.Requests[corev1.ResourceMemory]; !memory.IsZero() {
			totalMemory.Add(memory)
		}
	}

	// Get node allocatable resources
	nodeCPU := node.Status.Allocatable[corev1.ResourceCPU]
	nodeMemory := node.Status.Allocatable[corev1.ResourceMemory]

	var issues []string
	if totalCPU.Cmp(nodeCPU) > 0 {
		issues = append(issues, fmt.Sprintf("CPU request (%s) exceeds node capacity (%s)", totalCPU.String(), nodeCPU.String()))
	}
	if totalMemory.Cmp(nodeMemory) > 0 {
		issues = append(issues, fmt.Sprintf("Memory request (%s) exceeds node capacity (%s)", totalMemory.String(), nodeMemory.String()))
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "Resource Fit",
			Status:     output.StatusWarning,
			Message:    "Resource requests may exceed node capacity",
			Suggestion: "Review resource requests or scale to larger nodes",
			Details: map[string]string{
				"issues": strings.Join(issues, "; "),
			},
		}
	}

	return output.CheckResult{
		Name:    "Resource Fit",
		Status:  output.StatusPassed,
		Message: "Resource requests fit within node capacity",
		Details: map[string]string{
			"cpu":    fmt.Sprintf("%s/%s", totalCPU.String(), nodeCPU.String()),
			"memory": fmt.Sprintf("%s/%s", totalMemory.String(), nodeMemory.String()),
		},
	}
}

func (d *PodDiagnostic) checkRBACEvents(info *PodInfo) []output.CheckResult {
	checks := make([]output.CheckResult, 0, 5) // Pre-allocate for expected RBAC event checks

	for _, event := range info.Events {
		if strings.Contains(event.Message, "forbidden") ||
			strings.Contains(event.Message, "unauthorized") ||
			strings.Contains(event.Reason, "FailedMount") && strings.Contains(event.Message, "secret") {

			checks = append(checks, output.CheckResult{
				Name:       "RBAC - Permission Check",
				Status:     output.StatusFailed,
				Message:    fmt.Sprintf("Permission denied: %s", event.Message),
				Suggestion: "Check RBAC bindings and service account permissions",
				Details: map[string]string{
					"event": event.Reason,
				},
			})
		}
	}

	if len(checks) == 0 {
		checks = append(checks, output.CheckResult{
			Name:    "RBAC - Permission Check",
			Status:  output.StatusPassed,
			Message: "No RBAC permission issues detected in events",
		})
	}

	return checks
}

func (d *PodDiagnostic) analyzeContainerLogs(containerName, logs string) output.CheckResult {
	lines := strings.Split(logs, "\n")

	errorPatterns := []struct {
		pattern    string
		message    string
		suggestion string
	}{
		{`(?i)(connection refused|connection denied)`, "Connection refused error detected", "Check service availability and network connectivity"},
		{`(?i)(no such host|host not found)`, "DNS resolution failure detected", "Check DNS configuration and hostname"},
		{`(?i)(permission denied|access denied)`, "Permission denied error detected", "Check file permissions and RBAC settings"},
		{`(?i)(out of memory|oom|memory limit)`, "Out of memory error detected", "Increase memory limits or optimize memory usage"},
		{`(?i)(disk.*full|no space left)`, "Disk space error detected", "Check available disk space and cleanup"},
		{`(?i)(authentication.*fail|login.*fail)`, "Authentication failure detected", "Check credentials and authentication configuration"},
		{`(?i)(timeout|timed out)`, "Timeout error detected", "Check network latency and increase timeout values"},
		{`(?i)(panic|fatal|error|exception)`, "Application error detected", "Check application logs and configuration"},
	}

	for _, pattern := range errorPatterns {
		re := regexp.MustCompile(pattern.pattern)
		for _, line := range lines {
			if re.MatchString(line) {
				return output.CheckResult{
					Name:       fmt.Sprintf("Container %s - Log Analysis", containerName),
					Status:     output.StatusFailed,
					Message:    pattern.message,
					Suggestion: pattern.suggestion,
					Details: map[string]string{
						"logLine": strings.TrimSpace(line),
					},
				}
			}
		}
	}

	return output.CheckResult{
		Name:    fmt.Sprintf("Container %s - Log Analysis", containerName),
		Status:  output.StatusPassed,
		Message: "No critical errors detected in recent logs",
		Details: map[string]string{
			"analyzed": fmt.Sprintf("%d log lines", len(lines)),
		},
	}
}

func (d *PodDiagnostic) isContainerCrashLooping(pod *corev1.Pod, containerName string) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			return status.RestartCount > 0 &&
				status.State.Waiting != nil &&
				status.State.Waiting.Reason == "CrashLoopBackOff"
		}
	}
	return false
}

func (d *PodDiagnostic) analyzeCrashLoopBackOff(containerName, logs string, pod *corev1.Pod) output.CheckResult {
	var restartCount int32
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			restartCount = status.RestartCount
			break
		}
	}

	// Get last few lines for crash analysis
	lines := strings.Split(logs, "\n")
	lastLines := make([]string, 0)
	for i := len(lines) - 1; i >= 0 && len(lastLines) < 3; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastLines = append([]string{strings.TrimSpace(lines[i])}, lastLines...)
		}
	}

	suggestion := "Check container startup configuration and resource limits"
	if len(lastLines) > 0 {
		lastLog := strings.Join(lastLines, "; ")
		if strings.Contains(strings.ToLower(lastLog), "exit") ||
			strings.Contains(strings.ToLower(lastLog), "killed") {
			suggestion = "Container is being killed - check exit codes and resource limits"
		}
	}

	return output.CheckResult{
		Name:       fmt.Sprintf("Container %s - CrashLoopBackOff", containerName),
		Status:     output.StatusFailed,
		Message:    fmt.Sprintf("Container is crash looping (restart count: %d)", restartCount),
		Suggestion: suggestion,
		Details: map[string]string{
			"recentLogs": strings.Join(lastLines, "; "),
		},
	}
}

func (d *PodDiagnostic) checkInitContainerStatus(name string, status corev1.ContainerStatus) output.CheckResult {
	if status.State.Terminated != nil {
		if status.State.Terminated.ExitCode == 0 {
			return output.CheckResult{
				Name:    fmt.Sprintf("Init Container %s", name),
				Status:  output.StatusPassed,
				Message: "Init container completed successfully",
				Details: map[string]string{
					"exitCode": fmt.Sprintf("%d", status.State.Terminated.ExitCode),
				},
			}
		}

		return output.CheckResult{
			Name:       fmt.Sprintf("Init Container %s", name),
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Init container failed with exit code %d", status.State.Terminated.ExitCode),
			Suggestion: "Check init container logs and configuration",
			Details: map[string]string{
				"reason":   status.State.Terminated.Reason,
				"exitCode": fmt.Sprintf("%d", status.State.Terminated.ExitCode),
			},
		}
	}

	if status.State.Waiting != nil {
		return output.CheckResult{
			Name:       fmt.Sprintf("Init Container %s", name),
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Init container is waiting: %s", status.State.Waiting.Reason),
			Suggestion: "Check init container image and dependencies",
			Details: map[string]string{
				"reason": status.State.Waiting.Reason,
			},
		}
	}

	if status.State.Running != nil {
		return output.CheckResult{
			Name:    fmt.Sprintf("Init Container %s", name),
			Status:  output.StatusWarning,
			Message: "Init container is still running",
			Details: map[string]string{
				"started": status.State.Running.StartedAt.Format(time.RFC3339),
			},
		}
	}

	return output.CheckResult{
		Name:    fmt.Sprintf("Init Container %s", name),
		Status:  output.StatusWarning,
		Message: "Init container status unknown",
	}
}

func (d *PodDiagnostic) checkQoSClass(pod *corev1.Pod) output.CheckResult {
	qosClass := pod.Status.QOSClass

	var message, suggestion string
	status := output.StatusPassed

	switch qosClass {
	case corev1.PodQOSGuaranteed:
		message = "Pod has Guaranteed QoS class"
		suggestion = "Excellent - pod has dedicated resources"
	case corev1.PodQOSBurstable:
		message = "Pod has Burstable QoS class"
		suggestion = "Good - pod can burst above requests but may be throttled"
	case corev1.PodQOSBestEffort:
		message = "Pod has BestEffort QoS class"
		suggestion = "Warning - pod may be evicted first during resource pressure"
		status = output.StatusWarning
	default:
		message = "Pod QoS class unknown"
		status = output.StatusWarning
	}

	return output.CheckResult{
		Name:       "Resource QoS",
		Status:     status,
		Message:    message,
		Suggestion: suggestion,
		Details: map[string]string{
			"qosClass": string(qosClass),
		},
	}
}

func (d *PodDiagnostic) checkResourceRequests(pod *corev1.Pod) output.CheckResult {
	var hasRequests, hasLimits bool
	containers := make([]string, 0, len(pod.Spec.Containers)) // Pre-allocate based on container count

	for _, container := range pod.Spec.Containers {
		if len(container.Resources.Requests) > 0 {
			hasRequests = true
		}
		if len(container.Resources.Limits) > 0 {
			hasLimits = true
		}
		containers = append(containers, container.Name)
	}

	var status output.CheckStatus
	var message, suggestion string

	switch {
	case hasRequests && hasLimits:
		status = output.StatusPassed
		message = "Resource requests and limits are configured"
		suggestion = "Good practice - helps with scheduling and prevents resource abuse"
	case hasRequests:
		status = output.StatusWarning
		message = "Resource requests configured but no limits"
		suggestion = "Consider adding resource limits to prevent containers from consuming excessive resources"
	case hasLimits:
		status = output.StatusWarning
		message = "Resource limits configured but no requests"
		suggestion = "Consider adding resource requests to help with pod scheduling"
	default:
		status = output.StatusWarning
		message = "No resource requests or limits configured"
		suggestion = "Consider adding resource requests and limits for better scheduling and resource management"
	}

	return output.CheckResult{
		Name:       "Resource Configuration",
		Status:     status,
		Message:    message,
		Suggestion: suggestion,
		Details: map[string]string{
			"containers": strings.Join(containers, ", "),
		},
	}
}

func (d *PodDiagnostic) checkResourceEvents(info *PodInfo) output.CheckResult {
	for _, event := range info.Events {
		if strings.Contains(event.Reason, "FailedScheduling") &&
			strings.Contains(event.Message, "Insufficient") {
			return output.CheckResult{
				Name:       "Resource Events",
				Status:     output.StatusFailed,
				Message:    "Insufficient resources for scheduling",
				Suggestion: "Scale cluster or reduce resource requests",
				Details: map[string]string{
					"message": event.Message,
				},
			}
		}

		if event.Reason == "OOMKilling" || strings.Contains(event.Message, "OOMKilled") {
			return output.CheckResult{
				Name:       "Resource Events",
				Status:     output.StatusFailed,
				Message:    "Container was killed due to out of memory",
				Suggestion: "Increase memory limits or optimize memory usage",
				Details: map[string]string{
					"message": event.Message,
				},
			}
		}
	}

	return output.CheckResult{
		Name:    "Resource Events",
		Status:  output.StatusPassed,
		Message: "No resource-related issues in events",
	}
}

func (d *PodDiagnostic) checkDNSConfiguration(pod *corev1.Pod) output.CheckResult {
	if pod.Spec.DNSPolicy == corev1.DNSNone && pod.Spec.DNSConfig == nil {
		return output.CheckResult{
			Name:       "DNS Configuration",
			Status:     output.StatusFailed,
			Message:    "DNS policy is None but no DNS config provided",
			Suggestion: "Configure DNS settings when using DNSPolicy: None",
			Details: map[string]string{
				"dnsPolicy": string(pod.Spec.DNSPolicy),
			},
		}
	}

	return output.CheckResult{
		Name:    "DNS Configuration",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("DNS policy configured: %s", pod.Spec.DNSPolicy),
		Details: map[string]string{
			"dnsPolicy": string(pod.Spec.DNSPolicy),
		},
	}
}

func (d *PodDiagnostic) checkNetworkEvents(info *PodInfo) output.CheckResult {
	for _, event := range info.Events {
		if strings.Contains(event.Reason, "FailedCreatePodSandBox") ||
			strings.Contains(event.Reason, "NetworkNotReady") ||
			strings.Contains(event.Message, "CNI") {
			return output.CheckResult{
				Name:       "Network Events",
				Status:     output.StatusFailed,
				Message:    "Network-related errors detected",
				Suggestion: "Check CNI plugin status and network configuration",
				Details: map[string]string{
					"event": fmt.Sprintf("%s: %s", event.Reason, event.Message),
				},
			}
		}
	}

	return output.CheckResult{
		Name:    "Network Events",
		Status:  output.StatusPassed,
		Message: "No network-related issues in events",
	}
}

// Helper functions

func getPodCondition(pod *corev1.Pod, conditionType corev1.PodConditionType) *corev1.PodCondition {
	for i := range pod.Status.Conditions {
		if pod.Status.Conditions[i].Type == conditionType {
			return &pod.Status.Conditions[i]
		}
	}
	return nil
}

func formatConditionStatus(condition *corev1.PodCondition) string {
	if condition == nil {
		return "Unknown"
	}
	return string(condition.Status)
}
