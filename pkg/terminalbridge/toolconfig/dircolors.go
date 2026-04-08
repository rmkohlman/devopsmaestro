package toolconfig

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroPalette"
)

// DircolorsGenerator generates LS_COLORS env var string from palette colors.
// Maps palette semantic colors to file type ANSI escape sequences.
type DircolorsGenerator struct{}

// ToolName returns "dircolors".
func (g *DircolorsGenerator) ToolName() string { return "dircolors" }

// Description returns a description of what this generator produces.
func (g *DircolorsGenerator) Description() string {
	return "LS_COLORS env var for colorized directory listings"
}

// Generate produces an LS_COLORS string using palette colors mapped to
// ANSI 256-color escape sequences for common file types.
//
// LS_COLORS format: "key=value:key=value:..."
// Values use ANSI SGR codes: 38;2;R;G;B for 24-bit foreground color.
func (g *DircolorsGenerator) Generate(pal *palette.Palette) (string, error) {
	if pal == nil {
		return "", fmt.Errorf("palette is required for dircolors config generation")
	}

	// Map palette colors to file type categories
	fg := paletteGet(pal, palette.ColorFg, "#c0caf5")
	info := paletteGet(pal, palette.ColorInfo, "#7aa2f7")
	success := paletteGet(pal, palette.ColorSuccess, "#9ece6a")
	warning := paletteGet(pal, palette.ColorWarning, "#e0af68")
	errorC := paletteGet(pal, palette.ColorError, "#f7768e")
	accent := paletteGet(pal, palette.ColorAccent, "#7aa2f7")
	comment := paletteGet(pal, palette.ColorComment, "#565f89")
	primary := paletteGet(pal, palette.ColorPrimary, "#7aa2f7")
	secondary := paletteGet(pal, palette.ColorSecondary, "#bb9af7")

	// Build LS_COLORS entries
	// Format: 38;2;R;G;B for truecolor foreground
	entries := []string{
		// Directories — accent/primary color, bold
		fmt.Sprintf("di=1;%s", hexToANSI(info)),
		// Symbolic links — cyan/accent
		fmt.Sprintf("ln=%s", hexToANSI(accent)),
		// Executables — green/success, bold
		fmt.Sprintf("ex=1;%s", hexToANSI(success)),
		// Regular files — default fg
		fmt.Sprintf("fi=%s", hexToANSI(fg)),
		// Pipes — warning color
		fmt.Sprintf("pi=%s", hexToANSI(warning)),
		// Sockets — secondary
		fmt.Sprintf("so=%s", hexToANSI(secondary)),
		// Block devices — warning, bold
		fmt.Sprintf("bd=1;%s", hexToANSI(warning)),
		// Char devices — warning
		fmt.Sprintf("cd=%s", hexToANSI(warning)),
		// Orphan symlinks — error
		fmt.Sprintf("or=%s", hexToANSI(errorC)),
		// Missing target — error, strikethrough
		fmt.Sprintf("mi=9;%s", hexToANSI(errorC)),
		// Setuid — error bg
		fmt.Sprintf("su=1;%s", hexToANSI(errorC)),
		// Archives — primary
		archiveExts(primary),
		// Images — secondary
		imageExts(secondary),
		// Source code — success
		sourceExts(success),
		// Docs — info
		docExts(info),
		// Config/dot files — comment
		configExts(comment),
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# LS_COLORS from palette: %s\n", pal.Name)
	b.WriteString(strings.Join(entries, ":"))

	return b.String(), nil
}

// hexToANSI converts a hex color like "#7aa2f7" to ANSI truecolor format "38;2;R;G;B".
func hexToANSI(hex string) string {
	r, g, b := parseHex(hex)
	return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
}

// extEntries generates LS_COLORS entries for file extensions.
func extEntries(color string, exts ...string) string {
	ansi := hexToANSI(color)
	parts := make([]string, len(exts))
	for i, ext := range exts {
		parts[i] = fmt.Sprintf("*%s=%s", ext, ansi)
	}
	return strings.Join(parts, ":")
}

func archiveExts(color string) string {
	return extEntries(color,
		".tar", ".gz", ".tgz", ".zip", ".7z", ".bz2", ".xz",
		".rar", ".zst", ".deb", ".rpm",
	)
}

func imageExts(color string) string {
	return extEntries(color,
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".bmp", ".ico", ".webp",
	)
}

func sourceExts(color string) string {
	return extEntries(color,
		".go", ".rs", ".py", ".js", ".ts", ".rb", ".java", ".c", ".h",
		".cpp", ".lua", ".sh", ".bash", ".zsh", ".fish",
	)
}

func docExts(color string) string {
	return extEntries(color, ".md", ".txt", ".pdf", ".doc", ".rst", ".org")
}

func configExts(color string) string {
	return extEntries(color,
		".yaml", ".yml", ".toml", ".json", ".xml", ".ini", ".conf", ".cfg",
	)
}
