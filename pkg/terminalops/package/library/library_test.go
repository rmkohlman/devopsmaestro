package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLibrary(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)
	require.NotNil(t, lib)

	// Should load embedded packages
	assert.True(t, lib.Count() > 0, "library should have packages")

	// Should find core package
	core, ok := lib.Get("core")
	assert.True(t, ok, "should find core package")
	assert.Equal(t, "core", core.Name)
	assert.Equal(t, "base", core.Category)
	assert.Contains(t, core.Tags, "essentials")
	assert.Contains(t, core.Plugins, "zsh-autosuggestions")
}

func TestLibraryGet(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test existing package
	pkg, ok := lib.Get("core")
	assert.True(t, ok)
	assert.NotNil(t, pkg)
	assert.Equal(t, "core", pkg.Name)

	// Test non-existent package
	pkg, ok = lib.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, pkg)
}

func TestLibraryList(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	packages := lib.List()
	assert.Greater(t, len(packages), 0, "should have packages")

	// Should be sorted by name
	for i := 1; i < len(packages); i++ {
		assert.True(t, packages[i-1].Name <= packages[i].Name,
			"packages should be sorted by name")
	}
}

func TestLibraryListByCategory(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test base category
	basePackages := lib.ListByCategory("base")
	assert.Greater(t, len(basePackages), 0, "should have base packages")

	for _, pkg := range basePackages {
		assert.Equal(t, "base", pkg.Category)
	}

	// Test non-existent category
	nonExistent := lib.ListByCategory("nonexistent")
	assert.Empty(t, nonExistent)
}

func TestLibraryListByTag(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test essentials tag
	essentialPackages := lib.ListByTag("essentials")
	assert.Greater(t, len(essentialPackages), 0, "should have essential packages")

	for _, pkg := range essentialPackages {
		assert.Contains(t, pkg.Tags, "essentials")
	}

	// Test non-existent tag
	nonExistent := lib.ListByTag("nonexistent")
	assert.Empty(t, nonExistent)
}

func TestLibraryHas(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	// Test existing packages
	assert.True(t, lib.Has("core"))

	// Test non-existent package
	assert.False(t, lib.Has("nonexistent"))
}

func TestLibraryCategories(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	categories := lib.Categories()
	assert.Greater(t, len(categories), 0, "should have categories")
	assert.Contains(t, categories, "base")

	// Should be sorted
	for i := 1; i < len(categories); i++ {
		assert.True(t, categories[i-1] <= categories[i],
			"categories should be sorted")
	}
}

func TestLibraryTags(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	tags := lib.Tags()
	assert.Greater(t, len(tags), 0, "should have tags")
	assert.Contains(t, tags, "essentials")

	// Should be sorted
	for i := 1; i < len(tags); i++ {
		assert.True(t, tags[i-1] <= tags[i],
			"tags should be sorted")
	}
}

func TestLibraryInfo(t *testing.T) {
	lib, err := NewLibrary()
	require.NoError(t, err)

	info := lib.Info()
	assert.Greater(t, len(info), 0, "should have package info")

	// Find core package info
	var coreInfo *PackageInfo
	for i := range info {
		if info[i].Name == "core" {
			coreInfo = &info[i]
			break
		}
	}

	require.NotNil(t, coreInfo, "should find core package info")
	assert.Equal(t, "core", coreInfo.Name)
	assert.Equal(t, "base", coreInfo.Category)
	assert.Greater(t, coreInfo.PluginCount, 0, "core should have plugins")

	// Should be sorted by name
	for i := 1; i < len(info); i++ {
		assert.True(t, info[i-1].Name <= info[i].Name,
			"info should be sorted by name")
	}
}

func TestNewLibraryFromDir(t *testing.T) {
	// This test would require creating temporary files
	// For now, just test that it doesn't panic
	lib, err := NewLibraryFromDir("/nonexistent")
	assert.Error(t, err)
	assert.Nil(t, lib)
}
