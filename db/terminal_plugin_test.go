package db

import (
	"database/sql"
	"testing"
	"time"

	"devopsmaestro/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLDataStore_CreateTerminalPlugin(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create terminal plugin
	plugin := &models.TerminalPluginDB{
		Name:         "oh-my-zsh",
		Description:  sql.NullString{String: "Framework for zsh configuration", Valid: true},
		Repo:         "ohmyzsh/ohmyzsh",
		Category:     sql.NullString{String: "framework", Valid: true},
		Shell:        "zsh",
		Manager:      "manual",
		LoadCommand:  sql.NullString{String: "source ~/.oh-my-zsh/oh-my-zsh.sh", Valid: true},
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	err := store.CreateTerminalPlugin(plugin)
	assert.NoError(t, err)
	assert.NotZero(t, plugin.ID)
	assert.False(t, plugin.CreatedAt.IsZero())
	assert.False(t, plugin.UpdatedAt.IsZero())
}

func TestSQLDataStore_GetTerminalPlugin(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create terminal plugin
	plugin := &models.TerminalPluginDB{
		Name:         "powerlevel10k",
		Description:  sql.NullString{String: "Fast and flexible zsh theme", Valid: true},
		Repo:         "romkatv/powerlevel10k",
		Category:     sql.NullString{String: "theme", Valid: true},
		Shell:        "zsh",
		Manager:      "oh-my-zsh",
		Dependencies: `["zsh"]`,
		EnvVars:      `{"POWERLEVEL9K_MODE": "nerdfont-complete"}`,
		Labels:       `{"type": "theme", "performance": "fast"}`,
		Enabled:      true,
	}

	err := store.CreateTerminalPlugin(plugin)
	require.NoError(t, err)

	// Get terminal plugin
	retrieved, err := store.GetTerminalPlugin("powerlevel10k")
	assert.NoError(t, err)
	assert.Equal(t, plugin.Name, retrieved.Name)
	assert.Equal(t, plugin.Description, retrieved.Description)
	assert.Equal(t, plugin.Repo, retrieved.Repo)
	assert.Equal(t, plugin.Category, retrieved.Category)
	assert.Equal(t, plugin.Shell, retrieved.Shell)
	assert.Equal(t, plugin.Manager, retrieved.Manager)
	assert.Equal(t, plugin.Dependencies, retrieved.Dependencies)
	assert.Equal(t, plugin.EnvVars, retrieved.EnvVars)
	assert.Equal(t, plugin.Labels, retrieved.Labels)
	assert.Equal(t, plugin.Enabled, retrieved.Enabled)
}

func TestSQLDataStore_GetTerminalPlugin_NotFound(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Get non-existent terminal plugin
	_, err := store.GetTerminalPlugin("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terminal plugin not found")
}

func TestSQLDataStore_UpdateTerminalPlugin(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create terminal plugin
	plugin := &models.TerminalPluginDB{
		Name:         "zsh-autosuggestions",
		Repo:         "zsh-users/zsh-autosuggestions",
		Shell:        "zsh",
		Manager:      "oh-my-zsh",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	err := store.CreateTerminalPlugin(plugin)
	require.NoError(t, err)
	originalCreatedAt := plugin.CreatedAt

	// Update the plugin
	plugin.Description = sql.NullString{String: "Fish-like autosuggestions for zsh", Valid: true}
	plugin.Category = sql.NullString{String: "completion", Valid: true}

	// Wait a bit to see timestamp difference
	time.Sleep(1 * time.Millisecond)

	err = store.UpdateTerminalPlugin(plugin)
	assert.NoError(t, err)

	// Verify update
	updated, err := store.GetTerminalPlugin("zsh-autosuggestions")
	assert.NoError(t, err)
	assert.Equal(t, plugin.Description, updated.Description)
	assert.Equal(t, plugin.Category, updated.Category)
	// Verify timestamps - CreatedAt should be preserved, UpdatedAt should be newer or equal
	assert.True(t, !updated.CreatedAt.IsZero())
	assert.True(t, !updated.UpdatedAt.IsZero())
	assert.True(t, updated.UpdatedAt.After(originalCreatedAt) || updated.UpdatedAt.Equal(originalCreatedAt))
}

func TestSQLDataStore_UpsertTerminalPlugin(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Test insert (plugin doesn't exist)
	plugin := &models.TerminalPluginDB{
		Name:         "zsh-syntax-highlighting",
		Repo:         "zsh-users/zsh-syntax-highlighting",
		Shell:        "zsh",
		Manager:      "manual",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	err := store.UpsertTerminalPlugin(plugin)
	assert.NoError(t, err)
	assert.NotZero(t, plugin.ID)

	// Test update (plugin exists)
	plugin.Description = sql.NullString{String: "Syntax highlighting for zsh", Valid: true}
	originalID := plugin.ID

	err = store.UpsertTerminalPlugin(plugin)
	assert.NoError(t, err)
	assert.Equal(t, originalID, plugin.ID) // ID should be preserved

	// Verify the update
	updated, err := store.GetTerminalPlugin("zsh-syntax-highlighting")
	assert.NoError(t, err)
	assert.Equal(t, plugin.Description, updated.Description)
}

func TestSQLDataStore_DeleteTerminalPlugin(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create terminal plugin
	plugin := &models.TerminalPluginDB{
		Name:         "fzf",
		Repo:         "junegunn/fzf",
		Shell:        "bash",
		Manager:      "manual",
		Dependencies: "[]",
		EnvVars:      "{}",
		Labels:       "{}",
		Enabled:      true,
	}

	err := store.CreateTerminalPlugin(plugin)
	require.NoError(t, err)

	// Verify it exists
	_, err = store.GetTerminalPlugin("fzf")
	assert.NoError(t, err)

	// Delete it
	err = store.DeleteTerminalPlugin("fzf")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = store.GetTerminalPlugin("fzf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terminal plugin not found")
}

func TestSQLDataStore_ListTerminalPlugins(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create multiple plugins
	plugins := []*models.TerminalPluginDB{
		{
			Name:         "plugin1",
			Repo:         "user/plugin1",
			Shell:        "zsh",
			Manager:      "oh-my-zsh",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "plugin2",
			Repo:         "user/plugin2",
			Shell:        "bash",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
	}

	for _, plugin := range plugins {
		err := store.CreateTerminalPlugin(plugin)
		require.NoError(t, err)
	}

	// List all plugins
	listed, err := store.ListTerminalPlugins()
	assert.NoError(t, err)
	assert.Len(t, listed, 2)

	// Verify order (should be by name)
	assert.Equal(t, "plugin1", listed[0].Name)
	assert.Equal(t, "plugin2", listed[1].Name)
}

func TestSQLDataStore_ListTerminalPluginsByCategory(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create plugins with different categories
	plugins := []*models.TerminalPluginDB{
		{
			Name:         "theme1",
			Repo:         "user/theme1",
			Category:     sql.NullString{String: "theme", Valid: true},
			Shell:        "zsh",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "completion1",
			Repo:         "user/completion1",
			Category:     sql.NullString{String: "completion", Valid: true},
			Shell:        "zsh",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "theme2",
			Repo:         "user/theme2",
			Category:     sql.NullString{String: "theme", Valid: true},
			Shell:        "bash",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
	}

	for _, plugin := range plugins {
		err := store.CreateTerminalPlugin(plugin)
		require.NoError(t, err)
	}

	// List theme plugins
	themes, err := store.ListTerminalPluginsByCategory("theme")
	assert.NoError(t, err)
	assert.Len(t, themes, 2)

	// List completion plugins
	completions, err := store.ListTerminalPluginsByCategory("completion")
	assert.NoError(t, err)
	assert.Len(t, completions, 1)
	assert.Equal(t, "completion1", completions[0].Name)
}

func TestSQLDataStore_ListTerminalPluginsByShell(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create plugins for different shells
	plugins := []*models.TerminalPluginDB{
		{
			Name:         "zsh-plugin1",
			Repo:         "user/zsh-plugin1",
			Shell:        "zsh",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "bash-plugin1",
			Repo:         "user/bash-plugin1",
			Shell:        "bash",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "zsh-plugin2",
			Repo:         "user/zsh-plugin2",
			Shell:        "zsh",
			Manager:      "oh-my-zsh",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
	}

	for _, plugin := range plugins {
		err := store.CreateTerminalPlugin(plugin)
		require.NoError(t, err)
	}

	// List zsh plugins
	zshPlugins, err := store.ListTerminalPluginsByShell("zsh")
	assert.NoError(t, err)
	assert.Len(t, zshPlugins, 2)

	// List bash plugins
	bashPlugins, err := store.ListTerminalPluginsByShell("bash")
	assert.NoError(t, err)
	assert.Len(t, bashPlugins, 1)
	assert.Equal(t, "bash-plugin1", bashPlugins[0].Name)
}

func TestSQLDataStore_ListTerminalPluginsByManager(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Create plugins with different managers
	plugins := []*models.TerminalPluginDB{
		{
			Name:         "manual-plugin1",
			Repo:         "user/manual-plugin1",
			Shell:        "zsh",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "omz-plugin1",
			Repo:         "user/omz-plugin1",
			Shell:        "zsh",
			Manager:      "oh-my-zsh",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
		{
			Name:         "manual-plugin2",
			Repo:         "user/manual-plugin2",
			Shell:        "bash",
			Manager:      "manual",
			Dependencies: "[]",
			EnvVars:      "{}",
			Labels:       "{}",
			Enabled:      true,
		},
	}

	for _, plugin := range plugins {
		err := store.CreateTerminalPlugin(plugin)
		require.NoError(t, err)
	}

	// List manual plugins
	manualPlugins, err := store.ListTerminalPluginsByManager("manual")
	assert.NoError(t, err)
	assert.Len(t, manualPlugins, 2)

	// List oh-my-zsh plugins
	omzPlugins, err := store.ListTerminalPluginsByManager("oh-my-zsh")
	assert.NoError(t, err)
	assert.Len(t, omzPlugins, 1)
	assert.Equal(t, "omz-plugin1", omzPlugins[0].Name)
}

func TestIntegration_TerminalPlugin_FullWorkflow(t *testing.T) {
	store := createTestDataStore(t)
	defer store.Close()

	// Test complete workflow
	plugin := &models.TerminalPluginDB{
		Name:         "test-plugin",
		Description:  sql.NullString{String: "A test plugin", Valid: true},
		Repo:         "test/test-plugin",
		Category:     sql.NullString{String: "testing", Valid: true},
		Shell:        "zsh",
		Manager:      "manual",
		LoadCommand:  sql.NullString{String: "source test-plugin.sh", Valid: true},
		Dependencies: `["git", "curl"]`,
		EnvVars:      `{"TEST_VAR": "test_value"}`,
		Labels:       `{"environment": "development", "type": "utility"}`,
		Enabled:      true,
	}

	// Create
	err := store.CreateTerminalPlugin(plugin)
	assert.NoError(t, err)

	// Read
	retrieved, err := store.GetTerminalPlugin("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, plugin.Name, retrieved.Name)

	// Test JSON fields
	deps, err := retrieved.GetDependencies()
	assert.NoError(t, err)
	assert.Equal(t, []string{"git", "curl"}, deps)

	envVars, err := retrieved.GetEnvVars()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"TEST_VAR": "test_value"}, envVars)

	labels, err := retrieved.GetLabels()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"environment": "development", "type": "utility"}, labels)

	// Update
	plugin.Description = sql.NullString{String: "Updated test plugin", Valid: true}
	err = store.UpdateTerminalPlugin(plugin)
	assert.NoError(t, err)

	// Verify update
	updated, err := store.GetTerminalPlugin("test-plugin")
	assert.NoError(t, err)
	assert.Equal(t, "Updated test plugin", updated.Description.String)

	// List by category
	byCategory, err := store.ListTerminalPluginsByCategory("testing")
	assert.NoError(t, err)
	assert.Len(t, byCategory, 1)

	// Delete
	err = store.DeleteTerminalPlugin("test-plugin")
	assert.NoError(t, err)

	// Verify deletion
	_, err = store.GetTerminalPlugin("test-plugin")
	assert.Error(t, err)
}
