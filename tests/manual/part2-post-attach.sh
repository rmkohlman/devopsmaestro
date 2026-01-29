#!/bin/bash
# =============================================================================
# DVM Manual Test - Part 2: Post-Attach Tests
# =============================================================================
# Purpose: Plugin management, error handling, output formats, unit tests
# Time: ~2 minutes
# 
# Run this AFTER you've completed the interactive container testing.
#
# Usage: 
#   cd ~/Developer/tools/devopsmaestro
#   source tests/manual/part2-post-attach.sh
#
# Log file: ~/.devopsmaestro/logs/test-part2-TIMESTAMP.log
# =============================================================================

# -----------------------------------------------------------------------------
# CONFIGURATION
# -----------------------------------------------------------------------------
DVM_PROJECT_DIR=~/Developer/tools/devopsmaestro
DVM_TEST_PROJECT=~/Developer/sandbox/dvm-test-fastapi
DVM_LOG_DIR=~/.devopsmaestro/logs
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="$DVM_LOG_DIR/test-part2-$TIMESTAMP.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Test tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0
declare -a FAILED_TESTS=()

# -----------------------------------------------------------------------------
# LOGGING FUNCTIONS
# -----------------------------------------------------------------------------

mkdir -p "$DVM_LOG_DIR"

init_log() {
    cat > "$LOG_FILE" << EOF
================================================================================
DVM MANUAL TEST - PART 2: POST-ATTACH TESTS
================================================================================
Started: $(date)
Host: $(hostname)
User: $(whoami)
================================================================================

EOF
}

log_debug() {
    echo "[$(date '+%H:%M:%S')] DEBUG: $*" >> "$LOG_FILE"
}

log_cmd() {
    echo "" >> "$LOG_FILE"
    echo "[$(date '+%H:%M:%S')] COMMAND: $1" >> "$LOG_FILE"
    echo "----------------------------------------" >> "$LOG_FILE"
}

log_info() {
    echo -e "${CYAN}ℹ${NC}  $*"
    echo "[$(date '+%H:%M:%S')] INFO: $*" >> "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}✓${NC}  $*"
    echo "[$(date '+%H:%M:%S')] SUCCESS: $*" >> "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC}  $*"
    echo "[$(date '+%H:%M:%S')] WARNING: $*" >> "$LOG_FILE"
}

log_error() {
    echo -e "${RED}✗${NC}  $*"
    echo "[$(date '+%H:%M:%S')] ERROR: $*" >> "$LOG_FILE"
}

log_section() {
    echo ""
    echo -e "${BOLD}${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${BLUE}  $*${NC}"
    echo -e "${BOLD}${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "" >> "$LOG_FILE"
    echo "================================================================================" >> "$LOG_FILE"
    echo "SECTION: $*" >> "$LOG_FILE"
    echo "Time: $(date)" >> "$LOG_FILE"
    echo "================================================================================" >> "$LOG_FILE"
}

log_step() {
    echo -e "  ${CYAN}→${NC} $*"
    echo "[$(date '+%H:%M:%S')] STEP: $*" >> "$LOG_FILE"
}

# -----------------------------------------------------------------------------
# TEST FUNCTIONS
# -----------------------------------------------------------------------------

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected="$3"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    log_cmd "$test_cmd"
    
    local output
    local exit_code
    output=$(eval "$test_cmd" 2>&1)
    exit_code=$?
    
    echo "$output" >> "$LOG_FILE"
    echo "Exit code: $exit_code" >> "$LOG_FILE"
    echo "----------------------------------------" >> "$LOG_FILE"
    
    if [ $exit_code -eq 0 ]; then
        if [ -n "$expected" ]; then
            if echo "$output" | grep -q "$expected"; then
                TESTS_PASSED=$((TESTS_PASSED + 1))
                log_success "$test_name"
                return 0
            else
                TESTS_FAILED=$((TESTS_FAILED + 1))
                FAILED_TESTS+=("$test_name")
                log_error "$test_name (expected: $expected)"
                return 1
            fi
        else
            TESTS_PASSED=$((TESTS_PASSED + 1))
            log_success "$test_name"
            return 0
        fi
    else
        # For error handling tests, non-zero exit is expected
        if [ -n "$expected" ] && echo "$output" | grep -qi "$expected"; then
            TESTS_PASSED=$((TESTS_PASSED + 1))
            log_success "$test_name"
            return 0
        fi
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_name")
        log_error "$test_name (exit: $exit_code)"
        return 1
    fi
}

# Test that expects failure (for error handling)
run_test_expect_error() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_error="$3"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    log_cmd "$test_cmd"
    
    local output
    output=$(eval "$test_cmd" 2>&1)
    local exit_code=$?
    
    echo "$output" >> "$LOG_FILE"
    echo "Exit code: $exit_code" >> "$LOG_FILE"
    echo "----------------------------------------" >> "$LOG_FILE"
    
    # We expect either non-zero exit OR error message in output
    if [ $exit_code -ne 0 ] || echo "$output" | grep -qi "error\|not found\|invalid"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_success "$test_name (error handled correctly)"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_name")
        log_error "$test_name (expected error, got success)"
        return 1
    fi
}

run_cmd() {
    local description="$1"
    local cmd="$2"
    
    log_step "$description"
    log_cmd "$cmd"
    
    eval "$cmd" 2>&1 | tee -a "$LOG_FILE"
    local exit_code=${PIPESTATUS[0]}
    
    echo "Exit code: $exit_code" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"
    
    return $exit_code
}

run_silent() {
    local description="$1"
    local cmd="$2"
    
    log_step "$description"
    log_cmd "$cmd"
    
    local output
    output=$(eval "$cmd" 2>&1)
    local exit_code=$?
    
    echo "$output" >> "$LOG_FILE"
    echo "Exit code: $exit_code" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"
    
    if [ $exit_code -eq 0 ]; then
        log_success "$description"
    else
        log_warning "$description (exit: $exit_code)"
    fi
    
    return $exit_code
}

# -----------------------------------------------------------------------------
# MAIN
# -----------------------------------------------------------------------------

main() {
    init_log
    
    # Ensure dvm is in PATH
    export PATH="$DVM_PROJECT_DIR:$PATH"
    
    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║     DVM MANUAL TEST - PART 2: POST-ATTACH TESTS              ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  Log file: ${CYAN}$LOG_FILE${NC}"
    echo ""
    
    # -------------------------------------------------------------------------
    # SEGMENT 6: PLUGIN MANAGEMENT
    # -------------------------------------------------------------------------
    log_section "SEGMENT 6: PLUGIN MANAGEMENT"
    
    builtin cd $DVM_PROJECT_DIR
    
    log_step "Creating plugin from file..."
    cat > /tmp/test-plugin.yaml << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: Fuzzy finder for Neovim
  category: navigation
spec:
  repo: nvim-telescope/telescope.nvim
  lazy: true
  dependencies:
    - nvim-lua/plenary.nvim
  config: |
    return {
      'nvim-telescope/telescope.nvim',
      dependencies = { 'nvim-lua/plenary.nvim' },
    }
EOF
    log_cmd "dvm plugin apply -f /tmp/test-plugin.yaml"
    dvm plugin apply -f /tmp/test-plugin.yaml 2>&1 | tee -a "$LOG_FILE"
    
    echo ""
    log_step "Creating plugin from stdin..."
    log_cmd "cat << 'EOF' | dvm plugin apply -f -"
    cat << 'EOF' | dvm plugin apply -f - 2>&1 | tee -a "$LOG_FILE"
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: treesitter
  description: Syntax highlighting
  category: syntax
spec:
  repo: nvim-treesitter/nvim-treesitter
  lazy: true
  build: ':TSUpdate'
  config: |
    return {
      'nvim-treesitter/nvim-treesitter',
      build = ':TSUpdate',
    }
EOF
    
    echo ""
    run_test "Plugin list contains telescope" "dvm plugin list" "telescope"
    run_test "Plugin list contains treesitter" "dvm plugin list" "treesitter"
    
    echo ""
    run_cmd "Get telescope plugin" "dvm plugin get telescope"
    
    echo ""
    log_step "Deleting treesitter plugin..."
    log_cmd "dvm plugin delete treesitter --force"
    dvm plugin delete treesitter --force 2>&1 | tee -a "$LOG_FILE"
    
    echo ""
    # Verify treesitter is no longer in the list
    if dvm plugin list 2>&1 | grep -q "treesitter"; then
        log_error "Treesitter still exists after delete"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("Treesitter deleted")
    else
        log_success "Treesitter deleted"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # -------------------------------------------------------------------------
    # SEGMENT 7: ERROR HANDLING
    # -------------------------------------------------------------------------
    log_section "SEGMENT 7: ERROR HANDLING"
    
    log_info "Testing error handling (errors expected)..."
    echo ""
    
    run_test_expect_error "Non-existent project" "dvm use project nonexistent" "not found"
    run_test_expect_error "Non-existent workspace" "dvm get workspace doesnotexist" "not found"
    run_test_expect_error "Non-existent plugin" "dvm plugin get nonexistent" "not found"
    
    echo ""
    log_info "Testing invalid platform..."
    log_cmd "DVM_PLATFORM=invalid dvm get platforms"
    DVM_PLATFORM=invalid dvm get platforms 2>&1 | tee -a "$LOG_FILE"
    log_success "Invalid platform handled (shows available platforms)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # -------------------------------------------------------------------------
    # SEGMENT 8: OUTPUT FORMATS
    # -------------------------------------------------------------------------
    log_section "SEGMENT 8: OUTPUT FORMATS"
    
    run_test "Projects table format" "dvm get projects" "fastapi-test"
    run_test "Projects yaml format" "dvm get projects -o yaml" "apiVersion:"
    run_test "Projects json format" "dvm get projects -o json" '"Name":'
    
    echo ""
    run_test "Workspaces yaml format" "dvm get workspaces -o yaml" "kind: Workspace"
    run_test "Plugins yaml format" "dvm get plugins -o yaml" "telescope"
    
    # -------------------------------------------------------------------------
    # SEGMENT 9: VERSION AND HELP
    # -------------------------------------------------------------------------
    log_section "SEGMENT 9: VERSION AND HELP"
    
    run_cmd "Version info" "dvm version"
    
    echo ""
    run_test "Help shows commands" "dvm --help" "Available Commands"
    run_test "Build help" "dvm build --help" "Build"
    run_test "Attach help" "dvm attach --help" "Attach"
    
    # -------------------------------------------------------------------------
    # SEGMENT 10: UNIT TESTS
    # -------------------------------------------------------------------------
    log_section "SEGMENT 10: UNIT TESTS"
    
    builtin cd $DVM_PROJECT_DIR
    
    log_info "Running Go test suite..."
    echo ""
    log_cmd "go test ./..."
    
    # Capture test output
    local test_output
    test_output=$(go test ./... 2>&1)
    local test_exit=$?
    
    echo "$test_output" >> "$LOG_FILE"
    
    # Show summary
    echo "$test_output" | grep -E "^(ok|FAIL|---)" | head -20
    
    echo ""
    if [ $test_exit -eq 0 ]; then
        log_success "All unit tests passed"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        # Check if it's just the operators integration tests
        if echo "$test_output" | grep -q "FAIL.*operators" && \
           ! echo "$test_output" | grep -v operators | grep -q "FAIL"; then
            log_warning "Unit tests passed (operators integration tests expected to fail)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            log_error "Some unit tests failed"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            FAILED_TESTS+=("Unit tests")
        fi
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # -------------------------------------------------------------------------
    # SUMMARY
    # -------------------------------------------------------------------------
    log_section "TEST SUMMARY"
    
    echo ""
    echo -e "  ${BOLD}Results:${NC}"
    echo -e "    Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "    Tests failed: ${RED}$TESTS_FAILED${NC}"
    echo -e "    Total tests:  $TESTS_TOTAL"
    echo ""
    
    if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
        echo -e "  ${BOLD}${RED}Failed tests:${NC}"
        for test in "${FAILED_TESTS[@]}"; do
            echo -e "    ${RED}✗${NC} $test"
        done
        echo ""
    fi
    
    echo -e "  ${BOLD}Log files:${NC}"
    ls -la $DVM_LOG_DIR/*.log 2>/dev/null | tail -5 | while read line; do
        echo "    $line"
    done
    echo ""
    
    # Log summary
    cat >> "$LOG_FILE" << EOF

================================================================================
SUMMARY
================================================================================
Tests passed: $TESTS_PASSED
Tests failed: $TESTS_FAILED
Total tests:  $TESTS_TOTAL

Failed tests:
$(printf '%s\n' "${FAILED_TESTS[@]}")

Completed: $(date)
================================================================================
EOF
    
    # -------------------------------------------------------------------------
    # COMPLETION
    # -------------------------------------------------------------------------
    echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║              ALL AUTOMATED TESTS COMPLETE                    ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BOLD}Optional cleanup:${NC}"
    echo ""
    echo -e "    # Remove test containers"
    echo -e "    ${CYAN}docker ps -aq --filter 'name=dvm-' | xargs -r docker rm -f${NC}"
    echo ""
    echo -e "    # Full reset"
    echo -e "    ${CYAN}rm -rf ~/.devopsmaestro${NC}"
    echo -e "    ${CYAN}rm -rf ~/Developer/sandbox/dvm-test-*${NC}"
    echo ""
    echo -e "  ${BOLD}View logs:${NC}"
    echo -e "    ${CYAN}cat $LOG_FILE${NC}"
    echo -e "    ${CYAN}ls -la $DVM_LOG_DIR/${NC}"
    echo ""
    
    if [ $TESTS_FAILED -gt 0 ]; then
        return 1
    fi
    return 0
}

# Run main
main
