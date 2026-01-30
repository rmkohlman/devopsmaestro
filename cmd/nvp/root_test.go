package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/pkg/nvimops"
	"devopsmaestro/pkg/nvimops/store"
)

// Test fetchURL with GitHub shorthand conversion
func TestFetchURL_GitHubShorthand(t *testing.T) {
	// Create a test server that mimics GitHub raw content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the path matches expected GitHub raw format
		expectedPath := "/rmkohlman/nvim-yaml-plugins/main/plugins/telescope.yaml"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: Test plugin
spec:
  repo: "nvim-telescope/telescope.nvim"
`))
	}))
	defer server.Close()

	// We can't easily test the GitHub shorthand without mocking the HTTP client
	// So we'll test direct URL fetching
	data, source, err := fetchURL(server.URL + "/rmkohlman/nvim-yaml-plugins/main/plugins/telescope.yaml")
	if err != nil {
		t.Fatalf("fetchURL failed: %v", err)
	}

	if source != server.URL+"/rmkohlman/nvim-yaml-plugins/main/plugins/telescope.yaml" {
		t.Errorf("unexpected source: %s", source)
	}

	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// Test fetchURL with invalid URL
func TestFetchURL_InvalidURL(t *testing.T) {
	_, _, err := fetchURL("http://localhost:99999/nonexistent.yaml")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

// Test fetchURL with 404 response
func TestFetchURL_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, _, err := fetchURL(server.URL + "/notfound.yaml")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

// Test GitHub shorthand URL conversion
func TestGitHubShorthandConversion(t *testing.T) {
	tests := []struct {
		input   string
		wantURL string
		wantErr bool
	}{
		{
			input:   "github:user/repo/path/file.yaml",
			wantURL: "https://raw.githubusercontent.com/user/repo/main/path/file.yaml",
		},
		{
			input:   "github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml",
			wantURL: "https://raw.githubusercontent.com/rmkohlman/nvim-yaml-plugins/main/plugins/telescope.yaml",
		},
		{
			input:   "github:a/b/c.yaml",
			wantURL: "https://raw.githubusercontent.com/a/b/main/c.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// We need to extract the URL conversion logic to test it
			// For now, we'll test the full fetchURL which will fail on network
			// but we can at least verify it doesn't panic

			// The conversion happens inside fetchURL, so we can't test it directly
			// without refactoring. This test documents the expected behavior.
			_ = tt.wantURL
		})
	}
}

// Test applyPluginData with valid YAML
func TestApplyPluginData(t *testing.T) {
	// Create temp directory for test store
	tmpDir, err := os.MkdirTemp("", "nvp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pluginsDir := filepath.Join(tmpDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	fileStore, err := store.NewFileStore(pluginsDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	mgr, err := nvimops.NewWithOptions(nvimops.Options{
		Store: fileStore,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Close()

	validYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test-plugin
  description: A test plugin
  category: test
spec:
  repo: "test/test-plugin"
`)

	err = applyPluginData(mgr, validYAML, "test-source")
	if err != nil {
		t.Fatalf("applyPluginData failed: %v", err)
	}

	// Verify plugin was created
	p, err := mgr.Get("test-plugin")
	if err != nil {
		t.Fatalf("failed to get plugin: %v", err)
	}

	if p.Name != "test-plugin" {
		t.Errorf("unexpected plugin name: %s", p.Name)
	}

	if p.Description != "A test plugin" {
		t.Errorf("unexpected description: %s", p.Description)
	}
}

// Test applyPluginData with invalid YAML
func TestApplyPluginData_InvalidYAML(t *testing.T) {
	// Create temp directory for test store
	tmpDir, err := os.MkdirTemp("", "nvp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pluginsDir := filepath.Join(tmpDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	fileStore, err := store.NewFileStore(pluginsDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	mgr, err := nvimops.NewWithOptions(nvimops.Options{
		Store: fileStore,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Close()

	invalidYAML := []byte(`not: valid: yaml: content`)

	err = applyPluginData(mgr, invalidYAML, "test-source")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// Test applyPluginData updates existing plugin
func TestApplyPluginData_Update(t *testing.T) {
	// Create temp directory for test store
	tmpDir, err := os.MkdirTemp("", "nvp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pluginsDir := filepath.Join(tmpDir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins dir: %v", err)
	}

	fileStore, err := store.NewFileStore(pluginsDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	mgr, err := nvimops.NewWithOptions(nvimops.Options{
		Store: fileStore,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Close()

	// Create initial plugin
	initialYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test-plugin
  description: Initial description
spec:
  repo: "test/test-plugin"
`)

	err = applyPluginData(mgr, initialYAML, "test-source")
	if err != nil {
		t.Fatalf("applyPluginData (create) failed: %v", err)
	}

	// Update plugin
	updatedYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test-plugin
  description: Updated description
spec:
  repo: "test/test-plugin"
`)

	err = applyPluginData(mgr, updatedYAML, "test-source")
	if err != nil {
		t.Fatalf("applyPluginData (update) failed: %v", err)
	}

	// Verify plugin was updated
	p, err := mgr.Get("test-plugin")
	if err != nil {
		t.Fatalf("failed to get plugin: %v", err)
	}

	if p.Description != "Updated description" {
		t.Errorf("plugin was not updated: got %s, want 'Updated description'", p.Description)
	}
}

// Test getConfigDir with environment variable
func TestGetConfigDir(t *testing.T) {
	// Save original value
	originalDir := configDir
	originalEnv := os.Getenv("NVP_CONFIG_DIR")
	defer func() {
		configDir = originalDir
		os.Setenv("NVP_CONFIG_DIR", originalEnv)
	}()

	// Test with flag
	configDir = "/custom/path"
	if dir := getConfigDir(); dir != "/custom/path" {
		t.Errorf("expected /custom/path, got %s", dir)
	}

	// Test with environment variable
	configDir = ""
	os.Setenv("NVP_CONFIG_DIR", "/env/path")
	if dir := getConfigDir(); dir != "/env/path" {
		t.Errorf("expected /env/path, got %s", dir)
	}

	// Test default
	configDir = ""
	os.Unsetenv("NVP_CONFIG_DIR")
	dir := getConfigDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".nvp")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
