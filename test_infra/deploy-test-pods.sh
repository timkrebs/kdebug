#!/bin/bash

# Deploy test pods for kdebug diagnostics after cluster is ready
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Deploying kdebug Test Pods${NC}"
echo -e "${BLUE}==============================${NC}"

# Check if kubectl is configured
if ! kubectl cluster-info >/dev/null 2>&1; then
    echo -e "${RED}❌ kubectl is not configured or cluster is not accessible${NC}"
    echo "Please run: ./connect-eks.sh"
    exit 1
fi

# Check if test namespace exists
NAMESPACE="kdebug-test"
if ! kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
    echo -e "${YELLOW}📁 Creating namespace $NAMESPACE...${NC}"
    kubectl create namespace $NAMESPACE
else
    echo -e "${GREEN}✅ Namespace $NAMESPACE already exists${NC}"
fi

echo -e "${YELLOW}🚀 Deploying test workloads...${NC}"

# Create test pod manifests
cat > /tmp/kdebug-test-pods.yaml << 'EOF'
---
# Service account for RBAC testing
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kdebug-test-sa
  namespace: kdebug-test
---
# Limited role for RBAC testing (intentionally restricted)
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: kdebug-test
  name: limited-role
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
  # Intentionally missing "create", "delete" to test RBAC issues
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: limited-binding
  namespace: kdebug-test
subjects:
- kind: ServiceAccount
  name: kdebug-test-sa
  namespace: kdebug-test
roleRef:
  kind: Role
  name: limited-role
  apiGroup: rbac.authorization.k8s.io
---
# 1. Healthy pod for baseline testing
apiVersion: v1
kind: Pod
metadata:
  name: healthy-test-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: healthy
spec:
  containers:
  - name: nginx
    image: nginx:1.21
    ports:
    - containerPort: 80
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 100m
        memory: 128Mi
    livenessProbe:
      httpGet:
        path: /
        port: 80
      initialDelaySeconds: 10
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /
        port: 80
      initialDelaySeconds: 5
      periodSeconds: 5
  restartPolicy: Always
---
# 2. Pod with image pull error (invalid image)
apiVersion: v1
kind: Pod
metadata:
  name: image-pull-error-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: image-pull-error
spec:
  containers:
  - name: failing-container
    image: nonexistent-registry.example.com/invalid-image:latest
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
  restartPolicy: Never
---
# 3. Pod that will crash loop (exits with error)
apiVersion: v1
kind: Pod
metadata:
  name: crash-loop-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: crash-loop
spec:
  containers:
  - name: crashing-container
    image: busybox:1.35
    command: ["sh", "-c"]
    args:
    - "echo 'Starting application...'; echo 'Error: connection refused to database at db.example.com:5432'; echo 'Application failed to start'; exit 1"
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
  restartPolicy: Always
---
# 4. Pod with resource constraints that can't be scheduled
apiVersion: v1
kind: Pod
metadata:
  name: unschedulable-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: unschedulable
spec:
  containers:
  - name: resource-hungry
    image: nginx:1.21
    resources:
      requests:
        cpu: "10000m"  # 10 CPUs - more than t3.small can provide
        memory: "16Gi"  # 16GB - more than t3.small has
  restartPolicy: Never
---
# 5. Pod with init container that fails
apiVersion: v1
kind: Pod
metadata:
  name: init-failure-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: init-failure
spec:
  initContainers:
  - name: failing-init
    image: busybox:1.35
    command: ["sh", "-c"]
    args:
    - "echo 'Initializing...'; echo 'Checking database connection...'; echo 'Error: database not reachable'; exit 1"
  containers:
  - name: main-container
    image: nginx:1.21
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
  restartPolicy: Never
---
# 6. Pod with RBAC issues (uses restricted service account)
apiVersion: v1
kind: Pod
metadata:
  name: rbac-issue-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: rbac-issue
spec:
  serviceAccountName: kdebug-test-sa
  containers:
  - name: kubectl-test
    image: bitnami/kubectl:1.28
    command: ["sh", "-c"]
    args:
    - "echo 'Testing RBAC permissions...'; kubectl create configmap test-cm --from-literal=key=value; sleep 3600"
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
  restartPolicy: Never
---
# 7. Pod with memory issues (OOM)
apiVersion: v1
kind: Pod
metadata:
  name: oom-test-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: oom
spec:
  containers:
  - name: memory-hungry
    image: alpine:3.18
    command: ["sh", "-c"]
    args:
    - "echo 'Starting memory allocation test...'; dd if=/dev/zero of=/tmp/memory.dat bs=1M count=200; sleep 3600"
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
      limits:
        cpu: 50m
        memory: 64Mi  # Will be exceeded by the dd command
  restartPolicy: Never
---
# 8. Pod with complex dependencies (database connection simulation)
apiVersion: v1
kind: Pod
metadata:
  name: dependency-test-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: dependency-issues
spec:
  initContainers:
  - name: db-check
    image: busybox:1.35
    command: ["sh", "-c"]
    args:
    - "echo 'Checking database connectivity...'; nslookup postgres.database.svc.cluster.local || echo 'DNS resolution failed for database'"
  containers:
  - name: app
    image: alpine:3.18
    command: ["sh", "-c"]
    args:
    - "echo 'Application starting...'; echo 'Connecting to Redis...'; nc -z redis.cache.svc.cluster.local 6379 || echo 'Redis connection failed'; sleep 3600"
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
    env:
    - name: DATABASE_URL
      value: "postgresql://user:pass@postgres.database.svc.cluster.local:5432/app"
    - name: REDIS_URL
      value: "redis://redis.cache.svc.cluster.local:6379"
  restartPolicy: Always
---
# 9. Pod without resource requests/limits (BestEffort QoS)
apiVersion: v1
kind: Pod
metadata:
  name: best-effort-pod
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: best-effort-qos
spec:
  containers:
  - name: no-resources
    image: nginx:1.21
    ports:
    - containerPort: 80
    # Intentionally no resource requests or limits
  restartPolicy: Always
---
# 10. Deployment for testing multiple pods at once
apiVersion: apps/v1
kind: Deployment
metadata:
  name: multi-pod-deployment
  namespace: kdebug-test
  labels:
    app: kdebug-test
    scenario: multi-pod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: multi-pod-test
  template:
    metadata:
      labels:
        app: multi-pod-test
        scenario: multi-pod
    spec:
      containers:
      - name: web
        image: nginx:1.21
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 25m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 128Mi
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 10
      restartPolicy: Always
EOF

# Apply the manifests
echo -e "${YELLOW}📋 Applying test pod manifests...${NC}"
kubectl apply -f /tmp/kdebug-test-pods.yaml

# Clean up temporary file
rm /tmp/kdebug-test-pods.yaml

echo -e "${GREEN}✅ Test pods deployed successfully!${NC}"
echo ""

# Show status
echo -e "${YELLOW}📊 Current Pod Status:${NC}"
kubectl get pods -n kdebug-test
echo ""

echo -e "${YELLOW}💡 Tips:${NC}"
echo "• Some pods are designed to fail (this is expected)"
echo "• Wait a few minutes for all pods to reach their expected states"
echo "• Check pod status: kubectl get pods -n kdebug-test"
echo "• Run kdebug tests: ./test-kdebug.sh"
echo ""

echo -e "${GREEN}🎉 Test pod deployment completed!${NC}"
