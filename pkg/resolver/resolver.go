package resolver

import (
	"devopsmaestro/db"
	"devopsmaestro/models"
)

// workspaceResolver is the default implementation of WorkspaceResolver.
type workspaceResolver struct {
	store db.DataStore
}

// NewWorkspaceResolver creates a new WorkspaceResolver with the given DataStore.
// This is the primary constructor for creating resolvers.
func NewWorkspaceResolver(store db.DataStore) WorkspaceResolver {
	return &workspaceResolver{
		store: store,
	}
}

// Resolve finds workspaces matching the given filter criteria.
// Returns:
// - A single WorkspaceWithHierarchy if exactly one match is found
// - AmbiguousError if multiple matches are found (contains the matches)
// - ErrNoWorkspaceFound if no matches are found
func (r *workspaceResolver) Resolve(filter models.WorkspaceFilter) (*models.WorkspaceWithHierarchy, error) {
	matches, err := r.store.FindWorkspaces(filter)
	if err != nil {
		return nil, err
	}

	switch len(matches) {
	case 0:
		return nil, ErrNoWorkspaceFound
	case 1:
		return matches[0], nil
	default:
		return nil, NewAmbiguousError(matches)
	}
}

// ResolveAll finds all workspaces matching the given filter criteria.
// Unlike Resolve, this returns all matches without treating multiple results as an error.
func (r *workspaceResolver) ResolveAll(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error) {
	return r.store.FindWorkspaces(filter)
}

// defaultResolverFactory is the default implementation of ResolverFactory.
type defaultResolverFactory struct{}

// DefaultFactory is a singleton factory for creating WorkspaceResolver instances.
var DefaultFactory ResolverFactory = &defaultResolverFactory{}

// Create creates a new WorkspaceResolver with the given DataStore.
func (f *defaultResolverFactory) Create(store db.DataStore) WorkspaceResolver {
	return NewWorkspaceResolver(store)
}

// Ensure workspaceResolver implements WorkspaceResolver
var _ WorkspaceResolver = (*workspaceResolver)(nil)

// Ensure defaultResolverFactory implements ResolverFactory
var _ ResolverFactory = (*defaultResolverFactory)(nil)
