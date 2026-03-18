// Package cmd provides the 'dvm set build-arg' command for setting hierarchical build args.
// Build args cascade down the hierarchy: global < ecosystem < domain < app < workspace.
// More-specific levels override less-specific levels (workspace wins).
//
// For secrets, use 'dvm credential' instead — build args are stored in plain text.
package cmd

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/pkg/envvalidation"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for set build-arg command
var (
	setBuildArgEcosystem string
	setBuildArgDomain    string
	setBuildArgApp       string
	setBuildArgWorkspace string
	setBuildArgGlobal    bool
	setBuildArgDryRun    bool
)

// setBuildArgCmd sets a build arg at a specific hierarchy level
var setBuildArgCmd = &cobra.Command{
	Use:   "build-arg KEY VALUE",
	Short: "Set a build arg at hierarchy level",
	Long: `Set a Docker build argument at ecosystem, domain, app, workspace, or global level.

Build args cascade down the hierarchy (more-specific levels override parents):
  global → ecosystem → domain → app → workspace

For secrets, use 'dvm credential' instead — build args are stored in plain text.

Examples:
  dvm set build-arg PIP_INDEX_URL https://pypi.example.com --ecosystem my-eco
  dvm set build-arg EXTRA_PACKAGES "numpy pandas" --domain data-sci
  dvm set build-arg CGO_ENABLED 0 --app ml-api
  dvm set build-arg DEBUG_BUILD true --workspace dev
  dvm set build-arg PIP_INDEX_URL https://pypi.example.com --global

Flags:
  --ecosystem   Set at ecosystem level
  --domain      Set at domain level
  --app         Set at app level
  --workspace   Set at workspace level
  --global      Set as DVM-wide default (applies to all workspaces)
  --dry-run     Preview changes without applying`,
	Args: cobra.ExactArgs(2),
	RunE: runSetBuildArg,
}

func init() {
	setCmd.AddCommand(setBuildArgCmd)

	setBuildArgCmd.Flags().StringVar(&setBuildArgEcosystem, "ecosystem", "", "Set at ecosystem level")
	setBuildArgCmd.Flags().StringVar(&setBuildArgDomain, "domain", "", "Set at domain level")
	setBuildArgCmd.Flags().StringVar(&setBuildArgApp, "app", "", "Set at app level")
	setBuildArgCmd.Flags().StringVar(&setBuildArgWorkspace, "workspace", "", "Set at workspace level")
	setBuildArgCmd.Flags().BoolVar(&setBuildArgGlobal, "global", false, "Set as DVM-wide default")
	setBuildArgCmd.Flags().BoolVar(&setBuildArgDryRun, "dry-run", false, "Preview changes without applying")
}

func runSetBuildArg(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Validate that at least one target flag is provided
	if setBuildArgEcosystem == "" && setBuildArgDomain == "" && setBuildArgApp == "" &&
		setBuildArgWorkspace == "" && !setBuildArgGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	// Validate that --global is exclusive with all other level flags
	if setBuildArgGlobal && (setBuildArgEcosystem != "" || setBuildArgDomain != "" ||
		setBuildArgApp != "" || setBuildArgWorkspace != "") {
		return fmt.Errorf("--global cannot be used with --ecosystem, --domain, --app, or --workspace")
	}

	// Validate key at the CLI entry point
	if err := envvalidation.ValidateEnvKey(key); err != nil {
		return err
	}

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Dispatch to level-specific handler
	var levelName, objectName string
	switch {
	case setBuildArgWorkspace != "":
		levelName, objectName, err = setBuildArgAtWorkspace(cmd, ctx, setBuildArgWorkspace, setBuildArgApp, key, value)
	case setBuildArgApp != "":
		levelName, objectName, err = setBuildArgAtApp(ctx, setBuildArgApp, key, value)
	case setBuildArgDomain != "":
		levelName, objectName, err = setBuildArgAtDomain(ctx, setBuildArgDomain, key, value)
	case setBuildArgEcosystem != "":
		levelName, objectName, err = setBuildArgAtEcosystem(ctx, setBuildArgEcosystem, key, value)
	case setBuildArgGlobal:
		levelName, objectName, err = setBuildArgGlobalLevel(ctx, key, value)
	default:
		return fmt.Errorf("no hierarchy level specified")
	}

	if err != nil {
		return err
	}

	suffix := ""
	if setBuildArgDryRun {
		suffix = " (dry-run)"
		objectName = objectName + suffix
	}

	render.Success(fmt.Sprintf("Build arg set: %s=%s", key, value))
	render.Info(fmt.Sprintf("Level: %s", levelName))
	render.Info(fmt.Sprintf("Object: %s", objectName))
	if setBuildArgDryRun {
		render.Info("(No changes applied — dry-run mode)")
	}
	return nil
}

// setBuildArgAtEcosystem sets a build arg at the ecosystem level.
func setBuildArgAtEcosystem(ctx resource.Context, ecosystemName, key, value string) (levelName, objectName string, err error) {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return "", "", fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecosystem := ecosystemRes.Ecosystem()

	if setBuildArgDryRun {
		return "ecosystem", ecosystemName, nil
	}

	// Merge new key into existing build args
	ecoYAML := ecosystem.ToYAML(nil)
	if ecoYAML.Spec.Build.Args == nil {
		ecoYAML.Spec.Build.Args = make(map[string]string)
	}
	ecoYAML.Spec.Build.Args[key] = value

	data, err := yaml.Marshal(ecoYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-build-arg"); err != nil {
		return "", "", fmt.Errorf("failed to update ecosystem: %w", err)
	}

	return "ecosystem", ecosystemName, nil
}

// setBuildArgAtDomain sets a build arg at the domain level.
func setBuildArgAtDomain(ctx resource.Context, domainName, key, value string) (levelName, objectName string, err error) {
	res, err := resource.Get(ctx, handlers.KindDomain, domainName)
	if err != nil {
		return "", "", fmt.Errorf("domain %q not found: %w", domainName, err)
	}

	domainRes := res.(*handlers.DomainResource)
	domain := domainRes.Domain()

	if setBuildArgDryRun {
		return "domain", domainName, nil
	}

	// Lookup ecosystem name for ToYAML
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get DataStore: %w", err)
	}
	ecosystem, err := ds.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get ecosystem for domain: %w", err)
	}

	domainYAML := domain.ToYAML(ecosystem.Name, nil)
	if domainYAML.Spec.Build.Args == nil {
		domainYAML.Spec.Build.Args = make(map[string]string)
	}
	domainYAML.Spec.Build.Args[key] = value

	data, err := yaml.Marshal(domainYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal domain YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-build-arg"); err != nil {
		return "", "", fmt.Errorf("failed to update domain: %w", err)
	}

	return "domain", domainName, nil
}

// setBuildArgAtApp sets a build arg at the app level.
func setBuildArgAtApp(ctx resource.Context, appName, key, value string) (levelName, objectName string, err error) {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return "", "", fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	if setBuildArgDryRun {
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
	if appYAML.Spec.Build.Args == nil {
		appYAML.Spec.Build.Args = make(map[string]string)
	}
	appYAML.Spec.Build.Args[key] = value

	data, err := yaml.Marshal(appYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal app YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-build-arg"); err != nil {
		return "", "", fmt.Errorf("failed to update app: %w", err)
	}

	return "app", appName, nil
}

// setBuildArgAtWorkspace sets a build arg at the workspace level.
// When scopeAppName is non-empty, it scopes the workspace lookup to that app.
func setBuildArgAtWorkspace(cmd *cobra.Command, ctx resource.Context, workspaceName, scopeAppName, key, value string) (levelName, objectName string, err error) {
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
		// Fall back to active app context
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

	if setBuildArgDryRun {
		return "workspace", workspaceName, nil
	}

	// Resolve git repo name if set
	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		if gitRepo, err := ds.GetGitRepoByID(workspace.GitRepoID.Int64); err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}

	wsYAML := workspace.ToYAML(appName, gitRepoName)
	if wsYAML.Spec.Build.Args == nil {
		wsYAML.Spec.Build.Args = make(map[string]string)
	}
	wsYAML.Spec.Build.Args[key] = value

	data, err := yaml.Marshal(wsYAML)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal workspace YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "set-build-arg"); err != nil {
		return "", "", fmt.Errorf("failed to update workspace: %w", err)
	}

	return "workspace", workspaceName, nil
}

// setBuildArgGlobalLevel sets a build arg at the global (DVM-wide) level.
func setBuildArgGlobalLevel(ctx resource.Context, key, value string) (levelName, objectName string, err error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get DataStore: %w", err)
	}

	if setBuildArgDryRun {
		return "global", "global-defaults", nil
	}

	// SetGlobalBuildArg handles validation + read-modify-write
	if err := SetGlobalBuildArg(ds, key, value); err != nil {
		return "", "", fmt.Errorf("failed to set global build arg: %w", err)
	}

	return "global", "global-defaults", nil
}

// deleteBuildArgFromJSON removes a key from a JSON-encoded map[string]string blob
// and returns the updated JSON. Used by delete_build_arg.go.
func deleteBuildArgFromJSON(raw, key string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return "", fmt.Errorf("parsing build args JSON: %w", err)
	}
	delete(m, key)
	if len(m) == 0 {
		return "", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("encoding build args JSON: %w", err)
	}
	return string(b), nil
}
