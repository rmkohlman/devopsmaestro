package operators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/moby/term"
)

// GetWorkspaceStatus returns the status of a workspace
func (r *ContainerdRuntimeV2) GetWorkspaceStatus(ctx context.Context, containerID string) (string, error) {
	// For Colima, use nerdctl via SSH (containerd API doesn't work well across VM boundary)
	if r.platform.Type == PlatformColima {
		return r.getStatusViaColima(ctx, containerID)
	}

	// For other platforms, use containerd API directly
	return r.getStatusDirectAPI(ctx, containerID)
}

// getStatusViaColima gets container status using nerdctl via SSH for Colima
func (r *ContainerdRuntimeV2) getStatusViaColima(ctx context.Context, containerID string) (string, error) {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// Check if container exists and get its status via nerdctl
	statusCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{.State.Status}}' %s 2>/dev/null || echo not_found",
		r.namespace, containerID)
	statusExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", statusCmd)
	statusOutput, err := statusExec.Output()
	if err != nil {
		return "unknown", fmt.Errorf("failed to check container status: %w", err)
	}

	status := strings.TrimSpace(string(statusOutput))
	if status == "not_found" {
		return "not_found", nil
	}

	return status, nil
}

// getStatusDirectAPI gets container status using containerd API directly
func (r *ContainerdRuntimeV2) getStatusDirectAPI(ctx context.Context, containerID string) (string, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	container, err := r.client.LoadContainer(ctx, containerID)
	if err != nil {
		return "not_found", nil
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "created", nil
	}

	status, err := task.Status(ctx)
	if err != nil {
		return "unknown", err
	}

	return string(status.Status), nil
}

// ListWorkspaces lists all DVM-managed workspaces
func (r *ContainerdRuntimeV2) ListWorkspaces(ctx context.Context) ([]WorkspaceInfo, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	containers, err := r.client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var workspaces []WorkspaceInfo
	for _, c := range containers {
		labels, err := c.Labels(ctx)
		if err != nil {
			continue
		}

		// Check for DVM management label
		if labels["io.devopsmaestro.managed"] != "true" {
			continue
		}

		// Get status
		status := "created"
		task, err := c.Task(ctx, nil)
		if err == nil {
			taskStatus, err := task.Status(ctx)
			if err == nil {
				status = string(taskStatus.Status)
			}
		}

		// Get image
		image, _ := c.Image(ctx)
		imageName := ""
		if image != nil {
			imageName = image.Name()
		}

		workspaces = append(workspaces, WorkspaceInfo{
			ID:        c.ID()[:12],
			Name:      c.ID(), // containerd uses ID as name
			Status:    status,
			Image:     imageName,
			App:       labels["io.devopsmaestro.app"],
			Workspace: labels["io.devopsmaestro.workspace"],
			Ecosystem: labels["io.devopsmaestro.ecosystem"],
			Domain:    labels["io.devopsmaestro.domain"],
			Labels:    labels,
		})
	}

	return workspaces, nil
}

// FindWorkspace finds a workspace by name and returns its info
func (r *ContainerdRuntimeV2) FindWorkspace(ctx context.Context, name string) (*WorkspaceInfo, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	container, err := r.client.LoadContainer(ctx, name)
	if err != nil {
		return nil, nil // Not found
	}

	labels, err := container.Labels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	// Verify it's DVM-managed
	if labels["io.devopsmaestro.managed"] != "true" {
		return nil, nil
	}

	// Get status
	status := "created"
	task, err := container.Task(ctx, nil)
	if err == nil {
		taskStatus, err := task.Status(ctx)
		if err == nil {
			status = string(taskStatus.Status)
		}
	}

	// Get image
	image, _ := container.Image(ctx)
	imageName := ""
	if image != nil {
		imageName = image.Name()
	}

	return &WorkspaceInfo{
		ID:        container.ID()[:12],
		Name:      container.ID(),
		Status:    status,
		Image:     imageName,
		App:       labels["io.devopsmaestro.app"],
		Workspace: labels["io.devopsmaestro.workspace"],
		Ecosystem: labels["io.devopsmaestro.ecosystem"],
		Domain:    labels["io.devopsmaestro.domain"],
		Labels:    labels,
	}, nil
}

// StopAllWorkspaces stops all DVM-managed workspaces
func (r *ContainerdRuntimeV2) StopAllWorkspaces(ctx context.Context) (int, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	workspaces, err := r.ListWorkspaces(ctx)
	if err != nil {
		return 0, err
	}

	stopped := 0
	for _, ws := range workspaces {
		if ws.Status == "running" {
			if err := r.StopWorkspace(ctx, ws.Name); err == nil {
				stopped++
			}
		}
	}

	return stopped, nil
}

// setupTerminal sets up the terminal for raw mode
func setupTerminal() error {
	fd := os.Stdin.Fd()
	state, err := term.SetRawTerminal(fd)
	if err != nil {
		return err
	}
	// Store state for restoration
	terminalState = state
	return nil
}

// restoreTerminal restores the terminal to its original state
func restoreTerminal() {
	if terminalState != nil {
		fd := os.Stdin.Fd()
		term.RestoreTerminal(fd, terminalState)
		terminalState = nil
	}
}

var terminalState *term.State
