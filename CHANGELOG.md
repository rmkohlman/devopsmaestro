# Changelog

All notable changes to DevOpsMaestro (dvm) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

## [Unreleased]

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
