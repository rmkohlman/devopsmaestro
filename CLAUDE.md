# DevOpsMaestro - AI Assistant Context

> **Purpose:** Public architecture reference for AI assistants working on DevOpsMaestro.  
> **For detailed session context:** See private `devopsmaestro-toolkit` repository.

---

## Quick Start

### For AI Assistants

**⚠️ IMPORTANT: Before writing any code, read these files in order:**

1. **This file (CLAUDE.md)** - Architecture overview
2. **STANDARDS.md** - Design patterns, coding standards, and development philosophy (REQUIRED)
3. **.claude/instructions.md** - Mandatory patterns checklist and quick reference (REQUIRED)
4. **docs/vision/architecture.md** - Complete architecture vision document
5. **MANUAL_TEST_PLAN.md** - Testing procedures for new features

**Detailed planning and session docs are in the private toolkit repository:**

```
~/Developer/tools/devopsmaestro_toolkit/
├── MASTER_VISION.md      # Complete vision, architecture, roadmap
├── CLAUDE.md             # Detailed AI context with session protocols  
├── current-session.md    # What's in progress RIGHT NOW
├── decisions.md          # Technical decisions history
└── repos/dvm/            # This repository (cloned here)
```

If you have access to the toolkit folder, read those files first.

---

## Project Overview

**DevOpsMaestro** is a kubectl-style CLI toolkit for managing containerized development environments with a GitOps mindset.

### Core Object Hierarchy

```
Ecosystem → Domain → App → Workspace (dev mode)
                      ↓
                  (live mode - managed by Operator)
```

| Object | Purpose | Status |
|--------|---------|--------|
| **Ecosystem** | Top-level platform grouping | 🟢 Complete (v0.8.0) |
| **Domain** | Bounded context (was "Project") | 🟢 Complete (v0.8.0) |
| **App** | The codebase/application | 🟢 Complete (v0.8.0) |
| **Project** | ⚠️ DEPRECATED - use Domain/App | 🟡 To migrate |
| **Workspace** | Dev environment for an App | 🟢 Exists |
| **Context** | Active selection state | 🟢 Exists |

### Two Operating Modes

| Mode | Requirements | Use Case |
|------|--------------|----------|
| **dvm alone** | Docker only | Basic workspace management |
| **dvm + Operator** | Docker + k8s (Colima/OrbStack) | Full DevOps: live mode, CI/CD |

### Two Binaries

| Binary | Purpose | Entry Point |
|--------|---------|-------------|
| `dvm` | Workspace/app management | `main.go` |
| `nvp` | Neovim plugin/theme management (standalone) | `cmd/nvp/main.go` |

### Key Value Proposition

- **"GitOps for Dev Environments"** - Declarative, YAML-based configuration
- **Single command setup** - `dvm attach` gets a fully-configured workspace
- **Neovim integration** - Pre-configured editor with LSP, plugins, themes

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer (cmd/)                         │
│  Commands: create, get, apply, delete, use, build, attach       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Render Layer (render/)                      │
│  Decoupled output: JSON, YAML, Colored, Plain, Table renderers  │
│  Commands prepare data → Renderers decide how to display        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Database Layer (db/)                          │
│  DataStore interface → SQLDataStore (SQLite)                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Container Layer (operators/)                   │
│  ContainerRuntime interface → Docker, OrbStack, Podman          │
└─────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
devopsmaestro/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # dvm root command
│   ├── create.go          # dvm create app/workspace
│   ├── get.go             # dvm get apps/workspaces/plugins
│   └── nvp/               # NvimOps CLI (standalone)
│       ├── main.go        # nvp entry point
│       └── root.go        # All nvp commands
│
├── render/                # Decoupled rendering system
│   ├── interface.go       # Renderer interface, data types
│   ├── registry.go        # Register(), Output(), Msg() helpers
│   ├── renderer_*.go      # JSON, YAML, Colored, Plain, Table
│   └── types.go           # RenderType, Options, Config
│
├── pkg/nvimops/           # NvimOps library
│   ├── plugin/            # Plugin types, parser, generator
│   ├── theme/             # Theme types, parser, generator
│   ├── store/             # Storage interfaces
│   └── library/           # Embedded plugin/theme library
│
├── db/                    # Database layer (dvm only)
├── operators/             # Container runtime layer
├── builders/              # Image builder layer
├── models/                # Data models
└── migrations/            # Database migrations
```

---

## Core Design Principles

1. **Decoupling** - Interface → Implementation → Factory pattern
2. **kubectl Patterns** - Familiar commands: `create`, `get`, `apply`, `delete`
3. **Declarative** - YAML-based configuration
4. **Modular** - Sub-tools work independently (nvp doesn't need dvm)
5. **Testable** - Mocks for all major interfaces

---

## Quick Commands

```bash
# Build
go build -o dvm .              # DevOpsMaestro
go build -o nvp ./cmd/nvp/     # NvimOps

# Test
go test ./...                  # All tests
go test ./... -race            # With race detector (CI uses this)
go test ./pkg/nvimops/... -v   # nvp tests
go test ./db/... -v            # database tests

# Lint (requires golangci-lint)
golangci-lint run
golangci-lint run --timeout=5m # With extended timeout
```

---

## CI/CD

### GitHub Actions Workflows

| Workflow | File | Trigger | Jobs |
|----------|------|---------|------|
| CI | `.github/workflows/ci.yml` | Push/PR to main | Test, Build |
| Release | `.github/workflows/release.yml` | Tag push (v*) | GoReleaser |

### CI Jobs

- **Test**: Runs `go test ./... -v -race -coverprofile=coverage.out`
- **Build**: Builds both `dvm` and `nvp` binaries, verifies with `version` command
- **Lint**: *Temporarily disabled* - waiting for golangci-lint to support Go 1.25

### Requirements

- **Go version**: 1.25.0 (set in `go.mod`)
- **Race detector**: All tests must pass with `-race` flag

### Checking CI Status

```bash
gh run list --limit 3          # Recent runs
gh run view <RUN_ID>           # View specific run
gh run watch <RUN_ID>          # Watch live
```

---

## Key Interfaces

### Renderer (`render/interface.go`)
```go
type Renderer interface {
    Render(w io.Writer, data any, opts Options) error
    RenderMessage(w io.Writer, msg Message) error
    Name() RendererName
    SupportsColor() bool
}
```

### DataStore (`db/interfaces.go`)
```go
type DataStore interface {
    // Ecosystem operations (v0.8.0+)
    CreateEcosystem(ecosystem *models.Ecosystem) error
    GetEcosystemByName(name string) (*models.Ecosystem, error)
    
    // Domain operations (v0.8.0+)
    CreateDomain(domain *models.Domain) error
    GetDomainByName(name string) (*models.Domain, error)
    
    // App operations (v0.8.0+)
    CreateApp(app *models.App) error
    GetAppByName(name string) (*models.App, error)
    
    // Project operations (DEPRECATED - migrate to App)
    CreateProject(project *models.Project) error
    GetProjectByName(name string) (*models.Project, error)
    
    // Workspace, plugins, context...
}
```

### ContainerRuntime (`operators/runtime_interface.go`)
```go
type ContainerRuntime interface {
    BuildImage(ctx context.Context, opts BuildOptions) error
    StartWorkspace(ctx context.Context, opts StartOptions) (string, error)
    AttachToWorkspace(ctx context.Context, containerID string) error
    // ...
}
```

### PluginStore (`pkg/nvimops/store/interface.go`)
```go
type PluginStore interface {
    Save(plugin *plugin.Plugin) error
    Get(name string) (*plugin.Plugin, error)
    List() ([]*plugin.Plugin, error)
    Delete(name string) error
}
```

---

## Documentation Reference

| Document | Purpose |
|----------|---------|
| `README.md` | User documentation |
| `CHANGELOG.md` | Version history |
| `STANDARDS.md` | Code standards |
| `ARCHITECTURE.md` | Quick architecture reference |
| `docs/vision/architecture.md` | **Complete architecture vision** |
| `MANUAL_TEST_PLAN.md` | Testing procedures |
| `docs/development/release-process.md` | Release workflow |

---

## GitHub Resources

| Resource | URL |
|----------|-----|
| Main Repo | [github.com/rmkohlman/devopsmaestro](https://github.com/rmkohlman/devopsmaestro) |
| Homebrew Tap | [github.com/rmkohlman/homebrew-tap](https://github.com/rmkohlman/homebrew-tap) |
| Plugin Library | [github.com/rmkohlman/nvim-yaml-plugins](https://github.com/rmkohlman/nvim-yaml-plugins) |

---

## Installation

```bash
# Homebrew
brew tap rmkohlman/tap
brew install nvimops        # nvp only
brew install devopsmaestro  # dvm (includes nvp)

# From source
go build -o dvm .
go build -o nvp ./cmd/nvp/
```

---

**For detailed architecture, roadmap, and session context, see the private `devopsmaestro-toolkit` repository.**
