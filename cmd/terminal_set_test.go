//go:build integration
// +build integration

package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestSetTerminalCommandStructure verifies the 'set terminal' command hierarchy exists
func TestSetTerminalCommandStructure(t *testing.T) {
	t.Run("set terminal command exists", func(t *testing.T) {
		// Verify setCmd exists (from nvim_set.go)
		assert.NotNil(t, setCmd, "setCmd should exist")

		// Verify setTerminalCmd exists as a subcommand
		subcommands := setCmd.Commands()
		var terminalCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "terminal" {
				terminalCmd = cmd
				break
			}
		}
		require.NotNil(t, terminalCmd, "setTerminalCmd should exist under setCmd")
		assert.Equal(t, "terminal", terminalCmd.Use, "setTerminalCmd should have Use 'terminal'")
	})

	t.Run("set terminal has prompt subcommand", func(t *testing.T) {
		var terminalCmd *cobra.Command
		for _, cmd := range setCmd.Commands() {
			if cmd.Name() == "terminal" {
				terminalCmd = cmd
				break
			}
		}
		require.NotNil(t, terminalCmd)

		subcommands := terminalCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "prompt", "should have 'prompt' subcommand")
	})

	t.Run("set terminal has plugin subcommand", func(t *testing.T) {
		var terminalCmd *cobra.Command
		for _, cmd := range setCmd.Commands() {
			if cmd.Name() == "terminal" {
				terminalCmd = cmd
				break
			}
		}
		require.NotNil(t, terminalCmd)

		subcommands := terminalCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "plugin", "should have 'plugin' subcommand")
	})

	t.Run("set terminal has package subcommand", func(t *testing.T) {
		var terminalCmd *cobra.Command
		for _, cmd := range setCmd.Commands() {
			if cmd.Name() == "terminal" {
				terminalCmd = cmd
				break
			}
		}
		require.NotNil(t, terminalCmd)

		subcommands := terminalCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "package", "should have 'package' subcommand")
	})
}

// TestSetTerminalPrompt tests the 'dvm set terminal prompt' command
func TestSetTerminalPrompt(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("set prompt for workspace", func(t *testing.T) {
		// Create test app and workspace
		app, ws := createTestAppAndWorkspace(t, "test-app", "dev")

		// Execute command: dvm set terminal prompt -w dev starship
		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship"})

		err := cmd.Execute()
		assert.NoError(t, err, "setting prompt should succeed")

		// Verify workspace was updated
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPrompt field needs to be added to Workspace model
		// This test will fail until implementation adds the field
		assert.NotNil(t, updated.TerminalPrompt, "workspace should have terminal prompt set")
		assert.Equal(t, "starship", updated.TerminalPrompt.String)
	})

	t.Run("change prompt for workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "test-app-2", "dev")

		// Set initial prompt
		cmd := getRootCommand()
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship"})
		err := cmd.Execute()
		require.NoError(t, err)

		// Change to different prompt
		cmd = getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship-minimal"})

		err = cmd.Execute()
		assert.NoError(t, err, "changing prompt should succeed")

		// Verify change
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPrompt field needs to be added to Workspace model
		assert.Equal(t, "starship-minimal", updated.TerminalPrompt.String)
	})

	t.Run("clear prompt from workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "test-app-3", "dev")

		// Set prompt first
		cmd := getRootCommand()
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship"})
		err := cmd.Execute()
		require.NoError(t, err)

		// Clear prompt
		cmd = getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "--clear"})

		err = cmd.Execute()
		assert.NoError(t, err, "clearing prompt should succeed")

		// Verify cleared
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPrompt field needs to be added to Workspace model
		assert.False(t, updated.TerminalPrompt.Valid, "workspace should have no prompt after clear")
	})

	t.Run("error without workspace flag", func(t *testing.T) {
		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "starship"})

		err := cmd.Execute()
		assert.Error(t, err, "should require workspace flag")
		assert.Contains(t, err.Error(), "workspace", "error should mention workspace")
	})

	t.Run("error with unknown prompt", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "test-app-4", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "unknown-prompt"})

		err := cmd.Execute()
		assert.Error(t, err, "should fail for unknown prompt")
		assert.Contains(t, err.Error(), "not found", "error should indicate prompt not found")
	})

	t.Run("error without prompt name", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "test-app-5", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name})

		err := cmd.Execute()
		assert.Error(t, err, "should require prompt name or --clear")
	})

	t.Run("dry-run shows preview", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "test-app-6", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship", "--dry-run"})

		err := cmd.Execute()
		assert.NoError(t, err, "dry-run should succeed")

		// Verify workspace was NOT updated
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPrompt field needs to be added to Workspace model
		assert.False(t, updated.TerminalPrompt.Valid, "workspace should not be updated during dry-run")
	})

	t.Run("with explicit app flag", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "test-app-7", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-a", app.Name, "-w", ws.Name, "starship"})

		err := cmd.Execute()
		assert.NoError(t, err, "should work with explicit app flag")

		// Verify
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPrompt field needs to be added to Workspace model
		assert.Equal(t, "starship", updated.TerminalPrompt.String)
	})
}

// TestSetTerminalPlugin tests the 'dvm set terminal plugin' command
func TestSetTerminalPlugin(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("add single plugin to workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "plugin-app-1", "dev")
		createTestTerminalPlugin(t, "zsh-autosuggestions")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "zsh-autosuggestions"})

		err := cmd.Execute()
		assert.NoError(t, err, "adding plugin should succeed")

		// Verify plugin was added
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: GetTerminalPlugins method needs to be added to Workspace model
		plugins := updated.GetTerminalPlugins()
		assert.Contains(t, plugins, "zsh-autosuggestions")
	})

	t.Run("add multiple plugins to workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "plugin-app-2", "dev")
		createTestTerminalPlugin(t, "zsh-autosuggestions")
		createTestTerminalPlugin(t, "zsh-syntax-highlighting")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "zsh-autosuggestions", "zsh-syntax-highlighting"})

		err := cmd.Execute()
		assert.NoError(t, err, "adding multiple plugins should succeed")

		// Verify both plugins were added
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: GetTerminalPlugins method needs to be added to Workspace model
		plugins := updated.GetTerminalPlugins()
		assert.Contains(t, plugins, "zsh-autosuggestions")
		assert.Contains(t, plugins, "zsh-syntax-highlighting")
	})

	t.Run("add all plugins with --all flag", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "plugin-app-3", "dev")
		createTestTerminalPlugin(t, "plugin-a")
		createTestTerminalPlugin(t, "plugin-b")
		createTestTerminalPlugin(t, "plugin-c")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "--all"})

		err := cmd.Execute()
		assert.NoError(t, err, "adding all plugins should succeed")

		// Verify all plugins were added
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: GetTerminalPlugins method needs to be added to Workspace model
		plugins := updated.GetTerminalPlugins()
		assert.GreaterOrEqual(t, len(plugins), 3, "should have at least 3 plugins")
	})

	t.Run("clear plugins from workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "plugin-app-4", "dev")
		createTestTerminalPlugin(t, "test-plugin")

		// Add plugin first
		cmd := getRootCommand()
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "test-plugin"})
		err := cmd.Execute()
		require.NoError(t, err)

		// Clear plugins
		cmd = getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "--clear"})

		err = cmd.Execute()
		assert.NoError(t, err, "clearing plugins should succeed")

		// Verify cleared
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: GetTerminalPlugins method needs to be added to Workspace model
		plugins := updated.GetTerminalPlugins()
		assert.Empty(t, plugins, "workspace should have no plugins after clear")
	})

	t.Run("error without workspace flag", func(t *testing.T) {
		createTestTerminalPlugin(t, "test-plugin-2")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "test-plugin-2"})

		err := cmd.Execute()
		assert.Error(t, err, "should require workspace flag")
		assert.Contains(t, err.Error(), "workspace", "error should mention workspace")
	})

	t.Run("error without plugin names or flags", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "plugin-app-5", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name})

		err := cmd.Execute()
		assert.Error(t, err, "should require plugin names, --all, or --clear")
	})

	t.Run("error with unknown plugin", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "plugin-app-6", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "unknown-plugin"})

		err := cmd.Execute()
		assert.Error(t, err, "should fail for unknown plugin")
	})

	t.Run("skip already configured plugins", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "plugin-app-7", "dev")
		createTestTerminalPlugin(t, "duplicate-plugin")

		// Add plugin first time
		cmd := getRootCommand()
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "duplicate-plugin"})
		err := cmd.Execute()
		require.NoError(t, err)

		// Try to add same plugin again
		cmd = getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "duplicate-plugin"})

		err = cmd.Execute()
		assert.NoError(t, err, "should skip already configured plugin without error")

		// Verify only one instance
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: GetTerminalPlugins method needs to be added to Workspace model
		plugins := updated.GetTerminalPlugins()
		count := 0
		for _, p := range plugins {
			if p == "duplicate-plugin" {
				count++
			}
		}
		assert.Equal(t, 1, count, "should have exactly one instance of the plugin")
	})

	t.Run("dry-run shows preview", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "plugin-app-8", "dev")
		createTestTerminalPlugin(t, "preview-plugin")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "preview-plugin", "--dry-run"})

		err := cmd.Execute()
		assert.NoError(t, err, "dry-run should succeed")

		// Verify workspace was NOT updated
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: GetTerminalPlugins method needs to be added to Workspace model
		plugins := updated.GetTerminalPlugins()
		assert.NotContains(t, plugins, "preview-plugin", "workspace should not be updated during dry-run")
	})
}

// TestSetTerminalPackage tests the 'dvm set terminal package' command
func TestSetTerminalPackage(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("set package for workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "pkg-app-1", "dev")
		createTestTerminalPackage(t, "poweruser")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "package", "-w", ws.Name, "poweruser"})

		err := cmd.Execute()
		assert.NoError(t, err, "setting package should succeed")

		// Verify package was set
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPackage field needs to be added to Workspace model
		assert.NotNil(t, updated.TerminalPackage, "workspace should have terminal package set")
		assert.Equal(t, "poweruser", updated.TerminalPackage.String)
	})

	t.Run("clear package from workspace", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "pkg-app-2", "dev")
		createTestTerminalPackage(t, "test-package")

		// Set package first
		cmd := getRootCommand()
		cmd.SetArgs([]string{"set", "terminal", "package", "-w", ws.Name, "test-package"})
		err := cmd.Execute()
		require.NoError(t, err)

		// Clear package
		cmd = getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "package", "-w", ws.Name, "--clear"})

		err = cmd.Execute()
		assert.NoError(t, err, "clearing package should succeed")

		// Verify cleared
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPackage field needs to be added to Workspace model
		assert.False(t, updated.TerminalPackage.Valid, "workspace should have no package after clear")
	})

	t.Run("error without workspace flag", func(t *testing.T) {
		createTestTerminalPackage(t, "test-package-2")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "package", "test-package-2"})

		err := cmd.Execute()
		assert.Error(t, err, "should require workspace flag")
		assert.Contains(t, err.Error(), "workspace", "error should mention workspace")
	})

	t.Run("error with unknown package", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "pkg-app-3", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "package", "-w", ws.Name, "unknown-package"})

		err := cmd.Execute()
		assert.Error(t, err, "should fail for unknown package")
		assert.Contains(t, err.Error(), "not found", "error should indicate package not found")
	})

	t.Run("error without package name", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "pkg-app-4", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "package", "-w", ws.Name})

		err := cmd.Execute()
		assert.Error(t, err, "should require package name or --clear")
	})

	t.Run("dry-run shows preview", func(t *testing.T) {
		app, ws := createTestAppAndWorkspace(t, "pkg-app-5", "dev")
		createTestTerminalPackage(t, "preview-package")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "package", "-w", ws.Name, "preview-package", "--dry-run"})

		err := cmd.Execute()
		assert.NoError(t, err, "dry-run should succeed")

		// Verify workspace was NOT updated
		ds := getTestDataStore(t)
		updated, err := ds.GetWorkspaceByName(app.ID, ws.Name)
		require.NoError(t, err)
		// NOTE: TerminalPackage field needs to be added to Workspace model
		assert.False(t, updated.TerminalPackage.Valid, "workspace should not be updated during dry-run")
	})
}

// TestSetTerminalOutputFormats tests output format options
func TestSetTerminalOutputFormats(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("yaml output format for prompt", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "output-app-1", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship", "-o", "yaml"})

		err := cmd.Execute()
		assert.NoError(t, err, "should succeed with yaml output")

		// Verify output is valid YAML
		var result map[string]interface{}
		err = yaml.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err, "output should be valid YAML")
	})

	t.Run("json output format for prompt", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "output-app-2", "dev")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "prompt", "-w", ws.Name, "starship", "-o", "json"})

		err := cmd.Execute()
		assert.NoError(t, err, "should succeed with json output")

		// Verify output is valid JSON
		var result map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err, "output should be valid JSON")
	})

	t.Run("yaml output format for plugin", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "output-app-3", "dev")
		createTestTerminalPlugin(t, "output-plugin")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "output-plugin", "-o", "yaml"})

		err := cmd.Execute()
		assert.NoError(t, err, "should succeed with yaml output")

		// Verify output is valid YAML
		var result map[string]interface{}
		err = yaml.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err, "output should be valid YAML")
	})

	t.Run("json output format for plugin", func(t *testing.T) {
		_, ws := createTestAppAndWorkspace(t, "output-app-4", "dev")
		createTestTerminalPlugin(t, "output-plugin-2")

		cmd := getRootCommand()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"set", "terminal", "plugin", "-w", ws.Name, "output-plugin-2", "-o", "json"})

		err := cmd.Execute()
		assert.NoError(t, err, "should succeed with json output")

		// Verify output is valid JSON
		var result map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err, "output should be valid JSON")
	})
}

// Helper functions for test setup

func setupTestEnvironment(t *testing.T) {
	// Initialize test database
	// This is a no-op now, as each test creates its own datastore
}

func getTestDataStore(t *testing.T) db.DataStore {
	t.Helper()

	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	require.NoError(t, err)

	err = driver.Connect()
	require.NoError(t, err)

	// Create minimal schema for workspaces
	err = createMinimalTestSchema(driver)
	require.NoError(t, err)

	return db.NewSQLDataStore(driver, nil)
}

func createMinimalTestSchema(driver db.Driver) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			theme TEXT,
			nvim_package TEXT,
			terminal_package TEXT,
			build_args TEXT,
			ca_certs TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			theme TEXT,
			nvim_package TEXT,
			terminal_package TEXT,
			build_args TEXT,
			ca_certs TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE,
			UNIQUE(ecosystem_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER,
			system_id INTEGER,
			name TEXT NOT NULL,
			path TEXT,
			description TEXT,
			language TEXT,
			build_config TEXT,
			theme TEXT,
			nvim_package TEXT,
			terminal_package TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id INTEGER NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT,
			image_name TEXT NOT NULL,
			container_id TEXT,
			status TEXT NOT NULL DEFAULT 'stopped',
			ssh_agent_forwarding INTEGER DEFAULT 0,
			nvim_structure TEXT,
			nvim_plugins TEXT,
			theme TEXT,
			terminal_prompt TEXT,
			terminal_plugins TEXT,
			terminal_package TEXT,
			nvim_package TEXT,
			env TEXT NOT NULL DEFAULT '{}',
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(app_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS terminal_plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			repo TEXT NOT NULL,
			shell TEXT NOT NULL,
			manager TEXT NOT NULL,
			enabled INTEGER DEFAULT 1,
			dependencies TEXT,
			env_vars TEXT,
			labels TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS terminal_packages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			plugins TEXT,
			prompts TEXT,
			profiles TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		_, err := driver.Execute(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRootCommand() *cobra.Command {
	// Reset and return a fresh root command for each test
	// This prevents flag pollution between tests
	cmd := &cobra.Command{Use: "dvm"}
	// Re-initialize all commands
	return cmd
}

func createTestAppAndWorkspace(t *testing.T, appName, wsName string) (*models.App, *models.Workspace) {
	ds := getTestDataStore(t)

	app := &models.App{
		Name:        appName,
		Description: sql.NullString{String: "Test app", Valid: true},
	}
	err := ds.CreateApp(app)
	require.NoError(t, err)

	ws := &models.Workspace{
		Name:      wsName,
		AppID:     app.ID,
		ImageName: "test-image",
		Status:    "created",
		Slug:      appName + "-" + wsName,
	}
	err = ds.CreateWorkspace(ws)
	require.NoError(t, err)

	return app, ws
}

func createTestTerminalPlugin(t *testing.T, name string) *models.TerminalPluginDB {
	ds := getTestDataStore(t)

	plugin := &models.TerminalPluginDB{
		Name:         name,
		Repo:         "https://github.com/example/" + name,
		Shell:        "zsh",
		Manager:      "manual",
		Enabled:      true,
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
	}

	err := ds.CreateTerminalPlugin(plugin)
	require.NoError(t, err)

	return plugin
}

func createTestTerminalPackage(t *testing.T, name string) *models.TerminalPackageDB {
	ds := getTestDataStore(t)

	pkg := &models.TerminalPackageDB{
		Name:     name,
		Plugins:  "[]",
		Prompts:  "[]",
		Profiles: "[]",
	}

	err := ds.CreateTerminalPackage(pkg)
	require.NoError(t, err)

	return pkg
}
