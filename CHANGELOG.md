# Changelog

All notable changes to DevOpsMaestro (dvm) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.3.0] - 2026-01-24

### 🚀 Added

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
