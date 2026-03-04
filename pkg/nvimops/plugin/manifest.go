// Package plugin provides types and utilities for Neovim plugin management.
package plugin

import "strings"

// PluginManifest tracks what plugins are resolved for a build.
// This is used to conditionally enable features in generated configurations.
type PluginManifest struct {
	// InstalledPlugins is the list of plugin names that will be installed
	InstalledPlugins []string

	// Features tracks which plugin features are available
	Features PluginFeatures
}

// PluginFeatures tracks which major plugin features are installed.
// Used to conditionally enable build-time optimizations.
type PluginFeatures struct {
	// HasMason indicates if mason.nvim (LSP installer) is present
	HasMason bool

	// HasTreesitter indicates if nvim-treesitter is present
	HasTreesitter bool

	// HasTelescope indicates if telescope.nvim (fuzzy finder) is present
	HasTelescope bool

	// HasLSPConfig indicates if nvim-lspconfig is present
	HasLSPConfig bool
}

// ResolveManifest analyzes resolved plugins and creates a manifest.
// It detects which major plugins are installed by checking plugin names and repos.
func ResolveManifest(plugins []*Plugin) *PluginManifest {
	manifest := &PluginManifest{
		InstalledPlugins: make([]string, 0, len(plugins)),
	}

	for _, p := range plugins {
		if !p.Enabled {
			continue
		}

		manifest.InstalledPlugins = append(manifest.InstalledPlugins, p.Name)

		// Check repo and name for known plugins
		repo := strings.ToLower(p.Repo)
		name := strings.ToLower(p.Name)

		if strings.Contains(repo, "mason") || strings.Contains(name, "mason") {
			manifest.Features.HasMason = true
		}

		if strings.Contains(repo, "treesitter") || strings.Contains(name, "treesitter") {
			manifest.Features.HasTreesitter = true
		}

		if strings.Contains(repo, "telescope") || strings.Contains(name, "telescope") {
			manifest.Features.HasTelescope = true
		}

		if strings.Contains(repo, "lspconfig") || strings.Contains(name, "lspconfig") {
			manifest.Features.HasLSPConfig = true
		}
	}

	return manifest
}

// ResolveManifestFromNames creates a manifest from a list of plugin names.
// This is a lightweight version used when only plugin names are available.
func ResolveManifestFromNames(pluginNames []string) *PluginManifest {
	manifest := &PluginManifest{
		InstalledPlugins: pluginNames,
	}

	for _, name := range pluginNames {
		nameLower := strings.ToLower(name)

		if strings.Contains(nameLower, "mason") {
			manifest.Features.HasMason = true
		}

		if strings.Contains(nameLower, "treesitter") {
			manifest.Features.HasTreesitter = true
		}

		if strings.Contains(nameLower, "telescope") {
			manifest.Features.HasTelescope = true
		}

		if strings.Contains(nameLower, "lspconfig") {
			manifest.Features.HasLSPConfig = true
		}
	}

	return manifest
}
