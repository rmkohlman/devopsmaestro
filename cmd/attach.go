package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
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

Press Ctrl+D to detach from the workspace.

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name

Examples:
  dvm attach                           # Use current context
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

		// Update context to the resolved workspace
		if err := updateContextFromHierarchy(ds, result); err != nil {
			slog.Warn("failed to update context", "error", err)
			// Continue anyway - this is not fatal
		}

		render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	} else {
		// Fall back to existing context-based behavior
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to initialize context manager: %w", err)
		}

		appName, err = contextMgr.GetActiveApp()
		if err != nil {
			render.Info("Hint: Set active app with: dvm use app <name>")
			render.Info("      Or use flags: dvm attach -a <app>")
			return err
		}

		workspaceName, err = contextMgr.GetActiveWorkspace()
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

	containerName := fmt.Sprintf("dvm-%s-%s", appName, workspaceName)
	slog.Debug("container details", "name", containerName, "image", imageName)

	// Start workspace (handles existing containers automatically)
	render.Progress("Starting workspace container...")

	containerID, err := runtime.StartWorkspace(context.Background(), operators.StartOptions{
		ImageName:     imageName,
		WorkspaceName: workspaceName,
		ContainerName: containerName,
		AppName:       appName,
		AppPath:       app.Path,
	})
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}

	slog.Info("workspace started", "container_id", containerID)

	// Attach to workspace
	render.Progress("Attaching to workspace...")
	slog.Info("attaching to container", "name", containerName)

	if err := runtime.AttachToWorkspace(context.Background(), containerName); err != nil {
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

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
	AddHierarchyFlags(attachCmd, &attachFlags)
}
