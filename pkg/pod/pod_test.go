package pod

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kdebug/internal/client"
	"kdebug/internal/output"
)

func TestNewPodDiagnostic(t *testing.T) {
	client := &client.KubernetesClient{}
	outputManager := output.NewOutputManager("table", false)

	diagnostic := NewPodDiagnostic(client, outputManager)

	if diagnostic.client != client {
		t.Error("Client not set correctly")
	}
	if diagnostic.output != outputManager {
		t.Error("Output manager not set correctly")
	}
}

func TestCalculateSummary(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name     string
		checks   []output.CheckResult
		expected output.Summary
	}{
		{
			name: "all passed",
			checks: []output.CheckResult{
				{Status: "PASSED"},
				{Status: "PASSED"},
			},
			expected: output.Summary{Total: 2, Passed: 2},
		},
		{
			name: "mixed results",
			checks: []output.CheckResult{
				{Status: "PASSED"},
				{Status: "FAILED"},
				{Status: "WARNING"},
				{Status: "SKIPPED"},
			},
			expected: output.Summary{Total: 4, Passed: 1, Failed: 1, Warnings: 1, Skipped: 1},
		},
		{
			name:     "empty checks",
			checks:   []output.CheckResult{},
			expected: output.Summary{Total: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diagnostic.calculateSummary(tt.checks)
			if result != tt.expected {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestIsPodFailing(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
	}{
		{
			name: "running pod with ready containers",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
							RestartCount: 0,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "pending pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			expected: true,
		},
		{
			name: "failed pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			},
			expected: true,
		},
		{
			name: "crash loop back off",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
							RestartCount: 3,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "image pull back off",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ImagePullBackOff",
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "container with restart count",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
							RestartCount: 1,
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diagnostic.isPodFailing(tt.pod)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCheckPodBasicStatus(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name           string
		pod            *corev1.Pod
		expectedStatus string
		expectedName   string
	}{
		{
			name: "running pod with ready containers",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: true},
					},
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedStatus: "PASSED",
			expectedName:   "Pod Status",
		},
		{
			name: "running pod with some containers not ready",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true},
						{Ready: false},
					},
				},
			},
			expectedStatus: "WARNING",
			expectedName:   "Pod Status",
		},
		{
			name: "pending pod",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			expectedStatus: "FAILED",
			expectedName:   "Pod Status",
		},
		{
			name: "failed pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase:  corev1.PodFailed,
					Reason: "Evicted",
				},
			},
			expectedStatus: "FAILED",
			expectedName:   "Pod Status",
		},
		{
			name: "succeeded pod",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
			expectedStatus: "PASSED",
			expectedName:   "Pod Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PodInfo{Pod: tt.pod}
			result := diagnostic.checkPodBasicStatus(info)

			if string(result.Status) != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, result.Name)
			}
		})
	}
}

func TestCheckImageIssues(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name           string
		pod            *corev1.Pod
		expectedChecks int
		expectedStatus string
	}{
		{
			name: "successful image pull",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Image: "nginx:latest",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			expectedChecks: 1,
			expectedStatus: "PASSED",
		},
		{
			name: "image pull back off",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Image: "invalid-registry/nonexistent:latest",
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason:  "ImagePullBackOff",
									Message: "Failed to pull image",
								},
							},
						},
					},
				},
			},
			expectedChecks: 1,
			expectedStatus: "FAILED",
		},
		{
			name: "invalid image name",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "app",
							Image: "invalid-image-name",
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "InvalidImageName",
								},
							},
						},
					},
				},
			},
			expectedChecks: 1,
			expectedStatus: "FAILED",
		},
		{
			name: "init container image issue",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "init",
							Image: "init:latest",
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason:  "ErrImagePull",
									Message: "Image not found",
								},
							},
						},
					},
				},
			},
			expectedChecks: 1,
			expectedStatus: "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PodInfo{Pod: tt.pod}
			results := diagnostic.checkImageIssues(info)

			if len(results) != tt.expectedChecks {
				t.Errorf("Expected %d checks, got %d", tt.expectedChecks, len(results))
			}

			if len(results) > 0 && string(results[0].Status) != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, results[0].Status)
			}
		})
	}
}

func TestGetImagePullSuggestion(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name          string
		image         string
		message       string
		shouldContain string
	}{
		{
			name:          "image not found",
			image:         "nginx:nonexistent",
			message:       "manifest unknown: manifest unknown",
			shouldContain: "verify image name and tag",
		},
		{
			name:          "unauthorized access",
			image:         "private/repo:latest",
			message:       "unauthorized: authentication required",
			shouldContain: "registry credentials",
		},
		{
			name:          "connection timeout",
			image:         "registry.example.com/app:v1",
			message:       "net/http: request timeout",
			shouldContain: "network connectivity",
		},
		{
			name:          "rate limit",
			image:         "nginx:latest",
			message:       "pull rate limit exceeded",
			shouldContain: "rate limit",
		},
		{
			name:          "generic error",
			image:         "app:latest",
			message:       "some other error",
			shouldContain: "image name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := diagnostic.getImagePullSuggestion(tt.image, tt.message)
			if !containsSubstring(suggestion, tt.shouldContain) {
				t.Errorf("Expected suggestion to contain %q, got %q", tt.shouldContain, suggestion)
			}
		})
	}
}

func TestCheckQoSClass(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name           string
		qosClass       corev1.PodQOSClass
		expectedStatus string
	}{
		{
			name:           "guaranteed QoS",
			qosClass:       corev1.PodQOSGuaranteed,
			expectedStatus: "PASSED",
		},
		{
			name:           "burstable QoS",
			qosClass:       corev1.PodQOSBurstable,
			expectedStatus: "PASSED",
		},
		{
			name:           "best effort QoS",
			qosClass:       corev1.PodQOSBestEffort,
			expectedStatus: "WARNING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					QOSClass: tt.qosClass,
				},
			}

			result := diagnostic.checkQoSClass(pod)

			if string(result.Status) != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}
			if result.Name != "Resource QoS" {
				t.Errorf("Expected name 'Resource QoS', got %s", result.Name)
			}
		})
	}
}

func TestCheckResourceRequests(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name           string
		pod            *corev1.Pod
		expectedStatus string
	}{
		{
			name: "has both requests and limits",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
			expectedStatus: "PASSED",
		},
		{
			name: "has requests only",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
			expectedStatus: "WARNING",
		},
		{
			name: "has limits only",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("200m"),
								},
							},
						},
					},
				},
			},
			expectedStatus: "WARNING",
		},
		{
			name: "no resources configured",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "app",
							Resources: corev1.ResourceRequirements{},
						},
					},
				},
			},
			expectedStatus: "WARNING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diagnostic.checkResourceRequests(tt.pod)

			if string(result.Status) != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}
			if result.Name != "Resource Configuration" {
				t.Errorf("Expected name 'Resource Configuration', got %s", result.Name)
			}
		})
	}
}

func TestAnalyzeContainerLogs(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name           string
		containerName  string
		logs           string
		expectedStatus string
	}{
		{
			name:           "clean logs",
			containerName:  "app",
			logs:           "Starting application\nServer listening on port 8080\nReady to accept connections",
			expectedStatus: "PASSED",
		},
		{
			name:           "connection refused error",
			containerName:  "app",
			logs:           "Starting application\nError: connection refused to database\nShutting down",
			expectedStatus: "FAILED",
		},
		{
			name:           "DNS error",
			containerName:  "app",
			logs:           "Starting application\nError: no such host: db.example.com\nRetrying...",
			expectedStatus: "FAILED",
		},
		{
			name:           "permission denied",
			containerName:  "app",
			logs:           "Starting application\nError: permission denied accessing /etc/secret\nFailed to start",
			expectedStatus: "FAILED",
		},
		{
			name:           "out of memory",
			containerName:  "app",
			logs:           "Processing request\nError: out of memory\nProcess killed",
			expectedStatus: "FAILED",
		},
		{
			name:           "authentication failure",
			containerName:  "app",
			logs:           "Connecting to service\nAuthentication failed: invalid credentials\nRetrying with new token",
			expectedStatus: "FAILED",
		},
		{
			name:           "timeout error",
			containerName:  "app",
			logs:           "Making request to API\nError: request timed out after 30s\nRetrying",
			expectedStatus: "FAILED",
		},
		{
			name:           "panic in application",
			containerName:  "app",
			logs:           "Processing request\npanic: runtime error: nil pointer dereference\ngoroutine 1 [running]:",
			expectedStatus: "FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diagnostic.analyzeContainerLogs(tt.containerName, tt.logs)

			if string(result.Status) != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, result.Status)
			}

			expectedName := "Container " + tt.containerName + " - Log Analysis"
			if result.Name != expectedName {
				t.Errorf("Expected name %s, got %s", expectedName, result.Name)
			}
		})
	}
}

func TestIsContainerCrashLooping(t *testing.T) {
	diagnostic := &PodDiagnostic{}

	tests := []struct {
		name          string
		pod           *corev1.Pod
		containerName string
		expected      bool
	}{
		{
			name: "crash looping container",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 5,
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					},
				},
			},
			containerName: "app",
			expected:      true,
		},
		{
			name: "healthy container",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			containerName: "app",
			expected:      false,
		},
		{
			name: "restarted but not crash looping",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "app",
							RestartCount: 2,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			containerName: "app",
			expected:      false,
		},
		{
			name: "container not found",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "other",
							RestartCount: 0,
						},
					},
				},
			},
			containerName: "app",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diagnostic.isContainerCrashLooping(tt.pod, tt.containerName)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) &&
		len(substr) > 0 &&
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
