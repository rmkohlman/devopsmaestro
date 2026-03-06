package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaestroPackage tests the complete maestro package
func TestMaestroPackage(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package exists
	assert.True(t, lib.Has("maestro"))

	// Test package can be retrieved
	pkg, ok := lib.Get("maestro")
	require.True(t, ok)
	require.NotNil(t, pkg)

	// Test basic metadata
	assert.Equal(t, "maestro", pkg.Name)
	assert.Equal(t, "complete", pkg.Category)
	assert.Contains(t, pkg.Tags, "complete")
	assert.Contains(t, pkg.Tags, "ide")
	assert.Contains(t, pkg.Tags, "enhanced")
	assert.Contains(t, pkg.Description, "Maestro complete Neovim IDE setup")

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

// TestMaestroPackageInLibraryList tests that the package appears in library listings
func TestMaestroPackageInLibraryList(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package appears in full list
	packages := lib.List()
	var found bool
	for _, pkg := range packages {
		if pkg.Name == "maestro" {
			found = true
			assert.Equal(t, "complete", pkg.Category)
			break
		}
	}
	assert.True(t, found, "maestro should be in package list")

	// Test package appears in complete category
	completePackages := lib.ListByCategory("complete")
	found = false
	for _, pkg := range completePackages {
		if pkg.Name == "maestro" {
			found = true
			break
		}
	}
	assert.True(t, found, "maestro should be in complete category")

	// Test package appears when searching by tags
	idePackages := lib.ListByTag("ide")
	found = false
	for _, pkg := range idePackages {
		if pkg.Name == "maestro" {
			found = true
			break
		}
	}
	assert.True(t, found, "maestro should be found by 'ide' tag")

	// Test library count includes maestro
	count := lib.Count()
	assert.Equal(t, 12, count) // core, full, maestro, go-dev, python-dev, maestro-go, maestro-python, maestro-rust, maestro-node, maestro-java, maestro-gleam, maestro-dotnet
}

// TestMaestroPackageInfo tests package info functionality
func TestMaestroPackageInfo(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test package info
	info := lib.Info()

	var maestroInfo *PackageInfo
	for _, pkg := range info {
		if pkg.Name == "maestro" {
			maestroInfo = &pkg
			break
		}
	}

	require.NotNil(t, maestroInfo, "maestro should appear in package info")
	assert.Equal(t, "maestro", maestroInfo.Name)
	assert.Equal(t, "complete", maestroInfo.Category)
	assert.Equal(t, 37, maestroInfo.PluginCount)
	assert.Contains(t, maestroInfo.Tags, "complete")
	assert.Contains(t, maestroInfo.Tags, "ide")
	assert.Contains(t, maestroInfo.Tags, "enhanced")
}
