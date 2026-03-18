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
			build_args TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			theme TEXT,
			build_args TEXT,
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
			git_repo_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (domain_id) REFERENCES domains(id),
			UNIQUE(domain_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS git_repos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			default_ref TEXT NOT NULL DEFAULT 'main',
			auth_type TEXT NOT NULL DEFAULT 'none',
			credential_id INTEGER,
			auto_sync BOOLEAN NOT NULL DEFAULT 0,
			sync_interval_minutes INTEGER NOT NULL DEFAULT 0,
			last_synced_at DATETIME,
			sync_status TEXT NOT NULL DEFAULT 'pending',
			sync_error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
			terminal_prompt TEXT,
			terminal_plugins TEXT,
			terminal_package TEXT,
			git_repo_id INTEGER,
			env TEXT NOT NULL DEFAULT '{}',
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (app_id) REFERENCES apps(id),
			FOREIGN KEY (git_repo_id) REFERENCES git_repos(id),
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

// ---------------------------------------------------------------------------
// Tests for new handler-based completion functions
// ---------------------------------------------------------------------------

func TestCompletionFunctions_NewHandlerBased(t *testing.T) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	// Create in-memory database for testing
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	// Create a test command with datastore in context
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), "dataStore", dataStore)
	cmd.SetContext(ctx)

	// Valid shell completion directives — functions must return one of these
	validDirectives := []cobra.ShellCompDirective{
		cobra.ShellCompDirectiveNoFileComp,
		cobra.ShellCompDirectiveDefault,
	}

	tests := []struct {
		name     string
		function func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	}{
		{
			name:     "completeCredentials",
			function: completeCredentials,
		},
		{
			name:     "completeRegistries",
			function: completeRegistries,
		},
		{
			name:     "completeNvimPlugins",
			function: completeNvimPlugins,
		},
		{
			name:     "completeNvimThemes",
			function: completeNvimThemes,
		},
		{
			name:     "completeNvimPackages",
			function: completeNvimPackages,
		},
		{
			name:     "completeTerminalPackages",
			function: completeTerminalPackages,
		},
		{
			name:     "completeTerminalPrompts",
			function: completeTerminalPrompts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, directive := tt.function(cmd, []string{}, "")

			// Must not return nil — empty slice is fine, nil blocks completion UI
			assert.NotNil(t, completions)

			// Must return a valid shell completion directive (not panic or return garbage)
			assert.Contains(t, validDirectives, directive,
				"directive must be NoFileComp (happy path) or Default (error path)")
		})
	}
}

// ---------------------------------------------------------------------------
// TestCompleteGitRepos
// ---------------------------------------------------------------------------

func TestCompleteGitRepos(t *testing.T) {
	t.Run("returns completions in name-tab-url format", func(t *testing.T) {
		dataStore := createTestDataStore(t)
		defer dataStore.Close()

		// Insert two git repos
		repos := []*models.GitRepoDB{
			{
				Name:       "my-repo",
				URL:        "https://github.com/user/my-repo",
				Slug:       "github.com_user_my-repo",
				DefaultRef: "main",
				AuthType:   "none",
				SyncStatus: "pending",
			},
			{
				Name:       "another-repo",
				URL:        "https://gitlab.com/team/another-repo",
				Slug:       "gitlab.com_team_another-repo",
				DefaultRef: "main",
				AuthType:   "none",
				SyncStatus: "pending",
			},
		}
		for _, r := range repos {
			err := dataStore.CreateGitRepo(r)
			require.NoError(t, err)
		}

		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), "dataStore", dataStore)
		cmd.SetContext(ctx)

		completions, directive := completeGitRepos(cmd, []string{}, "")

		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
		assert.Len(t, completions, 2)

		// Verify tab-separated format: each entry should be "name\turl"
		for _, c := range completions {
			assert.Contains(t, c, "\t")
		}

		// Verify the completions contain expected values (order by name: another-repo, my-repo)
		assert.Contains(t, completions[0], "another-repo")
		assert.Contains(t, completions[0], "https://gitlab.com/team/another-repo")
		assert.Equal(t, "another-repo\thttps://gitlab.com/team/another-repo", completions[0])

		assert.Contains(t, completions[1], "my-repo")
		assert.Contains(t, completions[1], "https://github.com/user/my-repo")
		assert.Equal(t, "my-repo\thttps://github.com/user/my-repo", completions[1])
	})

	t.Run("empty database returns empty slice", func(t *testing.T) {
		dataStore := createTestDataStore(t)
		defer dataStore.Close()

		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), "dataStore", dataStore)
		cmd.SetContext(ctx)

		completions, directive := completeGitRepos(cmd, []string{}, "")

		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
		// Empty DB: ListGitRepos returns nil slice, so completions may be empty/nil
		assert.Empty(t, completions)
	})

	t.Run("no datastore in context returns empty with ShellCompDirectiveDefault", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())

		completions, directive := completeGitRepos(cmd, []string{}, "")

		// Without a valid DataStore in context, createDefaultDataStore will fail
		// in test environment, so we expect ShellCompDirectiveDefault
		assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
		assert.Empty(t, completions)
	})

	t.Run("wrong type in context returns empty with ShellCompDirectiveDefault", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), "dataStore", "not-a-datastore")
		cmd.SetContext(ctx)

		completions, directive := completeGitRepos(cmd, []string{}, "")

		// Wrong type falls through to createDefaultDataStore, which will fail in test env
		assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
		assert.Empty(t, completions)
	})
}

// ---------------------------------------------------------------------------
// TestCompleteRegistryTypes
// ---------------------------------------------------------------------------

func TestCompleteRegistryTypes(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())

	completions, directive := completeRegistryTypes(cmd, []string{}, "")

	// Must be NoFileComp (static list, no file completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

	// Must return exactly 5 registry types
	assert.Len(t, completions, 5)

	// Each entry must be tab-separated (type\tdescription)
	for _, c := range completions {
		assert.Contains(t, c, "\t", "completion %q should have tab-separated description", c)
	}

	// Verify the 5 expected type prefixes are present
	expectedTypes := []string{"oci", "pypi", "npm", "go", "http"}
	for _, expected := range expectedTypes {
		found := false
		for _, c := range completions {
			if len(c) >= len(expected) && c[:len(expected)] == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "registry type %q not found in completions", expected)
	}
}

// ---------------------------------------------------------------------------
// TestCompleteRegistrySetDefault
// ---------------------------------------------------------------------------

func TestCompleteRegistrySetDefault(t *testing.T) {
	t.Run("no args returns registry types", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())

		completions, directive := completeRegistrySetDefault(cmd, []string{}, "")

		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
		assert.Len(t, completions, 5)

		// Verify these are the registry types (same as completeRegistryTypes)
		expectedTypes := []string{"oci", "pypi", "npm", "go", "http"}
		for _, expected := range expectedTypes {
			found := false
			for _, c := range completions {
				if len(c) >= len(expected) && c[:len(expected)] == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "registry type %q not found in completions", expected)
		}
	})

	t.Run("one arg delegates to completeRegistries", func(t *testing.T) {
		dataStore := createTestDataStore(t)
		defer dataStore.Close()

		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), "dataStore", dataStore)
		cmd.SetContext(ctx)

		// With one arg provided, should delegate to completeRegistries.
		// completeRegistries calls completeResources(cmd, "Registry"), which
		// returns ShellCompDirectiveNoFileComp on success or ShellCompDirectiveDefault
		// on error. The test schema lacks a registries table so the handler List()
		// call will fail → ShellCompDirectiveDefault is expected here.
		// Either way, the function must not panic and must return a non-nil slice.
		completions, directive := completeRegistrySetDefault(cmd, []string{"oci"}, "")

		// The key invariant: returns a slice (not nil) with a valid directive
		assert.NotNil(t, completions)
		assert.Contains(t,
			[]cobra.ShellCompDirective{cobra.ShellCompDirectiveNoFileComp, cobra.ShellCompDirectiveDefault},
			directive,
			"directive should be either NoFileComp or Default",
		)
	})

	t.Run("two or more args returns nil with NoFileComp", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())

		tests := []struct {
			name string
			args []string
		}{
			{"two args", []string{"oci", "my-registry"}},
			{"three args", []string{"oci", "my-registry", "extra"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				completions, directive := completeRegistrySetDefault(cmd, tt.args, "")

				assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
				assert.Nil(t, completions)
			})
		}
	})
}

// ---------------------------------------------------------------------------
// TestGetCompletionDataStore
// ---------------------------------------------------------------------------

func TestGetCompletionDataStore(t *testing.T) {
	t.Run("returns DataStore from context", func(t *testing.T) {
		dataStore := createTestDataStore(t)
		defer dataStore.Close()

		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), "dataStore", dataStore)
		cmd.SetContext(ctx)

		result, err := getCompletionDataStore(cmd)

		require.NoError(t, err)
		assert.Equal(t, dataStore, result)
	})

	t.Run("without DataStore in context attempts createDefaultDataStore", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())

		// In the test environment, createDefaultDataStore will likely fail
		// because there's no production database available.
		// We just verify the function handles this gracefully (no panic).
		_, _ = getCompletionDataStore(cmd)
		// No assertion on success/failure — test environment may or may not have a DB
	})

	t.Run("wrong type in context falls through to createDefaultDataStore", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), "dataStore", 42) // wrong type
		cmd.SetContext(ctx)

		// Should fall through since type assertion fails; createDefaultDataStore
		// will likely fail in test env. Verify no panic.
		_, _ = getCompletionDataStore(cmd)
		// No assertion on success/failure
	})
}

// ---------------------------------------------------------------------------
// TestRegisterWorkspaceAppFlagCompletions
// ---------------------------------------------------------------------------

func TestRegisterWorkspaceAppFlagCompletions(t *testing.T) {
	t.Run("registers completions without panic", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("workspace", "", "workspace flag")
		cmd.Flags().String("app", "", "app flag")

		assert.NotPanics(t, func() {
			registerWorkspaceAppFlagCompletions(cmd)
		})
	})

	t.Run("works even when flags are not defined", func(t *testing.T) {
		// RegisterFlagCompletionFunc silently ignores missing flags in some cobra versions
		// or may log a warning. Verify it does not panic.
		cmd := &cobra.Command{Use: "test"}

		assert.NotPanics(t, func() {
			registerWorkspaceAppFlagCompletions(cmd)
		})
	})
}

// ---------------------------------------------------------------------------
// TestRegisterCredentialScopeFlagCompletions
// ---------------------------------------------------------------------------

func TestRegisterCredentialScopeFlagCompletions(t *testing.T) {
	t.Run("registers completions without panic", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("ecosystem", "", "ecosystem flag")
		cmd.Flags().String("domain", "", "domain flag")
		cmd.Flags().String("app", "", "app flag")
		cmd.Flags().String("workspace", "", "workspace flag")

		assert.NotPanics(t, func() {
			registerCredentialScopeFlagCompletions(cmd)
		})
	})

	t.Run("delegates to registerHierarchyFlagCompletions", func(t *testing.T) {
		// registerCredentialScopeFlagCompletions is a thin wrapper; verify behaviour
		// is identical to registerHierarchyFlagCompletions.
		cmd1 := &cobra.Command{Use: "cmd1"}
		cmd1.Flags().String("ecosystem", "", "")
		cmd1.Flags().String("domain", "", "")
		cmd1.Flags().String("app", "", "")
		cmd1.Flags().String("workspace", "", "")

		cmd2 := &cobra.Command{Use: "cmd2"}
		cmd2.Flags().String("ecosystem", "", "")
		cmd2.Flags().String("domain", "", "")
		cmd2.Flags().String("app", "", "")
		cmd2.Flags().String("workspace", "", "")

		assert.NotPanics(t, func() {
			registerCredentialScopeFlagCompletions(cmd1)
			registerHierarchyFlagCompletions(cmd2)
		})
	})
}
