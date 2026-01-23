package db

import (
	"database/sql"
	"devopsmaestro/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPluginTestDB(t *testing.T) (*InMemorySQLiteDB, *SQLDataStore, func()) {
	db, err := NewInMemoryTestDB()
	require.NoError(t, err)

	// Get the underlying SQLite connection
	sqliteDB := db.(*InMemorySQLiteDB)

	// Create the nvim_plugins table
	_, err = sqliteDB.conn.Exec(`
		CREATE TABLE IF NOT EXISTS nvim_plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			repo TEXT NOT NULL,
			branch TEXT,
			version TEXT,
			priority INTEGER DEFAULT 0,
			lazy INTEGER DEFAULT 0,
			enabled INTEGER DEFAULT 1,
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
			description TEXT,
			category TEXT,
			tags TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Create workspace_plugins junction table
	_, err = sqliteDB.conn.Exec(`
		CREATE TABLE IF NOT EXISTS workspace_plugins (
			workspace_id INTEGER NOT NULL,
			plugin_id INTEGER NOT NULL,
			PRIMARY KEY (workspace_id, plugin_id)
		)
	`)
	require.NoError(t, err)

	dataStore := &SQLDataStore{db: db, queryBuilder: &SQLQueryBuilder{}}

	cleanup := func() {
		db.Close()
	}

	return sqliteDB, dataStore, cleanup
}

func TestCreatePlugin(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	plugin := &models.NvimPluginDB{
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Branch:      sql.NullString{String: "0.1.x", Valid: true},
		Description: sql.NullString{String: "Fuzzy finder", Valid: true},
		Category:    sql.NullString{String: "fuzzy-finder", Valid: true},
		Enabled:     true,
		Event:       sql.NullString{String: "VeryLazy", Valid: true},
		Config:      sql.NullString{String: `require("telescope").setup()`, Valid: true},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)
	assert.Greater(t, plugin.ID, 0)
}

func TestGetPluginByName(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create a plugin first
	plugin := &models.NvimPluginDB{
		Name:        "mason",
		Repo:        "williamboman/mason.nvim",
		Description: sql.NullString{String: "LSP installer", Valid: true},
		Category:    sql.NullString{String: "lsp", Valid: true},
		Enabled:     true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := dataStore.GetPluginByName("mason")
	require.NoError(t, err)
	assert.Equal(t, "mason", retrieved.Name)
	assert.Equal(t, "williamboman/mason.nvim", retrieved.Repo)
	assert.True(t, retrieved.Description.Valid)
	assert.Equal(t, "LSP installer", retrieved.Description.String)
}

func TestGetPluginByName_NotFound(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	_, err := dataStore.GetPluginByName("nonexistent")
	assert.Error(t, err)
}

func TestUpdatePlugin(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create a plugin
	plugin := &models.NvimPluginDB{
		Name:      "copilot",
		Repo:      "zbirenbaum/copilot.lua",
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)

	// Update it
	plugin.Description = sql.NullString{String: "GitHub Copilot", Valid: true}
	plugin.Category = sql.NullString{String: "ai", Valid: true}
	plugin.Config = sql.NullString{String: `require("copilot").setup()`, Valid: true}
	plugin.UpdatedAt = time.Now().UTC()

	err = dataStore.UpdatePlugin(plugin)
	require.NoError(t, err)

	// Verify update
	retrieved, err := dataStore.GetPluginByName("copilot")
	require.NoError(t, err)
	assert.Equal(t, "GitHub Copilot", retrieved.Description.String)
	assert.Equal(t, "ai", retrieved.Category.String)
	assert.True(t, retrieved.Config.Valid)
}

func TestListPlugins(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create multiple plugins
	plugins := []*models.NvimPluginDB{
		{
			Name:      "telescope",
			Repo:      "nvim-telescope/telescope.nvim",
			Category:  sql.NullString{String: "fuzzy-finder", Valid: true},
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			Name:      "mason",
			Repo:      "williamboman/mason.nvim",
			Category:  sql.NullString{String: "lsp", Valid: true},
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			Name:      "lspconfig",
			Repo:      "neovim/nvim-lspconfig",
			Category:  sql.NullString{String: "lsp", Valid: true},
			Enabled:   false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	for _, p := range plugins {
		err := dataStore.CreatePlugin(p)
		require.NoError(t, err)
	}

	// List all plugins
	result, err := dataStore.ListPlugins()
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestListPluginsByCategory(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create plugins with different categories
	plugins := []*models.NvimPluginDB{
		{
			Name:      "telescope",
			Repo:      "nvim-telescope/telescope.nvim",
			Category:  sql.NullString{String: "fuzzy-finder", Valid: true},
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			Name:      "mason",
			Repo:      "williamboman/mason.nvim",
			Category:  sql.NullString{String: "lsp", Valid: true},
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			Name:      "lspconfig",
			Repo:      "neovim/nvim-lspconfig",
			Category:  sql.NullString{String: "lsp", Valid: true},
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	for _, p := range plugins {
		err := dataStore.CreatePlugin(p)
		require.NoError(t, err)
	}

	// List by category
	lspPlugins, err := dataStore.ListPluginsByCategory("lsp")
	require.NoError(t, err)
	assert.Len(t, lspPlugins, 2)

	for _, p := range lspPlugins {
		assert.Equal(t, "lsp", p.Category.String)
	}
}

func TestDeletePlugin(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create a plugin
	plugin := &models.NvimPluginDB{
		Name:      "test-plugin",
		Repo:      "test/test",
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)

	// Delete it
	err = dataStore.DeletePlugin("test-plugin")
	require.NoError(t, err)

	// Verify deletion
	_, err = dataStore.GetPluginByName("test-plugin")
	assert.Error(t, err)
}

func TestPluginWithDependencies(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create plugin with dependencies (stored as JSON array)
	plugin := &models.NvimPluginDB{
		Name:         "telescope",
		Repo:         "nvim-telescope/telescope.nvim",
		Dependencies: sql.NullString{String: `["nvim-lua/plenary.nvim","nvim-telescope/telescope-fzf-native.nvim"]`, Valid: true},
		Enabled:      true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := dataStore.GetPluginByName("telescope")
	require.NoError(t, err)
	assert.True(t, retrieved.Dependencies.Valid)
	assert.Contains(t, retrieved.Dependencies.String, "plenary")
}

func TestPluginWithKeymaps(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	// Create plugin with keymaps (stored as JSON)
	plugin := &models.NvimPluginDB{
		Name:      "telescope",
		Repo:      "nvim-telescope/telescope.nvim",
		Keymaps:   sql.NullString{String: `[{"key":"<leader>ff","mode":"n","action":"<cmd>Telescope find_files<cr>"}]`, Valid: true},
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := dataStore.GetPluginByName("telescope")
	require.NoError(t, err)
	assert.True(t, retrieved.Keymaps.Valid)
	assert.Contains(t, retrieved.Keymaps.String, "<leader>ff")
}

func TestPluginWithComplexConfig(t *testing.T) {
	_, dataStore, cleanup := setupPluginTestDB(t)
	defer cleanup()

	multilineConfig := `local telescope = require("telescope")
telescope.setup({
  defaults = {
    mappings = {
      i = {
        ["<C-u>"] = false,
      },
    },
  },
})`

	plugin := &models.NvimPluginDB{
		Name:      "telescope",
		Repo:      "nvim-telescope/telescope.nvim",
		Config:    sql.NullString{String: multilineConfig, Valid: true},
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := dataStore.CreatePlugin(plugin)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := dataStore.GetPluginByName("telescope")
	require.NoError(t, err)
	assert.True(t, retrieved.Config.Valid)
	assert.Contains(t, retrieved.Config.String, "telescope.setup")
	assert.Contains(t, retrieved.Config.String, "mappings")
}
