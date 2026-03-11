package workspace

import (
	"database/sql"
	"errors"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Wave 1B: PrepareDefaults Tests (Sprint 4.3)
// Tests for workspace default preparation: NvimStructure and slug generation.
// =============================================================================

// mockHierarchyReader is a test double for HierarchyReader.
// It holds pre-configured return values and errors for each lookup method,
// enabling fine-grained control over hierarchy traversal in tests.
type mockHierarchyReader struct {
	apps       map[int]*models.App
	domains    map[int]*models.Domain
	ecosystems map[int]*models.Ecosystem
	appErr     error
	domainErr  error
	ecoErr     error
}

func (m *mockHierarchyReader) GetAppByID(id int) (*models.App, error) {
	if m.appErr != nil {
		return nil, m.appErr
	}
	if app, ok := m.apps[id]; ok {
		return app, nil
	}
	return nil, errors.New("app not found")
}

func (m *mockHierarchyReader) GetDomainByID(id int) (*models.Domain, error) {
	if m.domainErr != nil {
		return nil, m.domainErr
	}
	if domain, ok := m.domains[id]; ok {
		return domain, nil
	}
	return nil, errors.New("domain not found")
}

func (m *mockHierarchyReader) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	if m.ecoErr != nil {
		return nil, m.ecoErr
	}
	if eco, ok := m.ecosystems[id]; ok {
		return eco, nil
	}
	return nil, errors.New("ecosystem not found")
}

// newFullHierarchy builds a standard mock reader with a complete, traversable
// hierarchy: ecosystem(id=1) → domain(id=2) → app(id=3).
func newFullHierarchy() *mockHierarchyReader {
	return &mockHierarchyReader{
		apps: map[int]*models.App{
			3: {ID: 3, DomainID: 2, Name: "myapp"},
		},
		domains: map[int]*models.Domain{
			2: {ID: 2, EcosystemID: 1, Name: "mydomain"},
		},
		ecosystems: map[int]*models.Ecosystem{
			1: {ID: 1, Name: "myeco"},
		},
	}
}

func TestPrepareDefaults(t *testing.T) {
	tests := []struct {
		name          string
		workspace     *models.Workspace
		hierarchy     HierarchyReader
		wantErr       bool
		errContains   string
		wantNvim      string
		wantSlug      string
		slugPreserved bool
		nvimPreserved bool
	}{
		{
			name: "defaults applied: empty workspace gets nvim structure and generated slug",
			workspace: &models.Workspace{
				AppID: 3,
				Name:  "dev",
				// NvimStructure: zero value (not valid, empty string)
				// Slug: ""
			},
			hierarchy: newFullHierarchy(),
			wantErr:   false,
			wantNvim:  "lazyvim",
			wantSlug:  "myeco-mydomain-myapp-dev",
		},
		{
			name: "nvim structure preserved when already set",
			workspace: &models.Workspace{
				AppID:         3,
				Name:          "dev",
				NvimStructure: sql.NullString{String: "nvchad", Valid: true},
			},
			hierarchy:     newFullHierarchy(),
			wantErr:       false,
			wantNvim:      "nvchad", // Not overwritten
			wantSlug:      "myeco-mydomain-myapp-dev",
			nvimPreserved: true,
		},
		{
			name: "slug preserved when already set",
			workspace: &models.Workspace{
				AppID: 3,
				Name:  "dev",
				Slug:  "custom-slug-already-set",
				// NvimStructure: zero value → should get default
			},
			hierarchy:     newFullHierarchy(),
			wantErr:       false,
			wantNvim:      "lazyvim",
			wantSlug:      "custom-slug-already-set", // Not overwritten
			slugPreserved: true,
		},
		{
			name: "both nvim structure and slug preserved when both already set",
			workspace: &models.Workspace{
				AppID:         3,
				Name:          "dev",
				NvimStructure: sql.NullString{String: "astronvim", Valid: true},
				Slug:          "pre-existing-slug",
			},
			hierarchy:     newFullHierarchy(),
			wantErr:       false,
			wantNvim:      "astronvim",         // Not overwritten
			wantSlug:      "pre-existing-slug", // Not overwritten
			nvimPreserved: true,
			slugPreserved: true,
		},
		{
			name: "app lookup error is propagated",
			workspace: &models.Workspace{
				AppID: 3,
				Name:  "dev",
			},
			hierarchy: &mockHierarchyReader{
				appErr: errors.New("db connection refused"),
			},
			wantErr:     true,
			errContains: "failed to get app for slug generation",
		},
		{
			name: "domain lookup error is propagated",
			workspace: &models.Workspace{
				AppID: 3,
				Name:  "dev",
			},
			hierarchy: &mockHierarchyReader{
				apps: map[int]*models.App{
					3: {ID: 3, DomainID: 2, Name: "myapp"},
				},
				domainErr: errors.New("domain table missing"),
			},
			wantErr:     true,
			errContains: "failed to get domain for slug generation",
		},
		{
			name: "ecosystem lookup error is propagated",
			workspace: &models.Workspace{
				AppID: 3,
				Name:  "dev",
			},
			hierarchy: &mockHierarchyReader{
				apps: map[int]*models.App{
					3: {ID: 3, DomainID: 2, Name: "myapp"},
				},
				domains: map[int]*models.Domain{
					2: {ID: 2, EcosystemID: 1, Name: "mydomain"},
				},
				ecoErr: errors.New("ecosystem not found in store"),
			},
			wantErr:     true,
			errContains: "failed to get ecosystem for slug generation",
		},
		{
			name: "nvim structure valid but empty string still gets default applied",
			workspace: &models.Workspace{
				AppID:         3,
				Name:          "dev",
				NvimStructure: sql.NullString{String: "", Valid: true},
			},
			hierarchy: newFullHierarchy(),
			wantErr:   false,
			wantNvim:  "lazyvim", // Empty-but-valid is treated as unset
			wantSlug:  "myeco-mydomain-myapp-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PrepareDefaults(tt.workspace, tt.hierarchy)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)

			// Verify NvimStructure
			assert.True(t, tt.workspace.NvimStructure.Valid,
				"NvimStructure.Valid should be true after PrepareDefaults")
			assert.Equal(t, tt.wantNvim, tt.workspace.NvimStructure.String,
				"NvimStructure.String should be %q", tt.wantNvim)

			// Verify Slug
			assert.Equal(t, tt.wantSlug, tt.workspace.Slug,
				"Slug should be %q", tt.wantSlug)
		})
	}
}

// TestPrepareDefaults_NvimDefaultIsLazyvim documents the specific expected
// default value for NvimStructure so it's obvious if it ever changes.
func TestPrepareDefaults_NvimDefaultIsLazyvim(t *testing.T) {
	ws := &models.Workspace{
		AppID: 3, // matches the app in newFullHierarchy()
		Name:  "dev",
	}

	err := PrepareDefaults(ws, newFullHierarchy())
	require.NoError(t, err)

	assert.Equal(t, "lazyvim", ws.NvimStructure.String,
		"default NvimStructure should be 'lazyvim' per nvimops.DefaultNvimConfig()")
}

// TestPrepareDefaults_SlugUsesHierarchyNames documents that slug generation
// traverses the full ecosystem → domain → app → workspace chain.
func TestPrepareDefaults_SlugUsesHierarchyNames(t *testing.T) {
	hierarchy := &mockHierarchyReader{
		apps: map[int]*models.App{
			10: {ID: 10, DomainID: 20, Name: "payments-api"},
		},
		domains: map[int]*models.Domain{
			20: {ID: 20, EcosystemID: 30, Name: "finance"},
		},
		ecosystems: map[int]*models.Ecosystem{
			30: {ID: 30, Name: "corp"},
		},
	}

	ws := &models.Workspace{
		AppID: 10,
		Name:  "dev",
	}

	err := PrepareDefaults(ws, hierarchy)
	require.NoError(t, err)

	assert.Equal(t, "corp-finance-payments-api-dev", ws.Slug,
		"slug should concatenate ecosystem-domain-app-workspace with hyphens")
}
