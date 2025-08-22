#!/bin/bash

# Automated deployment script for kdebug test infrastructure
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 kdebug Test Infrastructure Deployment${NC}"
echo -e "${BLUE}=======================================${NC}"

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"
    
    # Check terraform
    if ! command -v terraform &> /dev/null; then
        echo -e "${RED}❌ Terraform not found. Please install Terraform >= 1.3${NC}"
        exit 1
    fi
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        echo -e "${RED}❌ AWS CLI not found. Please install AWS CLI v2${NC}"
        exit 1
    fi
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}❌ kubectl not found. Please install kubectl${NC}"
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        echo -e "${RED}❌ AWS credentials not configured. Please run 'aws configure'${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All prerequisites satisfied${NC}"
}

# Function to setup terraform variables
setup_variables() {
    echo -e "${YELLOW}📝 Setting up Terraform variables...${NC}"
    
    if [ ! -f "terraform.tfvars" ]; then
        echo "Creating terraform.tfvars from example..."
        cp terraform.tfvars.example terraform.tfvars
        echo -e "${YELLOW}⚠️  Please review and customize terraform.tfvars if needed${NC}"
        echo "Default configuration:"
        echo "  Region: us-east-2"
        echo "  Instance Type: t3.small"
        echo "  Cluster Version: 1.29"
        echo ""
        read -p "Continue with default configuration? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Please edit terraform.tfvars and run this script again"
            exit 1
        fi
    else
        echo -e "${GREEN}✅ terraform.tfvars already exists${NC}"
    fi
}

# Function to deploy infrastructure
deploy_infrastructure() {
    echo -e "${YELLOW}🏗️  Deploying infrastructure...${NC}"
    
    # Initialize terraform
    echo "Initializing Terraform..."
    terraform init
    
    # Plan deployment
    echo "Creating deployment plan..."
    terraform plan -out=tfplan
    
    # Show cost estimate reminder
    echo -e "${YELLOW}💰 Cost Reminder:${NC}"
    echo "This deployment will create:"
    echo "  • 1 EKS cluster (~$73/month)"
    echo "  • 5 t3.small nodes (~$75/month)"
    echo "  • 1 NAT Gateway (~$45/month)"
    echo "  • Estimated total: ~$200/month"
    echo ""
    read -p "Continue with deployment? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deployment cancelled"
        exit 0
    fi
    
    # Apply deployment
    echo "Applying deployment (this will take 15-20 minutes)..."
    terraform apply tfplan
    
    # Clean up plan file
    rm -f tfplan
    
    echo -e "${GREEN}✅ Infrastructure deployed successfully!${NC}"
}

# Function to configure kubectl
configure_kubectl() {
    echo -e "${YELLOW}⚙️  Configuring kubectl...${NC}"
    
    ./connect-eks.sh
    
    echo -e "${GREEN}✅ kubectl configured successfully!${NC}"
}

# Function to wait for cluster readiness
wait_for_cluster() {
    echo -e "${YELLOW}⏳ Waiting for cluster to be fully ready...${NC}"
    
    # Wait for nodes to be ready
    echo "Waiting for nodes to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=600s
    
    echo -e "${GREEN}✅ Cluster is ready!${NC}"
}

# Function to deploy test pods
deploy_test_pods() {
    echo -e "${YELLOW}🧪 Deploying test pods...${NC}"
    
    read -p "Deploy test pods for kdebug testing? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        ./deploy-test-pods.sh
    else
        echo "Skipping test pod deployment. You can deploy them later with: ./deploy-test-pods.sh"
    fi
}

# Function to show deployment summary
show_summary() {
    echo -e "${BLUE}📊 Deployment Summary${NC}"
    echo -e "${BLUE}===================${NC}"
    
    # Cluster information
    echo -e "${GREEN}Cluster Information:${NC}"
    terraform output
    echo ""
    
    # Node status
    echo -e "${GREEN}Node Status:${NC}"
    kubectl get nodes
    echo ""
    
    # Test pods status (if namespace exists)
    if kubectl get namespace kdebug-test >/dev/null 2>&1; then
        echo -e "${GREEN}Test Pods Status:${NC}"
        kubectl get pods -n kdebug-test
        echo ""
    else
        echo -e "${YELLOW}Test pods not deployed yet. Run: ./deploy-test-pods.sh${NC}"
        echo ""
    fi
    
    # Next steps
    echo -e "${YELLOW}🎯 Next Steps:${NC}"
    echo "1. Build kdebug: cd .. && go build -o kdebug"
    echo "2. Run tests: cd test_infra && ./test-kdebug.sh"
    echo "3. Test manually: kdebug pod <pod-name> --namespace kdebug-test"
    echo ""
    echo -e "${YELLOW}📚 Useful Commands:${NC}"
    echo "• View all test pods: kubectl get pods -n kdebug-test"
    echo "• Check pod events: kubectl get events -n kdebug-test --sort-by='.lastTimestamp'"
    echo "• Test kdebug: ./test-kdebug.sh"
    echo "• Cleanup: terraform destroy"
    echo ""
    echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
}

# Main deployment flow
main() {
    check_prerequisites
    setup_variables
    deploy_infrastructure
    configure_kubectl
    wait_for_cluster
    deploy_test_pods
    show_summary
}

# Handle script interruption
trap 'echo -e "\n${RED}Deployment interrupted${NC}"; exit 1' INT

# Run main function
main
