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
dvm version   # Should show v0.22.0
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
dvm init    # Auto-migrates database to latest schema

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
dvm init    # Auto-migrates database to latest schema

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
dvm init                                # One-time setup (auto-migrates database)
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

# Local OCI Registry Management (v0.21.0+)
dvm registry start            # Start Zot registry (pull-through cache)
dvm registry start --port 5001 # Start on custom port
dvm registry start --foreground # Run in foreground mode
dvm registry stop             # Stop registry
dvm registry stop --force     # Force stop (kill process)
dvm registry status           # Show registry status
dvm registry status -o wide   # Show detailed status with uptime
dvm registry logs             # View registry logs
dvm registry logs -n 100      # Show last 100 log lines
dvm registry logs --since 10m # Show logs from last 10 minutes
dvm registry prune            # Clean up cached images (interactive)
dvm registry prune --dry-run  # Preview what would be deleted
dvm registry prune --all      # Remove all cached images
dvm registry prune --older-than 7d  # Remove images older than 7 days

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

## Local OCI Registry (Zot)

DevOpsMaestro includes an integrated local OCI registry (Zot) that acts as a pull-through cache for container images. This provides faster builds, offline support, and helps avoid Docker Hub rate limits.

### Benefits

- **Faster builds** - Base images are cached locally after the first pull
- **Offline support** - Build workspaces without network access if images are cached
- **Rate limit avoidance** - Reduce pulls from Docker Hub (avoids 100 pulls/6 hours limit)
- **Local image storage** - Store built workspace images in the local registry
- **Automatic management** - Zot binary auto-downloaded and managed by DevOpsMaestro

### Quick Start

```bash
# Start the registry (one-time setup)
dvm registry start

# Check status
dvm registry status

# Build with caching (default behavior)
dvm build

# Build without cache (fresh pull from upstream)
dvm build --no-cache

# Build and push to local registry
dvm build --push

# View cached images
dvm registry status -o wide

# Clean up old cached images
dvm registry prune --older-than 7d
```

### Configuration

The registry can be configured in `~/.devopsmaestro/config.yaml`:

```yaml
registry:
  enabled: true
  lifecycle: persistent  # persistent | on-demand | manual
  port: 5001
  storage: ~/.devopsmaestro/registry
  idle_timeout: 30m  # for on-demand lifecycle
  mirrors:
    - name: docker-hub
      url: https://index.docker.io
      on_demand: true
      prefix: docker.io
```

### Lifecycle Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `persistent` | Registry runs continuously | Heavy development, frequent builds |
| `on-demand` | Starts on first use, stops after idle timeout | Occasional builds, save resources |
| `manual` | User explicitly starts/stops | Full control over registry |

### Storage Location

- **Zot binary**: `~/.devopsmaestro/bin/zot`
- **Registry data**: `~/.devopsmaestro/registry/`
- **Registry config**: `~/.devopsmaestro/registry/config.json`
- **Logs**: `~/.devopsmaestro/registry/zot.log`

### Port Configuration

The registry uses port **5001** by default (avoids conflict with Docker registry on port 5000). You can customize this:

```bash
dvm registry start --port 5050
```

Or in `config.yaml`:

```yaml
registry:
  port: 5050
```

### Managing Cached Images

```bash
# View registry status with image count
dvm registry status -o wide

# Clean up interactively (prompts for confirmation)
dvm registry prune

# Preview what would be deleted
dvm registry prune --dry-run

# Remove all cached images
dvm registry prune --all --force

# Remove images older than 30 days
dvm registry prune --older-than 30d
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

### Advanced Usage

```bash
# Run registry in foreground (see logs in real-time)
dvm registry start --foreground

# Stop registry gracefully
dvm registry stop

# Force stop (if graceful stop hangs)
dvm registry stop --force

# View recent logs
dvm registry logs -n 50

# View logs from last hour
dvm registry logs --since 1h

# Use custom registry endpoint
dvm build --registry localhost:5050
```

### Troubleshooting

**Registry won't start:**
```bash
# Check if port is already in use
lsof -i :5001

# Try a different port
dvm registry start --port 5050
```

**Zot binary not found:**
The binary is auto-downloaded on first use. If download fails:
```bash
# Check internet connectivity
# The binary is downloaded from: https://github.com/project-zot/zot/releases
```

**Builds still slow:**
```bash
# Verify registry is running
dvm registry status

# Check if images are being cached
dvm registry status -o wide
```

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
