# Plugins

Managing Neovim plugins with nvp. Use the curated library of 38+ plugins or plugin packages for complete setups.

---

## Installing Plugins

### From Library

The easiest way - use the built-in library of 38+ curated plugins:

```bash
# See available plugins
nvp plugin list

# Install complete plugin package (39 plugins)
nvp apply -f package:rkohlman-full

# Install individual plugins
nvp apply -f plugin:telescope
nvp apply -f plugin:treesitter
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
nvp plugin list
```

### Get Plugin Details

```bash
nvp plugin get telescope
nvp plugin get telescope -o yaml
```

### Delete a Plugin

```bash
nvp plugin delete telescope
```

---

## Library Plugins

nvp includes 38+ curated plugins with a plugin package system for complete setups:

### Plugin Packages

Complete plugin configurations for specific use cases:

| Package | Plugins | Description |
|---------|---------|-------------|
| `rkohlman-full` | 39 plugins | Complete development environment with LSP, fuzzy finding, Git, AI, and more |

Install a complete package:

```bash
nvp apply -f package:rkohlman-full
```

### Individual Plugins

Choose from 38+ curated plugins:

#### Core Dependencies

| Plugin | Description |
|--------|-------------|
| `plenary` | Lua utility functions (dependency for many plugins) |
| `nvim-web-devicons` | File icons |

#### Fuzzy Finding & Navigation

| Plugin | Description |
|--------|-------------|
| `telescope` | Fuzzy finder for everything |
| `harpoon` | Quick file navigation |

#### Syntax & Parsing

| Plugin | Description |
|--------|-------------|
| `treesitter` | Advanced syntax highlighting |
| `treesitter-textobjects` | Text objects based on syntax |

#### LSP & Completion

| Plugin | Description |
|--------|-------------|
| `lspconfig` | LSP configuration |
| `mason` | LSP/DAP/Linter installer |
| `nvim-cmp` | Autocompletion |
| `cmp-nvim-lsp` | LSP completion source |
| `cmp-buffer` | Buffer completion source |
| `cmp-path` | Path completion source |
| `luasnip` | Snippet engine |

#### Git Integration

| Plugin | Description |
|--------|-------------|
| `gitsigns` | Git decorations and hunks |
| `fugitive` | Git commands |
| `diffview` | Git diff viewer |

#### UI & Interface

| Plugin | Description |
|--------|-------------|
| `lualine` | Status line |
| `bufferline` | Buffer/tab line |
| `which-key` | Keybinding hints |
| `alpha-nvim` | Dashboard |
| `neo-tree` | File tree |
| `dressing` | Better UI for inputs/selects |
| `notify` | Better notifications |

#### Editing & Text Manipulation

| Plugin | Description |
|--------|-------------|
| `comment` | Easy commenting |
| `surround` | Surround text with pairs |
| `autopairs` | Auto-close brackets |
| `conform` | Formatting |
| `nvim-lint` | Linting |

#### Terminal & System

| Plugin | Description |
|--------|-------------|
| `toggleterm` | Terminal management |

#### AI & Assistance

| Plugin | Description |
|--------|-------------|
| `copilot` | GitHub Copilot |
| `copilot-cmp` | Copilot completion source |

#### Language-Specific

| Plugin | Description |
|--------|-------------|
| `rust-tools` | Rust development |
| `go` | Go development |
| `typescript-tools` | TypeScript/JavaScript |

#### And More...

Plus additional plugins for debugging, testing, note-taking, and more specialized workflows.

See full list:

```bash
nvp plugin list
```

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
