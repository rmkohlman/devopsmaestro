// Package store provides storage abstractions for nvim-manager.
// This allows plugins to be stored in different backends (files, database, memory).
package store

import (
	"devopsmaestro/pkg/nvimops/plugin"
)

// PluginStore defines the interface for plugin storage operations.
// Implementations can store plugins in files, databases, or memory.
type PluginStore interface {
	// Create adds a new plugin to the store.
	// Returns an error if a plugin with the same name already exists.
	Create(p *plugin.Plugin) error

	// Update modifies an existing plugin in the store.
	// Returns an error if the plugin doesn't exist.
	Update(p *plugin.Plugin) error

	// Upsert creates or updates a plugin (create if not exists, update if exists).
	Upsert(p *plugin.Plugin) error

	// Delete removes a plugin from the store by name.
	// Returns an error if the plugin doesn't exist.
	Delete(name string) error

	// Get retrieves a plugin by name.
	// Returns nil and an error if the plugin doesn't exist.
	Get(name string) (*plugin.Plugin, error)

	// List returns all plugins in the store.
	List() ([]*plugin.Plugin, error)

	// ListByCategory returns plugins in a specific category.
	ListByCategory(category string) ([]*plugin.Plugin, error)

	// ListByTag returns plugins that have a specific tag.
	ListByTag(tag string) ([]*plugin.Plugin, error)

	// Exists checks if a plugin with the given name exists.
	Exists(name string) (bool, error)

	// Close releases any resources held by the store.
	Close() error
}

// ErrNotFound is returned when a plugin is not found.
type ErrNotFound struct {
	Name string
}

func (e *ErrNotFound) Error() string {
	return "plugin not found: " + e.Name
}

// ErrAlreadyExists is returned when trying to create a plugin that already exists.
type ErrAlreadyExists struct {
	Name string
}

func (e *ErrAlreadyExists) Error() string {
	return "plugin already exists: " + e.Name
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	_, ok := err.(*ErrNotFound)
	return ok
}

// IsAlreadyExists returns true if the error is an already exists error.
func IsAlreadyExists(err error) bool {
	_, ok := err.(*ErrAlreadyExists)
	return ok
}
