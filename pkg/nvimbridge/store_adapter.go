// Package nvimbridge provides database adapters that bridge MaestroNvim types
// with dvm's database layer (models, db packages).
// This file implements a database adapter that bridges the store.PluginStore
// interface with the db.DataStore interface, enabling nvimops to use SQLite storage.
package nvimbridge

import (
	"fmt"
	"strings"

	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroNvim/nvimops/store"
)

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

// PluginDBStoreAdapter adapts db.DataStore to implement store.PluginStore.
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
		return &store.ErrAlreadyExists{Name: p.Name}
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
		return &store.ErrNotFound{Name: p.Name}
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
		return &store.ErrNotFound{Name: name}
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
			return nil, &store.ErrNotFound{Name: name}
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

// Ensure PluginDBStoreAdapter implements store.PluginStore interface
var _ store.PluginStore = (*PluginDBStoreAdapter)(nil)
