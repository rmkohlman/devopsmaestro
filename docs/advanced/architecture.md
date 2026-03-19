# Architecture

Internal architecture of DevOpsMaestro.

---

## Overview

```
dvm/nvp/dvt CLI
    |
    ├── cmd/              # CLI commands (dvm, nvp, dvt entry points)
    ├── models/           # Data models (all resource types)
    ├── db/               # SQLite database layer
    │   └── migrations/   # Database migrations (001-017)
    ├── config/           # Configuration, vault integration, credentials
    ├── pkg/
    │   ├── resource/     # Unified resource interface & handlers
    │   │   └── handlers/ # 12 resource type handlers
    │   ├── colorbridge/  # Bridge: MaestroTheme → dvm color system
    │   ├── nvimbridge/   # Bridge: MaestroNvim → dvm DataStore
    │   ├── themebridge/  # Bridge: MaestroTheme → dvm DataStore
    │   ├── terminalbridge/ # Bridge: MaestroTerminal → dvm DataStore
    │   ├── source/       # Source resolution (file, URL, stdin, GitHub)
    │   ├── registry/     # OCI registry management (Zot)
    │   ├── mirror/       # Git bare repo mirror management
    │   ├── secrets/      # Secret provider abstraction
    │   ├── crd/          # Custom Resource Definition support
    │   ├── resolver/     # Theme/config resolution
    │   ├── preflight/    # Pre-build validation
    │   ├── workspace/    # Workspace operations
    │   └── client/       # External API clients
    ├── operators/        # Container runtime abstraction
    ├── builders/         # Image building (Docker, BuildKit, nerdctl)
    └── utils/            # Shared utilities
```

---

## External Modules

DevOpsMaestro consumes several extracted modules as Go dependencies:

| Module | Package | Purpose |
|--------|---------|---------|
| `github.com/rmkohlman/MaestroPalette` v0.1.0 | Color palette primitives | Generic color palette types |
| `github.com/rmkohlman/MaestroSDK` v0.1.0 | paths/, resource/, colors/, render/ | Shared foundation packages |
| `github.com/rmkohlman/MaestroNvim` v0.2.0 | nvimops/ | Neovim plugin/theme management |
| `github.com/rmkohlman/MaestroTheme` v0.1.0 | theme/ | Theme system |
| `github.com/rmkohlman/MaestroTerminal` v0.1.0 | terminalops/ | Terminal prompt/package management |

### Bridge Pattern

Each extracted module has a bridge package in dvm (`pkg/*bridge/`) that adapts the module's interfaces to dvm's DataStore. This allows modules to be independently testable while sharing dvm's SQLite database at runtime.

```
MaestroNvim (nvimops/)  ──▶  pkg/nvimbridge/  ──▶  DataStore (SQLite)
MaestroTheme (theme/)   ──▶  pkg/themebridge/  ──▶  DataStore (SQLite)
MaestroTerminal         ──▶  pkg/terminalbridge/ ──▶  DataStore (SQLite)
```

---

## Resource Types

12 registered resource handlers:

| Handler | Kind |
|---------|------|
| Ecosystem | Top-level organizational unit |
| Domain | Bounded context within an ecosystem |
| App | Application (maps to a code repository) |
| Workspace | Development environment for an app |
| Credential | Stored credentials for registries or repos |
| Registry | OCI registry configuration (Zot) |
| GitRepo | Git bare repo mirror |
| NvimPlugin | Neovim plugin definition |
| NvimTheme | Neovim color theme |
| NvimPackage | Curated plugin package |
| TerminalPrompt | Terminal prompt configuration |
| TerminalPackage | Terminal extension package |

CRD support via `CustomResourceDefinition` allows user-extensible resource types.

---

## Core Packages

### db/

Database layer:

- `DataStore` interface
- SQLite implementation at `~/.devopsmaestro/devopsmaestro.db`
- 17 migrations (001-017)

### pkg/source/

Source resolution for the `-f` flag:

- `Source` interface
- FileSource, URLSource, StdinSource, GitHubSource
- Automatic type detection

### pkg/resource/

Unified resource handling:

- `Resource` interface (GetKind, GetName, Validate)
- `Handler` interface (Apply, Get, List, Delete, ToYAML)
- Registry pattern for handler lookup

### operators/

Container runtime abstraction:

- `ContainerRuntime` interface
- Platform detection (OrbStack, Docker, Podman, Colima, containerd/nerdctl)
- Container lifecycle management

### builders/

Image building:

- `ImageBuilder` interface
- DockerBuilder, BuildKitBuilder, NerdctlBuilder
- Dockerfile generation

---

## Design Principles

### 1. Decoupling

Interface → Implementation → Factory pattern:

```go
// Interface
type DataStore interface {
    CreateApp(app *models.App) error
    GetAppByName(name string) (*models.App, error)
    // ...
}

// Implementation
type SQLDataStore struct {
    db *sql.DB
}

// Factory
func NewSQLDataStore(path string) (*SQLDataStore, error) {
    // ...
}
```

### 2. kubectl Patterns

Familiar command structure:

```
dvm get apps
dvm create app <name>
dvm delete workspace <name>
dvm apply -f config.yaml
```

### 3. Separation of Concerns

- Commands handle CLI interaction
- Render packages handle output formatting
- Stores handle persistence
- Handlers handle resource operations
- Bridge packages adapt external modules to dvm's DataStore

### 4. Testability

Mock implementations for all interfaces:

```go
type MockDataStore struct {
    apps map[string]*models.App
}

func (m *MockDataStore) CreateApp(a *models.App) error {
    m.apps[a.Name] = a
    return nil
}
```

---

## Resource Pipeline

How `apply` works:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Source    │────▶│   Parse     │────▶│   Handler   │
│  (resolve)  │     │   (YAML)    │     │   (apply)   │
└─────────────┘     └─────────────┘     └─────────────┘
     │                    │                    │
     ▼                    ▼                    ▼
 File/URL/            Detect             Save to
 Stdin/GitHub         Kind               store
```

1. **Source** resolves the input (file, URL, stdin, GitHub shorthand)
2. **Parse** reads YAML and detects the resource Kind
3. **Handler** applies the resource (validates, saves to store)

---

## Database

SQLite at `~/.devopsmaestro/devopsmaestro.db` with 17 migrations.

Tables:

| Table | Purpose |
|-------|---------|
| ecosystems | Top-level organizational units |
| domains | Bounded contexts |
| apps | Application registrations |
| workspaces | Development environments |
| nvim_plugins | Neovim plugin definitions |
| nvim_themes | Neovim color themes |
| nvim_packages | Curated plugin packages |
| terminal_prompts | Terminal prompt configurations |
| terminal_packages | Terminal extension packages |
| terminal_emulators | Terminal emulator settings |
| terminal_profiles | Terminal profile configurations |
| credentials | Stored credentials |
| registries | OCI registry configurations |
| git_repos | Git bare repo mirrors |
| crds | Custom Resource Definitions |
| crd_instances | CRD resource instances |
| context | Active context (current app/workspace) |

---

## Container Runtime Detection

Priority order:

1. `DVM_PLATFORM` environment variable
2. OrbStack (if installed)
3. Docker Desktop
4. Colima
5. Podman
6. containerd (nerdctl)

Detection checks socket paths:

```go
var orbstackSockets = []string{
    "~/.orbstack/run/docker.sock",
    "/var/run/docker.sock",
}
```

---

## Next Steps

- [Source Types](source-types.md) - Source resolution details
- [Build Architecture](../build-architecture.md) - Build pipeline internals
- [Contributing](../development/contributing.md) - Contribute to development
