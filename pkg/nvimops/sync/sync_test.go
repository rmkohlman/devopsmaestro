package sync

import (
	"context"
	"testing"
)

func TestSyncOptionsBuilder(t *testing.T) {
	// Test Builder pattern
	opts := NewSyncOptions().
		WithFilter("category", "lang").
		WithFilter("enabled", "true").
		DryRun(true).
		WithTargetDir("/tmp/plugins").
		Overwrite(true).
		Build()

	if !opts.DryRun {
		t.Error("Expected DryRun to be true")
	}

	if opts.GetFilter("category") != "lang" {
		t.Error("Expected category filter to be 'lang'")
	}

	if opts.GetFilter("enabled") != "true" {
		t.Error("Expected enabled filter to be 'true'")
	}

	if opts.TargetDir != "/tmp/plugins" {
		t.Error("Expected TargetDir to be '/tmp/plugins'")
	}

	if !opts.Overwrite {
		t.Error("Expected Overwrite to be true")
	}
}

func TestSyncOptionsFiltering(t *testing.T) {
	opts := NewSyncOptions().
		WithFilter("category", "lang").
		Build()

	// Test plugin that matches
	plugin1 := AvailablePlugin{
		Name:     "nvim-lspconfig",
		Category: "lang",
		Repo:     "neovim/nvim-lspconfig",
	}

	if !opts.MatchesAvailablePlugin(plugin1) {
		t.Error("Expected plugin1 to match filter")
	}

	// Test plugin that doesn't match
	plugin2 := AvailablePlugin{
		Name:     "telescope",
		Category: "navigation",
		Repo:     "nvim-telescope/telescope.nvim",
	}

	if opts.MatchesAvailablePlugin(plugin2) {
		t.Error("Expected plugin2 to not match filter")
	}
}

func TestSyncResult(t *testing.T) {
	result := &SyncResult{
		SourceName: "test-source",
	}

	// Test adding plugins
	result.AddPluginCreated("telescope")
	result.AddPluginCreated("lspconfig")
	result.AddPluginUpdated("treesitter")

	if len(result.PluginsCreated) != 2 {
		t.Errorf("Expected 2 plugins created, got %d", len(result.PluginsCreated))
	}

	if len(result.PluginsUpdated) != 1 {
		t.Errorf("Expected 1 plugin updated, got %d", len(result.PluginsUpdated))
	}

	if result.TotalSynced != 3 {
		t.Errorf("Expected TotalSynced to be 3, got %d", result.TotalSynced)
	}

	// Test summary
	summary := result.Summary()
	expected := "2 created, 1 updated"
	if summary != expected {
		t.Errorf("Expected summary '%s', got '%s'", expected, summary)
	}
}

func TestSourceRegistry(t *testing.T) {
	registry := NewSourceRegistry()

	// Test registration
	registration := HandlerRegistration{
		Name: "test-source",
		Info: SourceInfo{
			Name:        "test-source",
			Description: "A test source",
			Type:        string(SourceTypeLocal),
		},
		CreateFunc: func() SourceHandler {
			return &NotImplementedHandler{info: SourceInfo{Name: "test-source"}}
		},
	}

	err := registry.Register(registration)
	if err != nil {
		t.Fatalf("Failed to register source: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(registration)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test retrieval
	retrieved, exists := registry.GetRegistration("test-source")
	if !exists {
		t.Error("Expected source to exist after registration")
	}

	if retrieved.Name != "test-source" {
		t.Errorf("Expected name 'test-source', got '%s'", retrieved.Name)
	}

	// Test listing
	sources := registry.ListSources()
	if len(sources) != 1 || sources[0] != "test-source" {
		t.Errorf("Expected sources ['test-source'], got %v", sources)
	}
}

func TestSourceHandlerFactory(t *testing.T) {
	registry := NewSourceRegistry()
	factory := NewSourceHandlerFactoryWithRegistry(registry)

	// Test with empty registry
	sources := factory.ListSources()
	if len(sources) != 0 {
		t.Errorf("Expected empty source list, got %v", sources)
	}

	// Register a test source
	registration := HandlerRegistration{
		Name: "test-source",
		Info: SourceInfo{
			Name:        "test-source",
			Description: "A test source",
			Type:        string(SourceTypeLocal),
		},
		CreateFunc: func() SourceHandler {
			return &NotImplementedHandler{info: SourceInfo{Name: "test-source"}}
		},
	}

	err := registry.Register(registration)
	if err != nil {
		t.Fatalf("Failed to register source: %v", err)
	}

	// Test factory methods
	if !factory.IsSupported("test-source") {
		t.Error("Expected test-source to be supported")
	}

	handler, err := factory.CreateHandler("test-source")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	if handler.Name() != "test-source" {
		t.Errorf("Expected handler name 'test-source', got '%s'", handler.Name())
	}

	// Test unsupported source
	_, err = factory.CreateHandler("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent source")
	}
}

func TestNotImplementedHandler(t *testing.T) {
	info := SourceInfo{
		Name:        "test-source",
		Description: "A test source",
	}

	handler := &NotImplementedHandler{info: info}

	if handler.Name() != "test-source" {
		t.Errorf("Expected name 'test-source', got '%s'", handler.Name())
	}

	if handler.Description() != "A test source" {
		t.Errorf("Expected description 'A test source', got '%s'", handler.Description())
	}

	// Test that all methods return not implemented errors
	ctx := context.Background()

	_, err := handler.Sync(ctx, SyncOptions{})
	if err == nil {
		t.Error("Expected Sync to return error")
	}

	_, err = handler.ListAvailable(ctx)
	if err == nil {
		t.Error("Expected ListAvailable to return error")
	}

	err = handler.Validate(ctx)
	if err == nil {
		t.Error("Expected Validate to return error")
	}
}
