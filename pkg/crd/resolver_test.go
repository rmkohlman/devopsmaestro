package crd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Fixtures
// =============================================================================

func createTestCRD(kind, singular, plural string, shortNames []string) *CRDDefinition {
	return &CRDDefinition{
		Group: "devopsmaestro.io",
		Names: CRDNames{
			Kind:       kind,
			Singular:   singular,
			Plural:     plural,
			ShortNames: shortNames,
		},
		Scope: ScopeWorkspace,
		Versions: []CRDVersion{
			{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
				Schema: CRDSchema{
					OpenAPIV3Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"spec": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
			},
		},
	}
}

// =============================================================================
// Mock CRDStore
// =============================================================================

type MockCRDStore struct {
	crds   map[string]*CRDDefinition
	listFn func() ([]*CRDDefinition, error)
}

func NewMockCRDStore() *MockCRDStore {
	return &MockCRDStore{
		crds: make(map[string]*CRDDefinition),
	}
}

func (m *MockCRDStore) CreateCRD(crd *CRDDefinition) error {
	if _, exists := m.crds[crd.Names.Kind]; exists {
		return &CRDAlreadyExistsError{Kind: crd.Names.Kind}
	}
	m.crds[crd.Names.Kind] = crd
	return nil
}

func (m *MockCRDStore) GetCRD(kind string) (*CRDDefinition, error) {
	crd, exists := m.crds[kind]
	if !exists {
		return nil, &CRDNotFoundError{Kind: kind}
	}
	return crd, nil
}

func (m *MockCRDStore) ListCRDs() ([]*CRDDefinition, error) {
	if m.listFn != nil {
		return m.listFn()
	}
	crds := make([]*CRDDefinition, 0, len(m.crds))
	for _, crd := range m.crds {
		crds = append(crds, crd)
	}
	return crds, nil
}

func (m *MockCRDStore) UpdateCRD(crd *CRDDefinition) error {
	if _, exists := m.crds[crd.Names.Kind]; !exists {
		return &CRDNotFoundError{Kind: crd.Names.Kind}
	}
	m.crds[crd.Names.Kind] = crd
	return nil
}

func (m *MockCRDStore) DeleteCRD(kind string) error {
	if _, exists := m.crds[kind]; !exists {
		return &CRDNotFoundError{Kind: kind}
	}
	delete(m.crds, kind)
	return nil
}

// =============================================================================
// Mock CRDResolver (for testing)
// =============================================================================

type MockCRDResolver struct {
	crds map[string]*CRDDefinition
}

func NewMockCRDResolver() *MockCRDResolver {
	return &MockCRDResolver{
		crds: make(map[string]*CRDDefinition),
	}
}

func (m *MockCRDResolver) Resolve(name string) *CRDDefinition {
	// Case-insensitive lookup by kind, singular, plural, or short name
	lowerName := toLower(name)
	for _, crd := range m.crds {
		if toLower(crd.Names.Kind) == lowerName ||
			toLower(crd.Names.Singular) == lowerName ||
			toLower(crd.Names.Plural) == lowerName {
			return crd
		}
		for _, short := range crd.Names.ShortNames {
			if toLower(short) == lowerName {
				return crd
			}
		}
	}
	return nil
}

func (m *MockCRDResolver) Register(crd *CRDDefinition) error {
	if _, exists := m.crds[crd.Names.Kind]; exists {
		return &CRDAlreadyExistsError{Kind: crd.Names.Kind}
	}
	m.crds[crd.Names.Kind] = crd
	return nil
}

func (m *MockCRDResolver) Unregister(kind string) error {
	if _, exists := m.crds[kind]; !exists {
		return &CRDNotFoundError{Kind: kind}
	}
	delete(m.crds, kind)
	return nil
}

func (m *MockCRDResolver) List() []*CRDDefinition {
	crds := make([]*CRDDefinition, 0, len(m.crds))
	for _, crd := range m.crds {
		crds = append(crds, crd)
	}
	return crds
}

func (m *MockCRDResolver) Refresh() error {
	// Mock refresh does nothing
	return nil
}

// Helper function for case-insensitive comparison
func toLower(s string) string {
	return strings.ToLower(s)
}

// =============================================================================
// CRDResolver Tests - Resolution
// =============================================================================

func TestResolver_Resolve_ByKind(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	resolved := resolver.Resolve("Database")
	assert.NotNil(t, resolved, "Should resolve by Kind")
	assert.Equal(t, "Database", resolved.Names.Kind)
}

func TestResolver_Resolve_BySingular(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	resolved := resolver.Resolve("database")
	assert.NotNil(t, resolved, "Should resolve by singular name")
	assert.Equal(t, "Database", resolved.Names.Kind)
}

func TestResolver_Resolve_ByPlural(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	resolved := resolver.Resolve("databases")
	assert.NotNil(t, resolved, "Should resolve by plural name")
	assert.Equal(t, "Database", resolved.Names.Kind)
}

func TestResolver_Resolve_ByShortName(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	resolved := resolver.Resolve("db")
	assert.NotNil(t, resolved, "Should resolve by short name")
	assert.Equal(t, "Database", resolved.Names.Kind)
}

func TestResolver_Resolve_CaseInsensitive(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	tests := []struct {
		name  string
		input string
	}{
		{"uppercase kind", "DATABASE"},
		{"mixed case kind", "DataBase"},
		{"uppercase singular", "DATABASE"},
		{"uppercase plural", "DATABASES"},
		{"uppercase short", "DB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := resolver.Resolve(tt.input)
			assert.NotNil(t, resolved, "Should resolve case-insensitively")
			assert.Equal(t, "Database", resolved.Names.Kind)
		})
	}
}

func TestResolver_Resolve_UnknownReturnsNil(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	resolved := resolver.Resolve("Unknown")
	assert.Nil(t, resolved, "Should return nil for unknown kind")
}

func TestResolver_Resolve_MultipleShortNames(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db", "dbase"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	resolved1 := resolver.Resolve("db")
	assert.NotNil(t, resolved1)
	assert.Equal(t, "Database", resolved1.Names.Kind)

	resolved2 := resolver.Resolve("dbase")
	assert.NotNil(t, resolved2)
	assert.Equal(t, "Database", resolved2.Names.Kind)
}

// =============================================================================
// CRDResolver Tests - Registration
// =============================================================================

func TestResolver_Register_Success(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	assert.NoError(t, err)

	resolved := resolver.Resolve("Database")
	assert.NotNil(t, resolved)
	assert.Equal(t, "Database", resolved.Names.Kind)
}

func TestResolver_Register_DuplicateKindFails(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd1 := createTestCRD("Database", "database", "databases", []string{"db"})
	crd2 := createTestCRD("Database", "database", "databases", []string{"db"})

	err1 := resolver.Register(crd1)
	assert.NoError(t, err1)

	err2 := resolver.Register(crd2)
	assert.Error(t, err2)
	assert.IsType(t, &CRDAlreadyExistsError{}, err2)
}

func TestResolver_Register_MultipleCRDs(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd1 := createTestCRD("Database", "database", "databases", []string{"db"})
	crd2 := createTestCRD("Cache", "cache", "caches", []string{"ch"})

	err1 := resolver.Register(crd1)
	assert.NoError(t, err1)

	err2 := resolver.Register(crd2)
	assert.NoError(t, err2)

	assert.NotNil(t, resolver.Resolve("Database"))
	assert.NotNil(t, resolver.Resolve("Cache"))
}

// =============================================================================
// CRDResolver Tests - Unregister
// =============================================================================

func TestResolver_Unregister_Success(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd := createTestCRD("Database", "database", "databases", []string{"db"})

	err := resolver.Register(crd)
	require.NoError(t, err)

	err = resolver.Unregister("Database")
	assert.NoError(t, err)

	resolved := resolver.Resolve("Database")
	assert.Nil(t, resolved, "Should not resolve after unregister")
}

func TestResolver_Unregister_NotFoundReturnsError(t *testing.T) {
	resolver := NewMockCRDResolver()

	err := resolver.Unregister("NonExistent")
	assert.Error(t, err)
	assert.IsType(t, &CRDNotFoundError{}, err)
}

// =============================================================================
// CRDResolver Tests - List
// =============================================================================

func TestResolver_List_ReturnsAllCRDs(t *testing.T) {
	resolver := NewMockCRDResolver()
	crd1 := createTestCRD("Database", "database", "databases", []string{"db"})
	crd2 := createTestCRD("Cache", "cache", "caches", []string{"ch"})

	resolver.Register(crd1)
	resolver.Register(crd2)

	crds := resolver.List()
	assert.Len(t, crds, 2)
}

func TestResolver_List_EmptyWhenNoCRDs(t *testing.T) {
	resolver := NewMockCRDResolver()

	crds := resolver.List()
	assert.Empty(t, crds)
}

// =============================================================================
// CRDResolver Tests - Refresh
// =============================================================================

func TestResolver_Refresh_ReloadsFromStore(t *testing.T) {
	resolver := NewMockCRDResolver()

	err := resolver.Refresh()
	assert.NoError(t, err)
}
