package operators

import (
	"fmt"
	"os"
	"runtime"
)

// GetSSHAgentSocketPath returns the path to the SSH agent socket
// based on the runtime type and platform.
//
// Platform-specific behavior:
// - Linux native: Uses $SSH_AUTH_SOCK from environment
// - macOS Docker Desktop: /run/host-services/ssh-auth.sock
// - macOS Colima: Uses $SSH_AUTH_SOCK forwarded from host
// - macOS OrbStack: /run/host-services/ssh-auth.sock (similar to Docker Desktop)
//
// Returns empty string if SSH agent is not available.
func GetSSHAgentSocketPath(runtimeType string) (string, error) {
	// Check environment variable first (works for most cases)
	authSock := os.Getenv("SSH_AUTH_SOCK")

	// Platform-specific overrides
	if runtime.GOOS == "darwin" {
		// On macOS, some runtimes provide a special socket path
		switch runtimeType {
		case "docker":
			// Docker Desktop on macOS provides SSH agent at this path
			return "/run/host-services/ssh-auth.sock", nil
		case "orbstack":
			// OrbStack also provides SSH agent at the same path
			return "/run/host-services/ssh-auth.sock", nil
		case "colima":
			// Colima forwards the host's SSH_AUTH_SOCK
			if authSock == "" {
				return "", fmt.Errorf("SSH_AUTH_SOCK not set (required for Colima)")
			}
			return authSock, nil
		}
	}

	// For Linux or other platforms, use SSH_AUTH_SOCK
	if authSock == "" {
		return "", fmt.Errorf("SSH_AUTH_SOCK not set (SSH agent not running)")
	}

	return authSock, nil
}

// GetSSHAgentMountPath is an alias for GetSSHAgentSocketPath for test compatibility
func GetSSHAgentMountPath(runtimeType string) (string, error) {
	return GetSSHAgentSocketPath(runtimeType)
}

// GetSSHAgentMount returns the mount configuration for SSH agent forwarding.
// Returns the host socket path and the container socket path.
//
// The container path is standardized to /tmp/ssh-agent.sock and the
// SSH_AUTH_SOCK environment variable should be set to this path in the container.
//
// Returns error if SSH agent forwarding is requested but agent is not available.
func GetSSHAgentMount(runtimeType string) (hostPath, containerPath string, err error) {
	hostPath, err = GetSSHAgentSocketPath(runtimeType)
	if err != nil {
		return "", "", err
	}

	// Standardize container path
	containerPath = "/tmp/ssh-agent.sock"

	return hostPath, containerPath, nil
}

// ValidateSSHAgent checks if SSH agent is available when forwarding is requested
func ValidateSSHAgent(opts StartOptions, runtimeType string) error {
	if !opts.SSHAgentForwarding {
		return nil // No validation needed if not requested
	}

	_, err := GetSSHAgentSocketPath(runtimeType)
	if err != nil {
		return fmt.Errorf("SSH agent forwarding requested but SSH agent not available: %w", err)
	}

	return nil
}
