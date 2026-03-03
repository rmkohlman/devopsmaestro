package cmd

import (
	"fmt"

	"devopsmaestro/pkg/registry"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// registrySetDefaultCmd sets the default registry for a type
var registrySetDefaultCmd = &cobra.Command{
	Use:   "set-default <type> <registry-name>",
	Short: "Set the default registry for a type",
	Long: `Set the default registry to use for a specific type (oci, pypi, npm, go, http).

Examples:
  dvm registry set-default oci zot-local
  dvm registry set-default pypi devpi-local`,
	Args: cobra.ExactArgs(2),
	RunE: runRegistrySetDefault,
}

// registryGetDefaultsCmd displays all default registry settings
var registryGetDefaultsCmd = &cobra.Command{
	Use:   "get-defaults",
	Short: "Display default registry settings",
	Long:  `Display the default registry configured for each type.`,
	Args:  cobra.NoArgs,
	RunE:  runRegistryGetDefaults,
}

func init() {
	// Register set-default and get-defaults subcommands
	registryCmd.AddCommand(registrySetDefaultCmd)
	registryCmd.AddCommand(registryGetDefaultsCmd)
}

// runRegistrySetDefault implements the set-default command
func runRegistrySetDefault(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	aliasType := args[0]
	registryName := args[1]

	// Get dataStore from context
	store, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("database not initialized: %w", err)
	}

	// Validate type is a known alias
	allAliases := registry.GetAllAliases()
	expectedType, isValidAlias := allAliases[aliasType]
	if !isValidAlias {
		return fmt.Errorf("unknown type '%s'. Valid types: oci, pypi, npm, go, http", aliasType)
	}

	// Validate registry exists
	reg, err := store.GetRegistryByName(registryName)
	if err != nil {
		return fmt.Errorf("registry '%s' not found: %w", registryName, err)
	}

	// Validate registry type matches the alias
	if reg.Type != expectedType {
		return fmt.Errorf("registry '%s' is type '%s' but alias '%s' expects type '%s'",
			registryName, reg.Type, aliasType, expectedType)
	}

	// Set the default using RegistryDefaults
	defaults := registry.NewRegistryDefaults(store)
	if err := defaults.SetByAlias(ctx, aliasType, registryName); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	// Output success message
	successMsg := fmt.Sprintf("Set default %s registry to '%s'", aliasType, registryName)
	return render.Success(successMsg)
}

// runRegistryGetDefaults implements the get-defaults command
func runRegistryGetDefaults(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get dataStore from context
	store, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("database not initialized: %w", err)
	}

	// Get all defaults
	defaults := registry.NewRegistryDefaults(store)
	allDefaults, err := defaults.GetAllDefaults(ctx)
	if err != nil {
		return fmt.Errorf("failed to get defaults: %w", err)
	}

	// Build table data
	// Map from default keys to display aliases
	keyToAlias := map[string]string{
		registry.DefaultKeyOCI:  registry.AliasOCI,
		registry.DefaultKeyPyPI: registry.AliasPyPI,
		registry.DefaultKeyNPM:  registry.AliasNPM,
		registry.DefaultKeyGo:   registry.AliasGo,
		registry.DefaultKeyHTTP: registry.AliasHTTP,
	}

	// Get all aliases in order
	orderedAliases := []string{
		registry.AliasOCI,
		registry.AliasPyPI,
		registry.AliasNPM,
		registry.AliasGo,
		registry.AliasHTTP,
	}

	// Build rows
	rows := make([][]string, 0, len(orderedAliases))
	for _, alias := range orderedAliases {
		// Get the default key for this alias
		var defaultKey string
		for key, a := range keyToAlias {
			if a == alias {
				defaultKey = key
				break
			}
		}

		// Get the registry name for this default
		registryName, exists := allDefaults[defaultKey]

		var registryDisplay, endpoint, status string
		if !exists || registryName == "" {
			registryDisplay = "-"
			endpoint = "-"
			status = "not configured"
		} else {
			// Look up registry details
			reg, err := store.GetRegistryByName(registryName)
			if err != nil {
				registryDisplay = registryName
				endpoint = "-"
				status = "not found"
			} else {
				registryDisplay = reg.Name
				endpoint = fmt.Sprintf("http://localhost:%d", reg.Port)
				status = reg.Status
				if status == "" {
					status = "stopped"
				}
			}
		}

		rows = append(rows, []string{
			alias,
			registryDisplay,
			endpoint,
			status,
		})
	}

	tableData := render.TableData{
		Headers: []string{"TYPE", "REGISTRY", "ENDPOINT", "STATUS"},
		Rows:    rows,
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}
