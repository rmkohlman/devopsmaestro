// Package parametric provides parametric theme generation from base colors.
// This package analyzes the CoolNight Ocean theme to extract HSL relationships
// and generates new themes by applying the same relationships to different hues.
package parametric

import (
	"fmt"

	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/palette"
)

// Generator creates themes using parametric generation from base colors.
type Generator struct {
	// Original CoolNight Ocean color relationships extracted via HSL analysis
	baseHue        float64 // The base hue (210° for ocean blue)
	relationships  map[string]ColorRelationship
	semanticColors map[string]string // Fixed semantic colors (red, green, yellow)
}

// ColorRelationship defines how a color relates to the base theme hue.
type ColorRelationship struct {
	HueDelta    float64 // Difference from base hue (-180 to 180)
	Saturation  float64 // Absolute saturation (0-1)
	Lightness   float64 // Absolute lightness (0-1)
	IsRelative  bool    // Whether to use relative or absolute positioning
	Description string  // Human-readable description
}

// CoolNightOceanColors contains the exact reference colors from the user's config.
var CoolNightOceanColors = map[string]string{
	// Nvim colors
	"bg":           "#011628", // Main background
	"bg_dark":      "#011423", // Darker background
	"bg_highlight": "#143652", // Selection highlight
	"bg_search":    "#0A64AC", // Search highlight
	"bg_visual":    "#275378", // Visual selection
	"fg":           "#CBE0F0", // Main foreground
	"fg_dark":      "#B4D0E9", // Darker foreground
	"fg_gutter":    "#627E97", // Gutter text
	"border":       "#547998", // Border color

	// Terminal ANSI colors (from wezterm)
	"ansi_black":          "#214969",
	"ansi_red":            "#E52E2E", // Fixed semantic
	"ansi_green":          "#44FFB1", // Fixed semantic
	"ansi_yellow":         "#FFE073", // Fixed semantic
	"ansi_blue":           "#0FC5ED", // Shifted from base hue
	"ansi_magenta":        "#a277ff", // Shifted from base hue
	"ansi_cyan":           "#24EAF7", // Shifted from base hue
	"ansi_white":          "#24EAF7",
	"ansi_bright_black":   "#214969",
	"ansi_bright_red":     "#E52E2E",
	"ansi_bright_green":   "#44FFB1",
	"ansi_bright_yellow":  "#FFE073",
	"ansi_bright_blue":    "#A277FF", // Shifted from base hue
	"ansi_bright_magenta": "#a277ff",
	"ansi_bright_cyan":    "#24EAF7",
	"ansi_bright_white":   "#24EAF7",
	"cursor":              "#47FF9C", // Fixed semantic (green)
	"cursor_text":         "#011423", // Same as bg_dark
	"selection":           "#033259", // Darker variant
}

// NewGenerator creates a new parametric theme generator.
func NewGenerator() *Generator {
	g := &Generator{
		baseHue:        210.0, // CoolNight Ocean base hue
		relationships:  make(map[string]ColorRelationship),
		semanticColors: make(map[string]string),
	}

	// Analyze CoolNight Ocean colors to extract relationships
	g.analyzeColorRelationships()

	return g
}

// analyzeColorRelationships extracts HSL relationships from CoolNight Ocean colors.
func (g *Generator) analyzeColorRelationships() {
	// Fixed semantic colors that don't change with theme hue
	g.semanticColors["ansi_red"] = "#E52E2E"
	g.semanticColors["ansi_green"] = "#44FFB1"
	g.semanticColors["ansi_yellow"] = "#FFE073"
	g.semanticColors["ansi_bright_red"] = "#E52E2E"
	g.semanticColors["ansi_bright_green"] = "#44FFB1"
	g.semanticColors["ansi_bright_yellow"] = "#FFE073"
	g.semanticColors["cursor"] = "#47FF9C" // Green cursor

	// Analyze each color to extract its relationship to the base hue
	for key, hexColor := range CoolNightOceanColors {
		// Skip semantic colors
		if _, isFixed := g.semanticColors[key]; isFixed {
			continue
		}

		hsl, err := palette.HexToHSL(hexColor)
		if err != nil {
			continue
		}

		// Calculate hue delta from base
		hueDelta := hsl.H - g.baseHue

		// Normalize delta to -180 to 180 range
		for hueDelta > 180 {
			hueDelta -= 360
		}
		for hueDelta <= -180 {
			hueDelta += 360
		}

		// Store the relationship
		g.relationships[key] = ColorRelationship{
			HueDelta:    hueDelta,
			Saturation:  hsl.S,
			Lightness:   hsl.L,
			IsRelative:  isColorHueRelative(key),
			Description: getColorDescription(key),
		}
	}
}

// isColorHueRelative determines if a color should shift with the base hue.
func isColorHueRelative(colorKey string) bool {
	// Colors that should shift with base hue (blues, cyans, magentas)
	hueRelativeColors := []string{
		"ansi_blue", "ansi_cyan", "ansi_magenta",
		"ansi_bright_blue", "ansi_bright_cyan", "ansi_bright_magenta",
		"bg_search", // Blue search highlight
	}

	for _, relativeColor := range hueRelativeColors {
		if colorKey == relativeColor {
			return true
		}
	}

	// Most UI colors (backgrounds, foregrounds, borders) are hue-relative
	// Only semantic colors (red, green, yellow) are fixed
	return true
}

// getColorDescription returns a human-readable description for debugging.
func getColorDescription(colorKey string) string {
	descriptions := map[string]string{
		"bg":           "Main background",
		"bg_dark":      "Darker background",
		"bg_highlight": "Selection highlight",
		"bg_search":    "Search highlight",
		"bg_visual":    "Visual selection",
		"fg":           "Main foreground",
		"fg_dark":      "Darker foreground",
		"fg_gutter":    "Gutter text",
		"border":       "Border color",
		"ansi_blue":    "ANSI blue",
		"ansi_cyan":    "ANSI cyan",
		"ansi_magenta": "ANSI magenta",
	}

	if desc, ok := descriptions[colorKey]; ok {
		return desc
	}
	return colorKey
}

// GenerateFromHue creates a new theme using the specified hue.
func (g *Generator) GenerateFromHue(hue float64, name, description string) *theme.Theme {
	colors := make(map[string]string)

	// Add fixed semantic colors
	for key, hexColor := range g.semanticColors {
		colors[key] = hexColor
	}

	// Generate colors based on relationships
	for key, rel := range g.relationships {
		var newHue float64
		if rel.IsRelative {
			// Shift hue relative to new base hue
			newHue = hue + rel.HueDelta
		} else {
			// Keep original hue (for grays and some UI elements)
			newHue = g.baseHue + rel.HueDelta
		}

		// Normalize hue
		for newHue < 0 {
			newHue += 360
		}
		for newHue >= 360 {
			newHue -= 360
		}

		// Create new HSL color
		newHSL := palette.HSL{
			H: newHue,
			S: rel.Saturation,
			L: rel.Lightness,
		}

		colors[key] = newHSL.ToHex()
	}

	// Create theme as standalone (no external plugin repo required)
	// Standalone themes apply colors directly via nvim_set_hl()
	return &theme.Theme{
		Name:        name,
		Description: description,
		Author:      "DevOpsMaestro (parametric generator)",
		Category:    "dark",              // All CoolNight variants are dark themes
		Plugin:      theme.ThemePlugin{}, // Empty repo = standalone theme
		Colors:      colors,
	}
}

// GenerateFromHex creates a theme from a hex color string.
func (g *Generator) GenerateFromHex(hex string, name, description string) (*theme.Theme, error) {
	hsl, err := palette.HexToHSL(hex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color %s: %w", hex, err)
	}

	return g.GenerateFromHue(hsl.H, name, description), nil
}

// AnalyzeColor returns HSL analysis of a color for debugging.
func (g *Generator) AnalyzeColor(hex string) (palette.HSL, ColorRelationship, error) {
	hsl, err := palette.HexToHSL(hex)
	if err != nil {
		return palette.HSL{}, ColorRelationship{}, err
	}

	hueDelta := hsl.H - g.baseHue
	for hueDelta > 180 {
		hueDelta -= 360
	}
	for hueDelta <= -180 {
		hueDelta += 360
	}

	rel := ColorRelationship{
		HueDelta:    hueDelta,
		Saturation:  hsl.S,
		Lightness:   hsl.L,
		IsRelative:  true,
		Description: fmt.Sprintf("Custom color (%.1f° from base)", hueDelta),
	}

	return hsl, rel, nil
}

// GetBaseHue returns the base hue used for analysis.
func (g *Generator) GetBaseHue() float64 {
	return g.baseHue
}

// GetColorRelationships returns all analyzed color relationships.
func (g *Generator) GetColorRelationships() map[string]ColorRelationship {
	return g.relationships
}
