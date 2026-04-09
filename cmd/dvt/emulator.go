package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"devopsmaestro/db"
	"devopsmaestro/pkg/terminalbridge"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/emulator"
	"github.com/rmkohlman/MaestroTerminal/terminalops/emulator/library"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// emulatorCmd is the main emulator command
var emulatorCmd = &cobra.Command{
	Use:     "emulator",
	Aliases: []string{"emu"},
	Short:   "Manage terminal emulator configurations",
	Long: `Terminal emulator management for terminal configurations.

Manage configurations for terminal emulators like WezTerm, Alacritty, Kitty, and iTerm2.
Emulator configurations can be associated with workspaces and themes.

Examples:
  dvt emulator get                     # List all emulators
  dvt emulator get wezterm-default     # Show emulator details
  dvt emulator enable wezterm-default  # Enable emulator
  dvt emulator disable old-config     # Disable emulator`,
}

// emulatorGetCmd shows emulator details, or lists all emulators when no name is given
var emulatorGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get emulator configuration details (or list all)",
	Long: `Show details of a specific terminal emulator configuration, or list all emulators.

Examples:
  dvt emulator get                       # List all emulators
  dvt emulator get --type wezterm        # Filter by emulator type
  dvt emulator get --category dev        # Filter by category
  dvt emulator get -o yaml               # YAML output
  dvt emulator get wezterm-default       # Show specific emulator
  dvt emulator get my-config -o yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getEmulatorStore(cmd)
		if err != nil {
			return err
		}
		defer store.Close()

		if len(args) == 0 {
			// List mode
			var emulators []*emulator.Emulator

			// Filter by type if specified
			emulatorType, _ := cmd.Flags().GetString("type")
			if emulatorType != "" {
				emulators, err = store.ListByType(emulatorType)
				if err != nil {
					return fmt.Errorf("failed to list emulators by type: %w", err)
				}
			} else {
				emulators, err = store.List()
				if err != nil {
					return fmt.Errorf("failed to list emulators: %w", err)
				}
			}

			// Filter by category if specified
			category, _ := cmd.Flags().GetString("category")
			if category != "" {
				var filtered []*emulator.Emulator
				for _, emu := range emulators {
					if emu.Category == category {
						filtered = append(filtered, emu)
					}
				}
				emulators = filtered
			}

			if len(emulators) == 0 {
				render.Info("No emulators found")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputEmulators(emulators, format)
		}

		// Single get mode
		name := args[0]
		emu, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("emulator not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputEmulator(emu, format)
	},
}

// emulatorEnableCmd enables an emulator
var emulatorEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a terminal emulator configuration",
	Long: `Enable a terminal emulator configuration by setting enabled=true.

Examples:
  dvt emulator enable wezterm-default`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := getEmulatorStore(cmd)
		if err != nil {
			return err
		}
		defer store.Close()

		// Get the emulator first
		emu, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("emulator not found: %s", name)
		}

		// Check if already enabled
		if emu.Enabled {
			render.Infof("Emulator '%s' is already enabled", name)
			return nil
		}

		// Enable and update
		emu.Enabled = true
		if err := store.Update(emu); err != nil {
			return fmt.Errorf("failed to enable emulator: %w", err)
		}

		render.Successf("Emulator '%s' enabled", name)
		return nil
	},
}

// emulatorDisableCmd disables an emulator
var emulatorDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a terminal emulator configuration",
	Long: `Disable a terminal emulator configuration by setting enabled=false.

Examples:
  dvt emulator disable old-config`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := getEmulatorStore(cmd)
		if err != nil {
			return err
		}
		defer store.Close()

		// Get the emulator first
		emu, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("emulator not found: %s", name)
		}

		// Check if already disabled
		if !emu.Enabled {
			render.Infof("Emulator '%s' is already disabled", name)
			return nil
		}

		// Disable and update
		emu.Enabled = false
		if err := store.Update(emu); err != nil {
			return fmt.Errorf("failed to disable emulator: %w", err)
		}

		render.Successf("Emulator '%s' disabled", name)
		return nil
	},
}

// emulatorLibraryImportCmd imports an emulator from the library
var emulatorLibraryImportCmd = &cobra.Command{
	Use:   "import <name>",
	Short: "Import a terminal emulator from the library",
	Long: `Import a terminal emulator configuration from the built-in library.

Examples:
  dvt emulator library import maestro            # Import maestro emulator
  dvt emulator library import minimal --force    # Overwrite if exists
  dvt emulator library import developer --dry-run # Preview import`,
	Args: cobra.ExactArgs(1),
	RunE: runEmulatorLibraryImport,
}

// runEmulatorLibraryImport is the shared implementation for emulator library import
func runEmulatorLibraryImport(cmd *cobra.Command, args []string) error {
	name := args[0]

	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get emulator from library
	lib := library.Default()
	libEmulator, err := lib.Get(name)
	if err != nil {
		return fmt.Errorf("emulator not found in library: %s", name)
	}

	if dryRun {
		render.Infof("Would install emulator: %s", name)
		render.Plainf("Type: %s", libEmulator.Type)
		render.Plainf("Description: %s", libEmulator.Description)
		render.Plainf("Category: %s", libEmulator.Category)
		return nil
	}

	store, err := getEmulatorStore(cmd)
	if err != nil {
		return err
	}
	defer store.Close()

	// Check if emulator already exists
	_, err = store.Get(name)
	if err == nil && !force {
		return fmt.Errorf("emulator '%s' already exists (use --force to overwrite)", name)
	}

	// Install emulator
	if err := store.Upsert(libEmulator); err != nil {
		return fmt.Errorf("failed to install emulator: %w", err)
	}

	render.Successf("Emulator '%s' installed successfully", name)
	return nil
}

// emulatorApplyCmd applies an emulator configuration from a YAML file
var emulatorApplyCmd = &cobra.Command{
	Use:   "apply -f <file>",
	Short: "Apply an emulator configuration from a YAML file",
	Long: `Apply a terminal emulator configuration from a YAML file.

Examples:
  dvt emulator apply -f my-emulator.yaml  # Apply from file
  dvt emulator apply -f -                 # Apply from stdin
  dvt emulator apply -f config.yaml --dry-run # Preview changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, _ := cmd.Flags().GetString("filename")
		if filename == "" {
			return fmt.Errorf("filename is required (use -f)")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Read YAML content
		var data []byte
		var err error

		if filename == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(filename)
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Parse emulator YAML
		emu, err := emulator.Parse(data)
		if err != nil {
			return fmt.Errorf("failed to parse emulator YAML: %w", err)
		}

		if dryRun {
			render.Infof("Would apply emulator: %s", emu.Name)
			render.Plainf("Type: %s", emu.Type)
			render.Plainf("Description: %s", emu.Description)
			render.Plainf("Category: %s", emu.Category)
			return nil
		}

		store, err := getEmulatorStore(cmd)
		if err != nil {
			return err
		}
		defer store.Close()

		// Apply emulator (upsert)
		if err := store.Upsert(emu); err != nil {
			return fmt.Errorf("failed to apply emulator: %w", err)
		}

		render.Successf("Emulator '%s' applied successfully", emu.Name)
		return nil
	},
}

// emulatorLibraryCmd manages library operations
var emulatorLibraryCmd = &cobra.Command{
	Use:     "library",
	Aliases: []string{"lib"},
	Short:   "Manage emulator library",
	Long: `Access the built-in library of curated terminal emulator configurations.

Examples:
  dvt emulator library get                   # List all library emulators
  dvt emulator library get --type wezterm    # Filter by emulator type
  dvt emulator library describe maestro      # Show library emulator details`,
}

// emulatorLibraryListCmd lists library emulators
var emulatorLibraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available library emulators",
	Long: `List terminal emulator configurations from the built-in library.

Examples:
  dvt emulator library get                   # List all library emulators
  dvt emulator library get --type wezterm    # Filter by emulator type
  dvt emulator library get -o yaml           # YAML output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lib := library.Default()

		// Filter by type if specified
		emulatorType, _ := cmd.Flags().GetString("type")
		var emulators []*emulator.Emulator

		if emulatorType != "" {
			emulators = lib.ListByType(emulatorType)
		} else {
			emulators = lib.All()
		}

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			var filtered []*emulator.Emulator
			for _, emu := range emulators {
				if emu.Category == category {
					filtered = append(filtered, emu)
				}
			}
			emulators = filtered
		}

		if len(emulators) == 0 {
			render.Info("No library emulators found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputEmulators(emulators, format)
	},
}

// emulatorLibraryShowCmd shows library emulator details
var emulatorLibraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show library emulator details",
	Long: `Show details of a specific emulator from the built-in library.

Examples:
  dvt emulator library describe maestro
  dvt emulator library describe minimal -o yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib := library.Default()
		emu, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("library emulator not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputEmulator(emu, format)
	},
}

// getEmulatorStore extracts DataStore from command context and returns database-backed emulator store
func getEmulatorStore(cmd *cobra.Command) (emulator.EmulatorStore, error) {
	// Extract DataStore from context (following established dvt pattern)
	dataStoreInterface := cmd.Context().Value("dataStore")
	if dataStoreInterface == nil {
		return nil, fmt.Errorf("database not initialized - run 'dvt init' or check configuration")
	}

	dataStore := dataStoreInterface.(*db.DataStore)

	// Return database-backed emulator store via factory
	return terminalbridge.NewDBEmulatorStore(*dataStore), nil
}

// outputEmulators outputs emulators in the specified format
func outputEmulators(emulators []*emulator.Emulator, format string) error {
	// Sort by name
	sort.Slice(emulators, func(i, j int) bool {
		return emulators[i].Name < emulators[j].Name
	})

	switch format {
	case "yaml":
		for i, emu := range emulators {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(emu)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		data, err := json.MarshalIndent(emulators, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "TYPE", "ENABLED", "CATEGORY", "DESCRIPTION")

		for _, emu := range emulators {
			enabled := "false"
			if emu.Enabled {
				enabled = "true"
			}

			category := emu.Category
			if category == "" {
				category = "-"
			}

			tb.AddRow(emu.Name, string(emu.Type), enabled, category, render.Truncate(emu.Description, 40))
		}
		return render.OutputWith("", tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// outputEmulator outputs a single emulator in the specified format
func outputEmulator(emu *emulator.Emulator, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(emu)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(emu, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func init() {
	// Add subcommands
	emulatorCmd.AddCommand(emulatorGetCmd)
	emulatorCmd.AddCommand(emulatorEnableCmd)
	emulatorCmd.AddCommand(emulatorDisableCmd)
	emulatorCmd.AddCommand(emulatorApplyCmd)
	emulatorCmd.AddCommand(emulatorLibraryCmd)

	// Library subcommands
	emulatorLibraryCmd.AddCommand(emulatorLibraryListCmd)
	emulatorLibraryCmd.AddCommand(emulatorLibraryShowCmd)
	emulatorLibraryCmd.AddCommand(emulatorLibraryImportCmd)

	// Emulator get flags (merged from list + get)
	emulatorGetCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	emulatorGetCmd.Flags().String("type", "", "Filter by emulator type (wezterm, alacritty, kitty, iterm2)")
	emulatorGetCmd.Flags().String("category", "", "Filter by category")

	// Emulator library import flags
	emulatorLibraryImportCmd.Flags().Bool("force", false, "Overwrite if emulator already exists")
	emulatorLibraryImportCmd.Flags().Bool("dry-run", false, "Show what would be installed without installing")

	// Emulator apply flags
	emulatorApplyCmd.Flags().StringP("filename", "f", "", "YAML file to apply (required)")
	emulatorApplyCmd.Flags().Bool("dry-run", false, "Show what would be applied without applying")
	emulatorApplyCmd.MarkFlagRequired("filename")

	// Library list flags
	emulatorLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	emulatorLibraryListCmd.Flags().String("type", "", "Filter by emulator type (wezterm, alacritty, kitty, iterm2)")
	emulatorLibraryListCmd.Flags().String("category", "", "Filter by category")

	// Library show flags
	emulatorLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")

	// Hidden backward-compat aliases for deprecated verbs in emulator (after flags)
	emulatorCmd.AddCommand(hiddenAlias("list", emulatorGetCmd))
	emulatorCmd.AddCommand(hiddenAlias("install", emulatorLibraryImportCmd))
	emulatorLibraryCmd.AddCommand(hiddenAlias("list", emulatorLibraryListCmd))
	emulatorLibraryCmd.AddCommand(hiddenAlias("show", emulatorLibraryShowCmd))
}
