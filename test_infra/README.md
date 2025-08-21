# kdebug Test Infrastructure

This directory contains Terraform configuration to deploy a complete AWS EKS cluster with various test workloads designed to validate the `kdebug pod` diagnostics functionality.

## 🎯 Overview

The test infrastructure creates:
- **AWS EKS Cluster** with multiple node groups
- **Test Workloads** covering all kdebug pod diagnostic scenarios  
- **RBAC Configuration** for permission testing
- **Node Groups** with different taints and labels
- **Automated Test Scripts** for validation

## 📋 Prerequisites

### Required Tools
```bash
# Terraform (>= 1.0)
terraform --version

# AWS CLI v2
aws --version

# kubectl
kubectl version --client

# Go (for building kdebug)
go version
```

### AWS Configuration
```bash
# Configure AWS credentials
aws configure

# Or use AWS SSO
aws configure sso

# Verify access
aws sts get-caller-identity
```

## 🚀 Quick Start

### 1. Deploy Infrastructure
```bash
# Navigate to test infrastructure directory
cd test_infra

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy (takes ~15-20 minutes)
terraform apply
```

### 2. Configure kubectl
```bash
# Connect to the EKS cluster
./connect-eks.sh

# Verify connection
kubectl cluster-info
kubectl get nodes
```

### 3. Build and Test kdebug
```bash
# Build kdebug binary
cd ..
go build -o kdebug

# Return to test directory
cd test_infra

# Run comprehensive test suite
./test-kdebug.sh
```

## 🧪 Test Scenarios

The infrastructure deploys the following test pods to validate kdebug diagnostics:

### Pod Scenarios

| Pod Name | Scenario | Expected kdebug Result |
|----------|----------|----------------------|
| `healthy-test-pod` | ✅ Healthy running pod | PASSED |
| `image-pull-error-pod` | ❌ Invalid image registry | FAILED - Image pull issues |
| `crash-loop-pod` | ❌ Container exits with error | FAILED - CrashLoopBackOff |
| `unschedulable-pod` | ❌ Resource requests too high | FAILED - Scheduling issues |
| `init-failure-pod` | ❌ Init container fails | FAILED - Init container issues |
| `rbac-issue-pod` | ❌ Insufficient RBAC permissions | FAILED - Permission denied |
| `oom-test-pod` | ❌ Memory limit exceeded | FAILED - Out of memory |
| `dependency-test-pod` | ❌ Service dependencies missing | FAILED - DNS/connectivity |
| `best-effort-pod` | ⚠️ No resource requests/limits | WARNING - BestEffort QoS |
| `multi-pod-deployment` | ✅ Multiple healthy pods | PASSED - Deployment analysis |

### Node Groups

| Node Group | Purpose | Configuration |
|------------|---------|---------------|
| `kdebug-general-nodes` | Tainted nodes for scheduling tests | t3.small, taint: `kdebug-test=true:NoSchedule` |
| `kdebug-workload-nodes` | Clean nodes for workloads | t3.small, no taints |

## 🔧 Configuration Options

### Variables

Customize the deployment by modifying `terraform.tfvars`:

```hcl
# terraform.tfvars
region = "us-west-2"
cluster_version = "1.29"
node_instance_type = "t3.medium"
enable_test_workloads = true
```

### Available Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `region` | AWS region | `us-east-2` |
| `cluster_version` | Kubernetes version | `1.29` |
| `node_instance_type` | EC2 instance type | `t3.small` |
| `enable_test_workloads` | Deploy test pods | `true` |

## 🧪 Testing kdebug

### Basic Usage
```bash
# Test specific pod
kdebug pod healthy-test-pod --namespace kdebug-test

# Test all pods in namespace
kdebug pod --all --namespace kdebug-test

# Test with JSON output
kdebug pod image-pull-error-pod --namespace kdebug-test --output json

# Test with log analysis
kdebug pod crash-loop-pod --namespace kdebug-test --include-logs --log-lines 20
```

### Advanced Testing
```bash
# Focus on specific diagnostic areas
kdebug pod rbac-issue-pod --namespace kdebug-test --checks=rbac,basic

# Watch mode for real-time monitoring
kdebug pod healthy-test-pod --namespace kdebug-test --watch

# Verbose output for debugging
kdebug pod unschedulable-pod --namespace kdebug-test --verbose
```

### Automated Test Suite
```bash
# Run comprehensive test suite
./test-kdebug.sh

# The script tests:
# - All pod scenarios
# - Output formats (table, JSON, YAML)
# - Specific check types
# - Error handling
# - Advanced features
```

## 📊 Expected Results

### Healthy Pod Example
```bash
$ kdebug pod healthy-test-pod --namespace kdebug-test

✅ Pod Status: PASSED - Pod is running with 1/1 containers ready
✅ Pod Scheduling: PASSED - Pod is scheduled on node 'ip-10-0-1-xxx'
✅ Container nginx - Image Pull: PASSED - Image pulled successfully
✅ RBAC - Service Account: PASSED - Using default service account
✅ Resource QoS: PASSED - Pod has Burstable QoS class
✅ Network - Pod IP: PASSED - Pod IP assigned: 10.0.1.123
```

### Failed Pod Example
```bash
$ kdebug pod image-pull-error-pod --namespace kdebug-test

❌ Pod Status: FAILED - Pod is stuck in Pending state
❌ Container failing-container - Image Pull: FAILED - Failed to pull image
   ↳ Suggestion: Image not found - verify image name and tag
✅ RBAC - Service Account: PASSED - Using default service account
```

## 🗂️ File Structure

```
test_infra/
├── README.md                 # This documentation
├── main.tf                   # Core EKS cluster configuration
├── test-workloads.tf         # Test pod definitions
├── variables.tf              # Input variables
├── outputs.tf                # Output values
├── terraform.tf              # Terraform settings
├── connect-eks.sh            # kubectl configuration script
├── test-kdebug.sh           # Automated test suite
└── .gitignore                # Git ignore patterns
```

## 💰 Cost Considerations

### Estimated Monthly Costs (us-east-2)
- **EKS Cluster**: ~$73/month
- **t3.small instances (5 nodes)**: ~$75/month  
- **EBS volumes**: ~$10/month
- **NAT Gateway**: ~$45/month
- **Total**: ~$203/month

### Cost Optimization Tips
```bash
# Use t3.micro for development (not recommended for production testing)
node_instance_type = "t3.micro"

# Reduce node count
# Edit main.tf node group desired_size values

# Use Spot instances (requires additional configuration)
# Add spot configuration to node groups
```

## 🧹 Cleanup

### Destroy Infrastructure
```bash
# Destroy all resources
terraform destroy

# Confirm destruction
# Type: yes
```

### Cleanup kubectl Context
```bash
# Remove cluster from kubectl config
kubectl config delete-context arn:aws:eks:region:account:cluster/cluster-name

# Clean up local files
rm -f kubeconfig_*
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Terraform Apply Fails
```bash
# Check AWS permissions
aws sts get-caller-identity

# Verify region availability
aws ec2 describe-regions

# Check service quotas
aws service-quotas get-service-quota --service-code eks --quota-code L-1194D53C
```

#### 2. kubectl Connection Issues
```bash
# Re-run connection script
./connect-eks.sh

# Manual configuration
aws eks --region $(terraform output -raw region) update-kubeconfig --name $(terraform output -raw cluster_name)

# Test connection
kubectl cluster-info
```

#### 3. Test Pods Not Creating
```bash
# Check namespace
kubectl get namespace kdebug-test

# Check pod status
kubectl get pods -n kdebug-test

# Check events
kubectl get events -n kdebug-test --sort-by='.lastTimestamp'
```

#### 4. Node Group Issues
```bash
# Check node status
kubectl get nodes

# Check node group status in AWS console
aws eks describe-nodegroup --cluster-name $(terraform output -raw cluster_name) --nodegroup-name kdebug-general-nodes

# Check autoscaling
kubectl get pods -n kube-system | grep cluster-autoscaler
```

## 🔒 Security Notes

### IAM Permissions
The infrastructure creates minimal required permissions:
- EKS cluster management
- EC2 instance management
- VPC networking
- IAM role assumptions

### Network Security
- Private subnets for worker nodes
- Public subnets for load balancers only
- Security groups with minimal required access
- NAT Gateway for outbound internet access

### RBAC Testing
- Limited service account for RBAC testing
- Intentionally restricted permissions
- Tests permission denied scenarios

## 📚 Additional Resources

### Documentation
- [AWS EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [kubectl Reference](https://kubernetes.io/docs/reference/kubectl/)

### Related Tools
- [eksctl](https://eksctl.io/) - Alternative EKS cluster management
- [k9s](https://k9scli.io/) - Terminal UI for Kubernetes
- [stern](https://github.com/stern/stern) - Multi-pod log tailing

## 🤝 Contributing

To add new test scenarios:

1. Add pod definitions to `test-workloads.tf`
2. Update `outputs.tf` with new pod names
3. Add test cases to `test-kdebug.sh`
4. Update this README with scenario descriptions

## 📄 License

This test infrastructure follows the same license as the main kdebug project.
