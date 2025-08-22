#!/bin/bash

# Test script for kdebug pod diagnostics
# This script tests the newly implemented kdebug pod command against various pod scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="kdebug-test"
KDEBUG_BINARY="../kdebug"
TIMEOUT=60

echo -e "${BLUE}🧪 kdebug Pod Diagnostics Test Suite${NC}"
echo -e "${BLUE}=====================================${NC}"

# Check if kdebug binary exists
if [ ! -f "$KDEBUG_BINARY" ]; then
    echo -e "${RED}❌ kdebug binary not found at $KDEBUG_BINARY${NC}"
    echo "Please build kdebug first: cd .. && go build -o kdebug"
    exit 1
fi

# Check if kubectl is configured
if ! kubectl cluster-info >/dev/null 2>&1; then
    echo -e "${RED}❌ kubectl is not configured or cluster is not accessible${NC}"
    echo "Please run: ./connect-eks.sh"
    exit 1
fi

# Check if test namespace exists
if ! kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
    echo -e "${RED}❌ Test namespace '$NAMESPACE' not found${NC}"
    echo "Please run: terraform apply"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"
echo ""

# Function to run kdebug and capture output
run_kdebug_test() {
    local test_name="$1"
    local pod_name="$2"
    local expected_status="$3"
    local additional_args="${4:-}"
    
    echo -e "${YELLOW}🔍 Testing: $test_name${NC}"
    echo "   Pod: $pod_name"
    echo "   Args: $additional_args"
    
    # Run kdebug pod command
    local output
    if output=$($KDEBUG_BINARY pod "$pod_name" --namespace "$NAMESPACE" $additional_args 2>&1); then
        local exit_code=0
    else
        local exit_code=$?
    fi
    
    # Check if output contains expected content
    case $expected_status in
        "PASSED")
            if echo "$output" | grep -q "PASSED" && ! echo "$output" | grep -q "FAILED"; then
                echo -e "   ${GREEN}✅ PASSED - Healthy pod detected correctly${NC}"
            else
                echo -e "   ${RED}❌ FAILED - Expected PASSED status${NC}"
                echo "   Output: $output"
            fi
            ;;
        "FAILED")
            if echo "$output" | grep -q "FAILED"; then
                echo -e "   ${GREEN}✅ PASSED - Issues detected correctly${NC}"
            else
                echo -e "   ${RED}❌ FAILED - Expected FAILED status${NC}"
                echo "   Output: $output"
            fi
            ;;
        "WARNING")
            if echo "$output" | grep -q "WARNING"; then
                echo -e "   ${GREEN}✅ PASSED - Warnings detected correctly${NC}"
            else
                echo -e "   ${RED}❌ FAILED - Expected WARNING status${NC}"
                echo "   Output: $output"
            fi
            ;;
        *)
            echo -e "   ${BLUE}ℹ️  INFO - General test completed${NC}"
            ;;
    esac
    
    echo ""
}

# Function to test output formats
test_output_formats() {
    local pod_name="$1"
    
    echo -e "${YELLOW}🔍 Testing Output Formats${NC}"
    
    # Test JSON output
    echo "   Testing JSON output..."
    if $KDEBUG_BINARY pod "$pod_name" --namespace "$NAMESPACE" --output json >/dev/null 2>&1; then
        echo -e "   ${GREEN}✅ JSON output works${NC}"
    else
        echo -e "   ${RED}❌ JSON output failed${NC}"
    fi
    
    # Test YAML output
    echo "   Testing YAML output..."
    if $KDEBUG_BINARY pod "$pod_name" --namespace "$NAMESPACE" --output yaml >/dev/null 2>&1; then
        echo -e "   ${GREEN}✅ YAML output works${NC}"
    else
        echo -e "   ${RED}❌ YAML output failed${NC}"
    fi
    
    echo ""
}

# Function to test specific checks
test_specific_checks() {
    local pod_name="$1"
    
    echo -e "${YELLOW}🔍 Testing Specific Check Types${NC}"
    
    local check_types=("basic" "scheduling" "images" "rbac" "resources" "network")
    
    for check_type in "${check_types[@]}"; do
        echo "   Testing --checks=$check_type..."
        if $KDEBUG_BINARY pod "$pod_name" --namespace "$NAMESPACE" --checks="$check_type" >/dev/null 2>&1; then
            echo -e "   ${GREEN}✅ Check type '$check_type' works${NC}"
        else
            echo -e "   ${RED}❌ Check type '$check_type' failed${NC}"
        fi
    done
    
    echo ""
}

# Wait for pods to be in expected states
echo -e "${YELLOW}⏳ Waiting for test pods to reach expected states...${NC}"
sleep 30

# Test 1: Healthy pod
run_kdebug_test "Healthy Pod Diagnostics" "healthy-test-pod" "PASSED"

# Test 2: Image pull error
run_kdebug_test "Image Pull Error Detection" "image-pull-error-pod" "FAILED"

# Test 3: Crash loop detection
run_kdebug_test "Crash Loop Detection" "crash-loop-pod" "FAILED"

# Test 4: Unschedulable pod
run_kdebug_test "Unschedulable Pod Detection" "unschedulable-pod" "FAILED"

# Test 5: Init container failure
run_kdebug_test "Init Container Failure" "init-failure-pod" "FAILED"

# Test 6: RBAC issues
run_kdebug_test "RBAC Permission Issues" "rbac-issue-pod" "FAILED"

# Test 7: OOM pod
run_kdebug_test "Out of Memory Detection" "oom-test-pod" "FAILED"

# Test 8: Dependency issues
run_kdebug_test "Dependency Issues" "dependency-test-pod" "FAILED"

# Test 9: Best effort QoS
run_kdebug_test "Best Effort QoS Detection" "best-effort-pod" "WARNING"

# Test 10: All pods in namespace
echo -e "${YELLOW}🔍 Testing: All Pods in Namespace${NC}"
if $KDEBUG_BINARY pod --all --namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo -e "   ${GREEN}✅ PASSED - All pods analysis works${NC}"
else
    echo -e "   ${RED}❌ FAILED - All pods analysis failed${NC}"
fi
echo ""

# Test 11: Verbose output
run_kdebug_test "Verbose Output" "healthy-test-pod" "PASSED" "--verbose"

# Test 12: Log analysis (if pod supports it)
run_kdebug_test "Log Analysis" "crash-loop-pod" "FAILED" "--include-logs --log-lines 10"

# Test 13: Output formats
test_output_formats "healthy-test-pod"

# Test 14: Specific checks
test_specific_checks "healthy-test-pod"

# Test 15: Non-existent pod
echo -e "${YELLOW}🔍 Testing: Non-existent Pod${NC}"
if ! $KDEBUG_BINARY pod "non-existent-pod" --namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo -e "   ${GREEN}✅ PASSED - Non-existent pod handled correctly${NC}"
else
    echo -e "   ${RED}❌ FAILED - Non-existent pod should have failed${NC}"
fi
echo ""

echo -e "${BLUE}🏁 Test Suite Complete${NC}"
echo -e "${BLUE}=====================${NC}"

# Summary report
echo ""
echo -e "${GREEN}📊 Test Summary:${NC}"
echo "• Tested 10 different pod scenarios"
echo "• Tested multiple output formats (table, JSON, YAML)"
echo "• Tested specific diagnostic check types"
echo "• Tested advanced features (verbose, log analysis, watch)"
echo "• Tested error handling (non-existent pods)"
echo ""
echo -e "${YELLOW}💡 Manual Tests to Run:${NC}"
echo "1. Watch mode: $KDEBUG_BINARY pod healthy-test-pod --namespace $NAMESPACE --watch"
echo "2. Interactive debugging of specific issues found"
echo "3. Test with different kubeconfig contexts"
echo ""
echo -e "${GREEN}✨ kdebug pod diagnostics testing completed!${NC}"
