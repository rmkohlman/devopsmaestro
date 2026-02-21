// Package store provides storage abstractions for terminal prompt management.
// This file implements a database adapter that bridges the prompt.PromptStore
// interface with the db.DataStore interface, enabling terminalops to use SQLite storage.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/prompt"
)

// PromptDataStore is a subset of db.DataStore containing only terminal prompt operations.
// This interface allows the adapter to work with any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type PromptDataStore interface {
	// CreateTerminalPrompt inserts a new terminal prompt.
	CreateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// GetTerminalPromptByName retrieves a prompt by its name.
	GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error)

	// UpdateTerminalPrompt updates an existing prompt.
	UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// UpsertTerminalPrompt creates or updates a prompt.
	UpsertTerminalPrompt(prompt *models.TerminalPromptDB) error

	// DeleteTerminalPrompt removes a prompt by name.
	DeleteTerminalPrompt(name string) error

	// ListTerminalPrompts retrieves all prompts.
	ListTerminalPrompts() ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByType retrieves prompts filtered by type.
	ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByCategory retrieves prompts filtered by category.
	ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error)
}

// DBPromptStore adapts db.DataStore to implement prompt.PromptStore.
// This enables terminalops to use SQLite storage via the DataStore interface,
// providing a unified storage location for both nvp and dvm.
type DBPromptStore struct {
	store     PromptDataStore
	ownedConn bool // true if we own the connection and should close it
}

// NewDBPromptStore creates a new adapter wrapping the given PromptDataStore.
// The adapter does NOT take ownership of the connection - caller is responsible
// for closing the underlying DataStore.
func NewDBPromptStore(store PromptDataStore) *DBPromptStore {
	return &DBPromptStore{
		store:     store,
		ownedConn: false,
	}
}

// NewDBPromptStoreOwned creates a new adapter that owns the connection.
// When Close() is called, it will attempt to close the underlying store
// if it implements io.Closer.
func NewDBPromptStoreOwned(store PromptDataStore) *DBPromptStore {
	return &DBPromptStore{
		store:     store,
		ownedConn: true,
	}
}

// Create adds a new prompt to the store.
// Returns an error if a prompt with the same name already exists.
func (a *DBPromptStore) Create(p *prompt.Prompt) error {
	// Check if prompt already exists
	_, err := a.store.GetTerminalPromptByName(p.Name)
	if err == nil {
		return &prompt.ErrAlreadyExists{Name: p.Name}
	}

	dbPrompt := promptToDBModel(p)
	if err := a.store.CreateTerminalPrompt(dbPrompt); err != nil {
		return fmt.Errorf("failed to create prompt: %w", err)
	}
	return nil
}

// Update modifies an existing prompt in the store.
// Returns an error if the prompt doesn't exist.
func (a *DBPromptStore) Update(p *prompt.Prompt) error {
	// Check if prompt exists
	_, err := a.store.GetTerminalPromptByName(p.Name)
	if err != nil {
		return &prompt.ErrNotFound{Name: p.Name}
	}

	dbPrompt := promptToDBModel(p)
	if err := a.store.UpdateTerminalPrompt(dbPrompt); err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}
	return nil
}

// Upsert creates or updates a prompt (create if not exists, update if exists).
func (a *DBPromptStore) Upsert(p *prompt.Prompt) error {
	dbPrompt := promptToDBModel(p)
	return a.store.UpsertTerminalPrompt(dbPrompt)
}

// Delete removes a prompt from the store by name.
// Returns an error if the prompt doesn't exist.
func (a *DBPromptStore) Delete(name string) error {
	// Check if prompt exists first
	_, err := a.store.GetTerminalPromptByName(name)
	if err != nil {
		return &prompt.ErrNotFound{Name: name}
	}

	if err := a.store.DeleteTerminalPrompt(name); err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}
	return nil
}

// Get retrieves a prompt by name.
// Returns nil and an error if the prompt doesn't exist.
func (a *DBPromptStore) Get(name string) (*prompt.Prompt, error) {
	dbPrompt, err := a.store.GetTerminalPromptByName(name)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return nil, &prompt.ErrNotFound{Name: name}
		}
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}
	return dbModelToPrompt(dbPrompt), nil
}

// List returns all prompts in the store.
func (a *DBPromptStore) List() ([]*prompt.Prompt, error) {
	dbPrompts, err := a.store.ListTerminalPrompts()
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	prompts := make([]*prompt.Prompt, len(dbPrompts))
	for i, dbPrompt := range dbPrompts {
		prompts[i] = dbModelToPrompt(dbPrompt)
	}
	return prompts, nil
}

// ListByType returns prompts of a specific type.
func (a *DBPromptStore) ListByType(promptType prompt.PromptType) ([]*prompt.Prompt, error) {
	dbPrompts, err := a.store.ListTerminalPromptsByType(string(promptType))
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts by type: %w", err)
	}

	prompts := make([]*prompt.Prompt, len(dbPrompts))
	for i, dbPrompt := range dbPrompts {
		prompts[i] = dbModelToPrompt(dbPrompt)
	}
	return prompts, nil
}

// ListByCategory returns prompts in a specific category.
func (a *DBPromptStore) ListByCategory(category string) ([]*prompt.Prompt, error) {
	dbPrompts, err := a.store.ListTerminalPromptsByCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts by category: %w", err)
	}

	prompts := make([]*prompt.Prompt, len(dbPrompts))
	for i, dbPrompt := range dbPrompts {
		prompts[i] = dbModelToPrompt(dbPrompt)
	}
	return prompts, nil
}

// Exists checks if a prompt with the given name exists.
func (a *DBPromptStore) Exists(name string) (bool, error) {
	_, err := a.store.GetTerminalPromptByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check prompt existence: %w", err)
	}
	return true, nil
}

// Close releases any resources held by the store.
func (a *DBPromptStore) Close() error {
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

// promptToDBModel converts a prompt.Prompt to models.TerminalPromptDB.
func promptToDBModel(p *prompt.Prompt) *models.TerminalPromptDB {
	db := &models.TerminalPromptDB{
		Name:       p.Name,
		Type:       string(p.Type),
		AddNewline: p.AddNewline,
		Enabled:    p.Enabled,
	}

	// String fields with sql.NullString
	if p.Description != "" {
		db.Description = sql.NullString{String: p.Description, Valid: true}
	}
	if p.Palette != "" {
		db.Palette = sql.NullString{String: p.Palette, Valid: true}
	}
	if p.Format != "" {
		db.Format = sql.NullString{String: p.Format, Valid: true}
	}
	if p.PaletteRef != "" {
		db.PaletteRef = sql.NullString{String: p.PaletteRef, Valid: true}
	}
	if p.RawConfig != "" {
		db.RawConfig = sql.NullString{String: p.RawConfig, Valid: true}
	}
	if p.Category != "" {
		db.Category = sql.NullString{String: p.Category, Valid: true}
	}

	// Modules as JSON
	if p.Modules != nil && len(p.Modules) > 0 {
		if data, err := json.Marshal(p.Modules); err == nil {
			db.Modules = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Character config as JSON
	if p.Character != nil {
		if data, err := json.Marshal(p.Character); err == nil {
			db.Character = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Colors as JSON
	if p.Colors != nil && len(p.Colors) > 0 {
		if data, err := json.Marshal(p.Colors); err == nil {
			db.Colors = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Tags as JSON array
	if len(p.Tags) > 0 {
		if data, err := json.Marshal(p.Tags); err == nil {
			db.Tags = sql.NullString{String: string(data), Valid: true}
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

// dbModelToPrompt converts a models.TerminalPromptDB to prompt.Prompt.
func dbModelToPrompt(db *models.TerminalPromptDB) *prompt.Prompt {
	p := &prompt.Prompt{
		Name:       db.Name,
		Type:       prompt.PromptType(db.Type),
		AddNewline: db.AddNewline,
		Enabled:    db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Palette.Valid {
		p.Palette = db.Palette.String
	}
	if db.Format.Valid {
		p.Format = db.Format.String
	}
	if db.PaletteRef.Valid {
		p.PaletteRef = db.PaletteRef.String
	}
	if db.RawConfig.Valid {
		p.RawConfig = db.RawConfig.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}

	// Parse modules JSON
	if db.Modules.Valid && db.Modules.String != "" {
		var modules map[string]prompt.ModuleConfig
		if err := json.Unmarshal([]byte(db.Modules.String), &modules); err == nil {
			p.Modules = modules
		}
	}

	// Parse character config JSON
	if db.Character.Valid && db.Character.String != "" {
		var character prompt.CharacterConfig
		if err := json.Unmarshal([]byte(db.Character.String), &character); err == nil {
			p.Character = &character
		}
	}

	// Parse colors JSON
	if db.Colors.Valid && db.Colors.String != "" {
		var colors map[string]string
		if err := json.Unmarshal([]byte(db.Colors.String), &colors); err == nil {
			p.Colors = colors
		}
	}

	// Parse tags JSON
	if db.Tags.Valid && db.Tags.String != "" {
		var tags []string
		if err := json.Unmarshal([]byte(db.Tags.String), &tags); err == nil {
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

	return p
}

// Ensure DBPromptStore implements PromptStore interface
var _ prompt.PromptStore = (*DBPromptStore)(nil)
