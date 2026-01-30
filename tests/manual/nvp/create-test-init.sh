#!/bin/bash
#
# create-test-init.sh - Creates a minimal init.lua for testing nvp generated plugins
#

mkdir -p /tmp/nvp-test/lua/plugins/managed

cat > /tmp/nvp-test/init.lua << 'INITEOF'
-- Minimal init.lua for testing nvp generated plugins
vim.g.mapleader = " "
vim.g.maplocalleader = " "

-- Bootstrap lazy.nvim
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git", "clone", "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable", lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)

-- Load lazy.nvim with our plugins
require("lazy").setup({
  spec = {
    { import = "plugins.managed" },
  },
  defaults = { lazy = true },
})
INITEOF

echo "Created /tmp/nvp-test/init.lua"
echo ""
echo "Next steps:"
echo "  1. Generate plugins: ./nvp generate --output /tmp/nvp-test/lua/plugins/managed"
echo "  2. First run (bootstraps lazy.nvim): XDG_CONFIG_HOME=/tmp NVIM_APPNAME=nvp-test nvim"
echo "     (Run :Lazy sync then :qa to exit)"
echo "  3. Headless test: XDG_CONFIG_HOME=/tmp NVIM_APPNAME=nvp-test nvim --headless -c 'qa'"
