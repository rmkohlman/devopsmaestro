package registry

import (
	"context"
	"encoding/json"

	"devopsmaestro/models"
)

// ServiceManager is a unified interface for all registry runtime managers.
// It provides a common set of lifecycle methods that all registry types must implement.
// This interface bridges the gap between database-persisted Registry resources
// and the runtime service managers (ZotManager, AthensManager, etc.).
type ServiceManager interface {
	// Start starts the registry service.
	// Returns an error if the service fails to start.
	// Idempotent - safe to call if already running.
	Start(ctx context.Context) error

	// Stop stops the registry service gracefully.
	// Returns an error if the service fails to stop.
	// Idempotent - safe to call if already stopped.
	Stop(ctx context.Context) error

	// IsRunning checks if the registry service is currently running.
	IsRunning(ctx context.Context) bool

	// GetEndpoint returns the service endpoint (e.g., "localhost:5000").
	// The format may vary by registry type (some return full URLs, others just host:port).
	GetEndpoint() string
}

// RegistryStrategy defines the strategy pattern for type-specific behavior.
// Each registry type (zot, athens, devpi, etc.) implements this interface
// to provide its own configuration validation, manager creation, and defaults.
type RegistryStrategy interface {
	// ValidateConfig validates the registry-specific configuration JSON.
	// Returns an error if the configuration is invalid.
	ValidateConfig(config json.RawMessage) error

	// CreateManager creates a ServiceManager instance for this registry type.
	// The manager is configured based on the provided Registry model.
	CreateManager(reg *models.Registry) (ServiceManager, error)

	// GetDefaultPort returns the default port for this registry type.
	GetDefaultPort() int

	// GetDefaultStorage returns the default storage path for this registry type.
	GetDefaultStorage() string

	// GetDefaultVersion returns the default binary version for this registry type.
	// Returns empty string for externally-managed registries (athens, devpi, verdaccio, squid).
	GetDefaultVersion() string
}
