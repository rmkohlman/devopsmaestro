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

### 1. NvimTheme (NEW in v0.12.0)

Represents a Neovim theme that can be applied and shared via Infrastructure as Code.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-custom-theme
  description: "My beautiful custom theme"
  author: "john-doe"
  category: "dark"  # dark, light, monochrome
spec:
  # Plugin that provides the colorscheme
  plugin:
    repo: "user/awesome-theme.nvim"          # GitHub repository
    branch: "main"                           # Optional: git branch
    tag: "v1.2.0"                           # Optional: git tag/version
  
  # Theme style/variant (plugin-specific)
  style: "custom"                            # e.g., "storm", "night", "day"
  
  # Transparency support
  transparent: false                         # Enable transparent background
  
  # Custom color overrides
  colors:
    bg: "#1a1b26"                           # Background color
    fg: "#c0caf5"                           # Foreground color  
    accent: "#7aa2f7"                        # Accent color
    error: "#f7768e"                         # Error color
    warning: "#e0af68"                       # Warning color
    info: "#7dcfff"                          # Info color
    hint: "#1abc9c"                          # Hint color
    
  # Plugin-specific options
  options:
    italic_comments: true
    bold_keywords: false
    underline_errors: true
    custom_highlights:
      - group: "Comment"
        style: "italic"
        fg: "#565f89"
```

### 2. NvimPlugin (NEW in v0.12.0)

Represents a Neovim plugin configuration that can be shared and applied.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: "Fuzzy finder over lists"
  category: "navigation"
  tags: ["fuzzy-finder", "telescope", "navigation"]
spec:
  # Plugin repository
  repo: "nvim-telescope/telescope.nvim"
  branch: "master"                          # Optional
  version: "0.1.4"                         # Optional git tag
  
  # Lazy loading configuration
  lazy: true
  priority: 1000                            # Load priority (higher = earlier)
  
  # Lazy loading triggers
  event: ["VeryLazy"]                       # Load on events
  ft: ["lua", "vim"]                        # Load on filetypes
  cmd: ["Telescope", "Tele"]                # Load on commands
  keys:                                     # Load on key mappings
    - key: "<leader>ff"
      mode: "n"
      action: "<cmd>Telescope find_files<cr>"
      desc: "Find files"
    - key: "<leader>fg"
      mode: "n"  
      action: "<cmd>Telescope live_grep<cr>"
      desc: "Live grep"
  
  # Dependencies
  dependencies:
    - "nvim-lua/plenary.nvim"               # Simple string
    - repo: "nvim-tree/nvim-web-devicons"   # Detailed dependency
      build: ""
      config: false
  
  # Build command (if needed)
  build: "make"
  
  # Configuration (Lua code)
  config: |
    require('telescope').setup({
      defaults = {
        file_ignore_patterns = {"node_modules"},
        layout_strategy = 'horizontal',
        layout_config = {
          width = 0.95,
          height = 0.85,
        },
      }
    })
  
  # Init code (runs before plugin loads)
  init: |
    vim.g.telescope_theme = 'dropdown'
  
  # Options passed to plugin setup
  opts:
    defaults:
      prompt_prefix: "🔍 "
      selection_caret: "👉 "
      multi_icon: "📌"
    extensions:
      fzf:
        fuzzy: true
        override_generic_sorter: true
        override_file_sorter: true
  
  # Additional keymaps
  keymaps:
    - key: "<leader>fb"
      mode: ["n", "v"]
      action: "<cmd>Telescope buffers<cr>"
      desc: "Find buffers"
    - key: "<leader>fh"
      mode: "n"
      action: "<cmd>Telescope help_tags<cr>"
      desc: "Find help"
```

### 3. App

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
  
  # Repository information (NEW in v0.12.0)
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
    
  # Theme configuration (NEW in v0.12.0) - inherits to workspaces
  theme: coolnight-synthwave      # Optional: theme name
  
  # Associated workspaces
  workspaces:
    - main
    - debug
```

### 4. Workspace

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
      
  # Neovim configuration (ENHANCED in v0.12.0)
  nvim:
    structure: lazyvim            # lazyvim, custom, nvchad, astronvim, none
    theme: coolnight-synthwave    # Override app/domain/ecosystem theme
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

### Export resources to YAML

```bash
# Export theme
dvm get nvim theme coolnight-synthwave -o yaml > my-theme.yaml

# Export plugin
dvm get nvim plugin telescope -o yaml > telescope-plugin.yaml

# Export workspace
dvm get workspace main -o yaml > workspace.yaml

# Export app  
dvm get app my-api -o yaml > my-api.yaml
```

### Import resources from YAML

```bash
# Apply theme
dvm apply -f my-theme.yaml

# Apply plugin
dvm apply -f telescope-plugin.yaml

# Apply workspace
dvm apply -f workspace.yaml

# Apply from remote sources
dvm apply -f https://themes.example.com/coolnight-variants.yaml
dvm apply -f github:user/nvim-configs/plugins/telescope.yaml
```

### Create comprehensive configurations

```bash
# Multi-document YAML file
cat > full-config.yaml <<EOF
---
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-custom-theme
spec:
  plugin:
    repo: "user/my-theme.nvim"
  colors:
    bg: "#1a1b26"
    fg: "#c0caf5"
---
apiVersion: devopsmaestro.io/v1  
kind: NvimPlugin
metadata:
  name: my-telescope
spec:
  repo: "nvim-telescope/telescope.nvim"
  config: |
    require('telescope').setup({})
---
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev-environment
spec:
  nvim:
    theme: my-custom-theme
    plugins: ["my-telescope"]
EOF

dvm apply -f full-config.yaml
```

### Theme IaC Examples

```bash
# Apply CoolNight theme variant
dvm apply -f https://library.devopsmaestro.io/themes/coolnight-synthwave.yaml

# Apply custom team theme
dvm apply -f github:myteam/themes/company-theme.yaml

# Set theme at different hierarchy levels
dvm set theme company-theme --ecosystem production
dvm set theme coolnight-ocean --domain backend  
dvm set theme gruvbox-dark --app my-api
dvm set theme "" --workspace dev  # Clear, inherit from app
```

---

## Validation

DevOpsMaestro validates YAML against the schema on import:

**General validation:**
- Required fields (`apiVersion`, `kind`, `metadata.name`)
- Valid resource kinds (`NvimTheme`, `NvimPlugin`, `App`, `Workspace`)
- API version compatibility

**NvimTheme validation:**
- Valid GitHub repository format for `spec.plugin.repo`
- Color values in hex format (`#rrggbb`)
- Category enum values (`dark`, `light`, `monochrome`)

**NvimPlugin validation:**
- Valid repository format
- Lua syntax validation for `config` and `init` fields
- Keymap structure validation
- Dependency format validation

**Workspace/App validation:**
- Path existence checks for local resources
- Valid enum values (shell.type, nvim.structure, terminal.type)
- Image name format validation
- Theme name existence (against library + user themes)

**Example validation errors:**

```bash
$ dvm apply -f invalid-theme.yaml
Error: validation failed for NvimTheme/my-theme:
  - spec.plugin.repo: must be a valid GitHub repository (format: owner/repo)
  - spec.colors.bg: must be a valid hex color (format: #rrggbb)
  - metadata.category: must be one of: dark, light, monochrome
```

---

## Theme Cascade System (v0.12.0)

DevOpsMaestro now supports hierarchical theme inheritance:

```
Ecosystem → Domain → App → Workspace
   (org)    (context) (code)  (dev env)
```

### Theme Resolution Order

1. **Workspace theme** - Highest priority, overrides everything
2. **App theme** - Applies to all workspaces unless overridden  
3. **Domain theme** - Applies to all apps/workspaces unless overridden
4. **Ecosystem theme** - Global theme for organization
5. **System default** - `coolnight-ocean` (fallback)

### Theme Configuration Examples

```yaml
# Set ecosystem-wide theme
apiVersion: devopsmaestro.io/v1
kind: Ecosystem  
metadata:
  name: my-platform
spec:
  theme: catppuccin-mocha  # All apps inherit this

---
# Override for specific domain
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: my-platform
spec:
  theme: gruvbox-dark  # Backend apps use gruvbox

---
# Override for specific app
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: my-api
  domain: backend
spec:
  theme: tokyonight-night  # This app uses tokyonight

---
# Override for specific workspace
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  nvim:
    theme: coolnight-synthwave  # Dev workspace uses coolnight
```

### Built-in Theme Library (v0.12.0)

DevOpsMaestro includes 34+ themes that are instantly available:

**CoolNight Variants (21 themes):**
- `coolnight-ocean` (default)
- `coolnight-arctic`, `coolnight-midnight`
- `coolnight-synthwave`, `coolnight-violet`, `coolnight-grape`
- `coolnight-matrix`, `coolnight-forest`, `coolnight-mint`
- `coolnight-sunset`, `coolnight-ember`, `coolnight-gold`
- `coolnight-rose`, `coolnight-crimson`, `coolnight-sakura`
- `coolnight-charcoal`, `coolnight-slate`, `coolnight-warm`
- `coolnight-nord`, `coolnight-dracula`, `coolnight-solarized`

**Popular Themes:**
- `tokyonight-night`, `tokyonight-storm`, `tokyonight-day`
- `catppuccin-mocha`, `catppuccin-macchiato`, `catppuccin-frappe`
- `gruvbox-dark`, `gruvbox-light`
- `nord`, `dracula`, `onedark`

All library themes work immediately without installation:

```bash
dvm set theme coolnight-synthwave --workspace dev
dvm set theme tokyonight-storm --app my-api
dvm get nvim themes  # See all 34+ available themes
```

---

## Migration from Database

The `dvm get` command can export any resource to YAML format:

```bash
# Export themes
dvm get nvim theme coolnight-ocean -o yaml
dvm get nvim themes -o yaml  # All themes

# Export plugins  
dvm get nvim plugin telescope -o yaml
dvm get nvim plugins -o yaml  # All plugins

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
# - All custom themes and plugins
# - Theme assignments at all hierarchy levels
# - Complete configuration state
```

### Restore Example

```bash
# Restore from backup
dvm apply -f devopsmaestro-backup-20260219.yaml

# Selectively restore themes
grep -A 20 "kind: NvimTheme" backup.yaml > themes-only.yaml
dvm apply -f themes-only.yaml
```
