package handlers

import (
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/pkg/resource"
)

// =============================================================================
// Resource/Handler Integration Tests
//
// These tests verify the full flow from YAML resources through handlers
// to the database and back, testing the complete Resource pattern.
// =============================================================================

// createIntegrationContext creates a real resource context with SQLite backend for testing
func createIntegrationContext(t *testing.T) (resource.Context, func()) {
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

	// Create DataStore
	dataStore := db.NewSQLDataStore(driver, nil)

	// Create a resource context with the DataStore
	ctx := resource.Context{
		DataStore: dataStore,
	}

	// Return cleanup function
	cleanup := func() {
		dataStore.Close()
	}

	return ctx, cleanup
}

func TestTerminalPromptHandler_Integration_Apply(t *testing.T) {
	ctx, cleanup := createIntegrationContext(t)
	defer cleanup()

	handler := NewTerminalPromptHandler()

	// Test YAML data for a complex Starship prompt
	yamlData := `apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: integration-starship
  description: "Integration test Starship prompt"
  category: "powerline"
  tags:
    - git
    - docker
    - performance
spec:
  type: starship
  addNewline: true
  palette: "catppuccin"
  paletteRef: "catppuccin-mocha"
  format: "$all$character"
  colors:
    primary: "#89b4fa"
    secondary: "#a6e3a1"
    accent: "#f38ba8"
  modules:
    character:
      symbol: "❯"
      style: "bold green"
    directory:
      style: "bold cyan"
      disabled: false
    git_branch:
      symbol: ""
      style: "purple bold"
  character:
    success_symbol: "❯"
    error_symbol: "✗"
    vicmd_symbol: "❮"
  rawConfig: |
    format = "$all$character"
    [character]
    success_symbol = "[❯](bold green)"
    error_symbol = "[❯](bold red)"
  enabled: true`

	// Apply the YAML
	result, err := handler.Apply(ctx, []byte(yamlData))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if result == nil {
		t.Fatal("Apply() returned nil result")
	}

	// Verify the resource was created
	retrievedResult, err := handler.Get(ctx, "integration-starship")
	if err != nil {
		t.Fatalf("Get() after Apply() error = %v", err)
	}

	// Verify it's the correct type
	terminalPromptResource, ok := retrievedResult.(*TerminalPromptResource)
	if !ok {
		t.Fatalf("Get() returned wrong type: %T", retrievedResult)
	}

	// Verify the data made it through correctly
	prompt := terminalPromptResource.prompt
	if prompt.Name != "integration-starship" {
		t.Errorf("Name = %q, want %q", prompt.Name, "integration-starship")
	}
	if prompt.Type != "starship" {
		t.Errorf("Type = %q, want %q", prompt.Type, "starship")
	}
	if prompt.Description != "Integration test Starship prompt" {
		t.Errorf("Description = %q, want %q", prompt.Description, "Integration test Starship prompt")
	}
	if !prompt.AddNewline {
		t.Errorf("AddNewline = %v, want %v", prompt.AddNewline, true)
	}
	if prompt.Palette != "catppuccin" {
		t.Errorf("Palette = %q, want %q", prompt.Palette, "catppuccin")
	}
	if prompt.PaletteRef != "catppuccin-mocha" {
		t.Errorf("PaletteRef = %q, want %q", prompt.PaletteRef, "catppuccin-mocha")
	}
	if prompt.Format != "$all$character" {
		t.Errorf("Format = %q, want %q", prompt.Format, "$all$character")
	}
	if prompt.Category != "powerline" {
		t.Errorf("Category = %q, want %q", prompt.Category, "powerline")
	}

	// Verify tags
	expectedTags := []string{"git", "docker", "performance"}
	if len(prompt.Tags) != 3 {
		t.Errorf("Tags length = %d, want 3", len(prompt.Tags))
	} else {
		for i, tag := range expectedTags {
			if prompt.Tags[i] != tag {
				t.Errorf("Tags[%d] = %q, want %q", i, prompt.Tags[i], tag)
			}
		}
	}

	// Verify colors
	expectedColors := map[string]string{
		"primary":   "#89b4fa",
		"secondary": "#a6e3a1",
		"accent":    "#f38ba8",
	}
	if len(prompt.Colors) != 3 {
		t.Errorf("Colors length = %d, want 3", len(prompt.Colors))
	}
	for key, expectedValue := range expectedColors {
		if actualValue, exists := prompt.Colors[key]; !exists {
			t.Errorf("Colors[%q] missing", key)
		} else if actualValue != expectedValue {
			t.Errorf("Colors[%q] = %q, want %q", key, actualValue, expectedValue)
		}
	}

	// Verify modules
	if len(prompt.Modules) != 3 {
		t.Errorf("Modules length = %d, want 3", len(prompt.Modules))
	}
	if charModule, exists := prompt.Modules["character"]; exists {
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
	if prompt.Character == nil {
		t.Error("Character is nil")
	} else {
		if prompt.Character.SuccessSymbol != "❯" {
			t.Errorf("Character.SuccessSymbol = %q, want %q", prompt.Character.SuccessSymbol, "❯")
		}
		if prompt.Character.ErrorSymbol != "✗" {
			t.Errorf("Character.ErrorSymbol = %q, want %q", prompt.Character.ErrorSymbol, "✗")
		}
		if prompt.Character.ViCmdSymbol != "❮" {
			t.Errorf("Character.ViCmdSymbol = %q, want %q", prompt.Character.ViCmdSymbol, "❮")
		}
	}

	// Verify raw config
	expectedRawConfig := `format = "$all$character"
[character]
success_symbol = "[❯](bold green)"
error_symbol = "[❯](bold red)"
`
	if prompt.RawConfig != expectedRawConfig {
		t.Errorf("RawConfig = %q, want %q", prompt.RawConfig, expectedRawConfig)
	}

	if !prompt.Enabled {
		t.Errorf("Enabled = %v, want %v", prompt.Enabled, true)
	}
}

func TestTerminalPromptHandler_Integration_ApplyUpdate(t *testing.T) {
	ctx, cleanup := createIntegrationContext(t)
	defer cleanup()

	handler := NewTerminalPromptHandler()

	// First, apply an initial prompt
	initialYAML := `apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: update-test
  description: "Initial description"
  category: "minimal"
spec:
  type: oh-my-posh
  addNewline: true`

	_, err := handler.Apply(ctx, []byte(initialYAML))
	if err != nil {
		t.Fatalf("Initial Apply() error = %v", err)
	}

	// Now apply an update
	updatedYAML := `apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: update-test
  description: "Updated description"
  category: "powerline"
  tags:
    - updated
    - complex
spec:
  type: oh-my-posh
  addNewline: false
  colors:
    primary: "#ff0000"
    secondary: "#00ff00"
  enabled: false`

	result, err := handler.Apply(ctx, []byte(updatedYAML))
	if err != nil {
		t.Fatalf("Update Apply() error = %v", err)
	}

	// Verify the update took effect
	terminalPromptResource, ok := result.(*TerminalPromptResource)
	if !ok {
		t.Fatalf("Apply() returned wrong type: %T", result)
	}

	prompt := terminalPromptResource.prompt
	if prompt.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", prompt.Description, "Updated description")
	}
	if prompt.AddNewline {
		t.Errorf("AddNewline = %v, want %v", prompt.AddNewline, false)
	}
	if prompt.Category != "powerline" {
		t.Errorf("Category = %q, want %q", prompt.Category, "powerline")
	}
	if len(prompt.Colors) != 2 {
		t.Errorf("Colors length = %d, want 2", len(prompt.Colors))
	}
	if prompt.Colors["primary"] != "#ff0000" {
		t.Errorf("Colors[primary] = %q, want %q", prompt.Colors["primary"], "#ff0000")
	}
	if len(prompt.Tags) != 2 || prompt.Tags[0] != "updated" || prompt.Tags[1] != "complex" {
		t.Errorf("Tags = %v, want [updated complex]", prompt.Tags)
	}
	if prompt.Enabled {
		t.Errorf("Enabled = %v, want %v", prompt.Enabled, false)
	}
}

func TestTerminalPromptHandler_Integration_Delete(t *testing.T) {
	ctx, cleanup := createIntegrationContext(t)
	defer cleanup()

	handler := NewTerminalPromptHandler()

	// First, create a prompt to delete
	yamlData := `apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: delete-test
  description: "Prompt to delete"
spec:
  type: starship`

	_, err := handler.Apply(ctx, []byte(yamlData))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Verify it exists
	_, err = handler.Get(ctx, "delete-test")
	if err != nil {
		t.Fatalf("Get() before delete error = %v", err)
	}

	// Delete it
	err = handler.Delete(ctx, "delete-test")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = handler.Get(ctx, "delete-test")
	if err == nil {
		t.Error("Get() after delete should return error")
	}
}

func TestTerminalPromptHandler_Integration_List(t *testing.T) {
	ctx, cleanup := createIntegrationContext(t)
	defer cleanup()

	handler := NewTerminalPromptHandler()

	// Create multiple prompts
	promptData := []string{
		`apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: alpha-list-test
  category: minimal
spec:
  type: starship`,
		`apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: beta-list-test
  category: powerline
spec:
  type: oh-my-posh`,
		`apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: gamma-list-test
  category: advanced
spec:
  type: powerlevel10k`,
	}

	for _, yaml := range promptData {
		_, err := handler.Apply(ctx, []byte(yaml))
		if err != nil {
			t.Fatalf("Setup Apply() error = %v", err)
		}
	}

	// List all prompts
	resources, err := handler.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("List() returned %d resources, want 3", len(resources))
	}

	// Verify order (should be alphabetical)
	expectedNames := []string{"alpha-list-test", "beta-list-test", "gamma-list-test"}
	for i, resource := range resources {
		terminalPromptResource, ok := resource.(*TerminalPromptResource)
		if !ok {
			t.Fatalf("List()[%d] returned wrong type: %T", i, resource)
		}
		if terminalPromptResource.prompt.Name != expectedNames[i] {
			t.Errorf("List()[%d].Name = %q, want %q", i, terminalPromptResource.prompt.Name, expectedNames[i])
		}
	}
}

func TestTerminalPromptHandler_Integration_ErrorHandling(t *testing.T) {
	ctx, cleanup := createIntegrationContext(t)
	defer cleanup()

	handler := NewTerminalPromptHandler()

	// Test invalid YAML
	invalidYAML := `this is not valid yaml:`
	_, err := handler.Apply(ctx, []byte(invalidYAML))
	if err == nil {
		t.Error("Apply() with invalid YAML should return error")
	}

	// Test getting non-existent prompt
	_, err = handler.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) should return error")
	}

	// Test deleting non-existent prompt
	err = handler.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("Delete(nonexistent) should return error")
	}

	// Test duplicate prompt creation
	yamlData := `apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: duplicate-test
spec:
  type: starship`

	_, err = handler.Apply(ctx, []byte(yamlData))
	if err != nil {
		t.Fatalf("First Apply() error = %v", err)
	}

	// Try to create another with same name (should update, not fail)
	updatedYAML := `apiVersion: terminal/v1alpha1
kind: TerminalPrompt
metadata:
  name: duplicate-test
  description: "Updated prompt"
spec:
  type: oh-my-posh`

	result, err := handler.Apply(ctx, []byte(updatedYAML))
	if err != nil {
		t.Fatalf("Second Apply() (update) error = %v", err)
	}

	// Verify it was updated, not duplicated
	terminalPromptResource := result.(*TerminalPromptResource)
	if terminalPromptResource.prompt.Type != "oh-my-posh" {
		t.Errorf("Type = %q, want %q", terminalPromptResource.prompt.Type, "oh-my-posh")
	}
	if terminalPromptResource.prompt.Description != "Updated prompt" {
		t.Errorf("Description = %q, want %q", terminalPromptResource.prompt.Description, "Updated prompt")
	}
}
