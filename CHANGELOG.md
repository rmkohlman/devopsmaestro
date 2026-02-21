# Changelog

All notable changes to DevOpsMaestro (dvm) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.18.5] - 2026-02-20

### ➕ Added

#### Terminal Emulator Management (Phase 2)
- **Terminal emulator database support** - Added `terminal_emulators` table with proper indexes for storing emulator configurations
- **Multi-emulator type support** - Support for wezterm, alacritty, kitty, and iterm2 terminal emulator types
- **Emulator domain layer** - Created `pkg/terminalops/emulator/` package with types and store interface following established patterns
- **Database adapter for terminal emulators** - `DBEmulatorStore` adapter implementing EmulatorStore interface with proper JSON serialization
- **Terminal emulator CLI commands** - Added complete `dvt emulator` command suite:
  - `dvt emulator list` - List installed emulators with filtering by type and category
  - `dvt emulator get <name>` - Get detailed emulator configuration
  - `dvt emulator enable <name>` - Enable emulator for workspace use
  - `dvt emulator disable <name>` - Disable emulator
- **JSON config storage** - Store emulator-specific configuration as JSON for flexibility
- **Theme and workspace associations** - Link emulators to themes and workspaces for coordinated styling

### 🔧 Technical Improvements

#### Database Infrastructure
- **Migration 0005** - Added terminal_emulators table with proper foreign key constraints and indexes
- **Domain error mapping** - Proper error handling with domain-specific error types
- **Type-safe emulator types** - Enum-like constants for emulator type validation

#### Testing
- **Integration tests** - Database integration tests for emulator operations
- **Migration tests** - Verify database migration correctness
- **Mock implementations** - Updated mock store with emulator interface methods

---

## [0.18.4] - 2026-02-20

### ➕ Added

#### Terminal Plugin Database Support
- **Terminal plugin database support** - Added `terminal_plugins` table for persistent storage of terminal plugins
- **Database adapter for terminal plugins** - Created `pkg/terminalops/store/db_adapter.go` following the nvimops pattern
- **`dvt package install` now persists to database** - Installing a package stores plugins in the database (like nvp)

### 🔄 Changed

#### Plugin Storage Consistency
- **`dvt plugin list/get` now reads from database** - Plugin commands use the database as source of truth instead of file storage
- **Consistent data storage** - All dvt plugin operations now use the same database as dvm and nvp
- **Enhanced cross-command integration** - Installing packages via `dvt package install` now immediately updates plugin commands without requiring file system synchronization

### ⚠️ Migration

#### Plugin Storage Migration
- **Plugin storage location changed** - Plugin data moved from files (`~/.dvt/plugins/`) to database (`~/.dvt/dvt.db`)
- **Database initialization required** - Plugin commands now require database initialization via `dvt init`
- **Existing file-based plugins** - Users with existing file-based plugin installations may need to reinstall plugins from the library using `dvt plugin library install <name>` or `dvt package install <package>`

---

## [0.18.3] - 2026-02-20

### 🔧 Fixed

#### Build and Configuration Issues
- **Fixed Starship config TOML parsing error** - Changed `echo "[app]"` to `echo '[app]'` in generated starship config to fix TOML quoting issue that caused parsing errors inside containers
- **Fixed `dvm build` to use default nvim package** - Now respects the user's default nvim package (set via `dvm use nvim package <name>`) when a workspace has no explicit plugins configured
- **Fixed `dvm get defaults` display** - Now shows user-set defaults from database instead of hardcoded values for nvim-package, terminal-package, and theme

### 🔄 Changed

#### Package Naming Consistency
- **Renamed nvp package `rkohlman-full` → `rmkohlman`** - Updated package name for consistency with GitHub username and terminal package naming conventions

---

## [0.18.2] - 2026-02-20

### 🔧 Fixed

#### ARM64/Apple Silicon Stability
- **Fixed `dvm build` failing on ARM64 with dpkg errors** - Replaced gcc + python3-dev with build-essential for better ARM64 compatibility
- **Added `--fix-broken` flag to all apt-get install commands** - Prevents package conflicts on ARM64 systems
- **Enhanced Docker image cleanup** - Added `apt-get clean` before removing package cache
- **Pinned Python base images to `bookworm` variant** - Ensures reproducible builds across architectures (python:X.XX-slim-bookworm)
- **Improved package installation stability** - More robust dependency resolution for Apple Silicon M3 Max and other ARM64 systems

---

## [0.18.1] - 2026-02-20

### 🔧 Fixed

#### Neovim Installation
- **Fixed `dvm build` failures on slim Docker images** - Install Neovim directly from GitHub releases instead of apt-get
- **Works on all base images** - Supports Debian, Ubuntu, Alpine, and slim variants
- **Multi-architecture support** - Automatically detects and installs correct binary for amd64 and arm64
- **Enhanced Alpine Linux support** - Added proper Alpine package manager commands for Neovim dependencies

---

## [0.18.0] - 2026-02-20

### 🚀 Added

#### DVT Package Management - Terminal Package Library
- **`dvt package list`** - List all available terminal packages from embedded library
- **`dvt package get <name>`** - Show detailed package information with inheritance resolution
- **`dvt package install <name>`** - Install plugins/prompts/profiles from a package with `--dry-run` support
- **Embedded terminal package library** - Built-in packages: core, developer, rmkohlman
- **Package inheritance support** - Packages can extend other packages (e.g., developer extends core)
- **Parity with NvimPackage system** - Consistent package management across nvp and dvt tools

#### Terminal Package Library Structure
- **Core package** - Essential terminal tools and configurations
- **Developer package** - Extended development tools (extends core)
- **RMKohlman package** - Personal terminal environment (extends developer)
- **YAML-based definitions** - Standard package format with metadata and specifications
- **Automatic inheritance resolution** - Packages automatically include inherited content

### 🔧 Enhanced

#### Command Consistency
- **DVT package commands** now follow the same patterns as `nvp package` commands
- **Unified package management** across both Neovim and terminal environments
- **Consistent flag support** including `--dry-run` for safe preview operations

---

## [0.17.0] - 2026-02-20

### 🚀 Added

#### DVT (TerminalOps) Binary Release
- **DVT binary now included in releases** - All three binaries (dvm, nvp, dvt) are now built and released together
- **TerminalOps Homebrew formula** - `brew install rmkohlman/tap/terminalops` now available
- **Enhanced release workflow** - CI/CD builds and releases all three binaries with proper versioning

#### TerminalPackage Resource Type
- **New TerminalPackage resource** - Group terminal configuration into packages like NvimPackages
- **YAML format support** - Define terminal packages with `kind: TerminalPackage` in YAML
- **Database integration** - TerminalPackages stored in database with proper migrations
- **Supports plugins, prompts, profiles, wezterm configs** - Complete terminal environment bundling

#### Terminal Defaults Management
- **`dvm use terminal package <name>`** - Set default terminal package for new workspaces
- **`dvm use terminal package none`** - Clear default terminal package
- **Database persistence** - Terminal package defaults stored and retrieved consistently
- **Parallel to NvimPackage defaults** - Same pattern as `dvm use nvim package`

#### Terminal Resource Commands
- **`dvm get terminal packages`** - List all available terminal packages
- **`dvm get terminal package <name>`** - Show specific terminal package details
- **`dvm get terminal defaults`** - Show current terminal defaults alongside other defaults
- **Multiple output formats** - Supports `-o json`, `-o yaml`, `-o table` for all terminal commands

#### NvimOps Auto-Package Creation
- **Auto-create packages after sync** - `nvp source sync lazyvim` now automatically creates "lazyvim" package
- **Package contains all synced plugins** - All plugins from the source sync are bundled into the package
- **Metadata labels** - Auto-generated packages include source, auto-generated, and sync-time labels
- **Seamless workflow** - Sync from source, get package automatically for easy reuse

#### Test Gate Requirement
- **Mandatory 100% test success** - All tests must pass before any release-related documentation updates
- **Updated agent policies** - Test, release, and document agents now enforce the test gate
- **Quality assurance** - Ensures releases only happen with verified, working code

### 🔧 Enhanced

#### Release Process
- **Multi-binary support** - Release workflow now handles dvm, nvp, and dvt binaries
- **Version consistency** - All binaries use the same version tag for releases
- **Improved CI verification** - DVT build is now verified in CI workflow alongside dvm and nvp

---

## [0.16.0] - 2026-02-20

### 🚀 Added

#### Package Management (NvimPackage) - kubectl-style CRUD Operations
- **NvimPackage resource type** - Collections of plugin names with kubectl-style operations
- **`dvm get nvim packages`** - List all available packages
- **`dvm get nvim package <name>`** - Show specific package details
- **`dvm apply -f package.yaml`** - Apply package from YAML file
- **`dvm edit nvim package <name>`** - Edit package in default editor
- **`dvm delete nvim package <name>`** - Remove a package
- **Package YAML format** with apiVersion/kind/metadata/spec structure:
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
    extends: core  # optional package inheritance
  ```

#### Defaults Management - Set Default Packages for New Workspaces
- **`dvm use nvim package <name>`** - Set default nvim package for new workspaces
- **`dvm use nvim package none`** - Clear default package setting
- **`dvm get nvim defaults`** - Show current default package and other defaults
- **Validates packages** exist in database or library before setting as default
- **Helpful error messages** with hints when packages don't exist

#### External Source Sync - Import Plugins from External Sources
- **`nvp source list`** - List all available external sources (LazyVim, etc.)
- **`nvp source show <name>`** - Show detailed source information
- **`nvp source sync <name>`** - Sync plugins from external source to local store
- **`nvp source sync <name> --dry-run`** - Preview sync without making changes
- **`nvp source sync <name> -l category=lang`** - Filter plugins by labels during sync
- **`nvp source sync <name> --tag v15.0.0`** - Sync from specific version/tag
- **LazyVim integration** - Built-in support for syncing plugins from LazyVim configs

#### Auto-Migration on Startup
- **Database migrations run automatically** on dvm startup
- **No manual migration required** after upgrades
- **Seamless version upgrades** without user intervention
- **Backward compatibility** maintained for all existing data

### 🔧 Enhanced

#### Package System
- **Package inheritance** via `extends` field in YAML spec
- **Plugin resolution** automatically includes inherited plugins
- **Category and labeling** support for better package organization
- **Validation** ensures all referenced plugins exist

#### Error Handling
- **Better error messages** for package operations with actionable hints
- **Validation feedback** when setting defaults with non-existent packages
- **Clear success messages** for package and defaults operations

### 🧪 Testing

- **Package CRUD operations** - All kubectl-style operations tested
- **Defaults management** - Setting and clearing defaults verified
- **Auto-migration** - Database migration automation tested
- **External source sync** - Plugin syncing from external sources validated

---

## [0.15.1] - 2026-02-19

### Fixed

- **NvimPlugin Opts Field**: Support both map format (`{key: value}`) and raw Lua string format for `opts` field in NvimPlugin YAML
  - Previously `opts` was typed as `map[string]interface{}`, causing YAML parsing errors with raw Lua code
  - Now accepts both formats: `opts: {key: value}` and `opts: | raw lua code`
  - Fixed issue where 19 of 34 plugins from `github:rmkohlman/dvm-config/plugins/` failed to apply
  - Updated types in `pkg/nvimops/plugin/types.go`, `models/nvim_plugin.go`, and related files

---

## [0.15.0] - 2026-02-19

### Added

- **GitHub Directory URL Support**: Apply all YAML files from a GitHub directory with a single command
  - `dvm apply -f github:user/repo/plugins/` applies all .yaml files in directory
  - Supports trailing slash or no extension to indicate directory
  - Shows progress: "Applying 1/34: telescope.yaml..."
  - Continues on individual failures and reports summary
  
- **Secret Provider System**: Pluggable secret resolution for YAML resources
  - New `pkg/secrets/` package with interface-based design
  - **Keychain Provider** (macOS): Reads secrets from macOS Keychain
  - **Environment Provider**: Reads from `DVM_SECRET_<NAME>` or fallback env vars
  - Inline syntax: `${secret:name}` or `${secret:name:provider}`
  - GITHUB_TOKEN now resolved via secret providers (Keychain first, then env)
  
- **DirectorySource Interface**: Extensible interface for directory-based sources
  - `DirectorySource` interface in `pkg/source/`
  - `GitHubDirectorySource` implementation
  - Foundation for future local directory and other VCS support

### Changed

- `dvm apply` help text updated with directory and secret examples
- `dvm apply nvim plugin` and `dvm apply nvim theme` help updated
- `GitHubDirectorySource.ListFiles()` now returns `[]Source` for proper abstraction

### Fixed

- Architecture compliance: Proper Interface → Implementation → Factory pattern for sources

---

## [0.13.1] - 2026-02-19

### ✨ Features
- **`dvm get defaults`**: New command to display all default configuration values
  - Shows theme, shell, nvim, and container defaults
  - Supports `-o yaml` and `-o json` output formats

---

## [0.13.0] - 2026-02-19

### 🐳 Container Build
- **Staging directory**: Build artifacts now use `~/.devopsmaestro/build-staging/` instead of polluting app source directory
- **Shell configuration**: Generate proper `.zshrc` to prevent Zsh new-user wizard
- **Starship prompt**: Install starship and generate `starship.toml` with app name in prompt
- **Plugin errors**: Remove error masking (`|| true`) from nvim plugin installation
- **ARM64 support**: Fixed starship installation for ARM64 architecture

---

## [0.12.3] - 2026-02-19

### 📖 Documentation
- **Comprehensive YAML Reference Documentation** - Added 9 detailed reference pages covering all resource types:
  - Complete workspace.yaml specification with all fields and examples
  - App.yaml configuration guide with language-specific templates
  - Domain, ecosystem, and project YAML schemas
  - Plugin and theme YAML reference documentation
  - Package configuration and inheritance examples
  - Comprehensive field descriptions and validation rules

---

## [0.12.2] - 2026-02-19

### 🚀 Added

#### WezTerm CLI Commands with Theme Integration
- **`dvt wezterm list`** - List available WezTerm presets
- **`dvt wezterm show <name>`** - Show preset details  
- **`dvt wezterm generate <name>`** - Generate wezterm.lua with theme colors
- **`dvt wezterm apply <name>`** - Apply configuration to ~/.wezterm.lua
- **Automatic theme color resolution** - Theme colors from library embedded into generated configurations

### 📖 Documentation
- **Complete setup workflow** - Added to quickstart guide for seamless onboarding
- **Updated WezTerm documentation** - New CLI commands and usage examples

### 🧹 Fixed  
- **Removed temporary test files** - Cleaned up root directory

---

## [0.12.1] - 2026-02-19

### 🚀 Added

#### Default Nvim Configuration for New Workspaces
- **Automatic workspace setup** - New workspaces now get lazyvim structure with core plugin package by default
- **Core plugin package** - 6 essential plugins: treesitter, telescope, which-key, lspconfig, nvim-cmp, gitsigns
- **DefaultNvimConfig() function** - Programmatic API in `pkg/nvimops/defaults.go` for consistent configuration
- **Seamless integration** - `dvm create workspace` automatically applies defaults during workspace creation

### 📖 Documentation
- **Updated quickstart guide** - Reflects new default behavior for workspace creation
- **Enhanced workspace documentation** - Details about automatic nvim configuration
- **NvimOps documentation updates** - Core package and default configuration coverage

---

## [0.12.0] - 2026-02-19

### 🚀 Added

#### Hierarchical Theme System
- **Multi-level theme configuration** - Themes now cascade through hierarchy: Workspace → App → Domain → Ecosystem → Global Default
- **`dvm set theme` command** - Set themes at any level with `--workspace`, `--app`, `--domain`, or `--ecosystem` flags
- **Theme resolver with Strategy pattern** - New `pkg/colors/resolver/` package handles theme inheritance and cascading
- **Database migration 011** - Adds theme columns to ecosystems, domains, and apps tables for hierarchical theme storage

#### 21 CoolNight Theme Variants
- **Parametric theme generator** - Algorithmic theme generation with consistent color harmony
- **Blue variants**: ocean (default), arctic, midnight
- **Purple variants**: synthwave, violet, grape  
- **Green variants**: matrix, forest, mint
- **Warm variants**: sunset, ember, gold
- **Red/Pink variants**: rose, crimson, sakura
- **Monochrome variants**: charcoal, slate, warm
- **Special variants**: nord, dracula, solarized

#### kubectl-Style Theme IaC
- **Apply themes from files** - `dvm apply -f theme.yaml` for declarative theme management
- **Apply from URLs** - `dvm apply -f https://example.com/theme.yaml` for remote theme sharing
- **Apply from GitHub** - `dvm apply -f github:user/repo/theme.yaml` for version-controlled themes
- **Export themes** - `dvm get nvim theme <name> -o yaml` for sharing and backup

#### Library Fallback System
- **Embedded theme library** - NvimThemeHandler now falls back to 34+ embedded library themes
- **User theme priority** - User-defined themes override library themes with the same name
- **Instant availability** - Library themes available immediately without manual installation

#### WezTerm Configuration Support
- **Complete wezterm.lua generator** - Full configuration file generation for WezTerm terminal
- **Type-safe configuration** - Structured types for all WezTerm settings
- **Library presets** - Pre-built configurations (minimal, tmux-style, default)

#### rmkohlman Plugin Package
- **Complete Neovim setup** - 39-plugin configuration with LSP, treesitter, telescope, and more
- **Production-ready** - Fully configured development environment
- **Modular architecture** - Clean plugin organization and configuration

#### Enhanced Documentation
- **Comprehensive theme documentation** - New dedicated pages for advanced theme features
- **CoolNight Collection guide** - Complete documentation for all 21 CoolNight variants with usage recommendations
- **Plugin Packages documentation** - Detailed guide for using and creating plugin packages
- **WezTerm Integration guide** - Step-by-step terminal configuration documentation
- **Theme Infrastructure as Code guide** - Complete IaC workflow documentation with team sharing examples
- **Hierarchical Theme System guide** - In-depth explanation of theme cascade and resolution

### 🔧 Enhanced

#### Theme Management
- **Improved theme resolution** - Smarter inheritance with proper fallback chains
- **Better error handling** - Clear messages for missing themes and invalid configurations
- **Performance optimizations** - Cached theme resolution for faster startup

### 📦 Technical Changes

#### New Files
- `cmd/set_theme.go` - Theme setting command implementation
- `migrations/sqlite/011_*` - Database migration for hierarchical themes
- `pkg/colors/resolver/` - Theme resolution engine
- `pkg/nvimops/theme/library/themes/coolnight-*.yaml` - 21 CoolNight variants
- `pkg/terminalops/wezterm/` - WezTerm configuration support
- `pkg/nvimops/package/library/packages/rmkohlman.yaml` - Complete plugin package

#### Modified Files
- Enhanced theme handlers with library fallback
- Updated CLI commands for hierarchical theme support
- Improved database models with theme fields
- Enhanced documentation for new features

---

## [0.11.0] - 2026-02-19

### 🚀 Added

#### Terminal Theme Integration
- **Terminal theme integration** - Theme colors are now passed to container shell sessions for consistent visual experience
- **`AttachOptions` struct** - Enhanced container attach with environment variables, shell configuration, and login shell support
- **`Theme` field in `NvimConfig`** - Workspace-level theme configuration support for future workspace-specific themes
- **Terminal color environment variable generator** - New `pkg/nvimops/theme/terminal.go` converts theme colors to shell environment variables
- **Enhanced environment variables on attach** - Containers now receive comprehensive context:
  - `TERM=xterm-256color` - Proper terminal capabilities for backspace, autocomplete, and colors
  - `DVM_WORKSPACE` - Current workspace name for shell prompt integration
  - `DVM_APP` - Current app name for context awareness  
  - `DVM_THEME` - Active theme name for terminal theme matching
  - `DVM_COLOR_*` - Complete theme color palette as environment variables

### 🐛 Fixed

#### Terminal Experience
- **Terminal issues in `dvm attach`** - Fixed backspace and autocomplete functionality by setting `TERM=xterm-256color`
- **Container shell environment** - Shell sessions now have proper terminal capabilities and workspace context

### 🔧 Enhanced

#### Container Attachment
- **All container runtimes updated** - Docker, Containerd, and Colima now support `AttachOptions` with environment variables
- **Shell configuration** - Support for custom shell (`/bin/zsh` default) and login shell mode
- **Theme color mapping** - Automatic mapping from DVM config theme names to library theme names

### 📦 Files Changed

```
models/workspace.go                          # Added Theme field to NvimConfig
pkg/nvimops/theme/terminal.go               # NEW: Terminal env var generator  
pkg/nvimops/theme/terminal_test.go          # NEW: Terminal integration tests
cmd/attach.go                               # Enhanced with theme loading and env vars
operators/runtime_interface.go             # Added AttachOptions struct
operators/containerd_runtime.go            # Updated AttachToWorkspace signature
operators/containerd_runtime_v2.go         # Updated AttachToWorkspace with env support
operators/docker_runtime.go                # Updated AttachToWorkspace with env support
operators/mock_runtime.go                  # Updated for AttachOptions compatibility
operators/mock_runtime_test.go             # Updated tests for new interface
```

---

## [0.10.0] - 2026-02-19

### 🚀 Added

#### Plugin Packages System
- **Plugin Packages** - Group plugins into reusable packages with inheritance support
  - `nvp package list` - List all available packages
  - `nvp package get <name>` - Show package details with resolved plugins
  - `nvp package install <name>` - Install all plugins from a package
  - **Default packages**: `core`, `go-dev`, `python-dev`, `full`
  - **Package inheritance** - Packages can extend other packages (e.g., `go-dev` extends `core`)
  - **Workspace integration** - Support via `pluginPackage` and `mergeMode` fields in workspace YAML

#### Package Library System
- **Embedded package definitions** - Pre-built packages ready to use
- **Inheritance resolution** - Automatically resolves all plugins from package hierarchy
- **Category and tag support** - Organize packages by development context
- **YAML-based definitions** - Following kubectl pattern with apiVersion/kind/metadata/spec

#### Workspace Configuration Enhancement
- **`pluginPackage` field** - Reference a plugin package by name in workspace config
- **`mergeMode` field** - Control how package plugins merge with individual plugins ("append" or "replace")

### 🐛 Fixed

#### Keymap Generation
- **Fixed keymap generation** - The `keymaps` field now properly generates `vim.keymap.set()` calls
- **Combined config support** - Keymaps are now appended to existing plugin config functions
- **Multi-mode support** - Properly handles single mode strings and multi-mode arrays

### 📦 Package Structure

```
pkg/nvimops/package/
├── types.go               # Package, PackageYAML types
├── parser.go              # YAML parsing and validation
├── package_test.go        # Package type tests
└── library/
    ├── library.go         # Embedded package library
    ├── library_test.go    # Library tests
    └── packages/          # Default package definitions
        ├── core.yaml      # Essential plugins for any development
        ├── go-dev.yaml    # Go development essentials (extends core)
        ├── python-dev.yaml # Python development essentials (extends core)
        └── full.yaml      # Complete plugin collection (extends core)

cmd/nvp/
└── package.go             # Package CLI commands
```

### 🧪 Testing

- **Comprehensive package tests** - Full test coverage for package parsing, library operations, and inheritance
- **Generator tests** - Verified keymap generation produces correct vim.keymap.set() calls
- **Integration tests** - Package installation and plugin resolution workflows

---

## [0.9.7] - 2026-02-18

### 🐛 Fixed

- **Colima SSH command** - Removed invalid `-t` flag from `colima ssh` command in `attachViaColima` function. The `colima ssh` command doesn't support TTY allocation flags like regular SSH - TTY is automatically allocated.

---

## [0.9.6] - 2026-02-18

### 🐛 Fixed

- **Colima path lookup** - Fixed hardcoded `/usr/bin/colima` path to use PATH lookup for better system compatibility. Colima can now be found automatically regardless of installation method.

### 🔧 Improved

- **Container runtime standardization** - Added helper functions for container naming and command defaults. Standardized container runtime implementations with consistent helper methods for improved maintainability.

---

## [0.9.5] - 2026-02-18

### 🐛 Fixed

- **Container detached mode** - Fixed containers exiting immediately after start by using `/bin/sleep infinity` instead of `/bin/zsh -l` as the default command in detached mode. This ensures containers remain running when not attached to a terminal.

---

## [0.9.4] - 2026-02-18

### 🐛 Fixed

- **Colima containerd SSH operations** - Fixed container attach, stop, and status operations in Colima by using `nerdctl` via SSH instead of direct containerd client calls. This resolves connection issues when using Colima as the container runtime.
- **Container runtime platform detection** - Fixed platform detection mismatch by passing the detected platform to the containerd runtime. This prevents architecture conflicts when creating containers on different platforms.

---

## [0.9.3] - 2026-02-18

### 🐛 Fixed

- **Container attach consistency** - Fixed "container not found" error in `dvm attach` command. Containers were being created with `WorkspaceName` but attach was looking for `ContainerName`. Now uses `ContainerName` consistently across all runtime implementations for reliable workspace attachment.

---

## [0.9.1] - 2026-02-17

### 🚀 Added

- **`dvm get workspaces -A`** - New flag to list ALL workspaces across all apps/domains/ecosystems
- **`-A` shorthand** - Added `-A` shorthand to `get apps --all` and `get domains --all` for consistency

### 🐛 Fixed

- **Colima containerd mount error** - Fixed "failed to mount ... not implemented" error when using Colima with containerd runtime on macOS. Container creation now uses `nerdctl` via SSH which properly handles host path mounting through Colima's mount system.

---

## [0.9.0] - 2026-02-17

### 🚀 Added

#### Smart Workspace Resolution
- **Hierarchy flags** - All workspace commands now support `-e`, `-d`, `-a`, `-w` flags for smart resolution
- **No more sequential `dvm use` commands** - Specify criteria directly on the command line
- **Automatic disambiguation** - When multiple workspaces match, shows full paths to help you choose
- **Context auto-update** - Resolved workspace automatically becomes the active context

#### New Flags for Commands
- **`dvm attach`** - Added `-e/--ecosystem`, `-d/--domain`, `-a/--app`, `-w/--workspace`
- **`dvm build`** - Added `-e/--ecosystem`, `-d/--domain`, `-a/--app`, `-w/--workspace`
- **`dvm detach`** - Added `-e/--ecosystem`, `-d/--domain`, `-a/--app`, `-w/--workspace`
- **`dvm get workspaces`** - Added hierarchy flags for filtering
- **`dvm get workspace`** - Added hierarchy flags, workspace name now optional with flags

#### Resolver Package
- **`pkg/resolver/`** - New package for workspace resolution logic
  - `WorkspaceResolver` interface and implementation
  - `AmbiguousError` with `FormatDisambiguation()` for helpful output
  - `ErrNoWorkspaceFound` for clear error handling
- **`FindWorkspaces()` DataStore method** - Query workspaces across full hierarchy with JOINs

### 🔧 Changed

- **`dvm detach --all`** - Changed shorthand from `-a` to `-A` (frees `-a` for `--app`)
- **`dvm get workspace`** - Workspace name argument now optional when using flags

### 📖 Examples

```bash
# Before (verbose - required multiple commands)
dvm use ecosystem healthcare
dvm use domain billing  
dvm use app portal
dvm use workspace staging
dvm attach

# After (smart resolution - single command)
dvm attach -a portal                 # Find workspace in 'portal' app
dvm attach -e healthcare -a portal   # Specify ecosystem and app
dvm build -a portal -w staging       # Build specific workspace
dvm detach -A                        # Stop ALL workspaces (note: -A not -a)
dvm get workspaces -e healthcare     # List all workspaces in ecosystem
```

---

## [0.8.0] - 2025-01-06

### 🚀 Added

#### New Object Hierarchy
- **Ecosystem** - Top-level platform grouping (e.g., "acme-corp")
- **Domain** - Bounded context within an ecosystem (replaces "Project")
- **App** - The codebase/application within a domain
- **Workspace** - Development environment for an App

#### App Model Enhancements
- **`spec.language`** - Primary language configuration (name, version)
- **`spec.build`** - Build configuration (dockerfile, buildpack, args, target, context)
- **`spec.dependencies`** - Dependency management (file, install command, extras)
- **`spec.services`** - Service dependencies (postgres, redis, etc. with version, port, env)
- **`spec.ports`** - Port exposure for the application
- **JSON storage** - Language and build config stored as JSON in database

#### Workspace Model Enhancements
- **`spec.terminal`** - Terminal multiplexer config (tmux, zellij, screen)
- **`spec.build.devStage.devTools`** - Developer tools (gopls, delve, pylsp, etc.)
- **Cleaner separation** - Workspace now focuses purely on dev environment

#### New Commands
- **`dvm create ecosystem`** - Create a new ecosystem
- **`dvm create domain`** - Create a domain within an ecosystem
- **`dvm create app`** - Create an app within a domain
- **`dvm get ecosystems`** - List all ecosystems
- **`dvm get domains`** - List domains in current ecosystem
- **`dvm get apps`** - List apps in current domain
- **`dvm use ecosystem`** - Set active ecosystem
- **`dvm use domain`** - Set active domain
- **`dvm use app`** - Set active app

### 🔧 Changed

#### Model Separation (App vs Workspace)
- **App owns codebase concerns**: language, build config, services, ports, dependencies
- **Workspace owns dev environment**: nvim, shell, terminal, dev tools, mounts
- **Renamed `LanguageTools` to `DevTools`** in workspace spec (clearer intent)
- **Removed `Languages` from Workspace** - moved to App's language config
- **Removed `Ports` from Workspace container** - App owns port exposure
- **Renamed `BuildConfig` to `DevBuildConfig`** in workspace (dev-specific)

#### Terminology Migration
- **Project → Domain** - "Project" was overloaded, "Domain" is clearer (DDD concept)
- **Backward compatibility** - Old "project" commands still work with deprecation warnings

### 📚 Documentation

- **Updated YAML schema documentation** - Complete rewrite showing App/Workspace separation
- **Clear separation guide** - Table showing which concerns belong where
- **Language-specific examples** - Python, Go, Node.js App + Workspace pairs
- **Updated quickstart guide** - Full hierarchy workflow (ecosystem → domain → app → workspace)
- **Updated command reference** - All new hierarchy commands documented

### 🧪 Testing

- **All tests passing** with race detector
- **JSON marshal/unmarshal** implemented for App language and build config

---

## [0.7.2] - 2025-01-05

### 🐛 Fixed
- Minor bug fixes and stability improvements

---

## [0.7.1] - 2026-02-04

### 🚀 Added

#### Unified Resource Pipeline
- **`pkg/resource/` package** - Unified resource interface following kubectl patterns
  - `Resource` interface - Common interface for all resource types (NvimPlugin, NvimTheme, etc.)
  - `Handler` interface - CRUD operations per resource kind (Apply, Get, List, Delete, ToYAML)
  - `Context` struct - Carries DataStore, PluginStore, ThemeStore, ConfigDir
  - Registry pattern - Handlers registered at startup, looked up by Kind
- **`pkg/source/` package** - Source resolution for kubectl-style `-f` flag
  - `FileSource` - Read from local files
  - `URLSource` - Fetch from HTTP/HTTPS URLs
  - `StdinSource` - Read from stdin (`-f -`)
  - `GitHubSource` - GitHub shorthand (`github:user/repo/path.yaml`)
  - Automatic source type detection from path/URL

#### Consistent Command Architecture
- **`dvm apply`** - Refactored to use unified resource pipeline
- **`dvm get nvim plugins/themes`** - Now uses `resource.List()` and `resource.Get()`
- **`dvm delete nvim plugin`** - Now uses `resource.Delete()`
- **`nvp apply`** - Refactored to use unified source/resource pipeline

### 🔧 Changed

#### Architecture Improvements
- **Separation of concerns** - "How to get data" (Source) vs "What to do with data" (Handler)
- **Extensible design** - Add new resource types by implementing Handler interface
- **Testable** - All handlers work with mock stores for unit testing
- **Consistent patterns** - All nvim resource operations go through unified interface

### 📦 New Packages

```
pkg/source/
├── source.go          # Source interface, Resolve(), DetectSourceType()
└── source_test.go     # Comprehensive tests

pkg/resource/
├── resource.go        # Resource, Handler, Context interfaces
├── registry.go        # Register(), Get(), List(), Delete(), Apply()
├── resource_test.go   # Interface tests
└── handlers/
    ├── nvim_plugin.go # NvimPluginHandler, NvimPluginResource
    └── nvim_theme.go  # NvimThemeHandler, NvimThemeResource
```

---

## [0.7.0] - 2026-02-03

### 🚀 Added

#### Terminal Resize Support
- **Full terminal window on attach** - Container now uses the full terminal size
- **Dynamic resize handling** - Terminal automatically adjusts when you resize your window
- **SIGWINCH signal monitoring** - Proper signal handling for resize events

#### Timestamp-Based Image Versioning
- **Timestamp tags** - Images now tagged with `YYYYMMDD-HHMMSS` instead of `:latest`
- **Auto-recreate on image change** - `dvm attach` detects when image has changed and recreates container
- **Build history** - Each build creates a uniquely tagged image for rollback capability

#### kubectl-Style Workspace Plugin Commands
- **`dvm apply nvim plugin -f file.yaml`** - Apply plugin from YAML file
- **`dvm get nvim plugins`** - List all nvim plugins from database
- **`dvm get nvim plugin <name>`** - Get specific plugin details
- **`dvm delete nvim plugin <name>`** - Delete a plugin
- **Database as source of truth** - Plugins stored in SQLite, generated to Lua at build time

#### Nvim Plugin Library
- **16+ curated plugins** - Complete library matching nvim-config repo
- **Improved plugin configs** - Better treesitter, telescope, and LSP configurations
- **Array-of-maps rendering** - Fixed complex dependency rendering in generator

#### Terminal Operations (dvt)
- **New `dvt` CLI** - Terminal configuration management tool
- **Terminalops package** - Decoupled interfaces for terminal management
- **Shell, prompt, and plugin management** - Modular terminal customization

#### Theme System Enhancements
- **Theme preview command** - Preview themes before installing
- **5 new themes** - Additional color schemes
- **Database storage for themes** - Persistent theme configuration
- **Shared palette package** - Unified color management across nvp/dvm

### 🐛 Fixed

- **Leader key not working** - Set `vim.g.mapleader` in init.lua before lazy.nvim loads
- **Nvim config not in container** - Generate nvim config before Dockerfile so COPY instruction is included
- **Complex plugin dependencies** - Dependencies with config/build options now stored correctly
- **Platform detection** - Improved platform detection in status, attach, and detach commands

### 🔧 Changed

- **ContainerRuntime interface** - Commands now use decoupled runtime interface
- **Image naming** - Changed from `:latest` to timestamp-based tags
- **Workspace creation** - New workspaces get `:pending` tag until first build

---

## [0.6.0] - 2026-02-03

### 🚀 Added

#### `dvm status` Command
- **New status command** - Shows current context, runtime info, and running containers
- **Context display** - Active project and workspace at a glance
- **Runtime info** - Colima/nerdctl profile, status, container runtime
- **Running containers** - List DVM workspace containers with status
- **Output formats** - Supports `-o json` and `-o yaml` for scripting

```bash
dvm status           # Human-readable status
dvm status -o json   # JSON output for scripts
dvm status -o yaml   # YAML output
```

#### kubectl-style Resource Aliases
- **Short aliases** for common resources - faster commands!
- **Consistent across commands** - Works with `get`, `create`, `delete`, `use`

| Resource | Alias | Example |
|----------|-------|---------|
| projects | `proj` | `dvm get proj` |
| workspaces | `ws` | `dvm get ws` |
| context | `ctx` | `dvm get ctx` |
| platforms | `plat` | `dvm get plat` |

```bash
dvm get proj          # Same as 'dvm get projects'
dvm get ws            # Same as 'dvm get workspaces'
dvm use ws main       # Same as 'dvm use workspace main'
dvm create proj api   # Same as 'dvm create project api'
dvm delete ws dev     # Same as 'dvm delete workspace dev'
```

#### `dvm detach` Command
- **Stop workspace containers** - Cleanly stop running workspace containers
- **Context-aware** - Uses current workspace if none specified

```bash
dvm detach            # Stop current workspace container
dvm detach myworkspace # Stop specific workspace
```

#### `dvm get context` Command
- **View current context** - Show active project and workspace
- **Multiple formats** - Table, JSON, YAML output

```bash
dvm get context       # or: dvm get ctx
dvm get ctx -o yaml
```

#### Context Clear Commands
- **`--clear` flag** - Clear current project or workspace context
- **`none` argument** - Alternative way to clear context

```bash
dvm use project --clear    # Clear active project
dvm use workspace none     # Clear active workspace
```

#### Delete Commands
- **`dvm delete project`** - Delete a project and optionally its workspaces
- **`dvm delete workspace`** - Delete a workspace
- **`-p` flag** - Specify project for workspace commands

```bash
dvm delete project myproj
dvm delete workspace dev -p myproj
```

### 🔧 Changed

#### Render Package Migration
- **Decoupled CLI output** - All commands now use the `render/` package
- **Consistent formatting** - Unified output across all commands
- **Better separation** - Commands prepare data, renderers display it
- **Functions**: `render.Success()`, `render.Warning()`, `render.Info()`, `render.Error()`

#### CI/CD with GitHub Actions
- **Automated testing** - Tests run on push/PR to main
- **Race detection** - All tests run with `-race` flag
- **Build verification** - Both `dvm` and `nvp` binaries built and verified

### 📚 Documentation

- **ARCHITECTURE.md** - Decoupled pattern diagrams and code review checklist
- **Streamlined docs** - Cleaner CLAUDE.md overview

### 🧪 Testing

- **Alias tests** - `cmd/aliases_test.go` with comprehensive coverage
- **All tests passing** with race detector

---

## [0.5.1] - 2026-02-02

### 🐛 Fixed
- **BuildKitBuilder socket validation** - Validate containerd/buildkit sockets exist before attempting connection (fixes flaky behavior due to lazy connection)

### 📚 Documentation
- Updated README with two-tool structure and theme system documentation
- Added nvp installation instructions to INSTALL.md
- Added Part 9 (Theme Operations) to NVIMOPS_TEST_PLAN.md
- Updated Homebrew docs with current tap status (devopsmaestro + nvimops formulas)
- Added nvp shell completions to SHELL_COMPLETION.md
- Documented GoReleaser automation in release-process.md
- Updated CLAUDE.md with nvp architecture details

---

## [0.5.0] - 2026-01-30

### 🚀 Added

#### NvimTheme System - YAML-based Colorscheme Management
- **New `NvimTheme` resource type** - Define colorschemes in YAML with full palette control
- **Exported color palette** - Other plugins can `require("theme").palette` for consistent styling
- **Theme library** - 8 pre-defined themes ready to install:
  - `tokyonight-custom` - Custom deep blue variant (from rmkohlman/nvim-config)
  - `tokyonight-night` - Standard Tokyo Night
  - `catppuccin-mocha` - Catppuccin dark pastel
  - `catppuccin-latte` - Catppuccin light pastel
  - `gruvbox-dark` - Retro groove colors
  - `nord` - Arctic north-bluish theme
  - `rose-pine` - Natural pine with soho vibes
  - `kanagawa` - Inspired by the famous painting

#### Theme CLI Commands
- `nvp theme library list` - Browse available themes
- `nvp theme library show <name>` - View theme details
- `nvp theme library install <name>` - Install from library
- `nvp theme apply -f file.yaml` - Apply custom theme from file
- `nvp theme apply --url github:user/repo/theme.yaml` - Apply from URL
- `nvp theme list` - List installed themes
- `nvp theme get [name]` - Show theme details (defaults to active)
- `nvp theme use <name>` - Set active theme
- `nvp theme delete <name>` - Remove a theme
- `nvp theme generate` - Generate Lua files for active theme

#### Generated Theme Files
- `theme/palette.lua` - Color palette module with all theme colors
- `theme/init.lua` - Theme setup with helper functions:
  - `lualine_theme()` - Generate lualine theme from palette
  - `bufferline_highlights()` - Generate bufferline highlights
  - `telescope_border()` - Get telescope border color
  - `highlight(group, opts)` - Apply highlights using palette
- `plugins/nvp/colorscheme.lua` - Lazy.nvim plugin spec

#### Plugin Palette Integration
Other plugins can now use the active theme's colors:
```lua
local palette = require("theme").palette
local bg = palette.colors.bg
local fg = palette.colors.fg

-- Built-in helpers
local lualine_theme = require("theme").lualine_theme()
```

#### nvim-yaml-plugins Repository Update
- **Added 8 theme YAMLs** to https://github.com/rmkohlman/nvim-yaml-plugins
- Install themes via URL: `nvp theme apply --url github:rmkohlman/nvim-yaml-plugins/themes/catppuccin-mocha.yaml`

### 🧪 Testing

- **Theme system tests** - 14 tests across theme package:
  - `theme_test.go` - ParseYAML, Validate, ToYAML, Store tests
  - `generator_test.go` - Lua generation for multiple theme plugins
  - `library/library_test.go` - Library listing, categories, retrieval

### 📦 Files Created

```
pkg/nvimops/theme/
├── types.go           # Theme, ThemeYAML, ThemePlugin types
├── parser.go          # YAML parsing, validation, color checking
├── generator.go       # Lua code generation for all supported themes
├── store.go           # FileStore, MemoryStore implementations
├── theme_test.go      # Theme tests
├── generator_test.go  # Generator tests
└── library/
    ├── library.go     # Embedded theme library
    ├── library_test.go
    └── themes/        # 8 pre-defined theme YAMLs
```

---

## [0.4.1] - 2026-01-29

### 🚀 Added

#### URL Support for `nvp apply`
- **`--url` flag** - Fetch and apply plugin YAML directly from URLs
- **GitHub shorthand** - `github:user/repo/path/file.yaml` expands to raw GitHub URL
- **Multiple URLs** - Apply multiple plugins in one command: `--url url1 --url url2`
- **Combine with `-f`** - Use both local files and remote URLs together

**Example usage:**
```bash
nvp apply --url github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
nvp apply --url github:rmkohlman/nvim-yaml-plugins/plugins/treesitter.yaml \
          --url github:rmkohlman/nvim-yaml-plugins/plugins/lspconfig.yaml
```

#### Structured Logging for nvp
- **`-v/--verbose` flag** - Enable debug output to stderr
- **`--log-file` flag** - JSON logging to file for debugging
- **Silent by default** - Following CLI best practices
- **slog integration** - Same logging pattern as dvm

#### nvim-yaml-plugins Repository
- **New public repo** - https://github.com/rmkohlman/nvim-yaml-plugins
- **16 plugin YAMLs** - All embedded plugins exported as standalone files
- **Clean naming** - `telescope.yaml` instead of `02-telescope.yaml`

### 🧪 Testing

- **Added `cmd/nvp/root_test.go`** with 8 tests:
  - `TestFetchURL_GitHubShorthand` - HTTP mock server testing
  - `TestFetchURL_InvalidURL` - Error handling
  - `TestFetchURL_NotFound` - 404 response handling
  - `TestGitHubShorthandConversion` - Shorthand expansion (3 subtests)
  - `TestApplyPluginData` - Plugin creation
  - `TestApplyPluginData_InvalidYAML` - Error handling
  - `TestApplyPluginData_Update` - Plugin updates
  - `TestGetConfigDir` - Config directory resolution

### 📦 Files Modified

```
cmd/nvp/root.go       - Added URL support, logging flags, fetchURL()
cmd/nvp/root_test.go  - New test file with 8 tests
```

---

## [0.4.0] - 2026-01-29

### 🚀 Added

#### nvp (NvimOps) - Standalone Neovim Plugin Manager CLI
- **New `nvp` binary** - Standalone CLI for managing Neovim plugins using DevOps-style YAML configuration
- **Plugin Store** - File-based plugin storage at `~/.nvp/plugins/`
- **Plugin Library** - 16 pre-configured plugins ready to install:
  - telescope, treesitter, nvim-cmp, lspconfig, mason, gitsigns
  - lualine, which-key, copilot, comment, alpha, neo-tree
  - conform, nvim-lint, trouble, toggleterm
- **Lua Generation** - Generate lazy.nvim compatible Lua files from YAML definitions

#### nvp Commands
- `nvp init` - Initialize nvp store at `~/.nvp/`
- `nvp plugin add <name>` - Add plugin from YAML file or stdin
- `nvp plugin list` - List installed plugins
- `nvp plugin get <name>` - Show plugin details (YAML/JSON/table output)
- `nvp plugin delete <name>` - Remove a plugin
- `nvp library list` - List available plugins in the library
- `nvp library get <name>` - Show library plugin details
- `nvp library install <name>` - Install plugin from library to store
- `nvp generate` - Generate Lua files from installed plugins
- `nvp version` - Show nvp version
- `nvp completion` - Generate shell completions (bash/zsh/fish/powershell)

#### Decoupled Architecture (pkg/nvimops)
- **PluginStore interface** - Swappable storage backends:
  - `MemoryStore` - In-memory storage for testing
  - `FileStore` - File-based storage for production
  - `ReadOnlyStore` - Wrapper for read-only sources (library)
  - Future: `DBPluginStore` for dvm integration
- **LuaGenerator interface** - Swappable Lua generation:
  - `Generator` - Default lazy.nvim compatible generator
  - `MockGenerator` - For testing
  - Extensible for other plugin managers (packer, vim-plug)
- **ReadOnlySource interface** - Wrap any read-only source as a PluginStore
- **Comprehensive mock implementations** for all interfaces

#### Testing Infrastructure
- **Automated test script** - `tests/manual/nvp/test-nvp.sh`
  - 50+ automated tests covering all nvp functionality
  - Parts 1-4, 6-8 of the test plan
  - Verbose mode: `NVP_VERBOSE=1`
  - Keep output: `NVP_KEEP_OUTPUT=1`
- **Nvim config replica test** - `tests/manual/nvp/test-nvim-config-replica.sh`
  - Clones real nvim-config repo
  - Installs plugins from library
  - Generates Lua files
  - Verifies integration with Neovim
- **Interface compliance tests** - Verify all implementations satisfy interfaces
- **Swappability tests** - Same code works with different implementations

### 🔧 Changed

#### Package Rename
- **`pkg/nvimmanager` → `pkg/nvimops`** - Renamed for consistency with CLI name
- All imports updated across the codebase

#### GoReleaser Configuration
- **Fixed deprecation warnings** - Updated to latest GoReleaser syntax
- **`archives.builds` → `archives.ids`** - New archive syntax
- **Added `homebrew_casks`** - Recommended for pre-built binaries
- **Quarantine removal hooks** - For unsigned macOS binaries
- **`zap` section for nvp** - Clean up `~/.nvp` on Homebrew uninstall

### 📦 Files Created

```
pkg/nvimops/                      # Standalone nvim plugin management library
├── nvimops.go                    # Manager with swappable Store + Generator
├── nvimops_test.go
├── plugin/
│   ├── types.go                  # Plugin, PluginYAML types
│   ├── interfaces.go             # LuaGenerator interface
│   ├── yaml.go                   # YAML unmarshaling
│   ├── parser.go                 # YAML parsing
│   ├── generator.go              # Default Lua generator
│   ├── plugin_test.go
│   └── interface_test.go         # Generator interface tests
├── store/
│   ├── interface.go              # PluginStore interface
│   ├── readonly.go               # ReadOnlyStore wrapper
│   ├── memory.go                 # MemoryStore implementation
│   ├── file.go                   # FileStore implementation
│   ├── store_test.go
│   └── interface_test.go         # Store interface tests
└── library/
    ├── library.go                # Embedded plugin library
    ├── library_test.go
    └── plugins/                  # 16 embedded plugin YAMLs

cmd/nvp/                          # nvp CLI
├── root.go                       # Root command with subcommands
└── (Cobra command tree)

tests/manual/nvp/
├── test-nvp.sh                   # Automated test suite
└── test-nvim-config-replica.sh   # Real nvim config integration test

NVIMOPS_TEST_PLAN.md              # Comprehensive test plan for nvp
```

### 🧪 Testing

- **All Go tests passing** ✅
- **GoReleaser check passing** ✅
- **Interface compliance tests** - All implementations verified
- **Swappability tests** - Implementations are interchangeable

### 📚 Documentation

- **NVIMOPS_TEST_PLAN.md** - Comprehensive 8-part test plan
- **Architecture diagram** in test plan
- **Extensibility examples** for custom stores and generators

### 🎯 What's Next (v0.5.0)

- Integrate nvp with dvm (`dvm workspace add-plugin/remove-plugin`)
- Create `internal/db/plugin_store.go` - DBPluginStore for dvm
- Add more plugins to the library (indent-blankline, etc.)

---

## [0.3.3] - 2026-01-29

### 🚀 Added

#### Pre-generated Shell Completions
- **Shell completions included in release archives** - Bash, Zsh, and Fish completion scripts are now pre-generated during the build process and included in the release archives
- **Automatic completion installation via Homebrew** - `brew install rmkohlman/tap/dvm` now automatically installs shell completions without requiring manual user action
- **Bypass macOS sandbox restrictions** - Pre-built binaries no longer need to execute during Homebrew install, which was previously blocked by macOS sandbox

### 🔧 Changed

#### Release Archive Format
- **Archives now include `completions/` directory** with:
  - `dvm.bash` - Bash completion script
  - `_dvm` - Zsh completion script  
  - `dvm.fish` - Fish completion script
- **GoReleaser post-build hooks** generate completions after each platform build

### 📝 Notes

- This release resolves the Homebrew completion generation issue where pre-built binaries couldn't be executed during `brew install` due to macOS sandbox restrictions
- The completions are identical across all platforms (they're shell scripts, not platform-specific)

---

## [0.3.1] - 2026-01-29

### 🚀 Added

#### Multi-Platform Container Runtime Support
- **Platform detection** for OrbStack, Docker Desktop, Podman, and Colima
- **`dvm get platforms`** - List all detected container platforms with status
- **Automatic runtime selection** based on detected platform
- **Containerd support** for Colima with containerd runtime
- **Multiple socket path detection** for improved OrbStack reliability

#### Decoupled Architecture
- **ImageBuilder interface** with implementations:
  - `DockerBuilder` - Standard Docker build
  - `BuildKitBuilder` - BuildKit-based builds for containerd
  - `NerdctlBuilder` - nerdctl for Colima/containerd
- **Driver/DataStore/QueryBuilder interfaces** for database abstraction
- **ContainerRuntime interface** for multi-platform support
- **Formatter interface** with Plain and Colored implementations
- **Mock implementations** for all major interfaces (testing)

#### Structured Logging
- **slog integration** using Go's standard library
- **`-v/--verbose` flag** for debug output to stderr
- **`--log-file` flag** for JSON logging to file
- **Silent by default** following CLI best practices

#### Testing Infrastructure
- **Manual test scripts** in `tests/manual/`:
  - `part1-setup-and-build.sh` - 18 automated setup/build tests
  - `part2-post-attach.sh` - 16 automated post-attach tests
- **Comprehensive mock implementations** for unit testing
- **All 34 manual tests passing**

#### Documentation
- **CLAUDE.md** - AI assistant context and project architecture
- **STANDARDS.md** - Development standards and patterns
- **MANUAL_TEST_PLAN.md** - Comprehensive testing procedures

#### Neovim Configuration Management
- **`dvm nvim init` command** - Initialize local Neovim configuration from templates
  - Built-in templates: `minimal`, `kickstart`, `lazyvim`, `astronvim`
  - Remote URL support: Clone from any Git repository
  - GitHub shorthand: `github:user/repo` → `https://github.com/user/repo.git`
  - GitLab/Bitbucket support: `gitlab:user/repo`, `bitbucket:user/repo`
  - Subdirectory extraction: `--subdir` flag to use specific folder from repo
  - Overwrite protection: `--overwrite` flag required to replace existing config
- **`dvm nvim status` command** - Show local Neovim configuration status
  - Display config path, template used, last sync time
  - Show existence of config files
  - Track local/remote changes (stub for workspace sync)
- **`dvm nvim sync` command** - Pull config from workspace (stub implementation)
- **`dvm nvim push` command** - Push local config to workspace (stub implementation)

#### Remote Template System
- **Git URL auto-detection** - Automatically detect and normalize Git URLs
- **URL normalization** - Convert shorthand formats to full URLs
- **Subdirectory support** - Extract specific folders from repositories
- **`.git` removal** - Automatically remove Git metadata after cloning
- **Minimal template** - Full-featured minimal Neovim config with lazy.nvim

#### Shell Completion
- **Dynamic autocompletion** for template names with descriptions
- **Bash, Zsh, Fish, PowerShell** completion support via Cobra
- **Custom completion functions** for enhanced developer experience
- **Documentation** - Comprehensive shell completion guide

#### Build & Release
- **GoReleaser configuration** - Automated multi-platform releases
  - macOS (amd64, arm64)
  - Linux (amd64, arm64, 386)
  - Windows (amd64, 386)
- **Homebrew tap support** - Ready for distribution via Homebrew
- **Checksums and archives** - Secure distribution with verification
- **Version display fix** - Proper semver handling with `v` prefix

### 🧪 Testing

- **Added 19+ comprehensive tests** for Neovim functionality:
  - `nvim/url_test.go` (13 tests) - URL parsing and normalization
  - `nvim/manager_test.go` (8 tests) - Manager operations and status
  - `nvim/templates_test.go` (19 tests) - Template cloning and initialization
- **All tests passing** ✅ (38+ nvim tests, 66+ total)
- **Integration testing** - Manual testing of all URL formats and templates
- **Error handling coverage** - Invalid URLs, missing subdirectories, network failures

### 📚 Documentation

- **Created `docs/SHELL_COMPLETION.md`** - Shell completion installation guide
- **Created `docs/development/ADR-008-shared-nvim-library.md`** - Architecture decision for shared library
- **Created `docs/development/nvim-templates-repo-blueprint.md`** - Template repository design
- **Created `templates/README.md`** - Template usage guide
- **Created `templates/minimal/README.md`** - Minimal template documentation
- **Enhanced command help** - Comprehensive examples and usage information

### 🔧 Changed

- **Fixed version command** - Handle `v` prefix in git tags correctly
- **Improved error messages** - Clear feedback for common issues
- **Enhanced CLI UX** - Better help text and examples

### 🐛 Fixed

- **Build failing when `.config/nvim` doesn't exist** - Now shows skip message
- **`dvm attach` warning when image not built** - Clear warning with instructions
- **Podman buildkit compatibility** - Added `--load` flag for image loading
- **OrbStack detection** - Check multiple socket paths for reliability
- **Plugin delete UX** - Clearer messaging about what gets deleted

### 📦 Files Created

```
nvim/
├── manager.go (213 lines)       - Core Manager interface & implementation
├── templates.go (159 lines)     - Template initialization logic
├── url.go (89 lines)            - URL parsing utilities
├── manager_test.go (230 lines)  - Comprehensive unit tests
├── templates_test.go (400+ lines) - Template cloning tests
└── url_test.go (145 lines)      - URL parsing tests

cmd/
├── nvim.go (295 lines)          - Cobra commands for dvm nvim
└── completion.go (97 lines)     - Custom completion functions

templates/
├── README.md                    - Template documentation
└── minimal/
    ├── init.lua                 - Full-featured minimal config
    └── README.md                - Minimal template guide

docs/
├── SHELL_COMPLETION.md          - Shell completion guide
└── development/
    ├── ADR-008-shared-nvim-library.md
    └── nvim-templates-repo-blueprint.md
```

### 🎯 What's Next (v0.4.0)

- Extract shared library to `nvim-maestro-lib` repository
- Implement actual workspace sync functionality
- Add YAML remote fetch support (`url:` field in configs)
- Create public nvim-templates repository
- Support project/workspace templates from URLs

---

## [0.2.0] - 2026-01-24

### 🎨 Added

#### Professional Theme System
- **8 beautiful themes** for enhanced terminal output:
  - `catppuccin-mocha` - Soothing dark pastel colors
  - `catppuccin-latte` - Warm light pastel colors  
  - `tokyo-night` - Vibrant blue-purple dark theme
  - `nord` - Cool bluish minimal theme
  - `dracula` - Classic purple-pink dark theme
  - `gruvbox-dark` - Warm retro dark theme
  - `gruvbox-light` - Warm retro light theme
  - `auto` (default) - Auto-detects terminal light/dark mode
- **Auto-detection** of terminal color scheme using adaptive colors
- **Theme configuration** via `DVM_THEME` environment variable
- **Config file support** at `~/.devopsmaestro/config.yaml` for persistent theme settings
- **Dynamic theme switching** without restart

#### YAML Syntax Highlighting
- **Colored YAML keys** (cyan, bold) for better readability
- **Colored YAML values** (yellow) to distinguish from keys
- **Colored YAML comments** (gray) for subtle appearance
- **Applied to all `dvm get` commands** with `-o yaml` output format

### 🔧 Changed

- **Improved output readability** with themed color schemes across all commands
- **Enhanced YAML output** with syntax highlighting for better scanning
- **Made UI color system dynamic** (previously hardcoded constants)
- **Theme priority order**: environment variable > config file > auto-detection

### 🧪 Testing

- **Added 17 comprehensive theme tests** in `ui/themes_test.go`:
  - Theme switching and retrieval
  - Auto-detection logic
  - Environment variable override
  - Theme availability checks
- **Added 12 config system tests** in `config/config_test.go`:
  - Theme loading from environment
  - Theme loading from config file
  - Priority order verification
  - Default config creation
- **All 66 tests passing** ✅ (UI: 25, Theme: 17, Config: 12, Commands: 12)

### 📚 Documentation

- **Updated README.md** with comprehensive theme system documentation
- **Created LICENSING.md** - Dual-license guide (GPL-3.0 + Commercial)
- **Enhanced LICENSE-COMMERCIAL.txt** with professional terms and pricing
- **Added theme usage examples** with environment variable and config file setup

---

## [0.1.0] - 2026-01-23

### Initial Release

#### Core Features
- ✨ Project and workspace management
- 🔌 Database-backed Neovim plugin system
- 🐳 Container-native development environments
- 📦 Declarative YAML configuration
- 🎯 kubectl-style commands (projects and workspaces)

#### Commands Implemented
- `dvm init` - Initialize development environment
- `dvm project create/list/delete` - Project management
- `dvm workspace create/list/delete` - Workspace management
- `dvm get projects/workspaces` - Resource listing
- `dvm use project/workspace` - Context switching
- `dvm plugin apply/list/get/edit/delete` - Plugin management
- `dvm build` - Container image building
- `dvm attach` - Workspace attachment
- `dvm version` - Version information

#### Database
- SQLite-based storage at `~/.devopsmaestro/devopsmaestro.db`
- Tables for projects, workspaces, plugins, and workspace-plugin relations
- Database migrations support

#### Plugin System
- 16+ pre-configured plugins ready to use
- Support for lazy loading, dependencies, keymaps, and configuration
- YAML-based plugin definitions
- Database storage for plugin configurations

#### Documentation
- Comprehensive README with examples
- Installation guide (INSTALL.md)
- Homebrew tap setup (HOMEBREW.md)
- Architecture documentation

---

## [0.14.0] - 2026-02-19

### ✨ Features

#### TerminalPrompt Resource System
- **New Resource Kind**: `TerminalPrompt` for managing shell prompt configurations
- **Multi-prompt support**: Starship, Powerlevel10k, Oh-My-Posh
- **Theme integration**: Theme variable resolution (`${theme.red}`, `${theme.sky}`, etc.)
- **Database persistence**: Prompts stored in database for workspace sharing
- **Workspace-qualified naming**: Prompts named per app/workspace for isolation

#### dvt prompt CLI Commands
- **`dvt get prompts`** - List all terminal prompts with optional type filtering (`--type starship`)
- **`dvt get prompt <name>`** - Get specific prompt details and configuration
- **`dvt prompt apply -f <file>`** - Apply prompt from YAML file or URL
- **`dvt prompt delete <name>`** - Delete a terminal prompt
- **`dvt prompt generate <name>`** - Generate starship.toml configuration file
- **`dvt prompt set <name>`** - Set active prompt for current workspace

#### Personal Config Repository
- **rmkohlman/dvm-config repository** - User configuration storage and sharing
- **GitHub integration**: Apply configs with `dvm apply -f github:rmkohlman/dvm-config/...`
- **Version-controlled configurations**: Track and share personal setups across environments

### 🏗️ Architecture

#### Resource/Handler Integration
- **Unified CRUD system**: TerminalPrompt uses standardized Resource/Handler pattern
- **build.go refactored**: Now uses Resource/Handler pattern for prompt operations
- **Proper dependency injection**: Consistent with other resource types

#### Database Schema
- **`terminal_prompts` table**: Complete persistence layer for prompt configurations
- **Migration support**: Database schema versioning for prompt storage
- **Workspace relationships**: Prompts linked to workspace context

---

## [Unreleased]

### Planned Features

---

## [0.9.2] - 2026-02-18

### 🚀 Added

#### ColorProvider Architecture
- **`pkg/colors/` package** - Decoupled color/theme system with ColorProvider interface
- **Command-level integration** - Commands can now access ColorProvider from context
- **Render context support** - Theme integration through ColorProvider context

#### Dynamic Shell Completions
- **Resource-aware completions** - Dynamic completion for bash, zsh, and fish shells
- **Smart resource suggestions** - Context-aware completion based on current hierarchy

#### OpenCode Sub-agents
- **11 specialized agents** - Distributed AI assistance for development workflow:
  - `architecture` (advisory) - Design patterns and code review
  - `cli-architect` (advisory) - kubectl-style command design
  - `security` (advisory) - Security review and validation
  - `database` - Database schema, migrations, DataStore interface
  - `container-runtime` - Container operations and runtime management
  - `builder` - Image building and Dockerfile optimization
  - `render` - Output formatting, tables, and color rendering
  - `nvimops` - Neovim plugin/theme management
  - `theme` - Color systems, palettes, and ColorProvider
  - `test` - Test writing and execution
  - `document` - Documentation maintenance
  - `release` - **ALL git operations**, CI/CD, and release management
- **Workflow protocol** - Structured coordination between agents
- **Microservice mindset** - Clear interfaces and responsibility boundaries

#### Agent Coordination System
- **Mandatory task start checklist** - Ensures proper agent delegation
- **Workflow protocols** - Pre/post-invocation requirements for each agent
- **Git operation routing** - All git commands now route through release agent

### 🐛 Fixed

- **Release workflow race condition** - Resolved parallel job conflicts in GitHub Actions

### 📚 Documentation

- **Enhanced sub-agent documentation** - Updated with actual codebase structure
- **Workflow coordination guides** - Added protocols for agent coordination
- **Git operation routing** - Documented that release agent owns ALL git operations

### Planned Features (v0.4.0 and beyond)

#### Local Neovim Management (v0.4.0)
- [x] `dvm nvim init` - Initialize local Neovim configuration ✅ v0.3.0
- [x] `dvm nvim status` - Show local Neovim configuration status ✅ v0.3.0
- [x] Shell completion support ✅ v0.3.0
- [ ] `dvm nvim apply -f file.yaml` - Apply plugins to local Neovim
- [ ] `dvm nvim sync <workspace>` - Sync workspace config to local (full implementation)
- [ ] `dvm nvim push <workspace>` - Push local config to workspace (full implementation)
- [ ] `dvm nvim diff <workspace>` - Compare local vs workspace configs
- [ ] Fresh machine setup workflow
- [ ] Team configuration sharing

#### Shared Library Architecture (v0.4.0)
- [x] Design ADR for shared library ✅ v0.3.0
- [ ] Extract to `nvim-maestro-lib` repository
- [ ] Create standalone `nvim-maestro` CLI (v0.5.0)
- [ ] Publish shared library to Go modules

#### Template System (v0.4.0)
- [x] Remote URL template support ✅ v0.3.0
- [x] GitHub/GitLab/Bitbucket shorthand ✅ v0.3.0
- [x] Subdirectory extraction ✅ v0.3.0
- [ ] Create public `nvim-templates` repository
- [ ] Add more pre-configured templates
- [ ] YAML remote fetch (`url:` field in workspace/project configs)

#### Documentation & Guides (v0.4.0)
- [x] Shell completion guide ✅ v0.3.0
- [x] Neovim management documentation ✅ v0.3.0
- [ ] Comprehensive command documentation with status badges
- [ ] Getting-started guides
- [ ] Step-by-step tutorials
- [ ] YAML configuration examples
- [ ] Troubleshooting guides

#### kubectl-style Command Completeness
- [ ] `dvm apply -f file.yaml` - Top-level apply for all resource types
- [ ] `dvm edit plugin <name>` - Top-level edit command
- [ ] `dvm delete plugin <name>` - Enhanced delete with plugin support
- [ ] `dvm context` - Beautiful current context display

#### Enhanced UX
- [ ] Add deprecation warnings to old `dvm plugin` commands
- [ ] Add `--watch` flag for real-time resource updates
- [x] Shell completion (bash, zsh, fish, powershell) ✅ v0.3.0
- [ ] Progress bars for long-running operations
- [ ] Custom theme creation (v0.4.0)
- [ ] Theme preview command (`dvm theme list --preview`)

#### Advanced Features
- [ ] Resource labels and selectors
- [ ] Namespace support for isolation
- [ ] Backup and restore functionality
- [ ] Plugin marketplace/catalog
- [x] GoReleaser automation ✅ v0.3.0
- [x] Homebrew tap configuration ✅ v0.3.0

#### Quality & Testing
- [ ] Integration tests for full workflows
- [ ] Performance benchmarks
- [ ] CI/CD pipeline enhancements
- [ ] Code coverage reports

---

## Version History

- **[0.8.0]** - 2025-01-06 - New object hierarchy (Ecosystem/Domain/App/Workspace), model separation
- **[0.7.2]** - 2025-01-05 - Bug fixes and stability improvements
- **[0.7.1]** - 2026-02-04 - Unified resource pipeline, consistent command architecture
- **[0.7.0]** - 2026-02-03 - Terminal resize, timestamp-based image tags, auto-recreate containers
- **[0.6.0]** - 2026-02-03 - `dvm status`, kubectl aliases, `dvm detach`, context commands
- **[0.5.1]** - 2026-02-02 - BuildKit socket validation fix + documentation updates
- **[0.5.0]** - 2026-01-30 - NvimTheme system + exported palette for plugins
- **[0.4.1]** - 2026-01-29 - URL support for nvp apply + logging + tests
- **[0.4.0]** - 2026-01-29 - nvp (NvimOps) standalone CLI + decoupled architecture
- **[0.3.3]** - 2026-01-29 - Pre-generated shell completions in release archives
- **[0.3.1]** - 2026-01-29 - Multi-platform support + decoupled architecture
- **[0.3.0]** - 2026-01-24 - Neovim configuration management + remote URL templates
- **[0.2.0]** - 2026-01-24 - Theme system + YAML syntax highlighting
- **[0.1.0]** - 2026-01-23 - Initial release

---

## Links

- [GitHub Repository](https://github.com/rmkohlman/devopsmaestro)
- [Issue Tracker](https://github.com/rmkohlman/devopsmaestro/issues)
- [Releases](https://github.com/rmkohlman/devopsmaestro/releases)
- [Documentation](https://github.com/rmkohlman/devopsmaestro#readme)

---

## Notes

### Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR** version (X.0.0) - Incompatible API changes
- **MINOR** version (0.X.0) - New functionality (backward compatible)
- **PATCH** version (0.0.X) - Bug fixes (backward compatible)

### Backward Compatibility Promise

For v0.x releases:
- We maintain backward compatibility whenever possible
- Deprecation warnings given before command removal
- Breaking changes documented clearly
- Migration guides provided when needed

For v1.0+ releases:
- Strong backward compatibility guarantees
- Breaking changes only in major versions
- 6-month deprecation period minimum
