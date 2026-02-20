// Package sources provides initialization for built-in source handlers.
package sources

import "devopsmaestro/pkg/nvimops/sync"

// RegisterAllHandlers registers all available source handlers in the provided registry.
// This replaces placeholder handlers with actual implementations.
func RegisterAllHandlers(registry *sync.SourceRegistry) error {
	// Register LazyVim handler
	if err := RegisterLazyVimHandler(registry); err != nil {
		return err
	}

	// TODO: Register other handlers (AstroNvim, NvChad, etc.) when implemented

	return nil
}

// RegisterAllGlobalHandlers registers all handlers in the global registry.
// This is a convenience function for application initialization.
func RegisterAllGlobalHandlers() error {
	registry := sync.GetGlobalRegistry()
	return RegisterAllHandlers(registry)
}
