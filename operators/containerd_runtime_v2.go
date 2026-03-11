package operators

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"
)

// ContainerdRuntimeV2 is a clean implementation using containerd v2 API
type ContainerdRuntimeV2 struct {
	client    *client.Client
	platform  *Platform
	namespace string
}

// NewContainerdRuntimeV2 creates a new containerd v2 runtime instance
func NewContainerdRuntimeV2() (*ContainerdRuntimeV2, error) {
	// Detect platform
	detector, err := NewPlatformDetector()
	if err != nil {
		return nil, err
	}

	platform, err := detector.Detect()
	if err != nil {
		return nil, err
	}

	return NewContainerdRuntimeV2WithPlatform(platform)
}

// NewContainerdRuntimeV2WithPlatform creates a new containerd v2 runtime with a specific platform
func NewContainerdRuntimeV2WithPlatform(platform *Platform) (*ContainerdRuntimeV2, error) {
	// Verify this platform supports containerd
	if !platform.IsContainerd() {
		return nil, fmt.Errorf("platform %s does not support containerd API directly. Use docker runtime instead", platform.Name)
	}

	// Set namespace for DVM containers
	namespace := "devopsmaestro"

	// Create containerd client
	ctdClient, err := client.New(platform.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w\n%s", err, platform.GetStartHint())
	}

	// Verify connection
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	version, err := ctdClient.Version(ctx)
	if err != nil {
		ctdClient.Close()
		return nil, fmt.Errorf("failed to connect to containerd: %w\n%s", err, platform.GetStartHint())
	}

	fmt.Printf("Connected to containerd %s (%s, namespace: %s)\n",
		version.Version, platform.Name, namespace)

	return &ContainerdRuntimeV2{
		client:    ctdClient,
		platform:  platform,
		namespace: namespace,
	}, nil
}

// GetRuntimeType returns the runtime type
func (r *ContainerdRuntimeV2) GetRuntimeType() string {
	return "containerd-v2"
}

// BuildImage is not implemented (use BuildKit API instead)
func (r *ContainerdRuntimeV2) BuildImage(ctx context.Context, opts BuildOptions) error {
	return fmt.Errorf("use 'dvm build' command which uses BuildKit API")
}

// Close closes the containerd client
func (r *ContainerdRuntimeV2) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// GetPlatformName returns the human-readable platform name
func (r *ContainerdRuntimeV2) GetPlatformName() string {
	return r.platform.Name
}
