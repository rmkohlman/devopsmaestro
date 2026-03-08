package cmd

import (
	"context"
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/mirror"
	"devopsmaestro/pkg/nvimops/theme/library"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/render"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
)

// attachFlags holds the hierarchy flags for the attach command
var attachFlags HierarchyFlags

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

Examples:
  dvm attach                           # Use current context, sync mirror
  dvm attach --no-sync                 # Use current context, skip sync
  dvm attach -a portal                 # Attach to workspace in 'portal' app
  dvm attach -e healthcare -a portal   # Specify ecosystem and app
  dvm attach -a portal -w staging      # Specify app and workspace name`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAttach(cmd); err != nil {
			render.Error(err.Error())
		}
	},
}

func runAttach(cmd *cobra.Command) error {
	slog.Info("starting attach")

	// Get datastore from context
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return fmt.Errorf("dataStore not initialized")
	}
	ds := *dataStore

	var app *models.App
	var workspace *models.Workspace
	var appName, workspaceName string
	var ecosystemName, domainName string // For hierarchical container naming

	// Check if hierarchy flags were provided
	if attachFlags.HasAnyFlag() {
		// Use resolver to find workspace
		slog.Debug("using hierarchy flags", "ecosystem", attachFlags.Ecosystem,
			"domain", attachFlags.Domain, "app", attachFlags.App, "workspace", attachFlags.Workspace)

		wsResolver := resolver.NewWorkspaceResolver(ds)
		result, err := wsResolver.Resolve(attachFlags.ToFilter())
		if err != nil {
			// Check if ambiguous and provide helpful output
			if ambiguousErr, ok := resolver.IsAmbiguousError(err); ok {
				render.Warning("Multiple workspaces match your criteria")
				fmt.Println(ambiguousErr.FormatDisambiguation())
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Info("Hint: Use 'dvm get workspaces' to see available workspaces")
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

		// Update context to the resolved workspace
		if err := updateContextFromHierarchy(ds, result); err != nil {
			slog.Warn("failed to update context", "error", err)
			// Continue anyway - this is not fatal
		}

		render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	} else {
		// Fall back to existing context-based behavior (DB-backed)
		var err error
		appName, err = getActiveAppFromContext(ds)
		if err != nil {
			render.Info("Hint: Set active app with: dvm use app <name>")
			render.Info("      Or use flags: dvm attach -a <app>")
			return err
		}

		workspaceName, err = getActiveWorkspaceFromContext(ds)
		if err != nil {
			render.Info("Hint: Set active workspace with: dvm use workspace <name>")
			render.Info("      Or use flags: dvm attach -w <workspace>")
			return err
		}

		// Get app (search globally across all domains)
		app, err = ds.GetAppByNameGlobal(appName)
		if err != nil {
			return fmt.Errorf("failed to get app '%s': %w", appName, err)
		}

		// Get workspace
		workspace, err = ds.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			return fmt.Errorf("failed to get workspace '%s': %w", workspaceName, err)
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
		render.Info("Hint: Install OrbStack, Docker Desktop, or Colima")
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
		render.Info("Run 'dvm build' first to build the development container.")
		fmt.Println()
		return fmt.Errorf("workspace not built: run 'dvm build' first")
	}

	// Compute container name using hierarchical naming strategy
	namingStrategy := operators.NewHierarchicalNamingStrategy()
	containerName := namingStrategy.GenerateName(ecosystemName, domainName, appName, workspaceName)
	slog.Debug("container details", "name", containerName, "image", imageName)

	// Start workspace (handles existing containers automatically)
	render.Progress("Starting workspace container...")

	// Get correct mount path (workspace repo path if GitRepoID set, else app.Path)
	mountPath, err := getMountPath(ds, workspace, app.Path)
	if err != nil {
		return fmt.Errorf("failed to get mount path: %w", err)
	}

	containerID, err := runtime.StartWorkspace(context.Background(), operators.StartOptions{
		ImageName:     imageName,
		WorkspaceName: workspaceName,
		ContainerName: containerName,
		AppName:       appName,
		EcosystemName: ecosystemName,
		DomainName:    domainName,
		AppPath:       mountPath,
	})
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}

	slog.Info("workspace started", "container_id", containerID)

	// Attach to workspace
	render.Progress("Attaching to workspace...")
	slog.Info("attaching to container", "name", containerName)

	// Build base environment variables
	envVars := map[string]string{
		"TERM":          "xterm-256color",
		"DVM_WORKSPACE": workspaceName,
		"DVM_APP":       appName,
	}

	// Add ecosystem and domain if available
	if ecosystemName != "" {
		envVars["DVM_ECOSYSTEM"] = ecosystemName
	}
	if domainName != "" {
		envVars["DVM_DOMAIN"] = domainName
	}

	// Load theme colors and add to environment
	themeName := getThemeName(workspace)
	if themeName != "" {
		if themeEnvVars, err := loadThemeEnvVars(themeName); err == nil {
			// Merge theme env vars
			for k, v := range themeEnvVars {
				envVars[k] = v
			}
			slog.Info("loaded theme colors", "theme", themeName, "colors", len(themeEnvVars))
		} else {
			slog.Warn("failed to load theme colors", "theme", themeName, "error", err)
		}
	}

	// Build AttachOptions with environment variables for proper terminal and workspace context
	attachOpts := operators.AttachOptions{
		WorkspaceID: containerName,
		Env:         envVars,
		Shell:       "/bin/zsh",
		LoginShell:  true,
	}

	if err := runtime.AttachToWorkspace(context.Background(), attachOpts); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

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
		repoPath, err := ds.GetWorkspaceRepoPath(workspace.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get workspace repo path: %w", err)
		}
		return repoPath, nil
	}
	return appPath, nil
}

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
	AddHierarchyFlags(attachCmd, &attachFlags)
	attachCmd.Flags().Bool("no-sync", false, "Skip syncing git mirror before attach")
}
