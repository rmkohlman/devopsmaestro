# WeztermConfig YAML Reference

**Kind:** `WeztermConfig`  
**APIVersion:** `devopsmaestro.dev/v1alpha1`

A WeztermConfig represents a WezTerm terminal configuration that can be applied and managed through DevOpsMaestro.

## Full Example

```yaml
apiVersion: devopsmaestro.dev/v1alpha1
kind: WeztermConfig
metadata:
  name: synthwave-config
  description: "Synthwave-themed WezTerm configuration with custom keybinds"
  category: "dark"
  tags: ["synthwave", "retro", "neon"]
  labels:
    theme: "synthwave"
    maintainer: "devopsmaestro"
  annotations:
    version: "1.0.0"
    last-updated: "2026-02-19"
spec:
  font:
    family: "JetBrainsMono Nerd Font"
    size: 14.0
    weight: "Regular"
  window:
    opacity: 0.95
    blur: true
    decorations: "RESIZE"
    padding:
      left: 8
      right: 8
      top: 8
      bottom: 8
  colors:
    foreground: "#e1e1e6"
    background: "#0a0a0a"
    cursor_bg: "#bd93f9"
    cursor_fg: "#0a0a0a"
    selection_bg: "#44475a"
    selection_fg: "#f8f8f2"
    ansi:
      - "#21222c"  # black
      - "#ff5555"  # red
      - "#50fa7b"  # green
      - "#f1fa8c"  # yellow
      - "#bd93f9"  # blue
      - "#ff79c6"  # magenta
      - "#8be9fd"  # cyan
      - "#f8f8f2"  # white
    brights:
      - "#6272a4"  # bright black
      - "#ff6e6e"  # bright red
      - "#69ff94"  # bright green
      - "#ffffa5"  # bright yellow
      - "#d6acff"  # bright blue
      - "#ff92df"  # bright magenta
      - "#a4ffff"  # bright cyan
      - "#ffffff"  # bright white
  themeRef: "coolnight-synthwave"  # Alternative to colors
  leader:
    key: "a"
    mods: "CTRL"
    timeout: 1000
  keys:
    - key: "c"
      mods: "LEADER"
      action: "SpawnTab"
      args: ["CurrentPaneDomain"]
    - key: "x"
      mods: "LEADER"
      action: "CloseCurrentTab"
      args: ["confirm:true"]
    - key: "v"
      mods: "LEADER"
      action: "SplitHorizontal"
      args: ["CurrentPaneDomain"]
    - key: "h"
      mods: "LEADER"
      action: "SplitVertical"
      args: ["CurrentPaneDomain"]
    - key: "LeftArrow"
      mods: "LEADER"
      action: "ActivatePaneDirection"
      args: ["Left"]
    - key: "RightArrow"
      mods: "LEADER"
      action: "ActivatePaneDirection"
      args: ["Right"]
    - key: "UpArrow"
      mods: "LEADER"
      action: "ActivatePaneDirection"
      args: ["Up"]
    - key: "DownArrow"
      mods: "LEADER"
      action: "ActivatePaneDirection"
      args: ["Down"]
    - key: "r"
      mods: "LEADER"
      action: "ReloadConfiguration"
    - key: "Enter"
      mods: "ALT"
      action: "ToggleFullScreen"
  tabBar:
    enabled: true
    position: "bottom"
    fancyTabBar: true
    style:
      newTabButton: ">"
      activeTabLeftEdge: ""
      activeTabRightEdge: ""
      inactiveTabLeftEdge: ""
      inactiveTabRightEdge: ""
  scrollback: 10000
  workspace: "default"
  shell:
    program: "/bin/zsh"
    args: ["-l"]
  environment:
    EDITOR: "nvim"
    COLORTERM: "truecolor"
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.dev/v1alpha1` |
| `kind` | string | ✅ | Must be `WeztermConfig` |
| `metadata.name` | string | ✅ | Unique name for the configuration |
| `metadata.description` | string | ❌ | Configuration description |
| `metadata.category` | string | ❌ | Configuration category |
| `metadata.tags` | array | ❌ | Tags for organization |
| `metadata.labels` | object | ❌ | Key-value labels |
| `metadata.annotations` | object | ❌ | Key-value annotations |
| `spec.font` | object | ✅ | Font configuration |
| `spec.window` | object | ❌ | Window appearance settings |
| `spec.colors` | object | ❌ | Color scheme definition |
| `spec.themeRef` | string | ❌ | Reference to theme library |
| `spec.leader` | object | ❌ | Leader key configuration |
| `spec.keys` | array | ❌ | Key binding definitions |
| `spec.tabBar` | object | ❌ | Tab bar configuration |
| `spec.scrollback` | integer | ❌ | Scrollback buffer size |
| `spec.workspace` | string | ❌ | Default workspace |
| `spec.shell` | object | ❌ | Default shell configuration |
| `spec.environment` | object | ❌ | Environment variables |

## Field Details

### metadata.name (required)
The unique identifier for the WezTerm configuration.

**Examples:**
- `synthwave-config`
- `minimal-setup`
- `developer-workspace`

### metadata.category (optional)
Configuration category for organization.

**Valid values:**
- `dark` - Dark theme configurations
- `light` - Light theme configurations
- `minimal` - Minimalist setups
- `feature-rich` - Full-featured configurations

### spec.font (required)
Font family and sizing configuration.

```yaml
spec:
  font:
    family: "JetBrainsMono Nerd Font"  # Font family name (required)
    size: 14.0                         # Font size in points (required)
    weight: "Regular"                  # Font weight (optional)
    italic: false                      # Italic font (optional)
```

**Popular programming fonts:**
- `"JetBrainsMono Nerd Font"`
- `"FiraCode Nerd Font"`
- `"CascadiaCode"`
- `"Menlo"`
- `"SF Mono"`

### spec.window (optional)
Window appearance and behavior settings.

```yaml
spec:
  window:
    opacity: 0.95                      # Window opacity (0.0-1.0)
    blur: true                         # Enable background blur
    decorations: "RESIZE"              # Window decorations
    padding:                           # Window padding
      left: 8
      right: 8
      top: 8
      bottom: 8
    initialRows: 24                    # Initial window height
    initialCols: 80                    # Initial window width
```

**Decoration options:**
- `"TITLE"` - Title bar only
- `"RESIZE"` - Resizable borders
- `"NONE"` - No decorations
- `"INTEGRATED_BUTTONS"` - Integrated title bar

### spec.colors (optional)
Complete color scheme definition.

```yaml
spec:
  colors:
    foreground: "#c0caf5"              # Default text color
    background: "#1a1b26"              # Background color
    cursor_bg: "#c0caf5"               # Cursor background
    cursor_fg: "#1a1b26"               # Cursor foreground
    selection_bg: "#33467c"            # Selection background
    selection_fg: "#c0caf5"            # Selection foreground
    
    # ANSI colors (0-7)
    ansi:
      - "#15161e"  # black (0)
      - "#f7768e"  # red (1)
      - "#9ece6a"  # green (2)
      - "#e0af68"  # yellow (3)
      - "#7aa2f7"  # blue (4)
      - "#bb9af7"  # magenta (5)
      - "#7dcfff"  # cyan (6)
      - "#a9b1d6"  # white (7)
    
    # Bright colors (8-15)
    brights:
      - "#414868"  # bright black (8)
      - "#f7768e"  # bright red (9)
      - "#9ece6a"  # bright green (10)
      - "#e0af68"  # bright yellow (11)
      - "#7aa2f7"  # bright blue (12)
      - "#bb9af7"  # bright magenta (13)
      - "#7dcfff"  # bright cyan (14)
      - "#c0caf5"  # bright white (15)
```

### spec.themeRef (optional)
Alternative to `spec.colors` - references a theme from the library.

```yaml
spec:
  themeRef: "coolnight-synthwave"      # Use library theme instead of colors
```

**Built-in themes:**
- `coolnight-ocean`
- `coolnight-synthwave`
- `tokyonight-night`
- `catppuccin-mocha`
- `gruvbox-dark`

### spec.leader (optional)
Leader key configuration for key sequences.

```yaml
spec:
  leader:
    key: "a"                           # Leader key
    mods: "CTRL"                       # Modifier keys
    timeout: 1000                      # Timeout in milliseconds
```

**Modifier options:**
- `"CTRL"` - Control key
- `"ALT"` - Alt key
- `"SHIFT"` - Shift key
- `"CMD"` - Command key (macOS)
- `"SUPER"` - Super/Windows key

### spec.keys (optional)
Key binding definitions for shortcuts and actions.

```yaml
spec:
  keys:
    - key: "c"                         # Key to bind
      mods: "LEADER"                   # Modifier keys
      action: "SpawnTab"               # Action to execute
      args: ["CurrentPaneDomain"]      # Action arguments
    - key: "Enter"
      mods: "ALT"
      action: "ToggleFullScreen"
```

**Common actions:**
- `SpawnTab` - Create new tab
- `CloseCurrentTab` - Close current tab
- `SplitHorizontal` - Split pane horizontally
- `SplitVertical` - Split pane vertically
- `ActivatePaneDirection` - Navigate panes
- `ReloadConfiguration` - Reload config
- `ToggleFullScreen` - Toggle fullscreen

### spec.tabBar (optional)
Tab bar appearance and behavior.

```yaml
spec:
  tabBar:
    enabled: true                      # Show tab bar
    position: "bottom"                 # Tab bar position
    fancyTabBar: true                  # Use fancy styling
    style:                             # Tab styling
      newTabButton: "+"
      activeTabLeftEdge: ""
      activeTabRightEdge: ""
      inactiveTabLeftEdge: ""
      inactiveTabRightEdge: ""
```

**Position options:**
- `"top"` - Top of window
- `"bottom"` - Bottom of window

### spec.shell (optional)
Default shell configuration.

```yaml
spec:
  shell:
    program: "/bin/zsh"                # Shell program
    args: ["-l"]                       # Shell arguments
```

### spec.environment (optional)
Environment variables for new shells.

```yaml
spec:
  environment:
    EDITOR: "nvim"                     # Default editor
    COLORTERM: "truecolor"             # Color support
    TERM: "wezterm"                    # Terminal type
```

## Configuration Examples

### Minimal Configuration

```yaml
apiVersion: devopsmaestro.dev/v1alpha1
kind: WeztermConfig
metadata:
  name: minimal
  category: minimal
spec:
  font:
    family: "SF Mono"
    size: 12.0
  themeRef: "coolnight-ocean"
```

### Developer Configuration

```yaml
apiVersion: devopsmaestro.dev/v1alpha1
kind: WeztermConfig
metadata:
  name: developer
  category: feature-rich
spec:
  font:
    family: "JetBrainsMono Nerd Font"
    size: 14.0
  window:
    opacity: 0.95
    padding: { left: 8, right: 8, top: 8, bottom: 8 }
  themeRef: "coolnight-synthwave"
  leader:
    key: "a"
    mods: "CTRL"
  keys:
    - key: "c"
      mods: "LEADER"
      action: "SpawnTab"
    - key: "v"
      mods: "LEADER"
      action: "SplitHorizontal"
  tabBar:
    enabled: true
    fancyTabBar: true
```

### Light Theme Configuration

```yaml
apiVersion: devopsmaestro.dev/v1alpha1
kind: WeztermConfig
metadata:
  name: light-theme
  category: light
spec:
  font:
    family: "SF Mono"
    size: 13.0
  colors:
    foreground: "#383a42"
    background: "#fafafa"
    cursor_bg: "#526fff"
    ansi: ["#383a42", "#e45649", "#50a14f", "#c18401", "#4078f2", "#a626a4", "#0184bc", "#fafafa"]
```

## Usage Examples

### Apply Configuration

```bash
# From YAML file
dvm apply -f wezterm-config.yaml

# From URL
dvm apply -f https://configs.example.com/wezterm.yaml

# From GitHub
dvm apply -f github:user/configs/wezterm.yaml
```

### List Configurations

```bash
# List all WezTerm configs
dvm get wezterm configs

# List by category
dvm get wezterm configs --category dark
```

### Export Configuration

```bash
# Export to YAML
dvm get wezterm config synthwave-config -o yaml

# Export current active config
dvm get wezterm config --active -o yaml > my-config.yaml
```

### Set Active Configuration

```bash
# Set as active configuration
dvm set wezterm synthwave-config

# Apply to current workspace
dvm set wezterm synthwave-config --workspace dev
```

## Theme Integration

WezTerm configurations can reference DevOpsMaestro themes:

```yaml
spec:
  themeRef: "coolnight-synthwave"      # Use theme from library
```

This automatically applies colors from the theme to the terminal, maintaining consistency with Neovim themes.

## Related Resources

- [NvimTheme](nvim-theme.md) - Neovim themes that can be referenced
- [Workspace](workspace.md) - Workspaces can specify WezTerm configs

## Validation Rules

- `metadata.name` must be unique across all WezTerm configurations
- `metadata.name` must be a valid DNS subdomain
- `spec.font.family` and `spec.font.size` are required
- `spec.colors.*` must be valid hex colors (`#rrggbb`)
- `spec.themeRef` must reference an existing theme (if specified)
- `spec.keys[].action` must be valid WezTerm actions
- `spec.window.opacity` must be between 0.0 and 1.0
- Cannot specify both `spec.colors` and `spec.themeRef`