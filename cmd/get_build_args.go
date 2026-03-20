// Package cmd provides the 'dvm get build-args' command for viewing hierarchical build args.
// Without --effective: shows args at the specified level only.
// With --effective:    shows the fully merged cascade result (requires --workspace).
package cmd

import (
	"fmt"
	"sort"

	"devopsmaestro/db"
	"devopsmaestro/pkg/buildargs/resolver"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

// Flags for get build-args command
var (
	getBuildArgsEcosystem string
	getBuildArgsDomain    string
	getBuildArgsApp       string
	getBuildArgsWorkspace string
	getBuildArgsGlobal    bool
	getBuildArgsEffective bool
)

// getBuildArgsCmd displays build args at a specific hierarchy level or the merged result
var getBuildArgsCmd = &cobra.Command{
	Use:   "build-args",
	Short: "Get build args at hierarchy level",
	Long: `Display build arguments at a specific hierarchy level, or show the merged cascade.

Without --effective: shows args set directly at the specified level only.
With --effective:    shows the fully merged cascade result from all levels
                     (requires --workspace to identify the hierarchy path).

The SOURCE column (in --effective mode) shows which level each key came from.

Build args cascade: global → ecosystem → domain → app → workspace (workspace wins)

For secrets, use 'dvm credential' instead — build args are stored in plain text.

Examples:
  dvm get build-args --global                        # Global defaults
  dvm get build-args --ecosystem my-eco              # Ecosystem-level args
  dvm get build-args --domain data-sci               # Domain-level args
  dvm get build-args --app ml-api                    # App-level args
  dvm get build-args --workspace dev                 # Workspace-level args
  dvm get build-args --workspace dev --effective     # Fully merged cascade
  dvm get build-args -A                              # All args across all scopes
  dvm get build-args -o json                         # Output as JSON`,
	RunE: runGetBuildArgs,
}

func init() {
	getCmd.AddCommand(getBuildArgsCmd)

	getBuildArgsCmd.Flags().StringVar(&getBuildArgsEcosystem, "ecosystem", "", "Get build args at ecosystem level")
	getBuildArgsCmd.Flags().StringVar(&getBuildArgsDomain, "domain", "", "Get build args at domain level")
	getBuildArgsCmd.Flags().StringVar(&getBuildArgsApp, "app", "", "Get build args at app level")
	getBuildArgsCmd.Flags().StringVar(&getBuildArgsWorkspace, "workspace", "", "Get build args at workspace level")
	getBuildArgsCmd.Flags().BoolVar(&getBuildArgsGlobal, "global", false, "Get global default build args")
	getBuildArgsCmd.Flags().BoolVar(&getBuildArgsEffective, "effective", false, "Show fully merged cascade result (requires --workspace)")
	getBuildArgsCmd.Flags().BoolP("all", "A", false, "List all build args across all scopes")
	// NOTE: --output/-o is inherited from getCmd PersistentFlags — do not re-register
}

func runGetBuildArgs(cmd *cobra.Command, args []string) error {
	allFlag, _ := cmd.Flags().GetBool("all")

	// --all cannot be combined with level-specific flags
	if allFlag {
		if getBuildArgsEcosystem != "" || getBuildArgsDomain != "" || getBuildArgsApp != "" ||
			getBuildArgsWorkspace != "" || getBuildArgsGlobal || getBuildArgsEffective {
			return fmt.Errorf("--all/-A cannot be combined with --global, --ecosystem, --domain, --app, --workspace, or --effective")
		}

		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}
		return runGetAllBuildArgs(cmd, ctx)
	}

	// Validate that at least one target flag is provided
	if getBuildArgsEcosystem == "" && getBuildArgsDomain == "" && getBuildArgsApp == "" &&
		getBuildArgsWorkspace == "" && !getBuildArgsGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, --global, or --all must be specified")
	}

	// --effective requires --workspace (to identify the full hierarchy path)
	if getBuildArgsEffective && getBuildArgsWorkspace == "" {
		return fmt.Errorf("--effective requires --workspace to identify the hierarchy path")
	}

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// --effective mode: show merged cascade
	if getBuildArgsEffective {
		return runGetBuildArgsEffective(cmd, ctx)
	}

	// Level-specific display
	switch {
	case getBuildArgsWorkspace != "":
		return getBuildArgsAtWorkspace(cmd, ctx, getBuildArgsWorkspace, getBuildArgsApp)
	case getBuildArgsApp != "":
		return getBuildArgsAtApp(cmd, ctx, getBuildArgsApp)
	case getBuildArgsDomain != "":
		return getBuildArgsAtDomain(cmd, ctx, getBuildArgsDomain)
	case getBuildArgsEcosystem != "":
		return getBuildArgsAtEcosystem(cmd, ctx, getBuildArgsEcosystem)
	case getBuildArgsGlobal:
		return getBuildArgsGlobalLevel(cmd, ctx)
	default:
		return fmt.Errorf("no hierarchy level specified")
	}
}

// runGetBuildArgsEffective resolves the full cascade and displays with SOURCE provenance.
func runGetBuildArgsEffective(cmd *cobra.Command, ctx resource.Context) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	// Resolve workspace to get its ID
	var workspaceID int
	if getBuildArgsApp != "" {
		app, err := ds.GetAppByNameGlobal(getBuildArgsApp)
		if err != nil {
			return fmt.Errorf("app %q not found: %w", getBuildArgsApp, err)
		}
		ws, err := ds.GetWorkspaceByName(app.ID, getBuildArgsWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q not found under app %q: %w", getBuildArgsWorkspace, getBuildArgsApp, err)
		}
		workspaceID = ws.ID
	} else {
		// Fall back to active app context
		activeApp, err := getActiveAppFromContext(ds)
		if err != nil {
			return fmt.Errorf("no app context. Use --app <name> or 'dvm use app <name>' first")
		}
		app, err := ds.GetAppByNameGlobal(activeApp)
		if err != nil {
			return fmt.Errorf("app %q not found: %w", activeApp, err)
		}
		ws, err := ds.GetWorkspaceByName(app.ID, getBuildArgsWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q not found under app %q: %w", getBuildArgsWorkspace, activeApp, err)
		}
		workspaceID = ws.ID
	}

	r := resolver.NewHierarchyBuildArgsResolver(ds)
	resolution, err := r.Resolve(cmd.Context(), workspaceID)
	if err != nil {
		return fmt.Errorf("failed to resolve build args cascade: %w", err)
	}

	if len(resolution.Args) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No build args set anywhere in the hierarchy",
		})
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(resolution.Args))
	for k := range resolution.Args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// For JSON/YAML, build structured output
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		type argOutput struct {
			Key    string `json:"key" yaml:"key"`
			Value  string `json:"value" yaml:"value"`
			Source string `json:"source" yaml:"source"`
		}
		out := make([]argOutput, 0, len(keys))
		for _, k := range keys {
			source := resolution.Sources[k]
			out = append(out, argOutput{
				Key:    k,
				Value:  resolution.Args[k],
				Source: source.String(),
			})
		}
		return render.OutputWith(getOutputFormat, out, render.Options{})
	}

	// Table/human output
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		source := resolution.Sources[k]
		rows = append(rows, []string{k, resolution.Args[k], source.String()})
	}
	return render.OutputWith(getOutputFormat, render.TableData{
		Headers: []string{"KEY", "VALUE", "SOURCE"},
		Rows:    rows,
	}, render.Options{
		Type:  render.TypeTable,
		Title: fmt.Sprintf("Effective build args for workspace: %s (merged cascade)", getBuildArgsWorkspace),
	})
}

// runGetAllBuildArgs lists build args from ALL hierarchy levels.
func runGetAllBuildArgs(cmd *cobra.Command, ctx resource.Context) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	type scopedArg struct {
		Key   string `json:"key" yaml:"key"`
		Value string `json:"value" yaml:"value"`
		Scope string `json:"scope" yaml:"scope"`
	}

	var allArgs []scopedArg

	// Helper to add args from a map
	addArgs := func(argMap map[string]string, scope string) {
		keys := make([]string, 0, len(argMap))
		for k := range argMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			allArgs = append(allArgs, scopedArg{Key: k, Value: argMap[k], Scope: scope})
		}
	}

	// 1. Global
	globalArgs, err := GetGlobalBuildArgs(ds)
	if err == nil {
		addArgs(globalArgs, "global")
	}

	// 2. All ecosystems
	ecosystems, _ := ds.ListEcosystems()
	for _, eco := range ecosystems {
		ecoYAML := eco.ToYAML(nil)
		addArgs(ecoYAML.Spec.Build.Args, fmt.Sprintf("ecosystem: %s", eco.Name))
	}

	// 3. All domains
	domains, _ := ds.ListAllDomains()
	for _, dom := range domains {
		eco, _ := ds.GetEcosystemByID(dom.EcosystemID)
		ecoName := ""
		if eco != nil {
			ecoName = eco.Name
		}
		domYAML := dom.ToYAML(ecoName, nil)
		addArgs(domYAML.Spec.Build.Args, fmt.Sprintf("domain: %s", dom.Name))
	}

	// 4. All apps
	apps, _ := ds.ListAllApps()
	for _, app := range apps {
		buildConfig := app.GetBuildConfig()
		if buildConfig == nil {
			continue
		}
		addArgs(buildConfig.Args, fmt.Sprintf("app: %s", app.Name))
	}

	// 5. All workspaces
	workspaces, _ := ds.ListAllWorkspaces()
	for _, ws := range workspaces {
		app, _ := ds.GetAppByID(ws.AppID)
		appName := ""
		if app != nil {
			appName = app.Name
		}
		gitRepoName := ""
		if ws.GitRepoID.Valid {
			if gitRepo, err := ds.GetGitRepoByID(ws.GitRepoID.Int64); err == nil && gitRepo != nil {
				gitRepoName = gitRepo.Name
			}
		}
		wsYAML := ws.ToYAML(appName, gitRepoName)
		addArgs(wsYAML.Spec.Build.Args, fmt.Sprintf("workspace: %s/%s", appName, ws.Name))
	}

	if len(allArgs) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No build args found across any scope",
			EmptyHints:   []string{"dvm set build-arg --global <KEY> <VALUE>"},
		})
	}

	// For JSON/YAML, output structured data
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, allArgs, render.Options{})
	}

	// Table output
	rows := make([][]string, len(allArgs))
	for i, a := range allArgs {
		rows[i] = []string{a.Key, a.Value, a.Scope}
	}
	return render.OutputWith(getOutputFormat, render.TableData{
		Headers: []string{"KEY", "VALUE", "SCOPE"},
		Rows:    rows,
	}, render.Options{Type: render.TypeTable})
}

// getBuildArgsAtEcosystem displays build args at the ecosystem level.
func getBuildArgsAtEcosystem(cmd *cobra.Command, ctx resource.Context, ecosystemName string) error {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecoYAML := ecosystemRes.Ecosystem().ToYAML(nil)
	return displayBuildArgs("ecosystem", ecosystemName, ecoYAML.Spec.Build.Args)
}

// getBuildArgsAtDomain displays build args at the domain level.
func getBuildArgsAtDomain(cmd *cobra.Command, ctx resource.Context, domainName string) error {
	res, err := resource.Get(ctx, handlers.KindDomain, domainName)
	if err != nil {
		return fmt.Errorf("domain %q not found: %w", domainName, err)
	}

	domainRes := res.(*handlers.DomainResource)
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}
	domain := domainRes.Domain()
	eco, err := ds.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return fmt.Errorf("failed to get ecosystem for domain: %w", err)
	}
	domainYAML := domain.ToYAML(eco.Name, nil)
	return displayBuildArgs("domain", domainName, domainYAML.Spec.Build.Args)
}

// getBuildArgsAtApp displays build args at the app level.
func getBuildArgsAtApp(cmd *cobra.Command, ctx resource.Context, appName string) error {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	buildConfig := app.GetBuildConfig()
	var argMap map[string]string
	if buildConfig != nil {
		argMap = buildConfig.Args
	}
	return displayBuildArgs("app", appName, argMap)
}

// getBuildArgsAtWorkspace displays build args at the workspace level.
func getBuildArgsAtWorkspace(cmd *cobra.Command, ctx resource.Context, workspaceName, scopeAppName string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	var appID int
	var appName string

	if scopeAppName != "" {
		app, err := ds.GetAppByNameGlobal(scopeAppName)
		if err != nil {
			return fmt.Errorf("app %q not found: %w", scopeAppName, err)
		}
		appID = app.ID
		appName = scopeAppName
	} else {
		activeApp, err := getActiveAppFromContext(ds)
		if err != nil {
			return fmt.Errorf("no app context. Use --app <name> or 'dvm use app <name>' first")
		}
		app, err := ds.GetAppByNameGlobal(activeApp)
		if err != nil {
			return fmt.Errorf("app %q not found: %w", activeApp, err)
		}
		appID = app.ID
		appName = activeApp
	}

	workspace, err := ds.GetWorkspaceByName(appID, workspaceName)
	if err != nil {
		return fmt.Errorf("workspace %q not found under app %q: %w", workspaceName, appName, err)
	}

	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		if gitRepo, err := ds.GetGitRepoByID(workspace.GitRepoID.Int64); err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}

	wsYAML := workspace.ToYAML(appName, gitRepoName)
	return displayBuildArgs("workspace", workspaceName, wsYAML.Spec.Build.Args)
}

// getBuildArgsGlobalLevel displays the global default build args.
func getBuildArgsGlobalLevel(cmd *cobra.Command, ctx resource.Context) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	argMap, err := GetGlobalBuildArgs(ds)
	if err != nil {
		return fmt.Errorf("failed to get global build args: %w", err)
	}

	return displayBuildArgs("global", "global-defaults", argMap)
}

// displayBuildArgs renders the build args map for a given level/object.
// Uses render.OutputWith to support JSON/YAML/table output via the parent -o flag.
func displayBuildArgs(level, objectName string, argMap map[string]string) error {
	if len(argMap) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: fmt.Sprintf("No build args set at %s level (%s)", level, objectName),
		})
	}

	// For JSON/YAML, output the map directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, argMap, render.Options{})
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(argMap))
	for k := range argMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Table output
	rows := make([][]string, len(keys))
	for i, k := range keys {
		rows[i] = []string{k, argMap[k]}
	}
	return render.OutputWith(getOutputFormat, render.TableData{
		Headers: []string{"KEY", "VALUE"},
		Rows:    rows,
	}, render.Options{
		Type:  render.TypeTable,
		Title: fmt.Sprintf("Build args at %s level: %s", level, objectName),
	})
}
