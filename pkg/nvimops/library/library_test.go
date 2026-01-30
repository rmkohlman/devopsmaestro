package library

import (
	"testing"
)

func TestLibrary(t *testing.T) {
	lib, err := NewLibrary()
	if err != nil {
		t.Fatalf("NewLibrary failed: %v", err)
	}

	// Should have plugins loaded
	count := lib.Count()
	if count == 0 {
		t.Error("Library should have plugins loaded")
	}
	t.Logf("Loaded %d plugins from library", count)

	// Test Get
	telescope, ok := lib.Get("telescope")
	if !ok {
		t.Error("Library should have telescope plugin")
	}
	if telescope != nil && telescope.Repo != "nvim-telescope/telescope.nvim" {
		t.Errorf("Telescope repo = %q, want nvim-telescope/telescope.nvim", telescope.Repo)
	}

	// Test List
	plugins := lib.List()
	if len(plugins) != count {
		t.Errorf("List() returned %d plugins, want %d", len(plugins), count)
	}

	// Test Categories
	categories := lib.Categories()
	if len(categories) == 0 {
		t.Error("Library should have categories")
	}
	t.Logf("Categories: %v", categories)

	// Test Tags
	tags := lib.Tags()
	if len(tags) == 0 {
		t.Error("Library should have tags")
	}
	t.Logf("Tags: %v", tags)

	// Test ListByCategory
	lspPlugins := lib.ListByCategory("lsp")
	t.Logf("LSP plugins: %d", len(lspPlugins))

	// Test ListByTag
	finderPlugins := lib.ListByTag("finder")
	t.Logf("Finder plugins: %d", len(finderPlugins))

	// Test Info
	info := lib.Info()
	if len(info) != count {
		t.Errorf("Info() returned %d items, want %d", len(info), count)
	}
}

func TestLibraryPluginContent(t *testing.T) {
	lib, err := NewLibrary()
	if err != nil {
		t.Fatalf("NewLibrary failed: %v", err)
	}

	// Check some expected plugins
	expectedPlugins := []string{"telescope", "treesitter", "lspconfig"}

	for _, name := range expectedPlugins {
		p, ok := lib.Get(name)
		if !ok {
			t.Errorf("Expected plugin %q not found", name)
			continue
		}

		// Basic validation
		if p.Name == "" {
			t.Errorf("Plugin %q has empty name", name)
		}
		if p.Repo == "" {
			t.Errorf("Plugin %q has empty repo", name)
		}
	}
}
