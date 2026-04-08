package resolver

import (
	"devopsmaestro/db"
	"devopsmaestro/models"
)

// InferenceResolver enriches a WorkspaceFilter with smart inference
// (single-ecosystem auto-fill, cascade via app/domain name lookups)
// before delegating to the standard WorkspaceResolver.
//
// This enables commands like `dvm build -a myapi -w main` to work
// without requiring prior `dvm use app` or explicit --ecosystem flags
// when the hierarchy is unambiguous.
type InferenceResolver interface {
	// ResolveWithInference enriches the filter via inference and resolves
	// to a single workspace. Returns:
	//   - *WorkspaceWithHierarchy if exactly one match after inference
	//   - AmbiguousError if app/domain name matches multiple hierarchies
	//   - ErrNoWorkspaceFound if no workspace matches after inference
	ResolveWithInference(filter models.WorkspaceFilter) (*models.WorkspaceWithHierarchy, error)
}

// inferenceResolver is the default implementation of InferenceResolver.
type inferenceResolver struct {
	store    db.DataStore
	delegate WorkspaceResolver
}

// NewInferenceResolver creates an InferenceResolver backed by the given DataStore.
// It uses the standard WorkspaceResolver for final resolution after inference.
func NewInferenceResolver(store db.DataStore) InferenceResolver {
	return &inferenceResolver{
		store:    store,
		delegate: NewWorkspaceResolver(store),
	}
}

// ResolveWithInference enriches the given filter with inferred hierarchy
// information, then delegates to the standard WorkspaceResolver.
//
// Inference order:
//  1. App name inference: if AppName is set but DomainName/EcosystemName are not,
//     look up the app by name. If exactly one match, fill in domain and ecosystem.
//     If multiple matches, return AmbiguousError.
//  2. Domain name inference: if DomainName is set but EcosystemName is not,
//     look up the domain by name. If exactly one match, fill in ecosystem.
//     If multiple matches, return AmbiguousError.
//  3. Single-ecosystem inference: if EcosystemName is still empty, check
//     CountEcosystems(). If exactly one ecosystem exists, auto-fill it.
//  4. Delegate to WorkspaceResolver.Resolve() with the enriched filter.
func (ir *inferenceResolver) ResolveWithInference(filter models.WorkspaceFilter) (*models.WorkspaceWithHierarchy, error) {
	enriched := filter

	// Step 1: App name inference — cascade to domain and ecosystem
	if enriched.AppName != "" && enriched.DomainName == "" && enriched.EcosystemName == "" {
		matches, err := ir.store.FindAppsByName(enriched.AppName)
		if err != nil {
			return nil, err
		}

		switch len(matches) {
		case 0:
			// No app with this name exists — let the delegate return ErrNoWorkspaceFound
			return nil, ErrNoWorkspaceFound
		case 1:
			enriched.DomainName = matches[0].Domain.Name
			enriched.EcosystemName = matches[0].Ecosystem.Name
		default:
			// Multiple apps with same name across hierarchy — ambiguous.
			// Convert AppWithHierarchy matches into workspace search to
			// build an AmbiguousError with WorkspaceWithHierarchy entries.
			return ir.resolveAmbiguousApp(enriched, matches)
		}
	}

	// Step 2: Domain name inference — cascade to ecosystem
	if enriched.DomainName != "" && enriched.EcosystemName == "" {
		matches, err := ir.store.FindDomainsByName(enriched.DomainName)
		if err != nil {
			return nil, err
		}

		switch len(matches) {
		case 0:
			// Domain name doesn't exist — fall through and let delegate handle it
		case 1:
			enriched.EcosystemName = matches[0].Ecosystem.Name
		default:
			// Multiple domains with same name across ecosystems — ambiguous.
			// Delegate with the filter as-is to collect workspace matches for error.
			return ir.resolveAmbiguousDomain(enriched, matches)
		}
	}

	// Step 3: Single-ecosystem inference
	if enriched.EcosystemName == "" {
		count, err := ir.store.CountEcosystems()
		if err != nil {
			return nil, err
		}

		if count == 1 {
			ecosystems, err := ir.store.ListEcosystems()
			if err != nil {
				return nil, err
			}
			if len(ecosystems) == 1 {
				enriched.EcosystemName = ecosystems[0].Name
			}
		}
	}

	// Step 4: Delegate to the standard resolver with the enriched filter
	return ir.delegate.Resolve(enriched)
}

// resolveAmbiguousApp handles the case where FindAppsByName returned multiple
// matches. It runs FindWorkspaces with the original filter to collect workspace-level
// matches, then returns an AmbiguousError.
func (ir *inferenceResolver) resolveAmbiguousApp(filter models.WorkspaceFilter, appMatches []*models.AppWithHierarchy) (*models.WorkspaceWithHierarchy, error) {
	// Try to find all matching workspaces across the ambiguous apps
	allWorkspaces, err := ir.delegate.ResolveAll(filter)
	if err != nil {
		return nil, err
	}

	if len(allWorkspaces) == 0 {
		return nil, ErrNoWorkspaceFound
	}

	return nil, NewAmbiguousError(allWorkspaces)
}

// resolveAmbiguousDomain handles the case where FindDomainsByName returned
// multiple matches. It runs FindWorkspaces to collect workspace-level matches,
// then returns an AmbiguousError.
func (ir *inferenceResolver) resolveAmbiguousDomain(filter models.WorkspaceFilter, domainMatches []*models.DomainWithHierarchy) (*models.WorkspaceWithHierarchy, error) {
	allWorkspaces, err := ir.delegate.ResolveAll(filter)
	if err != nil {
		return nil, err
	}

	if len(allWorkspaces) == 0 {
		return nil, ErrNoWorkspaceFound
	}

	return nil, NewAmbiguousError(allWorkspaces)
}

// Ensure inferenceResolver implements InferenceResolver
var _ InferenceResolver = (*inferenceResolver)(nil)
