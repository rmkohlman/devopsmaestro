package parametric

import (
	"fmt"

	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/palette"
)

// ThemePreset defines a parametric theme variant with its metadata.
type ThemePreset struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Family      string  `json:"family"`
	Hue         float64 `json:"hue"`
	Notes       string  `json:"notes,omitempty"`
}

// AllPresets contains all 21 CoolNight theme presets organized by family.
var AllPresets = map[string]ThemePreset{
	// Blue Family (original ocean theme)
	"coolnight-ocean": {
		Name:        "coolnight-ocean",
		Description: "CoolNight Ocean - The original deep blue theme",
		Family:      "blue",
		Hue:         210.0, // Base hue - original CoolNight Ocean
		Notes:       "Original theme - base for all parametric generation",
	},
	"coolnight-arctic": {
		Name:        "coolnight-arctic",
		Description: "CoolNight Arctic - Cooler blue with ice-like tones",
		Family:      "blue",
		Hue:         200.0,
		Notes:       "Slightly cooler, more cyan-leaning blue",
	},
	"coolnight-midnight": {
		Name:        "coolnight-midnight",
		Description: "CoolNight Midnight - Deep midnight blue",
		Family:      "blue",
		Hue:         230.0,
		Notes:       "Deeper, more purple-leaning blue",
	},

	// Purple Family
	"coolnight-synthwave": {
		Name:        "coolnight-synthwave",
		Description: "CoolNight Synthwave - Neon purple inspired by 80s aesthetics",
		Family:      "purple",
		Hue:         270.0,
		Notes:       "Classic purple, perfect for synthwave vibes",
	},
	"coolnight-violet": {
		Name:        "coolnight-violet",
		Description: "CoolNight Violet - Soft violet tones",
		Family:      "purple",
		Hue:         280.0,
		Notes:       "Lighter purple with more blue undertones",
	},
	"coolnight-grape": {
		Name:        "coolnight-grape",
		Description: "CoolNight Grape - Rich grape purple",
		Family:      "purple",
		Hue:         290.0,
		Notes:       "Deeper purple with red undertones",
	},

	// Green Family
	"coolnight-matrix": {
		Name:        "coolnight-matrix",
		Description: "CoolNight Matrix - Digital green like the movie",
		Family:      "green",
		Hue:         150.0,
		Notes:       "True green, perfect for Matrix-style coding",
	},
	"coolnight-forest": {
		Name:        "coolnight-forest",
		Description: "CoolNight Forest - Deep forest green",
		Family:      "green",
		Hue:         140.0,
		Notes:       "Darker green with yellow undertones",
	},
	"coolnight-mint": {
		Name:        "coolnight-mint",
		Description: "CoolNight Mint - Fresh mint green",
		Family:      "green",
		Hue:         165.0,
		Notes:       "Lighter green with cyan undertones",
	},

	// Warm Family
	"coolnight-sunset": {
		Name:        "coolnight-sunset",
		Description: "CoolNight Sunset - Warm orange sunset tones",
		Family:      "warm",
		Hue:         25.0,
		Notes:       "Orange-red, reminiscent of sunset skies",
	},
	"coolnight-ember": {
		Name:        "coolnight-ember",
		Description: "CoolNight Ember - Glowing ember red-orange",
		Family:      "warm",
		Hue:         15.0,
		Notes:       "Warmer, more red-leaning orange",
	},
	"coolnight-gold": {
		Name:        "coolnight-gold",
		Description: "CoolNight Gold - Rich golden yellow",
		Family:      "warm",
		Hue:         45.0,
		Notes:       "Golden yellow, luxury aesthetic",
	},

	// Red/Pink Family
	"coolnight-rose": {
		Name:        "coolnight-rose",
		Description: "CoolNight Rose - Elegant rose pink",
		Family:      "pink",
		Hue:         350.0,
		Notes:       "Pink with red undertones, elegant feel",
	},
	"coolnight-crimson": {
		Name:        "coolnight-crimson",
		Description: "CoolNight Crimson - Deep crimson red",
		Family:      "red",
		Hue:         0.0,
		Notes:       "Pure red, bold and striking",
	},
	"coolnight-sakura": {
		Name:        "coolnight-sakura",
		Description: "CoolNight Sakura - Soft cherry blossom pink",
		Family:      "pink",
		Hue:         330.0,
		Notes:       "Softer pink, inspired by cherry blossoms",
	},

	// Monochrome Family (special handling for desaturated themes)
	"coolnight-mono-charcoal": {
		Name:        "coolnight-mono-charcoal",
		Description: "CoolNight Mono Charcoal - Pure grayscale theme",
		Family:      "monochrome",
		Hue:         0.0, // No hue - pure grayscale
		Notes:       "Completely desaturated, pure black/white/gray",
	},
	"coolnight-mono-slate": {
		Name:        "coolnight-mono-slate",
		Description: "CoolNight Mono Slate - Cool gray with subtle blue tint",
		Family:      "monochrome",
		Hue:         210.0, // Keep original hue but low saturation
		Notes:       "Low saturation blue-gray, sophisticated",
	},
	"coolnight-mono-warm": {
		Name:        "coolnight-mono-warm",
		Description: "CoolNight Mono Warm - Warm gray with subtle brown tint",
		Family:      "monochrome",
		Hue:         30.0, // Warm brown hue but low saturation
		Notes:       "Low saturation brown-gray, cozy feel",
	},

	// Special Themes (inspired by popular themes)
	"coolnight-nord": {
		Name:        "coolnight-nord",
		Description: "CoolNight Nord - Inspired by the Nord color palette",
		Family:      "special",
		Hue:         210.0, // Similar to original but with Nord-like adjustments
		Notes:       "Nord-inspired blue-gray palette",
	},
	"coolnight-dracula": {
		Name:        "coolnight-dracula",
		Description: "CoolNight Dracula - Inspired by Dracula theme colors",
		Family:      "special",
		Hue:         260.0, // Purple-ish like Dracula
		Notes:       "Dracula-inspired purple tones",
	},
	"coolnight-solarized": {
		Name:        "coolnight-solarized",
		Description: "CoolNight Solarized - Inspired by Solarized Dark",
		Family:      "special",
		Hue:         195.0, // Blue-cyan like Solarized
		Notes:       "Solarized-inspired blue-cyan base",
	},
}

// FamilyNames returns all theme family names.
func FamilyNames() []string {
	return []string{
		"blue",
		"purple",
		"green",
		"warm",
		"red",
		"pink",
		"monochrome",
		"special",
	}
}

// PresetsByFamily returns all presets organized by family.
func PresetsByFamily() map[string][]ThemePreset {
	families := make(map[string][]ThemePreset)

	for _, preset := range AllPresets {
		families[preset.Family] = append(families[preset.Family], preset)
	}

	return families
}

// GetPreset returns a preset by name.
func GetPreset(name string) (ThemePreset, bool) {
	preset, exists := AllPresets[name]
	return preset, exists
}

// ListPresets returns all preset names sorted by family.
func ListPresets() []string {
	var names []string

	// Add in family order for better organization
	for _, family := range FamilyNames() {
		for _, preset := range AllPresets {
			if preset.Family == family {
				names = append(names, preset.Name)
			}
		}
	}

	return names
}

// GeneratePreset creates a theme from a preset definition.
func GeneratePreset(presetName string) (*theme.Theme, error) {
	preset, exists := GetPreset(presetName)
	if !exists {
		return nil, fmt.Errorf("preset not found: %s", presetName)
	}

	generator := NewGenerator()

	// Handle special cases for monochrome themes
	if preset.Family == "monochrome" {
		return generator.generateMonochromeTheme(preset), nil
	}

	return generator.GenerateFromHue(preset.Hue, preset.Name, preset.Description), nil
}

// generateMonochromeTheme creates special low-saturation themes.
func (g *Generator) generateMonochromeTheme(preset ThemePreset) *theme.Theme {
	// Generate base theme
	baseTheme := g.GenerateFromHue(preset.Hue, preset.Name, preset.Description)

	// Apply desaturation based on preset
	switch preset.Name {
	case "coolnight-mono-charcoal":
		// Pure grayscale - remove all saturation
		return g.desaturateTheme(baseTheme, 0.0)
	case "coolnight-mono-slate", "coolnight-mono-warm":
		// Low saturation - keep subtle tint
		return g.desaturateTheme(baseTheme, 0.15)
	}

	return baseTheme
}

// desaturateTheme reduces saturation of all colors in a theme.
func (g *Generator) desaturateTheme(t *theme.Theme, targetSaturation float64) *theme.Theme {
	desaturated := &theme.Theme{
		Name:        t.Name,
		Description: t.Description,
		Author:      t.Author,
		Category:    t.Category,
		Plugin:      t.Plugin,
		Colors:      make(map[string]string),
	}

	for key, hexColor := range t.Colors {
		// Skip semantic colors (keep red, green, yellow vibrant)
		if g.isSemanticColor(key) {
			desaturated.Colors[key] = hexColor
			continue
		}

		hsl, err := palette.HexToHSL(hexColor)
		if err != nil {
			desaturated.Colors[key] = hexColor
			continue
		}

		// Apply target saturation
		hsl.S = targetSaturation
		desaturated.Colors[key] = hsl.ToHex()
	}

	return desaturated
}

// isSemanticColor checks if a color should remain saturated in monochrome themes.
func (g *Generator) isSemanticColor(colorKey string) bool {
	semanticColors := []string{
		"ansi_red", "ansi_green", "ansi_yellow",
		"ansi_bright_red", "ansi_bright_green", "ansi_bright_yellow",
		"cursor",
	}

	for _, semantic := range semanticColors {
		if colorKey == semantic {
			return true
		}
	}

	return false
}
