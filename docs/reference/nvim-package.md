# NvimPackage YAML Reference

**Kind:** `NvimPackage`  
**APIVersion:** `devopsmaestro.io/v1`

An NvimPackage represents a collection of related Neovim plugins that work together to provide a complete development environment for a specific use case.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: golang-dev
  description: "Complete Go development environment with LSP, debugging, and testing"
  category: "language"
  tags: ["go", "golang", "lsp", "debugging", "testing"]
  labels:
    language: "go"
    maintainer: "devopsmaestro"
  annotations:
    version: "1.0.0"
    last-updated: "2026-02-19"
    documentation: "https://github.com/devopsmaestro/packages/golang-dev"
spec:
  extends: "core"
  plugins:
    # LSP and completion
    - neovim/nvim-lspconfig
    - williamboman/mason.nvim
    - williamboman/mason-lspconfig.nvim
    - hrsh7th/nvim-cmp
    - hrsh7th/cmp-nvim-lsp
    
    # Go-specific plugins
    - fatih/vim-go
    - ray-x/go.nvim
    - ray-x/guihua.lua
    
    # Debugging
    - mfussenegger/nvim-dap
    - leoluz/nvim-dap-go
    - rcarriga/nvim-dap-ui
    
    # Testing
    - nvim-neotest/neotest
    - nvim-neotest/neotest-go
    
    # Code formatting
    - stevearc/conform.nvim
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `NvimPackage` |
| `metadata.name` | string | ✅ | Unique name for the package |
| `metadata.description` | string | ❌ | Package description |
| `metadata.category` | string | ❌ | Package category |
| `metadata.tags` | array | ❌ | Tags for organization |
| `metadata.labels` | object | ❌ | Key-value labels |
| `metadata.annotations` | object | ❌ | Key-value annotations |
| `spec.extends` | string | ❌ | Parent package to extend |
| `spec.plugins` | array | ✅ | List of plugin names |
| `spec.enabled` | boolean | ❌ | Enable or disable the package (default: `true`) |

## Field Details

### metadata.name (required)
The unique identifier for the package.

**Naming conventions:**
- Use descriptive names: `golang-dev`, `typescript-full`, `python-data`
- Include language/purpose: `rust-dev`, `web-frontend`, `data-science`
- Be specific: `react-typescript` vs `javascript-basic`

### metadata.category (optional)
Package category for organization.

**Common categories:**
- `language` - Language-specific packages
- `framework` - Framework-specific (React, Vue, etc.)
- `purpose` - Purpose-specific (data-science, devops, etc.)
- `core` - Base/foundation packages
- `specialty` - Specialized packages

### metadata.tags (optional)
Tags for filtering and searching packages.

```yaml
metadata:
  tags: ["go", "golang", "lsp", "debugging", "testing", "backend"]
```

### spec.extends (optional)
Parent package to inherit plugins from.

```yaml
spec:
  extends: "core"                    # Inherit from core package
```

**Package inheritance:**
```
core → language-specific → framework-specific
```

Example:
- `core` (base plugins)
- `golang-dev` extends `core` (adds Go plugins)
- `golang-web` extends `golang-dev` (adds web-specific Go plugins)

### spec.plugins (required)
List of plugin names to include in the package.

```yaml
spec:
  plugins:
    - neovim/nvim-lspconfig          # GitHub repo format
    - telescope                      # Local plugin name
    - cmp-nvim-lsp                   # Short name
```

**Plugin reference formats:**
- Full GitHub repo: `neovim/nvim-lspconfig`
- Short name: `telescope` (references local plugin)
- Hyphenated: `cmp-nvim-lsp`

### spec.enabled (optional)
Enable or disable the package. Defaults to `true`. When set to `false`, the package is stored but not applied.

```yaml
spec:
  enabled: false   # Disable this package
```

## Built-in Packages

### Core Package

Base package with essential plugins:

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: core
  description: "Essential plugins for any Neovim setup"
  category: core
spec:
  plugins:
    - nvim-telescope/telescope.nvim
    - nvim-treesitter/nvim-treesitter
    - neovim/nvim-lspconfig
    - hrsh7th/nvim-cmp
    - lewis6991/gitsigns.nvim
    - folke/which-key.nvim
    - nvim-lualine/lualine.nvim
```

### Language Packages

#### Golang Development

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: golang-dev
  category: language
  tags: ["go", "golang"]
spec:
  extends: core
  plugins:
    - fatih/vim-go
    - ray-x/go.nvim
    - leoluz/nvim-dap-go
```

#### Python Development

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: python-dev
  category: language
  tags: ["python"]
spec:
  extends: core
  plugins:
    - nvim-neotest/neotest-python
    - mfussenegger/nvim-dap-python
```

#### TypeScript Development

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: typescript-dev
  category: language
  tags: ["typescript", "javascript", "node"]
spec:
  extends: core
  plugins:
    - nvim-neotest/neotest-jest
    - mfussenegger/nvim-dap-node2
```

### Framework Packages

#### React Development

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: react-dev
  category: framework
  tags: ["react", "jsx", "typescript"]
spec:
  extends: typescript-dev
  plugins:
    - windwp/nvim-ts-autotag
    - JoosepAlviste/nvim-ts-context-commentstring
```

### Specialty Packages

#### Data Science

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: data-science
  category: purpose
  tags: ["python", "r", "jupyter", "data"]
spec:
  extends: python-dev
  plugins:
    - jupyter-vim/jupyter-vim
    - goerz/jupytext.vim
    - jalvesaq/Nvim-R
```

## Package Inheritance Examples

### Linear Inheritance

```yaml
# Base package
core:
  plugins: [telescope, treesitter, lspconfig]

# Language package  
golang-dev:
  extends: core
  plugins: [vim-go, go.nvim]
  # Result: telescope, treesitter, lspconfig, vim-go, go.nvim

# Framework package
golang-web:
  extends: golang-dev
  plugins: [rest.nvim]
  # Result: all golang-dev plugins + rest.nvim
```

### Multiple Extensions

```yaml
# Web development could extend multiple bases
web-dev:
  extends: typescript-dev
  plugins: [emmet-vim, vim-css-color]

# Full-stack combines frontend + backend  
fullstack-dev:
  extends: web-dev
  plugins: [vim-go, dockerfile.vim]
```

## Usage Examples

### Create Custom Package

```bash
# From YAML file
dvm apply -f my-package.yaml

# From URL
dvm apply -f https://packages.example.com/golang-dev.yaml

# From GitHub
dvm apply -f github:user/packages/my-stack.yaml
```

### List Packages

```bash
# List all packages
dvm get nvim packages

# List by category
dvm get nvim packages --category language

# Search packages  
dvm get nvim packages --name "*golang*"
```

### Use in Workspace

```yaml
# Use package in workspace
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: go-api
spec:
  nvim:
    pluginPackage: golang-dev       # Use the package
    plugins:                        # Add extra plugins
      - github/copilot.vim
    mergeMode: append               # Extend package
```

### Export Package

```bash
# Export to YAML
dvm get nvim package golang-dev -o yaml

# Export for sharing
dvm get nvim package my-custom-stack -o yaml > my-stack.yaml
```

## Package Merge Modes

When using packages in workspaces, you can control how plugins are combined:

```yaml
spec:
  nvim:
    pluginPackage: golang-dev
    plugins: [github/copilot.vim]
    mergeMode: append               # How to merge
```

**Merge modes:**
- `append` - Add workspace plugins to package plugins (default)
- `replace` - Replace package plugins with workspace plugins

## Best Practices

### Package Design

1. **Start with core** - Extend the `core` package for consistency
2. **Be specific** - Create focused packages for specific use cases
3. **Layer properly** - Use inheritance for related packages
4. **Document well** - Include good descriptions and tags

### Plugin Selection

1. **Choose essential plugins** - Include only necessary plugins
2. **Avoid conflicts** - Test plugin combinations
3. **Consider performance** - Balance features with startup time
4. **Stay updated** - Use maintained plugins

### Configuration

1. **Provide sane defaults** - Good out-of-the-box experience  
2. **Keep it minimal** - Let users customize further
3. **Use lazy loading** - Configure appropriate lazy loading
4. **Add helpful keymaps** - Include common workflows

## Related Resources

- [Workspace](workspace.md) - Use packages in workspaces
- [NvimPlugin](nvim-plugin.md) - Individual plugins
- [NvimTheme](nvim-theme.md) - Theme configurations

## Validation Rules

- `metadata.name` must be unique across all packages
- `metadata.name` must be a valid DNS subdomain
- `spec.extends` must reference an existing package
- `spec.plugins` must be valid plugin references
- Package inheritance must not create circular dependencies
- Package names must not conflict with built-in packages