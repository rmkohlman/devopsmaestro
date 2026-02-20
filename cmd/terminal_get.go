package cmd

import (
	"fmt"
	"strings"

	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	terminalpkg "devopsmaestro/pkg/terminalops/package"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// terminalGetCmd is the 'terminal' subcommand under 'get' for kubectl-style namespacing
// Usage: dvm get terminal packages, dvm get terminal package <name>
var terminalGetCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Get terminal resources (packages)",
	Long: `Get terminal-related resources in kubectl-style namespaced format.

Examples:
  dvm get terminal packages                 # List all terminal packages
  dvm get terminal package dev-essentials   # Get specific package details
  dvm get terminal packages -o yaml         # Output as YAML
  dvm get terminal defaults                 # Show terminal defaults
`,
}

// terminalGetPackagesCmd lists all terminal packages
// Usage: dvm get terminal packages
var terminalGetPackagesCmd = &cobra.Command{
	Use:     "packages",
	Aliases: []string{"pkg", "pkgs"},
	Short:   "List all terminal packages",
	Long: `List all terminal packages stored in the database.

Examples:
  dvm get terminal packages
  dvm get terminal packages -o yaml
  dvm get terminal packages -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getTerminalPackages(cmd)
	},
}

// terminalGetPackageCmd gets a specific terminal package
// Usage: dvm get terminal package <name>
var terminalGetPackageCmd = &cobra.Command{
	Use:   "package [name]",
	Short: "Get a specific terminal package",
	Long: `Get a specific terminal package by name.

Examples:
  dvm get terminal package dev-essentials
  dvm get terminal package dev-essentials -o yaml
  dvm get terminal package poweruser -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getTerminalPackage(cmd, args[0])
	},
}

// terminalGetDefaultsCmd shows terminal defaults
// Usage: dvm get terminal defaults
var terminalGetDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Show current terminal defaults",
	Long: `Show current terminal default configuration values.

Displays terminal-related defaults including:
- terminal-package: Default terminal package for workspaces

Examples:
  dvm get terminal defaults                   # Show table format
  dvm get terminal defaults -o yaml          # Show as YAML
  dvm get terminal defaults -o json          # Show as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getTerminalDefaults(cmd)
	},
}

func init() {
	// Add terminal subcommand to get
	getCmd.AddCommand(terminalGetCmd)

	// Add resource types under terminal
	terminalGetCmd.AddCommand(terminalGetPackagesCmd)
	terminalGetCmd.AddCommand(terminalGetPackageCmd)
	terminalGetCmd.AddCommand(terminalGetDefaultsCmd)
}

// getTerminalPackages lists all terminal packages
func getTerminalPackages(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindTerminalPackage)
	if err != nil {
		return fmt.Errorf("failed to list terminal packages: %w", err)
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No terminal packages found",
			EmptyHints:   []string{"dvm apply -f terminal-package.yaml"},
		})
	}

	// Extract underlying packages from resources
	packages := make([]*terminalpkg.Package, len(resources))
	for i, res := range resources {
		pr := res.(*handlers.TerminalPackageResource)
		packages[i] = pr.Package()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		packagesYAML := make([]*terminalpkg.PackageYAML, len(packages))
		for i, p := range packages {
			packagesYAML[i] = p.ToYAML()
		}
		return render.OutputWith(getOutputFormat, packagesYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "CATEGORY", "PLUGINS", "PROMPTS", "EXTENDS"},
		Rows:    make([][]string, len(packages)),
	}

	for i, p := range packages {
		category := p.Category
		if category == "" {
			category = "-"
		}

		pluginCount := fmt.Sprintf("%d", len(p.Plugins))
		promptCount := fmt.Sprintf("%d", len(p.Prompts))

		extends := p.Extends
		if extends == "" {
			extends = "-"
		}

		tableData.Rows[i] = []string{
			p.Name,
			category,
			pluginCount,
			promptCount,
			extends,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// getTerminalPackage gets a specific terminal package
func getTerminalPackage(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindTerminalPackage, name)
	if err != nil {
		return fmt.Errorf("failed to get terminal package '%s': %w", name, err)
	}

	p := res.(*handlers.TerminalPackageResource).Package()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, p.ToYAML(), render.Options{})
	}

	// For human output, show detail view
	category := p.Category
	if category == "" {
		category = "-"
	}

	extends := p.Extends
	if extends == "" {
		extends = "-"
	}

	pluginsList := "-"
	if len(p.Plugins) > 0 {
		pluginsList = strings.Join(p.Plugins, ", ")
	}

	promptsList := "-"
	if len(p.Prompts) > 0 {
		promptsList = strings.Join(p.Prompts, ", ")
	}

	profilesList := "-"
	if len(p.Profiles) > 0 {
		profilesList = strings.Join(p.Profiles, ", ")
	}

	tagsList := "-"
	if len(p.Tags) > 0 {
		tagsList = strings.Join(p.Tags, ", ")
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: p.Name},
		render.KeyValue{Key: "Category", Value: category},
		render.KeyValue{Key: "Description", Value: p.Description},
		render.KeyValue{Key: "Extends", Value: extends},
		render.KeyValue{Key: "Plugins", Value: pluginsList},
		render.KeyValue{Key: "Prompts", Value: promptsList},
		render.KeyValue{Key: "Profiles", Value: profilesList},
		render.KeyValue{Key: "Tags", Value: tagsList},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Terminal Package Details",
	})
}

// getTerminalDefaults shows current terminal defaults
func getTerminalDefaults(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get all defaults from database
	allDefaults, err := ds.ListDefaults()
	if err != nil {
		return fmt.Errorf("failed to list defaults: %w", err)
	}

	// Filter to terminal-related keys
	terminalKeys := []string{"terminal-package"}
	terminalDefaults := make(map[string]interface{})

	for _, key := range terminalKeys {
		if value, exists := allDefaults[key]; exists {
			terminalDefaults[key] = value
		} else {
			// Show empty value for missing keys
			terminalDefaults[key] = ""
		}
	}

	// For structured output (JSON/YAML)
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, terminalDefaults, render.Options{})
	}

	// For table output
	tableData := render.TableData{
		Headers: []string{"KEY", "VALUE"},
		Rows:    make([][]string, 0, len(terminalKeys)),
	}

	for _, key := range terminalKeys {
		value := terminalDefaults[key]
		var displayValue string

		switch v := value.(type) {
		case string:
			if v == "" {
				displayValue = "(none)"
			} else {
				displayValue = v
			}
		default:
			displayValue = fmt.Sprintf("%v", v)
		}

		tableData.Rows = append(tableData.Rows, []string{key, displayValue})
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}
