package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"devopsmaestro/pkg/nvimops/plugin"

	"gopkg.in/yaml.v3"
)

// FileStore stores plugins as YAML files in a directory.
// Each plugin is stored in a separate file named <plugin-name>.yaml.
type FileStore struct {
	mu      sync.RWMutex
	baseDir string
	cache   map[string]*plugin.Plugin
	loaded  bool
}

// NewFileStore creates a new file-based plugin store.
// The baseDir will be created if it doesn't exist.
func NewFileStore(baseDir string) (*FileStore, error) {
	// Expand home directory
	if strings.HasPrefix(baseDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, baseDir[1:])
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &FileStore{
		baseDir: baseDir,
		cache:   make(map[string]*plugin.Plugin),
	}, nil
}

// DefaultFileStore creates a FileStore in the default location (~/.nvim-manager/plugins).
func DefaultFileStore() (*FileStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return NewFileStore(filepath.Join(home, ".nvim-manager", "plugins"))
}

// Create adds a new plugin to the store.
func (s *FileStore) Create(p *plugin.Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoaded(); err != nil {
		return err
	}

	if _, exists := s.cache[p.Name]; exists {
		return &ErrAlreadyExists{Name: p.Name}
	}

	return s.writePlugin(p)
}

// Update modifies an existing plugin in the store.
func (s *FileStore) Update(p *plugin.Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoaded(); err != nil {
		return err
	}

	if _, exists := s.cache[p.Name]; !exists {
		return &ErrNotFound{Name: p.Name}
	}

	return s.writePlugin(p)
}

// Upsert creates or updates a plugin.
func (s *FileStore) Upsert(p *plugin.Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoaded(); err != nil {
		return err
	}

	return s.writePlugin(p)
}

// Delete removes a plugin from the store by name.
func (s *FileStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoaded(); err != nil {
		return err
	}

	if _, exists := s.cache[name]; !exists {
		return &ErrNotFound{Name: name}
	}

	path := s.pluginPath(name)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete plugin file: %w", err)
	}

	delete(s.cache, name)
	return nil
}

// Get retrieves a plugin by name.
func (s *FileStore) Get(name string) (*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	p, exists := s.cache[name]
	if !exists {
		return nil, &ErrNotFound{Name: name}
	}

	return p, nil
}

// List returns all plugins in the store.
func (s *FileStore) List() ([]*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	result := make([]*plugin.Plugin, 0, len(s.cache))
	for _, p := range s.cache {
		result = append(result, p)
	}
	return result, nil
}

// ListByCategory returns plugins in a specific category.
func (s *FileStore) ListByCategory(category string) ([]*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	var result []*plugin.Plugin
	for _, p := range s.cache {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result, nil
}

// ListByTag returns plugins that have a specific tag.
func (s *FileStore) ListByTag(tag string) ([]*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	var result []*plugin.Plugin
	for _, p := range s.cache {
		for _, t := range p.Tags {
			if t == tag {
				result = append(result, p)
				break
			}
		}
	}
	return result, nil
}

// Exists checks if a plugin with the given name exists.
func (s *FileStore) Exists(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureLoaded(); err != nil {
		return false, err
	}

	_, exists := s.cache[name]
	return exists, nil
}

// Close is a no-op for file store (files are written immediately).
func (s *FileStore) Close() error {
	return nil
}

// Reload forces a reload of all plugins from disk.
func (s *FileStore) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.loaded = false
	s.cache = make(map[string]*plugin.Plugin)
	return s.loadPlugins()
}

// BaseDir returns the base directory of the store.
func (s *FileStore) BaseDir() string {
	return s.baseDir
}

// ensureLoaded loads plugins if not already loaded.
// Must be called with lock held.
func (s *FileStore) ensureLoaded() error {
	if s.loaded {
		return nil
	}
	return s.loadPlugins()
}

// loadPlugins loads all YAML files from the base directory.
// Must be called with lock held.
func (s *FileStore) loadPlugins() error {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read store directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		path := filepath.Join(s.baseDir, entry.Name())
		p, err := plugin.ParseYAMLFile(path)
		if err != nil {
			// Log warning but continue loading other plugins
			fmt.Fprintf(os.Stderr, "warning: failed to load plugin %s: %v\n", entry.Name(), err)
			continue
		}

		s.cache[p.Name] = p
	}

	s.loaded = true
	return nil
}

// writePlugin writes a plugin to disk and updates the cache.
// Must be called with lock held.
func (s *FileStore) writePlugin(p *plugin.Plugin) error {
	// Set timestamps
	now := time.Now()
	if p.CreatedAt == nil {
		p.CreatedAt = &now
	}
	p.UpdatedAt = &now

	// Convert to YAML format
	py := p.ToYAML()

	// Marshal to YAML
	data, err := yaml.Marshal(py)
	if err != nil {
		return fmt.Errorf("failed to marshal plugin: %w", err)
	}

	// Write to file
	path := s.pluginPath(p.Name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	// Update cache
	s.cache[p.Name] = p
	return nil
}

// pluginPath returns the file path for a plugin.
func (s *FileStore) pluginPath(name string) string {
	// Sanitize name for filesystem
	safeName := strings.ReplaceAll(name, "/", "-")
	safeName = strings.ReplaceAll(safeName, "\\", "-")
	return filepath.Join(s.baseDir, safeName+".yaml")
}

// Verify FileStore implements PluginStore
var _ PluginStore = (*FileStore)(nil)
