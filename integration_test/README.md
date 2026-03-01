# Integration Testing Framework Documentation

## Overview

This directory contains a comprehensive system integration testing framework for DevOpsMaestro. The framework enables end-to-end testing of CLI commands, complete workflows, state verification, and output validation.

The framework provides:
- **Isolated test environments** - Each test gets its own database and binary
- **CLI command testing** - Execute real dvm commands and verify results
- **Workflow verification** - Test complete user workflows (create hierarchy, manage workspaces, etc.)
- **State verification** - Verify database state after commands
- **Output verification** - Verify CLI output (JSON, YAML, table formats)

## Framework Components

### 1. Core Framework (`framework.go`)

Provides the `TestFramework` struct and utilities for integration testing:

```go
type TestFramework struct {
    TempDir    string  // Root temporary directory
    DBPath     string  // Isolated test database path
    BinaryPath string  // Path to dvm executable
    HomeDir    string  // Isolated home directory
}

// Create isolated test environment
func NewTestFramework(t *testing.T) *TestFramework

// Execute dvm command
func (f *TestFramework) RunDVM(args ...string) (stdout, stderr string, err error)

// Execute command with JSON output
func (f *TestFramework) RunDVMJSON(args ...string) (map[string]interface{}, error)
func (f *TestFramework) RunDVMJSONList(args ...string) ([]map[string]interface{}, error)

// Assert command success/failure
func (f *TestFramework) AssertCommandSuccess(t *testing.T, args ...string)
func (f *TestFramework) AssertCommandFails(t *testing.T, args ...string)

// Verify output
func (f *TestFramework) AssertOutput(t *testing.T, output string, contains ...string)
func (f *TestFramework) AssertOutputDoesNotContain(t *testing.T, output string, notContains ...string)

// Cleanup (call via defer)
func (f *TestFramework) Cleanup()
```

### 2. Test Suites

| File | Purpose | Test Count |
|------|---------|------------|
| `hierarchy_test.go` | Ecosystem/Domain/App hierarchy management | 5 tests |
| `workspace_test.go` | Workspace lifecycle and configuration | 7 tests |
| `gitrepo_test.go` | Git repository management | 8 tests |
| `crud_test.go` | Basic CRUD operations and validation | 10 tests |
| `defaults_integration_test.go` | GetDefaults() functions (existing) | 3 tests |

**Total: 33 integration tests**

### 3. Test Categories

#### Hierarchy Tests (`hierarchy_test.go`)
- `TestHierarchyCreation` - Creating full ecosystem → domain → app hierarchy
- `TestHierarchyMultipleResources` - Multiple resources at each level
- `TestHierarchyContextSwitching` - Context switching between resources
- `TestHierarchyDeletion` - Resource deletion and cleanup
- `TestHierarchyValidation` - Input validation and error handling

#### Workspace Tests (`workspace_test.go`)
- `TestWorkspaceCreation` - Basic workspace creation
- `TestWorkspaceMultiple` - Multiple workspaces per app
- `TestWorkspaceWithGitRepo` - Workspace with git repository
- `TestWorkspaceDelete` - Workspace deletion
- `TestWorkspaceValidation` - Input validation
- `TestWorkspaceContext` - Context management
- `TestWorkspaceTheme` - Theme configuration
- `TestWorkspaceNvimConfig` - Neovim configuration

#### GitRepo Tests (`gitrepo_test.go`)
- `TestGitRepoCreate` - Basic gitrepo creation
- `TestGitRepoMultiple` - Multiple git repositories
- `TestGitRepoGet` - Retrieving specific gitrepo
- `TestGitRepoDelete` - Gitrepo deletion
- `TestGitRepoValidation` - Input validation
- `TestGitRepoWithWorkspace` - Integration with workspaces
- `TestGitRepoWithApp` - Integration with apps
- `TestGitRepoURLFormats` - Various URL formats
- `TestGitRepoUpdate` - Updating gitrepo properties

#### CRUD Tests (`crud_test.go`)
- `TestCRUDEcosystem` - CRUD on ecosystems
- `TestCRUDDomain` - CRUD on domains
- `TestCRUDApp` - CRUD on apps
- `TestCRUDWorkspace` - CRUD on workspaces
- `TestOutputFormats` - JSON/YAML/table output
- `TestContextPersistence` - Context across commands
- `TestBulkOperations` - Multiple resources
- `TestConcurrentSafety` - Concurrent operations
- `TestErrorHandling` - Error cases and messages
- `TestCleanState` - Test isolation

## Usage Examples

### Basic Test Pattern

```go
func TestMyWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    f := NewTestFramework(t)
    defer f.Cleanup()
    
    // Create hierarchy
    f.AssertCommandSuccess(t, "create", "ecosystem", "test")
    f.AssertCommandSuccess(t, "use", "ecosystem", "test")
    
    // Verify ecosystem exists
    ecosystems, err := f.RunDVMJSONList("get", "ecosystems")
    require.NoError(t, err)
    assert.Len(t, ecosystems, 1)
    assert.Equal(t, "test", ecosystems[0]["name"])
}
```

### Testing Complete Workflow

```go
func TestWorkspaceLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    f := NewTestFramework(t)
    defer f.Cleanup()
    
    // Setup hierarchy
    f.AssertCommandSuccess(t, "create", "ecosystem", "eco")
    f.AssertCommandSuccess(t, "use", "ecosystem", "eco")
    f.AssertCommandSuccess(t, "create", "domain", "domain")
    f.AssertCommandSuccess(t, "use", "domain", "domain")
    f.AssertCommandSuccess(t, "create", "app", "app", "--path", "/workspace/app")
    f.AssertCommandSuccess(t, "use", "app", "app")
    
    // Create workspace
    f.AssertCommandSuccess(t, "create", "workspace", "ws1")
    
    // Verify workspace exists
    workspaces, err := f.RunDVMJSONList("get", "workspaces")
    require.NoError(t, err)
    assert.Len(t, workspaces, 1)
    
    // Delete workspace
    f.AssertCommandSuccess(t, "delete", "workspace", "ws1")
    
    // Verify cleanup
    workspaces, err = f.RunDVMJSONList("get", "workspaces")
    require.NoError(t, err)
    assert.Len(t, workspaces, 0)
}
```

### Testing Output Formats

```go
// Test JSON output
result, err := f.RunDVMJSON("get", "ecosystem", "test")
assert.NoError(t, err)
assert.Equal(t, "test", result["name"])

// Test list output
list, err := f.RunDVMJSONList("get", "ecosystems")
assert.NoError(t, err)
assert.Len(t, list, 1)

// Test raw output (YAML, table, etc.)
stdout, _, err := f.RunDVM("get", "ecosystems", "-o", "yaml")
assert.NoError(t, err)
assert.Contains(t, stdout, "name: test")
```

## Running the Tests

### Run All Integration Tests

```bash
# From repository root
cd repos/dvm

# Run all integration tests
go test ./integration_test/... -v

# Run without verbose output
go test ./integration_test/...

# Run specific test file
go test ./integration_test/hierarchy_test.go ./integration_test/framework.go -v

# Run specific test
go test ./integration_test/... -run TestHierarchyCreation -v
```

### Skip Integration Tests (Short Mode)

Integration tests are skipped in short mode for fast unit test runs:

```bash
# Skip integration tests (fast)
go test ./... -short

# Run only integration tests
go test ./integration_test/... -v
```

### CI/CD Integration

Integration tests run in CI with race detection:

```bash
# CI command
go test ./integration_test/... -race -v
```

## Test Isolation

Each test gets:
- ✅ **Isolated database** - Fresh SQLite database in temp directory
- ✅ **Isolated binary** - Compiled dvm binary per test run
- ✅ **Isolated home directory** - Separate `~/.devopsmaestro/` per test
- ✅ **Clean environment** - No interference between tests

This ensures:
- Tests can run in parallel
- No shared state between tests
- Reproducible results
- Safe to run locally or in CI

## Framework Features

### 1. Automatic Binary Compilation

The framework automatically compiles the dvm binary before running tests. If compilation fails, tests fail fast with clear error message.

### 2. Database Initialization

Each test gets a fresh database with proper schema. No manual setup required.

### 3. Environment Isolation

Tests set `HOME` and `DVM_DB_PATH` environment variables to ensure complete isolation.

### 4. Cleanup Guarantees

Using `defer f.Cleanup()` ensures temporary files are removed even if test fails.

### 5. JSON Parsing Utilities

Built-in JSON parsing for easy verification of structured output:
- `RunDVMJSON()` - Parse single object
- `RunDVMJSONList()` - Parse array of objects

### 6. Assertion Helpers

Convenient assertion methods:
- `AssertCommandSuccess()` - Fail if command fails
- `AssertCommandFails()` - Fail if command succeeds
- `AssertOutput()` - Verify output contains strings
- `AssertOutputDoesNotContain()` - Verify output doesn't contain strings

## Test Coverage

The integration tests verify:

✅ **Complete Workflows**
- Create full hierarchy (ecosystem → domain → app → workspace)
- Context switching between resources
- Resource deletion and cleanup

✅ **CRUD Operations**
- Create, Read, Update, Delete for all resource types
- Duplicate name handling
- Validation and error cases

✅ **Output Formats**
- JSON output parsing
- YAML output verification
- Table format verification

✅ **State Verification**
- Database state after operations
- Context persistence across commands
- Resource scoping (domains in ecosystems, etc.)

✅ **Error Handling**
- Empty names rejected
- Duplicates rejected
- Non-existent resources handled
- Error messages on stderr

## Adding New Tests

1. Create new test file in `integration_test/`
2. Use `package integration`
3. Import the framework
4. Follow the test pattern:

```go
func TestMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    f := NewTestFramework(t)
    defer f.Cleanup()
    
    // Your test code here
}
```

## Debugging Failed Tests

If a test fails:

1. **Check stderr output** - Error messages go to stderr
2. **Verify command args** - Framework logs failed commands
3. **Check database state** - Use `f.GetDatabasePath()` to inspect
4. **Run single test** - `go test -run TestName -v` for detailed output

## Future Enhancements

Potential additions to the framework:

- [ ] Container runtime testing (requires Docker/Colima)
- [ ] Image building tests
- [ ] Workspace start/stop/attach tests (requires container runtime)
- [ ] Performance benchmarks
- [ ] Parallel test execution optimization
- [ ] Test data fixtures
- [ ] Custom assertions for specific resources