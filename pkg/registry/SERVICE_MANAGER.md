# ServiceManager Architecture

## Overview

The ServiceManager system provides a unified interface for managing registry runtime services. It bridges the gap between database-persisted Registry resources and their runtime service managers (ZotManager, AthensManager, etc.).

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Command Layer                          │
│              dvm registry start <name>                        │
└────────────────────────┬──────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     DataStore                                │
│          GetRegistryByName("my-cache")                       │
│                ↓                                             │
│         models.Registry{                                     │
│           Type: "zot",                                       │
│           Port: 5001,                                        │
│           Config: {...}                                      │
│         }                                                    │
└────────────────────────┬──────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   ServiceFactory                             │
│         factory.CreateManager(registry)                      │
│                ↓                                             │
│         GetStrategy(registry.Type)                           │
└────────────────────────┬──────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  RegistryStrategy                            │
│         strategies["zot"].CreateManager(registry)            │
│                ↓                                             │
│         - Validate config                                    │
│         - Apply defaults                                     │
│         - Create type-specific manager                       │
└────────────────────────┬──────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   ServiceManager                             │
│              (ZotManager, AthensManager, etc.)               │
│                ↓                                             │
│         manager.Start(ctx)                                   │
│         manager.Stop(ctx)                                    │
│         manager.IsRunning(ctx)                               │
│         manager.GetEndpoint()                                │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. ServiceManager Interface

The unified interface all runtime managers implement:

```go
type ServiceManager interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    IsRunning(ctx context.Context) bool
    GetEndpoint() string
}
```

**Implementations:**
- `ZotManager` - Directly implements ServiceManager
- `AthensManagerAdapter` - Wraps AthensManager to implement ServiceManager

### 2. RegistryStrategy Interface

The strategy pattern for type-specific behavior:

```go
type RegistryStrategy interface {
    ValidateConfig(config json.RawMessage) error
    CreateManager(reg *models.Registry) (ServiceManager, error)
    GetDefaultPort() int
    GetDefaultStorage() string
}
```

**Implementations:**
- `ZotStrategy` - For Zot container registry
- `AthensStrategy` - For Athens Go module proxy
- `StubStrategy` - Base for unimplemented types (devpi, verdaccio, squid)

### 3. ServiceFactory

Creates ServiceManager instances based on registry type:

```go
factory := NewServiceFactory()

// Get strategy
strategy, _ := factory.GetStrategy("zot")

// Create manager from Registry resource
manager, _ := factory.CreateManager(registry)

// Use manager
manager.Start(ctx)
endpoint := manager.GetEndpoint()
```

## Integration Flow

### CLI to Runtime

```go
// In CLI command handler
func startRegistry(name string) error {
    // 1. Lookup registry from database
    registry, err := datastore.GetRegistryByName(name)
    if err != nil {
        return err
    }

    // 2. Create service manager
    factory := registry.NewServiceFactory()
    manager, err := factory.CreateManager(registry)
    if err != nil {
        return err
    }

    // 3. Start the service
    ctx := context.Background()
    if err := manager.Start(ctx); err != nil {
        return err
    }

    // 4. Get endpoint for user
    endpoint := manager.GetEndpoint()
    fmt.Printf("Registry started: %s\n", endpoint)

    // 5. Update database status
    registry.Status = "running"
    return datastore.UpdateRegistry(registry)
}
```

## Strategy Pattern Details

### Zot Strategy

```go
strategy := NewZotStrategy()

// Defaults
port := strategy.GetDefaultPort()           // 5000
storage := strategy.GetDefaultStorage()     // /var/lib/zot

// Create manager
manager, _ := strategy.CreateManager(&models.Registry{
    Name: "my-zot",
    Type: "zot",
    Port: 5001,
})

// Manager is ready to use
manager.Start(ctx)
```

### Athens Strategy

```go
strategy := NewAthensStrategy()

// Defaults
port := strategy.GetDefaultPort()           // 3000
storage := strategy.GetDefaultStorage()     // /var/lib/athens

// Create manager
manager, _ := strategy.CreateManager(&models.Registry{
    Name: "my-athens",
    Type: "athens",
    Port: 3001,
})

// Manager returns full URL
endpoint := manager.GetEndpoint()  // http://localhost:3001
```

### Stub Strategies (Future)

```go
// devpi, verdaccio, squid return errors for now
strategy := NewDevpiStrategy()
_, err := strategy.CreateManager(registry)
// err: "devpi registry not implemented yet"
```

## Configuration Handling

### Registry Resource → Manager Config

```yaml
# Registry YAML
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: my-cache
spec:
  type: zot
  port: 5001
  lifecycle: on-demand
  config:
    storage: /custom/path
    idleTimeout: 15m
```

```go
// Database model
registry := &models.Registry{
    Name: "my-cache",
    Type: "zot",
    Port: 5001,
    Lifecycle: "on-demand",
    Config: sql.NullString{
        Valid: true,
        String: `{"storage": "/custom/path", "idleTimeout": "15m"}`,
    },
}

// Strategy creates manager
manager, _ := factory.CreateManager(registry)

// Manager has merged config:
// - Port: 5001 (from registry)
// - Storage: /custom/path (from config)
// - IdleTimeout: 15m (from config)
// - Other settings: defaults
```

## Extensibility

### Adding a New Registry Type

1. **Create Strategy:**
```go
type DevpiStrategy struct{}

func NewDevpiStrategy() *DevpiStrategy {
    return &DevpiStrategy{}
}

func (s *DevpiStrategy) ValidateConfig(config json.RawMessage) error {
    // Validate devpi-specific config
}

func (s *DevpiStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
    // Create DevpiManager
}

func (s *DevpiStrategy) GetDefaultPort() int {
    return 3141
}

func (s *DevpiStrategy) GetDefaultStorage() string {
    return "/var/lib/devpi"
}
```

2. **Register Strategy:**
```go
// In NewServiceFactory()
strategies: map[string]RegistryStrategy{
    "zot":    NewZotStrategy(),
    "athens": NewAthensStrategy(),
    "devpi":  NewDevpiStrategy(), // Add here
}
```

3. **Implement Manager:**
```go
type DevpiManager struct {
    // ...
}

// Implement ServiceManager interface
func (m *DevpiManager) Start(ctx context.Context) error { ... }
func (m *DevpiManager) Stop(ctx context.Context) error { ... }
func (m *DevpiManager) IsRunning(ctx context.Context) bool { ... }
func (m *DevpiManager) GetEndpoint() string { ... }
```

## Testing

### Unit Tests

```go
// Test strategy
func TestDevpiStrategy(t *testing.T) {
    strategy := NewDevpiStrategy()
    
    // Test defaults
    assert.Equal(t, 3141, strategy.GetDefaultPort())
    
    // Test manager creation
    manager, err := strategy.CreateManager(registry)
    assert.NoError(t, err)
    assert.NotNil(t, manager)
}

// Test factory integration
func TestServiceFactory_Devpi(t *testing.T) {
    factory := NewServiceFactory()
    
    registry := &models.Registry{
        Name: "test-devpi",
        Type: "devpi",
        Port: 3141,
    }
    
    manager, err := factory.CreateManager(registry)
    assert.NoError(t, err)
    assert.NotNil(t, manager)
}
```

### Integration Tests

```go
func TestRegistryLifecycle(t *testing.T) {
    // Create registry in database
    registry := &models.Registry{
        Name: "test-zot",
        Type: "zot",
        Port: 5099, // Test port
    }
    err := datastore.CreateRegistry(registry)
    require.NoError(t, err)
    
    // Create manager via factory
    factory := registry.NewServiceFactory()
    manager, err := factory.CreateManager(registry)
    require.NoError(t, err)
    
    // Test lifecycle
    ctx := context.Background()
    
    assert.False(t, manager.IsRunning(ctx))
    
    err = manager.Start(ctx)
    assert.NoError(t, err)
    assert.True(t, manager.IsRunning(ctx))
    
    endpoint := manager.GetEndpoint()
    assert.Contains(t, endpoint, "5099")
    
    err = manager.Stop(ctx)
    assert.NoError(t, err)
    assert.False(t, manager.IsRunning(ctx))
}
```

## Benefits

1. **Decoupling**: CLI doesn't know about ZotManager, AthensManager details
2. **Extensibility**: Add new registry types by implementing RegistryStrategy
3. **Consistency**: All registries have the same lifecycle interface
4. **Testability**: Easy to mock ServiceManager for tests
5. **Type Safety**: Compile-time guarantees via interfaces
6. **Single Responsibility**: Each component has one job

## Future Enhancements

1. **Dependency Injection**: Make binary/process managers injectable
2. **Manager Pool**: Reuse manager instances per registry
3. **Health Checks**: Add Health() method to ServiceManager
4. **Metrics**: Track start/stop times, usage statistics
5. **Auto-discovery**: Detect running registries not in database
6. **Hot Reload**: Support config changes without restart
