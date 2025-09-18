// Package service provides diagnostic capabilities for Kubernetes service-level issues.
//
// This package implements comprehensive checks for common service problems including:
//   - Service configuration: selector mismatches, port configuration, service types
//   - Endpoint health: backend pod availability, readiness, health status
//   - DNS resolution: service discovery, name resolution within cluster
//   - Connectivity issues: network policies, service mesh configuration
//   - Load balancing: endpoint distribution, session affinity
//
// The diagnostics help users identify root causes and provide actionable recommendations
// for resolving service-related networking issues in Kubernetes clusters.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

// ServiceDiagnostic performs diagnostic checks for service-level issues.
type ServiceDiagnostic struct {
	client *client.KubernetesClient
	output *output.OutputManager
}

// DiagnosticConfig contains configuration options for service diagnostics.
type DiagnosticConfig struct {
	Namespace     string
	Checks        []string
	TestDNS       bool
	AllNamespaces bool
	Timeout       time.Duration
	Verbose       bool
}

// ServiceInfo contains comprehensive information about a service and its health.
type ServiceInfo struct {
	Service     *corev1.Service
	Endpoints   *corev1.Endpoints
	BackendPods []*corev1.Pod
	Events      []corev1.Event
}

// NewServiceDiagnostic creates a new service diagnostic instance.
func NewServiceDiagnostic(kubeClient *client.KubernetesClient, outputMgr *output.OutputManager) *ServiceDiagnostic {
	return &ServiceDiagnostic{
		client: kubeClient,
		output: outputMgr,
	}
}

// DiagnoseService performs comprehensive diagnostics on a specific service.
func (sd *ServiceDiagnostic) DiagnoseService(ctx context.Context, serviceName string, config DiagnosticConfig) (*output.DiagnosticReport, error) {
	sd.output.PrintInfo(fmt.Sprintf("🔍 Analyzing service: %s", serviceName))

	// Get service information
	serviceInfo, err := sd.getServiceInfo(ctx, serviceName, config.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get service information: %w", err)
	}

	// Create diagnostic report
	report := &output.DiagnosticReport{
		Target:    fmt.Sprintf("Service %s/%s", config.Namespace, serviceName),
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    []output.CheckResult{},
		Metadata: map[string]interface{}{
			"resourceType": "Service",
			"resourceName": serviceName,
			"namespace":    config.Namespace,
		},
	}

	// Run diagnostic checks
	checks := []func(context.Context, *ServiceInfo, DiagnosticConfig) output.CheckResult{
		sd.checkServiceExists,
		sd.checkServiceConfiguration,
		sd.checkServiceSelector,
		sd.checkEndpointHealth,
		sd.checkPortConfiguration,
	}

	// Run checks
	for _, checkFunc := range checks {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		check := checkFunc(ctx, serviceInfo, config)
		report.Checks = append(report.Checks, check)
	}

	// Calculate summary
	report.Summary = sd.calculateSummary(report.Checks)

	return report, nil
}

// DiagnoseAllServices performs diagnostics on all services in the specified namespace(s).
func (sd *ServiceDiagnostic) DiagnoseAllServices(ctx context.Context, config DiagnosticConfig) ([]*output.DiagnosticReport, error) {
	var reports []*output.DiagnosticReport

	// Get list of services
	services, err := sd.getServiceList(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	if len(services.Items) == 0 {
		sd.output.PrintInfo("No services found in the specified namespace(s)")
		return reports, nil
	}

	sd.output.PrintInfo(fmt.Sprintf("🔍 Analyzing %d services", len(services.Items)))

	// Diagnose each service
	for _, service := range services.Items {
		if ctx.Err() != nil {
			return reports, ctx.Err()
		}

		serviceConfig := config
		serviceConfig.Namespace = service.Namespace

		report, err := sd.DiagnoseService(ctx, service.Name, serviceConfig)
		if err != nil {
			sd.output.PrintError("Failed to diagnose service", fmt.Errorf("service %s: %w", service.Name, err))
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// getServiceInfo retrieves comprehensive information about a service.
func (sd *ServiceDiagnostic) getServiceInfo(ctx context.Context, serviceName, namespace string) (*ServiceInfo, error) {
	info := &ServiceInfo{}

	// Get service
	service, err := sd.client.Clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	info.Service = service

	// Get endpoints
	endpoints, err := sd.client.Clientset.CoreV1().Endpoints(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		// Endpoints might not exist yet, which is not necessarily an error
		if !strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("failed to get endpoints: %w", err)
		}
	} else {
		info.Endpoints = endpoints
	}

	// Get backend pods if service has selectors
	if len(service.Spec.Selector) > 0 {
		selector := labels.SelectorFromSet(service.Spec.Selector)
		pods, err := sd.client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err == nil {
			for i := range pods.Items {
				info.BackendPods = append(info.BackendPods, &pods.Items[i])
			}
		}
	}

	// Get recent events
	events, err := sd.client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Service", serviceName),
	})
	if err == nil {
		info.Events = events.Items
	}

	return info, nil
}

// getServiceList retrieves a list of services based on the configuration.
func (sd *ServiceDiagnostic) getServiceList(ctx context.Context, config DiagnosticConfig) (*corev1.ServiceList, error) {
	if config.AllNamespaces {
		return sd.client.Clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	}
	return sd.client.Clientset.CoreV1().Services(config.Namespace).List(ctx, metav1.ListOptions{})
}

// checkServiceExists verifies that the service exists and is accessible.
func (sd *ServiceDiagnostic) checkServiceExists(ctx context.Context, info *ServiceInfo, config DiagnosticConfig) output.CheckResult {
	if info.Service == nil {
		return output.CheckResult{
			Name:       "Service Existence",
			Status:     output.StatusFailed,
			Message:    "Service not found",
			Suggestion: fmt.Sprintf("Create service or check service name in namespace %s", config.Namespace),
			Details: map[string]string{
				"namespace": config.Namespace,
			},
		}
	}

	return output.CheckResult{
		Name:    "Service Existence",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("Service '%s' exists and is accessible", info.Service.Name),
		Details: map[string]string{
			"name":      info.Service.Name,
			"namespace": info.Service.Namespace,
			"type":      string(info.Service.Spec.Type),
			"created":   info.Service.CreationTimestamp.Time.Format(time.RFC3339),
		},
	}
}

// checkServiceConfiguration validates the service configuration.
func (sd *ServiceDiagnostic) checkServiceConfiguration(ctx context.Context, info *ServiceInfo, config DiagnosticConfig) output.CheckResult {
	if info.Service == nil {
		return output.CheckResult{
			Name:    "Service Configuration",
			Status:  output.StatusSkipped,
			Message: "Service not found, skipping configuration check",
		}
	}

	service := info.Service
	issues := []string{}
	suggestions := []string{}

	// Check if service has ports configured
	if len(service.Spec.Ports) == 0 {
		issues = append(issues, "No ports configured")
		suggestions = append(suggestions, "Add port configuration to service spec")
	}

	// Check for duplicate port names
	portNames := make(map[string]bool)
	for _, port := range service.Spec.Ports {
		if port.Name != "" {
			if portNames[port.Name] {
				issues = append(issues, fmt.Sprintf("Duplicate port name: %s", port.Name))
				suggestions = append(suggestions, "Ensure all port names are unique")
			}
			portNames[port.Name] = true
		}
	}

	// Check service type specific configuration
	switch service.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) == 0 {
			issues = append(issues, "LoadBalancer service has no external IP assigned")
			suggestions = append(suggestions, "Check if load balancer controller is running and configured")
		}
	case corev1.ServiceTypeNodePort:
		for _, port := range service.Spec.Ports {
			if port.NodePort < 30000 || port.NodePort > 32767 {
				issues = append(issues, fmt.Sprintf("NodePort %d is outside valid range (30000-32767)", port.NodePort))
				suggestions = append(suggestions, "Use NodePort in valid range or let Kubernetes assign one")
			}
		}
	case corev1.ServiceTypeExternalName:
		if service.Spec.ExternalName == "" {
			issues = append(issues, "ExternalName service has no external name specified")
			suggestions = append(suggestions, "Set spec.externalName to the target DNS name")
		}
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "Service Configuration",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Configuration issues found: %s", strings.Join(issues, ", ")),
			Suggestion: strings.Join(suggestions, "; "),
			Details: map[string]string{
				"issues":      strings.Join(issues, ", "),
				"serviceType": string(service.Spec.Type),
				"portCount":   fmt.Sprintf("%d", len(service.Spec.Ports)),
			},
		}
	}

	return output.CheckResult{
		Name:    "Service Configuration",
		Status:  output.StatusPassed,
		Message: "Service configuration is valid",
		Details: map[string]string{
			"serviceType": string(service.Spec.Type),
			"portCount":   fmt.Sprintf("%d", len(service.Spec.Ports)),
		},
	}
}

// checkServiceSelector validates the service selector and pod matching.
func (sd *ServiceDiagnostic) checkServiceSelector(ctx context.Context, info *ServiceInfo, config DiagnosticConfig) output.CheckResult {
	if info.Service == nil {
		return output.CheckResult{
			Name:    "Service Selector",
			Status:  output.StatusSkipped,
			Message: "Service not found, skipping selector check",
		}
	}

	service := info.Service

	// ExternalName services don't have selectors
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return output.CheckResult{
			Name:    "Service Selector",
			Status:  output.StatusSkipped,
			Message: "ExternalName service does not use selectors",
		}
	}

	// Check if service has selectors
	if len(service.Spec.Selector) == 0 {
		return output.CheckResult{
			Name:       "Service Selector",
			Status:     output.StatusWarning,
			Message:    "Service has no selector (headless service or manual endpoints)",
			Suggestion: "If this should select pods, add appropriate selector labels",
			Details: map[string]string{
				"hasSelector": "false",
			},
		}
	}

	// Count matching pods
	matchingPods := len(info.BackendPods)
	readyPods := 0

	for _, pod := range info.BackendPods {
		if isPodReady(pod) {
			readyPods++
		}
	}

	if matchingPods == 0 {
		return output.CheckResult{
			Name:       "Service Selector",
			Status:     output.StatusFailed,
			Message:    "Service selector matches no pods",
			Suggestion: "Check if pods exist with matching labels, or update service selector",
			Details: map[string]string{
				"matchingPods": "0",
				"selector":     fmt.Sprintf("%v", service.Spec.Selector),
			},
		}
	}

	if readyPods == 0 {
		return output.CheckResult{
			Name:       "Service Selector",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Service selector matches %d pods but none are ready", matchingPods),
			Suggestion: "Check pod readiness and health",
			Details: map[string]string{
				"matchingPods": fmt.Sprintf("%d", matchingPods),
				"readyPods":    "0",
				"selector":     fmt.Sprintf("%v", service.Spec.Selector),
			},
		}
	}

	return output.CheckResult{
		Name:    "Service Selector",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("Service selector matches %d pods (%d ready)", matchingPods, readyPods),
		Details: map[string]string{
			"matchingPods": fmt.Sprintf("%d", matchingPods),
			"readyPods":    fmt.Sprintf("%d", readyPods),
			"selector":     fmt.Sprintf("%v", service.Spec.Selector),
		},
	}
}

// checkEndpointHealth validates the health of service endpoints.
func (sd *ServiceDiagnostic) checkEndpointHealth(ctx context.Context, info *ServiceInfo, config DiagnosticConfig) output.CheckResult {
	if info.Service == nil {
		return output.CheckResult{
			Name:    "Endpoint Health",
			Status:  output.StatusSkipped,
			Message: "Service not found, skipping endpoint check",
		}
	}

	// ExternalName services don't have endpoints
	if info.Service.Spec.Type == corev1.ServiceTypeExternalName {
		return output.CheckResult{
			Name:    "Endpoint Health",
			Status:  output.StatusSkipped,
			Message: "ExternalName service does not have endpoints",
		}
	}

	// Check if endpoints exist
	if info.Endpoints == nil {
		return output.CheckResult{
			Name:       "Endpoint Health",
			Status:     output.StatusFailed,
			Message:    "No endpoints found for service",
			Suggestion: "Check if pods are running and ready, and if service selector is correct",
			Details: map[string]string{
				"hasEndpoints": "false",
			},
		}
	}

	// Count ready and not ready endpoints
	readyCount := 0
	notReadyCount := 0

	for _, subset := range info.Endpoints.Subsets {
		readyCount += len(subset.Addresses)
		notReadyCount += len(subset.NotReadyAddresses)
	}

	totalEndpoints := readyCount + notReadyCount

	if totalEndpoints == 0 {
		return output.CheckResult{
			Name:       "Endpoint Health",
			Status:     output.StatusFailed,
			Message:    "Service has no endpoints",
			Suggestion: "Check if pods are running and if service selector matches pod labels",
			Details: map[string]string{
				"readyEndpoints":    "0",
				"notReadyEndpoints": "0",
			},
		}
	}

	if readyCount == 0 {
		return output.CheckResult{
			Name:       "Endpoint Health",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("All %d endpoints are not ready", notReadyCount),
			Suggestion: "Check pod readiness probes and application health",
			Details: map[string]string{
				"readyEndpoints":    "0",
				"notReadyEndpoints": fmt.Sprintf("%d", notReadyCount),
			},
		}
	}

	if notReadyCount > 0 {
		return output.CheckResult{
			Name:       "Endpoint Health",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("%d/%d endpoints are ready", readyCount, totalEndpoints),
			Suggestion: "Some endpoints are not ready, check pod health",
			Details: map[string]string{
				"readyEndpoints":    fmt.Sprintf("%d", readyCount),
				"notReadyEndpoints": fmt.Sprintf("%d", notReadyCount),
			},
		}
	}

	return output.CheckResult{
		Name:    "Endpoint Health",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("All %d endpoints are ready", readyCount),
		Details: map[string]string{
			"readyEndpoints":    fmt.Sprintf("%d", readyCount),
			"notReadyEndpoints": "0",
		},
	}
}

// checkPortConfiguration validates service port configuration.
func (sd *ServiceDiagnostic) checkPortConfiguration(ctx context.Context, info *ServiceInfo, config DiagnosticConfig) output.CheckResult {
	if info.Service == nil {
		return output.CheckResult{
			Name:    "Port Configuration",
			Status:  output.StatusSkipped,
			Message: "Service not found, skipping port check",
		}
	}

	service := info.Service
	issues := []string{}
	suggestions := []string{}

	// Check each port
	for _, port := range service.Spec.Ports {
		// Check for valid port numbers
		if port.Port < 1 || port.Port > 65535 {
			issues = append(issues, fmt.Sprintf("Invalid service port: %d", port.Port))
			suggestions = append(suggestions, "Use valid port numbers (1-65535)")
		}

		if port.TargetPort.IntVal < 1 || port.TargetPort.IntVal > 65535 {
			if port.TargetPort.StrVal == "" {
				issues = append(issues, fmt.Sprintf("Invalid target port: %d", port.TargetPort.IntVal))
				suggestions = append(suggestions, "Use valid target port numbers (1-65535)")
			}
		}

		// Check protocol
		if port.Protocol != corev1.ProtocolTCP && port.Protocol != corev1.ProtocolUDP && port.Protocol != corev1.ProtocolSCTP {
			issues = append(issues, fmt.Sprintf("Invalid protocol: %s", port.Protocol))
			suggestions = append(suggestions, "Use valid protocols (TCP, UDP, SCTP)")
		}
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "Port Configuration",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Port configuration issues: %s", strings.Join(issues, ", ")),
			Suggestion: strings.Join(suggestions, "; "),
			Details: map[string]string{
				"issues":    strings.Join(issues, ", "),
				"portCount": fmt.Sprintf("%d", len(service.Spec.Ports)),
			},
		}
	}

	return output.CheckResult{
		Name:    "Port Configuration",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("All %d port configurations are valid", len(service.Spec.Ports)),
		Details: map[string]string{
			"portCount": fmt.Sprintf("%d", len(service.Spec.Ports)),
		},
	}
}

// calculateSummary generates a summary of the diagnostic results.
func (sd *ServiceDiagnostic) calculateSummary(checks []output.CheckResult) output.Summary {
	passed := 0
	failed := 0
	warnings := 0
	skipped := 0

	for _, check := range checks {
		switch check.Status {
		case output.StatusPassed:
			passed++
		case output.StatusFailed:
			failed++
		case output.StatusWarning:
			warnings++
		case output.StatusSkipped:
			skipped++
		}
	}

	return output.Summary{
		Total:    len(checks),
		Passed:   passed,
		Failed:   failed,
		Warnings: warnings,
		Skipped:  skipped,
	}
}

// isPodReady checks if a pod is ready.
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
