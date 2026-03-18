// Package resolver provides hierarchical build args resolution across the object hierarchy.
// It walks the cascade: global < ecosystem < domain < app < workspace,
// merging build args at each level (most-specific wins for overlapping keys).
package resolver

import (
	"context"
)

// HierarchyLevel defines a level in the build args cascade hierarchy.
// Levels are ordered from least specific (global) to most specific (workspace).
type HierarchyLevel int

const (
	// LevelGlobal is the least-specific level — DVM-wide defaults stored in defaults table.
	LevelGlobal HierarchyLevel = iota
	// LevelEcosystem is the ecosystem-scoped build args level.
	LevelEcosystem
	// LevelDomain is the domain-scoped build args level.
	LevelDomain
	// LevelApp is the app-scoped build args level.
	LevelApp
	// LevelWorkspace is the most-specific level — workspace-scoped build args.
	LevelWorkspace
)

// String returns the display string for the hierarchy level.
func (h HierarchyLevel) String() string {
	switch h {
	case LevelGlobal:
		return "global"
	case LevelEcosystem:
		return "ecosystem"
	case LevelDomain:
		return "domain"
	case LevelApp:
		return "app"
	case LevelWorkspace:
		return "workspace"
	default:
		return "unknown"
	}
}

// BuildArgsStep represents one level in the cascade resolution path.
type BuildArgsStep struct {
	Level HierarchyLevel    // Which hierarchy level this step represents
	Name  string            // Human-readable name of the object at this level
	Args  map[string]string // Build args set at this specific level (may be empty)
	Found bool              // true if this level contributed any build args
}

// BuildArgsResolution holds the complete result of resolving build args
// across the full hierarchy cascade.
type BuildArgsResolution struct {
	// Args is the merged result map — what the build actually uses.
	// Workspace args override app args, which override domain args, etc.
	Args map[string]string

	// Sources records which HierarchyLevel each key in Args came from.
	// Useful for --effective display to show provenance.
	Sources map[string]HierarchyLevel

	// Path contains one BuildArgsStep for every hierarchy level (always 5 entries),
	// ordered from global (index 0) to workspace (index 4).
	Path []BuildArgsStep
}

// BuildArgsResolver resolves hierarchical build args for a given workspace.
type BuildArgsResolver interface {
	// Resolve walks the full hierarchy for the given workspace ID and returns
	// the merged build args along with per-key provenance information.
	// Returns an error if the workspace (or its parents) cannot be found.
	Resolve(ctx context.Context, workspaceID int) (*BuildArgsResolution, error)
}
