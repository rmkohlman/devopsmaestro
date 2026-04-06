package handlers

import (
	"fmt"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalbridge"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// KindTerminalPlugin is the resource kind identifier for terminal plugins.
// NOTE: MaestroTerminal plugin package uses the string "TerminalPlugin" directly
// (no exported constant), so we define it here for test clarity.
const testKindTerminalPlugin = "TerminalPlugin"

// =============================================================================
// MockTerminalPluginDataStore — implements terminalbridge.PluginDataStore
// =============================================================================

// MockTerminalPluginDataStore implements terminalbridge.PluginDataStore for testing.
type MockTerminalPluginDataStore struct {
	plugins map[string]*models.TerminalPluginDB
	nextID  int
}

func NewMockTerminalPluginDataStore() *MockTerminalPluginDataStore {
	return &MockTerminalPluginDataStore{
		plugins: make(map[string]*models.TerminalPluginDB),
		nextID:  1,
	}
}

func (m *MockTerminalPluginDataStore) CreateTerminalPlugin(p *models.TerminalPluginDB) error {
	if _, exists := m.plugins[p.Name]; exists {
		return fmt.Errorf("terminal plugin %q already exists", p.Name)
	}
	p.ID = m.nextID
	m.nextID++
	m.plugins[p.Name] = p
	return nil
}

func (m *MockTerminalPluginDataStore) GetTerminalPlugin(name string) (*models.TerminalPluginDB, error) {
	p, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("terminal plugin %q not found", name)
	}
	return p, nil
}

func (m *MockTerminalPluginDataStore) UpdateTerminalPlugin(p *models.TerminalPluginDB) error {
	if _, exists := m.plugins[p.Name]; !exists {
		return fmt.Errorf("terminal plugin %q not found", p.Name)
	}
	m.plugins[p.Name] = p
	return nil
}

func (m *MockTerminalPluginDataStore) UpsertTerminalPlugin(p *models.TerminalPluginDB) error {
	if existing, exists := m.plugins[p.Name]; exists {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
	} else {
		p.ID = m.nextID
		m.nextID++
	}
	m.plugins[p.Name] = p
	return nil
}

func (m *MockTerminalPluginDataStore) DeleteTerminalPlugin(name string) error {
	if _, exists := m.plugins[name]; !exists {
		return fmt.Errorf("terminal plugin %q not found", name)
	}
	delete(m.plugins, name)
	return nil
}

func (m *MockTerminalPluginDataStore) ListTerminalPlugins() ([]*models.TerminalPluginDB, error) {
	var result []*models.TerminalPluginDB
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockTerminalPluginDataStore) ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error) {
	var result []*models.TerminalPluginDB
	for _, p := range m.plugins {
		if p.Category.Valid && p.Category.String == category {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockTerminalPluginDataStore) ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error) {
	var result []*models.TerminalPluginDB
	for _, p := range m.plugins {
		if p.Shell == shell {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockTerminalPluginDataStore) Close() error {
	return nil
}

// Verify MockTerminalPluginDataStore satisfies the PluginDataStore interface at compile time.
var _ terminalbridge.PluginDataStore = (*MockTerminalPluginDataStore)(nil)

// =============================================================================
// Tests
// =============================================================================

// TestTerminalPluginHandler_Apply verifies that a TerminalPlugin YAML can be
// applied via the handler and persisted in the DB store.
func TestTerminalPluginHandler_Apply(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-autosuggestions
  description: Fish-like autosuggestions for zsh
  category: productivity
spec:
  repo: zsh-users/zsh-autosuggestions
  shell: zsh
  manager: manual
`)

	res, err := handler.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if res.GetKind() != testKindTerminalPlugin {
		t.Errorf("expected kind %q, got %q", testKindTerminalPlugin, res.GetKind())
	}
	if res.GetName() != "zsh-autosuggestions" {
		t.Errorf("expected name %q, got %q", "zsh-autosuggestions", res.GetName())
	}

	// Verify the plugin was persisted in the mock store
	stored, err := mockStore.GetTerminalPlugin("zsh-autosuggestions")
	if err != nil {
		t.Fatalf("plugin was not stored: %v", err)
	}
	if stored.Name != "zsh-autosuggestions" {
		t.Errorf("stored plugin name mismatch: got %q", stored.Name)
	}
}

// TestTerminalPluginHandler_Apply_Update verifies that applying a plugin that
// already exists performs an upsert (update) rather than failing.
func TestTerminalPluginHandler_Apply_Update(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-syntax-highlighting
spec:
  repo: zsh-users/zsh-syntax-highlighting
  shell: zsh
  manager: manual
`)

	// First apply — creates
	_, err := handler.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("first Apply failed: %v", err)
	}

	// Second apply — should upsert without error
	_, err = handler.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("second Apply (upsert) failed: %v", err)
	}

	// Should still be exactly one entry
	all, err := mockStore.ListTerminalPlugins()
	if err != nil {
		t.Fatalf("ListTerminalPlugins failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 plugin after upsert, got %d", len(all))
	}
}

// TestTerminalPluginHandler_Get verifies that a stored terminal plugin can be
// retrieved by name and returns the correct resource.
func TestTerminalPluginHandler_Get(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Pre-populate the store
	mockStore.plugins["my-plugin"] = &models.TerminalPluginDB{
		ID:           1,
		Name:         "my-plugin",
		Repo:         "example/my-plugin",
		Shell:        "zsh",
		Manager:      "manual",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	res, err := handler.Get(ctx, "my-plugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if res.GetName() != "my-plugin" {
		t.Errorf("expected name %q, got %q", "my-plugin", res.GetName())
	}
	if res.GetKind() != testKindTerminalPlugin {
		t.Errorf("expected kind %q, got %q", testKindTerminalPlugin, res.GetKind())
	}

	// Not-found case
	_, err = handler.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent plugin, got nil")
	}
}

// TestTerminalPluginHandler_List verifies that multiple stored terminal plugins
// are all returned by the handler's List method.
func TestTerminalPluginHandler_List(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Pre-populate with two plugins
	for _, name := range []string{"plugin-a", "plugin-b"} {
		mockStore.plugins[name] = &models.TerminalPluginDB{
			ID:           mockStore.nextID,
			Name:         name,
			Repo:         "example/" + name,
			Shell:        "zsh",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		}
		mockStore.nextID++
	}

	resources, err := handler.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}

	// Verify all returned resources have the correct kind
	for _, res := range resources {
		if res.GetKind() != testKindTerminalPlugin {
			t.Errorf("unexpected kind %q in list result", res.GetKind())
		}
	}
}

// TestTerminalPluginHandler_List_Empty verifies that List returns an empty
// slice (not an error) when no plugins exist.
func TestTerminalPluginHandler_List_Empty(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	resources, err := handler.List(ctx)
	if err != nil {
		t.Fatalf("List on empty store failed: %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

// TestTerminalPluginHandler_Delete verifies that a stored terminal plugin is
// removed after calling Delete.
func TestTerminalPluginHandler_Delete(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Pre-populate
	mockStore.plugins["delete-me"] = &models.TerminalPluginDB{
		ID:           1,
		Name:         "delete-me",
		Repo:         "example/delete-me",
		Shell:        "zsh",
		Manager:      "manual",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	err := handler.Delete(ctx, "delete-me")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	if _, exists := mockStore.plugins["delete-me"]; exists {
		t.Error("plugin was not deleted from mock store")
	}
}

// TestTerminalPluginHandler_ToYAML verifies that a TerminalPluginResource can be
// serialized to YAML with the expected kind and name fields.
func TestTerminalPluginHandler_ToYAML(t *testing.T) {
	handler := NewTerminalPluginHandler()

	// Build the resource directly using the handler's resource wrapper
	res := NewTerminalPluginResource(&models.TerminalPluginDB{
		ID:           1,
		Name:         "zsh-autosuggestions",
		Repo:         "zsh-users/zsh-autosuggestions",
		Shell:        "zsh",
		Manager:      "manual",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	})

	yamlBytes, err := handler.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Error("ToYAML returned empty bytes")
	}

	yamlStr := string(yamlBytes)

	// Verify basic structure is present
	if !containsSubstring(yamlStr, "TerminalPlugin") {
		t.Error("YAML output missing 'TerminalPlugin' kind")
	}
	if !containsSubstring(yamlStr, "zsh-autosuggestions") {
		t.Error("YAML output missing plugin name")
	}
}

// TestTerminalPluginHandler_RoundTrip verifies the full backup→restore cycle:
// Apply YAML → export via List+ToYAML → re-apply exported YAML → verify restored.
func TestTerminalPluginHandler_RoundTrip(t *testing.T) {
	handler := NewTerminalPluginHandler()
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	originalYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-history-substring-search
  description: Fish-like history search for zsh
  category: ux
spec:
  repo: zsh-users/zsh-history-substring-search
  shell: zsh
  manager: zinit
`)

	// Step 1: Apply original YAML
	_, err := handler.Apply(ctx, originalYAML)
	if err != nil {
		t.Fatalf("initial Apply failed: %v", err)
	}

	// Step 2: Export via List + ToYAML
	resources, err := handler.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource after apply, got %d", len(resources))
	}

	exportedYAML, err := handler.ToYAML(resources[0])
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	// Step 3: Wipe and restore
	delete(mockStore.plugins, "zsh-history-substring-search")

	_, err = handler.Apply(ctx, exportedYAML)
	if err != nil {
		t.Fatalf("restore Apply failed: %v", err)
	}

	// Step 4: Verify restored correctly
	restored, err := handler.Get(ctx, "zsh-history-substring-search")
	if err != nil {
		t.Fatalf("Get after restore failed: %v", err)
	}

	if restored.GetName() != "zsh-history-substring-search" {
		t.Errorf("restored plugin name mismatch: got %q", restored.GetName())
	}
	if restored.GetKind() != testKindTerminalPlugin {
		t.Errorf("restored plugin kind mismatch: got %q", restored.GetKind())
	}
}

// TestTerminalPluginHandler_Registered verifies that the TerminalPlugin handler
// is registered in the resource system under the "TerminalPlugin" kind.
// This test drives registration in register.go.
func TestTerminalPluginHandler_Registered(t *testing.T) {
	// Ensure all handlers are registered
	RegisterAll()

	// Attempt to retrieve the handler for "TerminalPlugin" kind.
	// resource.Get is the lookup mechanism — if the kind isn't registered,
	// it will return an error.
	mockStore := NewMockTerminalPluginDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// List with the TerminalPlugin kind — if the handler isn't registered
	// this will return an error from the resource registry.
	_, err := resource.List(ctx, testKindTerminalPlugin)
	if err != nil {
		t.Errorf("TerminalPlugin handler not registered in resource system: %v", err)
	}
}

// TestTerminalPluginResource_Validate verifies the resource Validate() method
// correctly catches invalid plugin configurations.
func TestTerminalPluginResource_Validate(t *testing.T) {
	tests := []struct {
		name        string
		plugin      *models.TerminalPluginDB
		expectError bool
	}{
		{
			name: "valid plugin with repo",
			plugin: &models.TerminalPluginDB{
				Name:    "zsh-autosuggestions",
				Repo:    "zsh-users/zsh-autosuggestions",
				Shell:   "zsh",
				Manager: "manual",
			},
			expectError: false,
		},
		{
			name: "missing name",
			plugin: &models.TerminalPluginDB{
				Repo:    "zsh-users/zsh-autosuggestions",
				Shell:   "zsh",
				Manager: "manual",
			},
			expectError: true,
		},
		{
			name: "missing repo",
			plugin: &models.TerminalPluginDB{
				Name:    "some-plugin",
				Shell:   "zsh",
				Manager: "manual",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewTerminalPluginResource(tt.plugin)
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
