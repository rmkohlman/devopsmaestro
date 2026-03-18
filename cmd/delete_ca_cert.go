// Package cmd provides the 'dvm delete ca-cert' command for removing CA certificates.
// Supports deletion at every hierarchy level: global, ecosystem, domain, app, workspace.
// Prompts for confirmation by default; use --force to skip.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for delete ca-cert command
var (
	deleteCACertEcosystem string
	deleteCACertDomain    string
	deleteCACertApp       string
	deleteCACertWorkspace string
	deleteCACertGlobal    bool
	deleteCACertForce     bool
)

// deleteCACertCmd removes a CA certificate at a specific hierarchy level
var deleteCACertCmd = &cobra.Command{
	Use:   "ca-cert NAME",
	Short: "Delete a CA certificate at hierarchy level",
	Long: `Delete a CA certificate by name at ecosystem, domain, app, workspace, or global level.

If the cert name does not exist at the specified level, the operation is a no-op (no error).
By default, you will be prompted for confirmation. Use --force to skip.

Examples:
  dvm delete ca-cert corp-root --ecosystem my-eco
  dvm delete ca-cert proxy-ca --domain data-sci
  dvm delete ca-cert internal --app ml-api
  dvm delete ca-cert dev-ca --workspace dev
  dvm delete ca-cert corp-root --global
  dvm delete ca-cert corp-root --global --force   # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteCACert,
}

func init() {
	deleteCmd.AddCommand(deleteCACertCmd)

	deleteCACertCmd.Flags().StringVar(&deleteCACertEcosystem, "ecosystem", "", "Delete at ecosystem level")
	deleteCACertCmd.Flags().StringVar(&deleteCACertDomain, "domain", "", "Delete at domain level")
	deleteCACertCmd.Flags().StringVar(&deleteCACertApp, "app", "", "Delete at app level")
	deleteCACertCmd.Flags().StringVar(&deleteCACertWorkspace, "workspace", "", "Delete at workspace level")
	deleteCACertCmd.Flags().BoolVar(&deleteCACertGlobal, "global", false, "Delete from DVM-wide defaults")
	deleteCACertCmd.Flags().BoolVarP(&deleteCACertForce, "force", "f", false, "Skip confirmation prompt")
}

func runDeleteCACert(cmd *cobra.Command, args []string) error {
	certName := args[0]

	// Validate that at least one target flag is provided
	if deleteCACertEcosystem == "" && deleteCACertDomain == "" && deleteCACertApp == "" &&
		deleteCACertWorkspace == "" && !deleteCACertGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	// Validate that --global is exclusive with all other level flags
	if deleteCACertGlobal && (deleteCACertEcosystem != "" || deleteCACertDomain != "" ||
		deleteCACertApp != "" || deleteCACertWorkspace != "") {
		return fmt.Errorf("--global cannot be used with --ecosystem, --domain, --app, or --workspace")
	}

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Confirm deletion
	if !deleteCACertForce {
		levelDesc := resolveCACertLevelDesc()
		fmt.Printf("Delete CA cert %q at %s? (y/N): ", certName, levelDesc)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if response != "y" && response != "Y" {
			render.Info("Aborted")
			return nil
		}
	}

	switch {
	case deleteCACertWorkspace != "":
		return deleteCACertAtWorkspace(ctx, deleteCACertWorkspace, deleteCACertApp, certName)
	case deleteCACertApp != "":
		return deleteCACertAtApp(ctx, deleteCACertApp, certName)
	case deleteCACertDomain != "":
		return deleteCACertAtDomain(ctx, deleteCACertDomain, certName)
	case deleteCACertEcosystem != "":
		return deleteCACertAtEcosystem(ctx, deleteCACertEcosystem, certName)
	case deleteCACertGlobal:
		return deleteCACertGlobalLevel(ctx, certName)
	default:
		return fmt.Errorf("no hierarchy level specified")
	}
}

// resolveCACertLevelDesc returns a human-readable description of the target level.
func resolveCACertLevelDesc() string {
	switch {
	case deleteCACertWorkspace != "":
		return fmt.Sprintf("workspace %q", deleteCACertWorkspace)
	case deleteCACertApp != "":
		return fmt.Sprintf("app %q", deleteCACertApp)
	case deleteCACertDomain != "":
		return fmt.Sprintf("domain %q", deleteCACertDomain)
	case deleteCACertEcosystem != "":
		return fmt.Sprintf("ecosystem %q", deleteCACertEcosystem)
	case deleteCACertGlobal:
		return "global defaults"
	default:
		return "unknown level"
	}
}

// deleteCACertAtEcosystem removes a CA cert by name from the ecosystem level.
func deleteCACertAtEcosystem(ctx resource.Context, ecosystemName, certName string) error {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecosystem := ecosystemRes.Ecosystem()

	// No-op if CA certs are not set
	if !ecosystem.CACerts.Valid || ecosystem.CACerts.String == "" {
		render.Info(fmt.Sprintf("CA cert %q not set at ecosystem level (%s) — nothing to delete", certName, ecosystemName))
		return nil
	}

	updatedJSON, err := deleteCACertFromDirectJSON(ecosystem.CACerts.String, certName)
	if err != nil {
		return fmt.Errorf("failed to update ecosystem CA certs: %w", err)
	}

	ecoYAML := ecosystem.ToYAML(nil)
	ecoYAML.Spec.CACerts = parseCACertsFromDirectJSON(updatedJSON)

	data, err := yaml.Marshal(ecoYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-ca-cert"); err != nil {
		return fmt.Errorf("failed to update ecosystem: %w", err)
	}

	render.Success(fmt.Sprintf("CA cert %q deleted from ecosystem %q", certName, ecosystemName))
	return nil
}

// deleteCACertAtDomain removes a CA cert by name from the domain level.
func deleteCACertAtDomain(ctx resource.Context, domainName, certName string) error {
	res, err := resource.Get(ctx, handlers.KindDomain, domainName)
	if err != nil {
		return fmt.Errorf("domain %q not found: %w", domainName, err)
	}

	domainRes := res.(*handlers.DomainResource)
	domain := domainRes.Domain()

	if !domain.CACerts.Valid || domain.CACerts.String == "" {
		render.Info(fmt.Sprintf("CA cert %q not set at domain level (%s) — nothing to delete", certName, domainName))
		return nil
	}

	updatedJSON, err := deleteCACertFromDirectJSON(domain.CACerts.String, certName)
	if err != nil {
		return fmt.Errorf("failed to update domain CA certs: %w", err)
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}
	eco, err := ds.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return fmt.Errorf("failed to get ecosystem for domain: %w", err)
	}

	domainYAML := domain.ToYAML(eco.Name, nil)
	domainYAML.Spec.CACerts = parseCACertsFromDirectJSON(updatedJSON)

	data, err := yaml.Marshal(domainYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal domain YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-ca-cert"); err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	render.Success(fmt.Sprintf("CA cert %q deleted from domain %q", certName, domainName))
	return nil
}

// deleteCACertAtApp removes a CA cert by name from the app level.
func deleteCACertAtApp(ctx resource.Context, appName, certName string) error {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	if !app.BuildConfig.Valid || app.BuildConfig.String == "" {
		render.Info(fmt.Sprintf("CA cert %q not set at app level (%s) — nothing to delete", certName, appName))
		return nil
	}

	// Check if the cert actually exists in the build config
	existingCerts := parseCACertsFromWrappedJSON(app.BuildConfig.String)
	certFound := false
	for _, c := range existingCerts {
		if c.Name == certName {
			certFound = true
			break
		}
	}
	if !certFound {
		render.Info(fmt.Sprintf("CA cert %q not set at app level (%s) — nothing to delete", certName, appName))
		return nil
	}

	updatedJSON, err := deleteCACertFromWrappedJSON(app.BuildConfig.String, certName)
	if err != nil {
		return fmt.Errorf("failed to update app build config: %w", err)
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}
	domain, err := ds.GetDomainByID(app.DomainID)
	if err != nil {
		return fmt.Errorf("failed to get domain for app: %w", err)
	}

	appYAML := app.ToYAML(domain.Name, nil)
	appYAML.Spec.Build.CACerts = parseCACertsFromWrappedJSON(updatedJSON)

	data, err := yaml.Marshal(appYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal app YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-ca-cert"); err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	render.Success(fmt.Sprintf("CA cert %q deleted from app %q", certName, appName))
	return nil
}

// deleteCACertAtWorkspace removes a CA cert by name from the workspace level.
func deleteCACertAtWorkspace(ctx resource.Context, workspaceName, scopeAppName, certName string) error {
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

	if !workspace.BuildConfig.Valid || workspace.BuildConfig.String == "" {
		render.Info(fmt.Sprintf("CA cert %q not set at workspace level (%s) — nothing to delete", certName, workspaceName))
		return nil
	}

	// Check if the cert actually exists
	existingCerts := parseCACertsFromWrappedJSON(workspace.BuildConfig.String)
	certFound := false
	for _, c := range existingCerts {
		if c.Name == certName {
			certFound = true
			break
		}
	}
	if !certFound {
		render.Info(fmt.Sprintf("CA cert %q not set at workspace level (%s) — nothing to delete", certName, workspaceName))
		return nil
	}

	updatedJSON, err := deleteCACertFromWrappedJSON(workspace.BuildConfig.String, certName)
	if err != nil {
		return fmt.Errorf("failed to update workspace build config: %w", err)
	}

	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		if gitRepo, err := ds.GetGitRepoByID(workspace.GitRepoID.Int64); err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}

	wsYAML := workspace.ToYAML(appName, gitRepoName)
	wsYAML.Spec.Build.CACerts = parseCACertsFromWrappedJSON(updatedJSON)

	data, err := yaml.Marshal(wsYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-ca-cert"); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("CA cert %q deleted from workspace %q", certName, workspaceName))
	return nil
}

// deleteCACertGlobalLevel removes a CA cert by name from the global defaults.
func deleteCACertGlobalLevel(ctx resource.Context, certName string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	if err := DeleteGlobalCACert(ds, certName); err != nil {
		return fmt.Errorf("failed to delete global CA cert: %w", err)
	}

	render.Success(fmt.Sprintf("CA cert %q deleted from global defaults", certName))
	return nil
}
