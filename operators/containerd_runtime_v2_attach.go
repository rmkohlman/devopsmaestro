package operators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/containerd/containerd/v2/pkg/namespaces"
)

// AttachToWorkspace attaches to a running workspace container
// For Colima, we use nerdctl via SSH since direct FIFO-based IO doesn't work across host/VM boundary
func (r *ContainerdRuntimeV2) AttachToWorkspace(ctx context.Context, opts AttachOptions) error {
	// For Colima, use nerdctl via SSH (containerd API doesn't work well across VM boundary)
	if r.platform.Type == PlatformColima {
		return r.attachViaColima(ctx, opts)
	}

	// For other platforms, use containerd API directly
	return r.attachDirectAPI(ctx, opts)
}

// attachViaColima attaches to a container using nerdctl via SSH for Colima
func (r *ContainerdRuntimeV2) attachViaColima(ctx context.Context, opts AttachOptions) error {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// Check if container exists and is running via nerdctl
	statusCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{.State.Status}}' %s 2>/dev/null || echo not_found",
		shellEscape(r.namespace), shellEscape(opts.WorkspaceID))
	statusExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", statusCmd)
	statusOutput, err := statusExec.Output()
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	status := strings.TrimSpace(string(statusOutput))
	if status == "not_found" {
		return fmt.Errorf("container not found: %s", opts.WorkspaceID)
	}

	if status != "running" {
		return fmt.Errorf("container is not running (status: %s)", status)
	}

	// Build shell command with options
	shell := opts.Shell
	if shell == "" {
		shell = "/bin/zsh"
	}

	// Compute effective UID/GID for exec session (default to 1000 if not set)
	uid := opts.UID
	if uid == 0 {
		uid = 1000
	}
	gid := opts.GID
	if gid == 0 {
		gid = 1000
	}
	userStr := fmt.Sprintf("%d:%d", uid, gid)

	// Start building the nerdctl exec command
	var cmdParts []string
	cmdParts = append(cmdParts, "sudo", "nerdctl", "--namespace", shellEscape(r.namespace), "exec", "-it")

	// Defense-in-depth: explicitly set user for exec sessions
	cmdParts = append(cmdParts, "--user", userStr)

	// Add environment variables
	for key, value := range opts.Env {
		// Shell-escape the value to prevent shell metacharacter interpretation
		cmdParts = append(cmdParts, "-e", fmt.Sprintf("%s=%s", key, shellEscape(value)))
	}

	// Add container name
	cmdParts = append(cmdParts, shellEscape(opts.WorkspaceID))

	// Add shell and login flag
	cmdParts = append(cmdParts, shellEscape(shell))
	if opts.LoginShell {
		cmdParts = append(cmdParts, "-l")
	}

	// Convert to command string for SSH execution
	cmd := strings.Join(cmdParts, " ")
	execCmd := []string{"colima", "--profile", profile, "ssh", "--", "sh", "-c", cmd}

	// Find colima in PATH
	colimaPath, err := exec.LookPath("colima")
	if err != nil {
		return fmt.Errorf("colima not found in PATH: %w", err)
	}

	// Create exec command
	execProc := &exec.Cmd{
		Path:   colimaPath,
		Args:   execCmd,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// Run the command
	if err := execProc.Run(); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

	return nil
}

// attachDirectAPI attaches to a container using containerd API directly
func (r *ContainerdRuntimeV2) attachDirectAPI(ctx context.Context, opts AttachOptions) error {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	// Load the container
	container, err := r.client.LoadContainer(ctx, opts.WorkspaceID)
	if err != nil {
		return fmt.Errorf("container not found: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task is running
	status, err := task.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get task status: %w", err)
	}

	if status.Status != "running" {
		return fmt.Errorf("container is not running (status: %s)", status.Status)
	}

	// For other platforms, return not implemented for now
	return fmt.Errorf("attach not implemented for platform %s with containerd runtime", r.platform.Name)
}
