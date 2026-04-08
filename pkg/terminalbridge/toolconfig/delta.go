package toolconfig

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroPalette"
)

// DeltaGenerator generates git delta configuration that maps palette colors
// to delta's color settings for diff viewing.
type DeltaGenerator struct{}

// ToolName returns "delta".
func (g *DeltaGenerator) ToolName() string { return "delta" }

// Description returns a description of what this generator produces.
func (g *DeltaGenerator) Description() string {
	return "Git delta theme configuration (gitconfig format)"
}

// Generate produces a [delta] gitconfig section with colors from the palette.
func (g *DeltaGenerator) Generate(pal *palette.Palette) (string, error) {
	if pal == nil {
		return "", fmt.Errorf("palette is required for delta config generation")
	}

	// Map palette colors to delta config keys
	bg := paletteGet(pal, palette.ColorBg, "#1a1b26")
	fg := paletteGet(pal, palette.ColorFg, "#c0caf5")
	success := paletteGet(pal, palette.ColorSuccess, "#9ece6a")
	errorC := paletteGet(pal, palette.ColorError, "#f7768e")
	info := paletteGet(pal, palette.ColorInfo, "#7aa2f7")
	highlight := paletteGet(pal, palette.ColorBgHighlight, "#292e42")
	comment := paletteGet(pal, palette.ColorComment, "#565f89")

	var b strings.Builder
	fmt.Fprintf(&b, "# Delta theme generated from palette: %s\n", pal.Name)
	fmt.Fprintf(&b, "[delta]\n")
	fmt.Fprintf(&b, "    syntax-theme = ansi\n")
	fmt.Fprintf(&b, "    minus-style = syntax \"%s\"\n", blendColor(errorC, bg, 0.15))
	fmt.Fprintf(&b, "    minus-emph-style = syntax \"%s\"\n", blendColor(errorC, bg, 0.30))
	fmt.Fprintf(&b, "    plus-style = syntax \"%s\"\n", blendColor(success, bg, 0.15))
	fmt.Fprintf(&b, "    plus-emph-style = syntax \"%s\"\n", blendColor(success, bg, 0.30))
	fmt.Fprintf(&b, "    hunk-header-style = \"%s\" bold\n", info)
	fmt.Fprintf(&b, "    hunk-header-decoration-style = \"%s\" box\n", comment)
	fmt.Fprintf(&b, "    file-style = \"%s\" bold\n", info)
	fmt.Fprintf(&b, "    file-decoration-style = \"%s\" ul\n", info)
	fmt.Fprintf(&b, "    line-numbers = true\n")
	fmt.Fprintf(&b, "    line-numbers-minus-style = \"%s\"\n", errorC)
	fmt.Fprintf(&b, "    line-numbers-plus-style = \"%s\"\n", success)
	fmt.Fprintf(&b, "    line-numbers-zero-style = \"%s\"\n", comment)
	fmt.Fprintf(&b, "    line-numbers-left-format = \"{nm:>3} \"\n")
	fmt.Fprintf(&b, "    line-numbers-right-format = \"{np:>3} \"\n")
	fmt.Fprintf(&b, "    navigate = true\n")
	fmt.Fprintf(&b, "    side-by-side = false\n")
	fmt.Fprintf(&b, "    zero-style = syntax \"%s\"\n", bg)
	fmt.Fprintf(&b, "    blame-palette = \"%s\" \"%s\"\n", bg, highlight)
	_ = fg // used in comment for clarity

	return b.String(), nil
}

// blendColor creates a simple "overlay" effect by returning the overlay color.
// A proper implementation would alpha-blend, but for config generation we
// just return the overlay color at the given intensity tier.
func blendColor(overlay, base string, intensity float64) string {
	// Parse hex colors and blend
	or, og, ob := parseHex(overlay)
	br, bg, bb := parseHex(base)

	// Linear interpolation: result = base + (overlay - base) * intensity
	r := uint8(float64(br) + (float64(or)-float64(br))*intensity)
	g := uint8(float64(bg) + (float64(og)-float64(bg))*intensity)
	b := uint8(float64(bb) + (float64(ob)-float64(bb))*intensity)

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// parseHex parses a hex color string like "#7aa2f7" into RGB components.
func parseHex(hex string) (r, g, b uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
