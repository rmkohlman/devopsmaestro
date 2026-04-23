# DevOpsMaestro YAML Schema Documentation

## Overview

DevOpsMaestro uses YAML for configuration import/export, following Kubernetes-style patterns with `apiVersion`, `kind`, and `metadata`/`spec` structure.

### App vs Workspace Separation

DevOpsMaestro separates concerns between two resource types:

| Resource | Purpose | Contains |
|----------|---------|----------|
| **App** | The codebase/application | Language, build config, services, ports, dependencies |
| **Workspace** | Developer environment | Editor (nvim), shell, terminal (tmux), dev tools, mounts |

**Rule of thumb:**
- **App** = "What the code needs to run" (would go in production)
- **Workspace** = "How the developer wants to work on it" (dev experience)

## Base Structure

All resources follow this pattern:

```yaml
apiVersion: devopsmaestro.io/v1
kind: <ResourceType>
metadata:
  name: <resource-name>
  labels: {}
  annotations: {}
spec:
  # Resource-specific configuration
```

---

## Resource Types

### 1. App

Represents a codebase/application within a domain. App configuration focuses on **what the code needs to build and run**.

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: my-api
  domain: backend  # Parent domain
  labels:
    team: platform
    type: api
  annotations:
    description: "RESTful API for user management"
spec:
  # Path to source code
  path: /Users/rkohlman/apps/my-api
  
  # Language configuration (what the app is written in)
  language:
    name: go          # go, python, node, rust, java, etc.
    version: "1.22"   # Language version
  
  # Repository information
  repo:
    url: "https://github.com/user/my-api"
    branch: "main"
    
  # Build configuration (how to build the app)
  build:
    dockerfile: ./Dockerfile      # Path to Dockerfile (if exists)
    buildpack: auto               # Or: go, python, node, etc.
    args:
      GITHUB_TOKEN: ${GITHUB_TOKEN}
    target: production            # Build target stage
    context: .                    # Build context path
  
  # Dependencies (where app deps come from)
  dependencies:
    file: go.mod                  # go.mod, requirements.txt, package.json
    install: go mod download      # Command to install deps
    extra:                        # Additional dependencies
      - github.com/some/package
  
  # Services the app needs (databases, caches, etc.)
  services:
    - name: postgres
      version: "15"
      port: 5432
      env:
        POSTGRES_USER: myapp
        POSTGRES_PASSWORD: secret
    - name: redis
      version: "7"
      port: 6379
  
  # Ports the app exposes
  ports:
    - "8080:8080"     # HTTP
    - "8443:8443"     # HTTPS
    - "9090:9090"     # Metrics
  
  # App environment variables
  env:
    DATABASE_URL: postgres://localhost:5432/myapp
    REDIS_URL: redis://localhost:6379
    LOG_LEVEL: debug
    
  # Theme configuration - inherits to workspaces
  theme: coolnight-synthwave      # Optional: theme name
  
  # Associated workspaces
  workspaces:
    - main
    - debug
```

### 2. Workspace

Represents a development environment for an app. Workspace configuration focuses on **developer experience**.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: my-api
  domain: backend               # Optional — parent domain name
  ecosystem: my-platform        # Optional — parent ecosystem name
  labels:
    purpose: development
  annotations:
    description: "Primary development workspace"
spec:
  # Image configuration
  image:
    name: dvm-main-my-api:latest
    buildFrom: ./Dockerfile       # Production Dockerfile to extend
    baseImage: my-api:prod        # Or use pre-built prod image
    
  # Dev build configuration (developer tools on top of app)
  build:
    args:
      GITHUB_USERNAME: ${GITHUB_USERNAME}
      GITHUB_PAT: ${GITHUB_PAT}
    caCerts:
      - name: corporate-ca          # REQUIRED — ^[a-zA-Z0-9][a-zA-Z0-9_-]*$; max 10
        vaultSecret: corp-ca-cert   # REQUIRED — MaestroVault secret name
        vaultEnvironment: prod      # Optional — vault environment override
        vaultField: cert            # Optional — field within secret (default: "cert")
    baseStage:
      packages:                   # System packages installed in the base (app) stage
        - ca-certificates
    devStage:
      packages:                   # System packages for dev
        - git
        - curl
        - wget
        - ripgrep
        - fd-find
      devTools:                   # Language dev tools (LSP, debugger, etc.)
        - gopls                   # Go LSP
        - delve                   # Go debugger
        - golangci-lint           # Linter
      customCommands:             # Custom setup commands
        - curl -fsSL https://starship.rs/install.sh | sh -s -- -y
    
  # Shell configuration
  shell:
    type: zsh                     # zsh, bash
    framework: oh-my-zsh          # oh-my-zsh, prezto
    theme: starship               # starship, powerlevel10k
    plugins:
      - git
      - zsh-autosuggestions
      - zsh-syntax-highlighting
      - docker
      - kubectl
    customRc: |
      # Custom zsh configuration
      export EDITOR=nvim
      alias k=kubectl
      
  # Terminal multiplexer configuration
  terminal:
    type: tmux                    # tmux, zellij, screen
    configPath: ~/.tmux.conf      # Mount this config
    autostart: true               # Start on attach
    prompt: starship              # Terminal prompt name (references a TerminalPrompt resource)
    plugins:                      # Terminal plugins to install
      - zsh-autosuggestions
    package: my-terminal-pkg      # Reference to a TerminalPackage resource by name
      
  # Neovim configuration
  nvim:
    structure: lazyvim            # lazyvim, custom, nvchad, astronvim, none
    theme: coolnight-synthwave    # Override app/domain/ecosystem theme
    pluginPackage: go-dev         # Reference to a NvimPackage resource by name
    plugins:                      # Individual NvimPlugin names to include
      - neovim/nvim-lspconfig
      - nvim-treesitter/nvim-treesitter
      - hrsh7th/nvim-cmp
      - nvim-telescope/telescope.nvim
      - lewis6991/gitsigns.nvim
    mergeMode: append             # How to merge package + plugins: append (default), replace
    customConfig: |
      -- Custom Lua configuration
      vim.opt.relativenumber = true
      vim.opt.expandtab = true
      vim.opt.shiftwidth = 2
    extraMasonTools:              # Additional Mason tools to install at build time
      - gopls
      - delve
    extraTreesitterParsers:       # Additional Treesitter parsers to install at build time
      - go
      - gomod
  
  # Optional workspace-level tools installed into the container image
  tools:
    opencode: true                # Install opencode TUI binary (default: false)
  
  # Container mounts (dev-specific)
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
  
  # SSH Key configuration
  sshKey:
    mode: mount_host              # mount_host, global_dvm, per_project, generate
    path: ${HOME}/.ssh
  
  # Optional: GitRepo resource to clone into the workspace
  gitrepo: my-app-repo            # Name of a GitRepo resource
  
  # Workspace environment variables (dev-specific)
  env:
    EDITOR: nvim
    TERM: xterm-256color
    COLORTERM: truecolor
  
  # Container settings (dev user setup)
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
    command: ["/bin/zsh", "-l"]
    entrypoint: []
    resources:
      cpus: "2.0"
      memory: "2G"
    sshAgentForwarding: false     # Forward SSH agent into container
    gitCredentialMounting: false  # Mount host git credentials
    networkMode: ""               # Optional: custom Docker network mode
```

---

## Type Definitions

### CACertConfig

Used in `spec.build.caCerts` on a Workspace. Describes a CA certificate to fetch from MaestroVault and inject into the container image at build time.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Certificate identifier. Must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`. Used as the certificate filename. |
| `vaultSecret` | string | Yes | MaestroVault secret name containing the PEM certificate. |
| `vaultEnvironment` | string | No | Vault environment override. If omitted, the default vault environment is used. |
| `vaultField` | string | No | Field within the vault secret to read. Defaults to `"cert"`. |

**Validation:**
- `name` must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- Maximum 10 certificates per workspace
- PEM content must contain both `BEGIN CERTIFICATE` and `END CERTIFICATE` markers
- A missing or invalid certificate causes a **fatal build error**

**Dockerfile effect:**
```dockerfile
COPY certs/ /usr/local/share/ca-certificates/custom/
RUN update-ca-certificates
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt
ENV NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
```

On Alpine-based images, `ca-certificates` is automatically added to the `apk add` package list when `caCerts` is configured.

---

## Key Differences: App vs Workspace

| Concern | App | Workspace |
|---------|-----|-----------|
| Language/Version | `spec.language` | - |
| Build config | `spec.build` | - |
| App dependencies | `spec.dependencies` | - |
| Services (DB, Redis) | `spec.services` | - |
| App ports | `spec.ports` | - |
| App env vars | `spec.env` | - |
| Dev tools (LSP, debugger) | - | `spec.build.devStage.devTools` |
| System packages (dev stage) | - | `spec.build.devStage.packages` |
| System packages (base stage) | - | `spec.build.baseStage.packages` |
| Workspace tools (opencode, etc.) | - | `spec.tools` |
| Shell config | - | `spec.shell` |
| Terminal/tmux | - | `spec.terminal` |
| Neovim config | - | `spec.nvim` |
| Neovim plugin package | - | `spec.nvim.pluginPackage` |
| SSH keys | - | `spec.sshKey` |
| Git repo to clone | - | `spec.gitrepo` |
| Dev user (UID/GID) | - | `spec.container` |
| SSH agent forwarding | - | `spec.container.sshAgentForwarding` |
| Dev mounts | - | `spec.mounts` |

---

## Language-Specific Examples

### Python FastAPI App + Workspace

**App (the codebase):**
```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: fastapi-service
  domain: backend
spec:
  path: /Users/dev/projects/fastapi-service
  language:
    name: python
    version: "3.11"
  build:
    dockerfile: ./Dockerfile
  dependencies:
    file: requirements.txt
    install: pip install -r requirements.txt
  services:
    - name: postgres
      version: "15"
      port: 5432
  ports:
    - "8000:8000"
  env:
    DATABASE_URL: postgres://localhost:5432/mydb
```

**Workspace (dev environment):**
```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: fastapi-service
spec:
  image:
    name: dvm-fastapi-service:latest
  build:
    devStage:
      packages: [python3-pip, python3-venv]
      devTools: [python-lsp-server, black, isort, pytest, mypy]
  shell:
    type: zsh
    framework: oh-my-zsh
  nvim:
    structure: lazyvim
  container:
    user: dev
    uid: 1000
    gid: 1000
```

### Golang API App + Workspace

**App (the codebase):**
```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: go-api
  domain: platform
spec:
  path: /Users/dev/projects/go-api
  language:
    name: go
    version: "1.22"
  build:
    buildpack: go
  dependencies:
    file: go.mod
    install: go mod download
  ports:
    - "8080:8080"
  env:
    GO111MODULE: "on"
```

**Workspace (dev environment):**
```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: go-api
spec:
  image:
    name: dvm-go-api:latest
  build:
    devStage:
      devTools: [gopls, delve, golangci-lint]
  shell:
    type: zsh
    theme: starship
  terminal:
    type: tmux
    autostart: true
  nvim:
    structure: lazyvim
```

### Node.js App + Workspace

**App (the codebase):**
```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: node-api
  domain: frontend
spec:
  path: /Users/dev/projects/node-api
  language:
    name: node
    version: "20"
  dependencies:
    file: package.json
    install: npm install
  ports:
    - "3000:3000"
    - "9229:9229"    # Debug port
```

**Workspace (dev environment):**
```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: node-api
spec:
  image:
    name: dvm-node-api:latest
  build:
    devStage:
      devTools: [typescript, ts-node, typescript-language-server, prettier, eslint]
  nvim:
    structure: lazyvim
```

---

## Variable Substitution

DevOpsMaestro supports environment variable substitution in YAML:

```yaml
spec:
  build:
    args:
      GITHUB_PAT: ${GITHUB_PAT}          # From environment
      APP_NAME: ${DVM_APP_NAME}          # From context
  mounts:
    - source: ${APP_PATH}                # Auto-filled from app
      destination: /workspace
```

### Built-in Variables

- `${APP_NAME}` - Current app name
- `${APP_PATH}` - Full path to app directory
- `${WORKSPACE_NAME}` - Current workspace name
- `${DVM_HOME}` - DevOpsMaestro home directory (~/.devopsmaestro)
- `${HOME}` - User home directory
- `${USER}` - Current username

---

## Usage Examples

### Export resources to YAML

```bash
# Export workspace
dvm get workspace main -o yaml > workspace.yaml

# Export app  
dvm get app my-api -o yaml > my-api.yaml
```

### Import resources from YAML

```bash
# Apply workspace
dvm apply -f workspace.yaml

# Apply app config
dvm apply -f my-api.yaml

# Apply from remote source
dvm apply -f https://configs.example.com/workspace.yaml
```

### Multi-document YAML

```bash
# Multi-document YAML file (App + Workspace together)
cat > full-config.yaml <<EOF
---
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: my-api
  domain: backend
spec:
  path: /Users/dev/projects/my-api
  language:
    name: go
    version: "1.22"
---
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: my-api
spec:
  image:
    name: dvm-my-api:latest
  build:
    devStage:
      devTools: [gopls, delve]
EOF

dvm apply -f full-config.yaml
```

---

## Validation

DevOpsMaestro validates YAML against the schema on import:

**General validation:**
- Required fields (`apiVersion`, `kind`, `metadata.name`)
- Valid resource kinds (`Ecosystem`, `Domain`, `App`, `Workspace`, `Credential`, `Registry`, `GitRepo`, `CustomResourceDefinition`)
- API version compatibility

**App validation:**
- Valid language names and versions
- Valid build configuration
- Port format validation

**Workspace validation:**
- Path existence checks for local resources
- Valid enum values (`shell.type`, `nvim.structure`, `terminal.type`)
- Image name format validation
- `caCerts` name format (`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), max 10 per workspace

**Example validation errors:**

```bash
$ dvm apply -f invalid-workspace.yaml
Error: validation failed for Workspace/main:
  - spec.shell.type: must be one of: zsh, bash
  - spec.build.caCerts[0].name: must match ^[a-zA-Z0-9][a-zA-Z0-9_-]*$
  - metadata.app: required field missing
```

---

## Export and Backup

The `dvm get` command can export any resource to YAML format:

```bash
# Export single workspace
dvm get workspace main -o yaml

# Export all workspaces in an app
dvm get workspaces --app my-api -o yaml

# Export app with metadata
dvm get app my-api -o yaml

# Export domain structure
dvm get domains --ecosystem my-platform -o yaml

# Export entire ecosystem
dvm get ecosystem my-platform -o yaml

# Backup everything (multi-document YAML)
dvm admin backup -o yaml
```

### Complete Backup Example

```bash
# Create comprehensive backup
dvm admin backup -o yaml > devopsmaestro-backup-$(date +%Y%m%d).yaml

# The backup includes:
# - All ecosystems, domains, apps, workspaces
# - All credentials and registries
# - Complete configuration state
```

### Restore Example

```bash
# Restore from backup
dvm apply -f devopsmaestro-backup-20260219.yaml
```
```
