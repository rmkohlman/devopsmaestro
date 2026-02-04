# Plugins

Managing Neovim plugins with nvp.

---

## Installing Plugins

### From Library

The easiest way - use the built-in library:

```bash
# See available plugins
nvp library list

# Install one
nvp library install telescope

# Install multiple
nvp library install telescope treesitter lspconfig nvim-cmp
```

### From File

Apply a plugin from a YAML file:

```bash
nvp apply -f my-plugin.yaml
```

### From URL

Apply directly from a URL:

```bash
nvp apply -f https://example.com/plugin.yaml
```

### From GitHub

Use GitHub shorthand:

```bash
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
```

### From Stdin

Pipe YAML content:

```bash
cat plugin.yaml | nvp apply -f -
```

---

## Plugin YAML Format

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  category: fuzzy-finder
  description: Highly extendable fuzzy finder
spec:
  repo: nvim-telescope/telescope.nvim
  branch: master                    # Optional
  version: "0.1.5"                  # Optional (tag)
  enabled: true                     # Default: true
  lazy: true                        # Default: true
  event:                            # Lazy-load triggers
    - VimEnter
  cmd:                              # Commands that trigger load
    - Telescope
  dependencies:
    - nvim-lua/plenary.nvim
    - nvim-tree/nvim-web-devicons
  config: |
    require("telescope").setup({
      defaults = {
        file_ignore_patterns = { "node_modules", ".git" },
      },
    })
  keys:
    - key: "<leader>ff"
      action: "<cmd>Telescope find_files<cr>"
      desc: "Find files"
    - key: "<leader>fg"
      action: "<cmd>Telescope live_grep<cr>"
      desc: "Live grep"
```

---

## Managing Plugins

### List Installed Plugins

```bash
nvp list
```

### Get Plugin Details

```bash
nvp get telescope
nvp get telescope -o yaml
```

### Delete a Plugin

```bash
nvp delete telescope
```

---

## Library Plugins

nvp includes these pre-configured plugins:

### Core

| Plugin | Description |
|--------|-------------|
| `plenary` | Lua utility functions (dependency for many plugins) |
| `nvim-web-devicons` | File icons |

### Fuzzy Finding

| Plugin | Description |
|--------|-------------|
| `telescope` | Fuzzy finder for everything |

### Syntax & Parsing

| Plugin | Description |
|--------|-------------|
| `treesitter` | Advanced syntax highlighting |

### LSP & Completion

| Plugin | Description |
|--------|-------------|
| `lspconfig` | LSP configuration |
| `mason` | LSP/DAP/Linter installer |
| `nvim-cmp` | Autocompletion |

### Git

| Plugin | Description |
|--------|-------------|
| `gitsigns` | Git decorations and hunks |

### UI

| Plugin | Description |
|--------|-------------|
| `lualine` | Status line |
| `which-key` | Keybinding hints |
| `alpha` | Dashboard |
| `neo-tree` | File tree |

### Editing

| Plugin | Description |
|--------|-------------|
| `comment` | Easy commenting |
| `conform` | Formatting |
| `nvim-lint` | Linting |

### Terminal

| Plugin | Description |
|--------|-------------|
| `toggleterm` | Terminal management |

### AI

| Plugin | Description |
|--------|-------------|
| `copilot` | GitHub Copilot |

---

## Generating Lua Files

After installing plugins, generate the Lua files:

```bash
nvp generate
```

This creates files in `~/.config/nvim/lua/plugins/nvp/`.

### Custom Output Directory

```bash
nvp generate --output ~/my-config/lua/plugins
```

---

## Plugin Dependencies

Dependencies are automatically handled:

```yaml
spec:
  dependencies:
    - nvim-lua/plenary.nvim
    - repo: nvim-tree/nvim-web-devicons
      config: |
        require("nvim-web-devicons").setup()
```

Simple dependencies are just repo strings. Complex dependencies can have their own config.

---

## Lazy Loading

Control when plugins load:

```yaml
spec:
  lazy: true          # Don't load on startup
  event:              # Load on these events
    - BufReadPost
    - BufNewFile
  cmd:                # Load when these commands run
    - Telescope
  ft:                 # Load for these filetypes
    - python
    - go
  keys:               # Load when these keys are pressed
    - key: "<leader>ff"
      action: "<cmd>Telescope find_files<cr>"
```

---

## Next Steps

- [Themes](themes.md) - Manage colorschemes
- [Commands Reference](commands.md) - Full command list
