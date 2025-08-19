package cluster

import (
	"testing"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

func TestNewClusterDiagnostic(t *testing.T) {
	k8sClient := &client.KubernetesClient{}
	outputMgr := output.NewOutputManager("table", false)

	diag := NewClusterDiagnostic(k8sClient, outputMgr)

	if diag == nil {
		t.Fatal("NewClusterDiagnostic() returned nil")
	}
	if diag.client != k8sClient {
		t.Error("NewClusterDiagnostic() client not set correctly")
	}
	if diag.output != outputMgr {
		t.Error("NewClusterDiagnostic() output manager not set correctly")
	}
}

func TestCalculateSummary(t *testing.T) {
	cd := &ClusterDiagnostic{}

	checks := []output.CheckResult{
		{Status: output.StatusPassed},
		{Status: output.StatusPassed},
		{Status: output.StatusFailed},
		{Status: output.StatusWarning},
		{Status: output.StatusSkipped},
		{Status: output.StatusFailed},
	}

	summary := cd.calculateSummary(checks)

	expected := output.Summary{
		Total:    6,
		Passed:   2,
		Failed:   2,
		Warnings: 1,
		Skipped:  1,
	}

	if summary != expected {
		t.Errorf("calculateSummary() = %+v, want %+v", summary, expected)
	}
}

func TestGetNodeSuggestion(t *testing.T) {
	cd := &ClusterDiagnostic{}

	tests := []struct {
		name     string
		issues   []string
		wantText string
	}{
		{
			name:     "not ready",
			issues:   []string{"NotReady"},
			wantText: "Check node status",
		},
		{
			name:     "memory pressure",
			issues:   []string{"MemoryPressure"},
			wantText: "Free up memory",
		},
		{
			name:     "disk pressure",
			issues:   []string{"DiskPressure"},
			wantText: "Clean up disk space",
		},
		{
			name:     "multiple issues",
			issues:   []string{"NotReady", "MemoryPressure"},
			wantText: "Check node status", // Should return first suggestion
		},
		{
			name:     "unknown issue",
			issues:   []string{"UnknownIssue"},
			wantText: "Check node logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := cd.getNodeSuggestion(tt.issues)
			if suggestion == "" {
				t.Error("getNodeSuggestion() returned empty string")
			}
			// Check if suggestion contains expected text
			found := false
			for _, expectedText := range []string{tt.wantText} {
				if len(suggestion) >= len(expectedText) {
					for i := 0; i <= len(suggestion)-len(expectedText); i++ {
						if suggestion[i:i+len(expectedText)] == expectedText {
							found = true
							break
						}
					}
				}
			}
			if !found {
				t.Errorf("getNodeSuggestion() = %q, want to contain %q", suggestion, tt.wantText)
			}
		})
	}
}

func TestGetControlPlaneSuggestion(t *testing.T) {
	cd := &ClusterDiagnostic{}

	tests := []struct {
		name      string
		component string
		running   int
		total     int
		wantEmpty bool
	}{
		{
			name:      "all running",
			component: "etcd",
			running:   3,
			total:     3,
			wantEmpty: true,
		},
		{
			name:      "none running",
			component: "etcd",
			running:   0,
			total:     3,
			wantEmpty: false,
		},
		{
			name:      "partial running",
			component: "kube-scheduler",
			running:   1,
			total:     2,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := cd.getControlPlaneSuggestion(tt.component, tt.running, tt.total)
			isEmpty := suggestion == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("getControlPlaneSuggestion() empty = %v, want %v. Suggestion: %q",
					isEmpty, tt.wantEmpty, suggestion)
			}
		})
	}
}

// Test with fake Kubernetes client - currently disabled due to type compatibility
// TODO: Implement interface-based client for better testability
func TestCheckNodeHealthWithFakeClient(t *testing.T) {
	t.Skip("Fake client tests disabled - TODO: implement interface-based client")
}

func TestCheckDNSWithFakeClient(t *testing.T) {
	t.Skip("Fake client tests disabled - TODO: implement interface-based client")
}

func TestCheckControlPlaneWithFakeClient(t *testing.T) {
	t.Skip("Fake client tests disabled - TODO: implement interface-based client")
}
