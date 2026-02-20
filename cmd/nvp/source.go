package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	nvimpackage "devopsmaestro/pkg/nvimops/package"
	"devopsmaestro/pkg/nvimops/sync"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// SOURCE COMMANDS
// =============================================================================

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage external plugin sources",
	Long: `Manage external plugin sources like LazyVim, AstroNvim, NvChad, etc.
	
External sources allow you to import plugin configurations from popular
Neovim distributions and configurations. This provides a starting point
for your own customizations while following proven patterns.

Available Commands:
  list      List available sources with descriptions  
  show      Show detailed information about a source
  sync      Sync plugins from an external source

Examples:
  nvp source list                    # List all available sources
  nvp source show lazyvim           # Show LazyVim source details
  nvp source sync lazyvim           # Sync all LazyVim plugins
  nvp source sync lazyvim --dry-run # Preview what would be synced`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior is to list sources
		return sourceListCmd.RunE(cmd, args)
	},
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available sources",
	Long: `List all available external plugin sources.
	
Sources are external Neovim configurations that can be used to import
plugin definitions. Each source may have different plugin collections,
configurations, and approaches.

Examples:
  nvp source list                    # List with table format
  nvp source list -o yaml           # YAML format
  nvp source list -o json           # JSON format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		factory := sync.NewSourceHandlerFactory()
		sources := factory.ListSources()

		if len(sources) == 0 {
			fmt.Println("No external sources available")
			fmt.Println("\nCheck the documentation for how to register external sources.")
			return nil
		}

		// Get info for each source
		var sourceInfos []*sync.SourceInfo
		for _, sourceName := range sources {
			info, err := factory.GetHandlerInfo(sourceName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not get info for source %s: %v\n", sourceName, err)
				continue
			}
			sourceInfos = append(sourceInfos, info)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputSources(sourceInfos, format)
	},
}

var sourceShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a source",
	Long: `Show detailed information about an external plugin source.
	
This displays metadata about the source including description, URL,
type, authentication requirements, and available configuration options.

Examples:
  nvp source show lazyvim           # Show LazyVim source details
  nvp source show astronvim -o json # JSON format`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName := args[0]

		factory := sync.NewSourceHandlerFactory()

		// Check if source exists
		if !factory.IsSupported(sourceName) {
			return fmt.Errorf("source not found: %s\n\nUse 'nvp source list' to see available sources", sourceName)
		}

		info, err := factory.GetHandlerInfo(sourceName)
		if err != nil {
			return fmt.Errorf("failed to get source info: %w", err)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputSource(info, format)
	},
}

var sourceSyncCmd = &cobra.Command{
	Use:   "sync <name>",
	Short: "Sync plugins from an external source",
	Long: `Sync plugins from an external source to your local plugin store.
	
This command imports plugin definitions from external Neovim configurations
like LazyVim, AstroNvim, NvChad, etc. You can filter which plugins to sync
using labels and control whether to overwrite existing plugins.

The sync operation:
1. Connects to the external source
2. Lists available plugins  
3. Applies filters (if specified)
4. Downloads/converts plugin definitions to YAML
5. Saves to your local plugin store

Label Filtering:
  Use -l/--selector to filter plugins by labels. Common labels include:
  - category=lang,ui,navigation,lsp,completion
  - enabled=true,false  
  - lazy=true,false
  - priority=high,medium,low

Version/Tag Selection:
  Use --tag to sync from a specific version or branch of the source.

Output Control:
  - --dry-run: Preview what would be synced without making changes
  - --force: Overwrite existing plugins 
  - -o/--output: Control output format (table, yaml, json)

Examples:
  nvp source sync lazyvim                    # Sync all LazyVim plugins
  nvp source sync lazyvim --dry-run          # Preview sync operation
  nvp source sync lazyvim -l category=lsp    # Sync only LSP plugins  
  nvp source sync lazyvim --tag v15.0.0      # Sync from specific version
  nvp source sync lazyvim --force            # Overwrite existing plugins
  nvp source sync lazyvim -o yaml            # YAML output format`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName := args[0]

		// Get flags
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")
		outputFormat, _ := cmd.Flags().GetString("output")
		selectors, _ := cmd.Flags().GetStringSlice("selector")
		tag, _ := cmd.Flags().GetString("tag")

		// Create factory and handler
		factory := sync.NewSourceHandlerFactory()

		if !factory.IsSupported(sourceName) {
			return fmt.Errorf("source not found: %s\n\nUse 'nvp source list' to see available sources", sourceName)
		}

		handler, err := factory.CreateHandler(sourceName)
		if err != nil {
			return fmt.Errorf("failed to create source handler: %w", err)
		}

		// Build sync options using builder pattern
		optionsBuilder := sync.NewSyncOptions().
			DryRun(dryRun).
			Overwrite(force)

		// Add target directory
		targetDir := getConfigDir() + "/plugins"
		optionsBuilder.WithTargetDir(targetDir)

		// Parse selectors (format: key=value)
		for _, selector := range selectors {
			parts := strings.SplitN(selector, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid selector format '%s'. Use key=value format", selector)
			}
			optionsBuilder.WithFilter(parts[0], parts[1])
		}

		// Add tag filter if specified
		if tag != "" {
			optionsBuilder.WithFilter("tag", tag)
		}

		// Create package creator for auto-generating packages
		packagesDir := getConfigDir() + "/packages"
		packageCreator := nvimpackage.NewFilePackageCreator(packagesDir)
		optionsBuilder.WithPackageCreator(packageCreator)

		options := optionsBuilder.Build()

		// Validate source before syncing
		if err := handler.Validate(cmd.Context()); err != nil {
			return fmt.Errorf("source validation failed: %w", err)
		}

		// Show what we're about to do
		if dryRun {
			fmt.Printf("Would sync from source '%s':\n", sourceName)
		} else {
			fmt.Printf("Syncing from source '%s'...\n", sourceName)
		}

		if len(options.Filters) > 0 {
			fmt.Printf("Filters: ")
			var filters []string
			for k, v := range options.Filters {
				filters = append(filters, fmt.Sprintf("%s=%s", k, v))
			}
			fmt.Printf("%s\n", strings.Join(filters, ", "))
		}

		if options.Overwrite {
			fmt.Println("Mode: Overwrite existing plugins")
		}

		fmt.Println()

		// Perform the sync
		result, err := handler.Sync(cmd.Context(), options)
		if err != nil {
			return fmt.Errorf("sync operation failed: %w", err)
		}

		// Display results
		return outputSyncResult(result, outputFormat, dryRun)
	},
}

func init() {
	// Add source command to root
	rootCmd.AddCommand(sourceCmd)

	// Add subcommands to source
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceShowCmd)
	sourceCmd.AddCommand(sourceSyncCmd)

	// Flags for list command
	sourceListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")

	// Flags for show command
	sourceShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")

	// Flags for sync command
	sourceSyncCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	sourceSyncCmd.Flags().StringSliceP("selector", "l", nil, "Label selector to filter plugins (key=value)")
	sourceSyncCmd.Flags().String("tag", "", "Specific version/tag to sync from")
	sourceSyncCmd.Flags().BoolP("force", "f", false, "Overwrite existing plugins")
	sourceSyncCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
}

// =============================================================================
// OUTPUT FUNCTIONS
// =============================================================================

// outputSources renders a list of sources in the specified format
func outputSources(sources []*sync.SourceInfo, format string) error {
	switch format {
	case "yaml":
		return render.OutputWith("yaml", sources, render.Options{})
	case "json":
		return render.OutputWith("json", sources, render.Options{})
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION")
		for _, source := range sources {
			desc := source.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", source.Name, source.Type, desc)
		}
		w.Flush()
		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// outputSource renders a single source in the specified format
func outputSource(source *sync.SourceInfo, format string) error {
	switch format {
	case "yaml", "":
		return render.OutputWith("yaml", source, render.Options{})
	case "json":
		return render.OutputWith("json", source, render.Options{})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// outputSyncResult renders the sync result in the specified format
func outputSyncResult(result *sync.SyncResult, format string, dryRun bool) error {
	// Handle errors first
	if result.HasErrors() {
		fmt.Fprintf(os.Stderr, "Sync completed with %d errors:\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %d. %v\n", i+1, err)
		}
		fmt.Fprintln(os.Stderr)
	}

	switch format {
	case "yaml":
		return render.OutputWith("yaml", result, render.Options{})
	case "json":
		return render.OutputWith("json", result, render.Options{})
	case "table", "":
		// Table format with summary
		if dryRun {
			fmt.Printf("Dry run complete for source '%s'\n", result.SourceName)
		} else {
			fmt.Printf("Sync complete for source '%s'\n", result.SourceName)
		}

		fmt.Printf("Available: %d plugins\n", result.TotalAvailable)
		fmt.Printf("Synced: %d plugins\n", result.TotalSynced)

		if len(result.PluginsCreated) > 0 {
			fmt.Printf("\nCreated (%d):\n", len(result.PluginsCreated))
			for _, name := range result.PluginsCreated {
				if dryRun {
					fmt.Printf("  %s (would create)\n", name)
				} else {
					fmt.Printf("  ✓ %s\n", name)
				}
			}
		}

		if len(result.PluginsUpdated) > 0 {
			fmt.Printf("\nUpdated (%d):\n", len(result.PluginsUpdated))
			for _, name := range result.PluginsUpdated {
				if dryRun {
					fmt.Printf("  %s (would update)\n", name)
				} else {
					fmt.Printf("  ✓ %s\n", name)
				}
			}
		}

		if len(result.PackagesCreated) > 0 {
			fmt.Printf("\nPackages Created (%d):\n", len(result.PackagesCreated))
			for _, name := range result.PackagesCreated {
				if dryRun {
					fmt.Printf("  %s (would create)\n", name)
				} else {
					fmt.Printf("  ✓ %s\n", name)
				}
			}
		}

		if len(result.PackagesUpdated) > 0 {
			fmt.Printf("\nPackages Updated (%d):\n", len(result.PackagesUpdated))
			for _, name := range result.PackagesUpdated {
				if dryRun {
					fmt.Printf("  %s (would update)\n", name)
				} else {
					fmt.Printf("  ✓ %s\n", name)
				}
			}
		}

		// Show next steps
		if !dryRun && result.TotalSynced > 0 {
			fmt.Println("\nNext steps:")
			fmt.Println("  nvp list                    # View synced plugins")
			fmt.Println("  nvp generate                # Generate Lua files")
			fmt.Println("  nvp get <plugin-name>       # Customize specific plugins")
		}

		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
