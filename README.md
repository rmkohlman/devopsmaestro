# DevOpsMaestro

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![CI](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml/badge.svg)](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**kubectl-style CLI toolkit for containerized development environments.**

DevOpsMaestro provides two tools:

| Tool | Binary | Description |
|------|--------|-------------|
| **DevOpsMaestro** | `dvm` | App and workspace management with container-native dev environments |
| **NvimOps** | `nvp` | Standalone Neovim plugin & theme manager using YAML |

### Object Hierarchy (v0.8.0+)

```
Ecosystem → Domain → App → Workspace
```

| Object | Purpose |
|--------|---------|
| **Ecosystem** | Top-level platform grouping |
| **Domain** | Bounded context (team area) |
| **App** | Your codebase (the thing you build) |
| **Workspace** | Dev environment for an App |

---

## Installation

### Homebrew (Recommended)

```bash
brew tap rmkohlman/tap

# Install DevOpsMaestro (includes dvm)
brew install devopsmaestro

# Or install NvimOps only (no containers needed)
brew install nvimops

# Verify installation
dvm version   # Should show v0.7.1
nvp version
```

### From Source

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro

# Build both binaries
go build -o dvm .
go build -o nvp ./cmd/nvp/

# Install
sudo mv dvm nvp /usr/local/bin/
```

### Requirements

- **Go 1.25+** for building from source
- **Docker** (for dvm) - OrbStack, Docker Desktop, Podman, or Colima

---

## Shell Completions

Both `dvm` and `nvp` support shell completions for commands, flags, and resource names. Completions provide:

- Tab completion for commands and subcommands
- Dynamic completion for resource names (ecosystems, domains, apps, workspaces)
- Flag value completion (e.g., `dvm attach -a <TAB>` shows available apps)
- Descriptions alongside resource names

### Installation Instructions

#### Bash (Linux)

```bash
dvm completion bash > /etc/bash_completion.d/dvm
nvp completion bash > /etc/bash_completion.d/nvp
```

#### Bash (macOS with Homebrew)

```bash
dvm completion bash > $(brew --prefix)/etc/bash_completion.d/dvm
nvp completion bash > $(brew --prefix)/etc/bash_completion.d/nvp
```

#### Zsh (macOS with Homebrew)

```bash
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm
nvp completion zsh > $(brew --prefix)/share/zsh/site-functions/_nvp
```

#### Zsh (Linux)

```bash
dvm completion zsh > "${fpath[1]}/_dvm"
nvp completion zsh > "${fpath[1]}/_nvp"
```

#### Fish

```bash
dvm completion fish > ~/.config/fish/completions/dvm.fish
nvp completion fish > ~/.config/fish/completions/nvp.fish
```

**Note:** After installing completions, restart your shell or source your shell configuration file.

---

## Quick Start

### NvimOps (nvp) - Neovim Plugin Manager

```bash
# Initialize
nvp init

# Browse and install plugins from library
nvp library list
nvp library install telescope treesitter lspconfig

# Install a theme
nvp theme library list
nvp theme library install tokyonight-custom --use

# Generate Lua files for Neovim
nvp generate

# Files created in ~/.config/nvim/lua/plugins/nvp/
```

### DevOpsMaestro (dvm) - Workspace Manager

#### Option 1: Add an Existing App

Already have a codebase on your laptop? Add it to dvm:

```bash
# Initialize dvm (one-time setup)
dvm init

# Set up the hierarchy (one-time or when starting new projects)
dvm create ecosystem my-platform    # Top-level grouping
dvm create domain backend           # Bounded context

# Go to your existing codebase
cd ~/Developer/my-existing-app

# Create an app from the current directory
dvm create app my-app --from-cwd

# Or specify the path explicitly
dvm create app my-app --path ~/Developer/my-existing-app

# Create a workspace (defines your container environment)
dvm create workspace dev
dvm use workspace dev

# Build the container image
dvm build

# Attach to the container (your code is mounted inside)
dvm attach
```

#### Option 2: Start a New App

Starting fresh? Create a new directory for your app:

```bash
# Initialize dvm (one-time setup)
dvm init

# Set up the hierarchy (one-time or when starting new projects)
dvm create ecosystem my-platform    # Top-level grouping
dvm create domain backend           # Bounded context

# Create a new directory for your app
mkdir ~/Developer/my-new-app
cd ~/Developer/my-new-app

# Initialize your code (e.g., git, go mod, npm init, etc.)
git init
go mod init github.com/myuser/my-new-app

# Create an app in dvm
dvm create app my-new-app --from-cwd

# Create workspace
dvm create workspace dev
dvm use workspace dev

# Build and attach
dvm build
dvm attach
```

#### Shorthand Version (using aliases)

```bash
cd ~/Developer/my-app
dvm init                                # One-time setup
dvm create eco my-platform              # Create ecosystem (one-time)
dvm create dom backend                  # Create domain (one-time)
dvm create app myapp --from-cwd         # Create app
dvm create ws dev                       # Create workspace
dvm use ws dev                          # Set active
dvm build && dvm attach                 # Build and enter container
```

#### Verify Your Setup

```bash
dvm get ctx          # Show current ecosystem/domain/app/workspace
dvm get apps         # List all apps
dvm get ws           # List workspaces in current app
dvm get plat         # Check detected container platforms
dvm status           # Full status overview
```

---

## Features

### dvm - Workspace Management

- **kubectl-style commands** - Familiar `get`, `create`, `delete`, `apply` patterns
- **Multi-platform** - OrbStack, Docker Desktop, Podman, Colima
- **Container-native** - Isolated dev environments with Neovim pre-configured
- **Database-backed** - SQLite storage for apps, workspaces, plugins
- **YAML configuration** - Declarative workspace definitions

### nvp - Neovim Plugin Manager

- **YAML-based plugins** - Define plugins in YAML, generate Lua
- **Built-in library** - 16+ curated plugins ready to install
- **Theme system** - 8 pre-built themes with palette export
- **URL support** - Install from GitHub repositories
- **Standalone** - Works without containers

---

## Command Reference

### Resource Aliases

kubectl-style short aliases for faster commands:

| Resource | Alias | Example |
|----------|-------|---------|
| ecosystems | `eco` | `dvm get eco` |
| domains | `dom` | `dvm get dom` |
| apps | `app` | `dvm get app` |
| workspaces | `ws` | `dvm get ws` |
| context | `ctx` | `dvm get ctx` |
| platforms | `plat` | `dvm get plat` |
| projects | `proj` | `dvm get proj` *(deprecated)* |

### dvm Commands

```bash
# Status
dvm status                    # Show current context and runtime info
dvm status -o json            # JSON output

# Ecosystems (v0.8.0+)
dvm create ecosystem <name>   # Create ecosystem
dvm get ecosystems            # List ecosystems
dvm use ecosystem <name>      # Set active ecosystem

# Domains (v0.8.0+)
dvm create domain <name>      # Create domain
dvm get domains               # List domains
dvm use domain <name>         # Set active domain

# Apps (v0.8.0+)
dvm create app <name>         # Create app
dvm get apps                  # List apps
dvm delete app <name>         # Delete app
dvm use app <name>            # Set active app

# Projects (DEPRECATED - use Apps)
dvm create project <name>     # Create project (deprecated)
dvm get projects              # List projects (deprecated)
dvm use project <name>        # Set active project (deprecated)

# Workspaces
dvm create workspace <name>   # Create workspace
dvm get workspaces            # List workspaces (or: dvm get ws)
dvm delete workspace <name>   # Delete workspace
dvm use workspace <name>      # Set active workspace

# Context
dvm get context               # Show active ecosystem/domain/app/workspace

# Build & Runtime
dvm build                     # Build workspace container
dvm attach                    # Attach to workspace
dvm detach                    # Stop workspace container

# Configuration
dvm apply -f workspace.yaml   # Apply YAML configuration
dvm get platforms             # List detected container platforms
```

### nvp Commands

```bash
# Plugins
nvp library list              # List available plugins
nvp library install <name>    # Install from library
nvp apply -f plugin.yaml      # Apply plugin from file
nvp apply -f https://example.com/plugin.yaml  # Apply from URL (auto-detected)
nvp apply -f github:user/repo/plugin.yaml     # GitHub shorthand
nvp apply -f -                # Apply from stdin

# Themes
nvp theme library list        # List available themes
nvp theme library install <name> --use
nvp theme use <name>          # Set active theme
nvp theme apply -f theme.yaml # Apply theme from file

# Generate
nvp generate                  # Generate Lua files
```

---

## Configuration

### Workspace YAML

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
spec:
  language: python
  version: "3.11"
  nvim:
    structure: custom
    plugins:
      - telescope
      - treesitter
      - lspconfig
```

### Plugin YAML

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  category: fuzzy-finder
spec:
  repo: nvim-telescope/telescope.nvim
  dependencies:
    - nvim-lua/plenary.nvim
  config: |
    require("telescope").setup({})
```

### Theme YAML

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-theme
spec:
  palette:
    bg: "#1a1b26"
    fg: "#c0caf5"
    # ...
```

---

## Source Types (kubectl-style)

The `-f` flag accepts multiple source types, auto-detected from the path:

| Source Type | Example | Description |
|-------------|---------|-------------|
| **File** | `-f plugin.yaml` | Local file path |
| **URL** | `-f https://example.com/plugin.yaml` | HTTP/HTTPS URL |
| **GitHub** | `-f github:user/repo/path.yaml` | GitHub shorthand |
| **Stdin** | `-f -` | Read from stdin |

```bash
# Apply from local file
dvm apply -f workspace.yaml
nvp apply -f plugin.yaml

# Apply from URL (auto-detected)
nvp apply -f https://raw.githubusercontent.com/user/repo/main/plugin.yaml

# Apply from GitHub (shorthand)
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml

# Apply from stdin
cat plugin.yaml | nvp apply -f -
echo '...' | dvm apply -f -
```

---

## Architecture

```
dvm/nvp CLI
    │
    ├── render/          # Decoupled output formatting
    ├── db/              # SQLite database layer (dvm)
    ├── pkg/source/      # Source resolution (file, URL, stdin, GitHub)
    ├── pkg/resource/    # Unified resource interface & handlers
    │   └── handlers/    # NvimPlugin, NvimTheme handlers
    ├── pkg/nvimops/     # Plugin/theme management (nvp)
    │   ├── plugin/      # Plugin types, parser, generator
    │   ├── theme/       # Theme types, parser, generator
    │   ├── store/       # Storage interfaces
    │   └── library/     # Embedded plugin/theme library
    ├── operators/       # Container runtime abstraction
    └── builders/        # Image building (Docker, BuildKit)
```

---

## Development

```bash
# Build
go build -o dvm .
go build -o nvp ./cmd/nvp/

# Test
go test ./...
go test ./... -race

# Lint (requires golangci-lint)
golangci-lint run
```

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Run tests (`go test ./... -race`)
4. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

[GPL-3.0](LICENSE) - Free for personal and open source use.

Commercial license available for business use. See [LICENSING.md](LICENSING.md).

---

## Links

- [Releases](https://github.com/rmkohlman/devopsmaestro/releases)
- [Changelog](CHANGELOG.md)
- [Homebrew Tap](https://github.com/rmkohlman/homebrew-tap)
- [Plugin Library](https://github.com/rmkohlman/nvim-yaml-plugins)
