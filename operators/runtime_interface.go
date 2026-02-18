package operators

import "context"

// ContainerRuntime defines the interface for container runtime operations
// This abstraction allows DevOpsMaestro to work with Docker, Kubernetes, or any other runtime
type ContainerRuntime interface {
	// BuildImage builds a container image from the app
	BuildImage(ctx context.Context, opts BuildOptions) error

	// StartWorkspace creates and starts a workspace container.
	//
	// Contracts:
	// - If ContainerName is empty, uses WorkspaceName
	// - If Command is empty, uses "/bin/sleep infinity" to keep container alive
	// - Attach to the running container with AttachToWorkspace
	// - Returns container ID (or container name for some runtimes)
	StartWorkspace(ctx context.Context, opts StartOptions) (string, error)

	// AttachToWorkspace attaches an interactive terminal to a running workspace
	AttachToWorkspace(ctx context.Context, workspaceID string) error

	// StopWorkspace stops a running workspace
	StopWorkspace(ctx context.Context, workspaceID string) error

	// GetWorkspaceStatus returns the current status of a workspace
	GetWorkspaceStatus(ctx context.Context, workspaceID string) (string, error)

	// GetRuntimeType returns the runtime type (docker, kubernetes, etc.)
	GetRuntimeType() string

	// ListWorkspaces lists all DVM-managed workspaces
	ListWorkspaces(ctx context.Context) ([]WorkspaceInfo, error)

	// FindWorkspace finds a workspace by name and returns its info
	FindWorkspace(ctx context.Context, name string) (*WorkspaceInfo, error)

	// GetPlatformName returns the human-readable platform name
	GetPlatformName() string

	// StopAllWorkspaces stops all DVM-managed workspaces
	StopAllWorkspaces(ctx context.Context) (int, error)
}

// WorkspaceInfo contains information about a running workspace
type WorkspaceInfo struct {
	ID        string            // Container/pod ID
	Name      string            // Workspace name (container name)
	Status    string            // Running, Stopped, etc.
	Image     string            // Image name
	App       string            // App name from labels
	Workspace string            // Workspace name from labels
	Labels    map[string]string // All labels
}

// BuildOptions contains options for building container images
type BuildOptions struct {
	AppPath      string            // Path to the app on the host
	AppName      string            // Name of the app
	ImageName    string            // Name of the image to build
	Dockerfile   string            // Path to Dockerfile
	BuildContext string            // Path to build context
	Tags         []string          // Additional tags for the image
	BuildArgs    map[string]string // Build arguments
}

// StartOptions contains options for starting a workspace
type StartOptions struct {
	ImageName     string            // Image to run
	WorkspaceName string            // Logical workspace name (used in labels)
	ContainerName string            // Physical container name (if empty, uses WorkspaceName)
	AppName       string            // App name for labeling
	AppPath       string            // Host path to mount at /workspace
	WorkingDir    string            // Container working directory (default: /workspace)
	Command       []string          // Command to run (default: /bin/sleep infinity for keep-alive)
	Env           map[string]string // Environment variables
}

// computeContainerName returns the container name to use.
// If ContainerName is set, use it. Otherwise fall back to WorkspaceName.
func (opts StartOptions) ComputeContainerName() string {
	if opts.ContainerName != "" {
		return opts.ContainerName
	}
	return opts.WorkspaceName
}

// DefaultKeepAliveCommand returns the standard command to keep containers running.
// Use this when Command is empty in StartOptions.
func DefaultKeepAliveCommand() []string {
	return []string{"/bin/sleep", "infinity"}
}

// ComputeCommand returns the command to use for starting a container.
// If Command is set, use it. Otherwise use the default keep-alive command.
func (opts StartOptions) ComputeCommand() []string {
	if len(opts.Command) > 0 {
		return opts.Command
	}
	return DefaultKeepAliveCommand()
}
