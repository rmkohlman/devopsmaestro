// Package resolver provides package hierarchy resolution tests.
// These tests drive the implementation of HierarchyPackageResolver.
// File is .pending — CI skips it until the implementation exists.
package resolver

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers: build mock entities with package fields
// ---------------------------------------------------------------------------

func addEcosystemWithPackages(ds *PackageMockDataStore, id int, name string, nvimPkg, termPkg *string) {
	eco := newPackageMockEcosystem(id, name, nvimPkg, termPkg)
	ds.ecosystems[id] = eco
}

func addDomainWithPackages(ds *PackageMockDataStore, id, ecoID int, name string, nvimPkg, termPkg *string) {
	dom := newPackageMockDomain(id, ecoID, name, nvimPkg, termPkg)
	ds.domains[id] = dom
}

func addAppWithPackages(ds *PackageMockDataStore, id, domainID int, name string, nvimPkg, termPkg *string) {
	app := newPackageMockApp(id, domainID, name, nvimPkg, termPkg)
	ds.apps[id] = app
}

func addWorkspaceWithPackages(ds *PackageMockDataStore, id, appID int, name string, nvimPkg, termPkg *string) {
	ws := newPackageMockWorkspace(id, appID, name, nvimPkg, termPkg)
	ds.workspaces[id] = ws
}

func pkgPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_WorkspaceLevelOverride
// workspace has nvim_package set → resolves to it, doesn't walk up
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_WorkspaceLevelOverride(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", pkgPtr("eco-nvim-pkg"), pkgPtr("eco-term-pkg"))
	addDomainWithPackages(ds, 1, 1, "dom", pkgPtr("dom-nvim-pkg"), pkgPtr("dom-term-pkg"))
	addAppWithPackages(ds, 1, 1, "app", pkgPtr("app-nvim-pkg"), pkgPtr("app-term-pkg"))
	addWorkspaceWithPackages(ds, 1, 1, "ws", pkgPtr("ws-nvim-pkg"), pkgPtr("ws-term-pkg"))

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "ws-nvim-pkg", res.PackageName)
	assert.Equal(t, PackageLevelWorkspace, res.Source)
	assert.Equal(t, "ws", res.SourceName)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_AppLevelFallback
// workspace has no package, app has it → resolves to app
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_AppLevelFallback(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", pkgPtr("eco-nvim-pkg"), nil)
	addDomainWithPackages(ds, 1, 1, "dom", nil, nil)
	addAppWithPackages(ds, 1, 1, "app", pkgPtr("app-nvim-pkg"), nil)
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil) // no package at workspace

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "app-nvim-pkg", res.PackageName)
	assert.Equal(t, PackageLevelApp, res.Source)
	assert.Equal(t, "app", res.SourceName)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_DomainLevelFallback
// workspace+app empty, domain has package → resolves to domain
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_DomainLevelFallback(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", nil, nil)
	addDomainWithPackages(ds, 1, 1, "dom", pkgPtr("dom-nvim-pkg"), nil)
	addAppWithPackages(ds, 1, 1, "app", nil, nil)
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil)

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "dom-nvim-pkg", res.PackageName)
	assert.Equal(t, PackageLevelDomain, res.Source)
	assert.Equal(t, "dom", res.SourceName)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_EcosystemLevelFallback
// all levels empty except ecosystem → resolves to ecosystem
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_EcosystemLevelFallback(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", pkgPtr("eco-nvim-pkg"), nil)
	addDomainWithPackages(ds, 1, 1, "dom", nil, nil)
	addAppWithPackages(ds, 1, 1, "app", nil, nil)
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil)

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "eco-nvim-pkg", res.PackageName)
	assert.Equal(t, PackageLevelEcosystem, res.Source)
	assert.Equal(t, "eco", res.SourceName)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_GlobalDefault
// all levels empty, defaults table has value → returns global default
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_GlobalDefault(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", nil, nil)
	addDomainWithPackages(ds, 1, 1, "dom", nil, nil)
	addAppWithPackages(ds, 1, 1, "app", nil, nil)
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil)
	ds.SetDefault("nvim-package", "global-nvim-default")

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "global-nvim-default", res.PackageName)
	assert.Equal(t, PackageLevelGlobal, res.Source)
	assert.Equal(t, "global default", res.SourceName)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_NoPackageAnywhere
// nothing set at any level, no global default → returns empty/error
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_NoPackageAnywhere(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", nil, nil)
	addDomainWithPackages(ds, 1, 1, "dom", nil, nil)
	addAppWithPackages(ds, 1, 1, "app", nil, nil)
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil)
	// No default set in ds.defaults

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	// Either an error or an empty PackageName is acceptable
	if err == nil {
		require.NotNil(t, res)
		assert.Empty(t, res.PackageName,
			"expected empty package name when nothing is configured anywhere")
	}
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_ClearAtLevel
// setting to empty string at a level clears it; inherits from parent
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_ClearAtLevel(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", pkgPtr("eco-nvim-pkg"), nil)
	addDomainWithPackages(ds, 1, 1, "dom", nil, nil)
	addAppWithPackages(ds, 1, 1, "app", nil, nil)

	// Workspace explicitly cleared (stored as sql.NullString with Valid=true, String="")
	ws := newPackageMockWorkspaceFromNullable(1, 1, "ws",
		sql.NullString{String: "", Valid: true},  // cleared
		sql.NullString{String: "", Valid: false}, // not set
	)
	ds.workspaces[1] = ws

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, res)
	// Empty string at workspace level should skip; inherit from ecosystem
	assert.Equal(t, "eco-nvim-pkg", res.PackageName)
	assert.Equal(t, PackageLevelEcosystem, res.Source)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_NvimAndTerminalIndependent
// nvim and terminal packages resolve independently at each level
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_NvimAndTerminalIndependent(t *testing.T) {
	ds := NewPackageMockDataStore()
	// Ecosystem sets nvim; app sets terminal; workspace sets nothing
	addEcosystemWithPackages(ds, 1, "eco", pkgPtr("eco-nvim-pkg"), nil)
	addDomainWithPackages(ds, 1, 1, "dom", nil, nil)
	addAppWithPackages(ds, 1, 1, "app", nil, pkgPtr("app-term-pkg"))
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil)

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	nvimRes, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)
	require.NoError(t, err)
	require.NotNil(t, nvimRes)
	assert.Equal(t, "eco-nvim-pkg", nvimRes.PackageName)
	assert.Equal(t, PackageLevelEcosystem, nvimRes.Source)

	termRes, err := r.ResolveTerminalPackage(ctx, PackageLevelWorkspace, 1)
	require.NoError(t, err)
	require.NotNil(t, termRes)
	assert.Equal(t, "app-term-pkg", termRes.PackageName)
	assert.Equal(t, PackageLevelApp, termRes.Source)
}

// ---------------------------------------------------------------------------
// TestPackageHierarchyResolver_ShowCascade
// resolution path includes all levels walked, with correct Found flags
// ---------------------------------------------------------------------------

func TestPackageHierarchyResolver_ShowCascade(t *testing.T) {
	ds := NewPackageMockDataStore()
	addEcosystemWithPackages(ds, 1, "eco", nil, nil)
	addDomainWithPackages(ds, 1, 1, "dom", pkgPtr("dom-nvim-pkg"), nil)
	addAppWithPackages(ds, 1, 1, "app", nil, nil)
	addWorkspaceWithPackages(ds, 1, 1, "ws", nil, nil)

	r := NewHierarchyPackageResolver(ds)
	ctx := context.Background()

	res, err := r.ResolveNvimPackage(ctx, PackageLevelWorkspace, 1)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Path should record all steps walked
	require.NotEmpty(t, res.Path)

	// Verify each level appears in path with correct Found status
	foundLevels := make(map[PackageHierarchyLevel]bool)
	for _, step := range res.Path {
		foundLevels[step.Level] = step.Found
	}

	assert.False(t, foundLevels[PackageLevelWorkspace], "workspace should not be found")
	assert.False(t, foundLevels[PackageLevelApp], "app should not be found")
	assert.True(t, foundLevels[PackageLevelDomain], "domain should be found")

	// Resolution came from domain
	assert.Equal(t, "dom-nvim-pkg", res.PackageName)
	assert.Equal(t, PackageLevelDomain, res.Source)
}
