# Changelog

All notable changes to DevOpsMaestro (dvm) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

### Planned Features (v0.3.0 and beyond)

#### Local Neovim Management (v0.3.0)
- [ ] `dvm nvim init` - Initialize local Neovim configuration
- [ ] `dvm nvim apply -f file.yaml` - Apply plugins to local Neovim
- [ ] `dvm nvim sync <workspace>` - Sync workspace config to local
- [ ] `dvm nvim push <workspace>` - Push local config to workspace
- [ ] `dvm nvim diff <workspace>` - Compare local vs workspace configs
- [ ] Fresh machine setup workflow
- [ ] Team configuration sharing

#### Documentation & Guides (v0.3.0)
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
- [ ] Shell completion (bash, zsh, fish)
- [ ] Progress bars for long-running operations
- [ ] Custom theme creation (v0.3.0)
- [ ] Theme preview command (`dvm theme list --preview`)

#### Advanced Features
- [ ] Resource labels and selectors
- [ ] Namespace support for isolation
- [ ] Backup and restore functionality
- [ ] Plugin marketplace/catalog
- [ ] GoReleaser automation (v0.3.0)
- [ ] Homebrew tap (v0.3.0)

#### Quality & Testing
- [ ] Integration tests for full workflows
- [ ] Performance benchmarks
- [ ] CI/CD pipeline enhancements
- [ ] Code coverage reports

---

## Version History

- **[0.2.0]** - 2026-01-24 - kubectl-style plugin commands + beautiful UI
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
