// Package sync provides a strategy pattern for syncing plugins from external sources.
//
// This package enables DevOpsMaestro's nvp tool to import plugin configurations from
// popular Neovim distributions like LazyVim, AstroNvim, NvChad, and others.
//
// # Architecture
//
// The package follows the Interface → Implementation → Factory pattern:
//   - SourceHandler interface defines the contract
//   - Implementations handle specific sources (LazyVim, AstroNvim, etc.)
//   - SourceHandlerFactory creates handlers by name
//   - SourceRegistry manages registration and discovery
//
// # Basic Usage
//
//	// Get factory and create handler
//	factory := sync.NewSourceHandlerFactory()
//	handler, err := factory.CreateHandler("lazyvim")
//	if err != nil {
//	    return err
//	}
//
//	// Configure sync options
//	opts := sync.NewSyncOptions().
//	    WithFilter("category", "lang").
//	    WithTargetDir("~/.config/nvp/plugins").
//	    DryRun(false).
//	    Build()
//
//	// Perform sync
//	result, err := handler.Sync(ctx, opts)
//	if err != nil {
//	    return err
//	}
//
// # Extending with New Sources
//
// To add support for a new source:
//
//	type MySourceHandler struct{}
//
//	func (h *MySourceHandler) Name() string { return "mysource" }
//	func (h *MySourceHandler) Description() string { return "My custom source" }
//	func (h *MySourceHandler) Sync(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
//	    // Implementation
//	}
//	func (h *MySourceHandler) ListAvailable(ctx context.Context) ([]AvailablePlugin, error) {
//	    // Implementation
//	}
//	func (h *MySourceHandler) Validate(ctx context.Context) error {
//	    // Implementation
//	}
//
//	// Register
//	registration := sync.HandlerRegistration{
//	    Name: "mysource",
//	    Info: sync.SourceInfo{
//	        Name:        "mysource",
//	        Description: "My custom source",
//	        Type:        string(sync.SourceTypeGitHub),
//	    },
//	    CreateFunc: func() sync.SourceHandler {
//	        return &MySourceHandler{}
//	    },
//	}
//
//	err := sync.RegisterGlobalSource(registration)
//
// # Builtin Sources
//
// The package includes metadata for common sources:
//   - lazyvim - LazyVim configuration
//   - astronvim - AstroNvim configuration
//   - nvchad - NvChad configuration
//   - kickstart - Kickstart.nvim configuration
//   - lunarvim - LunarVim configuration
//   - local - Local filesystem source
//
// Note: Source metadata is registered by default, but handlers must be
// implemented separately. Unimplemented sources use NotImplementedHandler
// which provides helpful error messages.
package sync
