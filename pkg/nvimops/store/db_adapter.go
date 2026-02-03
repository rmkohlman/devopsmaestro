// Package store provides storage abstractions for nvim-manager.
// This file implements a database adapter that bridges the store.PluginStore
// interface with the db.DataStore interface, enabling nvimops to use SQLite storage.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/plugin"
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

// DBStoreAdapter adapts db.DataStore to implement store.PluginStore.
// This enables nvimops.Manager to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type DBStoreAdapter struct {
	store     PluginDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewDBStoreAdapter creates a new adapter wrapping the given PluginDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewDBStoreAdapter(store PluginDataStore) *DBStoreAdapter {
	return &DBStoreAdapter{
		store:     store,
		ownedConn: false,
	}
}

// NewDBStoreAdapterOwned creates a new adapter that owns the connection.
// When Close() is called, it will attempt to close the underlying store
// if it implements io.Closer.
func NewDBStoreAdapterOwned(store PluginDataStore) *DBStoreAdapter {
	return &DBStoreAdapter{
		store:     store,
		ownedConn: true,
	}
}

// Create adds a new plugin to the store.
// Returns an error if a plugin with the same name already exists.
func (a *DBStoreAdapter) Create(p *plugin.Plugin) error {
	// Check if plugin already exists
	_, err := a.store.GetPluginByName(p.Name)
	if err == nil {
		return &ErrAlreadyExists{Name: p.Name}
	}

	dbPlugin := pluginToDBModel(p)
	if err := a.store.CreatePlugin(dbPlugin); err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	return nil
}

// Update modifies an existing plugin in the store.
// Returns an error if the plugin doesn't exist.
func (a *DBStoreAdapter) Update(p *plugin.Plugin) error {
	// Check if plugin exists
	_, err := a.store.GetPluginByName(p.Name)
	if err != nil {
		return &ErrNotFound{Name: p.Name}
	}

	dbPlugin := pluginToDBModel(p)
	if err := a.store.UpdatePlugin(dbPlugin); err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}
	return nil
}

// Upsert creates or updates a plugin (create if not exists, update if exists).
func (a *DBStoreAdapter) Upsert(p *plugin.Plugin) error {
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
func (a *DBStoreAdapter) Delete(name string) error {
	// Check if plugin exists first
	_, err := a.store.GetPluginByName(name)
	if err != nil {
		return &ErrNotFound{Name: name}
	}

	if err := a.store.DeletePlugin(name); err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}

// Get retrieves a plugin by name.
// Returns nil and an error if the plugin doesn't exist.
func (a *DBStoreAdapter) Get(name string) (*plugin.Plugin, error) {
	dbPlugin, err := a.store.GetPluginByName(name)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return nil, &ErrNotFound{Name: name}
		}
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	return dbModelToPlugin(dbPlugin), nil
}

// List returns all plugins in the store.
func (a *DBStoreAdapter) List() ([]*plugin.Plugin, error) {
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
func (a *DBStoreAdapter) ListByCategory(category string) ([]*plugin.Plugin, error) {
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
func (a *DBStoreAdapter) ListByTag(tag string) ([]*plugin.Plugin, error) {
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
func (a *DBStoreAdapter) Exists(name string) (bool, error) {
	_, err := a.store.GetPluginByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check plugin existence: %w", err)
	}
	return true, nil
}

// Close releases any resources held by the store.
func (a *DBStoreAdapter) Close() error {
	if a.ownedConn {
		// Try to close if the store implements io.Closer
		if closer, ok := a.store.(interface{ Close() error }); ok {
			return closer.Close()
		}
	}
	return nil
}

// =============================================================================
// Model Conversion Functions
// =============================================================================

// pluginToDBModel converts a plugin.Plugin to models.NvimPluginDB.
func pluginToDBModel(p *plugin.Plugin) *models.NvimPluginDB {
	db := &models.NvimPluginDB{
		Name:    p.Name,
		Repo:    p.Repo,
		Lazy:    p.Lazy,
		Enabled: p.Enabled,
	}

	// String fields with sql.NullString
	if p.Description != "" {
		db.Description = sql.NullString{String: p.Description, Valid: true}
	}
	if p.Branch != "" {
		db.Branch = sql.NullString{String: p.Branch, Valid: true}
	}
	if p.Version != "" {
		db.Version = sql.NullString{String: p.Version, Valid: true}
	}
	if p.Category != "" {
		db.Category = sql.NullString{String: p.Category, Valid: true}
	}
	if p.Build != "" {
		db.Build = sql.NullString{String: p.Build, Valid: true}
	}
	if p.Config != "" {
		db.Config = sql.NullString{String: p.Config, Valid: true}
	}
	if p.Init != "" {
		db.Init = sql.NullString{String: p.Init, Valid: true}
	}

	// Priority
	if p.Priority != 0 {
		db.Priority = sql.NullInt64{Int64: int64(p.Priority), Valid: true}
	}

	// JSON array fields
	if len(p.Event) > 0 {
		if data, err := json.Marshal(p.Event); err == nil {
			db.Event = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(p.Ft) > 0 {
		if data, err := json.Marshal(p.Ft); err == nil {
			db.Ft = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(p.Cmd) > 0 {
		if data, err := json.Marshal(p.Cmd); err == nil {
			db.Cmd = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(p.Tags) > 0 {
		if data, err := json.Marshal(p.Tags); err == nil {
			db.Tags = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Keys (convert to compatible format)
	if len(p.Keys) > 0 {
		keysData := make([]models.PluginKeymap, len(p.Keys))
		for i, k := range p.Keys {
			keysData[i] = models.PluginKeymap{
				Key:    k.Key,
				Mode:   modeSliceToInterface(k.Mode),
				Action: k.Action,
				Desc:   k.Desc,
			}
		}
		if data, err := json.Marshal(keysData); err == nil {
			db.Keys = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Keymaps
	if len(p.Keymaps) > 0 {
		keymapsData := make([]models.PluginKeymap, len(p.Keymaps))
		for i, k := range p.Keymaps {
			keymapsData[i] = models.PluginKeymap{
				Key:    k.Key,
				Mode:   modeSliceToInterface(k.Mode),
				Action: k.Action,
				Desc:   k.Desc,
			}
		}
		if data, err := json.Marshal(keymapsData); err == nil {
			db.Keymaps = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Dependencies
	if len(p.Dependencies) > 0 {
		depsData := make([]interface{}, len(p.Dependencies))
		for i, d := range p.Dependencies {
			// Use full struct if has extra fields, otherwise just repo string
			if d.Build != "" || d.Version != "" || d.Branch != "" || d.Config {
				depsData[i] = models.PluginDependency{
					Repo:    d.Repo,
					Build:   d.Build,
					Version: d.Version,
					Branch:  d.Branch,
				}
			} else {
				depsData[i] = d.Repo
			}
		}
		if data, err := json.Marshal(depsData); err == nil {
			db.Dependencies = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Opts (map)
	if len(p.Opts) > 0 {
		if data, err := json.Marshal(p.Opts); err == nil {
			db.Opts = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Timestamps
	if p.CreatedAt != nil {
		db.CreatedAt = *p.CreatedAt
	}
	if p.UpdatedAt != nil {
		db.UpdatedAt = *p.UpdatedAt
	}

	return db
}

// dbModelToPlugin converts a models.NvimPluginDB to plugin.Plugin.
func dbModelToPlugin(db *models.NvimPluginDB) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:    db.Name,
		Repo:    db.Repo,
		Lazy:    db.Lazy,
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Branch.Valid {
		p.Branch = db.Branch.String
	}
	if db.Version.Valid {
		p.Version = db.Version.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}
	if db.Build.Valid {
		p.Build = db.Build.String
	}
	if db.Config.Valid {
		p.Config = db.Config.String
	}
	if db.Init.Valid {
		p.Init = db.Init.String
	}

	// Priority
	if db.Priority.Valid {
		p.Priority = int(db.Priority.Int64)
	}

	// JSON array fields
	if db.Event.Valid {
		var event []string
		if err := json.Unmarshal([]byte(db.Event.String), &event); err == nil {
			p.Event = event
		}
	}
	if db.Ft.Valid {
		var ft []string
		if err := json.Unmarshal([]byte(db.Ft.String), &ft); err == nil {
			p.Ft = ft
		}
	}
	if db.Cmd.Valid {
		var cmd []string
		if err := json.Unmarshal([]byte(db.Cmd.String), &cmd); err == nil {
			p.Cmd = cmd
		}
	}
	if db.Tags.Valid {
		var tags []string
		if err := json.Unmarshal([]byte(db.Tags.String), &tags); err == nil {
			p.Tags = tags
		}
	}

	// Keys
	if db.Keys.Valid {
		var keysData []models.PluginKeymap
		if err := json.Unmarshal([]byte(db.Keys.String), &keysData); err == nil {
			p.Keys = make([]plugin.Keymap, len(keysData))
			for i, k := range keysData {
				p.Keys[i] = plugin.Keymap{
					Key:    k.Key,
					Mode:   interfaceToModeSlice(k.Mode),
					Action: k.Action,
					Desc:   k.Desc,
				}
			}
		}
	}

	// Keymaps
	if db.Keymaps.Valid {
		var keymapsData []models.PluginKeymap
		if err := json.Unmarshal([]byte(db.Keymaps.String), &keymapsData); err == nil {
			p.Keymaps = make([]plugin.Keymap, len(keymapsData))
			for i, k := range keymapsData {
				p.Keymaps[i] = plugin.Keymap{
					Key:    k.Key,
					Mode:   interfaceToModeSlice(k.Mode),
					Action: k.Action,
					Desc:   k.Desc,
				}
			}
		}
	}

	// Dependencies
	if db.Dependencies.Valid {
		var depsRaw []interface{}
		if err := json.Unmarshal([]byte(db.Dependencies.String), &depsRaw); err == nil {
			for _, dep := range depsRaw {
				switch d := dep.(type) {
				case string:
					p.Dependencies = append(p.Dependencies, plugin.Dependency{Repo: d})
				case map[string]interface{}:
					depObj := plugin.Dependency{}
					if repo, ok := d["repo"].(string); ok {
						depObj.Repo = repo
					}
					if build, ok := d["build"].(string); ok {
						depObj.Build = build
					}
					if version, ok := d["version"].(string); ok {
						depObj.Version = version
					}
					if branch, ok := d["branch"].(string); ok {
						depObj.Branch = branch
					}
					if config, ok := d["config"].(bool); ok {
						depObj.Config = config
					}
					p.Dependencies = append(p.Dependencies, depObj)
				}
			}
		}
	}

	// Opts
	if db.Opts.Valid {
		var opts map[string]interface{}
		if err := json.Unmarshal([]byte(db.Opts.String), &opts); err == nil {
			p.Opts = opts
		}
	}

	// Timestamps
	if !db.CreatedAt.IsZero() {
		p.CreatedAt = &db.CreatedAt
	}
	if !db.UpdatedAt.IsZero() {
		p.UpdatedAt = &db.UpdatedAt
	}

	return p
}

// =============================================================================
// Helper Functions
// =============================================================================

// modeSliceToInterface converts []string mode to interface{} for JSON storage.
func modeSliceToInterface(modes []string) interface{} {
	if len(modes) == 0 {
		return nil
	}
	if len(modes) == 1 {
		return modes[0]
	}
	return modes
}

// interfaceToModeSlice converts interface{} mode from JSON to []string.
func interfaceToModeSlice(mode interface{}) []string {
	if mode == nil {
		return nil
	}
	switch m := mode.(type) {
	case string:
		return []string{m}
	case []interface{}:
		result := make([]string, len(m))
		for i, v := range m {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	case []string:
		return m
	}
	return nil
}

// Ensure DBStoreAdapter implements PluginStore interface
var _ PluginStore = (*DBStoreAdapter)(nil)
