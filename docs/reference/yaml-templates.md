# YAML Templates

Quick-reference YAML templates for all DevOpsMaestro resource types.
Copy any template as a starting point for your resource definitions.

## Usage

```bash
# 1. Copy a template below into a file
# 2. Fill in the values (remove fields you don't need)
# 3. Apply with dvm
dvm apply -f my-resource.yaml

# You can also combine multiple resources in one file separated by ---
```

All resources follow the Kubernetes-style structure: `apiVersion`, `kind`, `metadata`, and `spec`.
Required fields are marked with `# REQUIRED` in the templates below.

---

## Core Resources

### Ecosystem

The top-level organizational grouping (e.g., a company or platform).

```yaml
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: ""                        # REQUIRED - Unique ecosystem name
  labels:                         # Optional - Key-value labels for filtering
    team: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  theme: ""                       # Optional - Theme name applied to this ecosystem
  build:                          # Optional - Build configuration
    args:                         # Build arguments cascaded to all child resources
      KEY: "value"
  domains:                        # Optional - List of child domain names
    - ""
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique ecosystem name | [Ecosystem](ecosystem.md) |
| `metadata.labels` | map[string]string | No | Key-value labels for filtering | [Ecosystem](ecosystem.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [Ecosystem](ecosystem.md) |
| `spec.theme` | string | No | Theme name applied to this ecosystem | [Ecosystem](ecosystem.md) |
| `spec.build.args` | map[string]string | No | Build args cascaded down the hierarchy; overridden by Domain, App, or Workspace | [Ecosystem](ecosystem.md) |
| `spec.domains` | []string | No | List of child domain names | [Ecosystem](ecosystem.md) |

---

### Domain

A bounded context within an ecosystem (e.g., "backend", "frontend", "infra").

```yaml
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: ""                        # REQUIRED - Unique domain name
  ecosystem: ""                   # REQUIRED - Parent ecosystem name
  labels:                         # Optional - Key-value labels for filtering
    team: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  theme: ""                       # Optional - Theme name applied to this domain
  build:                          # Optional - Build configuration
    args:                         # Build arguments cascaded to all child resources
      KEY: "value"
  apps:                           # Optional - List of child app names
    - ""
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique domain name | [Domain](domain.md) |
| `metadata.ecosystem` | string | Yes | Parent ecosystem name | [Domain](domain.md) |
| `metadata.labels` | map[string]string | No | Key-value labels for filtering | [Domain](domain.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [Domain](domain.md) |
| `spec.theme` | string | No | Theme name applied to this domain | [Domain](domain.md) |
| `spec.build.args` | map[string]string | No | Build args cascaded down the hierarchy; overridden by App or Workspace | [Domain](domain.md) |
| `spec.apps` | []string | No | List of child app names | [Domain](domain.md) |

---

### App

An application or codebase within a domain.

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: ""                        # REQUIRED - Unique app name
  domain: ""                      # REQUIRED - Parent domain name
  labels:                         # Optional - Key-value labels for filtering
    language: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  path: ""                        # Optional - Path to the application source code
  theme: ""                       # Optional - Theme name applied to this app
  language:                       # Optional - Language configuration
    name: ""                      # Language name (golang, python, nodejs, rust, java, etc.)
    version: ""                   # Language version (e.g., "1.25", "3.12")
  build:                          # Optional - Build configuration
    dockerfile: ""                # Path to an existing Dockerfile
    buildpack: ""                 # Buildpack to use
    args:                         # Build arguments (key-value pairs)
      KEY: "value"
    target: ""                    # Multi-stage build target stage
    context: ""                   # Build context path (defaults to app path)
  dependencies:                   # Optional - Dependency configuration
    file: ""                      # Dependencies file (e.g., go.mod, requirements.txt)
    install: ""                   # Install command (e.g., "pip install -r requirements.txt")
    extra:                        # Extra dependency files to include
      - ""
  services:                       # Optional - Sidecar services
    - name: ""                    # Service name
      image: ""                   # Docker image
      version: ""                 # Image version/tag
      port: 0                     # Port number
      env:                        # Service environment variables
        KEY: "value"
  env:                            # Optional - App-level environment variables
    KEY: "value"
  ports:                          # Optional - Port mappings
    - "8080:8080"
  workspaces:                     # Optional - Associated workspace names
    - ""
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique app name | [App](app.md) |
| `metadata.domain` | string | Yes | Parent domain name | [App](app.md) |
| `spec.path` | string | No | Path to application source code | [App](app.md) |
| `spec.theme` | string | No | Theme name applied to this app | [App](app.md) |
| `spec.language` | object | No | Language name and version | [App](app.md) |
| `spec.build` | object | No | Dockerfile, buildpack, args, target, context | [App](app.md) |
| `spec.dependencies` | object | No | Dependency file, install command, extras | [App](app.md) |
| `spec.services` | []object | No | Sidecar services with name, image, version, port, env | [App](app.md) |
| `spec.env` | map[string]string | No | App-level environment variables | [App](app.md) |
| `spec.ports` | []string | No | Port mappings | [App](app.md) |
| `spec.workspaces` | []string | No | Associated workspace names | [App](app.md) |

---

### Workspace

A development environment configuration for an app. This is the most detailed resource type.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: ""                        # REQUIRED - Unique workspace name
  app: ""                         # REQUIRED - Parent app name
  labels:                         # Optional - Key-value labels for filtering
    environment: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  image:                          # Optional - Container image configuration
    name: ""                      # Custom image name
    buildFrom: ""                 # Build from source reference
    baseImage: ""                 # Base Docker image to use
  build:                          # Optional - Build-time configuration
    args:                         # Build arguments (emitted as ARG declarations — not ENV)
      KEY: "value"
    caCerts:                      # CA certificates to inject from MaestroVault
      - name: ""                  # REQUIRED - Certificate identifier (alphanumeric, _ -)
        vaultSecret: ""           # REQUIRED - MaestroVault secret name
        vaultEnvironment: ""      # Optional - Vault environment override
        vaultField: ""            # Optional - Field within vault secret (default: "cert")
    devStage:                     # Development stage customizations
      packages:                   # System packages to install (apt-get)
        - ""
      devTools:                   # Development tools to install
        - ""
      customCommands:             # Custom commands to run during build
        - ""
  shell:                          # Optional - Shell configuration
    type: ""                      # Shell type: zsh, bash, fish
    framework: ""                 # Shell framework (e.g., oh-my-zsh)
    theme: ""                     # Shell theme name
    plugins:                      # Shell plugins to install
      - ""
    customRc: ""                  # Custom shell rc content (appended to .zshrc/.bashrc)
  terminal:                       # Optional - Terminal configuration
    type: ""                      # Terminal emulator type
    configPath: ""                # Path to terminal config file
    autostart: false              # Auto-start terminal on attach
    prompt: ""                    # Prompt name reference (links to TerminalPrompt)
    plugins:                      # Terminal plugins
      - ""
    package: ""                   # Terminal package name reference (links to TerminalPackage)
  nvim:                           # Optional - Neovim configuration
    structure: ""                 # Nvim structure (e.g., lazyvim)
    theme: ""                     # Nvim theme name (links to NvimTheme)
    pluginPackage: ""             # Nvim package name reference (links to NvimPackage)
    plugins:                      # Individual plugin names to include
      - ""
    mergeMode: ""                 # How to merge plugins: append, replace
    customConfig: ""              # Custom nvim Lua configuration
  mounts:                         # Optional - Volume mounts
    - type: ""                    # Mount type: bind, volume, tmpfs
      source: ""                  # Source path or volume name
      destination: ""             # Container destination path
      readOnly: false             # Whether mount is read-only
  sshKey:                         # Optional - SSH key configuration
    mode: ""                      # SSH key mode
    path: ""                      # Path to SSH key file
  env:                            # Optional - Environment variables for the workspace
    KEY: "value"
  container:                      # Optional - Container runtime configuration
    user: ""                      # Container user name
    uid: 0                        # User ID
    gid: 0                        # Group ID
    workingDir: ""                # Working directory inside container
    command:                       # Container command (list of strings)
      - ""
    entrypoint:                    # Container entrypoint (list of strings)
      - ""
    resources:                    # Resource limits
      cpus: ""                    # CPU limit (e.g., "2")
      memory: ""                  # Memory limit (e.g., "4096m")
  gitrepo: ""                     # Optional - Git repository URL to clone
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique workspace name | [Workspace](workspace.md) |
| `metadata.app` | string | Yes | Parent app name | [Workspace](workspace.md) |
| `spec.image` | object | No | Container image: name, buildFrom, baseImage | [Workspace](workspace.md) |
| `spec.build` | object | No | Build args and dev stage customizations | [Workspace](workspace.md) |
| `spec.build.args` | map[string]string | No | Build arguments; emitted as `ARG` declarations (not `ENV`) — values with credentials are not persisted in the image | [Workspace](workspace.md) |
| `spec.build.caCerts` | []object | No | CA certificates to inject from MaestroVault | [Workspace](workspace.md) |
| `spec.build.caCerts[].name` | string | Yes | Certificate identifier; must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 10 certs | [Workspace](workspace.md) |
| `spec.build.caCerts[].vaultSecret` | string | Yes | MaestroVault secret name containing the PEM certificate | [Workspace](workspace.md) |
| `spec.build.caCerts[].vaultEnvironment` | string | No | Vault environment override (optional) | [Workspace](workspace.md) |
| `spec.build.caCerts[].vaultField` | string | No | Field within the vault secret (default: `"cert"`) | [Workspace](workspace.md) |
| `spec.shell` | object | No | Shell type, framework, theme, plugins, customRc | [Workspace](workspace.md) |
| `spec.terminal` | object | No | Terminal type, prompt, plugins, package | [Workspace](workspace.md) |
| `spec.nvim` | object | No | Neovim structure, theme, plugins, package | [Workspace](workspace.md) |
| `spec.mounts` | []object | No | Volume mounts with type, source, destination | [Workspace](workspace.md) |
| `spec.sshKey` | object | No | SSH key mode and path | [Workspace](workspace.md) |
| `spec.env` | map[string]string | No | Environment variables | [Workspace](workspace.md) |
| `spec.container` | object | No | User, UID/GID, workingDir, command, entrypoint, resources | [Workspace](workspace.md) |
| `spec.gitrepo` | string | No | Git repository URL to clone | [Workspace](workspace.md) |

---

### Credential

A secret reference scoped to a level in the hierarchy. Credentials are resolved from MaestroVault or environment variables at build and runtime.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: ""                        # REQUIRED - Credential name (used as default env var name)
  # REQUIRED - Exactly ONE scope field:
  ecosystem: ""                   # Scope to an ecosystem
  # domain: ""                    # Scope to a domain
  # app: ""                       # Scope to an app
  # workspace: ""                 # Scope to a workspace
spec:
  source: ""                      # REQUIRED - "vault" or "env"

  # --- Vault source fields ---
  vaultSecret: ""                 # Vault secret name in MaestroVault
  vaultEnvironment: ""            # Vault environment (e.g., "production", "staging")
  vaultUsernameSecret: ""         # Separate vault secret for username (dual-field)
  vaultFields:                    # Map ENV_VAR -> vault field name (v0.41.0+, max 50)
    ENV_VAR_NAME: "field_name"

  # --- Env source fields ---
  envVar: ""                      # Environment variable name to read from host

  # --- Common fields ---
  description: ""                 # Human-readable description
  usernameVar: ""                 # Env var name for username output (dual-field)
  passwordVar: ""                 # Env var name for password output (dual-field)
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Credential name (default env var name) | [Credential](credential.md) |
| `metadata.ecosystem` | string | One scope required | Scope to ecosystem | [Credential](credential.md) |
| `metadata.domain` | string | One scope required | Scope to domain | [Credential](credential.md) |
| `metadata.app` | string | One scope required | Scope to app | [Credential](credential.md) |
| `metadata.workspace` | string | One scope required | Scope to workspace | [Credential](credential.md) |
| `spec.source` | string | Yes | `"vault"` or `"env"` | [Credential](credential.md) |
| `spec.vaultSecret` | string | No | MaestroVault secret name | [Credential](credential.md) |
| `spec.vaultEnvironment` | string | No | MaestroVault environment | [Credential](credential.md) |
| `spec.vaultUsernameSecret` | string | No | Separate vault secret for username | [Credential](credential.md) |
| `spec.vaultFields` | map[string]string | No | ENV_VAR to vault field mapping (max 50) | [Credential](credential.md) |
| `spec.envVar` | string | No | Host environment variable name | [Credential](credential.md) |
| `spec.description` | string | No | Human-readable description | [Credential](credential.md) |
| `spec.usernameVar` | string | No | Env var for username (dual-field) | [Credential](credential.md) |
| `spec.passwordVar` | string | No | Env var for password (dual-field) | [Credential](credential.md) |

!!! note "Mutual Exclusivity"
    `vaultFields` cannot be combined with `usernameVar`, `passwordVar`, or `vaultUsernameSecret`.
    Use either vault fields mode OR dual-field mode, not both.

---

## Infrastructure Resources

### Registry

A local package registry for caching dependencies (OCI, Python, Go, npm, HTTP).

```yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: ""                        # REQUIRED - Unique registry name
  description: ""                 # Optional - Human-readable description
spec:
  type: ""                        # REQUIRED - zot, athens, devpi, verdaccio, or squid
  version: ""                     # Optional - Registry software version
  port: 0                         # Optional - Port number (defaults per type, see below)
  lifecycle: ""                   # Optional - persistent, on-demand, or manual
  config:                         # Optional - Registry-specific configuration
    key: "value"

# status:                         # READ-ONLY - Set by the system, not user-configurable
#   state: ""                     # Running state (running, stopped, etc.)
#   endpoint: ""                  # Access endpoint (e.g., http://localhost:5001)
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique registry name | [Registry](registry.md) |
| `metadata.description` | string | No | Human-readable description | [Registry](registry.md) |
| `spec.type` | string | Yes | Registry type: `zot`, `athens`, `devpi`, `verdaccio`, `squid` | [Registry](registry.md) |
| `spec.version` | string | No | Registry software version | [Registry](registry.md) |
| `spec.port` | int | No | Port number | [Registry](registry.md) |
| `spec.lifecycle` | string | No | `persistent`, `on-demand`, or `manual` | [Registry](registry.md) |
| `spec.config` | map[string]any | No | Registry-specific key-value configuration | [Registry](registry.md) |

**Default ports by type:**

| Type | Default Port | Purpose |
|------|-------------|---------|
| `zot` | 5001 | OCI container registry |
| `athens` | 3000 | Go module proxy |
| `devpi` | 3141 | Python package index |
| `verdaccio` | 4873 | npm registry |
| `squid` | 3128 | HTTP caching proxy |

---

## Nvim Resources

### NvimTheme

A Neovim colorscheme theme definition.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: ""                        # REQUIRED - Unique theme name
  description: ""                 # Optional - Human-readable description
  author: ""                      # Optional - Theme author
  category: ""                    # Optional - Category (e.g., "dark", "light")
spec:
  plugin:                         # Optional - Plugin that provides this theme
    repo: ""                      # GitHub repository (e.g., "folke/tokyonight.nvim")
    branch: ""                    # Git branch
    tag: ""                       # Git tag
  style: ""                       # Optional - Theme style variant (e.g., "storm", "night")
  transparent: false              # Optional - Enable transparent background
  colors:                         # Optional - Color overrides (hex values)
    bg: ""
    fg: ""
  options:                        # Optional - Theme-specific options
    key: "value"
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique theme name | [NvimTheme](nvim-theme.md) |
| `metadata.description` | string | No | Human-readable description | [NvimTheme](nvim-theme.md) |
| `metadata.author` | string | No | Theme author | [NvimTheme](nvim-theme.md) |
| `metadata.category` | string | No | Category (dark, light, etc.) | [NvimTheme](nvim-theme.md) |
| `spec.plugin` | object | No | Plugin repo, branch, and tag | [NvimTheme](nvim-theme.md) |
| `spec.style` | string | No | Theme style variant | [NvimTheme](nvim-theme.md) |
| `spec.transparent` | bool | No | Enable transparent background | [NvimTheme](nvim-theme.md) |
| `spec.colors` | map[string]string | No | Color overrides (hex values) | [NvimTheme](nvim-theme.md) |
| `spec.options` | map[string]any | No | Theme-specific options | [NvimTheme](nvim-theme.md) |

---

### NvimPlugin

A Neovim plugin configuration with lazy-loading support.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: ""                        # REQUIRED - Unique plugin name
  description: ""                 # Optional - Human-readable description
  category: ""                    # Optional - Category (e.g., "lsp", "ui", "editing")
  tags:                           # Optional - Tags for searching/filtering
    - ""
  labels:                         # Optional - Key-value labels
    language: ""
  annotations:                    # Optional - Non-identifying metadata
    source: ""
spec:
  repo: ""                        # REQUIRED - GitHub repository (e.g., "nvim-telescope/telescope.nvim")
  branch: ""                      # Optional - Git branch
  version: ""                     # Optional - Version constraint
  priority: 0                     # Optional - Load priority (lower = earlier)
  lazy: false                     # Optional - Lazy-load the plugin
  event:                          # Optional - Events that trigger loading (string or list)
    - ""
  ft:                             # Optional - Filetypes that trigger loading (string or list)
    - ""
  keys:                           # Optional - Key mappings that trigger loading
    - key: ""                     # Key sequence (e.g., "<leader>ff")
      mode: ""                    # Vim mode (e.g., "n", "v", "i")
      action: ""                  # Action to perform
      desc: ""                    # Description shown in which-key
  cmd:                            # Optional - Commands that trigger loading (string or list)
    - ""
  dependencies:                   # Optional - Plugin dependencies (strings or objects)
    - ""                          # Simple: just a plugin name
    # - repo: ""                  # Detailed: object with repo, build, version, branch, config
    #   build: ""
    #   version: ""
    #   branch: ""
    #   config: false
  build: ""                       # Optional - Build command to run after install
  config: ""                      # Optional - Lua configuration code
  init: ""                        # Optional - Lua init code (runs before plugin loads)
  opts: {}                        # Optional - Plugin options passed to setup()
  keymaps:                        # Optional - Additional key mappings
    - key: ""                     # Key sequence
      mode: ""                    # Vim mode
      action: ""                  # Action to perform
      desc: ""                    # Description
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique plugin name | [NvimPlugin](nvim-plugin.md) |
| `metadata.description` | string | No | Human-readable description | [NvimPlugin](nvim-plugin.md) |
| `metadata.category` | string | No | Category for organization | [NvimPlugin](nvim-plugin.md) |
| `metadata.tags` | []string | No | Tags for searching and filtering | [NvimPlugin](nvim-plugin.md) |
| `spec.repo` | string | Yes | GitHub repository path | [NvimPlugin](nvim-plugin.md) |
| `spec.branch` | string | No | Git branch | [NvimPlugin](nvim-plugin.md) |
| `spec.version` | string | No | Version constraint | [NvimPlugin](nvim-plugin.md) |
| `spec.priority` | int | No | Load priority (lower = earlier) | [NvimPlugin](nvim-plugin.md) |
| `spec.lazy` | bool | No | Lazy-load the plugin | [NvimPlugin](nvim-plugin.md) |
| `spec.event` | string or []string | No | Events that trigger loading | [NvimPlugin](nvim-plugin.md) |
| `spec.ft` | string or []string | No | Filetypes that trigger loading | [NvimPlugin](nvim-plugin.md) |
| `spec.keys` | []object | No | Key mappings that trigger loading | [NvimPlugin](nvim-plugin.md) |
| `spec.cmd` | string or []string | No | Commands that trigger loading | [NvimPlugin](nvim-plugin.md) |
| `spec.dependencies` | []string or []object | No | Plugin dependencies | [NvimPlugin](nvim-plugin.md) |
| `spec.build` | string | No | Build command after install | [NvimPlugin](nvim-plugin.md) |
| `spec.config` | string | No | Lua configuration code | [NvimPlugin](nvim-plugin.md) |
| `spec.init` | string | No | Lua init code (pre-load) | [NvimPlugin](nvim-plugin.md) |
| `spec.opts` | any | No | Options passed to setup() | [NvimPlugin](nvim-plugin.md) |
| `spec.keymaps` | []object | No | Additional key mappings | [NvimPlugin](nvim-plugin.md) |

---

### NvimPackage

A collection of related Neovim plugins with single inheritance.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: ""                        # REQUIRED - Unique package name
  description: ""                 # Optional - Human-readable description
  category: ""                    # Optional - Category (e.g., "language", "core")
  tags:                           # Optional - Tags for searching/filtering
    - ""
  labels:                         # Optional - Key-value labels
    language: ""
  annotations:                    # Optional - Non-identifying metadata
    source: ""
spec:
  extends: ""                     # Optional - Parent package name (single inheritance)
  plugins:                        # Optional - List of plugin names to include
    - ""
  enabled: true                   # Optional - Whether package is enabled (default: true)
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique package name | [NvimPackage](nvim-package.md) |
| `metadata.description` | string | No | Human-readable description | [NvimPackage](nvim-package.md) |
| `metadata.category` | string | No | Category for organization | [NvimPackage](nvim-package.md) |
| `metadata.tags` | []string | No | Tags for searching and filtering | [NvimPackage](nvim-package.md) |
| `metadata.labels` | map[string]string | No | Key-value labels | [NvimPackage](nvim-package.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [NvimPackage](nvim-package.md) |
| `spec.extends` | string | No | Parent package (single inheritance) | [NvimPackage](nvim-package.md) |
| `spec.plugins` | []string | No | Plugin names to include | [NvimPackage](nvim-package.md) |
| `spec.enabled` | bool | No | Whether package is enabled (default: true) | [NvimPackage](nvim-package.md) |

---

## Terminal Resources

### TerminalPrompt

A shell prompt configuration supporting Starship, Powerlevel10k, and Oh-My-Posh.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: ""                        # REQUIRED - Unique prompt name
  description: ""                 # Optional - Human-readable description
  category: ""                    # Optional - Category (e.g., "minimal", "powerline")
  tags:                           # Optional - Tags for searching/filtering
    - ""
  labels:                         # Optional - Key-value labels
    style: ""
  annotations:                    # Optional - Non-identifying metadata
    source: ""
spec:
  type: ""                        # REQUIRED - "starship", "powerlevel10k", or "oh-my-posh"
  addNewline: false               # Optional - Add newline before prompt
  palette: ""                     # Optional - Starship palette name
  format: ""                      # Optional - Prompt format string
  modules:                        # Optional - Module configurations (keyed by module name)
    git_branch:                   # Example module
      disabled: false             # Disable this module
      format: ""                  # Module format string
      style: ""                   # Module style (color/formatting)
      symbol: ""                  # Module symbol
      options:                    # Module-specific options
        key: "value"
  character:                      # Optional - Prompt character configuration
    success_symbol: ""            # Shown when last command succeeded
    error_symbol: ""              # Shown when last command failed
    vicmd_symbol: ""              # Shown in vi command mode
    viins_symbol: ""              # Shown in vi insert mode
  paletteRef: ""                  # Optional - Reference to a color palette
  colors:                         # Optional - Custom color overrides
    key: "#hexvalue"
  rawConfig: ""                   # Optional - Raw config for advanced users
  enabled: true                   # Optional - Whether prompt is enabled (default: true)
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique prompt name | [TerminalPrompt](terminal-prompt.md) |
| `metadata.description` | string | No | Human-readable description | [TerminalPrompt](terminal-prompt.md) |
| `metadata.category` | string | No | Category for organization | [TerminalPrompt](terminal-prompt.md) |
| `metadata.tags` | []string | No | Tags for filtering | [TerminalPrompt](terminal-prompt.md) |
| `metadata.labels` | map[string]string | No | Key-value labels | [TerminalPrompt](terminal-prompt.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [TerminalPrompt](terminal-prompt.md) |
| `spec.type` | string | Yes | `starship`, `powerlevel10k`, or `oh-my-posh` | [TerminalPrompt](terminal-prompt.md) |
| `spec.addNewline` | bool | No | Add newline before prompt | [TerminalPrompt](terminal-prompt.md) |
| `spec.palette` | string | No | Starship palette name | [TerminalPrompt](terminal-prompt.md) |
| `spec.format` | string | No | Prompt format string | [TerminalPrompt](terminal-prompt.md) |
| `spec.modules` | map[string]ModuleConfig | No | Per-module configuration | [TerminalPrompt](terminal-prompt.md) |
| `spec.character` | object | No | Prompt character symbols | [TerminalPrompt](terminal-prompt.md) |
| `spec.paletteRef` | string | No | Color palette reference | [TerminalPrompt](terminal-prompt.md) |
| `spec.colors` | map[string]string | No | Custom color overrides | [TerminalPrompt](terminal-prompt.md) |
| `spec.rawConfig` | string | No | Raw config for advanced use | [TerminalPrompt](terminal-prompt.md) |
| `spec.enabled` | bool | No | Whether enabled (default: true) | [TerminalPrompt](terminal-prompt.md) |

---

### TerminalPackage

A collection of terminal configuration: shell plugins, prompts, profiles, and optional WezTerm settings.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: ""                        # REQUIRED - Unique package name
  description: ""                 # Optional - Human-readable description
  category: ""                    # Optional - Category (e.g., "development", "devops")
  tags:                           # Optional - Tags for searching/filtering
    - ""
  labels:                         # Optional - Key-value labels
    shell: ""
  annotations:                    # Optional - Non-identifying metadata
    source: ""
spec:
  extends: ""                     # Optional - Parent package name (single inheritance)
  plugins:                        # Optional - Shell plugin names
    - ""
  prompts:                        # Optional - Prompt names to include
    - ""
  profiles:                       # Optional - Profile preset names
    - ""
  wezterm:                        # Optional - Embedded WezTerm configuration
    fontSize: 0                   # Font size
    colorScheme: ""               # Color scheme name
    fontFamily: ""                # Font family name
  promptStyle: ""                 # Optional - Modular prompt style name
  promptExtensions:               # Optional - Prompt extension names
    - ""
  enabled: true                   # Optional - Whether package is enabled (default: true)
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique package name | [WeztermConfig](wezterm-config.md) |
| `metadata.description` | string | No | Human-readable description | [WeztermConfig](wezterm-config.md) |
| `metadata.category` | string | No | Category for organization | [WeztermConfig](wezterm-config.md) |
| `metadata.tags` | []string | No | Tags for filtering | [WeztermConfig](wezterm-config.md) |
| `metadata.labels` | map[string]string | No | Key-value labels | [WeztermConfig](wezterm-config.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [WeztermConfig](wezterm-config.md) |
| `spec.extends` | string | No | Parent package (single inheritance) | [WeztermConfig](wezterm-config.md) |
| `spec.plugins` | []string | No | Shell plugin names | [WeztermConfig](wezterm-config.md) |
| `spec.prompts` | []string | No | Prompt names to include | [WeztermConfig](wezterm-config.md) |
| `spec.profiles` | []string | No | Profile preset names | [WeztermConfig](wezterm-config.md) |
| `spec.wezterm` | object | No | WezTerm config: fontSize, colorScheme, fontFamily | [WeztermConfig](wezterm-config.md) |
| `spec.promptStyle` | string | No | Modular prompt style name | [WeztermConfig](wezterm-config.md) |
| `spec.promptExtensions` | []string | No | Prompt extension names | [WeztermConfig](wezterm-config.md) |
| `spec.enabled` | bool | No | Whether enabled (default: true) | [WeztermConfig](wezterm-config.md) |

---

### TerminalEmulator

A terminal emulator configuration (WezTerm, Alacritty, Kitty).

!!! warning "Not Yet Available via `dvm apply`"
    TerminalEmulator has YAML types defined but is not currently registered with `dvm apply -f`.
    This template is provided for reference and future compatibility.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  name: ""                        # REQUIRED - Unique emulator config name
  description: ""                 # Optional - Human-readable description
  category: ""                    # Optional - Category
  labels:                         # Optional - Key-value labels
    emulator: ""
  annotations:                    # Optional - Non-identifying metadata
    source: ""
spec:
  type: ""                        # REQUIRED - "wezterm", "alacritty", or "kitty"
  config:                         # Optional - Emulator-specific configuration
    key: "value"
  themeRef: ""                    # Optional - Reference to a theme name
  workspace: ""                   # Optional - Associated workspace name
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique emulator config name | [WeztermConfig](wezterm-config.md) |
| `metadata.description` | string | No | Human-readable description | [WeztermConfig](wezterm-config.md) |
| `metadata.category` | string | No | Category for organization | [WeztermConfig](wezterm-config.md) |
| `metadata.labels` | map[string]string | No | Key-value labels | [WeztermConfig](wezterm-config.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [WeztermConfig](wezterm-config.md) |
| `spec.type` | string | Yes | `wezterm`, `alacritty`, or `kitty` | [WeztermConfig](wezterm-config.md) |
| `spec.config` | map[string]any | No | Emulator-specific configuration | [WeztermConfig](wezterm-config.md) |
| `spec.themeRef` | string | No | Reference to a theme | [WeztermConfig](wezterm-config.md) |
| `spec.workspace` | string | No | Associated workspace name | [WeztermConfig](wezterm-config.md) |

---

## Extensibility Resources

### CustomResourceDefinition

Register a custom resource type to extend DevOpsMaestro with your own kinds.

```yaml
apiVersion: devopsmaestro.io/v1alpha1          # NOTE: v1alpha1, not v1
kind: CustomResourceDefinition
metadata:
  name: ""                        # REQUIRED - CRD name (typically plural form)
spec:
  group: ""                       # REQUIRED - API group (e.g., "mycompany.io")
  names:                          # REQUIRED - Resource naming
    kind: ""                      # REQUIRED - Resource kind (e.g., "DatabaseConfig")
    singular: ""                  # REQUIRED - Singular name (e.g., "databaseconfig")
    plural: ""                    # REQUIRED - Plural name (e.g., "databaseconfigs")
    shortNames:                   # Optional - Short aliases for CLI
      - ""
  scope: ""                       # REQUIRED - "Global", "Ecosystem", "Domain", "App", or "Workspace"
  versions:                       # REQUIRED - At least one version
    - name: ""                    # Version name (e.g., "v1")
      served: true                # Whether this version is served by the API
      storage: true               # Whether this is the storage version
      schema: {}                  # Optional - JSON Schema for validation
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | CRD name | [CRD](custom-resource-definition.md) |
| `spec.group` | string | Yes | API group for the custom resource | [CRD](custom-resource-definition.md) |
| `spec.names.kind` | string | Yes | Resource kind name | [CRD](custom-resource-definition.md) |
| `spec.names.singular` | string | Yes | Singular name | [CRD](custom-resource-definition.md) |
| `spec.names.plural` | string | Yes | Plural name | [CRD](custom-resource-definition.md) |
| `spec.names.shortNames` | []string | No | Short aliases for CLI use | [CRD](custom-resource-definition.md) |
| `spec.scope` | string | Yes | `Global`, `Ecosystem`, `Domain`, `App`, or `Workspace` | [CRD](custom-resource-definition.md) |
| `spec.versions` | []object | Yes | Version definitions with name, served, storage | [CRD](custom-resource-definition.md) |

!!! note "Built-in Kind Restrictions"
    CRD kind names cannot collide with any of the 15 built-in kinds:
    Ecosystem, Domain, App, Workspace, Credential, Registry,
    NvimTheme, NvimPlugin, NvimPackage,
    TerminalPrompt, TerminalPackage, TerminalPlugin, TerminalEmulator,
    CustomResourceDefinition, GitRepo.

---

## See Also

- [YAML Reference Overview](index.md) -- Resource type descriptions and hierarchy
- [YAML Schema](../configuration/yaml-schema.md) -- Schema validation rules
- [Commands Reference](../dvm/commands.md) -- CLI commands including `dvm apply`
