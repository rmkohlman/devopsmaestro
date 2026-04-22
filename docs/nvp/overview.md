# nvp Overview

`nvp` (NvimOps) is a standalone Neovim plugin and theme manager using YAML.

---

## What is nvp?

nvp lets you:

- **Define plugins in YAML** instead of Lua
- **Use a curated library** of 54 pre-configured plugins
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
- :material-library: **Built-in library** - 54 curated plugins and plugin packages
- :material-palette: **Theme system** - 34+ themes with parametric generator
- :material-link: **URL support** - Install from GitHub
- :material-package-variant: **Standalone** - No containers required

---

## Quick Start

```bash
# Initialize
nvp init

# Browse and install plugins from library
nvp library get
nvp library import telescope
nvp library import treesitter

# Apply a plugin from a file or URL
nvp apply -f my-plugin.yaml
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml

# List and use themes
nvp theme list
nvp theme create --from "210" --name my-blue-theme --use

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

nvp includes 54 curated plugins with plugin package system:

### Plugin Packages

| Package | Plugins | Description |
|---------|---------|-------------|
| `core` | 6 plugins | Essential base — telescope, treesitter, lspconfig, nvim-cmp, gitsigns, which-key |
| `maestro` | 30+ plugins | Complete IDE with AI, Git, database, notes, and more |
| `go-dev` | 11 plugins (incl. core) | Go development essentials |
| `maestro-go` | Full Go IDE | Extends core with Go tools, DAP, neotest, formatting |
| `maestro-python` | Full Python IDE | Extends core with Python tools, DAP, neotest, formatting |
| `full` | 17 plugins (incl. core) | Full plugin collection |
| and more... | | maestro-rust, maestro-node, maestro-java, maestro-gleam, maestro-dotnet, python-dev |

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

### Complete Plugin Library (54)

| Plugin | Category | Description |
|--------|----------|-------------|
| telescope | fuzzy-finder | Fuzzy finder for files, grep, etc. |
| treesitter | syntax | Advanced syntax highlighting |
| lspconfig | lsp | LSP configuration |
| nvim-cmp | completion | Autocompletion |
| mason | lsp | LSP/DAP/Linter installer |
| gitsigns | git | Git decorations |
| lazygit | git | LazyGit terminal integration |
| lualine | ui | Status line |
| bufferline | ui | Buffer/tab line |
| which-key | ui | Keybinding hints |
| nvim-tree | file-explorer | File tree (nvim-tree/nvim-tree.lua) |
| toggleterm | terminal | Terminal management |
| copilot | ai | GitHub Copilot |
| copilot-cmp | ai | Copilot completion source |
| copilot-chat | ai | Copilot Chat integration |
| snacks | utility | QoL utilities (input, picker, opencode integration) |
| opencode | ai | opencode AI assistant integration |
| formatting | lsp | Code formatting (conform.nvim) |
| linting | lsp | Code linting |
| trouble | ui | Diagnostics and quickfix list |
| todo-comments | ui | TODO comment highlighting |
| indent-blankline | ui | Indentation guides |
| auto-session | utility | Session management |
| vim-maximizer | utility | Window maximizer |
| substitute | editing | Substitute text motions |
| dadbod | database | Database client |
| dadbod-ui | database | Database UI |
| dadbod-completion | database | Database completion source |
| dbee | database | Advanced database explorer |
| render-markdown | markdown | Enhanced markdown rendering |
| markdown-preview | markdown | Markdown browser preview |
| obsidian | notes | Obsidian note-taking integration |
| nvim-dap | debug | Debug Adapter Protocol |
| neotest | test | Test runner framework |
| and more... | | gopher-nvim, neotest-go, nvim-dap-go, rustaceanvim, crates-nvim, neotest-rust, neotest-python, nvim-dap-python, venv-selector, neotest-jest, nvim-jdtls, euporie |

See full list:

```bash
nvp library get
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

### Popular Themes (13 others)

| Theme | Style | Description |
|-------|-------|-------------|
| tokyonight-night | dark | Standard Tokyo Night |
| tokyonight-ocean | dark | Tokyo Night ocean variant |
| catppuccin-mocha | dark | Soothing pastel |
| catppuccin-latte | light | Warm light pastel |
| gruvbox-dark | dark | Retro warm |
| nord | dark | Arctic bluish |
| dracula | dark | Purple theme |
| onedark | dark | Atom-inspired dark |
| rose-pine | dark | Rose Pinè natural tones |
| kanagawa | dark | Inspired by Kanagawa art |
| everforest | dark | Warm green |
| solarized-dark | dark | Scientific blue-green |
| tokyonight-custom | dark | Custom Tokyo Night |

### Parametric Generator

Create custom CoolNight variants using a hue angle, hex color, or preset name:

```bash
nvp theme create --from "210" --name my-blue-theme
nvp theme create --from "350" --name my-rose-theme
nvp theme create --from "#8B00FF" --name my-violet-theme
nvp theme create --from "synthwave" --name my-synth
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
