# Plugins

Managing Neovim plugins with nvp. Use the curated library of 54 plugins or plugin packages for complete setups.

---

## Installing Plugins

### From Library

The easiest way — use the built-in library of 54 curated plugins:

```bash
# Browse available plugins
nvp library get

# Filter by category
nvp library get --category lsp

# Install individual plugins
nvp library import telescope
nvp library import treesitter
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
nvp get
```

### Get Plugin Details

```bash
nvp get telescope
nvp get telescope -o yaml
```

### Enable / Disable a Plugin

```bash
nvp enable telescope
nvp disable copilot
```

### Delete a Plugin

```bash
nvp delete telescope
```

---

## Library Plugins

nvp includes 54 curated plugins with a plugin package system for complete setups:

### Plugin Packages

Complete plugin configurations for specific use cases are available via the library:

| Package | Plugins | Description |
|---------|---------|-------------|
| `core` | 6 plugins | Essential base — telescope, treesitter, lspconfig, nvim-cmp, gitsigns, which-key |
| `maestro` | 30+ plugins | Complete IDE with AI, Git, database, notes, and more |
| `go-dev` | 11 plugins (incl. core) | Go development essentials |
| `maestro-go` | Full Go IDE | Extends core with Go tools, DAP, neotest, formatting |
| `maestro-python` | Full Python IDE | Extends core with Python tools, DAP, neotest, formatting |
| `maestro-rust` | Full Rust IDE | Extends core with rustaceanvim, crates, neotest |
| `maestro-node` | Full Node IDE | Extends core with neotest-jest, TypeScript tools |
| `maestro-java` | Full Java IDE | Extends core with nvim-jdtls |
| `maestro-gleam` | Gleam IDE | Extends core with Gleam tools |
| `maestro-dotnet` | .NET IDE | Extends core with .NET tools |
| `python-dev` | Python dev | Python development setup |
| `full` | 17 plugins (incl. core) | Full plugin collection |

Install all plugins from a package:

```bash
nvp package install core
nvp package install maestro-go
nvp package install --dry-run maestro-python  # preview first
```

### Individual Plugins

Choose from 54 curated plugins:

#### Core Dependencies

> **Note:** `plenary` and `nvim-web-devicons` are used as dependencies by other plugins and are referenced by repo URL inside plugin YAML specs. They are not installed as standalone library plugins but are included automatically when needed.

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
| `formatting` | Code formatting (conform.nvim) |
| `linting` | Code linting |

#### Git Integration

| Plugin | Description |
|--------|-------------|
| `gitsigns` | Git decorations and hunks |
| `lazygit` | LazyGit terminal integration |

#### UI & Interface

| Plugin | Description |
|--------|-------------|
| `lualine` | Status line |
| `bufferline` | Buffer/tab line |
| `which-key` | Keybinding hints |
| `alpha` | Dashboard |
| `nvim-tree` | File tree (nvim-tree/nvim-tree.lua) |
| `dressing` | Better UI for inputs/selects |
| `indent-blankline` | Indentation guides |
| `trouble` | Diagnostics and quickfix list |
| `todo-comments` | TODO comment highlighting |

#### Editing & Text Manipulation

| Plugin | Description |
|--------|-------------|
| `comment` | Easy commenting |
| `surround` | Surround text with pairs |
| `autopairs` | Auto-close brackets |
| `substitute` | Substitute text motions |
| `vim-maximizer` | Window maximizer |
| `auto-session` | Session management |

#### Terminal & System

| Plugin | Description |
|--------|-------------|
| `toggleterm` | Terminal management |

#### AI & Assistance

| Plugin | Description |
|--------|-------------|
| `copilot` | GitHub Copilot |
| `copilot-cmp` | Copilot completion source |
| `copilot-chat` | Copilot Chat with glob support |
| `snacks` | QoL utility library (input, picker, opencode integration) |
| `opencode` | opencode AI assistant integration |

#### Database

| Plugin | Description |
|--------|-------------|
| `dadbod` | Database client |
| `dadbod-ui` | Database UI |
| `dadbod-completion` | Database completion source |
| `dbee` | Advanced database explorer |

#### Markdown & Notes

| Plugin | Description |
|--------|-------------|
| `render-markdown` | Enhanced markdown rendering |
| `markdown-preview` | Markdown browser preview |
| `obsidian` | Obsidian note-taking integration |

#### Debug & Test

| Plugin | Description |
|--------|-------------|
| `nvim-dap` | Debug Adapter Protocol |
| `neotest` | Test runner framework |

#### Language-Specific

| Plugin | Description |
|--------|-------------|
| `gopher-nvim` | Go development tools |
| `nvim-dap-go` | Go debugger (Delve) |
| `neotest-go` | Go test runner |
| `rustaceanvim` | Rust development tools |
| `crates-nvim` | Rust crates.io integration |
| `neotest-rust` | Rust test runner |
| `nvim-dap-python` | Python debugger |
| `neotest-python` | Python test runner |
| `venv-selector` | Python virtualenv selector |
| `neotest-jest` | JavaScript/TypeScript test runner |
| `nvim-jdtls` | Java LSP (eclipse.jdt.ls) |
| `euporie` | Jupyter notebook editing |

See full list:

```bash
nvp library get
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
nvp generate --output-dir ~/my-config/lua/plugins
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
