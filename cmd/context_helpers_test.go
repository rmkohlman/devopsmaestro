package cmd

import (
	"database/sql"
	"fmt"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// intPtr is a helper that returns a pointer to an int value.
func intPtr(i int) *int {
	return &i
}

// =============================================================================
// TestGetActiveAppFromContext
// =============================================================================

func TestGetActiveAppFromContext(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, mock *db.MockDataStore)
		envKey    string
		envVal    string
		wantName  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:   "env var DVM_APP overrides DB",
			envKey: "DVM_APP",
			envVal: "env-app",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				// DB should NOT be consulted when env var is set; leave context empty
				mock.Context = nil
			},
			wantName: "env-app",
			wantErr:  false,
		},
		{
			name: "DB context has active app ID and GetAppByID succeeds",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				appID := 42
				mock.Context = &models.Context{ID: 1, ActiveAppID: &appID}
				mock.Apps[42] = &models.App{ID: 42, Name: "my-app", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
			},
			wantName: "my-app",
			wantErr:  false,
		},
		{
			name: "DB context is nil returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.Context = nil
			},
			wantErr:   true,
			errSubstr: "no active app context",
		},
		{
			name: "DB context has nil ActiveAppID returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.Context = &models.Context{ID: 1, ActiveAppID: nil}
			},
			wantErr:   true,
			errSubstr: "no active app context",
		},
		{
			name: "GetContext returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.GetContextErr = fmt.Errorf("db error")
			},
			wantErr:   true,
			errSubstr: "no active app context",
		},
		{
			name: "GetAppByID returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.Context = &models.Context{ID: 1, ActiveAppID: intPtr(99)}
				// ID 99 is not in mock.Apps, so the mock returns "app not found: 99"
				// We also inject a dedicated error to be explicit.
				mock.GetAppByIDErr = fmt.Errorf("app not found")
			},
			wantErr:   true,
			errSubstr: "no active app context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up env var if specified (t.Setenv cleans up automatically)
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envVal)
			}

			mock := db.NewMockDataStore()
			if tt.setup != nil {
				tt.setup(t, mock)
			}

			got, err := getActiveAppFromContext(mock)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantName, got)
			}
		})
	}
}

// =============================================================================
// TestGetActiveWorkspaceFromContext
// =============================================================================

func TestGetActiveWorkspaceFromContext(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, mock *db.MockDataStore)
		envKey    string
		envVal    string
		wantName  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:   "env var DVM_WORKSPACE overrides DB",
			envKey: "DVM_WORKSPACE",
			envVal: "env-workspace",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				// DB should NOT be consulted when env var is set; leave context empty
				mock.Context = nil
			},
			wantName: "env-workspace",
			wantErr:  false,
		},
		{
			name: "DB context has active workspace ID and GetWorkspaceByID succeeds",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				wsID := 7
				mock.Context = &models.Context{ID: 1, ActiveWorkspaceID: &wsID}
				mock.Workspaces[7] = &models.Workspace{ID: 7, Name: "dev", AppID: 1}
			},
			wantName: "dev",
			wantErr:  false,
		},
		{
			name: "DB context is nil returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.Context = nil
			},
			wantErr:   true,
			errSubstr: "no active workspace context",
		},
		{
			name: "DB context has nil ActiveWorkspaceID returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.Context = &models.Context{ID: 1, ActiveWorkspaceID: nil}
			},
			wantErr:   true,
			errSubstr: "no active workspace context",
		},
		{
			name: "GetContext returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.GetContextErr = fmt.Errorf("db error")
			},
			wantErr:   true,
			errSubstr: "no active workspace context",
		},
		{
			name: "GetWorkspaceByID returns error",
			setup: func(t *testing.T, mock *db.MockDataStore) {
				mock.Context = &models.Context{ID: 1, ActiveWorkspaceID: intPtr(999)}
				// ID 999 is not in mock.Workspaces, so the mock returns "workspace not found: 999"
				// We also inject a dedicated error to be explicit.
				mock.GetWorkspaceByIDErr = fmt.Errorf("workspace not found")
			},
			wantErr:   true,
			errSubstr: "no active workspace context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up env var if specified (t.Setenv cleans up automatically)
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envVal)
			}

			mock := db.NewMockDataStore()
			if tt.setup != nil {
				tt.setup(t, mock)
			}

			got, err := getActiveWorkspaceFromContext(mock)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantName, got)
			}
		})
	}
}

// ===========================================================================
// Sprint 3 Tests: getActiveEcosystemFromContext / getActiveDomainFromContext
// ===========================================================================

// ---------------------------------------------------------------------------
// getActiveEcosystemFromContext
// ---------------------------------------------------------------------------

func TestGetActiveEcosystemFromContext_EnvVarOverride(t *testing.T) {
	// DVM_ECOSYSTEM=foo should return "foo" regardless of DB context
	t.Setenv("DVM_ECOSYSTEM", "foo")
	mock := db.NewMockDataStore()
	mock.Context = nil // DB should NOT be consulted when env var is set

	name, err := getActiveEcosystemFromContext(mock)
	require.NoError(t, err)
	assert.Equal(t, "foo", name)
}

func TestGetActiveEcosystemFromContext_DBContext(t *testing.T) {
	// No env var, DB has active ecosystem -> returns ecosystem name
	mock := db.NewMockDataStore()
	ecoID := 42
	mock.Ecosystems["production"] = &models.Ecosystem{ID: 42, Name: "production"}
	mock.Context = &models.Context{ID: 1, ActiveEcosystemID: &ecoID}

	name, err := getActiveEcosystemFromContext(mock)
	require.NoError(t, err)
	assert.Equal(t, "production", name)
}

func TestGetActiveEcosystemFromContext_NoContextReturnsError(t *testing.T) {
	// No env var, no DB context -> error with hint
	mock := db.NewMockDataStore()
	mock.Context = &models.Context{ID: 1} // No active ecosystem set

	_, err := getActiveEcosystemFromContext(mock)
	assert.Error(t, err)
}

func TestGetActiveEcosystemFromContext_NilContextReturnsError(t *testing.T) {
	// No env var, nil context -> error with hint
	mock := db.NewMockDataStore()
	mock.Context = nil

	_, err := getActiveEcosystemFromContext(mock)
	assert.Error(t, err)
}

func TestGetActiveEcosystemFromContext_GetContextError(t *testing.T) {
	mock := db.NewMockDataStore()
	mock.GetContextErr = fmt.Errorf("db error")

	_, err := getActiveEcosystemFromContext(mock)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// getActiveDomainFromContext
// ---------------------------------------------------------------------------

func TestGetActiveDomainFromContext_EnvVarOverride(t *testing.T) {
	// DVM_DOMAIN=bar should return "bar" regardless of DB context
	t.Setenv("DVM_DOMAIN", "bar")
	mock := db.NewMockDataStore()
	mock.Context = nil // DB should NOT be consulted when env var is set

	name, err := getActiveDomainFromContext(mock)
	require.NoError(t, err)
	assert.Equal(t, "bar", name)
}

func TestGetActiveDomainFromContext_DBContext(t *testing.T) {
	// No env var, DB has active domain -> returns domain name
	mock := db.NewMockDataStore()
	domID := 10
	mock.Domains[10] = &models.Domain{ID: 10, Name: "backend", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Context = &models.Context{ID: 1, ActiveDomainID: &domID}

	name, err := getActiveDomainFromContext(mock)
	require.NoError(t, err)
	assert.Equal(t, "backend", name)
}

func TestGetActiveDomainFromContext_NoContextReturnsError(t *testing.T) {
	// No env var, no DB context -> error with hint
	mock := db.NewMockDataStore()
	mock.Context = &models.Context{ID: 1} // No active domain set

	_, err := getActiveDomainFromContext(mock)
	assert.Error(t, err)
}

func TestGetActiveDomainFromContext_NilContextReturnsError(t *testing.T) {
	// No env var, nil context -> error with hint
	mock := db.NewMockDataStore()
	mock.Context = nil

	_, err := getActiveDomainFromContext(mock)
	assert.Error(t, err)
}

func TestGetActiveDomainFromContext_GetContextError(t *testing.T) {
	mock := db.NewMockDataStore()
	mock.GetContextErr = fmt.Errorf("db error")

	_, err := getActiveDomainFromContext(mock)
	assert.Error(t, err)
}
