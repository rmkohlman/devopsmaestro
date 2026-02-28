package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAML(t *testing.T) {
	yaml := `
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: test-theme
  description: A test theme
  author: test
  category: dark
spec:
  plugin:
    repo: "folke/tokyonight.nvim"
  style: night
  transparent: false
  colors:
    bg: "#1a1b26"
    fg: "#c0caf5"
`
	theme, err := ParseYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if theme.Name != "test-theme" {
		t.Errorf("expected name 'test-theme', got '%s'", theme.Name)
	}
	if theme.Plugin.Repo != "folke/tokyonight.nvim" {
		t.Errorf("expected repo 'folke/tokyonight.nvim', got '%s'", theme.Plugin.Repo)
	}
	if theme.Style != "night" {
		t.Errorf("expected style 'night', got '%s'", theme.Style)
	}
	if theme.Colors["bg"] != "#1a1b26" {
		t.Errorf("expected bg '#1a1b26', got '%s'", theme.Colors["bg"])
	}
}

func TestParseYAML_InvalidKind(t *testing.T) {
	yaml := `
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test
spec:
  plugin:
    repo: "test/test"
`
	_, err := ParseYAML([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
}

func TestParseYAML_MissingName(t *testing.T) {
	yaml := `
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  description: test
spec:
  plugin:
    repo: "test/test"
`
	_, err := ParseYAML([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestTheme_Validate(t *testing.T) {
	tests := []struct {
		name    string
		theme   Theme
		wantErr bool
	}{
		{
			name: "valid theme",
			theme: Theme{
				Name:   "test",
				Plugin: ThemePlugin{Repo: "user/repo"},
				Colors: map[string]string{"bg": "#1a1b26"},
			},
			wantErr: false,
		},
		{
			name: "valid standalone theme",
			theme: Theme{
				Name:   "test-standalone",
				Plugin: ThemePlugin{Repo: ""}, // Empty repo = standalone
				Colors: map[string]string{"bg": "#1a1b26", "fg": "#ffffff"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			theme: Theme{
				Plugin: ThemePlugin{Repo: "user/repo"},
			},
			wantErr: true,
		},
		{
			name: "standalone theme without colors",
			theme: Theme{
				Name:   "test",
				Plugin: ThemePlugin{Repo: ""}, // Empty repo = standalone
				// No colors defined - should fail for standalone
			},
			wantErr: true,
		},
		{
			name: "invalid color format",
			theme: Theme{
				Name:   "test",
				Plugin: ThemePlugin{Repo: "user/repo"},
				Colors: map[string]string{"bg": "invalid"},
			},
			wantErr: true,
		},
		{
			name: "valid 3-char hex",
			theme: Theme{
				Name:   "test",
				Plugin: ThemePlugin{Repo: "user/repo"},
				Colors: map[string]string{"bg": "#abc"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.theme.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTheme_ToYAML(t *testing.T) {
	theme := Theme{
		Name:        "test-theme",
		Description: "Test description",
		Plugin:      ThemePlugin{Repo: "user/repo"},
		Style:       "dark",
		Colors:      map[string]string{"bg": "#000000"},
	}

	data, err := theme.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	// Parse it back
	parsed, err := ParseYAML(data)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if parsed.Name != theme.Name {
		t.Errorf("round-trip name mismatch: got '%s', want '%s'", parsed.Name, theme.Name)
	}
}

func TestGetSetupName(t *testing.T) {
	tests := []struct {
		repo     string
		expected string
	}{
		{"folke/tokyonight.nvim", "tokyonight"},
		{"catppuccin/nvim", "catppuccin"},
		{"ellisonleao/gruvbox.nvim", "gruvbox"},
		{"unknown/theme", ""},
	}

	for _, tt := range tests {
		got := GetSetupName(tt.repo)
		if got != tt.expected {
			t.Errorf("GetSetupName(%s) = %s, want %s", tt.repo, got, tt.expected)
		}
	}
}

func TestFileStore(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "theme-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFileStore(tmpDir)

	// Test Init
	if err := store.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify themes dir created
	if _, err := os.Stat(filepath.Join(tmpDir, "themes")); os.IsNotExist(err) {
		t.Error("themes directory not created")
	}

	// Test Save
	theme := &Theme{
		Name:   "test-theme",
		Plugin: ThemePlugin{Repo: "user/repo"},
		Colors: map[string]string{"bg": "#000000"},
	}

	if err := store.Save(theme); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	retrieved, err := store.Get("test-theme")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != theme.Name {
		t.Errorf("Get returned wrong theme: got '%s', want '%s'", retrieved.Name, theme.Name)
	}

	// Test List
	themes, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(themes) != 1 {
		t.Errorf("List returned %d themes, want 1", len(themes))
	}

	// Test SetActive
	if err := store.SetActive("test-theme"); err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	// Test GetActive
	active, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active.Name != "test-theme" {
		t.Errorf("GetActive returned wrong theme: got '%s', want 'test-theme'", active.Name)
	}

	// Test Delete
	if err := store.Delete("test-theme"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	if _, err := store.Get("test-theme"); err == nil {
		t.Error("Get should fail after delete")
	}
}

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	theme := &Theme{
		Name:   "test-theme",
		Plugin: ThemePlugin{Repo: "user/repo"},
		Colors: map[string]string{"bg": "#000000"},
	}

	// Test Save
	if err := store.Save(theme); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	retrieved, err := store.Get("test-theme")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != theme.Name {
		t.Errorf("Get returned wrong theme")
	}

	// Test SetActive
	if err := store.SetActive("test-theme"); err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	// Test GetActive
	active, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active.Name != "test-theme" {
		t.Errorf("GetActive returned wrong theme")
	}

	// Test Delete
	if err := store.Delete("test-theme"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Active should be cleared
	active, err = store.GetActive()
	if err != nil {
		t.Fatalf("GetActive after delete failed: %v", err)
	}
	if active != nil {
		t.Error("Active should be nil after deleting active theme")
	}
}
