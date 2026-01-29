# DevOpsMaestro (dvm)

**Kubernetes-style development environment orchestration with database-backed Neovim plugin management**

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![License](https://img.shields.io/badge/license-GPL--3.0%20%2B%20Commercial-blue)](https://github.com/rmkohlman/devopsmaestro/blob/main/LICENSING.md)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)
[![Tests](https://img.shields.io/badge/tests-66%20passing-brightgreen)](https://github.com/rmkohlman/devopsmaestro)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey)](https://github.com/rmkohlman/devopsmaestro/releases)

DevOpsMaestro is a professional CLI tool that brings Kubernetes-style declarative configuration to development environments. Work inside containerized environments with consistent, reproducible Neovim setups managed through a database-backed plugin system.

## 🆕 What's New in v0.3.1

- 🖥️ **Multi-platform support**: OrbStack, Docker Desktop, Podman, Colima
- 🔌 **Platform detection**: `dvm get platforms` shows available container runtimes
- 🏗️ **Decoupled architecture**: Swappable builders, runtimes, and database drivers
- 📝 **Structured logging**: `-v` for debug output, `--log-file` for JSON logs
- 🧪 **Test infrastructure**: 34 automated tests in manual test scripts
- 📚 **New documentation**: CLAUDE.md, STANDARDS.md, MANUAL_TEST_PLAN.md
- 🐛 **Bug fixes**: Build without nvim config, improved attach warnings, Podman compatibility

---

## ✨ Key Features

### 🚀 **kubectl-Style Workflow**
- Declarative YAML configurations for everything
- `dvm apply -f plugin.yaml` to manage Neovim plugins
- `dvm get`, `dvm list`, `dvm delete` commands you already know
- Multiple output formats: `table`, `yaml`, `json`
- Beautiful colored output with emoji icons

### 🎨 **Professional Theme System**
- **8 popular themes**: Catppuccin, Tokyo Night, Nord, Dracula, Gruvbox (light & dark)
- **Auto-detection**: Automatically adapts to your terminal's light/dark theme
- **Zero config**: Beautiful by default, customizable when you need it
- **Flexible**: Use environment variable (`DVM_THEME=catppuccin-mocha`) or config file
- **YAML syntax highlighting**: Keys in cyan, values in yellow

### 🔌 **Database-Backed Plugin System**
- Define Neovim plugins once in YAML, store in database
- Reference plugins by name across workspaces
- Share plugin configurations across teams via git
- 16+ pre-built plugins ready to use (Telescope, LSP, Treesitter, Copilot, etc.)

### 🐳 **Container-Native Development**
- Each workspace runs in an isolated Docker container
- Neovim pre-installed and configured automatically
- Your project files mounted at `/workspace`
- Consistent environment across machines and teams

### 📦 **Reproducible Environments**
- GitOps-friendly configuration management
- Backup and restore workspace states
- Version-controlled development setups
- One command to rebuild entire environment

---

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Core Concepts](#-core-concepts)
- [Plugin System](#-plugin-system)
- [Theme System](#-theme-system)
- [Usage](#-usage)
- [Commands](#-commands)
- [Configuration](#-configuration)
- [Architecture](#-architecture)
- [Examples](#-examples)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🚀 Quick Start

```bash
# 1. Install dvm
cd /path/to/devopsmaestro
make install-dev

# 2. Add to PATH (one-time setup)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 3. Initialize environment
dvm init

# 4. Create a project
dvm project create my-project
cd my-project

# 5. Apply plugin definitions
dvm plugin apply -f https://raw.githubusercontent.com/.../telescope.yaml

# 6. Create workspace with plugins
cat > workspace.yaml <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
spec:
  nvim:
    structure: custom
    plugins:
      - telescope
      - treesitter
      - lspconfig
      - copilot
EOF

dvm apply -f workspace.yaml

# 7. Build and enter your environment
dvm build
dvm attach

# Inside container: nvim with all plugins configured! 🎉
```

---

## 💾 Installation

### Download Pre-Built Binary (Recommended)

**macOS (Apple Silicon - M1/M2/M3):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.3.1/dvm-darwin-arm64 -o dvm
chmod +x dvm
sudo mv dvm /usr/local/bin/
dvm version
```

**macOS (Intel):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.3.1/dvm-darwin-amd64 -o dvm
chmod +x dvm
sudo mv dvm /usr/local/bin/
dvm version
```

**Linux (x86_64):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.3.1/dvm-linux-amd64 -o dvm
chmod +x dvm
sudo mv dvm /usr/local/bin/
dvm version
```

**Linux (ARM64):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.3.1/dvm-linux-arm64 -o dvm
chmod +x dvm
sudo mv dvm /usr/local/bin/
dvm version
```

### Build from Source

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro
git checkout v0.3.1
go build -o dvm
sudo mv dvm /usr/local/bin/
```

### Homebrew (Coming Soon)

```bash
brew install devopsmaestro/tap/dvm
```

**See [Releases](https://github.com/rmkohlman/devopsmaestro/releases) for all versions and checksums.**

### Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Docker** - [Get Docker](https://www.docker.com/get-started)
- **Colima** (macOS) - `brew install colima`
- **Git** - Version control

---

## 🧠 Core Concepts

### Projects
A **project** is the top-level container for your work. It contains workspaces, configuration, and manages the Docker context.

```bash
dvm project create my-app
cd my-app
```

### Workspaces
A **workspace** is an isolated development environment running in a Docker container. Each workspace has:
- Pre-configured Neovim with plugins
- Programming language tools (LSPs, formatters, linters)
- Your project files mounted at `/workspace`
- Isolated dependencies and services

```bash
dvm workspace create main --language python
```

### Plugins (Neovim)
**Plugins** are Neovim extensions stored in a database and referenced by name. Define once, use everywhere.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  category: fuzzy-finder
spec:
  repo: nvim-telescope/telescope.nvim
  branch: 0.1.x
  dependencies:
    - nvim-lua/plenary.nvim
  config: |
    require("telescope").setup({...})
```

---

## 🔌 Plugin System

### How It Works

```
┌─────────────────────────────────────┐
│  1. Define Plugin (YAML)            │
│     templates/nvim-plugins/*.yaml   │
└─────────────────────────────────────┘
         ↓ (dvm plugin apply -f)
┌─────────────────────────────────────┐
│  2. Store in Database               │
│     ~/.devopsmaestro/db.sqlite      │
└─────────────────────────────────────┘
         ↓ (dvm build)
┌─────────────────────────────────────┐
│  3. Generate Lua Files              │
│     .config/nvim/lua/.../*.lua      │
└─────────────────────────────────────┘
         ↓ (Container build)
┌─────────────────────────────────────┐
│  4. Neovim Loads Plugins            │
│     lazy.nvim → All configured!     │
└─────────────────────────────────────┘
```

### Managing Plugins

#### New kubectl-style Commands (v0.2.0+)
```bash
# List all available plugins (beautiful table)
dvm get plugins

# List plugins in YAML format
dvm get plugins -o yaml

# Get specific plugin details
dvm get plugin telescope

# Get plugin in YAML format (ready to edit/share)
dvm get plugin telescope -o yaml > telescope.yaml
```

#### Legacy Commands (still supported)
```bash
# List all available plugins
dvm plugin list -o table

# View plugin details
dvm plugin get telescope -o yaml

# Apply a plugin definition
dvm plugin apply -f telescope.yaml

# Apply from stdin
cat plugin.yaml | dvm plugin apply -f -

# Edit a plugin
dvm plugin edit telescope

# Delete a plugin
dvm plugin delete telescope
```

### Pre-Built Plugins (16+ Available)

| Plugin | Category | Description |
|--------|----------|-------------|
| **tokyonight** | colorscheme | Beautiful color scheme |
| **telescope** | fuzzy-finder | Fuzzy finder for files/grep |
| **treesitter** | syntax | Advanced syntax highlighting |
| **lspconfig** | lsp | LSP configurations |
| **nvim-cmp** | completion | Autocompletion engine |
| **mason** | lsp | LSP/formatter installer |
| **copilot** | ai | GitHub Copilot integration |
| **gitsigns** | git | Git integration |
| **lazygit** | git | Terminal UI for git |
| **which-key** | ui | Keybinding hints |
| **lualine** | ui | Status line |
| **autopairs** | editing | Auto-close brackets |
| **comment** | editing | Smart commenting |
| **surround** | editing | Surround text objects |
| **alpha** | ui | Dashboard/greeter |

**All plugins are stored in `templates/nvim-plugins/` and ready to apply!**

---

## 🎨 Theme System

### Beautiful Colors, Anywhere

DevOpsMaestro features a professional theme system that makes your CLI output look stunning on any terminal. Choose from 8 popular themes or let it auto-detect your terminal's theme!

### Available Themes

| Theme | Style | Description |
|-------|-------|-------------|
| **auto** ⭐ | adaptive | Auto-detects light/dark terminal (default) |
| **catppuccin-mocha** | dark | Soothing pastel purple-pink |
| **catppuccin-latte** | light | Warm pastel purple-pink |
| **tokyo-night** | dark | Vibrant blue-purple |
| **nord** | dark | Cool bluish minimal |
| **dracula** | dark | Classic purple-pink |
| **gruvbox-dark** | dark | Warm retro colors |
| **gruvbox-light** | light | Warm retro light |

### Using Themes

#### Option 1: Environment Variable (Quick Override)

```bash
# Try different themes
DVM_THEME=catppuccin-mocha dvm get plugins
DVM_THEME=tokyo-night dvm version
DVM_THEME=nord dvm get plugin telescope -o yaml

# Set for entire session
export DVM_THEME=dracula
dvm get plugins
```

#### Option 2: Config File (Permanent Setting)

Create or edit `~/.devopsmaestro/config.yaml`:

```yaml
# DevOpsMaestro Configuration File

# UI Theme
# Options: auto, catppuccin-mocha, catppuccin-latte, tokyo-night,
#          nord, dracula, gruvbox-dark, gruvbox-light
# Default: auto (automatically adapts to your terminal's theme)
theme: catppuccin-mocha
```

#### Option 3: Default (Zero Config)

```bash
# Just works! Auto-detects your terminal's light/dark theme
dvm get plugins
```

### Theme Features

**🎯 Auto-Detection (Default)**
- Automatically detects if your terminal has a light or dark background
- Adjusts colors for perfect contrast
- Works with ANY terminal theme (Catppuccin, Tokyo Night, Solarized, etc.)

**🌈 YAML Syntax Highlighting**
- Keys displayed in cyan + bold
- Values displayed in yellow
- Comments in gray
- Makes YAML output easy to read!

```bash
# Beautiful colored YAML output
dvm get plugin telescope -o yaml
```

**🎨 Consistent Experience**
- All output (tables, YAML, version info) uses your chosen theme
- Emoji icons + colored text = professional appearance
- Works across all commands

### Theme Priority

Themes are selected in this order:
1. **Environment variable** (`DVM_THEME=nord`) - Highest priority
2. **Config file** (`~/.devopsmaestro/config.yaml`) - Medium priority
3. **Auto-detection** (`auto` theme) - Default

### Examples

```bash
# Test all themes to find your favorite!
for theme in auto catppuccin-mocha tokyo-night nord dracula gruvbox-dark; do
    echo "Theme: $theme"
    DVM_THEME=$theme dvm version
    echo ""
done

# Use Catppuccin Mocha for everything
echo "theme: catppuccin-mocha" >> ~/.devopsmaestro/config.yaml

# Override config for one command
DVM_THEME=tokyo-night dvm get plugins

# Beautiful YAML with syntax highlighting
DVM_THEME=nord dvm get plugin telescope -o yaml
```

### Why Themes?

**For Consistency**: Match your terminal's existing color scheme  
**For Readability**: High-contrast colors for better visibility  
**For Customization**: Pick your favorite from popular themes  
**For Professionalism**: Same approach as kubectl, gh, bat, delta

---

## 📖 Usage

### Basic Workflow

```bash
# 1. Initialize (first time only)
dvm init

# 2. Create project
dvm project create my-app
cd my-app

# 3. Create workspace
dvm workspace create main --language python

# 4. Apply plugins (optional)
dvm plugin apply -f ~/devopsmaestro/templates/nvim-plugins/telescope.yaml
dvm plugin apply -f ~/devopsmaestro/templates/nvim-plugins/lspconfig.yaml

# 5. Configure workspace plugins
cat > workspace.yaml <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
spec:
  nvim:
    structure: custom
    plugins:
      - telescope
      - lspconfig
      - copilot
EOF

dvm apply -f workspace.yaml

# 6. Build container
dvm build

# 7. Enter workspace
dvm attach

# Inside container:
# - Neovim is configured with all plugins
# - Your project is at /workspace
# - All tools installed and ready
```

### Working with Workspaces

```bash
# List workspaces
dvm workspace list

# Get workspace details
dvm get workspace main -o yaml

# Update workspace configuration
dvm apply -f workspace.yaml

# Rebuild workspace
dvm build --force

# Attach to workspace
dvm attach

# Stop workspace
dvm stop
```

---

## 🎯 Commands

### Project Management

```bash
dvm project create <name>              # Create new project
dvm project list                       # List all projects
dvm project delete <name>              # Delete project
dvm get project <name> -o yaml         # Get project details
```

### Workspace Management

```bash
dvm workspace create <name>            # Create workspace
dvm workspace list                     # List workspaces
dvm workspace delete <name>            # Delete workspace
dvm get workspace <name> -o yaml       # Get workspace config
dvm apply -f workspace.yaml            # Apply configuration
```

### Plugin Management (kubectl-style)

#### New kubectl-style Commands (v0.2.0+)
```bash
# List plugins (kubectl-style)
dvm get plugins                        # List all plugins (table)
dvm get plugins -o yaml                # List all plugins (YAML)
dvm get plugins -o json                # List all plugins (JSON)

# Get specific plugin (kubectl-style)
dvm get plugin <name>                  # Get plugin details (table)
dvm get plugin <name> -o yaml          # Get plugin (YAML)
dvm get plugin <name> -o json          # Get plugin (JSON)
```

#### Legacy Commands (still supported)
```bash
dvm plugin apply -f plugin.yaml        # Create/update plugin
dvm plugin apply -f -                  # Apply from stdin
dvm plugin get <name> -o yaml          # Get plugin details
dvm plugin list -o table               # List all plugins
dvm plugin edit <name>                 # Edit in $EDITOR
dvm plugin delete <name>               # Delete plugin
```

### Build & Runtime

```bash
dvm build                              # Build container image
dvm build --force                      # Force rebuild (no cache)
dvm attach                             # Attach to workspace
dvm stop                               # Stop workspace
dvm exec -- <command>                  # Run command in workspace
```

### Administration

```bash
dvm init                               # Initialize environment
dvm version                            # Show version info
dvm admin migrate                      # Run database migrations
dvm admin backup                       # Backup database
```

### Output Formats

All `get` and `list` commands support multiple output formats:

```bash
dvm plugin list -o table               # Table (default)
dvm plugin list -o yaml                # YAML
dvm plugin list -o json                # JSON
dvm get workspace main -o yaml         # YAML output
```

---

## ⚙️ Configuration

### Workspace Configuration

Create `workspace.yaml`:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  project: my-app
spec:
  # Language and version
  language: python
  version: "3.11"
  
  # Neovim configuration
  nvim:
    structure: custom
    plugins:
      - tokyonight
      - telescope
      - treesitter
      - lspconfig
      - nvim-cmp
      - mason
      - copilot
      - gitsigns
      - lazygit
      - which-key
      - lualine
  
  # Container settings
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
  
  # Build customization
  build:
    devStage:
      packages:
        - git
        - curl
        - zsh
        - neovim
      customCommands:
        - pip install pylsp black isort
```

### Plugin Configuration

Create plugin YAML files in `templates/nvim-plugins/`:

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: Fuzzy finder for files, grep, buffers
  category: fuzzy-finder
  tags:
    - finder
    - search
spec:
  repo: nvim-telescope/telescope.nvim
  branch: 0.1.x
  
  dependencies:
    - nvim-lua/plenary.nvim
    - repo: nvim-telescope/telescope-fzf-native.nvim
      build: make
    - nvim-tree/nvim-web-devicons
  
  config: |
    local telescope = require("telescope")
    telescope.setup({
      defaults = {
        path_display = { "smart" },
      },
    })
    telescope.load_extension("fzf")
  
  keymaps:
    - key: <leader>ff
      mode: "n"
      action: <cmd>Telescope find_files<cr>
      desc: Fuzzy find files in cwd
    - key: <leader>fs
      mode: "n"
      action: <cmd>Telescope live_grep<cr>
      desc: Find string in cwd
```

### Environment Variables

```bash
# UI Theme (see Theme System section)
export DVM_THEME=catppuccin-mocha  # Override theme

# Database location
export DVM_DATABASE_PATH=~/.devopsmaestro/devopsmaestro.db

# Active project/workspace
export DVM_PROJECT=my-app
export DVM_WORKSPACE=main

# Container runtime
export COLIMA_PROFILE=local-lite
```

### Application Configuration

Create `~/.devopsmaestro/config.yaml` for persistent settings:

```yaml
# DevOpsMaestro Configuration File

# UI Theme (see Theme System section for all options)
# Default: auto (adapts to terminal light/dark theme)
theme: auto
```

---

## 🏗️ Architecture

### System Components

```
┌─────────────────────────────────────────────────────┐
│  CLI (cmd/)                                         │
│  ├─ Project commands                                │
│  ├─ Workspace commands                              │
│  ├─ Plugin commands (kubectl-style)                 │
│  └─ Build & runtime commands                        │
└─────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────┐
│  Database Layer (db/)                               │
│  ├─ SQLite backend                                  │
│  ├─ Projects, Workspaces, Plugins                   │
│  └─ CRUD operations                                 │
└─────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────┐
│  Template Engine (templates/)                       │
│  ├─ Plugin Lua generator                            │
│  ├─ Dockerfile generator                            │
│  └─ Neovim config generator                         │
└─────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────┐
│  Container Runtime (builders/)                      │
│  ├─ Docker/BuildKit integration                     │
│  ├─ Colima profile management                       │
│  └─ Image building & caching                        │
└─────────────────────────────────────────────────────┘
```

### Plugin System Architecture

```
YAML Definitions              Database                Container
─────────────────            ─────────               ─────────
telescope.yaml ─────┐
lspconfig.yaml ─────┤─apply─→ nvim_plugins ─build─→ .config/nvim/
copilot.yaml   ─────┤          └─ telescope           └─ telescope.lua
mason.yaml     ─────┘             └─ lspconfig           └─ lspconfig.lua
                                  └─ copilot              └─ copilot.lua
                                  └─ mason                └─ mason.lua
                                  
                    workspace.yaml
                         └─ plugins: [telescope, lspconfig, ...]
```

### File Structure

```
devopsmaestro/
├── cmd/                    # CLI commands
│   ├── root.go
│   ├── project.go
│   ├── workspace.go
│   ├── plugin.go          # kubectl-style plugin commands
│   ├── build.go
│   ├── attach.go
│   └── version.go
├── db/                     # Database layer
│   ├── database.go
│   ├── sqldatastore.go
│   └── migrations/
├── models/                 # Data models
│   ├── project.go
│   ├── workspace.go
│   └── nvim_plugin.go     # Plugin model with YAML conversion
├── templates/              # Template generators
│   ├── db_plugin_manager.go
│   ├── dockerfile_generator.go
│   └── nvim-plugins/      # Pre-built plugin YAMLs
│       ├── telescope.yaml
│       ├── lspconfig.yaml
│       └── ... (16+ plugins)
├── builders/               # Container builders
│   ├── builder.go
│   └── colima.go
├── Makefile               # Build & install
├── README.md              # This file!
├── INSTALL.md             # Installation guide
└── HOMEBREW.md            # Homebrew distribution guide
```

---

## 💡 Examples

### Example 1: Python Data Science Workspace

```yaml
# workspace.yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: data-science
spec:
  language: python
  version: "3.11"
  nvim:
    structure: custom
    plugins:
      - tokyonight
      - telescope
      - treesitter
      - lspconfig
      - nvim-cmp
      - copilot
  build:
    devStage:
      customCommands:
        - pip install jupyterlab pandas numpy matplotlib scikit-learn
        - pip install pylsp python-lsp-server[all]
```

```bash
dvm apply -f workspace.yaml
dvm build
dvm attach
# Inside: nvim, Python LSP, Copilot, Jupyter all configured!
```

### Example 2: Go Microservice Workspace

```yaml
# workspace.yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: api-service
spec:
  language: golang
  version: "1.21"
  nvim:
    structure: custom
    plugins:
      - telescope
      - treesitter
      - lspconfig
      - mason
      - gitsigns
      - lazygit
  build:
    devStage:
      languageTools:
        - gopls
        - delve
        - golangci-lint
```

### Example 3: Custom Plugin Definition

```yaml
# my-custom-plugin.yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: oil
  description: File explorer that works like a buffer
  category: file-explorer
spec:
  repo: stevearc/oil.nvim
  dependencies:
    - nvim-tree/nvim-web-devicons
  config: |
    require("oil").setup({
      columns = {
        "icon",
        "permissions",
        "size",
        "mtime",
      },
    })
  keymaps:
    - key: <leader>e
      mode: "n"
      action: <cmd>Oil<cr>
      desc: Open file explorer
```

```bash
# Apply your custom plugin
dvm plugin apply -f my-custom-plugin.yaml

# Reference it in workspace
# workspace.yaml:
#   nvim:
#     plugins:
#       - oil

dvm apply -f workspace.yaml
dvm build
```

---

## 🛠️ Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/yourusername/devopsmaestro.git
cd devopsmaestro

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Install locally
make install-dev
```

### Makefile Targets

```bash
make help          # Show all targets
make build         # Build binary
make test          # Run tests
make install       # Install to /usr/local/bin (sudo)
make install-dev   # Install to ~/.local/bin (no sudo)
make uninstall     # Remove binary
make clean         # Clean build artifacts
make fmt           # Format code
make lint          # Run linters
make release       # Build for multiple platforms
make version       # Show version info
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./db/...
go test ./models/...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...
```

---

## 🤝 Contributing

Contributions are welcome! We'd love your help making DevOpsMaestro better.

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Run tests** (`make test`)
5. **Format code** (`make fmt`)
6. **Commit** (`git commit -m 'Add amazing feature'`)
7. **Push** (`git push origin feature/amazing-feature`)
8. **Open a Pull Request**

### Areas We Need Help

- 🔌 **More pre-built plugins** - Add your favorite Neovim plugins!
- 📚 **Documentation** - Improve guides and examples
- 🐛 **Bug fixes** - Found an issue? Fix it!
- ✨ **Features** - Workspace templates, language support, etc.
- 🧪 **Testing** - Add more test coverage

### Guidelines

- Follow Go best practices
- Add tests for new features
- Update documentation
- Keep commits focused and descriptive
- Be respectful and collaborative

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## 📝 License

DevOpsMaestro uses a **dual-license model**:

### 🆓 Free for Personal Use (GPL-3.0)

- ✅ Personal projects
- ✅ Individual developers
- ✅ Open source projects (GPL-compatible)
- ✅ Learning and education
- ✅ Non-profit organizations

### 💼 Commercial License Required

For corporate or business use, a commercial license is required:

- 🏢 Corporations, LLCs, business entities
- 💰 Commercial development work
- 👥 Employees using DevOpsMaestro at work
- 🔒 Proprietary/closed-source software

**📖 See [LICENSING.md](LICENSING.md) for detailed license guide with FAQs and examples.**

**📜 License Files:**
- [LICENSE](LICENSE) - Full GPL-3.0 text (free tier)
- [LICENSE-COMMERCIAL.txt](LICENSE-COMMERCIAL.txt) - Commercial terms and pricing
- [LICENSING.md](LICENSING.md) - Human-friendly licensing guide

**💬 Contact for Commercial Licensing:**
- Email: support@devopsmaestro.io
- GitHub: [Open an issue](https://github.com/rmkohlman/devopsmaestro/issues)

---

## 👥 Authors

- **Robert Kohlman** - *Creator & Lead Developer* - [@rmkohlman](https://github.com/rmkohlman)

---

## 🙏 Acknowledgements

- **[kubectl](https://kubernetes.io/docs/reference/kubectl/)** - Inspiration for the command structure
- **[lazy.nvim](https://github.com/folke/lazy.nvim)** - Plugin manager used in DevOpsMaestro
- **[Docker](https://www.docker.com/)** - Container runtime
- **[Colima](https://github.com/abiosoft/colima)** - macOS container runtime
- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[Viper](https://github.com/spf13/viper)** - Configuration management

---

## 📚 Additional Resources

- [Installation Guide](INSTALL.md) - Detailed installation instructions
- [Homebrew Guide](HOMEBREW.md) - Publishing to Homebrew
- [Plugin Development](docs/plugins.md) - Creating custom plugins (coming soon)
- [Workspace Templates](docs/templates.md) - Pre-built templates (coming soon)

---

## 🎯 Roadmap

### v0.2.0 - Enhanced UX ✅
- ✅ kubectl-style plugin commands (`dvm get plugins`)
- ✅ Beautiful colored output with emoji icons
- ✅ Professional theme system (8 themes)
- ✅ Auto-detection of terminal theme
- ✅ YAML syntax highlighting
- ✅ Multiple output formats (table, yaml, json)
- ✅ 54+ comprehensive tests

### v0.3.1 - Multi-Platform Support ✅ (Current)
- ✅ Multi-platform container runtime (OrbStack, Docker Desktop, Podman, Colima)
- ✅ Platform detection (`dvm get platforms`)
- ✅ Decoupled architecture (swappable builders, runtimes, database)
- ✅ Structured logging with slog (`-v`, `--log-file`)
- ✅ Neovim configuration management (`dvm nvim init/status`)
- ✅ Shell completions (bash/zsh/fish/powershell)
- ✅ 34 automated manual tests

### v0.4.0 - Developer Experience (Next)
- ⏳ Plugin-to-nvim integration (use stored plugins in containers)
- ⏳ Expanded logging across all commands
- ⏳ `dvm logs` command for viewing logs
- ⏳ Workspace templates
- ⏳ Better error messages

### v0.5.0 - Collaboration
- ⏳ Team workspaces
- ⏳ Plugin marketplace
- ⏳ Cloud sync (optional)
- ⏳ VS Code integration

### v1.0.0 - Production Ready
- ⏳ Homebrew official release
- ⏳ Full test coverage (90%+)
- ⏳ Performance optimizations
- ⏳ Enterprise features
- ⏳ Windows support (WSL2)

---

## ❓ FAQ

**Q: How is this different from devcontainers?**  
A: DevOpsMaestro uses a database-backed plugin system with kubectl-style commands. Plugins are defined once and shared across workspaces. It's more focused on Neovim-centric development workflows.

**Q: Can I use this with VS Code?**  
A: Currently focused on Neovim, but VS Code integration is on the roadmap!

**Q: Does this work on Windows?**  
A: Not yet. macOS and Linux are currently supported. Windows via WSL2 is planned.

**Q: How do I share plugins with my team?**  
A: Plugin YAML files can be committed to git. Your team applies them with `dvm plugin apply -f plugin.yaml`.

**Q: Can I use my own Neovim config?**  
A: Yes! You can customize plugins or bring your entire config. See Configuration section.

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/devopsmaestro/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/devopsmaestro/discussions)
- **Email**: support@devopsmaestro.io

---

<div align="center">

**⭐ Star us on GitHub if DevOpsMaestro helps you! ⭐**

Made with ❤️ by developers, for developers.

</div>
