# WezTerm Integration

DevOpsMaestro supports generating WezTerm configuration files that integrate with your Neovim theme system.

---

## Overview

WezTerm integration allows you to:

- Generate `wezterm.lua` configuration files
- Apply theme colors from your Neovim theme to WezTerm
- Choose from predefined presets for different workflows
- Keep terminal and editor styling consistent

---

## Available Presets

### minimal
Clean, distraction-free terminal with basic functionality:

```lua
-- Minimal preset features:
-- - Hide title bar
-- - No tabs visible
-- - Simple status line
-- - Theme colors applied
-- - Basic key bindings
```

### tmux-style
Terminal multiplexer-style configuration similar to tmux:

```lua
-- tmux-style preset features:
-- - Bottom status bar with session info
-- - Tab bar visible
-- - tmux-like key bindings (Ctrl-b prefix)
-- - Pane management
-- - Theme colors applied
```

### default
Standard WezTerm configuration with DevOpsMaestro theme integration:

```lua
-- Default preset features:
-- - Standard WezTerm behavior
-- - Tab bar visible
-- - Window decorations
-- - Theme colors applied
-- - Standard key bindings
```

---

## Generating WezTerm Configuration

### Basic Usage

```bash
# Generate wezterm.lua with current theme and default preset
nvp wezterm generate

# Generate with specific preset
nvp wezterm generate --preset minimal
nvp wezterm generate --preset tmux-style
nvp wezterm generate --preset default

# Generate with specific theme
nvp wezterm generate --theme coolnight-ocean
nvp wezterm generate --theme gruvbox-dark --preset minimal

# Generate to specific location
nvp wezterm generate --output ~/.config/wezterm/wezterm.lua
```

### Integration with Theme Hierarchy

WezTerm generation automatically uses the resolved theme from the DevOpsMaestro hierarchy:

```bash
# Set theme at app level
dvm set theme coolnight-synthwave --app my-project

# Generate WezTerm config using the hierarchical theme
cd ~/projects/my-project/workspace
nvp wezterm generate --preset minimal

# The generated config will use coolnight-synthwave colors
```

---

## Generated Configuration Structure

The generated `wezterm.lua` file includes:

```lua
local wezterm = require 'wezterm'
local config = {}

-- Theme colors (from your Neovim theme)
local theme = {
  background = "#1a1b26",
  foreground = "#c0caf5",
  cursor_bg = "#c0caf5",
  cursor_border = "#c0caf5",
  -- ... additional colors
}

-- Preset configuration
-- (varies based on chosen preset)

-- Apply theme colors
config.colors = {
  foreground = theme.foreground,
  background = theme.background,
  cursor_bg = theme.cursor_bg,
  cursor_border = theme.cursor_border,
  cursor_fg = theme.background,
  -- ... full color palette
}

return config
```

---

## Preset Details

### Minimal Preset

Perfect for focused development work:

```lua
-- Hide window decorations
config.window_decorations = "RESIZE"

-- No tab bar
config.enable_tab_bar = false

-- Simple appearance
config.window_padding = {
  left = 8,
  right = 8,
  top = 8,
  bottom = 8,
}

-- Basic key bindings
config.keys = {
  {key="t", mods="CMD", action=wezterm.action{SpawnTab="CurrentPaneDomain"}},
  {key="w", mods="CMD", action=wezterm.action{CloseCurrentTab={confirm=true}}},
  -- ... minimal key set
}
```

### tmux-style Preset

For users who prefer terminal multiplexer workflows:

```lua
-- Show tab bar at bottom
config.tab_bar_at_bottom = true
config.enable_tab_bar = true

-- Status bar configuration
config.status_update_interval = 1000

-- tmux-like key bindings with Ctrl-b prefix
local act = wezterm.action
config.leader = {key="b", mods="CTRL", timeout_milliseconds=1000}
config.keys = {
  -- Pane splitting
  {key="|", mods="LEADER", action=act{SplitHorizontal={domain="CurrentPaneDomain"}}},
  {key="-", mods="LEADER", action=act{SplitVertical={domain="CurrentPaneDomain"}}},
  
  -- Pane navigation
  {key="h", mods="LEADER", action=act{ActivatePaneDirection="Left"}},
  {key="j", mods="LEADER", action=act{ActivatePaneDirection="Down"}},
  {key="k", mods="LEADER", action=act{ActivatePaneDirection="Up"}},
  {key="l", mods="LEADER", action=act{ActivatePaneDirection="Right"}},
  
  -- ... additional tmux-style bindings
}
```

### Default Preset

Standard WezTerm configuration with theme integration:

```lua
-- Standard window decorations
config.window_decorations = "TITLE | RESIZE"

-- Tab bar visible at top
config.enable_tab_bar = true
config.tab_bar_at_bottom = false

-- Standard key bindings
config.keys = {
  {key="t", mods="CMD", action=wezterm.action{SpawnTab="CurrentPaneDomain"}},
  {key="w", mods="CMD", action=wezterm.action{CloseCurrentTab={confirm=true}}},
  {key="n", mods="CMD", action=wezterm.action{SpawnWindow}},
  -- ... full standard key set
}
```

---

## Theme Color Mapping

DevOpsMaestro maps Neovim theme colors to appropriate WezTerm colors:

| Neovim Color | WezTerm Color | Usage |
|--------------|---------------|--------|
| `bg` | `background` | Terminal background |
| `fg` | `foreground` | Default text color |
| `accent` | `cursor_bg`, `cursor_border` | Cursor colors |
| `comment` | `ansi[8]` | Bright black |
| `keyword` | `ansi[5]` | Magenta |
| `string` | `ansi[2]` | Green |
| `function` | `ansi[4]` | Blue |
| `variable` | `foreground` | Default text |
| `type` | `ansi[6]` | Cyan |
| `constant` | `ansi[3]` | Yellow |
| `error` | `ansi[1]` | Red |
| `warning` | `ansi[3]` | Yellow |
| `info` | `ansi[4]` | Blue |
| `selection` | `selection_bg` | Text selection |

---

## Automatic Theme Updates

WezTerm configuration automatically updates when you change themes:

```bash
# Change theme in DevOpsMaestro
dvm set theme coolnight-matrix --app

# Regenerate WezTerm config with new theme
nvp wezterm generate --preset minimal

# WezTerm will automatically reload the configuration
```

---

## Custom Configuration

### Extending Generated Config

You can extend the generated configuration:

```lua
-- ~/.config/wezterm/wezterm.lua
local wezterm = require 'wezterm'
local config = {}

-- Include generated theme and preset config
-- (generated content will be here)

-- Add your custom settings
config.font_size = 14.0
config.font = wezterm.font('JetBrains Mono', {weight='Medium'})

-- Custom key bindings
table.insert(config.keys, {
  key='r',
  mods='CMD|SHIFT',
  action=wezterm.action.ReloadConfiguration,
})

return config
```

### Theme Overrides

Override specific theme colors:

```bash
# Generate base config
nvp wezterm generate --preset minimal

# Then manually edit wezterm.lua to override colors
```

```lua
-- In the generated file, modify the theme table:
local theme = {
  background = "#1a1b26",  -- Keep original
  foreground = "#c0caf5",  -- Keep original
  cursor_bg = "#ff0000",   -- Override to red
  -- ... other colors
}
```

---

## Usage Workflows

### Development Setup

```bash
# 1. Set up your development hierarchy
dvm create ecosystem my-platform
dvm create domain backend --ecosystem my-platform
dvm create app user-service --domain backend

# 2. Set theme at app level
dvm set theme coolnight-ocean --app user-service

# 3. Generate consistent terminal config
cd ~/projects/user-service/workspace
nvp wezterm generate --preset minimal

# 4. Generate Neovim config
nvp generate

# Now terminal and editor use matching colors
```

### Team Consistency

```bash
# 1. Set team theme at domain level
dvm set theme company-theme --domain platform-team

# 2. Generate team WezTerm config
nvp wezterm generate --preset tmux-style --output ~/team-config/wezterm.lua

# 3. Share the generated config file
git add ~/team-config/wezterm.lua
git commit -m "Add team WezTerm config"

# Team members can copy the config file
cp ~/team-config/wezterm.lua ~/.config/wezterm/
```

---

## Troubleshooting

### WezTerm Not Loading Configuration

1. **Check file location:**
   ```bash
   ls -la ~/.config/wezterm/wezterm.lua
   ```

2. **Verify syntax:**
   ```bash
   # Test the configuration
   wezterm start --config ~/.config/wezterm/wezterm.lua
   ```

3. **Check WezTerm logs:**
   ```bash
   wezterm show-config
   ```

### Colors Not Matching Neovim

1. **Verify theme is active in nvp:**
   ```bash
   nvp theme get
   ```

2. **Regenerate with specific theme:**
   ```bash
   nvp wezterm generate --theme $(nvp theme get --name-only)
   ```

3. **Check theme color values:**
   ```bash
   nvp theme get coolnight-ocean -o yaml | grep colors: -A 20
   ```

---

## Supported WezTerm Versions

- **Minimum:** WezTerm 20230712-072601-f4abf8fd
- **Recommended:** Latest stable release
- **Features used:** Color schemes, key bindings, tab configuration

---

## Next Steps

- [Themes Documentation](../nvp/themes.md) - Available themes and variants
- [Theme Hierarchy](theme-hierarchy.md) - Setting themes at different levels
- [nvp Commands](../nvp/commands.md) - Full nvp command reference