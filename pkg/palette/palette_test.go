package palette

import (
	"testing"
)

func TestPalette_Get(t *testing.T) {
	p := &Palette{
		Name: "test",
		Colors: map[string]string{
			ColorBg: "#1a1b26",
			ColorFg: "#c0caf5",
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{ColorBg, "#1a1b26"},
		{ColorFg, "#c0caf5"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := p.Get(tt.key)
			if result != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestPalette_GetOrDefault(t *testing.T) {
	p := &Palette{
		Name: "test",
		Colors: map[string]string{
			ColorBg: "#1a1b26",
		},
	}

	tests := []struct {
		key          string
		defaultColor string
		expected     string
	}{
		{ColorBg, "#000000", "#1a1b26"},
		{"nonexistent", "#ffffff", "#ffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := p.GetOrDefault(tt.key, tt.defaultColor)
			if result != tt.expected {
				t.Errorf("GetOrDefault(%q, %q) = %q, want %q", tt.key, tt.defaultColor, result, tt.expected)
			}
		})
	}
}

func TestPalette_SetAndHas(t *testing.T) {
	p := &Palette{Name: "test"}

	if p.Has(ColorBg) {
		t.Error("expected Has(bg) to be false initially")
	}

	p.Set(ColorBg, "#1a1b26")

	if !p.Has(ColorBg) {
		t.Error("expected Has(bg) to be true after Set")
	}

	if got := p.Get(ColorBg); got != "#1a1b26" {
		t.Errorf("Get(bg) = %q, want %q", got, "#1a1b26")
	}
}

func TestPalette_Merge(t *testing.T) {
	base := &Palette{
		Name: "base",
		Colors: map[string]string{
			ColorBg: "#000000",
			ColorFg: "#ffffff",
		},
	}

	overlay := &Palette{
		Name: "overlay",
		Colors: map[string]string{
			ColorBg:    "#1a1b26",
			ColorError: "#ff0000",
		},
	}

	t.Run("without overwrite", func(t *testing.T) {
		p := base.Clone()
		p.Merge(overlay, false)

		if p.Get(ColorBg) != "#000000" {
			t.Error("expected bg to remain unchanged without overwrite")
		}
		if p.Get(ColorError) != "#ff0000" {
			t.Error("expected error color to be added")
		}
	})

	t.Run("with overwrite", func(t *testing.T) {
		p := base.Clone()
		p.Merge(overlay, true)

		if p.Get(ColorBg) != "#1a1b26" {
			t.Error("expected bg to be overwritten")
		}
	})
}

func TestPalette_Clone(t *testing.T) {
	original := &Palette{
		Name:        "original",
		Description: "test palette",
		Author:      "test",
		Category:    CategoryDark,
		Colors: map[string]string{
			ColorBg: "#1a1b26",
		},
	}

	clone := original.Clone()

	// Verify values are copied
	if clone.Name != original.Name {
		t.Error("Name not cloned correctly")
	}
	if clone.Get(ColorBg) != original.Get(ColorBg) {
		t.Error("Colors not cloned correctly")
	}

	// Verify independence
	clone.Set(ColorBg, "#000000")
	if original.Get(ColorBg) == clone.Get(ColorBg) {
		t.Error("Clone is not independent of original")
	}
}

func TestPalette_Validate(t *testing.T) {
	tests := []struct {
		name    string
		palette *Palette
		wantErr bool
	}{
		{
			name:    "valid palette",
			palette: &Palette{Name: "test", Colors: map[string]string{ColorBg: "#1a1b26"}},
			wantErr: false,
		},
		{
			name:    "missing name",
			palette: &Palette{Colors: map[string]string{ColorBg: "#1a1b26"}},
			wantErr: true,
		},
		{
			name:    "invalid color format",
			palette: &Palette{Name: "test", Colors: map[string]string{ColorBg: "not-a-color"}},
			wantErr: true,
		},
		{
			name:    "valid category",
			palette: &Palette{Name: "test", Category: CategoryDark},
			wantErr: false,
		},
		{
			name:    "invalid category",
			palette: &Palette{Name: "test", Category: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.palette.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#fff", true},
		{"#FFF", true},
		{"#ffffff", true},
		{"#FFFFFF", true},
		{"#1a1b26", true},
		{"#1a1b26ff", true}, // with alpha
		{"fff", false},
		{"#gg0000", false},
		{"#1234", false},
		{"rgb(0,0,0)", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := IsValidHexColor(tt.color)
			if result != tt.valid {
				t.Errorf("IsValidHexColor(%q) = %v, want %v", tt.color, result, tt.valid)
			}
		})
	}
}

func TestNormalizeHexColor(t *testing.T) {
	tests := []struct {
		color    string
		expected string
	}{
		{"#fff", "#ffffff"},
		{"#FFF", "#ffffff"},
		{"#1a1b26", "#1a1b26"},
		{"#1a1b26ff", "#1a1b26"}, // strips alpha
		{"invalid", "invalid"},   // returns unchanged
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := NormalizeHexColor(tt.color)
			if result != tt.expected {
				t.Errorf("NormalizeHexColor(%q) = %q, want %q", tt.color, result, tt.expected)
			}
		})
	}
}

func TestParseRGB(t *testing.T) {
	tests := []struct {
		color   string
		wantR   int
		wantG   int
		wantB   int
		wantErr bool
	}{
		{"#ffffff", 255, 255, 255, false},
		{"#000000", 0, 0, 0, false},
		{"#1a1b26", 26, 27, 38, false},
		{"#ff8040", 255, 128, 64, false},
		{"invalid", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			r, g, b, err := ParseRGB(tt.color)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRGB(%q) error = %v, wantErr %v", tt.color, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if r != tt.wantR || g != tt.wantG || b != tt.wantB {
					t.Errorf("ParseRGB(%q) = (%d, %d, %d), want (%d, %d, %d)",
						tt.color, r, g, b, tt.wantR, tt.wantG, tt.wantB)
				}
			}
		})
	}
}

func TestToANSI256(t *testing.T) {
	tests := []struct {
		color    string
		expected int
		wantErr  bool
	}{
		{"#000000", 16, false},  // black
		{"#ffffff", 231, false}, // white
		{"#808080", 244, false}, // gray (grayscale range)
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result, err := ToANSI256(tt.color)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToANSI256(%q) error = %v, wantErr %v", tt.color, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ToANSI256(%q) = %d, want %d", tt.color, result, tt.expected)
			}
		})
	}
}

func TestAllColorKeys(t *testing.T) {
	keys := AllColorKeys()
	if len(keys) == 0 {
		t.Error("AllColorKeys() returned empty slice")
	}

	// Check that required colors are in the list
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, required := range RequiredColors {
		if !keyMap[required] {
			t.Errorf("Required color %q not in AllColorKeys()", required)
		}
	}
}
