// Package store provides storage abstractions for terminal plugin management.
// This file implements a database adapter that bridges the plugin.PluginStore
// interface with the db.DataStore interface, enabling terminalops to use SQLite storage.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/plugin"
)

// PluginDataStore is a subset of db.DataStore containing only terminal plugin operations.
// This interface allows the adapter to work with any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type PluginDataStore interface {
	// CreateTerminalPlugin inserts a new terminal plugin.
	CreateTerminalPlugin(plugin *models.TerminalPluginDB) error

	// GetTerminalPlugin retrieves a plugin by its name.
	GetTerminalPlugin(name string) (*models.TerminalPluginDB, error)

	// UpdateTerminalPlugin updates an existing plugin.
	UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error

	// UpsertTerminalPlugin creates or updates a plugin.
	UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error

	// DeleteTerminalPlugin removes a plugin by name.
	DeleteTerminalPlugin(name string) error

	// ListTerminalPlugins retrieves all plugins.
	ListTerminalPlugins() ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByCategory retrieves plugins filtered by category.
	ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByShell retrieves plugins filtered by shell.
	ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error)
}

// DBPluginStore adapts db.DataStore to implement plugin.PluginStore.
// This enables terminalops to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type DBPluginStore struct {
	store     PluginDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewDBPluginStore creates a new adapter wrapping the given PluginDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewDBPluginStore(store PluginDataStore) *DBPluginStore {
	return &DBPluginStore{
		store:     store,
		ownedConn: false,
	}
}

// NewDBPluginStoreOwned creates a new adapter that owns the connection.
// When Close() is called, it will attempt to close the underlying store
// if it implements io.Closer.
func NewDBPluginStoreOwned(store PluginDataStore) *DBPluginStore {
	return &DBPluginStore{
		store:     store,
		ownedConn: true,
	}
}

// Create adds a new plugin to the store.
// Returns an error if a plugin with the same name already exists.
func (a *DBPluginStore) Create(p *plugin.Plugin) error {
	// Check if plugin already exists
	_, err := a.store.GetTerminalPlugin(p.Name)
	if err == nil {
		return &plugin.ErrAlreadyExists{Name: p.Name}
	}

	dbPlugin := pluginToDBModel(p)
	if err := a.store.CreateTerminalPlugin(dbPlugin); err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	return nil
}

// Update modifies an existing plugin in the store.
// Returns an error if the plugin doesn't exist.
func (a *DBPluginStore) Update(p *plugin.Plugin) error {
	// Check if plugin exists
	_, err := a.store.GetTerminalPlugin(p.Name)
	if err != nil {
		return &plugin.ErrNotFound{Name: p.Name}
	}

	dbPlugin := pluginToDBModel(p)
	if err := a.store.UpdateTerminalPlugin(dbPlugin); err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}
	return nil
}

// Upsert creates or updates a plugin (create if not exists, update if exists).
func (a *DBPluginStore) Upsert(p *plugin.Plugin) error {
	dbPlugin := pluginToDBModel(p)
	return a.store.UpsertTerminalPlugin(dbPlugin)
}

// Delete removes a plugin from the store by name.
// Returns an error if the plugin doesn't exist.
func (a *DBPluginStore) Delete(name string) error {
	// Check if plugin exists first
	_, err := a.store.GetTerminalPlugin(name)
	if err != nil {
		return &plugin.ErrNotFound{Name: name}
	}

	if err := a.store.DeleteTerminalPlugin(name); err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}

// Get retrieves a plugin by name.
// Returns nil and an error if the plugin doesn't exist.
func (a *DBPluginStore) Get(name string) (*plugin.Plugin, error) {
	dbPlugin, err := a.store.GetTerminalPlugin(name)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return nil, &plugin.ErrNotFound{Name: name}
		}
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	return dbModelToPlugin(dbPlugin), nil
}

// List returns all plugins in the store.
func (a *DBPluginStore) List() ([]*plugin.Plugin, error) {
	dbPlugins, err := a.store.ListTerminalPlugins()
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
func (a *DBPluginStore) ListByCategory(category string) ([]*plugin.Plugin, error) {
	dbPlugins, err := a.store.ListTerminalPluginsByCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by category: %w", err)
	}

	plugins := make([]*plugin.Plugin, len(dbPlugins))
	for i, dbPlugin := range dbPlugins {
		plugins[i] = dbModelToPlugin(dbPlugin)
	}
	return plugins, nil
}

// ListByManager returns plugins using a specific plugin manager.
func (a *DBPluginStore) ListByManager(manager plugin.PluginManager) ([]*plugin.Plugin, error) {
	// Get all plugins and filter by manager (could be optimized with a DB method)
	allPlugins, err := a.List()
	if err != nil {
		return nil, err
	}

	var filtered []*plugin.Plugin
	for _, p := range allPlugins {
		if p.Manager == manager {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// Exists checks if a plugin with the given name exists.
func (a *DBPluginStore) Exists(name string) (bool, error) {
	_, err := a.store.GetTerminalPlugin(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check plugin existence: %w", err)
	}
	return true, nil
}

// Close releases any resources held by the store.
func (a *DBPluginStore) Close() error {
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

// pluginToDBModel converts a plugin.Plugin to models.TerminalPluginDB.
func pluginToDBModel(p *plugin.Plugin) *models.TerminalPluginDB {
	db := &models.TerminalPluginDB{
		Name:    p.Name,
		Repo:    p.Repo,
		Enabled: p.Enabled,
	}

	// String fields with sql.NullString
	if p.Description != "" {
		db.Description = sql.NullString{String: p.Description, Valid: true}
	}
	if p.Category != "" {
		db.Category = sql.NullString{String: p.Category, Valid: true}
	}

	// Map plugin manager enum to string
	if p.Manager != "" {
		db.Manager = string(p.Manager)
	} else {
		db.Manager = string(plugin.PluginManagerManual)
	}

	// Shell - default to zsh if not specified
	if p.Source != "" {
		// Store the full source URL or path in the repo field if it's not a GitHub repo
		if !strings.HasPrefix(p.Source, "https://github.com/") {
			db.Repo = p.Source
		}
	}

	// Load command - construct from manager and plugin type
	if p.OhMyZshPlugin != "" {
		db.LoadCommand = sql.NullString{
			String: fmt.Sprintf("plugins+=%s", p.OhMyZshPlugin),
			Valid:  true,
		}
	} else if p.Config != "" {
		db.LoadCommand = sql.NullString{String: p.Config, Valid: true}
	}

	// Source file
	if len(p.SourceFiles) > 0 {
		db.SourceFile = sql.NullString{String: p.SourceFiles[0], Valid: true}
	}

	// Set shell - default to zsh since this is terminal plugin focused
	db.Shell = "zsh"

	// Dependencies as JSON array
	if len(p.Dependencies) > 0 {
		if data, err := json.Marshal(p.Dependencies); err == nil {
			db.Dependencies = string(data)
		} else {
			db.Dependencies = "[]"
		}
	} else {
		db.Dependencies = "[]"
	}

	// Environment variables as JSON object
	if len(p.Env) > 0 {
		if data, err := json.Marshal(p.Env); err == nil {
			db.EnvVars = string(data)
		} else {
			db.EnvVars = "{}"
		}
	} else {
		db.EnvVars = "{}"
	}

	// Labels (tags + metadata) as JSON object
	labels := make(map[string]string)

	// Add tags as labels
	for _, tag := range p.Tags {
		labels["tag:"+tag] = "true"
	}

	// Add other metadata
	if p.LoadMode != "" {
		labels["load_mode"] = string(p.LoadMode)
	}
	if p.Branch != "" {
		labels["branch"] = p.Branch
	}
	if p.Tag != "" {
		labels["tag"] = p.Tag
	}
	if p.Priority > 0 {
		labels["priority"] = fmt.Sprintf("%d", p.Priority)
	}

	if len(labels) > 0 {
		if data, err := json.Marshal(labels); err == nil {
			db.Labels = string(data)
		} else {
			db.Labels = "{}"
		}
	} else {
		db.Labels = "{}"
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

// dbModelToPlugin converts a models.TerminalPluginDB to plugin.Plugin.
func dbModelToPlugin(db *models.TerminalPluginDB) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:    db.Name,
		Repo:    db.Repo,
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}

	// Plugin manager
	p.Manager = plugin.PluginManager(db.Manager)

	// Load command and config
	if db.LoadCommand.Valid {
		p.Config = db.LoadCommand.String

		// If it looks like an oh-my-zsh plugin, extract the plugin name
		if strings.HasPrefix(db.LoadCommand.String, "plugins+=") {
			p.OhMyZshPlugin = strings.TrimPrefix(db.LoadCommand.String, "plugins+=")
		}
	}

	// Source file
	if db.SourceFile.Valid {
		p.SourceFiles = []string{db.SourceFile.String}
	}

	// Parse dependencies JSON
	if db.Dependencies != "" && db.Dependencies != "[]" {
		var deps []string
		if err := json.Unmarshal([]byte(db.Dependencies), &deps); err == nil {
			p.Dependencies = deps
		}
	}

	// Parse env vars JSON
	if db.EnvVars != "" && db.EnvVars != "{}" {
		var envVars map[string]string
		if err := json.Unmarshal([]byte(db.EnvVars), &envVars); err == nil {
			p.Env = envVars
		}
	}

	// Parse labels JSON and extract metadata
	if db.Labels != "" && db.Labels != "{}" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(db.Labels), &labels); err == nil {
			// Extract tags
			var tags []string
			for key, value := range labels {
				if strings.HasPrefix(key, "tag:") && value == "true" {
					tags = append(tags, strings.TrimPrefix(key, "tag:"))
				}
			}
			p.Tags = tags

			// Extract other metadata
			if loadMode, ok := labels["load_mode"]; ok {
				p.LoadMode = plugin.LoadMode(loadMode)
			}
			if branch, ok := labels["branch"]; ok {
				p.Branch = branch
			}
			if tag, ok := labels["tag"]; ok {
				p.Tag = tag
			}
			if priorityStr, ok := labels["priority"]; ok {
				var priority int
				fmt.Sscanf(priorityStr, "%d", &priority)
				p.Priority = priority
			}
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

// Ensure DBPluginStore implements PluginStore interface
var _ plugin.PluginStore = (*DBPluginStore)(nil)
