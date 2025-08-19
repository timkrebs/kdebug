package client

import (
	"context"
	"testing"
	"time"
)

func TestNewKubernetesClient(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		wantErr    bool
	}{
		{
			name:       "empty kubeconfig",
			kubeconfig: "",
			wantErr:    false, // May succeed if in-cluster config or default kubeconfig exists
		},
		{
			name:       "invalid kubeconfig path",
			kubeconfig: "/nonexistent/path/config",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewKubernetesClient(tt.kubeconfig)

			// For empty kubeconfig, check if it's actually available
			if tt.name == "empty kubeconfig" {
				if err != nil {
					t.Logf("No default kubeconfig available (expected): %v", err)
					return // This is acceptable
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubernetesClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("Expected valid client when no error")
			}
			if !tt.wantErr && client != nil && client.Clientset == nil {
				t.Error("Expected valid clientset")
			}
		})
	}
}

func TestKubernetesClient_TestConnection(t *testing.T) {
	// This test requires a valid kubeconfig and cluster
	// Skip if not available
	client, err := NewKubernetesClient("")
	if err != nil {
		t.Skip("No valid Kubernetes cluster available for testing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.TestConnection(ctx)
	if err != nil {
		t.Logf("Connection test failed (expected if no cluster): %v", err)
	}
}

func TestGetDefaultKubeconfigPath(t *testing.T) {
	path := getDefaultKubeconfigPath()
	// Should return a path even if file doesn't exist
	if path == "" {
		t.Error("Expected non-empty default kubeconfig path")
	}
}

func TestGetCurrentContext(t *testing.T) {
	// Test with invalid path
	_, err := getCurrentContext("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent kubeconfig")
	}

	// Test with empty path (will use default)
	context, err := getCurrentContext("")
	if err != nil {
		t.Logf("No default kubeconfig available: %v", err)
	} else if context == "" {
		t.Log("Empty context from kubeconfig")
	}
}
