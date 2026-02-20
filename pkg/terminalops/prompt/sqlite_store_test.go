package prompt

import (
	"testing"

	"devopsmaestro/db"
)

func TestSQLitePromptStore_Create(t *testing.T) {
	// Use mock data store for testing
	dataStore := db.NewMockDataStore()
	store := NewSQLitePromptStore(dataStore)

	// Create a test prompt
	prompt := NewPrompt("test-starship", PromptTypeStarship)
	prompt.Description = "Test starship prompt"
	prompt.AddNewline = true
	prompt.Palette = "catppuccin"
	prompt.Category = "minimal"
	prompt.Tags = []string{"fast", "simple"}
	prompt.Colors = map[string]string{"primary": "#89b4fa", "secondary": "#a6e3a1"}
	prompt.Modules = map[string]ModuleConfig{
		"character": {
			Symbol: "❯",
			Style:  "bold blue",
		},
		"directory": {
			Style:    "bold cyan",
			Disabled: false,
		},
	}
	prompt.Character = &CharacterConfig{
		SuccessSymbol: "❯",
		ErrorSymbol:   "❯",
	}

	// Test create
	err := store.Create(prompt)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify it was created
	retrieved, err := store.Get("test-starship")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Name != "test-starship" {
		t.Errorf("Expected name 'test-starship', got %s", retrieved.Name)
	}
	if retrieved.Type != PromptTypeStarship {
		t.Errorf("Expected type %s, got %s", PromptTypeStarship, retrieved.Type)
	}
	if retrieved.Description != "Test starship prompt" {
		t.Errorf("Expected description 'Test starship prompt', got %s", retrieved.Description)
	}
	if !retrieved.AddNewline {
		t.Errorf("Expected AddNewline to be true")
	}
	if retrieved.Category != "minimal" {
		t.Errorf("Expected category 'minimal', got %s", retrieved.Category)
	}
	if len(retrieved.Tags) != 2 || retrieved.Tags[0] != "fast" {
		t.Errorf("Expected tags [fast, simple], got %v", retrieved.Tags)
	}
}

func TestSQLitePromptStore_ListByType(t *testing.T) {
	// Use mock data store for testing
	dataStore := db.NewMockDataStore()
	store := NewSQLitePromptStore(dataStore)

	// Create prompts of different types
	starshipPrompt := NewPrompt("starship-1", PromptTypeStarship)
	starshipPrompt.Category = "starship"

	p10kPrompt := NewPrompt("p10k-1", PromptTypePowerlevel10k)
	p10kPrompt.Category = "powerlevel10k"

	ompPrompt := NewPrompt("omp-1", PromptTypeOhMyPosh)
	ompPrompt.Category = "oh-my-posh"

	// Create them
	if err := store.Create(starshipPrompt); err != nil {
		t.Fatalf("Create starship error = %v", err)
	}
	if err := store.Create(p10kPrompt); err != nil {
		t.Fatalf("Create p10k error = %v", err)
	}
	if err := store.Create(ompPrompt); err != nil {
		t.Fatalf("Create omp error = %v", err)
	}

	// List by type
	starshipPrompts, err := store.ListByType(PromptTypeStarship)
	if err != nil {
		t.Fatalf("ListByType starship error = %v", err)
	}
	if len(starshipPrompts) != 1 {
		t.Errorf("Expected 1 starship prompt, got %d", len(starshipPrompts))
	}
	if starshipPrompts[0].Name != "starship-1" {
		t.Errorf("Expected starship-1, got %s", starshipPrompts[0].Name)
	}

	p10kPrompts, err := store.ListByType(PromptTypePowerlevel10k)
	if err != nil {
		t.Fatalf("ListByType p10k error = %v", err)
	}
	if len(p10kPrompts) != 1 {
		t.Errorf("Expected 1 p10k prompt, got %d", len(p10kPrompts))
	}
}

func TestSQLitePromptStore_Update(t *testing.T) {
	// Use mock data store for testing
	dataStore := db.NewMockDataStore()
	store := NewSQLitePromptStore(dataStore)

	// Create a test prompt
	prompt := NewPrompt("test-update", PromptTypeStarship)
	prompt.Description = "Original description"
	prompt.Category = "original"

	err := store.Create(prompt)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the prompt
	prompt.Description = "Updated description"
	prompt.Category = "updated"
	prompt.Tags = []string{"updated"}

	err = store.Update(prompt)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify it was updated
	retrieved, err := store.Get("test-update")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Description != "Updated description" {
		t.Errorf("Expected updated description, got %s", retrieved.Description)
	}
	if retrieved.Category != "updated" {
		t.Errorf("Expected updated category, got %s", retrieved.Category)
	}
	if len(retrieved.Tags) != 1 || retrieved.Tags[0] != "updated" {
		t.Errorf("Expected updated tags, got %v", retrieved.Tags)
	}
}

func TestSQLitePromptStore_Delete(t *testing.T) {
	// Use mock data store for testing
	dataStore := db.NewMockDataStore()
	store := NewSQLitePromptStore(dataStore)

	// Create a test prompt
	prompt := NewPrompt("test-delete", PromptTypeStarship)
	err := store.Create(prompt)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify it exists
	exists, err := store.Exists("test-delete")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Expected prompt to exist")
	}

	// Delete it
	err = store.Delete("test-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it was deleted
	exists, err = store.Exists("test-delete")
	if err != nil {
		t.Fatalf("Exists() after delete error = %v", err)
	}
	if exists {
		t.Error("Expected prompt to be deleted")
	}

	// Verify Get returns error
	_, err = store.Get("test-delete")
	if err == nil {
		t.Error("Expected Get to return error after delete")
	}
	if !IsNotFound(err) {
		t.Error("Expected NotFound error")
	}
}
