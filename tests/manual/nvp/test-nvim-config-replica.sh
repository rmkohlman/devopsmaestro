#!/bin/bash
#
# test-nvim-config-replica.sh
#
# Creates a replica of your nvim-config in a temp directory and demonstrates
# how nvp can manage plugins. This allows testing without affecting your
# real Neovim configuration.
#
# Usage:
#   ./tests/manual/nvp/test-nvim-config-replica.sh
#
# Options:
#   NVP_KEEP_OUTPUT=1  - Don't clean up at the end (for inspection)
#   NVP_INTERACTIVE=1  - Launch Neovim at the end for manual testing
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Test directories
TEST_DIR="${TEST_DIR:-/tmp/nvp-nvim-test-$$}"
NVIM_CONFIG_DIR="$TEST_DIR/nvim"
NVP_CONFIG_DIR="$TEST_DIR/nvp"
NVIM_DATA_DIR="$TEST_DIR/nvim-data"

# Binary
NVP_BIN="$PROJECT_ROOT/nvp"

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  nvp Neovim Config Replica Test${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""
echo "Test directory: $TEST_DIR"
echo ""

# Step 0: Build nvp if needed
echo -e "${YELLOW}Step 0: Building nvp binary...${NC}"
if [[ ! -f "$NVP_BIN" ]] || [[ "$PROJECT_ROOT/cmd/nvp/root.go" -nt "$NVP_BIN" ]]; then
    cd "$PROJECT_ROOT"
    go build -o nvp ./cmd/nvp/
    echo -e "${GREEN}  Built nvp binary${NC}"
else
    echo -e "${GREEN}  nvp binary already exists${NC}"
fi

# Step 1: Clone your nvim-config
echo ""
echo -e "${YELLOW}Step 1: Cloning your nvim-config...${NC}"
mkdir -p "$TEST_DIR"
if [[ -d "/tmp/nvim-config-source" ]]; then
    cp -r /tmp/nvim-config-source "$NVIM_CONFIG_DIR"
    echo -e "${GREEN}  Copied from /tmp/nvim-config-source${NC}"
else
    gh repo clone rmkohlman/nvim-config "$NVIM_CONFIG_DIR" 2>/dev/null || {
        echo -e "${RED}  Failed to clone nvim-config. Make sure gh is authenticated.${NC}"
        exit 1
    }
    echo -e "${GREEN}  Cloned nvim-config${NC}"
fi

# Show structure
echo ""
echo "  Config structure:"
find "$NVIM_CONFIG_DIR" -name "*.lua" | head -10 | while read f; do
    echo "    ${f#$NVIM_CONFIG_DIR/}"
done
echo "    ..."

# Step 2: Initialize nvp store
echo ""
echo -e "${YELLOW}Step 2: Initializing nvp store...${NC}"
export NVP_CONFIG_DIR="$NVP_CONFIG_DIR"
"$NVP_BIN" init
echo -e "${GREEN}  Initialized nvp at $NVP_CONFIG_DIR${NC}"

# Step 3: Install plugins from library
echo ""
echo -e "${YELLOW}Step 3: Installing plugins from nvp library...${NC}"

# These match plugins in your nvim-config
PLUGINS=(
    telescope
    treesitter
    nvim-cmp
    lspconfig
    mason
    gitsigns
    lualine
    which-key
    copilot
    comment
    alpha
    indent-blankline
)

for plugin in "${PLUGINS[@]}"; do
    if "$NVP_BIN" library show "$plugin" &>/dev/null; then
        "$NVP_BIN" library install "$plugin" 2>/dev/null && echo -e "  ${GREEN}Installed: $plugin${NC}" || true
    else
        echo -e "  ${YELLOW}Not in library: $plugin${NC}"
    fi
done

# Step 4: List installed plugins
echo ""
echo -e "${YELLOW}Step 4: Listing installed plugins...${NC}"
"$NVP_BIN" list

# Step 5: Generate Lua files
echo ""
echo -e "${YELLOW}Step 5: Generating Lua files...${NC}"
MANAGED_PLUGINS_DIR="$NVIM_CONFIG_DIR/lua/workspace/plugins/managed"
mkdir -p "$MANAGED_PLUGINS_DIR"
"$NVP_BIN" generate --output "$MANAGED_PLUGINS_DIR"
echo ""
echo "  Generated files:"
ls -la "$MANAGED_PLUGINS_DIR" | tail -n +2 | head -10

# Step 6: Create an integration point
echo ""
echo -e "${YELLOW}Step 6: Creating integration with lazy.nvim...${NC}"

# Modify lazy.lua to also import managed plugins
LAZY_FILE="$NVIM_CONFIG_DIR/lua/workspace/lazy.lua"
if grep -q "workspace.plugins.managed" "$LAZY_FILE"; then
    echo -e "  ${GREEN}Already configured to import managed plugins${NC}"
else
    # Backup original
    cp "$LAZY_FILE" "$LAZY_FILE.bak"
    
    # Update to include managed plugins
    cat > "$LAZY_FILE" << 'LAZY_EOF'
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git",
    "clone",
    "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable",
    lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)

require("lazy").setup({
  { import = "workspace.plugins" },
  { import = "workspace.plugins.lsp" },
  { import = "workspace.plugins.managed" },  -- nvp managed plugins
}, {
  change_detection = {
    notify = false,
  },
})
LAZY_EOF
    echo -e "  ${GREEN}Updated lazy.lua to import managed plugins${NC}"
fi

# Step 7: Show summary
echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""
echo "Test Neovim config: $NVIM_CONFIG_DIR"
echo "nvp config dir:     $NVP_CONFIG_DIR"
echo "Managed plugins:    $MANAGED_PLUGINS_DIR"
echo ""
echo "Directory structure:"
echo "  $TEST_DIR/"
echo "  ├── nvim/                    # Your nvim-config replica"
echo "  │   └── lua/workspace/"
echo "  │       └── plugins/"
echo "  │           ├── *.lua        # Original plugins"
echo "  │           └── managed/     # nvp generated plugins"
echo "  └── nvp/                     # nvp plugin store"
echo "      └── plugins/"
echo ""

# Step 8: Verify Lua syntax
echo -e "${YELLOW}Verifying Lua syntax...${NC}"
ERRORS=0
for f in "$MANAGED_PLUGINS_DIR"/*.lua; do
    if command -v luac &>/dev/null; then
        if luac -p "$f" 2>/dev/null; then
            echo -e "  ${GREEN}OK${NC}: $(basename "$f")"
        else
            echo -e "  ${RED}FAIL${NC}: $(basename "$f")"
            ERRORS=$((ERRORS + 1))
        fi
    else
        echo -e "  ${YELLOW}SKIP${NC}: $(basename "$f") (luac not installed)"
    fi
done

if [[ $ERRORS -gt 0 ]]; then
    echo ""
    echo -e "${RED}$ERRORS Lua syntax errors found!${NC}"
fi

# Step 9: Interactive test (optional)
if [[ "${NVP_INTERACTIVE:-0}" == "1" ]]; then
    echo ""
    echo -e "${YELLOW}Launching Neovim with test config...${NC}"
    echo "  (use :Lazy to see plugins, :q to quit)"
    echo ""
    
    # Use NVIM_APPNAME to isolate data directory
    XDG_CONFIG_HOME="$TEST_DIR" \
    XDG_DATA_HOME="$NVIM_DATA_DIR" \
    NVIM_APPNAME="nvim" \
    nvim
fi

# Step 10: Headless test
echo ""
echo -e "${YELLOW}Running headless Neovim test...${NC}"
mkdir -p "$NVIM_DATA_DIR"
XDG_CONFIG_HOME="$TEST_DIR" \
XDG_DATA_HOME="$NVIM_DATA_DIR" \
NVIM_APPNAME="nvim" \
nvim --headless -c "lua print('Neovim loaded successfully')" -c "qa" 2>&1 | head -5 || {
    echo -e "  ${YELLOW}Note: Headless test may show warnings for missing plugins${NC}"
    echo -e "  ${YELLOW}(This is expected - plugins need to be installed via :Lazy sync)${NC}"
}

# Cleanup
echo ""
if [[ "${NVP_KEEP_OUTPUT:-0}" == "1" ]]; then
    echo -e "${GREEN}Test output preserved at: $TEST_DIR${NC}"
    echo ""
    echo "To test interactively:"
    echo "  XDG_CONFIG_HOME=$TEST_DIR NVIM_APPNAME=nvim nvim"
    echo ""
    echo "To install plugins:"
    echo "  XDG_CONFIG_HOME=$TEST_DIR NVIM_APPNAME=nvim nvim -c 'Lazy sync'"
else
    echo -e "${YELLOW}Cleaning up test directory...${NC}"
    # Use rm -rf with 2>/dev/null to suppress errors from tree-sitter compiled parsers
    rm -rf "$TEST_DIR" 2>/dev/null || {
        # If first attempt fails, try again after a short delay
        sleep 1
        rm -rf "$TEST_DIR" 2>/dev/null || true
    }
    echo -e "${GREEN}Done.${NC}"
fi

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  Test Complete${NC}"
echo -e "${BLUE}============================================${NC}"
