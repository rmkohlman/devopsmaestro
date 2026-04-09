// Package resolver provides package hierarchy resolution across the object hierarchy.
// It walks Workspace → App → Domain → Ecosystem → Global Default to find the
// effective nvim or terminal package. This mirrors the theme resolver pattern
// from pkg/colors/resolver/hierarchy.go.
package resolver

import (
	"context"
	"database/sql"
	"fmt"
)

// ---------------------------------------------------------------------------
// Hierarchy Level
// ---------------------------------------------------------------------------

// PackageHierarchyLevel defines where to start package resolution.
type PackageHierarchyLevel int

const (
	PackageLevelWorkspace PackageHierarchyLevel = iota
	PackageLevelApp
	PackageLevelDomain
	PackageLevelEcosystem
	PackageLevelGlobal
)

// String returns the string representation of the hierarchy level.
func (l PackageHierarchyLevel) String() string {
	switch l {
	case PackageLevelWorkspace:
		return "workspace"
	case PackageLevelApp:
		return "app"
	case PackageLevelDomain:
		return "domain"
	case PackageLevelEcosystem:
		return "ecosystem"
	case PackageLevelGlobal:
		return "global"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// Resolution Result Types
// ---------------------------------------------------------------------------

// PackageResolution contains the result of walking the hierarchy.
type PackageResolution struct {
	PackageName string                // resolved package name
	PackageType string                // "nvim" or "terminal"
	Source      PackageHierarchyLevel // which level resolved from
	SourceName  string                // e.g., "my-workspace" or "global default"
	SourceID    int                   // DB ID of the source entity
	Path        []PackageStep         // full walk trace for --show-cascade
}

// PackageStep represents one step in the hierarchy walk.
type PackageStep struct {
	Level       PackageHierarchyLevel
	ObjectID    int
	Name        string
	PackageName string // what was found at this level (may be empty)
	Found       bool
	Error       string
}

// ---------------------------------------------------------------------------
// Data Accessor Interface (narrow, testable)
// ---------------------------------------------------------------------------

// PackageDataAccessor provides the data needed for package hierarchy resolution.
// This is a narrow interface — implementations can wrap db.DataStore or be mocked.
type PackageDataAccessor interface {
	GetPackageEcosystemByID(id int) (PackageEcosystemData, error)
	GetPackageDomainByID(id int) (PackageDomainData, error)
	GetPackageAppByID(id int) (PackageAppData, error)
	GetPackageWorkspaceByID(id int) (PackageWorkspaceData, error)
	GetDefault(key string) (string, error)
}

// PackageEcosystemData holds the package-relevant fields from an ecosystem.
type PackageEcosystemData struct {
	ID              int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

// PackageDomainData holds the package-relevant fields from a domain.
type PackageDomainData struct {
	ID              int
	EcosystemID     int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

// PackageAppData holds the package-relevant fields from an app.
type PackageAppData struct {
	ID              int
	DomainID        int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

// PackageWorkspaceData holds the package-relevant fields from a workspace.
type PackageWorkspaceData struct {
	ID              int
	AppID           int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

// ---------------------------------------------------------------------------
// Resolver Interface
// ---------------------------------------------------------------------------

// PackageHierarchyResolver resolves effective packages by walking the hierarchy.
type PackageHierarchyResolver interface {
	ResolveNvimPackage(ctx context.Context, level PackageHierarchyLevel, objectID int) (*PackageResolution, error)
	ResolveTerminalPackage(ctx context.Context, level PackageHierarchyLevel, objectID int) (*PackageResolution, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type hierarchyPackageResolver struct {
	ds PackageDataAccessor
}

// NewHierarchyPackageResolver creates a resolver that walks the entity hierarchy.
func NewHierarchyPackageResolver(ds PackageDataAccessor) PackageHierarchyResolver {
	return &hierarchyPackageResolver{ds: ds}
}

// ResolveNvimPackage resolves the nvim package by walking up the hierarchy.
func (r *hierarchyPackageResolver) ResolveNvimPackage(ctx context.Context, level PackageHierarchyLevel, objectID int) (*PackageResolution, error) {
	return r.resolve(ctx, level, objectID, "nvim", func(nvim, _ sql.NullString) sql.NullString { return nvim })
}

// ResolveTerminalPackage resolves the terminal package by walking up the hierarchy.
func (r *hierarchyPackageResolver) ResolveTerminalPackage(ctx context.Context, level PackageHierarchyLevel, objectID int) (*PackageResolution, error) {
	return r.resolve(ctx, level, objectID, "terminal", func(_, term sql.NullString) sql.NullString { return term })
}

// resolve is the core walk logic, parameterized by a selector that picks
// nvim or terminal from the pair of NullString fields at each level.
func (r *hierarchyPackageResolver) resolve(
	ctx context.Context,
	level PackageHierarchyLevel,
	objectID int,
	pkgType string,
	selector func(nvim, term sql.NullString) sql.NullString,
) (*PackageResolution, error) {
	res := &PackageResolution{
		PackageType: pkgType,
		Path:        []PackageStep{},
	}

	curLevel := level
	curID := objectID

	for curLevel <= PackageLevelGlobal {
		step, parentID, parentLevel, err := r.resolveAtLevel(curLevel, curID, selector)
		if err != nil {
			step.Error = err.Error()
		}
		res.Path = append(res.Path, step)

		if step.Found && step.PackageName != "" {
			res.PackageName = step.PackageName
			res.Source = curLevel
			res.SourceName = step.Name
			res.SourceID = step.ObjectID
			return res, nil
		}

		curLevel = parentLevel
		curID = parentID
		if curLevel > PackageLevelGlobal {
			break
		}
	}

	// Nothing found anywhere
	return res, nil
}

// resolveAtLevel checks a single level and returns the step plus parent info.
func (r *hierarchyPackageResolver) resolveAtLevel(
	level PackageHierarchyLevel,
	objectID int,
	selector func(nvim, term sql.NullString) sql.NullString,
) (PackageStep, int, PackageHierarchyLevel, error) {
	step := PackageStep{Level: level, ObjectID: objectID}

	switch level {
	case PackageLevelWorkspace:
		return r.resolveWorkspace(objectID, step, selector)
	case PackageLevelApp:
		return r.resolveApp(objectID, step, selector)
	case PackageLevelDomain:
		return r.resolveDomain(objectID, step, selector)
	case PackageLevelEcosystem:
		return r.resolveEcosystem(objectID, step, selector)
	case PackageLevelGlobal:
		return r.resolveGlobal(step, selector)
	default:
		step.Error = "unknown hierarchy level"
		return step, 0, PackageLevelGlobal + 1, fmt.Errorf("unknown level: %d", level)
	}
}

func (r *hierarchyPackageResolver) resolveWorkspace(id int, step PackageStep, sel func(nvim, term sql.NullString) sql.NullString) (PackageStep, int, PackageHierarchyLevel, error) {
	ws, err := r.ds.GetPackageWorkspaceByID(id)
	if err != nil {
		return step, 0, PackageLevelGlobal, err
	}
	step.Name = ws.Name
	pkg := sel(ws.NvimPackage, ws.TerminalPackage)
	if pkg.Valid && pkg.String != "" {
		step.PackageName = pkg.String
		step.Found = true
	}
	return step, ws.AppID, PackageLevelApp, nil
}

func (r *hierarchyPackageResolver) resolveApp(id int, step PackageStep, sel func(nvim, term sql.NullString) sql.NullString) (PackageStep, int, PackageHierarchyLevel, error) {
	app, err := r.ds.GetPackageAppByID(id)
	if err != nil {
		return step, 0, PackageLevelGlobal, err
	}
	step.Name = app.Name
	pkg := sel(app.NvimPackage, app.TerminalPackage)
	if pkg.Valid && pkg.String != "" {
		step.PackageName = pkg.String
		step.Found = true
	}
	return step, app.DomainID, PackageLevelDomain, nil
}

func (r *hierarchyPackageResolver) resolveDomain(id int, step PackageStep, sel func(nvim, term sql.NullString) sql.NullString) (PackageStep, int, PackageHierarchyLevel, error) {
	dom, err := r.ds.GetPackageDomainByID(id)
	if err != nil {
		return step, 0, PackageLevelGlobal, err
	}
	step.Name = dom.Name
	pkg := sel(dom.NvimPackage, dom.TerminalPackage)
	if pkg.Valid && pkg.String != "" {
		step.PackageName = pkg.String
		step.Found = true
	}
	return step, dom.EcosystemID, PackageLevelEcosystem, nil
}

func (r *hierarchyPackageResolver) resolveEcosystem(id int, step PackageStep, sel func(nvim, term sql.NullString) sql.NullString) (PackageStep, int, PackageHierarchyLevel, error) {
	eco, err := r.ds.GetPackageEcosystemByID(id)
	if err != nil {
		return step, 0, PackageLevelGlobal, err
	}
	step.Name = eco.Name
	pkg := sel(eco.NvimPackage, eco.TerminalPackage)
	if pkg.Valid && pkg.String != "" {
		step.PackageName = pkg.String
		step.Found = true
	}
	return step, 0, PackageLevelGlobal, nil
}

func (r *hierarchyPackageResolver) resolveGlobal(step PackageStep, sel func(nvim, term sql.NullString) sql.NullString) (PackageStep, int, PackageHierarchyLevel, error) {
	step.Name = "global default"
	// Determine the defaults key based on which package type the selector picks
	// Test with nvim valid → if selector returns nvim, key is "nvim-package"
	testNvim := sql.NullString{String: "test", Valid: true}
	testTerm := sql.NullString{String: "test", Valid: true}
	key := "terminal-package"
	if sel(testNvim, sql.NullString{}).Valid {
		key = "nvim-package"
	} else if sel(sql.NullString{}, testTerm).Valid {
		key = "terminal-package"
	}

	val, err := r.ds.GetDefault(key)
	if err == nil && val != "" {
		step.PackageName = val
		step.Found = true
	}
	// After global, there's nowhere else to go
	return step, 0, PackageLevelGlobal + 1, nil
}

// Ensure hierarchyPackageResolver implements PackageHierarchyResolver.
var _ PackageHierarchyResolver = (*hierarchyPackageResolver)(nil)
