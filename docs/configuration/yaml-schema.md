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
      
  # Neovim configuration
  nvim:
    structure: lazyvim            # lazyvim, custom, nvchad, astronvim, none
    plugins:
      - neovim/nvim-lspconfig
      - nvim-treesitter/nvim-treesitter
      - hrsh7th/nvim-cmp
      - nvim-telescope/telescope.nvim
      - lewis6991/gitsigns.nvim
    customConfig: |
      -- Custom Lua configuration
      vim.opt.relativenumber = true
      vim.opt.expandtab = true
      vim.opt.shiftwidth = 2
  
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
    mode: mount_host              # mount_host, global_dvm, per_app, generate
    path: ${HOME}/.ssh
  
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
```

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
| System packages | - | `spec.build.devStage.packages` |
| Shell config | - | `spec.shell` |
| Terminal/tmux | - | `spec.terminal` |
| Neovim config | - | `spec.nvim` |
| SSH keys | - | `spec.sshKey` |
| Dev user (UID/GID) | - | `spec.container` |
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

### Export workspace to YAML

```bash
dvm get workspace main -o yaml > workspace.yaml
```

### Import workspace from YAML

```bash
dvm apply -f workspace.yaml
```

### Create app from YAML

```bash
cat > app.yaml <<EOF
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: my-new-api
  domain: backend
spec:
  path: /Users/dev/projects/my-new-api
  language:
    name: go
    version: "1.22"
EOF

dvm apply -f app.yaml
```

### Backup entire configuration

```bash
dvm admin backup --output backup.yaml
```

This creates a multi-document YAML file with all ecosystems, domains, apps, and workspaces.

---

## Validation

DevOpsMaestro validates YAML against the schema on import:

- Required fields (`apiVersion`, `kind`, `metadata.name`)
- Valid enum values (shell.type, nvim.structure, terminal.type)
- Path existence checks
- Image name format

---

## Migration from Database

The `dvm get` command can export any resource to YAML format:

```bash
# Export single workspace
dvm get workspace main -o yaml

# Export all workspaces in an app
dvm get workspaces -o yaml

# Export entire app with all workspaces
dvm get app my-api --include-workspaces -o yaml

# Export everything (multi-document YAML)
dvm admin backup -o yaml
```
