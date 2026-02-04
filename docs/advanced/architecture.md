# Architecture

Internal architecture of DevOpsMaestro.

---

## Overview

```
dvm/nvp CLI
    │
    ├── render/          # Decoupled output formatting
    ├── db/              # SQLite database layer (dvm)
    ├── pkg/source/      # Source resolution (file, URL, stdin, GitHub)
    ├── pkg/resource/    # Unified resource interface & handlers
    │   └── handlers/    # NvimPlugin, NvimTheme handlers
    ├── pkg/nvimops/     # Plugin/theme management (nvp)
    │   ├── plugin/      # Plugin types, parser, generator
    │   ├── theme/       # Theme types, parser, generator
    │   ├── store/       # Storage interfaces
    │   └── library/     # Embedded plugin/theme library
    ├── operators/       # Container runtime abstraction
    └── builders/        # Image building (Docker, BuildKit)
```

---

## Core Packages

### render/

Decoupled output formatting:

- `Renderer` interface with multiple implementations
- JSON, YAML, Table, Colored, Plain output
- Commands prepare data, renderers display it

### db/

Database layer (dvm only):

- `DataStore` interface
- SQLite implementation
- Projects, workspaces, plugins storage

### pkg/source/

Source resolution for `-f` flag:

- `Source` interface
- FileSource, URLSource, StdinSource, GitHubSource
- Automatic type detection

### pkg/resource/

Unified resource handling:

- `Resource` interface (GetKind, GetName, Validate)
- `Handler` interface (Apply, Get, List, Delete, ToYAML)
- Registry pattern for handler lookup

### pkg/nvimops/

Neovim plugin/theme management:

- Plugin types, parsing, Lua generation
- Theme types, parsing, palette export
- Store interfaces (File, Memory, DB adapters)

### operators/

Container runtime abstraction:

- `ContainerRuntime` interface
- Platform detection (OrbStack, Docker, Podman, Colima)
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
    CreateProject(project *Project) error
    GetProjectByName(name string) (*Project, error)
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
dvm get projects
dvm create project <name>
dvm delete workspace <name>
dvm apply -f config.yaml
```

### 3. Separation of Concerns

- Commands handle CLI interaction
- Render package handles output formatting
- Stores handle persistence
- Handlers handle resource operations

### 4. Testability

Mock implementations for all interfaces:

```go
type MockDataStore struct {
    projects map[string]*Project
}

func (m *MockDataStore) CreateProject(p *Project) error {
    m.projects[p.Name] = p
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

## Database Schema

```sql
-- Projects
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- Workspaces
CREATE TABLE workspaces (
    id INTEGER PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id),
    name TEXT NOT NULL,
    image_name TEXT,
    status TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- Plugins
CREATE TABLE plugins (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    repo TEXT NOT NULL,
    category TEXT,
    config TEXT,  -- JSON
    created_at DATETIME,
    updated_at DATETIME
);
```

---

## Container Runtime Detection

Priority order:

1. `DVM_PLATFORM` environment variable
2. OrbStack (if installed)
3. Docker Desktop
4. Colima
5. Podman

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
- [Contributing](../development/contributing.md) - Contribute to development
