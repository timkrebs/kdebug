---
layout: docs
title: Examples
description: Real-world examples and use cases for kdebug in different scenarios.
permalink: /docs/examples/
order: 4
---

# Examples

This page provides real-world examples and use cases for using kdebug to diagnose common Kubernetes issues.

## Scenario 1: Pod Stuck in Pending State

### Problem
Your application pod has been pending for several minutes and won't start.

### Solution

```bash
# Check the specific pod
kdebug pod myapp-deployment-7d4b8c6f9-x8k2l --verbose

# Check all pending pods in namespace
kdebug pod --selector app=myapp
```

### Expected Output

```
🔍 Analyzing pod: myapp-deployment-7d4b8c6f9-x8k2l

❌ CRITICAL: Pod is in Pending state
┌─────────────────────────────────────────────────────────────┐
│ Issue: Insufficient CPU resources                           │
│ Cause: Node has insufficient CPU to schedule this pod      │
│                                                             │
│ 🔧 Remediation:                                            │
│ 1. Reduce CPU requests in deployment spec                  │
│ 2. Add more nodes to cluster                               │
│ 3. Check if other pods can be scaled down                  │
└─────────────────────────────────────────────────────────────┘

📊 Resource Analysis:
• Requested: 2 CPU, 4Gi memory
• Available on best node: 1.5 CPU, 8Gi memory
• Taints: None blocking this pod

💡 Quick Fix: kubectl patch deployment myapp --patch '{"spec":{"template":{"spec":{"containers":[{"name":"myapp","resources":{"requests":{"cpu":"1"}}}]}}}}'
```

## Scenario 2: CrashLoopBackOff Issue

### Problem
Your application is constantly restarting and stuck in CrashLoopBackOff.

### Solution

```bash
# Analyze the crashing pod with logs
kdebug pod myapp-web-abc123 --since 10m --tail 200
```

### Expected Output

```
🔍 Analyzing pod: myapp-web-abc123

❌ CRITICAL: Pod in CrashLoopBackOff (restart count: 47)
┌─────────────────────────────────────────────────────────────┐
│ Issue: Application failing to start                        │
│ Cause: Connection refused to database                      │
│                                                             │
│ 🔧 Remediation:                                            │
│ 1. Check database service connectivity                     │
│ 2. Verify environment variables: DB_HOST, DB_PORT          │
│ 3. Check network policies allowing database access         │
└─────────────────────────────────────────────────────────────┘

📝 Recent Logs:
2025-01-22 10:15:23 [ERROR] Failed to connect to database: connection refused
2025-01-22 10:15:23 [ERROR] Database host: postgres-service:5432
2025-01-22 10:15:23 [FATAL] Application startup failed

🔗 Related Resources:
• Service: postgres-service (❌ No endpoints)
• ConfigMap: myapp-config (✅ Exists)
• Secret: myapp-secrets (✅ Exists)

💡 Next Steps:
1. Check postgres pod: kdebug pod --selector app=postgres
2. Verify service endpoints: kubectl get endpoints postgres-service
```

## Scenario 3: Cluster Node Issues

### Problem
Pods are failing to schedule and nodes appear unhealthy.

### Solution

```bash
# Check overall cluster health
kdebug cluster --verbose

# Focus on a specific node
kdebug cluster --node worker-node-2
```

### Expected Output

```
🏗️ Cluster Health Analysis

✅ HEALTHY: Control plane components
✅ HEALTHY: DNS resolution (CoreDNS)
❌ CRITICAL: 1 node unhealthy
⚠️  WARNING: High memory usage on 2 nodes

📊 Node Summary:
┌─────────────────┬─────────┬─────────┬──────────┬──────────┐
│ Node            │ Status  │ CPU     │ Memory   │ Pods     │
├─────────────────┼─────────┼─────────┼──────────┼──────────┤
│ master-1        │ Ready   │ 45%     │ 67%      │ 12/110   │
│ worker-node-1   │ Ready   │ 78%     │ 89% ⚠️   │ 45/110   │
│ worker-node-2   │ NotReady│ --      │ --       │ 0/110    │
│ worker-node-3   │ Ready   │ 82%     │ 91% ⚠️   │ 48/110   │
└─────────────────┴─────────┴─────────┴──────────┴──────────┘

❌ Node Issues:
worker-node-2:
  • Status: NotReady (5m ago)
  • Condition: KubeletNotReady
  • Message: container runtime not responding

🔧 Remediation Steps:
1. SSH to worker-node-2 and restart kubelet:
   systemctl restart kubelet
2. Check container runtime:
   systemctl status containerd
3. Monitor node recovery:
   watch kubectl get nodes
```

## Scenario 4: Application in Production Namespace

### Problem
Production application experiencing intermittent issues.

### Solution

```bash
# Check all production pods
kdebug pod --namespace production --verbose

# Focus on specific application
kdebug pod --namespace production --selector app=frontend
```

### Expected Output

```
🔍 Production Namespace Analysis: 47 pods found

✅ HEALTHY: 42 pods running normally
⚠️  WARNING: 3 pods with minor issues
❌ CRITICAL: 2 pods failing

📊 Pod Status Summary:
• Running: 42 pods
• Pending: 1 pod
• CrashLoopBackOff: 2 pods
• ImagePullBackOff: 1 pod
• Completed: 1 pod

❌ Critical Issues:
1. frontend-api-xyz789 (CrashLoopBackOff)
   └─ Memory limit exceeded, pod killed by OOMKiller
   
2. worker-queue-abc456 (CrashLoopBackOff)
   └─ Redis connection timeout

⚠️  Warnings:
1. frontend-web-def123 (High CPU usage: 95%)
2. database-read-ghi789 (High memory usage: 89%)
3. background-job-jkl012 (Pending: waiting for node resources)

🔧 Immediate Actions Required:
1. Increase memory limit for frontend-api
2. Check Redis service connectivity
3. Scale down non-critical workloads or add nodes
```

## Scenario 5: Image Pull Issues

### Problem
New deployment fails because pods can't pull container images.

### Solution

```bash
# Check pods with image pull issues
kdebug pod --selector app=newapp --since 5m
```

### Expected Output

```
🔍 Analyzing pods with selector: app=newapp

❌ CRITICAL: ImagePullBackOff on 3 pods
┌─────────────────────────────────────────────────────────────┐
│ Issue: Failed to pull image                                │
│ Image: registry.company.com/myapp:v2.1.0                   │
│ Error: unauthorized: authentication required               │
│                                                             │
│ 🔧 Remediation:                                            │
│ 1. Check image registry credentials                        │
│ 2. Verify imagePullSecrets in deployment                   │
│ 3. Test registry access from cluster                       │
└─────────────────────────────────────────────────────────────┘

🔍 Registry Analysis:
• Registry: registry.company.com
• Image: myapp:v2.1.0
• Pull Policy: Always
• Secrets: regcred (❌ Not found in namespace)

💡 Quick Fixes:
1. Create registry secret:
   kubectl create secret docker-registry regcred \
     --docker-server=registry.company.com \
     --docker-username=<username> \
     --docker-password=<password>

2. Add secret to deployment:
   kubectl patch deployment newapp -p '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"regcred"}]}}}}'
```

## Scenario 6: Cross-Namespace Communication Issues

### Problem
Service in one namespace can't communicate with service in another namespace.

### Solution

```bash
# Check networking and services
kdebug cluster --check-networking
kdebug pod --namespace app1 --selector app=frontend
```

### Expected Output

```
🔍 Network Connectivity Analysis

✅ HEALTHY: CoreDNS functioning
✅ HEALTHY: Cluster networking
❌ CRITICAL: Network policy blocking traffic

🚫 Network Policy Issues:
Policy: deny-all-ingress (namespace: app2)
Effect: Blocking traffic from app1/frontend to app2/backend

🔧 Remediation:
1. Create network policy to allow app1 → app2 communication:

apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-app1-to-app2
  namespace: app2
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: app1
```

## Scenario 7: Resource Constraints

### Problem
Cluster performance is degrading and pods are being evicted.

### Solution

```bash
# Check cluster resource usage
kdebug cluster --verbose

# Check resource-intensive pods
kdebug pod --all-namespaces --verbose
```

### Expected Output

```
🏗️ Cluster Resource Analysis

⚠️  WARNING: High resource utilization detected
❌ CRITICAL: Pods being evicted due to resource pressure

📊 Resource Utilization:
┌──────────────────┬─────────┬─────────┬──────────┐
│ Resource         │ Used    │ Total   │ Usage %  │
├──────────────────┼─────────┼─────────┼──────────┤
│ CPU              │ 15.2    │ 16.0    │ 95% ❌   │
│ Memory           │ 28.1Gi  │ 32.0Gi  │ 88% ⚠️   │
│ Storage          │ 180Gi   │ 200Gi   │ 90% ⚠️   │
└──────────────────┴─────────┴─────────┴──────────┘

🔥 Resource Pressure Events:
• Node worker-1: MemoryPressure, DiskPressure
• Evicted pods: 5 in last hour
• Failed scheduling: 12 pods pending

🔧 Immediate Actions:
1. Scale critical applications:
   kubectl scale deployment critical-app --replicas=2
2. Add cluster nodes or increase node capacity
3. Review resource requests/limits:
   kubectl top pods --all-namespaces --sort-by memory
```

## Scenario 8: Development Environment Setup

### Problem
Setting up a development environment and ensuring everything is working correctly.

### Solution

```bash
# Quick health check for development cluster
kdebug cluster

# Check development namespace
kdebug pod --namespace development
```

### Expected Output

```
🏗️ Development Cluster Health Check

✅ HEALTHY: All systems operational
✅ HEALTHY: 15 pods running normally
✅ HEALTHY: DNS resolution working
✅ HEALTHY: Node resources available

📊 Development Environment Status:
• Nodes: 3 ready
• Namespaces: 4 active
• Storage Classes: 2 available
• Ingress Controller: nginx (ready)

🎯 Development Ready:
All systems are healthy and ready for development work!

💡 Useful Commands:
• Deploy test app: kubectl apply -f examples/
• Port forward: kubectl port-forward svc/myapp 8080:80
• View logs: kubectl logs -f deployment/myapp
```

## Common Troubleshooting Workflow

Here's a systematic approach to troubleshooting with kdebug:

### 1. Start with Cluster Overview
```bash
kdebug cluster
```

### 2. Focus on Specific Namespace
```bash
kdebug pod --namespace <your-namespace>
```

### 3. Drill Down to Specific Pods
```bash
kdebug pod <pod-name> --verbose
```

### 4. Follow Logs if Needed
```bash
kdebug pod <pod-name> --follow --since 10m
```

### 5. Check Related Resources
```bash
# Based on kdebug suggestions
kubectl describe service <service-name>
kubectl get endpoints <service-name>
```

## Tips for Effective Debugging

1. **Use verbose mode** (`--verbose`) for detailed analysis
2. **Check logs with context** using `--since` and `--tail` flags
3. **Use label selectors** to focus on specific applications
4. **Start broad, then narrow down** - cluster → namespace → pod
5. **Follow kdebug suggestions** for quick fixes and next steps

## Next Steps

- [Commands Reference]({{ '/docs/commands/' | relative_url }}) - Complete command documentation
- [Contributing]({{ '/docs/contributing/' | relative_url }}) - Help improve kdebug
- [Installation]({{ '/docs/installation/' | relative_url }}) - Install kdebug on your system