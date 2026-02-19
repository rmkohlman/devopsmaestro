package library

import (
	"testing"
)

func TestNewWeztermLibrary(t *testing.T) {
	lib, err := NewWeztermLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	if lib.Count() == 0 {
		t.Error("Library should contain at least one preset")
	}

	names := lib.Names()
	if len(names) == 0 {
		t.Error("Library should have preset names")
	}

	// Check that we have expected presets
	expectedPresets := []string{"minimal", "tmux-style", "default"}
	for _, preset := range expectedPresets {
		if _, err := lib.Get(preset); err != nil {
			t.Errorf("Expected preset %s not found: %v", preset, err)
		}
	}
}

func TestWeztermLibraryGet(t *testing.T) {
	lib, err := NewWeztermLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	// Test getting existing preset
	config, err := lib.Get("minimal")
	if err != nil {
		t.Fatalf("Failed to get minimal preset: %v", err)
	}

	if config.Name != "minimal" {
		t.Errorf("Expected config name 'minimal', got %s", config.Name)
	}

	// Test getting non-existent preset
	_, err = lib.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent preset")
	}
}

func TestWeztermLibraryList(t *testing.T) {
	lib, err := NewWeztermLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	configs := lib.List()
	if len(configs) == 0 {
		t.Error("List should return at least one config")
	}

	// Check that results are sorted
	for i := 1; i < len(configs); i++ {
		if configs[i-1].Name >= configs[i].Name {
			t.Errorf("Results should be sorted by name: %s >= %s", configs[i-1].Name, configs[i].Name)
		}
	}
}

func TestWeztermLibraryCategories(t *testing.T) {
	lib, err := NewWeztermLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	categories := lib.Categories()
	if len(categories) == 0 {
		t.Error("Should have at least one category")
	}

	// Check that categories are sorted
	for i := 1; i < len(categories); i++ {
		if categories[i-1] >= categories[i] {
			t.Errorf("Categories should be sorted: %s >= %s", categories[i-1], categories[i])
		}
	}

	// Test getting configs by category
	for _, category := range categories {
		configs := lib.ListByCategory(category)
		if len(configs) == 0 {
			t.Errorf("Category %s should have at least one config", category)
		}

		// All configs should have the expected category
		for _, config := range configs {
			if config.Category != category {
				t.Errorf("Config %s should have category %s, got %s", config.Name, category, config.Category)
			}
		}
	}
}
