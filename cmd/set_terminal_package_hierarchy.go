// Package cmd implements the 'dvm set terminal-package' command for hierarchical
// terminal package assignment. Packages cascade down the hierarchy unless
// overridden: global → ecosystem → domain → app → workspace.
package cmd

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for set terminal-package command
var (
	setTermPkgEcosystem   string
	setTermPkgDomain      string
	setTermPkgApp         string
	setTermPkgWorkspace   string
	setTermPkgGlobal      bool
	setTermPkgOutput      string
	setTermPkgDryRun      bool
	setTermPkgShowCascade bool
)

// setTerminalPackageCmd sets terminal package at a hierarchy level.
// This is the kebab-case "dvm set terminal-package" command, registered
// directly on setCmd. The workspace-only "dvm set terminal package" lives in
// terminal_set.go as setTerminalPackageWorkspaceCmd.
var setTerminalPackageCmd = &cobra.Command{
	Use:   "terminal-package <name>",
	Short: "Set terminal package at hierarchy level",
	Long: `Set terminal package at ecosystem, domain, app, or workspace level.

Packages cascade down the hierarchy unless overridden:
  global → Ecosystem → Domain → App → Workspace

Use 'none' to clear override and inherit from parent level.

Examples:
  dvm set terminal-package poweruser --workspace dev
  dvm set terminal-package minimal --app my-api
  dvm set terminal-package none --workspace dev    # clear, inherit from app
  dvm set terminal-package standard --domain auth
  dvm set terminal-package poweruser --global      # Set global default
  dvm set terminal-package none --global           # Clear global default`,
	Args: cobra.ExactArgs(1),
	RunE: runSetTerminalPkg,
}

func init() {
	setCmd.AddCommand(setTerminalPackageCmd)

	setTerminalPackageCmd.Flags().StringVar(&setTermPkgEcosystem, "ecosystem", "", "Set at ecosystem level")
	setTerminalPackageCmd.Flags().StringVar(&setTermPkgDomain, "domain", "", "Set at domain level")
	setTerminalPackageCmd.Flags().StringVar(&setTermPkgApp, "app", "", "Set at app level")
	setTerminalPackageCmd.Flags().StringVar(&setTermPkgWorkspace, "workspace", "", "Set at workspace level")
	setTerminalPackageCmd.Flags().BoolVar(&setTermPkgGlobal, "global", false, "Set as global default")

	setTerminalPackageCmd.Flags().StringVarP(&setTermPkgOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	AddDryRunFlag(setTerminalPackageCmd, &setTermPkgDryRun)
	setTerminalPackageCmd.Flags().BoolVar(&setTermPkgShowCascade, "show-cascade", false, "Show package cascade effect")
}

func runSetTerminalPkg(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	if setTermPkgEcosystem == "" && setTermPkgDomain == "" && setTermPkgApp == "" &&
		setTermPkgWorkspace == "" && !setTermPkgGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	if setTermPkgGlobal && (setTermPkgEcosystem != "" || setTermPkgDomain != "" ||
		setTermPkgApp != "" || setTermPkgWorkspace != "") {
		return fmt.Errorf("--global cannot be used with --ecosystem, --domain, --app, or --workspace")
	}

	clearPkg := pkgName == "none"

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	var result *packageSetResult
	if setTermPkgWorkspace != "" {
		result, err = setTermPkgAtWorkspace(cmd, ctx, setTermPkgWorkspace, setTermPkgApp, pkgName, clearPkg)
	} else if setTermPkgApp != "" {
		result, err = setTermPkgAtApp(cmd, ctx, setTermPkgApp, pkgName, clearPkg)
	} else if setTermPkgDomain != "" {
		result, err = setTermPkgAtDomain(cmd, ctx, setTermPkgDomain, pkgName, clearPkg)
	} else if setTermPkgEcosystem != "" {
		result, err = setTermPkgAtEcosystem(cmd, ctx, setTermPkgEcosystem, pkgName, clearPkg)
	} else if setTermPkgGlobal {
		result, err = setTermPkgAtGlobal(cmd, ctx, pkgName, clearPkg)
	} else {
		return fmt.Errorf("no hierarchy level specified")
	}
	if err != nil {
		return err
	}

	if setTermPkgDryRun {
		result.ObjectName = result.ObjectName + " (dry-run)"
	}

	return renderPackageSetResult(result, setTermPkgOutput, setTermPkgShowCascade, "terminal")
}

// ---------------------------------------------------------------------------
// Level-specific setters for terminal-package
// ---------------------------------------------------------------------------

func setTermPkgAtEcosystem(cmd *cobra.Command, ctx resource.Context, ecoName, pkgName string, clear bool) (*packageSetResult, error) {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecoName)
	if err != nil {
		return nil, fmt.Errorf("ecosystem %q not found: %w", ecoName, err)
	}
	eco := res.(*handlers.EcosystemResource).Ecosystem()

	prev := nullStringValue(eco.TerminalPackage)

	if setTermPkgDryRun {
		return &packageSetResult{
			Level: "ecosystem", ObjectName: ecoName,
			Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
		}, nil
	}

	if clear {
		eco.TerminalPackage = sql.NullString{Valid: false}
	} else {
		eco.TerminalPackage = sql.NullString{String: pkgName, Valid: true}
	}

	ecoYAML := eco.ToYAML(nil)
	yamlData, err := yaml.Marshal(ecoYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-terminal-package"); err != nil {
		return nil, fmt.Errorf("failed to update ecosystem: %w", err)
	}

	return &packageSetResult{
		Level: "ecosystem", ObjectName: ecoName,
		Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
	}, nil
}

func setTermPkgAtDomain(cmd *cobra.Command, ctx resource.Context, domName, pkgName string, clear bool) (*packageSetResult, error) {
	res, err := resource.Get(ctx, handlers.KindDomain, domName)
	if err != nil {
		return nil, fmt.Errorf("domain %q not found: %w", domName, err)
	}
	dom := res.(*handlers.DomainResource).Domain()

	prev := nullStringValue(dom.TerminalPackage)

	if setTermPkgDryRun {
		return &packageSetResult{
			Level: "domain", ObjectName: domName,
			Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
		}, nil
	}

	if clear {
		dom.TerminalPackage = sql.NullString{Valid: false}
	} else {
		dom.TerminalPackage = sql.NullString{String: pkgName, Valid: true}
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}
	ecoName := ""
	if dom.EcosystemID.Valid {
		eco, err := ds.GetEcosystemByID(int(dom.EcosystemID.Int64))
		if err == nil {
			ecoName = eco.Name
		}
	}

	domYAML := dom.ToYAML(ecoName, nil)
	yamlData, err := yaml.Marshal(domYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal domain YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-terminal-package"); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	return &packageSetResult{
		Level: "domain", ObjectName: domName,
		Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
	}, nil
}

func setTermPkgAtApp(cmd *cobra.Command, ctx resource.Context, appName, pkgName string, clear bool) (*packageSetResult, error) {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return nil, fmt.Errorf("app %q not found: %w", appName, err)
	}
	app := res.(*handlers.AppResource).App()

	prev := nullStringValue(app.TerminalPackage)

	if setTermPkgDryRun {
		return &packageSetResult{
			Level: "app", ObjectName: appName,
			Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
		}, nil
	}

	if clear {
		app.TerminalPackage = sql.NullString{Valid: false}
	} else {
		app.TerminalPackage = sql.NullString{String: pkgName, Valid: true}
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}
	domName := ""
	if app.DomainID.Valid {
		dom, err := ds.GetDomainByID(int(app.DomainID.Int64))
		if err == nil {
			domName = dom.Name
		}
	}

	appYAML := app.ToYAML(domName, nil, "", "")
	yamlData, err := yaml.Marshal(appYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal app YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-terminal-package"); err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return &packageSetResult{
		Level: "app", ObjectName: appName,
		Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
	}, nil
}

func setTermPkgAtWorkspace(cmd *cobra.Command, ctx resource.Context, wsName, scopeApp, pkgName string, clear bool) (*packageSetResult, error) {
	sqlDS, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}

	var workspace *models.Workspace
	var appName string

	if scopeApp != "" {
		app, err := sqlDS.GetAppByNameGlobal(scopeApp)
		if err != nil {
			return nil, fmt.Errorf("app %q not found: %w", scopeApp, err)
		}
		workspace, err = sqlDS.GetWorkspaceByName(app.ID, wsName)
		if err != nil {
			return nil, fmt.Errorf("workspace %q not found under app %q: %w", wsName, scopeApp, err)
		}
		appName = scopeApp
	} else {
		res, err := resource.Get(ctx, handlers.KindWorkspace, wsName)
		if err != nil {
			return nil, fmt.Errorf("workspace %q not found: %w", wsName, err)
		}
		wsRes := res.(*handlers.WorkspaceResource)
		workspace = wsRes.Workspace()
		appName = wsRes.AppName()
	}

	prev := nullStringValue(workspace.TerminalPackage)

	if setTermPkgDryRun {
		return &packageSetResult{
			Level: "workspace", ObjectName: wsName,
			Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
		}, nil
	}

	if clear {
		workspace.TerminalPackage = sql.NullString{Valid: false}
	} else {
		workspace.TerminalPackage = sql.NullString{String: pkgName, Valid: true}
	}

	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		if repo, err := sqlDS.GetGitRepoByID(workspace.GitRepoID.Int64); err == nil && repo != nil {
			gitRepoName = repo.Name
		}
	}
	wsYAML := workspace.ToYAML(appName, gitRepoName)
	yamlData, err := yaml.Marshal(wsYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workspace YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-terminal-package"); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return &packageSetResult{
		Level: "workspace", ObjectName: wsName,
		Package: pkgName, PreviousPackage: prev, PackageType: "terminal",
	}, nil
}

func setTermPkgAtGlobal(_ *cobra.Command, ctx resource.Context, pkgName string, clear bool) (*packageSetResult, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}

	prevPkg, _ := ds.GetDefault("terminal-package")

	if setTermPkgDryRun {
		return &packageSetResult{
			Level: "global", ObjectName: "global-defaults",
			Package: pkgName, PreviousPackage: prevPkg, PackageType: "terminal",
		}, nil
	}

	if clear {
		if err := ds.DeleteDefault("terminal-package"); err != nil {
			return nil, fmt.Errorf("failed to clear global default terminal package: %w", err)
		}
	} else {
		if err := ds.SetDefault("terminal-package", pkgName); err != nil {
			return nil, fmt.Errorf("failed to set global default terminal package: %w", err)
		}
	}

	return &packageSetResult{
		Level: "global", ObjectName: "global-defaults",
		Package: pkgName, PreviousPackage: prevPkg, PackageType: "terminal",
	}, nil
}
