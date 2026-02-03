package library

import (
	"testing"

	"devopsmaestro/pkg/terminalops/prompt"
)

func TestNewPromptLibrary(t *testing.T) {
	lib, err := NewPromptLibrary()
	if err != nil {
		t.Fatalf("NewPromptLibrary failed: %v", err)
	}

	if lib.Count() == 0 {
		t.Error("expected at least one prompt in library")
	}
}

func TestPromptLibrary_Get(t *testing.T) {
	lib, err := NewPromptLibrary()
	if err != nil {
		t.Fatalf("NewPromptLibrary failed: %v", err)
	}

	// Get existing prompt
	p, err := lib.Get("starship-minimal")
	if err != nil {
		t.Fatalf("Get starship-minimal failed: %v", err)
	}
	if p.Name != "starship-minimal" {
		t.Errorf("expected name 'starship-minimal', got %s", p.Name)
	}
	if p.Type != prompt.PromptTypeStarship {
		t.Errorf("expected type starship, got %s", p.Type)
	}

	// Get non-existent prompt
	_, err = lib.Get("non-existent")
	if err == nil {
		t.Error("expected error for non-existent prompt")
	}
}

func TestPromptLibrary_List(t *testing.T) {
	lib, err := NewPromptLibrary()
	if err != nil {
		t.Fatalf("NewPromptLibrary failed: %v", err)
	}

	prompts := lib.List()
	if len(prompts) == 0 {
		t.Error("expected at least one prompt")
	}

	// Check that prompts are sorted by name
	for i := 1; i < len(prompts); i++ {
		if prompts[i].Name < prompts[i-1].Name {
			t.Error("prompts should be sorted by name")
		}
	}
}

func TestPromptLibrary_ListByCategory(t *testing.T) {
	lib, err := NewPromptLibrary()
	if err != nil {
		t.Fatalf("NewPromptLibrary failed: %v", err)
	}

	// Check minimal category
	minimal := lib.ListByCategory("minimal")
	if len(minimal) == 0 {
		t.Error("expected at least one minimal prompt")
	}
	for _, p := range minimal {
		if p.Category != "minimal" {
			t.Errorf("expected category 'minimal', got %s", p.Category)
		}
	}
}

func TestPromptLibrary_Names(t *testing.T) {
	lib, err := NewPromptLibrary()
	if err != nil {
		t.Fatalf("NewPromptLibrary failed: %v", err)
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

func TestPromptLibrary_Categories(t *testing.T) {
	lib, err := NewPromptLibrary()
	if err != nil {
		t.Fatalf("NewPromptLibrary failed: %v", err)
	}

	categories := lib.Categories()
	if len(categories) == 0 {
		t.Error("expected at least one category")
	}
}
