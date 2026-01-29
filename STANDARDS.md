# DevOpsMaestro Development Standards

> **Purpose:** This document defines the architectural principles, design patterns, and coding
> standards that guide all development on DevOpsMaestro. AI assistants and human contributors
> should follow these standards when adding new features or modifying existing code.

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
    CreateProject(project *models.Project) error
    GetProjectByName(name string) (*models.Project, error)
}

// 2. Create implementation
type SQLDataStore struct {
    driver Driver
}

func (s *SQLDataStore) CreateProject(project *models.Project) error {
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

### 4. ADAPTER PATTERN FOR BRIDGING

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

### 5. FACTORY PATTERN FOR CREATION

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
    Projects   map[string]*models.Project
    Workspaces map[int]map[string]*models.Workspace
    
    // Error injection
    ErrorToReturn error
    
    // Call recording
    Calls []string
}

func NewMockDataStore() *MockDataStore {
    return &MockDataStore{
        Projects:   make(map[string]*models.Project),
        Workspaces: make(map[int]map[string]*models.Workspace),
    }
}

func (m *MockDataStore) GetProjectByName(name string) (*models.Project, error) {
    m.Calls = append(m.Calls, "GetProjectByName")
    if m.ErrorToReturn != nil {
        return nil, m.ErrorToReturn
    }
    return m.Projects[name], nil
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
        testProjectOperations(t, store)
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
func (s *SQLDataStore) GetProject(id int) (*models.Project, error) {
    if id <= 0 {
        return nil, fmt.Errorf("invalid project ID: %d", id)
    }
    // ...
}

// BAD - panics
func (s *SQLDataStore) GetProject(id int) *models.Project {
    if id <= 0 {
        panic("invalid project ID")
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
// DataStore provides high-level operations for managing projects,
// workspaces, and plugins. It abstracts the underlying database
// driver and provides a clean interface for the application layer.
//
// Implementations must be safe for concurrent use.
type DataStore interface {
    // CreateProject creates a new project.
    // Returns an error if a project with the same name already exists.
    CreateProject(project *models.Project) error
    
    // GetProjectByName retrieves a project by name.
    // Returns nil, nil if the project does not exist.
    GetProjectByName(name string) (*models.Project, error)
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
