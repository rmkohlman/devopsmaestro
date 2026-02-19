# NvimPlugin YAML Reference

**Kind:** `NvimPlugin`  
**APIVersion:** `devopsmaestro.io/v1`

An NvimPlugin represents a Neovim plugin configuration that can be shared and applied across workspaces.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: "Fuzzy finder over lists"
  category: "navigation"
  tags: ["fuzzy-finder", "telescope", "navigation", "files"]
  labels:
    maintainer: "nvim-telescope"
    language: "lua"
  annotations:
    version: "0.1.4"
    last-updated: "2026-02-19"
    documentation: "https://github.com/nvim-telescope/telescope.nvim"
spec:
  repo: "nvim-telescope/telescope.nvim"
  branch: "master"
  version: "0.1.4"
  lazy: true
  priority: 1000
  event: ["VeryLazy"]
  ft: ["lua", "vim"]
  cmd: ["Telescope", "Tele"]
  keys:
    - key: "<leader>ff"
      mode: "n"
      action: "<cmd>Telescope find_files<cr>"
      desc: "Find files"
    - key: "<leader>fg"
      mode: "n"
      action: "<cmd>Telescope live_grep<cr>"
      desc: "Live grep"
    - key: "<leader>fb"
      mode: ["n", "v"]
      action: "<cmd>Telescope buffers<cr>"
      desc: "Find buffers"
    - key: "<leader>fh"
      mode: "n"
      action: "<cmd>Telescope help_tags<cr>"
      desc: "Find help"
  dependencies:
    - "nvim-lua/plenary.nvim"
    - repo: "nvim-tree/nvim-web-devicons"
      build: ""
      config: false
    - repo: "nvim-telescope/telescope-fzf-native.nvim"
      build: "make"
  build: "make"
  config: |
    require('telescope').setup({
      defaults = {
        file_ignore_patterns = {"node_modules", ".git/"},
        layout_strategy = 'horizontal',
        layout_config = {
          width = 0.95,
          height = 0.85,
          preview_cutoff = 120,
          horizontal = {
            preview_width = 0.6,
          },
        },
        mappings = {
          i = {
            ["<C-u>"] = false,
            ["<C-d>"] = false,
          },
        },
      },
      pickers = {
        find_files = {
          theme = "dropdown",
          previewer = false,
        },
        buffers = {
          theme = "dropdown",
          previewer = false,
        },
      },
      extensions = {
        fzf = {
          fuzzy = true,
          override_generic_sorter = true,
          override_file_sorter = true,
          case_mode = "smart_case",
        },
      },
    })
    
    -- Load extensions
    require('telescope').load_extension('fzf')
  init: |
    -- Set up before plugin loads
    vim.g.telescope_theme = 'dropdown'
  opts:
    defaults:
      prompt_prefix: "🔍 "
      selection_caret: "👉 "
      multi_icon: "📌"
      path_display: ["truncate"]
    extensions:
      fzf:
        fuzzy: true
        override_generic_sorter: true
        override_file_sorter: true
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `NvimPlugin` |
| `metadata.name` | string | ✅ | Unique name for the plugin |
| `metadata.description` | string | ❌ | Plugin description |
| `metadata.category` | string | ❌ | Plugin category |
| `metadata.tags` | array | ❌ | Tags for organization |
| `metadata.labels` | object | ❌ | Key-value labels |
| `metadata.annotations` | object | ❌ | Key-value annotations |
| `spec.repo` | string | ✅ | GitHub repository |
| `spec.branch` | string | ❌ | Git branch |
| `spec.version` | string | ❌ | Git tag/version |
| `spec.priority` | integer | ❌ | Load priority (higher = earlier) |
| `spec.lazy` | boolean | ❌ | Enable lazy loading |
| `spec.event` | array | ❌ | Load on events |
| `spec.ft` | array | ❌ | Load on filetypes |
| `spec.cmd` | array | ❌ | Load on commands |
| `spec.keys` | array | ❌ | Load on key mappings |
| `spec.dependencies` | array | ❌ | Plugin dependencies |
| `spec.build` | string | ❌ | Build command |
| `spec.config` | string | ❌ | Configuration Lua code |
| `spec.init` | string | ❌ | Initialization Lua code |
| `spec.opts` | object | ❌ | Options for plugin setup |

## Field Details

### metadata.name (required)
The unique identifier for the plugin.

**Naming conventions:**
- Use the plugin's common name: `telescope`, `lspconfig`, `treesitter`
- Be descriptive for custom configs: `telescope-custom`, `lsp-golang`

### metadata.category (optional)
Plugin category for organization.

**Common categories:**
- `navigation` - File/buffer navigation
- `lsp` - Language Server Protocol
- `completion` - Code completion
- `syntax` - Syntax highlighting
- `git` - Git integration
- `ui` - User interface enhancements
- `editing` - Text editing features
- `debugging` - Debug support
- `testing` - Test integration

### metadata.tags (optional)
Tags for filtering and searching plugins.

```yaml
metadata:
  tags: ["fuzzy-finder", "telescope", "navigation", "files", "grep"]
```

### spec.repo (required)
GitHub repository in `owner/repo` format.

```yaml
spec:
  repo: "nvim-telescope/telescope.nvim"
```

### spec.branch and spec.version (optional)
Git branch or version/tag to use.

```yaml
spec:
  branch: "master"      # Use specific branch
  # OR
  version: "0.1.4"      # Use specific tag/version
```

### spec.priority (optional)
Load priority for plugins. Higher numbers load earlier.

```yaml
spec:
  priority: 1000        # Load early (useful for colorschemes)
```

### Lazy Loading Configuration

#### spec.lazy (optional)
Enable lazy loading for the plugin.

```yaml
spec:
  lazy: true            # Enable lazy loading
```

#### spec.event (optional)
Load plugin on specific events.

```yaml
spec:
  event: ["BufReadPre", "BufNewFile"]  # Load on file open
  # OR
  event: ["VeryLazy"]                  # Load after startup
```

**Common events:**
- `VeryLazy` - After startup completion
- `BufReadPre` - Before reading a buffer
- `BufNewFile` - On new file creation
- `InsertEnter` - Entering insert mode
- `CmdlineEnter` - Entering command line

#### spec.ft (optional)
Load plugin on specific filetypes.

```yaml
spec:
  ft: ["go", "lua", "python"]          # Load for specific languages
```

#### spec.cmd (optional)
Load plugin when specific commands are used.

```yaml
spec:
  cmd: ["Telescope", "Tele"]           # Load on command usage
```

#### spec.keys (optional)
Load plugin on specific key mappings.

```yaml
spec:
  keys:
    - key: "<leader>ff"                # Key combination
      mode: "n"                        # Mode: n, i, v, x, o, c
      action: "<cmd>Telescope find_files<cr>"  # Action to execute
      desc: "Find files"               # Description
    - key: "<C-p>"
      mode: ["n", "i"]                 # Multiple modes
      action: "<cmd>Telescope find_files<cr>"
      desc: "Find files (Ctrl+P)"
```

### spec.dependencies (optional)
Plugin dependencies that must be loaded first.

```yaml
spec:
  dependencies:
    # Simple string format
    - "nvim-lua/plenary.nvim"
    
    # Detailed format
    - repo: "nvim-tree/nvim-web-devicons"
      build: ""                        # No build command
      config: false                    # Don't auto-configure
    - repo: "nvim-telescope/telescope-fzf-native.nvim"
      build: "make"                    # Build with make
      version: "1.0.0"                # Specific version
```

### spec.build (optional)
Build command to run after plugin installation/update.

```yaml
spec:
  build: "make"                        # Run make
  # OR
  build: ":TSUpdate"                   # Neovim command
  # OR  
  build: "npm install"                 # Shell command
```

### spec.config (optional)
Lua configuration code executed after plugin loads.

```yaml
spec:
  config: |
    require('telescope').setup({
      defaults = {
        file_ignore_patterns = {"node_modules"},
        layout_strategy = 'horizontal',
      }
    })
```

### spec.init (optional)
Lua initialization code executed before plugin loads.

```yaml
spec:
  init: |
    vim.g.telescope_theme = 'dropdown'
    vim.g.telescope_debug = false
```

### spec.opts (optional)
Options passed directly to the plugin's setup function.

```yaml
spec:
  opts:
    defaults:
      prompt_prefix: "🔍 "
      selection_caret: "👉 "
    pickers:
      find_files:
        theme: "dropdown"
```

## Plugin Categories and Examples

### Navigation Plugins

```yaml
# Telescope - Fuzzy finder
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  category: navigation
spec:
  repo: "nvim-telescope/telescope.nvim"
  keys:
    - key: "<leader>ff"
      action: "<cmd>Telescope find_files<cr>"

# Oil.nvim - File explorer
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: oil
  category: navigation
spec:
  repo: "stevearc/oil.nvim"
  keys:
    - key: "-"
      action: "<cmd>Oil<cr>"
```

### LSP Plugins

```yaml
# LSP Config
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: lspconfig
  category: lsp
spec:
  repo: "neovim/nvim-lspconfig"
  event: ["BufReadPre", "BufNewFile"]
  
# Mason (LSP installer)
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: mason
  category: lsp
spec:
  repo: "williamboman/mason.nvim"
  cmd: ["Mason", "MasonInstall"]
```

### Completion Plugins

```yaml
# nvim-cmp
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: nvim-cmp
  category: completion
spec:
  repo: "hrsh7th/nvim-cmp"
  event: "InsertEnter"
  dependencies:
    - "hrsh7th/cmp-nvim-lsp"
    - "hrsh7th/cmp-buffer"
    - "hrsh7th/cmp-path"
```

### Language-Specific Plugins

```yaml
# Go plugin
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: vim-go
  category: language
  tags: ["go", "golang"]
spec:
  repo: "fatih/vim-go"
  ft: ["go"]
  
# Rust plugin  
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: rust-tools
  category: language
  tags: ["rust"]
spec:
  repo: "simrat39/rust-tools.nvim"
  ft: ["rust"]
```

## Usage Examples

### Create Custom Plugin

```bash
# From YAML file
dvm apply -f my-plugin.yaml

# From URL
dvm apply -f https://plugins.example.com/telescope.yaml

# From GitHub
dvm apply -f github:user/configs/telescope.yaml
```

### List Plugins

```bash
# List all plugins
dvm get nvim plugins

# List by category
dvm get nvim plugins --category navigation

# Search plugins
dvm get nvim plugins --name "*telescope*"
```

### Export Plugin

```bash
# Export to YAML
dvm get nvim plugin telescope -o yaml

# Export for sharing
dvm get nvim plugin my-custom-config -o yaml > telescope-config.yaml
```

### Use in Workspace

```yaml
# Reference in workspace
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-app
spec:
  nvim:
    plugins:
      - telescope
      - lspconfig
      - nvim-cmp
```

## Related Resources

- [Workspace](workspace.md) - Use plugins in workspaces
- [NvimPackage](nvim-package.md) - Plugin collections
- [NvimTheme](nvim-theme.md) - Theme plugins

## Validation Rules

- `metadata.name` must be unique across all plugins
- `metadata.name` must be a valid DNS subdomain
- `spec.repo` must be a valid GitHub repository format (`owner/repo`)
- `spec.keys[].mode` must be valid Neovim mode(s)
- `spec.config` and `spec.init` must be valid Lua code
- `spec.dependencies` must reference valid repositories
- `spec.priority` must be a positive integer
- Plugin names must not conflict with built-in plugins