// Package cmd implements the 'dvm set nvim-package' command for hierarchical
// nvim package assignment. Packages cascade down the hierarchy unless
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

// Flags for set nvim-package command
var (
	setNvimPkgEcosystem   string
	setNvimPkgDomain      string
	setNvimPkgApp         string
	setNvimPkgWorkspace   string
	setNvimPkgGlobal      bool
	setNvimPkgOutput      string
	setNvimPkgDryRun      bool
	setNvimPkgShowCascade bool
)

// setNvimPackageCmd sets nvim package at a hierarchy level
var setNvimPackageCmd = &cobra.Command{
	Use:   "nvim-package <name>",
	Short: "Set nvim package at hierarchy level",
	Long: `Set nvim plugin package at ecosystem, domain, app, or workspace level.

Packages cascade down the hierarchy unless overridden:
  global → Ecosystem → Domain → App → Workspace

Use 'none' to clear override and inherit from parent level.

Examples:
  dvm set nvim-package full-stack --workspace dev
  dvm set nvim-package minimal --app my-api
  dvm set nvim-package none --workspace dev    # clear, inherit from app
  dvm set nvim-package standard --domain auth
  dvm set nvim-package full-stack --global     # Set global default
  dvm set nvim-package none --global           # Clear global default`,
	Args: cobra.ExactArgs(1),
	RunE: runSetNvimPackage,
}

func init() {
	setCmd.AddCommand(setNvimPackageCmd)

	setNvimPackageCmd.Flags().StringVar(&setNvimPkgEcosystem, "ecosystem", "", "Set at ecosystem level")
	setNvimPackageCmd.Flags().StringVar(&setNvimPkgDomain, "domain", "", "Set at domain level")
	setNvimPackageCmd.Flags().StringVar(&setNvimPkgApp, "app", "", "Set at app level")
	setNvimPackageCmd.Flags().StringVar(&setNvimPkgWorkspace, "workspace", "", "Set at workspace level")
	setNvimPackageCmd.Flags().BoolVar(&setNvimPkgGlobal, "global", false, "Set as global default")

	setNvimPackageCmd.Flags().StringVarP(&setNvimPkgOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	AddDryRunFlag(setNvimPackageCmd, &setNvimPkgDryRun)
	setNvimPackageCmd.Flags().BoolVar(&setNvimPkgShowCascade, "show-cascade", false, "Show package cascade effect")
}

func runSetNvimPackage(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	if setNvimPkgEcosystem == "" && setNvimPkgDomain == "" && setNvimPkgApp == "" &&
		setNvimPkgWorkspace == "" && !setNvimPkgGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	if setNvimPkgGlobal && (setNvimPkgEcosystem != "" || setNvimPkgDomain != "" ||
		setNvimPkgApp != "" || setNvimPkgWorkspace != "") {
		return fmt.Errorf("--global cannot be used with --ecosystem, --domain, --app, or --workspace")
	}

	// Determine the effective value: "none" clears the field
	clearPkg := pkgName == "none"

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	var result *packageSetResult
	if setNvimPkgWorkspace != "" {
		result, err = setNvimPkgAtWorkspace(cmd, ctx, setNvimPkgWorkspace, setNvimPkgApp, pkgName, clearPkg)
	} else if setNvimPkgApp != "" {
		result, err = setNvimPkgAtApp(cmd, ctx, setNvimPkgApp, pkgName, clearPkg)
	} else if setNvimPkgDomain != "" {
		result, err = setNvimPkgAtDomain(cmd, ctx, setNvimPkgDomain, pkgName, clearPkg)
	} else if setNvimPkgEcosystem != "" {
		result, err = setNvimPkgAtEcosystem(cmd, ctx, setNvimPkgEcosystem, pkgName, clearPkg)
	} else if setNvimPkgGlobal {
		result, err = setNvimPkgAtGlobal(cmd, ctx, pkgName, clearPkg)
	} else {
		return fmt.Errorf("no hierarchy level specified")
	}
	if err != nil {
		return err
	}

	if setNvimPkgDryRun {
		result.ObjectName = result.ObjectName + " (dry-run)"
	}

	return renderPackageSetResult(result, setNvimPkgOutput, setNvimPkgShowCascade, "nvim")
}

// ---------------------------------------------------------------------------
// Level-specific setters for nvim-package
// ---------------------------------------------------------------------------

func setNvimPkgAtEcosystem(cmd *cobra.Command, ctx resource.Context, ecoName, pkgName string, clear bool) (*packageSetResult, error) {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecoName)
	if err != nil {
		return nil, fmt.Errorf("ecosystem %q not found: %w", ecoName, err)
	}
	eco := res.(*handlers.EcosystemResource).Ecosystem()

	prev := nullStringValue(eco.NvimPackage)

	if setNvimPkgDryRun {
		return &packageSetResult{
			Level: "ecosystem", ObjectName: ecoName,
			Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
		}, nil
	}

	if clear {
		eco.NvimPackage = sql.NullString{Valid: false}
	} else {
		eco.NvimPackage = sql.NullString{String: pkgName, Valid: true}
	}

	ecoYAML := eco.ToYAML(nil)
	yamlData, err := yaml.Marshal(ecoYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-nvim-package"); err != nil {
		return nil, fmt.Errorf("failed to update ecosystem: %w", err)
	}

	return &packageSetResult{
		Level: "ecosystem", ObjectName: ecoName,
		Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
	}, nil
}

func setNvimPkgAtDomain(cmd *cobra.Command, ctx resource.Context, domName, pkgName string, clear bool) (*packageSetResult, error) {
	res, err := resource.Get(ctx, handlers.KindDomain, domName)
	if err != nil {
		return nil, fmt.Errorf("domain %q not found: %w", domName, err)
	}
	dom := res.(*handlers.DomainResource).Domain()

	prev := nullStringValue(dom.NvimPackage)

	if setNvimPkgDryRun {
		return &packageSetResult{
			Level: "domain", ObjectName: domName,
			Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
		}, nil
	}

	if clear {
		dom.NvimPackage = sql.NullString{Valid: false}
	} else {
		dom.NvimPackage = sql.NullString{String: pkgName, Valid: true}
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}
	eco, err := ds.GetEcosystemByID(dom.EcosystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ecosystem for domain: %w", err)
	}

	domYAML := dom.ToYAML(eco.Name, nil)
	yamlData, err := yaml.Marshal(domYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal domain YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-nvim-package"); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	return &packageSetResult{
		Level: "domain", ObjectName: domName,
		Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
	}, nil
}

func setNvimPkgAtApp(cmd *cobra.Command, ctx resource.Context, appName, pkgName string, clear bool) (*packageSetResult, error) {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return nil, fmt.Errorf("app %q not found: %w", appName, err)
	}
	app := res.(*handlers.AppResource).App()

	prev := nullStringValue(app.NvimPackage)

	if setNvimPkgDryRun {
		return &packageSetResult{
			Level: "app", ObjectName: appName,
			Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
		}, nil
	}

	if clear {
		app.NvimPackage = sql.NullString{Valid: false}
	} else {
		app.NvimPackage = sql.NullString{String: pkgName, Valid: true}
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}
	dom, err := ds.GetDomainByID(app.DomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain for app: %w", err)
	}

	appYAML := app.ToYAML(dom.Name, nil, "")
	yamlData, err := yaml.Marshal(appYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal app YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, yamlData, "set-nvim-package"); err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return &packageSetResult{
		Level: "app", ObjectName: appName,
		Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
	}, nil
}

func setNvimPkgAtWorkspace(cmd *cobra.Command, ctx resource.Context, wsName, scopeApp, pkgName string, clear bool) (*packageSetResult, error) {
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

	prev := nullStringValue(workspace.NvimPackage)

	if setNvimPkgDryRun {
		return &packageSetResult{
			Level: "workspace", ObjectName: wsName,
			Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
		}, nil
	}

	if clear {
		workspace.NvimPackage = sql.NullString{Valid: false}
	} else {
		workspace.NvimPackage = sql.NullString{String: pkgName, Valid: true}
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
	if _, err := resource.Apply(ctx, yamlData, "set-nvim-package"); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return &packageSetResult{
		Level: "workspace", ObjectName: wsName,
		Package: pkgName, PreviousPackage: prev, PackageType: "nvim",
	}, nil
}

func setNvimPkgAtGlobal(_ *cobra.Command, ctx resource.Context, pkgName string, clear bool) (*packageSetResult, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DataStore: %w", err)
	}

	prevPkg, _ := ds.GetDefault("nvim-package")

	if setNvimPkgDryRun {
		return &packageSetResult{
			Level: "global", ObjectName: "global-defaults",
			Package: pkgName, PreviousPackage: prevPkg, PackageType: "nvim",
		}, nil
	}

	if clear {
		if err := ds.DeleteDefault("nvim-package"); err != nil {
			return nil, fmt.Errorf("failed to clear global default nvim package: %w", err)
		}
	} else {
		if err := ds.SetDefault("nvim-package", pkgName); err != nil {
			return nil, fmt.Errorf("failed to set global default nvim package: %w", err)
		}
	}

	return &packageSetResult{
		Level: "global", ObjectName: "global-defaults",
		Package: pkgName, PreviousPackage: prevPkg, PackageType: "nvim",
	}, nil
}
