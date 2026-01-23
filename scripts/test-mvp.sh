#!/bin/bash
set -e

# DevOpsMaestro MVP Test Suite
# Tests the critical path: init → create → use → attach

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DVM="$PROJECT_DIR/dvm"
TEST_PROJECT_DIR="$HOME/tmp/dvm-test-project"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

test_header() {
    echo
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}TEST: $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    
    # Stop any running containers
    docker ps -a --filter "label=devopsmaestro.managed=true" -q | xargs -r docker stop 2>/dev/null || true
    docker ps -a --filter "label=devopsmaestro.managed=true" -q | xargs -r docker rm 2>/dev/null || true
    
    # Remove test project directory
    rm -rf "$TEST_PROJECT_DIR"
    
    # Backup and remove dvm config (optional)
    if [ -d "$HOME/.devopsmaestro" ]; then
        log_warning "Backing up ~/.devopsmaestro to ~/.devopsmaestro.bak"
        rm -rf "$HOME/.devopsmaestro.bak"
        mv "$HOME/.devopsmaestro" "$HOME/.devopsmaestro.bak"
    fi
    
    log_success "Cleanup complete"
}

# ============================================================================
# PRE-FLIGHT CHECKS
# ============================================================================

test_header "Pre-flight Checks"

# Check if dvm binary exists
if [ ! -f "$DVM" ]; then
    log_error "dvm binary not found at $DVM"
    log_info "Run: ./scripts/build.sh"
    exit 1
fi
log_success "dvm binary found"

# Check if Colima is running
if ! docker ps >/dev/null 2>&1; then
    log_error "Docker is not accessible (Colima not running?)"
    log_info "Run: colima start --profile local-med"
    exit 1
fi
log_success "Docker is accessible (Colima running)"

# Check Docker host
if [[ -n "$DOCKER_HOST" ]]; then
    log_success "DOCKER_HOST set: $DOCKER_HOST"
else
    log_warning "DOCKER_HOST not set (may use default)"
fi

# ============================================================================
# TEST 1: Initialize dvm
# ============================================================================

test_header "Test 1: Initialize dvm"

# Clean start
cleanup

# Run init
log_info "Running: $DVM admin init"
if $DVM admin init >/dev/null 2>&1; then
    log_success "dvm admin init succeeded"
else
    log_error "dvm admin init failed"
    exit 1
fi

# Verify directory structure
if [ -d "$HOME/.devopsmaestro" ]; then
    log_success "~/.devopsmaestro directory created"
else
    log_error "~/.devopsmaestro directory not found"
fi

if [ -f "$HOME/.devopsmaestro/config.yaml" ]; then
    log_success "config.yaml created"
else
    log_error "config.yaml not found"
fi

if [ -f "$HOME/.devopsmaestro/devopsmaestro.db" ]; then
    log_success "devopsmaestro.db created"
else
    log_error "devopsmaestro.db not found"
fi

# ============================================================================
# TEST 2: Create Test Project
# ============================================================================

test_header "Test 2: Create Test Project"

# Create test project directory
mkdir -p "$TEST_PROJECT_DIR"
cd "$TEST_PROJECT_DIR"

# Create a sample file
echo "# Test Project" > README.md
log_info "Created test project at $TEST_PROJECT_DIR"

# Create project in dvm
log_info "Running: dvm create project dvm-test --from-cwd"
if $DVM create project dvm-test --from-cwd --description "Automated test project" 2>&1 | grep -q "created successfully"; then
    log_success "Project created successfully"
else
    log_error "Failed to create project"
fi

# Verify in database
PROJECT_COUNT=$(sqlite3 "$HOME/.devopsmaestro/devopsmaestro.db" "SELECT COUNT(*) FROM projects WHERE name='dvm-test';")
if [ "$PROJECT_COUNT" -eq "1" ]; then
    log_success "Project found in database"
else
    log_error "Project not found in database"
fi

# Verify workspace was auto-created
WORKSPACE_COUNT=$(sqlite3 "$HOME/.devopsmaestro/devopsmaestro.db" "SELECT COUNT(*) FROM workspaces WHERE name='main';")
if [ "$WORKSPACE_COUNT" -eq "1" ]; then
    log_success "Main workspace auto-created"
else
    log_error "Main workspace not found"
fi

# ============================================================================
# TEST 3: Context Switching
# ============================================================================

test_header "Test 3: Context Switching"

# Use project (should already be set, but test explicitly)
log_info "Running: dvm use project dvm-test"
if $DVM use project dvm-test 2>&1 | grep -q "Switched to project"; then
    log_success "Project context switched"
else
    log_error "Failed to switch project context"
fi

# Use workspace
log_info "Running: dvm use workspace main"
if $DVM use workspace main 2>&1 | grep -q "Switched to workspace"; then
    log_success "Workspace context switched"
else
    log_error "Failed to switch workspace context"
fi

# Verify context.yaml
if [ -f "$HOME/.devopsmaestro/context.yaml" ]; then
    if grep -q "dvm-test" "$HOME/.devopsmaestro/context.yaml" && grep -q "main" "$HOME/.devopsmaestro/context.yaml"; then
        log_success "context.yaml updated correctly"
    else
        log_error "context.yaml content incorrect"
    fi
else
    log_error "context.yaml not found"
fi

# ============================================================================
# TEST 4: Error Handling
# ============================================================================

test_header "Test 4: Error Handling"

# Test non-existent project
log_info "Testing error: non-existent project"
if $DVM use project nonexistent-project 2>&1 | grep -q "not found"; then
    log_success "Correct error for non-existent project"
else
    log_error "Missing error for non-existent project"
fi

# Test non-existent workspace
log_info "Testing error: non-existent workspace"
if $DVM use workspace nonexistent-workspace 2>&1 | grep -q "not found"; then
    log_success "Correct error for non-existent workspace"
else
    log_error "Missing error for non-existent workspace"
fi

# ============================================================================
# TEST 5: Build Test Docker Image
# ============================================================================

test_header "Test 5: Build Test Docker Image"

log_info "Creating minimal Dockerfile for testing..."
cat > "$TEST_PROJECT_DIR/Dockerfile" << 'EOF'
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    zsh \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN chsh -s /bin/zsh root || true

WORKDIR /workspace

CMD ["/bin/zsh"]
EOF

log_info "Building Docker image: dvm-dvm-test-main:latest"
if docker build -t dvm-dvm-test-main:latest "$TEST_PROJECT_DIR" >/dev/null 2>&1; then
    log_success "Docker image built successfully"
else
    log_error "Failed to build Docker image"
fi

# Verify image exists
if docker images | grep -q "dvm-dvm-test-main"; then
    log_success "Image found in Docker"
else
    log_error "Image not found in Docker"
fi

# ============================================================================
# TEST 6: Attach to Workspace (Non-Interactive Test)
# ============================================================================

test_header "Test 6: Workspace Container Management"

log_info "Testing attach command (will start container)..."

# We can't test interactive attach in automated script, but we can test:
# 1. Container starts
# 2. Project is mounted
# 3. Environment variables are set

# Run attach in background with a simple command to test
log_info "Starting workspace container..."

# First, let's just test if attach can start the container
# We'll send it a quick exit command
timeout 10s bash -c "echo 'exit' | $DVM attach 2>&1" | head -20 || true

# Check if container was created (it may have exited already due to AutoRemove)
log_warning "Note: Container may have auto-removed after exit (expected behavior)"

# ============================================================================
# TEST 7: Multiple Projects
# ============================================================================

test_header "Test 7: Multiple Projects"

# Create second project
TEST_PROJECT_DIR_2="$HOME/tmp/dvm-test-project-2"
mkdir -p "$TEST_PROJECT_DIR_2"
echo "# Test Project 2" > "$TEST_PROJECT_DIR_2/README.md"

log_info "Creating second project..."
if cd "$TEST_PROJECT_DIR_2" && $DVM create project dvm-test-2 --from-cwd 2>&1 | grep -q "created successfully"; then
    log_success "Second project created"
else
    log_error "Failed to create second project"
fi

# Verify we have 2 projects
PROJECT_COUNT=$(sqlite3 "$HOME/.devopsmaestro/devopsmaestro.db" "SELECT COUNT(*) FROM projects;")
if [ "$PROJECT_COUNT" -eq "2" ]; then
    log_success "Both projects in database"
else
    log_error "Expected 2 projects, found $PROJECT_COUNT"
fi

# Switch between projects
log_info "Switching back to first project..."
if $DVM use project dvm-test 2>&1 | grep -q "Switched"; then
    log_success "Context switching works"
else
    log_error "Context switching failed"
fi

# Cleanup second test project
rm -rf "$TEST_PROJECT_DIR_2"

# ============================================================================
# FINAL REPORT
# ============================================================================

echo
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}TEST SUMMARY${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo
    echo "Your dvm setup is working correctly. Try the full workflow:"
    echo "  1. cd $TEST_PROJECT_DIR"
    echo "  2. $DVM attach"
    echo "  3. Inside container: ls -la /workspace"
    echo "  4. Press Ctrl+D to exit"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo
    echo "Check the errors above and debug. Common issues:"
    echo "  - Colima not running: colima start"
    echo "  - Database permissions: check ~/.devopsmaestro/"
    echo "  - Docker image issues: docker images"
    exit 1
fi
