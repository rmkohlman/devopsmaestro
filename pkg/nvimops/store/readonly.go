package store

import (
	"devopsmaestro/pkg/nvimops/plugin"
)

// ReadOnlyStore wraps a read-only plugin source (like Library) to implement PluginStore.
// Write operations (Create, Update, Upsert, Delete) return an error.
type ReadOnlyStore struct {
	source ReadOnlySource
}

// ReadOnlySource defines the minimal interface for a read-only plugin source.
// This is implemented by Library.
type ReadOnlySource interface {
	Get(name string) (*plugin.Plugin, bool)
	List() []*plugin.Plugin
	ListByCategory(category string) []*plugin.Plugin
	ListByTag(tag string) []*plugin.Plugin
}

// ErrReadOnly is returned when a write operation is attempted on a read-only store.
type ErrReadOnly struct {
	Operation string
}

func (e *ErrReadOnly) Error() string {
	return "operation not permitted on read-only store: " + e.Operation
}

// IsReadOnly returns true if the error is a read-only error.
func IsReadOnly(err error) bool {
	_, ok := err.(*ErrReadOnly)
	return ok
}

// NewReadOnlyStore creates a PluginStore wrapper around a read-only source.
func NewReadOnlyStore(source ReadOnlySource) *ReadOnlyStore {
	return &ReadOnlyStore{source: source}
}

// Create returns an error (read-only).
func (s *ReadOnlyStore) Create(p *plugin.Plugin) error {
	return &ErrReadOnly{Operation: "create"}
}

// Update returns an error (read-only).
func (s *ReadOnlyStore) Update(p *plugin.Plugin) error {
	return &ErrReadOnly{Operation: "update"}
}

// Upsert returns an error (read-only).
func (s *ReadOnlyStore) Upsert(p *plugin.Plugin) error {
	return &ErrReadOnly{Operation: "upsert"}
}

// Delete returns an error (read-only).
func (s *ReadOnlyStore) Delete(name string) error {
	return &ErrReadOnly{Operation: "delete"}
}

// Get retrieves a plugin by name.
func (s *ReadOnlyStore) Get(name string) (*plugin.Plugin, error) {
	p, ok := s.source.Get(name)
	if !ok {
		return nil, &ErrNotFound{Name: name}
	}
	return p, nil
}

// List returns all plugins.
func (s *ReadOnlyStore) List() ([]*plugin.Plugin, error) {
	return s.source.List(), nil
}

// ListByCategory returns plugins in a specific category.
func (s *ReadOnlyStore) ListByCategory(category string) ([]*plugin.Plugin, error) {
	return s.source.ListByCategory(category), nil
}

// ListByTag returns plugins that have a specific tag.
func (s *ReadOnlyStore) ListByTag(tag string) ([]*plugin.Plugin, error) {
	return s.source.ListByTag(tag), nil
}

// Exists checks if a plugin with the given name exists.
func (s *ReadOnlyStore) Exists(name string) (bool, error) {
	_, ok := s.source.Get(name)
	return ok, nil
}

// Close is a no-op for read-only store.
func (s *ReadOnlyStore) Close() error {
	return nil
}

// Verify ReadOnlyStore implements PluginStore
var _ PluginStore = (*ReadOnlyStore)(nil)
