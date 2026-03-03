package registry

import (
	"context"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// ResolutionSource indicates where a registry setting came from
type ResolutionSource int

const (
	SourceNone    ResolutionSource = iota // No default configured
	SourceDefault                         // From database default
	SourceCLI                             // From CLI override
)

// RegistryResolution contains the result of resolving a registry
type RegistryResolution struct {
	Source       ResolutionSource
	RegistryName string
	Endpoint     string
	Enabled      bool
}

// RegistryResolver resolves which registry to use for a given type
type RegistryResolver struct {
	store   db.DataStore
	manager RegistryManager
}

// NewRegistryResolver creates a new RegistryResolver
func NewRegistryResolver(store db.DataStore, manager RegistryManager) *RegistryResolver {
	return &RegistryResolver{
		store:   store,
		manager: manager,
	}
}

// Resolve determines which registry to use for a type alias
func (rr *RegistryResolver) Resolve(ctx context.Context, typeAlias string, cliOverride string) (*RegistryResolution, error) {
	// Priority 1: CLI override takes precedence
	if cliOverride != "" {
		return &RegistryResolution{
			Source:       SourceCLI,
			RegistryName: "",
			Endpoint:     cliOverride,
			Enabled:      true,
		}, nil
	}

	// Priority 2: Check for default registry setting
	// Try the direct key first (e.g., "registry-oci" or "registry-zot")
	defaultKey := "registry-" + typeAlias
	defaultName, err := rr.store.GetDefault(defaultKey)
	if err != nil {
		return nil, err
	}

	// If not found, check if this is a concrete type that has an alias
	if defaultName == "" {
		if alias, ok := GetAliasForType(typeAlias); ok {
			// Try the alias key (e.g., "registry-oci" for type "zot")
			defaultKey = "registry-" + alias
			defaultName, err = rr.store.GetDefault(defaultKey)
			if err != nil {
				return nil, err
			}
		}
	}

	// Priority 3: No default configured
	if defaultName == "" {
		return &RegistryResolution{
			Source:       SourceNone,
			RegistryName: "",
			Endpoint:     "",
			Enabled:      false,
		}, nil
	}

	// Look up the default registry by name
	registry, err := rr.store.GetRegistryByName(defaultName)
	if err != nil {
		return nil, err
	}

	// Build endpoint from registry
	endpoint := rr.BuildEndpoint(registry)

	return &RegistryResolution{
		Source:       SourceDefault,
		RegistryName: registry.Name,
		Endpoint:     endpoint,
		Enabled:      registry.Enabled,
	}, nil
}

// ResolveAll returns all enabled registries
func (rr *RegistryResolver) ResolveAll(ctx context.Context) ([]*RegistryResolution, error) {
	// Get all registries from the datastore
	registries, err := rr.store.ListRegistries()
	if err != nil {
		return nil, err
	}

	// Filter to only enabled registries and build resolutions
	var results []*RegistryResolution
	for _, registry := range registries {
		if registry.Enabled {
			endpoint := rr.BuildEndpoint(registry)
			results = append(results, &RegistryResolution{
				Source:       SourceDefault,
				RegistryName: registry.Name,
				Endpoint:     endpoint,
				Enabled:      true,
			})
		}
	}

	return results, nil
}

// BuildEndpoint constructs the endpoint URL for a registry
func (rr *RegistryResolver) BuildEndpoint(registry *models.Registry) string {
	// Build endpoint as http://localhost:port
	return fmt.Sprintf("http://localhost:%d", registry.Port)
}
