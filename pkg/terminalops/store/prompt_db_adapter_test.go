package store

import (
	"fmt"
	"testing"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalops/prompt"
)

// =============================================================================
// Mock Implementation of PromptDataStore
// =============================================================================

// MockPromptDataStore implements PromptDataStore for testing.
type MockPromptDataStore struct {
	prompts map[string]*models.TerminalPromptDB
	calls   []string
	err     error // inject errors for testing
}

func NewMockPromptDataStore() *MockPromptDataStore {
	return &MockPromptDataStore{
		prompts: make(map[string]*models.TerminalPromptDB),
	}
}

func (m *MockPromptDataStore) CreateTerminalPrompt(p *models.TerminalPromptDB) error {
	m.calls = append(m.calls, "CreateTerminalPrompt")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.prompts[p.Name]; exists {
		return fmt.Errorf("prompt already exists: %s", p.Name)
	}
	p.ID = len(m.prompts) + 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.prompts[p.Name] = p
	return nil
}

func (m *MockPromptDataStore) GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error) {
	m.calls = append(m.calls, "GetTerminalPromptByName")
	if m.err != nil {
		return nil, m.err
	}
	p, exists := m.prompts[name]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}
	return p, nil
}

func (m *MockPromptDataStore) UpdateTerminalPrompt(p *models.TerminalPromptDB) error {
	m.calls = append(m.calls, "UpdateTerminalPrompt")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.prompts[p.Name]; !exists {
		return fmt.Errorf("prompt not found: %s", p.Name)
	}
	p.UpdatedAt = time.Now()
	m.prompts[p.Name] = p
	return nil
}

func (m *MockPromptDataStore) UpsertTerminalPrompt(p *models.TerminalPromptDB) error {
	m.calls = append(m.calls, "UpsertTerminalPrompt")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.prompts[p.Name]; exists {
		p.UpdatedAt = time.Now()
	} else {
		p.ID = len(m.prompts) + 1
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
	}
	m.prompts[p.Name] = p
	return nil
}

func (m *MockPromptDataStore) DeleteTerminalPrompt(name string) error {
	m.calls = append(m.calls, "DeleteTerminalPrompt")
	if m.err != nil {
		return m.err
	}
	if _, exists := m.prompts[name]; !exists {
		return fmt.Errorf("prompt not found: %s", name)
	}
	delete(m.prompts, name)
	return nil
}

func (m *MockPromptDataStore) ListTerminalPrompts() ([]*models.TerminalPromptDB, error) {
	m.calls = append(m.calls, "ListTerminalPrompts")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalPromptDB
	for _, p := range m.prompts {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockPromptDataStore) ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error) {
	m.calls = append(m.calls, "ListTerminalPromptsByType")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalPromptDB
	for _, p := range m.prompts {
		if p.Type == promptType {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockPromptDataStore) ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error) {
	m.calls = append(m.calls, "ListTerminalPromptsByCategory")
	if m.err != nil {
		return nil, m.err
	}
	var result []*models.TerminalPromptDB
	for _, p := range m.prompts {
		if p.Category.Valid && p.Category.String == category {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockPromptDataStore) SetError(err error) {
	m.err = err
}

func (m *MockPromptDataStore) GetCalls() []string {
	return m.calls
}

// =============================================================================
// Interface Verification
// =============================================================================

func TestDBPromptStore_ImplementsPromptStore(t *testing.T) {
	var _ prompt.PromptStore = (*DBPromptStore)(nil)
}

func TestMockPromptDataStore_ImplementsPromptDataStore(t *testing.T) {
	var _ PromptDataStore = (*MockPromptDataStore)(nil)
}

// =============================================================================
// Basic CRUD Tests
// =============================================================================

func TestDBPromptStore_Create(t *testing.T) {
	mockStore := NewMockPromptDataStore()
	adapter := NewDBPromptStore(mockStore)

	testPrompt := &prompt.Prompt{
		Name:        "test-prompt",
		Type:        prompt.PromptTypeStarship,
		Description: "Test prompt",
		Enabled:     true,
	}

	err := adapter.Create(testPrompt)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the call was made
	calls := mockStore.GetCalls()
	if len(calls) != 2 { // GetTerminalPromptByName (for exists check) + CreateTerminalPrompt
		t.Fatalf("Expected 2 calls, got %d", len(calls))
	}
	if calls[1] != "CreateTerminalPrompt" {
		t.Fatalf("Expected CreateTerminalPrompt call, got %s", calls[1])
	}
}

func TestDBPromptStore_Get(t *testing.T) {
	mockStore := NewMockPromptDataStore()
	adapter := NewDBPromptStore(mockStore)

	// Create a prompt in the mock store
	dbPrompt := &models.TerminalPromptDB{
		Name:    "test-prompt",
		Type:    "starship",
		Enabled: true,
	}
	mockStore.CreateTerminalPrompt(dbPrompt)

	// Get the prompt through the adapter
	result, err := adapter.Get("test-prompt")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "test-prompt" {
		t.Fatalf("Expected name 'test-prompt', got '%s'", result.Name)
	}
	if result.Type != prompt.PromptTypeStarship {
		t.Fatalf("Expected type starship, got %s", result.Type)
	}
}

func TestDBPromptStore_Exists(t *testing.T) {
	mockStore := NewMockPromptDataStore()
	adapter := NewDBPromptStore(mockStore)

	// Prompt doesn't exist yet
	exists, err := adapter.Exists("nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Fatal("Expected prompt to not exist")
	}

	// Create a prompt
	dbPrompt := &models.TerminalPromptDB{
		Name:    "test-prompt",
		Type:    "starship",
		Enabled: true,
	}
	mockStore.CreateTerminalPrompt(dbPrompt)

	// Now it should exist
	exists, err = adapter.Exists("test-prompt")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Fatal("Expected prompt to exist")
	}
}
