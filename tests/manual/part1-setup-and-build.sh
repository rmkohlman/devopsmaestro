#!/bin/bash
# =============================================================================
# DVM Manual Test - Part 1: Setup and Build
# =============================================================================
# Purpose: Clean slate → init → create project → build image
# Time: ~5 minutes
# 
# Usage: 
#   cd ~/Developer/tools/devopsmaestro
#   source tests/manual/part1-setup-and-build.sh
#
# Log file: ~/.devopsmaestro/logs/test-part1-TIMESTAMP.log
# =============================================================================

# -----------------------------------------------------------------------------
# CONFIGURATION
# -----------------------------------------------------------------------------
DVM_PROJECT_DIR=~/Developer/tools/devopsmaestro
DVM_SANDBOX=~/Developer/sandbox
DVM_TEST_PROJECT=$DVM_SANDBOX/dvm-test-fastapi
DVM_LOG_DIR=~/.devopsmaestro/logs
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="$DVM_LOG_DIR/test-part1-$TIMESTAMP.log"

# Track if logging is enabled (disabled during cleanup)
LOG_ENABLED=false

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

# Initialize log file (call after cleanup)
init_log() {
    mkdir -p "$DVM_LOG_DIR"
    LOG_ENABLED=true
    
    cat > "$LOG_FILE" << EOF
================================================================================
DVM MANUAL TEST - PART 1: SETUP AND BUILD
================================================================================
Started: $(date)
Host: $(hostname)
User: $(whoami)
Shell: $SHELL
Working Dir: $(pwd)

Platform Info:
$(uname -a)

Go Version:
$(go version 2>/dev/null || echo "Go not found")

Docker Version:
$(docker --version 2>/dev/null || echo "Docker not found")

Colima Status:
$(colima status 2>/dev/null || echo "Colima not running")
================================================================================

EOF
}

# Log to file only (verbose) - only if logging enabled
log_debug() {
    if $LOG_ENABLED; then
        echo "[$(date '+%H:%M:%S')] DEBUG: $*" >> "$LOG_FILE"
    fi
}

# Log to file with full command output
log_cmd() {
    if $LOG_ENABLED; then
        echo "" >> "$LOG_FILE"
        echo "[$(date '+%H:%M:%S')] COMMAND: $1" >> "$LOG_FILE"
        echo "----------------------------------------" >> "$LOG_FILE"
    fi
}

# Log command output to file
log_output() {
    if $LOG_ENABLED; then
        cat >> "$LOG_FILE"
    else
        cat > /dev/null
    fi
}

# Log to both console and file
log_info() {
    echo -e "${CYAN}ℹ${NC}  $*"
    if $LOG_ENABLED; then
        echo "[$(date '+%H:%M:%S')] INFO: $*" >> "$LOG_FILE"
    fi
}

log_success() {
    echo -e "${GREEN}✓${NC}  $*"
    if $LOG_ENABLED; then
        echo "[$(date '+%H:%M:%S')] SUCCESS: $*" >> "$LOG_FILE"
    fi
}

log_warning() {
    echo -e "${YELLOW}⚠${NC}  $*"
    if $LOG_ENABLED; then
        echo "[$(date '+%H:%M:%S')] WARNING: $*" >> "$LOG_FILE"
    fi
}

log_error() {
    echo -e "${RED}✗${NC}  $*"
    if $LOG_ENABLED; then
        echo "[$(date '+%H:%M:%S')] ERROR: $*" >> "$LOG_FILE"
    fi
}

# Section headers
log_section() {
    echo ""
    echo -e "${BOLD}${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${BLUE}  $*${NC}"
    echo -e "${BOLD}${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    if $LOG_ENABLED; then
        echo "" >> "$LOG_FILE"
        echo "================================================================================" >> "$LOG_FILE"
        echo "SECTION: $*" >> "$LOG_FILE"
        echo "Time: $(date)" >> "$LOG_FILE"
        echo "================================================================================" >> "$LOG_FILE"
    fi
}

log_step() {
    echo -e "  ${CYAN}→${NC} $*"
    if $LOG_ENABLED; then
        echo "[$(date '+%H:%M:%S')] STEP: $*" >> "$LOG_FILE"
    fi
}

# -----------------------------------------------------------------------------
# TEST ASSERTION FUNCTIONS
# -----------------------------------------------------------------------------

# Run a test and track pass/fail
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected="$3"  # Optional expected output pattern
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    log_cmd "$test_cmd"
    
    # Run command and capture output
    local output
    local exit_code
    output=$(eval "$test_cmd" 2>&1)
    exit_code=$?
    
    # Log full output
    if $LOG_ENABLED; then
        echo "$output" >> "$LOG_FILE"
        echo "Exit code: $exit_code" >> "$LOG_FILE"
        echo "----------------------------------------" >> "$LOG_FILE"
    fi
    
    # Check result
    if [ $exit_code -eq 0 ]; then
        if [ -n "$expected" ]; then
            if echo "$output" | grep -q "$expected"; then
                TESTS_PASSED=$((TESTS_PASSED + 1))
                log_success "$test_name"
                return 0
            else
                TESTS_FAILED=$((TESTS_FAILED + 1))
                FAILED_TESTS+=("$test_name")
                log_error "$test_name (expected pattern not found: $expected)"
                return 1
            fi
        else
            TESTS_PASSED=$((TESTS_PASSED + 1))
            log_success "$test_name"
            return 0
        fi
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_name")
        log_error "$test_name (exit code: $exit_code)"
        return 1
    fi
}

# Run command with output displayed (for commands where we want to see output)
run_cmd() {
    local description="$1"
    local cmd="$2"
    
    log_step "$description"
    log_cmd "$cmd"
    
    # Run and tee to both console and log
    if $LOG_ENABLED; then
        eval "$cmd" 2>&1 | tee -a "$LOG_FILE"
    else
        eval "$cmd" 2>&1
    fi
    local exit_code=${PIPESTATUS[0]}
    
    if $LOG_ENABLED; then
        echo "Exit code: $exit_code" >> "$LOG_FILE"
        echo "" >> "$LOG_FILE"
    fi
    
    return $exit_code
}

# Run command silently (log only, console shows just status)
run_silent() {
    local description="$1"
    local cmd="$2"
    
    log_step "$description"
    log_cmd "$cmd"
    
    local output
    output=$(eval "$cmd" 2>&1)
    local exit_code=$?
    
    if $LOG_ENABLED; then
        echo "$output" >> "$LOG_FILE"
        echo "Exit code: $exit_code" >> "$LOG_FILE"
        echo "" >> "$LOG_FILE"
    fi
    
    if [ $exit_code -eq 0 ]; then
        log_success "$description"
    else
        log_warning "$description (exit: $exit_code)"
    fi
    
    return $exit_code
}

# -----------------------------------------------------------------------------
# MAIN TEST SCRIPT
# -----------------------------------------------------------------------------

main() {
    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║     DVM MANUAL TEST - PART 1: SETUP AND BUILD                ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # -------------------------------------------------------------------------
    # SEGMENT 1: CLEANUP (logging disabled during this phase)
    # -------------------------------------------------------------------------
    log_section "SEGMENT 1: CLEANUP AND INITIALIZATION"
    
    builtin cd ~
    
    log_step "Removing dvm data directory"
    rm -rf ~/.devopsmaestro
    log_success "Removed dvm data directory"
    
    log_step "Creating sandbox directory"
    mkdir -p $DVM_SANDBOX
    log_success "Created sandbox directory"
    
    log_step "Removing test project directories"
    rm -rf $DVM_SANDBOX/dvm-test-fastapi $DVM_SANDBOX/dvm-test-golang $DVM_SANDBOX/dvm-new-project $DVM_SANDBOX/dvm-attach-test
    log_success "Removed test project directories"
    
    log_step "Removing Docker test containers"
    docker ps -aq --filter "name=dvm-" | xargs -r docker rm -f 2>/dev/null || true
    log_success "Removed Docker test containers"
    
    log_step "Removing Docker test images"
    docker images --format '{{.Repository}}:{{.Tag}}' | grep "^dvm-" | xargs -r docker rmi 2>/dev/null || true
    log_success "Removed Docker test images"
    
    log_step "Cleaning Colima containerd containers"
    colima ssh -- sudo nerdctl --namespace devopsmaestro ps -aq 2>/dev/null | xargs -I {} colima ssh -- sudo nerdctl --namespace devopsmaestro rm -f {} 2>/dev/null || true
    log_success "Cleaned Colima containerd containers"
    
    log_step "Cleaning Colima containerd images"
    colima ssh -- sudo nerdctl --namespace devopsmaestro images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null | grep "^dvm-" | xargs -I {} colima ssh -- sudo nerdctl --namespace devopsmaestro rmi {} 2>/dev/null || true
    log_success "Cleaned Colima containerd images"
    
    log_step "Building dvm binary"
    builtin cd $DVM_PROJECT_DIR
    if go build -o dvm 2>&1; then
        log_success "Built dvm binary"
    else
        log_error "Failed to build dvm binary"
        return 1
    fi
    
    export PATH="$DVM_PROJECT_DIR:$PATH"
    
    log_step "Initializing dvm"
    dvm admin init 2>&1
    log_success "Initialized dvm"
    
    # NOW enable logging (after dvm admin init created the directory structure)
    init_log
    echo -e "  Log file: ${CYAN}$LOG_FILE${NC}"
    echo ""
    
    # Log what we did during cleanup
    cat >> "$LOG_FILE" << EOF

--- CLEANUP PHASE (completed before logging started) ---
✓ Removed ~/.devopsmaestro
✓ Created sandbox directory
✓ Removed test project directories
✓ Cleaned Docker containers and images
✓ Cleaned Colima containerd containers and images
✓ Built dvm binary
✓ Initialized dvm
---------------------------------------------------------

EOF
    
    # Verify initialization
    log_info "Verifying initialization..."
    run_test "Config directory exists" "test -d ~/.devopsmaestro" 
    run_test "Database exists" "test -f ~/.devopsmaestro/devopsmaestro.db"
    run_test "Database has tables" "sqlite3 ~/.devopsmaestro/devopsmaestro.db '.tables'" "projects"
    
    # -------------------------------------------------------------------------
    # SEGMENT 2: PLATFORM DETECTION
    # -------------------------------------------------------------------------
    log_section "SEGMENT 2: PLATFORM DETECTION"
    
    log_info "Detecting platforms..."
    echo ""
    
    run_cmd "Platform detection (table)" "dvm get platforms"
    echo ""
    
    run_test "Platform detection (yaml)" "dvm get platforms -o yaml" "type:"
    run_test "Platform detection (json)" "dvm get platforms -o json" '"type":'
    
    echo ""
    log_info "Testing platform selection..."
    run_test "DVM_PLATFORM=colima" "DVM_PLATFORM=colima dvm get platforms" ""
    run_test "DVM_PLATFORM=podman" "DVM_PLATFORM=podman dvm get platforms" ""
    
    # -------------------------------------------------------------------------
    # SEGMENT 3: CREATE PROJECT
    # -------------------------------------------------------------------------
    log_section "SEGMENT 3: CLONE AND CREATE PROJECT"
    
    builtin cd ~
    
    log_step "Cloning test project..."
    log_cmd "git clone https://github.com/rmkohlman/dvm-test-fastapi.git $DVM_TEST_PROJECT"
    if git clone https://github.com/rmkohlman/dvm-test-fastapi.git $DVM_TEST_PROJECT 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Cloned test project"
    else
        log_error "Failed to clone test project"
        return 1
    fi
    
    builtin cd $DVM_TEST_PROJECT
    log_info "Current directory: $(pwd)"
    log_debug "Directory contents: $(ls -la)"
    
    echo ""
    log_info "Verifying clean project (no dvm files)..."
    run_test "No .config/nvim" "test ! -d .config/nvim"
    run_test "No Dockerfile.dvm" "test ! -f Dockerfile.dvm"
    
    echo ""
    log_step "Creating dvm project..."
    log_cmd "dvm create project fastapi-test --from-cwd"
    dvm create project fastapi-test --from-cwd 2>&1 | tee -a "$LOG_FILE"
    
    echo ""
    log_info "Verifying project creation..."
    run_cmd "Project list" "dvm get projects"
    
    run_test "Project exists in database" \
        "sqlite3 ~/.devopsmaestro/devopsmaestro.db \"SELECT path FROM projects WHERE name='fastapi-test';\"" \
        "dvm-test-fastapi"
    
    echo ""
    run_cmd "Project details (yaml)" "dvm get project fastapi-test -o yaml"
    
    echo ""
    log_info "Verifying workspace..."
    run_cmd "Workspace list" "dvm get workspaces"
    run_test "Main workspace exists" "dvm get workspace main" "main"
    
    # -------------------------------------------------------------------------
    # SEGMENT 4: WORKSPACE MANAGEMENT
    # -------------------------------------------------------------------------
    log_section "SEGMENT 4: WORKSPACE MANAGEMENT"
    
    builtin cd $DVM_TEST_PROJECT
    
    log_step "Creating additional workspaces..."
    log_cmd "dvm create workspace dev"
    dvm create workspace dev 2>&1 | tee -a "$LOG_FILE"
    
    log_cmd "dvm create workspace staging --description 'Staging environment'"
    dvm create workspace staging --description "Staging environment" 2>&1 | tee -a "$LOG_FILE"
    
    echo ""
    run_cmd "All workspaces" "dvm get workspaces"
    
    run_test "Dev workspace exists" "dvm get workspace dev" "dev"
    run_test "Staging workspace exists" "dvm get workspace staging" "staging"
    
    echo ""
    log_info "Testing workspace switching..."
    run_test "Switch to main" "dvm use workspace main" "Switched to workspace"
    run_test "Switch to dev" "dvm use workspace dev" "Switched to workspace"
    run_test "Switch to staging" "dvm use workspace staging" "Switched to workspace"
    run_test "Switch back to main" "dvm use workspace main" "Switched to workspace"
    
    echo ""
    log_info "Verifying context..."
    run_cmd "Context file" "cat ~/.devopsmaestro/context.yaml"
    
    # -------------------------------------------------------------------------
    # SEGMENT 5: BUILD
    # -------------------------------------------------------------------------
    log_section "SEGMENT 5: BUILD IMAGE"
    
    builtin cd $DVM_TEST_PROJECT
    
    run_silent "Removing nvim config (test skip behavior)" "rm -rf .config/nvim"
    
    log_step "Building workspace image (this may take a few minutes)..."
    echo ""
    log_cmd "dvm build"
    
    # Show build output in real-time
    if dvm build 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Image built successfully"
    else
        log_error "Build failed"
        # Don't exit - continue to show summary
    fi
    
    echo ""
    log_info "Verifying build artifacts..."
    run_test "Dockerfile.dvm created" "test -f Dockerfile.dvm"
    
    echo ""
    log_info "Checking for image..."
    if colima status 2>/dev/null | grep -q containerd; then
        run_cmd "Images in containerd namespace" \
            "colima nerdctl -- --namespace devopsmaestro images | grep -E '(REPOSITORY|dvm-)' || echo 'No dvm images found'"
    else
        run_cmd "Images in Docker" \
            "docker images | grep -E '(REPOSITORY|dvm-)' || echo 'No dvm images found'"
    fi
    
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
    
    echo -e "  ${BOLD}Log file:${NC} $LOG_FILE"
    echo ""
    
    # Log summary to file
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
    # NEXT STEPS
    # -------------------------------------------------------------------------
    echo -e "${BOLD}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║                    PART 1 COMPLETE                           ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${BOLD}Next steps (interactive testing):${NC}"
    echo ""
    echo -e "    1. Attach to the workspace:"
    echo -e "       ${CYAN}builtin cd $DVM_TEST_PROJECT${NC}"
    echo -e "       ${CYAN}dvm attach${NC}"
    echo ""
    echo -e "    2. Inside the container, run:"
    echo -e "       ${CYAN}whoami${NC}          # Should be 'dev'"
    echo -e "       ${CYAN}pwd${NC}             # Should be '/workspace'"
    echo -e "       ${CYAN}nvim --version${NC}  # Should be 0.9+"
    echo -e "       ${CYAN}python3 --version${NC}"
    echo -e "       ${CYAN}exit${NC}            # When done"
    echo ""
    echo -e "    3. After exiting, run Part 2:"
    echo -e "       ${CYAN}source tests/manual/part2-post-attach.sh${NC}"
    echo ""
    
    # Return appropriate exit code
    if [ $TESTS_FAILED -gt 0 ]; then
        return 1
    fi
    return 0
}

# Run main
main
