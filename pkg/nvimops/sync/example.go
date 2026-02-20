package sync

import (
	"context"
	"fmt"
)

// Example demonstrates how the sync package would be used in the nvp command.
// This is not production code, just an illustration of the intended usage.

// CustomSourceHandler is an example implementation for demonstration
type CustomSourceHandler struct {
	name string
}

func (h *CustomSourceHandler) Name() string                       { return h.name }
func (h *CustomSourceHandler) Description() string                { return "Custom source example" }
func (h *CustomSourceHandler) Validate(ctx context.Context) error { return nil }
func (h *CustomSourceHandler) ListAvailable(ctx context.Context) ([]AvailablePlugin, error) {
	return []AvailablePlugin{
		{
			Name:        "example-plugin",
			Description: "An example plugin",
			Category:    "example",
			Repo:        "user/example-plugin",
			SourceName:  h.name,
		},
	}, nil
}
func (h *CustomSourceHandler) Sync(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{
		SourceName:     h.name,
		TotalAvailable: 1,
	}

	// Simulate plugin creation
	if !opts.DryRun {
		// Would write YAML file here
		result.AddPluginCreated("example-plugin")
	}

	return result, nil
}

// ExampleSyncCommand shows how `nvp sync lazyvim --category=lang --dry-run` would work
func ExampleSyncCommand() error {
	// Initialize the global registry (done once at startup)
	if err := InitializeGlobalRegistry(); err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}

	// Create factory and handler
	factory := NewSourceHandlerFactory()
	handler, err := factory.CreateHandler("lazyvim")
	if err != nil {
		return fmt.Errorf("failed to create lazyvim handler: %w", err)
	}

	// Validate source is accessible
	ctx := context.Background()
	if err := handler.Validate(ctx); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	// Build sync options from command flags
	opts := NewSyncOptions().
		WithFilter("category", "lang").         // --category=lang
		WithTargetDir("~/.config/nvp/plugins"). // default or --target-dir
		DryRun(true).                           // --dry-run
		Overwrite(false).                       // --no-overwrite
		Build()

	// Perform the sync
	result, err := handler.Sync(ctx, opts)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Report results
	fmt.Printf("Sync completed: %s\n", result.Summary())
	fmt.Printf("Source: %s\n", result.SourceName)
	fmt.Printf("Available: %d, Synced: %d\n", result.TotalAvailable, result.TotalSynced)

	if len(result.PluginsCreated) > 0 {
		fmt.Printf("Created: %v\n", result.PluginsCreated)
	}
	if len(result.PluginsUpdated) > 0 {
		fmt.Printf("Updated: %v\n", result.PluginsUpdated)
	}
	if result.HasErrors() {
		fmt.Printf("Errors: %d\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	return nil
}

// ExampleListSources shows how `nvp sync list` would work
func ExampleListSources() error {
	if err := InitializeGlobalRegistry(); err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}

	factory := NewSourceHandlerFactory()
	sources := factory.ListSources()

	fmt.Println("Available sources:")
	for _, source := range sources {
		info, err := factory.GetHandlerInfo(source)
		if err != nil {
			fmt.Printf("  %s - (info unavailable)\n", source)
			continue
		}

		status, err := GetSourceStatus(source)
		if err != nil {
			fmt.Printf("  %s - %s (status unknown)\n", source, info.Description)
			continue
		}

		statusText := "not implemented"
		if status.IsImplemented {
			statusText = "ready"
		}

		fmt.Printf("  %s - %s (%s)\n", source, info.Description, statusText)
	}

	return nil
}

// ExampleListAvailable shows how `nvp sync list-plugins lazyvim` would work
func ExampleListAvailable() error {
	if err := InitializeGlobalRegistry(); err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}

	factory := NewSourceHandlerFactory()
	handler, err := factory.CreateHandler("lazyvim")
	if err != nil {
		return fmt.Errorf("failed to create handler: %w", err)
	}

	ctx := context.Background()
	plugins, err := handler.ListAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	fmt.Printf("Available plugins from %s:\n", handler.Name())
	for _, plugin := range plugins {
		fmt.Printf("  %s - %s (%s)\n", plugin.Name, plugin.Description, plugin.Category)
		if plugin.Repo != "" {
			fmt.Printf("    Repo: %s\n", plugin.Repo)
		}
		if len(plugin.Dependencies) > 0 {
			fmt.Printf("    Dependencies: %v\n", plugin.Dependencies)
		}
	}

	return nil
}

// ExampleCustomSource shows how to register a custom source
func ExampleCustomSource() error {
	// Register the custom source
	registration := HandlerRegistration{
		Name: "custom",
		Info: SourceInfo{
			Name:        "custom",
			Description: "Custom source example",
			URL:         "https://example.com",
			Type:        string(SourceTypeRemote),
		},
		CreateFunc: func() SourceHandler {
			return &CustomSourceHandler{name: "custom"}
		},
	}

	err := RegisterGlobalSource(registration)
	if err != nil {
		return fmt.Errorf("failed to register custom source: %w", err)
	}

	// Now the custom source can be used like any other
	factory := NewSourceHandlerFactory()
	handler, err := factory.CreateHandler("custom")
	if err != nil {
		return fmt.Errorf("failed to create custom handler: %w", err)
	}

	fmt.Printf("Custom source registered: %s - %s\n",
		handler.Name(), handler.Description())

	return nil
}
