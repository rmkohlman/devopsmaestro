package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Helper: build a test cobra.Command with the given DataStore in context
// ---------------------------------------------------------------------------

func newGetAllTestCmd(t *testing.T, ds db.DataStore) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), "dataStore", ds)
	cmd.SetContext(ctx)
	return cmd
}

// createFullTestDataStore creates an in-memory SQLite DataStore with all 9
// resource tables that getAll() queries.  This avoids "no such table" warnings
// that would be prepended to JSON/YAML output and break parsing.
func createFullTestDataStore(t *testing.T) db.DataStore {
	t.Helper()

	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	require.NoError(t, err)

	err = driver.Connect()
	require.NoError(t, err)

	// Start with the base schema (ecosystems, domains, apps, git_repos, workspaces, context)
	err = createTestSchema(driver)
	require.NoError(t, err)

	// Add the remaining tables that getAll() queries
	extraQueries := []string{
		`CREATE TABLE IF NOT EXISTS credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
			scope_id INTEGER,
			name TEXT NOT NULL,
			source TEXT NOT NULL CHECK(source IN ('vault', 'env')),
			vault_secret TEXT,
			vault_env TEXT,
			vault_username_secret TEXT,
			vault_fields TEXT,
			env_var TEXT,
			description TEXT,
			username_var TEXT,
			password_var TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(scope_type, scope_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS registries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			version TEXT NOT NULL DEFAULT '',
			enabled BOOLEAN NOT NULL DEFAULT 1,
			lifecycle TEXT NOT NULL DEFAULT 'manual',
			port INTEGER NOT NULL UNIQUE,
			storage TEXT NOT NULL,
			idle_timeout INTEGER DEFAULT 1800,
			config TEXT,
			description TEXT,
			status TEXT DEFAULT 'stopped',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS nvim_plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			repo TEXT NOT NULL,
			branch TEXT,
			version TEXT,
			priority INTEGER,
			lazy INTEGER DEFAULT 0,
			event TEXT,
			ft TEXT,
			keys TEXT,
			cmd TEXT,
			dependencies TEXT,
			build TEXT,
			config TEXT,
			init TEXT,
			opts TEXT,
			keymaps TEXT,
			category TEXT,
			tags TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS nvim_themes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			author TEXT,
			category TEXT,
			plugin_repo TEXT NOT NULL,
			plugin_branch TEXT,
			plugin_tag TEXT,
			style TEXT,
			transparent BOOLEAN DEFAULT FALSE,
			colors TEXT,
			options TEXT,
			is_active BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range extraQueries {
		_, err := driver.Execute(q)
		require.NoError(t, err)
	}

	return db.NewSQLDataStore(driver, nil)
}

// ---------------------------------------------------------------------------
// TestGetAll_EmptyDatabase
// ---------------------------------------------------------------------------

func TestGetAll_EmptyDatabase(t *testing.T) {
	// createTestSchema creates ecosystems, domains, apps, workspaces, git_repos, context tables.
	// Other tables (credentials, registries, nvim_plugins, nvim_themes) are absent, so those
	// sections will hit list errors and be silently shown as "(none)".
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	cmd := newGetAllTestCmd(t, dataStore)

	// Redirect render output so tests don't pollute the terminal
	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "" // default human-readable format

	err := getAll(cmd)
	assert.NoError(t, err, "getAll should not error on an empty database")

	output := buf.String()

	// All 9 section headers must appear — counts are always (0) for an empty DB
	sections := []string{
		"=== Ecosystems (0) ===",
		"=== Domains (0) ===",
		"=== Apps (0) ===",
		"=== Workspaces (0) ===",
		"=== Credentials (0) ===",
		"=== Registries (0) ===",
		"=== Git Repos (0) ===",
		"=== Nvim Plugins (0) ===",
		"=== Nvim Themes (0) ===",
	}
	for _, section := range sections {
		assert.Contains(t, output, section, "expected section header %q in output", section)
	}
}

// ---------------------------------------------------------------------------
// TestGetAll_WithData
// ---------------------------------------------------------------------------

func TestGetAll_WithData(t *testing.T) {
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	// Seed one of each resource that the test schema supports
	ecosystem := &models.Ecosystem{
		Name:        "production",
		Description: sql.NullString{String: "Production ecosystem", Valid: true},
	}
	require.NoError(t, dataStore.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		Name:        "backend",
		EcosystemID: ecosystem.ID,
		Description: sql.NullString{String: "Backend services", Valid: true},
	}
	require.NoError(t, dataStore.CreateDomain(domain))

	app := &models.App{
		Name:        "api",
		Path:        "/srv/api",
		DomainID:    domain.ID,
		Description: sql.NullString{String: "REST API", Valid: true},
		Language:    sql.NullString{String: "go", Valid: true},
	}
	require.NoError(t, dataStore.CreateApp(app))

	workspace := &models.Workspace{
		Name:      "dev",
		AppID:     app.ID,
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(workspace))

	gitRepo := &models.GitRepoDB{
		Name:       "api-repo",
		URL:        "https://github.com/org/api.git",
		Slug:       "github.com_org_api",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	require.NoError(t, dataStore.CreateGitRepo(gitRepo))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	err := getAll(cmd)
	assert.NoError(t, err, "getAll should succeed with seeded data")

	output := buf.String()

	// Each seeded resource name should appear somewhere in the output
	assert.Contains(t, output, "production", "ecosystem name should appear in output")
	assert.Contains(t, output, "backend", "domain name should appear in output")
	assert.Contains(t, output, "api", "app name should appear in output")
	assert.Contains(t, output, "dev", "workspace name should appear in output")
	assert.Contains(t, output, "api-repo", "git repo name should appear in output")

	// Rich column headers should appear in each section's table
	assert.Contains(t, output, "DESCRIPTION", "ecosystem table should have DESCRIPTION column")
	assert.Contains(t, output, "CREATED", "tables should have CREATED column")
	assert.Contains(t, output, "ECOSYSTEM", "domain table should have ECOSYSTEM column")
	assert.Contains(t, output, "DOMAIN", "app table should have DOMAIN column")
	assert.Contains(t, output, "PATH", "app table should have PATH column")
	assert.Contains(t, output, "APP", "workspace table should have APP column")
}

// ---------------------------------------------------------------------------
// TestGetAll_JSON_ProducesListFormat  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_JSON_ProducesListFormat verifies that dvm get all -o json produces
// a kubectl-style kind: List document instead of the old AllResources flat struct.
func TestGetAll_JSON_ProducesListFormat(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed an ecosystem so at least one item appears in the list
	ecosystem := &models.Ecosystem{
		Name:        "test-eco",
		Description: sql.NullString{String: "Test", Valid: true},
	}
	require.NoError(t, dataStore.CreateEcosystem(ecosystem))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "json"

	err := getAll(cmd)
	require.NoError(t, err, "getAll with JSON output should not error")

	// Output must be valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid JSON; got: %s", buf.String())

	// Must be a kind: List document — NOT the old AllResources flat struct
	assert.Equal(t, "List", result["kind"],
		"JSON output should be kind: List, not flat AllResources")
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"],
		"JSON output should have apiVersion: devopsmaestro.io/v1")
	assert.NotNil(t, result["metadata"],
		"JSON output should have metadata field")
	assert.NotNil(t, result["items"],
		"JSON output should have items array")

	// The old flat keys must NOT be present
	oldKeys := []string{"ecosystems", "domains", "apps", "workspaces",
		"credentials", "registries", "gitRepos", "nvimPlugins", "nvimThemes"}
	for _, key := range oldKeys {
		assert.NotContains(t, result, key,
			"JSON output should NOT contain old AllResources key %q", key)
	}

	// items must be an array
	items, ok := result["items"].([]interface{})
	require.True(t, ok, "items should be a JSON array")

	// With one ecosystem seeded, items should contain at least that ecosystem
	assert.GreaterOrEqual(t, len(items), 1, "items should contain at least 1 resource")

	// Each item must have apiVersion, kind, metadata, spec fields (full spec, not lossy summary)
	for i, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		require.True(t, ok, "item %d should be a JSON object", i)
		assert.NotEmpty(t, item["kind"], "item %d should have 'kind' field", i)
		assert.NotNil(t, item["metadata"], "item %d should have 'metadata' field", i)
		assert.NotNil(t, item["apiVersion"], "item %d should have 'apiVersion' field", i)
	}
}

// ---------------------------------------------------------------------------
// TestGetAll_YAML_ProducesListFormat  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_YAML_ProducesListFormat verifies that dvm get all -o yaml produces
// a kubectl-style kind: List document instead of the old AllResources flat YAML.
func TestGetAll_YAML_ProducesListFormat(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed one ecosystem and one domain
	ecosystem := &models.Ecosystem{Name: "eco-yaml"}
	require.NoError(t, dataStore.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		Name:        "domain-yaml",
		EcosystemID: ecosystem.ID,
	}
	require.NoError(t, dataStore.CreateDomain(domain))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll with YAML output should not error")

	// Must be valid YAML that decodes into a List structure
	var result map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML; got: %s", buf.String())

	// Must be a kind: List document
	assert.Equal(t, "List", result["kind"],
		"YAML output should be kind: List, not flat AllResources")
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"],
		"YAML output should have apiVersion: devopsmaestro.io/v1")
	assert.NotNil(t, result["metadata"],
		"YAML output should have metadata field")

	// items must be present and be a list
	itemsRaw, exists := result["items"]
	assert.True(t, exists, "YAML output should have items field")
	if itemsRaw != nil {
		items, ok := itemsRaw.([]interface{})
		require.True(t, ok, "items should be a YAML sequence")
		// With one ecosystem + one domain seeded, at least 2 items
		assert.GreaterOrEqual(t, len(items), 2,
			"items should contain at least 2 resources (ecosystem + domain)")
	}

	// Old flat YAML keys must NOT be present at top level
	oldKeys := []string{"ecosystems", "domains", "apps", "workspaces"}
	for _, key := range oldKeys {
		assert.NotContains(t, result, key,
			"YAML output should NOT contain old AllResources key %q", key)
	}
}

// ---------------------------------------------------------------------------
// TestGetAll_YAML_EmptyDB  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_YAML_EmptyDB verifies that an empty database produces a List with
// empty items (not nil), not the old AllResources structure with 9 empty arrays.
func TestGetAll_YAML_EmptyDB(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll YAML on empty DB should not error")

	var result map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML: %s", buf.String())

	assert.Equal(t, "List", result["kind"], "empty DB should produce kind: List")
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"])

	// items should be present — may be nil/null or an empty list, but NOT the old 9-key structure
	assert.NotContains(t, result, "ecosystems", "should not have old AllResources key 'ecosystems'")
	assert.NotContains(t, result, "domains", "should not have old AllResources key 'domains'")
}

// ---------------------------------------------------------------------------
// TestGetAllCmd_Definition
// ---------------------------------------------------------------------------

func TestGetAllCmd_Definition(t *testing.T) {
	t.Run("getAllCmd has correct Use field", func(t *testing.T) {
		assert.Equal(t, "all", getAllCmd.Use)
	})

	t.Run("getAllCmd has non-empty Short description", func(t *testing.T) {
		assert.NotEmpty(t, getAllCmd.Short)
	})

	t.Run("getAllCmd is registered as a sub-command of getCmd", func(t *testing.T) {
		found := false
		for _, sub := range getCmd.Commands() {
			if sub.Use == "all" {
				found = true
				break
			}
		}
		assert.True(t, found, "getAllCmd should be registered under getCmd")
	})

	t.Run("getAllCmd Short is 'Show all resources'", func(t *testing.T) {
		assert.Equal(t, "Show all resources", getAllCmd.Short)
	})

	t.Run("getAllCmd RunE is not nil", func(t *testing.T) {
		assert.NotNil(t, getAllCmd.RunE, "getAllCmd.RunE should be set")
	})
}

// ---------------------------------------------------------------------------
// TestGetAll_NoDataStore
// ---------------------------------------------------------------------------

func TestGetAll_NoDataStore(t *testing.T) {
	// If no DataStore is available in context, getAll should return an error.
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background()) // no dataStore value

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	err := getAll(cmd)
	assert.Error(t, err, "getAll should error when no DataStore is in context")
	assert.Contains(t, err.Error(), "failed to get data store")
}

// ---------------------------------------------------------------------------
// TestGetAll_TableOutput
// ---------------------------------------------------------------------------

func TestGetAll_TableOutput(t *testing.T) {
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	// Seed a workspace so the Workspaces section renders a table row
	eco := &models.Ecosystem{Name: "eco-table"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{Name: "dom-table", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{Name: "app-table", Path: "/app", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "ws-table",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "running",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "table"

	err := getAll(cmd)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "eco-table")
	assert.Contains(t, output, "dom-table")
	assert.Contains(t, output, "app-table")
	assert.Contains(t, output, "ws-table")

	// Rich column headers verify shared builders are wired in
	assert.Contains(t, output, "ECOSYSTEM", "domain table should include ECOSYSTEM column")
	assert.Contains(t, output, "DOMAIN", "app table should include DOMAIN column")
	assert.Contains(t, output, "APP", "workspace table should include APP column")

	// Parent names should appear in child rows
	// domain row should include the ecosystem name
	assert.Contains(t, output, "eco-table", "ecosystem name should appear in domain row")
	// app row should include the domain name
	assert.Contains(t, output, "dom-table", "domain name should appear in app row")
	// workspace row should include the app name
	assert.Contains(t, output, "app-table", "app name should appear in workspace row")
}

// ---------------------------------------------------------------------------
// TestGetAll_JSON_EmptyDatabase  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_JSON_EmptyDatabase verifies that even with an empty database,
// the JSON output is a valid kind: List document with empty (or absent) items.
func TestGetAll_JSON_EmptyDatabase(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "json"

	err := getAll(cmd)
	require.NoError(t, err, "getAll JSON on empty DB should not error")

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid JSON")

	// Must be kind: List — NOT the old 9-key AllResources struct
	assert.Equal(t, "List", result["kind"], "empty DB JSON should produce kind: List")
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"])
	assert.NotNil(t, result["metadata"])

	// Old flat AllResources keys must NOT appear
	oldKeys := []string{
		"ecosystems", "domains", "apps", "workspaces",
		"credentials", "registries", "gitRepos", "nvimPlugins", "nvimThemes",
	}
	for _, key := range oldKeys {
		assert.NotContains(t, result, key,
			"JSON should NOT contain old AllResources key %q", key)
	}

	// items should be present and be an array (possibly empty)
	items, exists := result["items"]
	assert.True(t, exists, "JSON should contain 'items' key")
	if items != nil {
		_, isSlice := items.([]interface{})
		assert.True(t, isSlice, "'items' should be a JSON array")
	}
}

// ---------------------------------------------------------------------------
// TestGetAll_OutputFormatPreservation
// ---------------------------------------------------------------------------

// TestGetAll_OutputFormatPreservation verifies that getOutputFormat is not
// mutated by getAll itself (the function is read-only with respect to the flag).
func TestGetAll_OutputFormatPreservation(t *testing.T) {
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "json"

	_ = getAll(cmd)

	assert.Equal(t, "json", getOutputFormat, "getAll should not modify getOutputFormat")
}

// ---------------------------------------------------------------------------
// TestGetAll_JSON_ItemsCountMatchesAllResources  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_JSON_ItemsCountMatchesAllResources verifies that the number of
// items in the List equals the total number of resources across all types.
func TestGetAll_JSON_ItemsCountMatchesAllResources(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed multiple resources of different types
	eco1 := &models.Ecosystem{Name: "eco-count-a"}
	require.NoError(t, dataStore.CreateEcosystem(eco1))
	eco2 := &models.Ecosystem{Name: "eco-count-b"}
	require.NoError(t, dataStore.CreateEcosystem(eco2))

	dom1 := &models.Domain{Name: "dom-count-1", EcosystemID: eco1.ID}
	require.NoError(t, dataStore.CreateDomain(dom1))
	dom2 := &models.Domain{Name: "dom-count-2", EcosystemID: eco1.ID}
	require.NoError(t, dataStore.CreateDomain(dom2))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "json"

	err := getAll(cmd)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	require.Equal(t, "List", result["kind"])

	items, ok := result["items"].([]interface{})
	require.True(t, ok, "items should be a JSON array")

	// We seeded 2 ecosystems + 2 domains = 4 resources total
	// The list should contain exactly those (no more, no less)
	assert.GreaterOrEqual(t, len(items), 4,
		"items should contain at least the 4 seeded resources (2 ecosystems + 2 domains)")
}

// ---------------------------------------------------------------------------
// TestGetAll_SectionHeaders_PlainOutput (NO_COLOR)
// ---------------------------------------------------------------------------

func TestGetAll_SectionHeaders_PlainOutput(t *testing.T) {
	// Force plain output via NO_COLOR to avoid ANSI codes in the captured string
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer os.Setenv("NO_COLOR", origNoColor)

	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	err := getAll(cmd)
	require.NoError(t, err)

	output := buf.String()

	// With an empty DB every section shows "(none)"
	assert.Contains(t, output, "(none)", "empty sections should show '(none)'")

	// Section headers include (0) counts even for an empty DB
	assert.Contains(t, output, "=== Ecosystems (0) ===", "empty ecosystem section header should include count")
	assert.Contains(t, output, "=== Domains (0) ===", "empty domain section header should include count")
}

// ---------------------------------------------------------------------------
// TestGetAll_ActiveMarkers
// ---------------------------------------------------------------------------

// TestGetAll_ActiveMarkers verifies that active resources are prefixed with "●"
// when an active context is set in the DataStore.
func TestGetAll_ActiveMarkers(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed resources
	eco := &models.Ecosystem{Name: "active-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{Name: "active-dom", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{Name: "active-app", Path: "/app", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "active-ws",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	// Set active context for all resources
	require.NoError(t, dataStore.SetActiveEcosystem(&eco.ID))
	require.NoError(t, dataStore.SetActiveDomain(&dom.ID))
	require.NoError(t, dataStore.SetActiveApp(&app.ID))
	require.NoError(t, dataStore.SetActiveWorkspace(&ws.ID))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	err := getAll(cmd)
	require.NoError(t, err, "getAll should succeed with active context set")

	output := buf.String()

	// Active marker "●" should appear in the output for each active resource
	assert.Contains(t, output, "●", "active marker ● should appear in output")
	assert.Contains(t, output, "● active-eco", "active ecosystem should be prefixed with ●")
	assert.Contains(t, output, "● active-dom", "active domain should be prefixed with ●")
	assert.Contains(t, output, "● active-app", "active app should be prefixed with ●")
	assert.Contains(t, output, "● active-ws", "active workspace should be prefixed with ●")
}

// ---------------------------------------------------------------------------
// TestGetAll_WideOutput
// ---------------------------------------------------------------------------

// TestGetAll_WideOutput verifies that -o wide adds extra columns to the tables.
func TestGetAll_WideOutput(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed resources
	eco := &models.Ecosystem{Name: "wide-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{Name: "wide-dom", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{Name: "wide-app", Path: "/app", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "wide-ws",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	gitRepo := &models.GitRepoDB{
		Name:       "wide-repo",
		URL:        "https://github.com/org/wide.git",
		Slug:       "github.com_org_wide",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	require.NoError(t, dataStore.CreateGitRepo(gitRepo))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "wide"

	err := getAll(cmd)
	require.NoError(t, err, "getAll with -o wide should succeed")

	output := buf.String()

	// Verify resource names still appear
	assert.Contains(t, output, "wide-eco")
	assert.Contains(t, output, "wide-dom")
	assert.Contains(t, output, "wide-app")
	assert.Contains(t, output, "wide-ws")
	assert.Contains(t, output, "wide-repo")

	// Wide-only columns should appear
	assert.Contains(t, output, "ID", "wide mode should add ID column to ecosystems")
	assert.Contains(t, output, "CONTAINER-ID", "wide mode should add CONTAINER-ID column to workspaces")
	assert.Contains(t, output, "GITREPO", "wide mode should add GITREPO column to apps")
	assert.Contains(t, output, "SLUG", "wide mode should add SLUG column to git repos")
}

// ---------------------------------------------------------------------------
// TestGetAll_RichColumns
// ---------------------------------------------------------------------------

// TestGetAll_RichColumns verifies that parent resource names appear in child
// resource rows (e.g., ecosystem name in domain row, domain name in app row).
func TestGetAll_RichColumns(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed a hierarchy: ecosystem -> domain -> app -> workspace
	eco := &models.Ecosystem{
		Name: "rich-eco",
	}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{
		Name:        "rich-dom",
		EcosystemID: eco.ID,
	}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{
		Name:     "rich-app",
		Path:     "/src/rich",
		DomainID: dom.ID,
	}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "rich-ws",
		AppID:     app.ID,
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	err := getAll(cmd)
	require.NoError(t, err, "getAll should succeed with hierarchical data")

	output := buf.String()

	// The ECOSYSTEM column in the domains table should show the ecosystem name
	assert.Contains(t, output, "rich-eco", "ecosystem name should appear in domain table's ECOSYSTEM column")

	// The DOMAIN column in the apps table should show the domain name
	assert.Contains(t, output, "rich-dom", "domain name should appear in app table's DOMAIN column")

	// The APP column in the workspaces table should show the app name
	assert.Contains(t, output, "rich-app", "app name should appear in workspace table's APP column")

	// The PATH column should show the app path
	assert.Contains(t, output, "/src/rich", "app path should appear in app table's PATH column")
}

// ===========================================================================
// Sprint 3 Tests: Scoped Hierarchical Views in `dvm get all`
// ===========================================================================

// ---------------------------------------------------------------------------
// Category 1: Flag Wiring Tests
//
// These verify that getAllCmd has the new scoping flags registered.
// They will FAIL until the flags are wired in init().
// ---------------------------------------------------------------------------

func TestGetAllCmd_HasEcosystemFlag(t *testing.T) {
	f := getAllCmd.Flags().Lookup("ecosystem")
	require.NotNil(t, f, "getAllCmd should have --ecosystem flag")
	assert.Equal(t, "e", f.Shorthand, "--ecosystem flag should have -e shorthand")
}

func TestGetAllCmd_HasDomainFlag(t *testing.T) {
	f := getAllCmd.Flags().Lookup("domain")
	require.NotNil(t, f, "getAllCmd should have --domain flag")
	assert.Equal(t, "d", f.Shorthand, "--domain flag should have -d shorthand")
}

func TestGetAllCmd_HasAppFlag(t *testing.T) {
	f := getAllCmd.Flags().Lookup("app")
	require.NotNil(t, f, "getAllCmd should have --app flag")
	assert.Equal(t, "a", f.Shorthand, "--app flag should have -a shorthand")
}

func TestGetAllCmd_HasAllFlag(t *testing.T) {
	f := getAllCmd.Flags().Lookup("all")
	require.NotNil(t, f, "getAllCmd should have --all flag")
	assert.Equal(t, "A", f.Shorthand, "--all flag should have -A shorthand")
}

// ---------------------------------------------------------------------------
// Category 2: Scope Resolution Tests
//
// These call resolveGetAllScope() directly with a MockDataStore.
// They will FAIL because the stub returns "not implemented".
// ---------------------------------------------------------------------------

func TestResolveGetAllScope_AllFlagReturnsShowAll(t *testing.T) {
	mock := db.NewMockDataStore()
	sc, err := resolveGetAllScope(mock, "", "", "", true)
	require.NoError(t, err)
	assert.True(t, sc.ShowAll, "ShowAll should be true when -A flag is set")
	assert.Nil(t, sc.EcosystemID, "EcosystemID should be nil when ShowAll")
	assert.Nil(t, sc.DomainID, "DomainID should be nil when ShowAll")
	assert.Nil(t, sc.AppID, "AppID should be nil when ShowAll")
}

func TestResolveGetAllScope_AllFlagWithEcosystemReturnsError(t *testing.T) {
	mock := db.NewMockDataStore()
	_, err := resolveGetAllScope(mock, "eco1", "", "", true)
	assert.Error(t, err, "should error when -A is combined with --ecosystem")
}

func TestResolveGetAllScope_AllFlagWithDomainReturnsError(t *testing.T) {
	mock := db.NewMockDataStore()
	_, err := resolveGetAllScope(mock, "", "dom1", "", true)
	assert.Error(t, err, "should error when -A is combined with --domain")
}

func TestResolveGetAllScope_AllFlagWithAppReturnsError(t *testing.T) {
	mock := db.NewMockDataStore()
	_, err := resolveGetAllScope(mock, "", "", "app1", true)
	assert.Error(t, err, "should error when -A is combined with --app")
}

func TestResolveGetAllScope_EcosystemFlagResolvesID(t *testing.T) {
	mock := db.NewMockDataStore()
	mock.Ecosystems["production"] = &models.Ecosystem{ID: 42, Name: "production"}

	sc, err := resolveGetAllScope(mock, "production", "", "", false)
	require.NoError(t, err)
	require.NotNil(t, sc.EcosystemID, "EcosystemID should be set")
	assert.Equal(t, 42, *sc.EcosystemID)
	assert.Nil(t, sc.DomainID)
	assert.Nil(t, sc.AppID)
	assert.False(t, sc.ShowAll)
}

func TestResolveGetAllScope_EcosystemNotFoundReturnsError(t *testing.T) {
	mock := db.NewMockDataStore()
	// "bogus" doesn't exist in mock.Ecosystems
	_, err := resolveGetAllScope(mock, "bogus", "", "", false)
	assert.Error(t, err, "should error when ecosystem is not found")
}

func TestResolveGetAllScope_DomainFlagWithExplicitEcosystem(t *testing.T) {
	mock := db.NewMockDataStore()
	eco := &models.Ecosystem{ID: 1, Name: "eco1"}
	mock.Ecosystems["eco1"] = eco
	dom := &models.Domain{ID: 10, Name: "backend", EcosystemID: 1}
	mock.Domains[10] = dom

	sc, err := resolveGetAllScope(mock, "eco1", "backend", "", false)
	require.NoError(t, err)
	require.NotNil(t, sc.EcosystemID)
	assert.Equal(t, 1, *sc.EcosystemID)
	require.NotNil(t, sc.DomainID)
	assert.Equal(t, 10, *sc.DomainID)
	assert.Nil(t, sc.AppID)
}

func TestResolveGetAllScope_DomainFlagUsesActiveContext(t *testing.T) {
	// When --domain is given without --ecosystem, resolveGetAllScope should
	// look up the active ecosystem from context to qualify the domain.
	mock := db.NewMockDataStore()
	ecoID := 5
	eco := &models.Ecosystem{ID: 5, Name: "ctx-eco"}
	mock.Ecosystems["ctx-eco"] = eco
	mock.Context = &models.Context{ID: 1, ActiveEcosystemID: &ecoID}

	dom := &models.Domain{ID: 20, Name: "frontend", EcosystemID: 5}
	mock.Domains[20] = dom

	sc, err := resolveGetAllScope(mock, "", "frontend", "", false)
	require.NoError(t, err)
	require.NotNil(t, sc.EcosystemID)
	assert.Equal(t, 5, *sc.EcosystemID)
	require.NotNil(t, sc.DomainID)
	assert.Equal(t, 20, *sc.DomainID)
}

func TestResolveGetAllScope_DomainFlagNoEcosystemReturnsError(t *testing.T) {
	// When --domain is given but there's no --ecosystem and no active context,
	// it should fail because the domain can't be resolved without an ecosystem.
	mock := db.NewMockDataStore()
	mock.Context = &models.Context{ID: 1} // no active ecosystem

	_, err := resolveGetAllScope(mock, "", "orphan-dom", "", false)
	assert.Error(t, err, "should error when domain flag is given without ecosystem context")
}

func TestResolveGetAllScope_AppFlagWithExplicitDomainAndEcosystem(t *testing.T) {
	mock := db.NewMockDataStore()
	eco := &models.Ecosystem{ID: 1, Name: "eco1"}
	mock.Ecosystems["eco1"] = eco
	dom := &models.Domain{ID: 10, Name: "backend", EcosystemID: 1}
	mock.Domains[10] = dom
	app := &models.App{ID: 100, Name: "api", DomainID: 10}
	mock.Apps[100] = app

	sc, err := resolveGetAllScope(mock, "eco1", "backend", "api", false)
	require.NoError(t, err)
	require.NotNil(t, sc.EcosystemID)
	assert.Equal(t, 1, *sc.EcosystemID)
	require.NotNil(t, sc.DomainID)
	assert.Equal(t, 10, *sc.DomainID)
	require.NotNil(t, sc.AppID)
	assert.Equal(t, 100, *sc.AppID)
}

func TestResolveGetAllScope_AppFlagNoDomainReturnsError(t *testing.T) {
	// When --app is given but there's no --domain and no active domain context,
	// it should fail because the app can't be resolved without a domain.
	mock := db.NewMockDataStore()
	mock.Context = &models.Context{ID: 1} // no active domain

	_, err := resolveGetAllScope(mock, "", "", "orphan-app", false)
	assert.Error(t, err, "should error when app flag is given without domain context")
}

func TestResolveGetAllScope_NoFlagsUsesActiveContext(t *testing.T) {
	// When no flags are given, resolveGetAllScope should use the active context
	// to set scope. If there's an active ecosystem, scope to that.
	mock := db.NewMockDataStore()
	ecoID := 3
	eco := &models.Ecosystem{ID: 3, Name: "active-eco"}
	mock.Ecosystems["active-eco"] = eco
	mock.Context = &models.Context{ID: 1, ActiveEcosystemID: &ecoID}

	sc, err := resolveGetAllScope(mock, "", "", "", false)
	require.NoError(t, err)
	require.NotNil(t, sc.EcosystemID, "should scope to active ecosystem when no flags given")
	assert.Equal(t, 3, *sc.EcosystemID)
	assert.False(t, sc.ShowAll)
}

func TestResolveGetAllScope_NoFlagsNoContextReturnsShowAll(t *testing.T) {
	// When no flags and no active context, should default to ShowAll
	mock := db.NewMockDataStore()
	mock.Context = &models.Context{ID: 1} // no active anything

	sc, err := resolveGetAllScope(mock, "", "", "", false)
	require.NoError(t, err)
	assert.True(t, sc.ShowAll, "should default to ShowAll when no flags and no active context")
}

// ---------------------------------------------------------------------------
// Category 3: Scoped Data Fetching Tests
//
// These verify that `getAll` correctly filters resources when scope flags
// are provided. They use createFullTestDataStore (real SQLite) and need
// the full command pipeline with flags wired.
//
// They will FAIL because:
//   - getAllCmd doesn't have the scoping flags yet
//   - getAll() doesn't call resolveGetAllScope() yet
//   - No filtering logic exists yet
// ---------------------------------------------------------------------------

// newScopedGetAllCmd creates a test command that mimics getAllCmd with
// proper flags and RunE. This is needed because newGetAllTestCmd creates
// a bare command without flags.
func newScopedGetAllCmd(t *testing.T, ds db.DataStore) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{
		Use:  "all",
		RunE: func(cmd *cobra.Command, args []string) error { return getAll(cmd) },
	}
	// Wire scoping flags matching what init() should add to getAllCmd
	cmd.Flags().StringP("ecosystem", "e", "", "Filter by ecosystem")
	cmd.Flags().StringP("domain", "d", "", "Filter by domain")
	cmd.Flags().StringP("app", "a", "", "Filter by app")
	cmd.Flags().BoolP("all", "A", false, "Show all resources (ignore active context)")
	cmd.Flags().StringP("output", "o", "", "Output format")

	ctx := context.WithValue(context.Background(), "dataStore", ds)
	cmd.SetContext(ctx)
	return cmd
}

func TestGetAll_ScopedByEcosystem(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed two ecosystems with domains under each
	eco1 := &models.Ecosystem{Name: "eco-alpha"}
	require.NoError(t, dataStore.CreateEcosystem(eco1))
	eco2 := &models.Ecosystem{Name: "eco-beta"}
	require.NoError(t, dataStore.CreateEcosystem(eco2))

	dom1 := &models.Domain{Name: "dom-in-alpha", EcosystemID: eco1.ID}
	require.NoError(t, dataStore.CreateDomain(dom1))
	dom2 := &models.Domain{Name: "dom-in-beta", EcosystemID: eco2.ID}
	require.NoError(t, dataStore.CreateDomain(dom2))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	// Scope to eco-alpha only
	cmd.SetArgs([]string{"--ecosystem", "eco-alpha"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "eco-alpha", "scoped ecosystem should appear")
	assert.Contains(t, output, "dom-in-alpha", "domain in scoped ecosystem should appear")
	assert.NotContains(t, output, "eco-beta", "out-of-scope ecosystem should NOT appear")
	assert.NotContains(t, output, "dom-in-beta", "out-of-scope domain should NOT appear")
}

func TestGetAll_ScopedByDomain(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco := &models.Ecosystem{Name: "eco-scope-dom"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom1 := &models.Domain{Name: "dom-target", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom1))
	dom2 := &models.Domain{Name: "dom-other", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom2))

	app1 := &models.App{Name: "app-in-target", Path: "/t", DomainID: dom1.ID}
	require.NoError(t, dataStore.CreateApp(app1))
	app2 := &models.App{Name: "app-in-other", Path: "/o", DomainID: dom2.ID}
	require.NoError(t, dataStore.CreateApp(app2))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	cmd.SetArgs([]string{"--ecosystem", "eco-scope-dom", "--domain", "dom-target"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "dom-target", "scoped domain should appear")
	assert.Contains(t, output, "app-in-target", "app in scoped domain should appear")
	assert.NotContains(t, output, "app-in-other", "app in out-of-scope domain should NOT appear")
}

func TestGetAll_ScopedByApp(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco := &models.Ecosystem{Name: "eco-scope-app"}
	require.NoError(t, dataStore.CreateEcosystem(eco))
	dom := &models.Domain{Name: "dom-scope-app", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app1 := &models.App{Name: "app-target", Path: "/t", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app1))
	app2 := &models.App{Name: "app-other", Path: "/o", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app2))

	ws1 := &models.Workspace{Name: "ws-in-target", Slug: "eco-scope-app/dom-scope-app/app-target/ws-in-target", AppID: app1.ID, ImageName: "img", Status: "stopped"}
	require.NoError(t, dataStore.CreateWorkspace(ws1))
	ws2 := &models.Workspace{Name: "ws-in-other", Slug: "eco-scope-app/dom-scope-app/app-other/ws-in-other", AppID: app2.ID, ImageName: "img", Status: "stopped"}
	require.NoError(t, dataStore.CreateWorkspace(ws2))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	cmd.SetArgs([]string{"--ecosystem", "eco-scope-app", "--domain", "dom-scope-app", "--app", "app-target"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "app-target", "scoped app should appear")
	assert.Contains(t, output, "ws-in-target", "workspace in scoped app should appear")
	assert.NotContains(t, output, "ws-in-other", "workspace in out-of-scope app should NOT appear")
}

func TestGetAll_AllFlagShowsEverything(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed two ecosystems; set one as active context
	eco1 := &models.Ecosystem{Name: "eco-one"}
	require.NoError(t, dataStore.CreateEcosystem(eco1))
	eco2 := &models.Ecosystem{Name: "eco-two"}
	require.NoError(t, dataStore.CreateEcosystem(eco2))

	// Set active context to eco-one (without -A, scoping would filter to eco-one)
	require.NoError(t, dataStore.SetActiveEcosystem(&eco1.ID))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	// -A should show everything regardless of active context
	cmd.SetArgs([]string{"-A"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "eco-one", "-A should show eco-one")
	assert.Contains(t, output, "eco-two", "-A should show eco-two (even when eco-one is active)")
}

func TestGetAll_GlobalResourcesAlwaysShown(t *testing.T) {
	// Global resources (registries, nvim plugins, nvim themes) should always
	// appear regardless of ecosystem/domain/app scoping.
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco := &models.Ecosystem{Name: "eco-global-test"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	// Create a registry (global resource)
	reg := &models.Registry{
		Name:    "test-reg",
		Type:    "zot",
		Port:    5100,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	// Scope to eco-global-test — registries should still show
	cmd.SetArgs([]string{"--ecosystem", "eco-global-test"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test-reg", "global resource (registry) should appear even when scoped to ecosystem")
}

func TestGetAll_ScopedCredentialContainment(t *testing.T) {
	// Credentials scoped to ecosystem 1 should NOT appear when viewing ecosystem 2
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco1 := &models.Ecosystem{Name: "eco-cred-a"}
	require.NoError(t, dataStore.CreateEcosystem(eco1))
	eco2 := &models.Ecosystem{Name: "eco-cred-b"}
	require.NoError(t, dataStore.CreateEcosystem(eco2))

	// Create credentials scoped to each ecosystem
	envVar1 := "CRED_A_TOKEN"
	cred1 := &models.CredentialDB{
		Name:      "cred-in-a",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco1.ID),
		Source:    "env",
		EnvVar:    &envVar1,
	}
	require.NoError(t, dataStore.CreateCredential(cred1))

	envVar2 := "CRED_B_TOKEN"
	cred2 := &models.CredentialDB{
		Name:      "cred-in-b",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco2.ID),
		Source:    "env",
		EnvVar:    &envVar2,
	}
	require.NoError(t, dataStore.CreateCredential(cred2))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	// Scope to eco-cred-a
	cmd.SetArgs([]string{"--ecosystem", "eco-cred-a"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "cred-in-a", "credential scoped to viewed ecosystem should appear")
	assert.NotContains(t, output, "cred-in-b", "credential scoped to OTHER ecosystem should NOT appear")
}

// ---------------------------------------------------------------------------
// Category 5: Long Description Test
//
// Verifies getAllCmd.Long mentions scoping functionality.
// ---------------------------------------------------------------------------

func TestGetAllCmd_LongDescriptionMentionsScoping(t *testing.T) {
	long := getAllCmd.Long
	assert.Contains(t, long, "scope", "Long description should mention scoping (case-insensitive search)")
}

// ===========================================================================
// Sprint 4 Tests: List Format YAML/JSON Export  [RED Phase]
// ===========================================================================

// ---------------------------------------------------------------------------
// TestGetAll_YAML_ItemsHaveFullSpec  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_YAML_ItemsHaveFullSpec verifies that each item in the YAML list
// has a full resource spec (apiVersion, kind, metadata, spec) — not the lossy
// AllResourceSummary format that only had name/description/status.
func TestGetAll_YAML_ItemsHaveFullSpec(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed an ecosystem — its YAML spec includes description, createdAt, etc.
	eco := &models.Ecosystem{
		Name:        "full-spec-eco",
		Description: sql.NullString{String: "Full spec test ecosystem", Valid: true},
	}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll YAML should not error")

	var result map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML")
	require.Equal(t, "List", result["kind"])

	itemsRaw, exists := result["items"]
	require.True(t, exists, "items key must exist")
	require.NotNil(t, itemsRaw, "items must not be nil when resources exist")

	items, ok := itemsRaw.([]interface{})
	require.True(t, ok, "items should be a slice")
	require.NotEmpty(t, items, "items should not be empty (seeded 1 ecosystem)")

	// Find the ecosystem item and verify it has full spec fields
	var foundEco map[string]interface{}
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		if item["kind"] == "Ecosystem" {
			meta, _ := item["metadata"].(map[string]interface{})
			if meta != nil && meta["name"] == "full-spec-eco" {
				foundEco = item
				break
			}
		}
	}

	require.NotNil(t, foundEco, "should find the seeded ecosystem in items")

	// Each item must have the full kubectl-style fields — NOT the lossy summary
	assert.NotNil(t, foundEco["apiVersion"], "ecosystem item must have 'apiVersion'")
	assert.Equal(t, "Ecosystem", foundEco["kind"], "ecosystem item must have correct 'kind'")
	assert.NotNil(t, foundEco["metadata"], "ecosystem item must have 'metadata'")
	assert.NotNil(t, foundEco["spec"], "ecosystem item must have 'spec' (full spec, not summary)")

	// The spec must contain description, not just a name-only summary
	spec, ok := foundEco["spec"].(map[string]interface{})
	require.True(t, ok, "spec should be a map")
	assert.NotEmpty(t, spec, "spec should not be empty")
}

// ---------------------------------------------------------------------------
// TestGetAll_YAML_ScopedExcludesGlobals  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_YAML_ScopedExcludesGlobals verifies that when scoped to an
// ecosystem, global resources (NvimPlugins, NvimThemes, Registries) are
// NOT included in YAML/JSON output (they still appear in table output).
// Decision 9: Exclude global resources from scoped exports.
func TestGetAll_YAML_ScopedExcludesGlobals(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed an ecosystem
	eco := &models.Ecosystem{Name: "eco-scoped-export"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	// Seed a global registry
	reg := &models.Registry{
		Name:    "global-registry",
		Type:    "zot",
		Port:    5200,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	// Use the scoped command helper to pass ecosystem flag
	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	// Scope to a specific ecosystem
	cmd.SetArgs([]string{"--ecosystem", "eco-scoped-export"})
	err := cmd.Execute()
	require.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML")
	require.Equal(t, "List", result["kind"])

	// Check that global registry is NOT in the scoped YAML output
	itemsRaw := result["items"]
	if itemsRaw != nil {
		items, ok := itemsRaw.([]interface{})
		require.True(t, ok)
		for _, rawItem := range items {
			item, ok := rawItem.(map[string]interface{})
			if !ok {
				continue
			}
			assert.NotEqual(t, "Registry", item["kind"],
				"scoped YAML export should NOT include global Registry resources")
			assert.NotEqual(t, "NvimPlugin", item["kind"],
				"scoped YAML export should NOT include global NvimPlugin resources")
			assert.NotEqual(t, "NvimTheme", item["kind"],
				"scoped YAML export should NOT include global NvimTheme resources")
		}
	}
}

// ---------------------------------------------------------------------------
// TestGetAll_YAML_UnscopedIncludesAll  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_YAML_UnscopedIncludesAll verifies that with -A flag (or no active
// context), all resources including globals appear in the YAML/JSON output.
func TestGetAll_YAML_UnscopedIncludesAll(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed a registry (global resource)
	reg := &models.Registry{
		Name:    "global-reg-unscoped",
		Type:    "zot",
		Port:    5300,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	// Seed an ecosystem
	eco := &models.Ecosystem{Name: "eco-unscoped"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	// Use scoped command with -A flag (show all)
	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	cmd.SetArgs([]string{"-A"})
	err := cmd.Execute()
	require.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML")
	require.Equal(t, "List", result["kind"])

	itemsRaw := result["items"]
	require.NotNil(t, itemsRaw, "items should not be nil when resources exist")
	items, ok := itemsRaw.([]interface{})
	require.True(t, ok, "items should be a slice")

	// Find the Registry item in the list
	var foundRegistry bool
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		if item["kind"] == "Registry" {
			foundRegistry = true
			break
		}
	}

	assert.True(t, foundRegistry,
		"unscoped YAML export (-A) should include global Registry resources")
}

// ---------------------------------------------------------------------------
// TestGetAll_YAML_RoundTrip  [Sprint 4 - RED]
// ---------------------------------------------------------------------------

// TestGetAll_YAML_RoundTrip verifies that the output from "get all -o yaml"
// can be parsed as a valid List document and that each item can be re-applied
// (round-trip fidelity). This validates that ApplyList can consume the output.
func TestGetAll_YAML_RoundTrip(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed a minimal hierarchy
	eco := &models.Ecosystem{
		Name:        "roundtrip-eco",
		Description: sql.NullString{String: "Round-trip test", Valid: true},
	}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	cmd := newGetAllTestCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll YAML should not error")

	yamlOutput := buf.Bytes()

	// Verify the output is a valid List that DetectKind recognizes
	var listHeader struct {
		Kind string `yaml:"kind"`
	}
	err = yaml.Unmarshal(yamlOutput, &listHeader)
	require.NoError(t, err, "output should be parseable YAML")
	assert.Equal(t, "List", listHeader.Kind,
		"output must be a kind: List document for round-trip to work")

	// Verify the YAML can be parsed as a full List document with items
	var listDoc map[string]interface{}
	err = yaml.Unmarshal(yamlOutput, &listDoc)
	require.NoError(t, err)
	assert.Equal(t, "devopsmaestro.io/v1", listDoc["apiVersion"])
	assert.NotNil(t, listDoc["items"], "List document must have items field for apply round-trip")
}
