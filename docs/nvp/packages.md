# Plugin Packages

The nvp library contains curated collections of Neovim plugins organized into reusable sets. Instead of installing plugins one by one, you can browse the library and install plugins that work well together.

---

## What Are Plugin Packages?

The nvp library provides pre-configured plugin definitions with:

- **Complete setups** - Tested configurations that work together
- **Consistent experience** - Uniform key bindings and behavior across plugins
- **Quick setup** - Install individual plugins or entire sets from the library

---

## Available Library Plugins

### Core Dependencies

| Plugin | Description |
|--------|-------------|
| `plenary` | Lua utility functions (dependency for many plugins) |
| `nvim-web-devicons` | File icons |

### Fuzzy Finding & Navigation

| Plugin | Description |
|--------|-------------|
| `telescope` | Fuzzy finder for everything |
| `harpoon` | Quick file navigation |

### Syntax & Parsing

| Plugin | Description |
|--------|-------------|
| `treesitter` | Advanced syntax highlighting |
| `treesitter-textobjects` | Text objects based on syntax |

### LSP & Completion

| Plugin | Description |
|--------|-------------|
| `lspconfig` | LSP configuration |
| `mason` | LSP/DAP/Linter installer |
| `nvim-cmp` | Autocompletion |
| `cmp-nvim-lsp` | LSP completion source |
| `cmp-buffer` | Buffer completion source |
| `cmp-path` | Path completion source |
| `luasnip` | Snippet engine |

### Git Integration

| Plugin | Description |
|--------|-------------|
| `gitsigns` | Git decorations and hunks |
| `fugitive` | Git commands |
| `diffview` | Git diff viewer |

### UI & Interface

| Plugin | Description |
|--------|-------------|
| `lualine` | Status line |
| `bufferline` | Buffer/tab line |
| `which-key` | Keybinding hints |
| `alpha-nvim` | Dashboard |
| `neo-tree` | File tree |
| `dressing` | Better UI for inputs/selects |
| `notify` | Better notifications |

### Editing & Text Manipulation

| Plugin | Description |
|--------|-------------|
| `comment` | Easy commenting |
| `surround` | Surround text with pairs |
| `autopairs` | Auto-close brackets |
| `conform` | Formatting |
| `nvim-lint` | Linting |

### Terminal & AI

| Plugin | Description |
|--------|-------------|
| `toggleterm` | Terminal management |
| `copilot` | GitHub Copilot |
| `copilot-cmp` | Copilot completion source |

### Language-Specific

| Plugin | Description |
|--------|-------------|
| `rust-tools` | Rust development |
| `go` | Go development |
| `typescript-tools` | TypeScript/JavaScript |

---

## Installing from the Library

### Browse and Install

```bash
# See all available plugins
nvp library list

# Filter by category
nvp library list --category lsp
nvp library list --category fuzzy-finder

# See categories
nvp library categories

# See tags
nvp library tags

# Show details about a plugin
nvp library show telescope

# Install individual plugins
nvp library install telescope
nvp library install treesitter
nvp library install lspconfig
```

### Install from a YAML File or URL

If you have custom plugin definitions or want to install from GitHub:

```bash
# From local file
nvp apply -f my-plugin.yaml

# From URL
nvp apply -f https://example.com/plugin.yaml

# From GitHub
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml

# From stdin
cat plugin.yaml | nvp apply -f -
```

---

## Managing Installed Plugins

### List, Inspect, Enable/Disable, and Delete

```bash
# List installed plugins
nvp list

# Get plugin details
nvp get telescope
nvp get telescope -o yaml

# Enable or disable without deleting
nvp enable telescope
nvp disable copilot

# Delete a plugin
nvp delete telescope
```

### Override a Plugin Configuration

Apply a customized YAML to override a plugin's settings:

```bash
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

nvp apply -f my-telescope.yaml
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

## Generating Lua Files

After installing plugins, generate the Lua configuration files:

```bash
nvp generate
```

This creates files in `~/.config/nvim/lua/plugins/nvp/`.

### Custom Output Directory

```bash
nvp generate --output ~/my-config/lua/plugins
```

---

## Performance Considerations

### Lazy Loading

All library plugins come pre-configured with lazy loading:

- **Event-based triggers** - Load on file types, commands, or events
- **Dependency management** - Load dependencies in correct order
- **Startup optimization** - Only essential plugins load on startup

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

## Troubleshooting

### Plugin errors

```bash
# Check Neovim's health
nvim --headless -c 'checkhealth' -c 'quit'

# Regenerate clean configuration
rm -rf ~/.config/nvim/lua/plugins/nvp
nvp generate
```

### Check what is installed

```bash
# List all installed plugins
nvp list

# Get details on a specific plugin
nvp get telescope -o yaml
```

---

## Next Steps

- [Plugin Documentation](plugins.md) - Individual plugin management commands
- [Themes](themes.md) - Theme system integration
- [Commands Reference](commands.md) - Full nvp command list
- [Configuration](../configuration/yaml-schema.md) - YAML format details
