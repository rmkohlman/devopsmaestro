package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRkohlmanFullPackage tests the complete rkohlman-full package
func TestRkohlmanFullPackage(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package exists
	assert.True(t, lib.Has("rkohlman-full"))

	// Test package can be retrieved
	pkg, ok := lib.Get("rkohlman-full")
	require.True(t, ok)
	require.NotNil(t, pkg)

	// Test basic metadata
	assert.Equal(t, "rkohlman-full", pkg.Name)
	assert.Equal(t, "complete", pkg.Category)
	assert.Contains(t, pkg.Tags, "complete")
	assert.Contains(t, pkg.Tags, "ide")
	assert.Contains(t, pkg.Tags, "enhanced")
	assert.Contains(t, pkg.Description, "Complete Neovim IDE setup")

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

// TestRkohlmanPackageInLibraryList tests that the package appears in library listings
func TestRkohlmanPackageInLibraryList(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package appears in full list
	packages := lib.List()
	var found bool
	for _, pkg := range packages {
		if pkg.Name == "rkohlman-full" {
			found = true
			assert.Equal(t, "complete", pkg.Category)
			break
		}
	}
	assert.True(t, found, "rkohlman-full should be in package list")

	// Test package appears in complete category
	completePackages := lib.ListByCategory("complete")
	found = false
	for _, pkg := range completePackages {
		if pkg.Name == "rkohlman-full" {
			found = true
			break
		}
	}
	assert.True(t, found, "rkohlman-full should be in complete category")

	// Test package appears when searching by tags
	idePackages := lib.ListByTag("ide")
	found = false
	for _, pkg := range idePackages {
		if pkg.Name == "rkohlman-full" {
			found = true
			break
		}
	}
	assert.True(t, found, "rkohlman-full should be found by 'ide' tag")

	// Test library count includes rkohlman-full
	count := lib.Count()
	assert.Equal(t, 5, count) // core, go-dev, python-dev, full, rkohlman-full
}

// TestRkohlmanPackageInfo tests package info functionality
func TestRkohlmanPackageInfo(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package info
	info := lib.Info()

	var rkohlmanInfo *PackageInfo
	for _, pkg := range info {
		if pkg.Name == "rkohlman-full" {
			rkohlmanInfo = &pkg
			break
		}
	}

	require.NotNil(t, rkohlmanInfo, "rkohlman-full should appear in package info")
	assert.Equal(t, "rkohlman-full", rkohlmanInfo.Name)
	assert.Equal(t, "complete", rkohlmanInfo.Category)
	assert.Equal(t, 37, rkohlmanInfo.PluginCount)
	assert.Contains(t, rkohlmanInfo.Tags, "complete")
	assert.Contains(t, rkohlmanInfo.Tags, "ide")
	assert.Contains(t, rkohlmanInfo.Tags, "enhanced")
}
