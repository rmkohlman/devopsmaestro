package handlers

import (
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
)

// =============================================================================
// CredentialHandler Tests - Kind
// =============================================================================

func TestCredentialHandler_Kind(t *testing.T) {
	h := NewCredentialHandler()
	if h.Kind() != KindCredential {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindCredential)
	}
}

// =============================================================================
// CredentialHandler Tests - Apply
// =============================================================================

func TestCredentialHandler_Apply(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	// Pre-populate scope entities for tests that need them
	eco := &models.Ecosystem{Name: "testlab"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Setup: CreateEcosystem() error = %v", err)
	}

	domain := &models.Domain{EcosystemID: eco.ID, Name: "backend"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup: CreateDomain() error = %v", err)
	}

	app := &models.App{DomainID: domain.ID, Name: "api-server"}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup: CreateApp() error = %v", err)
	}

	workspace := &models.Workspace{AppID: app.ID, Name: "dev"}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup: CreateWorkspace() error = %v", err)
	}

	ctx := resource.Context{DataStore: ds}

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "create vault credential in ecosystem scope",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  ecosystem: testlab
spec:
  source: vault
  vaultSecret: github/token
  description: GitHub personal access token`,
			wantErr: false,
		},
		{
			name: "create env credential in app scope",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: db-url
  app: api-server
spec:
  source: env
  envVar: DATABASE_URL`,
			wantErr: false,
		},
		{
			name: "create credential in domain scope",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: domain-token
  domain: backend
spec:
  source: vault
  vaultSecret: internal/token`,
			wantErr: false,
		},
		{
			name: "create credential in workspace scope",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: ws-secret
  workspace: dev
spec:
  source: env
  envVar: WS_SECRET`,
			wantErr: false,
		},
		{
			name: "invalid source rejected",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: bad-cred
  ecosystem: testlab
spec:
  source: plaintext`,
			wantErr: true,
			errMsg:  "source must be",
		},
		{
			name: "missing name rejected",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  ecosystem: testlab
spec:
  source: vault
  vaultSecret: some/secret`,
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing scope rejected",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: no-scope-cred
spec:
  source: vault
  vaultSecret: some/secret`,
			wantErr: true,
			errMsg:  "exactly one scope",
		},
		{
			name: "unknown ecosystem rejected",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: ghost-cred
  ecosystem: nonexistent
spec:
  source: vault
  vaultSecret: ghost/secret`,
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:    "invalid YAML rejected",
			yaml:    "{\x00this is not: [valid yaml",
			wantErr: true,
			errMsg:  "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := h.Apply(ctx, []byte(tt.yaml))
			if tt.wantErr {
				if err == nil {
					t.Errorf("Apply() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Apply() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("Apply() error = %v", err)
				}
				if res.GetKind() != KindCredential {
					t.Errorf("Apply() resource.Kind = %v, want %v", res.GetKind(), KindCredential)
				}
			}
		})
	}
}

func TestCredentialHandler_Apply_Update(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	// Pre-populate ecosystem
	eco := &models.Ecosystem{Name: "update-lab"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Setup: CreateEcosystem() error = %v", err)
	}

	ctx := resource.Context{DataStore: ds}

	// Create the initial credential
	createYAML := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: update-cred
  ecosystem: update-lab
spec:
  source: vault
  vaultSecret: original/secret
  description: original description`

	_, err := h.Apply(ctx, []byte(createYAML))
	if err != nil {
		t.Fatalf("Apply() create error = %v", err)
	}

	// Update with different source fields
	updateYAML := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: update-cred
  ecosystem: update-lab
spec:
  source: env
  envVar: UPDATED_ENV_VAR
  description: updated description`

	res, err := h.Apply(ctx, []byte(updateYAML))
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}

	// Verify the update took effect by inspecting the returned resource
	credRes, ok := res.(*CredentialResource)
	if !ok {
		t.Fatalf("Apply() result is not *CredentialResource")
	}

	cred := credRes.Credential()
	if cred.Source != "env" {
		t.Errorf("Apply() updated source = %q, want %q", cred.Source, "env")
	}
	if cred.EnvVar == nil || *cred.EnvVar != "UPDATED_ENV_VAR" {
		t.Errorf("Apply() updated env_var = %v, want %q", cred.EnvVar, "UPDATED_ENV_VAR")
	}
}

// =============================================================================
// CredentialHandler Tests - ToYAML
// =============================================================================

func TestCredentialHandler_ToYAML(t *testing.T) {
	h := NewCredentialHandler()

	vaultSecret := "github/token"
	desc := "github token"
	cred := &models.CredentialDB{
		Name:        "github-token",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     1,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		Description: &desc,
	}

	res := &CredentialResource{
		credential: cred,
		scopeName:  "mylab",
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	if !contains(yamlStr, "kind: Credential") {
		t.Errorf("ToYAML() missing 'kind: Credential', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "name: github-token") {
		t.Errorf("ToYAML() missing 'name: github-token', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "source: vault") {
		t.Errorf("ToYAML() missing 'source: vault', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "vaultSecret: github/token") {
		t.Errorf("ToYAML() missing 'vaultSecret: github/token', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "ecosystem: mylab") {
		t.Errorf("ToYAML() missing 'ecosystem: mylab', got:\n%s", yamlStr)
	}
}

// =============================================================================
// CredentialResource Validation Tests
// =============================================================================

func TestCredentialResource_Validate(t *testing.T) {
	vaultSecret := "some/secret"
	envVar := "MY_ENV_VAR"

	tests := []struct {
		name      string
		cred      *models.CredentialDB
		scopeName string
		wantErr   bool
	}{
		{
			name: "valid vault credential",
			cred: &models.CredentialDB{
				Name:        "valid-cred",
				ScopeType:   models.CredentialScopeEcosystem,
				ScopeID:     1,
				Source:      "vault",
				VaultSecret: &vaultSecret,
			},
			scopeName: "mylab",
			wantErr:   false,
		},
		{
			name: "missing name",
			cred: &models.CredentialDB{
				Name:        "",
				ScopeType:   models.CredentialScopeEcosystem,
				ScopeID:     1,
				Source:      "vault",
				VaultSecret: &vaultSecret,
			},
			scopeName: "mylab",
			wantErr:   true,
		},
		{
			name: "invalid source",
			cred: &models.CredentialDB{
				Name:      "bad-source",
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   1,
				Source:    "plaintext",
			},
			scopeName: "mylab",
			wantErr:   true,
		},
		{
			name: "valid env credential",
			cred: &models.CredentialDB{
				Name:      "env-cred",
				ScopeType: models.CredentialScopeApp,
				ScopeID:   1,
				Source:    "env",
				EnvVar:    &envVar,
			},
			scopeName: "myapp",
			wantErr:   false,
		},
		{
			name: "missing scope name produces empty scope in YAML",
			cred: &models.CredentialDB{
				Name:        "no-scope",
				ScopeType:   models.CredentialScopeEcosystem,
				ScopeID:     1,
				Source:      "vault",
				VaultSecret: &vaultSecret,
			},
			scopeName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &CredentialResource{
				credential: tt.cred,
				scopeName:  tt.scopeName,
			}
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// createCredentialTestDataStore creates an in-memory SQLite DataStore with the
// full credential schema (all scope tables + credentials).
func createCredentialTestDataStore(t *testing.T) *db.SQLDataStore {
	t.Helper()

	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	if err := createCredentialTestSchema(driver); err != nil {
		driver.Close()
		t.Fatalf("Failed to create credential test schema: %v", err)
	}

	return db.NewSQLDataStore(driver, nil)
}

// === Dual-Field Credential Handler Tests (v0.37.1) ===

// TestCredentialHandler_Apply_DualField_BothVars verifies that applying a vault
// credential YAML with both usernameVar and passwordVar stores both fields correctly.
func TestCredentialHandler_Apply_DualField_BothVars(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "testlab"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Setup: CreateEcosystem() error = %v", err)
	}

	ctx := resource.Context{DataStore: ds}

	yamlData := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-creds
  ecosystem: testlab
spec:
  source: vault
  vaultSecret: github/token
  usernameVar: GITHUB_USERNAME
  passwordVar: GITHUB_PAT
  description: GitHub dual-field credential`

	res, err := h.Apply(ctx, []byte(yamlData))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindCredential {
		t.Errorf("Apply() resource.Kind = %q, want %q", res.GetKind(), KindCredential)
	}

	credRes, ok := res.(*CredentialResource)
	if !ok {
		t.Fatalf("Apply() result is not *CredentialResource")
	}

	cred := credRes.Credential()

	if cred.UsernameVar == nil {
		t.Fatalf("Apply() UsernameVar is nil, want %q", "GITHUB_USERNAME")
	}
	if *cred.UsernameVar != "GITHUB_USERNAME" {
		t.Errorf("Apply() UsernameVar = %q, want %q", *cred.UsernameVar, "GITHUB_USERNAME")
	}

	if cred.PasswordVar == nil {
		t.Fatalf("Apply() PasswordVar is nil, want %q", "GITHUB_PAT")
	}
	if *cred.PasswordVar != "GITHUB_PAT" {
		t.Errorf("Apply() PasswordVar = %q, want %q", *cred.PasswordVar, "GITHUB_PAT")
	}
}

// TestCredentialHandler_Apply_DualField_PasswordOnly verifies that applying a vault
// credential YAML with only passwordVar (no usernameVar) stores passwordVar and
// leaves UsernameVar nil.
func TestCredentialHandler_Apply_DualField_PasswordOnly(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "testlab"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Setup: CreateEcosystem() error = %v", err)
	}

	ctx := resource.Context{DataStore: ds}

	yamlData := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: docker-token
  ecosystem: testlab
spec:
  source: vault
  vaultSecret: docker/token
  passwordVar: DOCKER_TOKEN`

	res, err := h.Apply(ctx, []byte(yamlData))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	credRes, ok := res.(*CredentialResource)
	if !ok {
		t.Fatalf("Apply() result is not *CredentialResource")
	}

	cred := credRes.Credential()

	if cred.UsernameVar != nil {
		t.Errorf("Apply() UsernameVar = %q, want nil", *cred.UsernameVar)
	}

	if cred.PasswordVar == nil {
		t.Fatalf("Apply() PasswordVar is nil, want %q", "DOCKER_TOKEN")
	}
	if *cred.PasswordVar != "DOCKER_TOKEN" {
		t.Errorf("Apply() PasswordVar = %q, want %q", *cred.PasswordVar, "DOCKER_TOKEN")
	}
}

// TestCredentialHandler_Apply_DualField_Update verifies that applying a credential
// without vars and then re-applying with both usernameVar and passwordVar correctly
// updates the stored credential to include those vars.
func TestCredentialHandler_Apply_DualField_Update(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "dualfield-lab"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Setup: CreateEcosystem() error = %v", err)
	}

	ctx := resource.Context{DataStore: ds}

	// Step 1: create without vars
	createYAML := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: hub-creds
  ecosystem: dualfield-lab
spec:
  source: vault
  vaultSecret: hub/token
  description: initial no-vars credential`

	_, err := h.Apply(ctx, []byte(createYAML))
	if err != nil {
		t.Fatalf("Apply() create error = %v", err)
	}

	// Step 2: update with both vars
	updateYAML := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: hub-creds
  ecosystem: dualfield-lab
spec:
  source: vault
  vaultSecret: hub/token
  usernameVar: HUB_USER
  passwordVar: HUB_PASS
  description: updated with dual-field vars`

	res, err := h.Apply(ctx, []byte(updateYAML))
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}

	credRes, ok := res.(*CredentialResource)
	if !ok {
		t.Fatalf("Apply() result is not *CredentialResource")
	}

	cred := credRes.Credential()

	if cred.UsernameVar == nil {
		t.Fatalf("Apply() updated UsernameVar is nil, want %q", "HUB_USER")
	}
	if *cred.UsernameVar != "HUB_USER" {
		t.Errorf("Apply() updated UsernameVar = %q, want %q", *cred.UsernameVar, "HUB_USER")
	}

	if cred.PasswordVar == nil {
		t.Fatalf("Apply() updated PasswordVar is nil, want %q", "HUB_PASS")
	}
	if *cred.PasswordVar != "HUB_PASS" {
		t.Errorf("Apply() updated PasswordVar = %q, want %q", *cred.PasswordVar, "HUB_PASS")
	}
}

// TestCredentialHandler_Apply_DualField_WithEnvSource_Rejected verifies that
// applying a credential YAML with usernameVar or passwordVar combined with
// source: env is rejected because dual-field vars are only valid for vault.
func TestCredentialHandler_Apply_DualField_WithEnvSource_Rejected(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "testlab"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Setup: CreateEcosystem() error = %v", err)
	}

	ctx := resource.Context{DataStore: ds}

	yamlData := `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: bad-cred
  ecosystem: testlab
spec:
  source: env
  envVar: TOKEN
  usernameVar: MY_USER`

	_, err := h.Apply(ctx, []byte(yamlData))
	if err == nil {
		t.Fatalf("Apply() expected error for env source with usernameVar, but got none")
	}

	if !contains(err.Error(), "vault") {
		t.Errorf("Apply() error = %q, want error containing %q", err.Error(), "vault")
	}
}

// TestCredentialHandler_ToYAML_DualField verifies that a CredentialDB with
// UsernameVar and PasswordVar set produces YAML output containing
// the usernameVar and passwordVar fields.
func TestCredentialHandler_ToYAML_DualField(t *testing.T) {
	h := NewCredentialHandler()

	vaultSecret := "github/token"
	desc := "GitHub dual-field credential"
	user := "GITHUB_USERNAME"
	pass := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:        "github-creds",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     1,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		Description: &desc,
		UsernameVar: &user,
		PasswordVar: &pass,
	}

	res := &CredentialResource{
		credential: cred,
		scopeName:  "testlab",
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	if !contains(yamlStr, "usernameVar:") {
		t.Errorf("ToYAML() missing 'usernameVar:', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "GITHUB_USERNAME") {
		t.Errorf("ToYAML() missing 'GITHUB_USERNAME', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "passwordVar:") {
		t.Errorf("ToYAML() missing 'passwordVar:', got:\n%s", yamlStr)
	}
	if !contains(yamlStr, "GITHUB_PAT") {
		t.Errorf("ToYAML() missing 'GITHUB_PAT', got:\n%s", yamlStr)
	}
}

// createCredentialTestSchema creates all tables needed for credential handler tests.
// This includes the full scope hierarchy (ecosystems → domains → apps → workspaces)
// plus the credentials table and dvm_context table.
func createCredentialTestSchema(driver db.Driver) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT    NOT NULL UNIQUE,
			description TEXT,
			theme       TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS domains (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL REFERENCES ecosystems(id),
			name         TEXT    NOT NULL,
			description  TEXT,
			theme        TEXT,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(ecosystem_id, name)
		)`,

		`CREATE TABLE IF NOT EXISTS apps (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id    INTEGER NOT NULL REFERENCES domains(id),
			name         TEXT    NOT NULL,
			path         TEXT,
			description  TEXT,
			theme        TEXT,
			language     TEXT,
			build_config TEXT,
			git_repo_id  INTEGER,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(domain_id, name)
		)`,

		`CREATE TABLE IF NOT EXISTS workspaces (
			id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id                INTEGER NOT NULL REFERENCES apps(id),
			name                  TEXT    NOT NULL,
			slug                  TEXT,
			description           TEXT,
			image_name            TEXT    DEFAULT '',
			container_id          TEXT,
			status                TEXT    DEFAULT 'stopped',
			ssh_agent_forwarding  INTEGER DEFAULT 0,
			nvim_structure        TEXT,
			nvim_plugins          TEXT,
			theme                 TEXT,
			terminal_prompt       TEXT,
			terminal_plugins      TEXT,
			terminal_package      TEXT,
			git_repo_id           INTEGER,
			env                   TEXT    NOT NULL DEFAULT '{}',
			created_at            DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at            DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(app_id, name)
		)`,

		`CREATE TABLE IF NOT EXISTS credentials (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			scope_type           TEXT    NOT NULL CHECK(scope_type IN ('ecosystem','domain','app','workspace')),
			scope_id             INTEGER NOT NULL,
			name                 TEXT    NOT NULL,
			source               TEXT    NOT NULL CHECK(source IN ('vault','env')),
			env_var              TEXT,
			description          TEXT,
			username_var         TEXT,
			password_var         TEXT,
			vault_secret         TEXT,
			vault_env            TEXT,
			vault_username_secret TEXT,
			created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(scope_type, scope_id, name)
		)`,

		// context holds one row (id=1) with the active selection state.
		`CREATE TABLE IF NOT EXISTS context (
			id                  INTEGER PRIMARY KEY DEFAULT 1,
			active_ecosystem_id INTEGER REFERENCES ecosystems(id),
			active_domain_id    INTEGER REFERENCES domains(id),
			active_app_id       INTEGER REFERENCES apps(id),
			active_workspace_id INTEGER REFERENCES workspaces(id),
			updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Seed the single context row.
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
	}

	for _, stmt := range statements {
		if _, err := driver.Execute(stmt); err != nil {
			return err
		}
	}

	return nil
}
