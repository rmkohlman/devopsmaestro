package store

import (
	"devopsmaestro/pkg/nvimops/plugin"
	"sync"
)

// MemoryStore is an in-memory implementation of PluginStore.
// Useful for testing and temporary storage.
type MemoryStore struct {
	mu      sync.RWMutex
	plugins map[string]*plugin.Plugin
}

// NewMemoryStore creates a new in-memory plugin store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		plugins: make(map[string]*plugin.Plugin),
	}
}

// Create adds a new plugin to the store.
func (s *MemoryStore) Create(p *plugin.Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.plugins[p.Name]; exists {
		return &ErrAlreadyExists{Name: p.Name}
	}

	// Make a copy to avoid external mutations
	s.plugins[p.Name] = copyPlugin(p)
	return nil
}

// Update modifies an existing plugin in the store.
func (s *MemoryStore) Update(p *plugin.Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.plugins[p.Name]; !exists {
		return &ErrNotFound{Name: p.Name}
	}

	s.plugins[p.Name] = copyPlugin(p)
	return nil
}

// Upsert creates or updates a plugin.
func (s *MemoryStore) Upsert(p *plugin.Plugin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.plugins[p.Name] = copyPlugin(p)
	return nil
}

// Delete removes a plugin from the store by name.
func (s *MemoryStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.plugins[name]; !exists {
		return &ErrNotFound{Name: name}
	}

	delete(s.plugins, name)
	return nil
}

// Get retrieves a plugin by name.
func (s *MemoryStore) Get(name string) (*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.plugins[name]
	if !exists {
		return nil, &ErrNotFound{Name: name}
	}

	return copyPlugin(p), nil
}

// List returns all plugins in the store.
func (s *MemoryStore) List() ([]*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*plugin.Plugin, 0, len(s.plugins))
	for _, p := range s.plugins {
		result = append(result, copyPlugin(p))
	}
	return result, nil
}

// ListByCategory returns plugins in a specific category.
func (s *MemoryStore) ListByCategory(category string) ([]*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*plugin.Plugin
	for _, p := range s.plugins {
		if p.Category == category {
			result = append(result, copyPlugin(p))
		}
	}
	return result, nil
}

// ListByTag returns plugins that have a specific tag.
func (s *MemoryStore) ListByTag(tag string) ([]*plugin.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*plugin.Plugin
	for _, p := range s.plugins {
		for _, t := range p.Tags {
			if t == tag {
				result = append(result, copyPlugin(p))
				break
			}
		}
	}
	return result, nil
}

// Exists checks if a plugin with the given name exists.
func (s *MemoryStore) Exists(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.plugins[name]
	return exists, nil
}

// Close is a no-op for memory store.
func (s *MemoryStore) Close() error {
	return nil
}

// copyPlugin creates a shallow copy of a plugin.
func copyPlugin(p *plugin.Plugin) *plugin.Plugin {
	if p == nil {
		return nil
	}
	copy := *p
	// Copy slices
	if p.Event != nil {
		copy.Event = make([]string, len(p.Event))
		copy.Event = append(copy.Event[:0], p.Event...)
	}
	if p.Ft != nil {
		copy.Ft = make([]string, len(p.Ft))
		copy.Ft = append(copy.Ft[:0], p.Ft...)
	}
	if p.Cmd != nil {
		copy.Cmd = make([]string, len(p.Cmd))
		copy.Cmd = append(copy.Cmd[:0], p.Cmd...)
	}
	if p.Tags != nil {
		copy.Tags = make([]string, len(p.Tags))
		copy.Tags = append(copy.Tags[:0], p.Tags...)
	}
	if p.Dependencies != nil {
		copy.Dependencies = make([]plugin.Dependency, len(p.Dependencies))
		copy.Dependencies = append(copy.Dependencies[:0], p.Dependencies...)
	}
	if p.Keys != nil {
		copy.Keys = make([]plugin.Keymap, len(p.Keys))
		copy.Keys = append(copy.Keys[:0], p.Keys...)
	}
	if p.Keymaps != nil {
		copy.Keymaps = make([]plugin.Keymap, len(p.Keymaps))
		copy.Keymaps = append(copy.Keymaps[:0], p.Keymaps...)
	}
	// Note: Opts map is shallow copied (same reference)
	return &copy
}

// Verify MemoryStore implements PluginStore
var _ PluginStore = (*MemoryStore)(nil)
