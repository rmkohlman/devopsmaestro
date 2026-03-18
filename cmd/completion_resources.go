package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
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

// ---------------------------------------------------------------------------
// Completion functions for each resource type
// ---------------------------------------------------------------------------

// Core hierarchy resources

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

// Supporting resources (have registered handlers)

func completeCredentials(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Credential")
}

func completeRegistries(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Registry")
}

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

func completeTerminalPrompts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "TerminalPrompt")
}

// GitRepo completions — GitRepo has no handler in the resource registry,
// so we query the DataStore directly instead of using completeResources().
func completeGitRepos(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ds, err := getCompletionDataStore(cmd)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	repos, err := ds.ListGitRepos()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	completions := make([]string, 0, len(repos))
	for _, r := range repos {
		completions = append(completions, fmt.Sprintf("%s\t%s", r.Name, r.URL))
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Static completions for registry type aliases (oci, pypi, npm, go, http)
func completeRegistryTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	types := []string{
		"oci\tOCI container registry (Zot)",
		"pypi\tPython package index (devpi)",
		"npm\tNode.js package registry (Verdaccio)",
		"go\tGo module proxy (Athens)",
		"http\tHTTP caching proxy (Squid)",
	}
	return types, cobra.ShellCompDirectiveNoFileComp
}

// Multi-arg completion for registrySetDefaultCmd: arg[0]=type, arg[1]=registry name
func completeRegistrySetDefault(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return completeRegistryTypes(cmd, args, toComplete)
	case 1:
		return completeRegistries(cmd, args, toComplete)
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// ---------------------------------------------------------------------------
// Helper: get DataStore for completion context
// ---------------------------------------------------------------------------

// getCompletionDataStore gets a DataStore from context or creates a default one.
// Used by completion functions that query the DataStore directly (e.g. GitRepo).
func getCompletionDataStore(cmd *cobra.Command) (db.DataStore, error) {
	ctx := cmd.Context()
	if dataStore := ctx.Value("dataStore"); dataStore != nil {
		if ds, ok := dataStore.(db.DataStore); ok {
			return ds, nil
		}
	}
	return createDefaultDataStore()
}

// ---------------------------------------------------------------------------
// Registration helpers
// ---------------------------------------------------------------------------

// registerHierarchyFlagCompletions registers completion functions for hierarchy flags.
// Call this from command init() functions after AddHierarchyFlags().
func registerHierarchyFlagCompletions(cmd *cobra.Command) {
	cmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
	cmd.RegisterFlagCompletionFunc("domain", completeDomains)
	cmd.RegisterFlagCompletionFunc("app", completeApps)
	cmd.RegisterFlagCompletionFunc("workspace", completeWorkspaces)
}

// registerCredentialScopeFlagCompletions registers completion functions for
// credential scope flags (--ecosystem, --domain, --app, --workspace).
// These use the same flag names as hierarchy flags.
func registerCredentialScopeFlagCompletions(cmd *cobra.Command) {
	registerHierarchyFlagCompletions(cmd)
}

// registerWorkspaceAppFlagCompletions registers completions for --workspace and --app flags.
// Used by nvim and terminal set/get commands that operate on workspace-scoped resources.
func registerWorkspaceAppFlagCompletions(cmd *cobra.Command) {
	cmd.RegisterFlagCompletionFunc("workspace", completeWorkspaces)
	cmd.RegisterFlagCompletionFunc("app", completeApps)
}

// ---------------------------------------------------------------------------
// registerAllResourceCompletions registers ValidArgsFunction for all commands
// that accept resource names as positional arguments.
// This is called from registerDynamicCompletions() in completion.go.
// ---------------------------------------------------------------------------
func registerAllResourceCompletions() {
	// === Core hierarchy: get/use/delete ===

	// Ecosystem commands
	for _, cmd := range []*cobra.Command{getEcosystemCmd, useEcosystemCmd, deleteEcosystemCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeEcosystems
		}
	}

	// Domain commands
	for _, cmd := range []*cobra.Command{getDomainCmd, useDomainCmd, deleteDomainCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeDomains
		}
	}

	// App commands
	for _, cmd := range []*cobra.Command{getAppCmd, useAppCmd, deleteAppCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeApps
		}
	}

	// Workspace commands
	for _, cmd := range []*cobra.Command{getWorkspaceCmd, useWorkspaceCmd, deleteWorkspaceCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeWorkspaces
		}
	}

	// === Credential commands ===
	for _, cmd := range []*cobra.Command{getCredentialCmd, deleteCredentialCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeCredentials
		}
	}

	// === Registry commands ===
	for _, cmd := range []*cobra.Command{
		getRegistryCmd, deleteRegistryCmd,
		startRegistryCmd, stopRegistryCmd,
		rolloutRestartRegistryCmd, rolloutStatusRegistryCmd,
		rolloutHistoryRegistryCmd, rolloutUndoRegistryCmd,
	} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeRegistries
		}
	}

	// === GitRepo commands ===
	for _, cmd := range []*cobra.Command{getGitRepoCmd, deleteGitRepoCmd, syncGitRepoCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeGitRepos
		}
	}

	// === Nvim commands ===
	for _, cmd := range []*cobra.Command{nvimGetPluginCmd, setNvimPluginCmd, editNvimPluginCmd, deleteNvimPluginCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeNvimPlugins
		}
	}
	for _, cmd := range []*cobra.Command{nvimGetThemeCmd, editNvimThemeCmd, deleteNvimThemeCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeNvimThemes
		}
	}
	for _, cmd := range []*cobra.Command{nvimGetPackageCmd, useNvimPackageCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeNvimPackages
		}
	}

	// === Terminal commands ===
	if terminalGetPackageCmd != nil {
		terminalGetPackageCmd.ValidArgsFunction = completeTerminalPackages
	}
	if setTerminalPackageCmd != nil {
		setTerminalPackageCmd.ValidArgsFunction = completeTerminalPackages
	}
	if useTerminalPackageCmd != nil {
		useTerminalPackageCmd.ValidArgsFunction = completeTerminalPackages
	}
	if setTerminalPromptCmd != nil {
		setTerminalPromptCmd.ValidArgsFunction = completeTerminalPrompts
	}

	// === Set theme command (positional arg = theme name from NvimTheme library) ===
	if setThemeCmd != nil {
		setThemeCmd.ValidArgsFunction = completeNvimThemes
	}

	// === Registry config commands ===
	if registrySetDefaultCmd != nil {
		registrySetDefaultCmd.ValidArgsFunction = completeRegistrySetDefault
	}
	if registryEnableCmd != nil {
		registryEnableCmd.ValidArgsFunction = completeRegistryTypes
	}
	if registryDisableCmd != nil {
		registryDisableCmd.ValidArgsFunction = completeRegistryTypes
	}

	// Register all flag completions
	registerAllFlagCompletions()
}

// ---------------------------------------------------------------------------
// registerAllFlagCompletions registers flag completions for all commands
// that use --ecosystem, --domain, --app, --workspace, --repo, --credential, etc.
// ---------------------------------------------------------------------------
func registerAllFlagCompletions() {
	// === Commands with full hierarchy flags (-e/-d/-a/-w) ===
	for _, cmd := range []*cobra.Command{
		attachCmd,
		buildCmd,
		detachCmd,
		getWorkspacesCmd,
		getWorkspaceCmd,
	} {
		if cmd != nil {
			registerHierarchyFlagCompletions(cmd)
		}
	}

	// === Domain commands with --ecosystem flag ===
	for _, cmd := range []*cobra.Command{getDomainsCmd, getDomainCmd, deleteDomainCmd, createDomainCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
		}
	}

	// === App commands with --domain flag ===
	for _, cmd := range []*cobra.Command{getAppsCmd, getAppCmd, deleteAppCmd, createAppCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("domain", completeDomains)
		}
	}

	// === Workspace commands with --app flag ===
	for _, cmd := range []*cobra.Command{deleteWorkspaceCmd, createWorkspaceCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("app", completeApps)
		}
	}

	// === Create commands with --repo flag ===
	for _, cmd := range []*cobra.Command{createAppCmd, createWorkspaceCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("repo", completeGitRepos)
		}
	}

	// === Create git repo with --credential flag ===
	if createGitRepoCmd != nil {
		createGitRepoCmd.RegisterFlagCompletionFunc("credential", completeCredentials)
	}

	// === Create branch with --workspace and --app flags ===
	if createBranchCmd != nil {
		registerWorkspaceAppFlagCompletions(createBranchCmd)
	}

	// === Credential commands with scope flags (--ecosystem, --domain, --app, --workspace) ===
	for _, cmd := range []*cobra.Command{createCredentialCmd, getCredentialsCmd, getCredentialCmd, deleteCredentialCmd} {
		if cmd != nil {
			registerCredentialScopeFlagCompletions(cmd)
		}
	}

	// === Set theme command with hierarchy flags ===
	if setThemeCmd != nil {
		setThemeCmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
		setThemeCmd.RegisterFlagCompletionFunc("domain", completeDomains)
		setThemeCmd.RegisterFlagCompletionFunc("app", completeApps)
		setThemeCmd.RegisterFlagCompletionFunc("workspace", completeWorkspaces)
	}

	// === Nvim commands with --workspace and --app flags ===
	for _, cmd := range []*cobra.Command{nvimGetPluginsCmd, nvimGetPluginCmd, setNvimPluginCmd, deleteNvimPluginCmd, deleteNvimThemeCmd} {
		if cmd != nil {
			registerWorkspaceAppFlagCompletions(cmd)
		}
	}

	// === Terminal commands with --workspace and --app flags ===
	for _, cmd := range []*cobra.Command{setTerminalPromptCmd, setTerminalPluginCmd, setTerminalPackageCmd} {
		if cmd != nil {
			registerWorkspaceAppFlagCompletions(cmd)
		}
	}
}
