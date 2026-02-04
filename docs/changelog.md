# Changelog

All notable changes to DevOpsMaestro are documented in the [CHANGELOG.md](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) file in the repository.

## Latest Releases

### v0.7.1 (2026-02-04)

**Unified Resource Pipeline**

- Added `pkg/resource/` package with unified resource interface following kubectl patterns
- Added `pkg/source/` package for source resolution (file, URL, stdin, GitHub)
- Refactored `apply`, `get`, `delete` commands to use unified handlers
- Consistent architecture across all nvim resource operations

### v0.7.0 (2026-02-03)

**Terminal Resize + Image Versioning**

- Full terminal window on attach with dynamic resize handling
- Timestamp-based image tags (`YYYYMMDD-HHMMSS`) instead of `:latest`
- Auto-recreate containers when image changes
- kubectl-style workspace plugin commands (`dvm apply nvim plugin -f file.yaml`)

### v0.6.0 (2026-02-03)

**Status Command + Aliases**

- New `dvm status` command showing context, runtime, and containers
- kubectl-style aliases: `proj`, `ws`, `ctx`, `plat`
- `dvm detach` command to stop workspace containers
- `dvm get context` command

### v0.5.0 (2026-01-30)

**Theme System**

- YAML-based NvimTheme resource type
- 8 pre-defined themes (tokyonight, catppuccin, gruvbox, nord, etc.)
- Theme palette exported for plugin integration
- Full theme CLI commands

### v0.4.0 (2026-01-29)

**NvimOps Standalone CLI**

- New `nvp` binary for managing Neovim plugins
- 16+ pre-configured plugins in library
- Decoupled architecture with swappable stores
- Lua generation for lazy.nvim

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
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
