package nvimops

import (
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
)

func TestManager(t *testing.T) {
	// Create manager with memory store
	mgr, err := NewWithOptions(Options{
		Store: store.NewMemoryStore(),
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Test Apply
	p := &plugin.Plugin{
		Name:        "test-plugin",
		Description: "A test plugin",
		Repo:        "test/plugin",
		Enabled:     true,
	}

	err = mgr.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Test Get
	retrieved, err := mgr.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != p.Name {
		t.Errorf("Name = %q, want %q", retrieved.Name, p.Name)
	}

	// Test List
	plugins, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("List count = %d, want 1", len(plugins))
	}

	// Test Delete
	err = mgr.Delete("test-plugin")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	plugins, _ = mgr.List()
	if len(plugins) != 0 {
		t.Errorf("List count after delete = %d, want 0", len(plugins))
	}
}

func TestManagerGenerateLua(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nvim-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager with memory store
	mgr, err := NewWithOptions(Options{
		Store: store.NewMemoryStore(),
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Add some plugins
	plugins := []*plugin.Plugin{
		{
			Name:    "plugin1",
			Repo:    "test/plugin1",
			Enabled: true,
		},
		{
			Name:    "plugin2",
			Repo:    "test/plugin2",
			Enabled: true,
			Config:  "require('plugin2').setup({})",
		},
		{
			Name:    "disabled",
			Repo:    "test/disabled",
			Enabled: false,
		},
	}

	for _, p := range plugins {
		_ = mgr.Apply(p)
	}

	// Generate Lua
	outputDir := filepath.Join(tmpDir, "lua")
	err = mgr.GenerateLua(outputDir)
	if err != nil {
		t.Fatalf("GenerateLua failed: %v", err)
	}

	// Check files were created
	files, _ := os.ReadDir(outputDir)
	if len(files) != 2 { // disabled plugin should not be generated
		t.Errorf("Expected 2 Lua files, got %d", len(files))
	}

	// Verify content
	content, err := os.ReadFile(filepath.Join(outputDir, "plugin2.lua"))
	if err != nil {
		t.Fatalf("Failed to read plugin2.lua: %v", err)
	}
	if len(content) == 0 {
		t.Error("plugin2.lua is empty")
	}
}

func TestManagerApplyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nvim-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test YAML file
	yamlContent := `apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test-from-file
  description: Loaded from file
spec:
  repo: file/plugin
`
	yamlPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test YAML: %v", err)
	}

	// Create manager
	mgr, err := NewWithOptions(Options{
		Store: store.NewMemoryStore(),
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Apply from file
	err = mgr.ApplyFile(yamlPath)
	if err != nil {
		t.Fatalf("ApplyFile failed: %v", err)
	}

	// Verify plugin was added
	p, err := mgr.Get("test-from-file")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Description != "Loaded from file" {
		t.Errorf("Description = %q, want %q", p.Description, "Loaded from file")
	}
}
