# Standalone Theme Support

NvimOps now supports "standalone" or "inline" themes that don't require an external Neovim plugin repository. These themes apply colors directly via `vim.api.nvim_set_hl()`.

## How It Works

### Plugin-Based Themes (Traditional)
```yaml
spec:
  plugin:
    repo: "folke/tokyonight.nvim"  # Requires external plugin
```
- Requires cloning an external GitHub repository
- Uses the plugin's setup function and colorscheme command
- Limited to themes that have published plugins

### Standalone Themes (New)
```yaml
spec:
  plugin:
    repo: ""  # Empty repo = standalone theme
  colors:
    bg: "#1a1b26"
    fg: "#c0caf5"
    # ... more colors required
```
- No external dependencies - completely self-contained
- Applies highlights directly using Neovim's API
- Perfect for custom parametric themes like CoolNight

## Example: Standalone CoolNight Ocean

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: coolnight-ocean-standalone
  description: CoolNight Ocean - Standalone theme without external plugin
  author: DevOpsMaestro
  category: dark
spec:
  plugin:
    repo: ""  # This makes it standalone
  transparent: false
  colors:
    # UI colors
    bg: "#011628"
    bg_dark: "#011423"
    bg_highlight: "#143652"
    bg_visual: "#275378"
    fg: "#CBE0F0"
    fg_dark: "#B4D0E9"
    
    # Syntax colors
    red: "#E52E2E"
    green: "#44FFB1"
    blue: "#0FC5ED"
    yellow: "#FFE073"
    purple: "#a277ff"
    cyan: "#24EAF7"
    orange: "#FFE073"
    
    # Diagnostic colors
    error: "#E52E2E"
    warning: "#FFE073"
    info: "#0FC5ED"
    hint: "#44FFB1"
```

## Generated Files

Standalone themes generate these files:

### 1. `plugins/nvp/colorscheme.lua` - Lazy.nvim Plugin Spec
```lua
return {
  dir = vim.fn.stdpath("config") .. "/lua/theme",
  name = "coolnight-ocean-standalone",
  priority = 1000,
  lazy = false,
  config = function()
    require("theme.colorscheme").setup()
  end,
}
```

### 2. `theme/colorscheme.lua` - Self-Contained Colorscheme
```lua
local M = {}

function M.setup(opts)
  local palette = require("theme.palette")
  local colors = palette.colors
  
  -- Clear existing highlights
  vim.cmd("hi clear")
  vim.g.colors_name = "coolnight-ocean-standalone"
  
  -- Apply highlights
  local hl = function(name, val)
    vim.api.nvim_set_hl(0, name, val)
  end
  
  -- UI Groups
  hl("Normal", { fg = colors.fg, bg = colors.bg })
  hl("Comment", { fg = colors.comment, italic = true })
  hl("DiagnosticError", { fg = colors.error })
  -- ... many more highlight groups
end

M.setup() -- Auto-setup on load
return M
```

### 3. Standard Files (Same as Plugin Themes)
- `theme/palette.lua` - Color palette for other plugins
- `theme/init.lua` - Theme utilities and helpers

## Usage

```bash
# Create standalone theme
nvp theme apply -f my-standalone-theme.yaml

# Use it
nvp theme use my-standalone-theme

# Generate Lua files
nvp theme generate
```

## Validation Rules

- **Plugin-based themes**: Must have `spec.plugin.repo`
- **Standalone themes**: Must have empty `spec.plugin.repo` AND must define colors
- **Colors required**: Standalone themes must define at least basic colors (`bg`, `fg`, etc.)

## Highlight Groups Applied

Standalone themes automatically apply highlights for:

### UI Groups
- Normal, NormalFloat, CursorLine, Visual, Search
- StatusLine, TabLine, Pmenu, LineNr
- WinSeparator, VertSplit, SignColumn

### Syntax Groups  
- Comment, String, Number, Boolean, Function
- Keyword, Type, Operator, PreProc, Special

### Diagnostics
- DiagnosticError, DiagnosticWarn, DiagnosticInfo, DiagnosticHint
- Underline variants for each

### Git Integration
- GitSignsAdd, GitSignsChange, GitSignsDelete

### Treesitter
- Links treesitter highlights to standard syntax groups

## Migration from Plugin Themes

To convert a CoolNight theme from plugin-based to standalone:

1. Remove or empty the `spec.plugin.repo` field:
   ```yaml
   spec:
     plugin:
       repo: ""  # Was "rmkohlman/coolnight.nvim"
   ```

2. Ensure adequate colors are defined in `spec.colors`

3. Regenerate: `nvp theme generate`

The theme will now be completely self-contained with no external dependencies.