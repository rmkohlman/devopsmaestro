package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"devopsmaestro/db"
	"devopsmaestro/pkg/terminalops/emulator"
	"devopsmaestro/pkg/terminalops/store"

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
  dvt emulator list                    # List all emulators
  dvt emulator get wezterm-default     # Show emulator details
  dvt emulator enable wezterm-default  # Enable emulator
  dvt emulator disable old-config     # Disable emulator`,
}

// emulatorListCmd lists all emulators
var emulatorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List terminal emulators",
	Long: `List terminal emulator configurations from the database.

Examples:
  dvt emulator list                     # List all emulators
  dvt emulator list --type wezterm      # Filter by emulator type
  dvt emulator list --category dev      # Filter by category
  dvt emulator list -o yaml             # YAML output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getEmulatorStore(cmd)
		if err != nil {
			return err
		}
		defer store.Close()

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
			fmt.Println("No emulators found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputEmulators(emulators, format)
	},
}

// emulatorGetCmd shows emulator details
var emulatorGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get emulator configuration details",
	Long: `Show details of a specific terminal emulator configuration.

Examples:
  dvt emulator get wezterm-default
  dvt emulator get my-config -o yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := getEmulatorStore(cmd)
		if err != nil {
			return err
		}
		defer store.Close()

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
			fmt.Printf("Emulator '%s' is already enabled\n", name)
			return nil
		}

		// Enable and update
		emu.Enabled = true
		if err := store.Update(emu); err != nil {
			return fmt.Errorf("failed to enable emulator: %w", err)
		}

		fmt.Printf("Emulator '%s' enabled\n", name)
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
			fmt.Printf("Emulator '%s' is already disabled\n", name)
			return nil
		}

		// Disable and update
		emu.Enabled = false
		if err := store.Update(emu); err != nil {
			return fmt.Errorf("failed to disable emulator: %w", err)
		}

		fmt.Printf("Emulator '%s' disabled\n", name)
		return nil
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
	return store.NewDBEmulatorStore(*dataStore), nil
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tENABLED\tCATEGORY\tDESCRIPTION")

		for _, emu := range emulators {
			enabled := "false"
			if emu.Enabled {
				enabled = "true"
			}

			desc := emu.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}

			category := emu.Category
			if category == "" {
				category = "-"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				emu.Name, emu.Type, enabled, category, desc)
		}
		w.Flush()
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
	emulatorCmd.AddCommand(emulatorListCmd)
	emulatorCmd.AddCommand(emulatorGetCmd)
	emulatorCmd.AddCommand(emulatorEnableCmd)
	emulatorCmd.AddCommand(emulatorDisableCmd)

	// Emulator list flags
	emulatorListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	emulatorListCmd.Flags().String("type", "", "Filter by emulator type (wezterm, alacritty, kitty, iterm2)")
	emulatorListCmd.Flags().String("category", "", "Filter by category")

	// Emulator get flags
	emulatorGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
}
