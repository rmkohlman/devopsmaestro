package operators

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestStartOptions_Helpers tests the new helper functions for container naming and command computation
func TestStartOptions_Helpers(t *testing.T) {
	t.Run("ComputeContainerName with ContainerName set", func(t *testing.T) {
		opts := StartOptions{
			WorkspaceName: "my-workspace",
			ContainerName: "custom-container",
		}

		result := opts.ComputeContainerName()
		expected := "custom-container"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeContainerName with ContainerName empty", func(t *testing.T) {
		opts := StartOptions{
			WorkspaceName: "my-workspace",
			ContainerName: "",
		}

		result := opts.ComputeContainerName()
		expected := "my-workspace"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeContainerName with hierarchical naming - full hierarchy", func(t *testing.T) {
		opts := StartOptions{
			EcosystemName: "production",
			DomainName:    "backend",
			AppName:       "userservice",
			WorkspaceName: "dev",
		}

		result := opts.ComputeContainerName()
		expected := "dvm-production-backend-userservice-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeContainerName with hierarchical naming - ecosystem only", func(t *testing.T) {
		opts := StartOptions{
			EcosystemName: "staging",
			AppName:       "frontend",
			WorkspaceName: "test",
		}

		result := opts.ComputeContainerName()
		expected := "dvm-staging-frontend-test"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeContainerName with hierarchical naming - domain only", func(t *testing.T) {
		opts := StartOptions{
			DomainName:    "data",
			AppName:       "analytics",
			WorkspaceName: "dev",
		}

		result := opts.ComputeContainerName()
		expected := "dvm-data-analytics-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeContainerName with hierarchical naming - case normalization", func(t *testing.T) {
		opts := StartOptions{
			EcosystemName: "PROD",
			DomainName:    "Backend",
			AppName:       "UserService",
			WorkspaceName: "DEV",
		}

		result := opts.ComputeContainerName()
		expected := "dvm-prod-backend-userservice-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeContainerName explicit container name overrides hierarchy", func(t *testing.T) {
		opts := StartOptions{
			ContainerName: "custom-name",
			EcosystemName: "production",
			DomainName:    "backend",
			AppName:       "userservice",
			WorkspaceName: "dev",
		}

		result := opts.ComputeContainerName()
		expected := "custom-name"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ComputeCommand with Command set", func(t *testing.T) {
		opts := StartOptions{
			Command: []string{"/bin/zsh", "-l"},
		}

		result := opts.ComputeCommand()
		expected := []string{"/bin/zsh", "-l"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, cmd := range result {
			if cmd != expected[i] {
				t.Errorf("Expected %v, got %v", expected, result)
				break
			}
		}
	})

	t.Run("ComputeCommand with Command empty", func(t *testing.T) {
		opts := StartOptions{
			Command: []string{},
		}

		result := opts.ComputeCommand()
		expected := []string{"/bin/sleep", "infinity"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, cmd := range result {
			if cmd != expected[i] {
				t.Errorf("Expected %v, got %v", expected, result)
				break
			}
		}
	})

	t.Run("DefaultKeepAliveCommand", func(t *testing.T) {
		result := DefaultKeepAliveCommand()
		expected := []string{"/bin/sleep", "infinity"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, cmd := range result {
			if cmd != expected[i] {
				t.Errorf("Expected %v, got %v", expected, result)
				break
			}
		}
	})
}

// TestHierarchicalNamingStrategy tests the hierarchical container naming strategy
func TestHierarchicalNamingStrategy(t *testing.T) {
	strategy := NewHierarchicalNamingStrategy()

	t.Run("GenerateName - full hierarchy", func(t *testing.T) {
		result := strategy.GenerateName("production", "backend", "userservice", "dev")
		expected := "dvm-production-backend-userservice-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("GenerateName - ecosystem only", func(t *testing.T) {
		result := strategy.GenerateName("staging", "", "frontend", "test")
		expected := "dvm-staging-frontend-test"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("GenerateName - domain only", func(t *testing.T) {
		result := strategy.GenerateName("", "data", "analytics", "dev")
		expected := "dvm-data-analytics-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("GenerateName - legacy format (no hierarchy)", func(t *testing.T) {
		result := strategy.GenerateName("", "", "myapp", "dev")
		expected := "dvm-myapp-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("GenerateName - case normalization", func(t *testing.T) {
		result := strategy.GenerateName("PROD", "Backend", "UserService", "DEV")
		expected := "dvm-prod-backend-userservice-dev"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("ParseName - full hierarchy", func(t *testing.T) {
		ecosystem, domain, app, workspace, ok := strategy.ParseName("dvm-production-backend-userservice-dev")
		if !ok {
			t.Errorf("Expected parsing to succeed")
		}
		if ecosystem != "production" || domain != "backend" || app != "userservice" || workspace != "dev" {
			t.Errorf("Expected (production, backend, userservice, dev), got (%s, %s, %s, %s)",
				ecosystem, domain, app, workspace)
		}
	})

	t.Run("ParseName - ecosystem only", func(t *testing.T) {
		ecosystem, domain, app, workspace, ok := strategy.ParseName("dvm-staging-frontend-test")
		if !ok {
			t.Errorf("Expected parsing to succeed")
		}
		if ecosystem != "staging" || domain != "" || app != "frontend" || workspace != "test" {
			t.Errorf("Expected (staging, , frontend, test), got (%s, %s, %s, %s)",
				ecosystem, domain, app, workspace)
		}
	})

	t.Run("ParseName - legacy format", func(t *testing.T) {
		ecosystem, domain, app, workspace, ok := strategy.ParseName("dvm-myapp-dev")
		if !ok {
			t.Errorf("Expected parsing to succeed")
		}
		if ecosystem != "" || domain != "" || app != "myapp" || workspace != "dev" {
			t.Errorf("Expected (, , myapp, dev), got (%s, %s, %s, %s)",
				ecosystem, domain, app, workspace)
		}
	})

	t.Run("ParseName - invalid format (no dvm prefix)", func(t *testing.T) {
		_, _, _, _, ok := strategy.ParseName("invalid-name")
		if ok {
			t.Errorf("Expected parsing to fail for invalid format")
		}
	})

	t.Run("ParseName - invalid format (too few parts)", func(t *testing.T) {
		_, _, _, _, ok := strategy.ParseName("dvm-only")
		if ok {
			t.Errorf("Expected parsing to fail for too few parts")
		}
	})

	t.Run("ParseName - invalid format (too many parts)", func(t *testing.T) {
		_, _, _, _, ok := strategy.ParseName("dvm-too-many-parts-here-test-dev-extra")
		if ok {
			t.Errorf("Expected parsing to fail for too many parts")
		}
	})
}

// requireContainerdPlatform skips the test if no containerd platform is available
func requireContainerdPlatform(t *testing.T) *Platform {
	t.Helper()

	detector, err := NewPlatformDetector()
	if err != nil {
		t.Skipf("Failed to create platform detector: %v", err)
	}

	// Look for a containerd-capable platform (Colima with containerd)
	platforms := detector.DetectAll()
	for _, p := range platforms {
		if p.IsContainerd() {
			return p
		}
	}

	t.Skip("No containerd platform available, skipping integration test")
	return nil
}

// requireAlpineImage ensures alpine:latest is available in the containerd namespace.
// If not available, attempts to pull it. Skips the test if pull fails.
func requireAlpineImage(t *testing.T, platform *Platform) {
	t.Helper()

	// Check if image exists using nerdctl
	checkCmd := exec.Command("nerdctl",
		"--address", platform.SocketPath,
		"--namespace", "devopsmaestro",
		"image", "inspect", "alpine:latest",
	)
	if err := checkCmd.Run(); err == nil {
		// Image exists
		return
	}

	// Try to pull the image
	t.Log("Pulling alpine:latest for test...")
	pullCmd := exec.Command("nerdctl",
		"--address", platform.SocketPath,
		"--namespace", "devopsmaestro",
		"pull", "alpine:latest",
	)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr

	if err := pullCmd.Run(); err != nil {
		t.Skipf("Failed to pull alpine:latest: %v (test requires this image)", err)
	}
}

// TestContainerdRuntimeV2Creation tests runtime creation
func TestContainerdRuntimeV2Creation(t *testing.T) {
	platform := requireContainerdPlatform(t)

	runtime, err := NewContainerdRuntimeV2WithPlatform(platform)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	if runtime.GetRuntimeType() != "containerd-v2" {
		t.Errorf("Expected runtime type 'containerd-v2', got '%s'", runtime.GetRuntimeType())
	}

	if runtime.namespace != "devopsmaestro" {
		t.Errorf("Expected namespace 'devopsmaestro', got '%s'", runtime.namespace)
	}
}

// TestStartOptions validates StartOptions struct
func TestStartOptions(t *testing.T) {
	opts := StartOptions{
		ImageName:     "alpine:latest",
		WorkspaceName: "test-workspace",
		AppPath:       "/tmp/test",
		WorkingDir:    "/workspace",
		Command:       []string{"/bin/sh"},
		Env: map[string]string{
			"TEST": "value",
		},
	}

	if opts.ImageName != "alpine:latest" {
		t.Errorf("Expected ImageName 'alpine:latest', got '%s'", opts.ImageName)
	}

	if opts.Env["TEST"] != "value" {
		t.Errorf("Expected Env TEST=value, got TEST=%s", opts.Env["TEST"])
	}
}

// TestWorkspaceLifecycle tests full container lifecycle
func TestWorkspaceLifecycle(t *testing.T) {
	platform := requireContainerdPlatform(t)
	requireAlpineImage(t, platform)

	runtime, err := NewContainerdRuntimeV2WithPlatform(platform)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create temporary app directory
	appPath := t.TempDir()

	workspaceName := "test-workspace-lifecycle"

	// Test 1: Start workspace
	t.Run("StartWorkspace", func(t *testing.T) {
		containerID, err := runtime.StartWorkspace(ctx, StartOptions{
			ImageName:     "alpine:latest",
			WorkspaceName: workspaceName,
			AppPath:       appPath,
			WorkingDir:    "/workspace",
			Command:       []string{"/bin/sh", "-c", "sleep 30"},
			Env: map[string]string{
				"TEST_ENV": "test_value",
			},
		})
		if err != nil {
			t.Fatalf("Failed to start workspace: %v", err)
		}

		if containerID != workspaceName {
			t.Errorf("Expected containerID '%s', got '%s'", workspaceName, containerID)
		}

		// Give container time to start
		time.Sleep(2 * time.Second)
	})

	// Test 2: Get workspace status
	t.Run("GetWorkspaceStatus", func(t *testing.T) {
		status, err := runtime.GetWorkspaceStatus(ctx, workspaceName)
		if err != nil {
			t.Fatalf("Failed to get workspace status: %v", err)
		}

		if status != "running" {
			t.Errorf("Expected status 'running', got '%s'", status)
		}
	})

	// Test 3: Stop workspace
	t.Run("StopWorkspace", func(t *testing.T) {
		err := runtime.StopWorkspace(ctx, workspaceName)
		if err != nil {
			t.Fatalf("Failed to stop workspace: %v", err)
		}

		// Give container time to stop
		time.Sleep(2 * time.Second)

		// Verify it stopped
		status, err := runtime.GetWorkspaceStatus(ctx, workspaceName)
		if err != nil {
			t.Fatalf("Failed to get workspace status after stop: %v", err)
		}

		if status == "running" {
			t.Errorf("Expected container to be stopped, got status '%s'", status)
		}
	})

	// Test 4: Restart same workspace (should reuse name)
	t.Run("RestartWorkspace", func(t *testing.T) {
		containerID, err := runtime.StartWorkspace(ctx, StartOptions{
			ImageName:     "alpine:latest",
			WorkspaceName: workspaceName,
			AppPath:       appPath,
			WorkingDir:    "/workspace",
			Command:       []string{"/bin/sh", "-c", "sleep 10"},
		})
		if err != nil {
			t.Fatalf("Failed to restart workspace: %v", err)
		}

		if containerID != workspaceName {
			t.Errorf("Expected containerID '%s', got '%s'", workspaceName, containerID)
		}

		time.Sleep(2 * time.Second)

		status, err := runtime.GetWorkspaceStatus(ctx, workspaceName)
		if err != nil {
			t.Fatalf("Failed to get restarted workspace status: %v", err)
		}

		if status != "running" {
			t.Errorf("Expected restarted container to be running, got '%s'", status)
		}

		// Cleanup
		runtime.StopWorkspace(ctx, workspaceName)
	})
}

// TestStartWorkspaceWithMounts tests workspace with volume mounts
func TestStartWorkspaceWithMounts(t *testing.T) {
	platform := requireContainerdPlatform(t)
	requireAlpineImage(t, platform)

	runtime, err := NewContainerdRuntimeV2WithPlatform(platform)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create temporary app directory with a test file
	appPath := t.TempDir()
	testFile := filepath.Join(appPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello from host"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	workspaceName := "test-workspace-mounts"

	// Start workspace
	containerID, err := runtime.StartWorkspace(ctx, StartOptions{
		ImageName:     "alpine:latest",
		WorkspaceName: workspaceName,
		AppPath:       appPath,
		WorkingDir:    "/workspace",
		Command:       []string{"/bin/sh", "-c", "sleep 20"},
	})
	if err != nil {
		t.Fatalf("Failed to start workspace: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Verify container is running
	status, err := runtime.GetWorkspaceStatus(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status != "running" {
		t.Errorf("Expected running status, got '%s'", status)
	}

	// Cleanup
	runtime.StopWorkspace(ctx, containerID)
}

// TestGetWorkspaceStatusNotFound tests status of non-existent container
func TestGetWorkspaceStatusNotFound(t *testing.T) {
	platform := requireContainerdPlatform(t)

	runtime, err := NewContainerdRuntimeV2WithPlatform(platform)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx := context.Background()

	status, err := runtime.GetWorkspaceStatus(ctx, "non-existent-container")
	if err != nil {
		t.Fatalf("Expected no error for non-existent container, got: %v", err)
	}

	if status != "not_found" {
		t.Errorf("Expected status 'not_found', got '%s'", status)
	}
}

// TestBuildImageNotImplemented tests that BuildImage returns error
func TestBuildImageNotImplemented(t *testing.T) {
	runtime := &ContainerdRuntimeV2{}

	err := runtime.BuildImage(context.Background(), BuildOptions{})
	if err == nil {
		t.Error("Expected BuildImage to return error")
	}

	if err.Error() != "use 'dvm build' command which uses BuildKit API" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestStartWorkspaceWithEnvVars tests environment variable passing
func TestStartWorkspaceWithEnvVars(t *testing.T) {
	platform := requireContainerdPlatform(t)
	requireAlpineImage(t, platform)

	runtime, err := NewContainerdRuntimeV2WithPlatform(platform)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	appPath := t.TempDir()
	workspaceName := "test-workspace-env"

	// Start workspace with custom env vars
	containerID, err := runtime.StartWorkspace(ctx, StartOptions{
		ImageName:     "alpine:latest",
		WorkspaceName: workspaceName,
		AppPath:       appPath,
		WorkingDir:    "/workspace",
		Command:       []string{"/bin/sh", "-c", "sleep 15"},
		Env: map[string]string{
			"CUSTOM_VAR": "custom_value",
			"DVM_APP":    "test-app",
		},
	})
	if err != nil {
		t.Fatalf("Failed to start workspace: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Verify running
	status, err := runtime.GetWorkspaceStatus(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status != "running" {
		t.Errorf("Expected running status, got '%s'", status)
	}

	// Cleanup
	runtime.StopWorkspace(ctx, containerID)
}
