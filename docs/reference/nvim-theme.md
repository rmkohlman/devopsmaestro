# NvimTheme YAML Reference

**Kind:** `NvimTheme`  
**APIVersion:** `devopsmaestro.io/v1`

An NvimTheme represents a Neovim colorscheme theme that can be applied and shared via Infrastructure as Code.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: coolnight-synthwave
  description: "CoolNight Synthwave - Retro neon vibes with deep purples and electric blues"
  author: "devopsmaestro"
  category: "dark"
spec:
  plugin:
    repo: "rmkohlman/coolnight.nvim"
    branch: "main"
    tag: "v1.0.0"
  style: "synthwave"
  transparent: false
  colors:
    bg: "#0a0a0a"
    fg: "#e1e1e6"
    primary: "#bd93f9"
    secondary: "#ff79c6"
    accent: "#8be9fd"
    error: "#ff5555"
    warning: "#f1fa8c"
    info: "#8be9fd"
    hint: "#50fa7b"
    selection: "#44475a"
    comment: "#6272a4"
    cursor: "#f8f8f2"
    line_number: "#6272a4"
    line_highlight: "#282a36"
    popup_bg: "#282a36"
    popup_border: "#6272a4"
    statusline_bg: "#44475a"
    tabline_bg: "#282a36"
  options:
    italic_comments: true
    bold_keywords: false
    underline_errors: true
    transparent_background: false
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `NvimTheme` |
| `metadata.name` | string | ✅ | Unique name for the theme |
| `metadata.description` | string | ❌ | Theme description |
| `metadata.author` | string | ❌ | Theme author |
| `metadata.category` | string | ❌ | Theme category (dark/light/both) |
| `spec.plugin` | object | ❌ | Plugin repository information (omit for standalone themes) |
| `spec.plugin.repo` | string | ❌ | GitHub repository (required for plugin-based themes) |
| `spec.plugin.branch` | string | ❌ | Git branch |
| `spec.plugin.tag` | string | ❌ | Git tag/version |
| `spec.style` | string | ❌ | Theme style/variant (plugin-specific) |
| `spec.transparent` | boolean | ❌ | Enable transparent background |
| `spec.colors` | object | ❌ | Semantic color overrides (hex values) |
| `spec.promptColors` | object | ❌ | Starship prompt segment color overrides (hex values) |
| `spec.options` | object | ❌ | Plugin-specific key-value options |

## Field Details

### metadata.name (required)
The unique identifier for the theme.

**Naming conventions:**
- Use kebab-case: `coolnight-synthwave`
- Include collection: `coolnight-ocean`, `tokyonight-night`
- Be descriptive: `gruvbox-material-dark`

### metadata.category (optional)
Theme category for organization and filtering.

**Valid values:**
- `dark` - Dark theme
- `light` - Light theme  
- `both` - Theme with both variants
- `monochrome` - Black and white theme

### spec.plugin (optional)
Plugin repository that provides the colorscheme. Omit entirely for standalone themes — themes without a plugin repo apply colors directly via `vim.api.nvim_set_hl()`.

```yaml
spec:
  plugin:
    repo: "folke/tokyonight.nvim"      # GitHub repository (required for plugin-based themes)
    branch: "main"                     # Git branch (optional)
    tag: "v1.0.0"                     # Git tag/version (optional)
```

**Popular theme plugins:**
- `folke/tokyonight.nvim` - Tokyo Night themes
- `catppuccin/nvim` - Catppuccin themes  
- `ellisonleao/gruvbox.nvim` - Gruvbox themes
- `shaunsingh/nord.nvim` - Nord theme
- `Mofiqul/dracula.nvim` - Dracula theme

### spec.style (optional)
Theme style or variant name (plugin-specific).

```yaml
# Tokyo Night variants
spec:
  style: "night"    # or "storm", "day", "moon"

# Catppuccin variants  
spec:
  style: "mocha"    # or "macchiato", "frappe", "latte"

# Gruvbox variants
spec:
  style: "dark"     # or "light"
```

### spec.transparent (optional)
Enable transparent background for terminal integration.

```yaml
spec:
  transparent: true   # Enable transparent background
```

### spec.colors (optional)
Custom color overrides for semantic color names.

```yaml
spec:
  colors:
    # Basic colors
    bg: "#1a1b26"           # Background color
    fg: "#c0caf5"           # Foreground color
    
    # Semantic colors
    primary: "#7aa2f7"       # Primary accent
    secondary: "#bb9af7"     # Secondary accent
    accent: "#7dcfff"        # Tertiary accent
    
    # Status colors
    error: "#f7768e"         # Error messages
    warning: "#e0af68"       # Warning messages
    info: "#7dcfff"          # Info messages
    hint: "#1abc9c"          # Hint messages
    success: "#9ece6a"       # Success messages
    
    # UI colors
    selection: "#33467c"     # Selection highlight
    comment: "#565f89"       # Comments
    cursor: "#c0caf5"        # Cursor color
    line_number: "#3b4261"   # Line numbers
    line_highlight: "#1f2335" # Current line highlight
    
    # Popup/float colors
    popup_bg: "#1f2335"      # Popup background
    popup_border: "#27a1b9"  # Popup border
    
    # Statusline colors
    statusline_bg: "#1f2335" # Statusline background
    statusline_fg: "#c0caf5" # Statusline foreground
    
    # Tabline colors
    tabline_bg: "#1a1b26"    # Tabline background
    tabline_fg: "#565f89"    # Inactive tabs
    tabline_sel: "#7aa2f7"   # Active tab
```

### spec.options (optional)
Plugin-specific key-value options passed to the theme's `setup()` call. Keys and valid values are entirely plugin-defined.

```yaml
spec:
  options:
    # Typography options (plugin-specific — these are examples for common themes)
    italic_comments: true
    bold_keywords: false
    underline_errors: true
    
    # Background options
    transparent_background: false
    dim_inactive: false
    
    # Plugin integrations (Catppuccin example)
    integrations:
      telescope: true
      nvim_tree: true
      gitsigns: true
      lualine: true
```

### spec.promptColors (optional)
Color overrides for Starship prompt segment colors. These are separate from the Neovim editor colors in `spec.colors` and are applied when the theme is used with a Starship-based terminal prompt.

```yaml
spec:
  promptColors:
    directory: "#7aa2f7"    # Prompt directory segment color
    git_branch: "#9ece6a"   # Git branch segment color
    username: "#bb9af7"     # Username segment color
```

## Theme Collections

### CoolNight Collection

DevOpsMaestro includes 21 CoolNight theme variants:

```yaml
# Ocean (default)
metadata:
  name: coolnight-ocean
spec:
  plugin:
    repo: "rmkohlman/coolnight.nvim"
  style: "ocean"

# Synthwave  
metadata:
  name: coolnight-synthwave
spec:
  style: "synthwave"

# Arctic
metadata:
  name: coolnight-arctic
spec:
  style: "arctic"
```

### Popular Themes

```yaml
# Tokyo Night
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: tokyonight-night
  category: dark
spec:
  plugin:
    repo: "folke/tokyonight.nvim"
  style: "night"

# Catppuccin
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: catppuccin-mocha
  category: dark
spec:
  plugin:
    repo: "catppuccin/nvim"
  style: "mocha"

# Gruvbox
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: gruvbox-dark
  category: dark
spec:
  plugin:
    repo: "ellisonleao/gruvbox.nvim"
  style: "dark"
```

## Usage Examples

### Create Custom Theme

```bash
# From YAML file
dvm apply -f my-theme.yaml

# From URL
dvm apply -f https://themes.example.com/synthwave.yaml

# From GitHub
dvm apply -f github:user/themes/my-theme.yaml
```

### Set Theme

```bash
# Set at workspace level
dvm set theme coolnight-synthwave --workspace dev

# Set at app level
dvm set theme tokyonight-night --app my-api

# Set at domain level  
dvm set theme gruvbox-dark --domain backend
```

### List Themes

```bash
# List all available themes
dvm get nvim themes

# List themes by category
dvm get nvim themes --category dark

# Search themes
dvm get nvim themes --name "*coolnight*"
```

### Export Theme

```bash
# Export to YAML
dvm get nvim theme coolnight-ocean -o yaml

# Export for sharing
dvm get nvim theme my-custom-theme -o yaml > my-theme.yaml
```

## Color Guidelines

### Color Naming

Use semantic color names for better maintainability:

```yaml
colors:
  # Prefer semantic names
  primary: "#7aa2f7"      # ✅ Good
  error: "#f7768e"        # ✅ Good
  
  # Avoid generic names  
  blue: "#7aa2f7"         # ❌ Avoid
  red: "#f7768e"          # ❌ Avoid
```

### Color Accessibility

Ensure sufficient contrast for accessibility:

```yaml
colors:
  bg: "#1a1b26"           # Dark background
  fg: "#c0caf5"           # Light foreground (good contrast)
  comment: "#565f89"      # Muted but readable
```

### Color Consistency

Maintain consistent color usage across themes:

```yaml
colors:
  error: "#ff5555"        # Always red tones
  warning: "#f1fa8c"      # Always yellow tones
  info: "#8be9fd"         # Always blue/cyan tones
  success: "#50fa7b"      # Always green tones
```

## Related Resources

- [Workspace](workspace.md) - Apply themes to workspaces
- [App](app.md) - Set default app themes
- [Domain](domain.md) - Set domain-wide themes
- [Ecosystem](ecosystem.md) - Set ecosystem themes

## Validation Rules

- `metadata.name` must be unique across all themes
- `metadata.name` must be a valid DNS subdomain
- `spec.plugin.repo` must be a valid GitHub repository format
- `spec.colors.*` must be valid hex colors (`#rrggbb` or `#rgb`)
- `spec.promptColors.*` must be valid hex colors (`#rrggbb` or `#rgb`)
- Standalone themes (no `spec.plugin.repo`) must define `spec.colors`
- Theme names must not conflict with built-in themes