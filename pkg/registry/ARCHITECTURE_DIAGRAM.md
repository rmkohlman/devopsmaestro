# ServiceManager Architecture Diagram

## Complete System Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          USER INTERACTION                                 │
│                                                                           │
│  dvm registry start my-cache                                              │
│  dvm registry stop my-cache                                               │
│  dvm registry status my-cache                                             │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          CLI COMMAND LAYER                                │
│                         cmd/registry.go                                   │
│                                                                           │
│  • Parse command and flags                                                │
│  • Get DataStore from context                                             │
│  • Delegate to Resource Handler                                           │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        RESOURCE HANDLER LAYER                             │
│                     pkg/resource/registry_handler.go                      │
│                                                                           │
│  • Implements ResourceHandler interface                                   │
│  • Orchestrates DataStore + ServiceFactory                                │
│  • Handles YAML parsing and validation                                    │
└────────────────────────────────┬──────────────────────────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    ▼                         ▼
┌──────────────────────────────┐   ┌──────────────────────────────┐
│       DATABASE LAYER          │   │     SERVICE FACTORY          │
│       db/datastore.go         │   │  pkg/registry/factory.go     │
│                               │   │                              │
│  Registry CRUD operations:    │   │  • GetStrategy(type)         │
│  • CreateRegistry()           │   │  • CreateManager(registry)   │
│  • GetRegistryByName()        │   │  • Validates config          │
│  • UpdateRegistry()           │   │  • Applies defaults          │
│  • DeleteRegistry()           │   │                              │
│  • ListRegistries()           │   │                              │
│                               │   │                              │
│  Returns: models.Registry     │   │                              │
└───────────┬──────────────────┘   └─────────────┬────────────────┘
            │                                     │
            │ models.Registry                     │
            │ {                                   │
            │   Name: "my-cache"                  │
            │   Type: "zot"                       │
            │   Port: 5001                        │
            │   Config: {...}                     │
            │ }                                   │
            │                                     │
            └──────────────┬──────────────────────┘
                           │
                           ▼
          ┌────────────────────────────────────────┐
          │         REGISTRY STRATEGY              │
          │     pkg/registry/strategy.go           │
          │                                        │
          │  ZotStrategy                           │
          │  • ValidateConfig()                    │
          │  • GetDefaultPort() → 5000             │
          │  • GetDefaultStorage()                 │
          │  • CreateManager() → ServiceManager    │
          │                                        │
          │  AthensStrategy                        │
          │  • ValidateConfig()                    │
          │  • GetDefaultPort() → 3000             │
          │  • GetDefaultStorage()                 │
          │  • CreateManager() → ServiceManager    │
          │                                        │
          │  StubStrategy (devpi, verdaccio, squid)│
          │  • Returns "not implemented" error     │
          └─────────────────┬──────────────────────┘
                            │
                            ▼
          ┌────────────────────────────────────────┐
          │         SERVICE MANAGER                │
          │   pkg/registry/service_manager.go      │
          │                                        │
          │  Interface:                            │
          │  • Start(ctx) error                    │
          │  • Stop(ctx) error                     │
          │  • IsRunning(ctx) bool                 │
          │  • GetEndpoint() string                │
          └─────────────────┬──────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
┌─────────────────────────┐   ┌─────────────────────────┐
│      ZotManager         │   │  AthensManagerAdapter   │
│  pkg/registry/zot_*.go  │   │  pkg/registry/athens_*  │
│                         │   │                         │
│  • BinaryManager        │   │  Wraps:                 │
│  • ProcessManager       │   │  • AthensManager        │
│  • RegistryConfig       │   │  • GoModuleConfig       │
│                         │   │                         │
│  Implements:            │   │  Implements:            │
│  RegistryManager        │   │  GoModuleProxy          │
│  (already exists)       │   │  (already exists)       │
│                         │   │                         │
│  Also implements:       │   │  Adapter implements:    │
│  ServiceManager         │   │  ServiceManager         │
│  (no changes needed!)   │   │  (bridges interface)    │
└─────────────────────────┘   └─────────────────────────┘
```

## Strategy Pattern Detail

```
ServiceFactory.CreateManager(registry)
         ↓
   GetStrategy(registry.Type)
         ↓
┌─────────────────────────────────────────┐
│         Strategy Selection              │
│                                         │
│  switch registry.Type {                 │
│    case "zot":     → ZotStrategy        │
│    case "athens":  → AthensStrategy     │
│    case "devpi":   → DevpiStrategy      │
│    case "verdaccio": → VerdaccioStrategy│
│    case "squid":   → SquidStrategy      │
│    default: error                       │
│  }                                      │
└─────────────────────┬───────────────────┘
                      │
                      ▼
         strategy.CreateManager(registry)
                      │
                      ▼
         ┌────────────────────────────┐
         │  1. ValidateConfig()       │
         │  2. Apply defaults         │
         │  3. Convert Registry →     │
         │     RegistryConfig or      │
         │     GoModuleConfig         │
         │  4. Create manager         │
         │  5. Return ServiceManager  │
         └────────────────────────────┘
```

## Data Flow Example

### Starting a Registry

```
User: dvm registry start my-cache

1. CLI parses command
   ├─ registryName = "my-cache"
   └─ action = "start"

2. Resource Handler
   ├─ registry ← datastore.GetRegistryByName("my-cache")
   └─ Returns: Registry{Type: "zot", Port: 5001, ...}

3. Service Factory
   ├─ strategy ← GetStrategy("zot")
   └─ Returns: ZotStrategy

4. Strategy
   ├─ ValidateConfig(registry.Config)
   ├─ Convert Registry → RegistryConfig
   │  • Port: 5001 (from registry)
   │  • Storage: ~/.devopsmaestro/registries/my-cache
   │  • IdleTimeout: 30m (default)
   │  • Lifecycle: "on-demand" (from registry)
   └─ Create ZotManager(config)

5. Service Manager (ZotManager)
   ├─ manager.Start(ctx)
   │  • Download binary if needed
   │  • Generate Zot config.json
   │  • Start process
   │  • Wait for health check
   │  • Setup idle timer (on-demand)
   └─ Returns: nil (success)

6. Update Database
   ├─ registry.Status = "running"
   └─ datastore.UpdateRegistry(registry)

7. User Feedback
   └─ "✓ Registry 'my-cache' started at localhost:5001"
```

## Component Responsibilities

### ServiceManager
- **Lifecycle management**: Start, Stop, IsRunning
- **Endpoint information**: GetEndpoint()
- **Runtime agnostic**: Works for any registry type

### RegistryStrategy
- **Type-specific logic**: Each type has its own strategy
- **Configuration validation**: Ensures config is valid for type
- **Manager creation**: Converts Registry → ServiceManager
- **Defaults**: Provides sensible defaults for each type

### ServiceFactory
- **Strategy registry**: Maintains map of type → strategy
- **Delegation**: Routes to correct strategy
- **Validation**: Ensures Registry is valid before delegation

### Resource Handler
- **Orchestration**: Coordinates DataStore + ServiceFactory
- **YAML handling**: Parses and validates YAML
- **Error handling**: User-friendly error messages
- **Output rendering**: Pretty-prints status, lists, etc.

### DataStore
- **Persistence**: Stores Registry resources in database
- **CRUD operations**: Create, Read, Update, Delete
- **Queries**: Efficient lookups by name, type, status

## Future Extensions

### Adding a New Registry Type (e.g., Harbor)

```
1. Create Strategy
   ├─ harbor_strategy.go
   │  • Implement RegistryStrategy
   │  • GetDefaultPort() → 8080
   │  • GetDefaultStorage() → /var/lib/harbor
   │  • CreateManager() → HarborManager
   │
2. Create Manager
   ├─ harbor_manager.go
   │  • Implement ServiceManager
   │  • Start() - download Harbor, start containers
   │  • Stop() - stop containers
   │  • IsRunning() - check container status
   │  • GetEndpoint() - return URL
   │
3. Register Strategy
   ├─ service_factory.go
   │  • Add "harbor": NewHarborStrategy()
   │
4. Add to Models
   ├─ models/registry.go
   │  • Add "harbor" to validRegistryTypes
   │  • Add harbor defaults
   │
5. Create Migration
   └─ migrations/sqlite/012_harbor_defaults.sql (if needed)

Total effort: ~2-4 hours
```

### Manager Pool Extension

```
┌──────────────────────────────┐
│        Manager Pool          │
│   Caches ServiceManager      │
│      instances by ID         │
│                              │
│  managers: map[int]Manager   │
│                              │
│  • GetManager(registry)      │
│  • ReleaseManager(id)        │
│  • ShutdownAll()             │
└──────────────────────────────┘
         ↑                ↓
         │                │
    Cache hit          Cache miss
         │                │
         │                ▼
         │     ServiceFactory.CreateManager()
         │                │
         └────────────────┘
```

## Benefits of This Architecture

1. **Separation of Concerns**
   - CLI: User interaction
   - Handler: Orchestration
   - DataStore: Persistence
   - Factory: Creation logic
   - Strategy: Type-specific behavior
   - Manager: Runtime lifecycle

2. **Testability**
   - Each layer can be tested independently
   - Mock implementations at each boundary
   - Integration tests verify full flow

3. **Extensibility**
   - Add new registry types without modifying existing code
   - Open/Closed Principle: Open for extension, closed for modification

4. **Type Safety**
   - Compile-time interface checks
   - No reflection or type assertions

5. **Performance**
   - Manager pool can cache instances
   - Lazy loading (create manager only when needed)
   - Efficient database queries

6. **Maintainability**
   - Clear boundaries between components
   - Single Responsibility Principle
   - Easy to locate and fix bugs
