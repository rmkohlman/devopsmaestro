package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLibrary(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)
	assert.NotNil(t, lib)

	// Should load embedded packages
	assert.Greater(t, lib.Count(), 0)

	// Check that core packages are loaded
	expectedPackages := []string{"core", "go-dev", "python-dev", "full"}
	for _, name := range expectedPackages {
		assert.True(t, lib.Has(name), "Package %s should be loaded", name)
	}
}

func TestLibraryGet(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test getting existing package
	pkg, ok := lib.Get("core")
	assert.True(t, ok)
	assert.NotNil(t, pkg)
	assert.Equal(t, "core", pkg.Name)
	assert.Equal(t, "Essential Neovim plugins for any development", pkg.Description)
	assert.Equal(t, "core", pkg.Category)
	assert.Contains(t, pkg.Tags, "essential")
	assert.Contains(t, pkg.Tags, "base")

	// Test getting non-existing package
	_, ok = lib.Get("non-existent")
	assert.False(t, ok)
}

func TestLibraryList(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	packages := lib.List()
	assert.Greater(t, len(packages), 0)

	// Should be sorted by name
	for i := 1; i < len(packages); i++ {
		assert.LessOrEqual(t, packages[i-1].Name, packages[i].Name)
	}
}

func TestLibraryListByCategory(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test language category
	languagePackages := lib.ListByCategory("language")
	assert.Greater(t, len(languagePackages), 0)

	for _, pkg := range languagePackages {
		assert.Equal(t, "language", pkg.Category)
	}

	// Test non-existing category
	emptyPackages := lib.ListByCategory("non-existent")
	assert.Empty(t, emptyPackages)
}

func TestLibraryListByTag(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test essential tag
	essentialPackages := lib.ListByTag("essential")
	assert.Greater(t, len(essentialPackages), 0)

	for _, pkg := range essentialPackages {
		assert.Contains(t, pkg.Tags, "essential")
	}

	// Test golang tag
	golangPackages := lib.ListByTag("golang")
	assert.Greater(t, len(golangPackages), 0)

	for _, pkg := range golangPackages {
		assert.Contains(t, pkg.Tags, "golang")
	}

	// Test non-existing tag
	emptyPackages := lib.ListByTag("non-existent")
	assert.Empty(t, emptyPackages)
}

func TestLibraryCategories(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	categories := lib.Categories()
	assert.Greater(t, len(categories), 0)

	// Should be sorted
	for i := 1; i < len(categories); i++ {
		assert.LessOrEqual(t, categories[i-1], categories[i])
	}

	// Should contain expected categories
	assert.Contains(t, categories, "core")
	assert.Contains(t, categories, "language")
	assert.Contains(t, categories, "complete")
}

func TestLibraryTags(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	tags := lib.Tags()
	assert.Greater(t, len(tags), 0)

	// Should be sorted
	for i := 1; i < len(tags); i++ {
		assert.LessOrEqual(t, tags[i-1], tags[i])
	}

	// Should contain expected tags
	assert.Contains(t, tags, "essential")
	assert.Contains(t, tags, "base")
	assert.Contains(t, tags, "golang")
	assert.Contains(t, tags, "python")
	assert.Contains(t, tags, "lsp")
	assert.Contains(t, tags, "dap")
}

func TestLibraryInfo(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	info := lib.Info()
	assert.Greater(t, len(info), 0)

	// Should be sorted by name
	for i := 1; i < len(info); i++ {
		assert.LessOrEqual(t, info[i-1].Name, info[i].Name)
	}

	// Check core package info
	var coreInfo *PackageInfo
	for _, pkg := range info {
		if pkg.Name == "core" {
			coreInfo = &pkg
			break
		}
	}
	require.NotNil(t, coreInfo)
	assert.Equal(t, "core", coreInfo.Name)
	assert.Equal(t, "Essential Neovim plugins for any development", coreInfo.Description)
	assert.Equal(t, "core", coreInfo.Category)
	assert.Contains(t, coreInfo.Tags, "essential")
	assert.Equal(t, "", coreInfo.Extends) // core doesn't extend anything
	assert.Greater(t, coreInfo.PluginCount, 0)

	// Check go-dev package info
	var goDevInfo *PackageInfo
	for _, pkg := range info {
		if pkg.Name == "go-dev" {
			goDevInfo = &pkg
			break
		}
	}
	require.NotNil(t, goDevInfo)
	assert.Equal(t, "go-dev", goDevInfo.Name)
	assert.Equal(t, "core", goDevInfo.Extends)
	assert.Contains(t, goDevInfo.Tags, "golang")
}

func TestLibraryHas(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	assert.True(t, lib.Has("core"))
	assert.True(t, lib.Has("go-dev"))
	assert.False(t, lib.Has("non-existent"))
}

func TestLibraryCount(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	count := lib.Count()
	assert.Equal(t, 5, count) // core, go-dev, python-dev, full, rkohlman-full
}

func TestPackageExtensions(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test that go-dev extends core
	goDev, ok := lib.Get("go-dev")
	assert.True(t, ok)
	assert.Equal(t, "core", goDev.Extends)

	// Test that python-dev extends core
	pythonDev, ok := lib.Get("python-dev")
	assert.True(t, ok)
	assert.Equal(t, "core", pythonDev.Extends)

	// Test that full extends core
	full, ok := lib.Get("full")
	assert.True(t, ok)
	assert.Equal(t, "core", full.Extends)

	// Test that core doesn't extend anything
	core, ok := lib.Get("core")
	assert.True(t, ok)
	assert.Equal(t, "", core.Extends)
}

func TestPackagePlugins(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test core package has expected plugins
	core, ok := lib.Get("core")
	assert.True(t, ok)
	assert.Contains(t, core.Plugins, "lspconfig")
	assert.Contains(t, core.Plugins, "telescope")
	assert.Contains(t, core.Plugins, "treesitter")

	// Test go-dev package has Go-specific plugins
	goDev, ok := lib.Get("go-dev")
	assert.True(t, ok)
	assert.Contains(t, goDev.Plugins, "gopher-nvim")
	assert.Contains(t, goDev.Plugins, "nvim-dap-go")
	assert.Contains(t, goDev.Plugins, "neotest-go")

	// Test python-dev package has Python-specific plugins
	pythonDev, ok := lib.Get("python-dev")
	assert.True(t, ok)
	assert.Contains(t, pythonDev.Plugins, "nvim-dap-python")
	assert.Contains(t, pythonDev.Plugins, "neotest-python")
	assert.Contains(t, pythonDev.Plugins, "venv-selector")
}
