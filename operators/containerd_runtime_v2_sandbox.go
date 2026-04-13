package operators

import (
	"context"
	"fmt"
)

// ErrNotImplemented is returned when a containerd operation is not yet supported.
var ErrNotImplemented = fmt.Errorf("not implemented for containerd runtime")

// RemoveContainer removes a container by ID or name (containerd stub).
func (r *ContainerdRuntimeV2) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return fmt.Errorf("RemoveContainer: %w", ErrNotImplemented)
}

// RemoveImage removes a container image by name or ID (containerd stub).
func (r *ContainerdRuntimeV2) RemoveImage(ctx context.Context, imageID string) error {
	return fmt.Errorf("RemoveImage: %w", ErrNotImplemented)
}

// ListContainers lists containers matching the given label selectors (containerd stub).
func (r *ContainerdRuntimeV2) ListContainers(ctx context.Context, labels map[string]string) ([]ContainerInfo, error) {
	return nil, fmt.Errorf("ListContainers: %w", ErrNotImplemented)
}

// ImageExists checks whether a container image exists locally (containerd stub).
func (r *ContainerdRuntimeV2) ImageExists(ctx context.Context, imageName string) (bool, error) {
	return false, fmt.Errorf("ImageExists: %w", ErrNotImplemented)
}
