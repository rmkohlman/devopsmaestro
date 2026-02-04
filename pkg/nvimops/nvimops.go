// Package nvimops provides tools for managing Neovim configurations.
//
// This package is designed to be:
// - Standalone: Can be used independently to manage nvim config via YAML
// - Importable: Can be imported as a library by dvm for container integration
// - Portable: Enables sharing nvim setups as portable YAML files
//
// # Architecture
//
// The package is organized into sub-packages:
//   - plugin: Plugin types, YAML parsing, and Lua code generation
//   - store: Storage abstractions (file, memory, database adapters)
//   - config: Neovim configuration initialization and management
//   - library: Pre-built plugin definitions for common plugins
//
// # Basic Usage
//
//	import "devopsmaestro/pkg/nvimops"
//
//	// Create a manager with default file storage
//	mgr, err := nvimops.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Apply a plugin from YAML
//	err = mgr.ApplyFile("telescope.yaml")
//
//	// Generate Lua files for all plugins
//	err = mgr.GenerateLua("~/.config/nvim/lua/plugins/custom")
package nvimops

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
)

// Manager provides high-level operations for nvim-manager.
type Manager struct {
	store     store.PluginStore
	generator plugin.LuaGenerator
}

// Options configures the Manager.
type Options struct {
	// Store is the plugin store to use. If nil, creates a default FileStore.
	Store store.PluginStore

	// StoreDir is the directory for the file store. Ignored if Store is provided.
	// Defaults to ~/.nvim-manager/plugins
	StoreDir string

	// Generator is the Lua generator to use. If nil, creates a default Generator.
	Generator plugin.LuaGenerator
}

// New creates a new Manager with default options.
func New() (*Manager, error) {
	return NewWithOptions(Options{})
}

// NewWithOptions creates a new Manager with the specified options.
func NewWithOptions(opts Options) (*Manager, error) {
	var s store.PluginStore
	var err error

	if opts.Store != nil {
		s = opts.Store
	} else if opts.StoreDir != "" {
		s, err = store.NewFileStore(opts.StoreDir)
	} else {
		s, err = store.DefaultFileStore()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	// Use provided generator or create default
	gen := opts.Generator
	if gen == nil {
		gen = plugin.NewGenerator()
	}

	return &Manager{
		store:     s,
		generator: gen,
	}, nil
}

// ApplyFile applies a plugin from a YAML file.
func (m *Manager) ApplyFile(path string) error {
	p, err := plugin.ParseYAMLFile(path)
	if err != nil {
		return fmt.Errorf("failed to parse plugin file: %w", err)
	}
	return m.Apply(p)
}

// ApplyURL fetches a plugin YAML from a URL and applies it.
// Supports GitHub shorthand: github:user/repo/path/file.yaml
func (m *Manager) ApplyURL(url string) error {
	data, fetchedURL, err := FetchURL(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}

	p, err := plugin.ParseYAML(data)
	if err != nil {
		return fmt.Errorf("failed to parse plugin YAML from %s: %w", fetchedURL, err)
	}

	return m.Apply(p)
}

// FetchURL fetches content from a URL, supporting GitHub shorthand.
// GitHub shorthand format: github:user/repo/path/file.yaml
// Returns the data, the resolved URL, and any error.
func FetchURL(url string) ([]byte, string, error) {
	originalURL := url

	// Handle GitHub shorthand: github:user/repo/path/file.yaml
	if strings.HasPrefix(url, "github:") {
		path := strings.TrimPrefix(url, "github:")
		// Split into user/repo and the rest
		parts := strings.SplitN(path, "/", 3)
		if len(parts) < 3 {
			return nil, "", fmt.Errorf("invalid github URL format, expected github:user/repo/path/file.yaml")
		}
		user := parts[0]
		repo := parts[1]
		filePath := parts[2]
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", user, repo, filePath)
		slog.Debug("converted GitHub shorthand", "original", originalURL, "url", url)
	}

	slog.Debug("fetching URL", "url", url)

	// Fetch the URL
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		slog.Error("HTTP request failed", "url", url, "error", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("HTTP request returned error", "url", url, "status", resp.StatusCode)
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body", "url", url, "error", err)
		return nil, "", err
	}

	slog.Info("fetched URL successfully", "url", url, "bytes", len(data))
	return data, url, nil
}

// Apply applies (upserts) a plugin to the store.
func (m *Manager) Apply(p *plugin.Plugin) error {
	return m.store.Upsert(p)
}

// Get retrieves a plugin by name.
func (m *Manager) Get(name string) (*plugin.Plugin, error) {
	return m.store.Get(name)
}

// List returns all plugins.
func (m *Manager) List() ([]*plugin.Plugin, error) {
	return m.store.List()
}

// Delete removes a plugin by name.
func (m *Manager) Delete(name string) error {
	return m.store.Delete(name)
}

// GenerateLua generates Lua files for all enabled plugins in the output directory.
func (m *Manager) GenerateLua(outputDir string) error {
	plugins, err := m.store.List()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	// Expand home directory
	if outputDir[0] == '~' {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, outputDir[1:])
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, p := range plugins {
		if !p.Enabled {
			continue
		}

		lua, err := m.generator.GenerateLuaFile(p)
		if err != nil {
			return fmt.Errorf("failed to generate Lua for %s: %w", p.Name, err)
		}

		filename := filepath.Join(outputDir, p.Name+".lua")
		if err := os.WriteFile(filename, []byte(lua), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}

// GenerateLuaFor generates Lua code for a specific plugin without writing to disk.
func (m *Manager) GenerateLuaFor(name string) (string, error) {
	p, err := m.store.Get(name)
	if err != nil {
		return "", err
	}
	return m.generator.GenerateLuaFile(p)
}

// Store returns the underlying plugin store.
// Useful for advanced operations or using a different store implementation.
func (m *Manager) Store() store.PluginStore {
	return m.store
}

// Generator returns the underlying Lua generator.
// Useful for advanced operations or customizing Lua output.
func (m *Manager) Generator() plugin.LuaGenerator {
	return m.generator
}

// Close releases resources held by the manager.
func (m *Manager) Close() error {
	return m.store.Close()
}
