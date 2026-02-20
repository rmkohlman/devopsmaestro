package handlers

import (
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/terminalops/prompt"
)

// MockTerminalPromptDataStore implements prompt.PromptDataStore for testing.
type MockTerminalPromptDataStore struct {
	prompts map[string]*models.TerminalPromptDB
	nextID  int
}

func NewMockTerminalPromptDataStore() *MockTerminalPromptDataStore {
	return &MockTerminalPromptDataStore{
		prompts: make(map[string]*models.TerminalPromptDB),
		nextID:  1,
	}
}

func (m *MockTerminalPromptDataStore) CreateTerminalPrompt(p *models.TerminalPromptDB) error {
	p.ID = m.nextID
	m.nextID++
	m.prompts[p.Name] = p
	return nil
}

func (m *MockTerminalPromptDataStore) GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error) {
	p, exists := m.prompts[name]
	if !exists {
		return nil, &prompt.ErrNotFound{Name: name}
	}
	return p, nil
}

func (m *MockTerminalPromptDataStore) UpdateTerminalPrompt(p *models.TerminalPromptDB) error {
	if _, exists := m.prompts[p.Name]; !exists {
		return &prompt.ErrNotFound{Name: p.Name}
	}
	m.prompts[p.Name] = p
	return nil
}

func (m *MockTerminalPromptDataStore) DeleteTerminalPrompt(name string) error {
	if _, exists := m.prompts[name]; !exists {
		return &prompt.ErrNotFound{Name: name}
	}
	delete(m.prompts, name)
	return nil
}

func (m *MockTerminalPromptDataStore) ListTerminalPrompts() ([]*models.TerminalPromptDB, error) {
	var result []*models.TerminalPromptDB
	for _, p := range m.prompts {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockTerminalPromptDataStore) ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error) {
	var result []*models.TerminalPromptDB
	for _, p := range m.prompts {
		if p.Type == promptType {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockTerminalPromptDataStore) ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error) {
	var result []*models.TerminalPromptDB
	for _, p := range m.prompts {
		if p.Category.Valid && p.Category.String == category {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockTerminalPromptDataStore) Close() error {
	return nil
}

// Verify MockTerminalPromptDataStore implements prompt.PromptDataStore
var _ prompt.PromptDataStore = (*MockTerminalPromptDataStore)(nil)

func TestTerminalPromptHandler_Kind(t *testing.T) {
	h := NewTerminalPromptHandler()
	if h.Kind() != prompt.KindTerminalPrompt {
		t.Errorf("expected kind %q, got %q", prompt.KindTerminalPrompt, h.Kind())
	}
}

func TestTerminalPromptHandler_Apply(t *testing.T) {
	h := NewTerminalPromptHandler()
	mockStore := NewMockTerminalPromptDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: test-prompt
  description: A test prompt
spec:
  type: starship
  addNewline: true
  palette: theme
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if res.GetKind() != prompt.KindTerminalPrompt {
		t.Errorf("expected kind %q, got %q", prompt.KindTerminalPrompt, res.GetKind())
	}
	if res.GetName() != "test-prompt" {
		t.Errorf("expected name %q, got %q", "test-prompt", res.GetName())
	}

	// Verify the prompt was stored
	if _, exists := mockStore.prompts["test-prompt"]; !exists {
		t.Error("prompt was not stored in mock store")
	}
}

func TestTerminalPromptHandler_Get(t *testing.T) {
	h := NewTerminalPromptHandler()
	mockStore := NewMockTerminalPromptDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Pre-populate the store
	mockStore.prompts["my-prompt"] = &models.TerminalPromptDB{
		ID:      1,
		Name:    "my-prompt",
		Type:    "starship",
		Enabled: true,
	}

	res, err := h.Get(ctx, "my-prompt")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if res.GetName() != "my-prompt" {
		t.Errorf("expected name %q, got %q", "my-prompt", res.GetName())
	}
}

func TestTerminalPromptHandler_List(t *testing.T) {
	h := NewTerminalPromptHandler()
	mockStore := NewMockTerminalPromptDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Pre-populate the store
	mockStore.prompts["prompt-1"] = &models.TerminalPromptDB{
		ID:      1,
		Name:    "prompt-1",
		Type:    "starship",
		Enabled: true,
	}
	mockStore.prompts["prompt-2"] = &models.TerminalPromptDB{
		ID:      2,
		Name:    "prompt-2",
		Type:    "oh-my-posh",
		Enabled: true,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

func TestTerminalPromptHandler_Delete(t *testing.T) {
	h := NewTerminalPromptHandler()
	mockStore := NewMockTerminalPromptDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Pre-populate the store
	mockStore.prompts["delete-me"] = &models.TerminalPromptDB{
		ID:      1,
		Name:    "delete-me",
		Type:    "starship",
		Enabled: true,
	}

	err := h.Delete(ctx, "delete-me")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, exists := mockStore.prompts["delete-me"]; exists {
		t.Error("prompt was not deleted from mock store")
	}
}

func TestTerminalPromptHandler_ToYAML(t *testing.T) {
	h := NewTerminalPromptHandler()

	res := &TerminalPromptResource{
		prompt: &prompt.Prompt{
			Name:    "yaml-test",
			Type:    prompt.PromptTypeStarship,
			Enabled: true,
		},
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	yamlStr := string(yamlBytes)
	if len(yamlStr) == 0 {
		t.Error("ToYAML returned empty bytes")
	}

	// Verify basic structure
	if !containsSubstring(yamlStr, "kind: TerminalPrompt") {
		t.Error("YAML output missing kind field")
	}
	if !containsSubstring(yamlStr, "name: yaml-test") {
		t.Error("YAML output missing name field")
	}
}

func TestTerminalPromptResource_Validate(t *testing.T) {
	tests := []struct {
		name        string
		prompt      *prompt.Prompt
		expectError bool
	}{
		{
			name: "valid prompt",
			prompt: &prompt.Prompt{
				Name: "test",
				Type: prompt.PromptTypeStarship,
			},
			expectError: false,
		},
		{
			name: "missing name",
			prompt: &prompt.Prompt{
				Type: prompt.PromptTypeStarship,
			},
			expectError: true,
		},
		{
			name: "missing type",
			prompt: &prompt.Prompt{
				Name: "test",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &TerminalPromptResource{prompt: tt.prompt}
			err := res.Validate()
			if tt.expectError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
