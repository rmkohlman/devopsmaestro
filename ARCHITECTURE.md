# DevOpsMaestro Architecture

> **Purpose:** Quick reference for understanding the decoupled architecture patterns used throughout DevOpsMaestro. Read this before writing any new code.

---

## Core Design Pattern

**Every major subsystem follows the same pattern:**

```
┌─────────────────────────────────────────────────────────────────┐
│                         INTERFACE                                │
│              (defines the contract/behavior)                     │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Implementation A│ │ Implementation B│ │   Mock (test)   │
│  (e.g. SQLite)  │ │ (e.g. Postgres) │ │                 │
└─────────────────┘ └─────────────────┘ └─────────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FACTORY FUNCTION                              │
│     NewXxx(config) → returns Interface (not concrete type)      │
└─────────────────────────────────────────────────────────────────┘
```

**Key Rules:**
1. Functions accept **interfaces**, not concrete types
2. Factory functions return **interfaces**, not structs
3. Every interface has a **mock implementation** for testing
4. Configuration determines which implementation is created

---

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│                         cmd/*.go                                 │
│         Commands: create, get, use, build, attach, detach       │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  render/        │  │  db/            │  │  operators/     │
│  Renderer       │  │  DataStore      │  │  ContainerRun-  │
│  interface      │  │  interface      │  │  time interface │
└─────────────────┘  └─────────────────┘  └─────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  builders/      │  │  pkg/nvimops/   │  │  models/        │
│  ImageBuilder   │  │  PluginStore    │  │  Data structs   │
│  interface      │  │  interface      │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## Package: `render/`

**Purpose:** Decoupled output rendering. Commands prepare data, renderers decide how to display.

### Architecture Diagram

```
                         ┌─────────────────────────────┐
                         │     Command Layer (cmd/)    │
                         │                             │
                         │  // Prepare structured data │
                         │  data := render.TableData{} │
                         │                             │
                         │  // Call render - agnostic  │
                         │  render.OutputWith(format,  │
                         │    data, opts)              │
                         └──────────────┬──────────────┘
                                        │
                                        ▼
┌───────────────────────────────────────────────────────────────────┐
│                        render.OutputWith()                         │
│                                                                    │
│  1. Resolve renderer: flag (-o json) > env (DVM_RENDER) > default │
│  2. Get renderer from registry                                     │
│  3. Call renderer.Render(os.Stdout, data, opts)                   │
└───────────────────────────────────────┬───────────────────────────┘
                                        │
          ┌─────────────┬───────────────┼───────────────┬─────────────┐
          ▼             ▼               ▼               ▼             ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ JSONRenderer │ │ YAMLRenderer │ │ColoredRender │ │ PlainRender  │ │ TableRender  │
│              │ │              │ │   (default)  │ │  (no color)  │ │              │
│ json.Marshal │ │ yaml.Marshal │ │ icons+color  │ │  plain text  │ │  tabwriter   │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
          │             │               │               │             │
          └─────────────┴───────────────┴───────────────┴─────────────┘
                                        │
                                        ▼
                               ┌─────────────────┐
                               │    io.Writer    │
                               │   (os.Stdout)   │
                               └─────────────────┘
```

### Interface

```go
// render/interface.go
type Renderer interface {
    Render(w io.Writer, data any, opts Options) error
    RenderMessage(w io.Writer, msg Message) error
    Name() RendererName
    SupportsColor() bool
}
```

### Data Types

```go
type TableData struct {
    Headers []string
    Rows    [][]string
}

type KeyValueData struct {
    Pairs []KeyValue
}

type ListData struct {
    Items []string
}
```

### Implementations

| Renderer | File | Description |
|----------|------|-------------|
| `ColoredRenderer` | `renderer_colored.go` | Default, with icons and ANSI colors |
| `PlainRenderer` | `renderer_plain.go` | No colors (for piping) |
| `JSONRenderer` | `renderer_json.go` | Machine-readable JSON |
| `YAMLRenderer` | `renderer_yaml.go` | Machine-readable YAML |
| `TableRenderer` | `renderer_table.go` | Explicit table format |
| `CompactRenderer` | `renderer_table.go` | Compact single-line format |

### Usage Pattern

```go
// In a command - prepare data, let renderer decide display
data := render.TableData{
    Headers: []string{"NAME", "STATUS"},
    Rows:    [][]string{{"dev", "running"}, {"prod", "stopped"}},
}
render.OutputWith(outputFormat, data, render.Options{
    Type:  render.TypeTable,
    Title: "Workspaces",
})

// Messages
render.Success("Operation completed")
render.Warning("Check your config")
render.Error("Something failed")
render.Info("Hint: use --help")
render.Progress("Building image...")
```

---

## Package: `db/`

**Purpose:** Database abstraction with two layers - low-level Driver and high-level DataStore.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      Application Code                            │
│                                                                  │
│   store.CreateProject(project)                                  │
│   projects, _ := store.ListProjects()                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DataStore Interface                         │
│                      (High-level API)                            │
│                                                                  │
│   CreateProject, GetProjectByName, ListProjects                 │
│   CreateWorkspace, GetWorkspaceByName, ListWorkspaces           │
│   CreatePlugin, ListPlugins, GetContext, SetActiveProject       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
              ┌────────────────┴────────────────┐
              ▼                                 ▼
┌──────────────────────┐            ┌──────────────────────┐
│    SQLDataStore      │            │    MockDataStore     │
│    (Production)      │            │    (Testing)         │
│                      │            │                      │
│  store/store.go      │            │  mock_store.go       │
└──────────┬───────────┘            └──────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Driver Interface                           │
│                       (Low-level API)                            │
│                                                                  │
│   Connect, Close, Ping                                          │
│   Execute, Query, QueryRow                                      │
│   Begin (transactions)                                          │
└──────────────────────────────┬──────────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                    ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│  SQLiteDriver    │ │ PostgresDriver   │ │  MemoryDriver    │
│                  │ │   (future)       │ │   (testing)      │
│ sqlite_driver.go │ │                  │ │                  │
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

### Interfaces

```go
// db/datastore.go - High-level business operations
type DataStore interface {
    // Projects
    CreateProject(project *models.Project) error
    GetProjectByName(name string) (*models.Project, error)
    ListProjects() ([]*models.Project, error)
    
    // Workspaces
    CreateWorkspace(workspace *models.Workspace) error
    GetWorkspaceByName(projectID int, name string) (*models.Workspace, error)
    ListWorkspacesByProject(projectID int) ([]*models.Workspace, error)
    
    // Context
    GetContext() (*models.Context, error)
    SetActiveProject(projectID *int) error
    SetActiveWorkspace(workspaceID *int) error
    
    // Plugins
    CreatePlugin(plugin *models.NvimPluginDB) error
    ListPlugins() ([]*models.NvimPluginDB, error)
    
    // Lifecycle
    Driver() Driver
    Close() error
    Ping() error
}

// db/interfaces.go - Low-level database operations
type Driver interface {
    Connect() error
    Close() error
    Ping() error
    Execute(query string, args ...interface{}) (Result, error)
    QueryRow(query string, args ...interface{}) Row
    Query(query string, args ...interface{}) (Rows, error)
    Begin() (Transaction, error)
    Type() DriverType
}
```

### Factory Pattern

```go
// Create driver based on config
driver, err := db.NewDriver(db.DriverConfig{
    Type: db.DriverSQLite,
    Path: "~/.devopsmaestro/devopsmaestro.db",
})

// Create store using driver
store := db.NewSQLDataStore(driver)
```

---

## Package: `operators/`

**Purpose:** Container runtime abstraction. Same interface for Docker, containerd, Kubernetes.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/attach.go                            │
│                         cmd/build.go                             │
│                                                                  │
│   runtime, _ := operators.NewContainerRuntime()                 │
│   runtime.StartWorkspace(ctx, opts)                             │
│   runtime.AttachToWorkspace(ctx, containerID)                   │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                   ContainerRuntime Interface                     │
│                                                                  │
│   BuildImage(ctx, opts) error                                   │
│   StartWorkspace(ctx, opts) (string, error)                     │
│   AttachToWorkspace(ctx, workspaceID) error                     │
│   StopWorkspace(ctx, workspaceID) error                         │
│   GetWorkspaceStatus(ctx, workspaceID) (string, error)          │
│   GetRuntimeType() string                                       │
└──────────────────────────────┬──────────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                    ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│  DockerRuntime   │ │ContainerdRuntime │ │  MockRuntime     │
│                  │ │      V2          │ │   (testing)      │
│ OrbStack, Docker │ │ Colima+nerdctl   │ │                  │
│ Desktop, Podman  │ │                  │ │                  │
└──────────────────┘ └──────────────────┘ └──────────────────┘
          │                    │
          └────────────────────┼────────────────────┐
                               ▼                    │
┌─────────────────────────────────────────────────────────────────┐
│                    Platform Detection                            │
│                                                                  │
│   detector := operators.NewPlatformDetector()                   │
│   platform, _ := detector.Detect()                              │
│                                                                  │
│   platform.Type     // "docker" | "containerd"                  │
│   platform.Name     // "OrbStack" | "Colima" | "Docker Desktop" │
│   platform.SocketPath                                           │
└─────────────────────────────────────────────────────────────────┘
```

### Interface

```go
// operators/runtime_interface.go
type ContainerRuntime interface {
    BuildImage(ctx context.Context, opts BuildOptions) error
    StartWorkspace(ctx context.Context, opts StartOptions) (string, error)
    AttachToWorkspace(ctx context.Context, workspaceID string) error
    StopWorkspace(ctx context.Context, workspaceID string) error
    GetWorkspaceStatus(ctx context.Context, workspaceID string) (string, error)
    GetRuntimeType() string
}
```

### Factory Pattern

```go
// Auto-detects platform and creates appropriate runtime
runtime, err := operators.NewContainerRuntime()

// Or explicitly with platform detection
detector := operators.NewPlatformDetector()
platform, _ := detector.Detect()
// platform.Type determines which runtime implementation to use
```

---

## Package: `builders/`

**Purpose:** Image building abstraction. Same interface for Docker API and BuildKit.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/build.go                             │
│                                                                  │
│   platform, _ := detectPlatform()                               │
│   builder, _ := builders.NewImageBuilder(config)                │
│   defer builder.Close()                                         │
│   builder.Build(ctx, opts)                                      │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ImageBuilder Interface                        │
│                                                                  │
│   Build(ctx, opts BuildOptions) error                           │
│   ImageExists(ctx) (bool, error)                                │
│   Close() error                                                 │
└──────────────────────────────┬──────────────────────────────────┘
                               │
              ┌────────────────┴────────────────┐
              ▼                                 ▼
┌──────────────────────┐            ┌──────────────────────┐
│    DockerBuilder     │            │   BuildKitBuilder    │
│                      │            │                      │
│  Uses Docker API     │            │  Uses BuildKit gRPC  │
│  (OrbStack, Docker   │            │  (Colima+containerd) │
│   Desktop, Podman)   │            │                      │
└──────────────────────┘            └──────────────────────┘
```

### Interface

```go
// builders/interfaces.go
type ImageBuilder interface {
    Build(ctx context.Context, opts BuildOptions) error
    ImageExists(ctx context.Context) (bool, error)
    Close() error
}
```

### Factory Pattern

```go
builder, err := builders.NewImageBuilder(builders.BuilderConfig{
    Platform:    platform,      // Determines implementation
    Namespace:   "devopsmaestro",
    ProjectPath: "/path/to/project",
    ImageName:   "myimage:latest",
    Dockerfile:  "/path/to/Dockerfile",
})
```

---

## Package: `pkg/nvimops/`

**Purpose:** Neovim plugin management. Standalone package usable without dvm.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/nvp/                                 │
│                     (standalone CLI)                             │
│                                                                  │
│   nvp get plugins                                               │
│   nvp create plugin                                             │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      nvimops.Manager                             │
│                                                                  │
│   manager := nvimops.NewManager(store)                          │
│   manager.InstallPlugin(name)                                   │
│   manager.ListPlugins()                                         │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    PluginStore Interface                         │
│                                                                  │
│   Create(p *plugin.Plugin) error                                │
│   Update(p *plugin.Plugin) error                                │
│   Get(name string) (*plugin.Plugin, error)                      │
│   List() ([]*plugin.Plugin, error)                              │
│   Delete(name string) error                                     │
│   Exists(name string) (bool, error)                             │
└──────────────────────────────┬──────────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                    ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│   MemoryStore    │ │    FileStore     │ │  DBStoreAdapter  │
│                  │ │                  │ │                  │
│  In-memory map   │ │  YAML files on   │ │  Wraps DataStore │
│  (testing)       │ │  disk            │ │  for plugins     │
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

### Interface

```go
// pkg/nvimops/store/interface.go
type PluginStore interface {
    Create(p *plugin.Plugin) error
    Update(p *plugin.Plugin) error
    Upsert(p *plugin.Plugin) error
    Delete(name string) error
    Get(name string) (*plugin.Plugin, error)
    List() ([]*plugin.Plugin, error)
    ListByCategory(category string) ([]*plugin.Plugin, error)
    ListByTag(tag string) ([]*plugin.Plugin, error)
    Exists(name string) (bool, error)
    Close() error
}
```

### Plugin Data Types

```go
// pkg/nvimops/plugin/types.go
type Plugin struct {
    Name         string
    ShortName    string
    Description  string
    Category     string
    Tags         []string
    Repo         string
    Dependencies []Dependency
    Config       map[string]any
    Keymaps      []Keymap
    Events       []string
    Priority     int
    Lazy         bool
    Enabled      bool
}
```

---

## Package: `models/`

**Purpose:** Shared data structures used across packages.

```go
// models/project.go
type Project struct {
    ID          int
    Name        string
    Path        string
    Description sql.NullString
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// models/workspace.go
type Workspace struct {
    ID          int
    ProjectID   int
    Name        string
    Description sql.NullString
    ImageName   string
    Status      string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// models/context.go
type Context struct {
    ID                int
    ActiveProjectID   sql.NullInt64
    ActiveWorkspaceID sql.NullInt64
}
```

---

## Quick Reference Table

| Package | Primary Interface | Implementations | Factory Function |
|---------|-------------------|-----------------|------------------|
| `render/` | `Renderer` | Colored, Plain, JSON, YAML, Table, Compact | `registry.Get()` |
| `db/` | `DataStore` | SQLDataStore, MockDataStore | `NewSQLDataStore()` |
| `db/` | `Driver` | SQLiteDriver, (PostgresDriver future) | `NewDriver()` |
| `operators/` | `ContainerRuntime` | DockerRuntime, ContainerdRuntimeV2, MockRuntime | `NewContainerRuntime()` |
| `builders/` | `ImageBuilder` | DockerBuilder, BuildKitBuilder | `NewImageBuilder()` |
| `pkg/nvimops/store/` | `PluginStore` | MemoryStore, FileStore, DBStoreAdapter | constructor per type |

---

## Adding New Implementations

When adding a new implementation (e.g., new database driver, new container runtime):

1. **Create the implementation struct** in a new file (e.g., `mysql_driver.go`)
2. **Implement all interface methods** - compiler will enforce this
3. **Add to factory function** - update the switch/config logic
4. **Add mock if needed** for testing
5. **Document** in this file

Example:
```go
// 1. Create struct
type MySQLDriver struct {
    db *sql.DB
    // ...
}

// 2. Implement interface
func (d *MySQLDriver) Connect() error { ... }
func (d *MySQLDriver) Query(query string, args ...interface{}) (Rows, error) { ... }
// ... all Driver methods

// 3. Update factory
func NewDriver(config DriverConfig) (Driver, error) {
    switch config.Type {
    case DriverSQLite:
        return NewSQLiteDriver(config)
    case DriverMySQL:  // NEW
        return NewMySQLDriver(config)
    // ...
    }
}
```

---

## Anti-Patterns to Avoid

❌ **Don't do this:**
```go
// Accepting concrete type
func processData(store *SQLDataStore) { ... }

// Returning concrete type
func NewStore() *SQLDataStore { ... }

// Importing implementation in command layer
import "devopsmaestro/db/sqlite"
```

✅ **Do this instead:**
```go
// Accept interface
func processData(store DataStore) { ... }

// Return interface
func NewStore(driver Driver) DataStore { ... }

// Import only the interface package
import "devopsmaestro/db"
```

---

## Testing with Mocks

Every interface has a mock for testing:

```go
// Use mock in tests
func TestMyCommand(t *testing.T) {
    mockStore := &db.MockDataStore{
        Projects: []*models.Project{
            {ID: 1, Name: "test-project"},
        },
    }
    
    // Pass mock to code under test
    err := runCommand(mockStore)
    assert.NoError(t, err)
}
```

---

## File Structure Summary

```
devopsmaestro/
├── cmd/                    # CLI commands - uses interfaces only
│   ├── attach.go
│   ├── build.go
│   ├── create.go
│   ├── detach.go
│   ├── get.go
│   └── use.go
│
├── render/                 # Output rendering (decoupled)
│   ├── interface.go        # Renderer interface + data types
│   ├── registry.go         # Global registry + Output helpers
│   ├── types.go            # RenderType, Options, Config
│   ├── renderer_colored.go # ColoredRenderer
│   ├── renderer_plain.go   # PlainRenderer
│   ├── renderer_json.go    # JSONRenderer
│   ├── renderer_yaml.go    # YAMLRenderer
│   └── renderer_table.go   # TableRenderer, CompactRenderer
│
├── db/                     # Database layer (decoupled)
│   ├── interfaces.go       # Driver, Row, Rows, Result interfaces
│   ├── datastore.go        # DataStore interface
│   ├── driver.go           # NewDriver factory
│   ├── sqlite_driver.go    # SQLiteDriver implementation
│   ├── store.go            # SQLDataStore implementation
│   ├── factory.go          # CreateDataStore factory
│   └── mock_store.go       # MockDataStore for testing
│
├── operators/              # Container runtime (decoupled)
│   ├── runtime_interface.go    # ContainerRuntime interface
│   ├── runtime_factory.go      # NewContainerRuntime factory
│   ├── docker_runtime.go       # DockerRuntime implementation
│   ├── containerd_runtime_v2.go # ContainerdRuntimeV2 implementation
│   ├── platform.go             # Platform detection
│   └── mock_runtime.go         # MockContainerRuntime for testing
│
├── builders/               # Image building (decoupled)
│   ├── interfaces.go       # ImageBuilder interface
│   ├── factory.go          # NewImageBuilder factory
│   ├── docker_builder.go   # DockerBuilder implementation
│   └── buildkit_builder.go # BuildKitBuilder implementation
│
├── pkg/nvimops/            # Neovim plugin management (standalone)
│   ├── nvimops.go          # Manager
│   ├── plugin/             # Plugin types
│   │   ├── types.go
│   │   └── interfaces.go   # LuaGenerator interface
│   ├── store/              # Plugin storage
│   │   ├── interface.go    # PluginStore interface
│   │   ├── memory.go       # MemoryStore
│   │   ├── file.go         # FileStore
│   │   └── db_adapter.go   # DBStoreAdapter
│   └── library/            # Embedded plugin definitions
│
└── models/                 # Shared data structures
    ├── project.go
    ├── workspace.go
    └── context.go
```

---

**Remember:** The decoupling pattern enables:
- Easy testing with mocks
- Swapping implementations without changing consumers
- Clear separation of concerns
- Adding new implementations without modifying existing code
