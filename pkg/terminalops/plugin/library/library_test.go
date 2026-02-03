package library

import (
	"testing"

	"devopsmaestro/pkg/terminalops/plugin"
)

func TestNewPluginLibrary(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	if lib.Count() == 0 {
		t.Error("expected at least one plugin in library")
	}
}

func TestPluginLibrary_Get(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	// Get existing plugin
	p, err := lib.Get("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("Get zsh-autosuggestions failed: %v", err)
	}
	if p.Name != "zsh-autosuggestions" {
		t.Errorf("expected name 'zsh-autosuggestions', got %s", p.Name)
	}
	if p.Repo != "zsh-users/zsh-autosuggestions" {
		t.Errorf("expected repo 'zsh-users/zsh-autosuggestions', got %s", p.Repo)
	}

	// Get non-existent plugin
	_, err = lib.Get("non-existent")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}
}

func TestPluginLibrary_List(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	plugins := lib.List()
	if len(plugins) == 0 {
		t.Error("expected at least one plugin")
	}
}

func TestPluginLibrary_ListByCategory(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	// Check productivity category
	productivity := lib.ListByCategory("productivity")
	if len(productivity) == 0 {
		t.Error("expected at least one productivity plugin")
	}
	for _, p := range productivity {
		if p.Category != "productivity" {
			t.Errorf("expected category 'productivity', got %s", p.Category)
		}
	}
}

func TestPluginLibrary_ListByManager(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	// All library plugins use manual manager
	manual := lib.ListByManager(plugin.PluginManagerManual)
	if len(manual) == 0 {
		t.Error("expected at least one manual plugin")
	}
}

func TestPluginLibrary_Names(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	names := lib.Names()
	if len(names) == 0 {
		t.Error("expected at least one name")
	}

	// Check that names are sorted
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Error("names should be sorted")
		}
	}
}

func TestPluginLibrary_EssentialPlugins(t *testing.T) {
	lib, err := NewPluginLibrary()
	if err != nil {
		t.Fatalf("NewPluginLibrary failed: %v", err)
	}

	essentials := lib.EssentialPlugins()
	if len(essentials) == 0 {
		t.Error("expected at least one essential plugin")
	}

	// Check that zsh-autosuggestions is essential
	found := false
	for _, p := range essentials {
		if p.Name == "zsh-autosuggestions" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected zsh-autosuggestions to be essential")
	}
}
