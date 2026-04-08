package toolconfig

import (
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroPalette"
)

// testPalette returns a palette with all standard color keys populated.
func testPalette() *palette.Palette {
	return &palette.Palette{
		Name:     "test-theme",
		Category: palette.CategoryDark,
		Colors: map[string]string{
			palette.ColorBg:          "#1a1b26",
			palette.ColorFg:          "#c0caf5",
			palette.ColorPrimary:     "#7aa2f7",
			palette.ColorSecondary:   "#bb9af7",
			palette.ColorAccent:      "#7dcfff",
			palette.ColorSuccess:     "#9ece6a",
			palette.ColorWarning:     "#e0af68",
			palette.ColorError:       "#f7768e",
			palette.ColorInfo:        "#7aa2f7",
			palette.ColorComment:     "#565f89",
			palette.ColorBgHighlight: "#292e42",
			palette.ColorBorder:      "#27a1b9",
		},
	}
}

func TestBatGenerator_ToolName(t *testing.T) {
	g := &BatGenerator{}
	if g.ToolName() != "bat" {
		t.Errorf("ToolName() = %q, want %q", g.ToolName(), "bat")
	}
}

func TestBatGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		pal     *palette.Palette
		wantErr bool
		checks  []string
	}{
		{
			name: "with palette",
			pal:  testPalette(),
			checks: []string{
				"--theme=ansi",
				"test-theme",
			},
		},
		{
			name:   "nil palette defaults to ansi",
			pal:    nil,
			checks: []string{"ansi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &BatGenerator{}
			got, err := g.Generate(tt.pal)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, check := range tt.checks {
				if !strings.Contains(got, check) {
					t.Errorf("Generate() output missing %q\ngot: %s", check, got)
				}
			}
		})
	}
}

func TestBatGenerator_Description(t *testing.T) {
	g := &BatGenerator{}
	if g.Description() == "" {
		t.Error("Description() should not be empty")
	}
}
