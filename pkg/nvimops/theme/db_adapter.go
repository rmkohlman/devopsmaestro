// Package theme provides types and utilities for Neovim theme management.
// This file implements a database adapter that bridges the theme.Store
// interface with the db.DataStore interface, enabling dvm to use SQLite storage.
package theme

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
)

// ThemeDataStore is a subset of db.DataStore containing only theme operations.
// This interface allows the adapter to work with any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type ThemeDataStore interface {
	// CreateTheme inserts a new nvim theme.
	CreateTheme(theme *models.NvimThemeDB) error

	// GetThemeByName retrieves a theme by its name.
	GetThemeByName(name string) (*models.NvimThemeDB, error)

	// UpdateTheme updates an existing theme.
	UpdateTheme(theme *models.NvimThemeDB) error

	// DeleteTheme removes a theme by name.
	DeleteTheme(name string) error

	// ListThemes retrieves all themes.
	ListThemes() ([]*models.NvimThemeDB, error)

	// ListThemesByCategory retrieves themes filtered by category.
	ListThemesByCategory(category string) ([]*models.NvimThemeDB, error)

	// GetActiveTheme retrieves the currently active theme.
	GetActiveTheme() (*models.NvimThemeDB, error)

	// SetActiveTheme sets the active theme by name.
	SetActiveTheme(name string) error

	// ClearActiveTheme deactivates all themes.
	ClearActiveTheme() error
}

// DBStoreAdapter adapts db.DataStore to implement theme.Store.
// This enables nvimops to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type DBStoreAdapter struct {
	store     ThemeDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewDBStoreAdapter creates a new adapter wrapping the given ThemeDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewDBStoreAdapter(store ThemeDataStore) *DBStoreAdapter {
	return &DBStoreAdapter{
		store:     store,
		ownedConn: false,
	}
}

// NewDBStoreAdapterOwned creates a new adapter that owns the connection.
// When the adapter is done, it will attempt to close the underlying store
// if it implements io.Closer.
func NewDBStoreAdapterOwned(store ThemeDataStore) *DBStoreAdapter {
	return &DBStoreAdapter{
		store:     store,
		ownedConn: true,
	}
}

// Get retrieves a theme by name.
func (a *DBStoreAdapter) Get(name string) (*Theme, error) {
	dbTheme, err := a.store.GetThemeByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("theme %q not found", name)
		}
		return nil, fmt.Errorf("failed to get theme: %w", err)
	}
	return dbModelToTheme(dbTheme), nil
}

// List returns all stored themes.
func (a *DBStoreAdapter) List() ([]*Theme, error) {
	dbThemes, err := a.store.ListThemes()
	if err != nil {
		return nil, fmt.Errorf("failed to list themes: %w", err)
	}

	themes := make([]*Theme, len(dbThemes))
	for i, dbTheme := range dbThemes {
		themes[i] = dbModelToTheme(dbTheme)
	}
	return themes, nil
}

// Save stores a theme.
func (a *DBStoreAdapter) Save(theme *Theme) error {
	if err := theme.Validate(); err != nil {
		return fmt.Errorf("invalid theme: %w", err)
	}

	// Check if theme already exists
	existing, err := a.store.GetThemeByName(theme.Name)
	if err == nil {
		// Theme exists, update it
		dbTheme := themeToDBModel(theme)
		dbTheme.ID = existing.ID
		dbTheme.CreatedAt = existing.CreatedAt
		dbTheme.IsActive = existing.IsActive // Preserve active state
		return a.store.UpdateTheme(dbTheme)
	}

	// Theme doesn't exist, create it
	dbTheme := themeToDBModel(theme)
	return a.store.CreateTheme(dbTheme)
}

// Delete removes a theme by name.
func (a *DBStoreAdapter) Delete(name string) error {
	// Check if theme exists first
	_, err := a.store.GetThemeByName(name)
	if err != nil {
		return fmt.Errorf("theme %q not found", name)
	}

	if err := a.store.DeleteTheme(name); err != nil {
		return fmt.Errorf("failed to delete theme: %w", err)
	}
	return nil
}

// GetActive returns the currently active theme.
func (a *DBStoreAdapter) GetActive() (*Theme, error) {
	dbTheme, err := a.store.GetActiveTheme()
	if err != nil {
		return nil, fmt.Errorf("failed to get active theme: %w", err)
	}
	if dbTheme == nil {
		return nil, nil // No active theme
	}
	return dbModelToTheme(dbTheme), nil
}

// SetActive sets the active theme by name.
func (a *DBStoreAdapter) SetActive(name string) error {
	if err := a.store.SetActiveTheme(name); err != nil {
		return fmt.Errorf("failed to set active theme: %w", err)
	}
	return nil
}

// Path returns an empty path since DB storage doesn't have a file path.
func (a *DBStoreAdapter) Path() string {
	return ""
}

// =============================================================================
// Model Conversion Functions
// =============================================================================

// themeToDBModel converts a theme.Theme to models.NvimThemeDB.
func themeToDBModel(t *Theme) *models.NvimThemeDB {
	db := &models.NvimThemeDB{
		Name:        t.Name,
		PluginRepo:  t.Plugin.Repo,
		Transparent: t.Transparent,
	}

	// String fields with sql.NullString
	if t.Description != "" {
		db.Description = sql.NullString{String: t.Description, Valid: true}
	}
	if t.Author != "" {
		db.Author = sql.NullString{String: t.Author, Valid: true}
	}
	if t.Category != "" {
		db.Category = sql.NullString{String: t.Category, Valid: true}
	}
	if t.Plugin.Branch != "" {
		db.PluginBranch = sql.NullString{String: t.Plugin.Branch, Valid: true}
	}
	if t.Plugin.Tag != "" {
		db.PluginTag = sql.NullString{String: t.Plugin.Tag, Valid: true}
	}
	if t.Style != "" {
		db.Style = sql.NullString{String: t.Style, Valid: true}
	}

	// JSON fields
	if len(t.Colors) > 0 {
		if data, err := json.Marshal(t.Colors); err == nil {
			db.Colors = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(t.Options) > 0 {
		if data, err := json.Marshal(t.Options); err == nil {
			db.Options = sql.NullString{String: string(data), Valid: true}
		}
	}

	return db
}

// dbModelToTheme converts a models.NvimThemeDB to theme.Theme.
func dbModelToTheme(db *models.NvimThemeDB) *Theme {
	t := &Theme{
		Name: db.Name,
		Plugin: ThemePlugin{
			Repo: db.PluginRepo,
		},
		Transparent: db.Transparent,
	}

	// String fields
	if db.Description.Valid {
		t.Description = db.Description.String
	}
	if db.Author.Valid {
		t.Author = db.Author.String
	}
	if db.Category.Valid {
		t.Category = db.Category.String
	}
	if db.PluginBranch.Valid {
		t.Plugin.Branch = db.PluginBranch.String
	}
	if db.PluginTag.Valid {
		t.Plugin.Tag = db.PluginTag.String
	}
	if db.Style.Valid {
		t.Style = db.Style.String
	}

	// JSON fields
	if db.Colors.Valid {
		var colors map[string]string
		if err := json.Unmarshal([]byte(db.Colors.String), &colors); err == nil {
			t.Colors = colors
		}
	}
	if db.Options.Valid {
		var options map[string]any
		if err := json.Unmarshal([]byte(db.Options.String), &options); err == nil {
			t.Options = options
		}
	}

	return t
}

// Ensure DBStoreAdapter implements Store interface
var _ Store = (*DBStoreAdapter)(nil)
