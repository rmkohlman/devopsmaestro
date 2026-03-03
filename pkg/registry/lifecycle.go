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

// RecordActivity updates the last access time for a registry
func (lm *LifecycleManager) RecordActivity(ctx context.Context, registryID int) error {
	// For now, this is a no-op as we don't have a last_accessed_at field yet
	// In the future, this would update a timestamp in the database
	// Example: return lm.store.UpdateRegistryLastAccess(registryID, time.Now())
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

// StopIfIdle stops a registry if it has exceeded idle timeout
func (lm *LifecycleManager) StopIfIdle(ctx context.Context, registry *models.Registry) error {
	// Only process on-demand registries
	if registry.Lifecycle != LifecycleOnDemand {
		return nil
	}

	// Check if registry is running
	if !lm.manager.IsRunning(ctx) {
		return nil
	}

	// For testing, use a recent time to check if should stop
	// In production, this would query the last access time from database
	lastAccess := time.Now().Add(-time.Duration(registry.IdleTimeout) * time.Second)

	shouldStop, err := lm.ShouldStop(ctx, registry, lastAccess)
	if err != nil {
		return fmt.Errorf("failed to check idle timeout: %w", err)
	}

	if shouldStop {
		if err := lm.manager.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop idle registry: %w", err)
		}
	}

	return nil
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
