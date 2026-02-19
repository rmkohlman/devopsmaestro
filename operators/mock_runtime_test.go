package operators

import (
	"context"
	"errors"
	"testing"
)

// =============================================================================
// Mock Implementation Tests
// =============================================================================

func TestMockContainerRuntime_ImplementsInterface(t *testing.T) {
	// Compile-time check that MockContainerRuntime implements ContainerRuntime
	var _ ContainerRuntime = (*MockContainerRuntime)(nil)
}

func TestMockContainerRuntime_BuildImage(t *testing.T) {
	mock := NewMockContainerRuntime()

	opts := BuildOptions{
		ImageName:    "test-image:latest",
		AppPath:      "/test/app",
		Tags:         []string{"test-image:v1"},
		Dockerfile:   "Dockerfile",
		BuildContext: "/test/app",
	}

	err := mock.BuildImage(context.Background(), opts)
	if err != nil {
		t.Fatalf("BuildImage() error = %v", err)
	}

	// Verify image was marked as built
	if !mock.Images["test-image:latest"] {
		t.Error("Image 'test-image:latest' not marked as built")
	}
	if !mock.Images["test-image:v1"] {
		t.Error("Tag 'test-image:v1' not marked as built")
	}

	// Verify call was recorded
	if mock.CallCount("BuildImage") != 1 {
		t.Errorf("CallCount(BuildImage) = %d, want 1", mock.CallCount("BuildImage"))
	}
}

func TestMockContainerRuntime_BuildImage_Error(t *testing.T) {
	mock := NewMockContainerRuntime()
	mock.BuildImageError = errors.New("build failed")

	err := mock.BuildImage(context.Background(), BuildOptions{ImageName: "test"})
	if err == nil {
		t.Error("BuildImage() expected error, got nil")
	}
	if err.Error() != "build failed" {
		t.Errorf("BuildImage() error = %v, want 'build failed'", err)
	}
}

func TestMockContainerRuntime_StartWorkspace(t *testing.T) {
	mock := NewMockContainerRuntime()
	mock.AddImage("my-image:latest") // Pre-add image

	opts := StartOptions{
		ImageName:     "my-image:latest",
		WorkspaceName: "test-ws",
		AppPath:       "/app",
	}

	containerID, err := mock.StartWorkspace(context.Background(), opts)
	if err != nil {
		t.Fatalf("StartWorkspace() error = %v", err)
	}

	if containerID == "" {
		t.Error("StartWorkspace() returned empty container ID")
	}

	// Verify workspace status
	status, _ := mock.GetWorkspaceStatus(context.Background(), "test-ws")
	if status != "running" {
		t.Errorf("Workspace status = %q, want 'running'", status)
	}
}

func TestMockContainerRuntime_StartWorkspace_ImageNotFound(t *testing.T) {
	mock := NewMockContainerRuntime()
	mock.AddImage("other-image:latest") // Add a different image

	opts := StartOptions{
		ImageName:     "missing-image:latest",
		WorkspaceName: "test-ws",
	}

	_, err := mock.StartWorkspace(context.Background(), opts)
	if err == nil {
		t.Error("StartWorkspace() expected error for missing image")
	}
}

func TestMockContainerRuntime_AttachToWorkspace(t *testing.T) {
	mock := NewMockContainerRuntime()
	mock.SetWorkspaceStatus("test-ws", "running")

	err := mock.AttachToWorkspace(context.Background(), AttachOptions{WorkspaceID: "test-ws"})
	if err != nil {
		t.Fatalf("AttachToWorkspace() error = %v", err)
	}

	if mock.CallCount("AttachToWorkspace") != 1 {
		t.Errorf("CallCount(AttachToWorkspace) = %d, want 1", mock.CallCount("AttachToWorkspace"))
	}
}

func TestMockContainerRuntime_AttachToWorkspace_NotRunning(t *testing.T) {
	mock := NewMockContainerRuntime()
	mock.SetWorkspaceStatus("test-ws", "stopped")

	err := mock.AttachToWorkspace(context.Background(), AttachOptions{WorkspaceID: "test-ws"})
	if err == nil {
		t.Error("AttachToWorkspace() expected error for stopped workspace")
	}
}

func TestMockContainerRuntime_AttachToWorkspace_NotFound(t *testing.T) {
	mock := NewMockContainerRuntime()

	err := mock.AttachToWorkspace(context.Background(), AttachOptions{WorkspaceID: "nonexistent"})
	if err == nil {
		t.Error("AttachToWorkspace() expected error for nonexistent workspace")
	}
}

func TestMockContainerRuntime_StopWorkspace(t *testing.T) {
	mock := NewMockContainerRuntime()
	mock.SetWorkspaceStatus("test-ws", "running")

	err := mock.StopWorkspace(context.Background(), "test-ws")
	if err != nil {
		t.Fatalf("StopWorkspace() error = %v", err)
	}

	status, _ := mock.GetWorkspaceStatus(context.Background(), "test-ws")
	if status != "stopped" {
		t.Errorf("Workspace status after stop = %q, want 'stopped'", status)
	}
}

func TestMockContainerRuntime_StopWorkspace_NotFound(t *testing.T) {
	mock := NewMockContainerRuntime()

	err := mock.StopWorkspace(context.Background(), "nonexistent")
	if err == nil {
		t.Error("StopWorkspace() expected error for nonexistent workspace")
	}
}

func TestMockContainerRuntime_GetWorkspaceStatus(t *testing.T) {
	mock := NewMockContainerRuntime()

	// Test not found
	status, err := mock.GetWorkspaceStatus(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetWorkspaceStatus() error = %v", err)
	}
	if status != "not_found" {
		t.Errorf("GetWorkspaceStatus() = %q, want 'not_found'", status)
	}

	// Test found
	mock.SetWorkspaceStatus("my-ws", "running")
	status, err = mock.GetWorkspaceStatus(context.Background(), "my-ws")
	if err != nil {
		t.Fatalf("GetWorkspaceStatus() error = %v", err)
	}
	if status != "running" {
		t.Errorf("GetWorkspaceStatus() = %q, want 'running'", status)
	}
}

func TestMockContainerRuntime_GetRuntimeType(t *testing.T) {
	mock := NewMockContainerRuntime()

	if mock.GetRuntimeType() != "mock" {
		t.Errorf("GetRuntimeType() = %q, want 'mock'", mock.GetRuntimeType())
	}

	mock.RuntimeType = "test-runtime"
	if mock.GetRuntimeType() != "test-runtime" {
		t.Errorf("GetRuntimeType() = %q, want 'test-runtime'", mock.GetRuntimeType())
	}
}

func TestMockContainerRuntime_Reset(t *testing.T) {
	mock := NewMockContainerRuntime()

	// Add some state
	mock.AddImage("test-image")
	mock.SetWorkspaceStatus("test-ws", "running")
	mock.BuildImage(context.Background(), BuildOptions{ImageName: "img"})
	mock.BuildImageError = errors.New("error")

	// Reset
	mock.Reset()

	// Verify all state cleared
	if len(mock.Images) != 0 {
		t.Error("Reset() should clear Images")
	}
	if len(mock.Workspaces) != 0 {
		t.Error("Reset() should clear Workspaces")
	}
	if len(mock.Calls) != 0 {
		t.Error("Reset() should clear Calls")
	}
	if mock.BuildImageError != nil {
		t.Error("Reset() should clear BuildImageError")
	}
}

func TestMockContainerRuntime_CallTracking(t *testing.T) {
	mock := NewMockContainerRuntime()

	// Make several calls
	mock.BuildImage(context.Background(), BuildOptions{ImageName: "img1"})
	mock.BuildImage(context.Background(), BuildOptions{ImageName: "img2"})
	mock.SetWorkspaceStatus("ws1", "running")
	mock.GetWorkspaceStatus(context.Background(), "ws1")

	// Test CallCount
	if mock.CallCount("BuildImage") != 2 {
		t.Errorf("CallCount(BuildImage) = %d, want 2", mock.CallCount("BuildImage"))
	}
	if mock.CallCount("GetWorkspaceStatus") != 1 {
		t.Errorf("CallCount(GetWorkspaceStatus) = %d, want 1", mock.CallCount("GetWorkspaceStatus"))
	}

	// Test GetCalls
	buildCalls := mock.GetCalls("BuildImage")
	if len(buildCalls) != 2 {
		t.Errorf("GetCalls(BuildImage) returned %d calls, want 2", len(buildCalls))
	}

	// Test LastCall
	lastCall := mock.LastCall()
	if lastCall == nil {
		t.Fatal("LastCall() returned nil")
	}
	if lastCall.Method != "GetWorkspaceStatus" {
		t.Errorf("LastCall().Method = %q, want 'GetWorkspaceStatus'", lastCall.Method)
	}
}

// =============================================================================
// Interface Compliance Tests - Test through the interface
// =============================================================================

func TestContainerRuntime_Interface_Mock(t *testing.T) {
	// Create mock and use it through the interface
	var runtime ContainerRuntime = NewMockContainerRuntime()

	// Test all interface methods work through the interface type
	ctx := context.Background()

	// BuildImage
	err := runtime.BuildImage(ctx, BuildOptions{ImageName: "test:latest"})
	if err != nil {
		t.Errorf("Interface.BuildImage() error = %v", err)
	}

	// StartWorkspace - need to cast back to mock to add image first
	mock := runtime.(*MockContainerRuntime)
	mock.AddImage("workspace-image:latest")

	containerID, err := runtime.StartWorkspace(ctx, StartOptions{
		ImageName:     "workspace-image:latest",
		WorkspaceName: "my-workspace",
	})
	if err != nil {
		t.Errorf("Interface.StartWorkspace() error = %v", err)
	}
	if containerID == "" {
		t.Error("Interface.StartWorkspace() returned empty containerID")
	}

	// GetWorkspaceStatus
	status, err := runtime.GetWorkspaceStatus(ctx, "my-workspace")
	if err != nil {
		t.Errorf("Interface.GetWorkspaceStatus() error = %v", err)
	}
	if status != "running" {
		t.Errorf("Interface.GetWorkspaceStatus() = %q, want 'running'", status)
	}

	// AttachToWorkspace
	err = runtime.AttachToWorkspace(ctx, AttachOptions{WorkspaceID: "my-workspace"})
	if err != nil {
		t.Errorf("Interface.AttachToWorkspace() error = %v", err)
	}

	// StopWorkspace
	err = runtime.StopWorkspace(ctx, "my-workspace")
	if err != nil {
		t.Errorf("Interface.StopWorkspace() error = %v", err)
	}

	// GetRuntimeType
	runtimeType := runtime.GetRuntimeType()
	if runtimeType != "mock" {
		t.Errorf("Interface.GetRuntimeType() = %q, want 'mock'", runtimeType)
	}
}

// TestContainerRuntime_Swappability verifies that different implementations
// can be swapped through the interface
func TestContainerRuntime_Swappability(t *testing.T) {
	// This test demonstrates that code can be written to the interface
	// and work with any implementation

	testWorkflow := func(runtime ContainerRuntime, t *testing.T) {
		ctx := context.Background()

		// Build an image
		err := runtime.BuildImage(ctx, BuildOptions{
			ImageName:    "swappable-test:latest",
			AppPath:      "/test",
			BuildContext: "/test",
		})
		if err != nil {
			t.Errorf("BuildImage() error = %v", err)
			return
		}

		// Verify runtime type is accessible
		runtimeType := runtime.GetRuntimeType()
		if runtimeType == "" {
			t.Error("GetRuntimeType() returned empty string")
		}
	}

	t.Run("MockRuntime", func(t *testing.T) {
		mock := NewMockContainerRuntime()
		testWorkflow(mock, t)
	})

	// Additional implementations could be tested here:
	// t.Run("DockerRuntime", func(t *testing.T) { ... })
	// t.Run("ContainerdRuntime", func(t *testing.T) { ... })
}

// =============================================================================
// Error Injection Tests
// =============================================================================

func TestMockContainerRuntime_ErrorInjection(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockContainerRuntime)
		operation   func(*MockContainerRuntime) error
		wantErr     bool
		errContains string
	}{
		{
			name: "BuildImage error",
			setup: func(m *MockContainerRuntime) {
				m.BuildImageError = errors.New("docker daemon not running")
			},
			operation: func(m *MockContainerRuntime) error {
				return m.BuildImage(context.Background(), BuildOptions{})
			},
			wantErr:     true,
			errContains: "docker daemon not running",
		},
		{
			name: "StartWorkspace error",
			setup: func(m *MockContainerRuntime) {
				m.StartWorkspaceError = errors.New("out of memory")
			},
			operation: func(m *MockContainerRuntime) error {
				_, err := m.StartWorkspace(context.Background(), StartOptions{})
				return err
			},
			wantErr:     true,
			errContains: "out of memory",
		},
		{
			name: "AttachToWorkspace error",
			setup: func(m *MockContainerRuntime) {
				m.SetWorkspaceStatus("test", "running")
				m.AttachToWorkspaceError = errors.New("tty allocation failed")
			},
			operation: func(m *MockContainerRuntime) error {
				return m.AttachToWorkspace(context.Background(), AttachOptions{WorkspaceID: "test"})
			},
			wantErr:     true,
			errContains: "tty allocation failed",
		},
		{
			name: "StopWorkspace error",
			setup: func(m *MockContainerRuntime) {
				m.SetWorkspaceStatus("test", "running")
				m.StopWorkspaceError = errors.New("container busy")
			},
			operation: func(m *MockContainerRuntime) error {
				return m.StopWorkspace(context.Background(), "test")
			},
			wantErr:     true,
			errContains: "container busy",
		},
		{
			name: "GetWorkspaceStatus error",
			setup: func(m *MockContainerRuntime) {
				m.GetStatusError = errors.New("connection refused")
			},
			operation: func(m *MockContainerRuntime) error {
				_, err := m.GetWorkspaceStatus(context.Background(), "test")
				return err
			},
			wantErr:     true,
			errContains: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockContainerRuntime()
			tt.setup(mock)

			err := tt.operation(mock)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error containing %q, got nil", tt.errContains)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if err.Error() != tt.errContains {
					t.Errorf("Error = %q, want %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestMockContainerRuntime_ConcurrentAccess(t *testing.T) {
	mock := NewMockContainerRuntime()
	ctx := context.Background()

	// Run multiple goroutines accessing the mock concurrently
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			mock.BuildImage(ctx, BuildOptions{ImageName: "img"})
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func(id int) {
			mock.GetWorkspaceStatus(ctx, "ws")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify calls were recorded (exact count may vary due to race conditions in call recording)
	if len(mock.Calls) == 0 {
		t.Error("No calls recorded during concurrent access")
	}
}
