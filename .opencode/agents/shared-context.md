# Shared Project Context

> **This is a reference file, not an agent.** All agents should reference this file for shared project knowledge instead of duplicating it.

---

## Project Overview

**DevOpsMaestro** is a kubectl-style CLI toolkit for managing containerized development environments with a GitOps mindset.

- **Module**: `devopsmaestro` (Go 1.25.0)
- **Current Version**: v0.32.6
- **Codebase**: 150K+ lines of Go across 28 packages and 550+ files

### Two Binaries

| Binary | Purpose | Entry Point | Build Command |
|--------|---------|-------------|---------------|
| `dvm` | Workspace/app management | `main.go` | `go build -o dvm .` |
| `nvp` | Neovim plugin/theme management | `cmd/nvp/main.go` | `go build -o nvp ./cmd/nvp/` |

---

## Object Hierarchy

```
Ecosystem -> Domain -> App -> Workspace
   (org)    (context) (code)  (dev env)
```

| Object | Purpose | Status |
|--------|---------|--------|
| **Ecosystem** | Top-level platform grouping | Complete |
| **Domain** | Bounded context | Complete |
| **App** | The codebase/application | Complete |
| **Workspace** | Dev environment for an App | Complete |
| **Project** | DEPRECATED - use Domain/App | |

---

## Design Philosophy

**We build loosely coupled, decoupled, modular, cohesive code with responsibility segregation.**

### The Microservice Mindset

Each package should be treated like a microservice:
- Has a **clean interface boundary** (the contract)
- **Hides implementation details** (consumers don't know how it works)
- Can be **swapped without affecting consumers** (new database? new runtime? no problem)
- **Owns its domain completely** (no one else touches its internals)

### Core Design Patterns

| Pattern | Description | Example |
|---------|-------------|---------|
| **Interface -> Implementation -> Factory** | Everything is swappable | `DataStore`, `ContainerRuntime`, `ImageBuilder` |
| **Resource/Handler** | All CRUD goes through `pkg/resource/` | `EcosystemHandler`, `AppHandler`, `NvimPluginHandler` |
| **Dependency Injection** | Get from context, don't create internally | `getDataStore(cmd)` from Cobra context |
| **Registry** | Consumers use helpers, not direct instantiation | `render.Output()`, `render.Msg()` |
| **Factory** | Abstract creation logic | `NewContainerRuntime()`, `CreateDataStore()`, `NewImageBuilder()` |

### Dependency Flow

```
cmd/ (thin CLI layer, delegates to packages)
  |
  +-- render/       (Renderer interface)
  +-- db/           (DataStore interface)
  +-- operators/    (ContainerRuntime interface)
  +-- builders/     (ImageBuilder interface)
  +-- pkg/resource/ (Resource/Handler system)
  +-- pkg/nvimops/  (PluginStore, LuaGenerator interfaces)
  +-- pkg/colors/   (ColorProvider interface)
  +-- pkg/terminalops/ (PromptRenderer interface)
```

### What This Means in Practice

- **Loosely Coupled**: Components interact through interfaces, not concrete implementations
- **Decoupled**: Changes in one module don't cascade to others
- **Modular**: Each piece can be developed, tested, and replaced independently
- **Cohesive**: Related functionality is grouped together
- **Testability**: Mock any dependency for unit tests
- **Extensibility**: Add new features without breaking existing code

---

## Package Structure

```
cmd/           -> CLI commands (Cobra) - thin, delegates to packages (~22K lines)
db/            -> DataStore interface + SQLite implementation (~20K lines)
operators/     -> ContainerRuntime interface + implementations (~5.9K lines)
builders/      -> ImageBuilder interface + implementations (~4.5K lines)
render/        -> Renderer interface + implementations (~2.8K lines)
models/        -> Data models (no business logic) (~4.2K lines)
migrations/    -> Database migrations (sqlite/)
config/        -> Configuration handling (~741 lines)
utils/         -> Utility functions (~410 lines)
ui/            -> UI components (~1.5K lines)
nvim/          -> Nvim utilities (~2.1K lines)
templates/     -> Templates (~710 lines)
pkg/
  nvimops/     -> NvimOps library - plugins, themes, Lua generation (~15.9K lines)
  terminalops/ -> Terminal operations - prompts, shell, wezterm (~17.8K lines)
  colors/      -> ColorProvider interface, theme provider, context (~5.4K lines)
  palette/     -> Shared palette utilities (~1.4K lines)
  resource/    -> Resource/Handler system (kubectl patterns) (~6.3K lines)
  registry/    -> Registry system (~20.4K lines)
  crd/         -> CRD support (~3.1K lines)
  resolver/    -> Dependency resolution (~457 lines)
  mirror/      -> Mirror management (~1.9K lines)
  source/      -> Source management (~1.1K lines)
  secrets/     -> Secrets management (~2.1K lines)
  preflight/   -> Preflight checks (~920 lines)
  workspace/   -> Workspace utilities (~532 lines)
integration_test/ -> Integration tests (~4.4K lines)
```

---

## Resource Types

| Kind | Handler | Status |
|------|---------|--------|
| `Ecosystem` | `EcosystemHandler` | Complete |
| `Domain` | `DomainHandler` | Complete |
| `App` | `AppHandler` | Complete |
| `NvimPlugin` | `NvimPluginHandler` | Complete |
| `NvimTheme` | `NvimThemeHandler` | Complete |
| `Project` | -- | Deprecated - migrate to Domain/App |
| `Workspace` | -- | Needs migration |

---

## Parallel Work Segments

The orchestrator (CLAUDE.md) uses these safe parallel boundaries when fanning out multiple developer agent instances:

| Segment | Packages | Independence |
|---------|----------|-------------|
| **A: Container/Build Pipeline** | `operators/`, `builders/` | Independent (share interfaces only) |
| **B: Nvim Plugin Ecosystem** | `pkg/nvimops/**`, `nvim/` | Mostly independent |
| **C: Terminal Operations** | `pkg/terminalops/**` | Fully independent |
| **D: Color/Theme System** | `pkg/colors/**`, `pkg/palette/` | Independent |
| **E: Resource Framework** | `pkg/resource/**`, `pkg/crd/` | Handler implementations are very independent |
| **F: Database Layer** | `db/`, `migrations/sqlite/` | Adding methods safe; interface changes high-risk |
| **G: Registry System** | `pkg/registry/**` | Fully self-contained |
| **H: Standalone Utilities** | `pkg/mirror/`, `pkg/source/`, `pkg/secrets/`, `pkg/resolver/`, `pkg/preflight/`, `pkg/workspace/` | Independent |

**High-risk cross-cutting**: Changes to `DataStore` interface, `ContainerRuntime` interface, or `models/` affect multiple segments.

---

## v0.19.0+ Workspace Isolation Architecture

### Goal

Complete workspace isolation -- users only install Container Runtime + dvm, everything else lives inside workspaces.

### Directory Structure

```
~/.devopsmaestro/
  devopsmaestro.db          # Source of truth
  repos/                    # Bare repo mirrors
  registry/                 # Zot image registry data
  workspaces/
    {workspace-id}/
      repo/                 # Git clone from mirror
      volume/               # Persistent data (nvim-data/, nvim-state/, cache/)
      .dvm/                 # Generated configs (nvim/, shell/)
```

### Tool Hierarchy

| Tool | Scope | Target |
|------|-------|--------|
| **nvp** (standalone) | LOCAL | `~/.config/nvim/` for users who want local nvim setup |
| **dvt** (standalone) | LOCAL | `~/.config/starship.toml`, `~/.zshrc` for users who want local shell setup |
| **dvm** (workspaces) | ISOLATED | Workspace `.dvm/` directories only, never host paths |

### Key Requirements

| Requirement | Details |
|-------------|---------|
| **Parameterized Paths** | All config generators accept output path parameter (not hardcoded `~/.config/`) |
| **Volume Mounting** | Workspace data mounts to `/workspace/volume/` in container |
| **SSH Isolation** | SSH keys only mounted when workspace explicitly requests (`--mount-ssh`) |
| **Credential Scoping** | Credentials bound to scope (global/ecosystem/domain/app/workspace) |
| **No Host Pollution** | No writes to host `~/.config/`, `~/.local/`, `~/.zshrc` from dvm |

### Container Volume Strategy

```
Host paths:
~/.devopsmaestro/workspaces/{id}/
  .dvm/
    nvim/                    <- Generated configs
      lua/plugins/nvp/
  volume/
    nvim-data/               <- XDG_DATA_HOME/nvim (plugins, lazy)
    nvim-state/              <- XDG_STATE_HOME/nvim (shada, swap)
    nvim-cache/              <- XDG_CACHE_HOME/nvim

Container mounts:
/workspace/repo      <- Git clone (from bare mirror or direct)
/workspace/volume    <- Persistent data (nvim-data, nvim-state, cache)
/workspace/.dvm      <- Generated configs (nvim, shell)
/home/devuser/.ssh   <- ONLY if opts.MountSSH == true
```

---

## GitHub Resources

| Resource | URL | Visibility |
|----------|-----|------------|
| Main Repo | github.com/rmkohlman/devopsmaestro | Public |
| Toolkit | github.com/rmkohlman/devopsmaestro-toolkit | Private |
| Homebrew Tap | github.com/rmkohlman/homebrew-tap | Public |
| Plugin Library | github.com/rmkohlman/nvim-yaml-plugins | Public |

---

## Reference Documents

| Document | Location | Purpose |
|----------|----------|---------|
| `STANDARDS.md` | dvm repo | Design patterns, coding standards, development philosophy |
| `MASTER_VISION.md` | toolkit repo | Vision, architecture, roadmap, backlog |
| `current-session.md` | toolkit repo | Active work state |
| `decisions.md` | toolkit repo | Technical decisions with rationale |
