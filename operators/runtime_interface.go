package operators

import "context"

// ContainerRuntime defines the interface for container runtime operations
// This abstraction allows DevOpsMaestro to work with Docker, Kubernetes, or any other runtime
type ContainerRuntime interface {
	// BuildImage builds a container image from the project
	BuildImage(ctx context.Context, opts BuildOptions) error

	// StartWorkspace starts a workspace container/pod
	StartWorkspace(ctx context.Context, opts StartOptions) (string, error)

	// AttachToWorkspace attaches an interactive terminal to a running workspace
	AttachToWorkspace(ctx context.Context, workspaceID string) error

	// StopWorkspace stops a running workspace
	StopWorkspace(ctx context.Context, workspaceID string) error

	// GetWorkspaceStatus returns the current status of a workspace
	GetWorkspaceStatus(ctx context.Context, workspaceID string) (string, error)

	// GetRuntimeType returns the runtime type (docker, kubernetes, etc.)
	GetRuntimeType() string
}

// BuildOptions contains options for building container images
type BuildOptions struct {
	ProjectPath  string            // Path to the project on the host
	ProjectName  string            // Name of the project
	ImageName    string            // Name of the image to build
	Dockerfile   string            // Path to Dockerfile
	BuildContext string            // Path to build context
	Tags         []string          // Additional tags for the image
	BuildArgs    map[string]string // Build arguments
}

// StartOptions contains options for starting a workspace
type StartOptions struct {
	ImageName     string            // Container image to use
	WorkspaceName string            // Name of the workspace
	ProjectPath   string            // Path to mount as /workspace
	Env           map[string]string // Environment variables
	WorkingDir    string            // Working directory inside container
	Command       []string          // Command to run (default: /bin/zsh)
}
