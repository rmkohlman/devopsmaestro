package cmd

import (
	"devopsmaestro/db"
	nvimpackage "devopsmaestro/pkg/nvimops/package"
	packagelibrary "devopsmaestro/pkg/nvimops/package/library"
	"fmt"
	"strings"
)

// resolveDefaultPackagePlugins resolves plugins from a default package name.
// It first checks the embedded library, then falls back to database packages.
func resolveDefaultPackagePlugins(packageName string, ds db.DataStore) ([]string, error) {
	// First, try to load from embedded library
	lib, err := packagelibrary.NewLibrary()
	if err != nil {
		return nil, fmt.Errorf("failed to create package library: %w", err)
	}

	if pkg, ok := lib.Get(packageName); ok {
		// Package found in library - resolve plugins including inheritance
		return resolvePackagePlugins(pkg, lib)
	}

	// Package not in library - try database
	dbPkg, err := ds.GetPackage(packageName)
	if err != nil {
		return nil, fmt.Errorf("package '%s' not found in library or database: %w", packageName, err)
	}

	// Convert database model to package model
	pkg := &nvimpackage.Package{
		Name:        dbPkg.Name,
		Description: dbPkg.Description.String,
		Category:    dbPkg.Category.String,
		Tags:        []string{}, // Database packages don't have tags in current schema
		Extends:     dbPkg.Extends.String,
		Plugins:     dbPkg.GetPlugins(),
		Enabled:     true, // Database packages are enabled by default
	}

	// Clean up plugins (they come from JSON so should already be clean, but just in case)
	var cleanPlugins []string
	for _, plugin := range pkg.Plugins {
		plugin = strings.TrimSpace(plugin)
		if plugin != "" {
			cleanPlugins = append(cleanPlugins, plugin)
		}
	}
	pkg.Plugins = cleanPlugins

	// For database packages, we need to handle inheritance manually
	// since we can't use the library's resolution logic
	if pkg.Extends != "" {
		// Try to resolve parent from library first
		if parentPkg, ok := lib.Get(pkg.Extends); ok {
			parentPlugins, err := resolvePackagePlugins(parentPkg, lib)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve parent package '%s' from library: %w", pkg.Extends, err)
			}
			// Combine parent plugins with current package plugins
			allPlugins := append(parentPlugins, pkg.Plugins...)
			return removeDuplicates(allPlugins), nil
		}

		// Parent not in library - try database
		parentDBPkg, err := ds.GetPackage(pkg.Extends)
		if err != nil {
			return nil, fmt.Errorf("parent package '%s' not found in library or database: %w", pkg.Extends, err)
		}

		// Simple inheritance for database packages (no deep recursion to avoid complexity)
		parentPlugins := parentDBPkg.GetPlugins()

		// Combine parent and current plugins
		allPlugins := append(parentPlugins, pkg.Plugins...)
		return removeDuplicates(allPlugins), nil
	}

	// No inheritance - return current package plugins
	return pkg.Plugins, nil
}

// resolvePackagePlugins resolves all plugins from a package including inheritance.
// This is based on the same function in cmd/nvp/package.go.
func resolvePackagePlugins(pkg *nvimpackage.Package, lib *packagelibrary.Library) ([]string, error) {
	var result []string
	visited := make(map[string]bool)

	var resolve func(p *nvimpackage.Package) error
	resolve = func(p *nvimpackage.Package) error {
		if visited[p.Name] {
			return fmt.Errorf("circular dependency detected: %s", p.Name)
		}
		visited[p.Name] = true
		defer func() { visited[p.Name] = false }()

		// If this package extends another, resolve parent first
		if p.Extends != "" {
			parent, ok := lib.Get(p.Extends)
			if !ok {
				return fmt.Errorf("package %s extends %s, but %s not found in library", p.Name, p.Extends, p.Extends)
			}
			if err := resolve(parent); err != nil {
				return err
			}
		}

		// Add this package's plugins
		for _, pluginName := range p.Plugins {
			if !contains(result, pluginName) {
				result = append(result, pluginName)
			}
		}

		return nil
	}

	err := resolve(pkg)
	return result, err
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// removeDuplicates removes duplicate strings from a slice while preserving order
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
