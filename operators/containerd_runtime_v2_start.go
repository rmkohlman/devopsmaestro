package operators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"devopsmaestro/render"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// StartWorkspace creates and starts a workspace container
func (r *ContainerdRuntimeV2) StartWorkspace(ctx context.Context, opts StartOptions) (string, error) {
	// For Colima, use nerdctl via SSH to handle host path mounting correctly
	// The containerd API cannot handle macOS host paths directly since containerd
	// runs inside the Colima VM
	if r.platform.Type == PlatformColima {
		return r.startWorkspaceViaColima(ctx, opts)
	}

	// For other containerd platforms, use direct API (may need similar handling)
	return r.startWorkspaceDirectAPI(ctx, opts)
}

// startWorkspaceViaColima starts a workspace using nerdctl via colima ssh
// This handles the host path mounting correctly through Colima's mount system
func (r *ContainerdRuntimeV2) startWorkspaceViaColima(ctx context.Context, opts StartOptions) (string, error) {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// First check if container already exists and what image it's using
	containerName := opts.ComputeContainerName()

	// Check if container exists by trying to inspect it
	// We check for existence separately from the image label because
	// containers created before v0.18.18 don't have the image label
	existsCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect %s >/dev/null 2>&1 && echo EXISTS || echo NOTFOUND",
		r.namespace, containerName)
	existsExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", existsCmd)
	existsOutput, _ := existsExec.Output()
	containerExists := strings.TrimSpace(string(existsOutput)) == "EXISTS"

	if containerExists {
		// Get the image the existing container was created with from our label
		// Containers created before v0.18.18 won't have this label
		imageCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{index .Config.Labels \"io.devopsmaestro.image\"}}' %s 2>/dev/null || echo ''",
			r.namespace, containerName)
		imageExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", imageCmd)
		imageOutput, _ := imageExec.Output()
		existingImage := strings.TrimSpace(string(imageOutput))

		// If no label (pre-v0.18.18 container) or image changed, recreate the container
		hasImageLabel := existingImage != "" && existingImage != "''" && existingImage != "<no value>"
		imageChanged := hasImageLabel && existingImage != opts.ImageName

		if !hasImageLabel {
			render.Infof("Container exists without image label (pre-v0.18.18), recreating...")
			// Remove old container to recreate with proper labels
			rmCmd := fmt.Sprintf("sudo nerdctl --namespace %s rm -f %s 2>/dev/null || true",
				r.namespace, containerName)
			rmExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", rmCmd)
			rmExec.Run()
			// Fall through to create new container
		} else if imageChanged {
			render.Infof("Image changed: %s -> %s", existingImage, opts.ImageName)
			render.Info("Recreating container with new image...")

			// Stop and remove the old container regardless of state
			rmCmd := fmt.Sprintf("sudo nerdctl --namespace %s rm -f %s 2>/dev/null || true",
				r.namespace, containerName)
			rmExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", rmCmd)
			rmExec.Run()

			// Fall through to create new container
		} else {
			// Same image - check if it's running
			statusCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{.State.Status}}' %s 2>/dev/null || echo stopped",
				r.namespace, containerName)
			statusExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", statusCmd)
			statusOutput, _ := statusExec.Output()
			status := strings.TrimSpace(string(statusOutput))

			if status == "running" {
				// Already running with same image, return the container name as ID
				return containerName, nil
			}

			// Container exists but not running, try to start it
			startCmd := fmt.Sprintf("sudo nerdctl --namespace %s start %s 2>/dev/null",
				r.namespace, containerName)
			startExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", startCmd)
			if err := startExec.Run(); err == nil {
				return containerName, nil
			}

			// Start failed, remove and recreate
			rmCmd := fmt.Sprintf("sudo nerdctl --namespace %s rm -f %s 2>/dev/null || true",
				r.namespace, containerName)
			rmExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", rmCmd)
			rmExec.Run()
		}
	}

	// Build nerdctl run command
	// Set command to keep container running using helper
	command := opts.ComputeCommand()

	// Set default working directory
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	// Build the nerdctl run command parts
	nerdctlArgs := []string{
		"sudo", "nerdctl",
		"--namespace", r.namespace,
		"run", "-d",
		"--name", containerName,
		"-w", workingDir,
	}

	// Add volume mounts
	// nerdctl handles path translation for Colima's mounted volumes

	// Legacy mount for AppPath (if not using WorkspaceSlug)
	if opts.AppPath != "" {
		nerdctlArgs = append(nerdctlArgs, "-v", fmt.Sprintf("%s:/workspace", opts.AppPath))
	}

	// v0.19.0: Add workspace volume mounts from Mounts field
	for _, mount := range opts.Mounts {
		mountSpec := fmt.Sprintf("%s:%s", mount.Source, mount.Destination)
		if mount.ReadOnly {
			mountSpec += ":ro"
		}
		nerdctlArgs = append(nerdctlArgs, "-v", mountSpec)
	}

	// SSH agent forwarding (opt-in only)
	// SECURITY: SSH keys are NEVER mounted. Only the agent socket is forwarded.
	if opts.SSHAgentForwarding {
		hostSocket, containerSocket, err := GetSSHAgentMount(r.GetRuntimeType())
		if err != nil {
			return "", fmt.Errorf("SSH agent forwarding requested but not available: %w", err)
		}
		nerdctlArgs = append(nerdctlArgs, "-v", fmt.Sprintf("%s:%s", hostSocket, containerSocket))

		// Add SSH_AUTH_SOCK environment variable
		nerdctlArgs = append(nerdctlArgs, "-e", fmt.Sprintf("SSH_AUTH_SOCK=%s", containerSocket))
	}

	// Add environment variables
	for k, v := range opts.Env {
		nerdctlArgs = append(nerdctlArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add labels for DVM management
	nerdctlArgs = append(nerdctlArgs,
		"--label", "io.devopsmaestro.managed=true",
		"--label", fmt.Sprintf("io.devopsmaestro.app=%s", opts.AppName),
		"--label", fmt.Sprintf("io.devopsmaestro.workspace=%s", opts.WorkspaceName),
		"--label", fmt.Sprintf("io.devopsmaestro.image=%s", opts.ImageName),
	)

	// Add ecosystem label if provided
	if opts.EcosystemName != "" {
		nerdctlArgs = append(nerdctlArgs, "--label", fmt.Sprintf("io.devopsmaestro.ecosystem=%s", opts.EcosystemName))
	}

	// Add domain label if provided
	if opts.DomainName != "" {
		nerdctlArgs = append(nerdctlArgs, "--label", fmt.Sprintf("io.devopsmaestro.domain=%s", opts.DomainName))
	}

	// Add image and command
	nerdctlArgs = append(nerdctlArgs, opts.ImageName)
	nerdctlArgs = append(nerdctlArgs, command...)

	// Build the full command string for SSH
	cmdStr := ""
	for i, arg := range nerdctlArgs {
		if i > 0 {
			cmdStr += " "
		}
		// Quote arguments that contain spaces or special characters
		if needsQuoting(arg) {
			cmdStr += fmt.Sprintf("'%s'", arg)
		} else {
			cmdStr += arg
		}
	}

	// Execute via colima ssh
	execCmd := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", cmdStr)
	execCmd.Stderr = os.Stderr

	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to start workspace via nerdctl: %w", err)
	}

	// nerdctl run -d returns the container ID
	containerID := string(output)
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	// Return the container name as the ID (that's what we use to reference it)
	return containerName, nil
}

// needsQuoting returns true if a string needs shell quoting
func needsQuoting(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\'' || c == '"' || c == '=' || c == '$' || c == '\\' || c == '!' || c == '*' {
			return true
		}
	}
	return false
}

// startWorkspaceDirectAPI starts a workspace using the containerd API directly
// This is the original implementation, kept for non-Colima platforms
func (r *ContainerdRuntimeV2) startWorkspaceDirectAPI(ctx context.Context, opts StartOptions) (string, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	// Get the image
	image, err := r.client.GetImage(ctx, opts.ImageName)
	if err != nil {
		return "", fmt.Errorf("image not found: %s\nRun 'dvm build' first", opts.ImageName)
	}

	// Check if container already exists
	containerName := opts.ComputeContainerName()
	existingContainer, err := r.client.LoadContainer(ctx, containerName)
	if err == nil {
		// Container exists - clean it up
		task, err := existingContainer.Task(ctx, nil)
		if err == nil {
			// Task exists - kill and delete it
			status, err := task.Status(ctx)
			if err == nil && status.Status == "running" {
				// Container is already running - return its ID
				return existingContainer.ID(), nil
			}

			// Task exists but not running, try to kill it first
			task.Kill(ctx, syscall.SIGKILL)
			task.Wait(ctx) // Wait for it to die
			task.Delete(ctx, client.WithProcessKill)
		}
		// Delete the existing container with snapshot cleanup
		existingContainer.Delete(ctx, client.WithSnapshotCleanup)
	}

	// Set command to keep container running using helper
	command := opts.ComputeCommand()

	// Set default working directory
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	// Convert env map to slice
	envSlice := make([]string, 0, len(opts.Env))
	for k, v := range opts.Env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// Create mount specs
	mounts := []specs.Mount{}

	// Legacy mount for AppPath (if not using WorkspaceSlug)
	if opts.AppPath != "" {
		mounts = append(mounts, specs.Mount{
			Source:      opts.AppPath,
			Destination: "/workspace",
			Type:        "bind",
			Options:     []string{"rbind", "rw"},
		})
	}

	// v0.19.0: Add workspace volume mounts from Mounts field
	for _, mount := range opts.Mounts {
		mountOptions := []string{"rbind"}
		if mount.ReadOnly {
			mountOptions = append(mountOptions, "ro")
		} else {
			mountOptions = append(mountOptions, "rw")
		}

		mounts = append(mounts, specs.Mount{
			Source:      mount.Source,
			Destination: mount.Destination,
			Type:        "bind",
			Options:     mountOptions,
		})
	}

	// SSH agent forwarding (opt-in only)
	// SECURITY: SSH keys are NEVER mounted. Only the agent socket is forwarded.
	if opts.SSHAgentForwarding {
		hostSocket, containerSocket, err := GetSSHAgentMount(r.GetRuntimeType())
		if err != nil {
			return "", fmt.Errorf("SSH agent forwarding requested but not available: %w", err)
		}
		mounts = append(mounts, specs.Mount{
			Source:      hostSocket,
			Destination: containerSocket,
			Type:        "bind",
			Options:     []string{"rbind", "rw"},
		})

		// Add SSH_AUTH_SOCK environment variable
		envSlice = append(envSlice, fmt.Sprintf("SSH_AUTH_SOCK=%s", containerSocket))
	}

	// Create container with proper OCI spec
	// Use Compose to combine base spec with customizations
	container, err := r.client.NewContainer(
		ctx,
		containerName,
		client.WithImage(image),
		client.WithSnapshotter("overlayfs"),
		client.WithNewSnapshot(containerName+"-snapshot", image),
		client.WithRuntime("io.containerd.runc.v2", nil), // Explicitly use runc runtime
		client.WithNewSpec(
			oci.Compose(
				oci.WithDefaultSpec(),
				oci.WithDefaultUnixDevices,
				oci.WithImageConfig(image),
				oci.WithProcessArgs(command...),
				oci.WithProcessCwd(workingDir),
				oci.WithEnv(envSlice),
				oci.WithMounts(mounts),
				oci.WithUserID(0), // Run as root
				oci.WithUsername("root"),
				// Don't set TTY here - nerdctl exec will handle TTY allocation
			),
		),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Create the task (starts the container)
	// Use NullIO for now - we'll attach with proper IO later
	task, err := container.NewTask(ctx, cio.NullIO)
	if err != nil {
		container.Delete(ctx, client.WithSnapshotCleanup)
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Start the task
	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		container.Delete(ctx, client.WithSnapshotCleanup)
		return "", fmt.Errorf("failed to start task: %w", err)
	}

	return container.ID(), nil
}
