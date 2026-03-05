package registry

import (
	"context"
	"fmt"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// Lifecycle mode constants
const (
	LifecycleAuto       = "auto"
	LifecyclePersistent = "persistent" // Same as auto - starts when needed, stays running
	LifecycleOnDemand   = "on-demand"
	LifecycleManual     = "manual"
)

// LifecycleManager manages registry lifecycle (auto-start, idle timeout)
type LifecycleManager struct {
	store   db.DataStore
	manager RegistryManager
}

// NewLifecycleManager creates a new LifecycleManager
func NewLifecycleManager(store db.DataStore, manager RegistryManager) *LifecycleManager {
	return &LifecycleManager{
		store:   store,
		manager: manager,
	}
}

// EnsureRunning starts the registry if not running (based on lifecycle mode)
func (lm *LifecycleManager) EnsureRunning(ctx context.Context, registry *models.Registry) error {
	// Manual mode: do nothing (user must explicitly start)
	if registry.Lifecycle == LifecycleManual {
		return nil
	}

	// Check if already running
	if lm.manager.IsRunning(ctx) {
		return nil
	}

	// Auto/persistent or on-demand modes: start the registry
	if registry.Lifecycle == LifecyclePersistent ||
		registry.Lifecycle == LifecycleOnDemand ||
		registry.Lifecycle == LifecycleAuto {
		if err := lm.manager.Start(ctx); err != nil {
			return fmt.Errorf("failed to start registry: %w", err)
		}
	}

	return nil
}

// ShouldStop checks if a registry should be stopped due to idle timeout
func (lm *LifecycleManager) ShouldStop(ctx context.Context, registry *models.Registry, lastAccess time.Time) (bool, error) {
	// Only on-demand registries should auto-stop
	if registry.Lifecycle != LifecycleOnDemand {
		return false, nil
	}

	// Get effective timeout for this registry
	timeout := lm.GetEffectiveTimeout(registry)

	// Check if idle timeout has been exceeded
	idleDuration := time.Since(lastAccess)
	return idleDuration > timeout, nil
}

// StopIfIdle is not yet implemented — idle-timeout auto-stop requires
// last_accessed_at tracking in the database (deferred to a future release).
func (lm *LifecycleManager) StopIfIdle(ctx context.Context, registry *models.Registry) error {
	return fmt.Errorf("idle-timeout auto-stop not yet implemented: requires last_accessed_at tracking in database")
}

// GetEffectiveTimeout returns the effective timeout for a registry (registry-specific or global default)
func (lm *LifecycleManager) GetEffectiveTimeout(registry *models.Registry) time.Duration {
	// If registry has a specific timeout set, use it
	if registry.IdleTimeout > 0 {
		return time.Duration(registry.IdleTimeout) * time.Second
	}

	// Try to get global default from database
	globalTimeout, err := lm.store.GetDefault(DefaultKeyIdleTimeout)
	if err == nil && globalTimeout != "" {
		// Parse the duration string (e.g., "30m", "45m")
		if duration, err := time.ParseDuration(globalTimeout); err == nil {
			return duration
		}
	}

	// Fall back to hardcoded default
	return 30 * time.Minute
}
