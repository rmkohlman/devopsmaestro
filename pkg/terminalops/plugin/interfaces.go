// Package plugin provides types and utilities for shell plugin management.
package plugin

// PluginStore defines the interface for plugin storage operations.
type PluginStore interface {
	// Create adds a new plugin to the store.
	Create(p *Plugin) error

	// Update modifies an existing plugin in the store.
	Update(p *Plugin) error

	// Upsert creates or updates a plugin.
	Upsert(p *Plugin) error

	// Delete removes a plugin from the store by name.
	Delete(name string) error

	// Get retrieves a plugin by name.
	Get(name string) (*Plugin, error)

	// List returns all plugins in the store.
	List() ([]*Plugin, error)

	// ListByManager returns plugins using a specific plugin manager.
	ListByManager(manager PluginManager) ([]*Plugin, error)

	// ListByCategory returns plugins in a specific category.
	ListByCategory(category string) ([]*Plugin, error)

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
	return "terminal plugin not found: " + e.Name
}

// ErrAlreadyExists is returned when trying to create a plugin that already exists.
type ErrAlreadyExists struct {
	Name string
}

func (e *ErrAlreadyExists) Error() string {
	return "terminal plugin already exists: " + e.Name
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
