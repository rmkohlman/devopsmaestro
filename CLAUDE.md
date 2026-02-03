# DevOpsMaestro - AI Assistant Context

> **Purpose:** Public architecture reference for AI assistants working on DevOpsMaestro.  
> **For detailed session context:** See private `devopsmaestro-toolkit` repository.

---

## Quick Start

### For AI Assistants

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

### Two Binaries

| Binary | Purpose | Entry Point |
|--------|---------|-------------|
| `dvm` | Workspace/project management | `main.go` |
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
│                      Output Layer (output/)                      │
│  Formatter interface → table, yaml, json output                 │
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
│   ├── create.go          # dvm create project/workspace
│   ├── get.go             # dvm get projects/workspaces/plugins
│   └── nvp/               # NvimOps CLI (standalone)
│       ├── main.go        # nvp entry point
│       └── root.go        # All nvp commands
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
├── output/                # Output formatting
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
go test ./...
go test ./pkg/nvimops/... -v   # nvp tests
go test ./db/... -v            # database tests

# Lint
golangci-lint run
```

---

## Key Interfaces

### DataStore (`db/interfaces.go`)
```go
type DataStore interface {
    CreateProject(project *models.Project) error
    GetProjectByName(name string) (*models.Project, error)
    ListProjects() ([]*models.Project, error)
    // ... workspaces, plugins, context
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
