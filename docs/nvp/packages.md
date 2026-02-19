# Plugin Packages

Plugin packages are curated bundles of Neovim plugins that provide complete, opinionated setups for specific development workflows. Instead of installing plugins one by one, packages give you a complete development environment instantly.

---

## What Are Plugin Packages?

Plugin packages are pre-configured collections of plugins that work well together, providing:

- **Complete setups** - Everything you need for a specific workflow
- **Tested configurations** - All plugins are pre-configured to work together
- **Consistent experience** - Uniform key bindings and behavior across plugins
- **Quick setup** - Go from zero to fully configured in one command

---

## Available Packages

### core (Default Package)

**The default package automatically installed for new workspaces.** Provides essential IDE functionality with minimal overhead:

```bash
# This package is automatically applied to new workspaces
# You can also install it manually:
nvp apply -f package:core
```

**Included plugins (6 essential tools):**
- **nvim-treesitter** - Modern syntax highlighting and code understanding
- **telescope.nvim** - Fuzzy finder for files, grep, buffers, and more
- **which-key.nvim** - Keybinding discovery and help system
- **nvim-lspconfig** - Language Server Protocol configuration
- **nvim-cmp** - Intelligent autocompletion with multiple sources
- **gitsigns.nvim** - Git integration with inline status and blame

This minimal but powerful set gives you a complete development environment while keeping startup time fast and resource usage low.

### rkohlman-full

The flagship package with 39 carefully curated plugins for a complete development environment:

```bash
# Install the complete development environment
nvp apply -f package:rkohlman-full
```

**Included categories:**
- **LSP & Completion** - Full language server support with autocompletion
- **Fuzzy Finding** - Advanced file and content searching
- **Git Integration** - Complete Git workflow support
- **UI Enhancement** - Status line, buffer management, file tree
- **Text Editing** - Advanced text manipulation and formatting
- **Terminal Integration** - Embedded terminal management
- **AI Assistance** - GitHub Copilot integration
- **Language Support** - Rust, Go, TypeScript, and more

---

## Package Contents

### Default Core Package

The `core` package (automatically applied to new workspaces) includes:

| Category | Plugins | Description |
|----------|---------|-------------|
| **Syntax & Navigation** | treesitter, telescope | Modern syntax highlighting and fuzzy finding |
| **LSP & Completion** | lspconfig, nvim-cmp | Language server support and autocompletion |
| **UI & Git** | which-key, gitsigns | Keybinding help and git integration |

### Complete Development Tools (rkohlman-full)

| Category | Plugins | Description |
|----------|---------|-------------|
| **Dependencies** | plenary, nvim-web-devicons | Essential utilities and icons |
| **LSP & Completion** | lspconfig, mason, nvim-cmp, cmp-nvim-lsp, cmp-buffer, cmp-path, luasnip | Complete language server and autocompletion setup |
| **Fuzzy Finding** | telescope, harpoon | Advanced file navigation and quick jumping |
| **Syntax** | treesitter, treesitter-textobjects | Modern syntax highlighting and text objects |

### UI & Interface

| Category | Plugins | Description |
|----------|---------|-------------|
| **Status & Buffers** | lualine, bufferline | Professional status line and buffer management |
| **Navigation** | neo-tree, which-key | File tree and keybinding discovery |
| **Dashboard** | alpha-nvim | Beautiful startup dashboard |
| **Enhancements** | dressing, notify | Better UI for inputs and notifications |

### Git Integration

| Category | Plugins | Description |
|----------|---------|-------------|
| **Git Tools** | gitsigns, fugitive, diffview | Complete Git workflow integration |

### Text Editing

| Category | Plugins | Description |
|----------|---------|-------------|
| **Editing** | comment, surround, autopairs | Enhanced text manipulation |
| **Formatting** | conform, nvim-lint | Code formatting and linting |

### Terminal & AI

| Category | Plugins | Description |
|----------|---------|-------------|
| **Terminal** | toggleterm | Integrated terminal management |
| **AI** | copilot, copilot-cmp | GitHub Copilot integration |

### Language-Specific

| Category | Plugins | Description |
|----------|---------|-------------|
| **Languages** | rust-tools, go, typescript-tools | Specialized support for major languages |

---

## Using Plugin Packages

### Install Complete Package

```bash
# Install the full rkohlman-full package
nvp apply -f package:rkohlman-full

# Generate Neovim configuration
nvp generate

# Start using your fully configured Neovim
nvim
```

### Override Default Package

If you want to replace the default `core` package:

```bash
# Remove default core package
nvp package remove core

# Install different package
nvp apply -f package:rkohlman-full

# Or install individual plugins
nvp apply -f plugin:neo-tree
nvp apply -f plugin:copilot
```

### Package Installation Process

When you install a package, nvp:

1. **Downloads** all 39 plugin definitions
2. **Validates** plugin compatibility
3. **Installs** all plugins to `~/.nvp/plugins/`
4. **Configures** each plugin with optimized settings
5. **Sets up** key bindings and integrations

### Verify Package Installation

```bash
# List all installed plugins (should show 39 plugins)
nvp plugin list

# Check specific plugins from the package
nvp plugin get telescope
nvp plugin get lspconfig
nvp plugin get copilot
```

---

## Package Configuration

### Pre-configured Features

The rkohlman-full package comes with these pre-configured features:

#### LSP (Language Server Protocol)
- **Automatic installation** of language servers via Mason
- **Autocompletion** with multiple sources (LSP, buffer, path)
- **Snippets** support with LuaSnip
- **Diagnostics** display and navigation

#### Fuzzy Finding
- **Telescope** configured for files, grep, buffers, and more
- **Harpoon** for quick file switching
- **Which-key** for command discovery

#### Git Workflow
- **Gitsigns** for inline git status
- **Fugitive** for Git commands
- **Diffview** for reviewing changes

#### UI Enhancements
- **Lualine** status line with git and LSP info
- **Bufferline** for tab-like buffer management
- **Neo-tree** file explorer
- **Alpha-nvim** dashboard

#### Key Bindings

Pre-configured key bindings include:

| Key | Action | Plugin |
|-----|--------|--------|
| `<leader>ff` | Find files | telescope |
| `<leader>fg` | Live grep | telescope |
| `<leader>fb` | Find buffers | telescope |
| `<leader>e` | Toggle file tree | neo-tree |
| `<leader>gg` | Git status | fugitive |
| `<C-\>` | Toggle terminal | toggleterm |
| `<leader>ca` | Code actions | LSP |
| `gd` | Go to definition | LSP |
| `gr` | Go to references | LSP |

---

## Customizing Packages

### Override Individual Plugins

You can override specific plugins from a package:

```bash
# Install the full package
nvp apply -f package:rkohlman-full

# Override telescope configuration
cat > my-telescope.yaml << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
spec:
  repo: nvim-telescope/telescope.nvim
  config: |
    require("telescope").setup({
      defaults = {
        layout_strategy = "horizontal",
        layout_config = {
          width = 0.9,
          height = 0.8,
        },
      },
    })
EOF

# Apply override
nvp apply -f my-telescope.yaml
```

### Disable Plugins from Package

```bash
# Disable a plugin you don't want
nvp plugin delete copilot

# Or disable without deleting
cat > disable-copilot.yaml << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: copilot
spec:
  repo: github/copilot.vim
  enabled: false
EOF

nvp apply -f disable-copilot.yaml
```

### Add Additional Plugins

```bash
# Add plugins not in the package
nvp apply -f plugin:dashboard      # Alternative dashboard
nvp apply -f plugin:obsidian       # Note-taking integration
```

---

## Creating Custom Packages

### Package Definition Format

While the built-in packages cover most needs, you can create custom packages:

```yaml
# my-minimal-package.yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: my-minimal-dev
  description: Minimal development environment
  author: Your Name
  version: "1.0.0"
spec:
  plugins:
    - name: telescope
      source: plugin:telescope
    - name: treesitter
      source: plugin:treesitter
    - name: lspconfig
      source: plugin:lspconfig
    - name: nvim-cmp
      source: plugin:nvim-cmp
  theme:
    default: coolnight-ocean
```

### Package Categories

Consider organizing custom packages by use case:

```yaml
# frontend-package.yaml - For web development
spec:
  plugins:
    - plugin:telescope
    - plugin:treesitter
    - plugin:typescript-tools
    - plugin:tailwind-tools
    - plugin:emmet-vim

# backend-package.yaml - For API development  
spec:
  plugins:
    - plugin:telescope
    - plugin:treesitter
    - plugin:go
    - plugin:rust-tools
    - plugin:database-client

# data-science-package.yaml - For ML/Data work
spec:
  plugins:
    - plugin:telescope
    - plugin:jupyter-nvim
    - plugin:python-tools
    - plugin:r-nvim
    - plugin:markdown-preview
```

---

## Package Management

### Update Package

```bash
# Update to latest version of package
nvp apply -f package:rkohlman-full --update

# This will:
# - Update plugin versions
# - Add new plugins if added to package
# - Update configurations
```

### Package Information

```bash
# Show package details
nvp package get rkohlman-full

# List all available packages
nvp package list

# Show what's installed from a package
nvp plugin list --package rkohlman-full
```

### Remove Package

```bash
# Remove all plugins from a package
nvp package remove rkohlman-full

# Or remove individual plugins
nvp plugin delete telescope
nvp plugin delete lspconfig
# ... etc
```

---

## Integration with Themes

### Package + Theme Workflow

```bash
# 1. Set up development hierarchy
dvm create ecosystem my-company
dvm create domain backend --ecosystem my-company
dvm create app user-service --domain backend

# 2. Set theme at app level
dvm set theme coolnight-ocean --app user-service

# 3. Install complete plugin package
cd ~/projects/user-service/workspace
nvp apply -f package:rkohlman-full

# 4. Generate unified configuration
nvp generate

# Now you have 39 plugins + consistent theme
```

### Theme-Aware Packages

Some packages can adapt to your current theme:

```bash
# Package will use colors from current theme
nvp apply -f package:rkohlman-full

# Lualine, telescope, and other UI plugins
# will automatically use theme colors
```

---

## Performance Considerations

### Lazy Loading

The rkohlman-full package is optimized for performance:

- **Most plugins lazy load** - Only load when needed
- **Event-based triggers** - Load on file types, commands, or events
- **Dependency management** - Load dependencies in correct order
- **Startup optimization** - Core plugins load fast, others load on demand

### Startup Time

With 39 plugins, startup time is still fast due to:

```lua
-- Example lazy loading configuration (built into package)
{
  "nvim-telescope/telescope.nvim",
  lazy = true,
  cmd = "Telescope",
  keys = {
    {"<leader>ff", "<cmd>Telescope find_files<cr>"},
  },
}
```

### Memory Usage

Plugins only consume memory when active:
- **LSP servers** start per file type
- **Git plugins** activate in Git repositories
- **Language tools** load for specific languages

---

## Troubleshooting Packages

### Common Issues

1. **Plugin conflicts:**
   ```bash
   # Check for conflicting plugins
   nvp plugin list --conflicts
   
   # Remove conflicting plugin
   nvp plugin delete conflicting-plugin
   ```

2. **Missing dependencies:**
   ```bash
   # Package installation should handle this automatically
   # But if there are issues:
   nvp apply -f package:rkohlman-full --fix-dependencies
   ```

3. **Configuration errors:**
   ```bash
   # Check Neovim logs for plugin errors
   nvim --headless -c 'checkhealth' -c 'quit'
   
   # Regenerate clean configuration
   rm -rf ~/.config/nvim/lua/plugins/nvp
   nvp generate
   ```

### Getting Help

```bash
# Check plugin status
nvp plugin status

# Validate package installation
nvp package validate rkohlman-full

# Show generated configuration
cat ~/.config/nvim/lua/plugins/nvp/init.lua
```

---

## Future Packages

Additional packages are planned for specific use cases:

- **`minimal-dev`** - Essential plugins only (10-15 plugins)
- **`frontend-full`** - Web development focused
- **`backend-full`** - API and system development
- **`data-science`** - ML and analytics workflow
- **`writer`** - Documentation and writing focused

Request specific packages or contribute your own package definitions to the community.

---

## Next Steps

- [Plugin Documentation](plugins.md) - Individual plugin management
- [Themes](themes.md) - Theme system integration
- [Commands Reference](commands.md) - Full nvp command list
- [Configuration](../configuration/yaml-schema.md) - YAML format details