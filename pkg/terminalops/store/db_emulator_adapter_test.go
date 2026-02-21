package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/emulator"
)

// MockEmulatorDataStore implements EmulatorDataStore for testing
type MockEmulatorDataStore struct {
	emulators map[string]*models.TerminalEmulatorDB
	closed    bool
}

func NewMockEmulatorDataStore() *MockEmulatorDataStore {
	return &MockEmulatorDataStore{
		emulators: make(map[string]*models.TerminalEmulatorDB),
	}
}

func (m *MockEmulatorDataStore) CreateTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	if _, exists := m.emulators[emulator.Name]; exists {
		return assert.AnError // Simulate constraint violation
	}
	m.emulators[emulator.Name] = emulator
	return nil
}

func (m *MockEmulatorDataStore) GetTerminalEmulator(name string) (*models.TerminalEmulatorDB, error) {
	emulator, exists := m.emulators[name]
	if !exists {
		return nil, fmt.Errorf("terminal emulator not found: %s", name)
	}
	return emulator, nil
}

func (m *MockEmulatorDataStore) UpdateTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	if _, exists := m.emulators[emulator.Name]; !exists {
		return fmt.Errorf("terminal emulator not found: %s", emulator.Name)
	}
	m.emulators[emulator.Name] = emulator
	return nil
}

func (m *MockEmulatorDataStore) UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	m.emulators[emulator.Name] = emulator
	return nil
}

func (m *MockEmulatorDataStore) DeleteTerminalEmulator(name string) error {
	if _, exists := m.emulators[name]; !exists {
		return fmt.Errorf("terminal emulator not found: %s", name)
	}
	delete(m.emulators, name)
	return nil
}

func (m *MockEmulatorDataStore) ListTerminalEmulators() ([]*models.TerminalEmulatorDB, error) {
	var result []*models.TerminalEmulatorDB
	for _, emulator := range m.emulators {
		result = append(result, emulator)
	}
	return result, nil
}

func (m *MockEmulatorDataStore) ListTerminalEmulatorsByType(emulatorType string) ([]*models.TerminalEmulatorDB, error) {
	var result []*models.TerminalEmulatorDB
	for _, emulator := range m.emulators {
		if emulator.Type == emulatorType {
			result = append(result, emulator)
		}
	}
	return result, nil
}

func (m *MockEmulatorDataStore) ListTerminalEmulatorsByWorkspace(workspace string) ([]*models.TerminalEmulatorDB, error) {
	var result []*models.TerminalEmulatorDB
	for _, emulator := range m.emulators {
		if emulator.Workspace.Valid && emulator.Workspace.String == workspace {
			result = append(result, emulator)
		}
	}
	return result, nil
}

func (m *MockEmulatorDataStore) Close() error {
	m.closed = true
	return nil
}

func TestDBEmulatorStore_Create(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	emulator := &emulator.Emulator{
		Name:        "test-wezterm",
		Description: "Test WezTerm config",
		Type:        emulator.EmulatorTypeWezterm,
		Config: map[string]any{
			"font_size": float64(12), // JSON marshaling converts to float64
			"theme":     "dark",
		},
		Labels: map[string]string{
			"env": "test",
		},
		Enabled: true,
	}

	// Test successful creation
	err := adapter.Create(emulator)
	assert.NoError(t, err)

	// Verify emulator was stored
	stored, err := adapter.Get("test-wezterm")
	assert.NoError(t, err)
	assert.Equal(t, emulator.Name, stored.Name)
	assert.Equal(t, emulator.Type, stored.Type)
	assert.Equal(t, emulator.Config, stored.Config)
	assert.Equal(t, emulator.Labels, stored.Labels)

	// Test creating duplicate emulator
	err = adapter.Create(emulator)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDBEmulatorStore_Get(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	testEmulator := &emulator.Emulator{
		Name: "test-alacritty",
		Type: emulator.EmulatorTypeAlacritty,
		Config: map[string]any{
			"window": map[string]any{
				"opacity": 0.8,
			},
		},
		Enabled: true,
	}

	// Create emulator first
	err := adapter.Create(testEmulator)
	require.NoError(t, err)

	// Test successful get
	retrieved, err := adapter.Get("test-alacritty")
	assert.NoError(t, err)
	assert.Equal(t, testEmulator.Name, retrieved.Name)
	assert.Equal(t, testEmulator.Type, retrieved.Type)
	assert.Equal(t, testEmulator.Config, retrieved.Config)

	// Test get non-existent emulator
	_, err = adapter.Get("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDBEmulatorStore_Update(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	testEmulator := &emulator.Emulator{
		Name:    "test-kitty",
		Type:    emulator.EmulatorTypeKitty,
		Enabled: true,
	}

	// Create emulator first
	err := adapter.Create(testEmulator)
	require.NoError(t, err)

	// Update emulator
	testEmulator.Description = "Updated description"
	testEmulator.Config = map[string]any{
		"background_opacity": 0.9,
	}
	err = adapter.Update(testEmulator)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := adapter.Get("test-kitty")
	assert.NoError(t, err)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, 0.9, retrieved.Config["background_opacity"])

	// Test update non-existent emulator
	nonExistent := &emulator.Emulator{
		Name: "non-existent",
		Type: emulator.EmulatorTypeWezterm,
	}
	err = adapter.Update(nonExistent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDBEmulatorStore_Delete(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	testEmulator := &emulator.Emulator{
		Name:    "test-delete",
		Type:    emulator.EmulatorTypeITerm2,
		Enabled: true,
	}

	// Create emulator first
	err := adapter.Create(testEmulator)
	require.NoError(t, err)

	// Verify emulator exists
	exists, err := adapter.Exists("test-delete")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Delete emulator
	err = adapter.Delete("test-delete")
	assert.NoError(t, err)

	// Verify emulator is gone
	exists, err = adapter.Exists("test-delete")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test delete non-existent emulator
	err = adapter.Delete("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDBEmulatorStore_List(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	// Create multiple emulators
	emulators := []*emulator.Emulator{
		{
			Name:    "wezterm-1",
			Type:    emulator.EmulatorTypeWezterm,
			Enabled: true,
		},
		{
			Name:    "alacritty-1",
			Type:    emulator.EmulatorTypeAlacritty,
			Enabled: true,
		},
		{
			Name:    "kitty-1",
			Type:    emulator.EmulatorTypeKitty,
			Enabled: false,
		},
	}

	for _, e := range emulators {
		err := adapter.Create(e)
		require.NoError(t, err)
	}

	// List all emulators
	retrieved, err := adapter.List()
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Verify names are present
	names := make(map[string]bool)
	for _, e := range retrieved {
		names[e.Name] = true
	}
	assert.True(t, names["wezterm-1"])
	assert.True(t, names["alacritty-1"])
	assert.True(t, names["kitty-1"])
}

func TestDBEmulatorStore_ListByType(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	// Create emulators of different types
	emulators := []*emulator.Emulator{
		{Name: "wezterm-1", Type: emulator.EmulatorTypeWezterm, Enabled: true},
		{Name: "wezterm-2", Type: emulator.EmulatorTypeWezterm, Enabled: true},
		{Name: "alacritty-1", Type: emulator.EmulatorTypeAlacritty, Enabled: true},
	}

	for _, e := range emulators {
		err := adapter.Create(e)
		require.NoError(t, err)
	}

	// List WezTerm emulators
	weztermEmulators, err := adapter.ListByType(string(emulator.EmulatorTypeWezterm))
	assert.NoError(t, err)
	assert.Len(t, weztermEmulators, 2)

	// List Alacritty emulators
	alacrittyEmulators, err := adapter.ListByType(string(emulator.EmulatorTypeAlacritty))
	assert.NoError(t, err)
	assert.Len(t, alacrittyEmulators, 1)
	assert.Equal(t, "alacritty-1", alacrittyEmulators[0].Name)

	// List non-existent type
	noneEmulators, err := adapter.ListByType(string(emulator.EmulatorTypeKitty))
	assert.NoError(t, err)
	assert.Len(t, noneEmulators, 0)
}

func TestDBEmulatorStore_ListByWorkspace(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	// Create emulators for different workspaces
	emulators := []*emulator.Emulator{
		{Name: "web-wezterm", Type: emulator.EmulatorTypeWezterm, Workspace: "web-project", Enabled: true},
		{Name: "web-alacritty", Type: emulator.EmulatorTypeAlacritty, Workspace: "web-project", Enabled: true},
		{Name: "api-wezterm", Type: emulator.EmulatorTypeWezterm, Workspace: "api-project", Enabled: true},
	}

	for _, e := range emulators {
		err := adapter.Create(e)
		require.NoError(t, err)
	}

	// List emulators for web-project
	webEmulators, err := adapter.ListByWorkspace("web-project")
	assert.NoError(t, err)
	assert.Len(t, webEmulators, 2)

	// List emulators for api-project
	apiEmulators, err := adapter.ListByWorkspace("api-project")
	assert.NoError(t, err)
	assert.Len(t, apiEmulators, 1)
	assert.Equal(t, "api-wezterm", apiEmulators[0].Name)

	// List emulators for non-existent workspace
	noneEmulators, err := adapter.ListByWorkspace("non-existent")
	assert.NoError(t, err)
	assert.Len(t, noneEmulators, 0)
}

func TestDBEmulatorStore_Upsert(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	testEmulator := &emulator.Emulator{
		Name:    "test-upsert",
		Type:    emulator.EmulatorTypeWezterm,
		Enabled: true,
	}

	// Test create via upsert
	err := adapter.Upsert(testEmulator)
	assert.NoError(t, err)

	retrieved, err := adapter.Get("test-upsert")
	assert.NoError(t, err)
	assert.Equal(t, testEmulator.Name, retrieved.Name)

	// Test update via upsert
	testEmulator.Description = "Updated via upsert"
	err = adapter.Upsert(testEmulator)
	assert.NoError(t, err)

	retrieved, err = adapter.Get("test-upsert")
	assert.NoError(t, err)
	assert.Equal(t, "Updated via upsert", retrieved.Description)
}

func TestDBEmulatorStore_Exists(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()
	adapter := NewDBEmulatorStore(mockStore)

	// Test non-existent emulator
	exists, err := adapter.Exists("non-existent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create emulator
	testEmulator := &emulator.Emulator{
		Name:    "test-exists",
		Type:    emulator.EmulatorTypeWezterm,
		Enabled: true,
	}
	err = adapter.Create(testEmulator)
	require.NoError(t, err)

	// Test existing emulator
	exists, err = adapter.Exists("test-exists")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestDBEmulatorStore_Close(t *testing.T) {
	mockStore := NewMockEmulatorDataStore()

	// Test non-owned connection
	adapter := NewDBEmulatorStore(mockStore)
	err := adapter.Close()
	assert.NoError(t, err)
	assert.False(t, mockStore.closed)

	// Test owned connection
	adapterOwned := NewDBEmulatorStoreOwned(mockStore)
	err = adapterOwned.Close()
	assert.NoError(t, err)
	assert.True(t, mockStore.closed)
}

func TestEmulatorToDBModel(t *testing.T) {
	now := time.Now()
	testEmulator := &emulator.Emulator{
		Name:        "test-conversion",
		Description: "Test emulator for conversion",
		Type:        emulator.EmulatorTypeWezterm,
		Config: map[string]any{
			"font_size": 14,
			"colors": map[string]string{
				"background": "#1e1e1e",
			},
		},
		ThemeRef: "dark-theme",
		Category: "development",
		Labels: map[string]string{
			"env":     "test",
			"project": "dvm",
		},
		Workspace: "test-workspace",
		Enabled:   true,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	dbModel := emulatorToDBModel(testEmulator)

	assert.Equal(t, testEmulator.Name, dbModel.Name)
	assert.Equal(t, string(testEmulator.Type), dbModel.Type)
	assert.True(t, dbModel.Description.Valid)
	assert.Equal(t, testEmulator.Description, dbModel.Description.String)
	assert.True(t, dbModel.ThemeRef.Valid)
	assert.Equal(t, testEmulator.ThemeRef, dbModel.ThemeRef.String)
	assert.True(t, dbModel.Category.Valid)
	assert.Equal(t, testEmulator.Category, dbModel.Category.String)
	assert.True(t, dbModel.Workspace.Valid)
	assert.Equal(t, testEmulator.Workspace, dbModel.Workspace.String)
	assert.Equal(t, testEmulator.Enabled, dbModel.Enabled)

	// Check JSON fields are not empty
	assert.NotEmpty(t, dbModel.Config)
	assert.NotEqual(t, "{}", dbModel.Config)
	assert.NotEmpty(t, dbModel.Labels)
	assert.NotEqual(t, "{}", dbModel.Labels)

	assert.Equal(t, now, dbModel.CreatedAt)
	assert.Equal(t, now, dbModel.UpdatedAt)
}

func TestDBModelToEmulator(t *testing.T) {
	now := time.Now()
	dbModel := &models.TerminalEmulatorDB{
		Name:      "test-conversion",
		Type:      string(emulator.EmulatorTypeAlacritty),
		Config:    `{"font_size": 16, "window": {"opacity": 0.8}}`,
		Labels:    `{"env": "prod", "team": "backend"}`,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	dbModel.Description.String = "Test conversion"
	dbModel.Description.Valid = true
	dbModel.ThemeRef.String = "light-theme"
	dbModel.ThemeRef.Valid = true
	dbModel.Category.String = "production"
	dbModel.Category.Valid = true
	dbModel.Workspace.String = "prod-workspace"
	dbModel.Workspace.Valid = true

	emulatorObj := dbModelToEmulator(dbModel)

	assert.Equal(t, dbModel.Name, emulatorObj.Name)
	assert.Equal(t, emulator.EmulatorTypeAlacritty, emulatorObj.Type)
	assert.Equal(t, "Test conversion", emulatorObj.Description)
	assert.Equal(t, "light-theme", emulatorObj.ThemeRef)
	assert.Equal(t, "production", emulatorObj.Category)
	assert.Equal(t, "prod-workspace", emulatorObj.Workspace)
	assert.Equal(t, dbModel.Enabled, emulatorObj.Enabled)

	// Check JSON parsing
	assert.NotNil(t, emulatorObj.Config)
	assert.Equal(t, float64(16), emulatorObj.Config["font_size"]) // JSON unmarshals numbers as float64
	assert.NotNil(t, emulatorObj.Labels)
	assert.Equal(t, "prod", emulatorObj.Labels["env"])
	assert.Equal(t, "backend", emulatorObj.Labels["team"])

	assert.NotNil(t, emulatorObj.CreatedAt)
	assert.Equal(t, now, *emulatorObj.CreatedAt)
	assert.NotNil(t, emulatorObj.UpdatedAt)
	assert.Equal(t, now, *emulatorObj.UpdatedAt)
}

func TestJSONHandling(t *testing.T) {
	// Test empty config and labels
	emptyEmulator := &emulator.Emulator{
		Name:    "empty-test",
		Type:    emulator.EmulatorTypeWezterm,
		Enabled: true,
	}

	dbModel := emulatorToDBModel(emptyEmulator)
	assert.Equal(t, "{}", dbModel.Config)
	assert.Equal(t, "{}", dbModel.Labels)

	converted := dbModelToEmulator(dbModel)
	assert.NotNil(t, converted.Config)
	assert.Len(t, converted.Config, 0)
	assert.NotNil(t, converted.Labels)
	assert.Len(t, converted.Labels, 0)

	// Test invalid JSON handling
	dbModelInvalid := &models.TerminalEmulatorDB{
		Name:    "invalid-json",
		Type:    string(emulator.EmulatorTypeWezterm),
		Config:  `{invalid json`,
		Labels:  `{also invalid`,
		Enabled: true,
	}

	converted = dbModelToEmulator(dbModelInvalid)
	assert.NotNil(t, converted.Config)
	assert.Len(t, converted.Config, 0) // Should fallback to empty map
	assert.NotNil(t, converted.Labels)
	assert.Len(t, converted.Labels, 0) // Should fallback to empty map
}
