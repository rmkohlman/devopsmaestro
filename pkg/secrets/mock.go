package secrets

import (
	"context"
	"sync"
)

// MockProvider is a mock SecretProvider for testing.
// It allows setting specific secrets and errors for test scenarios.
//
// Thread-safe: all operations are protected by a mutex.
type MockProvider struct {
	mu        sync.RWMutex
	name      string
	available bool
	secrets   map[string]string
	errors    map[string]error
	calls     []MockCall
}

// MockCall records a call to GetSecret for verification in tests.
type MockCall struct {
	Name    string
	Key     string
	Options map[string]string
}

// MockProviderOption is a functional option for configuring MockProvider.
type MockProviderOption func(*MockProvider)

// WithMockName sets the provider name.
func WithMockName(name string) MockProviderOption {
	return func(m *MockProvider) {
		m.name = name
	}
}

// WithMockAvailable sets whether the provider is available.
func WithMockAvailable(available bool) MockProviderOption {
	return func(m *MockProvider) {
		m.available = available
	}
}

// WithMockSecrets sets initial secrets.
func WithMockSecrets(secrets map[string]string) MockProviderOption {
	return func(m *MockProvider) {
		for k, v := range secrets {
			m.secrets[k] = v
		}
	}
}

// NewMockProvider creates a mock provider for testing.
// By default, it uses name "mock" and is available.
func NewMockProvider(opts ...MockProviderOption) *MockProvider {
	m := &MockProvider{
		name:      "mock",
		available: true,
		secrets:   make(map[string]string),
		errors:    make(map[string]error),
		calls:     make([]MockCall, 0),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Name returns the provider identifier.
func (m *MockProvider) Name() string {
	return m.name
}

// IsAvailable returns whether the provider is available.
func (m *MockProvider) IsAvailable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.available
}

// GetSecret retrieves a mock secret.
// Records the call for test verification.
func (m *MockProvider) GetSecret(ctx context.Context, req SecretRequest) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.calls = append(m.calls, MockCall{
		Name:    req.Name,
		Key:     req.Key,
		Options: req.Options,
	})

	// Build the lookup key
	key := req.Name
	if req.Key != "" {
		key = req.Name + ":" + req.Key
	}

	// Check for configured error
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if err, ok := m.errors[req.Name]; ok && req.Key != "" {
		// Also check base name for errors
		return "", err
	}

	// Check for configured secret
	if val, ok := m.secrets[key]; ok {
		return val, nil
	}
	if val, ok := m.secrets[req.Name]; ok && req.Key == "" {
		return val, nil
	}

	return "", ErrSecretNotFound
}

// SetSecret configures a secret value for testing.
// Use "name:key" format for structured secrets.
func (m *MockProvider) SetSecret(name, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[name] = value
}

// SetError configures an error for a specific secret.
// Use "name:key" format for structured secrets.
func (m *MockProvider) SetError(name string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[name] = err
}

// SetAvailable changes the availability status.
func (m *MockProvider) SetAvailable(available bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.available = available
}

// GetCalls returns all recorded calls for test verification.
func (m *MockProvider) GetCalls() []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	calls := make([]MockCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// ClearCalls clears the recorded calls.
func (m *MockProvider) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]MockCall, 0)
}

// Reset clears all secrets, errors, and calls.
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets = make(map[string]string)
	m.errors = make(map[string]error)
	m.calls = make([]MockCall, 0)
}

// Ensure MockProvider implements SecretProvider
var _ SecretProvider = (*MockProvider)(nil)
