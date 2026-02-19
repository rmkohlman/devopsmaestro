// Package resolver provides hierarchical theme resolution across the object hierarchy.
// It implements the Strategy pattern to resolve effective themes by walking up:
// Workspace → App → Domain → Ecosystem → Global Default
package resolver

import (
	"context"
	"time"

	"devopsmaestro/pkg/nvimops/theme"
)

// DefaultTheme is the global fallback theme
const DefaultTheme = "coolnight-ocean"

// HierarchyLevel defines where to start theme resolution
type HierarchyLevel int

const (
	LevelWorkspace HierarchyLevel = iota
	LevelApp
	LevelDomain
	LevelEcosystem
	LevelGlobal
)

// String returns the string representation of the hierarchy level
func (h HierarchyLevel) String() string {
	switch h {
	case LevelWorkspace:
		return "workspace"
	case LevelApp:
		return "app"
	case LevelDomain:
		return "domain"
	case LevelEcosystem:
		return "ecosystem"
	case LevelGlobal:
		return "global"
	default:
		return "unknown"
	}
}

// ThemeResolver resolves effective themes using hierarchy walking
type ThemeResolver interface {
	// Resolve walks the hierarchy from the starting level upward
	// Returns the first theme found or the global default
	Resolve(ctx context.Context, level HierarchyLevel, objectID int) (*ThemeResolution, error)

	// ResolveDefault returns the global default theme
	ResolveDefault() (*ThemeResolution, error)

	// GetResolutionPath returns the complete resolution trace without loading themes
	// Useful for debugging and display purposes
	GetResolutionPath(ctx context.Context, level HierarchyLevel, objectID int) (*ThemeResolution, error)
}

// ThemeResolution contains complete resolution information
type ThemeResolution struct {
	// Effective result
	Theme      *theme.Theme   `json:"theme"`
	Source     HierarchyLevel `json:"source"`
	SourceName string         `json:"source_name"`
	SourceID   int            `json:"source_id"`

	// Full resolution path (for debugging/transparency)
	Path []ThemeStep `json:"path"`

	// Performance metadata
	CacheHit   bool      `json:"cache_hit,omitempty"`
	ResolvedAt time.Time `json:"resolved_at"`
}

// ThemeStep represents one step in the hierarchy walk
type ThemeStep struct {
	Level     HierarchyLevel `json:"level"`
	ObjectID  int            `json:"object_id,omitempty"`
	Name      string         `json:"name"`
	ThemeName string         `json:"theme_name,omitempty"`
	Found     bool           `json:"found"`
	Error     string         `json:"error,omitempty"`
}

// IsLight returns whether the effective theme is a light theme
func (r *ThemeResolution) IsLight() bool {
	if r.Theme == nil {
		return false
	}
	return r.Theme.Category == "light"
}

// GetEffectiveThemeName returns the name of the effective theme
func (r *ThemeResolution) GetEffectiveThemeName() string {
	if r.Theme != nil {
		return r.Theme.Name
	}

	// When theme isn't loaded (e.g., from GetResolutionPath), find the effective theme
	// from the resolution path by looking for the first found theme
	for _, step := range r.Path {
		if step.Found && step.ThemeName != "" {
			return step.ThemeName
		}
	}

	// No theme found in path, return default
	return DefaultTheme
}

// GetSourceDescription returns a human-readable description of where the theme came from
func (r *ThemeResolution) GetSourceDescription() string {
	if r.SourceName == "" {
		return r.Source.String()
	}
	return r.Source.String() + " '" + r.SourceName + "'"
}
