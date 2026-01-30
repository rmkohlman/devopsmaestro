package store

import (
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/pkg/nvimops/plugin"
)

func TestMemoryStore(t *testing.T) {
	testPluginStore(t, func() (PluginStore, func()) {
		return NewMemoryStore(), func() {}
	})
}

func TestFileStore(t *testing.T) {
	testPluginStore(t, func() (PluginStore, func()) {
		tmpDir, err := os.MkdirTemp("", "nvim-manager-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}

		store, err := NewFileStore(tmpDir)
		if err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create FileStore: %v", err)
		}

		cleanup := func() {
			store.Close()
			os.RemoveAll(tmpDir)
		}

		return store, cleanup
	})
}

// testPluginStore runs a suite of tests against any PluginStore implementation.
func testPluginStore(t *testing.T, factory func() (PluginStore, func())) {
	t.Run("Create", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		p := &plugin.Plugin{
			Name: "test-plugin",
			Repo: "test/plugin",
		}

		err := store.Create(p)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Verify it exists
		exists, err := store.Exists("test-plugin")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Plugin should exist after Create")
		}
	})

	t.Run("Create duplicate", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		p := &plugin.Plugin{
			Name: "test-plugin",
			Repo: "test/plugin",
		}

		_ = store.Create(p)
		err := store.Create(p)

		if !IsAlreadyExists(err) {
			t.Errorf("Expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		original := &plugin.Plugin{
			Name:        "test-plugin",
			Description: "A test plugin",
			Repo:        "test/plugin",
			Branch:      "main",
			Category:    "testing",
			Tags:        []string{"test", "demo"},
		}

		_ = store.Create(original)

		retrieved, err := store.Get("test-plugin")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved.Name != original.Name {
			t.Errorf("Name = %q, want %q", retrieved.Name, original.Name)
		}
		if retrieved.Repo != original.Repo {
			t.Errorf("Repo = %q, want %q", retrieved.Repo, original.Repo)
		}
		if retrieved.Category != original.Category {
			t.Errorf("Category = %q, want %q", retrieved.Category, original.Category)
		}
	})

	t.Run("Get not found", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		_, err := store.Get("nonexistent")

		if !IsNotFound(err) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		original := &plugin.Plugin{
			Name: "test-plugin",
			Repo: "test/plugin",
		}
		_ = store.Create(original)

		updated := &plugin.Plugin{
			Name:        "test-plugin",
			Repo:        "test/plugin",
			Description: "Updated description",
		}
		err := store.Update(updated)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		retrieved, _ := store.Get("test-plugin")
		if retrieved.Description != "Updated description" {
			t.Errorf("Description = %q, want %q", retrieved.Description, "Updated description")
		}
	})

	t.Run("Update not found", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		p := &plugin.Plugin{
			Name: "nonexistent",
			Repo: "test/plugin",
		}

		err := store.Update(p)
		if !IsNotFound(err) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Upsert create", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		p := &plugin.Plugin{
			Name: "test-plugin",
			Repo: "test/plugin",
		}

		err := store.Upsert(p)
		if err != nil {
			t.Fatalf("Upsert failed: %v", err)
		}

		exists, _ := store.Exists("test-plugin")
		if !exists {
			t.Error("Plugin should exist after Upsert")
		}
	})

	t.Run("Upsert update", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		p := &plugin.Plugin{
			Name: "test-plugin",
			Repo: "test/plugin",
		}
		_ = store.Create(p)

		p.Description = "Updated via upsert"
		err := store.Upsert(p)
		if err != nil {
			t.Fatalf("Upsert failed: %v", err)
		}

		retrieved, _ := store.Get("test-plugin")
		if retrieved.Description != "Updated via upsert" {
			t.Errorf("Description = %q, want %q", retrieved.Description, "Updated via upsert")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		p := &plugin.Plugin{
			Name: "test-plugin",
			Repo: "test/plugin",
		}
		_ = store.Create(p)

		err := store.Delete("test-plugin")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		exists, _ := store.Exists("test-plugin")
		if exists {
			t.Error("Plugin should not exist after Delete")
		}
	})

	t.Run("Delete not found", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		err := store.Delete("nonexistent")
		if !IsNotFound(err) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		plugins := []*plugin.Plugin{
			{Name: "plugin-1", Repo: "test/plugin-1"},
			{Name: "plugin-2", Repo: "test/plugin-2"},
			{Name: "plugin-3", Repo: "test/plugin-3"},
		}

		for _, p := range plugins {
			_ = store.Create(p)
		}

		list, err := store.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(list) != 3 {
			t.Errorf("List count = %d, want 3", len(list))
		}
	})

	t.Run("ListByCategory", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		plugins := []*plugin.Plugin{
			{Name: "plugin-1", Repo: "test/plugin-1", Category: "lsp"},
			{Name: "plugin-2", Repo: "test/plugin-2", Category: "ui"},
			{Name: "plugin-3", Repo: "test/plugin-3", Category: "lsp"},
		}

		for _, p := range plugins {
			_ = store.Create(p)
		}

		lspPlugins, err := store.ListByCategory("lsp")
		if err != nil {
			t.Fatalf("ListByCategory failed: %v", err)
		}

		if len(lspPlugins) != 2 {
			t.Errorf("ListByCategory count = %d, want 2", len(lspPlugins))
		}
	})

	t.Run("ListByTag", func(t *testing.T) {
		store, cleanup := factory()
		defer cleanup()

		plugins := []*plugin.Plugin{
			{Name: "plugin-1", Repo: "test/plugin-1", Tags: []string{"finder", "search"}},
			{Name: "plugin-2", Repo: "test/plugin-2", Tags: []string{"git"}},
			{Name: "plugin-3", Repo: "test/plugin-3", Tags: []string{"finder", "fuzzy"}},
		}

		for _, p := range plugins {
			_ = store.Create(p)
		}

		finderPlugins, err := store.ListByTag("finder")
		if err != nil {
			t.Fatalf("ListByTag failed: %v", err)
		}

		if len(finderPlugins) != 2 {
			t.Errorf("ListByTag count = %d, want 2", len(finderPlugins))
		}
	})
}

func TestFileStorePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nvim-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store and add a plugin
	store1, err := NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}

	p := &plugin.Plugin{
		Name:        "persistent-plugin",
		Description: "Should persist to disk",
		Repo:        "test/plugin",
		Category:    "testing",
	}

	if err := store1.Create(p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	store1.Close()

	// Verify file was created
	files, _ := os.ReadDir(tmpDir)
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Create new store instance and verify plugin loads
	store2, err := NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create second FileStore: %v", err)
	}
	defer store2.Close()

	retrieved, err := store2.Get("persistent-plugin")
	if err != nil {
		t.Fatalf("Get from new store failed: %v", err)
	}

	if retrieved.Name != p.Name {
		t.Errorf("Name = %q, want %q", retrieved.Name, p.Name)
	}
	if retrieved.Description != p.Description {
		t.Errorf("Description = %q, want %q", retrieved.Description, p.Description)
	}
}

func TestFileStoreReload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nvim-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}

	// Add a plugin
	p := &plugin.Plugin{
		Name: "test-plugin",
		Repo: "test/plugin",
	}
	_ = store.Create(p)

	// Manually create another YAML file
	newPluginYAML := `apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: external-plugin
spec:
  repo: external/plugin
`
	os.WriteFile(filepath.Join(tmpDir, "external-plugin.yaml"), []byte(newPluginYAML), 0644)

	// Before reload, external plugin should not be in cache
	exists, _ := store.Exists("external-plugin")
	if exists {
		t.Error("External plugin should not exist before reload")
	}

	// Reload
	if err := store.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	// After reload, external plugin should be available
	exists, _ = store.Exists("external-plugin")
	if !exists {
		t.Error("External plugin should exist after reload")
	}
}
