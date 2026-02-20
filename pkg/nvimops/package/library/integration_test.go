package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRmkohlmanPackage tests the complete rmkohlman package
func TestRmkohlmanPackage(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package exists
	assert.True(t, lib.Has("rmkohlman"))

	// Test package can be retrieved
	pkg, ok := lib.Get("rmkohlman")
	require.True(t, ok)
	require.NotNil(t, pkg)

	// Test basic metadata
	assert.Equal(t, "rmkohlman", pkg.Name)
	assert.Equal(t, "complete", pkg.Category)
	assert.Contains(t, pkg.Tags, "complete")
	assert.Contains(t, pkg.Tags, "ide")
	assert.Contains(t, pkg.Tags, "enhanced")
	assert.Contains(t, pkg.Description, "Robert's complete Neovim IDE setup")

	// Test plugin count (should be 37 plugins)
	assert.Len(t, pkg.Plugins, 37)

	// Test essential plugins are included
	pluginNames := pkg.Plugins

	// Core plugins
	assert.Contains(t, pluginNames, "telescope")
	assert.Contains(t, pluginNames, "treesitter")
	assert.Contains(t, pluginNames, "nvim-cmp")
	assert.Contains(t, pluginNames, "which-key")
	assert.Contains(t, pluginNames, "lualine")

	// LSP plugins
	assert.Contains(t, pluginNames, "mason")
	assert.Contains(t, pluginNames, "lspconfig")

	// AI/Copilot plugins
	assert.Contains(t, pluginNames, "copilot")
	assert.Contains(t, pluginNames, "copilot-chat")
	assert.Contains(t, pluginNames, "copilot-cmp")

	// Theme plugin
	assert.Contains(t, pluginNames, "tokyonight")

	// Git plugins
	assert.Contains(t, pluginNames, "gitsigns")
	assert.Contains(t, pluginNames, "lazygit")

	// Navigation plugins
	assert.Contains(t, pluginNames, "harpoon")
	assert.Contains(t, pluginNames, "nvim-tree")
	assert.Contains(t, pluginNames, "bufferline")

	// Database plugins
	assert.Contains(t, pluginNames, "dadbod")
	assert.Contains(t, pluginNames, "dadbod-ui")
	assert.Contains(t, pluginNames, "dadbod-completion")
	assert.Contains(t, pluginNames, "dbee")

	// Markdown plugins
	assert.Contains(t, pluginNames, "render-markdown")
	assert.Contains(t, pluginNames, "markdown-preview")
	assert.Contains(t, pluginNames, "obsidian")
}

// TestRmkohlmanPackageInLibraryList tests that the package appears in library listings
func TestRmkohlmanPackageInLibraryList(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package appears in full list
	packages := lib.List()
	var found bool
	for _, pkg := range packages {
		if pkg.Name == "rmkohlman" {
			found = true
			assert.Equal(t, "complete", pkg.Category)
			break
		}
	}
	assert.True(t, found, "rmkohlman should be in package list")

	// Test package appears in complete category
	completePackages := lib.ListByCategory("complete")
	found = false
	for _, pkg := range completePackages {
		if pkg.Name == "rmkohlman" {
			found = true
			break
		}
	}
	assert.True(t, found, "rmkohlman should be in complete category")

	// Test package appears when searching by tags
	idePackages := lib.ListByTag("ide")
	found = false
	for _, pkg := range idePackages {
		if pkg.Name == "rmkohlman" {
			found = true
			break
		}
	}
	assert.True(t, found, "rmkohlman should be found by 'ide' tag")

	// Test library count includes rmkohlman
	count := lib.Count()
	assert.Equal(t, 5, count) // core, go-dev, python-dev, full, rmkohlman
}

// TestRmkohlmanPackageInfo tests package info functionality
func TestRmkohlmanPackageInfo(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package info
	info := lib.Info()

	var rmkohlmanInfo *PackageInfo
	for _, pkg := range info {
		if pkg.Name == "rmkohlman" {
			rmkohlmanInfo = &pkg
			break
		}
	}

	require.NotNil(t, rmkohlmanInfo, "rmkohlman should appear in package info")
	assert.Equal(t, "rmkohlman", rmkohlmanInfo.Name)
	assert.Equal(t, "complete", rmkohlmanInfo.Category)
	assert.Equal(t, 37, rmkohlmanInfo.PluginCount)
	assert.Contains(t, rmkohlmanInfo.Tags, "complete")
	assert.Contains(t, rmkohlmanInfo.Tags, "ide")
	assert.Contains(t, rmkohlmanInfo.Tags, "enhanced")
}
