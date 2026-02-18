package resolver

import (
	"devopsmaestro/db"
	"devopsmaestro/models"
)

// WorkspaceResolver resolves workspace references from partial criteria.
// It abstracts the logic of finding workspaces based on hierarchy flags
// (-e ecosystem, -d domain, -a app, -w workspace) provided by commands.
type WorkspaceResolver interface {
	// Resolve finds workspaces matching the given filter criteria.
	// Returns:
	// - A single WorkspaceWithHierarchy if exactly one match is found
	// - AmbiguousError if multiple matches are found (contains the matches)
	// - ErrNoWorkspaceFound if no matches are found
	Resolve(filter models.WorkspaceFilter) (*models.WorkspaceWithHierarchy, error)

	// ResolveAll finds all workspaces matching the given filter criteria.
	// Unlike Resolve, this returns all matches without treating multiple results as an error.
	// Useful for listing/displaying matching workspaces.
	ResolveAll(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error)
}

// ResolverFactory creates WorkspaceResolver instances.
// This follows the Interface → Implementation → Factory pattern from STANDARDS.md.
type ResolverFactory interface {
	// Create creates a new WorkspaceResolver with the given DataStore.
	Create(store db.DataStore) WorkspaceResolver
}
