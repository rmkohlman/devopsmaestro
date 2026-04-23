# DevOpsMaestro

**kubectl-style CLI toolkit for containerized development environments.**

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![CI](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml/badge.svg)](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)

---

## What is DevOpsMaestro?

DevOpsMaestro (`dvm`) is a kubectl-style CLI for managing containerized development environments. It organizes your codebases and dev containers using a clear object hierarchy that matches real-world team structures.

## Object Hierarchy

```
Ecosystem → Domain → System → App → Workspace
   (org)    (context) (group) (code)  (dev env)
```

DevOpsMaestro organizes your development environments using a clear hierarchy that matches real-world organizational structures.

## Key Features

- :material-kubernetes: **kubectl-style commands** - Familiar `get`, `create`, `delete`, `apply` patterns
- :material-sitemap: **Object hierarchy** - Ecosystem → Domain → System → App → Workspace for organized development
- :material-docker: **Multi-platform** - OrbStack, Docker Desktop, Podman, Colima
- :material-cube-outline: **Container-native** - Isolated dev environments with Neovim pre-configured
- :material-database: **Database-backed** - SQLite storage for apps, workspaces, plugins
- :material-file-document: **YAML configuration** - Declarative workspace definitions
- :material-palette: **Hierarchical theme system** - Themes cascade through the object hierarchy

---

## Quick Install

=== "Homebrew (Recommended)"

    ```bash
    brew tap rmkohlman/tap
    brew install devopsmaestro
    
    # Verify installation
    dvm version
    ```

=== "From Source"

    ```bash
    git clone https://github.com/rmkohlman/devopsmaestro.git
    cd devopsmaestro
    go build -o dvm .
    sudo mv dvm /usr/local/bin/
    ```

---

## Quick Example

### Set up your development hierarchy

```bash
# Create organizational structure (one-time setup)
dvm admin init                                    # Initialize dvm
dvm create ecosystem mycompany                        # Top-level platform
dvm create domain mycompany/backend                   # Bounded context
dvm create system mycompany/backend/services          # System grouping
dvm create app mycompany/backend/services/api-service # Your application
dvm create workspace mycompany/backend/services/api-service/dev  # Dev environment
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
# Set themes at different levels (themes cascade down)
dvm set theme coolnight-ocean --workspace dev          # Workspace-specific
dvm set theme coolnight-synthwave --app api-service    # App-wide default
dvm set theme coolnight-nord --domain backend          # Domain-wide default

# Import themes from URL or GitHub
dvm apply -f https://example.com/theme.yaml           # Import from URL
dvm apply -f github:user/repo/theme.yaml              # Import from GitHub
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

    Complete reference for all dvm commands

    [:octicons-arrow-right-24: dvm Commands](dvm/commands.md)

-   :material-cog:{ .lg .middle } **Configuration**

    ---

    YAML schemas, shell completion, and customization

    [:octicons-arrow-right-24: Configuration](configuration/yaml-schema.md)

-   :material-book:{ .lg .middle } **YAML Reference**

    ---

    Complete YAML schemas for all resource types

    [:octicons-arrow-right-24: Reference](reference/index.md)

</div>
