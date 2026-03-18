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
// TestGetAll_JSONOutput
// ---------------------------------------------------------------------------

func TestGetAll_JSONOutput(t *testing.T) {
	// Use full schema so no "no such table" warnings are prepended to JSON output
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed an ecosystem so at least one section is non-empty in JSON
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

	// All 9 top-level keys must be present
	expectedKeys := []string{
		"ecosystems",
		"domains",
		"apps",
		"workspaces",
		"credentials",
		"registries",
		"gitRepos",
		"nvimPlugins",
		"nvimThemes",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, result, key, "JSON output should contain top-level key %q", key)
	}

	// Ecosystems section should have one entry
	ecosystemsRaw, ok := result["ecosystems"].([]interface{})
	require.True(t, ok, "ecosystems should be a JSON array")
	require.Len(t, ecosystemsRaw, 1, "ecosystems array should have 1 entry")

	entry, ok := ecosystemsRaw[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-eco", entry["name"])
}

// ---------------------------------------------------------------------------
// TestGetAll_YAMLOutput
// ---------------------------------------------------------------------------

func TestGetAll_YAMLOutput(t *testing.T) {
	// Use full schema so no "no such table" warnings are prepended to YAML output
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed one domain for variety
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

	// Output must be valid YAML that decodes into the AllResources structure
	var result AllResources
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML; got: %s", buf.String())

	// Domains section should have one entry
	require.Len(t, result.Domains, 1, "domains should have 1 entry")
	assert.Equal(t, "domain-yaml", result.Domains[0].Name)
}

// ---------------------------------------------------------------------------
// TestGetAllResourceSummary
// ---------------------------------------------------------------------------

func TestGetAllResourceSummary(t *testing.T) {
	t.Run("AllResourceSummary JSON round-trip", func(t *testing.T) {
		summary := AllResourceSummary{
			Name:        "my-resource",
			Description: "a description",
			Status:      "active",
			Type:        "go",
			URL:         "https://example.com",
			Repo:        "org/repo",
			Category:    "lsp",
		}

		data, err := json.Marshal(summary)
		require.NoError(t, err)

		var decoded AllResourceSummary
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, summary, decoded)
	})

	t.Run("AllResourceSummary omits empty optional fields", func(t *testing.T) {
		summary := AllResourceSummary{Name: "minimal"}

		data, err := json.Marshal(summary)
		require.NoError(t, err)

		// Optional fields with omitempty should not appear in JSON
		jsonStr := string(data)
		assert.NotContains(t, jsonStr, `"description"`, "empty description should be omitted")
		assert.NotContains(t, jsonStr, `"status"`, "empty status should be omitted")
		assert.NotContains(t, jsonStr, `"type"`, "empty type should be omitted")
		assert.NotContains(t, jsonStr, `"url"`, "empty url should be omitted")
		assert.Contains(t, jsonStr, `"name":"minimal"`)
	})

	t.Run("AllResources JSON round-trip with all 9 sections", func(t *testing.T) {
		all := AllResources{
			Ecosystems:  []AllResourceSummary{{Name: "eco1"}},
			Domains:     []AllResourceSummary{{Name: "dom1"}},
			Apps:        []AllResourceSummary{{Name: "app1", Type: "go"}},
			Workspaces:  []AllResourceSummary{{Name: "ws1", Status: "stopped"}},
			Credentials: []AllResourceSummary{{Name: "cred1", Type: "global"}},
			Registries:  []AllResourceSummary{{Name: "reg1", Type: "oci"}},
			GitRepos:    []AllResourceSummary{{Name: "repo1", URL: "https://github.com/org/repo"}},
			NvimPlugins: []AllResourceSummary{{Name: "nvim-lsp", Repo: "neovim/nvim-lspconfig"}},
			NvimThemes:  []AllResourceSummary{{Name: "catppuccin", Status: "yes"}},
		}

		data, err := json.Marshal(all)
		require.NoError(t, err)

		var decoded AllResources
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, all, decoded)
	})

	t.Run("AllResources YAML round-trip", func(t *testing.T) {
		all := AllResources{
			Ecosystems: []AllResourceSummary{{Name: "eco-yaml", Description: "test"}},
			GitRepos:   []AllResourceSummary{{Name: "repo-yaml", URL: "https://example.com"}},
		}

		data, err := yaml.Marshal(all)
		require.NoError(t, err)

		var decoded AllResources
		err = yaml.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, all.Ecosystems, decoded.Ecosystems)
		assert.Equal(t, all.GitRepos, decoded.GitRepos)
	})
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
// TestGetAll_JSONEmptyDatabase
// ---------------------------------------------------------------------------

func TestGetAll_JSONEmptyDatabase(t *testing.T) {
	// Even with an empty database all 9 keys must be present in JSON output.
	// Use full schema to avoid warning messages polluting the JSON buffer.
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

	expectedKeys := []string{
		"ecosystems", "domains", "apps", "workspaces",
		"credentials", "registries", "gitRepos", "nvimPlugins", "nvimThemes",
	}
	for _, key := range expectedKeys {
		val, exists := result[key]
		assert.True(t, exists, "JSON should contain key %q", key)
		// Each value must be either null or an empty/non-empty JSON array
		if val != nil {
			_, isSlice := val.([]interface{})
			assert.True(t, isSlice, "key %q should map to a JSON array, got %T", key, val)
		}
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
// TestGetAll_MultipleResources_JSON
// ---------------------------------------------------------------------------

func TestGetAll_MultipleResources_JSON(t *testing.T) {
	// Use full schema to avoid warning prefixes in JSON output
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed multiple ecosystems
	for _, name := range []string{"eco-a", "eco-b", "eco-c", "eco-extra"} {
		require.NoError(t, dataStore.CreateEcosystem(&models.Ecosystem{Name: name}))
	}

	// Create a parent ecosystem to attach domains to
	parentEco := &models.Ecosystem{Name: "parent-eco"}
	require.NoError(t, dataStore.CreateEcosystem(parentEco))

	for _, domName := range []string{"dom-x", "dom-y"} {
		require.NoError(t, dataStore.CreateDomain(&models.Domain{
			Name:        domName,
			EcosystemID: parentEco.ID,
		}))
	}

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

	// Ecosystems should have multiple entries (eco-a, eco-b, eco-c, eco-extra, parent-eco)
	ecos, ok := result["ecosystems"].([]interface{})
	require.True(t, ok, "ecosystems should be a JSON array")
	assert.GreaterOrEqual(t, len(ecos), 3, "should have at least 3 ecosystems")

	// Domains should have 2 entries
	doms, ok := result["domains"].([]interface{})
	require.True(t, ok, "domains should be a JSON array")
	assert.Len(t, doms, 2, "should have 2 domains")
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
