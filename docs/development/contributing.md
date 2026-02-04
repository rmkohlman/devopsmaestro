# Contributing to DevOpsMaestro

Thank you for your interest in contributing to DevOpsMaestro! This guide will help you get started.

## Getting Started

### Prerequisites

- **Go 1.25+** - DevOpsMaestro is written in Go
- **Git** - For version control
- **Docker or Colima** - For testing container features
- **golangci-lint** (optional) - For linting

### Clone the Repository

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro
```

### Build the Project

```bash
# Build dvm (DevOpsMaestro)
go build -o dvm .

# Build nvp (NvimOps)
go build -o nvp ./cmd/nvp/

# Verify builds
./dvm version
./nvp version
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with race detector (as CI does)
go test ./... -race

# Run specific package tests
go test ./pkg/nvimops/... -v
go test ./db/... -v
```

---

## Development Standards

Before writing code, please read our standards documents:

1. **[CLAUDE.md](https://github.com/rmkohlman/devopsmaestro/blob/main/CLAUDE.md)** - Architecture overview
2. **[STANDARDS.md](https://github.com/rmkohlman/devopsmaestro/blob/main/STANDARDS.md)** - Design patterns and coding standards

### Core Principles

1. **Decoupling** - Interface → Implementation → Factory pattern
2. **Single Responsibility** - Each component has one job
3. **Dependency Injection** - Dependencies injected, not created internally
4. **Testability** - Every interface has a mock

### Code Organization

```
package/
├── interfaces.go      # Interface definitions
├── implementation.go  # Concrete implementations
├── factory.go         # Factory functions
├── mock_*.go          # Mock implementations
├── *_test.go          # Tests
```

---

## Making Changes

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes

- Follow existing code patterns
- Add tests for new functionality
- Update documentation if needed

### 3. Test Your Changes

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### 4. Commit Your Changes

We use conventional commits:

```bash
# Features
git commit -m "feat: add new workspace command"

# Bug fixes
git commit -m "fix: correct plugin deletion error"

# Documentation
git commit -m "docs: update installation guide"

# Refactoring
git commit -m "refactor: extract builder interface"

# Tests
git commit -m "test: add plugin store tests"

# Chores
git commit -m "chore: update dependencies"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

---

## Code Review Checklist

Before submitting, ensure:

- [ ] Code follows decoupling principles
- [ ] Interface defined (if adding new component)
- [ ] Mock exists for testing
- [ ] Factory function exists (if applicable)
- [ ] Tests written and passing
- [ ] Error handling is proper
- [ ] Documentation updated

---

## Adding New Features

### Adding a New Command

1. Create command file in `cmd/`
2. Register with parent command
3. Add tests in `cmd/*_test.go`
4. Update MANUAL_TEST_PLAN.md

### Adding a New Interface

1. Define interface in `interfaces.go`
2. Create implementation
3. Create mock for testing
4. Add factory function
5. Write interface compliance tests

### Adding a New Resource Type

1. Define types in `models/` or relevant package
2. Implement `resource.Handler` interface
3. Register handler in `pkg/resource/registry.go`
4. Add YAML parsing support
5. Write tests

---

## Testing

### Running Tests

```bash
# All tests
go test ./...

# With verbose output
go test ./... -v

# With race detection
go test ./... -race

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Writing Tests

```go
func TestMyFeature(t *testing.T) {
    // Arrange
    store := NewMockDataStore()
    store.Projects["test"] = &models.Project{Name: "test"}
    
    // Act
    result, err := store.GetProjectByName("test")
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Name != "test" {
        t.Errorf("expected 'test', got '%s'", result.Name)
    }
}
```

### Interface Compliance Tests

```go
func TestMyStore_ImplementsDataStore(t *testing.T) {
    var _ DataStore = (*MyStore)(nil)
}
```

---

## Documentation

### When to Update Docs

| Change Type | Update |
|-------------|--------|
| New command | README.md, command docs |
| New feature | Feature docs, examples |
| Bug fix | Add test (regression) |
| API change | CHANGELOG.md |

### Building Docs Locally

```bash
# Install MkDocs
pip install mkdocs-material

# Serve locally
mkdocs serve

# Visit http://127.0.0.1:8000
```

---

## Release Process

Releases are handled by maintainers. See [Release Process](./release-process.md) for details.

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0) - Breaking changes
- **MINOR** (0.X.0) - New features (backward compatible)
- **PATCH** (0.0.X) - Bug fixes (backward compatible)

---

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/rmkohlman/devopsmaestro/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rmkohlman/devopsmaestro/discussions)

---

## License

By contributing, you agree that your contributions will be licensed under the GPL-3.0 License.

---

Thank you for contributing to DevOpsMaestro!
