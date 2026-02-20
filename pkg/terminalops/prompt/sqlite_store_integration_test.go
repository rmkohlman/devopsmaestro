package prompt

import (
	"testing"

	"devopsmaestro/db"
)

// =============================================================================
// Integration Tests for SQLitePromptStore
//
// These tests verify that the SQLitePromptStore adapter works correctly
// with a real SQLite database, testing the full flow from domain layer
// to database and back.
// =============================================================================

// createIntegrationPromptStore creates a real SQLitePromptStore with SQLite backend for testing
func createIntegrationPromptStore(t *testing.T) *SQLitePromptStore {
	t.Helper()

	// Create an in-memory SQLite database
	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	// Create the terminal_prompts table
	createTableSQL := `CREATE TABLE IF NOT EXISTS terminal_prompts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		type TEXT NOT NULL,
		add_newline BOOLEAN DEFAULT TRUE,
		palette TEXT,
		format TEXT,
		modules TEXT,
		character TEXT,
		palette_ref TEXT,
		colors TEXT,
		raw_config TEXT,
		category TEXT,
		tags TEXT,
		enabled BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := driver.Execute(createTableSQL); err != nil {
		driver.Close()
		t.Fatalf("Failed to create terminal_prompts table: %v", err)
	}

	// Create DataStore and adapter
	dataStore := db.NewSQLDataStore(driver, nil)
	return NewSQLitePromptStore(dataStore)
}

func TestSQLitePromptStore_Integration_Create(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Create a complex prompt
	prompt := NewPrompt("integration-starship", PromptTypeStarship)
	prompt.Description = "Integration test Starship prompt"
	prompt.AddNewline = true
	prompt.Palette = "catppuccin"
	prompt.Format = "$all$character"
	prompt.PaletteRef = "catppuccin-mocha"
	prompt.Category = "powerline"
	prompt.Tags = []string{"git", "docker", "performance"}
	prompt.Colors = map[string]string{
		"primary":   "#89b4fa",
		"secondary": "#a6e3a1",
		"accent":    "#f38ba8",
	}
	prompt.Modules = map[string]ModuleConfig{
		"character": {
			Symbol: "❯",
			Style:  "bold green",
		},
		"directory": {
			Style:    "bold cyan",
			Disabled: false,
		},
		"git_branch": {
			Symbol: "",
			Style:  "purple bold",
		},
	}
	prompt.Character = &CharacterConfig{
		SuccessSymbol: "❯",
		ErrorSymbol:   "✗",
		ViCmdSymbol:   "❮",
	}
	prompt.RawConfig = `format = "$all$character"
[character]
success_symbol = "[❯](bold green)"
error_symbol = "[❯](bold red)"`
	prompt.Enabled = true

	// Test create
	err := store.Create(prompt)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify it was created by retrieving it
	retrieved, err := store.Get("integration-starship")
	if err != nil {
		t.Fatalf("Get() after Create() error = %v", err)
	}

	// Verify all complex data
	if retrieved.Name != "integration-starship" {
		t.Errorf("Name = %q, want %q", retrieved.Name, "integration-starship")
	}
	if retrieved.Type != PromptTypeStarship {
		t.Errorf("Type = %q, want %q", retrieved.Type, PromptTypeStarship)
	}
	if retrieved.Description != "Integration test Starship prompt" {
		t.Errorf("Description = %q, want %q", retrieved.Description, "Integration test Starship prompt")
	}
	if !retrieved.AddNewline {
		t.Errorf("AddNewline = %v, want %v", retrieved.AddNewline, true)
	}
	if retrieved.Palette != "catppuccin" {
		t.Errorf("Palette = %q, want %q", retrieved.Palette, "catppuccin")
	}
	if retrieved.Format != "$all$character" {
		t.Errorf("Format = %q, want %q", retrieved.Format, "$all$character")
	}
	if retrieved.Category != "powerline" {
		t.Errorf("Category = %q, want %q", retrieved.Category, "powerline")
	}

	// Verify tags (slice comparison)
	if len(retrieved.Tags) != 3 {
		t.Errorf("Tags length = %d, want 3", len(retrieved.Tags))
	} else {
		expectedTags := []string{"git", "docker", "performance"}
		for i, tag := range expectedTags {
			if retrieved.Tags[i] != tag {
				t.Errorf("Tags[%d] = %q, want %q", i, retrieved.Tags[i], tag)
			}
		}
	}

	// Verify colors (map comparison)
	if len(retrieved.Colors) != 3 {
		t.Errorf("Colors length = %d, want 3", len(retrieved.Colors))
	}
	expectedColors := map[string]string{
		"primary":   "#89b4fa",
		"secondary": "#a6e3a1",
		"accent":    "#f38ba8",
	}
	for key, expectedValue := range expectedColors {
		if actualValue, exists := retrieved.Colors[key]; !exists {
			t.Errorf("Colors[%q] missing", key)
		} else if actualValue != expectedValue {
			t.Errorf("Colors[%q] = %q, want %q", key, actualValue, expectedValue)
		}
	}

	// Verify modules (complex nested structure)
	if len(retrieved.Modules) != 3 {
		t.Errorf("Modules length = %d, want 3", len(retrieved.Modules))
	}
	if charModule, exists := retrieved.Modules["character"]; exists {
		if charModule.Symbol != "❯" {
			t.Errorf("Modules[character].Symbol = %q, want %q", charModule.Symbol, "❯")
		}
		if charModule.Style != "bold green" {
			t.Errorf("Modules[character].Style = %q, want %q", charModule.Style, "bold green")
		}
	} else {
		t.Error("Modules[character] missing")
	}

	// Verify character config
	if retrieved.Character == nil {
		t.Error("Character is nil")
	} else {
		if retrieved.Character.SuccessSymbol != "❯" {
			t.Errorf("Character.SuccessSymbol = %q, want %q", retrieved.Character.SuccessSymbol, "❯")
		}
		if retrieved.Character.ErrorSymbol != "✗" {
			t.Errorf("Character.ErrorSymbol = %q, want %q", retrieved.Character.ErrorSymbol, "✗")
		}
		if retrieved.Character.ViCmdSymbol != "❮" {
			t.Errorf("Character.ViCmdSymbol = %q, want %q", retrieved.Character.ViCmdSymbol, "❮")
		}
	}

	// Verify raw config
	expectedRawConfig := `format = "$all$character"
[character]
success_symbol = "[❯](bold green)"
error_symbol = "[❯](bold red)"`
	if retrieved.RawConfig != expectedRawConfig {
		t.Errorf("RawConfig = %q, want %q", retrieved.RawConfig, expectedRawConfig)
	}

	if !retrieved.Enabled {
		t.Errorf("Enabled = %v, want %v", retrieved.Enabled, true)
	}
}

func TestSQLitePromptStore_Integration_Update(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Create initial prompt
	prompt := NewPrompt("update-test", PromptTypeOhMyPosh)
	prompt.Description = "Original description"
	prompt.AddNewline = true
	prompt.Category = "minimal"
	prompt.Tags = []string{"simple"}
	prompt.Colors = map[string]string{"primary": "#ff0000"}

	if err := store.Create(prompt); err != nil {
		t.Fatalf("Setup Create() error = %v", err)
	}

	// Update the prompt
	prompt.Description = "Updated description"
	prompt.AddNewline = false
	prompt.Category = "powerline"
	prompt.Tags = []string{"complex", "advanced"}
	prompt.Colors = map[string]string{
		"primary":   "#00ff00",
		"secondary": "#0000ff",
	}
	prompt.Modules = map[string]ModuleConfig{
		"directory": {
			Symbol:   "📁",
			Style:    "bold blue",
			Disabled: false,
		},
	}
	prompt.Enabled = false

	err := store.Update(prompt)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Retrieve and verify all updates
	updated, err := store.Get("update-test")
	if err != nil {
		t.Fatalf("Get() after Update() error = %v", err)
	}

	if updated.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", updated.Description, "Updated description")
	}
	if updated.AddNewline {
		t.Errorf("AddNewline = %v, want %v", updated.AddNewline, false)
	}
	if updated.Category != "powerline" {
		t.Errorf("Category = %q, want %q", updated.Category, "powerline")
	}
	if len(updated.Tags) != 2 || updated.Tags[0] != "complex" || updated.Tags[1] != "advanced" {
		t.Errorf("Tags = %v, want [complex advanced]", updated.Tags)
	}
	if len(updated.Colors) != 2 {
		t.Errorf("Colors length = %d, want 2", len(updated.Colors))
	}
	if updated.Colors["primary"] != "#00ff00" {
		t.Errorf("Colors[primary] = %q, want %q", updated.Colors["primary"], "#00ff00")
	}
	if updated.Colors["secondary"] != "#0000ff" {
		t.Errorf("Colors[secondary] = %q, want %q", updated.Colors["secondary"], "#0000ff")
	}
	if len(updated.Modules) != 1 {
		t.Errorf("Modules length = %d, want 1", len(updated.Modules))
	}
	if updated.Enabled {
		t.Errorf("Enabled = %v, want %v", updated.Enabled, false)
	}
}

func TestSQLitePromptStore_Integration_Upsert(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Test upsert on non-existing prompt (should create)
	prompt := NewPrompt("upsert-test", PromptTypePowerlevel10k)
	prompt.Description = "Upsert test prompt"
	prompt.Category = "test"

	err := store.Upsert(prompt)
	if err != nil {
		t.Fatalf("Upsert() (create) error = %v", err)
	}

	// Verify it was created
	retrieved, err := store.Get("upsert-test")
	if err != nil {
		t.Fatalf("Get() after Upsert() error = %v", err)
	}
	if retrieved.Description != "Upsert test prompt" {
		t.Errorf("Description = %q, want %q", retrieved.Description, "Upsert test prompt")
	}

	// Test upsert on existing prompt (should update)
	prompt.Description = "Updated via upsert"
	prompt.Category = "updated"
	prompt.Colors = map[string]string{"new": "#123456"}

	err = store.Upsert(prompt)
	if err != nil {
		t.Fatalf("Upsert() (update) error = %v", err)
	}

	// Verify it was updated
	updated, err := store.Get("upsert-test")
	if err != nil {
		t.Fatalf("Get() after Upsert() update error = %v", err)
	}
	if updated.Description != "Updated via upsert" {
		t.Errorf("Description = %q, want %q", updated.Description, "Updated via upsert")
	}
	if updated.Category != "updated" {
		t.Errorf("Category = %q, want %q", updated.Category, "updated")
	}
	if updated.Colors["new"] != "#123456" {
		t.Errorf("Colors[new] = %q, want %q", updated.Colors["new"], "#123456")
	}
}

func TestSQLitePromptStore_Integration_Delete(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Create a prompt to delete
	prompt := NewPrompt("delete-me", PromptTypeStarship)
	if err := store.Create(prompt); err != nil {
		t.Fatalf("Setup Create() error = %v", err)
	}

	// Verify it exists
	_, err := store.Get("delete-me")
	if err != nil {
		t.Fatalf("Setup verification error = %v", err)
	}

	// Delete it
	err = store.Delete("delete-me")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = store.Get("delete-me")
	if err == nil {
		t.Error("Delete() did not remove the prompt")
	}
}

func TestSQLitePromptStore_Integration_List(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Create multiple prompts
	prompts := []*Prompt{
		{Name: "alpha-starship", Type: PromptTypeStarship, Category: "minimal"},
		{Name: "beta-posh", Type: PromptTypeOhMyPosh, Category: "powerline"},
		{Name: "gamma-p10k", Type: PromptTypePowerlevel10k, Category: "advanced"},
	}

	for _, p := range prompts {
		if err := store.Create(p); err != nil {
			t.Fatalf("Setup Create() error = %v", err)
		}
	}

	// List all prompts
	all, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("List() returned %d prompts, want 3", len(all))
	}

	// Should be ordered by name
	expectedNames := []string{"alpha-starship", "beta-posh", "gamma-p10k"}
	for i, prompt := range all {
		if prompt.Name != expectedNames[i] {
			t.Errorf("List() prompt[%d].Name = %q, want %q", i, prompt.Name, expectedNames[i])
		}
	}
}

func TestSQLitePromptStore_Integration_ListByType(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Create prompts of different types
	prompts := []*Prompt{
		{Name: "starship-1", Type: PromptTypeStarship},
		{Name: "starship-2", Type: PromptTypeStarship},
		{Name: "posh-1", Type: PromptTypeOhMyPosh},
		{Name: "p10k-1", Type: PromptTypePowerlevel10k},
	}

	for _, p := range prompts {
		if err := store.Create(p); err != nil {
			t.Fatalf("Setup Create() error = %v", err)
		}
	}

	// List starship prompts
	starshipPrompts, err := store.ListByType(PromptTypeStarship)
	if err != nil {
		t.Fatalf("ListByType() error = %v", err)
	}

	if len(starshipPrompts) != 2 {
		t.Fatalf("ListByType(starship) returned %d prompts, want 2", len(starshipPrompts))
	}

	// Should be ordered by name
	expectedNames := []string{"starship-1", "starship-2"}
	for i, prompt := range starshipPrompts {
		if prompt.Name != expectedNames[i] {
			t.Errorf("ListByType() prompt[%d].Name = %q, want %q", i, prompt.Name, expectedNames[i])
		}
		if prompt.Type != PromptTypeStarship {
			t.Errorf("ListByType() prompt[%d].Type = %q, want %q", i, prompt.Type, PromptTypeStarship)
		}
	}

	// List oh-my-posh prompts
	poshPrompts, err := store.ListByType(PromptTypeOhMyPosh)
	if err != nil {
		t.Fatalf("ListByType(oh-my-posh) error = %v", err)
	}

	if len(poshPrompts) != 1 {
		t.Fatalf("ListByType(oh-my-posh) returned %d prompts, want 1", len(poshPrompts))
	}
	if poshPrompts[0].Name != "posh-1" {
		t.Errorf("ListByType() Name = %q, want %q", poshPrompts[0].Name, "posh-1")
	}

	// List nonexistent type
	emptyList, err := store.ListByType("nonexistent")
	if err != nil {
		t.Fatalf("ListByType(nonexistent) error = %v", err)
	}
	if len(emptyList) != 0 {
		t.Errorf("ListByType(nonexistent) returned %d prompts, want 0", len(emptyList))
	}
}

func TestSQLitePromptStore_Integration_ListByCategory(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Create prompts with different categories
	prompts := []*Prompt{
		{Name: "minimal-1", Type: PromptTypeStarship, Category: "minimal"},
		{Name: "minimal-2", Type: PromptTypeOhMyPosh, Category: "minimal"},
		{Name: "powerline-1", Type: PromptTypePowerlevel10k, Category: "powerline"},
		{Name: "no-category", Type: PromptTypeStarship}, // No category
	}

	for _, p := range prompts {
		if err := store.Create(p); err != nil {
			t.Fatalf("Setup Create() error = %v", err)
		}
	}

	// List minimal prompts
	minimalPrompts, err := store.ListByCategory("minimal")
	if err != nil {
		t.Fatalf("ListByCategory() error = %v", err)
	}

	if len(minimalPrompts) != 2 {
		t.Fatalf("ListByCategory(minimal) returned %d prompts, want 2", len(minimalPrompts))
	}

	// Should be ordered by name
	expectedNames := []string{"minimal-1", "minimal-2"}
	for i, prompt := range minimalPrompts {
		if prompt.Name != expectedNames[i] {
			t.Errorf("ListByCategory() prompt[%d].Name = %q, want %q", i, prompt.Name, expectedNames[i])
		}
		if prompt.Category != "minimal" {
			t.Errorf("ListByCategory() prompt[%d].Category = %q, want %q", i, prompt.Category, "minimal")
		}
	}

	// List powerline prompts
	powerlinePrompts, err := store.ListByCategory("powerline")
	if err != nil {
		t.Fatalf("ListByCategory(powerline) error = %v", err)
	}

	if len(powerlinePrompts) != 1 {
		t.Fatalf("ListByCategory(powerline) returned %d prompts, want 1", len(powerlinePrompts))
	}
	if powerlinePrompts[0].Name != "powerline-1" {
		t.Errorf("ListByCategory() Name = %q, want %q", powerlinePrompts[0].Name, "powerline-1")
	}

	// List nonexistent category
	emptyList, err := store.ListByCategory("nonexistent")
	if err != nil {
		t.Fatalf("ListByCategory(nonexistent) error = %v", err)
	}
	if len(emptyList) != 0 {
		t.Errorf("ListByCategory(nonexistent) returned %d prompts, want 0", len(emptyList))
	}
}

func TestSQLitePromptStore_Integration_ErrorHandling(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Test getting non-existent prompt
	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) should return error")
	}

	// Test updating non-existent prompt
	nonExistentPrompt := NewPrompt("nonexistent", PromptTypeStarship)
	err = store.Update(nonExistentPrompt)
	if err == nil {
		t.Error("Update(nonexistent) should return error")
	}

	// Test deleting non-existent prompt
	err = store.Delete("nonexistent")
	if err == nil {
		t.Error("Delete(nonexistent) should return error")
	}

	// Test unique constraint violation
	prompt1 := NewPrompt("duplicate-test", PromptTypeStarship)
	if err := store.Create(prompt1); err != nil {
		t.Fatalf("Setup Create() error = %v", err)
	}

	prompt2 := NewPrompt("duplicate-test", PromptTypeOhMyPosh) // Same name
	err = store.Create(prompt2)
	if err == nil {
		t.Error("Create() with duplicate name should fail")
	}
}

func TestSQLitePromptStore_Integration_DataIntegrity(t *testing.T) {
	store := createIntegrationPromptStore(t)
	defer store.dataStore.Close()

	// Test with complex JSON data to ensure serialization works correctly
	prompt := NewPrompt("json-test", PromptTypeStarship)
	prompt.Colors = map[string]string{
		"unicode":    "🚀",
		"special":    "\\\"quotes\\\"",
		"multiline":  "line1\nline2",
		"json_chars": "{\"test\": true}",
	}
	prompt.Modules = map[string]ModuleConfig{
		"complex": {
			Symbol:   "🌟",
			Style:    "bold \"blue\"",
			Disabled: false,
		},
	}
	prompt.Tags = []string{"json", "unicode", "special-chars"}

	// Create and retrieve
	if err := store.Create(prompt); err != nil {
		t.Fatalf("Create() with complex data error = %v", err)
	}

	retrieved, err := store.Get("json-test")
	if err != nil {
		t.Fatalf("Get() complex data error = %v", err)
	}

	// Verify complex data integrity
	if retrieved.Colors["unicode"] != "🚀" {
		t.Errorf("Unicode in Colors not preserved: %q", retrieved.Colors["unicode"])
	}
	if retrieved.Colors["special"] != "\\\"quotes\\\"" {
		t.Errorf("Special chars in Colors not preserved: %q", retrieved.Colors["special"])
	}
	if retrieved.Colors["multiline"] != "line1\nline2" {
		t.Errorf("Multiline in Colors not preserved: %q", retrieved.Colors["multiline"])
	}
	if retrieved.Colors["json_chars"] != "{\"test\": true}" {
		t.Errorf("JSON chars in Colors not preserved: %q", retrieved.Colors["json_chars"])
	}

	// Verify module data
	complexModule, exists := retrieved.Modules["complex"]
	if !exists {
		t.Error("Complex module not found")
	} else {
		if complexModule.Symbol != "🌟" {
			t.Errorf("Module Symbol not preserved: %q", complexModule.Symbol)
		}
		if complexModule.Style != "bold \"blue\"" {
			t.Errorf("Module Style not preserved: %q", complexModule.Style)
		}
	}

	// Verify tags
	expectedTags := []string{"json", "unicode", "special-chars"}
	if len(retrieved.Tags) != len(expectedTags) {
		t.Errorf("Tags length = %d, want %d", len(retrieved.Tags), len(expectedTags))
	} else {
		for i, tag := range expectedTags {
			if retrieved.Tags[i] != tag {
				t.Errorf("Tags[%d] = %q, want %q", i, retrieved.Tags[i], tag)
			}
		}
	}
}
