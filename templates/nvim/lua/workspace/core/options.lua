vim.cmd("let g:netrw_liststyle = 3")

local opt = vim.opt

-- timeout setting // Default is 1000
opt.timeoutlen = 500 -- Wait for 500 milliseconds (adjust as needed)

opt.relativenumber = true
opt.number = true

-- tabs & indentation
opt.tabstop = 2 -- spaces fro tabs (prettier default)
opt.shiftwidth = 2 -- spaces for indent width
opt.expandtab = true -- expand tab to spaces
opt.autoindent = true -- copy indent from current line when starting new one

opt.wrap = false

-- search settings
opt.ignorecase = true -- ignore case when searching
opt.smartcase = true -- if you include mixed case in your search, assumes you want case-sensitive
opt.cursorline = true

-- turn on termguicolors for colorscheme to work
opt.termguicolors = true
opt.background = "dark" -- colorschemes that can be light or dark will be made dark
opt.signcolumn = "yes" -- show sign column so that text doesn't shift

-- backspace
opt.backspace = "indent,eol,start" -- allow backspace on indent, end of line or insert modce start position

-- clipboard - use OSC 52 for container/SSH clipboard support
opt.clipboard:append("unnamedplus") -- use system clipboard as default register

-- OSC 52 clipboard provider for containers
-- This allows yanking to host clipboard over terminal escape sequences
-- Works with: WezTerm, iTerm2, Kitty, Alacritty, Windows Terminal
if os.getenv("CONTAINER") or os.getenv("SSH_TTY") or os.getenv("DVM_WORKSPACE") then
  local function paste()
    return {
      vim.fn.split(vim.fn.getreg(""), "\n"),
      vim.fn.getregtype(""),
    }
  end

  vim.g.clipboard = {
    name = "OSC 52",
    copy = {
      ["+"] = require("vim.ui.clipboard.osc52").copy("+"),
      ["*"] = require("vim.ui.clipboard.osc52").copy("*"),
    },
    paste = {
      ["+"] = paste,
      ["*"] = paste,
    },
  }
end

-- split windows
opt.splitright = true -- split vertical window to the right
opt.splitbelow = true -- split horizontal window to the bottom
