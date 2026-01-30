package library

import (
	"testing"
)

func TestList(t *testing.T) {
	themes, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(themes) == 0 {
		t.Error("expected at least one theme in library")
	}

	// Check for expected themes
	expectedThemes := []string{
		"tokyonight-night",
		"tokyonight-custom",
		"catppuccin-mocha",
		"gruvbox-dark",
		"nord",
	}

	themeMap := make(map[string]bool)
	for _, theme := range themes {
		themeMap[theme.Name] = true
	}

	for _, expected := range expectedThemes {
		if !themeMap[expected] {
			t.Errorf("expected theme '%s' not found in library", expected)
		}
	}
}

func TestGet(t *testing.T) {
	theme, err := Get("tokyonight-custom")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if theme.Name != "tokyonight-custom" {
		t.Errorf("expected name 'tokyonight-custom', got '%s'", theme.Name)
	}
	if theme.Plugin.Repo != "folke/tokyonight.nvim" {
		t.Errorf("expected repo 'folke/tokyonight.nvim', got '%s'", theme.Plugin.Repo)
	}
	if theme.Colors["bg"] != "#011628" {
		t.Errorf("expected custom bg color '#011628', got '%s'", theme.Colors["bg"])
	}
}

func TestGet_NotFound(t *testing.T) {
	_, err := Get("nonexistent-theme")
	if err == nil {
		t.Error("expected error for nonexistent theme")
	}
}

func TestHas(t *testing.T) {
	if !Has("catppuccin-mocha") {
		t.Error("expected catppuccin-mocha to exist in library")
	}
	if Has("nonexistent") {
		t.Error("expected nonexistent to not exist in library")
	}
}

func TestCategories(t *testing.T) {
	categories, err := Categories()
	if err != nil {
		t.Fatalf("Categories failed: %v", err)
	}

	if len(categories) == 0 {
		t.Error("expected at least one category")
	}

	// Check for expected categories
	hasLight := false
	hasDark := false
	for _, cat := range categories {
		if cat == "light" {
			hasLight = true
		}
		if cat == "dark" {
			hasDark = true
		}
	}

	if !hasDark {
		t.Error("expected 'dark' category")
	}
	if !hasLight {
		t.Error("expected 'light' category")
	}
}

func TestListByCategory(t *testing.T) {
	darkThemes, err := ListByCategory("dark")
	if err != nil {
		t.Fatalf("ListByCategory failed: %v", err)
	}

	if len(darkThemes) == 0 {
		t.Error("expected at least one dark theme")
	}

	for _, theme := range darkThemes {
		if theme.Category != "dark" {
			t.Errorf("expected category 'dark', got '%s' for theme '%s'", theme.Category, theme.Name)
		}
	}
}

func TestGetRaw(t *testing.T) {
	data, err := GetRaw("nord")
	if err != nil {
		t.Fatalf("GetRaw failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty YAML data")
	}
}
