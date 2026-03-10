package nvimops

import (
	"sync"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
)

// MockManager implements Manager for testing.
// It provides:
//   - Configurable return values for each method
//   - Call recording for verification
//   - Per-method error injection for testing error paths
//   - Optional mock store and generator
type MockManager struct {
	mu sync.RWMutex

	// Configurable return values
	GetResult            *plugin.Plugin
	ListResult           []*plugin.Plugin
	GenerateLuaForResult string

	// Pluggable mock dependencies
	MockStore     store.PluginStore
	MockGenerator plugin.LuaGenerator

	// Error injection (per-method for independent error testing)
	ApplyFileError      error
	ApplyURLError       error
	ApplyError          error
	GetError            error
	ListError           error
	DeleteError         error
	GenerateLuaError    error
	GenerateLuaForError error
	CloseError          error

	// Call recording
	Calls []MockManagerCall

	// Argument recording
	ApplyFileArgs      []string
	ApplyURLArgs       []string
	ApplyArgs          []*plugin.Plugin
	GetArgs            []string
	DeleteArgs         []string
	GenerateLuaArgs    []string
	GenerateLuaForArgs []string
}

// MockManagerCall records a single method call
type MockManagerCall struct {
	Method string
	Args   []interface{}
}

// Compile-time interface compliance check
var _ Manager = (*MockManager)(nil)

// NewMockManager creates a new mock manager with default settings
func NewMockManager() *MockManager {
	return &MockManager{
		Calls:              make([]MockManagerCall, 0),
		ApplyFileArgs:      make([]string, 0),
		ApplyURLArgs:       make([]string, 0),
		ApplyArgs:          make([]*plugin.Plugin, 0),
		GetArgs:            make([]string, 0),
		DeleteArgs:         make([]string, 0),
		GenerateLuaArgs:    make([]string, 0),
		GenerateLuaForArgs: make([]string, 0),
	}
}

// ApplyFile records the call and returns the configured error
func (m *MockManager) ApplyFile(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "ApplyFile",
		Args:   []interface{}{path},
	})
	m.ApplyFileArgs = append(m.ApplyFileArgs, path)

	return m.ApplyFileError
}

// ApplyURL records the call and returns the configured error
func (m *MockManager) ApplyURL(url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "ApplyURL",
		Args:   []interface{}{url},
	})
	m.ApplyURLArgs = append(m.ApplyURLArgs, url)

	return m.ApplyURLError
}

// Apply records the call and returns the configured error
func (m *MockManager) Apply(p *plugin.Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Apply",
		Args:   []interface{}{p},
	})
	m.ApplyArgs = append(m.ApplyArgs, p)

	return m.ApplyError
}

// Get returns the configured plugin or error
func (m *MockManager) Get(name string) (*plugin.Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Get",
		Args:   []interface{}{name},
	})
	m.GetArgs = append(m.GetArgs, name)

	if m.GetError != nil {
		return nil, m.GetError
	}

	return m.GetResult, nil
}

// List returns the configured plugin list or error
func (m *MockManager) List() ([]*plugin.Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "List",
	})

	if m.ListError != nil {
		return nil, m.ListError
	}

	return m.ListResult, nil
}

// Delete records the call and returns the configured error
func (m *MockManager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Delete",
		Args:   []interface{}{name},
	})
	m.DeleteArgs = append(m.DeleteArgs, name)

	return m.DeleteError
}

// GenerateLua records the call and returns the configured error
func (m *MockManager) GenerateLua(outputDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "GenerateLua",
		Args:   []interface{}{outputDir},
	})
	m.GenerateLuaArgs = append(m.GenerateLuaArgs, outputDir)

	return m.GenerateLuaError
}

// GenerateLuaFor returns the configured Lua code or error
func (m *MockManager) GenerateLuaFor(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "GenerateLuaFor",
		Args:   []interface{}{name},
	})
	m.GenerateLuaForArgs = append(m.GenerateLuaForArgs, name)

	if m.GenerateLuaForError != nil {
		return "", m.GenerateLuaForError
	}

	return m.GenerateLuaForResult, nil
}

// Store returns the configured mock store
func (m *MockManager) Store() store.PluginStore {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Store",
	})

	return m.MockStore
}

// Generator returns the configured mock generator
func (m *MockManager) Generator() plugin.LuaGenerator {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Generator",
	})

	return m.MockGenerator
}

// Close records the call and returns the configured error
func (m *MockManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockManagerCall{
		Method: "Close",
	})

	return m.CloseError
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetResult = nil
	m.ListResult = nil
	m.GenerateLuaForResult = ""
	m.MockStore = nil
	m.MockGenerator = nil
	m.ApplyFileError = nil
	m.ApplyURLError = nil
	m.ApplyError = nil
	m.GetError = nil
	m.ListError = nil
	m.DeleteError = nil
	m.GenerateLuaError = nil
	m.GenerateLuaForError = nil
	m.CloseError = nil
	m.Calls = make([]MockManagerCall, 0)
	m.ApplyFileArgs = make([]string, 0)
	m.ApplyURLArgs = make([]string, 0)
	m.ApplyArgs = make([]*plugin.Plugin, 0)
	m.GetArgs = make([]string, 0)
	m.DeleteArgs = make([]string, 0)
	m.GenerateLuaArgs = make([]string, 0)
	m.GenerateLuaForArgs = make([]string, 0)
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

// LastCall returns the last call made (for any method)
func (m *MockManager) LastCall() *MockManagerCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}
