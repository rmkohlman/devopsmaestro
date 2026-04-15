# Built-in Nvim Packages

Nvim packages are curated plugin bundles. Instead of selecting plugins one by one, choose a package that fits your workflow and get a complete, tested configuration immediately.

---

## What Is a Package?

A **package** bundles multiple related plugins into a single unit:

- **Built-in packages** — embedded in the binary, available without installation
- **User packages** — created with `dvm apply -f package.yaml`
- **Inheritance** — packages can extend other packages (`spec.extends: core`)

New workspaces automatically use the `core` package by default.

---

## Browsing the Library

```bash
# List all built-in packages
dvm library get nvim packages

# Short alias (multi-word form)
dvm lib ls "nvim packages"

# Show details for a specific package
dvm library describe nvim-package core
dvm library describe nvim-package lazyvim
```

To list packages you've imported into your database alongside library packages:

```bash
dvm get nvim packages
dvm get nvim package core    # Details for one package
```

---

## Built-in Packages

### `core`

The foundation package. Every new workspace starts with this.

**Includes (6 essential plugins):**

| Plugin | Role |
|--------|------|
| `nvim-telescope/telescope.nvim` | Fuzzy finder for files, text, and symbols |
| `nvim-treesitter/nvim-treesitter` | Modern syntax highlighting and parsing |
| `neovim/nvim-lspconfig` | Language Server Protocol (LSP) support |
| `hrsh7th/nvim-cmp` | Intelligent autocompletion |
| `lewis6991/gitsigns.nvim` | Git decorations and inline diff |
| `folke/which-key.nvim` | Keybinding hints and discovery |

```bash
dvm library describe nvim-package core
```

### `lazyvim`

LazyVim-based distribution. Provides a full IDE-like experience built on the LazyVim framework.

```bash
dvm library describe nvim-package lazyvim
```

### `maestro-python`

Python development package. Automatically selected for Python workspaces during `dvm build`.

**Extends:** `core`

**Adds:**

| Plugin | Role |
|--------|------|
| `nvim-neotest/neotest` | Test runner framework |
| `nvim-neotest/neotest-python` | Python test adapter |
| `mfussenegger/nvim-dap-python` | Python debugging |
| `stevearc/conform.nvim` | Formatting (black, isort) |

```bash
dvm library describe nvim-package maestro-python
```

### `maestro-go`

Go development package. Automatically selected for Go workspaces during `dvm build`.

**Extends:** `core`

**Adds:**

| Plugin | Role |
|--------|------|
| `ray-x/go.nvim` | Go tools integration |
| `leoluz/nvim-dap-go` | Go debugging (Delve) |
| `nvim-neotest/neotest-go` | Go test runner |
| `stevearc/conform.nvim` | Formatting (gofmt, goimports) |

```bash
dvm library describe nvim-package maestro-go
```

---

## Library Plugins (38+ available)

Individual plugins can be browsed and used to build custom packages:

```bash
# List all available library plugins
dvm library get plugins

# Short alias
dvm lib ls np

# Show details for a specific plugin
dvm library describe plugin telescope
dvm library describe plugin treesitter
```

### Core Dependencies

| Plugin | Description |
|--------|-------------|
| `plenary` | Lua utility functions (required by many plugins) |
| `nvim-web-devicons` | File type icons |

### Fuzzy Finding & Navigation

| Plugin | Description |
|--------|-------------|
| `telescope` | Fuzzy finder for files, grep, LSP symbols |
| `harpoon` | Quick file marks and navigation |

### Syntax & Parsing

| Plugin | Description |
|--------|-------------|
| `treesitter` | Advanced syntax highlighting |
| `treesitter-textobjects` | Syntax-aware text objects |

### LSP & Completion

| Plugin | Description |
|--------|-------------|
| `lspconfig` | LSP client configuration |
| `mason` | LSP/DAP/linter installer |
| `nvim-cmp` | Autocompletion engine |
| `cmp-nvim-lsp` | LSP completion source |
| `cmp-buffer` | Buffer word completion |
| `cmp-path` | Filesystem path completion |
| `luasnip` | Snippet engine |

### Git Integration

| Plugin | Description |
|--------|-------------|
| `gitsigns` | Git decorations and hunk navigation |
| `fugitive` | Git commands in Neovim |
| `diffview` | Side-by-side git diff viewer |

### UI & Interface

| Plugin | Description |
|--------|-------------|
| `lualine` | Status line |
| `bufferline` | Buffer/tab line |
| `which-key` | Keybinding hints |
| `alpha-nvim` | Dashboard / start screen |
| `neo-tree` | File tree explorer |
| `dressing` | Improved input and select UI |
| `notify` | Notification system |

### Editing & Text Manipulation

| Plugin | Description |
|--------|-------------|
| `comment` | Quick comment/uncomment |
| `surround` | Surround text with brackets, quotes |
| `autopairs` | Auto-close brackets and quotes |
| `conform` | Code formatting |
| `nvim-lint` | Linting integration |

### Terminal & AI

| Plugin | Description |
|--------|-------------|
| `toggleterm` | Terminal panel management |
| `copilot` | GitHub Copilot integration |
| `copilot-cmp` | Copilot as a completion source |
| `snacks` | QoL utilities (input, picker, opencode) |
| `opencode` | opencode AI assistant integration |

### Language-Specific

| Plugin | Description |
|--------|-------------|
| `rust-tools` | Rust development tools |
| `go` | Go tools integration |
| `typescript-tools` | TypeScript/JavaScript support |

---

## Setting the Default Package

Set which package new workspaces use automatically:

```bash
# Set default package for all new workspaces
dvm use nvim package core

# Set a different default
dvm use nvim package lazyvim

# Clear default (no package for new workspaces)
dvm use nvim package none

# Show current default
dvm get nvim defaults
```

---

## Applying a Package to a Workspace

```bash
# Set the nvim package for a specific workspace
dvm set nvim package --workspace dev maestro-python

# Set the package and app-scope the workspace lookup
dvm set nvim package --app my-api --workspace dev maestro-go
```

---

## Creating a Custom Package

Define a package as a YAML resource and apply it:

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: my-python-stack
  description: Full Python development setup
  category: language
  tags: ["python", "fastapi", "testing"]
spec:
  extends: core         # Inherits all core plugins
  plugins:
    - nvim-neotest/neotest
    - nvim-neotest/neotest-python
    - mfussenegger/nvim-dap-python
    - stevearc/conform.nvim
    - github/copilot.vim
```

```bash
dvm apply -f my-python-stack.yaml

# Apply from URL
dvm apply -f https://example.com/packages/my-stack.yaml

# Apply from GitHub
dvm apply -f github:user/repo/packages/my-stack.yaml
```

---

## Package Inheritance

Packages can extend other packages, adding plugins without duplication:

```
core
├── maestro-go      (adds Go tools)
├── maestro-python  (adds Python tools)
└── lazyvim         (full LazyVim distribution)
```

When a package specifies `extends: core`, it gets all of `core`'s plugins plus its own.

---

## Using a Package in Workspace YAML

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  nvim:
    pluginPackage: maestro-go    # Built-in or custom package
    plugins:                     # Additional plugins on top
      - github/copilot.vim
    mergeMode: append            # append (default) or replace
```

**Merge modes:**

| Mode | Behavior |
|------|----------|
| `append` | Package plugins + workspace plugins (default) |
| `replace` | Only workspace plugins, ignore package |

---

## Terminal Packages

Terminal packages bundle shell plugins and prompts. Use the same discovery pattern:

```bash
# List built-in terminal packages
dvm library get terminal packages

# Show details
dvm library describe terminal-package core

# List terminal prompts
dvm library get terminal prompts

# List terminal plugins (shell plugins)
dvm library get terminal plugins
```

---

## Related

- **[Quick Start: Themes](quick-start-themes.md)** — How to apply themes to your workspace
- **[Plugin Packages](packages.md)** — Full package management guide
- **[NvimPackage YAML Reference](../reference/nvim-package.md)** — Full YAML schema
- **[Workspace Reference](../reference/workspace.md)** — Using packages in workspace YAML
