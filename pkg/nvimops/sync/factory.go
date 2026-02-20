package sync

import (
	"fmt"
	"sort"
	"sync"
)

// DefaultSourceHandlerFactory is the default implementation of SourceHandlerFactory.
// It manages registration and creation of source handlers using the Registry pattern.
type DefaultSourceHandlerFactory struct {
	registry *SourceRegistry
}

// NewSourceHandlerFactory creates a new factory with the global source registry.
func NewSourceHandlerFactory() SourceHandlerFactory {
	return &DefaultSourceHandlerFactory{
		registry: GetGlobalRegistry(),
	}
}

// NewSourceHandlerFactoryWithRegistry creates a factory with a custom registry.
// This is useful for testing or when you need isolation.
func NewSourceHandlerFactoryWithRegistry(registry *SourceRegistry) SourceHandlerFactory {
	return &DefaultSourceHandlerFactory{
		registry: registry,
	}
}

// CreateHandler creates a handler for the specified source name.
func (f *DefaultSourceHandlerFactory) CreateHandler(sourceName string) (SourceHandler, error) {
	registration, exists := f.registry.GetRegistration(sourceName)
	if !exists {
		return nil, &ErrSourceNotFound{Source: sourceName}
	}

	// Call the factory function to create a new handler instance
	handler := registration.CreateFunc()

	// Verify the handler implements the interface correctly
	if handler == nil {
		return nil, fmt.Errorf("source factory function returned nil for source: %s", sourceName)
	}

	return handler, nil
}

// ListSources returns all available source names in alphabetical order.
func (f *DefaultSourceHandlerFactory) ListSources() []string {
	sources := f.registry.ListSources()
	sort.Strings(sources)
	return sources
}

// IsSupported checks if a source name is supported.
func (f *DefaultSourceHandlerFactory) IsSupported(sourceName string) bool {
	_, exists := f.registry.GetRegistration(sourceName)
	return exists
}

// GetHandlerInfo returns metadata about a supported source.
func (f *DefaultSourceHandlerFactory) GetHandlerInfo(sourceName string) (*SourceInfo, error) {
	registration, exists := f.registry.GetRegistration(sourceName)
	if !exists {
		return nil, &ErrSourceNotFound{Source: sourceName}
	}

	return &registration.Info, nil
}

// SourceRegistry manages the registration of source handlers.
// This allows new sources to be added without modifying the factory.
type SourceRegistry struct {
	mu            sync.RWMutex
	registrations map[string]HandlerRegistration
}

// NewSourceRegistry creates a new empty registry.
func NewSourceRegistry() *SourceRegistry {
	return &SourceRegistry{
		registrations: make(map[string]HandlerRegistration),
	}
}

// Register adds a new source handler to the registry.
func (r *SourceRegistry) Register(registration HandlerRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate registration
	if registration.Name == "" {
		return fmt.Errorf("registration name cannot be empty")
	}
	if registration.CreateFunc == nil {
		return fmt.Errorf("registration CreateFunc cannot be nil")
	}

	// Check for duplicates
	if _, exists := r.registrations[registration.Name]; exists {
		return &ErrSourceAlreadyRegistered{Source: registration.Name}
	}

	r.registrations[registration.Name] = registration
	return nil
}

// Unregister removes a source handler from the registry.
func (r *SourceRegistry) Unregister(sourceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.registrations[sourceName]; !exists {
		return &ErrSourceNotFound{Source: sourceName}
	}

	delete(r.registrations, sourceName)
	return nil
}

// GetRegistration returns the registration for a source.
func (r *SourceRegistry) GetRegistration(sourceName string) (HandlerRegistration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.registrations[sourceName]
	return registration, exists
}

// ListSources returns all registered source names.
func (r *SourceRegistry) ListSources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sources := make([]string, 0, len(r.registrations))
	for name := range r.registrations {
		sources = append(sources, name)
	}
	return sources
}

// ListRegistrations returns all registrations.
func (r *SourceRegistry) ListRegistrations() []HandlerRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registrations := make([]HandlerRegistration, 0, len(r.registrations))
	for _, registration := range r.registrations {
		registrations = append(registrations, registration)
	}
	return registrations
}

// Clear removes all registrations. Primarily for testing.
func (r *SourceRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.registrations = make(map[string]HandlerRegistration)
}

// Size returns the number of registered sources.
func (r *SourceRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.registrations)
}

// Global registry instance
var globalRegistry *SourceRegistry
var globalRegistryOnce sync.Once

// GetGlobalRegistry returns the global source registry.
// It's created once and reused across the application.
func GetGlobalRegistry() *SourceRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewSourceRegistry()
	})
	return globalRegistry
}

// RegisterGlobalSource is a convenience function to register a source in the global registry.
func RegisterGlobalSource(registration HandlerRegistration) error {
	return GetGlobalRegistry().Register(registration)
}

// Error types for the sync package

// ErrSourceNotFound indicates that a requested source is not registered.
type ErrSourceNotFound struct {
	Source string
}

func (e *ErrSourceNotFound) Error() string {
	return fmt.Sprintf("source not found: %s", e.Source)
}

// ErrSourceAlreadyRegistered indicates that a source is already registered.
type ErrSourceAlreadyRegistered struct {
	Source string
}

func (e *ErrSourceAlreadyRegistered) Error() string {
	return fmt.Sprintf("source already registered: %s", e.Source)
}

// ErrSyncFailed indicates that a sync operation failed.
type ErrSyncFailed struct {
	Source string
	Err    error
}

func (e *ErrSyncFailed) Error() string {
	return fmt.Sprintf("sync failed for source %s: %v", e.Source, e.Err)
}

func (e *ErrSyncFailed) Unwrap() error {
	return e.Err
}
