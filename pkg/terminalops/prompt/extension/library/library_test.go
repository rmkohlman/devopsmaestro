package library

import (
	"testing"
)

func TestExtensionLibrary_Load(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	// Should have all 9 extensions
	expectedCount := 9
	if lib.Count() != expectedCount {
		t.Errorf("Count() = %d, want %d", lib.Count(), expectedCount)
	}

	// Check expected extension names
	expectedNames := []string{
		"character",
		"colima",
		"conda",
		"git-detailed",
		"identity",
		"kubernetes",
		"languages-all",
		"path",
		"time",
	}

	names := lib.Names()
	if len(names) != len(expectedNames) {
		t.Errorf("Names() count = %d, want %d", len(names), len(expectedNames))
	}

	// Verify each expected name exists
	for _, expected := range expectedNames {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected extension %q not found in Names()", expected)
		}
	}
}

func TestExtensionLibrary_Get(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	tests := []struct {
		name       string
		extName    string
		wantErr    bool
		wantSeg    string
		wantModule string
	}{
		{
			name:       "get identity extension",
			extName:    "identity",
			wantErr:    false,
			wantSeg:    "identity",
			wantModule: "os",
		},
		{
			name:       "get git-detailed extension",
			extName:    "git-detailed",
			wantErr:    false,
			wantSeg:    "git",
			wantModule: "git_branch",
		},
		{
			name:       "get kubernetes extension",
			extName:    "kubernetes",
			wantErr:    false,
			wantSeg:    "kubernetes",
			wantModule: "kubernetes",
		},
		{
			name:    "get nonexistent extension",
			extName: "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := lib.Get(tt.extName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if ext.Segment != tt.wantSeg {
				t.Errorf("Segment = %q, want %q", ext.Segment, tt.wantSeg)
			}

			if _, ok := ext.Modules[tt.wantModule]; !ok {
				t.Errorf("Module %q not found in extension", tt.wantModule)
			}
		})
	}
}

func TestExtensionLibrary_ListBySegment(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	// Test git segment
	gitExts := lib.ListBySegment("git")
	if len(gitExts) != 1 {
		t.Errorf("ListBySegment(git) count = %d, want 1", len(gitExts))
	}
	if len(gitExts) > 0 && gitExts[0].Name != "git-detailed" {
		t.Errorf("ListBySegment(git)[0].Name = %q, want %q", gitExts[0].Name, "git-detailed")
	}

	// Test identity segment
	identityExts := lib.ListBySegment("identity")
	if len(identityExts) != 1 {
		t.Errorf("ListBySegment(identity) count = %d, want 1", len(identityExts))
	}
}

func TestExtensionLibrary_ListByCategory(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	// Test development category
	devExts := lib.ListByCategory("development")
	if len(devExts) != 2 { // languages-all, conda
		t.Errorf("ListByCategory(development) count = %d, want 2", len(devExts))
	}

	// Test system category
	sysExts := lib.ListByCategory("system")
	if len(sysExts) != 3 { // identity, time, character
		t.Errorf("ListByCategory(system) count = %d, want 3", len(sysExts))
	}
}

func TestExtensionLibrary_Categories(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	categories := lib.Categories()
	expectedCategories := []string{"development", "infrastructure", "navigation", "system", "vcs"}

	if len(categories) != len(expectedCategories) {
		t.Errorf("Categories() count = %d, want %d", len(categories), len(expectedCategories))
	}

	for i, cat := range categories {
		if cat != expectedCategories[i] {
			t.Errorf("Categories()[%d] = %q, want %q", i, cat, expectedCategories[i])
		}
	}
}

func TestExtensionLibrary_Segments(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	segments := lib.Segments()
	expectedSegments := []string{"character", "colima", "conda", "git", "identity", "kubernetes", "languages", "path", "time"}

	if len(segments) != len(expectedSegments) {
		t.Errorf("Segments() count = %d, want %d", len(segments), len(expectedSegments))
	}
}

func TestExtensionLibrary_ExtensionModules(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	// Test languages-all has all expected modules
	langExt, err := lib.Get("languages-all")
	if err != nil {
		t.Fatalf("Get(languages-all) error = %v", err)
	}

	expectedModules := []string{"c", "rust", "golang", "nodejs", "php", "java", "kotlin", "haskell", "python", "custom.db"}
	for _, mod := range expectedModules {
		if _, ok := langExt.Modules[mod]; !ok {
			t.Errorf("Module %q not found in languages-all extension", mod)
		}
	}

	// Test that provides matches modules
	if len(langExt.Provides) != len(langExt.Modules) {
		t.Errorf("Provides count = %d, Modules count = %d, should match", len(langExt.Provides), len(langExt.Modules))
	}
}

func TestExtensionLibrary_NoHexColors(t *testing.T) {
	lib, err := NewExtensionLibrary()
	if err != nil {
		t.Fatalf("NewExtensionLibrary() error = %v", err)
	}

	// All extensions should use palette keys, not hex colors
	for _, ext := range lib.List() {
		for modName, mod := range ext.Modules {
			if mod.Style != "" {
				// Check for hex color pattern
				if containsHexColor(mod.Style) {
					t.Errorf("Extension %q module %q contains hex color in style: %q", ext.Name, modName, mod.Style)
				}
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
