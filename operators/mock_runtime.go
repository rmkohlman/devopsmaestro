package operators

import (
	"context"
	"fmt"
	"sync"
)

// MockContainerRuntime implements ContainerRuntime for testing
// It provides:
//   - In-memory workspace tracking
//   - Call recording for verification
//   - Configurable error injection
//   - Status simulation
type MockContainerRuntime struct {
	mu sync.RWMutex

	// Workspaces tracks workspaces and their states
	// Key: workspaceName, Value: status (running, stopped, etc.)
	Workspaces map[string]string

	// Images tracks built images
	// Key: imageName, Value: true if built
	Images map[string]bool

	// Calls records all method calls for verification
	Calls []MockRuntimeCall

	// Error injection
	BuildImageError        error
	StartWorkspaceError    error
	AttachToWorkspaceError error
	StopWorkspaceError     error
	GetStatusError         error

	// Behavior configuration
	RuntimeType string
}

// MockRuntimeCall records a single method call
type MockRuntimeCall struct {
	Method string
	Args   []interface{}
}

// NewMockContainerRuntime creates a new mock runtime with default settings
func NewMockContainerRuntime() *MockContainerRuntime {
	return &MockContainerRuntime{
		Workspaces:  make(map[string]string),
		Images:      make(map[string]bool),
		Calls:       make([]MockRuntimeCall, 0),
		RuntimeType: "mock",
	}
}

// BuildImage simulates building an image
func (m *MockContainerRuntime) BuildImage(ctx context.Context, opts BuildOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockRuntimeCall{
		Method: "BuildImage",
		Args:   []interface{}{opts},
	})

	if m.BuildImageError != nil {
		return m.BuildImageError
	}

	// Mark image as built
	m.Images[opts.ImageName] = true
	for _, tag := range opts.Tags {
		m.Images[tag] = true
	}

	return nil
}

// StartWorkspace simulates starting a workspace
func (m *MockContainerRuntime) StartWorkspace(ctx context.Context, opts StartOptions) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockRuntimeCall{
		Method: "StartWorkspace",
		Args:   []interface{}{opts},
	})

	if m.StartWorkspaceError != nil {
		return "", m.StartWorkspaceError
	}

	// Check if image exists (optional validation)
	if len(m.Images) > 0 && !m.Images[opts.ImageName] {
		return "", fmt.Errorf("image not found: %s", opts.ImageName)
	}

	// Mark workspace as running
	m.Workspaces[opts.WorkspaceName] = "running"

	// Return a mock container ID
	containerID := fmt.Sprintf("mock-container-%s", opts.WorkspaceName)
	return containerID, nil
}

// AttachToWorkspace simulates attaching to a workspace
func (m *MockContainerRuntime) AttachToWorkspace(ctx context.Context, workspaceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockRuntimeCall{
		Method: "AttachToWorkspace",
		Args:   []interface{}{workspaceID},
	})

	if m.AttachToWorkspaceError != nil {
		return m.AttachToWorkspaceError
	}

	// Check workspace exists and is running
	status, exists := m.Workspaces[workspaceID]
	if !exists {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}
	if status != "running" {
		return fmt.Errorf("workspace not running: %s (status: %s)", workspaceID, status)
	}

	return nil
}

// StopWorkspace simulates stopping a workspace
func (m *MockContainerRuntime) StopWorkspace(ctx context.Context, workspaceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockRuntimeCall{
		Method: "StopWorkspace",
		Args:   []interface{}{workspaceID},
	})

	if m.StopWorkspaceError != nil {
		return m.StopWorkspaceError
	}

	// Check workspace exists
	if _, exists := m.Workspaces[workspaceID]; !exists {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}

	// Mark workspace as stopped
	m.Workspaces[workspaceID] = "stopped"
	return nil
}

// GetWorkspaceStatus returns the status of a workspace
func (m *MockContainerRuntime) GetWorkspaceStatus(ctx context.Context, workspaceID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockRuntimeCall{
		Method: "GetWorkspaceStatus",
		Args:   []interface{}{workspaceID},
	})

	if m.GetStatusError != nil {
		return "", m.GetStatusError
	}

	status, exists := m.Workspaces[workspaceID]
	if !exists {
		return "not_found", nil
	}

	return status, nil
}

// GetRuntimeType returns the mock runtime type
func (m *MockContainerRuntime) GetRuntimeType() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.RuntimeType
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockContainerRuntime) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Workspaces = make(map[string]string)
	m.Images = make(map[string]bool)
	m.Calls = make([]MockRuntimeCall, 0)
	m.BuildImageError = nil
	m.StartWorkspaceError = nil
	m.AttachToWorkspaceError = nil
	m.StopWorkspaceError = nil
	m.GetStatusError = nil
}

// CallCount returns the number of times a method was called
func (m *MockContainerRuntime) CallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// GetCalls returns all calls to a specific method
func (m *MockContainerRuntime) GetCalls(method string) []MockRuntimeCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MockRuntimeCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// LastCall returns the last call made (for any method)
func (m *MockContainerRuntime) LastCall() *MockRuntimeCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}

// SetWorkspaceStatus manually sets a workspace status (for test setup)
func (m *MockContainerRuntime) SetWorkspaceStatus(workspaceID, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Workspaces[workspaceID] = status
}

// AddImage marks an image as available (for test setup)
func (m *MockContainerRuntime) AddImage(imageName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Images[imageName] = true
}

// Ensure MockContainerRuntime implements ContainerRuntime
var _ ContainerRuntime = (*MockContainerRuntime)(nil)
