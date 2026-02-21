// Package store provides storage abstractions for terminal emulator management.
// This file implements a database adapter that bridges the emulator.EmulatorStore
// interface with the db.DataStore interface, enabling terminalops to use SQLite storage.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/emulator"
)

// EmulatorDataStore is a subset of db.DataStore containing only terminal emulator operations.
// This interface allows the adapter to work with any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type EmulatorDataStore interface {
	// CreateTerminalEmulator inserts a new terminal emulator.
	CreateTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// GetTerminalEmulator retrieves an emulator by its name.
	GetTerminalEmulator(name string) (*models.TerminalEmulatorDB, error)

	// UpdateTerminalEmulator updates an existing emulator.
	UpdateTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// UpsertTerminalEmulator creates or updates an emulator.
	UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// DeleteTerminalEmulator removes an emulator by name.
	DeleteTerminalEmulator(name string) error

	// ListTerminalEmulators retrieves all emulators.
	ListTerminalEmulators() ([]*models.TerminalEmulatorDB, error)

	// ListTerminalEmulatorsByType retrieves emulators filtered by type.
	ListTerminalEmulatorsByType(emulatorType string) ([]*models.TerminalEmulatorDB, error)

	// ListTerminalEmulatorsByWorkspace retrieves emulators filtered by workspace.
	ListTerminalEmulatorsByWorkspace(workspace string) ([]*models.TerminalEmulatorDB, error)
}

// DBEmulatorStore adapts db.DataStore to implement emulator.EmulatorStore.
// This enables terminalops to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type DBEmulatorStore struct {
	store     EmulatorDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewDBEmulatorStore creates a new adapter wrapping the given EmulatorDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewDBEmulatorStore(store EmulatorDataStore) *DBEmulatorStore {
	return &DBEmulatorStore{
		store:     store,
		ownedConn: false,
	}
}

// NewDBEmulatorStoreOwned creates a new adapter that owns the connection.
// When Close() is called, it will attempt to close the underlying store
// if it implements io.Closer.
func NewDBEmulatorStoreOwned(store EmulatorDataStore) *DBEmulatorStore {
	return &DBEmulatorStore{
		store:     store,
		ownedConn: true,
	}
}

// Create adds a new emulator to the store.
// Returns an error if an emulator with the same name already exists.
func (a *DBEmulatorStore) Create(e *emulator.Emulator) error {
	// Check if emulator already exists
	_, err := a.store.GetTerminalEmulator(e.Name)
	if err == nil {
		return &emulator.ErrAlreadyExists{Name: e.Name}
	}

	dbEmulator := emulatorToDBModel(e)
	if err := a.store.CreateTerminalEmulator(dbEmulator); err != nil {
		return fmt.Errorf("failed to create emulator: %w", err)
	}
	return nil
}

// Update modifies an existing emulator in the store.
// Returns an error if the emulator doesn't exist.
func (a *DBEmulatorStore) Update(e *emulator.Emulator) error {
	// Check if emulator exists
	_, err := a.store.GetTerminalEmulator(e.Name)
	if err != nil {
		return &emulator.ErrNotFound{Name: e.Name}
	}

	dbEmulator := emulatorToDBModel(e)
	if err := a.store.UpdateTerminalEmulator(dbEmulator); err != nil {
		return fmt.Errorf("failed to update emulator: %w", err)
	}
	return nil
}

// Upsert creates or updates an emulator (create if not exists, update if exists).
func (a *DBEmulatorStore) Upsert(e *emulator.Emulator) error {
	dbEmulator := emulatorToDBModel(e)
	return a.store.UpsertTerminalEmulator(dbEmulator)
}

// Delete removes an emulator from the store by name.
// Returns an error if the emulator doesn't exist.
func (a *DBEmulatorStore) Delete(name string) error {
	// Check if emulator exists first
	_, err := a.store.GetTerminalEmulator(name)
	if err != nil {
		return &emulator.ErrNotFound{Name: name}
	}

	if err := a.store.DeleteTerminalEmulator(name); err != nil {
		return fmt.Errorf("failed to delete emulator: %w", err)
	}
	return nil
}

// Get retrieves an emulator by name.
// Returns nil and an error if the emulator doesn't exist.
func (a *DBEmulatorStore) Get(name string) (*emulator.Emulator, error) {
	dbEmulator, err := a.store.GetTerminalEmulator(name)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return nil, &emulator.ErrNotFound{Name: name}
		}
		return nil, fmt.Errorf("failed to get emulator: %w", err)
	}
	return dbModelToEmulator(dbEmulator), nil
}

// List returns all emulators in the store.
func (a *DBEmulatorStore) List() ([]*emulator.Emulator, error) {
	dbEmulators, err := a.store.ListTerminalEmulators()
	if err != nil {
		return nil, fmt.Errorf("failed to list emulators: %w", err)
	}

	emulators := make([]*emulator.Emulator, len(dbEmulators))
	for i, dbEmulator := range dbEmulators {
		emulators[i] = dbModelToEmulator(dbEmulator)
	}
	return emulators, nil
}

// ListByType returns emulators of a specific type.
func (a *DBEmulatorStore) ListByType(emulatorType string) ([]*emulator.Emulator, error) {
	dbEmulators, err := a.store.ListTerminalEmulatorsByType(emulatorType)
	if err != nil {
		return nil, fmt.Errorf("failed to list emulators by type: %w", err)
	}

	emulators := make([]*emulator.Emulator, len(dbEmulators))
	for i, dbEmulator := range dbEmulators {
		emulators[i] = dbModelToEmulator(dbEmulator)
	}
	return emulators, nil
}

// ListByWorkspace returns emulators for a specific workspace.
func (a *DBEmulatorStore) ListByWorkspace(workspace string) ([]*emulator.Emulator, error) {
	dbEmulators, err := a.store.ListTerminalEmulatorsByWorkspace(workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to list emulators by workspace: %w", err)
	}

	emulators := make([]*emulator.Emulator, len(dbEmulators))
	for i, dbEmulator := range dbEmulators {
		emulators[i] = dbModelToEmulator(dbEmulator)
	}
	return emulators, nil
}

// Exists checks if an emulator with the given name exists.
func (a *DBEmulatorStore) Exists(name string) (bool, error) {
	_, err := a.store.GetTerminalEmulator(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check emulator existence: %w", err)
	}
	return true, nil
}

// Close releases any resources held by the store.
func (a *DBEmulatorStore) Close() error {
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

// emulatorToDBModel converts an emulator.Emulator to models.TerminalEmulatorDB.
func emulatorToDBModel(e *emulator.Emulator) *models.TerminalEmulatorDB {
	db := &models.TerminalEmulatorDB{
		Name:    e.Name,
		Type:    string(e.Type),
		Enabled: e.Enabled,
	}

	// String fields with sql.NullString
	if e.Description != "" {
		db.Description = sql.NullString{String: e.Description, Valid: true}
	}
	if e.Category != "" {
		db.Category = sql.NullString{String: e.Category, Valid: true}
	}
	if e.ThemeRef != "" {
		db.ThemeRef = sql.NullString{String: e.ThemeRef, Valid: true}
	}
	if e.Workspace != "" {
		db.Workspace = sql.NullString{String: e.Workspace, Valid: true}
	}

	// Configuration as JSON object
	if len(e.Config) > 0 {
		if data, err := json.Marshal(e.Config); err == nil {
			db.Config = string(data)
		} else {
			db.Config = "{}"
		}
	} else {
		db.Config = "{}"
	}

	// Labels as JSON object
	if len(e.Labels) > 0 {
		if data, err := json.Marshal(e.Labels); err == nil {
			db.Labels = string(data)
		} else {
			db.Labels = "{}"
		}
	} else {
		db.Labels = "{}"
	}

	// Timestamps
	if e.CreatedAt != nil {
		db.CreatedAt = *e.CreatedAt
	}
	if e.UpdatedAt != nil {
		db.UpdatedAt = *e.UpdatedAt
	}

	return db
}

// dbModelToEmulator converts a models.TerminalEmulatorDB to emulator.Emulator.
func dbModelToEmulator(db *models.TerminalEmulatorDB) *emulator.Emulator {
	e := &emulator.Emulator{
		Name:    db.Name,
		Type:    emulator.EmulatorType(db.Type),
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		e.Description = db.Description.String
	}
	if db.Category.Valid {
		e.Category = db.Category.String
	}
	if db.ThemeRef.Valid {
		e.ThemeRef = db.ThemeRef.String
	}
	if db.Workspace.Valid {
		e.Workspace = db.Workspace.String
	}

	// Parse configuration JSON
	if db.Config != "" && db.Config != "{}" {
		var config map[string]any
		if err := json.Unmarshal([]byte(db.Config), &config); err == nil {
			e.Config = config
		} else {
			e.Config = make(map[string]any)
		}
	} else {
		e.Config = make(map[string]any)
	}

	// Parse labels JSON
	if db.Labels != "" && db.Labels != "{}" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(db.Labels), &labels); err == nil {
			e.Labels = labels
		} else {
			e.Labels = make(map[string]string)
		}
	} else {
		e.Labels = make(map[string]string)
	}

	// Timestamps
	if !db.CreatedAt.IsZero() {
		e.CreatedAt = &db.CreatedAt
	}
	if !db.UpdatedAt.IsZero() {
		e.UpdatedAt = &db.UpdatedAt
	}

	return e
}

// Ensure DBEmulatorStore implements EmulatorStore interface
var _ emulator.EmulatorStore = (*DBEmulatorStore)(nil)
