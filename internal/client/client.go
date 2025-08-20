package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// KubernetesClient wraps the Kubernetes clientset with additional metadata
type KubernetesClient struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Context   string
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient(kubeconfig string) (*KubernetesClient, error) {
	var config *rest.Config

	var err error

	if kubeconfig == "" {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to kubeconfig file using proper loading rules
			// This handles KUBECONFIG env var and default paths correctly
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
			config, err = clientConfig.ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
			}
		}
	} else {
		// Load config from specific kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get current context
	context := getCurrentContextFromConfig(config, kubeconfig)

	return &KubernetesClient{
		Clientset: clientset,
		Config:    config,
		Context:   context,
	}, nil
}

// TestConnection tests the connection to the Kubernetes cluster
func (k *KubernetesClient) TestConnection(ctx context.Context) error {
	_, err := k.Clientset.Discovery().ServerVersion()
	if err != nil {
		// Provide more helpful error messages for common issues
		errMsg := err.Error()
		if strings.Contains(errMsg, "the server has asked for the client to provide credentials") {
			return fmt.Errorf("authentication failed - please check your credentials:\n"+
				"  • For EKS: run 'aws eks update-kubeconfig --region <region> --name <cluster-name>'\n"+
				"  • Ensure AWS credentials are valid: 'aws sts get-caller-identity'\n"+
				"  • Check kubeconfig: 'kubectl cluster-info'\n"+
				"Original error: %w", err)
		}

		if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "connection refused") {
			return fmt.Errorf("cluster unreachable - please check network connectivity:\n"+
				"  • Verify cluster is running\n"+
				"  • Check kubeconfig server URL\n"+
				"  • Ensure network access to cluster\n"+
				"Original error: %w", err)
		}

		return fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}

	return nil
}

// GetClusterInfo returns basic cluster information
func (k *KubernetesClient) GetClusterInfo(ctx context.Context) (map[string]string, error) {
	version, err := k.Clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	info := map[string]string{
		"context":    k.Context,
		"server":     k.Config.Host,
		"version":    version.String(),
		"gitVersion": version.GitVersion,
		"platform":   version.Platform,
	}

	return info, nil
}

// getDefaultKubeconfigPath returns the default kubeconfig path
func getDefaultKubeconfigPath() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}

// getCurrentContextFromConfig returns the current context from the loaded config
func getCurrentContextFromConfig(_ *rest.Config, kubeconfig string) string {
	// If we have a specific kubeconfig file, try to get context from it
	if kubeconfig != "" {
		if rawConfig, err := clientcmd.LoadFromFile(kubeconfig); err == nil {
			return rawConfig.CurrentContext
		}
	}

	// Fall back to getting context using default loading rules
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	if rawConfig, err := clientConfig.RawConfig(); err == nil {
		return rawConfig.CurrentContext
	}

	return "unknown"
}

// getCurrentContext returns the current context from kubeconfig
func getCurrentContext(kubeconfig string) (string, error) {
	if kubeconfig == "" {
		kubeconfig = getDefaultKubeconfigPath()
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return "", err
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil
}
