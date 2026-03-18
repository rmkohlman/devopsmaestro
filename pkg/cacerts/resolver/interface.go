// Package resolver provides hierarchical CA certificate resolution across the object hierarchy.
// It walks the cascade: global < ecosystem < domain < app < workspace,
// merging CA certs at each level (more-specific levels override by cert Name).
//
// RED PHASE (v0.56.0): This file defines the interface contract that the
// implementation (hierarchy.go) must satisfy. The interface is defined here
// to allow the test file to compile independently of the implementation.
package resolver

import (
	"context"
)

// HierarchyLevel defines a level in the CA certs cascade hierarchy.
// Levels are ordered from least specific (global) to most specific (workspace).
// NOTE: These mirror the same constants in pkg/buildargs/resolver — they are
// intentionally separate (SRP: each resolver owns its own level type).
type HierarchyLevel int

const (
	// LevelGlobal is the least-specific level — DVM-wide defaults.
	LevelGlobal HierarchyLevel = iota
	// LevelEcosystem is the ecosystem-scoped CA certs level.
	LevelEcosystem
	// LevelDomain is the domain-scoped CA certs level.
	LevelDomain
	// LevelApp is the app-scoped CA certs level.
	LevelApp
	// LevelWorkspace is the most-specific level — workspace-scoped CA certs.
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

// CACertEntry represents a single CA certificate in the resolved result.
// It carries the cert config plus provenance (which level it came from).
type CACertEntry struct {
	// Name is the cert's friendly name (used for deduplication — more-specific
	// levels override less-specific levels with the same Name).
	Name string

	// VaultSecret is the MaestroVault secret name to fetch the cert PEM from.
	VaultSecret string

	// VaultEnvironment is the optional vault environment override.
	VaultEnvironment string

	// VaultField is the optional field within the secret.
	VaultField string

	// Source is which hierarchy level this cert entry came from.
	Source HierarchyLevel
}

// CACertsResolutionStep represents one level in the cascade resolution path.
type CACertsResolutionStep struct {
	Level HierarchyLevel // Which hierarchy level this step represents
	Name  string         // Human-readable name of the object at this level
	Certs []CACertEntry  // CA certs defined at this specific level (may be empty)
	Found bool           // true if this level contributed any CA certs
}

// CACertsResolution holds the complete result of resolving CA certs
// across the full hierarchy cascade.
type CACertsResolution struct {
	// Certs is the merged result slice — what the build actually uses.
	// More-specific levels override less-specific levels by cert Name.
	// The merged result is validated against the 10-cert maximum.
	Certs []CACertEntry

	// Sources records which HierarchyLevel each cert Name came from.
	// Useful for --effective display to show provenance.
	Sources map[string]HierarchyLevel

	// Path contains one CACertsResolutionStep for every hierarchy level
	// (always 5 entries), ordered from global (index 0) to workspace (index 4).
	Path []CACertsResolutionStep
}

// CACertsResolver resolves hierarchical CA certs for a given workspace.
type CACertsResolver interface {
	// Resolve walks the full hierarchy for the given workspace ID and returns
	// the merged CA certs along with per-cert provenance information.
	// Returns an error if the workspace (or its parents) cannot be found,
	// or if the merged result exceeds the 10-cert maximum.
	Resolve(ctx context.Context, workspaceID int) (*CACertsResolution, error)
}
