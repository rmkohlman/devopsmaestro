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

### `maestro`

Complete Neovim IDE setup with 30+ plugins. Includes AI (Copilot, Copilot Chat), Git (gitsigns, lazygit), navigation (harpoon, nvim-tree, bufferline), database tools (dadbod, dbee), notes (obsidian), markdown, formatting, linting, testing, and more.

```bash
dvm library describe nvim-package maestro
```

### `go-dev`

Go development essentials.

**Extends:** `core`

**Adds:**

| Plugin | Role |
|--------|------|
| `gopher-nvim` | Go tools integration |
| `nvim-dap` | Debug Adapter Protocol |
| `nvim-dap-go` | Go debugging (Delve) |
| `neotest` | Test runner framework |
| `neotest-go` | Go test runner |

```bash
dvm library describe nvim-package go-dev
```

### `maestro-go`

Full Go IDE experience.

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

### Other Language Packages

| Package | Language | Description |
|---------|----------|-------------|
| `maestro-rust` | Rust | rustaceanvim, crates-nvim, neotest-rust |
| `maestro-node` | Node.js | neotest-jest, TypeScript tools |
| `maestro-java` | Java | nvim-jdtls (eclipse.jdt.ls) |
| `maestro-gleam` | Gleam | Gleam language tools |
| `maestro-dotnet` | .NET | .NET development tools |
| `python-dev` | Python | Alternative Python setup |
| `full` | All | Full plugin collection (extends core) |

```bash
# List all available packages
nvp package get

# Show details for any package
dvm library describe nvim-package maestro-rust
```

---

## Library Plugins (54 available)

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

### Core & Navigation

| Plugin | Description |
|--------|-------------|
| `telescope` | Fuzzy finder for files, grep, LSP symbols |
| `treesitter` | Advanced syntax highlighting |
| `treesitter-textobjects` | Syntax-aware text objects |
| `harpoon` | Quick file marks and navigation |
| `nvim-tree` | File tree explorer (nvim-tree/nvim-tree.lua) |

### LSP & Completion

| Plugin | Description |
|--------|-------------|
| `lspconfig` | LSP client configuration |
| `mason` | LSP/DAP/linter installer |
| `nvim-cmp` | Autocompletion engine |
| `formatting` | Code formatting (conform.nvim) |
| `linting` | Code linting |

### Git Integration

| Plugin | Description |
|--------|-------------|
| `gitsigns` | Git decorations and hunk navigation |
| `lazygit` | LazyGit terminal integration |

### UI & Interface

| Plugin | Description |
|--------|-------------|
| `lualine` | Status line |
| `bufferline` | Buffer/tab line |
| `which-key` | Keybinding hints |
| `alpha` | Dashboard / start screen |
| `dressing` | Improved input and select UI |
| `indent-blankline` | Indentation guides |
| `trouble` | Diagnostics and quickfix list |
| `todo-comments` | TODO comment highlighting |

### Editing & Utilities

| Plugin | Description |
|--------|-------------|
| `comment` | Quick comment/uncomment |
| `surround` | Surround text with brackets, quotes |
| `autopairs` | Auto-close brackets and quotes |
| `substitute` | Substitute text motions |
| `vim-maximizer` | Window maximizer |
| `auto-session` | Session management |
| `toggleterm` | Terminal panel management |

### AI & Copilot

| Plugin | Description |
|--------|-------------|
| `copilot` | GitHub Copilot integration |
| `copilot-cmp` | Copilot as a completion source |
| `copilot-chat` | Copilot Chat with glob support |
| `snacks` | QoL utilities (input, picker, opencode) |
| `opencode` | opencode AI assistant integration |

### Database

| Plugin | Description |
|--------|-------------|
| `dadbod` | Database client |
| `dadbod-ui` | Database UI |
| `dadbod-completion` | Database completion source |
| `dbee` | Advanced database explorer |

### Markdown & Notes

| Plugin | Description |
|--------|-------------|
| `render-markdown` | Enhanced markdown rendering |
| `markdown-preview` | Markdown browser preview |
| `obsidian` | Obsidian note-taking integration |

### Debug & Test

| Plugin | Description |
|--------|-------------|
| `nvim-dap` | Debug Adapter Protocol |
| `neotest` | Test runner framework |

### Language-Specific

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
├── go-dev          (Go dev essentials)
├── maestro-go      (full Go IDE)
├── maestro-python  (full Python IDE)
├── maestro-rust    (full Rust IDE)
├── maestro-node    (Node.js IDE)
├── maestro-java    (Java IDE)
├── full            (full plugin collection)
└── ...             (maestro-gleam, maestro-dotnet, python-dev)
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
