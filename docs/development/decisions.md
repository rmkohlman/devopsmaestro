# Architecture Decision Records

This document tracks major technical decisions made during DevOpsMaestro's development.

---

## ADR-001: Theme System with Pre-defined Themes

**Date:** 2026-01-24
**Status:** ✅ Accepted
**Version:** v0.2.0

### Context

Users wanted colorized terminal output that respects their theme preferences. We evaluated several approaches for providing theme support.

### Decision

Implement 8 pre-defined professional themes with automatic light/dark detection:
- catppuccin-mocha / catppuccin-latte
- tokyo-night
- nord
- dracula
- gruvbox-dark / gruvbox-light
- auto (adaptive)

Configuration via `DVM_THEME` environment variable or `~/.devopsmaestro/config.yaml`.

### Rationale

1. **Reliability** — Works consistently across all terminal emulators
2. **Industry Standard** — Used by popular tools (bat, gh, kubectl)
3. **Coverage** — 8 themes cover 95%+ of user preferences
4. **Simplicity** — Easy to configure, test, and maintain
5. **Extensibility** — Easy to add more themes in future releases

### Consequences

**Positive:**
- Consistent, beautiful output across all terminals
- Fast theme switching (< 1ms)
- Professional appearance out-of-the-box

**Negative:**
- No custom theme creation (planned for future releases)
- Limited to 8 themes initially

### Alternatives Considered

- **Terminal theme passthrough:** Unreliable across different terminals
- **Custom theme creation:** Too complex for initial release
- **Single fixed theme:** Not flexible enough for diverse user preferences

---

## ADR-002: Keep a Changelog Format

**Date:** 2026-01-24
**Status:** ✅ Accepted
**Version:** v0.2.0

### Context

Needed a consistent format for CHANGELOG.md and GitHub Release notes.

### Decision

Use [Keep a Changelog](https://keepachangelog.com/) format with standard categories: Added, Changed, Fixed, Deprecated, Removed, Security.

### Rationale

1. **Industry Standard** — Used by Docker, Terraform, many CLI tools
2. **Scannable** — Easy to quickly find relevant changes
3. **Semantic Versioning** — Maps naturally to semver
4. **Maintainable** — Fast to write and update

### Consequences

**Positive:**
- Quick scanning for specific change types
- Clear categorization
- Works well with semantic versioning

**Negative:**
- Less narrative than Kubernetes-style release notes

---

## ADR-003: Dual-License Model (GPL-3.0 + Commercial)

**Date:** 2026-01-24
**Status:** ✅ Accepted
**Version:** v0.2.0

### Context

Wanted to make DevOpsMaestro free for individuals while creating a sustainability model for long-term development.

### Decision

Implement dual-licensing:
- **GPL-3.0:** Free for personal use, open source projects, educational use
- **Commercial License:** Required for corporate/proprietary use

### Rationale

1. **Sustainability** — Revenue model for continued development
2. **Open Source Values** — Keeps tool free for individuals and the OSS community
3. **Fair Use** — Those profiting from it contribute back
4. **Proven Model** — Successfully used by MySQL, Qt, GitLab

### Consequences

**Positive:**
- Free for the vast majority of users
- Revenue potential from enterprise users
- Protects against proprietary forks

**Negative:**
- Some companies avoid GPL-licensed tools
- Requires license enforcement

---

## ADR-004: Cross-Platform Binary Distribution

**Date:** 2026-01-24
**Status:** ✅ Accepted
**Version:** v0.2.0

### Context

Users should be able to download and run DevOpsMaestro without installing Go.

### Decision

Build and distribute binaries for 4 platforms:
- macOS arm64 (Apple Silicon)
- macOS amd64 (Intel)
- Linux amd64
- Linux arm64

Include SHA256 checksums for verification.

### Rationale

1. **User Experience** — Download and run immediately, no toolchain required
2. **Coverage** — Covers 99%+ of developer machines
3. **Professional Standard** — Expected for CLI tools
4. **Security** — Checksums allow download verification

### Consequences

**Positive:**
- No Go installation required
- Fast, simple installation
- Verified downloads via checksums

**Negative:**
- Must build for each release (automated via GoReleaser)

---

## ADR-005: YAML Syntax Highlighting

**Date:** 2026-01-24
**Status:** ✅ Accepted
**Version:** v0.2.0

### Context

Users found YAML output difficult to scan. Wanted color distinction between keys and values.

### Decision

Apply syntax highlighting to all `-o yaml` output:
- Keys: Cyan + Bold
- Values: Yellow
- Comments: Gray

### Rationale

1. **Readability** — Easy to distinguish keys from values at a glance
2. **Standard Practice** — Matches behavior of editors like VS Code and Vim
3. **Theme Integration** — Uses the active theme's color palette

### Consequences

**Positive:**
- Much easier to scan YAML output
- Consistent with editor experience
- Works with all 8 themes

**Negative:**
- Small rendering overhead for large outputs (acceptable trade-off)

---

## ADR-006: `nvp` as a Standalone Binary

**Date:** 2026-02-04
**Status:** ✅ Accepted and Implemented
**Version:** v0.7.0+

### Context

An earlier proposal suggested integrating Neovim management into `dvm` via `dvm nvim` subcommands. During planning it became clear that:
- Neovim plugin/theme management is a distinct domain with its own release cadence
- Users who don't use Docker/containers still benefit from Neovim management
- A standalone binary enables Homebrew distribution without CGO/SQLite requirements
- Separation reduces binary size for Neovim-only users

### Decision

Ship `nvp` as a separate standalone binary alongside `dvm`, released from the same repository and version tag.

**Key commands:**
```bash
nvp list                        # List installed plugins
nvp get <name>                  # Show plugin details
nvp enable / disable <name>     # Toggle plugin
nvp library list/install        # Browse and install from library
nvp theme list/get/use          # Theme management
nvp theme generate              # Generate Lua config files
nvp apply -f FILE               # Apply YAML config (file, URL, GitHub shorthand)
nvp config init/show/edit       # Manage nvp configuration
```

### Rationale

1. **Separation of Concerns** — Neovim management is distinct from workspace/container management
2. **Smaller Install** — Users get only what they need
3. **Independent Feature Sets** — `nvp` and `dvm` ship from the same tag but evolve independently
4. **Homebrew-Friendly** — No CGO dependency enables seamless cross-compilation

### Consequences

**Positive:**
- Both binaries ship from a single release tag
- Clean separation: `dvm` owns workspace/container management; `nvp` owns Neovim plugin/theme management
- `nvp` is available via Homebrew tap

**Negative:**
- Two binaries to install (mitigated by the Homebrew tap installing both)

---

**Last Updated:** 2026-04-09 (v0.57.1)
