package nvim

import (
	"fmt"
	"sync"
	"time"
)

// MockManager implements Manager interface for testing
// It provides:
//   - In-memory state tracking
//   - Call recording for verification
//   - Configurable error injection
//   - Pre-configured workspaces and status
type MockManager struct {
	mu sync.RWMutex

	// State
	Initialized bool
	StatusData  *Status
	Workspaces  []Workspace

	// Call recording
	Calls []MockManagerCall

	// Error injection
	InitError           error
	SyncError           error
	PushError           error
	StatusError         error
	ListWorkspacesError error
}

// MockManagerCall records a single method call
type MockManagerCall struct {
	Method string
	Args   []interface{}
}

// NewMockManager creates a new mock manager with default settings
func NewMockManager() *MockManager {
	return &MockManager{
		StatusData: &Status{
			ConfigPath: "~/.config/nvim",
			Exists:     false,
		},
		Workspaces: make([]Workspace, 0),
		Calls:      make([]MockManagerCall, 0),
	}
}

// Init simulates initializing Neovim configuration
func (m *MockManager) Init(opts InitOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Init",
		Args:   []interface{}{opts},
	})

	if m.InitError != nil {
		return m.InitError
	}

	m.Initialized = true
	m.StatusData.Exists = true
	m.StatusData.LastSync = time.Now()
	m.StatusData.Template = opts.Template
	if opts.ConfigPath != "" {
		m.StatusData.ConfigPath = opts.ConfigPath
	}

	return nil
}

// Sync simulates syncing configuration with a workspace
func (m *MockManager) Sync(workspace string, direction SyncDirection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Sync",
		Args:   []interface{}{workspace, direction},
	})

	if m.SyncError != nil {
		return m.SyncError
	}

	m.StatusData.LastSync = time.Now()
	m.StatusData.SyncedWith = workspace
	m.StatusData.LocalChanges = false

	return nil
}

// Push simulates pushing configuration to a workspace
func (m *MockManager) Push(workspace string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Push",
		Args:   []interface{}{workspace},
	})

	if m.PushError != nil {
		return m.PushError
	}

	m.StatusData.LastSync = time.Now()
	m.StatusData.SyncedWith = workspace
	m.StatusData.LocalChanges = false

	return nil
}

// Status returns the current status
func (m *MockManager) Status() (*Status, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Status",
		Args:   nil,
	})

	if m.StatusError != nil {
		return nil, m.StatusError
	}

	// Return a copy to prevent modification
	statusCopy := *m.StatusData
	return &statusCopy, nil
}

// ListWorkspaces returns the list of available workspaces
func (m *MockManager) ListWorkspaces() ([]Workspace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "ListWorkspaces",
		Args:   nil,
	})

	if m.ListWorkspacesError != nil {
		return nil, m.ListWorkspacesError
	}

	// Return a copy
	workspaces := make([]Workspace, len(m.Workspaces))
	copy(workspaces, m.Workspaces)
	return workspaces, nil
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Initialized = false
	m.StatusData = &Status{
		ConfigPath: "~/.config/nvim",
		Exists:     false,
	}
	m.Workspaces = make([]Workspace, 0)
	m.Calls = make([]MockManagerCall, 0)
	m.InitError = nil
	m.SyncError = nil
	m.PushError = nil
	m.StatusError = nil
	m.ListWorkspacesError = nil
}

// CallCount returns the number of times a method was called
func (m *MockManager) CallCount(method string) int {
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
func (m *MockManager) GetCalls(method string) []MockManagerCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MockManagerCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// LastCall returns the last call made
func (m *MockManager) LastCall() *MockManagerCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}

// SetStatus sets the status data for testing
func (m *MockManager) SetStatus(status *Status) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StatusData = status
}

// AddWorkspace adds a workspace for testing
func (m *MockManager) AddWorkspace(ws Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Workspaces = append(m.Workspaces, ws)
}

// SetWorkspaces sets the full workspaces list
func (m *MockManager) SetWorkspaces(workspaces []Workspace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Workspaces = workspaces
}

// SimulateLocalChanges marks that local changes exist
func (m *MockManager) SimulateLocalChanges() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StatusData.LocalChanges = true
}

// SimulateRemoteChanges marks that remote changes exist
func (m *MockManager) SimulateRemoteChanges() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StatusData.RemoteChanges = true
}

// SetInitialized sets the initialized state
func (m *MockManager) SetInitialized(initialized bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Initialized = initialized
	m.StatusData.Exists = initialized
}

// InjectError injects an error for a specific method
func (m *MockManager) InjectError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch method {
	case "Init":
		m.InitError = err
	case "Sync":
		m.SyncError = err
	case "Push":
		m.PushError = err
	case "Status":
		m.StatusError = err
	case "ListWorkspaces":
		m.ListWorkspacesError = err
	default:
		panic(fmt.Sprintf("unknown method: %s", method))
	}
}

// Ensure MockManager implements Manager
var _ Manager = (*MockManager)(nil)
