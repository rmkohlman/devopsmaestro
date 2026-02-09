# DevOpsMaestro YAML Schema Documentation

## Overview

DevOpsMaestro uses YAML for configuration import/export, following Kubernetes-style patterns with `apiVersion`, `kind`, and `metadata`/`spec` structure.

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

Represents a top-level app (codebase).

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: my-api
  labels:
    team: backend
    language: golang
  annotations:
    description: "RESTful API for user management"
spec:
  path: /Users/rkohlman/apps/my-api
  workspaces:
    - main
    - debug
    - testing
```

### 2. Workspace

Represents a development container configuration.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  app: my-api
  labels:
    language: golang
    env: development
spec:
  # Image configuration
  image:
    name: dvm-main-my-api:latest
    buildFrom: ./Dockerfile  # Production Dockerfile
    baseImage: my-api:prod   # Optional: use pre-built prod image
    
  # Build configuration for dev stage
  build:
    args:
      GITHUB_USERNAME: ${GITHUB_USERNAME}
      GITHUB_PAT: ${GITHUB_PAT}
    devStage:
      packages:
        - git
        - zsh
        - neovim
        - curl
        - wget
        - build-essential
      languageTools:
        - gopls          # Go LSP
        - delve          # Go debugger
        - golangci-lint  # Linter
      customCommands:
        - curl -fsSL https://starship.rs/install.sh | sh -s -- -y
    
  # Shell configuration
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
    customRc: |
      # Custom zsh configuration
      export EDITOR=nvim
      alias k=kubectl
      
  # Neovim configuration
  nvim:
    structure: lazyvim  # or: custom, nvchad, astronvim
    plugins:
      # LSP
      - name: neovim/nvim-lspconfig
        config:
          servers: [gopls, pyright, tsserver]
      
      # Treesitter
      - name: nvim-treesitter/nvim-treesitter
        config:
          ensure_installed: [go, python, lua, bash, yaml, json]
          highlight: true
          indent: true
      
      # Completion
      - name: hrsh7th/nvim-cmp
      - name: hrsh7th/cmp-nvim-lsp
      - name: hrsh7th/cmp-buffer
      - name: hrsh7th/cmp-path
      
      # Telescope
      - name: nvim-telescope/telescope.nvim
        dependencies:
          - nvim-lua/plenary.nvim
      
      # Git
      - name: lewis6991/gitsigns.nvim
      
      # Language-specific
      - name: fatih/vim-go
        ft: [go]
        config:
          go_fmt_command: "goimports"
          go_def_mode: "gopls"
    
    customConfig: |
      -- Custom Lua configuration
      vim.opt.relativenumber = true
      vim.opt.expandtab = true
      vim.opt.shiftwidth = 2
  
  # Language-specific configuration
  languages:
    - name: go
      version: "1.22"
      lsp: gopls
      linter: golangci-lint
      formatter: gofmt
      debugger: delve
      envVars:
        GOPATH: /go
        GO111MODULE: "on"
    
  # Container mounts
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
    mode: mount_host  # or: global_dvm, per_app, generate
    path: ${HOME}/.ssh  # used when mode=mount_host
  
  # Environment variables
  env:
    EDITOR: nvim
    TERM: xterm-256color
    COLORTERM: truecolor
  
  # Container settings
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
    command: ["/bin/zsh", "-l"]
    entrypoint: []
    ports:
      - "8080:8080"
      - "9229:9229"  # Debug port
    resources:
      cpus: "2.0"
      memory: "2G"
```

### 3. Template (Future)

Reusable configuration templates.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Template
metadata:
  name: golang-api
  labels:
    language: golang
    type: api
spec:
  # References base workspace configuration
  workspaceDefaults:
    shell:
      type: zsh
      framework: oh-my-zsh
      theme: starship
    nvim:
      structure: lazyvim
      plugins:
        - name: fatih/vim-go
    languages:
      - name: go
        version: "1.22"
        lsp: gopls
```

---

## Language Presets

### Python (Base)

```yaml
spec:
  build:
    devStage:
      packages: [python3, python3-pip, python3-venv]
      languageTools: [pylsp, black, isort, pytest]
  languages:
    - name: python
      version: "3.11"
      lsp: pylsp
      formatter: black
      envVars:
        PYTHONPATH: /workspace
```

### Python (FastAPI)

```yaml
spec:
  build:
    devStage:
      packages: [python3, python3-pip, python3-venv]
      languageTools: [pylsp, black, isort, pytest, uvicorn]
  languages:
    - name: python
      version: "3.11"
      lsp: pylsp
      framework: fastapi
      envVars:
        PYTHONPATH: /workspace
  container:
    ports:
      - "8000:8000"
```

### Golang (Base)

```yaml
spec:
  build:
    devStage:
      languageTools: [gopls, delve, golangci-lint]
  languages:
    - name: go
      version: "1.22"
      lsp: gopls
      formatter: gofmt
      debugger: delve
      envVars:
        GOPATH: /go
        GO111MODULE: "on"
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
spec:
  path: /Users/rkohlman/apps/my-new-api
EOF

dvm apply -f app.yaml
```

### Backup entire configuration

```bash
dvm admin backup --output backup.yaml
```

This creates a multi-document YAML file with all apps, workspaces, and templates.

---

## Validation

DevOpsMaestro validates YAML against the schema on import:

- Required fields (`apiVersion`, `kind`, `metadata.name`)
- Valid enum values (shell.type, nvim.structure)
- Path existence checks
- Image name format
- Port range validation

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
