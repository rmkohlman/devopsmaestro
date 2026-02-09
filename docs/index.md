# DevOpsMaestro

**kubectl-style CLI toolkit for containerized development environments.**

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![CI](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml/badge.svg)](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)

---

## What is DevOpsMaestro?

DevOpsMaestro provides two powerful CLI tools for modern development workflows:

| Tool | Binary | Description |
|------|--------|-------------|
| **DevOpsMaestro** | `dvm` | Workspace and app management with container-native dev environments |
| **NvimOps** | `nvp` | Standalone Neovim plugin & theme manager using YAML |

## Key Features

### dvm - Workspace Management

- :material-kubernetes: **kubectl-style commands** - Familiar `get`, `create`, `delete`, `apply` patterns
- :material-docker: **Multi-platform** - OrbStack, Docker Desktop, Podman, Colima
- :material-cube-outline: **Container-native** - Isolated dev environments with Neovim pre-configured
- :material-database: **Database-backed** - SQLite storage for apps, workspaces, plugins
- :material-file-document: **YAML configuration** - Declarative workspace definitions

### nvp - Neovim Plugin Manager

- :material-file-code: **YAML-based plugins** - Define plugins in YAML, generate Lua
- :material-library: **Built-in library** - 16+ curated plugins ready to install
- :material-palette: **Theme system** - 8 pre-built themes with palette export
- :material-link: **URL support** - Install from GitHub repositories
- :material-package-variant: **Standalone** - Works without containers

---

## Quick Install

=== "Homebrew (Recommended)"

    ```bash
    brew tap rmkohlman/tap
    
    # Install DevOpsMaestro (includes dvm)
    brew install devopsmaestro
    
    # Or install NvimOps only (no containers needed)
    brew install nvimops
    ```

=== "From Source"

    ```bash
    git clone https://github.com/rmkohlman/devopsmaestro.git
    cd devopsmaestro
    go build -o dvm .
    go build -o nvp ./cmd/nvp/
    sudo mv dvm nvp /usr/local/bin/
    ```

---

## Quick Example

### Add an existing app to dvm

```bash
cd ~/Developer/my-existing-app

dvm init                              # One-time setup
dvm create app my-app --from-cwd      # Create app from current dir
dvm use app my-app                    # Set as active
dvm create workspace dev              # Create a workspace
dvm use workspace dev                 # Set as active
dvm build                             # Build container
dvm attach                            # Enter the container
```

### Manage Neovim plugins with nvp

```bash
nvp init                              # Initialize
nvp library list                      # Browse available plugins
nvp library install telescope lspconfig treesitter
nvp theme library install tokyonight-custom --use
nvp generate                          # Generate Lua files
```

---

## Next Steps

<div class="grid cards" markdown>

-   :material-rocket-launch:{ .lg .middle } **Getting Started**

    ---

    Install DevOpsMaestro and set up your first app

    [:octicons-arrow-right-24: Installation](getting-started/installation.md)

-   :material-book-open-variant:{ .lg .middle } **Commands Reference**

    ---

    Complete reference for all dvm and nvp commands

    [:octicons-arrow-right-24: dvm Commands](dvm/commands.md)

-   :material-cog:{ .lg .middle } **Configuration**

    ---

    YAML schemas, plugin system, and customization

    [:octicons-arrow-right-24: Configuration](configuration/yaml-schema.md)

-   :material-github:{ .lg .middle } **Contributing**

    ---

    Help improve DevOpsMaestro

    [:octicons-arrow-right-24: Contributing](development/contributing.md)

</div>
