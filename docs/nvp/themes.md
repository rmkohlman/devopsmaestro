# Themes

Managing Neovim themes with nvp. **Library themes are automatically available** - no installation needed for 34+ embedded themes including 21 CoolNight variants!

---

## Quick Start

```bash
# Library themes work out of the box
nvp theme list                        # Shows all available themes
dvm get nvim theme coolnight-ocean    # Use with dvm integration
nvp theme use coolnight-synthwave     # Direct nvp usage

# Parametric generator for CoolNight variants
nvp theme create --hue 210 --name my-blue-theme
nvp theme create --hue 350 --name my-rose-theme
```

---

## Theme Availability

DevOpsMaestro now includes **34+ embedded library themes** that are automatically available without installation:

- **21 CoolNight Variants** - Parametric theme generator with customizable hues
- **13+ Popular Themes** - TokyoNight, Catppuccin, Dracula, Gruvbox, Nord, and more
- **Automatic access** - Use any library theme directly: `nvp theme use <name>`
- **No installation required** - Library themes work out of the box
- **Parametric generator** - Create custom CoolNight variants with `nvp theme create`
- **User override** - Save a theme with the same name to override library version
- **Combined listing** - `nvp theme list` shows both user and library themes

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
nvp theme list                     # List all available themes
nvp theme use coolnight-ocean      # Use a theme
nvp theme use tokyonight-night
nvp theme use catppuccin-mocha

# Parametric generator for CoolNight variants
nvp theme create --hue 210 --name ocean-blue
nvp theme create --hue 280 --name synthwave-purple
nvp theme create --hue 120 --name matrix-green
```

### From Library (Manual Install)

```bash
# See available themes
nvp theme list

# Apply from file for customization
nvp apply -f my-theme.yaml
```

### From File

```bash
nvp apply -f my-theme.yaml
```

---

## Using Themes

### Set Active Theme

```bash
nvp theme use coolnight-ocean
nvp theme use tokyonight-night
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

**34+ embedded themes** are automatically available without installation, including parametric CoolNight variants and popular themes.

### CoolNight Collection (21 themes)

The **CoolNight Collection** features 21 parametrically generated themes optimized for extended coding sessions. Each variant uses scientifically calibrated color relationships for consistent readability and reduced eye strain.

**Popular variants:**
- `coolnight-ocean` (210°) - Professional blue, great default
- `coolnight-synthwave` (280°) - Retro neon purple  
- `coolnight-matrix` (120°) - High-contrast green
- `coolnight-sunset` (30°) - Warm orange
- `coolnight-mono-slate` - Minimalist grayscale

**Quick usage:**
```bash
# Try popular variants
nvp theme use coolnight-ocean
nvp theme use coolnight-synthwave
nvp theme use coolnight-matrix

# Create custom variant
nvp theme create --hue 165 --name coolnight-teal
```

**→ [Complete CoolNight Documentation](coolnight.md)** - See all 21 variants, color science, and usage recommendations

### Popular Themes (13+ others)

| Theme | Style | Plugin |
|-------|-------|--------|
| `tokyonight-night` | Dark blue | folke/tokyonight.nvim |
| `tokyonight-storm` | Stormy blue | folke/tokyonight.nvim |
| `tokyonight-day` | Light blue | folke/tokyonight.nvim |
| `catppuccin-mocha` | Dark pastel | catppuccin/nvim |
| `catppuccin-latte` | Light pastel | catppuccin/nvim |
| `catppuccin-frappe` | Medium pastel | catppuccin/nvim |
| `catppuccin-macchiato` | Dark warm pastel | catppuccin/nvim |
| `gruvbox-dark` | Warm retro | ellisonleao/gruvbox.nvim |
| `gruvbox-light` | Light retro | ellisonleao/gruvbox.nvim |
| `nord` | Arctic blue | shaunsingh/nord.nvim |
| `dracula` | Dark purple | Mofiqul/dracula.nvim |
| `one-dark` | Dark blue | navarasu/onedark.nvim |
| `solarized-dark` | Dark blue-green | ishan9299/nvim-solarized-lua |

### Parametric Generator

Create custom CoolNight variants with any hue:

```bash
# Create custom themes
nvp theme create --hue 210 --name my-blue-theme
nvp theme create --hue 350 --name my-rose-theme
nvp theme create --hue 120 --name my-green-theme

# Use immediately
nvp theme use my-blue-theme
```

**All themes work immediately** - no installation required!

```bash
# Use any theme directly
nvp theme use coolnight-ocean
nvp theme use dracula
nvp theme use catppuccin-mocha

# Create custom CoolNight variants
nvp theme create --hue 280 --name my-synthwave
```

View details:

```bash
nvp theme get coolnight-ocean       # Theme details
nvp theme list                      # All available themes
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
