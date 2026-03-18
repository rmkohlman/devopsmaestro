package cmd

import (
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// registryCmd represents the registry parent command
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage registry configuration and defaults",
	Long: `Manage registry configuration including enabling/disabling registry types,
setting default registries, and viewing registry defaults.`,
}

// registryEnableCmd enables a registry type
var registryEnableCmd = &cobra.Command{
	Use:   "enable <type>",
	Short: "Enable a registry type",
	Long: `Enable a registry type (oci, pypi, npm, go, http).
Optionally set lifecycle mode (persistent, on-demand, manual).`,
	Args: cobra.ExactArgs(1),
	RunE: runRegistryEnable,
}

// registryDisableCmd disables a registry type
var registryDisableCmd = &cobra.Command{
	Use:   "disable <type>",
	Short: "Disable a registry type",
	Long:  `Disable a registry type and stop any running instances.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryDisable,
}

// runRegistryEnable enables a registry type
func runRegistryEnable(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("database not initialized: %w", err)
	}

	// Parse the type alias from arguments
	typeAlias := args[0]

	// Validate the type alias
	allAliases := registry.GetAllAliases()
	concreteType, ok := allAliases[typeAlias]
	if !ok {
		return fmt.Errorf("unknown registry type: %s (valid types: oci, pypi, npm, go, http)", typeAlias)
	}

	// Get lifecycle flag
	lifecycle, err := cmd.Flags().GetString("lifecycle")
	if err != nil {
		return fmt.Errorf("failed to get lifecycle flag: %w", err)
	}

	// Validate lifecycle
	validLifecycles := map[string]bool{
		"persistent": true,
		"on-demand":  true,
		"manual":     true,
	}
	if !validLifecycles[lifecycle] {
		return fmt.Errorf("invalid lifecycle: %s (valid: persistent, on-demand, manual)", lifecycle)
	}

	// Check if a registry of this concrete type already exists
	existingRegistry, err := ds.GetRegistryByName(concreteType)
	if err == nil && existingRegistry != nil {
		// Registry already exists, just set it as default
		defaults := registry.NewRegistryDefaults(ds)
		if err := defaults.SetByAlias(ctx, typeAlias, existingRegistry.Name); err != nil {
			return fmt.Errorf("failed to set default registry: %w", err)
		}

		// Output success message
		render.Success(fmt.Sprintf("Registry type '%s' enabled using existing registry '%s'", typeAlias, existingRegistry.Name))
		return render.OutputWith("colored", render.KeyValueData{
			Pairs: []render.KeyValue{
				{Key: "Type", Value: typeAlias},
				{Key: "Registry", Value: existingRegistry.Name},
				{Key: "Lifecycle", Value: existingRegistry.Lifecycle},
			},
		}, render.Options{})
	}

	// No registry exists, create one with default name matching the type
	newRegistry := &models.Registry{
		Name:        concreteType,
		Type:        concreteType,
		Enabled:     true,
		Lifecycle:   lifecycle,
		Status:      "stopped",
		IdleTimeout: 1800, // 30 minutes default
	}

	// Apply defaults for port and storage
	newRegistry.ApplyDefaults()

	// Apply strategy-based defaults (version) — RC-1: versions belong to strategy layer
	if newRegistry.Version == "" {
		factory := registry.NewServiceFactory()
		if strategy, err := factory.GetStrategy(newRegistry.Type); err == nil {
			if v := strategy.GetDefaultVersion(); v != "" {
				newRegistry.Version = v
			}
		}
	}

	// Validate the registry
	if err := newRegistry.Validate(); err != nil {
		return fmt.Errorf("invalid registry configuration: %w", err)
	}

	// Create the registry
	if err := ds.CreateRegistry(newRegistry); err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Set as default
	defaults := registry.NewRegistryDefaults(ds)
	if err := defaults.SetByAlias(ctx, typeAlias, newRegistry.Name); err != nil {
		return fmt.Errorf("failed to set default registry: %w", err)
	}

	// Output success message
	render.Success(fmt.Sprintf("Registry type '%s' enabled with new registry '%s'", typeAlias, newRegistry.Name))
	return render.OutputWith("colored", render.KeyValueData{
		Pairs: []render.KeyValue{
			{Key: "Type", Value: typeAlias},
			{Key: "Registry", Value: newRegistry.Name},
			{Key: "Lifecycle", Value: newRegistry.Lifecycle},
			{Key: "Port", Value: fmt.Sprintf("%d", newRegistry.Port)},
		},
	}, render.Options{})
}

// runRegistryDisable disables a registry type
func runRegistryDisable(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("database not initialized: %w", err)
	}

	// Parse the type alias from arguments
	typeAlias := args[0]

	// Validate the type alias
	allAliases := registry.GetAllAliases()
	if _, ok := allAliases[typeAlias]; !ok {
		return fmt.Errorf("unknown registry type: %s (valid types: oci, pypi, npm, go, http)", typeAlias)
	}

	// Clear the default for this type
	defaults := registry.NewRegistryDefaults(ds)
	key := getDefaultKeyForAlias(typeAlias)
	if err := defaults.ClearDefault(ctx, key); err != nil {
		return fmt.Errorf("failed to clear default registry: %w", err)
	}

	// Output success message
	render.Success(fmt.Sprintf("Registry type '%s' disabled", typeAlias))
	return render.OutputWith("colored", render.KeyValueData{
		Pairs: []render.KeyValue{
			{Key: "Type", Value: typeAlias},
			{Key: "Status", Value: "Disabled"},
		},
	}, render.Options{})
}

// getDefaultKeyForAlias converts an alias to its default key
func getDefaultKeyForAlias(alias string) string {
	switch alias {
	case registry.AliasOCI:
		return registry.DefaultKeyOCI
	case registry.AliasPyPI:
		return registry.DefaultKeyPyPI
	case registry.AliasNPM:
		return registry.DefaultKeyNPM
	case registry.AliasGo:
		return registry.DefaultKeyGo
	case registry.AliasHTTP:
		return registry.DefaultKeyHTTP
	default:
		// If not a known alias, just prefix with "registry-"
		return "registry-" + alias
	}
}

func init() {
	// Register enable/disable subcommands
	registryCmd.AddCommand(registryEnableCmd)
	registryCmd.AddCommand(registryDisableCmd)

	// Add flags to enable command
	registryEnableCmd.Flags().String("lifecycle", "on-demand", "Lifecycle mode: persistent, on-demand, manual")

	// Register registry command to root
	rootCmd.AddCommand(registryCmd)
}
