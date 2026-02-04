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

// Test isURL detection
func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"http://example.com/file.yaml", true},
		{"https://example.com/file.yaml", true},
		{"github:user/repo/path/file.yaml", true},
		{"./local/file.yaml", false},
		{"/absolute/path/file.yaml", false},
		{"file.yaml", false},
		{"", false},
		{"httpnotaurl", false},
		{"githubnotaurl", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isURL(tt.input)
			if got != tt.want {
				t.Errorf("isURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Test nvimops.FetchURL with direct URL (using httptest server)
func TestFetchURL_DirectURL(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	data, source, err := nvimops.FetchURL(server.URL + "/plugins/telescope.yaml")
	if err != nil {
		t.Fatalf("FetchURL failed: %v", err)
	}

	if source != server.URL+"/plugins/telescope.yaml" {
		t.Errorf("unexpected source: %s", source)
	}

	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// Test nvimops.FetchURL with invalid URL
func TestFetchURL_InvalidURL(t *testing.T) {
	_, _, err := nvimops.FetchURL("http://localhost:99999/nonexistent.yaml")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

// Test nvimops.FetchURL with 404 response
func TestFetchURL_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, _, err := nvimops.FetchURL(server.URL + "/notfound.yaml")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

// Test GitHub shorthand URL conversion (documents expected behavior)
func TestGitHubShorthandConversion(t *testing.T) {
	tests := []struct {
		input   string
		wantURL string
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
			// The conversion happens inside FetchURL
			// We can't test it directly without making HTTP requests
			// This test documents the expected behavior
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
