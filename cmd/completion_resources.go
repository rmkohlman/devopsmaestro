package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"

	"github.com/spf13/cobra"
)

// completeResources is a generic helper that uses the Resource/Handler pattern
// to provide dynamic completions for resource names with descriptions.
func completeResources(cmd *cobra.Command, kind string) ([]string, cobra.ShellCompDirective) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	// Try to get datastore from context first (normal command context)
	ctx := cmd.Context()
	if dataStore := ctx.Value("dataStore"); dataStore != nil {
		if ds, ok := dataStore.(db.DataStore); ok {
			return getResourceCompletions(ds, kind)
		}
	}

	// For shell completion context, try to create a datastore with default config
	// This may fail if configuration isn't available, in which case we return empty
	createdDS, err := createDefaultDataStore()
	if err != nil {
		// Return empty on error, don't block completion
		return []string{}, cobra.ShellCompDirectiveDefault
	}
	defer func() {
		if closer, ok := createdDS.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	return getResourceCompletions(createdDS, kind)
}

// createDefaultDataStore creates a DataStore using default configuration
func createDefaultDataStore() (db.DataStore, error) {
	// Try the default factory first
	factory := db.NewDataStoreFactory()
	if ds, err := factory.Create(); err == nil {
		return ds, nil
	}

	// If that fails, try with explicit SQLite configuration
	cfg := db.DriverConfig{
		Type: db.DriverSQLite,
		Path: "~/.devopsmaestro/devopsmaestro.db", // Default path
	}

	driver, err := db.NewSQLiteDriver(cfg)
	if err != nil {
		return nil, err
	}

	if err := driver.Connect(); err != nil {
		return nil, err
	}

	return db.NewSQLDataStore(driver, nil), nil
}

// getResourceCompletions gets completions for a specific resource kind
func getResourceCompletions(ds db.DataStore, kind string) ([]string, cobra.ShellCompDirective) {
	// Get handler from registry
	handler := resource.GetHandler(kind)
	if handler == nil {
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	// Use Resource/Handler pattern to list resources
	resourceCtx := resource.Context{DataStore: ds}
	resources, err := handler.List(resourceCtx)
	if err != nil {
		// Return empty on error, don't block completion
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	// Convert to completion format with descriptions
	completions := make([]string, 0, len(resources))
	for _, res := range resources {
		name := res.GetName()
		desc := extractResourceDescription(res)
		if desc != "" {
			completions = append(completions, fmt.Sprintf("%s\t%s", name, desc))
		} else {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// extractResourceDescription extracts a description from different resource types.
// Returns empty string if no description is available.
func extractResourceDescription(res resource.Resource) string {
	switch r := res.(type) {
	case interface{ App() *models.App }:
		app := r.App()
		if app.Description.Valid {
			return app.Description.String
		}
	case interface{ Workspace() *models.Workspace }:
		workspace := r.Workspace()
		if workspace.Description.Valid {
			return workspace.Description.String
		}
	case interface{ Domain() *models.Domain }:
		domain := r.Domain()
		if domain.Description.Valid {
			return domain.Description.String
		}
	case interface{ Ecosystem() *models.Ecosystem }:
		ecosystem := r.Ecosystem()
		if ecosystem.Description.Valid {
			return ecosystem.Description.String
		}
	}
	return ""
}

// Specific completion functions for each resource type

func completeEcosystems(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Ecosystem")
}

func completeDomains(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Domain")
}

func completeApps(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "App")
}

func completeWorkspaces(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Workspace")
}

// registerHierarchyFlagCompletions registers completion functions for hierarchy flags.
// Call this from command init() functions after AddHierarchyFlags().
func registerHierarchyFlagCompletions(cmd *cobra.Command) {
	cmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
	cmd.RegisterFlagCompletionFunc("domain", completeDomains)
	cmd.RegisterFlagCompletionFunc("app", completeApps)
	cmd.RegisterFlagCompletionFunc("workspace", completeWorkspaces)
}

// NvimOps completion functions

func completeNvimPlugins(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "NvimPlugin")
}

func completeNvimThemes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "NvimTheme")
}

func completeNvimPackages(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "NvimPackage")
}

func completeTerminalPackages(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "TerminalPackage")
}

// registerAllResourceCompletions registers ValidArgsFunction for all commands
// that accept resource names as positional arguments.
// This is called from registerDynamicCompletions() in completion.go.
func registerAllResourceCompletions() {
	// Ecosystem commands
	if getEcosystemCmd != nil {
		getEcosystemCmd.ValidArgsFunction = completeEcosystems
	}
	if useEcosystemCmd != nil {
		useEcosystemCmd.ValidArgsFunction = completeEcosystems
	}
	if deleteEcosystemCmd != nil {
		deleteEcosystemCmd.ValidArgsFunction = completeEcosystems
	}

	// Domain commands
	if getDomainCmd != nil {
		getDomainCmd.ValidArgsFunction = completeDomains
	}
	if useDomainCmd != nil {
		useDomainCmd.ValidArgsFunction = completeDomains
	}
	if deleteDomainCmd != nil {
		deleteDomainCmd.ValidArgsFunction = completeDomains
	}

	// App commands
	if getAppCmd != nil {
		getAppCmd.ValidArgsFunction = completeApps
	}
	if useAppCmd != nil {
		useAppCmd.ValidArgsFunction = completeApps
	}
	if deleteAppCmd != nil {
		deleteAppCmd.ValidArgsFunction = completeApps
	}

	// Workspace commands
	if getWorkspaceCmd != nil {
		getWorkspaceCmd.ValidArgsFunction = completeWorkspaces
	}
	if useWorkspaceCmd != nil {
		useWorkspaceCmd.ValidArgsFunction = completeWorkspaces
	}
	if deleteWorkspaceCmd != nil {
		deleteWorkspaceCmd.ValidArgsFunction = completeWorkspaces
	}

	// Register flag completions for commands with hierarchy flags
	registerAllHierarchyFlagCompletions()
}

// registerAllHierarchyFlagCompletions registers flag completions for all commands
// that use --ecosystem, --domain, --app, or --workspace flags.
func registerAllHierarchyFlagCompletions() {
	// Commands with full hierarchy flags (use registerHierarchyFlagCompletions)
	commandsWithFullHierarchy := []*cobra.Command{
		attachCmd,
		buildCmd,
		detachCmd,
		getWorkspacesCmd,
		getWorkspaceCmd,
	}

	for _, cmd := range commandsWithFullHierarchy {
		if cmd != nil {
			registerHierarchyFlagCompletions(cmd)
		}
	}

	// Commands with specific hierarchy flags
	// Domain commands with --ecosystem flag
	if getDomainsCmd != nil {
		getDomainsCmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
	}
	if getDomainCmd != nil {
		getDomainCmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
	}
	if deleteDomainCmd != nil {
		deleteDomainCmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
	}

	// App commands with --domain flag
	if getAppsCmd != nil {
		getAppsCmd.RegisterFlagCompletionFunc("domain", completeDomains)
	}
	if getAppCmd != nil {
		getAppCmd.RegisterFlagCompletionFunc("domain", completeDomains)
	}
	if deleteAppCmd != nil {
		deleteAppCmd.RegisterFlagCompletionFunc("domain", completeDomains)
	}

	// Workspace commands with --app flag
	if deleteWorkspaceCmd != nil {
		deleteWorkspaceCmd.RegisterFlagCompletionFunc("app", completeApps)
	}

	// set theme command with hierarchy flags
	if setThemeCmd != nil {
		setThemeCmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
		setThemeCmd.RegisterFlagCompletionFunc("domain", completeDomains)
		setThemeCmd.RegisterFlagCompletionFunc("app", completeApps)
		setThemeCmd.RegisterFlagCompletionFunc("workspace", completeWorkspaces)
	}
}
