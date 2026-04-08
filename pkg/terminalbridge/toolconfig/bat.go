package toolconfig

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroPalette"
)

// BatGenerator generates configuration for the bat command-line tool.
// Since bat uses pre-built themes, we generate a BAT_THEME env var
// that selects the closest matching built-in theme, or "ansi" to use
// terminal ANSI colors (which inherit from the palette via terminal config).
type BatGenerator struct{}

// ToolName returns "bat".
func (g *BatGenerator) ToolName() string { return "bat" }

// Description returns a description of what this generator produces.
func (g *BatGenerator) Description() string {
	return "BAT_THEME env var for syntax-highlighted cat"
}

// Generate produces a BAT_THEME value. Since bat doesn't support arbitrary
// hex colors in its theme env var, we use "ansi" which inherits terminal
// ANSI colors from the palette's terminal color scheme.
//
// For light themes we use "ansi" as well — bat's ansi theme adapts to
// the terminal's color scheme.
func (g *BatGenerator) Generate(pal *palette.Palette) (string, error) {
	if pal == nil {
		return "ansi", nil
	}

	// Build a bat config snippet that sets the theme and paging behavior.
	// Using "ansi" ensures bat inherits colors from the terminal's ANSI palette,
	// which is set by the workspace theme.
	var b strings.Builder
	fmt.Fprintf(&b, "# bat configuration for theme: %s\n", pal.Name)
	fmt.Fprintf(&b, "# Uses ANSI colors from terminal palette\n")
	fmt.Fprintf(&b, "--theme=ansi\n")

	return b.String(), nil
}
