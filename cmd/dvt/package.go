package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	terminalpackage "devopsmaestro/pkg/terminalops/package"
	packagelibrary "devopsmaestro/pkg/terminalops/package/library"

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
  dvt package list                    # List all packages
  dvt package get developer           # Show package details
  dvt package install developer       # Install package components`,
}

// packageListCmd lists all packages
var packageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available packages",
	Long: `List packages from the library.

Examples:
  dvt package list                     # List all packages
  dvt package list --category development # Filter by category  
  dvt package list -o yaml             # YAML output
  dvt package list -w                  # Wide format with more details`,
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
	Long: `Show details of a specific package, including all components from inheritance.

Examples:
  dvt package get core
  dvt package get developer -o yaml`,
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
	Short: "Install a package (shows all its components)",
	Long: `Install a package by showing all its components from inheritance resolution.
This resolves inheritance, so installing 'developer' will also show all 'core' components.

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

		fmt.Printf("Installing package: %s\n\n", name)

		if len(components.Plugins) > 0 {
			fmt.Printf("Plugins to install (%d):\n", len(components.Plugins))
			for _, pluginName := range components.Plugins {
				// Get source package for this plugin
				source := getComponentSource(pluginName, pkg, lib, "plugin")
				if source != "" {
					fmt.Printf("  - %s (from %s)\n", pluginName, source)
				} else {
					fmt.Printf("  - %s\n", pluginName)
				}
			}
			fmt.Println()
		}

		if len(components.Prompts) > 0 {
			fmt.Printf("Prompts to install (%d):\n", len(components.Prompts))
			for _, promptName := range components.Prompts {
				source := getComponentSource(promptName, pkg, lib, "prompt")
				if source != "" {
					fmt.Printf("  - %s (from %s)\n", promptName, source)
				} else {
					fmt.Printf("  - %s\n", promptName)
				}
			}
			fmt.Println()
		}

		if len(components.Profiles) > 0 {
			fmt.Printf("Profiles to install (%d):\n", len(components.Profiles))
			for _, profileName := range components.Profiles {
				source := getComponentSource(profileName, pkg, lib, "profile")
				if source != "" {
					fmt.Printf("  - %s (from %s)\n", profileName, source)
				} else {
					fmt.Printf("  - %s\n", profileName)
				}
			}
			fmt.Println()
		}

		if totalComponents == 0 {
			fmt.Println("No components to install.")
			return nil
		}

		if dryRun {
			fmt.Printf("Use 'dvt package install %s' without --dry-run to install.\n", name)
			return nil
		}

		// For now, just show what would be installed
		// In the future, we can add actual installation logic here
		fmt.Printf("Package '%s' installation complete - %d components shown.\n", name, totalComponents)
		fmt.Println("\nNote: Actual installation of individual components not yet implemented.")
		fmt.Println("Use individual 'dvt plugin', 'dvt prompt', 'dvt profile' commands to install specific components.")

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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if wide {
			fmt.Fprintln(w, "NAME\tTYPE\tCATEGORY\tPLUGINS\tPROMPTS\tPROFILES\tEXTENDS\tDESCRIPTION")
		} else {
			fmt.Fprintln(w, "NAME\tTYPE\tPLUGINS\tCATEGORY\tDESCRIPTION")
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

			desc := p.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}

			if wide {
				fmt.Fprintf(w, "%s\tlibrary\t%s\t%d\t%d\t%d\t%s\t%s\n",
					p.Name, p.Category, pluginCount, promptCount, profileCount, p.Extends, desc)
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
