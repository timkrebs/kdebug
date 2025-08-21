# Test workloads for kdebug pod diagnostics
# This file creates various pod scenarios to test different kdebug diagnostic capabilities

# Note: Kubernetes provider is configured in main.tf

# Namespace for test workloads
resource "kubernetes_namespace" "kdebug_test" {
  metadata {
    name = "kdebug-test"
    labels = {
      purpose = "kdebug-diagnostics-testing"
    }
  }
}

# Service account for RBAC testing
resource "kubernetes_service_account" "test_sa" {
  metadata {
    name      = "kdebug-test-sa"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
  }
}

# Limited role for RBAC testing (intentionally restricted)
resource "kubernetes_role" "limited_role" {
  metadata {
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    name      = "limited-role"
  }

  rule {
    api_groups = [""]
    resources  = ["pods"]
    verbs      = ["get", "list"]
    # Intentionally missing "create", "delete" to test RBAC issues
  }
}

resource "kubernetes_role_binding" "limited_binding" {
  metadata {
    name      = "limited-binding"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.limited_role.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.test_sa.metadata[0].name
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
  }
}

# 1. Healthy pod for baseline testing
resource "kubernetes_pod" "healthy_pod" {
  metadata {
    name      = "healthy-test-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "healthy"
    }
  }

  spec {
    container {
      name  = "nginx"
      image = "nginx:1.21"
      
      port {
        container_port = 80
      }

      resources {
        requests = {
          cpu    = "50m"
          memory = "64Mi"
        }
        limits = {
          cpu    = "100m"
          memory = "128Mi"
        }
      }

      liveness_probe {
        http_get {
          path = "/"
          port = 80
        }
        initial_delay_seconds = 10
        period_seconds        = 10
      }

      readiness_probe {
        http_get {
          path = "/"
          port = 80
        }
        initial_delay_seconds = 5
        period_seconds        = 5
      }
    }

    restart_policy = "Always"
  }
}

# 2. Pod with image pull error (invalid image)
resource "kubernetes_pod" "image_pull_error_pod" {
  metadata {
    name      = "image-pull-error-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "image-pull-error"
    }
  }

  spec {
    container {
      name  = "failing-container"
      image = "nonexistent-registry.example.com/invalid-image:latest"
      
      resources {
        requests = {
          cpu    = "10m"
          memory = "32Mi"
        }
      }
    }

    restart_policy = "Never"
  }
}

# 3. Pod that will crash loop (exits with error)
resource "kubernetes_pod" "crash_loop_pod" {
  metadata {
    name      = "crash-loop-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "crash-loop"
    }
  }

  spec {
    container {
      name  = "crashing-container"
      image = "busybox:1.35"
      
      command = ["sh", "-c"]
      args = [
        "echo 'Starting application...'; echo 'Error: connection refused to database at db.example.com:5432'; echo 'Application failed to start'; exit 1"
      ]

      resources {
        requests = {
          cpu    = "10m"
          memory = "32Mi"
        }
      }
    }

    restart_policy = "Always"
  }
}

# 4. Pod with resource constraints that can't be scheduled
resource "kubernetes_pod" "unschedulable_pod" {
  metadata {
    name      = "unschedulable-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "unschedulable"
    }
  }

  spec {
    container {
      name  = "resource-hungry"
      image = "nginx:1.21"
      
      resources {
        requests = {
          cpu    = "10000m"  # 10 CPUs - more than t3.micro can provide
          memory = "16Gi"    # 16GB - more than t3.micro has
        }
      }
    }

    restart_policy = "Never"
  }
}

# 5. Pod with init container that fails
resource "kubernetes_pod" "init_container_failure_pod" {
  metadata {
    name      = "init-failure-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "init-failure"
    }
  }

  spec {
    init_container {
      name  = "failing-init"
      image = "busybox:1.35"
      
      command = ["sh", "-c"]
      args = [
        "echo 'Initializing...'; echo 'Checking database connection...'; echo 'Error: database not reachable'; exit 1"
      ]
    }

    container {
      name  = "main-container"
      image = "nginx:1.21"
      
      resources {
        requests = {
          cpu    = "10m"
          memory = "32Mi"
        }
      }
    }

    restart_policy = "Never"
  }
}

# 6. Pod with RBAC issues (uses restricted service account)
resource "kubernetes_pod" "rbac_issue_pod" {
  metadata {
    name      = "rbac-issue-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "rbac-issue"
    }
  }

  spec {
    service_account_name = kubernetes_service_account.test_sa.metadata[0].name

    container {
      name  = "kubectl-test"
      image = "bitnami/kubectl:1.28"
      
      command = ["sh", "-c"]
      args = [
        "echo 'Testing RBAC permissions...'; kubectl create configmap test-cm --from-literal=key=value; sleep 3600"
      ]

      resources {
        requests = {
          cpu    = "10m"
          memory = "32Mi"
        }
      }
    }

    restart_policy = "Never"
  }
}

# 7. Pod with memory issues (OOM)
resource "kubernetes_pod" "oom_pod" {
  metadata {
    name      = "oom-test-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "oom"
    }
  }

  spec {
    container {
      name  = "memory-hungry"
      image = "alpine:3.18"
      
      command = ["sh", "-c"]
      args = [
        "echo 'Starting memory allocation test...'; dd if=/dev/zero of=/tmp/memory.dat bs=1M count=200; sleep 3600"
      ]

      resources {
        requests = {
          cpu    = "10m"
          memory = "32Mi"
        }
        limits = {
          cpu    = "50m"
          memory = "64Mi"  # Will be exceeded by the dd command
        }
      }
    }

    restart_policy = "Never"
  }
}

# 8. Pod with complex dependencies (database connection simulation)
resource "kubernetes_pod" "dependency_pod" {
  metadata {
    name      = "dependency-test-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "dependency-issues"
    }
  }

  spec {
    init_container {
      name  = "db-check"
      image = "busybox:1.35"
      
      command = ["sh", "-c"]
      args = [
        "echo 'Checking database connectivity...'; nslookup postgres.database.svc.cluster.local || echo 'DNS resolution failed for database'"
      ]
    }

    container {
      name  = "app"
      image = "alpine:3.18"
      
      command = ["sh", "-c"]
      args = [
        "echo 'Application starting...'; echo 'Connecting to Redis...'; nc -z redis.cache.svc.cluster.local 6379 || echo 'Redis connection failed'; sleep 3600"
      ]

      resources {
        requests = {
          cpu    = "10m"
          memory = "32Mi"
        }
      }

      env {
        name  = "DATABASE_URL"
        value = "postgresql://user:pass@postgres.database.svc.cluster.local:5432/app"
      }

      env {
        name  = "REDIS_URL"
        value = "redis://redis.cache.svc.cluster.local:6379"
      }
    }

    restart_policy = "Always"
  }
}

# 9. Pod without resource requests/limits (BestEffort QoS)
resource "kubernetes_pod" "best_effort_pod" {
  metadata {
    name      = "best-effort-pod"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "best-effort-qos"
    }
  }

  spec {
    container {
      name  = "no-resources"
      image = "nginx:1.21"
      
      port {
        container_port = 80
      }
      
      # Intentionally no resource requests or limits
    }

    restart_policy = "Always"
  }
}

# 10. Deployment for testing multiple pods at once
resource "kubernetes_deployment" "multi_pod_test" {
  metadata {
    name      = "multi-pod-deployment"
    namespace = kubernetes_namespace.kdebug_test.metadata[0].name
    labels = {
      app     = "kdebug-test"
      scenario = "multi-pod"
    }
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "multi-pod-test"
      }
    }

    template {
      metadata {
        labels = {
          app = "multi-pod-test"
          scenario = "multi-pod"
        }
      }

      spec {
        container {
          name  = "web"
          image = "nginx:1.21"
          
          port {
            container_port = 80
          }

          resources {
            requests = {
              cpu    = "25m"
              memory = "32Mi"
            }
            limits = {
              cpu    = "100m"
              memory = "128Mi"
            }
          }

          liveness_probe {
            http_get {
              path = "/"
              port = 80
            }
            initial_delay_seconds = 10
            period_seconds        = 10
          }
        }

        restart_policy = "Always"
      }
    }
  }
}
