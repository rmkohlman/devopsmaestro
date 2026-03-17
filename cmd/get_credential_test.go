package cmd

// =============================================================================
// Tests for resolveScopeName() and formatTargetVars() helper functions
// in cmd/get_credential.go (added as part of credential output improvements).
// =============================================================================

import (
	"encoding/json"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr is a test helper that returns a pointer to the given string value.
func ptr(s string) *string { return &s }

// =============================================================================
// Tests for resolveScopeName()
// =============================================================================

// TestResolveScopeName_EcosystemScope verifies that an ecosystem-scoped credential
// resolves to "ecosystem: <name>" when the ecosystem exists in the store.
func TestResolveScopeName_EcosystemScope(t *testing.T) {
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "my-ecosystem"}
	require.NoError(t, store.CreateEcosystem(eco))

	result := resolveScopeName(store, models.CredentialScopeEcosystem, int64(eco.ID))

	assert.Equal(t, "ecosystem: my-ecosystem", result)
}

// TestResolveScopeName_DomainScope verifies that a domain-scoped credential
// resolves to "domain: <name>" when the domain exists in the store.
func TestResolveScopeName_DomainScope(t *testing.T) {
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "parent-eco"}
	require.NoError(t, store.CreateEcosystem(eco))

	domain := &models.Domain{Name: "my-domain", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(domain))

	result := resolveScopeName(store, models.CredentialScopeDomain, int64(domain.ID))

	assert.Equal(t, "domain: my-domain", result)
}

// TestResolveScopeName_AppScope verifies that an app-scoped credential
// resolves to "app: <name>" when the app exists in the store.
func TestResolveScopeName_AppScope(t *testing.T) {
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "eco"}
	require.NoError(t, store.CreateEcosystem(eco))
	dom := &models.Domain{Name: "dom", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(dom))
	app := &models.App{Name: "my-api", DomainID: dom.ID, Path: "/srv/api"}
	require.NoError(t, store.CreateApp(app))

	result := resolveScopeName(store, models.CredentialScopeApp, int64(app.ID))

	assert.Equal(t, "app: my-api", result)
}

// TestResolveScopeName_WorkspaceScope verifies that a workspace-scoped credential
// resolves to "workspace: <name>" when the workspace exists in the store.
func TestResolveScopeName_WorkspaceScope(t *testing.T) {
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "eco"}
	require.NoError(t, store.CreateEcosystem(eco))
	dom := &models.Domain{Name: "dom", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(dom))
	app := &models.App{Name: "app", DomainID: dom.ID, Path: "/srv/app"}
	require.NoError(t, store.CreateApp(app))
	ws := &models.Workspace{
		Name:      "my-workspace",
		AppID:     app.ID,
		Slug:      "eco-dom-app-my-workspace",
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	require.NoError(t, store.CreateWorkspace(ws))

	result := resolveScopeName(store, models.CredentialScopeWorkspace, int64(ws.ID))

	assert.Equal(t, "workspace: my-workspace", result)
}

// TestResolveScopeName_FallbackWhenEcosystemNotFound verifies that when an
// ecosystem lookup fails (resource deleted or ID doesn't exist), the function
// falls back to "ecosystem (ID: X)" format.
func TestResolveScopeName_FallbackWhenEcosystemNotFound(t *testing.T) {
	store := db.NewMockDataStore()
	// No ecosystems created — ID 99 does not exist

	result := resolveScopeName(store, models.CredentialScopeEcosystem, 99)

	assert.Equal(t, "ecosystem (ID: 99)", result)
}

// TestResolveScopeName_FallbackWhenDomainNotFound verifies fallback for domain.
func TestResolveScopeName_FallbackWhenDomainNotFound(t *testing.T) {
	store := db.NewMockDataStore()

	result := resolveScopeName(store, models.CredentialScopeDomain, 42)

	assert.Equal(t, "domain (ID: 42)", result)
}

// TestResolveScopeName_FallbackWhenAppNotFound verifies fallback for app.
func TestResolveScopeName_FallbackWhenAppNotFound(t *testing.T) {
	store := db.NewMockDataStore()

	result := resolveScopeName(store, models.CredentialScopeApp, 7)

	assert.Equal(t, "app (ID: 7)", result)
}

// TestResolveScopeName_FallbackWhenWorkspaceNotFound verifies fallback for workspace.
func TestResolveScopeName_FallbackWhenWorkspaceNotFound(t *testing.T) {
	store := db.NewMockDataStore()

	result := resolveScopeName(store, models.CredentialScopeWorkspace, 3)

	assert.Equal(t, "workspace (ID: 3)", result)
}

// TestResolveScopeName_AllScopeTypes is a table-driven test that covers all four
// scope types with a name-resolved result, verifying the "type: name" format.
func TestResolveScopeName_AllScopeTypes(t *testing.T) {
	store, app := setupTestContext()

	// Retrieve the domain so we can use its ID for scope testing
	eco, err := store.GetEcosystemByName("test-eco")
	require.NoError(t, err)

	domain, err := store.GetDomainByName(eco.ID, "test-domain")
	require.NoError(t, err)

	// Create a workspace for workspace scope
	ws := &models.Workspace{
		Name:      "scope-ws",
		AppID:     app.ID,
		Slug:      "test-eco-test-domain-test-app-scope-ws",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, store.CreateWorkspace(ws))

	tests := []struct {
		name      string
		scopeType models.CredentialScopeType
		scopeID   int64
		want      string
	}{
		{
			name:      "ecosystem scope resolves to name",
			scopeType: models.CredentialScopeEcosystem,
			scopeID:   int64(eco.ID),
			want:      "ecosystem: test-eco",
		},
		{
			name:      "domain scope resolves to name",
			scopeType: models.CredentialScopeDomain,
			scopeID:   int64(domain.ID),
			want:      "domain: test-domain",
		},
		{
			name:      "app scope resolves to name",
			scopeType: models.CredentialScopeApp,
			scopeID:   int64(app.ID),
			want:      "app: test-app",
		},
		{
			name:      "workspace scope resolves to name",
			scopeType: models.CredentialScopeWorkspace,
			scopeID:   int64(ws.ID),
			want:      "workspace: scope-ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveScopeName(store, tt.scopeType, tt.scopeID)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestResolveScopeName_ErrorInjection verifies that when the store returns an error
// for GetEcosystemByID (via error injection), the fallback format is used.
func TestResolveScopeName_ErrorInjection(t *testing.T) {
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "broken-eco"}
	require.NoError(t, store.CreateEcosystem(eco))

	// Inject an error so GetEcosystemByID fails even though the ecosystem exists
	store.GetEcosystemByIDErr = assert.AnError

	result := resolveScopeName(store, models.CredentialScopeEcosystem, int64(eco.ID))

	// Should fall back to "ecosystem (ID: X)" because lookup errored
	assert.Contains(t, result, "ecosystem (ID:")
}

// =============================================================================
// Tests for formatTargetVars()
// =============================================================================

// TestFormatTargetVars_VaultFields verifies that a credential with vault fields
// returns a sorted, comma-separated list of the env var names (the map keys).
func TestFormatTargetVars_VaultFields(t *testing.T) {
	// Vault fields JSON: {"DB_USER": "username", "DB_PASS": "password"}
	// sorted keys: DB_PASS, DB_USER
	vaultFields := `{"DB_USER":"username","DB_PASS":"password"}`
	cred := &models.CredentialDB{
		Name:        "db-creds",
		Source:      "vault",
		VaultFields: &vaultFields,
	}

	result := formatTargetVars(cred)

	// Keys are sorted alphabetically: DB_PASS, DB_USER
	assert.Equal(t, "DB_PASS, DB_USER", result)
}

// TestFormatTargetVars_VaultFields_SingleEntry verifies single vault field returns
// the env var name without a trailing comma.
func TestFormatTargetVars_VaultFields_SingleEntry(t *testing.T) {
	vaultFields := `{"GITHUB_TOKEN":"token"}`
	cred := &models.CredentialDB{
		Name:        "github-token",
		Source:      "vault",
		VaultFields: &vaultFields,
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "GITHUB_TOKEN", result)
}

// TestFormatTargetVars_VaultFields_SortedOrder verifies that when multiple vault
// fields are present, their env var keys are always sorted alphabetically.
func TestFormatTargetVars_VaultFields_SortedOrder(t *testing.T) {
	// Keys: ZEBRA, ALPHA, MIDDLE — sorted should be ALPHA, MIDDLE, ZEBRA
	fields := map[string]string{
		"ZEBRA":  "z_field",
		"ALPHA":  "a_field",
		"MIDDLE": "m_field",
	}
	vfJSON, err := json.Marshal(fields)
	require.NoError(t, err)
	vaultFieldsStr := string(vfJSON)

	cred := &models.CredentialDB{
		Name:        "multi-field",
		Source:      "vault",
		VaultFields: &vaultFieldsStr,
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "ALPHA, MIDDLE, ZEBRA", result)
}

// TestFormatTargetVars_DualField_BothVars verifies that a dual-field credential
// (UsernameVar + PasswordVar, no VaultFields) returns both var names joined with ", ".
func TestFormatTargetVars_DualField_BothVars(t *testing.T) {
	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		UsernameVar: ptr("GH_USER"),
		PasswordVar: ptr("GH_PAT"),
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "GH_USER, GH_PAT", result)
}

// TestFormatTargetVars_DualField_UsernameOnly verifies that a credential with
// only UsernameVar set (no PasswordVar) returns just the username var.
func TestFormatTargetVars_DualField_UsernameOnly(t *testing.T) {
	cred := &models.CredentialDB{
		Name:        "user-only",
		Source:      "vault",
		UsernameVar: ptr("MY_USER"),
		PasswordVar: nil,
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "MY_USER", result)
}

// TestFormatTargetVars_DualField_PasswordOnly verifies that a credential with
// only PasswordVar set (no UsernameVar) returns just the password var.
func TestFormatTargetVars_DualField_PasswordOnly(t *testing.T) {
	cred := &models.CredentialDB{
		Name:        "pass-only",
		Source:      "vault",
		UsernameVar: nil,
		PasswordVar: ptr("MY_PASS"),
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "MY_PASS", result)
}

// TestFormatTargetVars_SingleEnvVar verifies that a credential with only EnvVar set
// (no VaultFields, no UsernameVar, no PasswordVar) returns the env var name.
func TestFormatTargetVars_SingleEnvVar(t *testing.T) {
	cred := &models.CredentialDB{
		Name:   "api-key",
		Source: "env",
		EnvVar: ptr("MY_API_KEY"),
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "MY_API_KEY", result)
}

// TestFormatTargetVars_Fallback_CredentialName verifies that when no vault fields,
// no dual-field vars, and no EnvVar are set, the credential Name is used as fallback.
func TestFormatTargetVars_Fallback_CredentialName(t *testing.T) {
	cred := &models.CredentialDB{
		Name:   "LEGACY_TOKEN",
		Source: "vault",
		// No VaultFields, no UsernameVar, no PasswordVar, no EnvVar
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "LEGACY_TOKEN", result)
}

// TestFormatTargetVars_VaultFieldsPriority verifies that VaultFields takes priority
// over UsernameVar/PasswordVar when both are somehow set.
func TestFormatTargetVars_VaultFieldsPriority(t *testing.T) {
	vaultFields := `{"FIELD_A":"fa","FIELD_B":"fb"}`
	cred := &models.CredentialDB{
		Name:        "priority-test",
		Source:      "vault",
		VaultFields: &vaultFields,
		UsernameVar: ptr("SHOULD_NOT_APPEAR"),
		PasswordVar: ptr("SHOULD_NOT_APPEAR_EITHER"),
	}

	result := formatTargetVars(cred)

	// VaultFields takes priority — the dual-field vars should NOT appear
	assert.Equal(t, "FIELD_A, FIELD_B", result)
	assert.NotContains(t, result, "SHOULD_NOT_APPEAR")
}

// TestFormatTargetVars_TableDriven is a comprehensive table-driven test covering
// all branches of formatTargetVars priority logic.
func TestFormatTargetVars_TableDriven(t *testing.T) {
	vaultFieldsSingle := `{"TOKEN_VAR":"token"}`
	vaultFieldsMulti := `{"ZEBRA_VAR":"z","ALPHA_VAR":"a"}`

	tests := []struct {
		name string
		cred *models.CredentialDB
		want string
	}{
		{
			name: "vault fields - single key",
			cred: &models.CredentialDB{
				Name: "single-field", Source: "vault",
				VaultFields: &vaultFieldsSingle,
			},
			want: "TOKEN_VAR",
		},
		{
			name: "vault fields - multiple keys sorted",
			cred: &models.CredentialDB{
				Name: "multi-field", Source: "vault",
				VaultFields: &vaultFieldsMulti,
			},
			want: "ALPHA_VAR, ZEBRA_VAR",
		},
		{
			name: "dual field - both username and password",
			cred: &models.CredentialDB{
				Name: "dual", Source: "vault",
				UsernameVar: ptr("USER_VAR"),
				PasswordVar: ptr("PASS_VAR"),
			},
			want: "USER_VAR, PASS_VAR",
		},
		{
			name: "dual field - username only",
			cred: &models.CredentialDB{
				Name: "user-only", Source: "vault",
				UsernameVar: ptr("USER_ONLY"),
			},
			want: "USER_ONLY",
		},
		{
			name: "dual field - password only",
			cred: &models.CredentialDB{
				Name: "pass-only", Source: "vault",
				PasswordVar: ptr("PASS_ONLY"),
			},
			want: "PASS_ONLY",
		},
		{
			name: "single env var",
			cred: &models.CredentialDB{
				Name: "env-cred", Source: "env",
				EnvVar: ptr("MY_ENV_VAR"),
			},
			want: "MY_ENV_VAR",
		},
		{
			name: "fallback to credential name",
			cred: &models.CredentialDB{
				Name: "FALLBACK_CRED", Source: "vault",
			},
			want: "FALLBACK_CRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTargetVars(tt.cred)
			assert.Equal(t, tt.want, result,
				"formatTargetVars(%s) mismatch", tt.name)
		})
	}
}

// TestFormatTargetVars_EmptyVaultFields verifies that an empty VaultFields JSON
// object "{}" is treated as no vault fields (falls through to next priority).
func TestFormatTargetVars_EmptyVaultFields(t *testing.T) {
	emptyFields := "{}"
	cred := &models.CredentialDB{
		Name:        "empty-fields",
		Source:      "vault",
		VaultFields: &emptyFields,
		EnvVar:      ptr("FALLBACK_ENV"),
	}

	result := formatTargetVars(cred)

	// "{}" is treated as no vault fields — falls through to EnvVar
	assert.Equal(t, "FALLBACK_ENV", result)
}

// TestFormatTargetVars_NilVaultFields verifies that a nil VaultFields pointer
// is treated the same as no vault fields.
func TestFormatTargetVars_NilVaultFields(t *testing.T) {
	cred := &models.CredentialDB{
		Name:        "nil-fields",
		Source:      "vault",
		VaultFields: nil,
		EnvVar:      ptr("MY_TOKEN"),
	}

	result := formatTargetVars(cred)

	assert.Equal(t, "MY_TOKEN", result)
}
