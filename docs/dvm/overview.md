# dvm Overview

`dvm` (DevOpsMaestro) is a kubectl-style CLI for managing containerized development environments.

---

## What is dvm?

dvm provides:

- **Project management** - Track your codebases
- **Workspace management** - Isolated container environments per project
- **Container orchestration** - Build, attach, detach from dev containers
- **Neovim integration** - Pre-configured editor with LSP support

---

## Core Concepts

### Projects

A **project** represents a codebase on your filesystem:

```bash
dvm create project my-api --path ~/Developer/my-api
```

- Points to a directory containing your code
- Can have multiple workspaces
- Tracked in dvm's database

### Workspaces

A **workspace** is an isolated container environment:

```bash
dvm create workspace dev
```

- Belongs to a project
- Has its own container image
- Mounts the project directory
- Can have different configurations (plugins, tools, etc.)

### Context

The **context** tracks your currently active project and workspace:

```bash
dvm get ctx
# Project:   my-api
# Workspace: dev
```

Commands operate on the active context by default.

---

## Typical Workflow

```
┌─────────────────────────────────────────────────────┐
│  1. dvm init           Initialize dvm (one-time)    │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  2. dvm create project my-app --from-cwd            │
│     Create a project pointing to your code          │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  3. dvm use project my-app                          │
│     Set the active project                          │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  4. dvm create workspace dev                        │
│     Create a workspace (container environment)      │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  5. dvm use workspace dev                           │
│     Set the active workspace                        │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  6. dvm build                                       │
│     Build the container image                       │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  7. dvm attach                                      │
│     Enter the container and start coding!           │
└─────────────────────────────────────────────────────┘
```

---

## kubectl-style Commands

dvm follows kubectl patterns:

| Pattern | Example |
|---------|---------|
| `get` | `dvm get projects`, `dvm get workspaces` |
| `create` | `dvm create project`, `dvm create workspace` |
| `delete` | `dvm delete project`, `dvm delete workspace` |
| `apply` | `dvm apply -f workspace.yaml` |
| `use` | `dvm use project`, `dvm use workspace` |

### Resource Aliases

Use short aliases for faster commands:

| Resource | Alias | Example |
|----------|-------|---------|
| `projects` | `proj` | `dvm get proj` |
| `workspaces` | `ws` | `dvm get ws` |
| `context` | `ctx` | `dvm get ctx` |
| `platforms` | `plat` | `dvm get plat` |

---

## Container Platforms

dvm supports multiple container runtimes:

| Platform | Type | Notes |
|----------|------|-------|
| **OrbStack** | Docker | Recommended for macOS |
| **Docker Desktop** | Docker | Cross-platform |
| **Podman** | Docker-compatible | Rootless containers |
| **Colima** | Docker or containerd | Lightweight alternative |

Check detected platforms:

```bash
dvm get platforms
```

---

## Next Steps

- [Projects](projects.md) - Managing projects
- [Workspaces](workspaces.md) - Managing workspaces
- [Building & Attaching](build-attach.md) - Container lifecycle
- [Commands Reference](commands.md) - Complete command list
