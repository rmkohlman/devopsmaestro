package operators

import "sync"

// MockPlatformDetector implements PlatformDetector for testing.
// It provides:
//   - Configurable return values for each method
//   - Call recording for verification
//   - Error injection for testing error paths
type MockPlatformDetector struct {
	mu sync.RWMutex

	// Configurable return values
	DetectResult          *Platform
	DetectAllResult       []*Platform
	DetectReachableResult []*Platform

	// Error injection
	DetectError error

	// Call recording
	Calls []MockPlatformDetectorCall
}

// MockPlatformDetectorCall records a single method call
type MockPlatformDetectorCall struct {
	Method string
	Args   []interface{}
}

// Compile-time interface compliance check
var _ PlatformDetector = (*MockPlatformDetector)(nil)

// NewMockPlatformDetector creates a new mock platform detector with default settings
func NewMockPlatformDetector() *MockPlatformDetector {
	return &MockPlatformDetector{
		Calls: make([]MockPlatformDetectorCall, 0),
	}
}

// Detect returns the configured platform or error
func (m *MockPlatformDetector) Detect() (*Platform, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockPlatformDetectorCall{
		Method: "Detect",
	})

	if m.DetectError != nil {
		return nil, m.DetectError
	}

	return m.DetectResult, nil
}

// DetectAll returns the configured list of all platforms
func (m *MockPlatformDetector) DetectAll() []*Platform {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockPlatformDetectorCall{
		Method: "DetectAll",
	})

	return m.DetectAllResult
}

// DetectReachable returns the configured list of reachable platforms
func (m *MockPlatformDetector) DetectReachable() []*Platform {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockPlatformDetectorCall{
		Method: "DetectReachable",
	})

	return m.DetectReachableResult
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockPlatformDetector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DetectResult = nil
	m.DetectAllResult = nil
	m.DetectReachableResult = nil
	m.DetectError = nil
	m.Calls = make([]MockPlatformDetectorCall, 0)
}

// CallCount returns the number of times a method was called
func (m *MockPlatformDetector) CallCount(method string) int {
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
func (m *MockPlatformDetector) GetCalls(method string) []MockPlatformDetectorCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MockPlatformDetectorCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// LastCall returns the last call made (for any method)
func (m *MockPlatformDetector) LastCall() *MockPlatformDetectorCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}
