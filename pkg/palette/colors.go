package palette

import (
	"fmt"
	"regexp"
	"strings"
)

// hexColorRegex matches valid hex color formats: #RGB, #RRGGBB, #RRGGBBAA
var hexColorRegex = regexp.MustCompile(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$`)

// Validate checks if the palette is valid.
func (p *Palette) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("palette name is required")
	}

	// Validate color formats
	for key, color := range p.Colors {
		if color == "" {
			continue // Empty colors are allowed
		}
		if !IsValidHexColor(color) {
			return fmt.Errorf("invalid color format for %s: %s (expected hex like #RRGGBB)", key, color)
		}
	}

	// Validate category if set
	if p.Category != "" && p.Category != CategoryDark && p.Category != CategoryLight && p.Category != CategoryBoth {
		return fmt.Errorf("invalid category: %s (expected dark, light, or both)", p.Category)
	}

	return nil
}

// IsValidHexColor checks if a string is a valid hex color.
func IsValidHexColor(color string) bool {
	return hexColorRegex.MatchString(color)
}

// NormalizeHexColor normalizes a hex color to #RRGGBB format.
// - #RGB becomes #RRGGBB
// - #RRGGBBAA strips the alpha channel
// - Already #RRGGBB returns unchanged
func NormalizeHexColor(color string) string {
	if !IsValidHexColor(color) {
		return color
	}

	color = strings.ToLower(color)
	hex := color[1:] // Remove #

	switch len(hex) {
	case 3:
		// #RGB -> #RRGGBB
		return fmt.Sprintf("#%c%c%c%c%c%c", hex[0], hex[0], hex[1], hex[1], hex[2], hex[2])
	case 8:
		// #RRGGBBAA -> #RRGGBB (strip alpha)
		return "#" + hex[:6]
	default:
		return color
	}
}

// ParseRGB extracts RGB values from a hex color string.
// Returns red, green, blue values (0-255) and any error.
func ParseRGB(color string) (r, g, b int, err error) {
	if !IsValidHexColor(color) {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", color)
	}

	normalized := NormalizeHexColor(color)
	hex := normalized[1:] // Remove #

	_, err = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse color %s: %w", color, err)
	}

	return r, g, b, nil
}

// ToRGBString converts a hex color to an RGB string like "rgb(255, 128, 64)".
func ToRGBString(color string) (string, error) {
	r, g, b, err := ParseRGB(color)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b), nil
}

// ToANSI256 converts a hex color to the nearest ANSI 256-color code.
// This is useful for terminal applications that support 256 colors.
func ToANSI256(color string) (int, error) {
	r, g, b, err := ParseRGB(color)
	if err != nil {
		return 0, err
	}

	// Check for grayscale colors (232-255)
	if r == g && g == b {
		if r < 8 {
			return 16, nil // black
		}
		if r > 248 {
			return 231, nil // white
		}
		return int((float64(r)-8)/10) + 232, nil
	}

	// Convert to 6x6x6 color cube (16-231)
	rIdx := int(float64(r) / 255.0 * 5.0)
	gIdx := int(float64(g) / 255.0 * 5.0)
	bIdx := int(float64(b) / 255.0 * 5.0)

	return 16 + (36 * rIdx) + (6 * gIdx) + bIdx, nil
}

// =============================================================================
// Well-Known Palettes
// =============================================================================

// RequiredColors returns a list of color keys that are typically required
// for a complete theme.
var RequiredColors = []string{
	ColorBg,
	ColorFg,
	ColorError,
	ColorWarning,
	ColorInfo,
}

// AllColorKeys returns all defined color key constants.
func AllColorKeys() []string {
	return []string{
		// Backgrounds
		ColorBg, ColorBgDark, ColorBgHighlight, ColorBgSearch,
		ColorBgVisual, ColorBgFloat, ColorBgPopup, ColorBgSidebar, ColorBgStatusline,
		// Foregrounds
		ColorFg, ColorFgDark, ColorFgGutter, ColorFgSidebar,
		// UI
		ColorBorder, ColorComment,
		// Diagnostics
		ColorError, ColorWarning, ColorInfo, ColorHint, ColorSuccess,
		// Accents
		ColorPrimary, ColorSecondary, ColorAccent,
	}
}
