package secrets

import (
	"sync"
)

// ProviderFactory creates and manages secret providers.
// It maintains a registry of available providers and handles default provider selection.
//
// Thread-safe: all operations are protected by a mutex.
type ProviderFactory struct {
	mu              sync.RWMutex
	defaultProvider string
	providers       map[string]SecretProvider
}

// NewProviderFactory creates a factory with default configuration.
// The factory starts with no providers registered.
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]SecretProvider),
	}
}

// Register adds a provider to the factory.
// If this is the first provider and no default is set, it becomes the default.
// If a provider with the same name already exists, it is replaced.
func (f *ProviderFactory) Register(provider SecretProvider) {
	f.mu.Lock()
	defer f.mu.Unlock()

	name := provider.Name()
	f.providers[name] = provider

	// Set as default if this is the first provider
	if f.defaultProvider == "" {
		f.defaultProvider = name
	}
}

// Unregister removes a provider from the factory.
// If the removed provider was the default, the default is cleared.
func (f *ProviderFactory) Unregister(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.providers, name)

	if f.defaultProvider == name {
		f.defaultProvider = ""
	}
}

// Get returns a provider by name.
// Returns ErrProviderNotFound if the provider is not registered.
// Returns ErrProviderNotAvailable if the provider exists but IsAvailable() returns false.
func (f *ProviderFactory) Get(name string) (SecretProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, ok := f.providers[name]
	if !ok {
		return nil, &ProviderError{
			Provider: name,
			Op:       "get",
			Err:      ErrProviderNotFound,
		}
	}

	if !provider.IsAvailable() {
		return nil, &ProviderError{
			Provider: name,
			Op:       "get",
			Err:      ErrProviderNotAvailable,
		}
	}

	return provider, nil
}

// GetDefault returns the default provider.
// Returns ErrNoDefaultProvider if no default is configured.
// Returns ErrProviderNotAvailable if the default provider is not available.
func (f *ProviderFactory) GetDefault() (SecretProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.defaultProvider == "" {
		return nil, ErrNoDefaultProvider
	}

	provider, ok := f.providers[f.defaultProvider]
	if !ok {
		return nil, &ProviderError{
			Provider: f.defaultProvider,
			Op:       "get default",
			Err:      ErrProviderNotFound,
		}
	}

	if !provider.IsAvailable() {
		return nil, &ProviderError{
			Provider: f.defaultProvider,
			Op:       "get default",
			Err:      ErrProviderNotAvailable,
		}
	}

	return provider, nil
}

// SetDefault sets the default provider name.
// Returns ErrProviderNotFound if the provider is not registered.
func (f *ProviderFactory) SetDefault(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.providers[name]; !ok {
		return &ProviderError{
			Provider: name,
			Op:       "set default",
			Err:      ErrProviderNotFound,
		}
	}

	f.defaultProvider = name
	return nil
}

// DefaultName returns the name of the default provider.
// Returns an empty string if no default is set.
func (f *ProviderFactory) DefaultName() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.defaultProvider
}

// List returns the names of all registered providers.
func (f *ProviderFactory) List() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}

// ListAvailable returns the names of all available providers.
// A provider is available if IsAvailable() returns true.
func (f *ProviderFactory) ListAvailable() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name, provider := range f.providers {
		if provider.IsAvailable() {
			names = append(names, name)
		}
	}
	return names
}

// Ensure ProviderFactory methods exist (compile-time check)
var _ interface {
	Register(SecretProvider)
	Get(string) (SecretProvider, error)
	GetDefault() (SecretProvider, error)
	SetDefault(string) error
} = (*ProviderFactory)(nil)
