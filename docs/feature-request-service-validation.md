---
name: "🚀 New Feature Request"
description: "Service Health and Endpoint Validation for kdebug"
title: "[Feature]: Service Health and Endpoint Validation"
labels: ["feature-request", "needs-triage"]
assignees: []
---

# New Feature Request 🚀

**Feature Name:** Service Health and Endpoint Validation

**Feature Description:**
Implement comprehensive service diagnostics that analyze Kubernetes Service objects and their associated endpoints. This feature will help users quickly identify why services are not working correctly by checking service configuration, endpoint health, DNS resolution, and connectivity issues.

**Problem Statement:**
When applications fail to communicate in Kubernetes clusters, service-related issues are often the culprit. Currently, users must manually:
- Check if services exist and are properly configured
- Verify that services have healthy endpoints
- Investigate selector mismatches between services and pods
- Diagnose DNS resolution problems for service names
- Test actual connectivity to service endpoints

This manual process is time-consuming and error-prone, especially for Kubernetes beginners who may not know all the places to look.

**Use Cases and Scenarios:**
1. **As a DevOps engineer**, I want to quickly diagnose why my application can't reach a database service, so that I can resolve connectivity issues faster.
2. **When debugging microservice communication**, this feature would help by automatically checking if services are properly exposing pods and if endpoints are healthy.
3. **In production environments**, this would be useful for rapid troubleshooting of service discovery and load balancing issues.
4. **For Kubernetes beginners**, this provides guided diagnosis of service networking problems with clear explanations.

**Proposed Solution:**
Add a new `kdebug service` command that provides comprehensive service diagnostics:

```bash
# Check specific service
kdebug service <service-name> --namespace <namespace>

# Check all services in namespace
kdebug service --all --namespace <namespace>

# Include DNS resolution testing
kdebug service <service-name> --test-dns

# Check service across all namespaces
kdebug service --all-namespaces
```

The feature will analyze:
- Service existence and basic configuration
- Selector matching with available pods
- Endpoint health and readiness
- Port configuration and protocols
- DNS resolution within the cluster
- Service type-specific checks (ClusterIP, NodePort, LoadBalancer)

**Acceptance Criteria:**
- [ ] Feature can detect common service misconfigurations
- [ ] Users can run `kdebug service <name>` to get comprehensive service health report
- [ ] Output includes service status, endpoint health, and connectivity validation
- [ ] Works with all service types (ClusterIP, NodePort, LoadBalancer, ExternalName)
- [ ] Handles error cases gracefully (service not found, no endpoints, etc.)
- [ ] Provides actionable suggestions for fixing identified issues
- [ ] Integrates with existing kdebug output formats (table, JSON, YAML)

**Command Examples:**
```bash
# Diagnose a specific service
kdebug service frontend --namespace production

# Expected output:
✅ PASSED: Service 'frontend' exists and is accessible
✅ PASSED: Service selector matches 3 pods
❌ FAILED: Endpoint health check
   📍 Issue: 2/3 endpoints are not ready
   📄 Details: Pod frontend-7d4b8c6f9-abc12 is in CrashLoopBackOff
              Pod frontend-7d4b8c6f9-def34 is not ready (readiness probe failed)
   💡 Suggestions:
      1. Check pod logs: kubectl logs frontend-7d4b8c6f9-abc12
      2. Verify readiness probe configuration
      3. Check application startup process

✅ PASSED: DNS resolution test
✅ PASSED: Port configuration is valid

# Check all services in namespace
kdebug service --all --namespace kube-system

# Test DNS resolution
kdebug service api-gateway --test-dns
```

**Kubernetes Resources and Scope:**
This feature will analyze:
- Service objects (all types: ClusterIP, NodePort, LoadBalancer, ExternalName)
- Endpoints and EndpointSlices
- Related pods (via service selectors)
- DNS resolution (CoreDNS integration)
- Network connectivity within cluster
- Service annotations and labels

**Feature Category:** Networking & Connectivity (services, ingress, DNS)

**Target Users:** All kdebug users, especially:
- Application developers debugging service connectivity
- DevOps engineers troubleshooting networking issues
- Kubernetes beginners learning service networking

**Priority/Urgency:** High - would significantly improve our workflow

**Alternatives and Workarounds:**
Current manual processes:
- `kubectl get svc` - shows basic service info
- `kubectl describe svc` - shows detailed service configuration
- `kubectl get endpoints` - shows endpoint health
- `kubectl run debug-pod --rm -it --image=busybox -- nslookup service-name` - tests DNS
- Manual connectivity testing with curl/wget from debug pods

These tools require multiple commands and deep Kubernetes knowledge to piece together the full picture.

**Additional Context:**
- Should integrate with existing kdebug architecture patterns
- Leverage kubernetes client-go for API interactions
- Follow same output formatting as pod diagnostics
- Consider performance impact for large numbers of services
- Plan for future integration with Ingress diagnostics

**Implementation Interest:** Yes, I would like to design and implement it

**Implementation Considerations:**
- Use kubernetes client-go for Service and Endpoints API calls
- Implement service selector validation by querying pods with matching labels
- Add DNS resolution testing using ephemeral pods (similar to cluster DNS checks)
- Structure code following existing patterns in pkg/pod/ directory
- Add proper error handling for missing services, network timeouts
- Consider caching for performance when checking multiple services
- Implement comprehensive unit tests with fake clientsets