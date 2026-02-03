package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/output"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	detachAll bool
)

// detachCmd stops the active workspace container
var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Stop and detach from a workspace container",
	Long: `Stop and detach from a workspace container.

By default, stops the currently active workspace. The container is stopped
but not removed, so you can quickly re-attach later with 'dvm attach'.

Use --all to stop all DVM workspace containers.

Examples:
  dvm detach              # Stop active workspace
  dvm detach --all        # Stop all DVM workspaces
  dvm detach -a           # Short form`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDetach(cmd)
	},
}

func init() {
	rootCmd.AddCommand(detachCmd)
	detachCmd.Flags().BoolVarP(&detachAll, "all", "a", false, "Stop all DVM workspace containers")
}

func runDetach(cmd *cobra.Command) error {
	formatter := output.ForOutput("")

	if detachAll {
		return detachAllWorkspaces(formatter)
	}

	return detachActiveWorkspace(cmd, formatter)
}

func detachActiveWorkspace(cmd *cobra.Command, f output.Formatter) error {
	// Get context manager
	contextMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to initialize context manager: %w", err)
	}

	// Get active project and workspace
	projectName, err := contextMgr.GetActiveProject()
	if err != nil {
		f.Warning("No active project set")
		f.Println("Set active project with: dvm use project <name>")
		return nil
	}

	workspaceName, err := contextMgr.GetActiveWorkspace()
	if err != nil {
		f.Warning("No active workspace set")
		f.Println("Set active workspace with: dvm use workspace <name>")
		return nil
	}

	slog.Debug("detach context", "project", projectName, "workspace", workspaceName)

	// Get datastore from context
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return fmt.Errorf("dataStore not initialized")
	}

	ds := *dataStore

	// Verify project and workspace exist
	project, err := ds.GetProjectByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectName, err)
	}

	_, err = ds.GetWorkspaceByName(project.ID, workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace '%s': %w", workspaceName, err)
	}

	// Stop the container
	containerName := fmt.Sprintf("dvm-%s-%s", projectName, workspaceName)
	return stopContainer(containerName, f)
}

func detachAllWorkspaces(f output.Formatter) error {
	f.Progress("Finding all DVM workspace containers...")

	// Get Colima profile
	profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
	if profile == "" {
		profile = "default"
	}

	// List all DVM containers
	listCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "ps", "-q", "-f", "name=dvm-")
	listOutput, err := listCmd.Output()
	if err != nil {
		slog.Debug("no containers found or error listing", "error", err)
		f.Info("No running DVM workspace containers found")
		return nil
	}

	containers := splitLines(string(listOutput))
	if len(containers) == 0 {
		f.Info("No running DVM workspace containers found")
		return nil
	}

	f.Info(fmt.Sprintf("Found %d running workspace(s)", len(containers)))

	// Get container names for display
	namesCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "ps", "--format", "{{.Names}}", "-f", "name=dvm-")
	namesOutput, _ := namesCmd.Output()
	names := splitLines(string(namesOutput))

	// Stop all containers
	stopped := 0
	for i, containerID := range containers {
		name := containerID
		if i < len(names) && names[i] != "" {
			name = names[i]
		}

		f.Progress(fmt.Sprintf("Stopping %s...", name))

		stopCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "stop", containerID)
		if err := stopCmd.Run(); err != nil {
			f.Warning(fmt.Sprintf("Failed to stop %s: %v", name, err))
			continue
		}

		stopped++
		slog.Info("container stopped", "name", name, "id", containerID)
	}

	f.NewLine()
	f.Success(fmt.Sprintf("Stopped %d workspace container(s)", stopped))
	return nil
}

func stopContainer(containerName string, f output.Formatter) error {
	f.Progress(fmt.Sprintf("Stopping workspace '%s'...", containerName))

	// Get Colima profile
	profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
	if profile == "" {
		profile = "default"
	}

	// Check if container is running
	checkCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	checkOutput, _ := checkCmd.Output()

	if len(checkOutput) == 0 {
		f.Info(fmt.Sprintf("Workspace '%s' is not running", containerName))
		return nil
	}

	// Stop the container
	ctx := context.Background()
	stopCmd := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "stop", containerName)

	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	slog.Info("workspace stopped", "name", containerName)
	f.Success(fmt.Sprintf("Workspace '%s' stopped", containerName))
	f.NewLine()
	f.Println("Re-attach with: dvm attach")

	return nil
}

// splitLines splits a string by newlines and filters empty strings
func splitLines(s string) []string {
	var result []string
	for _, line := range splitByNewline(s) {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func splitByNewline(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
