# Example Terraform variables for kdebug test infrastructure
# Copy this file to terraform.tfvars and customize for your environment

# AWS Region (choose one close to you for better performance)
region = "us-east-2"  # Ohio
# region = "us-west-2"  # Oregon
# region = "eu-west-1"  # Ireland
# region = "ap-southeast-1"  # Singapore

# Kubernetes cluster version (use stable versions)
cluster_version = "1.29"
# cluster_version = "1.28"
# cluster_version = "1.30"

# Node instance type (affects cost and performance)
node_instance_type = "t3.small"   # Recommended for testing
# node_instance_type = "t3.micro"   # Cheapest option (may be unstable)
# node_instance_type = "t3.medium"  # Better performance
# node_instance_type = "t3.large"   # Production-like testing

# Enable test workloads (set to false to deploy only the cluster)
enable_test_workloads = true

# Additional customization examples:
# 
# For cost optimization:
# node_instance_type = "t3.micro"
# 
# For performance testing:
# node_instance_type = "t3.medium"
# 
# For production-like environment:
# node_instance_type = "t3.large"
# cluster_version = "1.28"  # Use stable LTS version
