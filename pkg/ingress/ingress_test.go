package ingress

import (
	"context"
	"testing"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

func TestNewIngressDiagnostic(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)

	diag := NewIngressDiagnostic(kubeClient, outputMgr)

	if diag == nil {
		t.Fatal("NewIngressDiagnostic returned nil")
	}
	if diag.client != kubeClient {
		t.Error("IngressDiagnostic client not set correctly")
	}
	if diag.output != outputMgr {
		t.Error("IngressDiagnostic output manager not set correctly")
	}
}

func TestGetIngressClass(t *testing.T) {
	tests := []struct {
		name     string
		ingress  *networkingv1.Ingress
		expected string
	}{
		{
			name: "ingress with ingressClassName",
			ingress: &networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: ptr.To("nginx"),
				},
			},
			expected: "nginx",
		},
		{
			name: "ingress with annotation",
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "traefik",
					},
				},
			},
			expected: "traefik",
		},
		{
			name: "ingress without class specification",
			ingress: &networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{},
			},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIngressClass(tt.ingress)
			if result != tt.expected {
				t.Errorf("getIngressClass() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCheckIngressExists(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	diag := NewIngressDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "test-namespace",
	}

	// Test with nil ingress (not found)
	t.Run("ingress not found", func(t *testing.T) {
		info := &IngressInfo{
			Ingress: nil,
		}

		result := diag.checkIngressExists(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %v", result.Status)
		}
		if result.Name != "Ingress Existence" {
			t.Errorf("Expected name 'Ingress Existence', got %v", result.Name)
		}
	})

	// Test with existing ingress
	t.Run("ingress exists", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "test-namespace",
				CreationTimestamp: metav1.Time{
					Time: time.Now().Add(-1 * time.Hour),
				},
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: ptr.To("nginx"),
			},
		}

		info := &IngressInfo{
			Ingress: ingress,
		}

		result := diag.checkIngressExists(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %v", result.Status)
		}
		if result.Name != "Ingress Existence" {
			t.Errorf("Expected name 'Ingress Existence', got %v", result.Name)
		}
		if result.Details["name"] != "test-ingress" {
			t.Errorf("Expected ingress name in details, got %v", result.Details)
		}
	})
}

func TestCheckIngressConfiguration(t *testing.T) {
	kubeClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)
	diag := NewIngressDiagnostic(kubeClient, outputMgr)

	config := DiagnosticConfig{
		Namespace: "test-namespace",
	}

	// Test with ingress without rules
	t.Run("ingress without rules", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "test-namespace",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{},
			},
		}

		info := &IngressInfo{
			Ingress: ingress,
		}

		result := diag.checkIngressConfiguration(context.Background(), info, config)

		if result.Status != output.StatusFailed {
			t.Errorf("Expected status FAILED, got %v", result.Status)
		}
	})

	// Test with valid ingress configuration
	t.Run("valid ingress configuration", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "test-namespace",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: "example.com",
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:     "/",
										PathType: (*networkingv1.PathType)(ptr.To("Prefix")),
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "test-service",
												Port: networkingv1.ServiceBackendPort{
													Number: 80,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		info := &IngressInfo{
			Ingress: ingress,
		}

		result := diag.checkIngressConfiguration(context.Background(), info, config)

		if result.Status != output.StatusPassed {
			t.Errorf("Expected status PASSED, got %v", result.Status)
		}
		if result.Details["rulesCount"] != "1" {
			t.Errorf("Expected rulesCount '1', got %v", result.Details["rulesCount"])
		}
	})
}

func TestDiagnosticConfig(t *testing.T) {
	config := DiagnosticConfig{
		Namespace:      "test",
		IngressName:    "test-ingress",
		All:            true,
		AllNamespaces:  false,
		CheckSSL:       true,
		TestBackends:   true,
		TestDNS:        false,
		CheckConflicts: true,
		Controllers:    true,
		Checks:         []string{"config", "backends"},
		Timeout:        30 * time.Second,
	}

	if config.Namespace != "test" {
		t.Errorf("Expected namespace 'test', got %v", config.Namespace)
	}
	if config.IngressName != "test-ingress" {
		t.Errorf("Expected ingress name 'test-ingress', got %v", config.IngressName)
	}
	if len(config.Checks) != 2 {
		t.Errorf("Expected 2 checks, got %v", len(config.Checks))
	}
}
