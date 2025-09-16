# Test workloads for kdebug pod diagnostics - Fixed version
# This file creates various pod scenarios to test different kdebug diagnostic capabilities

# Data source to wait for cluster to be ready
data "aws_eks_cluster" "cluster" {
  name = module.eks.cluster_name
  depends_on = [module.eks]
}

# Null resource to wait for nodes to be ready
resource "null_resource" "wait_for_cluster" {
  depends_on = [module.eks.eks_managed_node_groups]

  provisioner "local-exec" {
    command = <<-EOT
      echo "Waiting for cluster to be ready..."
      aws eks --region ${var.region} update-kubeconfig --name ${module.eks.cluster_name}
      
      # Wait for nodes to be ready
      echo "Waiting for nodes to be ready..."
      kubectl wait --for=condition=Ready nodes --all --timeout=600s
      
      # Wait for system pods to be ready
      echo "Waiting for system pods to be ready..."
      kubectl wait --for=condition=Ready pods -n kube-system -l k8s-app=kube-dns --timeout=300s || true
      kubectl wait --for=condition=Ready pods -n kube-system -l app.kubernetes.io/name=aws-load-balancer-controller --timeout=300s || true
      
      echo "Cluster is ready!"
    EOT
  }
}

# Create namespace only if test workloads are enabled
resource "kubernetes_namespace" "kdebug_test" {
  count = var.enable_test_workloads ? 1 : 0
  
  metadata {
    name = "kdebug-test"
    labels = {
      purpose = "kdebug-diagnostics-testing"
    }
  }
  
  depends_on = [null_resource.wait_for_cluster]
}

# Note: Due to the Terraform rate limiting issues, we'll create a separate
# script to apply workloads after the cluster is ready
# This provides a more reliable deployment experience
