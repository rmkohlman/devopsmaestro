package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalbridge"
	"github.com/rmkohlman/MaestroSDK/render"
	terminalpackage "github.com/rmkohlman/MaestroTerminal/terminalops/package"
	packagelibrary "github.com/rmkohlman/MaestroTerminal/terminalops/package/library"
	pluginlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/plugin/library"
	promptlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/library"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// packageCmd is the main package command
var packageCmd = &cobra.Command{
	Use:     "package",
	Aliases: []string{"pkg"},
	Short:   "Manage terminal packages",
	Long: `Terminal packages allow grouping related terminal configuration into reusable bundles with inheritance.
A package like 'developer' might extend 'core' and include development-specific plugins and prompts.

Packages can contain:
  - Shell plugins (zsh-autosuggestions, fzf, etc.)
  - Prompts (Starship, P10k configurations)
  - Profiles (combined terminal setups)
  - WezTerm configuration

Use these commands to explore and install packages from the library.

Examples:
  dvt package get                     # List all packages
  dvt package get developer           # Show package details
  dvt package install developer       # Install package components`,
}

// packageGetCmd shows package details, or lists all packages when no name is given
var packageGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Show package details (or list all packages)",
	Long: `Show details of a specific package, or list all packages when no name is given.

Examples:
  dvt package get                         # List all packages
  dvt package get --category development  # Filter by category
  dvt package get -o yaml                 # YAML output
  dvt package get -w                      # Wide format with more details
  dvt package get core                    # Show specific package
  dvt package get developer -o yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := packagelibrary.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load package library: %w", err)
		}

		if len(args) == 0 {
			// List mode
			packages := lib.List()

			// Filter by library flag (for now, all packages are library packages)
			libraryOnly, _ := cmd.Flags().GetBool("library")
			userOnly, _ := cmd.Flags().GetBool("user")

			if userOnly {
				// For now, no user packages - return empty
				packages = []*terminalpackage.Package{}
			} else if !libraryOnly {
				// Default: show all (which are all library packages for now)
			}

			// Filter by category if specified
			category, _ := cmd.Flags().GetString("category")
			if category != "" {
				packages = lib.ListByCategory(category)
			}

			if len(packages) == 0 {
				render.Info("No packages found")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			wide, _ := cmd.Flags().GetBool("wide")
			return outputPackages(packages, format, wide, lib)
		}

		// Single get mode
		name := args[0]
		pkg, ok := lib.Get(name)
		if !ok {
			return fmt.Errorf("package not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPackageDetails(pkg, format, lib)
	},
}

// packageInstallCmd installs a package
var packageInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a package and its components to the database",
	Long: `Install a package by resolving inheritance and storing components in the database.
This resolves inheritance, so installing 'developer' will also install all 'core' components.

Examples:
  dvt package install core
  dvt package install developer --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := packagelibrary.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load package library: %w", err)
		}

		pkg, ok := lib.Get(name)
		if !ok {
			return fmt.Errorf("package not found: %s", name)
		}

		// Resolve all components from inheritance
		components, err := resolvePackageComponents(pkg, lib)
		if err != nil {
			return fmt.Errorf("failed to resolve package components: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Calculate totals
		totalComponents := len(components.Plugins) + len(components.Prompts) + len(components.Profiles)

		render.Progressf("Installing package: %s", name)
		render.Blank()

		if totalComponents == 0 {
			render.Info("No components to install.")
			return nil
		}

		if dryRun {
			// Show preview for dry run
			if len(components.Plugins) > 0 {
				render.Infof("Plugins to install (%d):", len(components.Plugins))
				for _, pluginName := range components.Plugins {
					source := getComponentSource(pluginName, pkg, lib, "plugin")
					if source != "" {
						render.Plainf("  - %s (from %s)", pluginName, source)
					} else {
						render.Plainf("  - %s", pluginName)
					}
				}
				render.Blank()
			}

			if len(components.Prompts) > 0 {
				render.Infof("Prompts to install (%d):", len(components.Prompts))
				for _, promptName := range components.Prompts {
					source := getComponentSource(promptName, pkg, lib, "prompt")
					if source != "" {
						render.Plainf("  - %s (from %s)", promptName, source)
					} else {
						render.Plainf("  - %s", promptName)
					}
				}
				render.Blank()
			}

			if len(components.Profiles) > 0 {
				render.Infof("Profiles to install (%d):", len(components.Profiles))
				for _, profileName := range components.Profiles {
					source := getComponentSource(profileName, pkg, lib, "profile")
					if source != "" {
						render.Plainf("  - %s (from %s)", profileName, source)
					} else {
						render.Plainf("  - %s", profileName)
					}
				}
				render.Blank()
			}

			render.Infof("Use 'dvt package install %s' without --dry-run to install.", name)
			return nil
		}

		// Get DataStore from context (follows dvt pattern)
		dataStoreInterface := cmd.Context().Value("dataStore")
		if dataStoreInterface == nil {
			return fmt.Errorf("database not initialized")
		}
		dataStore := dataStoreInterface.(*db.DataStore)

		// Create stores
		pluginStore := terminalbridge.NewDBPluginStore(*dataStore)
		defer pluginStore.Close()

		promptStore := terminalbridge.NewDBPromptStore(*dataStore)
		defer promptStore.Close()

		profileStore := terminalbridge.NewDBProfileStore(*dataStore)
		defer profileStore.Close()

		// Load libraries
		pluginLib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load plugin library: %w", err)
		}

		promptLib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load prompt library: %w", err)
		}

		// Install components
		var installed, skipped, failed []string
		var promptsInstalled, promptsSkipped, promptsFailed []string
		var profilesFailed []string

		if len(components.Plugins) > 0 {
			render.Progressf("Installing %d plugins...", len(components.Plugins))
			for _, pluginName := range components.Plugins {
				// Get plugin from library
				plugin, err := pluginLib.Get(pluginName)
				if err != nil {
					render.Warningf("Plugin '%s' not found in library, skipping", pluginName)
					failed = append(failed, pluginName)
					continue
				}

				// Check if already exists
				if exists, _ := pluginStore.Exists(pluginName); exists {
					render.Infof("Plugin '%s' already installed", pluginName)
					skipped = append(skipped, pluginName)
					continue
				}

				// Install plugin using Upsert (safe for create/update)
				if err := pluginStore.Upsert(plugin); err != nil {
					render.Errorf("Failed to install '%s': %v", pluginName, err)
					failed = append(failed, pluginName)
				} else {
					render.Successf("Installed '%s'", pluginName)
					installed = append(installed, pluginName)
				}
			}
		}

		// Install prompts
		if len(components.Prompts) > 0 {
			render.Blank()
			render.Progressf("Installing %d prompts...", len(components.Prompts))
			for _, promptName := range components.Prompts {
				prompt, err := promptLib.Get(promptName)
				if err != nil {
					render.Warningf("Prompt '%s' not found in library, skipping", promptName)
					promptsFailed = append(promptsFailed, promptName)
					continue
				}

				// Check if already exists
				if exists, _ := promptStore.Exists(promptName); exists {
					render.Infof("Prompt '%s' already installed", promptName)
					promptsSkipped = append(promptsSkipped, promptName)
					continue
				}

				// Install prompt using Upsert (safe for create/update)
				if err := promptStore.Upsert(prompt); err != nil {
					render.Errorf("Failed to install '%s': %v", promptName, err)
					promptsFailed = append(promptsFailed, promptName)
				} else {
					render.Successf("Installed '%s'", promptName)
					promptsInstalled = append(promptsInstalled, promptName)
				}
			}
		}

		// Install profiles (note: profiles typically reference prompts/plugins, so no inline library yet)
		if len(components.Profiles) > 0 {
			render.Blank()
			render.Warningf("Profiles to install (%d) - profile library not yet implemented:", len(components.Profiles))
			for _, profileName := range components.Profiles {
				render.Plainf("  - %s (pending profile library)", profileName)
				profilesFailed = append(profilesFailed, profileName)
			}
		}

		// Store the package itself in terminal_packages table
		pkgDB := &models.TerminalPackageDB{
			Name: pkg.Name,
			Description: sql.NullString{
				String: pkg.Description,
				Valid:  pkg.Description != "",
			},
			Category: sql.NullString{
				String: pkg.Category,
				Valid:  pkg.Category != "",
			},
			Extends: sql.NullString{
				String: pkg.Extends,
				Valid:  pkg.Extends != "",
			},
		}
		// Set the resolved components (what was actually installed)
		pkgDB.SetPlugins(components.Plugins)
		pkgDB.SetPrompts(components.Prompts)
		pkgDB.SetProfiles(components.Profiles)

		if err := (*dataStore).UpsertTerminalPackage(pkgDB); err != nil {
			// Log warning but don't fail - components were installed
			render.Warningf("Failed to save package record: %v", err)
		}

		// Print summary
		render.Blank()
		render.Successf("Package '%s' installation complete:", name)
		render.Plainf("  Plugins installed: %d", len(installed))
		if len(promptsInstalled) > 0 {
			render.Plainf("  Prompts installed: %d", len(promptsInstalled))
		}
		if len(skipped) > 0 {
			render.Plainf("  Plugins skipped:   %d (already installed)", len(skipped))
		}
		if len(promptsSkipped) > 0 {
			render.Plainf("  Prompts skipped:   %d (already installed)", len(promptsSkipped))
		}
		if len(failed) > 0 {
			render.Plainf("  Plugins failed:    %d (not found in library)", len(failed))
		}
		if len(promptsFailed) > 0 {
			render.Plainf("  Prompts failed:    %d (not found in library)", len(promptsFailed))
		}
		if len(profilesFailed) > 0 {
			render.Plainf("  Profiles failed:   %d (library not implemented)", len(profilesFailed))
		}

		// Only fail if nothing was installed and there were failures, but no skipped items
		totalInstalled := len(installed) + len(promptsInstalled)
		totalSkipped := len(skipped) + len(promptsSkipped)
		totalFailed := len(failed) + len(promptsFailed) + len(profilesFailed)

		if totalInstalled == 0 && totalFailed > 0 && totalSkipped == 0 {
			return fmt.Errorf("no components could be installed")
		}

		render.Blank()
		render.Info("Use 'dvt plugin get' and 'dvt prompt get' to see installed components")

		return nil
	},
}

// ResolvedComponents holds all resolved items from a package
type ResolvedComponents struct {
	Plugins  []string
	Prompts  []string
	Profiles []string
}

// resolvePackageComponents resolves all components from a package including inheritance
func resolvePackageComponents(pkg *terminalpackage.Package, lib *packagelibrary.Library) (*ResolvedComponents, error) {
	result := &ResolvedComponents{
		Plugins:  []string{},
		Prompts:  []string{},
		Profiles: []string{},
	}
	visited := make(map[string]bool)

	var resolve func(p *terminalpackage.Package) error
	resolve = func(p *terminalpackage.Package) error {
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

		// Add this package's components
		for _, pluginName := range p.Plugins {
			if !contains(result.Plugins, pluginName) {
				result.Plugins = append(result.Plugins, pluginName)
			}
		}

		for _, promptName := range p.Prompts {
			if !contains(result.Prompts, promptName) {
				result.Prompts = append(result.Prompts, promptName)
			}
		}

		for _, profileName := range p.Profiles {
			if !contains(result.Profiles, profileName) {
				result.Profiles = append(result.Profiles, profileName)
			}
		}

		return nil
	}

	err := resolve(pkg)
	return result, err
}

// getComponentSource finds which package in the hierarchy provides a component
func getComponentSource(componentName string, pkg *terminalpackage.Package, lib *packagelibrary.Library, componentType string) string {
	// Check current package first
	switch componentType {
	case "plugin":
		for _, name := range pkg.Plugins {
			if name == componentName {
				return pkg.Name
			}
		}
	case "prompt":
		for _, name := range pkg.Prompts {
			if name == componentName {
				return pkg.Name
			}
		}
	case "profile":
		for _, name := range pkg.Profiles {
			if name == componentName {
				return pkg.Name
			}
		}
	}

	// Check parent packages
	if pkg.Extends != "" {
		parent, ok := lib.Get(pkg.Extends)
		if ok {
			return getComponentSource(componentName, parent, lib, componentType)
		}
	}

	return ""
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

// outputPackages outputs packages in the specified format
func outputPackages(packages []*terminalpackage.Package, format string, wide bool, lib *packagelibrary.Library) error {
	// Sort by name
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range packages {
			if i > 0 {
				fmt.Println("---")
			}
			yml := p.ToYAML()
			data, err := yaml.Marshal(yml)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*terminalpackage.PackageYAML
		for _, p := range packages {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		var tb *render.TableBuilder
		if wide {
			tb = render.NewTableBuilder("NAME", "TYPE", "CATEGORY", "PLUGINS", "PROMPTS", "PROFILES", "EXTENDS", "DESCRIPTION")
		} else {
			tb = render.NewTableBuilder("NAME", "TYPE", "PLUGINS", "CATEGORY", "DESCRIPTION")
		}

		for _, p := range packages {
			// Resolve component count including inheritance
			allComponents, err := resolvePackageComponents(p, lib)
			pluginCount := len(p.Plugins)
			promptCount := len(p.Prompts)
			profileCount := len(p.Profiles)
			if err == nil {
				pluginCount = len(allComponents.Plugins)
				promptCount = len(allComponents.Prompts)
				profileCount = len(allComponents.Profiles)
			}

			desc := render.Truncate(p.Description, 40)

			if wide {
				tb.AddRow(p.Name, "library", p.Category,
					fmt.Sprintf("%d", pluginCount), fmt.Sprintf("%d", promptCount),
					fmt.Sprintf("%d", profileCount), p.Extends, desc)
			} else {
				tb.AddRow(p.Name, "library", fmt.Sprintf("%d", pluginCount), p.Category, desc)
			}
		}
		return render.OutputWith("", tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// outputPackageDetails outputs detailed package information
func outputPackageDetails(pkg *terminalpackage.Package, format string, lib *packagelibrary.Library) error {
	switch format {
	case "yaml", "":
		// Show resolved package with all components
		resolved, err := createResolvedPackageYAML(pkg, lib)
		if err != nil {
			return fmt.Errorf("failed to resolve package: %w", err)
		}

		data, err := yaml.Marshal(resolved)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		resolved, err := createResolvedPackageYAML(pkg, lib)
		if err != nil {
			return fmt.Errorf("failed to resolve package: %w", err)
		}

		data, err := json.MarshalIndent(resolved, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// createResolvedPackageYAML creates a package YAML with all resolved components
func createResolvedPackageYAML(pkg *terminalpackage.Package, lib *packagelibrary.Library) (*terminalpackage.PackageYAML, error) {
	// Resolve all components
	allComponents, err := resolvePackageComponents(pkg, lib)
	if err != nil {
		return nil, err
	}

	// Create a copy with resolved components
	resolved := &terminalpackage.Package{
		Name:        pkg.Name,
		Description: pkg.Description + " (resolved)",
		Category:    pkg.Category,
		Tags:        pkg.Tags,
		Extends:     pkg.Extends,
		Plugins:     allComponents.Plugins,
		Prompts:     allComponents.Prompts,
		Profiles:    allComponents.Profiles,
		WezTerm:     pkg.WezTerm,
		Enabled:     pkg.Enabled,
		CreatedAt:   pkg.CreatedAt,
		UpdatedAt:   pkg.UpdatedAt,
	}

	yml := resolved.ToYAML()

	// Add a comment about resolution
	totalComponents := len(allComponents.Plugins) + len(allComponents.Prompts) + len(allComponents.Profiles)
	if pkg.Extends != "" {
		yml.Metadata.Description = fmt.Sprintf("%s (includes %d components from inheritance)",
			strings.TrimSuffix(pkg.Description, " (resolved)"), totalComponents)
	}

	return yml, nil
}

func init() {
	// Add subcommands
	packageCmd.AddCommand(packageGetCmd)
	packageCmd.AddCommand(packageInstallCmd)

	// Package get flags (merged from list + get)
	packageGetCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	packageGetCmd.Flags().Bool("library", false, "Show only library packages")
	packageGetCmd.Flags().Bool("user", false, "Show only user packages")
	packageGetCmd.Flags().StringP("category", "c", "", "Filter by category")
	packageGetCmd.Flags().BoolP("wide", "w", false, "Show extended output")

	// Package install flags
	packageInstallCmd.Flags().Bool("dry-run", false, "Show what would be installed without installing")

	// Hidden backward-compat alias for deprecated verb in package (after flags)
	packageCmd.AddCommand(hiddenAlias("list", packageGetCmd))
}
