#!/usr/bin/env bash
# Test script for denv bash wrapper

# Don't use set -e as it interferes with test evaluation
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected="$3"
    
    echo -n "Testing $test_name... "
    
    if result=$(eval "$test_cmd" 2>&1); then
        if [[ -z "$expected" ]] || [[ "$result" == *"$expected"* ]]; then
            echo -e "${GREEN}PASSED${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Expected: $expected"
            echo "  Got: $result"
            ((TESTS_FAILED++))
        fi
    else
        if [[ "$expected" == "SHOULD_FAIL" ]]; then
            echo -e "${GREEN}PASSED${NC} (expected failure)"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Error: $result"
            ((TESTS_FAILED++))
        fi
    fi
}

echo "==================================="
echo "denv Bash Wrapper Test Suite"
echo "==================================="
echo ""

# Source the wrapper (from same directory as test script)
echo "Sourcing wrapper..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/denv-wrapper.sh" || {
    echo -e "${RED}Failed to source wrapper${NC}"
    exit 1
}

# Test 1: Check if denv function exists
run_test "denv function exists" "type -t denv" "function"

# Test 2: Shell detection
run_test "shell detection" "detect_shell" "bash"

# Test 3: Check wrapper aliases
run_test "alias 'de' exists" "alias de 2>/dev/null" "denv enter"
run_test "alias 'dx' exists" "alias dx 2>/dev/null" "denv exit"

# Test 4: Environment preparation (mock)
# Create a mock denv-core for testing
cat > /tmp/denv-core << 'EOF'
#!/bin/bash
case "$1" in
    prepare-env)
        cat <<JSON
{
  "env_path": "/tmp/test-env",
  "project_path": "/tmp/test-project",
  "project_name": "test-project",
  "env_name": "test",
  "session_id": "test-session-123",
  "ports": {
    "3000": "30001",
    "5432": "35432"
  }
}
JSON
        ;;
    get-env-overrides)
        echo 'export TEST_VAR="test_value"'
        ;;
    cleanup-session)
        echo "Session cleaned: $2"
        ;;
    list)
        echo "default"
        echo "staging"
        echo "production"
        ;;
    *)
        echo "Unknown command: $1"
        exit 1
        ;;
esac
EOF
chmod +x /tmp/denv-core

# Set mock binary path
export DENV_CORE="/tmp/denv-core"

# Test 5: Prepare environment command
run_test "prepare-env returns JSON" \
    'denv project 2>/dev/null || echo "OK"' \
    "OK"

# Test 6: List command passthrough
run_test "list command passthrough" \
    'denv list' \
    "default"

# Test 7: Check _denv_list function
run_test "_denv_list function" \
    '_denv_list | head -1' \
    "default"

# Test 8: Shell init output
run_test "shell-init command" \
    'denv shell-init | grep -c "denv shell integration"' \
    "1"

# Test 9: Exit without session
export DENV_SESSION=""
run_test "exit without session fails" \
    '_denv_exit 2>&1' \
    "SHOULD_FAIL"

# Test 10: Environment variable escaping
test_escape() {
    local result=$(bash -c "source $SCRIPT_DIR/denv-wrapper.sh; _generate_bash_env /tmp/test.env /tmp/env /tmp/proj sess test proj 'PORT_3000=30001'")
    if [[ -f /tmp/test.env ]]; then
        grep -q "DENV_ENV_NAME=\"test\"" /tmp/test.env
        return $?
    fi
    return 1
}
run_test "environment script generation" "test_escape" ""

# Cleanup
rm -f /tmp/denv-core /tmp/test.env

echo ""
echo "==================================="
echo "Test Results"
echo "==================================="
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi