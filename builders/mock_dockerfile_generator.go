package builders

import (
	"sync"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
)

// MockDockerfileGenerator implements DockerfileGenerator for testing.
// It provides:
//   - Configurable return values
//   - Call recording for verification
//   - Error injection for testing error paths
//   - Argument recording for SetPluginManifest
type MockDockerfileGenerator struct {
	mu sync.RWMutex

	// Configurable return values
	GenerateResult string

	// Error injection
	GenerateError error

	// Call recording
	Calls []MockDockerfileGeneratorCall

	// Argument recording
	PluginManifestArg *plugin.PluginManifest
}

// MockDockerfileGeneratorCall records a single method call
type MockDockerfileGeneratorCall struct {
	Method string
	Args   []interface{}
}

// Compile-time interface compliance check
var _ DockerfileGenerator = (*MockDockerfileGenerator)(nil)

// NewMockDockerfileGenerator creates a new mock Dockerfile generator with default settings
func NewMockDockerfileGenerator() *MockDockerfileGenerator {
	return &MockDockerfileGenerator{
		Calls: make([]MockDockerfileGeneratorCall, 0),
	}
}

// SetPluginManifest records the call and stores the manifest argument
func (m *MockDockerfileGenerator) SetPluginManifest(manifest *plugin.PluginManifest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockDockerfileGeneratorCall{
		Method: "SetPluginManifest",
		Args:   []interface{}{manifest},
	})

	m.PluginManifestArg = manifest
}

// Generate returns the configured Dockerfile content or error
func (m *MockDockerfileGenerator) Generate() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockDockerfileGeneratorCall{
		Method: "Generate",
	})

	if m.GenerateError != nil {
		return "", m.GenerateError
	}

	return m.GenerateResult, nil
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockDockerfileGenerator) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GenerateResult = ""
	m.GenerateError = nil
	m.PluginManifestArg = nil
	m.Calls = make([]MockDockerfileGeneratorCall, 0)
}

// CallCount returns the number of times a method was called
func (m *MockDockerfileGenerator) CallCount(method string) int {
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
func (m *MockDockerfileGenerator) GetCalls(method string) []MockDockerfileGeneratorCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MockDockerfileGeneratorCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// LastCall returns the last call made (for any method)
func (m *MockDockerfileGenerator) LastCall() *MockDockerfileGeneratorCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}
