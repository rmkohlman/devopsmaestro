package operators

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/containerd/containerd/v2/pkg/namespaces"
)

// StopWorkspace stops a running workspace
func (r *ContainerdRuntimeV2) StopWorkspace(ctx context.Context, containerID string) error {
	// For Colima, use nerdctl via SSH (containerd API doesn't work well across VM boundary)
	if r.platform.Type == PlatformColima {
		return r.stopViaColima(ctx, containerID)
	}

	// For other platforms, use containerd API directly
	return r.stopDirectAPI(ctx, containerID)
}

// stopViaColima stops a container using nerdctl via SSH for Colima
func (r *ContainerdRuntimeV2) stopViaColima(ctx context.Context, containerID string) error {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// Stop the container using nerdctl
	stopCmd := fmt.Sprintf("sudo nerdctl --namespace %s stop %s 2>/dev/null || true", r.namespace, containerID)
	stopExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", stopCmd)
	if err := stopExec.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

// stopDirectAPI stops a container using containerd API directly
func (r *ContainerdRuntimeV2) stopDirectAPI(ctx context.Context, containerID string) error {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	container, err := r.client.LoadContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("container not found: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		// No task means container is not running
		return nil
	}

	// Kill the task
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill task: %w", err)
	}

	// Wait for task to exit
	statusC, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}

	<-statusC

	// Delete the task
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}
