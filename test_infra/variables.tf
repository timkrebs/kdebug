# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-2"
}

variable "cluster_version" {
  description = "Kubernetes cluster version"
  type        = string
  default     = "1.29"
}

variable "node_instance_type" {
  description = "Instance type for EKS worker nodes"
  type        = string
  default     = "t3.small"
}

variable "enable_test_workloads" {
  description = "Enable deployment of test workloads for kdebug testing"
  type        = bool
  default     = true
}
