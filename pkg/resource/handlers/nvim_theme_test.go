package handlers

import (
	"testing"

	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/resource"
)

func TestNvimThemeHandler_Kind(t *testing.T) {
	h := NewNvimThemeHandler()
	if h.Kind() != KindNvimTheme {
		t.Errorf("Kind() = %v, want %v", h.Kind(), KindNvimTheme)
	}
}

func TestNvimThemeHandler_Apply(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: catppuccin-mocha
  description: Soothing pastel theme
  category: dark
spec:
  plugin:
    repo: catppuccin/nvim
  style: mocha
  colors:
    bg: "#1e1e2e"
    fg: "#cdd6f4"`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindNvimTheme {
		t.Errorf("Apply() resource.Kind = %v, want %v", res.GetKind(), KindNvimTheme)
	}
	if res.GetName() != "catppuccin-mocha" {
		t.Errorf("Apply() resource.Name = %v, want catppuccin-mocha", res.GetName())
	}

	// Verify it's in the store
	stored, err := memStore.Get("catppuccin-mocha")
	if err != nil {
		t.Errorf("Theme not found in store: %v", err)
	}
	if stored.Plugin.Repo != "catppuccin/nvim" {
		t.Errorf("Stored theme plugin.repo = %v, want catppuccin/nvim", stored.Plugin.Repo)
	}
}

func TestNvimThemeHandler_Get(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Pre-populate store
	memStore.Save(&theme.Theme{
		Name:   "tokyonight",
		Plugin: theme.ThemePlugin{Repo: "folke/tokyonight.nvim"},
	})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	res, err := h.Get(ctx, "tokyonight")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if res.GetName() != "tokyonight" {
		t.Errorf("Get() resource.Name = %v, want tokyonight", res.GetName())
	}

	// Cast to access underlying theme
	tr, ok := res.(*NvimThemeResource)
	if !ok {
		t.Fatalf("Get() result is not *NvimThemeResource")
	}
	if tr.Theme().Plugin.Repo != "folke/tokyonight.nvim" {
		t.Errorf("Get() theme.Plugin.Repo = %v, want folke/tokyonight.nvim", tr.Theme().Plugin.Repo)
	}
}

func TestNvimThemeHandler_List(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Pre-populate store
	memStore.Save(&theme.Theme{Name: "theme1", Plugin: theme.ThemePlugin{Repo: "user/theme1"}})
	memStore.Save(&theme.Theme{Name: "theme2", Plugin: theme.ThemePlugin{Repo: "user/theme2"}})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("List() returned %d resources, want 2", len(resources))
	}
}

func TestNvimThemeHandler_Delete(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Pre-populate store
	memStore.Save(&theme.Theme{Name: "to-delete", Plugin: theme.ThemePlugin{Repo: "user/to-delete"}})

	ctx := resource.Context{
		ThemeStore: memStore,
	}

	err := h.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = memStore.Get("to-delete")
	if err == nil {
		t.Error("Delete() did not remove theme from store")
	}
}

func TestNvimThemeHandler_ToYAML(t *testing.T) {
	h := NewNvimThemeHandler()

	res := &NvimThemeResource{
		theme: &theme.Theme{
			Name:        "test-theme",
			Description: "Test description",
			Plugin:      theme.ThemePlugin{Repo: "user/test-theme"},
		},
	}

	yaml, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Verify YAML contains expected content
	yamlStr := string(yaml)
	if !contains(yamlStr, "kind: NvimTheme") {
		t.Error("ToYAML() missing 'kind: NvimTheme'")
	}
	if !contains(yamlStr, "name: test-theme") {
		t.Error("ToYAML() missing 'name: test-theme'")
	}
}

func TestNvimThemeResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		theme   *theme.Theme
		wantErr bool
	}{
		{
			name:    "valid theme",
			theme:   &theme.Theme{Name: "test", Plugin: theme.ThemePlugin{Repo: "user/test"}},
			wantErr: false,
		},
		{
			name:    "missing name",
			theme:   &theme.Theme{Name: "", Plugin: theme.ThemePlugin{Repo: "user/test"}},
			wantErr: true,
		},
		{
			name:    "missing repo",
			theme:   &theme.Theme{Name: "test", Plugin: theme.ThemePlugin{Repo: ""}},
			wantErr: true,
		},
		{
			name: "invalid color",
			theme: &theme.Theme{
				Name:   "test",
				Plugin: theme.ThemePlugin{Repo: "user/test"},
				Colors: map[string]string{"bg": "not-a-color"},
			},
			wantErr: true,
		},
		{
			name: "valid colors",
			theme: &theme.Theme{
				Name:   "test",
				Plugin: theme.ThemePlugin{Repo: "user/test"},
				Colors: map[string]string{"bg": "#1e1e2e", "fg": "#cdd6f4"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &NvimThemeResource{theme: tt.theme}
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
