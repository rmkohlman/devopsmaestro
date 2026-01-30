package store

import (
	"os"
	"testing"

	"devopsmaestro/pkg/nvimops/library"
	"devopsmaestro/pkg/nvimops/plugin"
)

// TestInterfaceCompliance verifies all store implementations satisfy PluginStore.
func TestInterfaceCompliance(t *testing.T) {
	t.Run("MemoryStore implements PluginStore", func(t *testing.T) {
		var _ PluginStore = (*MemoryStore)(nil)
		var _ PluginStore = NewMemoryStore()
	})

	t.Run("FileStore implements PluginStore", func(t *testing.T) {
		var _ PluginStore = (*FileStore)(nil)

		tmpDir, err := os.MkdirTemp("", "test-filestore-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		store, err := NewFileStore(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create FileStore: %v", err)
		}
		var _ PluginStore = store
	})

	t.Run("ReadOnlyStore implements PluginStore", func(t *testing.T) {
		var _ PluginStore = (*ReadOnlyStore)(nil)

		lib, err := library.NewLibrary()
		if err != nil {
			t.Fatalf("Failed to create library: %v", err)
		}
		store := NewReadOnlyStore(lib)
		var _ PluginStore = store
	})
}

// TestLibraryImplementsReadOnlySource verifies Library implements ReadOnlySource.
func TestLibraryImplementsReadOnlySource(t *testing.T) {
	lib, err := library.NewLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	// Verify Library implements ReadOnlySource
	var _ ReadOnlySource = lib
}

// TestReadOnlyStore tests the ReadOnlyStore wrapper with a Library.
func TestReadOnlyStore(t *testing.T) {
	lib, err := library.NewLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	store := NewReadOnlyStore(lib)
	defer store.Close()

	t.Run("Get existing plugin", func(t *testing.T) {
		p, err := store.Get("telescope")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if p == nil {
			t.Fatal("Get returned nil plugin")
		}
		if p.Name != "telescope" {
			t.Errorf("Name = %q, want telescope", p.Name)
		}
		if p.Repo == "" {
			t.Error("Repo should not be empty")
		}
	})

	t.Run("Get non-existent plugin", func(t *testing.T) {
		_, err := store.Get("nonexistent-plugin")
		if !IsNotFound(err) {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("List all plugins", func(t *testing.T) {
		plugins, err := store.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(plugins) == 0 {
			t.Error("List should return plugins")
		}
		t.Logf("Library has %d plugins accessible through store interface", len(plugins))
	})

	t.Run("ListByCategory", func(t *testing.T) {
		plugins, err := store.ListByCategory("lsp")
		if err != nil {
			t.Fatalf("ListByCategory failed: %v", err)
		}
		t.Logf("Found %d LSP plugins through store interface", len(plugins))
	})

	t.Run("ListByTag", func(t *testing.T) {
		plugins, err := store.ListByTag("finder")
		if err != nil {
			t.Fatalf("ListByTag failed: %v", err)
		}
		t.Logf("Found %d finder plugins through store interface", len(plugins))
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := store.Exists("telescope")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("telescope should exist")
		}

		exists, err = store.Exists("nonexistent")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("nonexistent should not exist")
		}
	})

	// Test write operations fail
	t.Run("Create fails (read-only)", func(t *testing.T) {
		err := store.Create(&plugin.Plugin{Name: "test", Repo: "test/test"})
		if !IsReadOnly(err) {
			t.Errorf("Expected ErrReadOnly, got %v", err)
		}
	})

	t.Run("Update fails (read-only)", func(t *testing.T) {
		err := store.Update(&plugin.Plugin{Name: "telescope", Repo: "test/test"})
		if !IsReadOnly(err) {
			t.Errorf("Expected ErrReadOnly, got %v", err)
		}
	})

	t.Run("Upsert fails (read-only)", func(t *testing.T) {
		err := store.Upsert(&plugin.Plugin{Name: "test", Repo: "test/test"})
		if !IsReadOnly(err) {
			t.Errorf("Expected ErrReadOnly, got %v", err)
		}
	})

	t.Run("Delete fails (read-only)", func(t *testing.T) {
		err := store.Delete("telescope")
		if !IsReadOnly(err) {
			t.Errorf("Expected ErrReadOnly, got %v", err)
		}
	})
}

// TestStoreSwappability verifies that code written against the interface works
// with any implementation.
func TestStoreSwappability(t *testing.T) {
	// Create test fixtures
	testPlugins := []*plugin.Plugin{
		{Name: "plugin-a", Repo: "test/plugin-a", Category: "testing", Tags: []string{"test"}},
		{Name: "plugin-b", Repo: "test/plugin-b", Category: "testing", Tags: []string{"test", "demo"}},
		{Name: "plugin-c", Repo: "test/plugin-c", Category: "other", Tags: []string{"demo"}},
	}

	// Test function that only uses the PluginStore interface
	testReadOperations := func(t *testing.T, s PluginStore) {
		// List
		plugins, err := s.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(plugins) == 0 {
			t.Error("Store should have plugins")
		}

		// Get first plugin
		if len(plugins) > 0 {
			name := plugins[0].Name
			p, err := s.Get(name)
			if err != nil {
				t.Errorf("Get(%q) failed: %v", name, err)
			}
			if p == nil || p.Name != name {
				t.Errorf("Get returned wrong plugin")
			}
		}

		// Exists
		if len(plugins) > 0 {
			exists, err := s.Exists(plugins[0].Name)
			if err != nil {
				t.Errorf("Exists failed: %v", err)
			}
			if !exists {
				t.Error("Exists should return true for known plugin")
			}
		}
	}

	// Test with MemoryStore
	t.Run("MemoryStore", func(t *testing.T) {
		store := NewMemoryStore()
		for _, p := range testPlugins {
			_ = store.Create(p)
		}
		testReadOperations(t, store)
	})

	// Test with FileStore
	t.Run("FileStore", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-filestore-swap-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		store, err := NewFileStore(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create FileStore: %v", err)
		}
		for _, p := range testPlugins {
			_ = store.Create(p)
		}
		testReadOperations(t, store)
	})

	// Test with ReadOnlyStore (Library)
	t.Run("ReadOnlyStore (Library)", func(t *testing.T) {
		lib, err := library.NewLibrary()
		if err != nil {
			t.Fatalf("Failed to create library: %v", err)
		}
		store := NewReadOnlyStore(lib)
		testReadOperations(t, store)
	})
}

// TestErrorTypes verifies error type checking works correctly.
func TestErrorTypes(t *testing.T) {
	t.Run("ErrNotFound", func(t *testing.T) {
		err := &ErrNotFound{Name: "test"}
		if !IsNotFound(err) {
			t.Error("IsNotFound should return true for ErrNotFound")
		}
		if err.Error() != "plugin not found: test" {
			t.Errorf("Error message = %q", err.Error())
		}
	})

	t.Run("ErrAlreadyExists", func(t *testing.T) {
		err := &ErrAlreadyExists{Name: "test"}
		if !IsAlreadyExists(err) {
			t.Error("IsAlreadyExists should return true for ErrAlreadyExists")
		}
		if err.Error() != "plugin already exists: test" {
			t.Errorf("Error message = %q", err.Error())
		}
	})

	t.Run("ErrReadOnly", func(t *testing.T) {
		err := &ErrReadOnly{Operation: "create"}
		if !IsReadOnly(err) {
			t.Error("IsReadOnly should return true for ErrReadOnly")
		}
		if err.Error() != "operation not permitted on read-only store: create" {
			t.Errorf("Error message = %q", err.Error())
		}
	})

	t.Run("Other errors", func(t *testing.T) {
		err := os.ErrNotExist
		if IsNotFound(err) {
			t.Error("IsNotFound should return false for os.ErrNotExist")
		}
		if IsAlreadyExists(err) {
			t.Error("IsAlreadyExists should return false for os.ErrNotExist")
		}
		if IsReadOnly(err) {
			t.Error("IsReadOnly should return false for os.ErrNotExist")
		}
	})
}
