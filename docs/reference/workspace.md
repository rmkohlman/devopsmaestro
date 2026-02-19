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
    mergeMode: extend
    customConfig: |
      -- Custom Lua configuration
      vim.opt.relativenumber = true
      vim.opt.expandtab = true
      vim.opt.shiftwidth = 2
      vim.opt.tabstop = 2
      
      -- Go-specific settings
      vim.api.nvim_create_autocmd("FileType", {
        pattern = "go",
        callback = function()
          vim.opt_local.shiftwidth = 4
          vim.opt_local.tabstop = 4
        end,
      })
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
    resources:
      cpus: "2.0"
      memory: "4G"
      storage: "20G"
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `Workspace` |
| `metadata.name` | string | ✅ | Unique name for the workspace |
| `metadata.app` | string | ✅ | Parent app name |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.image` | object | ❌ | Container image configuration |
| `spec.build` | object | ❌ | Build configuration for dev tools |
| `spec.shell` | object | ❌ | Shell configuration |
| `spec.terminal` | object | ❌ | Terminal multiplexer configuration |
| `spec.nvim` | object | ❌ | Neovim configuration |
| `spec.mounts` | array | ❌ | Container mount points |
| `spec.sshKey` | object | ❌ | SSH key configuration |
| `spec.env` | object | ❌ | Workspace environment variables |
| `spec.container` | object | ❌ | Container settings |

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
    args:                           # Build arguments
      GITHUB_USERNAME: ${GITHUB_USERNAME}
      GITHUB_PAT: ${GITHUB_PAT}
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
    mergeMode: extend               # extend, replace, merge
    customConfig: |
      vim.opt.relativenumber = true
```

**Available structures:**
- `lazyvim` - LazyVim distribution
- `custom` - Custom configuration
- `nvchad` - NvChad distribution
- `astronvim` - AstroNvim distribution

**Merge modes:**
- `extend` - Add plugins to package
- `replace` - Replace package plugins
- `merge` - Intelligent merge

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
    resources:
      cpus: "2.0"                  # CPU allocation
      memory: "4G"                 # Memory allocation
      storage: "20G"               # Storage allocation
```

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
- [NvimTheme](nvim-theme.md) - Theme definitions
- [NvimPlugin](nvim-plugin.md) - Plugin configurations
- [NvimPackage](nvim-package.md) - Plugin packages

## Validation Rules

- `metadata.name` must be unique within the parent app
- `metadata.name` must be a valid DNS subdomain
- `metadata.app` must reference an existing App resource
- `spec.shell.type` must be `zsh` or `bash`
- `spec.terminal.type` must be `tmux`, `zellij`, or `screen`
- `spec.nvim.structure` must be valid Neovim distribution
- `spec.nvim.theme` must reference an existing theme
- `spec.mounts[].type` must be `bind`, `volume`, or `tmpfs`
- `spec.sshKey.mode` must be valid SSH key mode
- `spec.container.resources` must have valid resource limits