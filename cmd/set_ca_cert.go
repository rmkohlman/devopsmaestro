// Package cmd provides the 'dvm set ca-cert' command for setting hierarchical CA certificates.
// CA certs cascade down the hierarchy: global < ecosystem < domain < app < workspace.
// More-specific levels override less-specific levels (matched by cert Name).
//
// Each CA cert references a MaestroVault secret that contains the PEM certificate data.
package cmd

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for set ca-cert command
var (
	setCACertEcosystem   string
	setCACertDomain      string
	setCACertApp         string
	setCACertWorkspace   string
	setCACertGlobal      bool
	setCACertDryRun      bool
	setCACertVaultSecret string
	setCACertVaultEnv    string
	setCACertVaultField  string
)

// setCACertCmd sets a CA certificate at a specific hierarchy level
var setCACertCmd = &cobra.Command{
	Use:   "ca-cert NAME",
	Short: "Set a CA certificate at hierarchy level",
	Long: `Set a CA certificate at ecosystem, domain, app, workspace, or global level.

CA certs cascade down the hierarchy (more-specific levels override by cert name):
  global → ecosystem → domain → app → workspace

Each cert references a MaestroVault secret containing the PEM certificate data.
Use --vault-secret to specify which secret to fetch.

Examples:
  dvm set ca-cert corp-root --vault-secret corp-root-ca --ecosystem my-eco
  dvm set ca-cert proxy-ca --vault-secret proxy-cert --vault-env production --domain data-sci
  dvm set ca-cert internal --vault-secret internal-ca --vault-field cert --app ml-api
  dvm set ca-cert dev-ca --vault-secret dev-cert --workspace dev
  dvm set ca-cert corp-root --vault-secret corp-root-ca --global

Flags:
  --vault-secret   MaestroVault secret name (required)
  --vault-env      Optional vault environment override
  --vault-field    Optional field within the secret
  --ecosystem      Set at ecosystem level
  --domain         Set at domain level
  --app            Set at app level
  --workspace      Set at workspace level
  --global         Set as DVM-wide default (applies to all workspaces)
  --dry-run        Preview changes without applying`,
	Args: cobra.ExactArgs(1),
	RunE: runSetCACert,
}

func init() {
	setCmd.AddCommand(setCACertCmd)

	setCACertCmd.Flags().StringVar(&setCACertVaultSecret, "vault-secret", "", "MaestroVault secret name (required)")
	setCACertCmd.Flags().StringVar(&setCACertVaultEnv, "vault-env", "", "Optional vault environment override")
	setCACertCmd.Flags().StringVar(&setCACertVaultField, "vault-field", "", "Optional field within the secret")
	setCACertCmd.Flags().StringVar(&setCACertEcosystem, "ecosystem", "", "Set at ecosystem level")
	setCACertCmd.Flags().StringVar(&setCACertDomain, "domain", "", "Set at domain level")
	setCACertCmd.Flags().StringVar(&setCACertApp, "app", "", "Set at app level")
	setCACertCmd.Flags().StringVar(&setCACertWorkspace, "workspace", "", "Set at workspace level")
	setCACertCmd.Flags().BoolVar(&setCACertGlobal, "global", false, "Set as DVM-wide default")
	setCACertCmd.Flags().BoolVar(&setCACertDryRun, "dry-run", false, "Preview changes without applying")

	_ = setCACertCmd.MarkFlagRequired("vault-secret")
}

func runSetCACert(cmd *cobra.Command, args []string) error {
	certName := args[0]

	// Validate that at least one target flag is provided
	if setCACertEcosystem == "" && setCACertDomain == "" && setCACertApp == "" &&
		setCACertWorkspace == "" && !setCACertGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	// Validate that --global is exclusive with all other level flags
	if setCACertGlobal && (setCACertEcosystem != "" || setCACertDomain != "" ||
		setCACertApp != "" || setCACertWorkspace != "") {
		return fmt.Errorf("--global cannot be used with --ecosystem, --domain, --app, or --workspace")
	}

	// Build the cert config from flags
	cert := models.CACertConfig{
		Name:             certName,
		VaultSecret:      setCACertVaultSecret,
		VaultEnvironment: setCACertVaultEnv,
		VaultField:       setCACertVaultField,
	}

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Dispatch to level-specific handler
	var levelName, objectName string
	switch {
	case setCACertWorkspace != "":
		levelName, objectName, err = setCACertAtWorkspace(cmd, ctx, setCACertWorkspace, setCACertApp, cert)
	case setCACertApp != "":
		levelName, objectName, err = setCACertAtApp(ctx, setCACertApp, cert)
	case setCACertDomain != "":
		levelName, objectName, err = setCACertAtDomain(ctx, setCACertDomain, cert)
	case setCACertEcosystem != "":
		levelName, objectName, err = setCACertAtEcosystem(ctx, setCACertEcosystem, cert)
	case setCACertGlobal:
		levelName, objectName, err = setCACertGlobalLevel(ctx, cert)
	default:
		return fmt.Errorf("no hierarchy level specified")
	}

	if err != nil {
		return err
	}

	if setCACertDryRun {
		objectName = objectName + " (dry-run)"
	}

	render.Success(fmt.Sprintf("CA cert set: %s (vault-secret=%s)", certName, setCACertVaultSecret))
	render.Info(fmt.Sprintf("Level: %s", levelName))
	render.Info(fmt.Sprintf("Object: %s", objectName))
	if setCACertDryRun {
		render.Info("(No changes applied — dry-run mode)")
	}
	return nil
}

// setCACertAtEcosystem sets a CA cert at the ecosystem level.
func setCACertAtEcosystem(ctx resource.Context, ecosystemName string, cert models.CACertConfig) (levelName, objectName string, err error) {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return "", "", fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecosystem := ecosystemRes.Ecosystem()

	if setCACertDryRun {
		return "ecosystem", ecosystemName, nil
	}

	ecoYAML := ecosystem.ToYAML(nil)
	ecoYAML.Spec.CACerts = upsertCACertInSlice(ecoYAML.Spec.CACerts, cert)

	data, err := yaml.Marshal(ecoYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-ca-cert"); err != nil {
		return "", "", fmt.Errorf("failed to update ecosystem: %w", err)
	}

	return "ecosystem", ecosystemName, nil
}

// setCACertAtDomain sets a CA cert at the domain level.
func setCACertAtDomain(ctx resource.Context, domainName string, cert models.CACertConfig) (levelName, objectName string, err error) {
	res, err := resource.Get(ctx, handlers.KindDomain, domainName)
	if err != nil {
		return "", "", fmt.Errorf("domain %q not found: %w", domainName, err)
	}

	domainRes := res.(*handlers.DomainResource)
	domain := domainRes.Domain()

	if setCACertDryRun {
		return "domain", domainName, nil
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get DataStore: %w", err)
	}
	eco, err := ds.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get ecosystem for domain: %w", err)
	}

	domainYAML := domain.ToYAML(eco.Name, nil)
	domainYAML.Spec.CACerts = upsertCACertInSlice(domainYAML.Spec.CACerts, cert)

	data, err := yaml.Marshal(domainYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal domain YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-ca-cert"); err != nil {
		return "", "", fmt.Errorf("failed to update domain: %w", err)
	}

	return "domain", domainName, nil
}

// setCACertAtApp sets a CA cert at the app level.
func setCACertAtApp(ctx resource.Context, appName string, cert models.CACertConfig) (levelName, objectName string, err error) {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return "", "", fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	if setCACertDryRun {
		return "app", appName, nil
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get DataStore: %w", err)
	}
	domain, err := ds.GetDomainByID(app.DomainID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get domain for app: %w", err)
	}

	appYAML := app.ToYAML(domain.Name, nil)
	appYAML.Spec.Build.CACerts = upsertCACertInSlice(appYAML.Spec.Build.CACerts, cert)

	data, err := yaml.Marshal(appYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal app YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-ca-cert"); err != nil {
		return "", "", fmt.Errorf("failed to update app: %w", err)
	}

	return "app", appName, nil
}

// setCACertAtWorkspace sets a CA cert at the workspace level.
// When scopeAppName is non-empty, it scopes the workspace lookup to that app.
func setCACertAtWorkspace(cmd *cobra.Command, ctx resource.Context, workspaceName, scopeAppName string, cert models.CACertConfig) (levelName, objectName string, err error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get DataStore: %w", err)
	}

	var appName string
	var appID int

	if scopeAppName != "" {
		app, err := ds.GetAppByNameGlobal(scopeAppName)
		if err != nil {
			return "", "", fmt.Errorf("app %q not found: %w", scopeAppName, err)
		}
		appName = scopeAppName
		appID = app.ID
	} else {
		activeApp, err := getActiveAppFromContext(ds)
		if err != nil {
			return "", "", fmt.Errorf("no app specified. Use --app <name> or 'dvm use app <name>' first")
		}
		app, err := ds.GetAppByNameGlobal(activeApp)
		if err != nil {
			return "", "", fmt.Errorf("app %q not found: %w", activeApp, err)
		}
		appName = activeApp
		appID = app.ID
	}

	workspace, err := ds.GetWorkspaceByName(appID, workspaceName)
	if err != nil {
		return "", "", fmt.Errorf("workspace %q not found under app %q: %w", workspaceName, appName, err)
	}

	if setCACertDryRun {
		return "workspace", workspaceName, nil
	}

	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		if gitRepo, err := ds.GetGitRepoByID(workspace.GitRepoID.Int64); err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}

	wsYAML := workspace.ToYAML(appName, gitRepoName)
	wsYAML.Spec.Build.CACerts = upsertCACertInSlice(wsYAML.Spec.Build.CACerts, cert)

	data, err := yaml.Marshal(wsYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal workspace YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-ca-cert"); err != nil {
		return "", "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return "workspace", workspaceName, nil
}

// setCACertGlobalLevel sets a CA cert at the global (DVM-wide) level.
func setCACertGlobalLevel(ctx resource.Context, cert models.CACertConfig) (levelName, objectName string, err error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get DataStore: %w", err)
	}

	if setCACertDryRun {
		return "global", "global-defaults", nil
	}

	if err := SetGlobalCACert(ds, cert); err != nil {
		return "", "", fmt.Errorf("failed to set global CA cert: %w", err)
	}

	return "global", "global-defaults", nil
}

// deleteCACertFromDirectJSON removes a CA cert by name from a direct JSON array
// (used for eco.ca_certs and domain.ca_certs columns).
func deleteCACertFromDirectJSON(raw, certName string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var certs []models.CACertConfig
	if err := json.Unmarshal([]byte(raw), &certs); err != nil {
		return "", fmt.Errorf("parsing CA certs JSON: %w", err)
	}
	certs, _ = removeCACertFromSlice(certs, certName)
	if len(certs) == 0 {
		return "", nil
	}
	b, err := json.Marshal(certs)
	if err != nil {
		return "", fmt.Errorf("encoding CA certs JSON: %w", err)
	}
	return string(b), nil
}
