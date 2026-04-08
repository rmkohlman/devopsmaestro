package toolconfig

import (
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroPalette"
)

func TestDeltaGenerator_ToolName(t *testing.T) {
	g := &DeltaGenerator{}
	if g.ToolName() != "delta" {
		t.Errorf("ToolName() = %q, want %q", g.ToolName(), "delta")
	}
}

func TestDeltaGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		pal     *palette.Palette
		wantErr bool
		checks  []string
	}{
		{
			name: "produces valid gitconfig delta section",
			pal:  testPalette(),
			checks: []string{
				"[delta]",
				"minus-style",
				"plus-style",
				"hunk-header-style",
				"line-numbers = true",
				"file-style",
				"navigate = true",
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
			g := &DeltaGenerator{}
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

func TestDeltaGenerator_ColorsAreHex(t *testing.T) {
	g := &DeltaGenerator{}
	got, err := g.Generate(testPalette())
	if err != nil {
		t.Fatal(err)
	}
	// All color references should be hex format
	if !strings.Contains(got, "#") {
		t.Error("expected hex color references in output")
	}
}

func TestBlendColor(t *testing.T) {
	tests := []struct {
		name      string
		overlay   string
		base      string
		intensity float64
		want      string
	}{
		{
			name:      "zero intensity returns base",
			overlay:   "#ff0000",
			base:      "#000000",
			intensity: 0.0,
			want:      "#000000",
		},
		{
			name:      "full intensity returns overlay",
			overlay:   "#ff0000",
			base:      "#000000",
			intensity: 1.0,
			want:      "#ff0000",
		},
		{
			name:      "half intensity blends",
			overlay:   "#ff0000",
			base:      "#000000",
			intensity: 0.5,
			// 0 + (255-0)*0.5 = 127.5 → 127
			want: "#7f0000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blendColor(tt.overlay, tt.base, tt.intensity)
			if got != tt.want {
				t.Errorf("blendColor(%q, %q, %f) = %q, want %q",
					tt.overlay, tt.base, tt.intensity, got, tt.want)
			}
		})
	}
}

func TestParseHex(t *testing.T) {
	r, g, b := parseHex("#7aa2f7")
	if r != 0x7a || g != 0xa2 || b != 0xf7 {
		t.Errorf("parseHex(#7aa2f7) = (%d,%d,%d), want (122,162,247)", r, g, b)
	}

	// Without hash prefix
	r, g, b = parseHex("ff0000")
	if r != 0xff || g != 0x00 || b != 0x00 {
		t.Errorf("parseHex(ff0000) = (%d,%d,%d), want (255,0,0)", r, g, b)
	}
}
