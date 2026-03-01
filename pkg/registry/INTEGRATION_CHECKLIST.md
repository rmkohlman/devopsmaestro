# ServiceManager Integration Checklist

## ✅ Completed

- [x] ServiceManager interface defined
- [x] RegistryStrategy interface defined
- [x] ZotStrategy implemented (wraps ZotManager)
- [x] AthensStrategy implemented (wraps AthensManager via adapter)
- [x] AthensManagerAdapter created (bridges GoModuleProxy to ServiceManager)
- [x] StubStrategy base class for future implementations
- [x] Stub strategies for devpi, verdaccio, squid
- [x] ServiceFactory implementation
- [x] Comprehensive test suite (all tests pass)
- [x] Example code demonstrating usage
- [x] Architecture documentation (SERVICE_MANAGER.md)
- [x] Implementation summary (SERVICE_MANAGER_SUMMARY.md)
- [x] Main binary still compiles

## 🔄 Next: CLI Integration

### Phase 1: DataStore Methods (Database Agent)

Add to `db/datastore.go`:

```go
// Registry CRUD operations
CreateRegistry(registry *models.Registry) error
GetRegistryByName(name string) (*models.Registry, error)
GetRegistryByID(id int) (*models.Registry, error)
UpdateRegistry(registry *models.Registry) error
DeleteRegistry(name string) error
ListRegistries() ([]*models.Registry, error)
```

Add to `db/store.go`:

```go
// Implement the above methods using SQL queries
// Example:
func (s *SQLDataStore) CreateRegistry(registry *models.Registry) error {
    // INSERT INTO registries ...
}

func (s *SQLDataStore) GetRegistryByName(name string) (*models.Registry, error) {
    // SELECT * FROM registries WHERE name = ?
}
```

Add migration: `db/migrations/sqlite/011_add_registries.up.sql`

### Phase 2: CLI Commands

Update `cmd/registry.go` to use ServiceFactory:

```go
// Before: Direct manager creation
manager := registry.NewZotManager(...)

// After: Factory-based creation
registry, err := datastore.GetRegistryByName(name)
if err != nil {
    return err
}

factory := registry.NewServiceFactory()
manager, err := factory.CreateManager(registry)
if err != nil {
    return err
}

// Use manager
ctx := context.Background()
if err := manager.Start(ctx); err != nil {
    return err
}

// Update database
registry.Status = "running"
datastore.UpdateRegistry(registry)
```

Commands to update:
- [x] `dvm registry start <name>` - Use factory to create manager
- [x] `dvm registry stop <name>` - Use factory to create manager
- [x] `dvm registry status <name>` - Use factory to create manager
- [x] `dvm registry list` - Show all registries from DB
- [x] `dvm registry create -f registry.yaml` - Parse YAML and store in DB
- [x] `dvm registry delete <name>` - Remove from DB (stop first if running)

### Phase 3: Resource Handler

Create `pkg/resource/registry_handler.go`:

```go
type RegistryHandler struct {
    datastore db.DataStore
    factory   *registry.ServiceFactory
}

func (h *RegistryHandler) Apply(yamlPath string) error {
    // 1. Parse YAML
    // 2. Convert to models.Registry
    // 3. Store in DB via datastore.CreateRegistry()
    // 4. Optionally start if lifecycle is "persistent"
}

func (h *RegistryHandler) Delete(name string) error {
    // 1. Lookup registry
    // 2. Stop if running
    // 3. Delete from DB
}

func (h *RegistryHandler) Get(name string) error {
    // 1. Lookup registry
    // 2. Get status via factory.CreateManager()
    // 3. Render output
}

func (h *RegistryHandler) List() error {
    // 1. Get all registries from DB
    // 2. Get status for each
    // 3. Render table
}
```

### Phase 4: Manager Pool (Optional)

For efficiency, cache manager instances:

```go
type ManagerPool struct {
    factory  *registry.ServiceFactory
    managers map[int]registry.ServiceManager  // registryID -> manager
    mu       sync.RWMutex
}

func (p *ManagerPool) GetManager(reg *models.Registry) (registry.ServiceManager, error) {
    p.mu.RLock()
    if manager, ok := p.managers[reg.ID]; ok {
        p.mu.RUnlock()
        return manager, nil
    }
    p.mu.RUnlock()

    // Create new manager
    p.mu.Lock()
    defer p.mu.Unlock()

    manager, err := p.factory.CreateManager(reg)
    if err != nil {
        return nil, err
    }

    p.managers[reg.ID] = manager
    return manager, nil
}
```

### Phase 5: Lifecycle Management

Implement lifecycle policies:

```go
// On DVM startup
func StartPersistentRegistries(datastore db.DataStore, factory *registry.ServiceFactory) error {
    registries, _ := datastore.ListRegistries()
    for _, reg := range registries {
        if reg.Lifecycle == "persistent" {
            manager, _ := factory.CreateManager(reg)
            manager.Start(context.Background())
        }
    }
}

// On workspace start (if lifecycle is "on-demand")
func EnsureRegistryRunning(datastore db.DataStore, factory *registry.ServiceFactory, name string) error {
    registry, _ := datastore.GetRegistryByName(name)
    if registry.Lifecycle == "on-demand" {
        manager, _ := factory.CreateManager(registry)
        return manager.Start(context.Background())
    }
    return nil
}
```

## 📊 Testing Strategy

### Unit Tests
- [x] ServiceManager interface tests
- [x] Strategy tests (Zot, Athens, stubs)
- [x] Factory tests
- [ ] DataStore registry methods tests
- [ ] CLI command tests (with mock datastore)
- [ ] Resource handler tests

### Integration Tests
- [ ] Full lifecycle test (create registry in DB → start → stop → delete)
- [ ] Multiple registry types test
- [ ] Lifecycle policy test (persistent, on-demand, manual)
- [ ] Configuration merge test
- [ ] Port conflict test

### Manual Testing
- [ ] `dvm registry create -f zot.yaml`
- [ ] `dvm registry start my-zot`
- [ ] `dvm registry status my-zot`
- [ ] `dvm registry list`
- [ ] `dvm registry stop my-zot`
- [ ] `dvm registry delete my-zot`

## 🎯 Success Criteria

- ✅ ServiceManager interface implemented
- ✅ Strategy pattern implemented for all types
- ✅ Factory creates managers from Registry resources
- ✅ All unit tests pass
- [ ] Database methods implemented
- [ ] CLI commands updated
- [ ] Resource handler implemented
- [ ] Integration tests pass
- [ ] Manual testing complete

## 📝 Notes

### Design Decisions

1. **Adapter Pattern for Athens**: AthensManager has slightly different interface (returns full URLs), so we use an adapter to conform to ServiceManager.

2. **Stub Strategies**: For registry types not yet implemented (devpi, verdaccio, squid), we provide stub strategies that:
   - Return correct defaults
   - Validate configuration
   - Return "not implemented" error on CreateManager()

3. **Storage Path Logic**: Strategies check Registry.Config for custom storage path, otherwise use `~/.devopsmaestro/registries/{name}`.

4. **No Dependency Injection Yet**: For simplicity, strategies create real BinaryManager and ProcessManager instances. Future enhancement: make these injectable.

### Future Enhancements

1. **Health Checks**: Add `Health() error` to ServiceManager
2. **Metrics**: Track start time, uptime, request count
3. **Auto-discovery**: Detect running registries not in DB
4. **Hot Reload**: Support config changes without restart
5. **Web UI**: Dashboard showing all registries and their status

### Breaking Changes

None! This is additive:
- Existing ZotManager, AthensManager continue to work
- New ServiceManager interface is a superset of their methods
- CLI can be updated incrementally

## 🚀 Deployment

No deployment needed yet - this is just the foundation. Once CLI integration is complete, we can:

1. Update CHANGELOG.md
2. Update README.md (add Registry resource documentation)
3. Create migration guide (if any breaking changes)
4. Release as part of v0.19.0 or v0.20.0
