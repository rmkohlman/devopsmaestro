# Workspace YAML Reference

**Kind:** `Workspace`  
**APIVersion:** `devopsmaestro.io/v1`

A Workspace represents a development environment for an app. Workspace configuration focuses on **developer experience** and tooling.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: api-service
  domain: backend
  system: payments
  ecosystem: my-platform
  labels:
    purpose: development
    user: john-doe
  annotations:
    description: "Primary development workspace"
    last-used: "2026-02-19T10:30:00Z"
spec:
  image:
    name: dvm-main-api-service:latest
    buildFrom: ./Dockerfile
    baseImage: api-service:prod
  build:
    args:
      GITHUB_USERNAME: ${GITHUB_USERNAME}
      GITHUB_PAT: ${GITHUB_PAT}
    caCerts:
      - name: corp-root-ca
        vaultSecret: corp-root-ca-pem
        vaultField: certificate
    baseStage:
      packages:
        - libpq-dev
    devStage:
      packages:
        - git
        - curl
        - wget
        - ripgrep
        - fd-find
        - jq
      devTools:
        - gopls
        - delve
        - golangci-lint
        - air
      customCommands:
        - curl -fsSL https://starship.rs/install.sh | sh -s -- -y
        - go install github.com/cosmtrek/air@latest
  shell:
    type: zsh
    framework: oh-my-zsh
    theme: starship
    plugins:
      - git
      - zsh-autosuggestions
      - zsh-syntax-highlighting
      - docker
      - kubectl
      - golang
    customRc: |
      # Custom zsh configuration
      export EDITOR=nvim
      export GO111MODULE=on
      alias k=kubectl
      alias g=git
      alias ll="ls -la"
  terminal:
    type: tmux
    configPath: ~/.tmux.conf
    autostart: true
  nvim:
    structure: lazyvim
    theme: coolnight-synthwave
    pluginPackage: golang-dev
    plugins:
      - neovim/nvim-lspconfig
      - nvim-treesitter/nvim-treesitter
      - hrsh7th/nvim-cmp
      - nvim-telescope/telescope.nvim
      - fatih/vim-go
    mergeMode: append
    extraMasonTools:
      - lua-language-server
      - stylua
    extraTreesitterParsers:
      - go
      - lua
    customConfig: |
      -- Custom Lua configuration
      vim.opt.relativenumber = true
      vim.opt.expandtab = true
      vim.opt.shiftwidth = 2
      vim.opt.tabstop = 2
  tools:
    opencode: true
  mounts:
    - type: bind
      source: ${APP_PATH}
      destination: /workspace
      readOnly: false
    - type: bind
      source: ${HOME}/.ssh
      destination: /home/dev/.ssh
      readOnly: true
    - type: bind
      source: ${HOME}/.gitconfig
      destination: /home/dev/.gitconfig
      readOnly: true
    - type: bind
      source: ${HOME}/.aws
      destination: /home/dev/.aws
      readOnly: true
  sshKey:
    mode: mount_host
    path: ${HOME}/.ssh
  env:
    EDITOR: nvim
    TERM: xterm-256color
    COLORTERM: truecolor
    LANG: en_US.UTF-8
    TZ: America/New_York
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
    command: ["/bin/zsh", "-l"]
    entrypoint: []
    sshAgentForwarding: true
    networkMode: bridge
    resources:
      cpus: "2.0"
      memory: "4G"
  gitrepo: api-service-repo
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `Workspace` |
| `metadata.name` | string | ✅ | Unique name for the workspace |
| `metadata.app` | string | ✅ | Parent app name |
| `metadata.domain` | string | ❌ | Parent domain name — enables context-free apply; required when domain name is ambiguous across ecosystems |
| `metadata.system` | string | ❌ | Parent system name — optional grouping layer between domain and app |
| `metadata.ecosystem` | string | ❌ | Parent ecosystem name — used with `metadata.domain` for fully-qualified context-free apply |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.image` | object | ❌ | Container image configuration |
| `spec.image.name` | string | ❌ | Image name (generated automatically if omitted) |
| `spec.image.buildFrom` | string | ❌ | Dockerfile path to build the image from |
| `spec.image.baseImage` | string | ❌ | Pre-built base image to use instead of building |
| `spec.build` | object | ❌ | Build configuration for dev tools |
| `spec.build.args` | map[string]string | ❌ | Build arguments emitted as `ARG` (not `ENV`) — not persisted in image layers |
| `spec.build.caCerts` | array | ❌ | CA certificates fetched from MaestroVault and injected at build time |
| `spec.build.caCerts[].name` | string | ✅ | Cert name — must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars |
| `spec.build.caCerts[].vaultSecret` | string | ✅ | MaestroVault secret name containing the PEM certificate |
| `spec.build.caCerts[].vaultEnvironment` | string | ❌ | Vault environment override |
| `spec.build.caCerts[].vaultField` | string | ❌ | Field within the secret (default: `cert`) |
| `spec.build.baseStage` | object | ❌ | Packages installed in the base (app) build stage |
| `spec.build.baseStage.packages` | array | ❌ | System packages installed via `apt-get` in the base stage |
| `spec.build.devStage` | object | ❌ | Developer tooling added on top of the base stage |
| `spec.build.devStage.packages` | array | ❌ | System packages installed in the dev stage (e.g., `ripgrep`, `fd-find`) |
| `spec.build.devStage.devTools` | array | ❌ | Language-specific dev tools (e.g., `gopls`, `delve`, `pylsp`) |
| `spec.build.devStage.customCommands` | array | ❌ | Arbitrary shell commands run during the dev stage build |
| `spec.shell` | object | ❌ | Shell configuration |
| `spec.shell.type` | string | ❌ | Shell type: `zsh`, `bash` |
| `spec.shell.framework` | string | ❌ | Shell framework: `oh-my-zsh`, `prezto` |
| `spec.shell.theme` | string | ❌ | Shell prompt theme: `starship`, `powerlevel10k`, `agnoster` |
| `spec.shell.plugins` | array | ❌ | Shell plugins to install |
| `spec.shell.customRc` | string | ❌ | Raw shell RC content appended to `.zshrc` / `.bashrc` |
| `spec.terminal` | object | ❌ | Terminal multiplexer configuration |
| `spec.terminal.type` | string | ❌ | Multiplexer type: `tmux`, `zellij`, `screen` |
| `spec.terminal.configPath` | string | ❌ | Host path to config file to mount into the container |
| `spec.terminal.autostart` | bool | ❌ | Start multiplexer automatically on container attach |
| `spec.terminal.prompt` | string | ❌ | TerminalPrompt resource name |
| `spec.terminal.plugins` | array | ❌ | Terminal plugin names to install |
| `spec.terminal.package` | string | ❌ | TerminalPackage resource name |
| `spec.nvim` | object | ❌ | Neovim configuration |
| `spec.nvim.structure` | string | ❌ | Nvim distribution: `lazyvim`, `custom`, `nvchad`, `astronvim` |
| `spec.nvim.theme` | string | ❌ | Theme name (overrides app/domain/ecosystem theme for nvim) |
| `spec.nvim.pluginPackage` | string | ❌ | NvimPackage resource name (pre-configured plugin collection) |
| `spec.nvim.plugins` | array | ❌ | Individual NvimPlugin names to include |
| `spec.nvim.mergeMode` | string | ❌ | How to merge package plugins with workspace plugins: `append` (default), `replace` |
| `spec.nvim.customConfig` | string | ❌ | Raw Lua configuration injected into the nvim setup |
| `spec.nvim.extraMasonTools` | array | ❌ | Additional Mason tools to install at image build time (e.g., `lua-language-server`) |
| `spec.nvim.extraTreesitterParsers` | array | ❌ | Additional Treesitter parsers to install at image build time (e.g., `go`, `python`) |
| `spec.tools` | object | ❌ | Optional workspace-level tool binaries installed at build time |
| `spec.tools.opencode` | bool | ❌ | Install [opencode](https://github.com/sst/opencode) AI assistant CLI (default: `false`) |
| `spec.mounts` | array | ❌ | Container mount points |
| `spec.mounts[].type` | string | ✅ | Mount type: `bind`, `volume`, `tmpfs` |
| `spec.mounts[].source` | string | ✅ | Host path or volume name |
| `spec.mounts[].destination` | string | ✅ | Container destination path |
| `spec.mounts[].readOnly` | bool | ❌ | Mount as read-only (default: `false`) |
| `spec.sshKey` | object | ❌ | SSH key configuration |
| `spec.sshKey.mode` | string | ✅ | Key mode: `mount_host`, `global_dvm`, `per_project`, `generate` |
| `spec.sshKey.path` | string | ❌ | Host path — used when `mode` is `mount_host` |
| `spec.env` | map[string]string | ❌ | Workspace environment variables injected into the container |
| `spec.container` | object | ❌ | Container runtime settings |
| `spec.container.user` | string | ❌ | Container username (sets `USER` in Dockerfile; default: `dev`) |
| `spec.container.uid` | int | ❌ | User ID (default: `1000`) |
| `spec.container.gid` | int | ❌ | Group ID (default: `1000`) |
| `spec.container.workingDir` | string | ❌ | Working directory inside the container (default: `/workspace`) |
| `spec.container.command` | array | ❌ | Container command (default: `["/bin/zsh", "-l"]`) |
| `spec.container.entrypoint` | array | ❌ | Container entrypoint override |
| `spec.container.resources` | object | ❌ | CPU and memory limits |
| `spec.container.resources.cpus` | string | ❌ | CPU limit (e.g., `"2.0"`) |
| `spec.container.resources.memory` | string | ❌ | Memory limit (e.g., `"4G"`) |
| `spec.container.sshAgentForwarding` | bool | ❌ | Forward SSH agent socket into the container (default: `false`) |
| `spec.container.networkMode` | string | ❌ | Docker network mode: `bridge` (default), `host`, `none` |
| `spec.gitrepo` | string | ❌ | GitRepo resource name to clone into the workspace on creation |

## Field Details

### metadata.name (required)
The unique identifier for the workspace within the app.

**Examples:**
- `main` - Primary development workspace
- `debug` - Debugging workspace
- `testing` - Testing workspace
- `hotfix` - Emergency hotfix workspace

### metadata.app (required)
The name of the parent app this workspace belongs to. Must reference an existing App resource.

```yaml
metadata:
  name: main
  app: api-service  # References App/api-service
```

### metadata.domain (optional)
The name of the parent domain. Optional but recommended — when present, `dvm apply` can resolve the workspace without requiring `dvm use domain` to be set. Required when the app's domain name exists in more than one ecosystem.

```yaml
metadata:
  name: main
  app: api-service
  domain: backend       # Enables context-free apply
  ecosystem: my-platform  # Add ecosystem to fully disambiguate
```

### metadata.system (optional)
The name of the parent system. Optional — when present, provides additional context for `dvm apply` resolution. Must reference an existing System resource when provided.

```yaml
metadata:
  name: main
  app: api-service
  domain: backend
  system: payments      # References System/payments
  ecosystem: my-platform
```

### metadata.ecosystem (optional)
The name of the parent ecosystem. Used together with `metadata.domain` for fully-qualified, context-free apply. Without this, `dvm apply` falls back to the active context.

```yaml
metadata:
  name: main
  app: api-service
  domain: backend
  ecosystem: my-platform
```

### spec.image (optional)
Container image configuration for the workspace.

```yaml
spec:
  image:
    name: dvm-main-api-service:latest  # Generated image name
    buildFrom: ./Dockerfile            # Dockerfile to extend
    baseImage: api-service:prod        # Or use pre-built image
```

### spec.build (optional)
Build configuration for adding development tools to the base image.

```yaml
spec:
  build:
    args:                           # Build arguments (emitted as ARG — not ENV)
      GITHUB_USERNAME: ${GITHUB_USERNAME}
      GITHUB_PAT: ${GITHUB_PAT}
    caCerts:                        # CA certificates injected from MaestroVault
      - name: corporate-ca          # REQUIRED — must match ^[a-zA-Z0-9][a-zA-Z0-9_-]*$
        vaultSecret: corp-ca-cert   # REQUIRED — MaestroVault secret name
        vaultEnvironment: prod      # Optional — vault environment override
        vaultField: cert            # Optional — field within secret (default: "cert")
    baseStage:
      packages:                     # System packages installed in the base (app) stage
        - libpq-dev
        - ca-certificates
    devStage:
      packages:                     # System packages for development
        - git
        - curl
        - wget
        - ripgrep
        - fd-find
      devTools:                     # Language-specific dev tools
        - gopls                     # Go LSP server
        - delve                     # Go debugger
        - golangci-lint            # Go linter
      customCommands:               # Custom setup commands
        - curl -fsSL https://starship.rs/install.sh | sh -s -- -y
```

**`spec.build.args`** — Build arguments are emitted as `ARG` declarations in the Dockerfile (not `ENV`). Values are available during the build but are not persisted in the final image. This is intentional: credentials such as `PIP_INDEX_URL` may contain tokens that must not be stored in image layers.

**`spec.build.baseStage.packages`** — System packages installed in the **base stage** of the generated Dockerfile, alongside any auto-detected language dependencies. Use this for packages your app runtime needs (e.g., `libpq-dev` for a PostgreSQL client). This is distinct from `devStage.packages`, which installs packages only in the developer layer.

**`spec.build.caCerts`** — CA certificates are fetched from MaestroVault at build time and injected into `/usr/local/share/ca-certificates/custom/`. The generator runs `update-ca-certificates` and sets `SSL_CERT_FILE`, `REQUESTS_CA_BUNDLE`, and `NODE_EXTRA_CA_CERTS` environment variables so Python, Node.js, and curl pick up the certificates automatically. On Alpine-based images, `ca-certificates` is automatically added to the `apk add` package list.

Validation rules for `caCerts`:
- `name` must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- Maximum 10 certificates per workspace
- PEM content must contain both `BEGIN CERTIFICATE` and `END CERTIFICATE` markers
- A missing or invalid certificate causes a **fatal build error** (not a warning)

### spec.shell (optional)
Shell configuration for the development environment.

```yaml
spec:
  shell:
    type: zsh                       # zsh, bash
    framework: oh-my-zsh           # oh-my-zsh, prezto
    theme: starship                # starship, powerlevel10k, agnoster
    plugins:
      - git
      - zsh-autosuggestions
      - zsh-syntax-highlighting
      - docker
      - kubectl
    customRc: |
      export EDITOR=nvim
      alias k=kubectl
```

### spec.terminal (optional)
Terminal multiplexer configuration.

```yaml
spec:
  terminal:
    type: tmux                      # tmux, zellij, screen
    configPath: ~/.tmux.conf        # Mount this config file
    autostart: true                 # Start on container attach
    prompt: my-starship-prompt      # TerminalPrompt resource name
    package: my-terminal-package    # Terminal package name
```

### spec.nvim (optional)
Neovim configuration for the workspace.

```yaml
spec:
  nvim:
    structure: lazyvim              # lazyvim, custom, nvchad, astronvim
    theme: coolnight-synthwave      # Override app/domain theme
    pluginPackage: golang-dev       # Pre-configured plugin package
    plugins:                        # Individual plugins
      - neovim/nvim-lspconfig
      - fatih/vim-go
    mergeMode: append               # append (default), replace
    customConfig: |
      vim.opt.relativenumber = true
    extraMasonTools:                # Additional Mason tools installed at build time
      - lua-language-server
      - stylua
    extraTreesitterParsers:         # Additional Treesitter parsers installed at build time
      - go
      - python
      - lua
```

**Available structures:**
- `lazyvim` - LazyVim distribution
- `custom` - Custom configuration
- `nvchad` - NvChad distribution
- `astronvim` - AstroNvim distribution

**Merge modes:**
- `append` - Add plugins to the package list (default)
- `replace` - Replace all package plugins with the workspace's plugin list

**`spec.nvim.extraMasonTools`** — Mason tool names installed into the container image at build time (not at container start). Use this for LSP servers, linters, and formatters that are not included in the chosen `pluginPackage`. Tools are installed via `MasonInstall` during the image build so they are immediately available when the container starts.

**`spec.nvim.extraTreesitterParsers`** — Treesitter parser names compiled and cached in the image at build time. Avoids the first-run compile delay inside the container. Use language short names as recognised by nvim-treesitter (e.g., `go`, `python`, `typescript`, `lua`).

### spec.tools (optional)

Optional workspace-level tool binaries installed into the container image at build time. Each tool is opt-in (default `false`) so images that do not need them stay lean.

```yaml
spec:
  tools:
    opencode: true    # Install opencode TUI binary (linux/amd64, linux/arm64)
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tools.opencode` | bool | `false` | Install [opencode](https://github.com/sst/opencode) AI assistant CLI |

When `tools.opencode` is `false` (or the `tools:` key is absent entirely), the section is omitted from YAML export. See [opencode CLI Tool](../dvm/opencode.md) for setup details.

### spec.mounts (optional)
Container mount points for accessing host filesystem.

```yaml
spec:
  mounts:
    - type: bind                    # bind, volume, tmpfs
      source: ${APP_PATH}          # Host path
      destination: /workspace      # Container path
      readOnly: false              # Mount as read-only
    - type: bind
      source: ${HOME}/.ssh
      destination: /home/dev/.ssh
      readOnly: true
```

### spec.sshKey (optional)
SSH key configuration for git operations.

```yaml
spec:
  sshKey:
    mode: mount_host                # mount_host, global_dvm, per_project, generate
    path: ${HOME}/.ssh             # Used when mode=mount_host
```

**SSH key modes:**
- `mount_host` - Mount host SSH keys
- `global_dvm` - Use global dvm SSH keys
- `per_project` - Use project-specific keys
- `generate` - Generate new SSH keys

### spec.env (optional)
Workspace-specific environment variables.

```yaml
spec:
  env:
    EDITOR: nvim
    TERM: xterm-256color
    COLORTERM: truecolor
    LANG: en_US.UTF-8
```

### spec.container (optional)
Container runtime settings.

```yaml
spec:
  container:
    user: dev                       # Container user name
    uid: 1000                       # User ID
    gid: 1000                       # Group ID
    workingDir: /workspace          # Working directory
    command: ["/bin/zsh", "-l"]     # Default command
    entrypoint: []                  # Container entrypoint
    sshAgentForwarding: true        # Forward SSH agent socket into the container
    networkMode: bridge             # Docker network mode: bridge, host, none
    resources:
      cpus: "2.0"                  # CPU allocation
      memory: "4G"                 # Memory allocation
```

**`container.user`** — Sets the `USER` directive in the generated Dockerfile. If unset, the default user `dev` is used.

**`container.sshAgentForwarding`** — When `true`, forwards the host SSH agent socket (`SSH_AUTH_SOCK`) into the running container. This lets git and other SSH-dependent tools inside the container use host SSH keys without copying them. Also stored in a dedicated DB column for fast querying.

**`container.networkMode`** — Sets the Docker `--network` flag when starting the container. Use `host` to share the host network stack (useful for services bound to `localhost`), `none` to disable networking entirely, or `bridge` (the default) for isolated networking.

## Language-Specific Examples

### Go Development Workspace

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: go-api
spec:
  build:
    devStage:
      devTools:
        - gopls
        - delve
        - golangci-lint
        - air
  shell:
    plugins:
      - golang
  nvim:
    structure: lazyvim
    pluginPackage: golang-dev
    plugins:
      - fatih/vim-go
      - ray-x/go.nvim
```

### Python Development Workspace

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: fastapi-service
spec:
  build:
    devStage:
      packages:
        - python3-pip
        - python3-venv
      devTools:
        - python-lsp-server
        - black
        - isort
        - pytest
        - mypy
  nvim:
    structure: lazyvim
    pluginPackage: python-dev
```

### Node.js Development Workspace

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: node-api
spec:
  build:
    devStage:
      devTools:
        - typescript
        - typescript-language-server
        - prettier
        - eslint
        - nodemon
  shell:
    plugins:
      - node
      - npm
  nvim:
    structure: lazyvim
    pluginPackage: typescript-dev
```

## Usage Examples

### Create Workspace

```bash
# From YAML file
dvm apply -f workspace.yaml

# Imperative command
dvm create workspace dev --app my-api
```

### Build and Attach

```bash
# Build workspace container
dvm build --workspace dev

# Attach to workspace
dvm attach --workspace dev
```

### Set Workspace Theme

```bash
# Set workspace-specific theme
dvm set theme coolnight-synthwave --workspace dev
```

### Export Workspace

```bash
# Export to YAML
dvm get workspace dev -o yaml

# Export with full configuration
dvm get workspace dev --include-config -o yaml
```

## Related Resources

- [App](app.md) - Parent application
- [Credential](credential.md) - Secrets scoped to this workspace
- [NvimTheme](nvim-theme.md) - Theme definitions
- [NvimPlugin](nvim-plugin.md) - Plugin configurations
- [NvimPackage](nvim-package.md) - Plugin packages
- [TerminalPrompt](terminal-prompt.md) - Shell prompt configurations
- [TerminalPackage](terminal-package.md) - Terminal configuration bundles

## Validation Rules

- `metadata.name` must be unique within the parent app
- `metadata.name` must be a valid DNS subdomain
- `metadata.app` must reference an existing App resource
- `metadata.domain`, if provided, must reference an existing Domain resource
- `metadata.system`, if provided, must reference an existing System resource
- `metadata.ecosystem`, if provided, must reference an existing Ecosystem resource
- `spec.shell.type` must be `zsh` or `bash`
- `spec.terminal.type` must be `tmux`, `zellij`, or `screen`
- `spec.nvim.structure` must be a valid Neovim distribution (`lazyvim`, `custom`, `nvchad`, `astronvim`)
- `spec.nvim.theme` must reference an existing theme
- `spec.build.caCerts[].name` must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars; max 10 certs
- `spec.mounts[].type` must be `bind`, `volume`, or `tmpfs`
- `spec.sshKey.mode` must be `mount_host`, `global_dvm`, `per_project`, or `generate`
- `spec.container.networkMode` must be `bridge`, `host`, or `none`
- `spec.container.resources.cpus` and `memory` must be valid Docker resource limit strings
- `spec.gitrepo`, if provided, must reference an existing GitRepo resource