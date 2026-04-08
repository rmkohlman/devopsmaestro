package cmd

import (
	"context"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"fmt"
	"github.com/rmkohlman/MaestroSDK/render"
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

func init() {
	rootCmd.AddCommand(statusCmd)
	AddOutputFlag(statusCmd, "")
}

// StatusInfo holds all status information
type StatusInfo struct {
	Context            ContextInfo         `json:"context" yaml:"context"`
	Containers         []ContainerInfo     `json:"containers" yaml:"containers"`
	Runtime            RuntimeInfo         `json:"runtime" yaml:"runtime"`
	CredentialWarnings []CredentialWarning `json:"credential_warnings,omitempty" yaml:"credential_warnings,omitempty"`
}

// CredentialWarning holds a credential expiration warning
type CredentialWarning struct {
	Name      string `json:"name" yaml:"name"`
	Scope     string `json:"scope" yaml:"scope"`
	Status    string `json:"status" yaml:"status"`
	ExpiresAt string `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
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
	outputFormat, _ := cmd.Flags().GetString("output")

	status := StatusInfo{
		Containers: []ContainerInfo{},
	}

	// Get context (DB-backed)
	ds, err := getDataStore(cmd)
	if err != nil {
		slog.Debug("failed to get data store", "error", err)
	} else {
		app, _ := getActiveAppFromContext(ds)
		workspace, _ := getActiveWorkspaceFromContext(ds)
		status.Context = ContextInfo{
			App:       app,
			Workspace: workspace,
		}

		// Check for expired/expiring credentials
		status.CredentialWarnings = collectCredentialWarnings(ds)
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
		if outputFormat == "json" || outputFormat == "yaml" {
			return render.OutputWith(outputFormat, status, render.Options{})
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
	if outputFormat == "json" || outputFormat == "yaml" {
		return render.OutputWith(outputFormat, status, render.Options{})
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

	render.Blank()

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

	render.Blank()

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

	// Credential warnings section
	renderCredentialWarnings(status.CredentialWarnings)
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// collectCredentialWarnings checks all credentials for expiration issues.
// Returns warnings for any credentials that are expired or expiring soon.
func collectCredentialWarnings(ds interface {
	ListAllCredentials() ([]*models.CredentialDB, error)
}) []CredentialWarning {
	creds, err := ds.ListAllCredentials()
	if err != nil {
		slog.Debug("failed to list credentials for expiration check", "error", err)
		return nil
	}

	var warnings []CredentialWarning
	for _, c := range creds {
		status := c.ExpirationStatus()
		if status == "expired" || status == "expiring soon" {
			w := CredentialWarning{
				Name:   c.Name,
				Scope:  fmt.Sprintf("%s (ID: %d)", c.ScopeType, c.ScopeID),
				Status: status,
			}
			if c.ExpiresAt != nil {
				w.ExpiresAt = c.ExpiresAt.Format("2006-01-02 15:04:05")
			}
			warnings = append(warnings, w)
		}
	}
	return warnings
}

// renderCredentialWarnings displays credential expiration warnings in the status output.
func renderCredentialWarnings(warnings []CredentialWarning) {
	if len(warnings) == 0 {
		return
	}

	render.Blank()
	render.Warning("Credential Warnings")
	for _, w := range warnings {
		switch w.Status {
		case "expired":
			render.Warning(fmt.Sprintf("  ⚠ %s — EXPIRED (was %s) [%s]", w.Name, w.ExpiresAt, w.Scope))
		case "expiring soon":
			render.Warning(fmt.Sprintf("  ⏰ %s — expiring soon (%s) [%s]", w.Name, w.ExpiresAt, w.Scope))
		}
	}
}
