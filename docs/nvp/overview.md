# nvp Overview

`nvp` (NvimOps) is a standalone Neovim plugin and theme manager using YAML.

---

## What is nvp?

nvp lets you:

- **Define plugins in YAML** instead of Lua
- **Use a curated library** of 38+ pre-configured plugins
- **Manage themes** with 34+ embedded themes and parametric generator
- **Generate Lua files** for lazy.nvim
- **Get default configurations** - new workspaces automatically include the "core" package

### Default Configuration

**New workspaces in dvm automatically get a pre-configured nvim setup** with:
- **Structure**: `lazyvim` framework 
- **Package**: `core` (6 essential plugins)
- **Theme**: Inherits from dvm's theme cascade system

This means you can start coding immediately without manual nvim configuration.

---

## Key Features

- :material-file-code: **YAML-based** - Write plugins in familiar YAML format
- :material-library: **Built-in library** - 38+ curated plugins and plugin packages
- :material-palette: **Theme system** - 34+ themes with parametric generator
- :material-link: **URL support** - Install from GitHub
- :material-package-variant: **Standalone** - No containers required

---

## Quick Start

```bash
# Initialize
nvp init

# Install plugins from library
nvp plugin list
nvp apply -f package:rmkohlman

# List and use themes
nvp theme list
nvp theme create --hue 210 --name my-blue-theme

# Generate Lua files
nvp generate
```

Files are created in `~/.config/nvim/lua/plugins/nvp/`.

---

## How It Works

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   YAML Files    │────▶│   nvp generate  │────▶│   Lua Files     │
│   (plugins)     │     │                 │     │   (lazy.nvim)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

1. You define/install plugins (stored as YAML)
2. Run `nvp generate`
3. Lua files are created for lazy.nvim
4. Neovim loads them on startup

---

## Plugin Library

nvp includes 38+ curated plugins with plugin package system:

### Plugin Packages

| Package | Plugins | Description |
|---------|---------|-------------|
| `rmkohlman` | 39 plugins | Complete development environment |

### Core Plugins (Default Package)

The `core` package (automatically installed for new dvm workspaces):

| Plugin | Category | Description |
|--------|----------|-------------|
| treesitter | syntax | Modern syntax highlighting and code understanding |
| telescope | fuzzy-finder | Fuzzy finder for files, grep, buffers, etc. |
| which-key | ui | Keybinding discovery and help system |
| lspconfig | lsp | Language Server Protocol configuration |
| nvim-cmp | completion | Intelligent autocompletion |
| gitsigns | git | Git integration with inline status |

### Complete Plugin Library (rmkohlman)

| Plugin | Category | Description |
|--------|----------|-------------|
| telescope | fuzzy-finder | Fuzzy finder for files, grep, etc. |
| treesitter | syntax | Advanced syntax highlighting |
| lspconfig | lsp | LSP configuration |
| nvim-cmp | completion | Autocompletion |
| mason | lsp | LSP/DAP/Linter installer |
| gitsigns | git | Git decorations |
| lualine | ui | Status line |
| which-key | ui | Keybinding hints |
| neo-tree | file-explorer | File tree |
| toggleterm | terminal | Terminal management |
| and 28+ more... | | |

See full list:

```bash
nvp plugin list
```

---

## Theme Library

34+ embedded themes with parametric generator:

### CoolNight Variants (21 themes)

| Theme | Hue | Description |
|-------|-----|-------------|
| coolnight-ocean | 210° | Deep blue (default) |
| coolnight-synthwave | 280° | Neon purple |
| coolnight-matrix | 120° | Matrix green |
| coolnight-sunset | 30° | Warm orange |
| coolnight-rose | 350° | Pink/rose |
| coolnight-nord | Nord | Arctic bluish |
| coolnight-dracula | Dracula | Purple variant |
| and 14 more... | | |

### Popular Themes (13+ others)

| Theme | Style | Description |
|-------|-------|-------------|
| tokyonight-night | dark | Standard Tokyo Night |
| tokyonight-storm | dark | Stormy blue variant |
| tokyonight-day | light | Light Tokyo Night |
| catppuccin-mocha | dark | Soothing pastel |
| catppuccin-latte | light | Warm light pastel |
| gruvbox-dark | dark | Retro warm |
| nord | dark | Arctic bluish |
| dracula | dark | Purple theme |
| and more... | | |

### Parametric Generator

Create custom CoolNight variants:

```bash
nvp theme create --hue 210 --name my-blue-theme
nvp theme create --hue 350 --name my-rose-theme
```

See full list:

```bash
nvp theme list
```

---

## File Structure

nvp stores data in `~/.nvp/`:

```
~/.nvp/
├── plugins/           # Installed plugin YAMLs
│   ├── telescope.yaml
│   ├── treesitter.yaml
│   └── ...
└── themes/            # Installed themes
    └── active.yaml    # Currently active theme
```

Generated files go to `~/.config/nvim/lua/plugins/nvp/`:

```
~/.config/nvim/lua/plugins/nvp/
├── telescope.lua
├── treesitter.lua
├── lspconfig.lua
└── ...
```

---

## Standalone Usage

nvp works completely independently:

- No Docker required
- No dvm required
- Just nvp + Neovim

Perfect for:

- Local Neovim setup
- Dotfiles management
- Team plugin sharing

---

## Integration with dvm

When used with dvm:

- Plugins configured via nvp are included in workspace builds
- Container Neovim has the same plugins
- Consistent setup local ↔ container

---

## Next Steps

- [Plugins](plugins.md) - Managing plugins
- [Themes](themes.md) - Managing themes
- [Commands Reference](commands.md) - Full command list
