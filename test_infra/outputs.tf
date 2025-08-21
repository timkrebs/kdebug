# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ids attached to the cluster control plane"
  value       = module.eks.cluster_security_group_id
}

output "region" {
  description = "AWS region"
  value       = var.region
}

output "cluster_name" {
  description = "Kubernetes Cluster Name"
  value       = module.eks.cluster_name
}

output "kubectl_config_command" {
  description = "Command to configure kubectl"
  value       = "aws eks --region ${var.region} update-kubeconfig --name ${module.eks.cluster_name}"
}

output "test_namespace" {
  description = "Namespace containing test pods for kdebug"
  value       = "kdebug-test"
}

output "test_pods_info" {
  description = "Information about test pods for kdebug diagnostics"
  value = {
    healthy_pod              = "healthy-test-pod"
    image_pull_error_pod     = "image-pull-error-pod"
    crash_loop_pod          = "crash-loop-pod"
    unschedulable_pod       = "unschedulable-pod"
    init_failure_pod        = "init-failure-pod"
    rbac_issue_pod          = "rbac-issue-pod"
    oom_pod                 = "oom-test-pod"
    dependency_pod          = "dependency-test-pod"
    best_effort_pod         = "best-effort-pod"
    multi_pod_deployment    = "multi-pod-deployment"
  }
}
