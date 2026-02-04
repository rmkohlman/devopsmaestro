# Themes

Managing Neovim themes with nvp.

---

## Installing Themes

### From Library

```bash
# See available themes
nvp theme library list

# Install a theme
nvp theme library install tokyonight-custom

# Install and set as active
nvp theme library install catppuccin-mocha --use
```

### From File

```bash
nvp theme apply -f my-theme.yaml
```

---

## Using Themes

### Set Active Theme

```bash
nvp theme use tokyonight-custom
```

### View Active Theme

```bash
nvp theme get
```

### Generate Theme Files

```bash
nvp generate
# or just theme:
nvp theme generate
```

---

## Theme Library

8 pre-built themes:

| Theme | Style | Plugin |
|-------|-------|--------|
| `tokyonight-custom` | Dark blue | folke/tokyonight.nvim |
| `tokyonight-night` | Dark blue | folke/tokyonight.nvim |
| `catppuccin-mocha` | Dark pastel | catppuccin/nvim |
| `catppuccin-latte` | Light pastel | catppuccin/nvim |
| `gruvbox-dark` | Warm retro | ellisonleao/gruvbox.nvim |
| `nord` | Arctic blue | shaunsingh/nord.nvim |
| `rose-pine` | Natural pine | rose-pine/neovim |
| `kanagawa` | Japanese art | rebelot/kanagawa.nvim |

View details:

```bash
nvp theme library show tokyonight-custom
```

---

## Theme YAML Format

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-custom-theme
  description: My personal colorscheme
  author: Your Name
  category: dark
spec:
  plugin:
    repo: folke/tokyonight.nvim
  style: night                    # Theme variant
  transparent: false              # Transparent background
  colors:
    bg: "#1a1b26"
    fg: "#c0caf5"
    accent: "#7aa2f7"
    comment: "#565f89"
    keyword: "#bb9af7"
    string: "#9ece6a"
    function: "#7aa2f7"
    variable: "#c0caf5"
    type: "#2ac3de"
    constant: "#ff9e64"
    error: "#f7768e"
    warning: "#e0af68"
    info: "#7dcfff"
    hint: "#1abc9c"
    selection: "#33467c"
    border: "#29a4bd"
  options:                        # Plugin-specific options
    dim_inactive: true
    styles:
      comments: italic
      keywords: bold
```

---

## Exported Palette

Themes export a color palette that other plugins can use:

```lua
-- In your Neovim config
local palette = require("theme").palette

-- Use colors
vim.api.nvim_set_hl(0, "MyHighlight", {
  fg = palette.colors.accent,
  bg = palette.colors.bg,
})

-- Built-in helpers
local lualine_theme = require("theme").lualine_theme()
```

---

## Generated Files

When you run `nvp generate`, theme files are created:

```
~/.config/nvim/lua/
├── theme/
│   ├── init.lua        # Theme setup and helpers
│   └── palette.lua     # Color palette module
└── plugins/nvp/
    └── colorscheme.lua # Lazy.nvim plugin spec
```

### palette.lua

```lua
return {
  colors = {
    bg = "#1a1b26",
    fg = "#c0caf5",
    accent = "#7aa2f7",
    -- ... all colors
  },
  name = "my-custom-theme",
  style = "night",
}
```

### init.lua

```lua
local M = {}

M.palette = require("theme.palette")

function M.lualine_theme()
  -- Returns lualine theme config
end

function M.highlight(group, opts)
  -- Apply highlights using palette
end

return M
```

---

## Managing Themes

### List Installed Themes

```bash
nvp theme list
```

### View Theme Details

```bash
nvp theme get my-theme
nvp theme get my-theme -o yaml
```

### Delete a Theme

```bash
nvp theme delete my-theme
```

---

## Switching Themes

```bash
# Switch to a different theme
nvp theme use gruvbox-dark

# Regenerate Lua files
nvp generate

# Restart Neovim to see changes
```

---

## Next Steps

- [Plugins](plugins.md) - Manage plugins
- [Commands Reference](commands.md) - Full command list
