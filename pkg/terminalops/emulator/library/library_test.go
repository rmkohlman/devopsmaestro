package library

import (
	"testing"
)

func TestNewEmulatorLibrary(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	if lib == nil {
		t.Fatal("Library is nil")
	}

	// Should have at least the embedded emulators
	count := lib.Count()
	if count == 0 {
		t.Fatal("Library is empty")
	}

	t.Logf("Loaded %d emulators", count)
}

func TestLibrary_List(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	names := lib.List()
	if len(names) == 0 {
		t.Fatal("No emulators found")
	}

	// Names should be sorted
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("Names not sorted: %s > %s", names[i-1], names[i])
		}
	}

	t.Logf("Emulators: %v", names)
}

func TestLibrary_Get(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	// Test getting existing emulator
	emulator, err := lib.Get("minimal")
	if err != nil {
		t.Fatalf("Failed to get minimal emulator: %v", err)
	}

	if emulator.Name != "minimal" {
		t.Errorf("Expected name 'minimal', got %s", emulator.Name)
	}

	if emulator.Type != "wezterm" {
		t.Errorf("Expected type wezterm, got %s", emulator.Type)
	}

	// Test getting non-existent emulator
	_, err = lib.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent emulator")
	}
}

func TestLibrary_ListByType(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	// Test WezTerm emulators
	weztermEmulators := lib.ListByType("wezterm")
	if len(weztermEmulators) == 0 {
		t.Error("No WezTerm emulators found")
	}

	for _, e := range weztermEmulators {
		if e.Type != "wezterm" {
			t.Errorf("Expected wezterm type, got %s", e.Type)
		}
	}

	// Test Alacritty emulators
	alacrittyEmulators := lib.ListByType("alacritty")
	for _, e := range alacrittyEmulators {
		if e.Type != "alacritty" {
			t.Errorf("Expected alacritty type, got %s", e.Type)
		}
	}

	t.Logf("WezTerm emulators: %d, Alacritty emulators: %d",
		len(weztermEmulators), len(alacrittyEmulators))
}

func TestLibrary_Categories(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	categories := lib.Categories()
	if len(categories) == 0 {
		t.Error("No categories found")
	}

	// Categories should be sorted
	for i := 1; i < len(categories); i++ {
		if categories[i-1] > categories[i] {
			t.Errorf("Categories not sorted: %s > %s", categories[i-1], categories[i])
		}
	}

	t.Logf("Categories: %v", categories)
}

func TestLibrary_Has(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	// Test existing emulator
	if !lib.Has("minimal") {
		t.Error("Expected minimal emulator to exist")
	}

	// Test non-existent emulator
	if lib.Has("nonexistent") {
		t.Error("Expected nonexistent emulator to not exist")
	}
}

func TestLibrary_All(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	emulators := lib.All()
	if len(emulators) == 0 {
		t.Error("No emulators found")
	}

	// Should be sorted by name
	for i := 1; i < len(emulators); i++ {
		if emulators[i-1].Name > emulators[i].Name {
			t.Errorf("Emulators not sorted: %s > %s",
				emulators[i-1].Name, emulators[i].Name)
		}
	}

	// Count should match
	if len(emulators) != lib.Count() {
		t.Errorf("Expected %d emulators, got %d", lib.Count(), len(emulators))
	}
}

func TestLibrary_ListByCategory(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	// Test starter category
	starters := lib.ListByCategory("starter")
	for _, e := range starters {
		if e.Category != "starter" {
			t.Errorf("Expected category 'starter', got %s", e.Category)
		}
	}

	// Test developer category
	developers := lib.ListByCategory("developer")
	for _, e := range developers {
		if e.Category != "developer" {
			t.Errorf("Expected category 'developer', got %s", e.Category)
		}
	}

	t.Logf("Starter emulators: %d, Developer emulators: %d",
		len(starters), len(developers))
}

func TestDefaultLibrary(t *testing.T) {
	lib := Default()
	if lib == nil {
		t.Fatal("Default library is nil")
	}

	count := lib.Count()
	if count == 0 {
		t.Fatal("Default library is empty")
	}

	t.Logf("Default library has %d emulators", count)
}

func TestExpectedEmulators(t *testing.T) {
	lib, err := NewEmulatorLibrary()
	if err != nil {
		t.Fatalf("Failed to create library: %v", err)
	}

	expectedEmulators := []string{
		"rmkohlman",
		"minimal",
		"developer",
		"alacritty-minimal",
		"kitty-poweruser",
		"iterm2-macos",
	}

	for _, name := range expectedEmulators {
		if !lib.Has(name) {
			t.Errorf("Expected emulator %s not found", name)
		}

		emulator, err := lib.Get(name)
		if err != nil {
			t.Errorf("Failed to get emulator %s: %v", name, err)
		}

		if emulator.Name != name {
			t.Errorf("Expected name %s, got %s", name, emulator.Name)
		}
	}
}
