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
Full reference: [ecosystem.md](ecosystem.md)

```yaml
# Full reference: https://devopsmaestro.io/reference/ecosystem
apiVersion: devopsmaestro.io/v1   # REQUIRED
kind: Ecosystem                   # REQUIRED
metadata:
  name: ""                        # REQUIRED - Unique ecosystem name (DNS subdomain)
  labels:                         # Optional - Key-value labels for filtering/organization
    team: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  description: ""                 # Optional - Human-readable description
  theme: ""                       # Optional - Default theme cascaded to all workspaces
                                  #   e.g. coolnight-ocean, tokyonight-night, gruvbox-dark
  nvimPackage: ""                 # Optional - Default NvimPackage cascaded to all workspaces
  terminalPackage: ""             # Optional - Default TerminalPackage cascaded to all workspaces
  build:                          # Optional - Build configuration inherited by all workspaces
    args:                         # Optional - Build args passed as --build-arg; cascades down
      KEY: "value"                #   global < ecosystem < domain < app < workspace
  caCerts:                        # Optional - CA certs cascaded to all workspace builds
    - name: ""                    # REQUIRED per cert - alphanumeric/_/- only; max 64 chars
      vaultSecret: ""             # REQUIRED per cert - MaestroVault secret name (PEM)
      vaultEnvironment: ""        # Optional - Vault environment override
      vaultField: ""              # Optional - Field within secret (default: "cert")
  domains:                        # Optional - Child domain names (populated by dvm get -o yaml)
    - ""
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique ecosystem name | [Ecosystem](ecosystem.md) |
| `metadata.labels` | map[string]string | No | Key-value labels for filtering | [Ecosystem](ecosystem.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [Ecosystem](ecosystem.md) |
| `spec.description` | string | No | Human-readable description | [Ecosystem](ecosystem.md) |
| `spec.theme` | string | No | Default theme cascaded to all workspaces | [Ecosystem](ecosystem.md) |
| `spec.nvimPackage` | string | No | Default NvimPackage cascaded to all workspaces | [Ecosystem](ecosystem.md) |
| `spec.terminalPackage` | string | No | Default TerminalPackage cascaded to all workspaces | [Ecosystem](ecosystem.md) |
| `spec.build.args` | map[string]string | No | Build args cascaded down the hierarchy; overridden by Domain, App, or Workspace | [Ecosystem](ecosystem.md) |
| `spec.caCerts[].name` | string | Yes (per cert) | Cert name; must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$` | [Ecosystem](ecosystem.md) |
| `spec.caCerts[].vaultSecret` | string | Yes (per cert) | MaestroVault secret name containing PEM | [Ecosystem](ecosystem.md) |
| `spec.caCerts[].vaultEnvironment` | string | No | Vault environment override | [Ecosystem](ecosystem.md) |
| `spec.caCerts[].vaultField` | string | No | Field within secret (default: `"cert"`) | [Ecosystem](ecosystem.md) |
| `spec.domains` | []string | No | List of child domain names | [Ecosystem](ecosystem.md) |

---

### Domain

A bounded context within an ecosystem (e.g., "backend", "frontend", "infra").  
Full reference: [domain.md](domain.md)

```yaml
# Full reference: https://devopsmaestro.io/reference/domain
apiVersion: devopsmaestro.io/v1   # REQUIRED
kind: Domain                      # REQUIRED
metadata:
  name: ""                        # REQUIRED - Unique domain name (DNS subdomain)
  ecosystem: ""                   # REQUIRED - Parent ecosystem name
  labels:                         # Optional - Key-value labels for filtering/organization
    team: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  theme: ""                       # Optional - Default theme; overrides ecosystem theme
                                  #   Theme hierarchy: Workspace > App > Domain > Ecosystem
  nvimPackage: ""                 # Optional - Default NvimPackage; overrides ecosystem nvimPackage
  terminalPackage: ""             # Optional - Default TerminalPackage; overrides ecosystem terminalPackage
  build:                          # Optional - Build configuration inherited by all workspaces
    args:                         # Optional - Build args; overrides ecosystem-level args
      KEY: "value"                #   global < ecosystem < domain < app < workspace
  caCerts:                        # Optional - CA certs cascaded to all workspace builds
    - name: ""                    # REQUIRED per cert - alphanumeric/_/- only; max 64 chars
      vaultSecret: ""             # REQUIRED per cert - MaestroVault secret name (PEM)
      vaultEnvironment: ""        # Optional - Vault environment override
      vaultField: ""              # Optional - Field within secret (default: "cert")
  apps:                           # Optional - Child app names (populated by dvm get -o yaml)
    - ""
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique domain name | [Domain](domain.md) |
| `metadata.ecosystem` | string | Yes | Parent ecosystem name | [Domain](domain.md) |
| `metadata.labels` | map[string]string | No | Key-value labels for filtering | [Domain](domain.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [Domain](domain.md) |
| `spec.theme` | string | No | Default theme; overrides ecosystem theme | [Domain](domain.md) |
| `spec.nvimPackage` | string | No | Default NvimPackage; overrides ecosystem nvimPackage | [Domain](domain.md) |
| `spec.terminalPackage` | string | No | Default TerminalPackage; overrides ecosystem terminalPackage | [Domain](domain.md) |
| `spec.build.args` | map[string]string | No | Build args cascaded down the hierarchy; overridden by App or Workspace | [Domain](domain.md) |
| `spec.caCerts[].name` | string | Yes (per cert) | Cert name; must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$` | [Domain](domain.md) |
| `spec.caCerts[].vaultSecret` | string | Yes (per cert) | MaestroVault secret name containing PEM | [Domain](domain.md) |
| `spec.caCerts[].vaultEnvironment` | string | No | Vault environment override | [Domain](domain.md) |
| `spec.caCerts[].vaultField` | string | No | Field within secret (default: `"cert"`) | [Domain](domain.md) |
| `spec.apps` | []string | No | List of child app names | [Domain](domain.md) |

---

### App

An application or codebase within a domain.  
Full reference: [app.md](app.md)

```yaml
# Full reference: https://devopsmaestro.io/reference/app
apiVersion: devopsmaestro.io/v1   # REQUIRED
kind: App                         # REQUIRED
metadata:
  name: ""                        # REQUIRED - Unique app name (DNS subdomain)
  domain: ""                      # REQUIRED - Parent domain name
  ecosystem: ""                   # Optional - Parent ecosystem; enables context-free apply
  labels:                         # Optional - Key-value labels for filtering/organization
    language: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  path: ""                        # REQUIRED - Absolute path to source code on local filesystem
                                  #   Variable substitution supported: ${HOME}/projects/my-app
  theme: ""                       # Optional - Default theme for workspaces; overrides domain theme
  nvimPackage: ""                 # Optional - Default NvimPackage for workspaces in this app
  terminalPackage: ""             # Optional - Default TerminalPackage for workspaces in this app
  gitRepo: ""                     # Optional - GitRepo resource name to associate with this app
  language:                       # Optional - Language/runtime configuration
    name: ""                      # Language: go, python, node, rust, java, dotnet
    version: ""                   # Version string e.g. "1.22", "3.11", "20"
  build:                          # Optional - Build/containerization configuration
    dockerfile: ""                # Optional - Path to existing Dockerfile
    buildpack: ""                 # Optional - Buildpack: auto, go, python, node, rust, java
    target: ""                    # Optional - Multi-stage Dockerfile build target
    context: ""                   # Optional - Build context path (default: app path)
    args:                         # Optional - Build args emitted as ARG (not ENV)
      KEY: "value"
    caCerts:                      # Optional - CA certs fetched from MaestroVault at build time
      - name: ""                  # REQUIRED per cert - alphanumeric/_/- only; max 64 chars
        vaultSecret: ""           # REQUIRED per cert - MaestroVault secret name (PEM)
        vaultEnvironment: ""      # Optional - Vault environment override
        vaultField: ""            # Optional - Field within secret (default: "cert")
  dependencies:                   # Optional - Dependency management
    file: ""                      # Dependency manifest: go.mod, requirements.txt, package.json
    install: ""                   # Install command: "go mod download", "pip install -r ..."
    extra:                        # Optional - Additional packages/modules to install
      - ""
  services:                       # Optional - Sidecar services (databases, caches, etc.)
    - name: ""                    # REQUIRED per service - e.g. postgres, redis, mongodb
      image: ""                   # Optional - Custom Docker image (default: official image)
      version: ""                 # Optional - Image version/tag e.g. "15", "7"
      port: 0                     # Optional - Port to expose
      env:                        # Optional - Service environment variables
        KEY: "value"
  ports:                          # Optional - Port mappings the app exposes
    - "8080:8080"                 #   Format: "host:container"
  env:                            # Optional - App-level environment variables
    KEY: "value"
  workspaces:                     # Optional - Child workspace names (populated by dvm get -o yaml)
    - ""
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique app name | [App](app.md) |
| `metadata.domain` | string | Yes | Parent domain name | [App](app.md) |
| `metadata.ecosystem` | string | No | Parent ecosystem; enables context-free apply | [App](app.md) |
| `spec.path` | string | Yes | Absolute path to source code | [App](app.md) |
| `spec.theme` | string | No | Default theme for workspaces in this app | [App](app.md) |
| `spec.nvimPackage` | string | No | Default NvimPackage for workspaces | [App](app.md) |
| `spec.terminalPackage` | string | No | Default TerminalPackage for workspaces | [App](app.md) |
| `spec.gitRepo` | string | No | GitRepo resource name to associate | [App](app.md) |
| `spec.language` | object | No | Language name and version | [App](app.md) |
| `spec.build` | object | No | Dockerfile, buildpack, args, target, context, caCerts | [App](app.md) |
| `spec.dependencies` | object | No | Dependency file, install command, extras | [App](app.md) |
| `spec.services` | []object | No | Sidecar services with name, image, version, port, env | [App](app.md) |
| `spec.ports` | []string | No | Port mappings (host:container) | [App](app.md) |
| `spec.env` | map[string]string | No | App-level environment variables | [App](app.md) |
| `spec.workspaces` | []string | No | Child workspace names | [App](app.md) |

---

### Workspace

A development environment configuration for an app. This is the most detailed resource type.  
Full reference: [workspace.md](workspace.md)

```yaml
# Full reference: https://devopsmaestro.io/reference/workspace
apiVersion: devopsmaestro.io/v1   # REQUIRED
kind: Workspace                   # REQUIRED
metadata:
  name: ""                        # REQUIRED - Unique workspace name (DNS subdomain)
  app: ""                         # REQUIRED - Parent app name
  domain: ""                      # Optional - Parent domain; enables context-free apply
  ecosystem: ""                   # Optional - Parent ecosystem; used with domain to fully disambiguate
  labels:                         # Optional - Key-value labels for filtering/organization
    environment: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:

  # ---------------------------------------------------------------------------
  # IMAGE — container image to build or use
  # ---------------------------------------------------------------------------
  image:                          # Optional - Container image configuration
    name: ""                      # Optional - Image name (auto-generated if omitted)
    buildFrom: ""                 # Optional - Dockerfile path to build from (e.g. ./Dockerfile)
    baseImage: ""                 # Optional - Pre-built base image (skips build stage)

  # ---------------------------------------------------------------------------
  # BUILD — what goes into the image at build time
  # ---------------------------------------------------------------------------
  build:                          # Optional - Build configuration
    args:                         # Optional - Build args emitted as ARG (not ENV); not persisted in layers
      KEY: "value"
    caCerts:                      # Optional - CA certs injected from MaestroVault at build time
      - name: ""                  # REQUIRED per cert - alphanumeric/_/- only; max 64 chars
        vaultSecret: ""           # REQUIRED per cert - MaestroVault secret name (PEM)
        vaultEnvironment: ""      # Optional - Vault environment override
        vaultField: ""            # Optional - Field within secret (default: "cert")
    baseStage:                    # Optional - Packages installed in the base (app) build stage
      packages:                   # System packages for the app runtime (e.g. libpq-dev)
        - ""
    devStage:                     # Optional - Developer tooling added on top of the base stage
      packages:                   # System packages for the dev layer (e.g. ripgrep, fd-find)
        - ""
      devTools:                   # Language dev tools (e.g. gopls, delve, pylsp, typescript-language-server)
        - ""
      customCommands:             # Arbitrary shell commands run during the dev stage build
        - ""

  # ---------------------------------------------------------------------------
  # SHELL — login shell inside the container
  # ---------------------------------------------------------------------------
  shell:                          # Optional - Shell configuration
    type: ""                      # Shell type: zsh, bash
    framework: ""                 # Shell framework: oh-my-zsh, prezto
    theme: ""                     # Prompt theme: starship, powerlevel10k, agnoster
    plugins:                      # Shell plugins to install
      - ""
    customRc: ""                  # Raw RC content appended to .zshrc / .bashrc

  # ---------------------------------------------------------------------------
  # TERMINAL — multiplexer configuration
  # ---------------------------------------------------------------------------
  terminal:                       # Optional - Terminal multiplexer configuration
    type: ""                      # Multiplexer: tmux, zellij, screen
    configPath: ""                # Host path to config file to mount (e.g. ~/.tmux.conf)
    autostart: false              # Start multiplexer automatically on container attach
    prompt: ""                    # TerminalPrompt resource name
    plugins:                      # Terminal plugin names to install
      - ""
    package: ""                   # TerminalPackage resource name

  # ---------------------------------------------------------------------------
  # NVIM — Neovim editor configuration
  # ---------------------------------------------------------------------------
  nvim:                           # Optional - Neovim configuration
    structure: ""                 # Distribution: lazyvim, custom, nvchad, astronvim
    theme: ""                     # Theme name (overrides app/domain/ecosystem theme for nvim)
    pluginPackage: ""             # NvimPackage resource name (pre-configured plugin collection)
    plugins:                      # Individual NvimPlugin names to include
      - ""
    mergeMode: ""                 # Plugin merge strategy: append (default), replace
    customConfig: ""              # Raw Lua config injected into the nvim setup
    extraMasonTools:              # Additional Mason tools installed at image build time
      - ""                        #   e.g. lua-language-server, stylua, prettier
    extraTreesitterParsers:       # Additional Treesitter parsers compiled at image build time
      - ""                        #   e.g. go, python, typescript, lua

  # ---------------------------------------------------------------------------
  # TOOLS — optional binary tools installed at build time
  # ---------------------------------------------------------------------------
  tools:                          # Optional - Workspace-level tool binaries (all default false)
    opencode: false               # Install opencode AI assistant CLI (linux/amd64, linux/arm64)

  # ---------------------------------------------------------------------------
  # MOUNTS — filesystem mounts into the container
  # ---------------------------------------------------------------------------
  mounts:                         # Optional - Container mount points
    - type: ""                    # REQUIRED per mount - bind, volume, tmpfs
      source: ""                  # REQUIRED per mount - host path or volume name
      destination: ""             # REQUIRED per mount - container destination path
      readOnly: false             # Optional - Mount as read-only (default: false)

  # ---------------------------------------------------------------------------
  # SSH KEY — how SSH keys reach the container
  # ---------------------------------------------------------------------------
  sshKey:                         # Optional - SSH key configuration
    mode: ""                      # REQUIRED if present - mount_host, global_dvm, per_project, generate
    path: ""                      # Optional - Host path; used when mode=mount_host

  # ---------------------------------------------------------------------------
  # ENV — environment variables injected at container start
  # ---------------------------------------------------------------------------
  env:                            # Optional - Workspace environment variables
    KEY: "value"

  # ---------------------------------------------------------------------------
  # CONTAINER — runtime settings
  # ---------------------------------------------------------------------------
  container:                      # Optional - Container runtime configuration
    user: ""                      # Container username (sets USER in Dockerfile; default: dev)
    uid: 0                        # User ID (default: 1000)
    gid: 0                        # Group ID (default: 1000)
    workingDir: ""                # Working directory inside container (default: /workspace)
    command:                      # Container command override (default: ["/bin/zsh", "-l"])
      - ""
    entrypoint:                   # Container entrypoint override
      - ""
    sshAgentForwarding: false     # Forward SSH agent socket (SSH_AUTH_SOCK) into container
    networkMode: ""               # Docker network mode: bridge (default), host, none
    resources:                    # Optional - Resource limits
      cpus: ""                    # CPU limit e.g. "2.0"
      memory: ""                  # Memory limit e.g. "4G", "512m"

  # ---------------------------------------------------------------------------
  # GITREPO — repository to clone into the workspace on creation
  # ---------------------------------------------------------------------------
  gitrepo: ""                     # Optional - GitRepo resource name to clone on first create
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique workspace name | [Workspace](workspace.md) |
| `metadata.app` | string | Yes | Parent app name | [Workspace](workspace.md) |
| `metadata.domain` | string | No | Parent domain; enables context-free apply | [Workspace](workspace.md) |
| `metadata.ecosystem` | string | No | Parent ecosystem; used with domain to fully disambiguate | [Workspace](workspace.md) |
| `spec.image` | object | No | Container image: name, buildFrom, baseImage | [Workspace](workspace.md) |
| `spec.build.args` | map[string]string | No | Build args emitted as `ARG` (not `ENV`) | [Workspace](workspace.md) |
| `spec.build.caCerts` | []object | No | CA certs injected from MaestroVault at build time | [Workspace](workspace.md) |
| `spec.build.caCerts[].name` | string | Yes (per cert) | Cert name; `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars; max 10 certs | [Workspace](workspace.md) |
| `spec.build.caCerts[].vaultSecret` | string | Yes (per cert) | MaestroVault secret containing PEM | [Workspace](workspace.md) |
| `spec.build.caCerts[].vaultEnvironment` | string | No | Vault environment override | [Workspace](workspace.md) |
| `spec.build.caCerts[].vaultField` | string | No | Field within secret (default: `"cert"`) | [Workspace](workspace.md) |
| `spec.build.baseStage.packages` | []string | No | System packages installed in the base (app) build stage | [Workspace](workspace.md) |
| `spec.build.devStage.packages` | []string | No | System packages installed in the dev layer | [Workspace](workspace.md) |
| `spec.build.devStage.devTools` | []string | No | Language-specific dev tools (gopls, delve, pylsp…) | [Workspace](workspace.md) |
| `spec.build.devStage.customCommands` | []string | No | Arbitrary shell commands run during dev stage build | [Workspace](workspace.md) |
| `spec.shell` | object | No | type, framework, theme, plugins, customRc | [Workspace](workspace.md) |
| `spec.terminal` | object | No | type, configPath, autostart, prompt, plugins, package | [Workspace](workspace.md) |
| `spec.nvim.structure` | string | No | `lazyvim`, `custom`, `nvchad`, `astronvim` | [Workspace](workspace.md) |
| `spec.nvim.theme` | string | No | Theme name for nvim | [Workspace](workspace.md) |
| `spec.nvim.pluginPackage` | string | No | NvimPackage resource name | [Workspace](workspace.md) |
| `spec.nvim.plugins` | []string | No | Individual NvimPlugin names | [Workspace](workspace.md) |
| `spec.nvim.mergeMode` | string | No | `append` (default) or `replace` | [Workspace](workspace.md) |
| `spec.nvim.customConfig` | string | No | Raw Lua configuration | [Workspace](workspace.md) |
| `spec.nvim.extraMasonTools` | []string | No | Additional Mason tools installed at image build time | [Workspace](workspace.md) |
| `spec.nvim.extraTreesitterParsers` | []string | No | Additional Treesitter parsers compiled at image build time | [Workspace](workspace.md) |
| `spec.tools.opencode` | bool | No | Install opencode AI assistant CLI (default: `false`) | [Workspace](workspace.md) |
| `spec.mounts` | []object | No | Volume mounts with type, source, destination, readOnly | [Workspace](workspace.md) |
| `spec.sshKey` | object | No | mode and path | [Workspace](workspace.md) |
| `spec.env` | map[string]string | No | Environment variables injected at container start | [Workspace](workspace.md) |
| `spec.container.user` | string | No | Container username (default: `dev`) | [Workspace](workspace.md) |
| `spec.container.uid` | int | No | User ID (default: `1000`) | [Workspace](workspace.md) |
| `spec.container.gid` | int | No | Group ID (default: `1000`) | [Workspace](workspace.md) |
| `spec.container.workingDir` | string | No | Working directory (default: `/workspace`) | [Workspace](workspace.md) |
| `spec.container.command` | []string | No | Container command (default: `["/bin/zsh", "-l"]`) | [Workspace](workspace.md) |
| `spec.container.entrypoint` | []string | No | Container entrypoint override | [Workspace](workspace.md) |
| `spec.container.sshAgentForwarding` | bool | No | Forward SSH agent socket into container | [Workspace](workspace.md) |
| `spec.container.networkMode` | string | No | Docker network mode: `bridge`, `host`, `none` | [Workspace](workspace.md) |
| `spec.container.resources.cpus` | string | No | CPU limit (e.g., `"2.0"`) | [Workspace](workspace.md) |
| `spec.container.resources.memory` | string | No | Memory limit (e.g., `"4G"`) | [Workspace](workspace.md) |
| `spec.gitrepo` | string | No | GitRepo resource name to clone on first create | [Workspace](workspace.md) |

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
  version: ""                     # Optional - Registry software version (semver, e.g., "2.1.15")
  enabled: true                   # Optional - Whether registry is enabled (default: true)
  port: 0                         # Optional - Port number (0 = use type default; must be 1024-65535)
  lifecycle: ""                   # Optional - persistent, on-demand, or manual (default: manual)
  storage: ""                     # Optional - Storage path (default: type-specific, e.g., /var/lib/zot)
  idleTimeout: 0                  # Optional - Seconds before auto-stop (on-demand only; default: 1800; min: 60)
  config:                         # Optional - Registry-specific configuration
    key: "value"

# status:                         # READ-ONLY - Set by the system, not user-configurable
#   state: ""                     # Running state (running, stopped, starting, error)
#   endpoint: ""                  # Access endpoint (e.g., http://localhost:5001)
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique registry name | [Registry](registry.md) |
| `metadata.description` | string | No | Human-readable description | [Registry](registry.md) |
| `spec.type` | string | Yes | Registry type: `zot`, `athens`, `devpi`, `verdaccio`, `squid` | [Registry](registry.md) |
| `spec.version` | string | No | Registry software version | [Registry](registry.md) |
| `spec.enabled` | bool | No | Whether registry is enabled (default: `true`) | [Registry](registry.md) |
| `spec.port` | int | No | Port number (0 = type default; range 1024–65535) | [Registry](registry.md) |
| `spec.lifecycle` | string | No | `persistent`, `on-demand`, or `manual` (default: `manual`) | [Registry](registry.md) |
| `spec.storage` | string | No | Storage path (default: type-specific) | [Registry](registry.md) |
| `spec.idleTimeout` | int | No | Seconds before auto-stop for `on-demand` (default: `1800`; min: `60`) | [Registry](registry.md) |
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

### GitRepo

A remote git repository mirrored locally for fast, offline-capable workspace builds.

```yaml
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: ""                        # REQUIRED - Unique repository name
  labels:                         # Optional - Key-value labels for filtering
    team: ""
    language: ""
  annotations:                    # Optional - Non-identifying metadata
    description: ""
spec:
  url: ""                         # REQUIRED - Remote repository URL (HTTPS or SSH)
  defaultRef: ""                  # Optional - Default branch/tag to check out (default: "main")
  authType: ""                    # Optional - none, ssh, or basic (default: "none")
  credential: ""                  # Optional - Credential name for private repo authentication
  autoSync: false                 # Optional - Automatically sync mirror on a schedule (default: false)
  syncIntervalMinutes: 0          # Optional - Sync interval in minutes when autoSync is true
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique repository name | [GitRepo](gitrepo.md) |
| `metadata.labels` | map[string]string | No | Key-value labels for filtering | [GitRepo](gitrepo.md) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [GitRepo](gitrepo.md) |
| `spec.url` | string | Yes | Remote repository URL (HTTPS or SSH) | [GitRepo](gitrepo.md) |
| `spec.defaultRef` | string | No | Default branch or tag (default: `main`) | [GitRepo](gitrepo.md) |
| `spec.authType` | string | No | `none`, `ssh`, or `basic` (default: `none`) | [GitRepo](gitrepo.md) |
| `spec.credential` | string | No | Credential resource name for authentication | [GitRepo](gitrepo.md) |
| `spec.autoSync` | bool | No | Periodically sync the local mirror (default: `false`) | [GitRepo](gitrepo.md) |
| `spec.syncIntervalMinutes` | int | No | Sync frequency in minutes (requires `autoSync: true`) | [GitRepo](gitrepo.md) |

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
  plugin:                         # Optional - Plugin that provides this theme (omit for standalone)
    repo: ""                      # GitHub repository (e.g., "folke/tokyonight.nvim")
    branch: ""                    # Git branch
    tag: ""                       # Git tag
  style: ""                       # Optional - Theme style variant (e.g., "storm", "night")
  transparent: false              # Optional - Enable transparent background
  colors:                         # Optional - Color overrides (hex values); REQUIRED for standalone themes
    bg: ""
    fg: ""
  promptColors:                   # Optional - Starship prompt segment color overrides (hex values)
    directory: ""
    git_branch: ""
  options:                        # Optional - Theme-specific key-value options (plugin-defined)
    key: "value"
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique theme name | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `metadata.description` | string | No | Human-readable description | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `metadata.author` | string | No | Theme author | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `metadata.category` | string | No | Category (dark, light, etc.) | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `spec.plugin` | object | No | Plugin repo, branch, and tag (omit for standalone) | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `spec.style` | string | No | Theme style variant | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `spec.transparent` | bool | No | Enable transparent background | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `spec.colors` | map[string]string | No | Color overrides (hex values); required for standalone | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `spec.promptColors` | map[string]string | No | Starship prompt segment color overrides | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| `spec.options` | map[string]any | No | Theme-specific key-value options | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |

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
  version: ""                     # Optional - Version constraint (git tag)
  priority: 0                     # Optional - Load priority (higher = earlier)
  lazy: false                     # Optional - Lazy-load the plugin
  enabled: true                   # Optional - Disable with false; omit when enabled (default: true)
  event:                          # Optional - Events that trigger loading (string or list)
    - ""
  ft:                             # Optional - Filetypes that trigger loading (string or list)
    - ""
  keys:                           # Optional - Key mappings that trigger loading
    - key: ""                     # Key sequence (e.g., "<leader>ff")
      mode: ""                    # Vim mode (e.g., "n", "v", "i") — string or list
      action: ""                  # Action to perform
      desc: ""                    # Description shown in which-key
  cmd:                            # Optional - Commands that trigger loading (string or list)
    - ""
  dependencies:                   # Optional - Plugin dependencies (strings or objects)
    - ""                          # Simple: just a repo path (e.g., "nvim-lua/plenary.nvim")
    # - repo: ""                  # Detailed: object with repo, build, version, branch, config
    #   build: ""
    #   version: ""
    #   branch: ""
    #   config: false
  build: ""                       # Optional - Build command to run after install
  config: ""                      # Optional - Lua configuration code (runs after plugin loads)
  init: ""                        # Optional - Lua init code (runs before plugin loads)
  opts: {}                        # Optional - Plugin options passed to setup()
  keymaps:                        # Optional - Additional key mappings (not lazy-load triggers)
    - key: ""                     # Key sequence
      mode: ""                    # Vim mode — string or list
      action: ""                  # Action to perform
      desc: ""                    # Description
  health_checks:                  # Optional - Health checks to verify plugin installation
    - type: ""                    # lua_module, command, treesitter, or lsp
      value: ""                   # Module name, command name, parser, or LSP server
      description: ""             # Human-readable description of what is checked
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Unique plugin name | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `metadata.description` | string | No | Human-readable description | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `metadata.category` | string | No | Category for organization | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `metadata.tags` | []string | No | Tags for searching and filtering | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.repo` | string | Yes | GitHub repository path | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.branch` | string | No | Git branch | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.version` | string | No | Version constraint (git tag) | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.priority` | int | No | Load priority (higher = earlier) | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.lazy` | bool | No | Lazy-load the plugin | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.enabled` | bool | No | Disable with `false`; omit when enabled (default: `true`) | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.event` | string or []string | No | Events that trigger loading | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.ft` | string or []string | No | Filetypes that trigger loading | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.keys` | []object | No | Key mappings that trigger loading | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.cmd` | string or []string | No | Commands that trigger loading | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.dependencies` | []string or []object | No | Plugin dependencies | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.build` | string | No | Build command after install | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.config` | string | No | Lua configuration code (post-load) | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.init` | string | No | Lua init code (pre-load) | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.opts` | any | No | Options passed to setup() | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.keymaps` | []object | No | Additional key mappings (not lazy triggers) | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| `spec.health_checks` | []object | No | Health checks: type, value, description | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |

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
| `metadata.name` | string | Yes | Unique package name | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `metadata.description` | string | No | Human-readable description | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `metadata.category` | string | No | Category for organization | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `metadata.tags` | []string | No | Tags for searching and filtering | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `metadata.labels` | map[string]string | No | Key-value labels | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `spec.extends` | string | No | Parent package (single inheritance) | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `spec.plugins` | []string | No | Plugin names to include | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| `spec.enabled` | bool | No | Whether package is enabled (default: true) | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |

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
| `metadata.name` | string | Yes | Unique prompt name | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `metadata.description` | string | No | Human-readable description | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `metadata.category` | string | No | Category for organization | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `metadata.tags` | []string | No | Tags for filtering | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `metadata.labels` | map[string]string | No | Key-value labels | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.type` | string | Yes | `starship`, `powerlevel10k`, or `oh-my-posh` | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.addNewline` | bool | No | Add newline before prompt | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.palette` | string | No | Starship palette name | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.format` | string | No | Prompt format string | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.modules` | map[string]ModuleConfig | No | Per-module configuration | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.character` | object | No | Prompt character symbols | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.paletteRef` | string | No | Color palette reference | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.colors` | map[string]string | No | Custom color overrides | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.rawConfig` | string | No | Raw config for advanced use | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| `spec.enabled` | bool | No | Whether enabled (default: true) | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |

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
| `metadata.name` | string | Yes | Unique package name | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.description` | string | No | Human-readable description | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.category` | string | No | Category for organization | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.tags` | []string | No | Tags for filtering | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.labels` | map[string]string | No | Key-value labels | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.extends` | string | No | Parent package (single inheritance) | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.plugins` | []string | No | Shell plugin names | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.prompts` | []string | No | Prompt names to include | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.profiles` | []string | No | Profile preset names | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.wezterm` | object | No | WezTerm config: fontSize, colorScheme, fontFamily | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.promptStyle` | string | No | Modular prompt style name | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.promptExtensions` | []string | No | Prompt extension names | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.enabled` | bool | No | Whether enabled (default: true) | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |

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
| `metadata.name` | string | Yes | Unique emulator config name | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.description` | string | No | Human-readable description | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.category` | string | No | Category for organization | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.labels` | map[string]string | No | Key-value labels | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.type` | string | Yes | `wezterm`, `alacritty`, or `kitty` | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.config` | map[string]any | No | Emulator-specific configuration | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.themeRef` | string | No | Reference to a theme | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| `spec.workspace` | string | No | Associated workspace name | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |

---

## Meta Resources

### GlobalDefaults

System-wide fallback values for theme, packages, build args, CA certs, and registry routing. Singleton — exactly one per installation.  
Full reference: [global-defaults.md](global-defaults.md)

```yaml
# Full reference: https://devopsmaestro.io/reference/global-defaults
apiVersion: devopsmaestro.io/v1   # REQUIRED — always "devopsmaestro.io/v1"
kind: GlobalDefaults              # REQUIRED
metadata:
  name: global-defaults           # REQUIRED — always "global-defaults"; value is informational
spec:
  theme: ""                       # Optional - Global fallback theme name (lowest priority in cascade)
                                  #   e.g. coolnight-ocean, tokyonight-night, gruvbox-dark
  nvimPackage: ""                 # Optional - Global fallback NvimPackage name
  terminalPackage: ""             # Optional - Global fallback TerminalPackage name
  plugins:                        # Optional - Global default plugin names
    - ""
  buildArgs:                      # Optional - Global build args passed as --build-arg (lowest priority)
    KEY: "value"                  #   global < ecosystem < domain < app < workspace
  caCerts:                        # Optional - CA certs injected globally into all workspace builds
    - name: ""                    # REQUIRED per cert - alphanumeric/_/- only; max 64 chars
      vaultSecret: ""             # REQUIRED per cert - MaestroVault secret name (PEM)
      vaultEnvironment: ""        # Optional - Vault environment override
      vaultField: ""              # Optional - Field within secret (default: "cert")
  registryOci: ""                 # Optional - Default OCI registry resource name (type: zot)
  registryPypi: ""                # Optional - Default PyPI registry resource name (type: devpi)
  registryNpm: ""                 # Optional - Default npm registry resource name (type: verdaccio)
  registryGo: ""                  # Optional - Default Go module proxy resource name (type: athens)
  registryHttp: ""                # Optional - Default HTTP caching proxy resource name (type: squid)
  registryIdleTimeout: ""         # Optional - Global idle timeout for on-demand registries (e.g. "30m", "1h")
```

| Field | Type | Required | Description | Reference |
|-------|------|----------|-------------|-----------|
| `metadata.name` | string | Yes | Always `global-defaults`; informational only | [GlobalDefaults](global-defaults.md) |
| `spec.theme` | string | No | Global fallback theme; lowest priority in cascade | [GlobalDefaults](global-defaults.md) |
| `spec.nvimPackage` | string | No | Global fallback NvimPackage name | [GlobalDefaults](global-defaults.md) |
| `spec.terminalPackage` | string | No | Global fallback TerminalPackage name | [GlobalDefaults](global-defaults.md) |
| `spec.plugins` | []string | No | Global default plugin names | [GlobalDefaults](global-defaults.md) |
| `spec.buildArgs` | map[string]string | No | Global build args; lowest priority in cascade | [GlobalDefaults](global-defaults.md) |
| `spec.caCerts[].name` | string | Yes (per cert) | Cert name; must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars | [GlobalDefaults](global-defaults.md) |
| `spec.caCerts[].vaultSecret` | string | Yes (per cert) | MaestroVault secret name containing PEM | [GlobalDefaults](global-defaults.md) |
| `spec.caCerts[].vaultEnvironment` | string | No | Vault environment override | [GlobalDefaults](global-defaults.md) |
| `spec.caCerts[].vaultField` | string | No | Field within secret (default: `"cert"`) | [GlobalDefaults](global-defaults.md) |
| `spec.registryOci` | string | No | Default OCI registry resource name | [GlobalDefaults](global-defaults.md) |
| `spec.registryPypi` | string | No | Default PyPI registry resource name | [GlobalDefaults](global-defaults.md) |
| `spec.registryNpm` | string | No | Default npm registry resource name | [GlobalDefaults](global-defaults.md) |
| `spec.registryGo` | string | No | Default Go module proxy resource name | [GlobalDefaults](global-defaults.md) |
| `spec.registryHttp` | string | No | Default HTTP caching proxy resource name | [GlobalDefaults](global-defaults.md) |
| `spec.registryIdleTimeout` | string | No | Global idle timeout for `on-demand` registries (e.g., `"30m"`, `"1h"`) | [GlobalDefaults](global-defaults.md) |

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
  group: ""                       # Optional - API group (e.g., "mycompany.io")
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
| `spec.group` | string | No | API group for the custom resource (e.g., `mycompany.io`) | [CRD](custom-resource-definition.md) |
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

---

## Nvim Annotated Blank Templates

Complete blank templates for nvim resources with every possible field, inline comments, and required/optional markers. Copy and fill in only the fields you need — remove commented-out optional fields you won't use.

---

### NvimPlugin — Complete Annotated Template

```yaml
# ============================================================
# NvimPlugin — Full annotated blank template
# Apply with: dvm apply -f my-plugin.yaml
# ============================================================
apiVersion: devopsmaestro.io/v1    # REQUIRED — always "devopsmaestro.io/v1"
kind: NvimPlugin                   # REQUIRED — always "NvimPlugin"

metadata:
  # ── Identity ──────────────────────────────────────────────
  name: ""                         # REQUIRED — unique name (e.g., "telescope", "lspconfig")
  description: ""                  # optional — human-readable description
  category: ""                     # optional — e.g., "lsp", "navigation", "completion",
                                   #   "ui", "editing", "git", "syntax", "debugging"
  tags:                            # optional — list of strings for searching/filtering
    - ""
  labels:                          # optional — arbitrary key-value labels
    key: "value"
  annotations:                     # optional — non-identifying metadata (e.g., docs URLs)
    key: "value"

spec:
  # ── Source ────────────────────────────────────────────────
  repo: ""                         # REQUIRED — GitHub repo path, e.g., "nvim-telescope/telescope.nvim"
  branch: ""                       # optional — pin to a git branch (mutually exclusive with version)
  version: ""                      # optional — pin to a git tag/version (e.g., "0.1.4")

  # ── Load order ────────────────────────────────────────────
  priority: 0                      # optional — higher number loads earlier; useful for colorschemes
                                   #   (e.g., priority: 1000 loads before priority: 100)

  # ── Lazy loading ──────────────────────────────────────────
  lazy: false                      # optional — true = defer load until a trigger fires
  event:                           # optional — load on Neovim events; string or list
    - "BufReadPre"                 #   common: BufReadPre, BufNewFile, VeryLazy,
    - "BufNewFile"                 #           InsertEnter, CmdlineEnter
  ft:                              # optional — load only for these filetypes; string or list
    - "go"
    - "lua"
  cmd:                             # optional — load when these ex-commands are first called; string or list
    - "Telescope"
  keys:                            # optional — load on keypress AND register the mapping
    - key: "<leader>ff"            #   key: key sequence (e.g., "<leader>ff", "<C-p>")
      mode: "n"                    #   mode: vim mode — "n", "i", "v", "x", "o", "c", or list
      action: "<cmd>Telescope find_files<cr>"  # action: Lua code or ex-command
      desc: "Find files"           #   desc: shown in which-key popup

  # ── Dependencies ──────────────────────────────────────────
  dependencies:                    # optional — plugins that must be loaded first
    - "nvim-lua/plenary.nvim"      #   simple format: just the repo path
    - repo: "nvim-tree/nvim-web-devicons"  # detailed format: full spec
      build: ""                    #     build: build command for this dep
      version: ""                  #     version: pin to git tag
      branch: ""                   #     branch: pin to git branch
      config: false                #     config: true = run this dep's config too

  # ── Build ─────────────────────────────────────────────────
  build: ""                        # optional — shell/neovim command after install/update
                                   #   e.g., "make", ":TSUpdate", "npm install"

  # ── Configuration ─────────────────────────────────────────
  init: |                          # optional — Lua code that runs BEFORE the plugin loads
    -- Set globals/options that the plugin reads at startup
    vim.g.example_setting = true
  config: |                        # optional — Lua code that runs AFTER the plugin loads
    require("example").setup({
      -- your config here
    })
  opts: {}                         # optional — table passed directly to setup(); alternative to config
                                   #   when you only need to pass options, not run arbitrary Lua

  # ── Additional keymaps ────────────────────────────────────
  keymaps:                         # optional — mappings registered after plugin loads
    - key: "<leader>tt"            #   unlike spec.keys, these do NOT trigger lazy loading
      mode: "n"
      action: "<cmd>SomeCommand<cr>"
      desc: "Description"

  # ── State ─────────────────────────────────────────────────
  enabled: true                    # optional — set to false to disable; omit when enabled (default: true)

  # ── Health checks ─────────────────────────────────────────
  health_checks:                   # optional — verified with: nvp health
    - type: "lua_module"           #   type options:
      value: "example"             #     lua_module  — checks require("value") succeeds
      description: "Core module"   #     command     — checks ex-command exists
    - type: "command"              #     treesitter  — checks parser is installed
      value: "ExampleCmd"          #     lsp         — checks LSP server is configured
      description: "Main command"
```

---

### NvimTheme — Complete Annotated Template

```yaml
# ============================================================
# NvimTheme — Full annotated blank template
# Apply with: dvm apply -f my-theme.yaml
# Two modes:
#   Plugin-based: spec.plugin.repo points to a colorscheme plugin
#   Standalone:   omit spec.plugin entirely; spec.colors is REQUIRED
# ============================================================
apiVersion: devopsmaestro.io/v1    # REQUIRED — always "devopsmaestro.io/v1"
kind: NvimTheme                    # REQUIRED — always "NvimTheme"

metadata:
  # ── Identity ──────────────────────────────────────────────
  name: ""                         # REQUIRED — unique name, e.g., "tokyonight-night", "gruvbox-dark"
  description: ""                  # optional — human-readable description
  author: ""                       # optional — theme author
  category: ""                     # optional — "dark", "light", or "both"

spec:
  # ── Plugin source ─────────────────────────────────────────
  # For plugin-based themes: fill in spec.plugin.repo
  # For standalone themes:   remove the entire plugin block and add spec.colors
  plugin:                          # optional — omit entirely for standalone themes
    repo: ""                       # GitHub repository, e.g., "folke/tokyonight.nvim"
    branch: ""                     # optional — pin to git branch
    tag: ""                        # optional — pin to git tag/version

  # ── Variant ───────────────────────────────────────────────
  style: ""                        # optional — plugin-specific variant, e.g.:
                                   #   tokyonight: "night", "storm", "day", "moon"
                                   #   catppuccin:  "mocha", "macchiato", "frappe", "latte"
                                   #   gruvbox:     "dark", "light"
                                   #   kanagawa:    "wave", "dragon", "lotus"

  # ── Background ────────────────────────────────────────────
  transparent: false               # optional — enable transparent background for terminal integration

  # ── Color overrides ───────────────────────────────────────
  # Semantic color names understood by DevOpsMaestro's color system.
  # Plugin-based themes: override individual colors; standalone: all required.
  colors:                          # optional for plugin-based; REQUIRED for standalone themes
    # Background palette
    bg: ""                         # main background
    bg_dark: ""                    # darker background (splits, inactive windows)
    bg_highlight: ""               # highlighted background (current line)
    bg_search: ""                  # search highlight background
    bg_visual: ""                  # visual selection background
    bg_float: ""                   # floating window background
    bg_popup: ""                   # popup/completion menu background
    bg_sidebar: ""                 # sidebar background (NvimTree, etc.)
    bg_statusline: ""              # statusline background

    # Foreground palette
    fg: ""                         # main foreground
    fg_dark: ""                    # muted foreground
    fg_gutter: ""                  # line numbers, gutter signs
    fg_sidebar: ""                 # sidebar foreground

    # Semantic/diagnostic colors
    error: ""                      # error highlight (DiagnosticError)
    warning: ""                    # warning highlight (DiagnosticWarn)
    info: ""                       # info highlight (DiagnosticInfo)
    hint: ""                       # hint highlight (DiagnosticHint)

    # UI colors
    border: ""                     # window/popup borders
    comment: ""                    # comment text

  # ── Prompt color overrides ────────────────────────────────
  # Applied when this theme is used with a Starship-based terminal prompt.
  # Keys are Starship module names; values are hex colors.
  promptColors:                    # optional — Starship prompt segment colors
    directory: ""                  # directory/path segment
    git_branch: ""                 # git branch segment
    username: ""                   # username segment
    hostname: ""                   # hostname segment

  # ── Plugin-specific options ───────────────────────────────
  # Passed directly to the colorscheme plugin's setup() call.
  # Keys and valid values are entirely plugin-defined.
  options:                         # optional — plugin-specific key-value options
    # Examples (actual keys depend on the plugin):
    italic_comments: true
    bold_keywords: false
    transparent_background: false
    dim_inactive: false
    terminal_colors: true
```

---

### NvimPackage — Complete Annotated Template

```yaml
# ============================================================
# NvimPackage — Full annotated blank template
# Apply with: dvm apply -f my-package.yaml
#
# A package is a named, reusable list of plugin references.
# Use spec.extends for single inheritance (one parent only).
# Packages are resolved at workspace build time; circular
# dependencies are rejected.
# ============================================================
apiVersion: devopsmaestro.io/v1    # REQUIRED — always "devopsmaestro.io/v1"
kind: NvimPackage                  # REQUIRED — always "NvimPackage"

metadata:
  # ── Identity ──────────────────────────────────────────────
  name: ""                         # REQUIRED — unique name, e.g., "golang-dev", "core", "typescript-full"
  description: ""                  # optional — human-readable description
  category: ""                     # optional — e.g., "language", "framework", "core", "purpose"
  tags:                            # optional — list of strings for searching/filtering
    - ""
  labels:                          # optional — arbitrary key-value labels
    key: "value"
  annotations:                     # optional — non-identifying metadata
    key: "value"

spec:
  # ── Inheritance ───────────────────────────────────────────
  extends: ""                      # optional — parent package name (single inheritance only)
                                   #   the parent's plugins are prepended before this package's plugins
                                   #   inheritance chain: core → lang-dev → framework-dev

  # ── Plugins ───────────────────────────────────────────────
  plugins:                         # REQUIRED — list of plugin names to include (at least one)
    - ""                           #   use the plugin's metadata.name as defined in its NvimPlugin resource
                                   #   e.g., "telescope", "lspconfig", "nvim-cmp"
                                   #   strings only — no inline plugin definitions

  # ── State ─────────────────────────────────────────────────
  enabled: true                    # optional — set to false to disable; omit when enabled (default: true)
                                   #   disabled packages are stored but not applied to workspaces
```

---

## See Also

- [YAML Reference Overview](index.md) -- Resource type descriptions and hierarchy
- [YAML Schema](../configuration/yaml-schema.md) -- Schema validation rules
- [Commands Reference](../dvm/commands.md) -- CLI commands including `dvm apply`

---

## Complete Setup Template

A production-ready multi-document YAML file that bootstraps an entire DevOpsMaestro environment from scratch. Save this as `complete-setup.yaml` and customize for your team.

This template sets up the **Acme Platform** — a realistic microservices development environment with Go/Python backends, Neovim IDE configuration, and WezTerm terminal setup.

```yaml
# Complete DevOpsMaestro Setup — Acme Platform
# Apply with: dvm apply -f complete-setup.yaml
# Resources are processed in document order.

# ─── 1. Workspace ─────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: acme-platform
spec:
  image: acme/dev-workspace:latest
  build:
    dockerfile: .devcontainer/Dockerfile
    context: .
  shell: zsh
  terminal: wezterm
  nvim: true
  tools:
    - go
    - python3
    - node
    - docker
    - kubectl
  mounts:
    - source: ~/.ssh
      target: /home/dev/.ssh
      readOnly: true
    - source: ~/.aws
      target: /home/dev/.aws
      readOnly: true
  env:
    GOPATH: /home/dev/go
    EDITOR: nvim
---
# ─── 2. Global Defaults ───────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: GlobalDefaults
metadata:
  name: acme-defaults
spec:
  theme: catppuccin-mocha
  nvimPackage: acme-ide
  terminalPackage: acme-terminal
  buildArgs:
    GO_VERSION: "1.22"
    PYTHON_VERSION: "3.12"
  caCerts:
    - /etc/ssl/certs/acme-root-ca.pem
  registryOci: registry.acme.io
  registryGo: https://goproxy.acme.io
  registryNpm: https://npm.acme.io
  registryPypi: https://pypi.acme.io
  registryIdleTimeout: 30m
---
# ─── 3. Ecosystem ─────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: backend
spec:
  description: "Backend services ecosystem — Go and Python microservices"
  theme: catppuccin-mocha
  nvimPackage: acme-ide
  terminalPackage: acme-terminal
  build:
    parallel: true
    timeout: 10m
  caCerts:
    - /etc/ssl/certs/acme-root-ca.pem
  domains:
    - payments
    - identity
    - notifications
---
# ─── 4. Domain ─────────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: payments
spec:
  theme: catppuccin-mocha
  nvimPackage: acme-ide
  terminalPackage: acme-terminal
  build:
    parallel: true
    timeout: 5m
  apps:
    - payment-service
    - payment-gateway
    - billing-worker
---
# ─── 5. App ───────────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: payment-service
  ecosystem: backend
spec:
  path: ./services/payment-service
  theme: catppuccin-mocha
  nvimPackage: acme-ide
  terminalPackage: acme-terminal
  gitRepo: payment-service-repo
  language: go
  build:
    command: make build
    testCommand: make test
    lintCommand: golangci-lint run
  dependencies:
    - billing-worker
  services:
    - name: postgres
      image: postgres:16
      ports: ["5432:5432"]
      env:
        POSTGRES_DB: payments
        POSTGRES_USER: dev
        POSTGRES_PASSWORD: dev
    - name: redis
      image: redis:7-alpine
      ports: ["6379:6379"]
  ports:
    - "8080:8080"
    - "9090:9090"
  env:
    SERVICE_NAME: payment-service
    LOG_LEVEL: debug
    DB_HOST: localhost
    DB_PORT: "5432"
---
# ─── 6. Credential ────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: acme-dockerhub
spec:
  source: vault
  vaultSecret: secret/ci/dockerhub
  vaultEnvironment: production
  vaultFields:
    username: docker_user
    password: docker_token
  description: "Acme DockerHub service account for pulling base images"
---
# ─── 7. Registry ──────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: acme-registry
spec:
  type: oci
  version: "2"
  enabled: true
  port: 5000
  lifecycle:
    deleteUntagged: true
    keepLastN: 10
  storage:
    driver: filesystem
    rootDirectory: /var/lib/registry
  idleTimeout: 30m
  config:
    proxy:
      remoteurl: https://registry.acme.io
    auth:
      credential: acme-dockerhub
---
# ─── 8. Git Repo ──────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: payment-service-repo
spec:
  url: https://github.com/acme-corp/payment-service.git
  defaultRef: main
  authType: ssh
  credential: acme-github-ssh
  autoSync: true
  syncIntervalMinutes: 15
---
# ─── 9. Nvim Plugin ───────────────────────────────────────────
# Telescope — fuzzy finder for files, grep, and more
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
spec:
  repo: nvim-telescope/telescope.nvim
  branch: master
  priority: 100
  lazy: true
  enabled: true
  event:
    - VimEnter
  cmd:
    - Telescope
  dependencies:
    - nvim-lua/plenary.nvim
    - nvim-tree/nvim-web-devicons
  build: make
  keymaps:
    - key: "<leader>ff"
      action: "<cmd>Telescope find_files<cr>"
      desc: "Find files"
    - key: "<leader>fg"
      action: "<cmd>Telescope live_grep<cr>"
      desc: "Live grep"
    - key: "<leader>fb"
      action: "<cmd>Telescope buffers<cr>"
      desc: "Find buffers"
  health_checks:
    - command: "Telescope"
      expected: "telescope"
---
# ─── 10. Nvim Theme ───────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: catppuccin-mocha
spec:
  plugin:
    repo: catppuccin/nvim
    branch: main
    priority: 1000
  style: mocha
  transparent: false
  colors:
    background: "#1e1e2e"
    foreground: "#cdd6f4"
    cursor: "#f5e0dc"
    selection: "#585b70"
  promptColors:
    primary: "#89b4fa"
    secondary: "#a6e3a1"
    accent: "#f5c2e7"
  options:
    integrations:
      treesitter: true
      telescope: true
      cmp: true
      gitsigns: true
      nvimtree: true
---
# ─── 11. Nvim Package ─────────────────────────────────────────
# Combines plugins + theme into a distributable IDE configuration
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: acme-ide
spec:
  extends: base-ide
  plugins:
    - telescope
    - nvim-treesitter
    - nvim-lspconfig
    - nvim-cmp
    - gitsigns
    - lualine
    - neo-tree
    - which-key
  enabled: true
---
# ─── 12. Terminal Prompt ───────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: acme-starship
spec:
  type: starship
  addNewline: true
  palette: acme-colors
  format: "$directory$git_branch$git_status$golang$python$kubernetes$line_break$character"
  modules:
    directory:
      truncation_length: 3
      style: "bold cyan"
    git_branch:
      format: "[$symbol$branch]($style) "
      style: "bold purple"
    git_status:
      format: "[$all_status$ahead_behind]($style) "
    golang:
      format: "[$symbol($version)]($style) "
      symbol: " "
    python:
      format: "[$symbol($version)]($style) "
      symbol: " "
    kubernetes:
      disabled: false
      format: "[$symbol$context(/$namespace)]($style) "
  character:
    success_symbol: "[❯](bold green)"
    error_symbol: "[❯](bold red)"
  colors:
    primary: "#89b4fa"
    secondary: "#a6e3a1"
  enabled: true
---
# ─── 13. Terminal Plugin ──────────────────────────────────────
# Note: TerminalPlugin is a built-in kind but uses TerminalPackage
# to bundle plugins. Individual plugins are referenced by name.
---
# ─── 14. Terminal Package ──────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: acme-terminal
spec:
  extends: base-terminal
  plugins:
    - zoxide
    - fzf
    - bat
    - eza
    - ripgrep
    - fd
    - lazygit
    - delta
  prompts:
    - acme-starship
  wezterm: acme-wezterm
  promptStyle: starship
  enabled: true
---
# ─── 15. Terminal Emulator (WezTerm Config) ────────────────────
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  name: acme-wezterm
spec:
  type: wezterm
  config:
    font_size: 13.0
    font:
      family: "JetBrains Mono"
      harfbuzz_features:
        - calt
        - liga
    window_padding:
      left: 8
      right: 8
      top: 8
      bottom: 8
    enable_tab_bar: true
    hide_tab_bar_if_only_one_tab: true
    window_decorations: RESIZE
    window_background_opacity: 0.95
    scrollback_lines: 10000
  themeRef: catppuccin-mocha
  workspace: acme-platform
---
# ─── 16. Custom Resource Definition ───────────────────────────
# Extend DevOpsMaestro with custom resource types
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: monitors
spec:
  group: observability.acme.io
  names:
    kind: Monitor
    plural: monitors
    singular: monitor
    shortNames:
      - mon
  scope: Domain
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        properties:
          type:
            type: string
            enum: [datadog, prometheus, grafana]
          endpoint:
            type: string
          alerts:
            type: array
            items:
              type: object
              properties:
                name:
                  type: string
                query:
                  type: string
                threshold:
                  type: number
```

### Usage

1. **Save** the template to a file:
   ```bash
   # Copy and customize the template above
   vim complete-setup.yaml
   ```

2. **Customize** values for your environment:
   - Replace `acme-*` names with your organization
   - Update image references, registry URLs, and Git repo URLs
   - Adjust tool lists, plugins, and theme preferences
   - Set credential sources to match your secrets backend

3. **Apply** the complete setup:
   ```bash
   dvm apply -f complete-setup.yaml
   ```
   Resources are created in document order. Parent resources (Workspace, Ecosystem) are processed before children (Domain, App) that reference them.

4. **Verify** the setup:
   ```bash
   dvm get workspaces
   dvm get ecosystems
   dvm get apps
   dvm get nvim-packages
   dvm get terminal-packages
   ```

### Resource Reference

Each resource type in this template has a dedicated reference page with full field documentation:

| Resource | Reference |
|----------|-----------|
| Workspace | [workspace.md](workspace.md) |
| GlobalDefaults | [global-defaults.md](global-defaults.md) |
| Ecosystem | [ecosystem.md](ecosystem.md) |
| Domain | [domain.md](domain.md) |
| App | [app.md](app.md) |
| Credential | [credential.md](credential.md) |
| Registry | [registry.md](registry.md) |
| GitRepo | [gitrepo.md](gitrepo.md) |
| NvimPlugin | [NvimPlugin](https://rmkohlman.github.io/MaestroNvim/reference/nvim-plugin/) |
| NvimTheme | [NvimTheme](https://rmkohlman.github.io/MaestroNvim/reference/nvim-theme/) |
| NvimPackage | [NvimPackage](https://rmkohlman.github.io/MaestroNvim/reference/nvim-package/) |
| TerminalPrompt | [TerminalPrompt](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-prompt/) |
| TerminalPackage | [TerminalPackage](https://rmkohlman.github.io/MaestroTerminal/reference/terminal-package/) |
| TerminalEmulator | [WeztermConfig](https://rmkohlman.github.io/MaestroTerminal/reference/wezterm-config/) |
| CustomResourceDefinition | [custom-resource-definition.md](custom-resource-definition.md) |
