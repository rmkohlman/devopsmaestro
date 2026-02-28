package operators

import (
	"context"
	"strings"
	"testing"
)

// =============================================================================
// Task 2.6: Container Volume Mounts Tests (v0.19.0)
// Tests verify workspace volumes are mounted correctly and no host paths leak
// =============================================================================

// TestWorkspaceVolumesMounted verifies nvim-data, nvim-state, cache are mounted
func TestWorkspaceVolumesMounted(t *testing.T) {
	workspaceSlug := "test-eco-domain-app-ws"

	tests := []struct {
		name             string
		volumeType       string
		expectedHostPath string
		expectedContPath string
		readOnly         bool
	}{
		{
			name:             "nvim data volume",
			volumeType:       "nvim-data",
			expectedHostPath: ".devopsmaestro/workspaces/" + workspaceSlug + "/volume/nvim-data",
			expectedContPath: "/home/dev/.local/share/nvim",
			readOnly:         false,
		},
		{
			name:             "nvim state volume",
			volumeType:       "nvim-state",
			expectedHostPath: ".devopsmaestro/workspaces/" + workspaceSlug + "/volume/nvim-state",
			expectedContPath: "/home/dev/.local/state/nvim",
			readOnly:         false,
		},
		{
			name:             "cache volume",
			volumeType:       "cache",
			expectedHostPath: ".devopsmaestro/workspaces/" + workspaceSlug + "/volume/cache",
			expectedContPath: "/home/dev/.cache",
			readOnly:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := StartOptions{
				WorkspaceName: "ws",
				ImageName:     "test:latest",
				WorkspaceSlug: workspaceSlug,
			}

			// Compute workspace mounts from slug
			mounts, err := options.ComputeWorkspaceMounts()
			if err != nil {
				t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
			}
			options.Mounts = mounts

			ctx := context.Background()

			mockRuntime := &MockContainerRuntime{
				StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
					// Verify the volume is in the mounts list
					found := false
					for _, mount := range opts.Mounts {
						if strings.Contains(mount.Source, tt.expectedHostPath) &&
							strings.Contains(mount.Destination, tt.expectedContPath) {
							found = true

							// Verify read-only flag
							if mount.ReadOnly != tt.readOnly {
								t.Errorf("Mount %s ReadOnly = %v, want %v", tt.volumeType, mount.ReadOnly, tt.readOnly)
							}
						}
					}

					if !found {
						t.Errorf("Volume mount for %s not found in options", tt.volumeType)
					}

					return "mock-container-id", nil
				},
			}

			_, err = mockRuntime.StartWorkspace(ctx, options)
			if err != nil {
				t.Fatalf("StartWorkspace() error = %v", err)
			}
		})
	}
}

// TestGeneratedConfigsMountedReadOnly verifies .dvm/nvim, .dvm/shell mounted :ro
func TestGeneratedConfigsMountedReadOnly(t *testing.T) {
	workspaceSlug := "test-eco-domain-app-ws"

	tests := []struct {
		name             string
		configType       string
		expectedHostPath string
		expectedContPath string
		mustBeReadOnly   bool
	}{
		{
			name:             "nvim config read-only",
			configType:       "nvim",
			expectedHostPath: ".devopsmaestro/workspaces/" + workspaceSlug + "/.dvm/nvim",
			expectedContPath: "/home/dev/.config/nvim",
			mustBeReadOnly:   true,
		},
		{
			name:             "shell config read-only",
			configType:       "shell",
			expectedHostPath: ".devopsmaestro/workspaces/" + workspaceSlug + "/.dvm/shell",
			expectedContPath: "/home/dev/.config/shell",
			mustBeReadOnly:   true,
		},
		{
			name:             "starship config read-only",
			configType:       "starship",
			expectedHostPath: ".devopsmaestro/workspaces/" + workspaceSlug + "/.dvm/starship",
			expectedContPath: "/home/dev/.config/starship.toml",
			mustBeReadOnly:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := StartOptions{
				WorkspaceName: "ws",
				ImageName:     "test:latest",
				WorkspaceSlug: workspaceSlug,
			}

			// Compute workspace mounts from slug
			mounts, err := options.ComputeWorkspaceMounts()
			if err != nil {
				t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
			}
			options.Mounts = mounts

			ctx := context.Background()

			mockRuntime := &MockContainerRuntime{
				StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
					// Verify generated configs are read-only

					found := false
					for _, mount := range opts.Mounts {
						if strings.Contains(mount.Source, tt.expectedHostPath) {
							found = true

							// CRITICAL: Generated configs must be read-only
							if !mount.ReadOnly {
								t.Errorf("Generated config %s must be read-only, got ReadOnly=false", tt.configType)
							}
						}
					}

					if !found {
						t.Errorf("Generated config mount for %s not found", tt.configType)
					}

					return "mock-container-id", nil
				},
			}

			_, err = mockRuntime.StartWorkspace(ctx, options)
			if err != nil {
				t.Fatalf("StartWorkspace() error = %v", err)
			}
		})
	}
}

// TestRepoDirectoryMounted verifies repo/ directory is mounted for git clone
func TestRepoDirectoryMounted(t *testing.T) {
	workspaceSlug := "test-eco-domain-app-ws"

	options := StartOptions{
		WorkspaceName: "ws",
		ImageName:     "test:latest",
		WorkspaceSlug: workspaceSlug,
	}

	// Compute workspace mounts from slug
	mounts, err := options.ComputeWorkspaceMounts()
	if err != nil {
		t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
	}
	options.Mounts = mounts

	ctx := context.Background()

	mockRuntime := &MockContainerRuntime{
		StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
			// Verify repo directory mount

			expectedHostPath := ".devopsmaestro/workspaces/" + workspaceSlug + "/repo"
			expectedContPath := "/workspace"

			found := false
			for _, mount := range opts.Mounts {
				if strings.Contains(mount.Source, expectedHostPath) &&
					mount.Destination == expectedContPath {
					found = true

					// Repo must be read-write (user needs to git commit)
					if mount.ReadOnly {
						t.Errorf("Repo directory must be read-write, got ReadOnly=true")
					}
				}
			}

			if !found {
				t.Errorf("Repo directory mount not found")
			}

			return "mock-container-id", nil
		},
	}

	_, err = mockRuntime.StartWorkspace(ctx, options)
	if err != nil {
		t.Fatalf("StartWorkspace() error = %v", err)
	}
}

// TestNoHostPathLeakage verifies no paths outside ~/.devopsmaestro/ are mounted
func TestNoHostPathLeakage(t *testing.T) {
	options := StartOptions{
		WorkspaceName: "ws",
		ImageName:     "test:latest",
		WorkspaceSlug: "test-eco-domain-app-ws",
	}

	// Compute workspace mounts from slug
	mounts, err := options.ComputeWorkspaceMounts()
	if err != nil {
		t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
	}
	options.Mounts = mounts

	ctx := context.Background()

	// Prohibited host paths that should NEVER be mounted
	prohibitedPaths := []string{
		"~/.config/nvim",      // Host nvim config
		"~/.local/share/nvim", // Host nvim data
		"~/.ssh",              // SSH keys (use agent instead)
		"~/.gitconfig",        // Host git config
		"~/.zshrc",            // Host shell config
		"/etc",                // System config
		"/usr",                // System binaries
		"/var",                // System data
		"/home/",              // Other user directories
	}

	mockRuntime := &MockContainerRuntime{
		StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
			// Verify no prohibited paths

			for _, mount := range opts.Mounts {
				// Check if mount source contains any prohibited path
				for _, prohibited := range prohibitedPaths {
					if strings.Contains(mount.Source, prohibited) {
						t.Errorf("Mount source %q contains prohibited path %q - host path leakage!", mount.Source, prohibited)
					}
				}

				// All mounts should be under .devopsmaestro/workspaces/
				if !strings.Contains(mount.Source, ".devopsmaestro/workspaces/") {
					t.Errorf("Mount source %q is not under .devopsmaestro/workspaces/ - possible leakage!", mount.Source)
				}
			}

			return "mock-container-id", nil
		},
	}

	_, err = mockRuntime.StartWorkspace(ctx, options)
	if err != nil {
		t.Fatalf("StartWorkspace() error = %v", err)
	}
}

// TestVolumePathsMatchWorkspaceSlug verifies mount paths use correct slug
func TestVolumePathsMatchWorkspaceSlug(t *testing.T) {
	tests := []struct {
		name          string
		workspaceSlug string
	}{
		{
			name:          "dev workspace",
			workspaceSlug: "personal-tools-dvm-dev",
		},
		{
			name:          "production workspace",
			workspaceSlug: "enterprise-payments-api-prod",
		},
		{
			name:          "multi-word slug",
			workspaceSlug: "my-org-data-platform-etl-service-local-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := StartOptions{
				WorkspaceName: "ws",
				ImageName:     "test:latest",
				WorkspaceSlug: tt.workspaceSlug,
			}

			// Compute workspace mounts from slug
			mounts, err := options.ComputeWorkspaceMounts()
			if err != nil {
				t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
			}
			options.Mounts = mounts

			ctx := context.Background()

			mockRuntime := &MockContainerRuntime{
				StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
					// Verify all mount paths contain slug

					for _, mount := range opts.Mounts {
						// Skip SSH agent socket (doesn't contain slug)
						if strings.Contains(mount.Source, "ssh-auth.sock") {
							continue
						}

						// All workspace-related mounts should contain the slug
						if !strings.Contains(mount.Source, tt.workspaceSlug) {
							t.Errorf("Mount source %q does not contain workspace slug %q", mount.Source, tt.workspaceSlug)
						}
					}

					return "mock-container-id", nil
				},
			}

			_, err = mockRuntime.StartWorkspace(ctx, options)
			if err != nil {
				t.Fatalf("StartWorkspace() error = %v", err)
			}
		})
	}
}

// TestVolumePermissions verifies volume directories have correct permissions
func TestVolumePermissions(t *testing.T) {
	workspaceSlug := "test-eco-domain-app-ws"

	options := StartOptions{
		WorkspaceName: "ws",
		ImageName:     "test:latest",
		WorkspaceSlug: workspaceSlug,
	}

	// Compute workspace mounts from slug
	mounts, err := options.ComputeWorkspaceMounts()
	if err != nil {
		t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
	}
	options.Mounts = mounts

	ctx := context.Background()

	mockRuntime := &MockContainerRuntime{
		StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
			// Verify volume permissions by checking mount configuration

			// This would be tested by checking the actual filesystem after creation
			// For now, we verify the mount configuration is correct

			volumeMounts := []string{
				"volume/nvim-data",
				"volume/nvim-state",
				"volume/cache",
			}

			for _, volPath := range volumeMounts {
				found := false
				for _, mount := range opts.Mounts {
					if strings.Contains(mount.Source, volPath) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Volume mount for %s not found", volPath)
				}
			}

			return "mock-container-id", nil
		},
	}

	_, err = mockRuntime.StartWorkspace(ctx, options)
	if err != nil {
		t.Fatalf("StartWorkspace() error = %v", err)
	}
}

// TestMountTypes verifies correct mount types are used
func TestMountTypes(t *testing.T) {
	workspaceSlug := "test-eco-domain-app-ws"

	tests := []struct {
		name         string
		mountPath    string
		expectedType string
		description  string
	}{
		{
			name:         "bind mount for repo",
			mountPath:    "repo",
			expectedType: "bind",
			description:  "Git repo should use bind mount",
		},
		{
			name:         "bind mount for nvim-data",
			mountPath:    "volume/nvim-data",
			expectedType: "bind",
			description:  "Plugin data should use bind mount",
		},
		{
			name:         "bind mount for generated configs",
			mountPath:    ".dvm/nvim",
			expectedType: "bind",
			description:  "Generated configs should use bind mount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := StartOptions{
				WorkspaceName: "ws",
				ImageName:     "test:latest",
				WorkspaceSlug: workspaceSlug,
			}

			// Compute workspace mounts from slug
			mounts, err := options.ComputeWorkspaceMounts()
			if err != nil {
				t.Fatalf("ComputeWorkspaceMounts() error = %v", err)
			}
			options.Mounts = mounts

			ctx := context.Background()

			mockRuntime := &MockContainerRuntime{
				StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
					// Verify mount type

					for _, mount := range opts.Mounts {
						if strings.Contains(mount.Source, tt.mountPath) {
							if mount.Type != tt.expectedType {
								t.Errorf("Mount %s has type %s, want %s (%s)",
									tt.mountPath, mount.Type, tt.expectedType, tt.description)
							}
						}
					}

					return "mock-container-id", nil
				},
			}

			_, err = mockRuntime.StartWorkspace(ctx, options)
			if err != nil {
				t.Fatalf("StartWorkspace() error = %v", err)
			}
		})
	}
}

// TestWorkspaceIsolation verifies multiple workspaces don't share volumes
func TestWorkspaceIsolation(t *testing.T) {
	workspace1 := "eco1-dom1-app1-ws1"
	workspace2 := "eco2-dom2-app2-ws2"

	// Start first workspace
	opts1 := StartOptions{
		WorkspaceName: "ws1",
		ImageName:     "test:latest",
		WorkspaceSlug: workspace1,
	}

	// Compute workspace mounts for workspace 1
	mounts1, err := opts1.ComputeWorkspaceMounts()
	if err != nil {
		t.Fatalf("ComputeWorkspaceMounts(ws1) error = %v", err)
	}
	opts1.Mounts = mounts1

	// Start second workspace
	opts2 := StartOptions{
		WorkspaceName: "ws2",
		ImageName:     "test:latest",
		WorkspaceSlug: workspace2,
	}

	// Compute workspace mounts for workspace 2
	mounts2, err := opts2.ComputeWorkspaceMounts()
	if err != nil {
		t.Fatalf("ComputeWorkspaceMounts(ws2) error = %v", err)
	}
	opts2.Mounts = mounts2

	ctx := context.Background()

	mockRuntime := &MockContainerRuntime{
		StartWorkspaceFunc: func(ctx context.Context, opts StartOptions) (string, error) {
			// Verify workspace isolation - each workspace has separate mount paths

			for _, mount := range opts.Mounts {
				// Verify mount contains the workspace's slug
				if !strings.Contains(mount.Source, opts.WorkspaceSlug) {
					t.Errorf("Mount %s does not contain workspace slug %s", mount.Source, opts.WorkspaceSlug)
				}

				// Verify it doesn't contain the OTHER workspace's slug
				otherSlug := workspace1
				if opts.WorkspaceSlug == workspace1 {
					otherSlug = workspace2
				}
				if strings.Contains(mount.Source, otherSlug) {
					t.Errorf("Mount %s contains other workspace slug %s - isolation breach!", mount.Source, otherSlug)
				}
			}

			return "mock-container-id", nil
		},
	}

	_, err = mockRuntime.StartWorkspace(ctx, opts1)
	if err != nil {
		t.Fatalf("StartWorkspace(ws1) error = %v", err)
	}

	_, err = mockRuntime.StartWorkspace(ctx, opts2)
	if err != nil {
		t.Fatalf("StartWorkspace(ws2) error = %v", err)
	}
}
