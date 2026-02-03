package cmd

import (
	"devopsmaestro/operators"
	"devopsmaestro/render"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// statusCmd shows the current status of DVM
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of DVM and running workspaces",
	Long: `Show the current status of DVM including:

- Active project and workspace context
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
	Project   string `json:"project,omitempty" yaml:"project,omitempty"`
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
	Socket string `json:"socket,omitempty" yaml:"socket,omitempty"`
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
		project, _ := contextMgr.GetActiveProject()
		workspace, _ := contextMgr.GetActiveWorkspace()
		status.Context = ContextInfo{
			Project:   project,
			Workspace: workspace,
		}
	}

	// Detect container platform using the platform detector
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		slog.Debug("failed to create platform detector", "error", err)
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

	platform, err := detector.Detect()
	if err != nil {
		slog.Debug("failed to detect platform", "error", err)
		status.Runtime = RuntimeInfo{
			Type:   "unknown",
			Status: "not found",
		}
	} else {
		status.Runtime = RuntimeInfo{
			Type:   string(platform.Type),
			Name:   platform.Name,
			Socket: platform.SocketPath,
			Status: "active",
		}

		// Get running dvm containers using docker CLI (works for OrbStack, Docker Desktop, Podman)
		listCmd := exec.Command("docker", "-H", "unix://"+platform.SocketPath,
			"ps", "--format", "{{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}",
			"--filter", "label=io.devopsmaestro.managed=true")
		output, err := listCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.Split(line, "\t")
				if len(parts) >= 3 {
					container := ContainerInfo{
						ID:     parts[0],
						Name:   parts[1],
						Status: parts[2],
					}
					if len(parts) >= 4 {
						container.Image = parts[3]
					}
					status.Containers = append(status.Containers, container)
				}
			}
		} else {
			slog.Debug("failed to list containers", "error", err)
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

func renderStatusColored(status StatusInfo) {
	// Context section
	render.Info("Context")
	if status.Context.Project != "" {
		render.Success(fmt.Sprintf("  Project:   %s", status.Context.Project))
	} else {
		render.Warning("  Project:   (none)")
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
