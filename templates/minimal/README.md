# Minimal Template

A simple, well-commented Neovim configuration perfect for beginners and those who want to build their setup from scratch.

## Features

- ✅ **~200 lines** - Easy to understand
- ✅ **Well-commented** - Learn as you go
- ✅ **Essential plugins only** - Colorscheme, statusline, git signs, commenting
- ✅ **Sensible defaults** - Modern editor settings
- ✅ **Lazy.nvim** - Fast, modern plugin manager

## Plugins Included

| Plugin | Purpose |
|--------|---------|
| [catppuccin](https://github.com/catppuccin/nvim) | Beautiful colorscheme |
| [lualine](https://github.com/nvim-lualine/lualine.nvim) | Statusline |
| [gitsigns](https://github.com/lewis6991/gitsigns.nvim) | Git integration |
| [Comment.nvim](https://github.com/numToStr/Comment.nvim) | Easy commenting |
| [vim-sleuth](https://github.com/tpope/vim-sleuth) | Auto-detect indentation |

## Quick Start

```bash
# Initialize with this template
dvm nvim init minimal

# Open Neovim (plugins will auto-install)
nvim

# Check health
:checkhealth
```

## Key Bindings

### General
- `<Space>` - Leader key
- `<Space>w` - Save file
- `<Space>q` - Quit window
- `<Esc>` - Clear search highlighting

### Navigation
- `<C-h/j/k/l>` - Move between windows
- `<C-d/u>` - Scroll down/up (centered)

### Editing
- `gcc` - Comment/uncomment line
- `gc` (visual) - Comment selection
- `</>` (visual) - Indent left/right (stay in visual mode)
- `J/K` (visual) - Move selection up/down

### Plugins
- `:Lazy` - Open plugin manager
- `:Lazy sync` - Update plugins

## Customization

### Change Colorscheme

Edit `init.lua` line 115:

```lua
flavour = 'mocha', -- Options: latte, frappe, macchiato, mocha
```

Or use a different colorscheme entirely:

```lua
{
  'folke/tokyonight.nvim',
  priority = 1000,
  config = function()
    vim.cmd.colorscheme 'tokyonight'
  end,
},
```

### Add More Plugins

Add to the `require('lazy').setup({})` table:

```lua
require('lazy').setup({
  -- ... existing plugins ...
  
  -- Your new plugin
  {
    'plugin-author/plugin-name',
    config = function()
      require('plugin-name').setup{}
    end,
  },
})
```

### Change Leader Key

Edit `init.lua` line 8:

```lua
vim.g.mapleader = ","  -- Or any other key
```

## Next Steps

1. **Learn Vim motions** - Run `:Tutor` in Neovim
2. **Add LSP** - Language Server Protocol for code intelligence
3. **Add Telescope** - Fuzzy finder for files
4. **Add Treesitter** - Better syntax highlighting
5. **Explore more plugins** - See [awesome-neovim](https://github.com/rockerBOO/awesome-neovim)

## Growing Your Config

When you're ready for more features:

```bash
# Check out the starter template
dvm nvim init starter

# Or explore full examples
# https://github.com/rmkohlman/devopsmaestro-nvim-templates
```

## Troubleshooting

### Plugins not loading

```vim
:Lazy sync
:Lazy restore
```

### Check for errors

```vim
:checkhealth
:messages
```

### Reset everything

```bash
rm -rf ~/.local/share/nvim ~/.local/state/nvim ~/.cache/nvim
dvm nvim init minimal --overwrite
```

## Resources

- [Neovim Docs](https://neovim.io/doc/)
- [lazy.nvim](https://github.com/folke/lazy.nvim)
- [Learn Vimscript the Hard Way](https://learnvimscriptthehardway.stevelosh.com/)
- [Lua Guide](https://neovim.io/doc/user/lua-guide.html)

---

**Tip:** This is a foundation. Build on it! Start simple, add features as you need them.
