// Package nvimbridge provides database adapters that bridge MaestroNvim types
// with dvm's database layer (models, db packages).
// This file implements a database adapter that bridges the NvimPluginStore
// interface with the db.DataStore interface, enabling nvimops to use SQLite storage.
//
// The bridge defines its own interfaces and error types so it depends on
// MaestroNvim only for data types (plugin.Plugin), not for store contracts.
// This allows the nvim module's store layer to be swapped or updated without
// breaking the bridge.
package nvimbridge

import (
	"fmt"
	"strings"

	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
)

// ---------------------------------------------------------------------------
// Bridge-local interfaces and error types
// ---------------------------------------------------------------------------

// NvimPluginStore defines the bridge-level interface for plugin storage
// operations. It mirrors the method set of store.PluginStore from MaestroNvim
// so that any type implementing NvimPluginStore also structurally satisfies
// store.PluginStore, but the bridge does not import that package.
type NvimPluginStore interface {
	Create(p *plugin.Plugin) error
	Update(p *plugin.Plugin) error
	Upsert(p *plugin.Plugin) error
	Delete(name string) error
	Get(name string) (*plugin.Plugin, error)
	List() ([]*plugin.Plugin, error)
	ListByCategory(category string) ([]*plugin.Plugin, error)
	ListByTag(tag string) ([]*plugin.Plugin, error)
	Exists(name string) (bool, error)
	Close() error
}

// ErrPluginNotFound is returned when a requested plugin does not exist.
type ErrPluginNotFound struct {
	Name string
}

func (e *ErrPluginNotFound) Error() string {
	return "plugin not found: " + e.Name
}

// ErrPluginAlreadyExists is returned when attempting to create a plugin
// that already exists.
type ErrPluginAlreadyExists struct {
	Name string
}

func (e *ErrPluginAlreadyExists) Error() string {
	return "plugin already exists: " + e.Name
}

// IsPluginNotFound reports whether err is an ErrPluginNotFound.
func IsPluginNotFound(err error) bool {
	_, ok := err.(*ErrPluginNotFound)
	return ok
}

// IsPluginAlreadyExists reports whether err is an ErrPluginAlreadyExists.
func IsPluginAlreadyExists(err error) bool {
	_, ok := err.(*ErrPluginAlreadyExists)
	return ok
}

// PluginDataStore is a subset of db.DataStore containing only plugin operations.
// This interface allows the adapter to work with any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type PluginDataStore interface {
	// CreatePlugin inserts a new nvim plugin.
	CreatePlugin(plugin *models.NvimPluginDB) error

	// GetPluginByName retrieves a plugin by its name.
	GetPluginByName(name string) (*models.NvimPluginDB, error)

	// UpdatePlugin updates an existing plugin.
	UpdatePlugin(plugin *models.NvimPluginDB) error

	// DeletePlugin removes a plugin by name.
	DeletePlugin(name string) error

	// ListPlugins retrieves all plugins.
	ListPlugins() ([]*models.NvimPluginDB, error)

	// ListPluginsByCategory retrieves plugins filtered by category.
	ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error)

	// ListPluginsByTags retrieves plugins that have any of the specified tags.
	ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error)
}

// PluginDBStoreAdapter adapts db.DataStore to implement NvimPluginStore.
// Because NvimPluginStore mirrors the upstream store.PluginStore method set,
// the adapter also structurally satisfies store.PluginStore without importing it.
// This enables nvimops.Manager to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type PluginDBStoreAdapter struct {
	store     PluginDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewPluginDBStoreAdapter creates a new adapter wrapping the given PluginDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewPluginDBStoreAdapter(ds PluginDataStore) *PluginDBStoreAdapter {
	return &PluginDBStoreAdapter{
		store:     ds,
		ownedConn: false,
	}
}

// NewPluginDBStoreAdapterOwned creates a new adapter that owns the connection.
// When Close() is called, it will attempt to close the underlying store
// if it implements io.Closer.
func NewPluginDBStoreAdapterOwned(ds PluginDataStore) *PluginDBStoreAdapter {
	return &PluginDBStoreAdapter{
		store:     ds,
		ownedConn: true,
	}
}

// Create adds a new plugin to the store.
// Returns an error if a plugin with the same name already exists.
func (a *PluginDBStoreAdapter) Create(p *plugin.Plugin) error {
	// Check if plugin already exists
	_, err := a.store.GetPluginByName(p.Name)
	if err == nil {
		return &ErrPluginAlreadyExists{Name: p.Name}
	}

	dbPlugin := pluginToDBModel(p)
	if err := a.store.CreatePlugin(dbPlugin); err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	return nil
}

// Update modifies an existing plugin in the store.
// Returns an error if the plugin doesn't exist.
func (a *PluginDBStoreAdapter) Update(p *plugin.Plugin) error {
	// Check if plugin exists
	_, err := a.store.GetPluginByName(p.Name)
	if err != nil {
		return &ErrPluginNotFound{Name: p.Name}
	}

	dbPlugin := pluginToDBModel(p)
	if err := a.store.UpdatePlugin(dbPlugin); err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}
	return nil
}

// Upsert creates or updates a plugin (create if not exists, update if exists).
func (a *PluginDBStoreAdapter) Upsert(p *plugin.Plugin) error {
	existing, err := a.store.GetPluginByName(p.Name)
	if err != nil {
		// Plugin doesn't exist, create it
		dbPlugin := pluginToDBModel(p)
		return a.store.CreatePlugin(dbPlugin)
	}

	// Plugin exists, update it
	dbPlugin := pluginToDBModel(p)
	dbPlugin.ID = existing.ID
	dbPlugin.CreatedAt = existing.CreatedAt
	return a.store.UpdatePlugin(dbPlugin)
}

// Delete removes a plugin from the store by name.
// Returns an error if the plugin doesn't exist.
func (a *PluginDBStoreAdapter) Delete(name string) error {
	// Check if plugin exists first
	_, err := a.store.GetPluginByName(name)
	if err != nil {
		return &ErrPluginNotFound{Name: name}
	}

	if err := a.store.DeletePlugin(name); err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}

// Get retrieves a plugin by name.
// Returns nil and an error if the plugin doesn't exist.
func (a *PluginDBStoreAdapter) Get(name string) (*plugin.Plugin, error) {
	dbPlugin, err := a.store.GetPluginByName(name)
	if err != nil {
		// Check if it's a "not found" error.
		// Cannot use db.IsNotFound() due to import cycle (nvimbridge cannot import db).
		// All db methods now return *db.ErrNotFound whose Error() contains "not found".
		if strings.Contains(err.Error(), "not found") {
			return nil, &ErrPluginNotFound{Name: name}
		}
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	return dbModelToPlugin(dbPlugin), nil
}

// List returns all plugins in the store.
func (a *PluginDBStoreAdapter) List() ([]*plugin.Plugin, error) {
	dbPlugins, err := a.store.ListPlugins()
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	plugins := make([]*plugin.Plugin, len(dbPlugins))
	for i, dbPlugin := range dbPlugins {
		plugins[i] = dbModelToPlugin(dbPlugin)
	}
	return plugins, nil
}

// ListByCategory returns plugins in a specific category.
func (a *PluginDBStoreAdapter) ListByCategory(category string) ([]*plugin.Plugin, error) {
	dbPlugins, err := a.store.ListPluginsByCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by category: %w", err)
	}

	plugins := make([]*plugin.Plugin, len(dbPlugins))
	for i, dbPlugin := range dbPlugins {
		plugins[i] = dbModelToPlugin(dbPlugin)
	}
	return plugins, nil
}

// ListByTag returns plugins that have a specific tag.
func (a *PluginDBStoreAdapter) ListByTag(tag string) ([]*plugin.Plugin, error) {
	dbPlugins, err := a.store.ListPluginsByTags([]string{tag})
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by tag: %w", err)
	}

	plugins := make([]*plugin.Plugin, len(dbPlugins))
	for i, dbPlugin := range dbPlugins {
		plugins[i] = dbModelToPlugin(dbPlugin)
	}
	return plugins, nil
}

// Exists checks if a plugin with the given name exists.
func (a *PluginDBStoreAdapter) Exists(name string) (bool, error) {
	_, err := a.store.GetPluginByName(name)
	if err != nil {
		// Cannot use db.IsNotFound() due to import cycle (nvimbridge cannot import db).
		// All db methods now return *db.ErrNotFound whose Error() contains "not found".
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check plugin existence: %w", err)
	}
	return true, nil
}

// Close releases any resources held by the store.
func (a *PluginDBStoreAdapter) Close() error {
	if a.ownedConn {
		// Try to close if the store implements io.Closer
		if closer, ok := a.store.(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	return nil
}

// Ensure PluginDBStoreAdapter implements NvimPluginStore at compile time.
// Because NvimPluginStore has the same method set as store.PluginStore,
// the adapter also structurally satisfies store.PluginStore without importing it.
var _ NvimPluginStore = (*PluginDBStoreAdapter)(nil)
