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
	getBuildArgsOutputFmt string
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
  dvm get build-args --workspace dev --effective     # Fully merged cascade`,
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
	getBuildArgsCmd.Flags().StringVarP(&getBuildArgsOutputFmt, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
}

func runGetBuildArgs(cmd *cobra.Command, args []string) error {
	// Validate that at least one target flag is provided
	if getBuildArgsEcosystem == "" && getBuildArgsDomain == "" && getBuildArgsApp == "" &&
		getBuildArgsWorkspace == "" && !getBuildArgsGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
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
		return getBuildArgsAtWorkspace(ctx, getBuildArgsWorkspace, getBuildArgsApp)
	case getBuildArgsApp != "":
		return getBuildArgsAtApp(ctx, getBuildArgsApp)
	case getBuildArgsDomain != "":
		return getBuildArgsAtDomain(ctx, getBuildArgsDomain)
	case getBuildArgsEcosystem != "":
		return getBuildArgsAtEcosystem(ctx, getBuildArgsEcosystem)
	case getBuildArgsGlobal:
		return getBuildArgsGlobalLevel(ctx)
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
		render.Info("No build args set anywhere in the hierarchy")
		return nil
	}

	render.Info(fmt.Sprintf("Effective build args for workspace: %s (merged cascade)", getBuildArgsWorkspace))
	render.Blank()

	// Sort keys for deterministic output
	keys := make([]string, 0, len(resolution.Args))
	for k := range resolution.Args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Display as table: KEY=VALUE  (SOURCE)
	for _, k := range keys {
		source := resolution.Sources[k]
		render.Plainf("  %-40s = %-30s  (%s)", k, resolution.Args[k], source.String())
	}
	render.Blank()
	render.Info("Legend: global < ecosystem < domain < app < workspace (workspace wins)")

	return nil
}

// getBuildArgsAtEcosystem displays build args at the ecosystem level.
func getBuildArgsAtEcosystem(ctx resource.Context, ecosystemName string) error {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecoYAML := ecosystemRes.Ecosystem().ToYAML(nil)
	return displayBuildArgs("ecosystem", ecosystemName, ecoYAML.Spec.Build.Args)
}

// getBuildArgsAtDomain displays build args at the domain level.
func getBuildArgsAtDomain(ctx resource.Context, domainName string) error {
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
func getBuildArgsAtApp(ctx resource.Context, appName string) error {
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
func getBuildArgsAtWorkspace(ctx resource.Context, workspaceName, scopeAppName string) error {
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
func getBuildArgsGlobalLevel(ctx resource.Context) error {
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
func displayBuildArgs(level, objectName string, argMap map[string]string) error {
	if len(argMap) == 0 {
		render.Info(fmt.Sprintf("No build args set at %s level (%s)", level, objectName))
		return nil
	}

	render.Info(fmt.Sprintf("Build args at %s level: %s", level, objectName))
	render.Blank()

	keys := make([]string, 0, len(argMap))
	for k := range argMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		render.Plainf("  %s = %s", k, argMap[k])
	}
	render.Blank()

	return nil
}
