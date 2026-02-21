package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/library"
	nvimpackage "devopsmaestro/pkg/nvimops/package"
	packagelibrary "devopsmaestro/pkg/nvimops/package/library"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// packageCmd is the main package command
var packageCmd = &cobra.Command{
	Use:     "package",
	Aliases: []string{"pkg"},
	Short:   "Manage Neovim plugin packages",
	Long: `Plugin packages allow grouping related plugins into reusable bundles with inheritance.
A package like 'go-dev' might extend 'core' and include Go-specific development tools.

Use these commands to explore and install packages from the library.`,
}

// packageListCmd lists all packages
var packageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available packages",
	Long: `List packages from the library.

Examples:
  nvp package list                     # List all packages
  nvp package list --category language # Filter by category  
  nvp package list -o yaml            # YAML output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := packagelibrary.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load package library: %w", err)
		}

		packages := lib.List()

		// Filter by library flag (for now, all packages are library packages)
		libraryOnly, _ := cmd.Flags().GetBool("library")
		userOnly, _ := cmd.Flags().GetBool("user")

		if userOnly {
			// For now, no user packages - return empty
			packages = []*nvimpackage.Package{}
		} else if !libraryOnly {
			// Default: show all (which are all library packages for now)
		}

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			packages = lib.ListByCategory(category)
		}

		if len(packages) == 0 {
			fmt.Println("No packages found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		wide, _ := cmd.Flags().GetBool("wide")
		return outputPackages(packages, format, wide, lib)
	},
}

// packageGetCmd shows package details
var packageGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Show package details",
	Long: `Show details of a specific package, including all plugins from inheritance.

Examples:
  nvp package get core
  nvp package get go-dev -o yaml`,
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

		format, _ := cmd.Flags().GetString("output")
		return outputPackageDetails(pkg, format, lib)
	},
}

// packageInstallCmd installs a package
var packageInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a package (adds all its plugins)",
	Long: `Install a package by adding all its plugins to your local store.
This resolves inheritance, so installing 'go-dev' will also install all 'core' plugins.

Examples:
  nvp package install core
  nvp package install go-dev --dry-run`,
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

		// Resolve all plugins from inheritance
		pluginNames, err := resolvePackagePlugins(pkg, lib)
		if err != nil {
			return fmt.Errorf("failed to resolve package plugins: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			fmt.Printf("Would install %d plugins from package '%s':\n", len(pluginNames), name)
			for _, pluginName := range pluginNames {
				fmt.Printf("  - %s\n", pluginName)
			}
			return nil
		}

		// Load plugin library
		pluginLib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load plugin library: %w", err)
		}

		// Get manager for plugin storage
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		// Get dataStore from context like dvt does
		dataStoreInterface := cmd.Context().Value("dataStore")
		if dataStoreInterface == nil {
			return fmt.Errorf("dataStore not found in context")
		}
		dataStore := dataStoreInterface.(*db.DataStore)

		// Install each plugin
		var installed, failed []string
		for _, pluginName := range pluginNames {
			// Get plugin from library
			plugin, ok := pluginLib.Get(pluginName)
			if !ok {
				fmt.Printf("⚠ Plugin '%s' not found in library, skipping\n", pluginName)
				failed = append(failed, pluginName)
				continue
			}

			// Check if already exists
			if _, err := mgr.Get(pluginName); err == nil {
				fmt.Printf("• Plugin '%s' already installed\n", pluginName)
				continue
			}

			// Install plugin to file store
			if err := mgr.Apply(plugin); err != nil {
				fmt.Printf("✗ Failed to install '%s': %v\n", pluginName, err)
				failed = append(failed, pluginName)
			} else {
				fmt.Printf("✓ Installed '%s'\n", pluginName)
				installed = append(installed, pluginName)

				// Also save to database for dvm compatibility
				pluginDB := &models.NvimPluginDB{}
				if err := pluginDB.FromNvimOpsPlugin(plugin); err != nil {
					fmt.Printf("⚠ Warning: Failed to convert plugin '%s' for database: %v\n", pluginName, err)
				} else if err := (*dataStore).UpsertPlugin(pluginDB); err != nil {
					fmt.Printf("⚠ Warning: Failed to save plugin '%s' to database: %v\n", pluginName, err)
				}
			}
		}

		// Summary
		fmt.Printf("\nPackage '%s' installation complete:\n", name)
		fmt.Printf("  Installed: %d\n", len(installed))
		if len(failed) > 0 {
			fmt.Printf("  Failed:    %d\n", len(failed))
			return fmt.Errorf("some plugins failed to install")
		}

		return nil
	},
}

// resolvePackagePlugins resolves all plugins from a package including inheritance
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

// outputPackages outputs packages in the specified format
func outputPackages(packages []*nvimpackage.Package, format string, wide bool, lib *packagelibrary.Library) error {
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
		var items []*nvimpackage.PackageYAML
		for _, p := range packages {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if wide {
			fmt.Fprintln(w, "NAME\tTYPE\tCATEGORY\tPLUGINS\tEXTENDS\tDESCRIPTION")
		} else {
			fmt.Fprintln(w, "NAME\tTYPE\tPLUGINS\tCATEGORY\tDESCRIPTION")
		}

		for _, p := range packages {
			// Resolve plugin count including inheritance
			allPlugins, err := resolvePackagePlugins(p, lib)
			pluginCount := len(p.Plugins)
			if err == nil {
				pluginCount = len(allPlugins)
			}

			desc := p.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}

			if wide {
				fmt.Fprintf(w, "%s\tlibrary\t%s\t%d\t%s\t%s\n",
					p.Name, p.Category, pluginCount, p.Extends, desc)
			} else {
				fmt.Fprintf(w, "%s\tlibrary\t%d\t%s\t%s\n",
					p.Name, pluginCount, p.Category, desc)
			}
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// outputPackageDetails outputs detailed package information
func outputPackageDetails(pkg *nvimpackage.Package, format string, lib *packagelibrary.Library) error {
	switch format {
	case "yaml", "":
		// Show resolved package with all plugins
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

// createResolvedPackageYAML creates a package YAML with all resolved plugins
func createResolvedPackageYAML(pkg *nvimpackage.Package, lib *packagelibrary.Library) (*nvimpackage.PackageYAML, error) {
	// Resolve all plugins
	allPlugins, err := resolvePackagePlugins(pkg, lib)
	if err != nil {
		return nil, err
	}

	// Create a copy with resolved plugins
	resolved := &nvimpackage.Package{
		Name:        pkg.Name,
		Description: pkg.Description + " (resolved)",
		Category:    pkg.Category,
		Tags:        pkg.Tags,
		Extends:     pkg.Extends,
		Plugins:     allPlugins,
		Enabled:     pkg.Enabled,
		CreatedAt:   pkg.CreatedAt,
		UpdatedAt:   pkg.UpdatedAt,
	}

	yml := resolved.ToYAML()

	// Add a comment about resolution
	if pkg.Extends != "" {
		yml.Metadata.Description = fmt.Sprintf("%s (includes %d plugins from inheritance)",
			strings.TrimSuffix(pkg.Description, " (resolved)"), len(allPlugins))
	}

	return yml, nil
}

func init() {
	// Add subcommands
	packageCmd.AddCommand(packageListCmd)
	packageCmd.AddCommand(packageGetCmd)
	packageCmd.AddCommand(packageInstallCmd)

	// Package list flags
	packageListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	packageListCmd.Flags().Bool("library", false, "Show only library packages")
	packageListCmd.Flags().Bool("user", false, "Show only user packages")
	packageListCmd.Flags().StringP("category", "c", "", "Filter by category")
	packageListCmd.Flags().BoolP("wide", "w", false, "Show extended output")

	// Package get flags
	packageGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")

	// Package install flags
	packageInstallCmd.Flags().Bool("dry-run", false, "Show what would be installed without installing")
}
