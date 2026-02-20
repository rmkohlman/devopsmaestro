package sources

import (
	"context"
	"testing"

	"devopsmaestro/pkg/nvimops/sync"
)

func TestLazyVimHandler_Name(t *testing.T) {
	handler := NewLazyVimHandler()
	if handler.Name() != "lazyvim" {
		t.Errorf("Expected name 'lazyvim', got %s", handler.Name())
	}
}

func TestLazyVimHandler_Description(t *testing.T) {
	handler := NewLazyVimHandler()
	expected := "LazyVim - A Neovim config for lazy people"
	if handler.Description() != expected {
		t.Errorf("Expected description %s, got %s", expected, handler.Description())
	}
}

func TestLazyVimHandler_Validate(t *testing.T) {
	handler := NewLazyVimHandler()
	ctx := context.Background()

	// Note: This will make a real HTTP request to GitHub
	// In a production test, you'd want to mock this
	err := handler.Validate(ctx)
	if err != nil {
		t.Skipf("GitHub API not accessible: %v", err)
	}
}

func TestRegisterLazyVimHandler(t *testing.T) {
	// Create a test registry
	registry := sync.NewSourceRegistry()

	// Register builtin sources first (to set up the placeholder)
	err := sync.RegisterBuiltinSources(registry)
	if err != nil {
		t.Fatalf("Failed to register builtin sources: %v", err)
	}

	// Verify the placeholder is registered
	if !registry.IsRegistered("lazyvim") {
		t.Fatal("LazyVim placeholder not registered")
	}

	// Register the actual handler
	err = RegisterLazyVimHandler(registry)
	if err != nil {
		t.Fatalf("Failed to register LazyVim handler: %v", err)
	}

	// Verify the handler is registered and not a placeholder
	factory := sync.NewSourceHandlerFactoryWithRegistry(registry)
	handler, err := factory.CreateHandler("lazyvim")
	if err != nil {
		t.Fatalf("Failed to create LazyVim handler: %v", err)
	}

	// Verify it's the actual handler, not the placeholder
	if _, isPlaceholder := handler.(*sync.NotImplementedHandler); isPlaceholder {
		t.Error("Handler is still a placeholder after registration")
	}

	if _, isLazyVim := handler.(*LazyVimHandler); !isLazyVim {
		t.Error("Handler is not a LazyVimHandler")
	}
}

func TestRegisterAllHandlers(t *testing.T) {
	// Create a test registry
	registry := sync.NewSourceRegistry()

	// Register builtin sources first
	err := sync.RegisterBuiltinSources(registry)
	if err != nil {
		t.Fatalf("Failed to register builtin sources: %v", err)
	}

	// Register all handlers
	err = RegisterAllHandlers(registry)
	if err != nil {
		t.Fatalf("Failed to register all handlers: %v", err)
	}

	// Verify LazyVim is registered with actual handler
	factory := sync.NewSourceHandlerFactoryWithRegistry(registry)
	handler, err := factory.CreateHandler("lazyvim")
	if err != nil {
		t.Fatalf("Failed to create LazyVim handler: %v", err)
	}

	if _, isLazyVim := handler.(*LazyVimHandler); !isLazyVim {
		t.Error("LazyVim handler is not registered correctly")
	}
}

func TestExtractCategoryFromFilename(t *testing.T) {
	handler := &LazyVimHandler{}

	tests := []struct {
		filename string
		expected string
	}{
		{"coding.lua", "coding"},
		{"colorscheme.lua", "theme"},
		{"editor.lua", "editor"},
		{"formatting.lua", "formatting"},
		{"linting.lua", "linting"},
		{"treesitter.lua", "syntax"},
		{"ui.lua", "ui"},
		{"util.lua", "utility"},
		{"lsp/init.lua", "lsp"},
		{"extras/lang/go.lua", "misc"},
		{"unknown.lua", "misc"},
	}

	for _, test := range tests {
		result := handler.extractCategoryFromFilename(test.filename)
		if result != test.expected {
			t.Errorf("extractCategoryFromFilename(%s) = %s, expected %s",
				test.filename, result, test.expected)
		}
	}
}

func TestExtractPluginNameFromRepo(t *testing.T) {
	handler := &LazyVimHandler{}

	tests := []struct {
		repo     string
		expected string
	}{
		{"nvim-telescope/telescope.nvim", "telescope"},
		{"nvim-treesitter/nvim-treesitter", "treesitter"},
		{"hrsh7th/nvim-cmp", "cmp"},
		{"folke/which-key.nvim", "which-key"},
		{"single-part", "single-part"},
		{"user/repo-name.vim", "repo-name"},
	}

	for _, test := range tests {
		result := handler.extractPluginNameFromRepo(test.repo)
		if result != test.expected {
			t.Errorf("extractPluginNameFromRepo(%s) = %s, expected %s",
				test.repo, result, test.expected)
		}
	}
}

func TestParseLuaContent(t *testing.T) {
	handler := &LazyVimHandler{version: "v1.0.0"}

	luaContent := `
return {
  { "nvim-telescope/telescope.nvim", dependencies = { "nvim-lua/plenary.nvim" } },
  { "hrsh7th/nvim-cmp", config = function() require("cmp").setup({}) end },
}
`

	plugins, err := handler.parseLuaContent(luaContent, "editor.lua")
	if err != nil {
		t.Fatalf("Failed to parse Lua content: %v", err)
	}

	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}

	// Check first plugin
	if len(plugins) > 0 {
		plugin := plugins[0]
		if plugin.Repo != "nvim-telescope/telescope.nvim" {
			t.Errorf("Expected repo 'nvim-telescope/telescope.nvim', got %s", plugin.Repo)
		}
		if plugin.Name != "lazyvim-telescope" {
			t.Errorf("Expected name 'lazyvim-telescope', got %s", plugin.Name)
		}
		if plugin.Category != "editor" {
			t.Errorf("Expected category 'editor', got %s", plugin.Category)
		}
		if plugin.Labels["source"] != "lazyvim" {
			t.Errorf("Expected source label 'lazyvim', got %s", plugin.Labels["source"])
		}
	}
}
