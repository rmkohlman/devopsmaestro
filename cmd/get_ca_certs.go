// Package cmd provides the 'dvm get ca-certs' command for viewing hierarchical CA certificates.
// Without --effective: shows certs at the specified level only.
// With --effective:    shows the fully merged cascade result (requires --workspace).
package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	cacertsresolver "devopsmaestro/pkg/cacerts/resolver"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

// Flags for get ca-certs command
var (
	getCACertsEcosystem string
	getCACertsDomain    string
	getCACertsApp       string
	getCACertsWorkspace string
	getCACertsGlobal    bool
	getCACertsEffective bool
	getCACertsOutputFmt string
)

// getCACertsCmd displays CA certs at a specific hierarchy level or the merged result
var getCACertsCmd = &cobra.Command{
	Use:   "ca-certs",
	Short: "Get CA certificates at hierarchy level",
	Long: `Display CA certificates at a specific hierarchy level, or show the merged cascade.

Without --effective: shows certs set directly at the specified level only.
With --effective:    shows the fully merged cascade result from all levels
                     (requires --workspace to identify the hierarchy path).

The SOURCE column (in --effective mode) shows which level each cert came from.

CA certs cascade: global → ecosystem → domain → app → workspace (workspace wins, matched by name)

Examples:
  dvm get ca-certs --global                        # Global defaults
  dvm get ca-certs --ecosystem my-eco              # Ecosystem-level certs
  dvm get ca-certs --domain data-sci               # Domain-level certs
  dvm get ca-certs --app ml-api                    # App-level certs
  dvm get ca-certs --workspace dev                 # Workspace-level certs
  dvm get ca-certs --workspace dev --effective     # Fully merged cascade`,
	RunE: runGetCACerts,
}

func init() {
	getCmd.AddCommand(getCACertsCmd)

	getCACertsCmd.Flags().StringVar(&getCACertsEcosystem, "ecosystem", "", "Get CA certs at ecosystem level")
	getCACertsCmd.Flags().StringVar(&getCACertsDomain, "domain", "", "Get CA certs at domain level")
	getCACertsCmd.Flags().StringVar(&getCACertsApp, "app", "", "Get CA certs at app level")
	getCACertsCmd.Flags().StringVar(&getCACertsWorkspace, "workspace", "", "Get CA certs at workspace level")
	getCACertsCmd.Flags().BoolVar(&getCACertsGlobal, "global", false, "Get global default CA certs")
	getCACertsCmd.Flags().BoolVar(&getCACertsEffective, "effective", false, "Show fully merged cascade result (requires --workspace)")
	getCACertsCmd.Flags().StringVarP(&getCACertsOutputFmt, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
}

func runGetCACerts(cmd *cobra.Command, args []string) error {
	// Validate that at least one target flag is provided
	if getCACertsEcosystem == "" && getCACertsDomain == "" && getCACertsApp == "" &&
		getCACertsWorkspace == "" && !getCACertsGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	// --effective requires --workspace (to identify the full hierarchy path)
	if getCACertsEffective && getCACertsWorkspace == "" {
		return fmt.Errorf("--effective requires --workspace to identify the hierarchy path")
	}

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// --effective mode: show merged cascade
	if getCACertsEffective {
		return runGetCACertsEffective(cmd, ctx)
	}

	// Level-specific display
	switch {
	case getCACertsWorkspace != "":
		return getCACertsAtWorkspace(ctx, getCACertsWorkspace, getCACertsApp)
	case getCACertsApp != "":
		return getCACertsAtApp(ctx, getCACertsApp)
	case getCACertsDomain != "":
		return getCACertsAtDomain(ctx, getCACertsDomain)
	case getCACertsEcosystem != "":
		return getCACertsAtEcosystem(ctx, getCACertsEcosystem)
	case getCACertsGlobal:
		return getCACertsGlobalLevel(ctx)
	default:
		return fmt.Errorf("no hierarchy level specified")
	}
}

// runGetCACertsEffective resolves the full cascade and displays with SOURCE provenance.
func runGetCACertsEffective(cmd *cobra.Command, ctx resource.Context) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	// Resolve workspace to get its ID
	var workspaceID int
	if getCACertsApp != "" {
		app, err := ds.GetAppByNameGlobal(getCACertsApp)
		if err != nil {
			return fmt.Errorf("app %q not found: %w", getCACertsApp, err)
		}
		ws, err := ds.GetWorkspaceByName(app.ID, getCACertsWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q not found under app %q: %w", getCACertsWorkspace, getCACertsApp, err)
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
		ws, err := ds.GetWorkspaceByName(app.ID, getCACertsWorkspace)
		if err != nil {
			return fmt.Errorf("workspace %q not found under app %q: %w", getCACertsWorkspace, activeApp, err)
		}
		workspaceID = ws.ID
	}

	r := cacertsresolver.NewHierarchyCACertsResolver(ds)
	resolution, err := r.Resolve(cmd.Context(), workspaceID)
	if err != nil {
		return fmt.Errorf("failed to resolve CA certs cascade: %w", err)
	}

	if len(resolution.Certs) == 0 {
		render.Info("No CA certs set anywhere in the hierarchy")
		return nil
	}

	render.Info(fmt.Sprintf("Effective CA certs for workspace: %s (merged cascade)", getCACertsWorkspace))
	render.Blank()

	// Display as table: NAME  VAULT-SECRET  (SOURCE)
	for _, cert := range resolution.Certs {
		source := resolution.Sources[cert.Name]
		vaultInfo := cert.VaultSecret
		if cert.VaultEnvironment != "" {
			vaultInfo += fmt.Sprintf(" (env=%s)", cert.VaultEnvironment)
		}
		if cert.VaultField != "" {
			vaultInfo += fmt.Sprintf(" (field=%s)", cert.VaultField)
		}
		render.Plainf("  %-30s  vault-secret=%-25s  (%s)", cert.Name, vaultInfo, source.String())
	}
	render.Blank()
	render.Info("Legend: global < ecosystem < domain < app < workspace (workspace wins, matched by name)")

	return nil
}

// getCACertsAtEcosystem displays CA certs at the ecosystem level.
func getCACertsAtEcosystem(ctx resource.Context, ecosystemName string) error {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecoYAML := ecosystemRes.Ecosystem().ToYAML(nil)
	return displayCACerts("ecosystem", ecosystemName, ecoYAML.Spec.CACerts)
}

// getCACertsAtDomain displays CA certs at the domain level.
func getCACertsAtDomain(ctx resource.Context, domainName string) error {
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
	return displayCACerts("domain", domainName, domainYAML.Spec.CACerts)
}

// getCACertsAtApp displays CA certs at the app level.
func getCACertsAtApp(ctx resource.Context, appName string) error {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	buildConfig := app.GetBuildConfig()
	var certs []models.CACertConfig
	if buildConfig != nil {
		certs = buildConfig.CACerts
	}
	return displayCACerts("app", appName, certs)
}

// getCACertsAtWorkspace displays CA certs at the workspace level.
func getCACertsAtWorkspace(ctx resource.Context, workspaceName, scopeAppName string) error {
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
	return displayCACerts("workspace", workspaceName, wsYAML.Spec.Build.CACerts)
}

// getCACertsGlobalLevel displays the global default CA certs.
func getCACertsGlobalLevel(ctx resource.Context) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	certs, err := GetGlobalCACerts(ds)
	if err != nil {
		return fmt.Errorf("failed to get global CA certs: %w", err)
	}

	return displayCACerts("global", "global-defaults", certs)
}

// displayCACerts renders a CA cert list for a given level/object.
func displayCACerts(level, objectName string, certs []models.CACertConfig) error {
	if len(certs) == 0 {
		render.Info(fmt.Sprintf("No CA certs set at %s level (%s)", level, objectName))
		return nil
	}

	render.Info(fmt.Sprintf("CA certs at %s level: %s", level, objectName))
	render.Blank()

	for _, c := range certs {
		vaultInfo := c.VaultSecret
		if c.VaultEnvironment != "" {
			vaultInfo += fmt.Sprintf(" (env=%s)", c.VaultEnvironment)
		}
		if c.VaultField != "" {
			vaultInfo += fmt.Sprintf(" (field=%s)", c.VaultField)
		}
		render.Plainf("  %-30s  vault-secret=%s", c.Name, vaultInfo)
	}
	render.Blank()

	return nil
}
