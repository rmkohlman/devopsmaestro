# NvimOps Sync Package

The `sync` package provides a strategy pattern for syncing plugins from external sources like LazyVim, AstroNvim, NvChad, and others. This enables `nvp sync <source>` commands to import plugin configurations from popular Neovim distributions.

## Architecture

The package follows the **Interface → Implementation → Factory** pattern from DevOpsMaestro's standards:

1. **SourceHandler Interface** - Defines the contract for syncing from external sources
2. **SourceHandlerFactory Interface** - Creates handler instances 
3. **SourceRegistry** - Manages registration and discovery of sources
4. **Builder Pattern** - Fluent interface for creating SyncOptions

## Core Types

### SourceHandler Interface

```go
type SourceHandler interface {
    Name() string
    Description() string
    Sync(ctx context.Context, options SyncOptions) (*SyncResult, error)
    ListAvailable(ctx context.Context) ([]AvailablePlugin, error)
    Validate(ctx context.Context) error
}
```

### SyncOptions (Builder Pattern)

```go
opts := sync.NewSyncOptions().
    WithFilter("category", "lang").
    WithTargetDir("~/.config/nvp/plugins").
    DryRun(true).
    Overwrite(false).
    Build()
```

### SyncResult

```go
type SyncResult struct {
    PluginsCreated  []string
    PluginsUpdated  []string
    PackagesCreated []string
    PackagesUpdated []string
    Errors         []error
    SourceName     string
    TotalAvailable int
    TotalSynced    int
}
```

## Usage Example

```go
// Get factory and create handler
factory := sync.NewSourceHandlerFactory()
handler, err := factory.CreateHandler("lazyvim")
if err != nil {
    return err
}

// Configure sync options
opts := sync.NewSyncOptions().
    WithFilter("category", "lang").
    WithTargetDir("~/.config/nvp/plugins").
    DryRun(false).
    Build()

// Perform sync
result, err := handler.Sync(ctx, opts)
if err != nil {
    return err
}

fmt.Printf("Synced %d plugins (%s)\n", result.TotalSynced, result.Summary())
```

## Builtin Sources

The package includes metadata for common Neovim distributions:

- **lazyvim** - LazyVim configuration
- **astronvim** - AstroNvim configuration  
- **nvchad** - NvChad configuration
- **kickstart** - Kickstart.nvim configuration
- **lunarvim** - LunarVim configuration
- **local** - Local filesystem source

## Extending with New Sources

To add a new source, implement the `SourceHandler` interface and register it:

```go
type MySourceHandler struct {
    // Implementation fields
}

func (h *MySourceHandler) Name() string { return "mysource" }
func (h *MySourceHandler) Description() string { return "My custom source" }
func (h *MySourceHandler) Sync(ctx context.Context, options SyncOptions) (*SyncResult, error) {
    // Implementation
}
func (h *MySourceHandler) ListAvailable(ctx context.Context) ([]AvailablePlugin, error) {
    // Implementation
}
func (h *MySourceHandler) Validate(ctx context.Context) error {
    // Implementation
}

// Register the source
registration := sync.HandlerRegistration{
    Name: "mysource",
    Info: sync.SourceInfo{
        Name:        "mysource",
        Description: "My custom source",
        Type:        string(sync.SourceTypeGitHub),
    },
    CreateFunc: func() sync.SourceHandler {
        return &MySourceHandler{}
    },
}

err := sync.RegisterGlobalSource(registration)
```

## Testing

The package includes comprehensive tests:

```bash
go test ./pkg/nvimops/sync/... -v
```

## Future Implementation

This package provides the foundation. Actual source handlers (LazyVim, AstroNvim, etc.) will be implemented in separate packages or files, following the same pattern as the `NotImplementedHandler` placeholder.

Each source handler will:

1. Fetch plugin configurations from the external source (GitHub API, file parsing, etc.)
2. Convert them to DevOpsMaestro's YAML plugin format
3. Write the YAML files to the target directory
4. Report sync results

The `nvp sync` command will use this package to provide a unified interface for importing from any supported source.