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

## See Also

- [YAML Reference Overview](index.md) -- Resource type descriptions and hierarchy
- [YAML Schema](../configuration/yaml-schema.md) -- Schema validation rules
- [Commands Reference](../dvm/commands.md) -- CLI commands including `dvm apply`

---

## Complete Setup Template

A production-ready multi-document YAML file that bootstraps an entire DevOpsMaestro environment from scratch. Save this as `complete-setup.yaml` and customize for your team.

This template sets up the **Acme Platform** — a realistic microservices development environment with a Go backend and workspace configuration.

```yaml
# Complete DevOpsMaestro Setup — Acme Platform
# Apply with: dvm apply -f complete-setup.yaml
# Resources are processed in document order.

# ─── 1. Global Defaults ───────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: GlobalDefaults
metadata:
  name: global-defaults
spec:
  theme: catppuccin-mocha
  nvimPackage: acme-ide
  terminalPackage: acme-terminal
  buildArgs:
    PIP_INDEX_URL: "https://pypi.acme.io/root/prod"
  caCerts:
    - name: acme-root-ca
      vaultSecret: acme-root-ca-pem
  registryOci: acme-zot
  registryGo: acme-athens
  registryIdleTimeout: 30m
---
# ─── 2. Ecosystem ─────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: acme-platform
  labels:
    organization: acme-corp
  annotations:
    description: "Acme Corp development platform"
spec:
  description: "Acme Corp development platform"
  theme: catppuccin-mocha
  nvimPackage: acme-ide
  terminalPackage: acme-terminal
  build:
    args:
      GOPROXY: "https://goproxy.acme.io,direct"
  domains:
    - backend
---
# ─── 3. Domain ─────────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: acme-platform
  labels:
    team: backend-team
spec:
  theme: catppuccin-mocha
  apps:
    - payment-service
---
# ─── 4. App ───────────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: payment-service
  domain: backend
  ecosystem: acme-platform
  labels:
    language: go
spec:
  path: /Users/dev/projects/payment-service
  language:
    name: go
    version: "1.22"
  build:
    buildpack: go
  dependencies:
    file: go.mod
    install: go mod download
  services:
    - name: postgres
      version: "16"
      port: 5432
      env:
        POSTGRES_DB: payments
        POSTGRES_USER: dev
        POSTGRES_PASSWORD: dev
    - name: redis
      version: "7"
      port: 6379
  ports:
    - "8080:8080"
    - "9090:9090"
  env:
    LOG_LEVEL: debug
  workspaces:
    - dev
---
# ─── 5. Workspace ─────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: payment-service
  domain: backend
  ecosystem: acme-platform
spec:
  build:
    args:
      GITHUB_PAT: ${GITHUB_PAT}
    devStage:
      packages:
        - git
        - curl
        - ripgrep
      devTools:
        - gopls
        - delve
        - golangci-lint
  shell:
    type: zsh
    framework: oh-my-zsh
    plugins:
      - git
      - golang
      - zsh-autosuggestions
  nvim:
    structure: lazyvim
    pluginPackage: acme-ide
    extraMasonTools:
      - gopls
  mounts:
    - type: bind
      source: ${APP_PATH}
      destination: /workspace
      readOnly: false
    - type: bind
      source: ${HOME}/.ssh
      destination: /home/dev/.ssh
      readOnly: true
  env:
    EDITOR: nvim
    GOPATH: /home/dev/go
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
---
# ─── 6. Credential ────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  ecosystem: acme-platform
spec:
  source: vault
  vaultSecret: "github-pat-shared"
  vaultEnvironment: production
  description: "GitHub PAT for pulling private Go modules"
---
# ─── 7. Registry ──────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: acme-zot
  description: "OCI container image registry for local development"
spec:
  type: zot
  version: "2.1.15"
  port: 5001
  lifecycle: on-demand
  idleTimeout: 1800
---
# ─── 8. Git Repo ──────────────────────────────────────────────
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: payment-service-repo
  labels:
    team: backend
    language: go
spec:
  url: "https://github.com/acme-corp/payment-service.git"
  defaultRef: main
  authType: ssh
  credential: github-token
  autoSync: true
  syncIntervalMinutes: 15
---
# ─── 9. Custom Resource Definition ───────────────────────────
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: monitors.observability.acme.io
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
        openAPIV3Schema:
          type: object
          properties:
            type:
              type: string
            endpoint:
              type: string
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
| CustomResourceDefinition | [custom-resource-definition.md](custom-resource-definition.md) |
