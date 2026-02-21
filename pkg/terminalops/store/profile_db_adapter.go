// Package store provides storage abstractions for terminal profile management.
// This file implements a database adapter that bridges the profile.ProfileStore
// interface with the db.DataStore interface, enabling terminalops to use SQLite storage.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/profile"
)

// ProfileDataStore is a subset of db.DataStore containing only terminal profile operations.
// This interface allows the adapter to work with any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type ProfileDataStore interface {
	// CreateTerminalProfile inserts a new terminal profile.
	CreateTerminalProfile(profile *models.TerminalProfileDB) error

	// GetTerminalProfileByName retrieves a profile by its name.
	GetTerminalProfileByName(name string) (*models.TerminalProfileDB, error)

	// UpdateTerminalProfile updates an existing profile.
	UpdateTerminalProfile(profile *models.TerminalProfileDB) error

	// UpsertTerminalProfile creates or updates a profile.
	UpsertTerminalProfile(profile *models.TerminalProfileDB) error

	// DeleteTerminalProfile removes a profile by name.
	DeleteTerminalProfile(name string) error

	// ListTerminalProfiles retrieves all profiles.
	ListTerminalProfiles() ([]*models.TerminalProfileDB, error)

	// ListTerminalProfilesByCategory retrieves profiles filtered by category.
	ListTerminalProfilesByCategory(category string) ([]*models.TerminalProfileDB, error)
}

// DBProfileStore adapts db.DataStore to implement profile.ProfileStore.
// This enables terminalops to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type DBProfileStore struct {
	store     ProfileDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewDBProfileStore creates a new adapter wrapping the given ProfileDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewDBProfileStore(store ProfileDataStore) *DBProfileStore {
	return &DBProfileStore{
		store:     store,
		ownedConn: false,
	}
}

// NewDBProfileStoreOwned creates a new adapter that owns the connection.
// When Close() is called, it will attempt to close the underlying store
// if it implements io.Closer.
func NewDBProfileStoreOwned(store ProfileDataStore) *DBProfileStore {
	return &DBProfileStore{
		store:     store,
		ownedConn: true,
	}
}

// Create adds a new profile to the store.
// Returns an error if a profile with the same name already exists.
func (a *DBProfileStore) Create(p *profile.Profile) error {
	// Check if profile already exists
	_, err := a.store.GetTerminalProfileByName(p.Name)
	if err == nil {
		return &profile.ErrAlreadyExists{Name: p.Name}
	}

	dbProfile := profileToDBModel(p)
	if err := a.store.CreateTerminalProfile(dbProfile); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}
	return nil
}

// Update modifies an existing profile in the store.
// Returns an error if the profile doesn't exist.
func (a *DBProfileStore) Update(p *profile.Profile) error {
	// Check if profile exists
	_, err := a.store.GetTerminalProfileByName(p.Name)
	if err != nil {
		return &profile.ErrNotFound{Name: p.Name}
	}

	dbProfile := profileToDBModel(p)
	if err := a.store.UpdateTerminalProfile(dbProfile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}
	return nil
}

// Upsert creates or updates a profile (create if not exists, update if exists).
func (a *DBProfileStore) Upsert(p *profile.Profile) error {
	dbProfile := profileToDBModel(p)
	return a.store.UpsertTerminalProfile(dbProfile)
}

// Delete removes a profile from the store by name.
// Returns an error if the profile doesn't exist.
func (a *DBProfileStore) Delete(name string) error {
	// Check if profile exists first
	_, err := a.store.GetTerminalProfileByName(name)
	if err != nil {
		return &profile.ErrNotFound{Name: name}
	}

	if err := a.store.DeleteTerminalProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// Get retrieves a profile by name.
// Returns nil and an error if the profile doesn't exist.
func (a *DBProfileStore) Get(name string) (*profile.Profile, error) {
	dbProfile, err := a.store.GetTerminalProfileByName(name)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return nil, &profile.ErrNotFound{Name: name}
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return dbModelToProfile(dbProfile), nil
}

// List returns all profiles in the store.
func (a *DBProfileStore) List() ([]*profile.Profile, error) {
	dbProfiles, err := a.store.ListTerminalProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	profiles := make([]*profile.Profile, len(dbProfiles))
	for i, dbProfile := range dbProfiles {
		profiles[i] = dbModelToProfile(dbProfile)
	}
	return profiles, nil
}

// ListByCategory returns profiles in a specific category.
func (a *DBProfileStore) ListByCategory(category string) ([]*profile.Profile, error) {
	dbProfiles, err := a.store.ListTerminalProfilesByCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles by category: %w", err)
	}

	profiles := make([]*profile.Profile, len(dbProfiles))
	for i, dbProfile := range dbProfiles {
		profiles[i] = dbModelToProfile(dbProfile)
	}
	return profiles, nil
}

// Exists checks if a profile with the given name exists.
func (a *DBProfileStore) Exists(name string) (bool, error) {
	_, err := a.store.GetTerminalProfileByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check profile existence: %w", err)
	}
	return true, nil
}

// Close releases any resources held by the store.
func (a *DBProfileStore) Close() error {
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

// profileToDBModel converts a profile.Profile to models.TerminalProfileDB.
func profileToDBModel(p *profile.Profile) *models.TerminalProfileDB {
	db := &models.TerminalProfileDB{
		Name:    p.Name,
		Enabled: p.Enabled,
	}

	// String fields with sql.NullString
	if p.Description != "" {
		db.Description = sql.NullString{String: p.Description, Valid: true}
	}
	if p.Category != "" {
		db.Category = sql.NullString{String: p.Category, Valid: true}
	}
	if p.PromptRef != "" {
		db.PromptRef = sql.NullString{String: p.PromptRef, Valid: true}
	}
	if p.ShellRef != "" {
		db.ShellRef = sql.NullString{String: p.ShellRef, Valid: true}
	}
	if p.ThemeRef != "" {
		db.ThemeRef = sql.NullString{String: p.ThemeRef, Valid: true}
	}

	// PluginRefs as JSON array
	if len(p.PluginRefs) > 0 {
		if data, err := json.Marshal(p.PluginRefs); err == nil {
			db.PluginRefs = string(data)
		} else {
			db.PluginRefs = "[]"
		}
	} else {
		db.PluginRefs = "[]"
	}

	// Tags as JSON array
	if len(p.Tags) > 0 {
		if data, err := json.Marshal(p.Tags); err == nil {
			db.Tags = string(data)
		} else {
			db.Tags = "[]"
		}
	} else {
		db.Tags = "[]"
	}

	// Labels as JSON object (can be populated from metadata if needed)
	db.Labels = "{}"

	// Timestamps
	if p.CreatedAt != nil {
		db.CreatedAt = *p.CreatedAt
	}
	if p.UpdatedAt != nil {
		db.UpdatedAt = *p.UpdatedAt
	}

	return db
}

// dbModelToProfile converts a models.TerminalProfileDB to profile.Profile.
func dbModelToProfile(db *models.TerminalProfileDB) *profile.Profile {
	p := &profile.Profile{
		Name:    db.Name,
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}
	if db.PromptRef.Valid {
		p.PromptRef = db.PromptRef.String
	}
	if db.ShellRef.Valid {
		p.ShellRef = db.ShellRef.String
	}
	if db.ThemeRef.Valid {
		p.ThemeRef = db.ThemeRef.String
	}

	// Parse PluginRefs JSON
	if db.PluginRefs != "" && db.PluginRefs != "[]" {
		var pluginRefs []string
		if err := json.Unmarshal([]byte(db.PluginRefs), &pluginRefs); err == nil {
			p.PluginRefs = pluginRefs
		}
	}

	// Parse Tags JSON
	if db.Tags != "" && db.Tags != "[]" {
		var tags []string
		if err := json.Unmarshal([]byte(db.Tags), &tags); err == nil {
			p.Tags = tags
		}
	}

	// Timestamps
	if !db.CreatedAt.IsZero() {
		p.CreatedAt = &db.CreatedAt
	}
	if !db.UpdatedAt.IsZero() {
		p.UpdatedAt = &db.UpdatedAt
	}

	// Note: Inline configurations (Prompt, Plugins, Shell) are not stored in DB
	// Those are references only. Inline configs would need separate resolution logic.

	return p
}

// Ensure DBProfileStore implements ProfileStore interface
var _ profile.ProfileStore = (*DBProfileStore)(nil)
