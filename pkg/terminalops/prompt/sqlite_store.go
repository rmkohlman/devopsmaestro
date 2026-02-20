package prompt

import (
	"database/sql"
	"encoding/json"

	"devopsmaestro/models"
)

// PromptDataStore is a subset of db.DataStore containing only terminal prompt operations.
// This allows SQLitePromptStore to accept any implementation that provides
// these methods, without depending on the full db.DataStore interface.
type PromptDataStore interface {
	// CreateTerminalPrompt inserts a new terminal prompt.
	CreateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// GetTerminalPromptByName retrieves a terminal prompt by its name.
	GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error)

	// UpdateTerminalPrompt updates an existing terminal prompt.
	UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// DeleteTerminalPrompt removes a terminal prompt by name.
	DeleteTerminalPrompt(name string) error

	// ListTerminalPrompts retrieves all terminal prompts.
	ListTerminalPrompts() ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByType retrieves terminal prompts filtered by type.
	ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByCategory retrieves terminal prompts filtered by category.
	ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error)

	// Close releases any resources held by the store.
	Close() error
}

// SQLitePromptStore implements the PromptStore interface using a DataStore backend.
// It acts as an adapter between the terminal prompt domain and the database layer.
type SQLitePromptStore struct {
	dataStore PromptDataStore
}

// NewSQLitePromptStore creates a new SQLitePromptStore with the given PromptDataStore.
func NewSQLitePromptStore(dataStore PromptDataStore) *SQLitePromptStore {
	return &SQLitePromptStore{
		dataStore: dataStore,
	}
}

// convertToPrompt converts a database terminal prompt to the canonical Prompt type.
func convertToPrompt(tp *models.TerminalPromptDB) (*Prompt, error) {
	p := &Prompt{
		Name:       tp.Name,
		Type:       PromptType(tp.Type),
		AddNewline: tp.AddNewline,
		Enabled:    tp.Enabled,
		CreatedAt:  &tp.CreatedAt,
		UpdatedAt:  &tp.UpdatedAt,
	}

	if tp.Description.Valid {
		p.Description = tp.Description.String
	}

	if tp.Palette.Valid {
		p.Palette = tp.Palette.String
	}

	if tp.Format.Valid {
		p.Format = tp.Format.String
	}

	if tp.PaletteRef.Valid {
		p.PaletteRef = tp.PaletteRef.String
	}

	if tp.RawConfig.Valid {
		p.RawConfig = tp.RawConfig.String
	}

	if tp.Category.Valid {
		p.Category = tp.Category.String
	}

	// Parse JSON fields
	if tp.Modules.Valid {
		var modules map[string]ModuleConfig
		if err := json.Unmarshal([]byte(tp.Modules.String), &modules); err == nil {
			p.Modules = modules
		}
	}

	if tp.Character.Valid {
		var character CharacterConfig
		if err := json.Unmarshal([]byte(tp.Character.String), &character); err == nil {
			p.Character = &character
		}
	}

	if tp.Colors.Valid {
		var colors map[string]string
		if err := json.Unmarshal([]byte(tp.Colors.String), &colors); err == nil {
			p.Colors = colors
		}
	}

	if tp.Tags.Valid {
		var tags []string
		if err := json.Unmarshal([]byte(tp.Tags.String), &tags); err == nil {
			p.Tags = tags
		}
	}

	return p, nil
}

// convertFromPrompt converts a canonical Prompt to the database format.
func convertFromPrompt(p *Prompt) (*models.TerminalPromptDB, error) {
	tp := &models.TerminalPromptDB{
		Name:       p.Name,
		Type:       string(p.Type),
		AddNewline: p.AddNewline,
		Enabled:    p.Enabled,
	}

	if p.Description != "" {
		tp.Description = sql.NullString{String: p.Description, Valid: true}
	}

	if p.Palette != "" {
		tp.Palette = sql.NullString{String: p.Palette, Valid: true}
	}

	if p.Format != "" {
		tp.Format = sql.NullString{String: p.Format, Valid: true}
	}

	if p.PaletteRef != "" {
		tp.PaletteRef = sql.NullString{String: p.PaletteRef, Valid: true}
	}

	if p.RawConfig != "" {
		tp.RawConfig = sql.NullString{String: p.RawConfig, Valid: true}
	}

	if p.Category != "" {
		tp.Category = sql.NullString{String: p.Category, Valid: true}
	}

	// Marshal JSON fields
	if len(p.Modules) > 0 {
		if modulesJSON, err := json.Marshal(p.Modules); err == nil {
			tp.Modules = sql.NullString{String: string(modulesJSON), Valid: true}
		}
	}

	if p.Character != nil {
		if characterJSON, err := json.Marshal(p.Character); err == nil {
			tp.Character = sql.NullString{String: string(characterJSON), Valid: true}
		}
	}

	if len(p.Colors) > 0 {
		if colorsJSON, err := json.Marshal(p.Colors); err == nil {
			tp.Colors = sql.NullString{String: string(colorsJSON), Valid: true}
		}
	}

	if len(p.Tags) > 0 {
		if tagsJSON, err := json.Marshal(p.Tags); err == nil {
			tp.Tags = sql.NullString{String: string(tagsJSON), Valid: true}
		}
	}

	return tp, nil
}

// Create adds a new prompt to the store.
func (s *SQLitePromptStore) Create(p *Prompt) error {
	// Check if prompt already exists
	_, err := s.dataStore.GetTerminalPromptByName(p.Name)
	if err == nil {
		return &ErrAlreadyExists{Name: p.Name}
	}

	// Convert to database model
	dbPrompt, err := convertFromPrompt(p)
	if err != nil {
		return err
	}

	// Create in database
	if err := s.dataStore.CreateTerminalPrompt(dbPrompt); err != nil {
		return err
	}

	// Update the prompt with the generated ID
	if convertedPrompt, err := convertToPrompt(dbPrompt); err == nil {
		*p = *convertedPrompt
	}

	return nil
}

// Update modifies an existing prompt in the store.
func (s *SQLitePromptStore) Update(p *Prompt) error {
	// Check if prompt exists
	existingPrompt, err := s.dataStore.GetTerminalPromptByName(p.Name)
	if err != nil {
		return &ErrNotFound{Name: p.Name}
	}

	// Convert to database model
	dbPrompt, err := convertFromPrompt(p)
	if err != nil {
		return err
	}

	// Preserve the ID and creation timestamp
	dbPrompt.ID = existingPrompt.ID
	dbPrompt.CreatedAt = existingPrompt.CreatedAt

	// Update in database
	return s.dataStore.UpdateTerminalPrompt(dbPrompt)
}

// Upsert creates or updates a prompt (create if not exists, update if exists).
func (s *SQLitePromptStore) Upsert(p *Prompt) error {
	// Try to get existing prompt
	_, err := s.dataStore.GetTerminalPromptByName(p.Name)
	if err != nil {
		// Prompt doesn't exist, create it
		return s.Create(p)
	}
	// Prompt exists, update it
	return s.Update(p)
}

// Delete removes a prompt from the store by name.
func (s *SQLitePromptStore) Delete(name string) error {
	// Check if prompt exists
	_, err := s.dataStore.GetTerminalPromptByName(name)
	if err != nil {
		return &ErrNotFound{Name: name}
	}

	// Delete from database
	return s.dataStore.DeleteTerminalPrompt(name)
}

// Get retrieves a prompt by name.
func (s *SQLitePromptStore) Get(name string) (*Prompt, error) {
	dbPrompt, err := s.dataStore.GetTerminalPromptByName(name)
	if err != nil {
		return nil, &ErrNotFound{Name: name}
	}

	return convertToPrompt(dbPrompt)
}

// List returns all prompts in the store.
func (s *SQLitePromptStore) List() ([]*Prompt, error) {
	dbPrompts, err := s.dataStore.ListTerminalPrompts()
	if err != nil {
		return nil, err
	}

	prompts := make([]*Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompt, err := convertToPrompt(dbPrompt)
		if err != nil {
			continue // Skip prompts that can't be converted
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// ListByType returns prompts of a specific type (starship, powerlevel10k, etc.).
func (s *SQLitePromptStore) ListByType(promptType PromptType) ([]*Prompt, error) {
	dbPrompts, err := s.dataStore.ListTerminalPromptsByType(string(promptType))
	if err != nil {
		return nil, err
	}

	prompts := make([]*Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompt, err := convertToPrompt(dbPrompt)
		if err != nil {
			continue // Skip prompts that can't be converted
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// ListByCategory returns prompts in a specific category.
func (s *SQLitePromptStore) ListByCategory(category string) ([]*Prompt, error) {
	dbPrompts, err := s.dataStore.ListTerminalPromptsByCategory(category)
	if err != nil {
		return nil, err
	}

	prompts := make([]*Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompt, err := convertToPrompt(dbPrompt)
		if err != nil {
			continue // Skip prompts that can't be converted
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// Exists checks if a prompt with the given name exists.
func (s *SQLitePromptStore) Exists(name string) (bool, error) {
	_, err := s.dataStore.GetTerminalPromptByName(name)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Close releases any resources held by the store.
func (s *SQLitePromptStore) Close() error {
	return s.dataStore.Close()
}
