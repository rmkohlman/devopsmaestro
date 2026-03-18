// Package cmd provides the 'dvm delete build-arg' command for removing build args.
// Supports deletion at every hierarchy level: global, ecosystem, domain, app, workspace.
// Prompts for confirmation by default; use --force to skip.
package cmd

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for delete build-arg command
var (
	deleteBuildArgEcosystem string
	deleteBuildArgDomain    string
	deleteBuildArgApp       string
	deleteBuildArgWorkspace string
	deleteBuildArgGlobal    bool
	deleteBuildArgForce     bool
)

// deleteBuildArgCmd removes a build arg key at a specific hierarchy level
var deleteBuildArgCmd = &cobra.Command{
	Use:   "build-arg KEY",
	Short: "Delete a build arg at hierarchy level",
	Long: `Delete a Docker build argument key at ecosystem, domain, app, workspace, or global level.

If the key does not exist at the specified level, the operation is a no-op (no error).
By default, you will be prompted for confirmation. Use --force to skip.

Examples:
  dvm delete build-arg PIP_INDEX_URL --ecosystem my-eco
  dvm delete build-arg EXTRA_PACKAGES --domain data-sci
  dvm delete build-arg CGO_ENABLED --app ml-api
  dvm delete build-arg DEBUG_BUILD --workspace dev
  dvm delete build-arg PIP_INDEX_URL --global
  dvm delete build-arg PIP_INDEX_URL --global --force   # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteBuildArg,
}

func init() {
	deleteCmd.AddCommand(deleteBuildArgCmd)

	deleteBuildArgCmd.Flags().StringVar(&deleteBuildArgEcosystem, "ecosystem", "", "Delete at ecosystem level")
	deleteBuildArgCmd.Flags().StringVar(&deleteBuildArgDomain, "domain", "", "Delete at domain level")
	deleteBuildArgCmd.Flags().StringVar(&deleteBuildArgApp, "app", "", "Delete at app level")
	deleteBuildArgCmd.Flags().StringVar(&deleteBuildArgWorkspace, "workspace", "", "Delete at workspace level")
	deleteBuildArgCmd.Flags().BoolVar(&deleteBuildArgGlobal, "global", false, "Delete from DVM-wide defaults")
	deleteBuildArgCmd.Flags().BoolVarP(&deleteBuildArgForce, "force", "f", false, "Skip confirmation prompt")
}

func runDeleteBuildArg(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Validate that at least one target flag is provided
	if deleteBuildArgEcosystem == "" && deleteBuildArgDomain == "" && deleteBuildArgApp == "" &&
		deleteBuildArgWorkspace == "" && !deleteBuildArgGlobal {
		return fmt.Errorf("at least one of --ecosystem, --domain, --app, --workspace, or --global must be specified")
	}

	// Validate that --global is exclusive with all other level flags
	if deleteBuildArgGlobal && (deleteBuildArgEcosystem != "" || deleteBuildArgDomain != "" ||
		deleteBuildArgApp != "" || deleteBuildArgWorkspace != "") {
		return fmt.Errorf("--global cannot be used with --ecosystem, --domain, --app, or --workspace")
	}

	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Confirm deletion
	if !deleteBuildArgForce {
		levelDesc := resolveBuildArgLevelDesc()
		fmt.Printf("Delete build arg %q at %s? (y/N): ", key, levelDesc)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if response != "y" && response != "Y" {
			render.Info("Aborted")
			return nil
		}
	}

	switch {
	case deleteBuildArgWorkspace != "":
		return deleteBuildArgAtWorkspace(ctx, deleteBuildArgWorkspace, deleteBuildArgApp, key)
	case deleteBuildArgApp != "":
		return deleteBuildArgAtApp(ctx, deleteBuildArgApp, key)
	case deleteBuildArgDomain != "":
		return deleteBuildArgAtDomain(ctx, deleteBuildArgDomain, key)
	case deleteBuildArgEcosystem != "":
		return deleteBuildArgAtEcosystem(ctx, deleteBuildArgEcosystem, key)
	case deleteBuildArgGlobal:
		return deleteBuildArgGlobalLevel(ctx, key)
	default:
		return fmt.Errorf("no hierarchy level specified")
	}
}

// resolveBuildArgLevelDesc returns a human-readable description of the target level.
func resolveBuildArgLevelDesc() string {
	switch {
	case deleteBuildArgWorkspace != "":
		return fmt.Sprintf("workspace %q", deleteBuildArgWorkspace)
	case deleteBuildArgApp != "":
		return fmt.Sprintf("app %q", deleteBuildArgApp)
	case deleteBuildArgDomain != "":
		return fmt.Sprintf("domain %q", deleteBuildArgDomain)
	case deleteBuildArgEcosystem != "":
		return fmt.Sprintf("ecosystem %q", deleteBuildArgEcosystem)
	case deleteBuildArgGlobal:
		return "global defaults"
	default:
		return "unknown level"
	}
}

// deleteBuildArgAtEcosystem removes a build arg key from the ecosystem level.
func deleteBuildArgAtEcosystem(ctx resource.Context, ecosystemName, key string) error {
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecosystem := ecosystemRes.Ecosystem()

	// No-op if build args are not set or key doesn't exist
	if !ecosystem.BuildArgs.Valid || ecosystem.BuildArgs.String == "" {
		render.Info(fmt.Sprintf("Build arg %q not set at ecosystem level (%s) — nothing to delete", key, ecosystemName))
		return nil
	}

	updatedJSON, err := deleteBuildArgFromJSON(ecosystem.BuildArgs.String, key)
	if err != nil {
		return fmt.Errorf("failed to update ecosystem build args: %w", err)
	}

	// Build an updated ecosystem YAML and apply
	ecoYAML := ecosystem.ToYAML(nil)
	ecoYAML.Spec.Build.Args = parseDirectJSONMap(updatedJSON)

	data, err := yaml.Marshal(ecoYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-build-arg"); err != nil {
		return fmt.Errorf("failed to update ecosystem: %w", err)
	}

	render.Success(fmt.Sprintf("Build arg %q deleted from ecosystem %q", key, ecosystemName))
	return nil
}

// deleteBuildArgAtDomain removes a build arg key from the domain level.
func deleteBuildArgAtDomain(ctx resource.Context, domainName, key string) error {
	res, err := resource.Get(ctx, handlers.KindDomain, domainName)
	if err != nil {
		return fmt.Errorf("domain %q not found: %w", domainName, err)
	}

	domainRes := res.(*handlers.DomainResource)
	domain := domainRes.Domain()

	if !domain.BuildArgs.Valid || domain.BuildArgs.String == "" {
		render.Info(fmt.Sprintf("Build arg %q not set at domain level (%s) — nothing to delete", key, domainName))
		return nil
	}

	updatedJSON, err := deleteBuildArgFromJSON(domain.BuildArgs.String, key)
	if err != nil {
		return fmt.Errorf("failed to update domain build args: %w", err)
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
	domainYAML.Spec.Build.Args = parseDirectJSONMap(updatedJSON)

	data, err := yaml.Marshal(domainYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal domain YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-build-arg"); err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	render.Success(fmt.Sprintf("Build arg %q deleted from domain %q", key, domainName))
	return nil
}

// deleteBuildArgAtApp removes a build arg key from the app level.
func deleteBuildArgAtApp(ctx resource.Context, appName, key string) error {
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	if !app.BuildConfig.Valid || app.BuildConfig.String == "" {
		render.Info(fmt.Sprintf("Build arg %q not set at app level (%s) — nothing to delete", key, appName))
		return nil
	}

	// App build_config is {"args": {...}, ...}
	updatedJSON, err := deleteArgFromWrappedJSON(app.BuildConfig.String, key)
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
	appYAML.Spec.Build.Args = parseArgsFromWrappedJSON(updatedJSON)

	data, err := yaml.Marshal(appYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal app YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-build-arg"); err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	render.Success(fmt.Sprintf("Build arg %q deleted from app %q", key, appName))
	return nil
}

// deleteBuildArgAtWorkspace removes a build arg key from the workspace level.
func deleteBuildArgAtWorkspace(ctx resource.Context, workspaceName, scopeAppName, key string) error {
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
		render.Info(fmt.Sprintf("Build arg %q not set at workspace level (%s) — nothing to delete", key, workspaceName))
		return nil
	}

	// Workspace build_config is {"args": {...}, ...}
	updatedJSON, err := deleteArgFromWrappedJSON(workspace.BuildConfig.String, key)
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
	wsYAML.Spec.Build.Args = parseArgsFromWrappedJSON(updatedJSON)

	data, err := yaml.Marshal(wsYAML)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace YAML: %w", err)
	}
	if _, err := resource.Apply(ctx, data, "delete-build-arg"); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("Build arg %q deleted from workspace %q", key, workspaceName))
	return nil
}

// deleteBuildArgGlobalLevel removes a build arg key from the global defaults.
func deleteBuildArgGlobalLevel(ctx resource.Context, key string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DataStore: %w", err)
	}

	if err := DeleteGlobalBuildArg(ds, key); err != nil {
		return fmt.Errorf("failed to delete global build arg: %w", err)
	}

	render.Success(fmt.Sprintf("Build arg %q deleted from global defaults", key))
	return nil
}

// ─── internal helpers ──────────────────────────────────────────────────────────

// deleteArgFromWrappedJSON removes a key from a JSON blob of the form {"args": {...}, ...}.
// Used for app.build_config and workspace.build_config.
func deleteArgFromWrappedJSON(raw, key string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var wrapper map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return "", fmt.Errorf("parsing build config JSON: %w", err)
	}

	argsRaw, ok := wrapper["args"]
	if !ok {
		return raw, nil // no "args" key — nothing to delete
	}

	argsMap, ok := argsRaw.(map[string]interface{})
	if !ok {
		return raw, nil
	}
	delete(argsMap, key)
	wrapper["args"] = argsMap

	b, err := json.Marshal(wrapper)
	if err != nil {
		return "", fmt.Errorf("encoding build config JSON: %w", err)
	}
	return string(b), nil
}

// parseDirectJSONMap parses a direct JSON map string (may be empty/blank).
func parseDirectJSONMap(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	return m
}

// parseArgsFromWrappedJSON extracts the "args" field from a wrapped build config JSON.
func parseArgsFromWrappedJSON(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var wrapper struct {
		Args map[string]string `json:"args"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return nil
	}
	return wrapper.Args
}

// nullString returns a valid sql.NullString if s is non-empty, else an invalid one.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
