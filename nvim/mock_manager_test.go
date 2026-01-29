package nvim

import (
	"errors"
	"testing"
	"time"
)

// =============================================================================
// Mock Implementation Tests
// =============================================================================

func TestMockManager_ImplementsInterface(t *testing.T) {
	// Compile-time check that MockManager implements Manager
	var _ Manager = (*MockManager)(nil)
}

func TestMockManager_Init(t *testing.T) {
	mock := NewMockManager()

	opts := InitOptions{
		ConfigPath: "/test/config",
		Template:   "kickstart",
		Overwrite:  true,
	}

	err := mock.Init(opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if !mock.Initialized {
		t.Error("Init() should set Initialized to true")
	}

	status, _ := mock.Status()
	if !status.Exists {
		t.Error("Init() should set Exists to true")
	}
	if status.Template != "kickstart" {
		t.Errorf("Template = %q, want 'kickstart'", status.Template)
	}
	if status.ConfigPath != "/test/config" {
		t.Errorf("ConfigPath = %q, want '/test/config'", status.ConfigPath)
	}

	if mock.CallCount("Init") != 1 {
		t.Errorf("CallCount(Init) = %d, want 1", mock.CallCount("Init"))
	}
}

func TestMockManager_Init_Error(t *testing.T) {
	mock := NewMockManager()
	mock.InitError = errors.New("config already exists")

	err := mock.Init(InitOptions{Template: "minimal"})
	if err == nil {
		t.Error("Init() expected error, got nil")
	}
	if err.Error() != "config already exists" {
		t.Errorf("Init() error = %v, want 'config already exists'", err)
	}
}

func TestMockManager_Sync(t *testing.T) {
	mock := NewMockManager()
	mock.SetInitialized(true)

	err := mock.Sync("workspace-1", SyncPull)
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	status, _ := mock.Status()
	if status.SyncedWith != "workspace-1" {
		t.Errorf("SyncedWith = %q, want 'workspace-1'", status.SyncedWith)
	}

	if mock.CallCount("Sync") != 1 {
		t.Errorf("CallCount(Sync) = %d, want 1", mock.CallCount("Sync"))
	}

	// Verify call args
	calls := mock.GetCalls("Sync")
	if len(calls) != 1 {
		t.Fatal("Expected 1 Sync call")
	}
	if calls[0].Args[0] != "workspace-1" {
		t.Errorf("Sync args[0] = %v, want 'workspace-1'", calls[0].Args[0])
	}
	if calls[0].Args[1] != SyncPull {
		t.Errorf("Sync args[1] = %v, want SyncPull", calls[0].Args[1])
	}
}

func TestMockManager_Sync_Error(t *testing.T) {
	mock := NewMockManager()
	mock.SyncError = errors.New("connection failed")

	err := mock.Sync("workspace-1", SyncPush)
	if err == nil {
		t.Error("Sync() expected error, got nil")
	}
}

func TestMockManager_Push(t *testing.T) {
	mock := NewMockManager()

	err := mock.Push("workspace-2")
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	status, _ := mock.Status()
	if status.SyncedWith != "workspace-2" {
		t.Errorf("SyncedWith = %q, want 'workspace-2'", status.SyncedWith)
	}

	if mock.CallCount("Push") != 1 {
		t.Errorf("CallCount(Push) = %d, want 1", mock.CallCount("Push"))
	}
}

func TestMockManager_Push_Error(t *testing.T) {
	mock := NewMockManager()
	mock.PushError = errors.New("permission denied")

	err := mock.Push("workspace-1")
	if err == nil {
		t.Error("Push() expected error, got nil")
	}
}

func TestMockManager_Status(t *testing.T) {
	mock := NewMockManager()
	mock.SetStatus(&Status{
		ConfigPath:   "/custom/path",
		Exists:       true,
		LastSync:     time.Now(),
		SyncedWith:   "my-workspace",
		LocalChanges: true,
		Template:     "lazyvim",
	})

	status, err := mock.Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if status.ConfigPath != "/custom/path" {
		t.Errorf("ConfigPath = %q, want '/custom/path'", status.ConfigPath)
	}
	if !status.Exists {
		t.Error("Exists should be true")
	}
	if status.SyncedWith != "my-workspace" {
		t.Errorf("SyncedWith = %q, want 'my-workspace'", status.SyncedWith)
	}
	if !status.LocalChanges {
		t.Error("LocalChanges should be true")
	}
}

func TestMockManager_Status_Error(t *testing.T) {
	mock := NewMockManager()
	mock.StatusError = errors.New("failed to read status")

	_, err := mock.Status()
	if err == nil {
		t.Error("Status() expected error, got nil")
	}
}

func TestMockManager_ListWorkspaces(t *testing.T) {
	mock := NewMockManager()
	mock.SetWorkspaces([]Workspace{
		{ID: "ws-1", Name: "project-1", Active: true},
		{ID: "ws-2", Name: "project-2", Active: false},
	})

	workspaces, err := mock.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces() error = %v", err)
	}

	if len(workspaces) != 2 {
		t.Errorf("ListWorkspaces() returned %d workspaces, want 2", len(workspaces))
	}

	if workspaces[0].Name != "project-1" {
		t.Errorf("First workspace name = %q, want 'project-1'", workspaces[0].Name)
	}
}

func TestMockManager_ListWorkspaces_Error(t *testing.T) {
	mock := NewMockManager()
	mock.ListWorkspacesError = errors.New("database error")

	_, err := mock.ListWorkspaces()
	if err == nil {
		t.Error("ListWorkspaces() expected error, got nil")
	}
}

func TestMockManager_AddWorkspace(t *testing.T) {
	mock := NewMockManager()

	mock.AddWorkspace(Workspace{ID: "ws-1", Name: "test"})
	mock.AddWorkspace(Workspace{ID: "ws-2", Name: "test2"})

	workspaces, _ := mock.ListWorkspaces()
	if len(workspaces) != 2 {
		t.Errorf("Expected 2 workspaces, got %d", len(workspaces))
	}
}

func TestMockManager_Reset(t *testing.T) {
	mock := NewMockManager()

	// Add state
	mock.Init(InitOptions{Template: "minimal"})
	mock.AddWorkspace(Workspace{ID: "ws-1"})
	mock.InitError = errors.New("error")

	// Reset
	mock.Reset()

	// Verify cleared
	if mock.Initialized {
		t.Error("Reset() should clear Initialized")
	}
	if len(mock.Calls) != 0 {
		t.Error("Reset() should clear Calls")
	}
	if len(mock.Workspaces) != 0 {
		t.Error("Reset() should clear Workspaces")
	}
	if mock.InitError != nil {
		t.Error("Reset() should clear InitError")
	}

	status, _ := mock.Status()
	if status.Exists {
		t.Error("Reset() should clear status.Exists")
	}
}

func TestMockManager_CallTracking(t *testing.T) {
	mock := NewMockManager()

	mock.Init(InitOptions{Template: "kickstart"})
	mock.Status()
	mock.Status()
	mock.Push("ws-1")

	// Test CallCount
	if mock.CallCount("Init") != 1 {
		t.Errorf("CallCount(Init) = %d, want 1", mock.CallCount("Init"))
	}
	if mock.CallCount("Status") != 2 {
		t.Errorf("CallCount(Status) = %d, want 2", mock.CallCount("Status"))
	}

	// Test GetCalls
	statusCalls := mock.GetCalls("Status")
	if len(statusCalls) != 2 {
		t.Errorf("GetCalls(Status) returned %d calls, want 2", len(statusCalls))
	}

	// Test LastCall
	lastCall := mock.LastCall()
	if lastCall == nil {
		t.Fatal("LastCall() returned nil")
	}
	if lastCall.Method != "Push" {
		t.Errorf("LastCall().Method = %q, want 'Push'", lastCall.Method)
	}
}

func TestMockManager_SimulateChanges(t *testing.T) {
	mock := NewMockManager()

	// Initially no changes
	status, _ := mock.Status()
	if status.LocalChanges {
		t.Error("LocalChanges should be false initially")
	}

	// Simulate local changes
	mock.SimulateLocalChanges()
	status, _ = mock.Status()
	if !status.LocalChanges {
		t.Error("LocalChanges should be true after SimulateLocalChanges()")
	}

	// Simulate remote changes
	mock.SimulateRemoteChanges()
	status, _ = mock.Status()
	if !status.RemoteChanges {
		t.Error("RemoteChanges should be true after SimulateRemoteChanges()")
	}
}

func TestMockManager_InjectError(t *testing.T) {
	mock := NewMockManager()

	mock.InjectError("Init", errors.New("init error"))
	mock.InjectError("Status", errors.New("status error"))

	err := mock.Init(InitOptions{})
	if err == nil || err.Error() != "init error" {
		t.Errorf("Init() error = %v, want 'init error'", err)
	}

	_, err = mock.Status()
	if err == nil || err.Error() != "status error" {
		t.Errorf("Status() error = %v, want 'status error'", err)
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestManager_Interface_Mock(t *testing.T) {
	// Use mock through the interface
	var mgr Manager = NewMockManager()

	// Test all interface methods work
	err := mgr.Init(InitOptions{Template: "minimal"})
	if err != nil {
		t.Errorf("Interface.Init() error = %v", err)
	}

	_, err = mgr.Status()
	if err != nil {
		t.Errorf("Interface.Status() error = %v", err)
	}

	err = mgr.Sync("workspace", SyncPull)
	if err != nil {
		t.Errorf("Interface.Sync() error = %v", err)
	}

	err = mgr.Push("workspace")
	if err != nil {
		t.Errorf("Interface.Push() error = %v", err)
	}

	_, err = mgr.ListWorkspaces()
	if err != nil {
		t.Errorf("Interface.ListWorkspaces() error = %v", err)
	}
}

// TestManager_Swappability verifies different manager implementations
// can be swapped through the interface
func TestManager_Swappability(t *testing.T) {
	testWorkflow := func(mgr Manager, t *testing.T) {
		// Get initial status
		status, err := mgr.Status()
		if err != nil {
			t.Errorf("Status() error = %v", err)
			return
		}
		if status == nil {
			t.Error("Status() returned nil")
			return
		}

		// List workspaces
		_, err = mgr.ListWorkspaces()
		if err != nil {
			t.Errorf("ListWorkspaces() error = %v", err)
		}
	}

	t.Run("MockManager", func(t *testing.T) {
		mock := NewMockManager()
		testWorkflow(mock, t)

		// Verify mock captured calls
		if mock.CallCount("Status") != 1 {
			t.Error("Mock should capture Status call")
		}
	})

	// Real manager would require filesystem, so we skip it in unit tests
	// t.Run("RealManager", func(t *testing.T) { ... })
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestMockManager_ConcurrentAccess(t *testing.T) {
	mock := NewMockManager()

	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func() {
			mock.Status()
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		go func() {
			mock.ListWorkspaces()
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify calls were recorded
	if len(mock.Calls) == 0 {
		t.Error("No calls recorded during concurrent access")
	}
}
