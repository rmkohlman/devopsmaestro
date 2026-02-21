package models

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginToYAML_SimplePlugin(t *testing.T) {
	plugin := &NvimPluginDB{
		ID:          1,
		Name:        "telescope",
		Repo:        "nvim-telescope/telescope.nvim",
		Description: sql.NullString{String: "Fuzzy finder", Valid: true},
		Category:    sql.NullString{String: "fuzzy-finder", Valid: true},
		Enabled:     true,
	}

	pluginYAML, err := plugin.ToYAML()
	require.NoError(t, err)

	assert.Equal(t, "devopsmaestro.io/v1", pluginYAML.APIVersion)
	assert.Equal(t, "NvimPlugin", pluginYAML.Kind)
	assert.Equal(t, "telescope", pluginYAML.Metadata.Name)
	assert.Equal(t, "Fuzzy finder", pluginYAML.Metadata.Description)
	assert.Equal(t, "fuzzy-finder", pluginYAML.Metadata.Category)
	assert.Equal(t, "nvim-telescope/telescope.nvim", pluginYAML.Spec.Repo)
}

func TestPluginFromYAML_SimplePlugin(t *testing.T) {
	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name:        "mason",
			Description: "LSP installer",
			Category:    "lsp",
			Tags:        []string{"lsp", "tools"},
		},
		Spec: PluginSpec{
			Repo:    "williamboman/mason.nvim",
			Branch:  "main",
			Version: "v1.0.0",
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.Equal(t, "mason", plugin.Name)
	assert.Equal(t, "williamboman/mason.nvim", plugin.Repo)
	assert.True(t, plugin.Description.Valid)
	assert.Equal(t, "LSP installer", plugin.Description.String)
	assert.True(t, plugin.Category.Valid)
	assert.Equal(t, "lsp", plugin.Category.String)
	assert.True(t, plugin.Branch.Valid)
	assert.Equal(t, "main", plugin.Branch.String)
}

func TestPluginWithDependencies_ToYAML(t *testing.T) {
	plugin := &NvimPluginDB{
		Name:         "telescope",
		Repo:         "nvim-telescope/telescope.nvim",
		Dependencies: sql.NullString{String: `["nvim-lua/plenary.nvim","nvim-telescope/telescope-fzf-native.nvim"]`, Valid: true},
		Enabled:      true,
	}

	pluginYAML, err := plugin.ToYAML()
	require.NoError(t, err)

	assert.NotNil(t, pluginYAML.Spec.Dependencies)
	assert.Len(t, pluginYAML.Spec.Dependencies, 2)
	assert.Contains(t, pluginYAML.Spec.Dependencies, "nvim-lua/plenary.nvim")
	assert.Contains(t, pluginYAML.Spec.Dependencies, "nvim-telescope/telescope-fzf-native.nvim")
}

func TestPluginWithDependencies_FromYAML(t *testing.T) {
	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: "telescope",
		},
		Spec: PluginSpec{
			Repo: "nvim-telescope/telescope.nvim",
			Dependencies: []interface{}{
				"nvim-lua/plenary.nvim",
				"nvim-tree/nvim-web-devicons",
			},
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.True(t, plugin.Dependencies.Valid)
	assert.Contains(t, plugin.Dependencies.String, "plenary")
	assert.Contains(t, plugin.Dependencies.String, "devicons")
}

func TestPluginWithKeymaps_ToYAML(t *testing.T) {
	plugin := &NvimPluginDB{
		Name:    "telescope",
		Repo:    "nvim-telescope/telescope.nvim",
		Keymaps: sql.NullString{String: `[{"key":"<leader>ff","mode":"n","action":"<cmd>Telescope find_files<cr>","desc":"Find files"}]`, Valid: true},
		Enabled: true,
	}

	pluginYAML, err := plugin.ToYAML()
	require.NoError(t, err)

	assert.NotNil(t, pluginYAML.Spec.Keymaps)
	assert.Len(t, pluginYAML.Spec.Keymaps, 1)
	assert.Equal(t, "<leader>ff", pluginYAML.Spec.Keymaps[0].Key)
	assert.Equal(t, "n", pluginYAML.Spec.Keymaps[0].Mode)
	assert.Equal(t, "Find files", pluginYAML.Spec.Keymaps[0].Desc)
}

func TestPluginWithKeymaps_FromYAML(t *testing.T) {
	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: "telescope",
		},
		Spec: PluginSpec{
			Repo: "nvim-telescope/telescope.nvim",
			Keymaps: []PluginKeymap{
				{
					Key:    "<leader>ff",
					Mode:   "n",
					Action: "<cmd>Telescope find_files<cr>",
					Desc:   "Find files",
				},
				{
					Key:    "<leader>fg",
					Mode:   "n",
					Action: "<cmd>Telescope live_grep<cr>",
					Desc:   "Live grep",
				},
			},
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.True(t, plugin.Keymaps.Valid)
	// Just verify it's valid JSON containing the keys
	assert.Contains(t, plugin.Keymaps.String, "leader")
	assert.Contains(t, plugin.Keymaps.String, "ff")
	assert.Contains(t, plugin.Keymaps.String, "fg")
}

func TestPluginWithKeys_ToYAML(t *testing.T) {
	plugin := &NvimPluginDB{
		Name:    "lazygit",
		Repo:    "kdheepak/lazygit.nvim",
		Keys:    sql.NullString{String: `[{"key":"<leader>gg"}]`, Valid: true},
		Enabled: true,
	}

	pluginYAML, err := plugin.ToYAML()
	require.NoError(t, err)

	assert.NotNil(t, pluginYAML.Spec.Keys)
	assert.Len(t, pluginYAML.Spec.Keys, 1)
	assert.Equal(t, "<leader>gg", pluginYAML.Spec.Keys[0].Key)
}

func TestPluginWithKeys_FromYAML(t *testing.T) {
	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: "lazygit",
		},
		Spec: PluginSpec{
			Repo: "kdheepak/lazygit.nvim",
			Keys: []PluginKeymap{
				{Key: "<leader>gg"},
				{Key: "<leader>gc"},
			},
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.True(t, plugin.Keys.Valid)
	assert.Contains(t, plugin.Keys.String, "leader")
	assert.Contains(t, plugin.Keys.String, "gg")
}

func TestPluginWithComplexConfig_ToYAML(t *testing.T) {
	configLua := `local telescope = require("telescope")
telescope.setup({
  defaults = {
    mappings = {
      i = {
        ["<C-u>"] = false,
      },
    },
  },
})`

	plugin := &NvimPluginDB{
		Name:    "telescope",
		Repo:    "nvim-telescope/telescope.nvim",
		Config:  sql.NullString{String: configLua, Valid: true},
		Enabled: true,
	}

	pluginYAML, err := plugin.ToYAML()
	require.NoError(t, err)

	assert.Equal(t, configLua, pluginYAML.Spec.Config)
	assert.Contains(t, pluginYAML.Spec.Config, "telescope.setup")
}

func TestPluginWithComplexConfig_FromYAML(t *testing.T) {
	configLua := `require("mason").setup({
  ui = {
    border = "rounded"
  }
})`

	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: "mason",
		},
		Spec: PluginSpec{
			Repo:   "williamboman/mason.nvim",
			Config: configLua,
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.True(t, plugin.Config.Valid)
	assert.Equal(t, configLua, plugin.Config.String)
	assert.Contains(t, plugin.Config.String, "require")
}

func TestPluginRoundTrip_ToYAML_FromYAML(t *testing.T) {
	// Original plugin
	original := &NvimPluginDB{
		Name:         "telescope",
		Repo:         "nvim-telescope/telescope.nvim",
		Branch:       sql.NullString{String: "0.1.x", Valid: true},
		Description:  sql.NullString{String: "Fuzzy finder", Valid: true},
		Category:     sql.NullString{String: "fuzzy-finder", Valid: true},
		Event:        sql.NullString{String: `"VeryLazy"`, Valid: true}, // JSON-encoded string
		Dependencies: sql.NullString{String: `["nvim-lua/plenary.nvim"]`, Valid: true},
		Config:       sql.NullString{String: `require("telescope").setup()`, Valid: true},
		Enabled:      true,
	}

	// Convert to YAML
	pluginYAML, err := original.ToYAML()
	require.NoError(t, err)

	// Convert back to DB model
	roundtrip := &NvimPluginDB{}
	err = roundtrip.FromYAML(pluginYAML)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, original.Name, roundtrip.Name)
	assert.Equal(t, original.Repo, roundtrip.Repo)
	assert.Equal(t, original.Branch.String, roundtrip.Branch.String)
	assert.Equal(t, original.Description.String, roundtrip.Description.String)
	assert.Equal(t, original.Category.String, roundtrip.Category.String)
	assert.Contains(t, roundtrip.Event.String, "VeryLazy") // JSON-encoded may have quotes
	assert.Equal(t, original.Config.String, roundtrip.Config.String)
	assert.Equal(t, original.Enabled, roundtrip.Enabled)
}

func TestPluginWithTags_ToYAML(t *testing.T) {
	plugin := &NvimPluginDB{
		Name:    "telescope",
		Repo:    "nvim-telescope/telescope.nvim",
		Tags:    sql.NullString{String: `["finder","search","fuzzy"]`, Valid: true},
		Enabled: true,
	}

	pluginYAML, err := plugin.ToYAML()
	require.NoError(t, err)

	assert.NotNil(t, pluginYAML.Metadata.Tags)
	assert.Len(t, pluginYAML.Metadata.Tags, 3)
	assert.Contains(t, pluginYAML.Metadata.Tags, "finder")
	assert.Contains(t, pluginYAML.Metadata.Tags, "search")
}

func TestPluginWithTags_FromYAML(t *testing.T) {
	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: "telescope",
			Tags: []string{"finder", "search", "fuzzy"},
		},
		Spec: PluginSpec{
			Repo: "nvim-telescope/telescope.nvim",
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.True(t, plugin.Tags.Valid)
	assert.Contains(t, plugin.Tags.String, "finder")
}

func TestPluginWithLazyLoading_FromYAML(t *testing.T) {
	pluginYAML := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: "telescope",
		},
		Spec: PluginSpec{
			Repo:  "nvim-telescope/telescope.nvim",
			Lazy:  true,
			Event: "VeryLazy",
			Ft:    []string{"lua", "vim"},
			Cmd:   []string{"Telescope"},
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromYAML(pluginYAML)
	require.NoError(t, err)

	assert.True(t, plugin.Lazy)
	assert.True(t, plugin.Event.Valid)
	assert.Contains(t, plugin.Event.String, "VeryLazy") // May be JSON-encoded
	assert.True(t, plugin.Ft.Valid)
	assert.Contains(t, plugin.Ft.String, "lua")
	assert.True(t, plugin.Cmd.Valid)
	assert.Contains(t, plugin.Cmd.String, "Telescope")
}

func TestPluginFromNvimOpsPlugin(t *testing.T) {
	// Create a mock plugin using the structure that would come from nvimops
	mockPlugin := map[string]interface{}{
		"name":        "telescope",
		"description": "Fuzzy finder for neovim",
		"repo":        "nvim-telescope/telescope.nvim",
		"category":    "navigation",
		"enabled":     true,
		"lazy":        true,
		"priority":    float64(100), // JSON numbers are float64
		"event":       []interface{}{"VeryLazy"},
		"ft":          []interface{}{"lua", "vim"},
		"tags":        []interface{}{"telescope", "fuzzy-finder"},
		"config":      "require('telescope').setup({})",
		"keymaps": []interface{}{
			map[string]interface{}{
				"key":    "<leader>ff",
				"action": "<cmd>Telescope find_files<cr>",
				"desc":   "Find files",
			},
		},
	}

	plugin := &NvimPluginDB{}
	err := plugin.FromNvimOpsPlugin(mockPlugin)
	require.NoError(t, err)

	assert.Equal(t, "telescope", plugin.Name)
	assert.Equal(t, "nvim-telescope/telescope.nvim", plugin.Repo)
	assert.True(t, plugin.Enabled)
	assert.True(t, plugin.Lazy)

	assert.True(t, plugin.Description.Valid)
	assert.Equal(t, "Fuzzy finder for neovim", plugin.Description.String)

	assert.True(t, plugin.Category.Valid)
	assert.Equal(t, "navigation", plugin.Category.String)

	assert.True(t, plugin.Priority.Valid)
	assert.Equal(t, int64(100), plugin.Priority.Int64)

	assert.True(t, plugin.Event.Valid)
	assert.Contains(t, plugin.Event.String, "VeryLazy")

	assert.True(t, plugin.Ft.Valid)
	assert.Contains(t, plugin.Ft.String, "lua")

	assert.True(t, plugin.Tags.Valid)
	assert.Contains(t, plugin.Tags.String, "telescope")

	assert.True(t, plugin.Config.Valid)
	assert.Equal(t, "require('telescope').setup({})", plugin.Config.String)

	assert.True(t, plugin.Keymaps.Valid)
	assert.Contains(t, plugin.Keymaps.String, "leader")
}
