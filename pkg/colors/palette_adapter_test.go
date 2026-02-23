package colors

import (
	"testing"

	"devopsmaestro/pkg/palette"
)

func TestColorToPaletteAdapter(t *testing.T) {
	// Create a mock ColorProvider for testing
	provider := NewMockColorProvider(
		WithMockName("test-theme"),
		WithMockColors(map[string]string{
			"primary":    "#7aa2f7",
			"secondary":  "#bb9af7",
			"accent":     "#7aa2f7",
			"success":    "#9ece6a",
			"warning":    "#e0af68",
			"error":      "#f7768e",
			"info":       "#7dcfff",
			"foreground": "#c0caf5",
			"background": "#1a1b26",
			"muted":      "#565f89",
			"highlight":  "#283457",
			"border":     "#414868",
		}),
	)

	// Create adapter
	adapter := NewColorToPaletteAdapter(provider)

	// Convert to palette
	p := adapter.ToPalette()

	// Verify palette metadata
	if p.Name != "test-theme" {
		t.Errorf("Expected name 'test-theme', got %s", p.Name)
	}

	if p.Category != palette.CategoryDark {
		t.Errorf("Expected category dark, got %s", p.Category)
	}

	if p.Description != "Generated from ColorProvider" {
		t.Errorf("Expected description 'Generated from ColorProvider', got %s", p.Description)
	}

	// Verify color mappings
	expectedMappings := map[string]string{
		palette.ColorPrimary:     "#7aa2f7",
		palette.ColorSecondary:   "#bb9af7",
		palette.ColorAccent:      "#7aa2f7",
		palette.ColorSuccess:     "#9ece6a",
		palette.ColorWarning:     "#e0af68",
		palette.ColorError:       "#f7768e",
		palette.ColorInfo:        "#7dcfff",
		palette.ColorFg:          "#c0caf5",
		palette.ColorBg:          "#1a1b26",
		palette.ColorComment:     "#565f89", // Muted maps to Comment
		palette.ColorBgHighlight: "#283457",
		palette.ColorBorder:      "#414868",
	}

	for key, expected := range expectedMappings {
		if actual := p.Colors[key]; actual != expected {
			t.Errorf("Expected %s = %s, got %s", key, expected, actual)
		}
	}

	// Verify all expected colors are present
	if len(p.Colors) != len(expectedMappings) {
		t.Errorf("Expected %d colors, got %d", len(expectedMappings), len(p.Colors))
	}
}

func TestColorToPaletteAdapter_LightTheme(t *testing.T) {
	// Create a light theme provider
	provider := NewMockColorProvider(
		WithMockName("test-light"),
		WithMockLight(),
		WithMockColors(map[string]string{
			"primary":    "#0969da",
			"foreground": "#24292e",
			"background": "#ffffff",
		}),
	)

	adapter := NewColorToPaletteAdapter(provider)
	p := adapter.ToPalette()

	// Verify light theme category
	if p.Category != palette.CategoryLight {
		t.Errorf("Expected category light, got %s", p.Category)
	}

	// Verify color mappings
	if p.Colors[palette.ColorPrimary] != "#0969da" {
		t.Errorf("Expected primary color #0969da, got %s", p.Colors[palette.ColorPrimary])
	}

	if p.Colors[palette.ColorFg] != "#24292e" {
		t.Errorf("Expected foreground color #24292e, got %s", p.Colors[palette.ColorFg])
	}

	if p.Colors[palette.ColorBg] != "#ffffff" {
		t.Errorf("Expected background color #ffffff, got %s", p.Colors[palette.ColorBg])
	}
}

func TestColorToPaletteAdapter_NoColorProvider(t *testing.T) {
	// Test with NoColorProvider (all colors empty)
	provider := NewNoColorProvider()

	adapter := NewColorToPaletteAdapter(provider)
	p := adapter.ToPalette()

	// Verify palette is not nil
	if p == nil {
		t.Fatal("Expected palette to not be nil for NoColorProvider")
	}

	// Verify metadata
	if p.Name != "no-color" {
		t.Errorf("Expected name 'no-color', got %s", p.Name)
	}

	if p.Category != palette.CategoryDark {
		t.Errorf("Expected category dark, got %s", p.Category)
	}

	// Verify no colors are set (all empty strings filtered out)
	if len(p.Colors) != 0 {
		t.Errorf("Expected no colors for NoColorProvider, got %d colors", len(p.Colors))
	}
}

func TestToPalette_ConvenienceFunction(t *testing.T) {
	// Create a mock provider
	provider := NewMockColorProvider(
		WithMockName("test-convenience"),
		WithMockColor("primary", "#ff5555"),
	)

	// Use convenience function
	p := ToPalette(provider)

	// Verify it works the same as the adapter
	if p == nil {
		t.Fatal("Expected palette to not be nil")
	}

	if p.Name != "test-convenience" {
		t.Errorf("Expected name 'test-convenience', got %s", p.Name)
	}

	if p.Colors[palette.ColorPrimary] != "#ff5555" {
		t.Errorf("Expected primary color #ff5555, got %s", p.Colors[palette.ColorPrimary])
	}
}

func TestColorToPaletteAdapter_EmptyColors(t *testing.T) {
	// Test with a provider that has some empty colors
	provider := NewMockColorProvider(
		WithMockName("partial-colors"),
		WithMockColors(map[string]string{
			"primary": "#ff0000",
			// Leave other colors as defaults or empty
		}),
	)

	// Clear some colors to test empty handling
	provider.SetColor("secondary", "")
	provider.SetColor("accent", "")

	adapter := NewColorToPaletteAdapter(provider)
	p := adapter.ToPalette()

	// Should have primary color
	if p.Colors[palette.ColorPrimary] != "#ff0000" {
		t.Errorf("Expected primary color #ff0000, got %s", p.Colors[palette.ColorPrimary])
	}

	// Empty colors should not be in the palette
	if _, exists := p.Colors[palette.ColorSecondary]; exists && p.Colors[palette.ColorSecondary] == "" {
		t.Error("Empty secondary color should not be in palette")
	}

	if _, exists := p.Colors[palette.ColorAccent]; exists && p.Colors[palette.ColorAccent] == "" {
		t.Error("Empty accent color should not be in palette")
	}
}
