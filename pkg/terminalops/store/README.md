# Terminal Plugin Database Adapter

This package provides a database adapter for terminal plugin management, following the nvimops pattern.

## Overview

The `DBPluginStore` implements the `plugin.PluginStore` interface and bridges the domain layer (`pkg/terminalops/plugin/`) with the database layer via the DataStore interface.

## Architecture

```
Domain Layer (pkg/terminalops/plugin/)
    ↓ Plugin struct
Database Adapter (pkg/terminalops/store/)  
    ↓ models.TerminalPluginDB
Database Layer (db/)
    ↓ SQLite/PostgreSQL
```

## Key Components

### DBPluginStore
- **Purpose**: Adapts `db.DataStore` to implement `plugin.PluginStore`
- **Benefits**: 
  - Unified storage location for both nvp and dvm
  - Swappable storage backends
  - Full CRUD operations with proper error handling

### PluginDataStore Interface
```go
type PluginDataStore interface {
    CreateTerminalPlugin(plugin *models.TerminalPluginDB) error
    GetTerminalPlugin(name string) (*models.TerminalPluginDB, error)
    UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error
    UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error
    DeleteTerminalPlugin(name string) error
    ListTerminalPlugins() ([]*models.TerminalPluginDB, error)
    ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error)
    ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error)
}
```

## Model Conversion

The adapter handles conversion between:

### Domain Plugin → Database Model
- **Basic fields**: name, repo, category, description, enabled
- **Manager enum**: Converts `plugin.PluginManager` to string
- **Dependencies**: JSON array serialization
- **Environment variables**: JSON object serialization  
- **Labels/metadata**: Extracts tags, load mode, priority into JSON labels
- **Oh-My-Zsh plugins**: Special handling for built-in plugins

### Database Model → Domain Plugin  
- **Reverse conversion**: All fields converted back to domain types
- **JSON deserialization**: Dependencies, env vars, and labels
- **Oh-My-Zsh extraction**: Detects and extracts plugin names from load commands
- **Type safety**: Proper enum conversion and null handling

## Usage

```go
// Create database connection
dataStore, err := db.NewDataStore(db.DataStoreConfig{
    Driver: driver,
})

// Create adapter
pluginStore := store.NewDBPluginStore(dataStore)
defer pluginStore.Close()

// Use through interface
var store plugin.PluginStore = pluginStore

// Standard CRUD operations
plugin := &plugin.Plugin{
    Name:     "zsh-autosuggestions", 
    Repo:     "zsh-users/zsh-autosuggestions",
    Manager:  plugin.PluginManagerZinit,
    Enabled:  true,
}

err = store.Upsert(plugin)
retrieved, err := store.Get("zsh-autosuggestions")
plugins, err := store.List()
```

## Error Handling

The adapter provides proper error types:
- `plugin.ErrNotFound`: When plugin doesn't exist
- `plugin.ErrAlreadyExists`: When creating duplicate plugins
- Wrapped database errors with context

## Testing

Comprehensive test suite includes:
- **Interface compliance**: Compile-time verification
- **Mock implementation**: For unit testing without database
- **CRUD operations**: Full create, read, update, delete testing
- **Model conversion**: Bidirectional conversion testing
- **Roundtrip testing**: Ensures no data loss in conversions
- **Interface polymorphism**: Testing through `PluginStore` interface

Run tests:
```bash
go test ./pkg/terminalops/store/... -v
```

## Integration

This adapter integrates with:
- **Terminal plugin system**: Primary storage backend
- **Database layer**: Uses existing DataStore interface  
- **Configuration system**: Inherits database configuration
- **Migration system**: Leverages existing migration infrastructure

## Future Enhancements

- **Caching layer**: Add optional caching for read operations
- **Bulk operations**: Batch insert/update operations
- **Search capabilities**: Full-text search across plugin metadata
- **Indexing**: Database indexes for common query patterns