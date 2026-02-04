# DevOpsMaestro

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![CI](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml/badge.svg)](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**kubectl-style CLI toolkit for containerized development environments.**

DevOpsMaestro provides two tools:

| Tool | Binary | Description |
|------|--------|-------------|
| **DevOpsMaestro** | `dvm` | Workspace and project management with container-native dev environments |
| **NvimOps** | `nvp` | Standalone Neovim plugin & theme manager using YAML |

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

```bash
# Initialize
dvm init

# Create and use a project
dvm create project myapp
dvm use project myapp

# Create a workspace
dvm create workspace dev
dvm use workspace dev

# Build and attach
dvm build
dvm attach
```

---

## Features

### dvm - Workspace Management

- **kubectl-style commands** - Familiar `get`, `create`, `delete`, `apply` patterns
- **Multi-platform** - OrbStack, Docker Desktop, Podman, Colima
- **Container-native** - Isolated dev environments with Neovim pre-configured
- **Database-backed** - SQLite storage for projects, workspaces, plugins
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
| projects | `proj` | `dvm get proj` |
| workspaces | `ws` | `dvm get ws` |
| context | `ctx` | `dvm get ctx` |
| platforms | `plat` | `dvm get plat` |

### dvm Commands

```bash
# Status
dvm status                    # Show current context and runtime info
dvm status -o json            # JSON output

# Projects
dvm create project <name>     # Create project
dvm get projects              # List projects (or: dvm get proj)
dvm delete project <name>     # Delete project
dvm use project <name>        # Set active project
dvm use project --clear       # Clear active project

# Workspaces
dvm create workspace <name>   # Create workspace
dvm get workspaces            # List workspaces (or: dvm get ws)
dvm delete workspace <name>   # Delete workspace
dvm use workspace <name>      # Set active workspace
dvm use ws none               # Clear active workspace

# Context
dvm get context               # Show active project/workspace (or: dvm get ctx)

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
