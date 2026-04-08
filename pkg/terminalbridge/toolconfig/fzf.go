package toolconfig

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroPalette"
)

// FzfGenerator generates FZF_DEFAULT_OPTS --color string from palette colors.
// Maps semantic palette colors to fzf's color slot names.
type FzfGenerator struct{}

// ToolName returns "fzf".
func (g *FzfGenerator) ToolName() string { return "fzf" }

// Description returns a description of what this generator produces.
func (g *FzfGenerator) Description() string {
	return "FZF_DEFAULT_OPTS --color string for fuzzy finder"
}

// Generate produces an fzf --color= string mapping palette colors to fzf slots.
//
// fzf color slots:
//   - fg, bg: default foreground/background
//   - hl: highlighted substrings
//   - fg+, bg+: current line foreground/background
//   - hl+: highlighted substrings on current line
//   - info: info line
//   - prompt: prompt
//   - pointer: pointer to current line
//   - marker: multi-select marker
//   - spinner: streaming input indicator
//   - header: header
//   - border: border
//   - gutter: gutter on the left
func (g *FzfGenerator) Generate(pal *palette.Palette) (string, error) {
	if pal == nil {
		return "", fmt.Errorf("palette is required for fzf config generation")
	}

	// Map palette to fzf color slots
	fg := paletteGet(pal, palette.ColorFg, "#c0caf5")
	bg := paletteGet(pal, palette.ColorBg, "#1a1b26")
	accent := paletteGet(pal, palette.ColorAccent, "#7aa2f7")
	highlight := paletteGet(pal, palette.ColorBgHighlight, "#292e42")
	info := paletteGet(pal, palette.ColorInfo, "#7aa2f7")
	success := paletteGet(pal, palette.ColorSuccess, "#9ece6a")
	warning := paletteGet(pal, palette.ColorWarning, "#e0af68")
	comment := paletteGet(pal, palette.ColorComment, "#565f89")
	border := paletteGet(pal, palette.ColorBorder, "#27a1b9")
	primary := paletteGet(pal, palette.ColorPrimary, "#7aa2f7")

	// Build fzf --color string
	pairs := []string{
		fmt.Sprintf("fg:%s", fg),
		fmt.Sprintf("bg:%s", bg),
		fmt.Sprintf("hl:%s", accent),
		fmt.Sprintf("fg+:%s", fg),
		fmt.Sprintf("bg+:%s", highlight),
		fmt.Sprintf("hl+:%s", success),
		fmt.Sprintf("info:%s", info),
		fmt.Sprintf("prompt:%s", primary),
		fmt.Sprintf("pointer:%s", accent),
		fmt.Sprintf("marker:%s", warning),
		fmt.Sprintf("spinner:%s", info),
		fmt.Sprintf("header:%s", comment),
		fmt.Sprintf("border:%s", border),
		fmt.Sprintf("gutter:%s", bg),
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# fzf colors from palette: %s\n", pal.Name)
	fmt.Fprintf(&b, "--color=%s", strings.Join(pairs, ","))

	return b.String(), nil
}
