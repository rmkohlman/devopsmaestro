// Package cmd provides shared helpers for the set nvim-package and set
// terminal-package commands. The types and rendering logic live here to
// avoid duplication between the two command files.
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/pkg/resolver"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// packageSetResult holds the outcome of setting a package at a hierarchy level.
type packageSetResult struct {
	Level           string `yaml:"level" json:"level"`
	ObjectName      string `yaml:"objectName" json:"objectName"`
	Package         string `yaml:"package" json:"package"`
	PreviousPackage string `yaml:"previousPackage,omitempty" json:"previousPackage,omitempty"`
	PackageType     string `yaml:"packageType" json:"packageType"` // "nvim" or "terminal"
}

// nullStringValue returns the value of a sql.NullString or "" if not valid.
func nullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// renderPackageSetResult outputs the result of a set package command.
func renderPackageSetResult(result *packageSetResult, outputFmt string, showCascade bool, pkgType string) error {
	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Level", Value: result.Level},
		render.KeyValue{Key: "Object", Value: result.ObjectName},
		render.KeyValue{Key: "Package", Value: result.Package},
	)
	if result.PreviousPackage != "" {
		kvData.Pairs = append(kvData.Pairs, render.KeyValue{Key: "Previous Package", Value: result.PreviousPackage})
	}

	opts := render.Options{
		Type:  render.TypeKeyValue,
		Title: fmt.Sprintf("%s Package Set: %s", strings.Title(pkgType), result.Level),
	}

	if err := render.OutputWith(outputFmt, kvData, opts); err != nil {
		return err
	}

	if showCascade {
		render.Blank()
		render.Info(fmt.Sprintf("Package type: %s", pkgType))
		render.Info(fmt.Sprintf("Cascade: global → ecosystem → domain → app → workspace"))
	}

	return nil
}

// resolvePackageCascade resolves the effective package from the hierarchy.
// This is used by the get commands to display the resolved package.
func resolvePackageCascade(ctx resource.Context, pkgType string, level resolver.PackageHierarchyLevel, objectID int) (*resolver.PackageResolution, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}

	adapter := resolver.NewDataStorePackageAdapter(ds)
	r := resolver.NewHierarchyPackageResolver(adapter)

	bgCtx := context.Background()
	if pkgType == "nvim" {
		return r.ResolveNvimPackage(bgCtx, level, objectID)
	}
	return r.ResolveTerminalPackage(bgCtx, level, objectID)
}

// resolveCurrentWorkspaceForPackage returns the current workspace name,
// the starting resolution level, and the object ID for hierarchy resolution.
// It resolves the active workspace from the command context.
func resolveCurrentWorkspaceForPackage(ctx resource.Context) (string, resolver.PackageHierarchyLevel, int, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to get DataStore: %w", err)
	}

	dbCtx, err := ds.GetContext()
	if err != nil || dbCtx == nil || dbCtx.ActiveWorkspaceID == nil {
		return "", 0, 0, fmt.Errorf("no active workspace context (use 'dvm use workspace <name>')")
	}

	ws, err := ds.GetWorkspaceByID(*dbCtx.ActiveWorkspaceID)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to get active workspace: %w", err)
	}

	return ws.Name, resolver.PackageLevelWorkspace, ws.ID, nil
}

// renderPackageResolution renders a package resolution result for get commands.
func renderPackageResolution(res *resolver.PackageResolution, wsName string, showCascade bool, outputFmt string) error {
	pkgName := res.PackageName
	if pkgName == "" {
		pkgName = "(none)"
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Workspace", Value: wsName},
		render.KeyValue{Key: "Package Type", Value: res.PackageType},
		render.KeyValue{Key: "Package", Value: pkgName},
		render.KeyValue{Key: "Source", Value: res.Source.String()},
	)
	if res.SourceName != "" {
		kvData.Pairs = append(kvData.Pairs, render.KeyValue{Key: "Source Name", Value: res.SourceName})
	}

	opts := render.Options{
		Type:  render.TypeKeyValue,
		Title: fmt.Sprintf("Resolved %s Package", strings.Title(res.PackageType)),
	}

	if err := render.OutputWith(outputFmt, kvData, opts); err != nil {
		return err
	}

	if showCascade && len(res.Path) > 0 {
		render.Blank()
		render.Info("Resolution path:")
		for _, step := range res.Path {
			status := "○"
			if step.Found {
				status = "●"
			}
			line := fmt.Sprintf("  %s %s '%s'", status, step.Level.String(), step.Name)
			if step.PackageName != "" {
				line += fmt.Sprintf(" → %s", step.PackageName)
			}
			if step.Error != "" {
				line += fmt.Sprintf(" (error: %s)", step.Error)
			}
			render.Plain(line)
		}
		render.Blank()
		render.Info("Legend: ● package set, ○ no package (inherits from parent)")
	}

	return nil
}
