package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

// getAllCmd shows all resources across the system.
var getAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Show all resources",
	Long: `Show a summary of all resources across the system.

Displays all ecosystems, domains, apps, workspaces, credentials,
registries, git repos, nvim plugins, and nvim themes.

By default, resources are scoped to the active context (ecosystem, domain,
or app). Use flags to override the scope:

  -e, --ecosystem   Filter to a specific ecosystem
  -d, --domain      Filter to a specific domain (requires ecosystem)
  -a, --app         Filter to a specific app (requires domain)
  -A, --all         Show all resources (ignore active context)

Global resources (registries, nvim plugins, nvim themes) are always shown
regardless of scope.

Examples:
  dvm get all              # Show resources in active scope
  dvm get all -A           # Show all resources (ignore context)
  dvm get all -e prod      # Show resources in 'prod' ecosystem
  dvm get all -e prod -d backend  # Show resources in 'backend' domain
  dvm get all -o wide      # Show additional columns
  dvm get all -o json      # Output as JSON
  dvm get all -o yaml      # Output as YAML
  dvm get all -o table     # Output as plain table`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAll(cmd)
	},
}

func getAll(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("failed to get data store: %w", err)
	}

	// Read scoping flags (may be absent on bare test commands, so ignore errors)
	ecoFlag, _ := cmd.Flags().GetString("ecosystem")
	domFlag, _ := cmd.Flags().GetString("domain")
	appFlag, _ := cmd.Flags().GetString("app")
	allFlag, _ := cmd.Flags().GetBool("all")

	// Resolve scope
	scope, err := resolveGetAllScope(ds, ecoFlag, domFlag, appFlag, allFlag)
	if err != nil {
		return err
	}

	// Collect all resources, treating errors as empty sections
	ecosystems, err := ds.ListEcosystems()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list ecosystems: %v", err))
		ecosystems = nil
	}

	domains, err := ds.ListAllDomains()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list domains: %v", err))
		domains = nil
	}

	apps, err := ds.ListAllApps()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list apps: %v", err))
		apps = nil
	}

	workspaces, err := ds.ListAllWorkspaces()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list workspaces: %v", err))
		workspaces = nil
	}

	credentials, err := ds.ListAllCredentials()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list credentials: %v", err))
		credentials = nil
	}

	// Global resources — always shown regardless of scope
	registries, err := ds.ListRegistries()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list registries: %v", err))
		registries = nil
	}

	gitRepos, err := ds.ListGitRepos()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list git repos: %v", err))
		gitRepos = nil
	}

	plugins, err := ds.ListPlugins()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list nvim plugins: %v", err))
		plugins = nil
	}

	themes, err := ds.ListThemes()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list nvim themes: %v", err))
		themes = nil
	}

	nvimPackages, err := ds.ListPackages()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list nvim packages: %v", err))
		nvimPackages = nil
	}

	terminalPrompts, err := ds.ListTerminalPrompts()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list terminal prompts: %v", err))
		terminalPrompts = nil
	}

	terminalPackages, err := ds.ListTerminalPackages()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list terminal packages: %v", err))
		terminalPackages = nil
	}

	// Apply scope filtering to hierarchical resources
	if !scope.ShowAll {
		ecosystems = filterEcosystems(ecosystems, scope)
		domains = filterDomains(domains, scope)
		apps = filterApps(apps, scope, domains)
		workspaces = filterWorkspaces(workspaces, scope, apps)
		credentials = filterCredentials(credentials, scope)
		gitRepos = filterGitRepos(gitRepos, scope, apps)
	}

	// JSON/YAML: build a kubectl-style kind: List document via resource.BuildList
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		// Warn when exporting YAML/JSON in a scoped context (global resources excluded)
		if !scope.ShowAll {
			render.Warning("Warning: Scoped export excludes global resources (GitRepos, Registries, NvimPlugins, NvimThemes, NvimPackages, TerminalPrompts, TerminalPackages, GlobalDefaults). Use -A for a complete backup.")
		}

		// Ensure all resource handlers are registered
		handlers.RegisterAll()

		// Build parent name lookup maps for hierarchical resources
		ecoNames := make(map[int]string)
		for _, e := range ecosystems {
			ecoNames[e.ID] = e.Name
		}
		domNames := make(map[int]string)
		domEcoIDs := make(map[int]int)
		for _, d := range domains {
			domNames[d.ID] = d.Name
			domEcoIDs[d.ID] = d.EcosystemID
		}
		appNames := make(map[int]string)
		appDomIDs := make(map[int]int)
		for _, a := range apps {
			appNames[a.ID] = a.Name
			appDomIDs[a.ID] = a.DomainID
		}

		// Workspace name lookup (for credential scope resolution)
		wsNames := make(map[int]string)
		for _, w := range workspaces {
			wsNames[w.ID] = w.Name
		}

		// GitRepo name lookup (for workspace export)
		gitRepoNames := make(map[int64]string)
		for i := range gitRepos {
			gitRepoNames[int64(gitRepos[i].ID)] = gitRepos[i].Name
		}

		// Collect resources in dependency order (DependencyOrder from resource package)
		var allResources []resource.Resource

		// GlobalDefaults — prepend so they're applied first during restore.
		// Only when unscoped (global-level configuration).
		if scope.ShowAll {
			resCtx := resource.Context{DataStore: ds}
			if gdRes, err := resource.List(resCtx, handlers.KindGlobalDefaults); err == nil {
				allResources = append(allResources, gdRes...)
			}
		}

		// Ecosystems
		for _, e := range ecosystems {
			allResources = append(allResources, handlers.NewEcosystemResource(e))
		}

		// Domains (need parent ecosystem name)
		for _, d := range domains {
			allResources = append(allResources, handlers.NewDomainResource(d, ecoNames[d.EcosystemID]))
		}

		// Apps (need parent domain name + ecosystem name for context-free apply)
		for _, a := range apps {
			ecoName := ecoNames[domEcoIDs[a.DomainID]]
			allResources = append(allResources, handlers.NewAppResource(a, domNames[a.DomainID], ecoName))
		}

		// GitRepos — global but filtered; include only when unscoped
		if scope.ShowAll {
			for i := range gitRepos {
				allResources = append(allResources, handlers.NewGitRepoResource(&gitRepos[i]))
			}
		}

		// Registries — global; include only when unscoped
		if scope.ShowAll {
			for _, r := range registries {
				allResources = append(allResources, handlers.NewRegistryResource(r))
			}
		}

		// Credentials (resolve scopeName from precomputed maps)
		for _, c := range credentials {
			scopeName := resolveCredScopeName(c, ecoNames, domNames, appNames, wsNames)
			allResources = append(allResources, handlers.NewCredentialResource(c, scopeName))
		}

		// Workspaces (need parent app name + resolve domain/gitrepo names)
		for _, w := range workspaces {
			domName := domNames[appDomIDs[w.AppID]]
			grName := ""
			if w.GitRepoID.Valid {
				grName = gitRepoNames[w.GitRepoID.Int64]
			}
			allResources = append(allResources, handlers.NewWorkspaceResource(w, appNames[w.AppID], domName, grName))
		}

		// Global resources using handler List() — only when unscoped (WI-4)
		if scope.ShowAll {
			resCtx := resource.Context{DataStore: ds}

			// NvimPlugins
			if pluginRes, err := resource.List(resCtx, handlers.KindNvimPlugin); err == nil {
				allResources = append(allResources, pluginRes...)
			}

			// NvimThemes
			if themeRes, err := resource.List(resCtx, handlers.KindNvimTheme); err == nil {
				allResources = append(allResources, themeRes...)
			}

			// NvimPackages (WI-5)
			if pkgRes, err := resource.List(resCtx, handlers.KindNvimPackage); err == nil {
				allResources = append(allResources, pkgRes...)
			}

			// TerminalPrompts (WI-5)
			if promptRes, err := resource.List(resCtx, "TerminalPrompt"); err == nil {
				allResources = append(allResources, promptRes...)
			}

			// TerminalPackages (WI-5)
			if termPkgRes, err := resource.List(resCtx, handlers.KindTerminalPackage); err == nil {
				allResources = append(allResources, termPkgRes...)
			}
		}

		resCtx := resource.Context{DataStore: ds}
		list, err := resource.BuildList(resCtx, allResources)
		if err != nil {
			return fmt.Errorf("failed to build resource list: %w", err)
		}

		return render.OutputWith(getOutputFormat, list, render.Options{Type: render.TypeAuto})
	}

	// Human-readable output: render each section using shared table builders

	// Fetch active context for markers (ignore errors - no active context is fine)
	var activeEcoID, activeDomID, activeAppID *int
	var activeWorkspaceName string
	if dbCtx, ctxErr := ds.GetContext(); ctxErr == nil && dbCtx != nil {
		activeEcoID = dbCtx.ActiveEcosystemID
		activeDomID = dbCtx.ActiveDomainID
		activeAppID = dbCtx.ActiveAppID
		if dbCtx.ActiveWorkspaceID != nil {
			if ws, wsErr := ds.GetWorkspaceByID(*dbCtx.ActiveWorkspaceID); wsErr == nil {
				activeWorkspaceName = ws.Name
			}
		}
	}

	wide := getOutputFormat == "wide"

	// === Ecosystems ===
	render.Info(fmt.Sprintf("=== Ecosystems (%d) ===", len(ecosystems)))
	if len(ecosystems) > 0 {
		b := &ecosystemTableBuilder{ActiveID: activeEcoID}
		td := BuildTable(b, ecosystems, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Domains ===
	render.Info(fmt.Sprintf("=== Domains (%d) ===", len(domains)))
	if len(domains) > 0 {
		b := &domainTableBuilder{DataStore: ds, ActiveID: activeDomID}
		td := BuildTable(b, domains, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Apps ===
	render.Info(fmt.Sprintf("=== Apps (%d) ===", len(apps)))
	if len(apps) > 0 {
		b := &appTableBuilder{DataStore: ds, ActiveID: activeAppID}
		td := BuildTable(b, apps, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Workspaces ===
	render.Info(fmt.Sprintf("=== Workspaces (%d) ===", len(workspaces)))
	if len(workspaces) > 0 {
		b := &workspaceTableBuilder{DataStore: ds, ActiveWorkspaceName: activeWorkspaceName}
		td := BuildTable(b, workspaces, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Credentials ===
	render.Info(fmt.Sprintf("=== Credentials (%d) ===", len(credentials)))
	if len(credentials) > 0 {
		b := &credentialTableBuilder{DataStore: ds}
		td := BuildTable(b, credentials, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Registries ===
	render.Info(fmt.Sprintf("=== Registries (%d) ===", len(registries)))
	if len(registries) > 0 {
		b := &registryTableBuilder{StatusMap: nil}
		td := BuildTable(b, registries, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Git Repos ===
	render.Info(fmt.Sprintf("=== Git Repos (%d) ===", len(gitRepos)))
	if len(gitRepos) > 0 {
		// ListGitRepos returns []models.GitRepoDB (value type), but
		// gitRepoTableBuilder.Row expects *models.GitRepoDB, so convert.
		gitRepoPtrs := make([]*models.GitRepoDB, len(gitRepos))
		for i := range gitRepos {
			gitRepoPtrs[i] = &gitRepos[i]
		}
		b := &gitRepoTableBuilder{}
		td := BuildTable(b, gitRepoPtrs, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Nvim Plugins ===
	render.Info(fmt.Sprintf("=== Nvim Plugins (%d) ===", len(plugins)))
	if len(plugins) > 0 {
		b := &nvimPluginTableBuilder{}
		td := BuildTable(b, plugins, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Nvim Themes ===
	render.Info(fmt.Sprintf("=== Nvim Themes (%d) ===", len(themes)))
	if len(themes) > 0 {
		b := &nvimThemeTableBuilder{}
		td := BuildTable(b, themes, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Nvim Packages ===
	render.Info(fmt.Sprintf("=== Nvim Packages (%d) ===", len(nvimPackages)))
	if len(nvimPackages) > 0 {
		b := &nvimPackageTableBuilder{}
		td := BuildTable(b, nvimPackages, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Terminal Prompts ===
	render.Info(fmt.Sprintf("=== Terminal Prompts (%d) ===", len(terminalPrompts)))
	if len(terminalPrompts) > 0 {
		b := &terminalPromptTableBuilder{}
		td := BuildTable(b, terminalPrompts, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Terminal Packages ===
	render.Info(fmt.Sprintf("=== Terminal Packages (%d) ===", len(terminalPackages)))
	if len(terminalPackages) > 0 {
		b := &terminalPackageTableBuilder{}
		td := BuildTable(b, terminalPackages, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === CA Certs ===
	// Gather CA certs across all scopes for the summary table
	var allCACerts []scopedCACert
	globalCerts, certsErr := GetGlobalCACerts(ds)
	if certsErr == nil {
		for _, c := range globalCerts {
			allCACerts = append(allCACerts, scopedCACert{Name: c.Name, Scope: "global"})
		}
	}
	for _, eco := range ecosystems {
		ecoYAML := eco.ToYAML(nil)
		for _, c := range ecoYAML.Spec.CACerts {
			allCACerts = append(allCACerts, scopedCACert{Name: c.Name, Scope: fmt.Sprintf("ecosystem: %s", eco.Name)})
		}
	}
	for _, dom := range domains {
		ecoName := ""
		if eco, err := ds.GetEcosystemByID(dom.EcosystemID); err == nil {
			ecoName = eco.Name
		}
		domYAML := dom.ToYAML(ecoName, nil)
		for _, c := range domYAML.Spec.CACerts {
			allCACerts = append(allCACerts, scopedCACert{Name: c.Name, Scope: fmt.Sprintf("domain: %s", dom.Name)})
		}
	}
	for _, app := range apps {
		buildConfig := app.GetBuildConfig()
		if buildConfig == nil {
			continue
		}
		for _, c := range buildConfig.CACerts {
			allCACerts = append(allCACerts, scopedCACert{Name: c.Name, Scope: fmt.Sprintf("app: %s", app.Name)})
		}
	}
	render.Info(fmt.Sprintf("=== CA Certs (%d) ===", len(allCACerts)))
	if len(allCACerts) > 0 {
		b := &caCertTableBuilder{}
		td := BuildTable(b, allCACerts, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Build Args ===
	// Gather build args across all scopes for the summary table
	var allBuildArgs []scopedBuildArg
	globalArgs, argsErr := GetGlobalBuildArgs(ds)
	if argsErr == nil {
		for k := range globalArgs {
			allBuildArgs = append(allBuildArgs, scopedBuildArg{Key: k, Scope: "global"})
		}
	}
	for _, eco := range ecosystems {
		ecoYAML := eco.ToYAML(nil)
		for k := range ecoYAML.Spec.Build.Args {
			allBuildArgs = append(allBuildArgs, scopedBuildArg{Key: k, Scope: fmt.Sprintf("ecosystem: %s", eco.Name)})
		}
	}
	for _, dom := range domains {
		ecoName := ""
		if eco, err := ds.GetEcosystemByID(dom.EcosystemID); err == nil {
			ecoName = eco.Name
		}
		domYAML := dom.ToYAML(ecoName, nil)
		for k := range domYAML.Spec.Build.Args {
			allBuildArgs = append(allBuildArgs, scopedBuildArg{Key: k, Scope: fmt.Sprintf("domain: %s", dom.Name)})
		}
	}
	for _, app := range apps {
		buildConfig := app.GetBuildConfig()
		if buildConfig == nil {
			continue
		}
		for k := range buildConfig.Args {
			allBuildArgs = append(allBuildArgs, scopedBuildArg{Key: k, Scope: fmt.Sprintf("app: %s", app.Name)})
		}
	}
	render.Info(fmt.Sprintf("=== Build Args (%d) ===", len(allBuildArgs)))
	if len(allBuildArgs) > 0 {
		b := &buildArgTableBuilder{}
		td := BuildTable(b, allBuildArgs, wide)
		renderTable(td)
	} else {
		render.Plainf("  (none)")
	}

	return nil
}

// scopeContext holds the resolved scope for a get all operation.
type scopeContext struct {
	EcosystemID *int
	DomainID    *int
	AppID       *int
	ShowAll     bool
}

// resolveGetAllScope resolves the scope for a get all operation based on flags and active context.
// Priority: -A flag > explicit flags (-e/-d/-a) > active context > ShowAll fallback
func resolveGetAllScope(ds db.DataStore, ecosystem, domain, app string, showAll bool) (*scopeContext, error) {
	// -A flag: show everything, but conflicts with explicit flags
	if showAll {
		if ecosystem != "" || domain != "" || app != "" {
			return nil, fmt.Errorf("--all/-A cannot be combined with --ecosystem, --domain, or --app flags")
		}
		return &scopeContext{ShowAll: true}, nil
	}

	sc := &scopeContext{}

	// Resolve ecosystem: explicit flag or active context
	var ecoID int
	if ecosystem != "" {
		eco, err := ds.GetEcosystemByName(ecosystem)
		if err != nil {
			return nil, fmt.Errorf("ecosystem not found: %s", ecosystem)
		}
		ecoID = eco.ID
		sc.EcosystemID = &ecoID
	}

	// Resolve domain: requires ecosystem (explicit or context)
	if domain != "" {
		// Domain needs an ecosystem to be resolved
		if sc.EcosystemID == nil {
			// Try to get ecosystem from active context
			dbCtx, err := ds.GetContext()
			if err != nil || dbCtx == nil || dbCtx.ActiveEcosystemID == nil {
				return nil, fmt.Errorf("--domain requires --ecosystem or an active ecosystem context")
			}
			ecoID = *dbCtx.ActiveEcosystemID
			sc.EcosystemID = &ecoID
		}

		dom, err := ds.GetDomainByName(*sc.EcosystemID, domain)
		if err != nil {
			return nil, fmt.Errorf("domain not found: %s", domain)
		}
		domID := dom.ID
		sc.DomainID = &domID
	}

	// Resolve app: requires domain (explicit or context)
	if app != "" {
		if sc.DomainID == nil {
			// Try to get domain from active context
			dbCtx, err := ds.GetContext()
			if err != nil || dbCtx == nil || dbCtx.ActiveDomainID == nil {
				return nil, fmt.Errorf("--app requires --domain or an active domain context")
			}
			domID := *dbCtx.ActiveDomainID
			sc.DomainID = &domID

			// Also ensure we have ecosystem set
			if sc.EcosystemID == nil {
				dom, err := ds.GetDomainByID(domID)
				if err == nil {
					ecoID := dom.EcosystemID
					sc.EcosystemID = &ecoID
				}
			}
		}

		a, err := ds.GetAppByName(*sc.DomainID, app)
		if err != nil {
			return nil, fmt.Errorf("app not found: %s", app)
		}
		appID := a.ID
		sc.AppID = &appID
	}

	// If no explicit flags, fall back to active context
	if ecosystem == "" && domain == "" && app == "" {
		dbCtx, err := ds.GetContext()
		if err != nil || dbCtx == nil {
			sc.ShowAll = true
			return sc, nil
		}

		// Use deepest active context level
		if dbCtx.ActiveAppID != nil {
			appID := *dbCtx.ActiveAppID
			sc.AppID = &appID

			// Walk up to get domain and ecosystem
			if a, err := ds.GetAppByID(appID); err == nil {
				domID := a.DomainID
				sc.DomainID = &domID
				if dom, err := ds.GetDomainByID(domID); err == nil {
					ecoID := dom.EcosystemID
					sc.EcosystemID = &ecoID
				}
			}
		} else if dbCtx.ActiveDomainID != nil {
			domID := *dbCtx.ActiveDomainID
			sc.DomainID = &domID
			if dom, err := ds.GetDomainByID(domID); err == nil {
				ecoID := dom.EcosystemID
				sc.EcosystemID = &ecoID
			}
		} else if dbCtx.ActiveEcosystemID != nil {
			ecoID := *dbCtx.ActiveEcosystemID
			sc.EcosystemID = &ecoID
		} else {
			// No active context at all
			sc.ShowAll = true
		}
	}

	return sc, nil
}

// ---------------------------------------------------------------------------
// Scope filter helpers for getAll
// ---------------------------------------------------------------------------

// filterEcosystems filters ecosystems to only those matching the scope.
func filterEcosystems(ecosystems []*models.Ecosystem, sc *scopeContext) []*models.Ecosystem {
	if sc.EcosystemID == nil {
		return ecosystems
	}
	var filtered []*models.Ecosystem
	for _, e := range ecosystems {
		if e.ID == *sc.EcosystemID {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// filterDomains filters domains to those in the scoped ecosystem and/or domain.
func filterDomains(domains []*models.Domain, sc *scopeContext) []*models.Domain {
	if sc.EcosystemID == nil {
		return domains
	}
	var filtered []*models.Domain
	for _, d := range domains {
		if d.EcosystemID != *sc.EcosystemID {
			continue
		}
		if sc.DomainID != nil && d.ID != *sc.DomainID {
			continue
		}
		filtered = append(filtered, d)
	}
	return filtered
}

// filterApps filters apps to those in the scoped domain and/or app.
// filteredDomains should be the already-filtered domain list (scoped to the
// target ecosystem) so that ecosystem-only scoping can determine which apps
// belong to the ecosystem via their DomainID.
func filterApps(apps []*models.App, sc *scopeContext, filteredDomains []*models.Domain) []*models.App {
	if sc.DomainID == nil && sc.EcosystemID == nil {
		return apps
	}

	// Build set of allowed domain IDs from the already-filtered domains
	allowedDomains := make(map[int]bool)
	for _, d := range filteredDomains {
		allowedDomains[d.ID] = true
	}

	var filtered []*models.App
	for _, a := range apps {
		if sc.AppID != nil {
			// If we have a specific app scope, only show that app
			if a.ID == *sc.AppID {
				filtered = append(filtered, a)
			}
			continue
		}
		if sc.DomainID != nil {
			if a.DomainID == *sc.DomainID {
				filtered = append(filtered, a)
			}
			continue
		}
		// Ecosystem scope only: include only apps whose domain is in the
		// filtered domain set (i.e., belongs to the scoped ecosystem).
		if allowedDomains[a.DomainID] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// filterWorkspaces filters workspaces to those belonging to the filtered apps.
func filterWorkspaces(workspaces []*models.Workspace, sc *scopeContext, filteredApps []*models.App) []*models.Workspace {
	if sc.EcosystemID == nil && sc.DomainID == nil && sc.AppID == nil {
		return workspaces
	}

	// Build allowed app ID set
	allowedApps := make(map[int]bool)
	for _, a := range filteredApps {
		allowedApps[a.ID] = true
	}

	var filtered []*models.Workspace
	for _, w := range workspaces {
		if allowedApps[w.AppID] {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

// filterCredentials filters credentials to those scoped within the hierarchy.
func filterCredentials(credentials []*models.CredentialDB, sc *scopeContext) []*models.CredentialDB {
	if sc.EcosystemID == nil {
		return credentials
	}

	var filtered []*models.CredentialDB
	for _, c := range credentials {
		switch c.ScopeType {
		case models.CredentialScopeEcosystem:
			if sc.EcosystemID != nil && c.ScopeID == int64(*sc.EcosystemID) {
				filtered = append(filtered, c)
			}
		case models.CredentialScopeDomain:
			if sc.DomainID != nil && c.ScopeID == int64(*sc.DomainID) {
				filtered = append(filtered, c)
			}
		case models.CredentialScopeApp:
			if sc.AppID != nil && c.ScopeID == int64(*sc.AppID) {
				filtered = append(filtered, c)
			}
		case models.CredentialScopeWorkspace:
			// Workspace-scoped credentials are shown when viewing the parent app scope
			// For now, include them if we have any scope narrower than ecosystem
			if sc.AppID != nil {
				filtered = append(filtered, c)
			}
		}
	}
	return filtered
}

// filterGitRepos filters git repos. Git repos are currently global, but if scoped
// we show all of them (they don't have a direct hierarchy relationship yet).
func filterGitRepos(gitRepos []models.GitRepoDB, sc *scopeContext, filteredApps []*models.App) []models.GitRepoDB {
	// Git repos don't have ecosystem/domain/app scoping in the DB schema yet,
	// so we show all of them regardless of scope (they're effectively global).
	return gitRepos
}

// resolveCredScopeName resolves a credential's scope ID to a human-readable name
// using precomputed lookup maps (avoids N+1 DB queries).
func resolveCredScopeName(c *models.CredentialDB, ecoNames, domNames, appNames, wsNames map[int]string) string {
	switch c.ScopeType {
	case models.CredentialScopeEcosystem:
		return ecoNames[int(c.ScopeID)]
	case models.CredentialScopeDomain:
		return domNames[int(c.ScopeID)]
	case models.CredentialScopeApp:
		return appNames[int(c.ScopeID)]
	case models.CredentialScopeWorkspace:
		return wsNames[int(c.ScopeID)]
	default:
		return ""
	}
}
