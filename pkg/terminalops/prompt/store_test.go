package prompt

import (
	"testing"
)

// TestMemoryStore_ImplementsPromptStore verifies interface compliance.
func TestMemoryStore_ImplementsPromptStore(t *testing.T) {
	var _ PromptStore = (*MemoryStore)(nil)
}

func TestMemoryStore_CRUD(t *testing.T) {
	store := NewMemoryStore()

	// Create
	p := &Prompt{
		Name:    "test-prompt",
		Type:    PromptTypeStarship,
		Enabled: true,
	}

	err := store.Create(p)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Create duplicate should fail
	err = store.Create(p)
	if err == nil {
		t.Error("expected error when creating duplicate")
	}
	if !IsAlreadyExists(err) {
		t.Error("expected ErrAlreadyExists error")
	}

	// Get
	retrieved, err := store.Get("test-prompt")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != "test-prompt" {
		t.Errorf("expected name 'test-prompt', got %s", retrieved.Name)
	}

	// Get non-existent
	_, err = store.Get("non-existent")
	if err == nil {
		t.Error("expected error when getting non-existent prompt")
	}
	if !IsNotFound(err) {
		t.Error("expected ErrNotFound error")
	}

	// Update
	p.Description = "Updated description"
	err = store.Update(p)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ = store.Get("test-prompt")
	if retrieved.Description != "Updated description" {
		t.Error("expected description to be updated")
	}

	// Update non-existent should fail
	nonExistent := &Prompt{Name: "non-existent", Type: PromptTypeStarship}
	err = store.Update(nonExistent)
	if err == nil {
		t.Error("expected error when updating non-existent prompt")
	}

	// Delete
	err = store.Delete("test-prompt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = store.Get("test-prompt")
	if err == nil {
		t.Error("expected error after delete")
	}

	// Delete non-existent should fail
	err = store.Delete("non-existent")
	if err == nil {
		t.Error("expected error when deleting non-existent prompt")
	}
}

func TestMemoryStore_Upsert(t *testing.T) {
	store := NewMemoryStore()

	p := &Prompt{
		Name:        "test-prompt",
		Type:        PromptTypeStarship,
		Description: "Initial",
	}

	// Upsert new prompt
	err := store.Upsert(p)
	if err != nil {
		t.Fatalf("Upsert (create) failed: %v", err)
	}

	retrieved, _ := store.Get("test-prompt")
	if retrieved.Description != "Initial" {
		t.Error("expected initial description")
	}

	// Upsert existing prompt
	p.Description = "Updated"
	err = store.Upsert(p)
	if err != nil {
		t.Fatalf("Upsert (update) failed: %v", err)
	}

	retrieved, _ = store.Get("test-prompt")
	if retrieved.Description != "Updated" {
		t.Error("expected updated description")
	}
}

func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore()

	// Empty list
	prompts, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts, got %d", len(prompts))
	}

	// Add some prompts
	_ = store.Create(&Prompt{Name: "p1", Type: PromptTypeStarship, Category: "minimal"})
	_ = store.Create(&Prompt{Name: "p2", Type: PromptTypePowerlevel10k, Category: "feature-rich"})
	_ = store.Create(&Prompt{Name: "p3", Type: PromptTypeStarship, Category: "minimal"})

	// List all
	prompts, err = store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(prompts) != 3 {
		t.Errorf("expected 3 prompts, got %d", len(prompts))
	}
}

func TestMemoryStore_ListByType(t *testing.T) {
	store := NewMemoryStore()

	_ = store.Create(&Prompt{Name: "p1", Type: PromptTypeStarship})
	_ = store.Create(&Prompt{Name: "p2", Type: PromptTypePowerlevel10k})
	_ = store.Create(&Prompt{Name: "p3", Type: PromptTypeStarship})

	// List Starship prompts
	prompts, err := store.ListByType(PromptTypeStarship)
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}
	if len(prompts) != 2 {
		t.Errorf("expected 2 Starship prompts, got %d", len(prompts))
	}

	// List Powerlevel10k prompts
	prompts, err = store.ListByType(PromptTypePowerlevel10k)
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 Powerlevel10k prompt, got %d", len(prompts))
	}
}

func TestMemoryStore_ListByCategory(t *testing.T) {
	store := NewMemoryStore()

	_ = store.Create(&Prompt{Name: "p1", Type: PromptTypeStarship, Category: "minimal"})
	_ = store.Create(&Prompt{Name: "p2", Type: PromptTypeStarship, Category: "feature-rich"})
	_ = store.Create(&Prompt{Name: "p3", Type: PromptTypeStarship, Category: "minimal"})

	// List minimal prompts
	prompts, err := store.ListByCategory("minimal")
	if err != nil {
		t.Fatalf("ListByCategory failed: %v", err)
	}
	if len(prompts) != 2 {
		t.Errorf("expected 2 minimal prompts, got %d", len(prompts))
	}
}

func TestMemoryStore_Exists(t *testing.T) {
	store := NewMemoryStore()

	_ = store.Create(&Prompt{Name: "test-prompt", Type: PromptTypeStarship})

	exists, err := store.Exists("test-prompt")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected prompt to exist")
	}

	exists, err = store.Exists("non-existent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected prompt to not exist")
	}
}

func TestMemoryStore_Close(t *testing.T) {
	store := NewMemoryStore()
	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
