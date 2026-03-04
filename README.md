# DevOpsMaestro

[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
[![CI](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml/badge.svg)](https://github.com/rmkohlman/devopsmaestro/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rmkohlman/devopsmaestro)](https://golang.org/)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**kubectl-style CLI toolkit for containerized development environments with hierarchical theme management.**

DevOpsMaestro provides three tools:

| Tool | Binary | Description |
|------|--------|-------------|
| **DevOpsMaestro** | `dvm` | App and workspace management with container-native dev environments |
| **NvimOps** | `nvp` | Standalone Neovim plugin & theme manager using YAML |
| **Terminal Operations** | `dvt` | Terminal prompt and configuration management with database-backed plugin system |

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
dvm version   # Should show v0.32.2
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
nvp library list                    # 38+ curated plugins available
nvp library install telescope treesitter lspconfig

# Or sync from external sources (NEW in v0.16.0)
nvp source list                     # Show available sources
nvp source sync lazyvim             # Import LazyVim plugins
nvp source sync lazyvim --dry-run   # Preview first

# Package Management (NEW in v0.16.0)
dvm get nvim packages               # List available packages
dvm get nvim package core           # Show package details
dvm use nvim package my-package     # Set default for new workspaces
dvm get nvim defaults               # Show current defaults

# Install a theme
nvp theme library list              # 34+ themes including 21 CoolNight variants
nvp theme library install tokyonight-custom --use

# Or use library themes directly (no installation needed)
dvm get nvim theme coolnight-ocean  # Works out of the box

# Set hierarchical themes
dvm set theme coolnight-ocean --workspace dev    # Set at workspace level
dvm set theme coolnight-synthwave --app myapp    # Set at app level

# Generate Lua files for Neovim
nvp generate

# Files created in ~/.config/nvim/lua/plugins/nvp/
```

### Terminal Operations (dvt) - Terminal Plugin Manager

```bash
# Initialize database (required for plugin commands)
dvt init

# Install plugins from packages (updates plugin database)
dvt package install core               # Install core terminal package
dvt package install developer          # Install developer package (extends core)

# View installed plugins (now database-backed)
dvt plugin list                        # List all plugins from database
dvt plugin get starship                # Show specific plugin details

# Generate terminal configuration files
dvt plugin generate                    # Generate configs from database

# Note: dvt commands now use consistent database storage
#       Package install immediately updates plugin commands
```

### DevOpsMaestro (dvm) - Workspace Manager

#### Option 1: Add an Existing App

Already have a codebase on your laptop? Add it to dvm:

```bash
# Initialize dvm (one-time setup)
dvm admin init    # Auto-migrates database to latest schema

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

# Optional: Enable SSH agent forwarding for git operations
dvm attach --ssh-agent
```

#### Option 2: Start a New App

Starting fresh? Create a new directory for your app:

```bash
# Initialize dvm (one-time setup)
dvm admin init    # Auto-migrates database to latest schema

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
dvm admin init                          # One-time setup (auto-migrates database)
dvm create eco my-platform              # Create ecosystem (one-time)
dvm create dom backend                  # Create domain (one-time)
dvm create app myapp --from-cwd         # Create app
dvm create ws dev                       # Create workspace
dvm use ws dev                          # Set active
dvm build && dvm attach                 # Build and enter container
# Or with SSH agent: dvm attach --ssh-agent
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
- **Object hierarchy** - Ecosystem → Domain → App → Workspace for organized development
- **Workspace isolation** - Each workspace has dedicated directories (repo, volume, configs)
- **Local OCI registry (Zot)** - Pull-through cache for faster builds and offline support
- **Git repository mirrors** - Store bare git mirrors locally for faster workspace cloning
- **GitRepo-Workspace integration** - Associate workspaces with GitRepos for automatic cloning
- **Auto-sync on attach** - Optionally sync mirrors before attaching (can be skipped with `--no-sync`)
- **SSH agent forwarding** - Opt-in SSH key access without mounting private keys into containers
- **Package management** - kubectl-style CRUD operations for NvimPackage resources
- **Defaults management** - Set default nvim packages for new workspaces
- **Auto-migration** - Database migrations run automatically on startup
- **GitHub directory URL support** - Apply entire directories of YAML files with `dvm apply -f github:user/repo/directory/`
- **Secret provider system** - Pluggable secret resolution (Keychain, environment variables)
- **Multi-platform** - OrbStack, Docker Desktop, Podman, Colima
- **Container-native** - Isolated dev environments with Neovim pre-configured
- **Default Nvim setup** - New workspaces auto-configured with lazyvim + "core" plugin package
- **Custom Resource Definitions** - Define custom resource types with OpenAPI V3 schema validation
- **Database-backed** - SQLite storage for apps, workspaces, plugins, git repositories
- **YAML configuration** - Declarative workspace definitions
- **Hierarchical theme system** - Themes cascade through the object hierarchy

### nvp - Neovim Plugin Manager

- **YAML-based plugins** - Define plugins in YAML, generate Lua
- **Built-in library** - 38+ curated plugins ready to install
- **External source sync** - Import plugins from external sources like LazyVim
- **Theme system** - 34+ embedded themes available instantly (no installation needed)
  - **21 CoolNight variants** - blue, purple, green, warm, red/pink, monochrome, special
  - **13+ additional themes** - Catppuccin, Dracula, Everforest, Gruvbox, and more
- **Theme hierarchy** - Themes cascade Workspace → App → Domain → Ecosystem → Global
- **kubectl-style IaC** - `dvm apply -f theme.yaml`, URL support, GitHub shorthand
- **Theme override** - User themes override library themes with same name
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
| gitrepo/gitrepos | `repo`, `gr` | `dvm get repo` |
| context | `ctx` | `dvm get ctx` |
| platforms | `plat` | `dvm get plat` |
| library | `lib` | `dvm lib ls plugins` |
| plugins | `np` | `dvm lib ls np` |
| themes | `nt` | `dvm lib ls nt` |
| terminal prompts | `tp` | `dvm lib ls tp` |

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

# Workspaces
dvm create workspace <name>   # Create workspace
dvm create workspace <name> --repo <gitrepo-name>  # Create with git repository
dvm get workspaces            # List workspaces (or: dvm get ws)
dvm delete workspace <name>   # Delete workspace
dvm use workspace <name>      # Set active workspace

# Context
dvm get context               # Show active ecosystem/domain/app/workspace

# Package Management (v0.16.0+)
dvm get nvim packages         # List all available packages
dvm get nvim package <name>   # Show package details
dvm apply -f package.yaml     # Apply package from YAML file
dvm edit nvim package <name>  # Edit package in default editor
dvm delete nvim package <name> # Remove a package

# Defaults Management (v0.16.0+)
dvm use nvim package <name>   # Set default package for new workspaces
dvm use nvim package none     # Clear default package
dvm get nvim defaults         # Show current defaults

# Library Browsing (v0.22.0+)
dvm library list plugins              # List available nvim plugins (38+)
dvm library list themes               # List available nvim themes (34+)
dvm library list nvim packages        # List nvim package bundles
dvm library list terminal prompts     # List terminal prompts
dvm library list terminal plugins     # List shell plugins
dvm library list terminal packages    # List terminal bundles
dvm library show plugin <name>        # Show plugin details
dvm library show theme <name>         # Show theme details
dvm lib ls np                         # Short aliases work

# Terminal Configuration (v0.22.0+)
dvm set terminal prompt -w <workspace> <name>    # Set workspace terminal prompt
dvm set terminal plugin -w <workspace> <plugins> # Set workspace terminal plugins
dvm set terminal package -w <workspace> <name>   # Set workspace terminal package

# Theme Management (v0.12.0+)
dvm get nvim themes           # List all 34+ available themes
dvm get nvim theme <name>     # Show theme details
dvm get nvim theme <name> -o yaml  # Export theme as YAML
dvm set theme <name>          # Set theme for current workspace
dvm set theme <name> --app    # Set theme for current app
dvm set theme <name> --domain # Set theme for current domain
dvm set theme <name> --ecosystem # Set theme for current ecosystem
dvm apply -f theme.yaml       # Apply theme from file
dvm apply -f https://example.com/theme.yaml  # Apply from URL
dvm apply -f github:user/repo/theme.yaml     # Apply from GitHub

# Build & Runtime
dvm build                     # Build workspace container
dvm build --no-cache          # Build without using registry cache
dvm build --push              # Build and push to local registry
dvm build --registry <url>    # Use custom registry endpoint
dvm attach                    # Attach to workspace (auto-syncs GitRepo if configured)
dvm attach --ssh-agent        # Attach with SSH agent forwarding
dvm attach --no-sync          # Attach without syncing GitRepo mirror
dvm detach                    # Stop workspace container

# Registry Management (v0.21.0+)
dvm create registry <name> --type <type> --port <port>  # Create registry resource
dvm get registries            # List all registry resources
dvm get registry <name>       # Show specific registry
dvm delete registry <name>    # Delete registry resource

# Registry Lifecycle (v0.25.0+)
dvm start registry <name>     # Start registry
dvm stop registry <name>      # Stop registry

# Registry Rollout Commands (v0.25.0+)
dvm rollout restart registry <name>  # Restart registry
dvm rollout status registry <name>   # Show rollout status
dvm rollout history registry <name>  # Show rollout history
dvm rollout undo registry <name>     # Rollback (not yet implemented)

# Registry Defaults (v0.30.0+)
dvm registry enable <type> --lifecycle <mode>     # Enable registry type (oci, pypi, npm, go, http)
dvm registry disable <type>                       # Disable registry type
dvm registry set-default <type> <registry-name>   # Set default registry for type
dvm registry get-defaults                         # Show all configured defaults
dvm registry get-defaults -o yaml                 # Output as YAML

# Git Repository Management (v0.20.0+)
dvm create gitrepo <name> --url <git-url>         # Create git repository mirror
dvm create gitrepo <name> --url <url> --auth-type ssh --credential <name>  # With authentication
dvm create gitrepo <name> --url <url> --no-sync  # Create without initial sync
dvm get gitrepos                                  # List all git repositories
dvm get gitrepo <name>                            # Get specific repository
dvm sync gitrepo <name>                           # Sync a repository mirror
dvm sync gitrepos                                 # Sync all repositories
dvm delete gitrepo <name>                         # Delete repository (removes mirror)
dvm delete gitrepo <name> --keep-mirror           # Delete metadata, keep mirror

# Custom Resource Definitions (v0.29.0+)
dvm apply -f crd.yaml                             # Register a CRD
dvm get crds                                      # List all CRDs
dvm get crd <name>                                # Show specific CRD
dvm delete crd <name>                             # Delete a CRD
dvm apply -f resource.yaml                        # Create custom resource instance
dvm get <plural-name>                             # List custom resources
dvm delete <kind> <name>                          # Delete custom resource

# Configuration
dvm apply -f workspace.yaml   # Apply YAML configuration
dvm get platforms             # List detected container platforms
```

### nvp Commands

```bash
# Plugins
nvp library list              # List available plugins (38+ curated plugins)
nvp library install <name>    # Install from library
nvp apply -f plugin.yaml      # Apply plugin from file
nvp apply -f https://example.com/plugin.yaml  # Apply from URL (auto-detected)
nvp apply -f github:user/repo/plugin.yaml     # GitHub shorthand
nvp apply -f -                # Apply from stdin

# External Sources (v0.16.0+)
nvp source list               # List available external sources
nvp source show <name>        # Show source details
nvp source sync <name>        # Sync plugins from external source
nvp source sync <name> --dry-run  # Preview sync without changes
nvp source sync <name> -l category=lang  # Filter by labels
nvp source sync <name> --tag v15.0.0     # Sync specific version

# Themes
nvp theme library list        # List available themes (34+ themes)
nvp theme library show <name> # View theme details  
nvp theme library install <name> --use
nvp theme apply -f theme.yaml # Apply theme from file

# Note: Library themes are automatically available, no installation needed
dvm get nvim themes           # Shows user + library themes (34+ total)
dvm get nvim theme <name>     # Works with any library theme

# CoolNight Theme Variants (21 variants available)
dvm get nvim theme coolnight-ocean      # Blue: default ocean theme
dvm get nvim theme coolnight-arctic     # Blue: ice-cold variant
dvm get nvim theme coolnight-midnight   # Blue: deep night variant
dvm get nvim theme coolnight-synthwave  # Purple: cyberpunk aesthetic
dvm get nvim theme coolnight-violet     # Purple: soft violet
dvm get nvim theme coolnight-grape      # Purple: rich grape
dvm get nvim theme coolnight-matrix     # Green: Matrix-inspired
dvm get nvim theme coolnight-forest     # Green: forest theme
dvm get nvim theme coolnight-mint       # Green: fresh mint
dvm get nvim theme coolnight-sunset     # Warm: golden sunset
dvm get nvim theme coolnight-ember      # Warm: burning ember
dvm get nvim theme coolnight-gold       # Warm: golden theme
dvm get nvim theme coolnight-rose       # Red/Pink: soft rose
dvm get nvim theme coolnight-crimson    # Red/Pink: bold crimson
dvm get nvim theme coolnight-sakura     # Red/Pink: cherry blossom
dvm get nvim theme coolnight-mono-charcoal  # Monochrome: dark charcoal
dvm get nvim theme coolnight-mono-slate     # Monochrome: blue-gray
dvm get nvim theme coolnight-mono-warm      # Monochrome: warm gray
dvm get nvim theme coolnight-nord       # Special: Nord-inspired
dvm get nvim theme coolnight-dracula    # Special: Dracula-inspired
dvm get nvim theme coolnight-solarized  # Special: Solarized-inspired

# Generate
nvp generate                  # Generate Lua files
```

### dvt Commands (Terminal Operations)

```bash
# Plugin Management (Database-backed)
dvt init                      # Initialize database (required for plugin commands)
dvt plugin list               # List terminal plugins from database
dvt plugin get <name>         # Show specific plugin details
dvt plugin generate           # Generate terminal configuration files

# Package Management  
dvt package list              # List available terminal packages
dvt package get <name>        # Show package details
dvt package install <name>    # Install plugins/prompts from package (updates plugin commands)

# Terminal Prompt Management
dvt get prompts               # List all terminal prompts
dvt get prompts --type starship # Filter by prompt type
dvt get prompt <name>         # Show specific prompt details
dvt prompt apply -f prompt.yaml  # Apply prompt from file
dvt prompt apply -f github:user/repo/prompt.yaml  # Apply from GitHub
dvt prompt generate <name>    # Generate starship.toml config
dvt prompt set <name>         # Set active prompt for workspace
dvt prompt delete <name>      # Delete a prompt

# WezTerm Configuration (v0.12.2+)
dvt wezterm list              # List available WezTerm presets
dvt wezterm show <name>       # Show preset details
dvt wezterm generate <name>   # Generate wezterm.lua with theme colors
dvt wezterm apply <name>      # Apply configuration to ~/.wezterm.lua
```

---

## Secrets

DevOpsMaestro supports pluggable secret providers for secure credential management.

### Supported Providers

| Provider | Backend | Platform | Priority |
|----------|---------|----------|----------|
| `keychain` | macOS Keychain | macOS | Default on macOS |
| `env` | Environment variables | All | Fallback |

### Usage

Secrets can be referenced in YAML using inline syntax:

```yaml
config: |
  api_key = "${secret:my-api-key}"           # Uses default provider
  token = "${secret:github-token:keychain}"  # Explicit provider
```

### Adding Secrets to Keychain

```bash
# Add a secret to macOS Keychain for dvm
security add-generic-password -s devopsmaestro -a github-token -w "your-token-here"
```

### Environment Variables

Secrets can also be set via environment variables:
- `DVM_SECRET_GITHUB_TOKEN` - GitHub API token
- `DVM_SECRET_<NAME>` - Any secret (uppercase, underscores for dashes)

For backward compatibility, `GITHUB_TOKEN` is also checked.

---

## Git Repository Mirrors

DevOpsMaestro can create and manage local bare git mirrors for faster workspace cloning and offline access.

### Benefits

- **Faster workspace cloning** - Clone from local mirrors instead of remote repositories
- **Workspace integration** - Associate workspaces with GitRepos for automatic cloning
- **Auto-sync on attach** - Optionally sync mirrors before attaching to workspaces
- **Offline access** - Work with cached repositories without network
- **URL normalization** - SSH and HTTPS URLs for the same repo map to the same mirror
- **One-to-many model** - One mirror can serve multiple workspaces

### Storage Location

Git mirrors are stored as bare repositories in:
```
~/.devopsmaestro/repos/<slug>/
```

Where `<slug>` is normalized from the repository URL (e.g., `github.com/user/repo`).

### Usage

```bash
# Create a git repository mirror
dvm create gitrepo devopsmaestro --url https://github.com/rmkohlman/devopsmaestro

# Create with SSH authentication
dvm create gitrepo my-private-repo \
  --url git@github.com:user/repo.git \
  --auth-type ssh \
  --credential my-ssh-key

# Create without syncing immediately
dvm create gitrepo future-repo --url https://github.com/user/repo --no-sync

# Create a workspace associated with a GitRepo
dvm create workspace dev --repo devopsmaestro
# → Clones from local mirror to ~/.devopsmaestro/workspaces/{slug}/repo/

# Attach with auto-sync (default if GitRepo.AutoSync=true)
dvm attach
# → Syncs mirror before attaching if GitRepo configured

# Attach without syncing mirror
dvm attach --no-sync
# → Skip mirror sync for faster attach or offline work

# List all git repositories
dvm get gitrepos
dvm get gitrepos -o yaml

# Get specific repository details
dvm get gitrepo devopsmaestro

# Sync a repository mirror (fetch updates)
dvm sync gitrepo devopsmaestro

# Sync all repository mirrors
dvm sync gitrepos

# Delete repository (removes mirror by default)
dvm delete gitrepo devopsmaestro

# Delete metadata but keep mirror
dvm delete gitrepo devopsmaestro --keep-mirror
```

### Workspace Integration Workflow

```bash
# 1. Create a git repository mirror
dvm create gitrepo my-project --url https://github.com/myorg/my-project

# 2. Create a workspace with the repository
dvm create workspace dev --repo my-project
# → Clones from local mirror to workspace's repo/ directory

# 3. Attach to workspace (auto-syncs if GitRepo.AutoSync=true)
dvm attach
# → Mirror synced before attach (can be skipped with --no-sync)

# 4. Multiple workspaces can share the same mirror
dvm create workspace staging --repo my-project
dvm create workspace prod --repo my-project
# → Each gets independent clone from the same mirror
```

### Authentication Types

| Auth Type | Description | Credential Required |
|-----------|-------------|---------------------|
| `none` | Public repositories (default) | No |
| `ssh` | SSH key authentication | Yes (SSH key name) |
| `token` | GitHub Personal Access Token | Yes (token name) |

### URL Normalization

Both SSH and HTTPS URLs normalize to the same slug:
```bash
# These both create the same mirror:
dvm create gitrepo repo1 --url https://github.com/user/repo.git
dvm create gitrepo repo2 --url git@github.com:user/repo.git
# → Both map to: ~/.devopsmaestro/repos/github.com/user/repo/
```

---

## Registry Management

DevOpsMaestro provides a flexible registry system supporting multiple registry types. Each registry is managed as a resource with database persistence, supporting OCI container images (Zot), Go modules (Athens), Python packages (devpi), npm packages (Verdaccio), and HTTP proxy caching (Squid).

### Supported Registry Types

| Type | Purpose | Status |
|------|---------|--------|
| `zot` | OCI container image registry | ✅ Full support |
| `devpi` | Python package index/proxy | ✅ Full support |
| `verdaccio` | npm package proxy | ✅ Full support |
| `athens` | Go module proxy | 🚧 Stub implementation |
| `squid` | HTTP proxy cache | ✅ Full support |

### Benefits

- **Faster builds** - Base images are cached locally after the first pull
- **Offline support** - Build workspaces without network access if images are cached
- **Rate limit avoidance** - Reduce pulls from Docker Hub (avoids 100 pulls/6 hours limit)
- **Python package caching** - Cache PyPI packages locally with devpi for faster pip installs
- **npm package caching** - Cache npm packages locally with verdaccio for faster npm installs
- **Multi-registry support** - Run multiple registries simultaneously (zot, devpi, verdaccio, athens, etc.)
- **Resource-based management** - Registries are database-backed resources like other DevOpsMaestro objects
- **ServiceFactory pattern** - Extensible design for adding new registry types

### Quick Start

```bash
# Create a registry resource
dvm create registry myregistry --type zot --port 5000

# Start the registry
dvm start registry myregistry

# Check rollout status
dvm rollout status registry myregistry

# Build with caching (default behavior)
dvm build

# Restart the registry (stop + start)
dvm rollout restart registry myregistry

# View rollout history
dvm rollout history registry myregistry

# Stop the registry
dvm stop registry myregistry

# List all registry resources
dvm get registries

# Delete registry resource
dvm delete registry myregistry
```

### Registry Resource Management

Registries are managed as database-backed resources:

```bash
# Create a Zot registry (OCI images)
dvm create registry myregistry --type zot --port 5000

# Create a devpi registry (Python packages)
dvm create registry pypi --type devpi --port 3141

# Create a verdaccio registry (npm packages)
dvm create registry npm --type verdaccio --port 4873

# Create registries for other types (stub implementations)
dvm create registry athens --type athens --port 3000

# List all registries
dvm get registries

# Get specific registry details
dvm get registry myregistry

# Delete a registry (stops if running)
dvm delete registry myregistry
```

### Registry Configuration

Each registry type has its own configuration stored in the database. For Zot registries, the configuration includes:

- **Port**: HTTP port for the registry
- **Storage path**: Local directory for cached images
- **Mirror configuration**: Pull-through cache settings for upstream registries

### Storage Location

**For Zot registries:**
- **Zot binary**: `~/.devopsmaestro/bin/zot`
- **Registry data**: `~/.devopsmaestro/registry/<registry-name>/`
- **Registry config**: `~/.devopsmaestro/registry/<registry-name>/config.json`
- **Logs**: `~/.devopsmaestro/registry/<registry-name>/zot.log`

**For devpi registries:**
- **devpi-server binary**: Managed by pipx (installed via `pipx install devpi-server`)
- **Registry data**: `~/.dvm/devpi/<registry-name>/`
- **Server config**: `~/.dvm/devpi/<registry-name>/devpi-server/`
- **Logs**: Managed by devpi-server

**For verdaccio registries:**
- **verdaccio binary**: Managed by npm (installed via `npm install -g verdaccio`)
- **Registry data**: `~/.devopsmaestro/registry/<registry-name>/storage/`
- **Registry config**: `~/.devopsmaestro/registry/<registry-name>/config.yaml`
- **Logs**: `~/.devopsmaestro/registry/<registry-name>/verdaccio.log`

### Runtime Operations

```bash
# Start a registry
dvm start registry myregistry

# Stop a registry
dvm stop registry myregistry

# Restart a registry (kubectl-style rollout)
dvm rollout restart registry myregistry

# Check rollout status
dvm rollout status registry myregistry

# View rollout history
dvm rollout history registry myregistry
```

### Build Integration

The registry automatically acts as a pull-through cache for base images:

```bash
# First build: Pulls from Docker Hub, caches locally
dvm build

# Second build: Uses cached image (much faster!)
dvm build

# Force fresh pull from upstream
dvm build --no-cache

# Build and store in local registry
dvm build --push
```

### Multiple Registries

You can run multiple registries simultaneously:

```bash
# Create different registry types
dvm create registry images --type zot --port 5000
dvm create registry pypi --type devpi --port 3141
dvm create registry npm --type verdaccio --port 4873
dvm create registry gomodules --type athens --port 3000

# Start them all
dvm start registry images
dvm start registry pypi
dvm start registry npm
dvm start registry gomodules

# Check status of each
dvm rollout status registry images
dvm rollout status registry pypi
dvm rollout status registry npm
dvm rollout status registry gomodules

# Stop specific registry
dvm stop registry images
```

### Registry Defaults (v0.30.0+)

DevOpsMaestro now supports registry type aliases and automatic environment variable injection for simplified registry management.

#### Type Aliases

Registry types can be referenced by intuitive aliases:

| Alias | Registry Type | Purpose |
|-------|---------------|---------|
| `oci` | Zot | OCI container images |
| `pypi` | devpi | Python packages |
| `npm` | Verdaccio | npm packages |
| `go` | Athens | Go modules |
| `http` | Squid | HTTP proxy cache |

#### Enable/Disable Registry Types

```bash
# Enable a registry type with lifecycle mode
dvm registry enable pypi --lifecycle auto       # Starts when needed, stays running
dvm registry enable npm --lifecycle on-demand   # Starts/stops as needed
dvm registry enable oci --lifecycle manual      # User controlled

# Disable a registry type (stops if running)
dvm registry disable pypi
```

#### Lifecycle Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| `auto` | Starts when needed, stays running | Development workstations |
| `on-demand` | Starts when needed, stops after idle | Resource-constrained systems |
| `manual` | User explicitly controls | Production-like environments |

#### Set Default Registries

Associate a concrete registry with a type alias:

```bash
# Create a registry
dvm create registry my-python-cache --type devpi --port 3141

# Set as default for pypi type
dvm registry set-default pypi my-python-cache

# Now all pypi operations use my-python-cache
```

#### View Registry Defaults

```bash
# Show all configured defaults
dvm registry get-defaults

# Output as YAML
dvm registry get-defaults -o yaml

# Example output:
# pypi: my-python-cache (auto)
# npm: my-npm-proxy (on-demand)
# oci: local-images (manual)
```

#### Automatic Environment Injection

When you attach to a workspace, DevOpsMaestro automatically injects environment variables for enabled registries:

| Registry Type | Environment Variables | Purpose |
|---------------|----------------------|---------|
| `pypi` (devpi) | `PIP_INDEX_URL`, `PIP_TRUSTED_HOST` | Points pip to local cache |
| `npm` (verdaccio) | `NPM_CONFIG_REGISTRY` | Points npm to local cache |
| `go` (athens) | `GOPROXY` | Points Go to local proxy |
| `http` (squid) | `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` | Routes HTTP through proxy |

**Example workflow:**

```bash
# Enable PyPI registry
dvm registry enable pypi --lifecycle auto

# Create and set default registry
dvm create registry my-pypi --type devpi --port 3141
dvm registry set-default pypi my-pypi

# Start the registry
dvm start registry my-pypi

# Attach to workspace (environment variables automatically set)
dvm attach
# Inside container:
# → PIP_INDEX_URL=http://host.docker.internal:3141/root/pypi/+simple/
# → PIP_TRUSTED_HOST=host.docker.internal

# pip install now uses local cache automatically!
pip install requests
```

#### Pre-flight Checks

DevOpsMaestro validates registry health before build and attach operations:

```bash
# Pre-flight checks run automatically
dvm build    # Validates enabled registries are healthy
dvm attach   # Validates enabled registries are healthy

# Warnings are non-fatal - operations continue if registry is unreachable
# [WARN] Registry 'my-pypi' is unreachable, continuing without cache
```

#### Security Features

Registry URLs are validated for security:

- **Scheme validation** - Only http://, https://, localhost allowed
- **No embedded credentials** - Rejects user:password@ format
- **External registry warnings** - Warns about non-localhost registries
- **Path traversal protection** - Blocks directory traversal attempts

### Troubleshooting

**Registry won't start:**
```bash
# Check if port is already in use
lsof -i :5000

# Check registry resource configuration
dvm get registry myregistry

# Check rollout status for errors
dvm rollout status registry myregistry
```

**Port conflict:**
```bash
# Create registry with different port
dvm delete registry myregistry
dvm create registry myregistry --type zot --port 5050
dvm start registry myregistry
```

**Check running registries:**
```bash
# List all registry resources
dvm get registries

# Check status of each registry
dvm rollout status registry myregistry
```

**Environment variables not set:**
```bash
# Verify registry type is enabled
dvm registry get-defaults

# Verify registry is running
dvm rollout status registry myregistry

# Check pre-flight status
dvm build  # Check for pre-flight warnings
```

---

## Custom Resource Definitions (CRDs)

DevOpsMaestro supports Kubernetes-style Custom Resource Definitions (CRDs) for defining custom resource types with schema validation. This allows you to extend DevOpsMaestro with your own resource types while maintaining the same kubectl-style workflow.

### Key Features

- **Define custom resource types** with OpenAPI V3 schema validation
- **Support for all scope types** (Workspace, App, Domain, Ecosystem, Global)
- **Full CRUD operations** on custom resources
- **Case-insensitive kind lookup** - `dvm get Database` and `dvm get database` both work
- **Built-in kind protection** - Cannot override core types like Workspace, App, Domain, etc.
- **kubectl-style workflow** - Familiar `apply`, `get`, `delete` commands

### Quick Start

```bash
# Register a CRD
dvm apply -f database-crd.yaml

# List registered CRDs
dvm get crds

# Create a custom resource
dvm apply -f mydb.yaml

# Get custom resources by kind
dvm get databases

# Delete a custom resource
dvm delete database mydb
```

### CRD YAML Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases
spec:
  group: mycompany.io
  names:
    kind: Database
    singular: database
    plural: databases
    shortNames:
      - db
  scope: App
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required:
            - spec
          properties:
            spec:
              type: object
              required:
                - engine
              properties:
                engine:
                  type: string
                  enum: ["postgres", "mysql", "sqlite"]
                replicas:
                  type: integer
                  minimum: 1
                  maximum: 10
```

### Custom Resource YAML Example

Once the CRD is registered, create instances of your custom resource:

```yaml
apiVersion: mycompany.io/v1
kind: Database
metadata:
  name: mydb
spec:
  engine: postgres
  replicas: 3
```

### Schema Validation

CRDs support full OpenAPI V3 schema validation features:

| Feature | Example | Description |
|---------|---------|-------------|
| **Types** | `type: string` | string, integer, number, boolean, object, array |
| **Required** | `required: ["engine"]` | Mark fields as required |
| **Enums** | `enum: ["a", "b"]` | Restrict to specific values |
| **Min/Max** | `minimum: 1, maximum: 10` | Numeric constraints |
| **Patterns** | `pattern: "^[a-z]+$"` | Regex validation for strings |
| **Nested** | `properties: { spec: { ... } }` | Nested object schemas |

### Scope Types

Custom resources can be scoped at different levels of the object hierarchy:

| Scope | Description | Example Use Case |
|-------|-------------|------------------|
| **Workspace** | Scoped to a workspace | Workspace-specific configurations |
| **App** | Scoped to an app | Application-level resources (databases, services) |
| **Domain** | Scoped to a domain | Domain-wide shared resources |
| **Ecosystem** | Scoped to an ecosystem | Ecosystem-level infrastructure |
| **Global** | Available globally | System-wide resource types |

### Working with Custom Resources

```bash
# Register a CRD
dvm apply -f database-crd.yaml

# Create custom resource instances
dvm apply -f production-db.yaml
dvm apply -f staging-db.yaml

# List all custom resources of a kind
dvm get databases           # Plural form
dvm get database            # Singular form works too
dvm get db                  # Short name also works

# Get specific custom resource
dvm get database production-db
dvm get database production-db -o yaml

# Delete custom resource
dvm delete database production-db

# List all registered CRDs
dvm get crds

# Get specific CRD details
dvm get crd databases

# Delete a CRD (removes all instances)
dvm delete crd databases
```

### Example: Service CRD

```yaml
apiVersion: devopsmaestro.io/v1
kind: CustomResourceDefinition
metadata:
  name: services
spec:
  group: platform.io
  names:
    kind: Service
    singular: service
    plural: services
    shortNames:
      - svc
  scope: App
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required:
            - spec
          properties:
            spec:
              type: object
              required:
                - port
                - protocol
              properties:
                port:
                  type: integer
                  minimum: 1
                  maximum: 65535
                protocol:
                  type: string
                  enum: ["http", "https", "tcp", "grpc"]
                healthCheck:
                  type: object
                  properties:
                    path:
                      type: string
                    interval:
                      type: integer
```

### Benefits

- **Extensibility** - Add custom resource types without modifying core DevOpsMaestro code
- **Validation** - OpenAPI V3 schemas ensure resources are valid before creation
- **Consistency** - Same kubectl-style workflow for all resources (core + custom)
- **Flexibility** - Define resources at the scope level that makes sense for your use case
- **Type Safety** - Schema validation catches errors early

---

## Applying Resources

```bash
# Apply single file
dvm apply -f plugin.yaml
dvm apply -f package.yaml        # NEW: Apply NvimPackage

# Apply from GitHub
dvm apply -f github:user/repo/plugins/telescope.yaml
dvm apply -f github:user/repo/packages/my-package.yaml  # NEW: Apply package

# Apply entire directory from GitHub
dvm apply -f github:user/repo/plugins/
dvm apply -f github:user/repo/packages/   # NEW: Apply all packages in directory

# Apply from your personal config repo
dvm apply -f github:rmkohlman/dvm-config/plugins/
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
  ssh_agent_forwarding: true  # Optional: Enable SSH agent forwarding
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

### Package YAML

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: my-package
  description: "My custom package"
  category: custom
  labels:
    source: user
spec:
  plugins:
    - telescope
    - treesitter
    - lspconfig
  extends: core  # optional - inherit from another package
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
| **GitHub Directory** | `-f github:user/repo/plugins/` | Entire GitHub directory |
| **Stdin** | `-f -` | Read from stdin |

```bash
# Apply from local file
dvm apply -f workspace.yaml
nvp apply -f plugin.yaml

# Apply from URL (auto-detected)
nvp apply -f https://raw.githubusercontent.com/user/repo/main/plugin.yaml

# Apply from GitHub (shorthand)
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml

# Apply entire directory from GitHub
dvm apply -f github:rmkohlman/nvim-yaml-plugins/plugins/
dvm apply -f github:rmkohlman/dvm-config/themes/

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
