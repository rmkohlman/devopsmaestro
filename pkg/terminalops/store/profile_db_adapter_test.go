package store

import (
	"fmt"
	"testing"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/profile"
)

// =============================================================================
// Mock Implementation of ProfileDataStore
// =============================================================================

// MockProfileDataStore implements ProfileDataStore for testing.
type MockProfileDataStore struct {
	profiles map[string]*models.TerminalProfileDB
	calls    []string
	err      error // inject errors for testing
}

func NewMockProfileDataStore() *MockProfileDataStore {
	return &MockProfileDataStore{
		profiles: make(map[string]*models.TerminalProfileDB),
	}
}

func (m *MockProfileDataStore) CreateTerminalProfile(p *models.TerminalProfileDB) error {
	m.calls = append(m.calls, "CreateTerminalProfile")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.profiles[p.Name]; exists {
		return fmt.Errorf("profile already exists: %s", p.Name)
	}
	p.ID = len(m.profiles) + 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.profiles[p.Name] = p
	return nil
}

func (m *MockProfileDataStore) GetTerminalProfileByName(name string) (*models.TerminalProfileDB, error) {
	m.calls = append(m.calls, "GetTerminalProfileByName")
	if m.err != nil {
		return nil, m.err
	}
	p, exists := m.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", name)
	}
	return p, nil
}

func (m *MockProfileDataStore) UpdateTerminalProfile(p *models.TerminalProfileDB) error {
	m.calls = append(m.calls, "UpdateTerminalProfile")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.profiles[p.Name]; !exists {
		return fmt.Errorf("profile not found: %s", p.Name)
	}
	p.UpdatedAt = time.Now()
	m.profiles[p.Name] = p
	return nil
}

func (m *MockProfileDataStore) UpsertTerminalProfile(p *models.TerminalProfileDB) error {
	m.calls = append(m.calls, "UpsertTerminalProfile")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.profiles[p.Name]; exists {
		p.UpdatedAt = time.Now()
	} else {
		p.ID = len(m.profiles) + 1
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
	}
	m.profiles[p.Name] = p
	return nil
}

func (m *MockProfileDataStore) DeleteTerminalProfile(name string) error {
	m.calls = append(m.calls, "DeleteTerminalProfile")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}
	delete(m.profiles, name)
	return nil
}

func (m *MockProfileDataStore) ListTerminalProfiles() ([]*models.TerminalProfileDB, error) {
	m.calls = append(m.calls, "ListTerminalProfiles")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalProfileDB
	for _, p := range m.profiles {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockProfileDataStore) ListTerminalProfilesByCategory(category string) ([]*models.TerminalProfileDB, error) {
	m.calls = append(m.calls, "ListTerminalProfilesByCategory")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalProfileDB
	for _, p := range m.profiles {
		if p.Category.Valid && p.Category.String == category {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockProfileDataStore) SetError(err error) {
	m.err = err
}

func (m *MockProfileDataStore) GetCalls() []string {
	return m.calls
}

// =============================================================================
// Interface Verification
// =============================================================================

func TestDBProfileStore_ImplementsProfileStore(t *testing.T) {
	var _ profile.ProfileStore = (*DBProfileStore)(nil)
}

func TestMockProfileDataStore_ImplementsProfileDataStore(t *testing.T) {
	var _ ProfileDataStore = (*MockProfileDataStore)(nil)
}

// =============================================================================
// Basic CRUD Tests
// =============================================================================

func TestDBProfileStore_Create(t *testing.T) {
	mockStore := NewMockProfileDataStore()
	adapter := NewDBProfileStore(mockStore)

	testProfile := &profile.Profile{
		Name:        "test-profile",
		Description: "Test profile",
		PromptRef:   "starship-default",
		PluginRefs:  []string{"fzf", "zsh-autosuggestions"},
		Enabled:     true,
	}

	err := adapter.Create(testProfile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the call was made
	calls := mockStore.GetCalls()
	if len(calls) != 2 { // GetTerminalProfileByName (for exists check) + CreateTerminalProfile
		t.Fatalf("Expected 2 calls, got %d", len(calls))
	}
	if calls[1] != "CreateTerminalProfile" {
		t.Fatalf("Expected CreateTerminalProfile call, got %s", calls[1])
	}
}

func TestDBProfileStore_Get(t *testing.T) {
	mockStore := NewMockProfileDataStore()
	adapter := NewDBProfileStore(mockStore)

	// Create a profile in the mock store
	dbProfile := &models.TerminalProfileDB{
		Name:       "test-profile",
		PluginRefs: `["fzf", "zsh-autosuggestions"]`,
		Enabled:    true,
	}
	mockStore.CreateTerminalProfile(dbProfile)

	// Get the profile through the adapter
	result, err := adapter.Get("test-profile")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "test-profile" {
		t.Fatalf("Expected name 'test-profile', got '%s'", result.Name)
	}
	if len(result.PluginRefs) != 2 {
		t.Fatalf("Expected 2 plugin refs, got %d", len(result.PluginRefs))
	}
}

func TestDBProfileStore_Exists(t *testing.T) {
	mockStore := NewMockProfileDataStore()
	adapter := NewDBProfileStore(mockStore)

	// Profile doesn't exist yet
	exists, err := adapter.Exists("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Fatal("Expected profile to not exist")
	}

	// Create a profile
	dbProfile := &models.TerminalProfileDB{
		Name:    "test-profile",
		Enabled: true,
	}
	mockStore.CreateTerminalProfile(dbProfile)

	// Now it should exist
	exists, err = adapter.Exists("test-profile")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Fatal("Expected profile to exist")
	}
}
