package toolconfig

import (
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroPalette"
)

func TestFzfGenerator_ToolName(t *testing.T) {
	g := &FzfGenerator{}
	if g.ToolName() != "fzf" {
		t.Errorf("ToolName() = %q, want %q", g.ToolName(), "fzf")
	}
}

func TestFzfGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		pal     *palette.Palette
		wantErr bool
		checks  []string
	}{
		{
			name: "produces valid fzf color string",
			pal:  testPalette(),
			checks: []string{
				"--color=",
				"fg:#c0caf5",
				"bg:#1a1b26",
				"hl:#7dcfff",
				"fg+:#c0caf5",
				"bg+:#292e42",
				"info:#7aa2f7",
				"prompt:#7aa2f7",
				"pointer:#7dcfff",
				"marker:#e0af68",
				"spinner:#7aa2f7",
				"header:#565f89",
				"border:#27a1b9",
			},
		},
		{
			name:    "nil palette returns error",
			pal:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &FzfGenerator{}
			got, err := g.Generate(tt.pal)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			for _, check := range tt.checks {
				if !strings.Contains(got, check) {
					t.Errorf("output missing %q\ngot:\n%s", check, got)
				}
			}
		})
	}
}

func TestFzfGenerator_DefaultColors(t *testing.T) {
	// Palette with missing colors should use defaults
	pal := &palette.Palette{
		Name:     "minimal",
		Category: palette.CategoryDark,
		Colors:   map[string]string{},
	}

	g := &FzfGenerator{}
	got, err := g.Generate(pal)
	if err != nil {
		t.Fatal(err)
	}

	// Should still produce valid output with defaults
	if !strings.Contains(got, "--color=") {
		t.Error("expected --color= prefix in output")
	}
	if !strings.Contains(got, "fg:") {
		t.Error("expected fg color slot in output")
	}
}
