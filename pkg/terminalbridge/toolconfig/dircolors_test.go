package toolconfig

import (
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroPalette"
)

func TestDircolorsGenerator_ToolName(t *testing.T) {
	g := &DircolorsGenerator{}
	if g.ToolName() != "dircolors" {
		t.Errorf("ToolName() = %q, want %q", g.ToolName(), "dircolors")
	}
}

func TestDircolorsGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		pal     *palette.Palette
		wantErr bool
		checks  []string
	}{
		{
			name: "produces valid LS_COLORS entries",
			pal:  testPalette(),
			checks: []string{
				"di=",   // directories
				"ln=",   // symlinks
				"ex=",   // executables
				"fi=",   // regular files
				"38;2;", // truecolor ANSI codes
				"*.go=",
				"*.py=",
				"*.md=",
				"*.yaml=",
				"*.tar=",
				"*.png=",
				"test-theme",
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
			g := &DircolorsGenerator{}
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

func TestDircolorsGenerator_FormatIsColonSeparated(t *testing.T) {
	g := &DircolorsGenerator{}
	got, err := g.Generate(testPalette())
	if err != nil {
		t.Fatal(err)
	}

	// Split into lines, get the actual LS_COLORS value (line after comment)
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatal("expected at least 2 lines (comment + value)")
	}

	// The value line should be colon-separated
	valueLine := lines[1]
	parts := strings.Split(valueLine, ":")
	if len(parts) < 5 {
		t.Errorf("expected many colon-separated entries, got %d parts", len(parts))
	}
}

func TestHexToANSI(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"#ff0000", "38;2;255;0;0"},
		{"#00ff00", "38;2;0;255;0"},
		{"#0000ff", "38;2;0;0;255"},
		{"#7aa2f7", "38;2;122;162;247"},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			got := hexToANSI(tt.hex)
			if got != tt.want {
				t.Errorf("hexToANSI(%q) = %q, want %q", tt.hex, got, tt.want)
			}
		})
	}
}
