package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/plugin"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for integration tests
)

// =============================================================================
// Mock Implementation of PluginDataStore
// =============================================================================

// MockPluginDataStore implements PluginDataStore for testing.
type MockPluginDataStore struct {
	plugins map[string]*models.TerminalPluginDB
	calls   []string
	err     error // inject errors for testing
}

func NewMockPluginDataStore() *MockPluginDataStore {
	return &MockPluginDataStore{
		plugins: make(map[string]*models.TerminalPluginDB),
	}
}

func (m *MockPluginDataStore) CreateTerminalPlugin(p *models.TerminalPluginDB) error {
	m.calls = append(m.calls, "CreateTerminalPlugin")
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

func (m *MockPluginDataStore) GetTerminalPlugin(name string) (*models.TerminalPluginDB, error) {
	m.calls = append(m.calls, "GetTerminalPlugin")
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

func (m *MockPluginDataStore) UpdateTerminalPlugin(p *models.TerminalPluginDB) error {
	m.calls = append(m.calls, "UpdateTerminalPlugin")
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

func (m *MockPluginDataStore) UpsertTerminalPlugin(p *models.TerminalPluginDB) error {
	m.calls = append(m.calls, "UpsertTerminalPlugin")
	if m.err != nil {
		return m.err
	}
	if existing, exists := m.plugins[p.Name]; exists {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
	} else {
		p.ID = len(m.plugins) + 1
		p.CreatedAt = time.Now()
	}
	p.UpdatedAt = time.Now()
	m.plugins[p.Name] = p
	return nil
}

func (m *MockPluginDataStore) DeleteTerminalPlugin(name string) error {
	m.calls = append(m.calls, "DeleteTerminalPlugin")
	if m.err != nil {
		return m.err
	}
	delete(m.plugins, name)
	return nil
}

func (m *MockPluginDataStore) ListTerminalPlugins() ([]*models.TerminalPluginDB, error) {
	m.calls = append(m.calls, "ListTerminalPlugins")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalPluginDB
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockPluginDataStore) ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error) {
	m.calls = append(m.calls, "ListTerminalPluginsByCategory")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalPluginDB
	for _, p := range m.plugins {
		if p.Category.Valid && p.Category.String == category {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockPluginDataStore) ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error) {
	m.calls = append(m.calls, "ListTerminalPluginsByShell")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalPluginDB
	for _, p := range m.plugins {
		if p.Shell == shell {
			result = append(result, p)
		}
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

// TestDBPluginStore_ImplementsPluginStore verifies that DBPluginStore
// satisfies the PluginStore interface at compile time.
func TestDBPluginStore_ImplementsPluginStore(t *testing.T) {
	var _ plugin.PluginStore = (*DBPluginStore)(nil)
}

// TestMockPluginDataStore_ImplementsPluginDataStore verifies the mock
// satisfies the PluginDataStore interface.
func TestMockPluginDataStore_ImplementsPluginDataStore(t *testing.T) {
	var _ PluginDataStore = (*MockPluginDataStore)(nil)
}

// =============================================================================
// Direct Struct Method Tests
// =============================================================================

func TestDBPluginStore_Create(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	p := &plugin.Plugin{
		Name:        "zsh-autosuggestions",
		Repo:        "zsh-users/zsh-autosuggestions",
		Description: "Fish-like autosuggestions for zsh",
		Category:    "productivity",
		Manager:     plugin.PluginManagerZinit,
		LoadMode:    plugin.LoadModeImmediate,
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
	if len(mock.calls) != 2 { // GetTerminalPlugin + CreateTerminalPlugin
		t.Errorf("Expected 2 calls, got %d: %v", len(mock.calls), mock.calls)
	}

	// Test duplicate create returns ErrAlreadyExists
	err = adapter.Create(p)
	if err == nil {
		t.Error("Expected error for duplicate create")
	}
	if !plugin.IsAlreadyExists(err) {
		t.Errorf("Expected ErrAlreadyExists, got %T: %v", err, err)
	}
}

func TestDBPluginStore_Get(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	// Add a plugin directly to mock
	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID:          1,
		Name:        "zsh-autosuggestions",
		Repo:        "zsh-users/zsh-autosuggestions",
		Description: sql.NullString{String: "Fish-like autosuggestions", Valid: true},
		Category:    sql.NullString{String: "productivity", Valid: true},
		Shell:       "zsh",
		Manager:     "zinit",
		Enabled:     true,
	}

	// Test successful get
	p, err := adapter.Get("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name != "zsh-autosuggestions" {
		t.Errorf("Expected name 'zsh-autosuggestions', got '%s'", p.Name)
	}
	if p.Description != "Fish-like autosuggestions" {
		t.Errorf("Expected description 'Fish-like autosuggestions', got '%s'", p.Description)
	}
	if p.Manager != plugin.PluginManagerZinit {
		t.Errorf("Expected manager zinit, got '%s'", p.Manager)
	}

	// Test get non-existent
	_, err = adapter.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
	if !plugin.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDBPluginStore_Update(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	// Add a plugin directly to mock
	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID:      1,
		Name:    "zsh-autosuggestions",
		Repo:    "zsh-users/zsh-autosuggestions",
		Shell:   "zsh",
		Manager: "zinit",
		Enabled: true,
	}

	// Test successful update
	p := &plugin.Plugin{
		Name:        "zsh-autosuggestions",
		Repo:        "zsh-users/zsh-autosuggestions",
		Description: "Updated description",
		Manager:     plugin.PluginManagerZinit,
		Enabled:     false,
	}
	err := adapter.Update(p)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	if mock.plugins["zsh-autosuggestions"].Description.String != "Updated description" {
		t.Error("Description not updated")
	}

	// Test update non-existent
	p2 := &plugin.Plugin{Name: "nonexistent", Repo: "some/repo"}
	err = adapter.Update(p2)
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}
	if !plugin.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDBPluginStore_Upsert(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	p := &plugin.Plugin{
		Name:    "zsh-autosuggestions",
		Repo:    "zsh-users/zsh-autosuggestions",
		Manager: plugin.PluginManagerZinit,
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
	if mock.plugins["zsh-autosuggestions"].Description.String != "Updated via upsert" {
		t.Error("Plugin not updated via upsert")
	}
}

func TestDBPluginStore_Delete(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	// Add a plugin
	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID:   1,
		Name: "zsh-autosuggestions",
		Repo: "zsh-users/zsh-autosuggestions",
	}

	// Test successful delete
	err := adapter.Delete("zsh-autosuggestions")
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
	if !plugin.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDBPluginStore_List(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	// Empty list
	plugins, err := adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}

	// Add plugins
	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID: 1, Name: "zsh-autosuggestions", Repo: "zsh-users/zsh-autosuggestions", Shell: "zsh", Manager: "zinit",
	}
	mock.plugins["zsh-syntax-highlighting"] = &models.TerminalPluginDB{
		ID: 2, Name: "zsh-syntax-highlighting", Repo: "zsh-users/zsh-syntax-highlighting", Shell: "zsh", Manager: "zinit",
	}

	plugins, err = adapter.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

func TestDBPluginStore_ListByCategory(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID: 1, Name: "zsh-autosuggestions", Repo: "zsh-users/zsh-autosuggestions",
		Category: sql.NullString{String: "productivity", Valid: true}, Shell: "zsh", Manager: "zinit",
	}
	mock.plugins["powerlevel10k"] = &models.TerminalPluginDB{
		ID: 2, Name: "powerlevel10k", Repo: "romkatv/powerlevel10k",
		Category: sql.NullString{String: "theme", Valid: true}, Shell: "zsh", Manager: "zinit",
	}

	plugins, err := adapter.ListByCategory("productivity")
	if err != nil {
		t.Fatalf("ListByCategory failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin in productivity, got %d", len(plugins))
	}
}

func TestDBPluginStore_ListByManager(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID: 1, Name: "zsh-autosuggestions", Repo: "zsh-users/zsh-autosuggestions", Shell: "zsh", Manager: "zinit",
	}
	mock.plugins["oh-my-zsh-plugin"] = &models.TerminalPluginDB{
		ID: 2, Name: "oh-my-zsh-plugin", Repo: "", Shell: "zsh", Manager: "oh-my-zsh",
	}

	// Test ListByManager
	plugins, err := adapter.ListByManager(plugin.PluginManagerZinit)
	if err != nil {
		t.Fatalf("ListByManager failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("Expected 1 zinit plugin, got %d", len(plugins))
	}
}

func TestDBPluginStore_Exists(t *testing.T) {
	mock := NewMockPluginDataStore()
	adapter := NewDBPluginStore(mock)

	mock.plugins["zsh-autosuggestions"] = &models.TerminalPluginDB{
		ID: 1, Name: "zsh-autosuggestions", Repo: "zsh-users/zsh-autosuggestions", Shell: "zsh", Manager: "zinit",
	}

	// Test exists
	exists, err := adapter.Exists("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected zsh-autosuggestions to exist")
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

func TestDBPluginStore_Close(t *testing.T) {
	mock := NewMockPluginDataStore()

	// Test non-owned connection (should not close)
	adapter := NewDBPluginStore(mock)
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
	adapter2 := NewDBPluginStoreOwned(mock2)
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
		Name:        "zsh-autosuggestions",
		Description: "Fish-like autosuggestions for zsh",
		Repo:        "zsh-users/zsh-autosuggestions",
		Category:    "productivity",
		Manager:     plugin.PluginManagerZinit,
		LoadMode:    plugin.LoadModeDeferred,
		Enabled:     true,
		Priority:    100,
		Branch:      "master",
		Tag:         "v0.7.0",
	}

	db := pluginToDBModel(p)

	if db.Name != "zsh-autosuggestions" {
		t.Errorf("Name mismatch: %s", db.Name)
	}
	if db.Description.String != "Fish-like autosuggestions for zsh" || !db.Description.Valid {
		t.Errorf("Description mismatch: %v", db.Description)
	}
	if db.Repo != "zsh-users/zsh-autosuggestions" {
		t.Errorf("Repo mismatch: %s", db.Repo)
	}
	if db.Category.String != "productivity" || !db.Category.Valid {
		t.Errorf("Category mismatch: %v", db.Category)
	}
	if db.Manager != "zinit" {
		t.Errorf("Manager mismatch: %s", db.Manager)
	}
	if db.Shell != "zsh" {
		t.Errorf("Shell mismatch: %s", db.Shell)
	}
	if !db.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestPluginToDBModel_OhMyZshPlugin(t *testing.T) {
	p := &plugin.Plugin{
		Name:          "git",
		OhMyZshPlugin: "git",
		Manager:       plugin.PluginManagerOhMyZsh,
		Enabled:       true,
	}

	db := pluginToDBModel(p)

	if db.LoadCommand.String != "plugins+=git" || !db.LoadCommand.Valid {
		t.Errorf("LoadCommand mismatch for oh-my-zsh: %v", db.LoadCommand)
	}
	if db.Manager != "oh-my-zsh" {
		t.Errorf("Manager mismatch: %s", db.Manager)
	}
}

func TestPluginToDBModel_Dependencies(t *testing.T) {
	p := &plugin.Plugin{
		Name:         "example",
		Repo:         "user/example",
		Dependencies: []string{"dep1", "dep2"},
		Enabled:      true,
	}

	db := pluginToDBModel(p)

	if db.Dependencies != `["dep1","dep2"]` {
		t.Errorf("Dependencies mismatch: %s", db.Dependencies)
	}
}

func TestPluginToDBModel_EnvVars(t *testing.T) {
	p := &plugin.Plugin{
		Name: "example",
		Repo: "user/example",
		Env: map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
		},
		Enabled: true,
	}

	db := pluginToDBModel(p)

	if db.EnvVars == "{}" {
		t.Error("EnvVars should not be empty")
	}
}

func TestDBModelToPlugin_BasicFields(t *testing.T) {
	db := &models.TerminalPluginDB{
		ID:           1,
		Name:         "zsh-autosuggestions",
		Description:  sql.NullString{String: "Fish-like autosuggestions", Valid: true},
		Repo:         "zsh-users/zsh-autosuggestions",
		Category:     sql.NullString{String: "productivity", Valid: true},
		Shell:        "zsh",
		Manager:      "zinit",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	p := dbModelToPlugin(db)

	if p.Name != "zsh-autosuggestions" {
		t.Errorf("Name mismatch: %s", p.Name)
	}
	if p.Description != "Fish-like autosuggestions" {
		t.Errorf("Description mismatch: %s", p.Description)
	}
	if p.Repo != "zsh-users/zsh-autosuggestions" {
		t.Errorf("Repo mismatch: %s", p.Repo)
	}
	if p.Category != "productivity" {
		t.Errorf("Category mismatch: %s", p.Category)
	}
	if p.Manager != plugin.PluginManagerZinit {
		t.Errorf("Manager mismatch: %s", p.Manager)
	}
	if !p.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestDBModelToPlugin_OhMyZshPlugin(t *testing.T) {
	db := &models.TerminalPluginDB{
		ID:           1,
		Name:         "git",
		Repo:         "",
		Shell:        "zsh",
		Manager:      "oh-my-zsh",
		LoadCommand:  sql.NullString{String: "plugins+=git", Valid: true},
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	p := dbModelToPlugin(db)

	if p.OhMyZshPlugin != "git" {
		t.Errorf("OhMyZshPlugin mismatch: %s", p.OhMyZshPlugin)
	}
	if p.Config != "plugins+=git" {
		t.Errorf("Config mismatch: %s", p.Config)
	}
}

func TestConversion_Roundtrip(t *testing.T) {
	// Create a fully-populated plugin
	original := &plugin.Plugin{
		Name:         "zsh-autosuggestions",
		Description:  "Fish-like autosuggestions for zsh",
		Repo:         "zsh-users/zsh-autosuggestions",
		Branch:       "master",
		Tag:          "v0.7.0",
		Manager:      plugin.PluginManagerZinit,
		LoadMode:     plugin.LoadModeDeferred,
		SourceFiles:  []string{"zsh-autosuggestions.zsh"},
		Config:       "ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE='fg=8'",
		Env:          map[string]string{"ZSH_AUTOSUGGEST_STRATEGY": "history"},
		Dependencies: []string{"zsh-users/zsh-completions"},
		Category:     "productivity",
		Tags:         []string{"autosuggestions", "zsh"},
		Enabled:      true,
		Priority:     100,
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
	if converted.Manager != original.Manager {
		t.Errorf("Manager mismatch: got %s, want %s", converted.Manager, original.Manager)
	}
	if converted.Category != original.Category {
		t.Errorf("Category mismatch: got %s, want %s", converted.Category, original.Category)
	}
	if converted.Enabled != original.Enabled {
		t.Errorf("Enabled mismatch: got %v, want %v", converted.Enabled, original.Enabled)
	}

	// Check arrays
	if len(converted.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies length mismatch: got %d, want %d", len(converted.Dependencies), len(original.Dependencies))
	}
	if len(converted.Tags) != len(original.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(converted.Tags), len(original.Tags))
	}

	// Check environment variables
	if len(converted.Env) != len(original.Env) {
		t.Errorf("Env length mismatch: got %d, want %d", len(converted.Env), len(original.Env))
	}
}

// =============================================================================
// Through Interface Tests (PluginStore interface)
// =============================================================================

// TestPluginStore_ThroughInterface tests the DBPluginStore through
// the PluginStore interface to ensure polymorphism works correctly.
func TestPluginStore_ThroughInterface(t *testing.T) {
	mock := NewMockPluginDataStore()

	// Use through interface
	var store plugin.PluginStore = NewDBPluginStore(mock)

	// Test Create through interface
	p := &plugin.Plugin{
		Name:        "zsh-autosuggestions",
		Repo:        "zsh-users/zsh-autosuggestions",
		Description: "Fish-like autosuggestions",
		Category:    "productivity",
		Manager:     plugin.PluginManagerZinit,
		Enabled:     true,
	}

	err := store.Upsert(p)
	if err != nil {
		t.Fatalf("Upsert through interface failed: %v", err)
	}

	// Test Get through interface
	retrieved, err := store.Get("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("Get through interface failed: %v", err)
	}
	if retrieved.Name != "zsh-autosuggestions" {
		t.Errorf("Expected name zsh-autosuggestions, got %s", retrieved.Name)
	}

	// Test Exists through interface
	exists, err := store.Exists("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("Exists through interface failed: %v", err)
	}
	if !exists {
		t.Error("Expected zsh-autosuggestions to exist")
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
	err = store.Delete("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("Delete through interface failed: %v", err)
	}

	exists, _ = store.Exists("zsh-autosuggestions")
	if exists {
		t.Error("Expected zsh-autosuggestions to not exist after delete")
	}

	// Test Close through interface
	err = store.Close()
	if err != nil {
		t.Fatalf("Close through interface failed: %v", err)
	}
}
