#!/bin/bash
#
# test-nvp.sh - Automated test script for nvp CLI
#
# Runs through the NVIMOPS_TEST_PLAN.md automated sections:
#   Part 1: Build & basic commands
#   Part 2: Library operations
#   Part 3: Plugin CRUD operations
#   Part 4: Lua generation
#   Part 6: Error handling & edge cases
#   Part 7: Shell completions
#
# Usage:
#   ./tests/manual/nvp/test-nvp.sh
#
# Options:
#   NVP_KEEP_OUTPUT=1  - Don't clean up test artifacts
#   NVP_VERBOSE=1      - Show all command output
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Test directories
TEST_DIR="${TEST_DIR:-/tmp/nvp-test-$$}"
NVP_BIN="$PROJECT_ROOT/nvp"

# Verbose mode
VERBOSE="${NVP_VERBOSE:-0}"

# Helper functions
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
}

log_section() {
    echo ""
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}============================================${NC}"
}

run_cmd() {
    local desc="$1"
    shift
    local cmd="$@"
    
    log_test "$desc"
    if [[ "$VERBOSE" == "1" ]]; then
        if eval "$cmd"; then
            log_pass "$desc"
            return 0
        else
            log_fail "$desc"
            return 1
        fi
    else
        if eval "$cmd" &>/dev/null; then
            log_pass "$desc"
            return 0
        else
            log_fail "$desc"
            return 1
        fi
    fi
}

expect_fail() {
    local desc="$1"
    shift
    local cmd="$@"
    
    log_test "$desc (expect failure)"
    if eval "$cmd" &>/dev/null; then
        log_fail "$desc - command should have failed"
        return 1
    else
        log_pass "$desc"
        return 0
    fi
}

check_output() {
    local desc="$1"
    local pattern="$2"
    shift 2
    local cmd="$@"
    
    log_test "$desc"
    local output
    output=$(eval "$cmd" 2>&1) || true
    # Use -- to prevent pattern from being interpreted as grep options
    if echo "$output" | grep -qE -- "$pattern"; then
        log_pass "$desc"
        return 0
    else
        log_fail "$desc - pattern '$pattern' not found"
        [[ "$VERBOSE" == "1" ]] && echo "Output: $output"
        return 1
    fi
}

# Setup
setup() {
    log_section "Setup"
    
    mkdir -p "$TEST_DIR"
    export NVP_CONFIG_DIR="$TEST_DIR/nvp"
    
    echo "Test directory: $TEST_DIR"
    echo "NVP config: $NVP_CONFIG_DIR"
}

# Cleanup
cleanup() {
    if [[ "${NVP_KEEP_OUTPUT:-0}" == "1" ]]; then
        echo ""
        echo -e "${GREEN}Test artifacts preserved at: $TEST_DIR${NC}"
    else
        rm -rf "$TEST_DIR"
        rm -f "$PROJECT_ROOT/nvp" 2>/dev/null || true
    fi
}

trap cleanup EXIT

# ============================================
# Part 1: Build and Basic Commands
# ============================================
test_part1() {
    log_section "Part 1: Build and Basic Commands"
    
    cd "$PROJECT_ROOT"
    
    # 1.1 Build
    run_cmd "Build nvp binary" "go build -o nvp ./cmd/nvp/"
    run_cmd "Binary exists" "test -f $NVP_BIN"
    run_cmd "Binary is executable" "test -x $NVP_BIN"
    
    # 1.2 Version
    check_output "Version command" "nvp.*NvimOps" "$NVP_BIN version"
    # Accept "dev" (local build) or version number (released build)
    check_output "Version --short" "^(dev|[0-9])" "$NVP_BIN version --short"
    
    # 1.3 Help
    check_output "Root help" "NvimOps.*DevOps-style" "$NVP_BIN --help"
    check_output "Library help" "library" "$NVP_BIN library --help"
    check_output "Apply help" "-f.*filename" "$NVP_BIN apply --help"
    check_output "Generate help" "output" "$NVP_BIN generate --help"
}

# ============================================
# Part 2: Library Operations
# ============================================
test_part2() {
    log_section "Part 2: Library Operations"
    
    # 2.1 List
    check_output "Library list" "telescope" "$NVP_BIN library list"
    check_output "Library list YAML" "apiVersion:" "$NVP_BIN library list -o yaml"
    check_output "Library list JSON" '^\[' "$NVP_BIN library list -o json"
    check_output "Library list by category" "lspconfig|mason" "$NVP_BIN library list --category lsp"
    
    # 2.2 Show
    check_output "Library show telescope" "telescope" "$NVP_BIN library show telescope"
    check_output "Library show YAML" "spec:" "$NVP_BIN library show telescope -o yaml"
    check_output "Library show JSON" '"APIVersion"' "$NVP_BIN library show telescope -o json"
    expect_fail "Library show nonexistent" "$NVP_BIN library show nonexistent-plugin-xyz"
    
    # 2.3 Metadata
    check_output "Library categories" "lsp|ui|completion" "$NVP_BIN library categories"
    check_output "Library tags" "finder|git|lsp" "$NVP_BIN library tags"
    
    # 2.4 Install (tested in Part 3 with full CRUD)
}

# ============================================
# Part 3: Plugin CRUD Operations
# ============================================
test_part3() {
    log_section "Part 3: Plugin CRUD Operations"
    
    # 3.1 Initialize
    run_cmd "Init store" "$NVP_BIN init"
    run_cmd "Plugins dir exists" "test -d $NVP_CONFIG_DIR/plugins"
    
    # 3.2 Install from library
    run_cmd "Install telescope" "$NVP_BIN library install telescope"
    run_cmd "Install multiple" "$NVP_BIN library install treesitter nvim-cmp"
    
    # 3.3 Apply from YAML
    cat > "$TEST_DIR/test-plugin.yaml" << 'EOF'
apiVersion: nvp.io/v1
kind: NvimPlugin
metadata:
  name: test-plugin
  description: Test plugin for automated tests
  category: testing
  tags: ["test", "automated"]
spec:
  repo: test/test-plugin
  branch: main
  event: VeryLazy
  config: |
    require("test-plugin").setup({})
EOF
    
    run_cmd "Apply from file" "$NVP_BIN apply -f $TEST_DIR/test-plugin.yaml"
    run_cmd "Apply update" "$NVP_BIN apply -f $TEST_DIR/test-plugin.yaml"
    run_cmd "Apply from stdin" "cat $TEST_DIR/test-plugin.yaml | $NVP_BIN apply -f -"
    
    # 3.4 List
    check_output "List plugins - telescope" "telescope" "$NVP_BIN list"
    check_output "List plugins - treesitter" "treesitter" "$NVP_BIN list"
    check_output "List YAML" "apiVersion:" "$NVP_BIN list -o yaml"
    check_output "List JSON" '^\[' "$NVP_BIN list -o json"
    check_output "List by category" "test-plugin" "$NVP_BIN list --category testing"
    
    # 3.5 Get
    check_output "Get telescope" "telescope" "$NVP_BIN get telescope"
    check_output "Get as JSON" '"APIVersion"' "$NVP_BIN get telescope -o json"
    expect_fail "Get nonexistent" "$NVP_BIN get nonexistent-plugin-xyz"
    
    # 3.6 Enable/Disable
    run_cmd "Disable telescope" "$NVP_BIN disable telescope"
    check_output "Telescope disabled" "no" "$NVP_BIN list | grep telescope"
    run_cmd "Enable telescope" "$NVP_BIN enable telescope"
    check_output "Telescope enabled" "yes" "$NVP_BIN list | grep telescope"
    
    # 3.7 Delete
    run_cmd "Delete test-plugin" "$NVP_BIN delete test-plugin --force"
    expect_fail "Verify deleted" "$NVP_BIN get test-plugin"
}

# ============================================
# Part 4: Lua Generation
# ============================================
test_part4() {
    log_section "Part 4: Lua Generation"
    
    OUTPUT_DIR="$TEST_DIR/lua-output"
    mkdir -p "$OUTPUT_DIR"
    
    # 4.1 Generate all
    run_cmd "Generate Lua" "$NVP_BIN generate --output $OUTPUT_DIR"
    run_cmd "Telescope.lua exists" "test -f $OUTPUT_DIR/telescope.lua"
    run_cmd "Treesitter.lua exists" "test -f $OUTPUT_DIR/treesitter.lua"
    
    # 4.2 Verify content
    check_output "Has return statement" "^return {" "cat $OUTPUT_DIR/telescope.lua"
    check_output "Has repo" "nvim-telescope" "cat $OUTPUT_DIR/telescope.lua"
    
    # 4.3 Dry run
    check_output "Dry run" "Would generate" "$NVP_BIN generate --dry-run"
    
    # 4.4 Generate single
    check_output "Generate-lua single" "return {" "$NVP_BIN generate-lua telescope"
    expect_fail "Generate-lua nonexistent" "$NVP_BIN generate-lua nonexistent-xyz"
    
    # 4.5 Verify Lua syntax (if luac available)
    if command -v luac &>/dev/null; then
        for f in "$OUTPUT_DIR"/*.lua; do
            run_cmd "Lua syntax: $(basename $f)" "luac -p $f"
        done
    else
        log_skip "Lua syntax check (luac not installed)"
    fi
}

# ============================================
# Part 6: Error Handling
# ============================================
test_part6() {
    log_section "Part 6: Error Handling & Edge Cases"
    
    # 6.1 Invalid input
    expect_fail "Apply missing file" "$NVP_BIN apply -f /nonexistent/path/file.yaml"
    expect_fail "Apply no flag" "$NVP_BIN apply"
    expect_fail "Get no arg" "$NVP_BIN get"
    
    # 6.2 Invalid YAML
    echo "not valid yaml {{{{" > "$TEST_DIR/bad.yaml"
    expect_fail "Apply invalid YAML" "$NVP_BIN apply -f $TEST_DIR/bad.yaml"
    
    # Missing required fields
    cat > "$TEST_DIR/missing-repo.yaml" << 'EOF'
apiVersion: nvp.io/v1
kind: NvimPlugin
metadata:
  name: missing-repo
spec: {}
EOF
    expect_fail "Apply missing repo" "$NVP_BIN apply -f $TEST_DIR/missing-repo.yaml"
}

# ============================================
# Part 7: Shell Completions
# ============================================
test_part7() {
    log_section "Part 7: Shell Completions"
    
    run_cmd "Bash completion" "$NVP_BIN completion bash > $TEST_DIR/nvp.bash"
    run_cmd "Zsh completion" "$NVP_BIN completion zsh > $TEST_DIR/_nvp"
    run_cmd "Fish completion" "$NVP_BIN completion fish > $TEST_DIR/nvp.fish"
    
    check_output "Bash script valid" "complete|_nvp" "cat $TEST_DIR/nvp.bash"
    check_output "Zsh script valid" "compdef|_nvp" "cat $TEST_DIR/_nvp"
    check_output "Fish script valid" "complete" "cat $TEST_DIR/nvp.fish"
}

# ============================================
# Part 8: Unit Tests
# ============================================
test_part8() {
    log_section "Part 8: Unit Tests (Go)"
    
    cd "$PROJECT_ROOT"
    run_cmd "go test ./pkg/nvimops/..." "go test ./pkg/nvimops/..."
}

# ============================================
# Main
# ============================================
main() {
    echo -e "${BLUE}"
    echo "============================================"
    echo "  nvp Automated Test Suite"
    echo "============================================"
    echo -e "${NC}"
    
    setup
    
    test_part1
    test_part2
    test_part3
    test_part4
    test_part6
    test_part7
    test_part8
    
    # Summary
    log_section "Test Summary"
    echo ""
    echo -e "${GREEN}Passed:${NC}  $TESTS_PASSED"
    echo -e "${RED}Failed:${NC}  $TESTS_FAILED"
    echo -e "${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
    echo ""
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"
