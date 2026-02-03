package palette

import (
	"testing"
)

func TestPalette_ToTerminalColors(t *testing.T) {
	// Create a palette similar to tokyonight theme
	p := &Palette{
		Name:     "tokyonight",
		Category: CategoryDark,
		Colors: map[string]string{
			ColorBg:          "#1a1b26",
			ColorBgDark:      "#16161e",
			ColorBgHighlight: "#292e42",
			ColorBgVisual:    "#283457",
			ColorFg:          "#c0caf5",
			ColorFgDark:      "#a9b1d6",
			ColorFgGutter:    "#3b4261",
			ColorComment:     "#565f89",
			ColorBorder:      "#27a1b9",
			// Standard colors
			"blue":    "#7aa2f7",
			"cyan":    "#7dcfff",
			"green":   "#9ece6a",
			"magenta": "#bb9af7",
			"orange":  "#ff9e64",
			"purple":  "#9d7cd8",
			"red":     "#f7768e",
			"teal":    "#1abc9c",
			"yellow":  "#e0af68",
		},
	}

	terminal := p.ToTerminalColors()

	// Verify key terminal colors are extracted
	tests := []struct {
		key      string
		expected string
	}{
		{TermBlack, "#16161e"},   // from bg_dark
		{TermRed, "#f7768e"},     // from red
		{TermGreen, "#9ece6a"},   // from green
		{TermYellow, "#e0af68"},  // from yellow
		{TermBlue, "#7aa2f7"},    // from blue
		{TermMagenta, "#bb9af7"}, // from magenta
		{TermCyan, "#7dcfff"},    // from cyan
		{TermWhite, "#c0caf5"},   // from fg

		{TermBrightBlack, "#565f89"}, // from comment
		{ColorBg, "#1a1b26"},
		{ColorFg, "#c0caf5"},
		{TermSelection, "#283457"}, // from bg_visual
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := terminal[tt.key]; got != tt.expected {
				t.Errorf("ToTerminalColors()[%q] = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestPalette_TerminalPalette(t *testing.T) {
	p := &Palette{
		Name:        "test-theme",
		Description: "Test description",
		Author:      "Test Author",
		Category:    CategoryDark,
		Colors: map[string]string{
			ColorBg: "#1a1b26",
			ColorFg: "#c0caf5",
			"red":   "#ff0000",
		},
	}

	terminal := p.TerminalPalette()

	if terminal.Name != "test-theme-terminal" {
		t.Errorf("TerminalPalette().Name = %q, want %q", terminal.Name, "test-theme-terminal")
	}

	if terminal.Description != p.Description {
		t.Error("TerminalPalette() should preserve Description")
	}

	if terminal.Author != p.Author {
		t.Error("TerminalPalette() should preserve Author")
	}

	if terminal.Category != p.Category {
		t.Error("TerminalPalette() should preserve Category")
	}

	// Verify it has terminal colors
	if terminal.Get(TermRed) == "" {
		t.Error("TerminalPalette() should have terminal colors extracted")
	}
}

func TestPalette_ToTerminalColors_Fallbacks(t *testing.T) {
	// Test with minimal palette - should use fallbacks
	p := &Palette{
		Name: "minimal",
		Colors: map[string]string{
			ColorBg:    "#000000",
			ColorFg:    "#ffffff",
			ColorError: "#ff0000",
		},
	}

	terminal := p.ToTerminalColors()

	// Should fallback to ColorError for red
	if got := terminal[TermRed]; got != "#ff0000" {
		t.Errorf("Expected TermRed to fallback to ColorError, got %q", got)
	}

	// Should fallback to bg for black
	if got := terminal[TermBlack]; got != "#000000" {
		t.Errorf("Expected TermBlack to fallback to bg, got %q", got)
	}
}

func TestPalette_ToTerminalColors_NilSafe(t *testing.T) {
	var p *Palette
	if p.ToTerminalColors() != nil {
		t.Error("ToTerminalColors() on nil palette should return nil")
	}

	p = &Palette{Name: "empty"}
	if colors := p.ToTerminalColors(); colors == nil {
		t.Error("ToTerminalColors() on empty palette should return empty map, not nil")
	}
}
