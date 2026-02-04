package handlers

import (
	"testing"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/resource"
)

func TestNvimPluginHandler_Kind(t *testing.T) {
	h := NewNvimPluginHandler()
	if h.Kind() != KindNvimPlugin {
		t.Errorf("Kind() = %v, want %v", h.Kind(), KindNvimPlugin)
	}
}

func TestNvimPluginHandler_Apply(t *testing.T) {
	h := NewNvimPluginHandler()
	memStore := store.NewMemoryStore()

	ctx := resource.Context{
		PluginStore: memStore,
	}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: Fuzzy finder
spec:
  repo: nvim-telescope/telescope.nvim
  lazy: true`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindNvimPlugin {
		t.Errorf("Apply() resource.Kind = %v, want %v", res.GetKind(), KindNvimPlugin)
	}
	if res.GetName() != "telescope" {
		t.Errorf("Apply() resource.Name = %v, want telescope", res.GetName())
	}

	// Verify it's in the store
	stored, err := memStore.Get("telescope")
	if err != nil {
		t.Errorf("Plugin not found in store: %v", err)
	}
	if stored.Repo != "nvim-telescope/telescope.nvim" {
		t.Errorf("Stored plugin repo = %v, want nvim-telescope/telescope.nvim", stored.Repo)
	}
}

func TestNvimPluginHandler_Get(t *testing.T) {
	h := NewNvimPluginHandler()
	memStore := store.NewMemoryStore()

	// Pre-populate store
	memStore.Create(&plugin.Plugin{
		Name:    "treesitter",
		Repo:    "nvim-treesitter/nvim-treesitter",
		Enabled: true,
	})

	ctx := resource.Context{
		PluginStore: memStore,
	}

	res, err := h.Get(ctx, "treesitter")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if res.GetName() != "treesitter" {
		t.Errorf("Get() resource.Name = %v, want treesitter", res.GetName())
	}

	// Cast to access underlying plugin
	pr, ok := res.(*NvimPluginResource)
	if !ok {
		t.Fatalf("Get() result is not *NvimPluginResource")
	}
	if pr.Plugin().Repo != "nvim-treesitter/nvim-treesitter" {
		t.Errorf("Get() plugin.Repo = %v, want nvim-treesitter/nvim-treesitter", pr.Plugin().Repo)
	}
}

func TestNvimPluginHandler_List(t *testing.T) {
	h := NewNvimPluginHandler()
	memStore := store.NewMemoryStore()

	// Pre-populate store
	memStore.Create(&plugin.Plugin{Name: "plugin1", Repo: "user/plugin1", Enabled: true})
	memStore.Create(&plugin.Plugin{Name: "plugin2", Repo: "user/plugin2", Enabled: true})

	ctx := resource.Context{
		PluginStore: memStore,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("List() returned %d resources, want 2", len(resources))
	}
}

func TestNvimPluginHandler_Delete(t *testing.T) {
	h := NewNvimPluginHandler()
	memStore := store.NewMemoryStore()

	// Pre-populate store
	memStore.Create(&plugin.Plugin{Name: "to-delete", Repo: "user/to-delete", Enabled: true})

	ctx := resource.Context{
		PluginStore: memStore,
	}

	err := h.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = memStore.Get("to-delete")
	if err == nil {
		t.Error("Delete() did not remove plugin from store")
	}
}

func TestNvimPluginHandler_ToYAML(t *testing.T) {
	h := NewNvimPluginHandler()

	res := &NvimPluginResource{
		plugin: &plugin.Plugin{
			Name:        "test-plugin",
			Description: "Test description",
			Repo:        "user/test-plugin",
			Enabled:     true,
		},
	}

	yaml, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Verify YAML contains expected content
	yamlStr := string(yaml)
	if !contains(yamlStr, "kind: NvimPlugin") {
		t.Error("ToYAML() missing 'kind: NvimPlugin'")
	}
	if !contains(yamlStr, "name: test-plugin") {
		t.Error("ToYAML() missing 'name: test-plugin'")
	}
}

func TestNvimPluginResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		plugin  *plugin.Plugin
		wantErr bool
	}{
		{
			name:    "valid plugin",
			plugin:  &plugin.Plugin{Name: "test", Repo: "user/test"},
			wantErr: false,
		},
		{
			name:    "missing name",
			plugin:  &plugin.Plugin{Name: "", Repo: "user/test"},
			wantErr: true,
		},
		{
			name:    "missing repo",
			plugin:  &plugin.Plugin{Name: "test", Repo: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &NvimPluginResource{plugin: tt.plugin}
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
