# Changelog

All notable changes to DevOpsMaestro are documented in the [CHANGELOG.md](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) file in the repository.

## Latest Releases

### v0.32.1 (2026-03-04)

**🐛 Bug Fixes & Refactoring**

Fixed error handling in `--repo` flag and extracted watchdog helper:

- **Error handling** - Fixed error message check causing "not found" errors to be misclassified as database errors
- **Error messages** - Improved slug conflict and GitRepo not found error messages with helpful examples
- **Watchdog refactor** - Extracted watchdog helper to dedicated module with `WatchdogConfig` and injection pattern
- **Test coverage** - Added 22 new tests (11 for watchdog, 11 for GitRepo resolution)
- **Test fixes** - Unskipped 4 previously skipped tests after error handling improvements

### v0.32.0 (2026-03-04)

**✨ Feature - `--repo` Flag for App Creation**

Streamlined GitRepo-backed app creation with new `--repo` flag:

- **Accept URL** - `dvm create app my-app --repo https://github.com/user/repo.git`
- **Accept name** - `dvm create app my-app --repo my-existing-repo`
- **Auto-create GitRepo** - Automatically creates GitRepo resource when given a URL
- **Detect duplicates** - Reuses existing GitRepos by URL to avoid duplicates
- **Mutually exclusive** - Cannot use with `--path` or `--from-cwd` flags

**🐛 Bug Fix - Docker Build Hang on Colima**

Fixed Docker buildx + Colima hang where build completes but process doesn't exit:

- **Watchdog mechanism** - Polls for image existence during build
- **Context cancellation** - Terminates hung docker process when image is detected
- **Goroutine execution** - Runs docker build in background for parallel monitoring
- **Timeout protection** - 30-minute overall timeout as fallback
- **Reliable Colima builds** - Enables reliable builds on Colima with Docker runtime (non-containerd)

### v0.31.0 (2026-03-03)

**✨ Feature - Lazygit in Containers**

All development containers now include lazygit for terminal-based git operations:

- **Automatic installation** - Lazygit installed from GitHub releases during container builds
- **Multi-architecture** - Supports x86_64 and ARM64 architectures
- **Multi-distro** - Works with Alpine (musl) and Debian-based images
- **Stable release** - Downloads latest stable version from jesseduffield/lazygit

Every workspace now has `lazygit` available at `/usr/local/bin/lazygit`.

### v0.30.4 (2026-03-03)

**🐛 Bug Fix - Attach Mount Path for GitRepo-backed Workspaces**

Fixed `dvm attach` mounting wrong directory for GitRepo-backed workspaces. Container now correctly mounts workspace repo path instead of empty app path.

### v0.30.3 (2026-03-03)

**🐛 Bug Fix - Database Schema Drift**

Fixed schema drift where `apps.git_repo_id` column existed in database but not in Go code. All 7 App DataStore methods now include `git_repo_id` field.

### v0.30.2 (2026-03-03)

**🐛 Bug Fixes - Theme & Terminal Package Integration**

Four critical fixes for workspace theme inheritance and terminal package prompt generation:

**Theme Inheritance Hierarchy Fix:**
- `dvm build` now properly walks Workspace → App → Domain → Ecosystem → Global hierarchy when resolving themes
- Previously bypassed hierarchy and always used global default theme
- Workspaces now correctly inherit themes from their parent resources

**Terminal Package Prompt Composition Fix:**
- `generateShellConfig()` now loads terminal packages from library and composes prompts from style + extensions
- Previously ignored workspace's `terminal-package` setting, always generating default prompts
- Custom terminal packages like `rmkohlman` now generate rich prompts as designed

**CoolNight Theme Prompt Colors:**
- Added monochromatic `promptColors` gradients to all 21 CoolNight theme variants (except ocean)
- Previously missing `promptColors` section caused harsh ANSI fallback colors
- Terminal prompts now have smooth, cohesive color gradients matching the workspace theme

**Neovim Colorscheme Generation:**
- `dvm build` now generates `theme/colorscheme.lua` containing `vim.api.nvim_set_hl()` calls
- Previously only generated 3 theme files, missing the colorscheme file
- Neovim inside workspaces now displays correct theme colors matching the terminal prompt

**Testing:**
- Added `cmd/build_terminal_package_test.go` (9 tests)
- Added `cmd/build_theme_test.go` (5 tests)
- Added `cmd/set_theme_test.go` (theme setting tests)

### v0.30.1 (2026-03-03)

**🐛 Bug Fixes - Database Schema & YAML Completeness**

Seven critical fixes for database schema completeness and YAML field parsing:

- **Registry database queries** - Fixed missing `storage`, `enabled`, `idle_timeout` columns (GitHub Issue #5)
- **Workspace database queries** - Fixed missing terminal configuration columns (GitHub Issue #8)
- **Terminal package validation** - Now checks database + embedded library (GitHub Issue #7)
- **Library show commands** - Fixed table output for library resources (GitHub Issue #6)
- **Git checkout** - Fixed workspace creation with `--repo` flag (GitHub Issue #9)
- **YAML apply workspace fields** - Fixed incomplete field parsing (GitHub Issue #10)
- **Workspace YAML serialization** - Fixed incomplete YAML output (GitHub Issue #11)

### v0.24.0 (2026-03-01)

**🔄 Registry Resources - Multi-Registry Support**

Major refactor of registry system to support multiple registry types with database-backed resources.

**⚠️ Breaking Changes:**

- **Registry commands now require name as positional argument**:
  - OLD: `dvm registry start` → NEW: `dvm registry start myregistry`
  - OLD: `dvm registry stop` → NEW: `dvm registry stop myregistry`
- **Removed `--name` flag** from all registry runtime commands
- **Removed config-based registry** - Must use Registry Resources instead
- **Must create registry resource first**: `dvm create registry <name> --type <type> --port <port>`

**New Features:**

- **Registry Resource Type** - Database-backed registry management:
  - `dvm create registry <name> --type <type> --port <port>` - Create registry
  - `dvm get registries` - List all registries
  - `dvm get registry <name>` - Show specific registry
  - `dvm delete registry <name>` - Delete registry
  
- **Multi-Registry Support** - Run multiple registry types simultaneously:
  - `zot` - OCI container images (full support)
  - `athens` - Go module proxy (stub)
  - `devpi` - Python package index (stub)
  - `verdaccio` - npm registry (stub)
  - `squid` - HTTP proxy cache (stub)

- **ServiceFactory Pattern** - Extensible architecture for registry services:
  - Each registry type has dedicated service implementation
  - Database persistence for all registry configurations
  - Independent lifecycle management per registry

**Updated Commands:**

- `dvm registry start <name>` - Start specific registry (name REQUIRED)
- `dvm registry stop <name>` - Stop specific registry (name REQUIRED)
- `dvm registry status` - List all registries (no name = show all)
- `dvm registry status <name>` - Show specific registry status

**Migration Guide:**

```bash
# OLD approach (no longer works):
dvm registry start

# NEW approach (required):
# Step 1: Create registry resource
dvm create registry myregistry --type zot --port 5000

# Step 2: Start the registry
dvm registry start myregistry

# Check status
dvm registry status myregistry
# or list all:
dvm registry status
```

**Technical Details:**

- New Registry Resource model with database persistence
- ServiceFactory pattern for multi-registry support
- Zot service implementation with full lifecycle management
- Stub implementations for Athens, Devpi, Verdaccio, Squid
- Full test coverage for CRUD operations and runtime commands

### v0.22.0 (2026-02-28)

**🔗 Integration - Consolidated Library & Terminal Configuration**

Complete integration of nvp/dvt functionality into dvm. Users can now browse libraries and configure terminal settings without leaving the dvm CLI.

**New Features:**

- **Library Browsing Commands** - Browse all libraries from dvm:
  - `dvm library list plugins` - 38+ nvim plugins
  - `dvm library list themes` - 34+ nvim themes
  - `dvm library list nvim packages` - Nvim bundles
  - `dvm library list terminal prompts` - 5 terminal prompts
  - `dvm library list terminal plugins` - 8 shell plugins
  - `dvm library list terminal packages` - Terminal bundles
  - `dvm library show <resource> <name>` - Detailed info
  - **Aliases**: `lib` → `library`, `ls` → `list`, `np` → `plugins`, `nt` → `themes`, `tp` → `terminal prompts`

- **Terminal Configuration** - Configure terminal per-workspace:
  - `dvm set terminal prompt -w <workspace> <name>` - Set prompt
  - `dvm set terminal plugin -w <workspace> <plugins...>` - Set plugins
  - `dvm set terminal package -w <workspace> <name>` - Set package bundle
  - Workspace-specific terminal configuration stored in database
  - Validation ensures resources exist in library

**Example Workflow:**
```bash
# Browse available themes
dvm lib ls themes

# Set terminal prompt for workspace
dvm set terminal prompt -w dev starship-minimal

# Set shell plugins
dvm set terminal plugin -w dev zsh-autosuggestions fzf

# Verify workspace config
dvm get workspace dev -o yaml
```

**Technical Details:**
- New database migration `004_add_terminal_fields`
- 104 new integration tests following TDD
- Workspace model extended with TerminalPrompt, TerminalPlugins, TerminalPackage fields

### v0.21.0 (2026-02-28)

**🚀 Local OCI Registry (Zot) - Container Image Caching**

> **Note**: This version has been superseded by v0.24.0 which introduces Registry Resources. See v0.24.0 for current registry usage.

**Legacy Features (replaced in v0.24.0):**

- Local OCI Registry with pull-through caching
- Registry CLI commands (now require positional arguments in v0.24.0)
- Build integration with `--no-cache`, `--push`, `--registry` flags

See [Full CHANGELOG](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) for complete v0.21.0 details.

### v0.20.1 (2026-02-28)

**🔗 GitRepo-Workspace Integration**

**New Features:**

- **Workspace Creation with GitRepo** - `--repo` flag for `dvm create workspace`
  - Associate workspaces with existing GitRepo resources
  - Automatically clones from local mirror to workspace's `repo/` directory
  - Each workspace gets independent clone for isolated development
  - Example: `dvm create workspace dev --repo my-project`
  
- **Auto-Sync Control** - `--no-sync` flag for `dvm attach`
  - Skip automatic mirror sync before attaching to workspace
  - Default: Syncs mirror if GitRepo has AutoSync=true
  - Use `--no-sync` for faster attach or offline work
  - Sync failures are warnings, not fatal errors
  - Example: `dvm attach --no-sync`

**Workflow:**
```bash
# 1. Create git mirror
dvm create gitrepo my-project --url https://github.com/myorg/my-project

# 2. Create workspace with repo
dvm create workspace dev --repo my-project

# 3. Attach (auto-syncs by default)
dvm attach

# Or skip sync for faster attach
dvm attach --no-sync
```

### v0.19.0 (2026-02-28)

**🚀 Full Workspace Isolation**

**🚀 Full Workspace Isolation**

**⚠️ BREAKING CHANGES - Fresh Database Required**

This is a major breaking release that requires a fresh database. **You must backup and delete your existing database before upgrading.**

- Fresh database schema required - existing databases incompatible
- Removed `projects` table - use Ecosystem → Domain → App hierarchy
- Removed credential `value` field - plaintext storage no longer supported
- Removed SSH key auto-mounting - use SSH agent forwarding instead

**New Features:**

- **Workspace Isolation** - Each workspace has dedicated isolated directories:
  - `~/.devopsmaestro/workspaces/{slug}/repo/` - Git repository clone
  - `~/.devopsmaestro/workspaces/{slug}/volume/` - Persistent data (nvim-data, cache)
  - `~/.devopsmaestro/workspaces/{slug}/.dvm/` - Generated configs
- **Workspace Slug** - Unique identifier format: `{ecosystem}-{domain}-{app}-{workspace}`
- **SSH Agent Forwarding** - Opt-in via `--ssh-agent` flag or `ssh_agent_forwarding: true` in YAML
- **Enhanced Security** - SSH keys never mounted, credentials limited to keychain/env only

**Migration Steps:**

1. Backup: `dvm get <resources> -o yaml > backup.yaml`
2. Delete database: `rm ~/.devopsmaestro/devopsmaestro.db`
3. Upgrade: `brew upgrade devopsmaestro`
4. Re-initialize: `dvm admin init`
5. Re-apply resources: `dvm apply -f backup.yaml`

See the [full migration guide](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md#migration-guide) for detailed instructions.

### v0.18.25 (2026-02-28)

**Fix Coolnight Theme Git Clone Errors**

- Fixed git clone failures for all 21 coolnight parametric themes by converting them to standalone mode
- Coolnight themes no longer require cloning external `rmkohlman/coolnight.nvim` repo (which doesn't exist)
- Standalone themes apply colors directly via `nvim_set_hl()` for better reliability
- **Breaking change**: Existing workspaces using coolnight themes need `dvm build` to regenerate config

### v0.18.24 (2026-02-23)

**Hierarchical Container Naming & Starship Theme Colors**

- Container names now include full hierarchy path: `dvm-{ecosystem}-{domain}-{app}-{workspace}` for better identification
- Added ecosystem/domain labels (`io.devopsmaestro.ecosystem`, `io.devopsmaestro.domain`) and environment variables to containers
- Starship prompts now automatically use active workspace theme colors via new ColorToPaletteAdapter
- Backward compatible fallback to `dvm-{app}-{workspace}` naming when hierarchy unavailable

### v0.18.23 (2026-02-23)

**Theme Database Persistence & Output Formatting**

- Fixed theme persistence bug where `dvm set theme --workspace X` wasn't saving to database
- Fixed theme output formatting that was showing raw struct instead of clean key-value format
- Theme values now properly appear in `dvm get workspace -o yaml` and all workspace queries

### v0.18.22 (2026-02-23)

**Shell Completion Enhancements**

- Added comprehensive tab completion for all resource commands (`dvm get ecosystem <TAB>`, etc.)
- All `--ecosystem`, `--domain`, `--app`, `--workspace` flags now complete resource names on TAB
- Fixed `dvm nvim sync <TAB>` and `dvm nvim push <TAB>` to complete workspace names correctly
- Provides kubectl-style CLI experience with real-time completion based on current database state

### v0.18.21 (2026-02-23)

**Theme Visibility Fix**

- Fixed bug where `dvm set theme` stored themes correctly but theme values were missing from YAML output
- Theme values now properly appear in `dvm get workspace -o yaml` and similar commands for all resource types
- Added dedicated theme column to workspaces table and updated all model ToYAML/FromYAML methods

### v0.18.20 (2026-02-23)

**Container Label Compatibility Fix**

- Fixed "name already in use" error when attaching to workspaces after upgrading from pre-v0.18.18 versions
- Containers created before the image tracking feature (v0.18.18) are now automatically detected and recreated with proper labels
- Provides seamless upgrade experience without requiring manual container cleanup

### v0.18.19 (2026-02-23)

**Mason Toolchain Auto-Installation**

- When Neovim is configured for a workspace, generated containers now automatically include npm, cargo, and pip toolchains
- Enables Mason to install language servers, formatters, and linters (stylua, prettier, pyright) without "executable not found" errors
- Supports both Alpine and Debian-based images with cross-platform package management

### v0.18.18 (2026-02-23)

**Containerd Runtime Image Change Detection**

- Fixed critical bug where containerd/Colima runtime was reusing running containers without checking if image had changed
- `dvm build --force --no-cache` now properly recreates containers with new images instead of reusing stale ones
- Added `io.devopsmaestro.image` label tracking to detect when underlying image changes
- Brings containerd runtime behavior in line with Docker runtime for consistent experience

### v0.18.17 (2026-02-21)

**Docker Build Context Fix**

- Fixed critical bug where `Dockerfile.dvm` was saved to the original app directory but Docker build used staging directory
- Docker COPY commands now work correctly with generated config files like `.config/starship.toml`
- Eliminates container configuration issues where generated files weren't found during build

### v0.18.16 (2026-02-21)

**Shell Configuration Fix**

- Fixed critical bug where shell configuration (starship.toml, .zshrc) was only generated when nvim was configured
- `dvm build` now always generates shell config regardless of nvim configuration, eliminating TOML parse errors in containers
- Refactored build flow to separate shell config generation from nvim config for better reliability

### v0.18.14 (2026-02-20)

**Plugin Storage Compatibility Fix**

- Fixed critical bug where `nvp package install` only saved to FileStore but `dvm build` reads from database
- `nvp package install` now saves plugins to BOTH FileStore and database for full compatibility
- Added plugin library fallback in `dvm build` when plugins not found in database
- Plugins installed via `nvp` are now immediately available to `dvm build`

### v0.18.6 (2026-02-20)

**Terminal Emulator Management (Phase 3 - Build Integration)**

- Added embedded emulator library with 6 curated configurations (rmkohlman, minimal, developer, alacritty-minimal, kitty-poweruser, iterm2-macos)
- New CLI commands: `dvt emulator install <name>`, `dvt emulator apply -f <file>`, `dvt emulator library list/show`
- Build integration: `dvm build` now generates WezTerm config and loads terminal plugins automatically
- Complete WezTerm configuration mapping from database (fonts, colors, keybindings, etc.)
- Terminal plugin loading in `.zshrc` with support for manual, oh-my-zsh, and zinit managers

### v0.18.5 (2026-02-20)

**Terminal Emulator Management (Phase 2)**

- Added complete terminal emulator database infrastructure with `terminal_emulators` table
- Support for wezterm, alacritty, kitty, and iterm2 emulator types with JSON configuration storage
- New `dvt emulator` CLI commands: `list`, `get`, `enable`, `disable` with filtering and formatting options
- Created `pkg/terminalops/emulator/` domain layer with proper interfaces and error handling
- Database adapter `DBEmulatorStore` following established DevOpsMaestro patterns
- Theme and workspace association support for coordinated terminal styling

### v0.18.4 (2026-02-20)

**Terminal Plugin Database Support**

- Added terminal plugin database support with `terminal_plugins` table for persistent storage
- Created database adapter for terminal plugins following nvimops pattern (`pkg/terminalops/store/db_adapter.go`)
- `dvt package install` now persists plugins to database for consistency with nvp
- `dvt plugin list` and `dvt plugin get` commands now read from database instead of file storage
- Enhanced cross-command integration - all terminal plugin operations use shared DevOpsMaestro database
- **Migration**: Plugin data moved from files to database; existing users may need to reinstall plugins from library

### v0.18.3 (2026-02-20)

**Build & Configuration Fixes**

- Fixed Starship TOML parsing error in containerized environments
- Fixed `dvm build` to respect user's default nvim package when workspace has no explicit plugins
- Fixed `dvm get defaults` to show actual user-configured values instead of hardcoded defaults
- Renamed nvp package `rkohlman-full` → `rmkohlman` for consistent naming

### v0.18.2 (2026-02-20)

**ARM64/Apple Silicon Stability**

- Fixed `dvm build` failing on ARM64 (Apple Silicon) with dpkg errors  
- Replaced gcc + python3-dev with build-essential for better ARM64 compatibility
- Added --fix-broken flag to apt-get commands for robust package installation
- Pinned Python base images to bookworm variant for reproducible builds
- Enhanced Docker cleanup process for better ARM64 system stability

### v0.18.1 (2026-02-20)

**Neovim Installation Fix**

- Fixed `dvm build` failures on slim Docker images (python:3.11-slim, etc.)
- Neovim now installed from GitHub releases instead of apt-get
- Works on all base images: Debian, Ubuntu, Alpine, and slim variants
- Multi-architecture support (amd64/arm64) with automatic detection

### v0.18.0 (2026-02-20)

**DVT Package Management**

- `dvt package list/get/install` commands for terminal package management
- Embedded terminal package library (core, developer, rmkohlman packages)
- Package inheritance support with automatic resolution
- Dry-run installation support for safe previews
- Parity with NvimPackage system for consistent package management

### v0.17.0 (2026-02-20)

**DVT Binary Release & TerminalPackages**

- DVT (TerminalOps) binary now included in releases alongside dvm and nvp
- New TerminalPackage resource type for terminal configuration bundles
- Terminal defaults management: `dvm use terminal package` and `dvm get terminal` commands  
- Auto-create NvimPackage after `nvp source sync` operations
- Mandatory test gate requirement for all releases (100% test success)

### v0.16.0 (2026-02-20)

**Package Management System**

- New NvimPackage resource type with kubectl-style CRUD operations
- Package defaults management for new workspaces
- External source sync (LazyVim integration)
- Auto-migration on startup for seamless upgrades

### v0.15.0 (2026-02-19)

**GitHub Directory & Secret Providers**

- Apply all YAML files from GitHub directories: `dvm apply -f github:user/repo/plugins/`
- Secret provider system with Keychain and environment support
- Inline secret syntax: `${secret:name}` in YAML resources
- DirectorySource interface for extensible directory-based sources

### v0.12.0 (2026-02-19)

**Hierarchical Theme System**

- Multi-level theme configuration: Workspace → App → Domain → Ecosystem → Global
- `dvm set theme` command with hierarchy flags
- 21 CoolNight theme variants with parametric generation
- kubectl-style theme IaC with file/URL/GitHub support

### v0.10.0 (2026-02-19)

**Plugin Packages & Enhanced Features**

- Plugin packages system with inheritance support
- Enhanced keymap generation for vim.keymap.set() calls
- Terminal theme integration with workspace context
- Smart workspace resolution with hierarchy flags

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| **0.32.1** | 2026-03-04 | Error handling fixes, watchdog refactor, improved GitRepo resolution |
| **0.32.0** | 2026-03-04 | `--repo` flag for app creation, Docker build hang fix on Colima |
| **0.31.0** | 2026-03-03 | Lazygit in containers, multi-architecture support |
| **0.30.4** | 2026-03-03 | Attach mount path fix for GitRepo-backed workspaces |
| **0.30.3** | 2026-03-03 | Database schema drift fix (git_repo_id column) |
| **0.30.2** | 2026-03-03 | Theme inheritance & terminal package prompt fixes, colorscheme generation |
| **0.30.1** | 2026-03-03 | Database schema completeness, YAML field parsing fixes |
| **0.24.0** | 2026-03-01 | Registry resource refactor, multi-registry support (BREAKING) |
| **0.22.0** | 2026-02-28 | Library browsing (dvm library), terminal configuration (dvm set terminal), integration |
| **0.21.0** | 2026-02-28 | Local OCI registry (Zot), pull-through cache, offline builds, registry CLI commands |
| **0.20.1** | 2026-02-28 | GitRepo-Workspace integration (--repo flag, --no-sync flag, auto-sync on attach) |
| **0.20.0** | 2026-02-28 | Git repository mirror management, bare mirrors, MirrorManager package, security validation |
| **0.19.0** | 2026-02-28 | Full workspace isolation, SSH agent forwarding, security hardening (BREAKING) |
| **0.18.17** | 2026-02-21 | Docker build context fix for generated config files |
| **0.17.0** | 2026-02-20 | DVT binary release, TerminalPackages, test gate requirement |
| **0.16.0** | 2026-02-20 | Package management system, auto-migration |
| **0.15.1** | 2026-02-19 | NvimPlugin opts field support fix |
| **0.15.0** | 2026-02-19 | GitHub directory support, secret providers |
| **0.14.0** | 2026-02-19 | TerminalPrompt resources, dvt prompt CLI |
| **0.13.1** | 2026-02-19 | `dvm get defaults` command |
| **0.13.0** | 2026-02-19 | Container build improvements, staging directory |
| **0.12.3** | 2026-02-19 | Comprehensive YAML reference documentation |
| **0.12.2** | 2026-02-19 | WezTerm CLI commands with theme integration |
| **0.12.1** | 2026-02-19 | Default nvim configuration for new workspaces |
| **0.12.0** | 2026-02-19 | Hierarchical theme system, 21 CoolNight variants |
| **0.11.0** | 2026-02-19 | Terminal theme integration, environment variables |
| **0.10.0** | 2026-02-19 | Plugin packages system, keymap generation fixes |
| **0.9.7** | 2026-02-18 | Colima SSH command fix |
| **0.9.6** | 2026-02-18 | Colima path lookup fix |
| **0.9.5** | 2026-02-18 | Container detached mode fix |
| **0.9.4** | 2026-02-18 | Colima containerd SSH operations fix |
| **0.9.3** | 2026-02-18 | Container attach consistency fix |
| **0.9.2** | 2026-02-18 | ColorProvider architecture, dynamic completions |
| **0.9.1** | 2026-02-17 | `dvm get workspaces -A` flag, Colima mount fix |
| **0.9.0** | 2026-02-17 | Smart workspace resolution with hierarchy flags |
| **0.8.0** | 2025-01-06 | New object hierarchy (Ecosystem/Domain/App/Workspace) |
| **0.7.1** | 2026-02-04 | Unified resource pipeline |
| **0.7.0** | 2026-02-03 | Terminal resize, timestamp tags |
| **0.6.0** | 2026-02-03 | Status command, aliases |
| **0.5.1** | 2026-02-02 | BuildKit socket validation fix |
| **0.5.0** | 2026-01-30 | Theme system |
| **0.4.1** | 2026-01-29 | URL support for nvp apply |
| **0.4.0** | 2026-01-29 | nvp standalone CLI |
| **0.3.3** | 2026-01-29 | Pre-generated shell completions |
| **0.3.1** | 2026-01-29 | Multi-platform support |
| **0.2.0** | 2026-01-24 | Theme system + YAML highlighting |
| **0.1.0** | 2026-01-23 | Initial release |

---

## Links

- [Full Changelog](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md)
- [GitHub Releases](https://github.com/rmkohlman/devopsmaestro/releases)
- [Release Process](./development/release-process.md)

---

## Semantic Versioning

DevOpsMaestro follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0) - Breaking changes
- **MINOR** (0.X.0) - New features (backward compatible)
- **PATCH** (0.0.X) - Bug fixes (backward compatible)
