package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/plugin"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for integration tests
)

// =============================================================================
// Mock Implementation of PluginDataStore
// =============================================================================

// MockPluginDataStore implements PluginDataStore for testing.
type MockPluginDataStore struct {
	plugins map[string]*models.NvimPluginDB
	calls   []string
	err     error // inject errors for testing
}

func NewMockPluginDataStore() *MockPluginDataStore {
	return &MockPluginDataStore{
		plugins: make(map[string]*models.NvimPluginDB),
	}
}

func (m *MockPluginDataStore) CreatePlugin(p *models.NvimPluginDB) error {
	m.calls = append(m.calls, "CreatePlugin")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.plugins[p.Name]; exists {
		return fmt.Errorf("plugin already exists: %s", p.Name)
	}
	p.ID = len(m.plugins) + 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.plugins[p.Name] = p
	return nil
}

func (m *MockPluginDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	m.calls = append(m.calls, "GetPluginByName")
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

func (m *MockPluginDataStore) UpdatePlugin(p *models.NvimPluginDB) error {
	m.calls = append(m.calls, "UpdatePlugin")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.plugins[p.Name]; !exists {
		return fmt.Errorf("plugin not found: %s", p.Name)
	}
	p.UpdatedAt = time.Now()
	m.plugins[p.Name] = p
	return nil
}

func (m *MockPluginDataStore) DeletePlugin(name string) error {
	m.calls = append(m.calls, "DeletePlugin")
	if m.err != nil {
		return m.err
	}
	delete(m.plugins, name)
	return nil
}

func (m *MockPluginDataStore) ListPlugins() ([]*models.NvimPluginDB, error) {
	m.calls = append(m.calls, "ListPlugins")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.NvimPluginDB
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockPluginDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	m.calls = append(m.calls, "ListPluginsByCategory")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.NvimPluginDB
	for _, p := range m.plugins {
		if p.Category.Valid && p.Category.String == category {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockPluginDataStore) ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error) {
	m.calls = append(m.calls, "ListPluginsByTags")
	if m.err != nil {
		return nil, m.err
	}
	// Simplified: return all plugins (real impl would filter)
	var result []*models.NvimPluginDB
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockPluginDataStore) Close() error {
	m.calls = append(m.calls, "Close")
	return nil
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

// TestDBStoreAdapter_ImplementsPluginStore verifies that DBStoreAdapter
// satisfies the PluginStore interface at compile time.
func TestDBStoreAdapter_ImplementsPluginStore(t *testing.T) {
	var _ PluginStore = (*DBStoreAdapter)(nil)
}

// TestMockPluginDataStore_ImplementsPluginDataStore verifies the mock
// satisfies the PluginDataStore interface.
func TestMockPluginDataStore_ImplementsPluginDataStore(t *testing.T) {
	var _ PluginDataStore = (*MockPluginDataStore)(nil)
}

// =============================================================================
// Direct Struct Method Tests
// =============================================================================

func TestDBStoreAdapter_Create(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	p := &plugin.Plugin{
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Description: "Fuzzy finder",
		Category:    "navigation",
		Enabled:     true,
	}

	// Test successful create
	err := adapter.Create(p)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify plugin was stored
	if len(mock.plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(mock.plugins))
	}

	// Verify calls
	if len(mock.calls) != 2 { // GetPluginByName + CreatePlugin
		t.Errorf("Expected 2 calls, got %d: %v", len(mock.calls), mock.calls)
	}

	// Test duplicate create returns ErrAlreadyExists
	err = adapter.Create(p)
	if err == nil {
		t.Error("Expected error for duplicate create")
	}
	if !IsAlreadyExists(err) {
		t.Errorf("Expected ErrAlreadyExists, got %T: %v", err, err)
	}
}

func TestDBStoreAdapter_Get(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Add a plugin directly to mock
	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID:          1,
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Description: sql.NullString{String: "Fuzzy finder", Valid: true},
		Category:    sql.NullString{String: "navigation", Valid: true},
		Enabled:     true,
	}

	// Test successful get
	p, err := adapter.Get("telescope")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name != "telescope" {
		t.Errorf("Expected name 'telescope', got '%s'", p.Name)
	}
	if p.Description != "Fuzzy finder" {
		t.Errorf("Expected description 'Fuzzy finder', got '%s'", p.Description)
	}

	// Test get non-existent
	_, err = adapter.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDBStoreAdapter_Update(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Add a plugin directly to mock
	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID:      1,
		Name:    "telescope",
		Repo:    "nvim-telescope/telescope.nvim",
		Enabled: true,
	}

	// Test successful update
	p := &plugin.Plugin{
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Description: "Updated description",
		Enabled:     false,
	}
	err := adapter.Update(p)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	if mock.plugins["telescope"].Description.String != "Updated description" {
		t.Error("Description not updated")
	}

	// Test update non-existent
	p2 := &plugin.Plugin{Name: "nonexistent", Repo: "some/repo"}
	err = adapter.Update(p2)
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDBStoreAdapter_Upsert(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	p := &plugin.Plugin{
		Name:    "telescope",
		Repo:    "nvim-telescope/telescope.nvim",
		Enabled: true,
	}

	// Upsert should create when not exists
	err := adapter.Upsert(p)
	if err != nil {
		t.Fatalf("Upsert (create) failed: %v", err)
	}
	if len(mock.plugins) != 1 {
		t.Errorf("Expected 1 plugin after create, got %d", len(mock.plugins))
	}

	// Upsert should update when exists
	p.Description = "Updated via upsert"
	err = adapter.Upsert(p)
	if err != nil {
		t.Fatalf("Upsert (update) failed: %v", err)
	}
	if mock.plugins["telescope"].Description.String != "Updated via upsert" {
		t.Error("Plugin not updated via upsert")
	}
}

func TestDBStoreAdapter_Delete(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Add a plugin
	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID:   1,
		Name: "telescope",
		Repo: "nvim-telescope/telescope.nvim",
	}

	// Test successful delete
	err := adapter.Delete("telescope")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(mock.plugins) != 0 {
		t.Errorf("Expected 0 plugins after delete, got %d", len(mock.plugins))
	}

	// Test delete non-existent
	err = adapter.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDBStoreAdapter_List(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	// Empty list
	plugins, err := adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}

	// Add plugins
	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID: 1, Name: "telescope", Repo: "nvim-telescope/telescope.nvim",
	}
	mock.plugins["treesitter"] = &models.NvimPluginDB{
		ID: 2, Name: "treesitter", Repo: "nvim-treesitter/nvim-treesitter",
	}

	plugins, err = adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestDBStoreAdapter_ListByCategory(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID: 1, Name: "telescope", Repo: "nvim-telescope/telescope.nvim",
		Category: sql.NullString{String: "navigation", Valid: true},
	}
	mock.plugins["treesitter"] = &models.NvimPluginDB{
		ID: 2, Name: "treesitter", Repo: "nvim-treesitter/nvim-treesitter",
		Category: sql.NullString{String: "syntax", Valid: true},
	}

	plugins, err := adapter.ListByCategory("navigation")
	if err != nil {
		t.Fatalf("ListByCategory failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin in navigation, got %d", len(plugins))
	}
}

func TestDBStoreAdapter_ListByTag(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID: 1, Name: "telescope", Repo: "nvim-telescope/telescope.nvim",
	}

	// ListByTag uses mock which returns all plugins
	plugins, err := adapter.ListByTag("fuzzy")
	if err != nil {
		t.Fatalf("ListByTag failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

func TestDBStoreAdapter_Exists(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBStoreAdapter(mock)

	mock.plugins["telescope"] = &models.NvimPluginDB{
		ID: 1, Name: "telescope", Repo: "nvim-telescope/telescope.nvim",
	}

	// Test exists
	exists, err := adapter.Exists("telescope")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected telescope to exist")
	}

	// Test not exists
	exists, err = adapter.Exists("nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected nonexistent to not exist")
	}
}

func TestDBStoreAdapter_Close(t *testing.T) {
	mock := NewMockPluginDataStore()

	// Test non-owned connection (should not close)
	adapter := NewDBStoreAdapter(mock)
	err := adapter.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	// Close should not be called on mock
	hasClose := false
	for _, call := range mock.calls {
		if call == "Close" {
			hasClose = true
		}
	}
	if hasClose {
		t.Error("Close should not be called on non-owned connection")
	}

	// Test owned connection (should close)
	mock2 := NewMockPluginDataStore()
	adapter2 := NewDBStoreAdapterOwned(mock2)
	err = adapter2.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	hasClose = false
	for _, call := range mock2.calls {
		if call == "Close" {
			hasClose = true
		}
	}
	if !hasClose {
		t.Error("Close should be called on owned connection")
	}
}

// =============================================================================
// Model Conversion Tests
// =============================================================================

func TestPluginToDBModel_BasicFields(t *testing.T) {
	p := &plugin.Plugin{
		Name:        "telescope",
		Description: "Fuzzy finder",
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      "main",
		Version:     "0.1.0",
		Priority:    100,
		Lazy:        true,
		Category:    "navigation",
		Enabled:     true,
		Build:       "make",
		Config:      "require('telescope').setup()",
		Init:        "vim.g.telescope_enabled = true",
	}

	db := pluginToDBModel(p)

	if db.Name != "telescope" {
		t.Errorf("Name mismatch: %s", db.Name)
	}
	if db.Description.String != "Fuzzy finder" || !db.Description.Valid {
		t.Errorf("Description mismatch: %v", db.Description)
	}
	if db.Repo != "nvim-telescope/telescope.nvim" {
		t.Errorf("Repo mismatch: %s", db.Repo)
	}
	if db.Branch.String != "main" || !db.Branch.Valid {
		t.Errorf("Branch mismatch: %v", db.Branch)
	}
	if db.Version.String != "0.1.0" || !db.Version.Valid {
		t.Errorf("Version mismatch: %v", db.Version)
	}
	if db.Priority.Int64 != 100 || !db.Priority.Valid {
		t.Errorf("Priority mismatch: %v", db.Priority)
	}
	if !db.Lazy {
		t.Error("Lazy should be true")
	}
	if db.Category.String != "navigation" || !db.Category.Valid {
		t.Errorf("Category mismatch: %v", db.Category)
	}
	if !db.Enabled {
		t.Error("Enabled should be true")
	}
	if db.Build.String != "make" || !db.Build.Valid {
		t.Errorf("Build mismatch: %v", db.Build)
	}
	if db.Config.String != "require('telescope').setup()" || !db.Config.Valid {
		t.Errorf("Config mismatch: %v", db.Config)
	}
	if db.Init.String != "vim.g.telescope_enabled = true" || !db.Init.Valid {
		t.Errorf("Init mismatch: %v", db.Init)
	}
}

func TestPluginToDBModel_Arrays(t *testing.T) {
	p := &plugin.Plugin{
		Name:  "telescope",
		Repo:  "nvim-telescope/telescope.nvim",
		Event: []string{"VimEnter", "BufRead"},
		Ft:    []string{"lua", "python"},
		Cmd:   []string{"Telescope", "TelescopeFind"},
		Tags:  []string{"fuzzy", "finder", "navigation"},
	}

	db := pluginToDBModel(p)

	if !db.Event.Valid || db.Event.String != `["VimEnter","BufRead"]` {
		t.Errorf("Event mismatch: %v", db.Event)
	}
	if !db.Ft.Valid || db.Ft.String != `["lua","python"]` {
		t.Errorf("Ft mismatch: %v", db.Ft)
	}
	if !db.Cmd.Valid || db.Cmd.String != `["Telescope","TelescopeFind"]` {
		t.Errorf("Cmd mismatch: %v", db.Cmd)
	}
	if !db.Tags.Valid || db.Tags.String != `["fuzzy","finder","navigation"]` {
		t.Errorf("Tags mismatch: %v", db.Tags)
	}
}

func TestPluginToDBModel_Keys(t *testing.T) {
	p := &plugin.Plugin{
		Name: "telescope",
		Repo: "nvim-telescope/telescope.nvim",
		Keys: []plugin.Keymap{
			{Key: "<leader>ff", Mode: []string{"n"}, Action: "find_files", Desc: "Find files"},
			{Key: "<leader>fg", Mode: []string{"n", "v"}, Action: "live_grep", Desc: "Live grep"},
		},
	}

	db := pluginToDBModel(p)

	if !db.Keys.Valid {
		t.Fatal("Keys should be valid")
	}
	// Just verify it's valid JSON with expected content
	if db.Keys.String == "" {
		t.Error("Keys should not be empty")
	}
}

func TestPluginToDBModel_Dependencies(t *testing.T) {
	p := &plugin.Plugin{
		Name: "telescope",
		Repo: "nvim-telescope/telescope.nvim",
		Dependencies: []plugin.Dependency{
			{Repo: "nvim-lua/plenary.nvim"},
			{Repo: "nvim-tree/nvim-web-devicons", Build: "make"},
		},
	}

	db := pluginToDBModel(p)

	if !db.Dependencies.Valid {
		t.Fatal("Dependencies should be valid")
	}
	if db.Dependencies.String == "" {
		t.Error("Dependencies should not be empty")
	}
}

func TestPluginToDBModel_Opts(t *testing.T) {
	p := &plugin.Plugin{
		Name: "telescope",
		Repo: "nvim-telescope/telescope.nvim",
		Opts: map[string]interface{}{
			"defaults": map[string]interface{}{
				"layout_strategy": "horizontal",
			},
		},
	}

	db := pluginToDBModel(p)

	if !db.Opts.Valid {
		t.Fatal("Opts should be valid")
	}
	if db.Opts.String == "" {
		t.Error("Opts should not be empty")
	}
}

func TestDBModelToPlugin_BasicFields(t *testing.T) {
	db := &models.NvimPluginDB{
		ID:          1,
		Name:        "telescope",
		Description: sql.NullString{String: "Fuzzy finder", Valid: true},
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      sql.NullString{String: "main", Valid: true},
		Version:     sql.NullString{String: "0.1.0", Valid: true},
		Priority:    sql.NullInt64{Int64: 100, Valid: true},
		Lazy:        true,
		Category:    sql.NullString{String: "navigation", Valid: true},
		Enabled:     true,
		Build:       sql.NullString{String: "make", Valid: true},
		Config:      sql.NullString{String: "require('telescope').setup()", Valid: true},
		Init:        sql.NullString{String: "vim.g.telescope_enabled = true", Valid: true},
	}

	p := dbModelToPlugin(db)

	if p.Name != "telescope" {
		t.Errorf("Name mismatch: %s", p.Name)
	}
	if p.Description != "Fuzzy finder" {
		t.Errorf("Description mismatch: %s", p.Description)
	}
	if p.Repo != "nvim-telescope/telescope.nvim" {
		t.Errorf("Repo mismatch: %s", p.Repo)
	}
	if p.Branch != "main" {
		t.Errorf("Branch mismatch: %s", p.Branch)
	}
	if p.Version != "0.1.0" {
		t.Errorf("Version mismatch: %s", p.Version)
	}
	if p.Priority != 100 {
		t.Errorf("Priority mismatch: %d", p.Priority)
	}
	if !p.Lazy {
		t.Error("Lazy should be true")
	}
	if p.Category != "navigation" {
		t.Errorf("Category mismatch: %s", p.Category)
	}
	if !p.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestDBModelToPlugin_Arrays(t *testing.T) {
	db := &models.NvimPluginDB{
		Name:  "telescope",
		Repo:  "nvim-telescope/telescope.nvim",
		Event: sql.NullString{String: `["VimEnter","BufRead"]`, Valid: true},
		Ft:    sql.NullString{String: `["lua","python"]`, Valid: true},
		Cmd:   sql.NullString{String: `["Telescope","TelescopeFind"]`, Valid: true},
		Tags:  sql.NullString{String: `["fuzzy","finder"]`, Valid: true},
	}

	p := dbModelToPlugin(db)

	if len(p.Event) != 2 || p.Event[0] != "VimEnter" {
		t.Errorf("Event mismatch: %v", p.Event)
	}
	if len(p.Ft) != 2 || p.Ft[0] != "lua" {
		t.Errorf("Ft mismatch: %v", p.Ft)
	}
	if len(p.Cmd) != 2 || p.Cmd[0] != "Telescope" {
		t.Errorf("Cmd mismatch: %v", p.Cmd)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "fuzzy" {
		t.Errorf("Tags mismatch: %v", p.Tags)
	}
}

func TestConversion_Roundtrip(t *testing.T) {
	// Create a fully-populated plugin
	original := &plugin.Plugin{
		Name:        "telescope",
		Description: "Fuzzy finder",
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      "main",
		Version:     "0.1.0",
		Priority:    100,
		Lazy:        true,
		Event:       []string{"VimEnter"},
		Ft:          []string{"lua"},
		Cmd:         []string{"Telescope"},
		Keys: []plugin.Keymap{
			{Key: "<leader>ff", Mode: []string{"n"}, Action: "find_files", Desc: "Find files"},
		},
		Dependencies: []plugin.Dependency{
			{Repo: "nvim-lua/plenary.nvim"},
		},
		Build:    "make",
		Config:   "require('telescope').setup()",
		Init:     "vim.g.telescope = true",
		Opts:     map[string]interface{}{"key": "value"},
		Keymaps:  []plugin.Keymap{{Key: "<C-p>", Mode: []string{"n"}, Desc: "Quick find"}},
		Category: "navigation",
		Tags:     []string{"fuzzy", "finder"},
		Enabled:  true,
	}

	// Convert to DB model and back
	db := pluginToDBModel(original)
	converted := dbModelToPlugin(db)

	// Verify all fields match
	if converted.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", converted.Name, original.Name)
	}
	if converted.Description != original.Description {
		t.Errorf("Description mismatch: got %s, want %s", converted.Description, original.Description)
	}
	if converted.Repo != original.Repo {
		t.Errorf("Repo mismatch: got %s, want %s", converted.Repo, original.Repo)
	}
	if converted.Branch != original.Branch {
		t.Errorf("Branch mismatch: got %s, want %s", converted.Branch, original.Branch)
	}
	if converted.Version != original.Version {
		t.Errorf("Version mismatch: got %s, want %s", converted.Version, original.Version)
	}
	if converted.Priority != original.Priority {
		t.Errorf("Priority mismatch: got %d, want %d", converted.Priority, original.Priority)
	}
	if converted.Lazy != original.Lazy {
		t.Errorf("Lazy mismatch: got %v, want %v", converted.Lazy, original.Lazy)
	}
	if converted.Category != original.Category {
		t.Errorf("Category mismatch: got %s, want %s", converted.Category, original.Category)
	}
	if converted.Enabled != original.Enabled {
		t.Errorf("Enabled mismatch: got %v, want %v", converted.Enabled, original.Enabled)
	}
	if converted.Build != original.Build {
		t.Errorf("Build mismatch: got %s, want %s", converted.Build, original.Build)
	}
	if converted.Config != original.Config {
		t.Errorf("Config mismatch: got %s, want %s", converted.Config, original.Config)
	}
	if converted.Init != original.Init {
		t.Errorf("Init mismatch: got %s, want %s", converted.Init, original.Init)
	}

	// Check arrays
	if len(converted.Event) != len(original.Event) {
		t.Errorf("Event length mismatch: got %d, want %d", len(converted.Event), len(original.Event))
	}
	if len(converted.Ft) != len(original.Ft) {
		t.Errorf("Ft length mismatch: got %d, want %d", len(converted.Ft), len(original.Ft))
	}
	if len(converted.Tags) != len(original.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(converted.Tags), len(original.Tags))
	}
	if len(converted.Keys) != len(original.Keys) {
		t.Errorf("Keys length mismatch: got %d, want %d", len(converted.Keys), len(original.Keys))
	}
	if len(converted.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies length mismatch: got %d, want %d", len(converted.Dependencies), len(original.Dependencies))
	}
}

// =============================================================================
// Through Interface Tests (PluginStore interface)
// =============================================================================

// TestPluginStore_ThroughInterface tests the DBStoreAdapter through
// the PluginStore interface to ensure polymorphism works correctly.
func TestPluginStore_ThroughInterface(t *testing.T) {
	mock := NewMockPluginDataStore()

	// Use through interface
	var store PluginStore = NewDBStoreAdapter(mock)

	// Test Create through interface
	p := &plugin.Plugin{
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Description: "Fuzzy finder",
		Category:    "navigation",
		Enabled:     true,
	}

	err := store.Upsert(p)
	if err != nil {
		t.Fatalf("Upsert through interface failed: %v", err)
	}

	// Test Get through interface
	retrieved, err := store.Get("telescope")
	if err != nil {
		t.Fatalf("Get through interface failed: %v", err)
	}
	if retrieved.Name != "telescope" {
		t.Errorf("Expected name telescope, got %s", retrieved.Name)
	}

	// Test Exists through interface
	exists, err := store.Exists("telescope")
	if err != nil {
		t.Fatalf("Exists through interface failed: %v", err)
	}
	if !exists {
		t.Error("Expected telescope to exist")
	}

	// Test List through interface
	plugins, err := store.List()
	if err != nil {
		t.Fatalf("List through interface failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	// Test Delete through interface
	err = store.Delete("telescope")
	if err != nil {
		t.Fatalf("Delete through interface failed: %v", err)
	}

	exists, _ = store.Exists("telescope")
	if exists {
		t.Error("Expected telescope to not exist after delete")
	}

	// Test Close through interface
	err = store.Close()
	if err != nil {
		t.Fatalf("Close through interface failed: %v", err)
	}
}

// TestPluginStore_Swappability verifies that the adapter can be swapped
// with other PluginStore implementations (like MemoryStore).
func TestPluginStore_Swappability(t *testing.T) {
	// Create a test function that works with any PluginStore
	testStore := func(store PluginStore, name string) {
		p := &plugin.Plugin{
			Name:    "test-plugin",
			Repo:    "test/repo",
			Enabled: true,
		}

		// Create
		if err := store.Upsert(p); err != nil {
			t.Errorf("%s: Upsert failed: %v", name, err)
			return
		}

		// Get
		retrieved, err := store.Get("test-plugin")
		if err != nil {
			t.Errorf("%s: Get failed: %v", name, err)
			return
		}
		if retrieved.Name != "test-plugin" {
			t.Errorf("%s: Name mismatch", name)
		}

		// Delete
		if err := store.Delete("test-plugin"); err != nil {
			t.Errorf("%s: Delete failed: %v", name, err)
		}

		store.Close()
	}

	// Test with DBStoreAdapter
	mock := NewMockPluginDataStore()
	testStore(NewDBStoreAdapter(mock), "DBStoreAdapter")

	// Test with MemoryStore (if available)
	memStore := NewMemoryStore()
	testStore(memStore, "MemoryStore")
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestModeSliceToInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected interface{}
	}{
		{"empty", []string{}, nil},
		{"single", []string{"n"}, "n"},
		{"multiple", []string{"n", "v"}, []string{"n", "v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modeSliceToInterface(tt.input)
			switch exp := tt.expected.(type) {
			case nil:
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			case string:
				if result != exp {
					t.Errorf("Expected %s, got %v", exp, result)
				}
			case []string:
				arr, ok := result.([]string)
				if !ok {
					t.Errorf("Expected []string, got %T", result)
					return
				}
				if len(arr) != len(exp) {
					t.Errorf("Expected length %d, got %d", len(exp), len(arr))
				}
			}
		})
	}
}

func TestInterfaceToModeSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{"nil", nil, nil},
		{"string", "n", []string{"n"}},
		{"slice_interface", []interface{}{"n", "v"}, []string{"n", "v"}},
		{"slice_string", []string{"n", "v"}, []string{"n", "v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interfaceToModeSlice(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Index %d: expected %s, got %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

// =============================================================================
// Integration Tests - Real In-Memory SQLite Database
//
// These tests verify the full stack: DBStoreAdapter -> DataStore -> SQLite
// They write to a real database and verify data with raw SQL queries.
// =============================================================================

// createIntegrationTestDataStore creates a real in-memory SQLite DataStore for testing
func createIntegrationTestDataStore(t *testing.T) *integrationTestDB {
	t.Helper()

	// Import the db package dynamically to avoid circular imports
	// We'll use database/sql directly with sqlite3

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite: %v", err)
	}

	// Create the nvim_plugins table (matching the real schema)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS nvim_plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			repo TEXT NOT NULL,
			branch TEXT,
			version TEXT,
			priority INTEGER,
			lazy INTEGER DEFAULT 0,
			event TEXT,
			ft TEXT,
			keys TEXT,
			cmd TEXT,
			dependencies TEXT,
			build TEXT,
			config TEXT,
			init TEXT,
			opts TEXT,
			keymaps TEXT,
			category TEXT,
			tags TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create nvim_plugins table: %v", err)
	}

	return &integrationTestDB{db: db, t: t}
}

// integrationTestDB wraps sql.DB and implements PluginDataStore for integration testing
type integrationTestDB struct {
	db *sql.DB
	t  *testing.T
}

func (d *integrationTestDB) CreatePlugin(p *models.NvimPluginDB) error {
	result, err := d.db.Exec(`
		INSERT INTO nvim_plugins (
			name, description, repo, branch, version, priority, lazy,
			event, ft, keys, cmd, dependencies, build, config, init,
			opts, keymaps, category, tags, enabled
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		p.Name,
		p.Description,
		p.Repo,
		p.Branch,
		p.Version,
		p.Priority,
		p.Lazy,
		p.Event,
		p.Ft,
		p.Keys,
		p.Cmd,
		p.Dependencies,
		p.Build,
		p.Config,
		p.Init,
		p.Opts,
		p.Keymaps,
		p.Category,
		p.Tags,
		p.Enabled,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	p.ID = int(id)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

func (d *integrationTestDB) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	p := &models.NvimPluginDB{}
	err := d.db.QueryRow(`
		SELECT id, name, description, repo, branch, version, priority, lazy,
			   event, ft, keys, cmd, dependencies, build, config, init,
			   opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE name = ?
	`, name).Scan(
		&p.ID, &p.Name, &p.Description, &p.Repo, &p.Branch, &p.Version,
		&p.Priority, &p.Lazy, &p.Event, &p.Ft, &p.Keys, &p.Cmd,
		&p.Dependencies, &p.Build, &p.Config, &p.Init, &p.Opts,
		&p.Keymaps, &p.Category, &p.Tags, &p.Enabled, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, err
}

func (d *integrationTestDB) UpdatePlugin(p *models.NvimPluginDB) error {
	_, err := d.db.Exec(`
		UPDATE nvim_plugins SET
			description = ?, repo = ?, branch = ?, version = ?, priority = ?,
			lazy = ?, event = ?, ft = ?, keys = ?, cmd = ?, dependencies = ?,
			build = ?, config = ?, init = ?, opts = ?, keymaps = ?,
			category = ?, tags = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE name = ?
	`,
		p.Description, p.Repo, p.Branch, p.Version, p.Priority,
		p.Lazy, p.Event, p.Ft, p.Keys, p.Cmd, p.Dependencies,
		p.Build, p.Config, p.Init, p.Opts, p.Keymaps,
		p.Category, p.Tags, p.Enabled, p.Name,
	)
	return err
}

func (d *integrationTestDB) DeletePlugin(name string) error {
	_, err := d.db.Exec(`DELETE FROM nvim_plugins WHERE name = ?`, name)
	return err
}

func (d *integrationTestDB) ListPlugins() ([]*models.NvimPluginDB, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, repo, branch, version, priority, lazy,
			   event, ft, keys, cmd, dependencies, build, config, init,
			   opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		p := &models.NvimPluginDB{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Repo, &p.Branch, &p.Version,
			&p.Priority, &p.Lazy, &p.Event, &p.Ft, &p.Keys, &p.Cmd,
			&p.Dependencies, &p.Build, &p.Config, &p.Init, &p.Opts,
			&p.Keymaps, &p.Category, &p.Tags, &p.Enabled, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func (d *integrationTestDB) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	rows, err := d.db.Query(`
		SELECT id, name, description, repo, branch, version, priority, lazy,
			   event, ft, keys, cmd, dependencies, build, config, init,
			   opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE category = ? ORDER BY name
	`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		p := &models.NvimPluginDB{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Repo, &p.Branch, &p.Version,
			&p.Priority, &p.Lazy, &p.Event, &p.Ft, &p.Keys, &p.Cmd,
			&p.Dependencies, &p.Build, &p.Config, &p.Init, &p.Opts,
			&p.Keymaps, &p.Category, &p.Tags, &p.Enabled, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func (d *integrationTestDB) ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error) {
	// Simplified: return all plugins (real impl would filter by tags)
	return d.ListPlugins()
}

func (d *integrationTestDB) Close() error {
	return d.db.Close()
}

// rawQuery executes a raw SQL query and returns the result for verification
func (d *integrationTestDB) rawQuery(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// rawExec executes a raw SQL statement
func (d *integrationTestDB) rawExec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// =============================================================================
// Integration Test: Write Plugin, Verify with Raw SQL
// =============================================================================

func TestIntegration_CreatePlugin_VerifyWithSQL(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create plugin through adapter
	p := &plugin.Plugin{
		Name:        "telescope",
		Description: "Fuzzy finder and picker",
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      "master",
		Version:     "0.1.0",
		Priority:    100,
		Lazy:        true,
		Event:       []string{"VimEnter", "BufRead"},
		Ft:          []string{"lua", "python"},
		Cmd:         []string{"Telescope", "TelescopeFindFiles"},
		Category:    "navigation",
		Tags:        []string{"fuzzy", "finder", "picker"},
		Enabled:     true,
		Build:       "make",
		Config:      "require('telescope').setup({})",
		Init:        "vim.g.telescope_enabled = true",
		Keys: []plugin.Keymap{
			{Key: "<leader>ff", Mode: []string{"n"}, Action: "find_files", Desc: "Find files"},
		},
		Dependencies: []plugin.Dependency{
			{Repo: "nvim-lua/plenary.nvim"},
		},
		Opts: map[string]interface{}{
			"defaults": map[string]interface{}{
				"layout_strategy": "horizontal",
			},
		},
	}

	err := adapter.Create(p)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify with raw SQL query
	var (
		name        string
		description sql.NullString
		repo        string
		branch      sql.NullString
		version     sql.NullString
		priority    sql.NullInt64
		lazy        bool
		event       sql.NullString
		category    sql.NullString
		enabled     bool
	)

	err = testDB.rawQuery(`
		SELECT name, description, repo, branch, version, priority, lazy, event, category, enabled
		FROM nvim_plugins WHERE name = ?
	`, "telescope").Scan(&name, &description, &repo, &branch, &version, &priority, &lazy, &event, &category, &enabled)

	if err != nil {
		t.Fatalf("Raw SQL query failed: %v", err)
	}

	// Verify each field
	if name != "telescope" {
		t.Errorf("SQL name = %q, want %q", name, "telescope")
	}
	if !description.Valid || description.String != "Fuzzy finder and picker" {
		t.Errorf("SQL description = %v, want %q", description, "Fuzzy finder and picker")
	}
	if repo != "nvim-telescope/telescope.nvim" {
		t.Errorf("SQL repo = %q, want %q", repo, "nvim-telescope/telescope.nvim")
	}
	if !branch.Valid || branch.String != "master" {
		t.Errorf("SQL branch = %v, want %q", branch, "master")
	}
	if !version.Valid || version.String != "0.1.0" {
		t.Errorf("SQL version = %v, want %q", version, "0.1.0")
	}
	if !priority.Valid || priority.Int64 != 100 {
		t.Errorf("SQL priority = %v, want %d", priority, 100)
	}
	if !lazy {
		t.Error("SQL lazy should be true")
	}
	if !event.Valid || event.String != `["VimEnter","BufRead"]` {
		t.Errorf("SQL event = %v, want JSON array", event)
	}
	if !category.Valid || category.String != "navigation" {
		t.Errorf("SQL category = %v, want %q", category, "navigation")
	}
	if !enabled {
		t.Error("SQL enabled should be true")
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
	p := &plugin.Plugin{
		Name:     "treesitter",
		Repo:     "nvim-treesitter/nvim-treesitter",
		Category: "syntax",
		Enabled:  true,
	}

	err := adapter.Create(p)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify count via SQL
	var count int
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_plugins`).Scan(&count)
	if count != 1 {
		t.Errorf("After CREATE: count = %d, want 1", count)
	}

	// --- READ ---
	retrieved, err := adapter.Get("treesitter")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != "treesitter" {
		t.Errorf("Get name = %q, want %q", retrieved.Name, "treesitter")
	}
	if retrieved.Category != "syntax" {
		t.Errorf("Get category = %q, want %q", retrieved.Category, "syntax")
	}

	// --- UPDATE ---
	retrieved.Description = "Updated description"
	retrieved.Enabled = false
	err = adapter.Update(retrieved)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update via SQL
	var desc sql.NullString
	var enabled bool
	testDB.rawQuery(`SELECT description, enabled FROM nvim_plugins WHERE name = ?`, "treesitter").Scan(&desc, &enabled)
	if !desc.Valid || desc.String != "Updated description" {
		t.Errorf("After UPDATE: description = %v, want %q", desc, "Updated description")
	}
	if enabled {
		t.Error("After UPDATE: enabled should be false")
	}

	// --- DELETE ---
	err = adapter.Delete("treesitter")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion via SQL
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_plugins`).Scan(&count)
	if count != 0 {
		t.Errorf("After DELETE: count = %d, want 0", count)
	}
}

// =============================================================================
// Integration Test: List and Filter Operations
// =============================================================================

func TestIntegration_ListOperations(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create multiple plugins
	plugins := []*plugin.Plugin{
		{Name: "telescope", Repo: "nvim-telescope/telescope.nvim", Category: "navigation", Enabled: true},
		{Name: "treesitter", Repo: "nvim-treesitter/nvim-treesitter", Category: "syntax", Enabled: true},
		{Name: "lspconfig", Repo: "neovim/nvim-lspconfig", Category: "lsp", Enabled: true},
		{Name: "neo-tree", Repo: "nvim-neo-tree/neo-tree.nvim", Category: "navigation", Enabled: false},
	}

	for _, p := range plugins {
		if err := adapter.Create(p); err != nil {
			t.Fatalf("Create %s failed: %v", p.Name, err)
		}
	}

	// Verify count via SQL
	var count int
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_plugins`).Scan(&count)
	if count != 4 {
		t.Errorf("SQL count = %d, want 4", count)
	}

	// Test List
	all, err := adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 4 {
		t.Errorf("List returned %d plugins, want 4", len(all))
	}

	// Test ListByCategory
	navPlugins, err := adapter.ListByCategory("navigation")
	if err != nil {
		t.Fatalf("ListByCategory failed: %v", err)
	}
	if len(navPlugins) != 2 {
		t.Errorf("ListByCategory(navigation) returned %d plugins, want 2", len(navPlugins))
	}

	// Verify navigation plugins are correct
	navNames := make(map[string]bool)
	for _, p := range navPlugins {
		navNames[p.Name] = true
	}
	if !navNames["telescope"] || !navNames["neo-tree"] {
		t.Errorf("ListByCategory should return telescope and neo-tree, got %v", navNames)
	}
}

// =============================================================================
// Integration Test: Upsert Behavior
// =============================================================================

func TestIntegration_UpsertBehavior(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	p := &plugin.Plugin{
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Description: "Original description",
		Enabled:     true,
	}

	// First upsert should create
	err := adapter.Upsert(p)
	if err != nil {
		t.Fatalf("First Upsert failed: %v", err)
	}

	var count int
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_plugins`).Scan(&count)
	if count != 1 {
		t.Errorf("After first Upsert: count = %d, want 1", count)
	}

	// Second upsert should update
	p.Description = "Updated description"
	p.Enabled = false
	err = adapter.Upsert(p)
	if err != nil {
		t.Fatalf("Second Upsert failed: %v", err)
	}

	// Should still be 1 record
	testDB.rawQuery(`SELECT COUNT(*) FROM nvim_plugins`).Scan(&count)
	if count != 1 {
		t.Errorf("After second Upsert: count = %d, want 1", count)
	}

	// Verify the update via SQL
	var desc sql.NullString
	var enabled bool
	testDB.rawQuery(`SELECT description, enabled FROM nvim_plugins WHERE name = ?`, "telescope").Scan(&desc, &enabled)
	if !desc.Valid || desc.String != "Updated description" {
		t.Errorf("After Upsert update: description = %v, want %q", desc, "Updated description")
	}
	if enabled {
		t.Error("After Upsert update: enabled should be false")
	}
}

// =============================================================================
// Integration Test: YAML Output Verification
// =============================================================================

func TestIntegration_YAMLOutput(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create a plugin with full configuration
	p := &plugin.Plugin{
		Name:        "telescope",
		Description: "Fuzzy finder",
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      "master",
		Category:    "navigation",
		Tags:        []string{"fuzzy", "finder"},
		Lazy:        true,
		Event:       []string{"VimEnter"},
		Enabled:     true,
	}

	err := adapter.Create(p)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Read back and convert to YAML
	retrieved, err := adapter.Get("telescope")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Convert to PluginYAML format
	yamlFormat := retrieved.ToYAML()

	// Verify YAML structure
	if yamlFormat.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("YAML APIVersion = %q, want %q", yamlFormat.APIVersion, "devopsmaestro.io/v1")
	}
	if yamlFormat.Kind != "NvimPlugin" {
		t.Errorf("YAML Kind = %q, want %q", yamlFormat.Kind, "NvimPlugin")
	}
	if yamlFormat.Metadata.Name != "telescope" {
		t.Errorf("YAML Metadata.Name = %q, want %q", yamlFormat.Metadata.Name, "telescope")
	}
	if yamlFormat.Metadata.Description != "Fuzzy finder" {
		t.Errorf("YAML Metadata.Description = %q, want %q", yamlFormat.Metadata.Description, "Fuzzy finder")
	}
	if yamlFormat.Metadata.Category != "navigation" {
		t.Errorf("YAML Metadata.Category = %q, want %q", yamlFormat.Metadata.Category, "navigation")
	}
	if len(yamlFormat.Metadata.Tags) != 2 {
		t.Errorf("YAML Metadata.Tags length = %d, want 2", len(yamlFormat.Metadata.Tags))
	}
	if yamlFormat.Spec.Repo != "nvim-telescope/telescope.nvim" {
		t.Errorf("YAML Spec.Repo = %q, want %q", yamlFormat.Spec.Repo, "nvim-telescope/telescope.nvim")
	}
	if yamlFormat.Spec.Branch != "master" {
		t.Errorf("YAML Spec.Branch = %q, want %q", yamlFormat.Spec.Branch, "master")
	}
	if !yamlFormat.Spec.Lazy {
		t.Error("YAML Spec.Lazy should be true")
	}
}

// =============================================================================
// Integration Test: Complex Plugin with All Fields
// =============================================================================

func TestIntegration_ComplexPluginRoundtrip(t *testing.T) {
	testDB := createIntegrationTestDataStore(t)
	defer testDB.Close()

	adapter := NewDBStoreAdapter(testDB)

	// Create a fully-populated plugin
	original := &plugin.Plugin{
		Name:        "telescope",
		Description: "Highly extendable fuzzy finder",
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      "master",
		Version:     "0.1.4",
		Priority:    1000,
		Lazy:        true,
		Event:       []string{"VimEnter", "BufReadPost"},
		Ft:          []string{"lua", "python", "go"},
		Cmd:         []string{"Telescope", "TelescopeFindFiles", "TelescopeGrepString"},
		Keys: []plugin.Keymap{
			{Key: "<leader>ff", Mode: []string{"n"}, Action: "find_files", Desc: "Find files"},
			{Key: "<leader>fg", Mode: []string{"n", "v"}, Action: "live_grep", Desc: "Live grep"},
			{Key: "<leader>fb", Mode: []string{"n"}, Action: "buffers", Desc: "List buffers"},
		},
		Dependencies: []plugin.Dependency{
			{Repo: "nvim-lua/plenary.nvim"},
			{Repo: "nvim-tree/nvim-web-devicons", Build: "make"},
			{Repo: "nvim-telescope/telescope-fzf-native.nvim", Build: "cmake -S. -Bbuild && cmake --build build"},
		},
		Build:  "make",
		Config: "require('telescope').setup({ defaults = { layout_strategy = 'horizontal' } })",
		Init:   "vim.g.telescope_loaded = true",
		Opts: map[string]interface{}{
			"defaults": map[string]interface{}{
				"layout_strategy":  "horizontal",
				"sorting_strategy": "ascending",
			},
			"pickers": map[string]interface{}{
				"find_files": map[string]interface{}{
					"hidden": true,
				},
			},
		},
		Keymaps: []plugin.Keymap{
			{Key: "<C-p>", Mode: []string{"n"}, Action: ":Telescope find_files<CR>", Desc: "Quick find"},
		},
		Category: "navigation",
		Tags:     []string{"fuzzy", "finder", "picker", "search", "navigation"},
		Enabled:  true,
	}

	// Create
	err := adapter.Create(original)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Read back
	retrieved, err := adapter.Get("telescope")
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
	if retrieved.Repo != original.Repo {
		t.Errorf("Repo: got %q, want %q", retrieved.Repo, original.Repo)
	}
	if retrieved.Branch != original.Branch {
		t.Errorf("Branch: got %q, want %q", retrieved.Branch, original.Branch)
	}
	if retrieved.Version != original.Version {
		t.Errorf("Version: got %q, want %q", retrieved.Version, original.Version)
	}
	if retrieved.Priority != original.Priority {
		t.Errorf("Priority: got %d, want %d", retrieved.Priority, original.Priority)
	}
	if retrieved.Lazy != original.Lazy {
		t.Errorf("Lazy: got %v, want %v", retrieved.Lazy, original.Lazy)
	}
	if retrieved.Category != original.Category {
		t.Errorf("Category: got %q, want %q", retrieved.Category, original.Category)
	}
	if retrieved.Enabled != original.Enabled {
		t.Errorf("Enabled: got %v, want %v", retrieved.Enabled, original.Enabled)
	}
	if retrieved.Build != original.Build {
		t.Errorf("Build: got %q, want %q", retrieved.Build, original.Build)
	}
	if retrieved.Config != original.Config {
		t.Errorf("Config: got %q, want %q", retrieved.Config, original.Config)
	}
	if retrieved.Init != original.Init {
		t.Errorf("Init: got %q, want %q", retrieved.Init, original.Init)
	}

	// Verify arrays
	if len(retrieved.Event) != len(original.Event) {
		t.Errorf("Event length: got %d, want %d", len(retrieved.Event), len(original.Event))
	}
	if len(retrieved.Ft) != len(original.Ft) {
		t.Errorf("Ft length: got %d, want %d", len(retrieved.Ft), len(original.Ft))
	}
	if len(retrieved.Cmd) != len(original.Cmd) {
		t.Errorf("Cmd length: got %d, want %d", len(retrieved.Cmd), len(original.Cmd))
	}
	if len(retrieved.Tags) != len(original.Tags) {
		t.Errorf("Tags length: got %d, want %d", len(retrieved.Tags), len(original.Tags))
	}
	if len(retrieved.Keys) != len(original.Keys) {
		t.Errorf("Keys length: got %d, want %d", len(retrieved.Keys), len(original.Keys))
	}
	if len(retrieved.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies length: got %d, want %d", len(retrieved.Dependencies), len(original.Dependencies))
	}
	if len(retrieved.Keymaps) != len(original.Keymaps) {
		t.Errorf("Keymaps length: got %d, want %d", len(retrieved.Keymaps), len(original.Keymaps))
	}

	// Verify specific key mapping
	if len(retrieved.Keys) > 0 {
		if retrieved.Keys[0].Key != original.Keys[0].Key {
			t.Errorf("Keys[0].Key: got %q, want %q", retrieved.Keys[0].Key, original.Keys[0].Key)
		}
		if retrieved.Keys[0].Desc != original.Keys[0].Desc {
			t.Errorf("Keys[0].Desc: got %q, want %q", retrieved.Keys[0].Desc, original.Keys[0].Desc)
		}
	}

	// Verify specific dependency
	if len(retrieved.Dependencies) > 0 {
		if retrieved.Dependencies[0].Repo != original.Dependencies[0].Repo {
			t.Errorf("Dependencies[0].Repo: got %q, want %q", retrieved.Dependencies[0].Repo, original.Dependencies[0].Repo)
		}
	}
}
