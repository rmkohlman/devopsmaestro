package cmd

import (
	"fmt"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/resource"
	"github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

// completeResources is a generic helper that uses the Resource/Handler pattern
// to provide dynamic completions for resource names with descriptions.
func completeResources(cmd *cobra.Command, kind string) ([]string, cobra.ShellCompDirective) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	// Try to get datastore from context first (normal command context)
	ctx := cmd.Context()
	if dataStore := ctx.Value(CtxKeyDataStore); dataStore != nil {
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
	case interface{ System() *models.System }:
		system := r.System()
		if system.Description.Valid {
			return system.Description.String
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

func completeSystems(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "System")
}

func completeApps(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "App")
}

func completeWorkspaces(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
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

// completeAllThemes returns both library and user themes for completion.
// Unlike completeNvimThemes (which only returns user-stored themes via the
// NvimTheme handler), this function merges the 34+ built-in library themes
// with any user-defined themes, matching what `dvm get themes` displays.
// Used by setThemeCmd where users need to see all available themes.
func completeAllThemes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Start with library themes (always available, no DB needed)
	libraryInfos, err := library.List()
	if err != nil {
		libraryInfos = nil
	}

	// Build completion list; use a map to deduplicate library vs user
	byName := make(map[string]string, len(libraryInfos)) // name → description tab string
	var order []string

	for _, info := range libraryInfos {
		desc := info.Category
		if desc == "" {
			desc = "library"
		}
		byName[info.Name] = fmt.Sprintf("%s\t%s", info.Name, desc)
		order = append(order, info.Name)
	}

	// Try to add user themes from the DB (override library entries with same name)
	userCompletions, _ := completeResources(cmd, "NvimTheme")
	for _, comp := range userCompletions {
		// comp is either "name" or "name\tdescription"
		name, _, _ := strings.Cut(comp, "\t")
		if _, exists := byName[name]; !exists {
			order = append(order, name)
		}
		byName[name] = comp // user theme overrides library entry
	}

	completions := make([]string, 0, len(order))
	for _, name := range order {
		completions = append(completions, byName[name])
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
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
	if dataStore := ctx.Value(CtxKeyDataStore); dataStore != nil {
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
	cmd.RegisterFlagCompletionFunc("system", completeSystems)
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

	// System commands
	for _, cmd := range []*cobra.Command{getSystemCmd, useSystemCmd, deleteSystemCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeSystems
		}
	}

	// App commands
	for _, cmd := range []*cobra.Command{getAppCmd, useAppCmd, deleteAppCmd} {
		if cmd != nil {
			cmd.ValidArgsFunction = completeApps
		}
	}

	// Workspace commands
	for _, cmd := range []*cobra.Command{getWorkspaceCmd, useWorkspaceCmd, deleteWorkspaceCmd, attachCmd, buildCmd, detachCmd} {
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
	if setTerminalPackageWorkspaceCmd != nil {
		setTerminalPackageWorkspaceCmd.ValidArgsFunction = completeTerminalPackages
	}
	if useTerminalPackageCmd != nil {
		useTerminalPackageCmd.ValidArgsFunction = completeTerminalPackages
	}
	if setTerminalPromptCmd != nil {
		setTerminalPromptCmd.ValidArgsFunction = completeTerminalPrompts
	}

	// === Set theme command (positional arg = library + user themes) ===
	if setThemeCmd != nil {
		setThemeCmd.ValidArgsFunction = completeAllThemes
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

	// === System commands with --domain and --ecosystem flags ===
	for _, cmd := range []*cobra.Command{getSystemsCmd, getSystemCmd, deleteSystemCmd, createSystemCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("domain", completeDomains)
		}
	}
	for _, cmd := range []*cobra.Command{createSystemCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
		}
	}

	// === App commands with --domain and --system flags ===
	for _, cmd := range []*cobra.Command{getAppsCmd, getAppCmd, deleteAppCmd, createAppCmd} {
		if cmd != nil {
			cmd.RegisterFlagCompletionFunc("domain", completeDomains)
			cmd.RegisterFlagCompletionFunc("system", completeSystems)
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
	for _, cmd := range []*cobra.Command{setTerminalPromptCmd, setTerminalPluginCmd, setTerminalPackageWorkspaceCmd} {
		if cmd != nil {
			registerWorkspaceAppFlagCompletions(cmd)
		}
	}

	// === Build-arg commands with hierarchy flags (-e/-d/-a/-w) ===
	for _, cmd := range []*cobra.Command{setBuildArgCmd, deleteBuildArgCmd, getBuildArgsCmd} {
		if cmd != nil {
			registerHierarchyFlagCompletions(cmd)
		}
	}

	// === CA-cert commands with hierarchy flags (-e/-d/-a/-w) ===
	for _, cmd := range []*cobra.Command{setCACertCmd, deleteCACertCmd, getCACertsCmd} {
		if cmd != nil {
			registerHierarchyFlagCompletions(cmd)
		}
	}

	// === Get-all command (ecosystem, domain, app only — no workspace flag) ===
	if getAllCmd != nil {
		getAllCmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
		getAllCmd.RegisterFlagCompletionFunc("domain", completeDomains)
		getAllCmd.RegisterFlagCompletionFunc("app", completeApps)
	}
}
