package resource

import (
	"fmt"
	"sync"
)

var (
	// handlers stores registered handlers by Kind
	handlers        = make(map[string]Handler)
	fallbackHandler Handler // For dynamic/custom resources
	mu              sync.RWMutex
)

// Register adds a handler to the registry.
// Handlers should be registered at application startup.
// Panics if a handler for the same Kind is already registered.
func Register(h Handler) {
	mu.Lock()
	defer mu.Unlock()

	kind := h.Kind()
	if _, exists := handlers[kind]; exists {
		panic(fmt.Sprintf("resource handler already registered for kind: %s", kind))
	}
	handlers[kind] = h
}

// SetFallbackHandler sets a handler to use for unknown kinds.
// This is typically the DynamicHandler for custom resources.
func SetFallbackHandler(h Handler) {
	mu.Lock()
	defer mu.Unlock()
	fallbackHandler = h
}

// GetHandler returns the handler for the given Kind.
// Returns the fallback handler if no specific handler is registered.
// Returns nil if no handler is found and no fallback is set.
func GetHandler(kind string) Handler {
	mu.RLock()
	defer mu.RUnlock()

	if h, exists := handlers[kind]; exists {
		return h
	}
	return fallbackHandler
}

// MustGetHandler returns the handler for the given Kind.
// Returns an error if no handler is registered.
func MustGetHandler(kind string) (Handler, error) {
	h := GetHandler(kind)
	if h == nil {
		return nil, fmt.Errorf("no handler registered for kind: %s", kind)
	}
	return h, nil
}

// RegisteredKinds returns all registered resource kinds.
func RegisteredKinds() []string {
	mu.RLock()
	defer mu.RUnlock()

	kinds := make([]string, 0, len(handlers))
	for kind := range handlers {
		kinds = append(kinds, kind)
	}
	return kinds
}

// ClearRegistry removes all registered handlers.
// This is primarily useful for testing.
func ClearRegistry() {
	mu.Lock()
	defer mu.Unlock()
	handlers = make(map[string]Handler)
	fallbackHandler = nil
}

// Apply parses YAML data, detects the Kind, and applies it using the appropriate handler.
// This is the main entry point for the unified apply pipeline.
func Apply(ctx Context, data []byte, source string) (Resource, error) {
	kind, err := DetectKind(data)
	if err != nil {
		return nil, fmt.Errorf("failed to detect resource kind: %w", err)
	}

	handler, err := MustGetHandler(kind)
	if err != nil {
		return nil, err
	}

	return handler.Apply(ctx, data)
}

// Get retrieves a resource by kind and name.
func Get(ctx Context, kind, name string) (Resource, error) {
	handler, err := MustGetHandler(kind)
	if err != nil {
		return nil, err
	}

	return handler.Get(ctx, name)
}

// List returns all resources of the given kind.
func List(ctx Context, kind string) ([]Resource, error) {
	handler, err := MustGetHandler(kind)
	if err != nil {
		return nil, err
	}

	return handler.List(ctx)
}

// Delete removes a resource by kind and name.
func Delete(ctx Context, kind, name string) error {
	handler, err := MustGetHandler(kind)
	if err != nil {
		return err
	}

	return handler.Delete(ctx, name)
}

// ToYAML serializes a resource to YAML using the appropriate handler.
func ToYAML(res Resource) ([]byte, error) {
	handler, err := MustGetHandler(res.GetKind())
	if err != nil {
		return nil, err
	}

	return handler.ToYAML(res)
}
