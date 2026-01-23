-- DevOpsMaestro Minimal Neovim Configuration
-- Perfect for beginners and those who want to build their config from scratch

-- ============================================================================
-- BASIC SETTINGS
-- ============================================================================

-- Set leader keys (must be set before loading plugins)
vim.g.mapleader = " "           -- Space as leader key
vim.g.maplocalleader = "\\"     -- Backslash as local leader

-- Line numbers
vim.opt.number = true           -- Show line numbers
vim.opt.relativenumber = true   -- Relative line numbers (great for motions)

-- Mouse support
vim.opt.mouse = 'a'             -- Enable mouse in all modes

-- Editor behavior
vim.opt.showmode = false        -- Don't show mode (statusline handles it)
vim.opt.clipboard = 'unnamedplus' -- Sync with system clipboard
vim.opt.breakindent = true      -- Wrapped lines continue indented
vim.opt.undofile = true         -- Save undo history
vim.opt.updatetime = 250        -- Faster completion (default 4000ms)
vim.opt.timeoutlen = 300        -- Faster mapped sequences

-- Search
vim.opt.ignorecase = true       -- Case-insensitive search
vim.opt.smartcase = true        -- Unless uppercase in search term
vim.opt.hlsearch = true         -- Highlight search results
vim.opt.incsearch = true        -- Incremental search

-- Visual
vim.opt.termguicolors = true    -- True color support
vim.opt.signcolumn = 'yes'      -- Always show sign column (prevents text shifting)
vim.opt.cursorline = true       -- Highlight current line
vim.opt.scrolloff = 10          -- Keep 10 lines above/below cursor
vim.opt.sidescrolloff = 8       -- Keep 8 columns left/right of cursor

-- Splits
vim.opt.splitright = true       -- Vertical splits go right
vim.opt.splitbelow = true       -- Horizontal splits go below

-- Whitespace
vim.opt.list = true             -- Show some invisible characters
vim.opt.listchars = {
  tab = '» ',
  trail = '·',
  nbsp = '␣'
}

-- Search/Replace preview
vim.opt.inccommand = 'split'    -- Preview substitutions live in split

-- ============================================================================
-- BASIC KEYMAPS
-- ============================================================================

-- Clear search highlighting
vim.keymap.set('n', '<Esc>', '<cmd>nohlsearch<CR>', { desc = 'Clear search highlighting' })

-- Quick save/quit
vim.keymap.set('n', '<leader>w', '<cmd>w<CR>', { desc = '[W]rite (save) file' })
vim.keymap.set('n', '<leader>q', '<cmd>q<CR>', { desc = '[Q]uit window' })

-- Better window navigation
vim.keymap.set('n', '<C-h>', '<C-w>h', { desc = 'Move to left window' })
vim.keymap.set('n', '<C-j>', '<C-w>j', { desc = 'Move to bottom window' })
vim.keymap.set('n', '<C-k>', '<C-w>k', { desc = 'Move to top window' })
vim.keymap.set('n', '<C-l>', '<C-w>l', { desc = 'Move to right window' })

-- Stay in visual mode when indenting
vim.keymap.set('v', '<', '<gv', { desc = 'Indent left' })
vim.keymap.set('v', '>', '>gv', { desc = 'Indent right' })

-- Move selected lines
vim.keymap.set('v', 'J', ":m '>+1<CR>gv=gv", { desc = 'Move selection down' })
vim.keymap.set('v', 'K', ":m '<-2<CR>gv=gv", { desc = 'Move selection up' })

-- Keep cursor centered when scrolling
vim.keymap.set('n', '<C-d>', '<C-d>zz', { desc = 'Scroll down (centered)' })
vim.keymap.set('n', '<C-u>', '<C-u>zz', { desc = 'Scroll up (centered)' })

-- ============================================================================
-- PLUGIN MANAGER (lazy.nvim)
-- ============================================================================

-- Bootstrap lazy.nvim (auto-install if not present)
local lazypath = vim.fn.stdpath 'data' .. '/lazy/lazy.nvim'
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system {
    'git',
    'clone',
    '--filter=blob:none',
    'https://github.com/folke/lazy.nvim.git',
    '--branch=stable',
    lazypath,
  }
end
vim.opt.rtp:prepend(lazypath)

-- ============================================================================
-- PLUGINS
-- ============================================================================

require('lazy').setup({
  -- Colorscheme: Catppuccin (popular, well-maintained)
  {
    'catppuccin/nvim',
    name = 'catppuccin',
    priority = 1000, -- Load colorscheme first
    config = function()
      require('catppuccin').setup {
        flavour = 'mocha', -- latte, frappe, macchiato, mocha
        transparent_background = false,
        term_colors = true,
      }
      vim.cmd.colorscheme 'catppuccin-mocha'
    end,
  },

  -- Auto-detect tabstop and shiftwidth
  {
    'tpope/vim-sleuth',
  },

  -- Better statusline (optional but nice)
  {
    'nvim-lualine/lualine.nvim',
    dependencies = { 'nvim-tree/nvim-web-devicons' },
    config = function()
      require('lualine').setup {
        options = {
          theme = 'catppuccin',
          component_separators = '|',
          section_separators = '',
        },
      }
    end,
  },

  -- Git signs in gutter
  {
    'lewis6991/gitsigns.nvim',
    config = function()
      require('gitsigns').setup {
        signs = {
          add          = { text = '+' },
          change       = { text = '~' },
          delete       = { text = '_' },
          topdelete    = { text = '‾' },
          changedelete = { text = '~' },
        },
      }
    end,
  },

  -- Comment lines easily (gcc to toggle)
  {
    'numToStr/Comment.nvim',
    config = function()
      require('Comment').setup()
    end,
  },
}, {
  ui = {
    icons = {
      cmd = "⌘",
      config = "🛠",
      event = "📅",
      ft = "📂",
      init = "⚙",
      keys = "🗝",
      plugin = "🔌",
      runtime = "💻",
      require = "🌙",
      source = "📄",
      start = "🚀",
      task = "📌",
    },
  },
})

-- ============================================================================
-- AUTOCMDS
-- ============================================================================

-- Highlight yanked text briefly
vim.api.nvim_create_autocmd('TextYankPost', {
  desc = 'Highlight when yanking text',
  group = vim.api.nvim_create_augroup('highlight-yank', { clear = true }),
  callback = function()
    vim.highlight.on_yank()
  end,
})

-- ============================================================================
-- TIPS
-- ============================================================================

-- Welcome message
print('🚀 DevOpsMaestro Minimal Config Loaded!')
print('   Leader key: <Space>')
print('   Help: :help')
print('   Plugins: :Lazy')

-- Quick reference:
-- <Space>w   - Save file
-- <Space>q   - Quit
-- gcc        - Comment line
-- <Esc>      - Clear search highlight
-- :Lazy      - Plugin manager
-- :checkhealth - Check Neovim health

-- vim: ts=2 sts=2 sw=2 et
