package operators

import (
	"context"
	"strings"
)

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
	AttachToWorkspace(ctx context.Context, opts AttachOptions) error

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

// AttachOptions contains options for attaching to a workspace
type AttachOptions struct {
	WorkspaceID string            // Container ID or name to attach to
	Env         map[string]string // Environment variables for the shell session
	Shell       string            // Shell to use (default: /bin/zsh)
	LoginShell  bool              // Use login shell (default: true)
}

// WorkspaceInfo contains information about a running workspace
type WorkspaceInfo struct {
	ID        string            // Container/pod ID
	Name      string            // Workspace name (container name)
	Status    string            // Running, Stopped, etc.
	Image     string            // Image name
	App       string            // App name from labels
	Workspace string            // Workspace name from labels
	Ecosystem string            // Ecosystem name from labels
	Domain    string            // Domain name from labels
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
	EcosystemName string            // Ecosystem name for hierarchical naming
	DomainName    string            // Domain name for hierarchical naming
	AppPath       string            // Host path to mount at /workspace
	WorkingDir    string            // Container working directory (default: /workspace)
	Command       []string          // Command to run (default: /bin/sleep infinity for keep-alive)
	Env           map[string]string // Environment variables
}

// ContainerNamingStrategy defines the interface for generating and parsing container names
type ContainerNamingStrategy interface {
	// GenerateName generates a container name from ecosystem, domain, app, and workspace
	GenerateName(ecosystem, domain, app, workspace string) string

	// ParseName parses a container name and returns its components
	// Returns ok=false if the name doesn't follow this strategy's format
	ParseName(containerName string) (ecosystem, domain, app, workspace string, ok bool)
}

// HierarchicalNamingStrategy implements a hierarchical container naming strategy
// Format: dvm-{ecosystem}-{domain}-{app}-{workspace} (all lowercase, dash-separated)
// If ecosystem/domain are empty, falls back to legacy dvm-{app}-{workspace} format
type HierarchicalNamingStrategy struct{}

// NewHierarchicalNamingStrategy creates a new hierarchical naming strategy
func NewHierarchicalNamingStrategy() ContainerNamingStrategy {
	return &HierarchicalNamingStrategy{}
}

// GenerateName generates a hierarchical container name
func (h *HierarchicalNamingStrategy) GenerateName(ecosystem, domain, app, workspace string) string {
	// Normalize all components to lowercase
	ecosystem = strings.ToLower(strings.TrimSpace(ecosystem))
	domain = strings.ToLower(strings.TrimSpace(domain))
	app = strings.ToLower(strings.TrimSpace(app))
	workspace = strings.ToLower(strings.TrimSpace(workspace))

	// Build name parts
	parts := []string{"dvm"}

	// Add ecosystem if present
	if ecosystem != "" {
		parts = append(parts, ecosystem)
	}

	// Add domain if present
	if domain != "" {
		parts = append(parts, domain)
	}

	// Always add app and workspace
	parts = append(parts, app, workspace)

	return strings.Join(parts, "-")
}

// ParseName parses a hierarchical container name into its components
func (h *HierarchicalNamingStrategy) ParseName(containerName string) (ecosystem, domain, app, workspace string, ok bool) {
	// Must start with "dvm-"
	if !strings.HasPrefix(containerName, "dvm-") {
		return "", "", "", "", false
	}

	// Split into parts
	parts := strings.Split(containerName, "-")
	if len(parts) < 3 { // At minimum: dvm-app-workspace
		return "", "", "", "", false
	}

	// Remove "dvm" prefix
	parts = parts[1:]

	switch len(parts) {
	case 2:
		// Legacy format: dvm-app-workspace
		return "", "", parts[0], parts[1], true
	case 3:
		// Single hierarchy: dvm-ecosystem-app-workspace or dvm-domain-app-workspace
		// We can't distinguish between ecosystem and domain without context
		// For parsing purposes, assume it's ecosystem
		return parts[0], "", parts[1], parts[2], true
	case 4:
		// Full hierarchy: dvm-ecosystem-domain-app-workspace
		return parts[0], parts[1], parts[2], parts[3], true
	default:
		// Too many parts, not a valid format
		return "", "", "", "", false
	}
}

// computeContainerName returns the container name to use.
// If ContainerName is set, use it. Otherwise, generate using hierarchical naming
// if ecosystem/domain are provided, or fall back to WorkspaceName.
func (opts StartOptions) ComputeContainerName() string {
	if opts.ContainerName != "" {
		return opts.ContainerName
	}

	// Use hierarchical naming if ecosystem or domain are provided
	if opts.EcosystemName != "" || opts.DomainName != "" {
		strategy := NewHierarchicalNamingStrategy()
		return strategy.GenerateName(opts.EcosystemName, opts.DomainName, opts.AppName, opts.WorkspaceName)
	}

	// Fall back to workspace name for backward compatibility
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
