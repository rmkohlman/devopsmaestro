package theme

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"devopsmaestro/models"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for integration tests
)

// =============================================================================
// Mock Implementation of ThemeDataStore
// =============================================================================

// MockThemeDataStore implements ThemeDataStore for testing.
type MockThemeDataStore struct {
	themes      map[string]*models.NvimThemeDB
	activeTheme string
	calls       []string
	err         error // inject errors for testing
}

func NewMockThemeDataStore() *MockThemeDataStore {
	return &MockThemeDataStore{
		themes: make(map[string]*models.NvimThemeDB),
	}
}

func (m *MockThemeDataStore) CreateTheme(t *models.NvimThemeDB) error {
	m.calls = append(m.calls, "CreateTheme")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.themes[t.Name]; exists {
		return fmt.Errorf("theme already exists: %s", t.Name)
	}
	t.ID = len(m.themes) + 1
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	m.themes[t.Name] = t
	return nil
}

func (m *MockThemeDataStore) GetThemeByName(name string) (*models.NvimThemeDB, error) {
	m.calls = append(m.calls, "GetThemeByName")
	if m.err != nil {
		return nil, m.err
	}
	t, ok := m.themes[name]
	if !ok {
		return nil, fmt.Errorf("theme not found: %s", name)
	}
	return t, nil
}

func (m *MockThemeDataStore) UpdateTheme(t *models.NvimThemeDB) error {
	m.calls = append(m.calls, "UpdateTheme")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.themes[t.Name]; !exists {
		return fmt.Errorf("theme not found: %s", t.Name)
	}
	t.UpdatedAt = time.Now()
	m.themes[t.Name] = t
	return nil
}

func (m *MockThemeDataStore) DeleteTheme(name string) error {
	m.calls = append(m.calls, "DeleteTheme")
	if m.err != nil {
		return m.err
	}
	delete(m.themes, name)
	if m.activeTheme == name {
		m.activeTheme = ""
	}
	return nil
}

func (m *MockThemeDataStore) ListThemes() ([]*models.NvimThemeDB, error) {
	m.calls = append(m.calls, "ListThemes")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.NvimThemeDB
	for _, t := range m.themes {
		result = append(result, t)
	}
	return result, nil
}

func (m *MockThemeDataStore) ListThemesByCategory(category string) ([]*models.NvimThemeDB, error) {
	m.calls = append(m.calls, "ListThemesByCategory")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.NvimThemeDB
	for _, t := range m.themes {
		if t.Category.Valid && t.Category.String == category {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockThemeDataStore) GetActiveTheme() (*models.NvimThemeDB, error) {
	m.calls = append(m.calls, "GetActiveTheme")
	if m.err != nil {
		return nil, m.err
	}
	if m.activeTheme == "" {
		return nil, nil
	}
	t, ok := m.themes[m.activeTheme]
	if !ok {
		return nil, fmt.Errorf("active theme not found: %s", m.activeTheme)
	}
	return t, nil
}

func (m *MockThemeDataStore) SetActiveTheme(name string) error {
	m.calls = append(m.calls, "SetActiveTheme")
	if m.err != nil {
		return m.err
	}
	if _, ok := m.themes[name]; !ok {
		return fmt.Errorf("theme not found: %s", name)
	}
	m.activeTheme = name
	return nil
}

func (m *MockThemeDataStore) ClearActiveTheme() error {
	m.calls = append(m.calls, "ClearActiveTheme")
	if m.err != nil {
		return m.err
	}
	m.activeTheme = ""
	return nil
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

// TestDBStoreAdapter_ImplementsStore verifies that DBStoreAdapter
// satisfies the Store interface at compile time.
func TestDBStoreAdapter_ImplementsStore(t *testing.T) {
	var _ Store = (*DBStoreAdapter)(nil)
}

// TestMockThemeDataStore_ImplementsThemeDataStore verifies the mock
// satisfies the ThemeDataStore interface.
func TestMockThemeDataStore_ImplementsThemeDataStore(t *testing.T) {
	var _ ThemeDataStore = (*MockThemeDataStore)(nil)
}

// =============================================================================
// Direct Struct Method Tests
// =============================================================================

func TestDBStoreAdapter_Save(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	theme := &Theme{
		Name:        "catppuccin",
		Description: "Soothing pastel theme",
		Category:    "dark",
		Plugin: ThemePlugin{
			Repo: "catppuccin/nvim",
		},
	}

	// Test successful save (create)
	err := adapter.Save(theme)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify theme was stored
	if len(mock.themes) != 1 {
		t.Errorf("Expected 1 theme, got %d", len(mock.themes))
	}

	// Verify calls (GetThemeByName to check existence + CreateTheme)
	if len(mock.calls) != 2 {
		t.Errorf("Expected 2 calls, got %d: %v", len(mock.calls), mock.calls)
	}

	// Test save again (should update)
	theme.Description = "Updated description"
	mock.calls = nil // Reset calls
	err = adapter.Save(theme)
	if err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	// Should still be 1 theme
	if len(mock.themes) != 1 {
		t.Errorf("Expected 1 theme after update, got %d", len(mock.themes))
	}
}

func TestDBStoreAdapter_Get(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Add a theme directly to mock
	mock.themes["catppuccin"] = &models.NvimThemeDB{
		ID:          1,
		Name:        "catppuccin",
		PluginRepo:  "catppuccin/nvim",
		Description: sql.NullString{String: "Soothing pastel theme", Valid: true},
		Category:    sql.NullString{String: "dark", Valid: true},
	}

	// Test successful get
	theme, err := adapter.Get("catppuccin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if theme.Name != "catppuccin" {
		t.Errorf("Expected name 'catppuccin', got '%s'", theme.Name)
	}
	if theme.Description != "Soothing pastel theme" {
		t.Errorf("Expected description 'Soothing pastel theme', got '%s'", theme.Description)
	}

	// Test get non-existent
	_, err = adapter.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent theme")
	}
}

func TestDBStoreAdapter_Delete(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Add a theme
	mock.themes["catppuccin"] = &models.NvimThemeDB{
		ID:         1,
		Name:       "catppuccin",
		PluginRepo: "catppuccin/nvim",
	}

	// Test successful delete
	err := adapter.Delete("catppuccin")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify theme was deleted from mock
	if len(mock.themes) != 0 {
		t.Errorf("Expected 0 themes after delete, got %d", len(mock.themes))
	}

	// Test delete non-existent (should check existence first)
	err = adapter.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent theme")
	}
}

func TestDBStoreAdapter_List(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Empty list
	themes, err := adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(themes) != 0 {
		t.Errorf("Expected 0 themes, got %d", len(themes))
	}

	// Add themes
	mock.themes["catppuccin"] = &models.NvimThemeDB{
		ID: 1, Name: "catppuccin", PluginRepo: "catppuccin/nvim",
	}
	mock.themes["tokyonight"] = &models.NvimThemeDB{
		ID: 2, Name: "tokyonight", PluginRepo: "folke/tokyonight.nvim",
	}

	themes, err = adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(themes) != 2 {
		t.Errorf("Expected 2 themes, got %d", len(themes))
	}
}

func TestDBStoreAdapter_GetActive(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	// No active theme
	active, err := adapter.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active != nil {
		t.Error("Expected nil for no active theme")
	}

	// Add a theme and set as active
	mock.themes["catppuccin"] = &models.NvimThemeDB{
		ID:         1,
		Name:       "catppuccin",
		PluginRepo: "catppuccin/nvim",
	}
	mock.activeTheme = "catppuccin"

	active, err = adapter.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active == nil {
		t.Fatal("Expected active theme, got nil")
	}
	if active.Name != "catppuccin" {
		t.Errorf("Expected active theme 'catppuccin', got '%s'", active.Name)
	}
}

func TestDBStoreAdapter_SetActive(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Add a theme
	mock.themes["catppuccin"] = &models.NvimThemeDB{
		ID:         1,
		Name:       "catppuccin",
		PluginRepo: "catppuccin/nvim",
	}

	// Set active
	err := adapter.SetActive("catppuccin")
	if err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	// Verify active was set in mock
	if mock.activeTheme != "catppuccin" {
		t.Errorf("Expected active theme 'catppuccin', got '%s'", mock.activeTheme)
	}

	// Test set active for non-existent theme
	err = adapter.SetActive("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent theme")
	}
}

func TestDBStoreAdapter_Path(t *testing.T) {
	mock := NewMockThemeDataStore()
	adapter := NewDBStoreAdapter(mock)

	// DB adapter should return empty path
	if adapter.Path() != "" {
		t.Errorf("Expected empty path, got '%s'", adapter.Path())
	}
}

// =============================================================================
// Model Conversion Tests
// =============================================================================

func TestThemeToDBModel_BasicFields(t *testing.T) {
	theme := &Theme{
		Name:        "catppuccin",
		Description: "Soothing pastel theme",
		Author:      "catppuccin",
		Category:    "dark",
		Plugin: ThemePlugin{
			Repo:   "catppuccin/nvim",
			Branch: "main",
			Tag:    "v1.0.0",
		},
		Style:       "mocha",
		Transparent: true,
	}

	db := themeToDBModel(theme)

	if db.Name != "catppuccin" {
		t.Errorf("Name mismatch: %s", db.Name)
	}
	if db.Description.String != "Soothing pastel theme" || !db.Description.Valid {
		t.Errorf("Description mismatch: %v", db.Description)
	}
	if db.Author.String != "catppuccin" || !db.Author.Valid {
		t.Errorf("Author mismatch: %v", db.Author)
	}
	if db.Category.String != "dark" || !db.Category.Valid {
		t.Errorf("Category mismatch: %v", db.Category)
	}
	if db.PluginRepo != "catppuccin/nvim" {
		t.Errorf("PluginRepo mismatch: %s", db.PluginRepo)
	}
	if db.PluginBranch.String != "main" || !db.PluginBranch.Valid {
		t.Errorf("PluginBranch mismatch: %v", db.PluginBranch)
	}
	if db.PluginTag.String != "v1.0.0" || !db.PluginTag.Valid {
		t.Errorf("PluginTag mismatch: %v", db.PluginTag)
	}
	if db.Style.String != "mocha" || !db.Style.Valid {
		t.Errorf("Style mismatch: %v", db.Style)
	}
	if !db.Transparent {
		t.Error("Transparent should be true")
	}
}

func TestThemeToDBModel_JSONFields(t *testing.T) {
	theme := &Theme{
		Name: "catppuccin",
		Plugin: ThemePlugin{
			Repo: "catppuccin/nvim",
		},
		Colors: map[string]string{
			"bg":      "#1e1e2e",
			"fg":      "#cdd6f4",
			"red":     "#f38ba8",
			"green":   "#a6e3a1",
			"blue":    "#89b4fa",
			"yellow":  "#f9e2af",
			"magenta": "#f5c2e7",
			"cyan":    "#94e2d5",
		},
		Options: map[string]any{
			"transparent_background": true,
			"integrations": map[string]any{
				"telescope": true,
				"nvimtree":  true,
			},
		},
	}

	db := themeToDBModel(theme)

	if !db.Colors.Valid || db.Colors.String == "" {
		t.Error("Colors should be valid JSON")
	}
	if !db.Options.Valid || db.Options.String == "" {
		t.Error("Options should be valid JSON")
	}
}

func TestDBModelToTheme_BasicFields(t *testing.T) {
	db := &models.NvimThemeDB{
		ID:           1,
		Name:         "catppuccin",
		Description:  sql.NullString{String: "Soothing pastel theme", Valid: true},
		Author:       sql.NullString{String: "catppuccin", Valid: true},
		Category:     sql.NullString{String: "dark", Valid: true},
		PluginRepo:   "catppuccin/nvim",
		PluginBranch: sql.NullString{String: "main", Valid: true},
		PluginTag:    sql.NullString{String: "v1.0.0", Valid: true},
		Style:        sql.NullString{String: "mocha", Valid: true},
		Transparent:  true,
	}

	theme := dbModelToTheme(db)

	if theme.Name != "catppuccin" {
		t.Errorf("Name mismatch: %s", theme.Name)
	}
	if theme.Description != "Soothing pastel theme" {
		t.Errorf("Description mismatch: %s", theme.Description)
	}
	if theme.Author != "catppuccin" {
		t.Errorf("Author mismatch: %s", theme.Author)
	}
	if theme.Category != "dark" {
		t.Errorf("Category mismatch: %s", theme.Category)
	}
	if theme.Plugin.Repo != "catppuccin/nvim" {
		t.Errorf("Plugin.Repo mismatch: %s", theme.Plugin.Repo)
	}
	if theme.Plugin.Branch != "main" {
		t.Errorf("Plugin.Branch mismatch: %s", theme.Plugin.Branch)
	}
	if theme.Plugin.Tag != "v1.0.0" {
		t.Errorf("Plugin.Tag mismatch: %s", theme.Plugin.Tag)
	}
	if theme.Style != "mocha" {
		t.Errorf("Style mismatch: %s", theme.Style)
	}
	if !theme.Transparent {
		t.Error("Transparent should be true")
	}
}

func TestDBModelToTheme_JSONFields(t *testing.T) {
	db := &models.NvimThemeDB{
		Name:       "catppuccin",
		PluginRepo: "catppuccin/nvim",
		Colors:     sql.NullString{String: `{"bg":"#1e1e2e","fg":"#cdd6f4"}`, Valid: true},
		Options:    sql.NullString{String: `{"transparent_background":true}`, Valid: true},
	}

	theme := dbModelToTheme(db)

	if len(theme.Colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(theme.Colors))
	}
	if theme.Colors["bg"] != "#1e1e2e" {
		t.Errorf("Colors[bg] mismatch: %s", theme.Colors["bg"])
	}
	if len(theme.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(theme.Options))
	}
}

func TestConversion_Roundtrip(t *testing.T) {
	// Create a fully-populated theme
	original := &Theme{
		Name:        "catppuccin",
		Description: "Soothing pastel theme for Neovim",
		Author:      "catppuccin",
		Category:    "dark",
		Plugin: ThemePlugin{
			Repo:   "catppuccin/nvim",
			Branch: "main",
			Tag:    "v1.0.0",
		},
		Style:       "mocha",
		Transparent: true,
		Colors: map[string]string{
			"bg": "#1e1e2e",
			"fg": "#cdd6f4",
		},
		Options: map[string]any{
			"transparent_background": true,
		},
	}

	// Convert to DB model and back
	db := themeToDBModel(original)
	converted := dbModelToTheme(db)

	// Verify all fields match
	if converted.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", converted.Name, original.Name)
	}
	if converted.Description != original.Description {
		t.Errorf("Description mismatch: got %s, want %s", converted.Description, original.Description)
	}
	if converted.Author != original.Author {
		t.Errorf("Author mismatch: got %s, want %s", converted.Author, original.Author)
	}
	if converted.Category != original.Category {
		t.Errorf("Category mismatch: got %s, want %s", converted.Category, original.Category)
	}
	if converted.Plugin.Repo != original.Plugin.Repo {
		t.Errorf("Plugin.Repo mismatch: got %s, want %s", converted.Plugin.Repo, original.Plugin.Repo)
	}
	if converted.Plugin.Branch != original.Plugin.Branch {
		t.Errorf("Plugin.Branch mismatch: got %s, want %s", converted.Plugin.Branch, original.Plugin.Branch)
	}
	if converted.Plugin.Tag != original.Plugin.Tag {
		t.Errorf("Plugin.Tag mismatch: got %s, want %s", converted.Plugin.Tag, original.Plugin.Tag)
	}
	if converted.Style != original.Style {
		t.Errorf("Style mismatch: got %s, want %s", converted.Style, original.Style)
	}
	if converted.Transparent != original.Transparent {
		t.Errorf("Transparent mismatch: got %v, want %v", converted.Transparent, original.Transparent)
	}
	if len(converted.Colors) != len(original.Colors) {
		t.Errorf("Colors length mismatch: got %d, want %d", len(converted.Colors), len(original.Colors))
	}
	if len(converted.Options) != len(original.Options) {
		t.Errorf("Options length mismatch: got %d, want %d", len(converted.Options), len(original.Options))
	}
}

// =============================================================================
// Through Interface Tests (Store interface)
// =============================================================================

// TestStore_ThroughInterface tests the DBStoreAdapter through
// the Store interface to ensure polymorphism works correctly.
func TestStore_ThroughInterface(t *testing.T) {
	mock := NewMockThemeDataStore()

	// Use through interface
	var store Store = NewDBStoreAdapter(mock)

	// Test Save through interface
	theme := &Theme{
		Name:        "catppuccin",
		Description: "Soothing pastel theme",
		Category:    "dark",
		Plugin: ThemePlugin{
			Repo: "catppuccin/nvim",
		},
	}

	err := store.Save(theme)
	if err != nil {
		t.Fatalf("Save through interface failed: %v", err)
	}

	// Test Get through interface
	retrieved, err := store.Get("catppuccin")
	if err != nil {
		t.Fatalf("Get through interface failed: %v", err)
	}
	if retrieved.Name != "catppuccin" {
		t.Errorf("Expected name catppuccin, got %s", retrieved.Name)
	}

	// Test List through interface
	themes, err := store.List()
	if err != nil {
		t.Fatalf("List through interface failed: %v", err)
	}
	if len(themes) != 1 {
		t.Errorf("Expected 1 theme, got %d", len(themes))
	}

	// Test SetActive through interface
	err = store.SetActive("catppuccin")
	if err != nil {
		t.Fatalf("SetActive through interface failed: %v", err)
	}

	// Test GetActive through interface
	active, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive through interface failed: %v", err)
	}
	if active == nil || active.Name != "catppuccin" {
		t.Errorf("Expected active theme catppuccin, got %v", active)
	}

	// Test Delete through interface
	err = store.Delete("catppuccin")
	if err != nil {
		t.Fatalf("Delete through interface failed: %v", err)
	}

	// Test Path through interface
	path := store.Path()
	if path != "" {
		t.Errorf("Expected empty path, got %s", path)
	}
}

// TestStore_Swappability verifies that the adapter can be swapped
// with other Store implementations (like MemoryStore).
func TestStore_Swappability(t *testing.T) {
	// Create a test function that works with any Store
	testStore := func(store Store, name string) {
		theme := &Theme{
			Name: "test-theme",
			Plugin: ThemePlugin{
				Repo: "test/repo",
			},
		}

		// Save
		if err := store.Save(theme); err != nil {
			t.Errorf("%s: Save failed: %v", name, err)
			return
		}

		// Get
		retrieved, err := store.Get("test-theme")
		if err != nil {
			t.Errorf("%s: Get failed: %v", name, err)
			return
		}
		if retrieved.Name != "test-theme" {
			t.Errorf("%s: Name mismatch", name)
		}

		// SetActive
		if err := store.SetActive("test-theme"); err != nil {
			t.Errorf("%s: SetActive failed: %v", name, err)
			return
		}

		// GetActive
		active, err := store.GetActive()
		if err != nil {
			t.Errorf("%s: GetActive failed: %v", name, err)
			return
		}
		if active == nil || active.Name != "test-theme" {
			t.Errorf("%s: Active theme mismatch", name)
		}

		// Delete
		if err := store.Delete("test-theme"); err != nil {
			t.Errorf("%s: Delete failed: %v", name, err)
		}
	}

	// Test with DBStoreAdapter
	mock := NewMockThemeDataStore()
	testStore(NewDBStoreAdapter(mock), "DBStoreAdapter")

	// Test with MemoryStore
	memStore := NewMemoryStore()
	testStore(memStore, "MemoryStore")
}

// =============================================================================
// Integration Tests - Real In-Memory SQLite Database
// =============================================================================

// createIntegrationTestDataStore creates a real in-memory SQLite DataStore for testing
func createIntegrationTestDataStore(t *testing.T) *integrationTestDB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite: %v", err)
	}

	// Create the nvim_themes table (matching the real schema)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS nvim_themes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			author TEXT,
			category TEXT,
			plugin_repo TEXT NOT NULL,
			plugin_branch TEXT,
			plugin_tag TEXT,
			style TEXT,
			colors TEXT,
			options TEXT,
			transparent INTEGER DEFAULT 0,
			is_active INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create nvim_themes table: %v", err)
	}

	return &integrationTestDB{db: db, t: t}
}

// integrationTestDB wraps sql.DB and implements ThemeDataStore for integration testing
type integrationTestDB struct {
	db *sql.DB
	t  *testing.T
}

func (d *integrationTestDB) CreateTheme(t *models.NvimThemeDB) error {
	result, err := d.db.Exec(`
		INSERT INTO nvim_themes (
			name, description, author, category, plugin_repo, plugin_branch,
			plugin_tag, style, colors, options, transparent
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		t.Name, t.Description, t.Author, t.Category, t.PluginRepo,
		t.PluginBranch, t.PluginTag, t.Style, t.Colors, t.Options, t.Transparent,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	t.ID = int(id)
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (d *integrationTestDB) GetThemeByName(name string) (*models.NvimThemeDB, error) {
	t := &models.NvimThemeDB{}
	err := d.db.QueryRow(`
		SELECT id, name, description, author, category, plugin_repo, plugin_branch,
			   plugin_tag, style, colors, options, transparent, is_active, created_at, updated_at
		FROM nvim_themes WHERE name = ?
	`, name).Scan(
		&t.ID, &t.Name, &t.Description, &t.Author, &t.Category, &t.PluginRepo,
		&t.PluginBranch, &t.PluginTag, &t.Style, &t.Colors, &t.Options,
		&t.Transparent, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("theme not found: %s", name)
	}
	return t, err
}

func (d *integrationTestDB) UpdateTheme(t *models.NvimThemeDB) error {
	_, err := d.db.Exec(`
		UPDATE nvim_themes SET
			description = ?, author = ?, category = ?, plugin_repo = ?,
			plugin_branch = ?, plugin_tag = ?, style = ?, colors = ?,
			options = ?, transparent = ?, updated_at = CURRENT_TIMESTAMP
		WHERE name = ?
	`,
		t.Description, t.Author, t.Category, t.PluginRepo,
		t.PluginBranch, t.PluginTag, t.Style, t.Colors,
		t.Options, t.Transparent, t.Name,
	)
	return err
}

func (d *integrationTestDB) DeleteTheme(name string) error {
	_, err := d.db.Exec(`DELETE FROM nvim_themes WHERE name = ?`, name)
	return err
}

func (d *integrationTestDB) ListThemes() ([]*models.NvimThemeDB, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, author, category, plugin_repo, plugin_branch,
			   plugin_tag, style, colors, options, transparent, is_active, created_at, updated_at
		FROM nvim_themes ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []*models.NvimThemeDB
	for rows.Next() {
		t := &models.NvimThemeDB{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.Author, &t.Category, &t.PluginRepo,
			&t.PluginBranch, &t.PluginTag, &t.Style, &t.Colors, &t.Options,
			&t.Transparent, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	return themes, nil
}

func (d *integrationTestDB) ListThemesByCategory(category string) ([]*models.NvimThemeDB, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, author, category, plugin_repo, plugin_branch,
			   plugin_tag, style, colors, options, transparent, is_active, created_at, updated_at
		FROM nvim_themes WHERE category = ? ORDER BY name
	`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []*models.NvimThemeDB
	for rows.Next() {
		t := &models.NvimThemeDB{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.Author, &t.Category, &t.PluginRepo,
			&t.PluginBranch, &t.PluginTag, &t.Style, &t.Colors, &t.Options,
			&t.Transparent, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	return themes, nil
}

func (d *integrationTestDB) GetActiveTheme() (*models.NvimThemeDB, error) {
	t := &models.NvimThemeDB{}
	err := d.db.QueryRow(`
		SELECT id, name, description, author, category, plugin_repo, plugin_branch,
			   plugin_tag, style, colors, options, transparent, is_active, created_at, updated_at
		FROM nvim_themes WHERE is_active = 1
	`).Scan(
		&t.ID, &t.Name, &t.Description, &t.Author, &t.Category, &t.PluginRepo,
		&t.PluginBranch, &t.PluginTag, &t.Style, &t.Colors, &t.Options,
		&t.Transparent, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (d *integrationTestDB) SetActiveTheme(name string) error {
	// First clear any existing active theme
	_, err := d.db.Exec(`UPDATE nvim_themes SET is_active = 0`)
	if err != nil {
		return err
	}
	// Set the new active theme
	result, err := d.db.Exec(`UPDATE nvim_themes SET is_active = 1 WHERE name = ?`, name)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("theme not found: %s", name)
	}
	return nil
}

func (d *integrationTestDB) ClearActiveTheme() error {
	_, err := d.db.Exec(`UPDATE nvim_themes SET is_active = 0`)
	return err
}

func (d *integrationTestDB) Close() error {
	return d.db.Close()
}

// rawQuery executes a raw SQL query and returns the result for verification
func (d *integrationTestDB) rawQuery(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// =============================================================================
// Integration Test: Write Theme, Verify with Raw SQL
// =============================================================================

func TestIntegration_CreateTheme_VerifyWithSQL(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create theme through adapter
	theme := &Theme{
		Name:        "catppuccin",
		Description: "Soothing pastel theme",
		Author:      "catppuccin",
		Category:    "dark",
		Plugin: ThemePlugin{
			Repo:   "catppuccin/nvim",
			Branch: "main",
		},
		Style:       "mocha",
		Transparent: true,
		Colors: map[string]string{
			"bg": "#1e1e2e",
			"fg": "#cdd6f4",
		},
	}

	err := adapter.Save(theme)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify with raw SQL query
	var (
		name         string
		description  sql.NullString
		author       sql.NullString
		category     sql.NullString
		pluginRepo   string
		pluginBranch sql.NullString
		style        sql.NullString
		transparent  bool
		colors       sql.NullString
	)

	err = testDB.rawQuery(`
		SELECT name, description, author, category, plugin_repo, plugin_branch, style, transparent, colors
		FROM nvim_themes WHERE name = ?
	`, "catppuccin").Scan(&name, &description, &author, &category, &pluginRepo, &pluginBranch, &style, &transparent, &colors)

	if err != nil {
		t.Fatalf("Raw SQL query failed: %v", err)
	}

	// Verify each field
	if name != "catppuccin" {
		t.Errorf("SQL name = %q, want %q", name, "catppuccin")
	}
	if !description.Valid || description.String != "Soothing pastel theme" {
		t.Errorf("SQL description = %v, want %q", description, "Soothing pastel theme")
	}
	if !author.Valid || author.String != "catppuccin" {
		t.Errorf("SQL author = %v, want %q", author, "catppuccin")
	}
	if !category.Valid || category.String != "dark" {
		t.Errorf("SQL category = %v, want %q", category, "dark")
	}
	if pluginRepo != "catppuccin/nvim" {
		t.Errorf("SQL plugin_repo = %q, want %q", pluginRepo, "catppuccin/nvim")
	}
	if !pluginBranch.Valid || pluginBranch.String != "main" {
		t.Errorf("SQL plugin_branch = %v, want %q", pluginBranch, "main")
	}
	if !style.Valid || style.String != "mocha" {
		t.Errorf("SQL style = %v, want %q", style, "mocha")
	}
	if !transparent {
		t.Error("SQL transparent should be true")
	}
	if !colors.Valid || colors.String == "" {
		t.Error("SQL colors should be valid JSON")
	}
}

// =============================================================================
// Integration Test: Full CRUD Workflow with SQL Verification
// =============================================================================

func TestIntegration_FullCRUDWorkflow(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// --- CREATE ---
	theme := &Theme{
		Name:     "tokyonight",
		Category: "dark",
		Plugin: ThemePlugin{
			Repo: "folke/tokyonight.nvim",
		},
	}

	err := adapter.Save(theme)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify count via SQL
	var count int
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_themes`).Scan(&count)
	if count != 1 {
		t.Errorf("After CREATE: count = %d, want 1", count)
	}

	// --- READ ---
	retrieved, err := adapter.Get("tokyonight")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != "tokyonight" {
		t.Errorf("Get name = %q, want %q", retrieved.Name, "tokyonight")
	}
	if retrieved.Category != "dark" {
		t.Errorf("Get category = %q, want %q", retrieved.Category, "dark")
	}

	// --- UPDATE ---
	retrieved.Description = "Updated description"
	retrieved.Transparent = true
	err = adapter.Save(retrieved)
	if err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	// Verify update via SQL
	var desc sql.NullString
	var transparent bool
	testDB.rawQuery(`SELECT description, transparent FROM nvim_themes WHERE name = ?`, "tokyonight").Scan(&desc, &transparent)
	if !desc.Valid || desc.String != "Updated description" {
		t.Errorf("After UPDATE: description = %v, want %q", desc, "Updated description")
	}
	if !transparent {
		t.Error("After UPDATE: transparent should be true")
	}

	// --- DELETE ---
	err = adapter.Delete("tokyonight")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion via SQL
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_themes`).Scan(&count)
	if count != 0 {
		t.Errorf("After DELETE: count = %d, want 0", count)
	}
}

// =============================================================================
// Integration Test: Active Theme Operations
// =============================================================================

func TestIntegration_ActiveThemeOperations(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create two themes
	themes := []*Theme{
		{Name: "catppuccin", Category: "dark", Plugin: ThemePlugin{Repo: "catppuccin/nvim"}},
		{Name: "tokyonight", Category: "dark", Plugin: ThemePlugin{Repo: "folke/tokyonight.nvim"}},
	}

	for _, theme := range themes {
		if err := adapter.Save(theme); err != nil {
			t.Fatalf("Save %s failed: %v", theme.Name, err)
		}
	}

	// No active theme initially
	active, err := adapter.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active != nil {
		t.Errorf("Expected no active theme, got %s", active.Name)
	}

	// Set catppuccin as active
	err = adapter.SetActive("catppuccin")
	if err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	active, err = adapter.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active == nil || active.Name != "catppuccin" {
		t.Errorf("Expected active theme catppuccin, got %v", active)
	}

	// Verify via SQL
	var isActive bool
	testDB.rawQuery(`SELECT is_active FROM nvim_themes WHERE name = ?`, "catppuccin").Scan(&isActive)
	if !isActive {
		t.Error("SQL is_active should be true for catppuccin")
	}

	// Switch to tokyonight
	err = adapter.SetActive("tokyonight")
	if err != nil {
		t.Fatalf("SetActive (switch) failed: %v", err)
	}

	// Verify catppuccin is no longer active
	testDB.rawQuery(`SELECT is_active FROM nvim_themes WHERE name = ?`, "catppuccin").Scan(&isActive)
	if isActive {
		t.Error("SQL is_active should be false for catppuccin after switch")
	}

	// Verify tokyonight is now active
	testDB.rawQuery(`SELECT is_active FROM nvim_themes WHERE name = ?`, "tokyonight").Scan(&isActive)
	if !isActive {
		t.Error("SQL is_active should be true for tokyonight")
	}
}

// =============================================================================
// Integration Test: List Operations
// =============================================================================

func TestIntegration_ListOperations(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create multiple themes
	themes := []*Theme{
		{Name: "catppuccin", Category: "dark", Plugin: ThemePlugin{Repo: "catppuccin/nvim"}},
		{Name: "tokyonight", Category: "dark", Plugin: ThemePlugin{Repo: "folke/tokyonight.nvim"}},
		{Name: "solarized", Category: "light", Plugin: ThemePlugin{Repo: "maxmx03/solarized.nvim"}},
		{Name: "gruvbox", Category: "dark", Plugin: ThemePlugin{Repo: "ellisonleao/gruvbox.nvim"}},
	}

	for _, theme := range themes {
		if err := adapter.Save(theme); err != nil {
			t.Fatalf("Save %s failed: %v", theme.Name, err)
		}
	}

	// Verify count via SQL
	var count int
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_themes`).Scan(&count)
	if count != 4 {
		t.Errorf("SQL count = %d, want 4", count)
	}

	// Test List
	all, err := adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 4 {
		t.Errorf("List returned %d themes, want 4", len(all))
	}
}

// =============================================================================
// Integration Test: Complex Theme Roundtrip
// =============================================================================

func TestIntegration_ComplexThemeRoundtrip(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create a fully-populated theme
	original := &Theme{
		Name:        "catppuccin",
		Description: "Soothing pastel theme for the high-spirited!",
		Author:      "catppuccin",
		Category:    "dark",
		Plugin: ThemePlugin{
			Repo:   "catppuccin/nvim",
			Branch: "main",
			Tag:    "v1.7.0",
		},
		Style:       "mocha",
		Transparent: true,
		Colors: map[string]string{
			"bg":     "#1e1e2e",
			"fg":     "#cdd6f4",
			"red":    "#f38ba8",
			"green":  "#a6e3a1",
			"blue":   "#89b4fa",
			"yellow": "#f9e2af",
		},
		Options: map[string]any{
			"transparent_background": true,
			"integrations": map[string]any{
				"telescope": true,
				"nvimtree":  true,
			},
		},
	}

	// Save
	err := adapter.Save(original)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read back
	retrieved, err := adapter.Get("catppuccin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Verify all fields match
	if retrieved.Name != original.Name {
		t.Errorf("Name: got %q, want %q", retrieved.Name, original.Name)
	}
	if retrieved.Description != original.Description {
		t.Errorf("Description: got %q, want %q", retrieved.Description, original.Description)
	}
	if retrieved.Author != original.Author {
		t.Errorf("Author: got %q, want %q", retrieved.Author, original.Author)
	}
	if retrieved.Category != original.Category {
		t.Errorf("Category: got %q, want %q", retrieved.Category, original.Category)
	}
	if retrieved.Plugin.Repo != original.Plugin.Repo {
		t.Errorf("Plugin.Repo: got %q, want %q", retrieved.Plugin.Repo, original.Plugin.Repo)
	}
	if retrieved.Plugin.Branch != original.Plugin.Branch {
		t.Errorf("Plugin.Branch: got %q, want %q", retrieved.Plugin.Branch, original.Plugin.Branch)
	}
	if retrieved.Plugin.Tag != original.Plugin.Tag {
		t.Errorf("Plugin.Tag: got %q, want %q", retrieved.Plugin.Tag, original.Plugin.Tag)
	}
	if retrieved.Style != original.Style {
		t.Errorf("Style: got %q, want %q", retrieved.Style, original.Style)
	}
	if retrieved.Transparent != original.Transparent {
		t.Errorf("Transparent: got %v, want %v", retrieved.Transparent, original.Transparent)
	}
	if len(retrieved.Colors) != len(original.Colors) {
		t.Errorf("Colors length: got %d, want %d", len(retrieved.Colors), len(original.Colors))
	}
	if len(retrieved.Options) != len(original.Options) {
		t.Errorf("Options length: got %d, want %d", len(retrieved.Options), len(original.Options))
	}

	// Verify specific color
	if retrieved.Colors["bg"] != original.Colors["bg"] {
		t.Errorf("Colors[bg]: got %q, want %q", retrieved.Colors["bg"], original.Colors["bg"])
	}
}
