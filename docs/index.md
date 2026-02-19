# DevOpsMaestro

**kubectl-style CLI toolkit for containerized development environments with hierarchical theme management.**

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![CI](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml/badge.svg)](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)

---

## What is DevOpsMaestro?

DevOpsMaestro provides two powerful CLI tools for modern development workflows:

| Tool | Binary | Description |
|------|--------|-------------|
| **DevOpsMaestro** | `dvm` | Workspace and app management with container-native dev environments and hierarchical theme system |
| **NvimOps** | `nvp` | Standalone Neovim plugin & theme manager with 38+ curated plugins and 34+ themes |

## Object Hierarchy

```
Ecosystem → Domain → App → Workspace
   (org)    (context) (code)  (dev env)
```

DevOpsMaestro organizes your development environments using a clear hierarchy that matches real-world organizational structures.

## Key Features

### dvm - Workspace Management

- :material-kubernetes: **kubectl-style commands** - Familiar `get`, `create`, `delete`, `apply` patterns
- :material-sitemap: **Object hierarchy** - Ecosystem → Domain → App → Workspace for organized development
- :material-docker: **Multi-platform** - OrbStack, Docker Desktop, Podman, Colima
- :material-cube-outline: **Container-native** - Isolated dev environments with Neovim pre-configured
- :material-cog-outline: **Default Nvim setup** - New workspaces auto-configured with lazyvim + "core" plugin package
- :material-database: **Database-backed** - SQLite storage for apps, workspaces, plugins
- :material-file-document: **YAML configuration** - Declarative workspace definitions
- :material-palette: **Hierarchical theme system** - Themes cascade through the object hierarchy

### nvp - Neovim Plugin Manager

- :material-file-code: **YAML-based plugins** - Define plugins in YAML, generate Lua
- :material-library: **Built-in library** - 38+ curated plugins ready to install
- :material-palette: **Theme system** - 34+ embedded themes with instant availability
  - **21 CoolNight variants** - Blue, purple, green, warm, red/pink, monochrome, special themes
  - **13+ additional themes** - Catppuccin, Dracula, Everforest, Gruvbox, Tokyo Night, and more
- :material-layers: **Theme hierarchy** - Themes cascade Workspace → App → Domain → Ecosystem → Global
- :material-kubernetes: **kubectl-style IaC** - `dvm apply -f theme.yaml`, URL support, GitHub shorthand
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
    
    # Verify installation (should show v0.12.0)
    dvm version && nvp version
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

### Set up your development hierarchy

```bash
# Create organizational structure (one-time setup)
dvm init                                    # Initialize dvm
dvm create ecosystem mycompany              # Top-level platform
dvm create domain mycompany/backend         # Bounded context  
dvm create app mycompany/backend/api-service  # Your application
dvm create workspace mycompany/backend/api-service/dev  # Dev environment
```

### Add an existing app to dvm

```bash
cd ~/Developer/my-existing-app
dvm create app my-app --from-cwd            # Create app from current dir
dvm use app my-app                          # Set as active
dvm create workspace dev                    # Create a workspace
dvm use workspace dev                       # Set as active
dvm build                                   # Build container
dvm attach                                  # Enter the container
```

### Hierarchical theme management

```bash
# Browse 34+ available themes
dvm get nvim themes

# Set themes at different levels (themes cascade down)
dvm set theme coolnight-ocean --workspace dev          # Workspace-specific
dvm set theme coolnight-synthwave --app api-service    # App-wide default
dvm set theme coolnight-nord --domain backend          # Domain-wide default

# kubectl-style theme management
dvm get nvim theme coolnight-ocean -o yaml             # Export for sharing
dvm apply -f https://example.com/theme.yaml           # Import from URL
dvm apply -f github:user/repo/theme.yaml              # Import from GitHub
```

### Manage Neovim plugins with nvp

```bash
nvp init                              # Initialize
nvp library list                      # Browse 38+ available plugins
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

-   :material-palette:{ .lg .middle } **Theme System**

    ---

    Hierarchical themes, CoolNight collection, and Infrastructure as Code

    [:octicons-arrow-right-24: Theme Hierarchy](advanced/theme-hierarchy.md)

-   :material-package-variant:{ .lg .middle } **Plugin Packages**

    ---

    Complete development environments with 39+ curated plugins

    [:octicons-arrow-right-24: Plugin Packages](nvp/packages.md)

-   :material-cog:{ .lg .middle } **Configuration**

    ---

    YAML schemas, plugin system, and customization

    [:octicons-arrow-right-24: Configuration](configuration/yaml-schema.md)

-   :material-github:{ .lg .middle } **Contributing**

    ---

    Help improve DevOpsMaestro

    [:octicons-arrow-right-24: Contributing](development/contributing.md)

</div>
