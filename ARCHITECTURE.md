# DevOpsMaestro Architecture

> Quick reference for the decoupled patterns. Read before writing new code.
> For the complete architecture vision, see `docs/vision/architecture.md`.

---

## Object Hierarchy

```
Ecosystem → Domain → App → Workspace (dev mode)
                      ↓
                  (live mode - managed by Operator)
```

| Object | Purpose | Status |
|--------|---------|--------|
| **Ecosystem** | Top-level platform grouping | 🔴 v0.8.0 |
| **Domain** | Bounded context (replaces Project) | 🔴 v0.8.0 |
| **App** | The codebase (has `path`) | 🔴 v0.8.0 |
| **Project** | ⚠️ DEPRECATED | 🟡 Migrate to App |
| **Workspace** | Dev environment for App | 🟢 Exists |

---

## ✅ Code Review Checklist

Before writing or reviewing code, verify:

```
□ Functions accept INTERFACES, not concrete structs
□ Factory functions return INTERFACES, not concrete types
□ No direct instantiation of structs outside their package
□ New implementations use existing interfaces (don't create new ones unnecessarily)
□ cmd/ layer only imports interfaces, never implementations
□ Business logic uses DataStore, not *SQLDataStore
□ Container ops use ContainerRuntime, not *DockerRuntime
□ Output uses render.Output(), not fmt.Print()
□ Tests use mocks, not real implementations
```

**Quick smell test:** If you see `*SQLDataStore`, `*DockerRuntime`, or `fmt.Printf` in `cmd/`, it's likely tightly coupled.

---

## Core Pattern

**Every subsystem follows: Interface → Implementations → Factory**

```
┌─────────────────────────────────────────────────────────────────┐
│                         INTERFACE                                │
│              (defines the contract/behavior)                     │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
       ┌───────────┐   ┌───────────┐   ┌───────────┐
       │  Impl A   │   │  Impl B   │   │   Mock    │
       └───────────┘   └───────────┘   └───────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FACTORY FUNCTION                              │
│          NewXxx(config) → returns Interface                      │
└─────────────────────────────────────────────────────────────────┘
```

**Rules:**
- Functions accept **interfaces**, not concrete types
- Factories return **interfaces**, not structs
- Every interface has a **mock** for testing

---

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      CLI Layer (cmd/)                            │
│            create, get, use, build, attach, detach              │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│   render/     │     │     db/       │     │  operators/   │
│   Renderer    │     │   DataStore   │     │ContainerRuntime│
└───────────────┘     └───────────────┘     └───────────────┘
        │                     │                     │
        │                     ▼                     ▼
        │             ┌───────────────┐     ┌───────────────┐
        │             │   builders/   │     │ pkg/nvimops/  │
        │             │  ImageBuilder │     │  PluginStore  │
        │             └───────────────┘     └───────────────┘
        │                     │                     │
        └─────────────────────┴─────────────────────┘
                              │
                              ▼
                      ┌───────────────┐
                      │   models/     │
                      │ Data structs  │
                      └───────────────┘
```

---

## Render Layer

**Commands prepare data → Renderers decide display**

```
  cmd/get.go                    render/
 ┌──────────────┐           ┌──────────────────────────────────┐
 │ data := ...  │           │         render.OutputWith()      │
 │              │ ────────► │                                  │
 │ OutputWith() │           │  flag -o > env DVM_RENDER > default
 └──────────────┘           └──────────────┬───────────────────┘
                                           │
                  ┌────────────────────────┼────────────────────────┐
                  ▼                        ▼                        ▼
           ┌────────────┐           ┌────────────┐           ┌────────────┐
           │   JSON     │           │  Colored   │           │   Plain    │
           │  Renderer  │           │ (default)  │           │  Renderer  │
           └────────────┘           └────────────┘           └────────────┘
```

| Interface | Implementations | Location |
|-----------|-----------------|----------|
| `Renderer` | Colored, Plain, JSON, YAML, Table | `render/renderer_*.go` |

---

## Database Layer

**Two-tier: High-level DataStore → Low-level Driver**

```
  Application Code                    db/
 ┌──────────────────┐           ┌─────────────────────┐
 │ store.ListApps()              │     DataStore       │  High-level API
 │ store.CreateWorkspace()  ──► │  (business logic)   │
 └──────────────────┘           └──────────┬──────────┘
                                           │
                                           ▼
                                ┌─────────────────────┐
                                │      Driver         │  Low-level SQL
                                │  (Connect, Query)   │
                                └──────────┬──────────┘
                                           │
                          ┌────────────────┴────────────────┐
                          ▼                                 ▼
                   ┌────────────┐                    ┌────────────┐
                   │   SQLite   │                    │  (future)  │
                   │   Driver   │                    │  Postgres  │
                   └────────────┘                    └────────────┘
```

| Interface | Implementations | Factory |
|-----------|-----------------|---------|
| `DataStore` | SQLDataStore, MockDataStore | `NewSQLDataStore()` |
| `Driver` | SQLiteDriver | `NewDriver()` |

---

## Container Runtime Layer

**Auto-detects platform, creates appropriate runtime**

```
  cmd/attach.go                 operators/
 ┌──────────────────┐    ┌────────────────────────────────┐
 │ runtime, _ :=    │    │  NewContainerRuntime()         │
 │   NewContainer() │───►│                                │
 │                  │    │  PlatformDetector → selects:   │
 └──────────────────┘    └────────────┬───────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
             ┌────────────┐    ┌────────────┐    ┌────────────┐
             │   Docker   │    │ Containerd │    │    Mock    │
             │  Runtime   │    │  Runtime   │    │  Runtime   │
             │            │    │            │    │            │
             │ OrbStack,  │    │  Colima    │    │  Testing   │
             │ Desktop    │    │            │    │            │
             └────────────┘    └────────────┘    └────────────┘
```

| Interface | Implementations | Factory |
|-----------|-----------------|---------|
| `ContainerRuntime` | DockerRuntime, ContainerdRuntimeV2, MockRuntime | `NewContainerRuntime()` |

---

## Image Builder Layer

**Platform determines Docker API vs BuildKit**

```
  cmd/build.go                  builders/
 ┌──────────────────┐    ┌────────────────────────────────┐
 │ builder, _ :=    │    │  NewImageBuilder(config)       │
 │   NewImageBuilder│───►│                                │
 │                  │    │  config.Platform → selects:    │
 └──────────────────┘    └────────────┬───────────────────┘
                                      │
                         ┌────────────┴────────────┐
                         ▼                         ▼
                  ┌────────────┐            ┌────────────┐
                  │   Docker   │            │  BuildKit  │
                  │  Builder   │            │  Builder   │
                  │            │            │            │
                  │ Docker API │            │ gRPC API   │
                  └────────────┘            └────────────┘
```

| Interface | Implementations | Factory |
|-----------|-----------------|---------|
| `ImageBuilder` | DockerBuilder, BuildKitBuilder | `NewImageBuilder()` |

---

## NvimOps Layer

**Standalone package for Neovim plugin management**

```
  cmd/nvp/                      pkg/nvimops/
 ┌──────────────────┐    ┌────────────────────────────────┐
 │ nvp get plugins  │───►│  Manager                       │
 │ nvp create plugin│    │    └── uses PluginStore        │
 └──────────────────┘    └────────────┬───────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
             ┌────────────┐    ┌────────────┐    ┌────────────┐
             │  Memory    │    │   File     │    │ DBStore    │
             │  Store     │    │   Store    │    │ Adapter    │
             │ (testing)  │    │ (YAML)     │    │ (SQLite)   │
             └────────────┘    └────────────┘    └────────────┘
```

| Interface | Implementations | Location |
|-----------|-----------------|----------|
| `PluginStore` | MemoryStore, FileStore, DBStoreAdapter | `pkg/nvimops/store/` |

---

## Quick Reference

| Package | Interface | Factory | Mock |
|---------|-----------|---------|------|
| `render/` | `Renderer` | `registry.Get()` | - |
| `db/` | `DataStore` | `NewSQLDataStore()` | `MockDataStore` |
| `db/` | `Driver` | `NewDriver()` | `MemoryDriver` |
| `operators/` | `ContainerRuntime` | `NewContainerRuntime()` | `MockRuntime` |
| `builders/` | `ImageBuilder` | `NewImageBuilder()` | - |
| `pkg/nvimops/store/` | `PluginStore` | per-type constructor | `MemoryStore` |

---

## Key Files

```
render/
├── interface.go          # Renderer interface
├── registry.go           # Output(), Msg() helpers
└── renderer_*.go         # Implementations

db/
├── datastore.go          # DataStore interface
├── interfaces.go         # Driver interface
├── store.go              # SQLDataStore impl
└── sqlite_driver.go      # SQLiteDriver impl

operators/
├── runtime_interface.go  # ContainerRuntime interface
├── runtime_factory.go    # NewContainerRuntime()
├── docker_runtime.go     # DockerRuntime impl
└── containerd_runtime_v2.go

builders/
├── interfaces.go         # ImageBuilder interface
├── factory.go            # NewImageBuilder()
├── docker_builder.go     # DockerBuilder impl
└── buildkit_builder.go   # BuildKitBuilder impl

pkg/nvimops/store/
├── interface.go          # PluginStore interface
├── memory.go             # MemoryStore impl
├── file.go               # FileStore impl
└── db_adapter.go         # DBStoreAdapter impl
```

---

## Anti-Patterns

```go
// ❌ Don't accept concrete types
func process(store *SQLDataStore) { ... }

// ✅ Accept interfaces
func process(store DataStore) { ... }

// ❌ Don't return concrete types from factories
func NewStore() *SQLDataStore { ... }

// ✅ Return interfaces
func NewStore() DataStore { ... }
```
