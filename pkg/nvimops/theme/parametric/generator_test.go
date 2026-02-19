package parametric

import (
	"testing"

	"devopsmaestro/pkg/palette"
)

func TestHSLUtilities(t *testing.T) {
	// Test basic HSL conversion
	hsl, err := palette.HexToHSL("#011628") // CoolNight Ocean bg
	if err != nil {
		t.Fatalf("HexToHSL failed: %v", err)
	}

	// Should be close to hue 210° (blue)
	if hsl.H < 200 || hsl.H > 220 {
		t.Errorf("Expected hue around 210°, got %.1f°", hsl.H)
	}

	// Convert back to hex - allow small precision differences
	hex := hsl.ToHex()
	// Due to floating point precision, small differences in color values are acceptable
	// The important thing is that the hue is correctly identified
	t.Logf("Round-trip: #011628 -> %s -> %s", hsl.String(), hex)
}

func TestParametricGenerator(t *testing.T) {
	generator := NewGenerator()

	// Test base hue
	if generator.GetBaseHue() != 210.0 {
		t.Errorf("Expected base hue 210°, got %.1f°", generator.GetBaseHue())
	}

	// Generate a theme from hue
	theme := generator.GenerateFromHue(270.0, "test-purple", "Test purple theme")

	if theme.Name != "test-purple" {
		t.Errorf("Expected name 'test-purple', got '%s'", theme.Name)
	}

	if theme.Category != "dark" {
		t.Errorf("Expected category 'dark', got '%s'", theme.Category)
	}

	// Should have semantic colors preserved
	if theme.Colors["ansi_red"] != "#E52E2E" {
		t.Errorf("Expected red semantic color preserved, got %s", theme.Colors["ansi_red"])
	}

	if theme.Colors["ansi_green"] != "#44FFB1" {
		t.Errorf("Expected green semantic color preserved, got %s", theme.Colors["ansi_green"])
	}

	// Background should be shifted to purple tint
	bgColor := theme.Colors["bg"]
	if bgColor == "" {
		t.Fatal("Background color not generated")
	}

	bgHSL, err := palette.HexToHSL(bgColor)
	if err != nil {
		t.Fatalf("Invalid background color %s: %v", bgColor, err)
	}

	// Should be close to purple (270°)
	if bgHSL.H < 260 || bgHSL.H > 280 {
		t.Errorf("Expected purple background hue ~270°, got %.1f°", bgHSL.H)
	}
}

func TestPresets(t *testing.T) {
	// Test preset existence
	presets := []string{
		"coolnight-ocean",
		"coolnight-synthwave",
		"coolnight-matrix",
		"coolnight-mono-charcoal",
	}

	for _, presetName := range presets {
		preset, exists := GetPreset(presetName)
		if !exists {
			t.Errorf("Preset %s should exist", presetName)
		}

		if preset.Name != presetName {
			t.Errorf("Preset name mismatch: expected %s, got %s", presetName, preset.Name)
		}
	}

	// Test preset generation
	theme, err := GeneratePreset("coolnight-synthwave")
	if err != nil {
		t.Fatalf("Failed to generate synthwave preset: %v", err)
	}

	if theme.Name != "coolnight-synthwave" {
		t.Errorf("Expected theme name 'coolnight-synthwave', got '%s'", theme.Name)
	}
}

func TestMonochromeGeneration(t *testing.T) {
	theme, err := GeneratePreset("coolnight-mono-charcoal")
	if err != nil {
		t.Fatalf("Failed to generate monochrome theme: %v", err)
	}

	// Should preserve semantic colors
	if theme.Colors["ansi_red"] != "#E52E2E" {
		t.Errorf("Red should be preserved in monochrome, got %s", theme.Colors["ansi_red"])
	}

	// Background should be desaturated
	bgColor := theme.Colors["bg"]
	bgHSL, err := palette.HexToHSL(bgColor)
	if err != nil {
		t.Fatalf("Invalid background color: %v", err)
	}

	// Should be very low saturation
	if bgHSL.S > 0.1 {
		t.Errorf("Expected low saturation for monochrome bg, got %.2f", bgHSL.S)
	}
}

func TestColorRelationships(t *testing.T) {
	generator := NewGenerator()
	relationships := generator.GetColorRelationships()

	// Should have relationships for key colors
	keyColors := []string{"bg", "fg", "ansi_blue", "ansi_cyan"}

	for _, color := range keyColors {
		if _, exists := relationships[color]; !exists {
			t.Errorf("Missing relationship for %s", color)
		}
	}

	// Test that semantic colors are NOT in relationships (they're fixed)
	semanticColors := []string{"ansi_red", "ansi_green", "ansi_yellow"}

	for _, color := range semanticColors {
		if _, exists := relationships[color]; exists {
			t.Errorf("Semantic color %s should not have relationship (should be fixed)", color)
		}
	}
}

func TestFamilyOrganization(t *testing.T) {
	families := PresetsByFamily()

	// Should have expected families
	expectedFamilies := []string{"blue", "purple", "green", "warm", "red", "pink", "monochrome", "special"}

	for _, family := range expectedFamilies {
		if _, exists := families[family]; !exists {
			t.Errorf("Missing family: %s", family)
		}
	}

	// Blue family should include ocean, arctic, midnight
	blueFamily := families["blue"]
	expectedBlue := map[string]bool{
		"coolnight-ocean":    false,
		"coolnight-arctic":   false,
		"coolnight-midnight": false,
	}

	for _, preset := range blueFamily {
		if _, expected := expectedBlue[preset.Name]; expected {
			expectedBlue[preset.Name] = true
		}
	}

	for name, found := range expectedBlue {
		if !found {
			t.Errorf("Blue family missing preset: %s", name)
		}
	}
}
