#!/bin/bash

# Fix Terraform state after the rate limiting error
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Fixing Terraform State After Rate Limiting Error${NC}"
echo -e "${BLUE}=================================================${NC}"

echo -e "${YELLOW}📋 Current situation:${NC}"
echo "• Terraform deployment failed due to Kubernetes API rate limiting"
echo "• The EKS cluster was created successfully"
echo "• Test pods failed to deploy via Terraform"
echo "• We'll clean up the failed state and continue with kubectl deployment"
echo ""

# Check terraform state
if [ ! -f "terraform.tfstate" ]; then
    echo -e "${RED}❌ terraform.tfstate not found${NC}"
    exit 1
fi

echo -e "${YELLOW}🧹 Cleaning up failed Kubernetes resources from state...${NC}"

# Remove the failed Kubernetes resources from terraform state
# This allows us to manage them via kubectl instead

resources_to_remove=(
    "kubernetes_pod.healthy_pod"
    "kubernetes_pod.image_pull_error_pod"
    "kubernetes_pod.crash_loop_pod"
    "kubernetes_pod.unschedulable_pod"
    "kubernetes_pod.init_container_failure_pod"
    "kubernetes_pod.rbac_issue_pod"
    "kubernetes_pod.oom_pod"
    "kubernetes_pod.dependency_pod"
    "kubernetes_pod.best_effort_pod"
    "kubernetes_deployment.multi_pod_test"
    "kubernetes_service_account.test_sa"
    "kubernetes_role.limited_role"
    "kubernetes_role_binding.limited_binding"
)

echo "Removing failed resources from Terraform state..."
for resource in "${resources_to_remove[@]}"; do
    if terraform state show "$resource" >/dev/null 2>&1; then
        echo "  Removing: $resource"
        terraform state rm "$resource" >/dev/null 2>&1 || true
    fi
done

echo -e "${GREEN}✅ Terraform state cleaned up${NC}"
echo ""

echo -e "${YELLOW}🏗️ Re-applying Terraform with fixed configuration...${NC}"

# Re-run terraform apply with the fixed configuration
terraform apply -auto-approve

echo -e "${GREEN}✅ Terraform deployment completed successfully!${NC}"
echo ""

echo -e "${YELLOW}🔗 Configuring kubectl...${NC}"
./connect-eks.sh

echo -e "${YELLOW}🧪 Deploying test pods via kubectl...${NC}"
./deploy-test-pods.sh

echo -e "${GREEN}🎉 Fix completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📋 Summary:${NC}"
echo "• EKS cluster is fully operational"
echo "• Test pods are deployed via kubectl"
echo "• You can now test kdebug with: ./test-kdebug.sh"
echo "• All infrastructure is ready for testing"
