# AI Assistant Instructions for DevOpsMaestro

> **⚠️ READ THIS FILE BEFORE WRITING ANY CODE**
>
> This file contains mandatory patterns and standards. Failure to follow these
> will result in inconsistent code that will need to be refactored later.

---

## Pre-Code Checklist

Before writing ANY code, complete this checklist:

- [ ] Read `CLAUDE.md` for architecture overview
- [ ] Read `STANDARDS.md` for design patterns (especially sections 1-7)
- [ ] Identify which pattern(s) apply to your task
- [ ] Check if similar code exists to follow as a template

---

## Mandatory Patterns

### 1. Resource/Handler Pattern (CLI Commands)

**ALL CLI commands that perform CRUD operations MUST use the Resource/Handler pattern.**

Location: `pkg/resource/` and `pkg/resource/handlers/`

**✅ CORRECT:**
```go
// Use resource.* functions for Get, List, Delete
func getApps(cmd *cobra.Command) error {
    ctx, err := buildResourceContext(cmd)
    if err != nil {
        return err
    }
    
    resources, err := resource.List(ctx, handlers.KindApp)
    if err != nil {
        return fmt.Errorf("failed to list apps: %w", err)
    }
    
    // Extract models from resources
    apps := make([]*models.App, len(resources))
    for i, res := range resources {
        apps[i] = res.(*handlers.AppResource).App()
    }
    
    return render.OutputWith(getOutputFormat, apps, render.Options{})
}
```

**❌ INCORRECT:**
```go
// Direct DataStore calls bypass the unified pipeline
func getApps(cmd *cobra.Command) error {
    ds, _ := getDataStore(cmd)
    apps, _ := ds.ListAllApps()  // BAD: Direct DataStore call
    return render.OutputWith(getOutputFormat, apps, render.Options{})
}
```

### 2. Interface → Implementation → Factory

**ALL new components must follow this pattern:**

```go
// 1. Interface (in interfaces.go or datastore.go)
type MyService interface {
    DoThing() error
}

// 2. Implementation
type myServiceImpl struct {
    deps SomeDependency
}

func (s *myServiceImpl) DoThing() error { /* ... */ }

// 3. Factory
func NewMyService(deps SomeDependency) MyService {
    return &myServiceImpl{deps: deps}
}

// 4. Mock (for testing)
type MockMyService struct {
    DoThingErr error
    Calls      []string
}
```

### 3. Dependency Injection

**Dependencies MUST be injected, not created internally.**

**✅ CORRECT:**
```go
func (cmd *cobra.Command) RunE(cmd *cobra.Command, args []string) error {
    ds, err := getDataStore(cmd)  // Injected via context
    // ...
}
```

**❌ INCORRECT:**
```go
func (cmd *cobra.Command) RunE(cmd *cobra.Command, args []string) error {
    db, _ := db.InitializeDBConnection()  // Creates its own
    defer db.Close()
    // ...
}
```

### 4. Decoupled Output Rendering

**ALL CLI output MUST use the render package.**

```go
// Use render.OutputWith for structured data
return render.OutputWith(getOutputFormat, data, render.Options{
    Type: render.TypeTable,
})

// Use render.Success/Error/Info for messages
render.Success("Operation completed")
render.Error("Something went wrong")
render.Info("Hint: try this command")
```

---

## Adding New Features Checklist

### Adding a New Resource Type (e.g., "Widget")

1. [ ] Create model in `models/widget.go`
2. [ ] Add DataStore interface methods in `db/datastore.go`
3. [ ] Implement in `db/store.go`
4. [ ] Add to `db/mock_store.go`
5. [ ] Create migration in `migrations/sqlite/`
6. [ ] Create handler in `pkg/resource/handlers/widget.go`
7. [ ] Register handler in `pkg/resource/handlers/register.go`
8. [ ] Create CLI commands in `cmd/widget.go`
9. [ ] Add tests for handler and commands
10. [ ] Update STANDARDS.md resource type table

### Adding a New CLI Command

1. [ ] Determine if it's a CRUD operation → Use Resource/Handler pattern
2. [ ] Use `buildResourceContext(cmd)` for resource operations
3. [ ] Use `getDataStore(cmd)` only for context state operations
4. [ ] Use `render.OutputWith()` for output
5. [ ] Add to parent command in `init()`
6. [ ] Add tests

### Adding a New DataStore Method

1. [ ] Add to `DataStore` interface in `db/datastore.go`
2. [ ] Implement in `db/store.go`
3. [ ] Add to `MockDataStore` in `db/mock_store.go`
4. [ ] Add error injection field if needed
5. [ ] Add tests in `db/store_test.go`

---

## Code Review Reminders

When reviewing your own code before committing:

- [ ] Did I use the Resource/Handler pattern for CRUD operations?
- [ ] Did I inject dependencies instead of creating them?
- [ ] Did I use the render package for all output?
- [ ] Did I add/update the mock for any new interfaces?
- [ ] Did I add tests?
- [ ] Did I run `go test ./... -race`?

---

## Quick Reference: Key Files

| Purpose | File |
|---------|------|
| Resource interface | `pkg/resource/resource.go` |
| Resource registry | `pkg/resource/registry.go` |
| Handlers | `pkg/resource/handlers/*.go` |
| Handler registration | `pkg/resource/handlers/register.go` |
| DataStore interface | `db/datastore.go` |
| DataStore implementation | `db/store.go` |
| Mock DataStore | `db/mock_store.go` |
| Render system | `render/*.go` |
| Build resource context | `cmd/apply.go` → `buildResourceContext()` |

---

## Common Mistakes to Avoid

1. **Bypassing the Resource/Handler pattern** for "simple" operations
2. **Creating DataStore connections** inside command handlers
3. **Using fmt.Println** instead of render package
4. **Forgetting to update mocks** when adding interface methods
5. **Not running tests** before committing

---

## When in Doubt

1. Look at existing code that does something similar
2. Check `pkg/resource/handlers/ecosystem.go` as a reference implementation
3. Ask: "Is this decoupled? Can it be tested with a mock?"

---

*Last updated: Session implementing Ecosystem/Domain/App hierarchy (v0.8.0)*
