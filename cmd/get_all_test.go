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
