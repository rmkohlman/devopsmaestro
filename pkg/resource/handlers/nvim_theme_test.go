package handlers

import (
	"testing"

	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/resource"
)

func TestNvimThemeHandler_Kind(t *testing.T) {
	h := NewNvimThemeHandler()
	if h.Kind() != KindNvimTheme {
		t.Errorf("Kind() = %v, want %v", h.Kind(), KindNvimTheme)
	}
}

func TestNvimThemeHandler_Apply(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: catppuccin-mocha
  description: Soothing pastel theme
  category: dark
spec:
  plugin:
    repo: catppuccin/nvim
  style: mocha
  colors:
    bg: "#1e1e2e"
    fg: "#cdd6f4"`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindNvimTheme {
		t.Errorf("Apply() resource.Kind = %v, want %v", res.GetKind(), KindNvimTheme)
	}
	if res.GetName() != "catppuccin-mocha" {
		t.Errorf("Apply() resource.Name = %v, want catppuccin-mocha", res.GetName())
	}

	// Verify it's in the store
	stored, err := memStore.Get("catppuccin-mocha")
	if err != nil {
		t.Errorf("Theme not found in store: %v", err)
	}
	if stored.Plugin.Repo != "catppuccin/nvim" {
		t.Errorf("Stored theme plugin.repo = %v, want catppuccin/nvim", stored.Plugin.Repo)
	}
}

func TestNvimThemeHandler_Get(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Pre-populate store
	memStore.Save(&theme.Theme{
		Name:   "tokyonight",
		Plugin: theme.ThemePlugin{Repo: "folke/tokyonight.nvim"},
	})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	res, err := h.Get(ctx, "tokyonight")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if res.GetName() != "tokyonight" {
		t.Errorf("Get() resource.Name = %v, want tokyonight", res.GetName())
	}

	// Cast to access underlying theme
	tr, ok := res.(*NvimThemeResource)
	if !ok {
		t.Fatalf("Get() result is not *NvimThemeResource")
	}
	if tr.Theme().Plugin.Repo != "folke/tokyonight.nvim" {
		t.Errorf("Get() theme.Plugin.Repo = %v, want folke/tokyonight.nvim", tr.Theme().Plugin.Repo)
	}
}

func TestNvimThemeHandler_List(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Pre-populate store
	memStore.Save(&theme.Theme{Name: "theme1", Plugin: theme.ThemePlugin{Repo: "user/theme1"}})
	memStore.Save(&theme.Theme{Name: "theme2", Plugin: theme.ThemePlugin{Repo: "user/theme2"}})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// With library fallback, we should have user themes + library themes
	// We know there are 2 user themes and ~34 library themes
	if len(resources) < 30 {
		t.Errorf("List() returned %d resources, expected at least 30 (2 user + library themes)", len(resources))
	}

	// Verify our user themes are present
	names := make(map[string]bool)
	for _, res := range resources {
		names[res.GetName()] = true
	}

	expectedUserThemes := []string{"theme1", "theme2"}
	for _, name := range expectedUserThemes {
		if !names[name] {
			t.Errorf("Expected user theme %q not found in list", name)
		}
	}
}

func TestNvimThemeHandler_Delete(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Pre-populate store
	memStore.Save(&theme.Theme{Name: "to-delete", Plugin: theme.ThemePlugin{Repo: "user/to-delete"}})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	err := h.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = memStore.Get("to-delete")
	if err == nil {
		t.Error("Delete() did not remove theme from store")
	}
}

func TestNvimThemeHandler_ToYAML(t *testing.T) {
	h := NewNvimThemeHandler()

	res := &NvimThemeResource{
		theme: &theme.Theme{
			Name:        "test-theme",
			Description: "Test description",
			Plugin:      theme.ThemePlugin{Repo: "user/test-theme"},
		},
	}

	yaml, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Verify YAML contains expected content
	yamlStr := string(yaml)
	if !contains(yamlStr, "kind: NvimTheme") {
		t.Error("ToYAML() missing 'kind: NvimTheme'")
	}
	if !contains(yamlStr, "name: test-theme") {
		t.Error("ToYAML() missing 'name: test-theme'")
	}
}

// Test library fallback functionality
func TestNvimThemeHandler_LibraryFallback_Get(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	// Test getting a library theme that doesn't exist in user store
	res, err := h.Get(ctx, "coolnight-ocean")
	if err != nil {
		t.Fatalf("Get(coolnight-ocean) error = %v", err)
	}

	if res.GetName() != "coolnight-ocean" {
		t.Errorf("Get(coolnight-ocean) resource.Name = %v, want coolnight-ocean", res.GetName())
	}

	// Cast to access underlying theme
	tr, ok := res.(*NvimThemeResource)
	if !ok {
		t.Fatalf("Get() result is not *NvimThemeResource")
	}

	// Verify it's actually from the library (should have specific plugin repo)
	if tr.Theme().Plugin.Repo == "" {
		t.Error("Library theme should have a plugin repo")
	}
}

func TestNvimThemeHandler_LibraryFallback_UserOverride(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Add a user theme with same name as library theme
	memStore.Save(&theme.Theme{
		Name:   "dracula",
		Plugin: theme.ThemePlugin{Repo: "user/custom-dracula"},
	})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	// Get should return user version, not library version
	res, err := h.Get(ctx, "dracula")
	if err != nil {
		t.Fatalf("Get(dracula) error = %v", err)
	}

	tr, ok := res.(*NvimThemeResource)
	if !ok {
		t.Fatalf("Get() result is not *NvimThemeResource")
	}

	// Should be user theme, not library theme
	if tr.Theme().Plugin.Repo != "user/custom-dracula" {
		t.Errorf("User theme should override library theme, got repo: %v", tr.Theme().Plugin.Repo)
	}
}

func TestNvimThemeHandler_LibraryFallback_NotFound(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	// Try to get a theme that doesn't exist in either user store or library
	_, err := h.Get(ctx, "nonexistent-theme")
	if err == nil {
		t.Error("Get(nonexistent-theme) should return error")
	}
}

func TestNvimThemeHandler_LibraryFallback_List(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Add some user themes
	memStore.Save(&theme.Theme{Name: "user-theme1", Plugin: theme.ThemePlugin{Repo: "user/theme1"}})
	memStore.Save(&theme.Theme{Name: "user-theme2", Plugin: theme.ThemePlugin{Repo: "user/theme2"}})
	// Add a user theme that overrides a library theme
	memStore.Save(&theme.Theme{Name: "dracula", Plugin: theme.ThemePlugin{Repo: "user/custom-dracula"}})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should have user themes + library themes (minus duplicates)
	// We expect at least 3 themes: 2 unique user themes + 1 overridden user theme + many library themes
	if len(resources) < 30 { // We know there are ~34 library themes
		t.Errorf("List() returned %d resources, expected at least 30 (library + user themes)", len(resources))
	}

	// Verify user themes are present
	names := make(map[string]bool)
	for _, res := range resources {
		names[res.GetName()] = true
	}

	expectedUserThemes := []string{"user-theme1", "user-theme2", "dracula"}
	for _, name := range expectedUserThemes {
		if !names[name] {
			t.Errorf("Expected user theme %q not found in list", name)
		}
	}

	// Verify some known library themes are present
	expectedLibraryThemes := []string{"coolnight-ocean", "catppuccin-mocha", "nord"}
	for _, name := range expectedLibraryThemes {
		if !names[name] {
			t.Errorf("Expected library theme %q not found in list", name)
		}
	}

	// Verify no duplicates
	seenNames := make(map[string]int)
	for _, res := range resources {
		seenNames[res.GetName()]++
	}
	for name, count := range seenNames {
		if count > 1 {
			t.Errorf("Theme %q appears %d times in list (should be unique)", name, count)
		}
	}
}

func TestNvimThemeHandler_LibraryFallback_ListWithLibraryError(t *testing.T) {
	// This test verifies that if library.List() fails, we still return user themes
	// We can't easily simulate library failure, but we can test with an empty store
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Add user themes
	memStore.Save(&theme.Theme{Name: "user-only", Plugin: theme.ThemePlugin{Repo: "user/theme"}})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() should not error even if library unavailable: %v", err)
	}

	// Should have at least the user theme plus library themes
	userFound := false
	for _, res := range resources {
		if res.GetName() == "user-only" {
			userFound = true
			break
		}
	}

	if !userFound {
		t.Error("User theme not found in list when library might be unavailable")
	}
}

func TestNvimThemeResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		theme   *theme.Theme
		wantErr bool
	}{
		{
			name:    "valid theme",
			theme:   &theme.Theme{Name: "test", Plugin: theme.ThemePlugin{Repo: "user/test"}},
			wantErr: false,
		},
		{
			name:    "missing name",
			theme:   &theme.Theme{Name: "", Plugin: theme.ThemePlugin{Repo: "user/test"}},
			wantErr: true,
		},
		{
			name:    "missing repo",
			theme:   &theme.Theme{Name: "test", Plugin: theme.ThemePlugin{Repo: ""}},
			wantErr: true,
		},
		{
			name: "invalid color",
			theme: &theme.Theme{
				Name:   "test",
				Plugin: theme.ThemePlugin{Repo: "user/test"},
				Colors: map[string]string{"bg": "not-a-color"},
			},
			wantErr: true,
		},
		{
			name: "valid colors",
			theme: &theme.Theme{
				Name:   "test",
				Plugin: theme.ThemePlugin{Repo: "user/test"},
				Colors: map[string]string{"bg": "#1e1e2e", "fg": "#cdd6f4"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &NvimThemeResource{theme: tt.theme}
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
