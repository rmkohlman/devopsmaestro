package operators

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

// =============================================================================
// Task 2.5: SSH Agent Forwarding Tests (v0.19.0)
// Tests verify SSH agent forwarding is opt-in and platform-aware
// =============================================================================

// TestSSHAgentForwardingDisabledByDefault verifies workspace without flag has no SSH
func TestSSHAgentForwardingDisabledByDefault(t *testing.T) {
	tests := []struct {
		name        string
		options     StartOptions
		expectSSH   bool
		description string
	}{
		{
			name: "default options no SSH",
			options: StartOptions{
				WorkspaceName: "test-ws",
				ImageName:     "test:latest",
				// SSHAgentForwarding not set (default: false)
			},
			expectSSH:   false,
			description: "By default, SSH agent should not be forwarded",
		},
		{
			name: "explicitly disabled",
			options: StartOptions{
				WorkspaceName:      "test-ws",
				ImageName:          "test:latest",
				SSHAgentForwarding: false,
			},
			expectSSH:   false,
			description: "Explicitly disabled SSH agent should not be forwarded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIXME: This test will FAIL - SSHAgentForwarding field doesn't exist yet
			// After Phase 3, StartOptions should have:
			// type StartOptions struct {
			//     ...
			//     SSHAgentForwarding bool
			// }

			// Verify field exists and is false by default
			if tt.options.SSHAgentForwarding != tt.expectSSH {
				t.Errorf("SSHAgentForwarding = %v, want %v", tt.options.SSHAgentForwarding, tt.expectSSH)
			}
		})
	}
}

// TestSSHAgentForwardingEnabled verifies workspace with flag forwards agent
func TestSSHAgentForwardingEnabled(t *testing.T) {
	options := StartOptions{
		WorkspaceName:      "test-ws",
		ImageName:          "test:latest",
		SSHAgentForwarding: true, // FIXME: Field doesn't exist yet
	}

	// FIXME: This test will FAIL - SSHAgentForwarding field doesn't exist yet
	if !options.SSHAgentForwarding {
		t.Errorf("SSHAgentForwarding = false, want true")
	}

	// Verify this translates to container mount
	// After Phase 3, container runtime should check this field and:
	// - On Linux: Mount $SSH_AUTH_SOCK
	// - On macOS with Docker Desktop: Mount /run/host-services/ssh-auth.sock
	// - On macOS with Colima: Mount socket from Colima VM
}

// TestSSHAgentSocketPath verifies correct socket path per platform
func TestSSHAgentSocketPath(t *testing.T) {
	tests := []struct {
		name         string
		platform     string
		runtime      string
		expectedPath string
		checkEnv     bool
	}{
		{
			name:         "Linux native",
			platform:     "linux",
			runtime:      "docker",
			expectedPath: "$SSH_AUTH_SOCK", // From environment
			checkEnv:     true,
		},
		{
			name:         "macOS Docker Desktop",
			platform:     "darwin",
			runtime:      "docker",
			expectedPath: "/run/host-services/ssh-auth.sock",
			checkEnv:     false,
		},
		{
			name:         "macOS Colima",
			platform:     "darwin",
			runtime:      "colima",
			expectedPath: "$SSH_AUTH_SOCK", // Colima forwards host socket
			checkEnv:     true,
		},
		{
			name:         "macOS OrbStack",
			platform:     "darwin",
			runtime:      "orbstack",
			expectedPath: "/run/host-services/ssh-auth.sock", // Similar to Docker Desktop
			checkEnv:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if not on the right platform
			if runtime.GOOS != tt.platform {
				t.Skipf("Test requires %s, running on %s", tt.platform, runtime.GOOS)
			}

			// FIXME: This test will FAIL - GetSSHAgentSocketPath() doesn't exist yet
			// After Phase 3, container runtime should have:
			// func GetSSHAgentSocketPath(runtimeType string) (string, error)
			socketPath, err := GetSSHAgentSocketPath(tt.runtime)
			if err != nil {
				t.Fatalf("GetSSHAgentSocketPath() error = %v", err)
			}

			// Verify path matches expected
			// Note: For env vars, we check the pattern not the actual value
			if tt.checkEnv {
				if socketPath == "" {
					t.Errorf("GetSSHAgentSocketPath() returned empty, want path from environment")
				}
			} else {
				if socketPath != tt.expectedPath {
					t.Errorf("GetSSHAgentSocketPath() = %q, want %q", socketPath, tt.expectedPath)
				}
			}
		})
	}
}

// TestSSHAgentNotAvailable tests graceful failure when no agent running
func TestSSHAgentNotAvailable(t *testing.T) {
	// Clear SSH_AUTH_SOCK to simulate no agent
	originalAuthSock := ""
	if val, exists := os.LookupEnv("SSH_AUTH_SOCK"); exists {
		originalAuthSock = val
		defer os.Setenv("SSH_AUTH_SOCK", originalAuthSock)
	}
	os.Unsetenv("SSH_AUTH_SOCK")

	options := StartOptions{
		WorkspaceName:      "test-ws",
		ImageName:          "test:latest",
		SSHAgentForwarding: true, // FIXME: Field doesn't exist yet
	}

	ctx := context.Background()

	// FIXME: This test will FAIL - StartWorkspace doesn't handle missing agent yet
	// After Phase 3, StartWorkspace should:
	// 1. Check if SSH agent is available when SSHAgentForwarding=true
	// 2. Return clear error if agent requested but not available
	// 3. NOT fail silently - user should know their SSH won't work

	mockRuntime := &MockContainerRuntime{
		StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
			if opts.SSHAgentForwarding {
				// Check SSH agent availability
				if os.Getenv("SSH_AUTH_SOCK") == "" {
					return "", fmt.Errorf("SSH agent forwarding requested but SSH_AUTH_SOCK not set")
				}
			}
			return "mock-container-id", nil
		},
	}

	_, err := mockRuntime.StartWorkspace(ctx, options)

	// Should get clear error about SSH agent not available
	if err == nil {
		t.Errorf("StartWorkspace() expected error when SSH agent not available, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "SSH") {
		t.Errorf("StartWorkspace() error should mention SSH, got: %v", err)
	}
}

// TestNoSSHKeyMounting verifies ~/.ssh is NEVER mounted
func TestNoSSHKeyMounting(t *testing.T) {
	options := StartOptions{
		WorkspaceName:      "test-ws",
		ImageName:          "test:latest",
		SSHAgentForwarding: false, // Even with agent disabled
	}

	ctx := context.Background()

	mockRuntime := &MockContainerRuntime{
		StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
			// FIXME: After Phase 3, need to verify container mounts
			// No mount should include ~/.ssh directory
			// This is the old behavior that must be REMOVED

			// For now, mock the check
			if hasSshMount(opts.Mounts) {
				return "", fmt.Errorf("container should not mount ~/.ssh directory")
			}

			return "mock-container-id", nil
		},
	}

	// FIXME: This test will FAIL - Mounts field structure may change
	_, err := mockRuntime.StartWorkspace(ctx, options)
	if err != nil {
		t.Errorf("StartWorkspace() should succeed without SSH mount, got error: %v", err)
	}

	// Verify SSH keys are NOT in container even with SSH agent forwarding
	optionsWithAgent := options
	optionsWithAgent.SSHAgentForwarding = true

	_, err = mockRuntime.StartWorkspace(ctx, optionsWithAgent)
	if err != nil {
		t.Errorf("StartWorkspace() should succeed with only agent socket, got error: %v", err)
	}
}

// TestSSHAgentSecurityDefault verifies opt-in security model
func TestSSHAgentSecurityDefault(t *testing.T) {
	// This test verifies the security requirement from @security:
	// "SSH agent forwarding must be opt-in (default: false)"
	// This follows principle of least privilege

	tests := []struct {
		name              string
		createOptions     func() StartOptions
		shouldHaveSSH     bool
		securityPrinciple string
	}{
		{
			name: "new workspace no flags",
			createOptions: func() StartOptions {
				return StartOptions{
					WorkspaceName: "test-ws",
					ImageName:     "test:latest",
					// No SSHAgentForwarding field set
				}
			},
			shouldHaveSSH:     false,
			securityPrinciple: "Principle of least privilege - no SSH by default",
		},
		{
			name: "workspace with explicit opt-in",
			createOptions: func() StartOptions {
				return StartOptions{
					WorkspaceName:      "test-ws",
					ImageName:          "test:latest",
					SSHAgentForwarding: true, // Explicit opt-in
				}
			},
			shouldHaveSSH:     true,
			securityPrinciple: "User explicitly opted in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.createOptions()

			// FIXME: This test will FAIL - SSHAgentForwarding field doesn't exist yet
			hasSSH := opts.SSHAgentForwarding

			if hasSSH != tt.shouldHaveSSH {
				t.Errorf("SSH agent forwarding = %v, want %v (%s)",
					hasSSH, tt.shouldHaveSSH, tt.securityPrinciple)
			}
		})
	}
}

// TestSSHAgentPlatformSpecific verifies platform-specific handling
func TestSSHAgentPlatformSpecific(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		runtime  string
		verify   func(t *testing.T, mountPath string)
	}{
		{
			name:     "Docker Desktop on macOS",
			platform: "darwin",
			runtime:  "docker",
			verify: func(t *testing.T, mountPath string) {
				// Docker Desktop provides SSH agent at /run/host-services/ssh-auth.sock
				if !strings.Contains(mountPath, "/run/host-services/ssh-auth.sock") {
					t.Errorf("macOS Docker Desktop should mount /run/host-services/ssh-auth.sock, got: %s", mountPath)
				}
			},
		},
		{
			name:     "Colima on macOS",
			platform: "darwin",
			runtime:  "colima",
			verify: func(t *testing.T, mountPath string) {
				// Colima forwards SSH_AUTH_SOCK from host
				if mountPath == "" {
					t.Errorf("Colima should forward SSH_AUTH_SOCK from host")
				}
			},
		},
		{
			name:     "Docker on Linux",
			platform: "linux",
			runtime:  "docker",
			verify: func(t *testing.T, mountPath string) {
				// Linux should use SSH_AUTH_SOCK from environment
				if mountPath == "" {
					t.Errorf("Linux should use SSH_AUTH_SOCK from environment")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS != tt.platform {
				t.Skipf("Test requires %s, running on %s", tt.platform, runtime.GOOS)
			}

			// FIXME: This test will FAIL - GetSSHAgentMountPath() doesn't exist yet
			// After Phase 3, should have platform-specific logic
			mountPath, err := GetSSHAgentMountPath(tt.runtime)
			if err != nil {
				t.Fatalf("GetSSHAgentMountPath() error = %v", err)
			}

			tt.verify(t, mountPath)
		})
	}
}

// hasSshMount checks if mounts include ~/.ssh directory (helper function)
func hasSshMount(mounts []MountConfig) bool {
	for _, mount := range mounts {
		if strings.Contains(mount.Source, ".ssh") || strings.Contains(mount.Destination, ".ssh") {
			return true
		}
	}
	return false
}
