package cmd

import (
	"context"
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/envvalidation"
	"devopsmaestro/pkg/mirror"
	"devopsmaestro/pkg/registry/envinjector"
	"devopsmaestro/pkg/resolver"
	ws "devopsmaestro/pkg/workspace"
	"fmt"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTheme/library"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// attachFlags holds the hierarchy flags for the attach command
var attachFlags HierarchyFlags

// attachTimeout holds the timeout duration for the attach operation
var attachTimeout time.Duration

// attachDryRun holds the dry-run flag for the attach command
var attachDryRun bool

// attachNetworkMode holds the network isolation mode for the container
var attachNetworkMode string

// attachCPUs holds the CPU limit for the container
var attachCPUs float64

// attachMemory holds the memory limit for the container
var attachMemory string

// attachCmd attaches to the active workspace
var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to your workspace container",
	Long: `Attach an interactive terminal to your active workspace container.
If the workspace is not running, it will be started automatically.
If the container image doesn't exist, it will be built first.

The workspace provides your complete dev environment with:
- Neovim configuration
- oh-my-zsh + Powerlevel10k theme
- Your app files mounted at /workspace

If the workspace is associated with a GitRepo, the mirror is synced
automatically before attach unless --no-sync is specified.

Press Ctrl+D to detach from the workspace.

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name
      --no-sync     Skip syncing git mirror before attach
      --network     Network mode: bridge (default), none, host, or custom name
      --cpus        CPU limit (e.g., 1.5 for 1.5 cores)
      --memory      Memory limit (e.g., 512m, 2g)

Examples:
  dvm attach                           # Use current context, sync mirror
  dvm attach --no-sync                 # Use current context, skip sync
  dvm attach -a portal                 # Attach to workspace in 'portal' app
  dvm attach -e healthcare -a portal   # Specify ecosystem and app
  dvm attach -a portal -w staging      # Specify app and workspace name
  dvm attach --network=none            # Isolate container from network
  dvm attach --cpus=2 --memory=4g      # Limit to 2 CPUs and 4GB RAM`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runAttach(cmd); err != nil {
			render.Error(err.Error())
			return errSilent
		}
		return nil
	},
}

func runAttach(cmd *cobra.Command) error {
	slog.Info("starting attach")

	// Dry-run: preview what would happen
	if attachDryRun {
		details := []string{"Would attach to workspace container"}
		if attachNetworkMode != "" {
			details = append(details, fmt.Sprintf("network=%s", attachNetworkMode))
		}
		if attachCPUs > 0 {
			details = append(details, fmt.Sprintf("cpus=%.1f", attachCPUs))
		}
		if attachMemory != "" {
			details = append(details, fmt.Sprintf("memory=%s", attachMemory))
		}
		render.Plain(strings.Join(details, ", "))
		return nil
	}

	// Create timeout context from flag
	ctx := context.Background()
	if attachTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, attachTimeout)
		defer cancel()
	}

	// Get datastore from context
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("dataStore not initialized: %w", err)
	}

	var app *models.App
	var workspace *models.Workspace
	var appName, workspaceName string
	var ecosystemName, domainName, systemName string // For hierarchical container naming

	// Check if hierarchy flags were provided
	if attachFlags.HasAnyFlag() {
		// Use resolver to find workspace
		slog.Debug("using hierarchy flags", "ecosystem", attachFlags.Ecosystem,
			"domain", attachFlags.Domain, "system", attachFlags.System, "app", attachFlags.App, "workspace", attachFlags.Workspace)

		wsResolver := resolver.NewWorkspaceResolver(ds)
		result, err := wsResolver.Resolve(attachFlags.ToFilter())
		if err != nil {
			// Check if ambiguous and provide helpful output
			if ambiguousErr, ok := resolver.IsAmbiguousError(err); ok {
				render.Warning("Multiple workspaces match your criteria")
				render.Plain(ambiguousErr.FormatDisambiguation())
				render.Plain(FormatSuggestions(SuggestAmbiguousWorkspace()...))
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Plain(FormatSuggestions(SuggestWorkspaceNotFound("")...))
				return err
			}
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}

		// Use resolved workspace and app
		workspace = result.Workspace
		app = result.App
		appName = app.Name
		workspaceName = workspace.Name
		ecosystemName = result.Ecosystem.Name
		domainName = result.Domain.Name
		if result.System != nil {
			systemName = result.System.Name
		}

		render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	} else {
		// Fall back to existing context-based behavior (DB-backed)
		var err error
		appName, err = getActiveAppFromContext(ds)
		if err != nil {
			render.Plain(FormatSuggestions(SuggestNoActiveApp()...))
			return err
		}

		workspaceName, err = getActiveWorkspaceFromContext(ds)
		if err != nil {
			render.Plain(FormatSuggestions(SuggestNoActiveWorkspace()...))
			return err
		}

		// Get app (search globally across all domains)
		app, err = ds.GetAppByNameGlobal(appName)
		if err != nil {
			return ErrorWithSuggestion(
				fmt.Sprintf("app %q not found", appName),
				SuggestAppNotFound(appName)...,
			)
		}

		// Get workspace
		workspace, err = ds.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			return ErrorWithSuggestion(
				fmt.Sprintf("workspace %q not found in app %q", workspaceName, appName),
				SuggestWorkspaceNotFound(workspaceName)...,
			)
		}
	}

	slog.Debug("attach context", "app", appName, "workspace", workspaceName)
	render.Info(fmt.Sprintf("App: %s | Workspace: %s", appName, workspaceName))

	slog.Debug("resolved workspace", "image", workspace.ImageName, "app_path", app.Path)

	// Sync git mirror if workspace has git_repo_id (unless --no-sync)
	noSync, _ := cmd.Flags().GetBool("no-sync")
	if workspace.GitRepoID.Valid && !noSync {
		gitRepo, err := ds.GetGitRepoByID(workspace.GitRepoID.Int64)
		if err == nil && gitRepo.AutoSync {
			render.Progress(fmt.Sprintf("Syncing mirror '%s'...", gitRepo.Name))
			baseDir := getGitRepoBaseDir()
			mirrorMgr := mirror.NewGitMirrorManager(baseDir)
			if err := mirrorMgr.Sync(gitRepo.Slug); err != nil {
				slog.Warn("failed to sync mirror", "repo", gitRepo.Name, "error", err)
				render.Warning(fmt.Sprintf("Mirror sync failed: %v", err))
				// Continue with attach - don't fail
			} else {
				render.Success("Mirror up to date")
			}
		}
	} else if workspace.GitRepoID.Valid && noSync {
		render.Info("Skipping mirror sync (--no-sync)")
	}

	// Create container runtime using factory
	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Plain(FormatSuggestions(SuggestNoContainerRuntime()...))
		return fmt.Errorf("failed to create container runtime: %w", err)
	}

	slog.Info("using runtime", "type", runtime.GetRuntimeType(), "platform", runtime.GetPlatformName())
	render.Info(fmt.Sprintf("Platform: %s", runtime.GetPlatformName()))

	// Use image name from workspace
	imageName := workspace.ImageName

	// Check if workspace has been built (pending tag means not yet built)
	if strings.HasSuffix(imageName, ":pending") || !strings.HasPrefix(imageName, "dvm-") {
		slog.Warn("workspace image may not be built", "image", imageName)
		render.Warning(fmt.Sprintf("Workspace image '%s' has not been built yet.", imageName))
		render.Plain(FormatSuggestions(SuggestWorkspaceNotBuilt()...))
		render.Blank()
		return fmt.Errorf("workspace not built: run 'dvm build' first")
	}

	// Compute container name using hierarchical naming strategy
	namingStrategy := operators.NewHierarchicalNamingStrategy()
	containerName := namingStrategy.GenerateName(ecosystemName, domainName, systemName, appName, workspaceName)
	slog.Debug("container details", "name", containerName, "image", imageName)

	// Start workspace (handles existing containers automatically)
	render.Progress("Starting workspace container...")

	// Get correct mount path (workspace repo path if GitRepoID set, else app.Path)
	mountPath, err := getMountPath(ds, workspace, app.Path)
	if err != nil {
		return fmt.Errorf("failed to get mount path: %w", err)
	}

	// Get workspace container config for UID/GID
	workspaceYAML := workspace.ToYAML(appName, "")
	containerUID := workspaceYAML.Spec.Container.UID
	containerGID := workspaceYAML.Spec.Container.GID

	// Validate container options (network mode and resource limits)
	if err := operators.ValidateNetworkMode(attachNetworkMode); err != nil {
		return err
	}
	if err := operators.ValidateCPUs(attachCPUs); err != nil {
		return err
	}
	if attachMemory != "" {
		if _, err := operators.ParseMemoryString(attachMemory); err != nil {
			return err
		}
	}

	containerID, err := runtime.StartWorkspace(ctx, operators.StartOptions{
		ImageName:             imageName,
		WorkspaceName:         workspaceName,
		ContainerName:         containerName,
		AppName:               appName,
		EcosystemName:         ecosystemName,
		DomainName:            domainName,
		SystemName:            systemName,
		AppPath:               mountPath,
		UID:                   containerUID,
		GID:                   containerGID,
		SSHAgentForwarding:    workspace.SSHAgentForwarding,
		GitCredentialMounting: workspace.GitCredentialMounting,
		NetworkMode:           attachNetworkMode,
		CPUs:                  attachCPUs,
		Memory:                attachMemory,
	})
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}

	slog.Info("workspace started", "container_id", containerID)

	// Attach to workspace
	render.Progress("Attaching to workspace...")
	slog.Info("attaching to container", "name", containerName)

	// Load workspace env
	wsEnv := workspace.GetEnv()

	// Load theme env
	themeEnv := map[string]string{}
	themeName := getThemeName(workspace)
	if themeName != "" {
		if te, err := loadThemeEnvVars(themeName); err == nil {
			themeEnv = te
			slog.Info("loaded theme colors", "theme", themeName, "colors", len(themeEnv))
		} else {
			slog.Warn("failed to load theme colors", "theme", themeName, "error", err)
		}
	}

	// Load registry env (WI-3)
	registryEnv, _ := loadRegistryEnv(ds)

	// Load credential env (WI-2)
	credentialEnv, credWarnings := loadBuildCredentials(ds, app, workspace)
	for _, w := range credWarnings {
		render.Warning(w)
	}

	// Build the merged env
	envVars := buildRuntimeEnv(appName, workspaceName, ecosystemName, domainName, systemName, themeEnv, registryEnv, credentialEnv, wsEnv)

	// Build AttachOptions with environment variables for proper terminal and workspace context
	attachOpts := operators.AttachOptions{
		WorkspaceID: containerName,
		Env:         envVars,
		Shell:       "/bin/zsh",
		LoginShell:  true,
		UID:         containerUID,
		GID:         containerGID,
	}

	// Set terminal tab title via OSC 0 escape sequence (standard xterm protocol).
	// Any terminal that supports OSC (WezTerm, iTerm2, Kitty, etc.) will update
	// the tab/window title automatically — no terminal-specific configuration needed.
	fmt.Fprintf(os.Stderr, "\x1b]0;[dvm] %s/%s\x07", appName, workspaceName)

	if err := runtime.AttachToWorkspace(ctx, attachOpts); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b]0;\x07") // reset title on error
		return fmt.Errorf("failed to attach: %w", err)
	}

	// Reset terminal tab title to default on detach
	fmt.Fprintf(os.Stderr, "\x1b]0;\x07")

	slog.Info("session ended", "container", containerName)
	render.Info("Session ended.")
	return nil
}

// updateContextFromHierarchy updates the database context with the resolved hierarchy.
// This ensures that subsequent commands without flags use the same workspace.
func updateContextFromHierarchy(ds db.DataStore, wh *models.WorkspaceWithHierarchy) error {
	if err := ds.SetActiveEcosystem(&wh.Ecosystem.ID); err != nil {
		return err
	}
	if err := ds.SetActiveDomain(&wh.Domain.ID); err != nil {
		return err
	}
	if wh.System != nil {
		if err := ds.SetActiveSystem(&wh.System.ID); err != nil {
			return err
		}
	} else {
		if err := ds.SetActiveSystem(nil); err != nil {
			return err
		}
	}
	if err := ds.SetActiveApp(&wh.App.ID); err != nil {
		return err
	}
	if err := ds.SetActiveWorkspace(&wh.Workspace.ID); err != nil {
		return err
	}
	return nil
}

// getThemeName determines which theme to use for terminal colors.
// Priority: workspace spec > DVM config > default
func getThemeName(workspace *models.Workspace) string {
	// TODO: In future, we could parse workspace YAML and get spec.nvim.theme
	// For now, use DVM config theme
	themeName := config.GetTheme()

	// Convert UI theme names to library theme names if needed
	// The config uses names like "tokyo-night" but library uses "tokyonight-night"
	themeMapping := map[string]string{
		"tokyo-night":   "tokyonight-night",
		"gruvbox-dark":  "gruvbox-dark",
		"gruvbox-light": "gruvbox-light",
		// Most names match directly
	}

	if mapped, ok := themeMapping[themeName]; ok {
		return mapped
	}

	// If "auto", default to a sensible theme
	if themeName == "auto" || themeName == "" {
		return "tokyonight-night"
	}

	return themeName
}

// loadThemeEnvVars loads a theme from the library and returns terminal color env vars.
func loadThemeEnvVars(themeName string) (map[string]string, error) {
	theme, err := library.Get(themeName)
	if err != nil {
		return nil, fmt.Errorf("theme %q not found in library: %w", themeName, err)
	}

	return theme.TerminalEnvVars(), nil
}

// getMountPath determines the source path for mounting into a workspace container.
// When a workspace has a GitRepoID (created with --repo flag), the source code
// is in the workspace repo path (~/.devopsmaestro/workspaces/{slug}/repo/),
// not in the original app.Path. This function returns the correct path to mount.
func getMountPath(ds db.DataStore, workspace *models.Workspace, appPath string) (string, error) {
	if workspace.GitRepoID.Valid {
		repoPath, err := ws.GetWorkspaceRepoPath(workspace.Slug)
		if err != nil {
			return "", fmt.Errorf("failed to get workspace repo path: %w", err)
		}
		return repoPath, nil
	}
	return appPath, nil
}

// buildRuntimeEnv assembles the environment variable map for a workspace shell session.
// It merges env vars in increasing priority order:
//
//	Layer 1 (lowest): themeEnv     — terminal color vars from the active theme
//	Layer 2:          registryEnv  — PIP_INDEX_URL, GOPROXY, NPM_CONFIG_REGISTRY, etc.
//	Layer 3:          credentialEnv — GITHUB_TOKEN, AWS_ACCESS_KEY_ID, etc. (dangerous vars filtered)
//	Layer 4:          wsEnv        — workspace spec.env (highest user-defined priority)
//	Layer 5 (highest): metadata    — TERM, DVM_WORKSPACE, DVM_APP, DVM_ECOSYSTEM, DVM_DOMAIN
//
// Metadata vars are applied last so they can never be overridden by any env layer.
func buildRuntimeEnv(appName, workspaceName, ecosystemName, domainName, systemName string, themeEnv, registryEnv, credentialEnv, wsEnv map[string]string) map[string]string {
	env := make(map[string]string)

	// Layer 1 (lowest priority): theme env
	for k, v := range themeEnv {
		env[k] = v
	}

	// Layer 2: registry env
	for k, v := range registryEnv {
		env[k] = v
	}

	// Layer 3: credential env (filter dangerous vars, re-validate keys)
	for k, v := range credentialEnv {
		if envvalidation.IsDangerousEnvVar(k) {
			slog.Warn("blocked dangerous credential env var", "key", k)
			continue
		}
		if err := envvalidation.ValidateEnvKey(k); err != nil {
			slog.Warn("skipped credential with invalid env key", "key", k, "error", err)
			continue
		}
		env[k] = v
	}

	// Layer 4: workspace env (highest user-defined priority)
	// Note: DVM_WORKSPACE, DVM_APP, DVM_ECOSYSTEM, DVM_DOMAIN, and TERM are
	// protected by Layer 5 metadata which is applied last and always wins.
	for k, v := range wsEnv {
		env[k] = v
	}

	// Layer 5 (highest priority): DVM metadata — CANNOT be overridden
	env["TERM"] = "xterm-256color"
	env["DVM_WORKSPACE"] = workspaceName
	env["DVM_APP"] = appName
	if ecosystemName != "" {
		env["DVM_ECOSYSTEM"] = ecosystemName
	}
	if domainName != "" {
		env["DVM_DOMAIN"] = domainName
	}
	if systemName != "" {
		env["DVM_SYSTEM"] = systemName
	}

	return env
}

// loadRegistryEnv loads env vars from all enabled registries.
// Returns a map of registry-injected env vars (e.g., PIP_INDEX_URL, GOPROXY).
func loadRegistryEnv(ds db.DataStore) (map[string]string, error) {
	registries, err := ds.ListRegistries()
	if err != nil {
		slog.Warn("failed to list registries for env injection", "error", err)
		return nil, err
	}

	var enabled []*models.Registry
	for _, reg := range registries {
		if reg.Enabled {
			enabled = append(enabled, reg)
		}
	}

	if len(enabled) == 0 {
		return map[string]string{}, nil
	}

	injector := envinjector.NewEnvironmentInjector()
	envVars := injector.InjectForAttachMultiple(enabled)
	if len(envVars) > 0 {
		slog.Info("injected registry env vars", "count", len(envVars), "registries", len(enabled))
	}
	return envVars, nil
}

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
	AddHierarchyFlags(attachCmd, &attachFlags)
	attachCmd.Flags().Bool("no-sync", false, "Skip syncing git mirror before attach")
	attachCmd.Flags().DurationVar(&attachTimeout, "timeout", 10*time.Minute, "Timeout for the attach operation (e.g., 10m, 30s)")
	attachCmd.Flags().StringVar(&attachNetworkMode, "network", "", "Network mode: bridge (default), none, host, or custom network name")
	attachCmd.Flags().Float64Var(&attachCPUs, "cpus", 0, "CPU limit (e.g., 1.5 for 1.5 cores; 0 = no limit)")
	attachCmd.Flags().StringVar(&attachMemory, "memory", "", "Memory limit (e.g., 512m, 2g; empty = no limit)")
	AddDryRunFlag(attachCmd, &attachDryRun)
}
