package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resolver"
	"log/slog"
)

// resolveNvimPackageFromHierarchy resolves the nvim package for a workspace
// by walking the hierarchy: workspace → app → domain → ecosystem → global default.
// Returns the resolved package name, or empty string if nothing found.
func resolveNvimPackageFromHierarchy(ds db.DataStore, workspace *models.Workspace) string {
	if workspace == nil {
		// No workspace context — fall back to global default
		defaultPkg, err := ds.GetDefault("nvim-package")
		if err == nil && defaultPkg != "" {
			return defaultPkg
		}
		return ""
	}

	adapter := resolver.NewDataStorePackageAdapter(ds)
	pkgResolver := resolver.NewHierarchyPackageResolver(adapter)

	ctx := context.Background()
	resolution, err := pkgResolver.ResolveNvimPackage(ctx, resolver.PackageLevelWorkspace, workspace.ID)
	if err != nil {
		slog.Warn("hierarchy resolution failed for nvim package, falling back to global default",
			"workspace", workspace.Name, "error", err)
		defaultPkg, _ := ds.GetDefault("nvim-package")
		return defaultPkg
	}

	if resolution.PackageName != "" {
		slog.Debug("resolved nvim package from hierarchy",
			"package", resolution.PackageName,
			"source", resolution.Source.String(),
			"sourceName", resolution.SourceName)
		return resolution.PackageName
	}

	return ""
}

// resolveTerminalPackageFromHierarchy resolves the terminal package for a workspace
// by walking the hierarchy: workspace → app → domain → ecosystem → global default.
// Returns the resolved package name, or empty string if nothing found.
func resolveTerminalPackageFromHierarchy(ds db.DataStore, workspace *models.Workspace) string {
	if workspace == nil {
		// No workspace context — fall back to global default
		defaultPkg, err := ds.GetDefault("terminal-package")
		if err == nil && defaultPkg != "" {
			return defaultPkg
		}
		return ""
	}

	adapter := resolver.NewDataStorePackageAdapter(ds)
	pkgResolver := resolver.NewHierarchyPackageResolver(adapter)

	ctx := context.Background()
	resolution, err := pkgResolver.ResolveTerminalPackage(ctx, resolver.PackageLevelWorkspace, workspace.ID)
	if err != nil {
		slog.Warn("hierarchy resolution failed for terminal package, falling back to global default",
			"workspace", workspace.Name, "error", err)
		defaultPkg, _ := ds.GetDefault("terminal-package")
		return defaultPkg
	}

	if resolution.PackageName != "" {
		slog.Debug("resolved terminal package from hierarchy",
			"package", resolution.PackageName,
			"source", resolution.Source.String(),
			"sourceName", resolution.SourceName)
		return resolution.PackageName
	}

	return ""
}
