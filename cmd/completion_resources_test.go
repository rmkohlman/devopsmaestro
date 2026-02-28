package cmd

import (
	"context"
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestDataStore creates an in-memory SQLite database with schema for testing
func createTestDataStore(t *testing.T) db.DataStore {
	t.Helper()

	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	require.NoError(t, err)

	err = driver.Connect()
	require.NoError(t, err)

	// Create schema
	err = createTestSchema(driver)
	require.NoError(t, err)

	return db.NewSQLDataStore(driver, nil)
}

// createTestSchema creates the required tables for testing
func createTestSchema(driver db.Driver) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			theme TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			theme TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(ecosystem_id, name),
			FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id)
		)`,
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			description TEXT,
			theme TEXT,
			language TEXT,
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (domain_id) REFERENCES domains(id),
			UNIQUE(domain_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT,
			image_name TEXT,
			container_id TEXT,
			status TEXT DEFAULT 'stopped',
			ssh_agent_forwarding BOOLEAN NOT NULL DEFAULT 0,
			nvim_structure TEXT,
			nvim_plugins TEXT,
			theme TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (app_id) REFERENCES apps(id),
			UNIQUE(app_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS context (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			active_ecosystem_id INTEGER,
			active_domain_id INTEGER,
			active_app_id INTEGER,
			active_workspace_id INTEGER,
			active_project_id INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (active_ecosystem_id) REFERENCES ecosystems(id),
			FOREIGN KEY (active_domain_id) REFERENCES domains(id),
			FOREIGN KEY (active_app_id) REFERENCES apps(id),
			FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id)
		)`,
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
	}

	for _, query := range queries {
		_, err := driver.Execute(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestCompleteResources(t *testing.T) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	// Create in-memory database for testing
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	// Create test data
	ecosystem := &models.Ecosystem{
		Name:        "test-ecosystem",
		Description: sql.NullString{String: "Test ecosystem description", Valid: true},
	}
	err := dataStore.CreateEcosystem(ecosystem)
	require.NoError(t, err)
	t.Logf("Created ecosystem with ID: %d", ecosystem.ID)

	domain := &models.Domain{
		Name:        "test-domain",
		EcosystemID: ecosystem.ID,
		Description: sql.NullString{String: "Test domain description", Valid: true},
	}
	err = dataStore.CreateDomain(domain)
	require.NoError(t, err)
	t.Logf("Created domain with ID: %d", domain.ID)

	app := &models.App{
		Name:        "test-app",
		Path:        "/path/to/test-app",
		DomainID:    domain.ID,
		Description: sql.NullString{String: "Test app description", Valid: true},
	}
	err = dataStore.CreateApp(app)
	require.NoError(t, err)
	t.Logf("Created app with ID: %d", app.ID)

	workspace := &models.Workspace{
		Name:        "test-workspace",
		AppID:       app.ID,
		Description: sql.NullString{String: "Test workspace description", Valid: true},
	}
	err = dataStore.CreateWorkspace(workspace)
	require.NoError(t, err)
	t.Logf("Created workspace with ID: %d", workspace.ID)

	// Create a test command with datastore in context
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), "dataStore", dataStore)
	cmd.SetContext(ctx)

	tests := []struct {
		name          string
		resourceKind  string
		expectedNames []string
		expectDesc    bool
	}{
		{
			name:          "complete ecosystems",
			resourceKind:  "Ecosystem",
			expectedNames: []string{"test-ecosystem"},
			expectDesc:    true,
		},
		{
			name:          "complete domains",
			resourceKind:  "Domain",
			expectedNames: []string{"test-domain"},
			expectDesc:    true,
		},
		{
			name:          "complete apps",
			resourceKind:  "App",
			expectedNames: []string{"test-app"},
			expectDesc:    true,
		},
		{
			name:          "complete workspaces",
			resourceKind:  "Workspace",
			expectedNames: []string{"test-workspace"},
			expectDesc:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, directive := completeResources(cmd, tt.resourceKind)

			// Should use NoFileComp directive
			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

			// Should have expected number of completions
			assert.Len(t, completions, len(tt.expectedNames))

			if len(completions) > 0 && tt.expectDesc {
				// Should include descriptions (format: "name\tdescription")
				assert.Contains(t, completions[0], "\t")
				assert.Contains(t, completions[0], tt.expectedNames[0])
				assert.Contains(t, completions[0], "description")
			}
		})
	}
}

func TestCompleteResourcesErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		kind     string
	}{
		{
			name: "no datastore in context",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.SetContext(context.Background())
				return cmd
			},
			kind: "App",
		},
		{
			name: "wrong datastore type in context",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				ctx := context.WithValue(context.Background(), "dataStore", "not-a-datastore")
				cmd.SetContext(ctx)
				return cmd
			},
			kind: "App",
		},
		{
			name: "invalid resource kind",
			setupCmd: func() *cobra.Command {
				dataStore := createTestDataStore(t)
				cmd := &cobra.Command{Use: "test"}
				ctx := context.WithValue(context.Background(), "dataStore", dataStore)
				cmd.SetContext(ctx)
				return cmd
			},
			kind: "InvalidKind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			completions, directive := completeResources(cmd, tt.kind)

			// Should return empty completions and default directive on error
			assert.Empty(t, completions)
			assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
		})
	}
}

func TestExtractResourceDescription(t *testing.T) {
	tests := []struct {
		name         string
		resource     resource.Resource
		expectedDesc string
	}{
		{
			name: "ecosystem with description",
			resource: handlers.NewEcosystemResource(&models.Ecosystem{
				Description: sql.NullString{String: "Ecosystem description", Valid: true},
			}),
			expectedDesc: "Ecosystem description",
		},
		{
			name: "ecosystem without description",
			resource: handlers.NewEcosystemResource(&models.Ecosystem{
				Description: sql.NullString{},
			}),
			expectedDesc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := extractResourceDescription(tt.resource)
			assert.Equal(t, tt.expectedDesc, desc)
		})
	}
}

func TestCompletionFunctions(t *testing.T) {
	// Create in-memory database for testing
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	// Create a test command with datastore in context
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), "dataStore", dataStore)
	cmd.SetContext(ctx)

	tests := []struct {
		name     string
		function func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	}{
		{
			name:     "completeEcosystems",
			function: completeEcosystems,
		},
		{
			name:     "completeDomains",
			function: completeDomains,
		},
		{
			name:     "completeApps",
			function: completeApps,
		},
		{
			name:     "completeWorkspaces",
			function: completeWorkspaces,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, directive := tt.function(cmd, []string{}, "")

			// Should not error and should use NoFileComp directive
			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

			// Empty database should return empty completions but not error
			assert.NotNil(t, completions)
		})
	}
}

func TestRegisterHierarchyFlagCompletions(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Add flags that completion functions can be registered for
	cmd.Flags().String("ecosystem", "", "ecosystem flag")
	cmd.Flags().String("domain", "", "domain flag")
	cmd.Flags().String("app", "", "app flag")
	cmd.Flags().String("workspace", "", "workspace flag")

	// Should not panic when registering completions
	assert.NotPanics(t, func() {
		registerHierarchyFlagCompletions(cmd)
	})
}
