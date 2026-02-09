package cmd

import (
	"context"
	"devopsmaestro/operators"
	"devopsmaestro/render"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

// statusCmd shows the current status of DVM
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of DVM and running workspaces",
	Long: `Show the current status of DVM including:

- Active app and workspace context
- Running workspace containers
- Container runtime information

Examples:
  dvm status              # Show full status
  dvm status -o json      # Output as JSON
  dvm status -o yaml      # Output as YAML`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd)
	},
}

var statusOutputFormat string

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVarP(&statusOutputFormat, "output", "o", "", "Output format (json, yaml)")
}

// StatusInfo holds all status information
type StatusInfo struct {
	Context    ContextInfo     `json:"context" yaml:"context"`
	Containers []ContainerInfo `json:"containers" yaml:"containers"`
	Runtime    RuntimeInfo     `json:"runtime" yaml:"runtime"`
}

// ContextInfo holds the current context
type ContextInfo struct {
	App       string `json:"app,omitempty" yaml:"app,omitempty"`
	Workspace string `json:"workspace,omitempty" yaml:"workspace,omitempty"`
}

// ContainerInfo holds container details
type ContainerInfo struct {
	Name   string `json:"name" yaml:"name"`
	ID     string `json:"id" yaml:"id"`
	Status string `json:"status" yaml:"status"`
	Image  string `json:"image,omitempty" yaml:"image,omitempty"`
}

// RuntimeInfo holds runtime information
type RuntimeInfo struct {
	Type   string `json:"type" yaml:"type"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Status string `json:"status" yaml:"status"`
}

func runStatus(cmd *cobra.Command) error {
	status := StatusInfo{
		Containers: []ContainerInfo{},
	}

	// Get context
	contextMgr, err := operators.NewContextManager()
	if err != nil {
		slog.Debug("failed to get context manager", "error", err)
	} else {
		app, _ := contextMgr.GetActiveApp()
		workspace, _ := contextMgr.GetActiveWorkspace()
		status.Context = ContextInfo{
			App:       app,
			Workspace: workspace,
		}
	}

	// Create container runtime using factory
	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		slog.Debug("failed to create runtime", "error", err)
		status.Runtime = RuntimeInfo{
			Type:   "unknown",
			Status: "not found",
		}
		// Handle output format
		if statusOutputFormat == "json" || statusOutputFormat == "yaml" {
			return render.OutputWith(statusOutputFormat, status, render.Options{})
		}
		renderStatusColored(status)
		return nil
	}

	// Get runtime info
	status.Runtime = RuntimeInfo{
		Type:   runtime.GetRuntimeType(),
		Name:   runtime.GetPlatformName(),
		Status: "active",
	}

	// List running workspaces using the runtime interface
	workspaces, err := runtime.ListWorkspaces(context.Background())
	if err != nil {
		slog.Debug("failed to list workspaces", "error", err)
	} else {
		for _, ws := range workspaces {
			// Only show running containers
			if isRunning(ws.Status) {
				status.Containers = append(status.Containers, ContainerInfo{
					ID:     ws.ID,
					Name:   ws.Name,
					Status: ws.Status,
					Image:  ws.Image,
				})
			}
		}
	}

	// Handle output format
	if statusOutputFormat == "json" || statusOutputFormat == "yaml" {
		return render.OutputWith(statusOutputFormat, status, render.Options{})
	}

	// Default: colored output
	renderStatusColored(status)
	return nil
}

// isRunning checks if the status indicates a running container
func isRunning(status string) bool {
	return status == "running" || (len(status) >= 2 && status[:2] == "Up")
}

func renderStatusColored(status StatusInfo) {
	// Context section
	render.Info("Context")
	if status.Context.App != "" {
		render.Success(fmt.Sprintf("  App:       %s", status.Context.App))
	} else {
		render.Warning("  App:       (none)")
	}
	if status.Context.Workspace != "" {
		render.Success(fmt.Sprintf("  Workspace: %s", status.Context.Workspace))
	} else {
		render.Warning("  Workspace: (none)")
	}

	fmt.Println()

	// Runtime section
	render.Info("Runtime")
	if status.Runtime.Name != "" {
		render.Info(fmt.Sprintf("  Platform: %s", status.Runtime.Name))
	} else {
		render.Info(fmt.Sprintf("  Type:     %s", status.Runtime.Type))
	}
	if status.Runtime.Status == "active" {
		render.Success(fmt.Sprintf("  Status:   %s", status.Runtime.Status))
	} else {
		render.Warning(fmt.Sprintf("  Status:   %s", status.Runtime.Status))
	}

	fmt.Println()

	// Containers section
	render.Info("Running Workspaces")
	if len(status.Containers) == 0 {
		render.Info("  (none)")
	} else {
		for _, c := range status.Containers {
			render.Success(fmt.Sprintf("  %s", c.Name))
			render.Info(fmt.Sprintf("    ID:     %s", truncateID(c.ID)))
			render.Info(fmt.Sprintf("    Status: %s", c.Status))
			if c.Image != "" {
				render.Info(fmt.Sprintf("    Image:  %s", c.Image))
			}
		}
	}
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
