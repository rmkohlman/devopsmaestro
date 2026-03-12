package registry

import (
	"context"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// EnsureDefaultRegistry ensures that a default registry exists for the given
// alias (e.g., "oci") and lifecycle (e.g., "on-demand", "persistent").
//
// It resolves the alias to a concrete registry type using the alias map, checks
// whether a registry with that name already exists, and either creates a new one
// or re-uses the existing one. In both cases the registry is set as the default
// for the alias.
//
// Returns (true, nil) when a new registry was created, (false, nil) when one
// already existed, or (false, err) on failure.
func EnsureDefaultRegistry(
	ctx context.Context,
	registryStore db.RegistryStore,
	defaultsStore db.DefaultsStore,
	alias string,
	lifecycle string,
) (bool, error) {
	// Resolve alias to concrete type.
	allAliases := GetAllAliases()
	concreteType, ok := allAliases[alias]
	if !ok {
		return false, fmt.Errorf("unknown registry alias: %q", alias)
	}

	defaults := NewRegistryDefaults(defaultsStore)

	// Check if a registry with this name already exists.
	existing, err := registryStore.GetRegistryByName(concreteType)
	if err == nil && existing != nil {
		// Registry already exists — just set it as the default.
		if err := defaults.SetByAlias(ctx, alias, existing.Name); err != nil {
			return false, fmt.Errorf("setting default for alias %q: %w", alias, err)
		}
		return false, nil
	}

	// Registry does not exist — create it.
	reg := &models.Registry{
		Name:        concreteType,
		Type:        concreteType,
		Enabled:     true,
		Lifecycle:   lifecycle,
		Status:      "stopped",
		IdleTimeout: 1800,
	}
	reg.ApplyDefaults()

	if err := reg.Validate(); err != nil {
		return false, fmt.Errorf("validating default registry %q: %w", concreteType, err)
	}

	if err := registryStore.CreateRegistry(reg); err != nil {
		return false, fmt.Errorf("creating default registry %q: %w", concreteType, err)
	}

	if err := defaults.SetByAlias(ctx, alias, reg.Name); err != nil {
		return false, fmt.Errorf("setting default for alias %q: %w", alias, err)
	}

	return true, nil
}
