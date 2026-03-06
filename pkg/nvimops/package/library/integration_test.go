package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resolvePlugins resolves the full plugin list for a package by walking the extends chain.
// Parent plugins come first; child plugins are appended with deduplication.
func resolvePlugins(lib *Library, packageName string) []string {
	p, ok := lib.Get(packageName)
	if !ok {
		return nil
	}

	var base []string
	if p.Extends != "" {
		base = resolvePlugins(lib, p.Extends)
	}

	// Combine: parent first, then child — deduplicated
	seen := make(map[string]bool, len(base)+len(p.Plugins))
	result := make([]string, 0, len(base)+len(p.Plugins))

	for _, pl := range base {
		if !seen[pl] {
			seen[pl] = true
			result = append(result, pl)
		}
	}
	for _, pl := range p.Plugins {
		if !seen[pl] {
			seen[pl] = true
			result = append(result, pl)
		}
	}

	return result
}

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

// TestMaestroLanguagePackages_ExtendsResolution verifies that the extends inheritance chain
// produces the correct total plugin count for all 7 language-specific maestro packages.
func TestMaestroLanguagePackages_ExtendsResolution(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Verify the maestro base package has exactly 37 plugins (foundational assumption)
	maestro, ok := lib.Get("maestro")
	require.True(t, ok, "maestro base package must exist")
	require.Len(t, maestro.Plugins, 37, "maestro base package must have 37 plugins")

	tests := []struct {
		name          string
		packageName   string
		ownPlugins    []string
		expectedTotal int
	}{
		{
			name:          "maestro-go",
			packageName:   "maestro-go",
			ownPlugins:    []string{"nvim-dap", "nvim-dap-go", "neotest", "neotest-go", "gopher-nvim"},
			expectedTotal: 42,
		},
		{
			name:          "maestro-python",
			packageName:   "maestro-python",
			ownPlugins:    []string{"nvim-dap", "nvim-dap-python", "neotest", "neotest-python", "venv-selector"},
			expectedTotal: 42,
		},
		{
			name:          "maestro-rust",
			packageName:   "maestro-rust",
			ownPlugins:    []string{"nvim-dap", "rustaceanvim", "crates-nvim", "neotest", "neotest-rust"},
			expectedTotal: 42,
		},
		{
			name:          "maestro-node",
			packageName:   "maestro-node",
			ownPlugins:    []string{"nvim-dap", "neotest", "neotest-jest"},
			expectedTotal: 40,
		},
		{
			name:          "maestro-java",
			packageName:   "maestro-java",
			ownPlugins:    []string{"nvim-dap", "nvim-jdtls", "neotest"},
			expectedTotal: 40,
		},
		{
			name:          "maestro-gleam",
			packageName:   "maestro-gleam",
			ownPlugins:    []string{},
			expectedTotal: 37,
		},
		{
			name:          "maestro-dotnet",
			packageName:   "maestro-dotnet",
			ownPlugins:    []string{"nvim-dap", "neotest"},
			expectedTotal: 39,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Package must exist in the library
			p, ok := lib.Get(tt.packageName)
			require.True(t, ok, "package %q must exist in library", tt.packageName)
			require.NotNil(t, p)

			// Must extend "maestro"
			assert.Equal(t, "maestro", p.Extends, "package %q must extend maestro", tt.packageName)

			// Own plugin count must match
			assert.Len(t, p.Plugins, len(tt.ownPlugins),
				"package %q: own plugin count mismatch", tt.packageName)

			// Each expected own plugin must be present
			for _, expectedPlugin := range tt.ownPlugins {
				assert.Contains(t, p.Plugins, expectedPlugin,
					"package %q must contain own plugin %q", tt.packageName, expectedPlugin)
			}

			// Resolved total (parent + own, deduplicated) must match expected
			resolved := resolvePlugins(lib, tt.packageName)
			assert.Len(t, resolved, tt.expectedTotal,
				"package %q: resolved total plugin count mismatch (parent 37 + own %d = %d)",
				tt.packageName, len(tt.ownPlugins), tt.expectedTotal)
		})
	}
}

// TestMaestroLanguagePackages_NoDuplicatePlugins verifies that no language-specific maestro
// package has duplicate plugin names in its own Plugins list.
func TestMaestroLanguagePackages_NoDuplicatePlugins(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	languagePackages := []string{
		"maestro-go",
		"maestro-python",
		"maestro-rust",
		"maestro-node",
		"maestro-java",
		"maestro-gleam",
		"maestro-dotnet",
	}

	for _, name := range languagePackages {
		t.Run(name, func(t *testing.T) {
			p, ok := lib.Get(name)
			require.True(t, ok, "package %q must exist in library", name)

			seen := make(map[string]bool, len(p.Plugins))
			for _, pl := range p.Plugins {
				assert.False(t, seen[pl],
					"package %q has duplicate plugin %q in its own Plugins list", name, pl)
				seen[pl] = true
			}
		})
	}
}
