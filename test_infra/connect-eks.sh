#!/bin/bash

# Connect to EKS cluster and configure kubectl
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🔗 Connecting to EKS cluster...${NC}"

# Check if terraform state exists
if [ ! -f "terraform.tfstate" ]; then
    echo -e "${RED}❌ terraform.tfstate not found. Please run 'terraform apply' first.${NC}"
    exit 1
fi

# Get cluster information from terraform outputs
REGION=$(terraform output -raw region 2>/dev/null)
CLUSTER_NAME=$(terraform output -raw cluster_name 2>/dev/null)

if [ -z "$REGION" ] || [ -z "$CLUSTER_NAME" ]; then
    echo -e "${RED}❌ Failed to get cluster information from terraform outputs${NC}"
    exit 1
fi

echo "Region: $REGION"
echo "Cluster: $CLUSTER_NAME"

# Update kubeconfig
echo -e "${YELLOW}Updating kubeconfig...${NC}"
aws eks --region "$REGION" update-kubeconfig --name "$CLUSTER_NAME"

# Test connection
echo -e "${YELLOW}Testing connection...${NC}"
if kubectl cluster-info >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Successfully connected to EKS cluster!${NC}"
    echo ""
    echo "Cluster info:"
    kubectl cluster-info
    echo ""
    echo "Nodes:"
    kubectl get nodes
    echo ""
    echo "Test namespace status:"
    kubectl get namespace kdebug-test 2>/dev/null || echo "kdebug-test namespace not found (this is normal if test workloads haven't been applied yet)"
else
    echo -e "${RED}❌ Failed to connect to cluster${NC}"
    echo "Please check:"
    echo "1. AWS credentials are configured correctly"
    echo "2. The cluster is running and accessible"
    echo "3. Your IP is allowed in cluster security groups"
    exit 1
fi
