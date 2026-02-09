# DevOpsMaestro Development Standards

> **Purpose:** This document defines the architectural principles, design patterns, and coding
> standards that guide all development on DevOpsMaestro. AI assistants and human contributors
> should follow these standards when adding new features or modifying existing code.

---

## Design Philosophy

**We build loosely coupled, decoupled, modular, cohesive code with responsibility segregation.**

This means:
- **Loosely Coupled**: Components interact through interfaces, not concrete implementations
- **Decoupled**: Changes in one module don't cascade to others
- **Modular**: Each piece can be developed, tested, and replaced independently
- **Cohesive**: Related functionality is grouped together
- **Responsibility Segregation**: Each component has a single, clear purpose

**Why this matters:**
- Testability: Mock any dependency for unit tests
- Flexibility: Swap implementations without changing consumers
- Maintainability: Fix bugs in isolation
- Extensibility: Add new features without breaking existing code

---

## Core Principles

### 1. DECOUPLING IS PARAMOUNT

**Everything must be modular and swappable.**

This enables:
- Easy unit testing with mocks
- Future extensibility (new databases, runtimes, output formats)
- Gradual migration without breaking changes
- Different deployment scenarios

**Pattern:** Interface -> Implementation -> Factory

```go
// 1. Define the interface (contract)
type DataStore interface {
    CreateApp(app *models.App) error
    GetAppByName(name string) (*models.App, error)
}

// 2. Create implementation
type SQLDataStore struct {
    driver Driver
}

func (s *SQLDataStore) CreateApp(app *models.App) error {
    // Implementation
}

// 3. Factory creates the implementation
func CreateDataStore() (DataStore, error) {
    driver, err := DriverFactory()
    if err != nil {
        return nil, err
    }
    return NewSQLDataStore(driver, nil), nil
}
```

### 2. SINGLE RESPONSIBILITY PRINCIPLE

Each component should have one reason to change.

| Component | Responsibility |
|-----------|---------------|
| `DataStore` | Data persistence operations |
| `ContainerRuntime` | Container lifecycle management |
| `ImageBuilder` | Building container images |
| `PlatformDetector` | Detecting available platforms |
| `Formatter` | Output formatting |

**Anti-pattern:** A function that detects platform AND builds images AND formats output.

### 3. LOOSE COUPLING

Components should depend on interfaces, not implementations.

```go
// GOOD: Depends on interface
func BuildWorkspace(builder ImageBuilder, store DataStore) error {
    // Can use any builder/store implementation
}

// BAD: Depends on concrete type
func BuildWorkspace(builder *DockerBuilder, store *SQLDataStore) error {
    // Tightly coupled
}
```

### 4. DEPENDENCY INJECTION

Dependencies should be injected, not created internally.

```go
// GOOD: Get dependency from context (injected by parent)
func (cmd *cobra.Command) RunE(cmd *cobra.Command, args []string) error {
    datastore, err := getDataStore(cmd)  // From context
    if err != nil {
        return err
    }
    return datastore.DeletePlugin(args[0])
}

// BAD: Create dependency internally (hard to test, tightly coupled)
func (cmd *cobra.Command) RunE(cmd *cobra.Command, args []string) error {
    database, _ := db.InitializeDBConnection()  // Creates its own connection
    defer database.Close()
    datastore, _ := db.StoreFactory(database)   // Creates its own store
    return datastore.DeletePlugin(args[0])
}
```

**Why injection matters:**
- Tests can inject mocks instead of real databases
- Parent can control lifecycle (connection pooling, transactions)
- Easier to swap implementations

### 5. ADAPTER PATTERN FOR BRIDGING

When migrating from old to new architecture, use adapters to bridge interfaces.

```go
// Adapter wraps legacy interface to implement new interface
type DatabaseDriverAdapter struct {
    db         Database      // Legacy interface
    driverType DriverType
}

// Implements new Driver interface by delegating to legacy Database
func (d *DatabaseDriverAdapter) Execute(query string, args ...interface{}) (Result, error) {
    result, err := d.db.Execute(query, args...)
    return &ResultAdapter{result}, err
}
```

### 6. FACTORY PATTERN FOR CREATION

Use factories to:
- Abstract creation logic
- Enable configuration-based creation
- Support dependency injection
- Facilitate testing with mocks

```go
// Factory interface allows swapping creation strategies
type DataStoreFactory interface {
    Create() (DataStore, error)
}

// Default factory uses viper config
type DefaultDataStoreFactory struct{}

// Mock factory for testing
type MockDataStoreFactory struct {
    store DataStore
}
```

### 7. RESOURCE/HANDLER PATTERN FOR CLI OPERATIONS

**ALL CLI commands that perform CRUD operations on resources MUST use the Resource/Handler pattern.**

This is the unified kubectl-style pattern for managing resources. It provides:
- Consistent API across all resource types
- Decoupled rendering (JSON, YAML, table output)
- YAML-based apply pipeline (`dvm apply -f resource.yaml`)
- Easy addition of new resource types

**Architecture:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    CLI Commands (cmd/)                          │
│  dvm get apps, dvm create ecosystem, dvm apply -f file.yaml    │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                 Resource Registry (pkg/resource/)               │
│  resource.Get(), resource.List(), resource.Apply()              │
│  Routes to appropriate handler by Kind                          │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│              Resource Handlers (pkg/resource/handlers/)         │
│  EcosystemHandler, DomainHandler, AppHandler, NvimPluginHandler │
│  Each implements: Apply, Get, List, Delete, ToYAML              │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DataStore (db/)                              │
│  Actual database operations                                     │
└─────────────────────────────────────────────────────────────────┘
```

**Resource Interface (`pkg/resource/resource.go`):**

```go
// Resource represents any manageable resource
type Resource interface {
    GetKind() string   // e.g., "Ecosystem", "App", "NvimPlugin"
    GetName() string   // Unique name
    Validate() error   // Validation logic
}

// Handler knows how to manage a specific resource type
type Handler interface {
    Kind() string
    Apply(ctx Context, data []byte) (Resource, error)
    Get(ctx Context, name string) (Resource, error)
    List(ctx Context) ([]Resource, error)
    Delete(ctx Context, name string) error
    ToYAML(res Resource) ([]byte, error)
}
```

**Adding a New Resource Type:**

1. **Create Handler** (`pkg/resource/handlers/myresource.go`):
```go
const KindMyResource = "MyResource"

type MyResourceHandler struct{}

func NewMyResourceHandler() *MyResourceHandler {
    return &MyResourceHandler{}
}

func (h *MyResourceHandler) Kind() string { return KindMyResource }

func (h *MyResourceHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
    // Parse YAML, upsert to database
}

func (h *MyResourceHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
    // Get from database, wrap in Resource
}

func (h *MyResourceHandler) List(ctx resource.Context) ([]resource.Resource, error) {
    // List from database
}

func (h *MyResourceHandler) Delete(ctx resource.Context, name string) error {
    // Delete from database
}

func (h *MyResourceHandler) ToYAML(res resource.Resource) ([]byte, error) {
    // Serialize to YAML
}

// Resource wrapper
type MyResourceResource struct {
    model *models.MyResource
}

func (r *MyResourceResource) GetKind() string { return KindMyResource }
func (r *MyResourceResource) GetName() string { return r.model.Name }
func (r *MyResourceResource) Validate() error { /* validation */ }
```

2. **Register Handler** (`pkg/resource/handlers/register.go`):
```go
func RegisterAll() {
    registerOnce.Do(func() {
        resource.Register(NewMyResourceHandler())
        // ... other handlers
    })
}
```

3. **Use in CLI Commands**:
```go
// In cmd/myresource.go
func getMyResources(cmd *cobra.Command) error {
    ctx, err := buildResourceContext(cmd)
    if err != nil {
        return err
    }

    // Use the unified resource API
    resources, err := resource.List(ctx, handlers.KindMyResource)
    if err != nil {
        return err
    }

    // Render output (decoupled)
    return render.OutputWith(getOutputFormat, resources, render.Options{
        Type: render.TypeTable,
    })
}
```

**Current Resource Types:**

| Kind | Handler | Status |
|------|---------|--------|
| `Ecosystem` | `EcosystemHandler` | ✅ Complete |
| `Domain` | `DomainHandler` | ✅ Complete |
| `App` | `AppHandler` | ✅ Complete |
| `NvimPlugin` | `NvimPluginHandler` | ✅ Complete |
| `NvimTheme` | `NvimThemeHandler` | ✅ Complete |
| `Project` | — | ⚠️ Needs migration (deprecated) |
| `Workspace` | — | ⚠️ Needs migration |

---

## Design Pattern Checklist

Before implementing a new feature, ask these questions:

### Interface Design
- [ ] Is there an interface defined for this component?
- [ ] Does the interface follow the single responsibility principle?
- [ ] Can this interface be easily mocked for testing?

### Implementation
- [ ] Does the implementation only implement the interface methods?
- [ ] Are dependencies injected via constructor, not created internally?
- [ ] Is the implementation package-private if possible?

### Factory
- [ ] Is there a factory function to create instances?
- [ ] Does the factory handle configuration/environment?
- [ ] Can the factory be swapped for testing?

### Testing
- [ ] Is there a mock implementation of the interface?
- [ ] Does the mock support call recording?
- [ ] Does the mock support error injection?
- [ ] Are there interface compliance tests?

---

## Code Organization

### Package Structure

```
package/
├── interfaces.go      # Interface definitions
├── implementation.go  # Concrete implementations
├── factory.go         # Factory functions
├── mock_*.go          # Mock implementations
├── *_test.go          # Tests
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Interface | Noun (capability) | `DataStore`, `Driver`, `Builder` |
| Implementation | Adjective + Interface | `SQLDataStore`, `DockerBuilder` |
| Mock | `Mock` + Interface | `MockDataStore`, `MockDriver` |
| Factory | `New` + Type or `Create` + Type | `NewDataStore()`, `CreateDriver()` |
| Adapter | Type + `Adapter` | `DatabaseDriverAdapter` |

---

## Testing Standards

### Every Interface Needs a Mock

```go
// MockDataStore implements DataStore for testing
type MockDataStore struct {
    // Storage for test data
    Apps       map[string]*models.App
    Workspaces map[int]map[string]*models.Workspace
    
    // Error injection
    ErrorToReturn error
    
    // Call recording
    Calls []string
}

func NewMockDataStore() *MockDataStore {
    return &MockDataStore{
        Apps:       make(map[string]*models.App),
        Workspaces: make(map[int]map[string]*models.Workspace),
    }
}

func (m *MockDataStore) GetAppByName(name string) (*models.App, error) {
    m.Calls = append(m.Calls, "GetAppByName")
    if m.ErrorToReturn != nil {
        return nil, m.ErrorToReturn
    }
    return m.Apps[name], nil
}
```

### Interface Compliance Tests

```go
// Verify implementations satisfy interface
func TestSQLDataStore_ImplementsDataStore(t *testing.T) {
    var _ DataStore = (*SQLDataStore)(nil)
}

func TestMockDataStore_ImplementsDataStore(t *testing.T) {
    var _ DataStore = (*MockDataStore)(nil)
}
```

### Swappability Tests

```go
func TestDataStoreSwappability(t *testing.T) {
    implementations := []DataStore{
        createRealStore(t),
        NewMockDataStore(),
    }
    
    for _, store := range implementations {
        // Same tests should pass for all implementations
        testAppOperations(t, store)
    }
}
```

---

## Flexibility Guidelines

### Adding New Features

1. **Define interface first** - What operations does this feature need?
2. **Create minimal implementation** - Just enough to work
3. **Create mock for testing** - Enable unit tests
4. **Add factory function** - Abstract creation
5. **Write tests** - Unit and integration

### Modifying Existing Code

1. **Don't break existing interfaces** - Add new methods, don't change signatures
2. **Use adapters for migration** - Bridge old and new code
3. **Maintain backward compatibility** - Legacy code paths should keep working
4. **Add deprecation notices** - Mark old code with migration path

### Supporting New Platforms/Backends

The architecture should make it easy to add:
- New database backends (PostgreSQL, MySQL)
- New container runtimes (Kubernetes, containerd)
- New build systems (BuildKit, Kaniko)
- New output formats (JSON, YAML, table)

```go
// Example: Adding a new database driver
// 1. Implement Driver interface
type PostgresDriver struct { ... }

// 2. Register with factory
func init() {
    RegisterDriver(DriverTypePostgres, NewPostgresDriver)
}

// 3. Users can now use it via config
// config.yaml: database.driver: postgres
```

---

## Error Handling

### Always Return Errors

```go
// GOOD
func (s *SQLDataStore) GetApp(id int) (*models.App, error) {
    if id <= 0 {
        return nil, fmt.Errorf("invalid app ID: %d", id)
    }
    // ...
}

// BAD - panics
func (s *SQLDataStore) GetApp(id int) *models.App {
    if id <= 0 {
        panic("invalid app ID")
    }
    // ...
}
```

### Wrap Errors with Context

```go
if err := driver.Connect(); err != nil {
    return fmt.Errorf("failed to connect to database: %w", err)
}
```

---

## Documentation

### Public API Documentation

```go
// DataStore provides high-level operations for managing ecosystems, domains,
// apps, workspaces, and plugins. It abstracts the underlying database
// driver and provides a clean interface for the application layer.
//
// Implementations must be safe for concurrent use.
type DataStore interface {
    // CreateApp creates a new app.
    // Returns an error if an app with the same name already exists.
    CreateApp(app *models.App) error
    
    // GetAppByName retrieves an app by name.
    // Returns nil, nil if the app does not exist.
    GetAppByName(name string) (*models.App, error)
}
```

### Update These Files When Making Changes

| Change Type | Files to Update |
|-------------|-----------------|
| New interface | `interfaces.go`, mock file, factory, STANDARDS.md if pattern |
| New database table | `migrations/`, `store.go`, `interfaces.go` |
| New command | `cmd/`, MANUAL_TEST_PLAN.md |
| New platform | `operators/platform.go`, `runtime_factory.go` |
| Bug fix | Test file (add regression test) |
| Architecture change | CLAUDE.md, STANDARDS.md |

---

## Quick Reference

### Before Writing Code

1. Read CLAUDE.md for architecture overview
2. Check STANDARDS.md (this file) for patterns
3. Check if similar code exists to follow patterns

### During Development

1. Follow interface -> implementation -> factory pattern
2. Create mock for testing
3. Write tests alongside code
4. Keep functions small and focused

### After Writing Code

1. Run `go test ./...`
2. Run `go fmt ./...`
3. Update documentation if needed
4. Update MANUAL_TEST_PLAN.md if adding user-facing features

### Code Review Checklist

- [ ] Follows decoupling principles
- [ ] Has interface defined
- [ ] Has mock for testing
- [ ] Has factory if applicable
- [ ] Has tests (unit + integration where needed)
- [ ] Error handling is proper
- [ ] Documentation is updated

---

## References

- [CLAUDE.md](./CLAUDE.md) - Full architecture documentation
- [MANUAL_TEST_PLAN.md](./MANUAL_TEST_PLAN.md) - Testing procedures
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
