# DevOpsMaestro Neovim Templates

This directory contains **quick-start templates** for initializing your local Neovim configuration.

## Quick Start

```bash
# Use one of these templates
dvm nvim init minimal
dvm nvim init kickstart --git-clone
dvm nvim init lazyvim --git-clone
dvm nvim init astronvim --git-clone
```

## Available Templates

### 📦 Built-in Templates (Local)

#### **minimal** (recommended for beginners)
Simple, well-commented configuration created by DevOpsMaestro.
- **Size:** ~100 lines
- **Plugins:** lazy.nvim, catppuccin theme, vim-sleuth
- **Setup time:** < 1 minute
- **Best for:** Learning Neovim, customizing from scratch

#### **starter** (batteries included)
Opinionated config with essential plugins and LSP support.
- **Size:** ~300 lines  
- **Plugins:** lazy.nvim, telescope, treesitter, LSP, catppuccin
- **Setup time:** 2-3 minutes
- **Best for:** Productive coding immediately

### 🌐 Upstream Templates (Git Clone)

#### **kickstart** (kickstart.nvim)
Official Neovim foundation project. Minimal but complete.
- **Source:** https://github.com/nvim-lua/kickstart.nvim
- **Philosophy:** Understand every line
- **Best for:** Learning Neovim Lua API

#### **lazyvim** (LazyVim)
Feature-rich, pre-configured IDE-like experience.
- **Source:** https://github.com/LazyVim/starter
- **Philosophy:** Batteries included
- **Best for:** Modern IDE features out of the box

#### **astronvim** (AstroNvim)
Beautiful, fully-featured configuration with easy customization.
- **Source:** https://github.com/AstroNvim/template
- **Philosophy:** Aesthetic + functional
- **Best for:** Polished experience with great UI

### 🎨 Custom Templates

Use your own or community templates:

```bash
dvm nvim init custom --git-url https://github.com/yourusername/nvim-config.git
```

## Template Structure

Local templates follow this structure:

```
templates/
├── minimal/
│   ├── init.lua          # Main configuration
│   ├── lua/
│   │   └── config/
│   │       ├── options.lua
│   │       ├── keymaps.lua
│   │       └── lazy.lua
│   └── README.md
└── starter/
    ├── init.lua
    ├── lua/
    │   └── plugins/
    │       ├── colorscheme.lua
    │       ├── telescope.lua
    │       ├── treesitter.lua
    │       └── lsp.lua
    └── README.md
```

## Full Examples

For more comprehensive, production-ready configurations, see:

**[devopsmaestro-nvim-templates](https://github.com/rmkohlman/devopsmaestro-nvim-templates)** 🎯

Contains:
- Complete workspace-optimized configs
- Language-specific setups (Go, Python, JavaScript, Rust, etc.)
- DevOps tooling integration (Docker, Kubernetes, Terraform)
- Advanced LSP configurations
- Custom plugins and extensions
- Dotfiles integration examples

## Creating Your Own Template

### Method 1: Local Template

Add to this directory and submit a PR:

```bash
templates/mytemplate/
├── init.lua
└── README.md
```

### Method 2: Git Repository

Create a repo with your config and use:

```bash
dvm nvim init custom --git-url https://github.com/yourname/nvim-config.git
```

### Method 3: Contribute to devopsmaestro-nvim-templates

Submit a PR with your configuration to the examples repo.

## Template Guidelines

**Good templates:**
- ✅ Well-commented
- ✅ Minimal plugin dependencies
- ✅ Clear README with features list
- ✅ Work out of the box
- ✅ Easy to customize

**Avoid:**
- ❌ Undocumented configurations
- ❌ Hundreds of plugins
- ❌ Hard-coded paths
- ❌ Opaque magic

## Initialization Options

```bash
# Initialize with overwrite
dvm nvim init minimal --overwrite

# Custom config path
dvm nvim init minimal --config-path ~/my-nvim-config

# Git clone from upstream
dvm nvim init kickstart --git-clone

# Custom Git URL
dvm nvim init custom --git-url https://github.com/you/config.git --git-clone
```

## See Also

- [Neovim Official Docs](https://neovim.io/doc/)
- [kickstart.nvim](https://github.com/nvim-lua/kickstart.nvim)
- [LazyVim](https://www.lazyvim.org/)
- [AstroNvim](https://astronvim.com/)
- [DevOpsMaestro Docs](../../docs/)

---

**Tip:** After initialization, run `nvim` and wait for plugins to install. Restart Neovim when done.
