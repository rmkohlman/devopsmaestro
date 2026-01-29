# DevOpsMaestro - AI Assistant Context

> **Purpose:** This file provides context for AI assistants (Claude, etc.) to understand
> the project architecture, design principles, and development patterns.

---

## QUICK START FOR AI ASSISTANTS

**Read this section first to get fully up to speed on the project.**

### Step 1: Read These Files (In Order)

| Priority | File | Purpose | Location |
|----------|------|---------|----------|
| 1 | `current-session.md` | **What's in progress RIGHT NOW** - read first! | `~/Developer/tools/devopsmaestro-private/` |
| 2 | `CLAUDE.md` | Project architecture & patterns (this file) | `~/Developer/tools/devopsmaestro/` |
| 3 | `decisions.md` | Technical decisions history | `~/Developer/tools/devopsmaestro-private/` |
| 4 | `project-context.md` | High-level architecture & roadmap | `~/Developer/tools/devopsmaestro-private/` |

### Step 2: Know the Key Locations

```
~/Developer/tools/devopsmaestro/           # PUBLIC - Main codebase
~/Developer/tools/devopsmaestro-private/   # PRIVATE - Session state & decisions
```

### Step 3: Reference Documents by Task

| When Doing... | Read This Document | Location |
|---------------|-------------------|----------|
| **Releasing a version** | `docs/development/release-process.md` | Main repo |
| **Running tests** | `MANUAL_TEST_PLAN.md` | Main repo |
| **Writing code** | `STANDARDS.md` | Main repo |
| **Making architecture decisions** | `decisions.md` | Private repo |
| **Understanding the codebase** | `project-context.md` | Private repo |

### Step 4: Session Protocol

**AT START of every session:**
```bash
# 1. Read current session state (MOST IMPORTANT)
cat ~/Developer/tools/devopsmaestro-private/current-session.md

# 2. Pull latest changes
cd ~/Developer/tools/devopsmaestro-private && git pull
cd ~/Developer/tools/devopsmaestro && git pull
```

**AT END of every session:**
```bash
# 1. Update current-session.md with what was done and next steps
# 2. Commit and push BOTH repos
cd ~/Developer/tools/devopsmaestro-private
git add . && git commit -m "Update session: <brief description>" && git push

cd ~/Developer/tools/devopsmaestro
git add . && git commit -m "<commit message>" && git push
```

### Quick Command Reference

```bash
# Build & Test
go build -o dvm              # Build binary
go test ./...                # Run all tests
go fmt ./...                 # Format code
go vet ./...                 # Check for issues

# Manual Tests
./tests/manual/part1-setup-and-build.sh   # Setup & build tests
./tests/manual/part2-post-attach.sh       # Post-attach tests

# Release (follow docs/development/release-process.md)
git tag -a vX.Y.Z -m "Release vX.Y.Z: description"
git push origin vX.Y.Z
```

### File Index

| Document | Purpose | When to Use |
|----------|---------|-------------|
| **Main Repo (`devopsmaestro/`)** | | |
| `CLAUDE.md` | AI context, architecture, patterns | Always - this is your primary reference |
| `STANDARDS.md` | Code standards, design principles | When writing/reviewing code |
| `MANUAL_TEST_PLAN.md` | Test procedures, test scripts | When testing changes |
| `README.md` | User documentation | Understanding features |
| `CHANGELOG.md` | Version history | Before releases |
| `docs/development/release-process.md` | Release checklist | When releasing versions |
| `verify-release.sh` | Release verification | After creating releases |
| `.goreleaser.yaml` | GoReleaser config | Automated releases |
| **Private Repo (`devopsmaestro-private/`)** | | |
| `current-session.md` | Active work state | **START OF EVERY SESSION** |
| `decisions.md` | Technical decisions (ADR format) | Making architecture decisions |
| `project-context.md` | Architecture overview | Understanding the project |
| `archive/` | Historical sessions | Reference past work |

---

## CRITICAL: Private Session Data Repository

### Location & Purpose

```
~/Developer/tools/devopsmaestro-private/
```

**GitHub:** `https://github.com/rmkohlman/devopsmaestro-private` (PRIVATE repo)

This is a **separate private Git repository** that maintains AI session state and development
context across conversations. It exists to provide continuity between AI sessions, allowing
seamless resumption of work without losing context.

### Why This Exists

1. **Session Continuity** - AI assistants lose context between conversations; this repo preserves it
2. **Multi-Machine Access** - Work from any machine with consistent context (synced via Git)
3. **Decision History** - Track why architectural decisions were made over time
4. **Progress Tracking** - Know exactly what was completed and what's next
5. **Collaboration** - Share context with trusted collaborators (private access controlled)

### Repository Contents

| File | Purpose | When to Read |
|------|---------|--------------|
| `current-session.md` | **Active session state** - what's in progress, next steps, recent work | **READ FIRST** at session start |
| `decisions.md` | Technical decisions log with rationale, alternatives considered, consequences | When making architecture decisions |
| `project-context.md` | High-level architecture, design patterns, data models, roadmap | For understanding project structure |
| `archive/` | Historical session states organized by date/feature | Reference for past work |
| `README.md` | Repository documentation and workflow instructions | First time setup |

### Session Workflow

**At the START of each session:**
```bash
# 1. Read current session state
cat ~/Developer/tools/devopsmaestro-private/current-session.md

# 2. Check for relevant past decisions
cat ~/Developer/tools/devopsmaestro-private/decisions.md

# 3. Pull latest changes (in case updated from another machine)
cd ~/Developer/tools/devopsmaestro-private && git pull
```

**At the END of each session:**
```bash
# 1. Update current-session.md with:
#    - What was completed
#    - What's in progress
#    - Next steps
#    - Any blockers or issues

# 2. Add significant technical decisions to decisions.md

# 3. Commit and push BOTH repos
cd ~/Developer/tools/devopsmaestro-private
git add . && git commit -m "Update session: <brief description>" && git push

cd ~/Developer/tools/devopsmaestro
git add . && git commit -m "<commit message>" && git push
```

### Contents Detail

**`current-session.md`** structure:
- Executive Summary - Quick overview of session focus
- Bugs Fixed - Details of any bug fixes with file locations
- Features Added - New functionality implemented
- Files Modified - List of changed files
- Test Status - Current test results
- Next Steps - What to work on next
- Commands to Test - Helpful verification commands

**`decisions.md`** structure (ADR format):
- Decision number and title
- Status (Proposed/Accepted/Deprecated)
- Context - Why the decision was needed
- Decision - What was decided
- Rationale - Why this approach was chosen
- Consequences - Pros and cons
- Alternatives Considered - Other options evaluated

**`project-context.md`** includes:
- Project vision and philosophy
- Technology stack details
- Architecture overview with diagrams
- Design patterns used
- Data model and database schema
- UI/theme system documentation
- Roadmap and future plans
- Development workflow

### Relationship Between Repos

```
devopsmaestro/                    (PUBLIC - main codebase)
├── CLAUDE.md                     ← Points to private repo
├── STANDARDS.md                  ← Development standards
├── MANUAL_TEST_PLAN.md           ← Testing procedures
└── [source code]

devopsmaestro-private/            (PRIVATE - session state)
├── current-session.md            ← Active work state
├── decisions.md                  ← Technical decisions
├── project-context.md            ← Architecture docs
└── archive/                      ← Historical sessions
```

---

## Project Overview

**DevOpsMaestro (dvm)** is a kubectl-style CLI for managing containerized development environments
with database-backed Neovim configuration management.

**Key Value Proposition:** Developers can spin up fully-configured, reproducible dev environments
with a single command, with Neovim pre-configured with LSP, formatters, and plugins.

## Core Design Principles

### 1. DECOUPLING IS PARAMOUNT

Everything must be modular and swappable. This allows:
- Easy testing with mocks
- Future extensibility (new databases, runtimes, output formats)
- Gradual migration without breaking changes
- Different deployment scenarios

**Pattern:** Interface → Implementation → Factory

```go
// Interface defines the contract
type DataStore interface {
    CreateProject(project *models.Project) error
    // ...
}

// Implementation fulfills the contract
type SQLDataStore struct {
    driver Driver
}

// Factory creates the implementation
func CreateDataStore() (DataStore, error) {
    driver, err := DriverFactory()
    // ...
    return NewSQLDataStore(driver, nil), nil
}
```

### 2. ADAPTER PATTERN FOR BRIDGING

When migrating from old to new architecture, use adapters:

```go
// DatabaseDriverAdapter wraps legacy Database to implement new Driver interface
type DatabaseDriverAdapter struct {
    db         Database      // Legacy interface
    driverType DriverType
}

// Implements Driver interface, delegates to Database
func (d *DatabaseDriverAdapter) Execute(query string, args ...interface{}) (Result, error) {
    result, err := d.db.Execute(query, args...)
    // ...
}
```

### 3. FACTORY PATTERN EVERYWHERE

Factories enable:
- Configuration-based creation
- Dependency injection
- Testing with mocks

```go
// DataStoreFactory interface allows swapping factory implementations
type DataStoreFactory interface {
    Create() (DataStore, error)
}

// DefaultDataStoreFactory uses viper config
type DefaultDataStoreFactory struct{}

// MockDataStoreFactory for testing
type MockDataStoreFactory struct {
    store DataStore
}
```

### 4. MOCK IMPLEMENTATIONS FOR TESTING

Every major interface has a mock:

| Interface | Mock | Location |
|-----------|------|----------|
| `Driver` | `MockDriver` | `db/mock_driver.go` |
| `DataStore` | `MockDataStore` | `db/mock_store.go` |
| `ContainerRuntime` | `MockContainerRuntime` | `operators/mock_runtime.go` |
| `Formatter` | `MockFormatter` | `output/mock_formatter.go` |
| `Manager` (nvim) | `MockManager` | `nvim/mock_manager.go` |

Mocks support:
- Call recording for verification
- Error injection for testing error paths
- In-memory state tracking

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer (cmd/)                         │
│  Commands: init, create, get, use, build, attach, plugin, etc.  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Output Layer (output/)                      │
│  Formatter interface → PlainFormatter, ColoredFormatter, etc.   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Database Layer (db/)                          │
│  DataStore interface → SQLDataStore                              │
│  Driver interface → SQLiteDriver, PostgresDriver, MemoryDriver   │
│  QueryBuilder interface → SQLiteQueryBuilder, PostgresBuilder    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Container Layer (operators/)                   │
│  ContainerRuntime interface → DockerRuntime, ContainerdRuntime  │
│  Platform detection → OrbStack, Colima, Podman, Docker Desktop  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Builder Layer (builders/)                     │
│  ImageBuilder interface → DockerBuilder, BuildKitBuilder        │
│  Dockerfile generation, image building                           │
└─────────────────────────────────────────────────────────────────┘
```

## Key Interfaces

### Database Layer (`db/interfaces.go`)

```go
// Driver - Low-level database operations
type Driver interface {
    Connect() error
    Close() error
    Execute(query string, args ...interface{}) (Result, error)
    Query(query string, args ...interface{}) (Rows, error)
    QueryRow(query string, args ...interface{}) Row
    Begin() (Transaction, error)
    Type() DriverType
    DSN() string
    MigrationDSN() string
}

// DataStore - High-level application operations
type DataStore interface {
    // Projects
    CreateProject(project *models.Project) error
    GetProjectByName(name string) (*models.Project, error)
    ListProjects() ([]*models.Project, error)
    
    // Workspaces
    CreateWorkspace(workspace *models.Workspace) error
    GetWorkspaceByName(projectID int, name string) (*models.Workspace, error)
    
    // Plugins
    CreatePlugin(plugin *models.NvimPluginDB) error
    GetPluginByName(name string) (*models.NvimPluginDB, error)
    ListPlugins() ([]*models.NvimPluginDB, error)
    
    // Context
    GetContext() (*models.Context, error)
    SetActiveProject(projectID *int) error
}

// QueryBuilder - SQL dialect abstraction
type QueryBuilder interface {
    Placeholder(index int) string  // ? vs $1
    Now() string                    // datetime('now') vs NOW()
    Boolean(value bool) string      // 0/1 vs TRUE/FALSE
}
```

### Container Layer (`operators/runtime_interface.go`)

```go
type ContainerRuntime interface {
    BuildImage(ctx context.Context, opts BuildOptions) error
    StartWorkspace(ctx context.Context, opts StartOptions) (string, error)
    AttachToWorkspace(ctx context.Context, containerID string) error
    StopWorkspace(ctx context.Context, containerID string) error
    GetWorkspaceStatus(ctx context.Context, containerID string) (string, error)
    GetRuntimeType() string
}
```

### Output Layer (`output/interfaces.go`)

```go
type Formatter interface {
    // Messages
    Info(message string)
    Success(message string)
    Warning(message string)
    Error(message string)
    
    // Structured output
    Table(headers []string, rows [][]string)
    KeyValue(pairs []KeyValuePair)
    Object(obj interface{}, format string) // yaml, json, table
}
```

### Logging (`log/slog` - Go Standard Library)

DevOpsMaestro uses Go's built-in `log/slog` for structured logging.

**Design Philosophy:**
- **User output** goes through `output.Formatter` (success/error/table messages)
- **Debug logging** goes through `slog` (only shown with `-v` flag)
- Silent by default - CLIs shouldn't spam users

**Global Flags:**
```bash
dvm -v build            # Enable debug logging to stderr
dvm --log-file /tmp/dvm.log build  # JSON logs to file
```

**Usage Pattern:**
```go
import "log/slog"

func someCommand() error {
    // User sees this (via formatter)
    formatter.Info("Building workspace...")
    
    // Debug log (only with -v flag)
    slog.Debug("starting operation", "workspace", name, "path", path)
    
    // Info for significant events
    slog.Info("build completed", "image", imageName, "duration", elapsed)
    
    // Errors (always log, even without -v)
    slog.Error("build failed", "error", err, "image", imageName)
    
    return nil
}
```

**Log Levels:**
| Level | When to Use | Example |
|-------|-------------|---------|
| `slog.Debug` | Verbose details, variable values | `slog.Debug("resolved path", "path", p)` |
| `slog.Info` | Significant events | `slog.Info("container started", "id", id)` |
| `slog.Warn` | Recoverable issues | `slog.Warn("image may not be built")` |
| `slog.Error` | Failures | `slog.Error("failed to connect", "error", err)` |

**Best Practices:**
1. Use key-value pairs for structured data: `slog.Debug("msg", "key1", val1, "key2", val2)`
2. Always include `"error", err` when logging errors
3. Log at function entry/exit for tracing: `slog.Info("starting build")` / `slog.Info("build completed")`
4. Don't log sensitive data (tokens, passwords)

## Directory Structure

```
devopsmaestro/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command, global flags
│   ├── init.go            # dvm admin init
│   ├── create.go          # dvm create project/workspace
│   ├── get.go             # dvm get projects/workspaces/plugins
│   ├── use.go             # dvm use project/workspace
│   ├── build.go           # dvm build
│   ├── attach.go          # dvm attach
│   └── plugin.go          # dvm plugin apply/list/get/delete
│
├── db/                     # Database layer (DECOUPLED)
│   ├── interfaces.go      # Driver, DataStore, QueryBuilder interfaces
│   ├── driver.go          # Driver factory and registration
│   ├── factory.go         # DataStore factory, legacy adapters
│   ├── store.go           # SQLDataStore implementation
│   ├── sqlite_driver.go   # SQLite Driver implementation
│   ├── querybuilder.go    # QueryBuilder implementations
│   ├── mock_driver.go     # MockDriver for testing
│   ├── mock_store.go      # MockDataStore for testing
│   └── database.go        # Legacy Database interface (deprecated)
│
├── operators/              # Container runtime layer (DECOUPLED)
│   ├── runtime_interface.go    # ContainerRuntime interface
│   ├── runtime_factory.go      # Runtime factory
│   ├── docker_runtime.go       # Docker/OrbStack/Podman implementation
│   ├── containerd_runtime_v2.go # Containerd implementation
│   ├── platform.go             # Platform detection
│   └── mock_runtime.go         # MockContainerRuntime for testing
│
├── builders/               # Image builder layer (DECOUPLED)
│   ├── interfaces.go      # ImageBuilder interface
│   ├── factory.go         # Builder factory
│   ├── docker_builder.go  # Docker build implementation
│   ├── buildkit_builder.go # BuildKit implementation
│   └── helpers.go         # Shared utilities
│
├── output/                 # Output formatting (DECOUPLED)
│   ├── interfaces.go      # Formatter interface
│   ├── formatter.go       # Base formatter
│   ├── plain_formatter.go # Plain text output
│   ├── colored_formatter.go # Colored output with themes
│   └── mock_formatter.go  # MockFormatter for testing
│
├── nvim/                   # Neovim configuration management
│   ├── interfaces.go      # Manager interface
│   ├── manager.go         # Default implementation
│   └── mock_manager.go    # MockManager for testing
│
├── models/                 # Data models
│   ├── project.go
│   ├── workspace.go
│   ├── nvim_plugin.go
│   └── context.go
│
├── config/                 # Configuration management
│   └── config.go
│
├── migrations/             # Database migrations
│   ├── sqlite/
│   └── postgres/
│
├── templates/              # Templates and pre-built configs
│   ├── nvim-plugins/      # Pre-built Neovim plugin YAMLs
│   └── minimal/           # Minimal nvim config template
│
├── main.go                # Entry point
├── embed.go               # Embedded filesystem
├── CLAUDE.md              # THIS FILE - AI context
├── MANUAL_TEST_PLAN.md    # Comprehensive manual test plan
└── README.md              # User documentation
```

## Factory Patterns in Detail

### Creating a DataStore (Recommended Path)

```go
// main.go - Simple, decoupled approach
func main() {
    store, err := db.CreateDataStore()  // Uses viper config
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()
    
    // Pass store to commands
    cmd.Execute(store, migrationsFS)
}

// CreateDataStore creates driver based on config, wraps in DataStore
func CreateDataStore() (DataStore, error) {
    driver, err := DriverFactory()  // Creates SQLiteDriver, PostgresDriver, etc.
    if err != nil {
        return nil, err
    }
    if err := driver.Connect(); err != nil {
        return nil, err
    }
    return NewSQLDataStore(driver, nil), nil
}
```

### Creating a ContainerRuntime

```go
// Auto-detect platform and create appropriate runtime
runtime, err := operators.NewContainerRuntime()

// Or create with specific platform
platform, _ := detector.Detect()
runtime, err := operators.NewDockerRuntime(platform)
```

### Creating an ImageBuilder

```go
// Factory selects builder based on platform
builder, err := builders.NewImageBuilder(platform, buildConfig)

// OrbStack/Docker Desktop/Podman → DockerBuilder
// Colima containerd → BuildKitBuilder
```

## Testing Patterns

### Unit Testing with Mocks

```go
func TestProjectCreation(t *testing.T) {
    // Create mock
    mockStore := db.NewMockDataStore()
    
    // Setup expectations
    mockStore.Projects["test"] = &models.Project{Name: "test"}
    
    // Test
    project, err := mockStore.GetProjectByName("test")
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, "test", project.Name)
    
    // Check calls
    assert.True(t, mockStore.WasCalled("GetProjectByName"))
}
```

### Interface Compliance Testing

```go
func TestSQLDataStore_ImplementsInterface(t *testing.T) {
    var _ db.DataStore = (*db.SQLDataStore)(nil)
}

func TestMockDataStore_ImplementsInterface(t *testing.T) {
    var _ db.DataStore = (*db.MockDataStore)(nil)
}
```

### Swappability Testing

```go
func TestDataStoreSwappability(t *testing.T) {
    // Both implementations should work through the interface
    implementations := []db.DataStore{
        createRealStore(t),
        db.NewMockDataStore(),
    }
    
    for _, store := range implementations {
        // Same tests pass for all implementations
        testProjectOperations(t, store)
    }
}
```

## Roadmap

### v0.3.0 - Current Development
- Multi-platform container runtime support (OrbStack, Colima, Podman, Docker Desktop)
- Decoupled builder architecture (ImageBuilder interface)
- Platform detection and selection
- Output formatting system
- Database layer decoupling (Driver/DataStore interfaces)
- Mock implementations for all major interfaces

### v0.4.0 - Planned
- nvim-yaml tool: Declarative Neovim configuration via YAML
- Workspace nvim integration: Per-workspace Neovim configurations
- Output refactoring: Migrate all commands to use output.Formatter

### v0.5.0 - Future (Kubernetes Operator)
- Kubernetes runtime support: Run dev environments as Kubernetes pods
- k3s/k3d integration: Lightweight k3s clusters on all platforms
- DevOpsMaestro Operator: Kubernetes operator for managing workspaces
- CRDs: Custom Resource Definitions for Workspace, Project, NvimConfig

## Important Patterns to Follow

### When Adding New Features

1. **Define an interface first**
2. **Create the implementation**
3. **Create a mock for testing**
4. **Add factory function**
5. **Write tests using both real and mock implementations**

### When Modifying Existing Code

1. **Don't break existing interfaces** - Add new methods, don't change signatures
2. **Use adapters for migration** - Bridge old and new code
3. **Maintain backward compatibility** - Legacy code paths should keep working
4. **Add deprecation notices** - Mark old code as deprecated with migration path

### Code Style

- Use explicit error handling (no panic)
- Return interfaces, accept interfaces
- Keep functions small and focused
- Document public APIs
- Use meaningful variable names
- Group related functionality in packages

## Quick Commands

```bash
# Build
go build -o dvm

# Test all
go test ./...

# Test specific package
go test ./db/... -v

# Test with coverage
go test -cover ./...

# Run specific test
go test -run TestProjectCreation ./db/...

# Format code
go fmt ./...

# Lint
golangci-lint run
```

## Files to Update When Making Changes

| Change Type | Files to Update |
|-------------|-----------------|
| New interface | `interfaces.go` in package, mock file, factory |
| New database table | `migrations/`, `store.go`, `interfaces.go` |
| New command | `cmd/`, possibly `executor.go` |
| New platform | `operators/platform.go`, `runtime_factory.go` |
| New builder | `builders/interfaces.go`, `factory.go`, new impl file |
| New output format | `output/formatter.go` |

## Contact

For questions about this codebase, refer to:
- This file (CLAUDE.md) for architecture
- README.md for user documentation
- MANUAL_TEST_PLAN.md for testing procedures
- `docs/` for detailed documentation
