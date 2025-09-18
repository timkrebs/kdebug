package service

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

func TestNewServiceDiagnostic(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)

	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	if serviceDiag == nil {
		t.Fatal("Expected ServiceDiagnostic to be created, got nil")
	}

	if serviceDiag.client != kubeClient {
		t.Error("Expected client to be set correctly")
	}

	if serviceDiag.output != outputMgr {
		t.Error("Expected output manager to be set correctly")
	}
}

func TestCheckServiceExists(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "default",
	}

	t.Run("service exists", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkServiceExists(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %s", result.Status)
		}

		if result.Name != "Service Existence" {
			t.Errorf("Expected name 'Service Existence', got %s", result.Name)
		}
	})

	t.Run("service not found", func(t *testing.T) {
		info := &ServiceInfo{
			Service: nil,
		}

		result := serviceDiag.checkServiceExists(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %s", result.Status)
		}

		if result.Message != "Service not found" {
			t.Errorf("Expected message 'Service not found', got %s", result.Message)
		}
	})
}

func TestCheckServiceConfiguration(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "default",
	}

	t.Run("valid configuration", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(8080),
						Protocol:   corev1.ProtocolTCP,
					},
				},
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkServiceConfiguration(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %s", result.Status)
		}
	})

	t.Run("no ports configured", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{},
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkServiceConfiguration(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "No ports configured") {
			t.Errorf("Expected message to contain 'No ports configured', got %s", result.Message)
		}
	})

	t.Run("external name without external name", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: "", // Missing external name
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkServiceConfiguration(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "ExternalName service has no external name specified") {
			t.Errorf("Expected message to contain external name error, got %s", result.Message)
		}
	})
}

func TestCheckServiceSelector(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "default",
	}

	t.Run("service with matching ready pods", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{
					"app": "test",
				},
			},
		}

		readyPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}

		info := &ServiceInfo{
			Service:     service,
			BackendPods: []*corev1.Pod{readyPod},
		}

		result := serviceDiag.checkServiceSelector(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "1 pods (1 ready)") {
			t.Errorf("Expected message to show 1 ready pod, got %s", result.Message)
		}
	})

	t.Run("service with no selector", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{}, // No selector
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkServiceSelector(context.Background(), info, config)

		if result.Status != output.StatusWarning {
			t.Errorf("Expected status WARNING, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "no selector") {
			t.Errorf("Expected message to mention no selector, got %s", result.Message)
		}
	})

	t.Run("external name service", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: "example.com",
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkServiceSelector(context.Background(), info, config)

		if result.Status != output.StatusSkipped {
			t.Errorf("Expected status SKIPPED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "does not use selectors") {
			t.Errorf("Expected message to mention no selectors for ExternalName, got %s", result.Message)
		}
	})
}

func TestCheckEndpointHealth(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "default",
	}

	t.Run("healthy endpoints", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		}

		endpoints := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{
						{IP: "10.0.0.1"},
						{IP: "10.0.0.2"},
					},
					NotReadyAddresses: []corev1.EndpointAddress{},
				},
			},
		}

		info := &ServiceInfo{
			Service:   service,
			Endpoints: endpoints,
		}

		result := serviceDiag.checkEndpointHealth(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "All 2 endpoints are ready") {
			t.Errorf("Expected message to show 2 ready endpoints, got %s", result.Message)
		}
	})

	t.Run("no endpoints", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		}

		info := &ServiceInfo{
			Service:   service,
			Endpoints: nil, // No endpoints
		}

		result := serviceDiag.checkEndpointHealth(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "No endpoints found") {
			t.Errorf("Expected message to mention no endpoints, got %s", result.Message)
		}
	})
}

func TestCheckPortConfiguration(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "default",
	}

	t.Run("valid port configuration", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(8080),
						Protocol:   corev1.ProtocolTCP,
					},
					{
						Name:       "https",
						Port:       443,
						TargetPort: intstr.FromInt(8443),
						Protocol:   corev1.ProtocolTCP,
					},
				},
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkPortConfiguration(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "All 2 port configurations are valid") {
			t.Errorf("Expected message to show 2 valid ports, got %s", result.Message)
		}
	})

	t.Run("invalid port numbers", func(t *testing.T) {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:       "invalid",
						Port:       0, // Invalid port
						TargetPort: intstr.FromInt(8080),
						Protocol:   corev1.ProtocolTCP,
					},
				},
			},
		}

		info := &ServiceInfo{
			Service: service,
		}

		result := serviceDiag.checkPortConfiguration(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %s", result.Status)
		}

		if !strings.Contains(result.Message, "Invalid service port") {
			t.Errorf("Expected message to mention invalid port, got %s", result.Message)
		}
	})
}

func TestCalculateSummary(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	serviceDiag := NewServiceDiagnostic(kubeClient, outputMgr)

	t.Run("all passed", func(t *testing.T) {
		checks := []output.CheckResult{
			{Status: output.StatusPassed},
			{Status: output.StatusPassed},
		}

		summary := serviceDiag.calculateSummary(checks)

		if summary.Total != 2 {
			t.Errorf("Expected total 2, got %d", summary.Total)
		}
		if summary.Passed != 2 {
			t.Errorf("Expected passed 2, got %d", summary.Passed)
		}
		if summary.Failed != 0 {
			t.Errorf("Expected failed 0, got %d", summary.Failed)
		}
	})

	t.Run("mixed results", func(t *testing.T) {
		checks := []output.CheckResult{
			{Status: output.StatusPassed},
			{Status: output.StatusFailed},
			{Status: output.StatusWarning},
			{Status: output.StatusSkipped},
		}

		summary := serviceDiag.calculateSummary(checks)

		if summary.Total != 4 {
			t.Errorf("Expected total 4, got %d", summary.Total)
		}
		if summary.Passed != 1 {
			t.Errorf("Expected passed 1, got %d", summary.Passed)
		}
		if summary.Failed != 1 {
			t.Errorf("Expected failed 1, got %d", summary.Failed)
		}
		if summary.Warnings != 1 {
			t.Errorf("Expected warnings 1, got %d", summary.Warnings)
		}
		if summary.Skipped != 1 {
			t.Errorf("Expected skipped 1, got %d", summary.Skipped)
		}
	})
}

func TestIsPodReady(t *testing.T) {
	t.Run("ready pod", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}

		if !isPodReady(pod) {
			t.Error("Expected pod to be ready")
		}
	})

	t.Run("not ready pod", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		}

		if isPodReady(pod) {
			t.Error("Expected pod to not be ready")
		}
	})

	t.Run("pod without ready condition", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{},
			},
		}

		if isPodReady(pod) {
			t.Error("Expected pod without ready condition to not be ready")
		}
	})
}
