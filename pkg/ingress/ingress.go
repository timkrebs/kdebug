package ingress

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

// IngressDiagnostic handles ingress-related diagnostics
type IngressDiagnostic struct {
	client *client.KubernetesClient
	output *output.OutputManager
}

// DiagnosticConfig holds configuration for ingress diagnostics
type DiagnosticConfig struct {
	Namespace      string
	IngressName    string
	All            bool
	AllNamespaces  bool
	CheckSSL       bool
	TestBackends   bool
	TestDNS        bool
	CheckConflicts bool
	Controllers    bool
	Checks         []string
	Timeout        time.Duration
}

// IngressInfo contains information about an ingress resource
type IngressInfo struct {
	Ingress          *networkingv1.Ingress
	BackendServices  []*corev1.Service
	BackendEndpoints []*corev1.Endpoints
	TLSSecrets       []*corev1.Secret
	LoadBalancerIP   string
}

// SummaryInfo provides a summary of ingress diagnostics
type SummaryInfo struct {
	Total         int
	Healthy       int
	Warning       int
	Critical      int
	Controllers   int
	Certificates  int
	BackendIssues int
}

// NewIngressDiagnostic creates a new ingress diagnostic instance
func NewIngressDiagnostic(kubeClient *client.KubernetesClient, outputMgr *output.OutputManager) *IngressDiagnostic {
	return &IngressDiagnostic{
		client: kubeClient,
		output: outputMgr,
	}
}

// DiagnoseIngress performs diagnostics on a single ingress resource
func (id *IngressDiagnostic) DiagnoseIngress(ctx context.Context, ingressName string, config DiagnosticConfig) (*output.DiagnosticReport, error) {
	id.output.PrintInfo(fmt.Sprintf("🔍 Analyzing ingress: %s", ingressName))

	// Analyze the ingress resource
	ingressInfo, err := id.analyzeIngress(ctx, config.Namespace, ingressName)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze ingress %s: %w", ingressName, err)
	}

	// Run diagnostic checks
	results := id.runIngressChecks(ctx, ingressInfo, config)

	// Create diagnostic report
	report := &output.DiagnosticReport{
		Target:    fmt.Sprintf("ingress/%s", ingressName),
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    results,
		ClusterInfo: map[string]string{
			"namespace": config.Namespace,
		},
		Metadata: map[string]interface{}{
			"resourceType": "Ingress",
			"resourceName": ingressName,
			"namespace":    config.Namespace,
		},
	}

	// Calculate summary
	summary := output.Summary{Total: len(results)}
	for _, result := range results {
		switch result.Status {
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
	report.Summary = summary

	return report, nil
}

// DiagnoseAllIngresses performs diagnostics on all ingress resources in specified namespaces
func (id *IngressDiagnostic) DiagnoseAllIngresses(ctx context.Context, config DiagnosticConfig) ([]*output.DiagnosticReport, error) {
	var reports []*output.DiagnosticReport

	// Get list of ingress resources
	ingresses, err := id.getIngressResources(ctx, config.Namespace, config.AllNamespaces)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress resources: %w", err)
	}

	if len(ingresses) == 0 {
		id.output.PrintInfo("No ingress resources found in the specified namespace(s)")
		return reports, nil
	}

	id.output.PrintInfo(fmt.Sprintf("🔍 Analyzing %d ingress resources", len(ingresses)))

	// Diagnose each ingress
	for _, ingress := range ingresses {
		if ctx.Err() != nil {
			return reports, ctx.Err()
		}

		ingressConfig := config
		ingressConfig.Namespace = ingress.Namespace

		report, err := id.DiagnoseIngress(ctx, ingress.Name, ingressConfig)
		if err != nil {
			id.output.PrintError("Failed to diagnose ingress", fmt.Errorf("ingress %s: %w", ingress.Name, err))
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// analyzeIngress performs comprehensive analysis of a single ingress resource
func (id *IngressDiagnostic) analyzeIngress(ctx context.Context, namespace, name string) (*IngressInfo, error) {
	// Get the ingress resource
	ingress, err := id.client.Clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	info := &IngressInfo{
		Ingress: ingress,
	}

	// Get backend services
	info.BackendServices, err = id.getBackendServices(ctx, ingress)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend services: %w", err)
	}

	// Get backend endpoints
	info.BackendEndpoints, err = id.getBackendEndpoints(ctx, ingress)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend endpoints: %w", err)
	}

	// Get TLS secrets
	if len(ingress.Spec.TLS) > 0 {
		info.TLSSecrets, err = id.getTLSSecrets(ctx, ingress)
		if err != nil {
			return nil, fmt.Errorf("failed to get TLS secrets: %w", err)
		}
	}

	// Get load balancer IP if available
	info.LoadBalancerIP = id.getLoadBalancerIP(ingress)

	return info, nil
}

// runIngressChecks runs all diagnostic checks for an ingress resource
func (id *IngressDiagnostic) runIngressChecks(ctx context.Context, info *IngressInfo, config DiagnosticConfig) []output.CheckResult {
	var results []output.CheckResult

	// Define available checks
	allChecks := map[string]func(context.Context, *IngressInfo, DiagnosticConfig) output.CheckResult{
		"existence": id.checkIngressExists,
		"config":    id.checkIngressConfiguration,
		"backends":  id.checkBackendServices,
		"endpoints": id.checkBackendEndpoints,
		"ssl":       id.checkSSLConfiguration,
	}

	// Determine which checks to run
	checksToRun := config.Checks
	if len(checksToRun) == 0 {
		// Default checks
		checksToRun = []string{"existence", "config", "backends", "endpoints"}

		// Add SSL checks if TLS is configured
		if len(info.Ingress.Spec.TLS) > 0 {
			checksToRun = append(checksToRun, "ssl")
		}
	}

	// Run selected checks
	for _, checkName := range checksToRun {
		if checkFunc, exists := allChecks[checkName]; exists {
			result := checkFunc(ctx, info, config)
			results = append(results, result)
		}
	}

	return results
}

// checkIngressExists verifies that the ingress resource exists and is accessible
func (id *IngressDiagnostic) checkIngressExists(ctx context.Context, info *IngressInfo, config DiagnosticConfig) output.CheckResult {
	if info.Ingress == nil {
		return output.CheckResult{
			Name:       "Ingress Existence",
			Status:     output.StatusFailed,
			Message:    "Ingress resource not found",
			Suggestion: fmt.Sprintf("Create ingress or check ingress name in namespace %s", config.Namespace),
			Details: map[string]string{
				"namespace": config.Namespace,
			},
		}
	}

	return output.CheckResult{
		Name:    "Ingress Existence",
		Status:  output.StatusPassed,
		Message: "Ingress resource exists and is accessible",
		Details: map[string]string{
			"name":      info.Ingress.Name,
			"namespace": info.Ingress.Namespace,
			"class":     getIngressClass(info.Ingress),
			"created":   info.Ingress.CreationTimestamp.Time.Format(time.RFC3339),
			"age":       time.Since(info.Ingress.CreationTimestamp.Time).Round(time.Second).String(),
		},
	}
}

// checkIngressConfiguration validates the ingress resource configuration
func (id *IngressDiagnostic) checkIngressConfiguration(ctx context.Context, info *IngressInfo, config DiagnosticConfig) output.CheckResult {
	var issues []string
	var warnings []string

	ingress := info.Ingress

	// Check if ingress has rules
	if len(ingress.Spec.Rules) == 0 {
		issues = append(issues, "No ingress rules defined")
	}

	// Check for empty hosts in rules
	hasEmptyHost := false
	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			hasEmptyHost = true
		}
	}
	if hasEmptyHost {
		warnings = append(warnings, "Some rules have empty host (catch-all)")
	}

	// Check for missing default backend
	if ingress.Spec.DefaultBackend == nil && hasEmptyHost {
		warnings = append(warnings, "No default backend defined for catch-all rules")
	}

	// Check for deprecated API version
	if ingress.APIVersion == "extensions/v1beta1" || ingress.APIVersion == "networking.k8s.io/v1beta1" {
		warnings = append(warnings, "Using deprecated ingress API version, consider upgrading to networking.k8s.io/v1")
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "Ingress Configuration",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Configuration issues found: %s", strings.Join(issues, ", ")),
			Suggestion: "Review and fix the configuration issues",
			Details: map[string]string{
				"issues":     strings.Join(issues, ", "),
				"warnings":   strings.Join(warnings, ", "),
				"rulesCount": fmt.Sprintf("%d", len(ingress.Spec.Rules)),
				"apiVersion": ingress.APIVersion,
			},
		}
	}

	if len(warnings) > 0 {
		return output.CheckResult{
			Name:       "Ingress Configuration",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Configuration warnings: %s", strings.Join(warnings, ", ")),
			Suggestion: "Consider addressing these warnings",
			Details: map[string]string{
				"warnings":   strings.Join(warnings, ", "),
				"rulesCount": fmt.Sprintf("%d", len(ingress.Spec.Rules)),
				"apiVersion": ingress.APIVersion,
			},
		}
	}

	return output.CheckResult{
		Name:    "Ingress Configuration",
		Status:  output.StatusPassed,
		Message: "Ingress configuration is valid",
		Details: map[string]string{
			"rulesCount": fmt.Sprintf("%d", len(ingress.Spec.Rules)),
			"apiVersion": ingress.APIVersion,
			"class":      getIngressClass(ingress),
		},
	}
}

// checkBackendServices validates that backend services exist and are properly configured
func (id *IngressDiagnostic) checkBackendServices(ctx context.Context, info *IngressInfo, config DiagnosticConfig) output.CheckResult {
	var issues []string
	var serviceCount int

	for _, rule := range info.Ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service == nil {
				issues = append(issues, fmt.Sprintf("Path %s has no service backend", path.Path))
				continue
			}

			serviceName := path.Backend.Service.Name
			servicePort := path.Backend.Service.Port

			// Check if service exists in backend services
			found := false
			for _, svc := range info.BackendServices {
				if svc.Name == serviceName {
					found = true
					serviceCount++

					// Validate port
					portFound := false
					for _, port := range svc.Spec.Ports {
						if (servicePort.Number != 0 && port.Port == servicePort.Number) ||
							(servicePort.Name != "" && port.Name == servicePort.Name) {
							portFound = true
							break
						}
					}
					if !portFound {
						issues = append(issues, fmt.Sprintf("Service %s does not expose port %v", serviceName, servicePort))
					}
					break
				}
			}

			if !found {
				issues = append(issues, fmt.Sprintf("Backend service %s not found", serviceName))
			}
		}
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "Backend Services",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Backend service issues found: %d problems", len(issues)),
			Suggestion: "Fix backend service configurations and ensure services exist",
			Details: map[string]string{
				"issues":       strings.Join(issues, ", "),
				"serviceCount": fmt.Sprintf("%d", serviceCount),
				"problems":     fmt.Sprintf("%d", len(issues)),
			},
		}
	}

	return output.CheckResult{
		Name:    "Backend Services",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("All %d backend services are properly configured", serviceCount),
		Details: map[string]string{
			"serviceCount": fmt.Sprintf("%d", serviceCount),
			"status":       "all services validated",
		},
	}
}

// checkBackendEndpoints validates that backend services have healthy endpoints
func (id *IngressDiagnostic) checkBackendEndpoints(ctx context.Context, info *IngressInfo, config DiagnosticConfig) output.CheckResult {
	var issues []string
	var totalEndpoints, readyEndpoints int

	for _, rule := range info.Ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service == nil {
				continue
			}

			serviceName := path.Backend.Service.Name

			// Find corresponding endpoints
			for _, endpoints := range info.BackendEndpoints {
				if endpoints.Name == serviceName {
					if len(endpoints.Subsets) == 0 {
						issues = append(issues, fmt.Sprintf("Service %s has no endpoints", serviceName))
						continue
					}

					for _, subset := range endpoints.Subsets {
						totalEndpoints += len(subset.Addresses) + len(subset.NotReadyAddresses)
						readyEndpoints += len(subset.Addresses)

						if len(subset.Addresses) == 0 {
							issues = append(issues, fmt.Sprintf("Service %s has no ready endpoints", serviceName))
						}
					}
					break
				}
			}
		}
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "Backend Endpoints",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("Endpoint issues found: %d problems", len(issues)),
			Suggestion: "Check pod health and service selectors",
			Details: map[string]string{
				"issues":         strings.Join(issues, ", "),
				"totalEndpoints": fmt.Sprintf("%d", totalEndpoints),
				"readyEndpoints": fmt.Sprintf("%d", readyEndpoints),
				"problems":       fmt.Sprintf("%d", len(issues)),
			},
		}
	}

	if readyEndpoints == 0 {
		return output.CheckResult{
			Name:       "Backend Endpoints",
			Status:     output.StatusFailed,
			Message:    "No ready endpoints found for any backend services",
			Suggestion: "Check pod status and ensure pods are running and ready",
			Details: map[string]string{
				"readyEndpoints": fmt.Sprintf("%d", readyEndpoints),
				"totalEndpoints": fmt.Sprintf("%d", totalEndpoints),
				"status":         "no ready endpoints",
			},
		}
	}

	if readyEndpoints < totalEndpoints {
		return output.CheckResult{
			Name:       "Backend Endpoints",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("Some endpoints not ready: %d/%d ready", readyEndpoints, totalEndpoints),
			Suggestion: "Check why some endpoints are not ready",
			Details: map[string]string{
				"readyEndpoints": fmt.Sprintf("%d", readyEndpoints),
				"totalEndpoints": fmt.Sprintf("%d", totalEndpoints),
				"readyRatio":     fmt.Sprintf("%.1f%%", float64(readyEndpoints)/float64(totalEndpoints)*100),
			},
		}
	}

	return output.CheckResult{
		Name:    "Backend Endpoints",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("All %d endpoints are ready", readyEndpoints),
		Details: map[string]string{
			"readyEndpoints": fmt.Sprintf("%d", readyEndpoints),
			"totalEndpoints": fmt.Sprintf("%d", totalEndpoints),
			"status":         "all endpoints healthy",
		},
	}
}

// checkSSLConfiguration validates SSL/TLS configuration
func (id *IngressDiagnostic) checkSSLConfiguration(ctx context.Context, info *IngressInfo, config DiagnosticConfig) output.CheckResult {
	if len(info.Ingress.Spec.TLS) == 0 {
		return output.CheckResult{
			Name:    "SSL Configuration",
			Status:  output.StatusSkipped,
			Message: "No TLS configuration found",
			Details: map[string]string{
				"tlsCount": "0",
				"status":   "no TLS configuration",
			},
		}
	}

	var issues []string
	var warnings []string

	for _, tls := range info.Ingress.Spec.TLS {
		if tls.SecretName == "" {
			issues = append(issues, "TLS block has no secret name specified")
			continue
		}

		// Check if secret exists
		secretFound := false
		for _, secret := range info.TLSSecrets {
			if secret.Name == tls.SecretName {
				secretFound = true

				// Validate secret type
				if secret.Type != corev1.SecretTypeTLS {
					warnings = append(warnings, fmt.Sprintf("Secret %s is not of type kubernetes.io/tls", tls.SecretName))
				}

				// Check for required keys
				if _, hasCert := secret.Data["tls.crt"]; !hasCert {
					issues = append(issues, fmt.Sprintf("Secret %s missing tls.crt", tls.SecretName))
				}
				if _, hasKey := secret.Data["tls.key"]; !hasKey {
					issues = append(issues, fmt.Sprintf("Secret %s missing tls.key", tls.SecretName))
				}
				break
			}
		}

		if !secretFound {
			issues = append(issues, fmt.Sprintf("TLS secret %s not found", tls.SecretName))
		}

		// Validate hosts
		if len(tls.Hosts) == 0 {
			warnings = append(warnings, fmt.Sprintf("TLS block for secret %s has no hosts specified", tls.SecretName))
		}
	}

	if len(issues) > 0 {
		return output.CheckResult{
			Name:       "SSL Configuration",
			Status:     output.StatusFailed,
			Message:    fmt.Sprintf("SSL configuration issues: %s", strings.Join(issues, ", ")),
			Suggestion: "Fix SSL/TLS configuration issues",
			Details: map[string]string{
				"issues":   strings.Join(issues, ", "),
				"warnings": strings.Join(warnings, ", "),
				"tlsCount": fmt.Sprintf("%d", len(info.Ingress.Spec.TLS)),
			},
		}
	}

	if len(warnings) > 0 {
		return output.CheckResult{
			Name:       "SSL Configuration",
			Status:     output.StatusWarning,
			Message:    fmt.Sprintf("SSL configuration warnings: %s", strings.Join(warnings, ", ")),
			Suggestion: "Consider addressing SSL/TLS warnings",
			Details: map[string]string{
				"warnings": strings.Join(warnings, ", "),
				"tlsCount": fmt.Sprintf("%d", len(info.Ingress.Spec.TLS)),
			},
		}
	}

	return output.CheckResult{
		Name:    "SSL Configuration",
		Status:  output.StatusPassed,
		Message: fmt.Sprintf("SSL configuration is valid for %d TLS blocks", len(info.Ingress.Spec.TLS)),
		Details: map[string]string{
			"tlsCount": fmt.Sprintf("%d", len(info.Ingress.Spec.TLS)),
			"status":   "all TLS configurations valid",
		},
	}
}

// Helper methods

// getIngressResources retrieves ingress resources from specified namespace(s)
func (id *IngressDiagnostic) getIngressResources(ctx context.Context, namespace string, allNamespaces bool) ([]*networkingv1.Ingress, error) {
	var ingresses []*networkingv1.Ingress

	if allNamespaces {
		ingressList, err := id.client.Clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for i := range ingressList.Items {
			ingresses = append(ingresses, &ingressList.Items[i])
		}
	} else {
		ingressList, err := id.client.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for i := range ingressList.Items {
			ingresses = append(ingresses, &ingressList.Items[i])
		}
	}

	return ingresses, nil
}

// getBackendServices retrieves all backend services referenced by the ingress
func (id *IngressDiagnostic) getBackendServices(ctx context.Context, ingress *networkingv1.Ingress) ([]*corev1.Service, error) {
	var services []*corev1.Service
	serviceNames := make(map[string]bool)

	// Extract service names from ingress rules
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil {
				serviceNames[path.Backend.Service.Name] = true
			}
		}
	}

	// Get each service
	for serviceName := range serviceNames {
		service, err := id.client.Clientset.CoreV1().Services(ingress.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			// Service not found - we'll report this in the check
			continue
		}
		services = append(services, service)
	}

	return services, nil
}

// getBackendEndpoints retrieves endpoints for backend services
func (id *IngressDiagnostic) getBackendEndpoints(ctx context.Context, ingress *networkingv1.Ingress) ([]*corev1.Endpoints, error) {
	var endpoints []*corev1.Endpoints
	serviceNames := make(map[string]bool)

	// Extract service names from ingress rules
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil {
				serviceNames[path.Backend.Service.Name] = true
			}
		}
	}

	// Get endpoints for each service
	for serviceName := range serviceNames {
		endpoint, err := id.client.Clientset.CoreV1().Endpoints(ingress.Namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			// Endpoints not found - we'll report this in the check
			continue
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// getTLSSecrets retrieves TLS secrets referenced by the ingress
func (id *IngressDiagnostic) getTLSSecrets(ctx context.Context, ingress *networkingv1.Ingress) ([]*corev1.Secret, error) {
	var secrets []*corev1.Secret
	secretNames := make(map[string]bool)

	// Extract secret names from TLS configuration
	for _, tls := range ingress.Spec.TLS {
		if tls.SecretName != "" {
			secretNames[tls.SecretName] = true
		}
	}

	// Get each secret
	for secretName := range secretNames {
		secret, err := id.client.Clientset.CoreV1().Secrets(ingress.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			// Secret not found - we'll report this in the check
			continue
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// getLoadBalancerIP extracts the load balancer IP from ingress status
func (id *IngressDiagnostic) getLoadBalancerIP(ingress *networkingv1.Ingress) string {
	if len(ingress.Status.LoadBalancer.Ingress) > 0 {
		if ingress.Status.LoadBalancer.Ingress[0].IP != "" {
			return ingress.Status.LoadBalancer.Ingress[0].IP
		}
		if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
			return ingress.Status.LoadBalancer.Ingress[0].Hostname
		}
	}
	return ""
}

// getIngressClass returns the ingress class name or "default" if not specified
func getIngressClass(ingress *networkingv1.Ingress) string {
	if ingress.Spec.IngressClassName != nil {
		return *ingress.Spec.IngressClassName
	}
	if class, exists := ingress.Annotations["kubernetes.io/ingress.class"]; exists {
		return class
	}
	return "default"
}
