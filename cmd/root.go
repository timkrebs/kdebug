/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kdebug",
	Short: "A CLI tool that automatically diagnoses common Kubernetes issues",
	Long: `kdebug is a diagnostic tool for Kubernetes clusters that automatically
identifies common issues and provides actionable suggestions.

Think of it as a "doctor" for Kubernetes clusters (like 'brew doctor', but for K8s).
Instead of manually digging through kubectl describe outputs and events, kdebug 
runs a series of checks and gives clear guidance on what's wrong and how to fix it.

Examples:
  kdebug cluster                           # Run cluster-wide health checks
  kdebug pod myapp-123 -n production      # Debug a specific pod
  kdebug service myservice                 # Check service and endpoints
  kdebug dns                               # Test DNS resolution`,
	Version: "0.1.0-dev",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	// Global flags
	kubeconfig   string
	namespace    string
	outputFormat string
	verbose      bool
)

func init() {
	// Global persistent flags that apply to all commands
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file (defaults to $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output for debugging")
}
