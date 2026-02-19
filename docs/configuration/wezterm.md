# WezTerm Integration

DevOpsMaestro supports generating WezTerm configuration files that integrate with your Neovim theme system through the new `dvt` (DevOps Terminal) command suite.

---

## Overview

WezTerm integration allows you to:

- Generate `wezterm.lua` configuration files with automatic theme integration
- Apply theme colors from your Neovim theme to WezTerm automatically
- Choose from predefined presets for different workflows
- Keep terminal and editor styling perfectly consistent
- Use simple CLI commands for quick setup

---

## Quick Commands

### dvt wezterm Commands

| Command | Description |
|---------|-------------|
| `dvt wezterm list` | List all available presets |
| `dvt wezterm show <preset>` | Display preset configuration details |
| `dvt wezterm generate <preset>` | Generate config file (preview mode) |
| `dvt wezterm apply <preset>` | Generate and save to `~/.wezterm.lua` |

### Basic Usage

```bash
# List available presets
dvt wezterm list

# View preset details
dvt wezterm show minimal
dvt wezterm show tmux-style

# Apply configuration directly
dvt wezterm apply default    # Creates ~/.wezterm.lua
dvt wezterm apply minimal    # Overwrites with minimal preset

# Generate to custom location
dvt wezterm generate minimal --output ~/.config/wezterm/wezterm.lua

# Preview configuration before applying
dvt wezterm generate default  # Shows config without saving
```

### Automatic Theme Integration

**The key benefit:** Colors are automatically resolved from the theme library and current workspace theme settings.

```bash
# Set workspace theme
dvm set theme coolnight-synthwave --workspace main

# Apply WezTerm config - automatically uses coolnight-synthwave colors
dvt wezterm apply minimal

# Terminal now matches your Neovim theme perfectly!
```

---

## Available Presets

### minimal
Clean, distraction-free terminal with basic functionality:

```lua
-- Minimal preset features:
-- - Hide title bar and decorations
-- - No tabs visible
-- - Clean appearance with minimal padding
-- - Theme colors applied automatically
-- - Essential key bindings only
```

### tmux-style
Terminal multiplexer-style configuration similar to tmux:

```lua
-- tmux-style preset features:
-- - Bottom status bar with session info
-- - Tab bar visible with tmux-style navigation
-- - tmux-like key bindings (Ctrl-b prefix)
-- - Pane splitting and management
-- - Theme colors applied automatically
```

### default
Standard WezTerm configuration with DevOpsMaestro theme integration:

```lua
-- Default preset features:
-- - Standard WezTerm behavior
-- - Tab bar visible at top
-- - Window decorations enabled
-- - Theme colors applied automatically
-- - Full standard key binding set
```

---

## Legacy Commands (Still Supported)

For backward compatibility, the original `nvp wezterm` commands still work:

```bash
# Legacy commands (still functional)
nvp wezterm generate --preset minimal
nvp wezterm generate --theme coolnight-ocean --preset default
```

### Integration with Theme Hierarchy

WezTerm generation automatically uses the resolved theme from the DevOpsMaestro hierarchy:

```bash
# Set theme at app level
dvm set theme coolnight-synthwave --app my-project

# Generate WezTerm config using the hierarchical theme
cd ~/projects/my-project/workspace
dvt wezterm apply minimal

# The generated config automatically uses coolnight-synthwave colors
```

### Custom Output Locations

```bash
# Save to custom location
dvt wezterm apply default --output ~/.config/wezterm/wezterm.lua

# Preview before saving
dvt wezterm generate minimal  # Shows config without saving
dvt wezterm apply minimal     # Saves to ~/.wezterm.lua
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
dvt wezterm apply minimal

# WezTerm will automatically reload the configuration
```

### Batch Updates

```bash
# Update all configurations at once after theme change
nvp config generate       # Update Neovim
nvp theme generate        # Update theme files
dvt wezterm apply minimal # Update terminal

# Everything stays in sync automatically
```

---

## Custom Configuration

### Extending Generated Config

You can extend the generated configuration:

```lua
-- ~/.wezterm.lua (generated with dvt wezterm apply)
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
dvt wezterm apply minimal

# Then manually edit ~/.wezterm.lua to override colors
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

### Complete Development Setup

```bash
# 1. Initialize and create workspace
dvm admin init
cd ~/projects/user-service
dvm create app user-service --from-cwd
dvm create workspace main

# 2. Set theme at app level
dvm set theme coolnight-ocean --app user-service

# 3. Generate all configurations with consistent theme
nvp config generate      # Neovim config
nvp theme generate       # Theme files
dvt wezterm apply minimal # Terminal config

# 4. Generate shell profile
dvt profile generate myprofile --output ~/.config

# Now terminal and editor use perfectly matching colors
```

### Team Consistency

```bash
# 1. Set team theme at domain level
dvm create ecosystem company
dvm create domain platform-team
dvm set theme company-theme --domain platform-team

# 2. Generate team WezTerm config
dvt wezterm generate tmux-style --output ~/team-config/wezterm.lua

# 3. Share the generated config file
git add ~/team-config/wezterm.lua
git commit -m "Add team WezTerm config with company theme"

# Team members can apply the config
cp ~/team-config/wezterm.lua ~/.wezterm.lua
```

### Quick Theme Changes

```bash
# Change theme across your entire setup
dvm set theme coolnight-matrix --app

# Update all configurations
nvp theme generate              # Update Neovim theme
dvt wezterm apply minimal       # Update terminal theme

# Everything now uses the new theme consistently
```

---

## Troubleshooting

### WezTerm Not Loading Configuration

1. **Check file location:**
   ```bash
   ls -la ~/.wezterm.lua
   ```

2. **Verify syntax:**
   ```bash
   # Test the configuration
   wezterm start --config ~/.wezterm.lua
   ```

3. **Check WezTerm logs:**
   ```bash
   wezterm show-config
   ```

### Colors Not Matching Neovim

1. **Verify theme is active:**
   ```bash
   dvm get context --show-theme
   ```

2. **Regenerate with current theme:**
   ```bash
   dvt wezterm apply minimal
   ```

3. **Check available presets:**
   ```bash
   dvt wezterm list
   dvt wezterm show minimal
   ```

### Command Not Found: dvt

If `dvt` command is not available, you may be using an older version:

```bash
# Check version
dvm version

# Update to latest
brew upgrade devopsmaestro

# Use legacy commands as fallback
nvp wezterm generate --preset minimal --output ~/.wezterm.lua
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