package operators

import "sync"

// MockContextManager implements ContextManager for testing.
// It provides:
//   - In-memory context state
//   - Call recording for verification
//   - Per-method error injection for testing error paths
type MockContextManager struct {
	mu sync.RWMutex

	// Configurable return values
	ActiveAppResult       string
	ActiveWorkspaceResult string
	ContextSummaryResult  string
	LoadContextResult     *ContextConfig

	// Error injection (per-method for independent error testing)
	GetActiveAppError       error
	GetActiveWorkspaceError error
	SetAppError             error
	SetWorkspaceError       error
	ClearAppError           error
	ClearWorkspaceError     error
	LoadContextError        error
	SaveContextError        error
	GetContextSummaryError  error

	// Call recording
	Calls []MockContextManagerCall

	// Argument recording
	SetAppArgs       []string
	SetWorkspaceArgs []string
	SaveContextArgs  []*ContextConfig
}

// MockContextManagerCall records a single method call
type MockContextManagerCall struct {
	Method string
	Args   []interface{}
}

// Compile-time interface compliance check
var _ ContextManager = (*MockContextManager)(nil)

// NewMockContextManager creates a new mock context manager with default settings
func NewMockContextManager() *MockContextManager {
	return &MockContextManager{
		Calls:             make([]MockContextManagerCall, 0),
		SetAppArgs:        make([]string, 0),
		SetWorkspaceArgs:  make([]string, 0),
		SaveContextArgs:   make([]*ContextConfig, 0),
		LoadContextResult: &ContextConfig{},
	}
}

// GetActiveApp returns the configured active app or error
func (m *MockContextManager) GetActiveApp() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "GetActiveApp",
	})

	if m.GetActiveAppError != nil {
		return "", m.GetActiveAppError
	}

	return m.ActiveAppResult, nil
}

// GetActiveWorkspace returns the configured active workspace or error
func (m *MockContextManager) GetActiveWorkspace() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "GetActiveWorkspace",
	})

	if m.GetActiveWorkspaceError != nil {
		return "", m.GetActiveWorkspaceError
	}

	return m.ActiveWorkspaceResult, nil
}

// SetApp records the call and returns the configured error
func (m *MockContextManager) SetApp(appName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "SetApp",
		Args:   []interface{}{appName},
	})
	m.SetAppArgs = append(m.SetAppArgs, appName)

	if m.SetAppError != nil {
		return m.SetAppError
	}

	// Update internal state for consistency
	m.ActiveAppResult = appName
	m.ActiveWorkspaceResult = "" // Clear workspace when switching apps

	return nil
}

// SetWorkspace records the call and returns the configured error
func (m *MockContextManager) SetWorkspace(workspaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "SetWorkspace",
		Args:   []interface{}{workspaceName},
	})
	m.SetWorkspaceArgs = append(m.SetWorkspaceArgs, workspaceName)

	if m.SetWorkspaceError != nil {
		return m.SetWorkspaceError
	}

	// Update internal state for consistency
	m.ActiveWorkspaceResult = workspaceName

	return nil
}

// ClearApp records the call and returns the configured error
func (m *MockContextManager) ClearApp() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "ClearApp",
	})

	if m.ClearAppError != nil {
		return m.ClearAppError
	}

	// Update internal state for consistency
	m.ActiveAppResult = ""
	m.ActiveWorkspaceResult = ""

	return nil
}

// ClearWorkspace records the call and returns the configured error
func (m *MockContextManager) ClearWorkspace() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "ClearWorkspace",
	})

	if m.ClearWorkspaceError != nil {
		return m.ClearWorkspaceError
	}

	// Update internal state for consistency
	m.ActiveWorkspaceResult = ""

	return nil
}

// LoadContext returns the configured context or error
func (m *MockContextManager) LoadContext() (*ContextConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "LoadContext",
	})

	if m.LoadContextError != nil {
		return nil, m.LoadContextError
	}

	return m.LoadContextResult, nil
}

// SaveContext records the call and returns the configured error
func (m *MockContextManager) SaveContext(ctx *ContextConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "SaveContext",
		Args:   []interface{}{ctx},
	})
	m.SaveContextArgs = append(m.SaveContextArgs, ctx)

	if m.SaveContextError != nil {
		return m.SaveContextError
	}

	// Update internal state for consistency
	m.LoadContextResult = ctx

	return nil
}

// GetContextSummary returns the configured summary or error
func (m *MockContextManager) GetContextSummary() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockContextManagerCall{
		Method: "GetContextSummary",
	})

	if m.GetContextSummaryError != nil {
		return "", m.GetContextSummaryError
	}

	return m.ContextSummaryResult, nil
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockContextManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ActiveAppResult = ""
	m.ActiveWorkspaceResult = ""
	m.ContextSummaryResult = ""
	m.LoadContextResult = &ContextConfig{}
	m.GetActiveAppError = nil
	m.GetActiveWorkspaceError = nil
	m.SetAppError = nil
	m.SetWorkspaceError = nil
	m.ClearAppError = nil
	m.ClearWorkspaceError = nil
	m.LoadContextError = nil
	m.SaveContextError = nil
	m.GetContextSummaryError = nil
	m.Calls = make([]MockContextManagerCall, 0)
	m.SetAppArgs = make([]string, 0)
	m.SetWorkspaceArgs = make([]string, 0)
	m.SaveContextArgs = make([]*ContextConfig, 0)
}

// CallCount returns the number of times a method was called
func (m *MockContextManager) CallCount(method string) int {
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
func (m *MockContextManager) GetCalls(method string) []MockContextManagerCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MockContextManagerCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// LastCall returns the last call made (for any method)
func (m *MockContextManager) LastCall() *MockContextManagerCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}
