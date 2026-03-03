package library

import (
	"testing"
)

func TestStyleLibrary_Load(t *testing.T) {
	lib, err := NewStyleLibrary()
	if err != nil {
		t.Fatalf("NewStyleLibrary() error = %v", err)
	}

	// Should have 2 styles
	expectedCount := 2
	if lib.Count() != expectedCount {
		t.Errorf("Count() = %d, want %d", lib.Count(), expectedCount)
	}

	// Check expected style names
	expectedNames := []string{
		"powerline-minimal",
		"powerline-segments",
	}

	names := lib.Names()
	if len(names) != len(expectedNames) {
		t.Errorf("Names() count = %d, want %d", len(names), len(expectedNames))
	}

	for i, name := range names {
		if name != expectedNames[i] {
			t.Errorf("Names()[%d] = %q, want %q", i, name, expectedNames[i])
		}
	}
}

func TestStyleLibrary_Get(t *testing.T) {
	lib, err := NewStyleLibrary()
	if err != nil {
		t.Fatalf("NewStyleLibrary() error = %v", err)
	}

	tests := []struct {
		name        string
		styleName   string
		wantErr     bool
		wantSegment string
	}{
		{
			name:        "get powerline-segments style",
			styleName:   "powerline-segments",
			wantErr:     false,
			wantSegment: "identity",
		},
		{
			name:        "get powerline-minimal style",
			styleName:   "powerline-minimal",
			wantErr:     false,
			wantSegment: "identity",
		},
		{
			name:      "get nonexistent style",
			styleName: "nonexistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := lib.Get(tt.styleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Check that segments are present
			segments := s.GetSegments()
			if len(segments) == 0 {
				t.Error("Style should have segments")
			}

			// Check first segment name
			if segments[0].Name != tt.wantSegment {
				t.Errorf("First segment name = %q, want %q", segments[0].Name, tt.wantSegment)
			}
		})
	}
}

func TestStyleLibrary_PowerlineSegmentsStructure(t *testing.T) {
	lib, err := NewStyleLibrary()
	if err != nil {
		t.Fatalf("NewStyleLibrary() error = %v", err)
	}

	s, err := lib.Get("powerline-segments")
	if err != nil {
		t.Fatalf("Get(powerline-segments) error = %v", err)
	}

	// Should have 9 segments in order
	segments := s.GetSegments()
	expectedSegments := []string{
		"identity",
		"kubernetes",
		"colima",
		"path",
		"git",
		"languages",
		"conda",
		"time",
		"character",
	}

	if len(segments) != len(expectedSegments) {
		t.Errorf("Segments count = %d, want %d", len(segments), len(expectedSegments))
	}

	for i, seg := range segments {
		if i >= len(expectedSegments) {
			break
		}
		if seg.Name != expectedSegments[i] {
			t.Errorf("Segment[%d].Name = %q, want %q", i, seg.Name, expectedSegments[i])
		}
	}

	// Check that identity has startTransition
	if segments[0].StartTransition == "" {
		t.Error("identity segment should have startTransition")
	}

	// Check that path has correct colors
	pathSeg := segments[3]
	if pathSeg.StartColor != "peach" {
		t.Errorf("path.StartColor = %q, want %q", pathSeg.StartColor, "peach")
	}
	if pathSeg.EndColor != "peach" {
		t.Errorf("path.EndColor = %q, want %q", pathSeg.EndColor, "peach")
	}
}

func TestStyleLibrary_Categories(t *testing.T) {
	lib, err := NewStyleLibrary()
	if err != nil {
		t.Fatalf("NewStyleLibrary() error = %v", err)
	}

	categories := lib.Categories()
	if len(categories) != 1 {
		t.Errorf("Categories() count = %d, want 1", len(categories))
	}

	if len(categories) > 0 && categories[0] != "powerline" {
		t.Errorf("Categories()[0] = %q, want %q", categories[0], "powerline")
	}
}

func TestStyleLibrary_ListByCategory(t *testing.T) {
	lib, err := NewStyleLibrary()
	if err != nil {
		t.Fatalf("NewStyleLibrary() error = %v", err)
	}

	powerlineStyles := lib.ListByCategory("powerline")
	if len(powerlineStyles) != 2 {
		t.Errorf("ListByCategory(powerline) count = %d, want 2", len(powerlineStyles))
	}
}

func TestStyleLibrary_NoHexColors(t *testing.T) {
	lib, err := NewStyleLibrary()
	if err != nil {
		t.Fatalf("NewStyleLibrary() error = %v", err)
	}

	// All styles should use palette keys, not hex colors
	for _, s := range lib.List() {
		for _, seg := range s.Segments {
			if containsHexColor(seg.StartColor) {
				t.Errorf("Style %q segment %q has hex color in StartColor: %q", s.Name, seg.Name, seg.StartColor)
			}
			if containsHexColor(seg.EndColor) {
				t.Errorf("Style %q segment %q has hex color in EndColor: %q", s.Name, seg.Name, seg.EndColor)
			}
		}
	}
}

func containsHexColor(s string) bool {
	// Simple check for #XXX or #XXXXXX patterns
	for i := 0; i < len(s); i++ {
		if s[i] == '#' && i+3 < len(s) {
			return true
		}
	}
	return false
}
