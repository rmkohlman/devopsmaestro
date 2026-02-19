# Themes

Managing Neovim themes with nvp. **Library themes are automatically available** - no installation needed for 34+ embedded themes!

---

## Quick Start

```bash
# Library themes work out of the box
dvm get nvim theme coolnight-ocean      # No installation needed
dvm get nvim themes                     # Shows all user + library themes

# Or install to user store (for customization)
nvp theme library install tokyonight-custom --use
```

---

## Theme Availability

DevOpsMaestro now includes **34+ embedded library themes** that are automatically available without installation:

- **Automatic access** - Use any library theme directly: `dvm get nvim theme <name>`
- **No installation required** - Library themes work out of the box
- **User override** - Save a theme with the same name to override library version
- **Combined listing** - `dvm get nvim themes` shows both user and library themes

### Library vs User Themes

| Type | Location | Access | Customizable |
|------|----------|--------|--------------|
| **Library** | Embedded in binary | Automatic | No (read-only) |  
| **User** | `~/.nvp/themes/` | Manual install | Yes (full control) |

**Override behavior**: User themes with the same name take precedence over library themes.

---

## Installing Themes

### From Library (Automatic)

```bash
# Library themes work immediately - no installation needed
dvm get nvim theme coolnight-ocean
dvm get nvim theme tokyonight-night
dvm get nvim theme catppuccin-mocha

# List all available themes (user + library)
dvm get nvim themes
```

### From Library (Manual Install)

```bash
# See available themes
nvp theme library list

# Install to user store (for customization)
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

**34+ embedded themes** are automatically available without installation, including all CoolNight variants, TokyoNight, Catppuccin, Dracula, Gruvbox, Nord, and more.

### Popular Library Themes

| Theme | Style | Plugin |
|-------|-------|--------|
| `coolnight-ocean` | Deep blue | CoolNight variant |
| `coolnight-synthwave` | Neon purple | CoolNight variant |
| `tokyonight-night` | Dark blue | folke/tokyonight.nvim |
| `tokyonight-storm` | Stormy blue | folke/tokyonight.nvim |
| `catppuccin-mocha` | Dark pastel | catppuccin/nvim |
| `catppuccin-latte` | Light pastel | catppuccin/nvim |
| `dracula` | Dark purple | Mofiqul/dracula.nvim |
| `gruvbox-dark` | Warm retro | ellisonleao/gruvbox.nvim |
| `nord` | Arctic blue | shaunsingh/nord.nvim |

**All themes work immediately** - no installation required!

```bash
# Use any theme directly
dvm get nvim theme coolnight-ocean
dvm get nvim theme dracula
dvm get nvim theme catppuccin-mocha
```

View details:

```bash
nvp theme library show tokyonight-custom    # Library theme details
dvm get nvim theme coolnight-ocean          # Works directly, no install needed
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
