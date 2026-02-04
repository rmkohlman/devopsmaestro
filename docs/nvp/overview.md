# nvp Overview

`nvp` (NvimOps) is a standalone Neovim plugin and theme manager using YAML.

---

## What is nvp?

nvp lets you:

- **Define plugins in YAML** instead of Lua
- **Use a curated library** of 16+ pre-configured plugins
- **Manage themes** with exported color palettes
- **Generate Lua files** for lazy.nvim

---

## Key Features

- :material-file-code: **YAML-based** - Write plugins in familiar YAML format
- :material-library: **Built-in library** - Telescope, Treesitter, LSP, and more
- :material-palette: **Theme system** - 8 themes with palette export
- :material-link: **URL support** - Install from GitHub
- :material-package-variant: **Standalone** - No containers required

---

## Quick Start

```bash
# Initialize
nvp init

# Install plugins from library
nvp library list
nvp library install telescope treesitter lspconfig

# Install a theme
nvp theme library install tokyonight-custom --use

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

nvp includes 16+ curated plugins:

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
| and more... | | |

See full list:

```bash
nvp library list
```

---

## Theme Library

8 pre-built themes:

| Theme | Style | Description |
|-------|-------|-------------|
| tokyonight-custom | dark | Custom deep blue variant |
| tokyonight-night | dark | Standard Tokyo Night |
| catppuccin-mocha | dark | Soothing pastel |
| catppuccin-latte | light | Warm light pastel |
| gruvbox-dark | dark | Retro warm |
| nord | dark | Arctic bluish |
| rose-pine | dark | Natural pine |
| kanagawa | dark | Japanese painting inspired |

See full list:

```bash
nvp theme library list
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
