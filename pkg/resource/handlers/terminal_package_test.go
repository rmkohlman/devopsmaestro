package handlers

import (
	"fmt"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	terminalpkg "devopsmaestro/pkg/terminalops/package"
)

// MockTerminalPackageDataStore implements the terminal package DataStore methods for testing.
type MockTerminalPackageDataStore struct {
	packages map[string]*models.TerminalPackageDB
	nextID   int
}

func NewMockTerminalPackageDataStore() *MockTerminalPackageDataStore {
	return &MockTerminalPackageDataStore{
		packages: make(map[string]*models.TerminalPackageDB),
		nextID:   1,
	}
}

func (m *MockTerminalPackageDataStore) CreateTerminalPackage(pkg *models.TerminalPackageDB) error {
	pkg.ID = m.nextID
	m.nextID++
	m.packages[pkg.Name] = pkg
	return nil
}

func (m *MockTerminalPackageDataStore) UpdateTerminalPackage(pkg *models.TerminalPackageDB) error {
	m.packages[pkg.Name] = pkg
	return nil
}

func (m *MockTerminalPackageDataStore) GetTerminalPackage(name string) (*models.TerminalPackageDB, error) {
	pkg, exists := m.packages[name]
	if !exists {
		return nil, fmt.Errorf("terminal package %q not found", name)
	}
	return pkg, nil
}

func (m *MockTerminalPackageDataStore) ListTerminalPackages() ([]*models.TerminalPackageDB, error) {
	var packages []*models.TerminalPackageDB
	for _, pkg := range m.packages {
		packages = append(packages, pkg)
	}
	return packages, nil
}

func (m *MockTerminalPackageDataStore) DeleteTerminalPackage(name string) error {
	delete(m.packages, name)
	return nil
}

func TestTerminalPackageHandler_Apply(t *testing.T) {
	handler := NewTerminalPackageHandler()
	mockStore := NewMockTerminalPackageDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: test-package
  description: Test terminal package
  category: test
  tags: [zsh, starship]
spec:
  extends: core
  plugins: [zsh-autosuggestions, zsh-syntax-highlighting]
  prompts: [starship-minimal]
  profiles: [developer]
  enabled: true
`)

	// Test apply (create)
	res, err := handler.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if res.GetKind() != KindTerminalPackage {
		t.Errorf("Expected kind %s, got %s", KindTerminalPackage, res.GetKind())
	}

	if res.GetName() != "test-package" {
		t.Errorf("Expected name 'test-package', got %s", res.GetName())
	}

	// Verify it was stored
	storedPkg, err := mockStore.GetTerminalPackage("test-package")
	if err != nil {
		t.Fatalf("Failed to get stored package: %v", err)
	}

	if storedPkg.Name != "test-package" {
		t.Errorf("Expected stored name 'test-package', got %s", storedPkg.Name)
	}

	plugins := storedPkg.GetPlugins()
	if len(plugins) != 2 || plugins[0] != "zsh-autosuggestions" || plugins[1] != "zsh-syntax-highlighting" {
		t.Errorf("Expected plugins [zsh-autosuggestions, zsh-syntax-highlighting], got %v", plugins)
	}

	prompts := storedPkg.GetPrompts()
	if len(prompts) != 1 || prompts[0] != "starship-minimal" {
		t.Errorf("Expected prompts [starship-minimal], got %v", prompts)
	}

	profiles := storedPkg.GetProfiles()
	if len(profiles) != 1 || profiles[0] != "developer" {
		t.Errorf("Expected profiles [developer], got %v", profiles)
	}
}

func TestTerminalPackageHandler_Get(t *testing.T) {
	handler := NewTerminalPackageHandler()
	mockStore := NewMockTerminalPackageDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Create a test package
	testPkg := &models.TerminalPackageDB{
		Name: "test-package",
	}
	testPkg.SetPlugins([]string{"zsh-autosuggestions"})
	testPkg.SetPrompts([]string{"starship-minimal"})
	testPkg.SetProfiles([]string{"developer"})

	err := mockStore.CreateTerminalPackage(testPkg)
	if err != nil {
		t.Fatalf("Failed to create test package: %v", err)
	}

	// Test get
	res, err := handler.Get(ctx, "test-package")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if res.GetName() != "test-package" {
		t.Errorf("Expected name 'test-package', got %s", res.GetName())
	}

	// Test get non-existent
	_, err = handler.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent package")
	}
}

func TestTerminalPackageHandler_List(t *testing.T) {
	handler := NewTerminalPackageHandler()
	mockStore := NewMockTerminalPackageDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Create test packages
	pkg1 := &models.TerminalPackageDB{Name: "pkg1"}
	pkg1.SetPlugins([]string{"plugin1"})
	pkg1.SetPrompts([]string{})
	pkg1.SetProfiles([]string{})

	pkg2 := &models.TerminalPackageDB{Name: "pkg2"}
	pkg2.SetPlugins([]string{"plugin2"})
	pkg2.SetPrompts([]string{})
	pkg2.SetProfiles([]string{})

	mockStore.CreateTerminalPackage(pkg1)
	mockStore.CreateTerminalPackage(pkg2)

	// Test list
	resources, err := handler.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(resources))
	}
}

func TestTerminalPackageHandler_Delete(t *testing.T) {
	handler := NewTerminalPackageHandler()
	mockStore := NewMockTerminalPackageDataStore()
	ctx := resource.Context{
		DataStore: mockStore,
	}

	// Create a test package
	testPkg := &models.TerminalPackageDB{Name: "test-package"}
	testPkg.SetPlugins([]string{"plugin1"})
	testPkg.SetPrompts([]string{})
	testPkg.SetProfiles([]string{})
	mockStore.CreateTerminalPackage(testPkg)

	// Test delete
	err := handler.Delete(ctx, "test-package")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = mockStore.GetTerminalPackage("test-package")
	if err == nil {
		t.Error("Expected error when getting deleted package")
	}
}

func TestTerminalPackageHandler_ToYAML(t *testing.T) {
	handler := NewTerminalPackageHandler()

	pkg := terminalpkg.NewPackage("test-package")
	pkg.Description = "Test package"
	pkg.Plugins = []string{"zsh-autosuggestions"}
	pkg.Prompts = []string{"starship-minimal"}

	res := &TerminalPackageResource{pkg: pkg}

	yamlBytes, err := handler.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Error("Expected non-empty YAML output")
	}

	// Verify we can parse it back
	parsedPkg, err := terminalpkg.ParseYAML(yamlBytes)
	if err != nil {
		t.Fatalf("Failed to parse generated YAML: %v", err)
	}

	if parsedPkg.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got %s", parsedPkg.Name)
	}
}
