// Package resolver provides hierarchical theme resolution implementation
package resolver

import (
	"context"
	"fmt"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/nvimops/theme/library"
)

// HierarchyThemeResolver implements ThemeResolver with database access
type HierarchyThemeResolver struct {
	dataStore    db.DataStore
	themeStore   theme.Store
	defaultTheme string
}

// NewHierarchyThemeResolver creates a new hierarchy theme resolver
func NewHierarchyThemeResolver(dataStore db.DataStore, themeStore theme.Store) *HierarchyThemeResolver {
	return &HierarchyThemeResolver{
		dataStore:    dataStore,
		themeStore:   themeStore,
		defaultTheme: DefaultTheme,
	}
}

// Resolve walks the hierarchy from the starting level upward
func (r *HierarchyThemeResolver) Resolve(ctx context.Context, level HierarchyLevel, objectID int) (*ThemeResolution, error) {
	resolution := &ThemeResolution{
		Path:       []ThemeStep{},
		ResolvedAt: time.Now(),
	}

	walker := &hierarchyWalker{
		level:      level,
		objectID:   objectID,
		resolution: resolution,
		dataStore:  r.dataStore,
	}

	return r.walkHierarchy(ctx, walker)
}

// ResolveDefault returns the global default theme
func (r *HierarchyThemeResolver) ResolveDefault() (*ThemeResolution, error) {
	resolution := &ThemeResolution{
		Source:     LevelGlobal,
		SourceName: "global default",
		SourceID:   0,
		Path: []ThemeStep{
			{
				Level:     LevelGlobal,
				Name:      "global default",
				ThemeName: r.defaultTheme,
				Found:     true,
			},
		},
		ResolvedAt: time.Now(),
	}

	// Load the default theme
	theme, err := r.loadTheme(r.defaultTheme)
	if err != nil {
		resolution.Path[0].Error = fmt.Sprintf("failed to load default theme: %v", err)
		return resolution, err
	}

	resolution.Theme = theme
	return resolution, nil
}

// GetResolutionPath returns the complete resolution trace without loading themes
func (r *HierarchyThemeResolver) GetResolutionPath(ctx context.Context, level HierarchyLevel, objectID int) (*ThemeResolution, error) {
	resolution := &ThemeResolution{
		Path:       []ThemeStep{},
		ResolvedAt: time.Now(),
	}

	walker := &hierarchyWalker{
		level:      level,
		objectID:   objectID,
		resolution: resolution,
		dataStore:  r.dataStore,
	}

	// Walk hierarchy but don't load themes, just trace the path
	return r.walkHierarchyTrace(ctx, walker)
}

// hierarchyWalker manages the state during hierarchy traversal
type hierarchyWalker struct {
	level      HierarchyLevel
	objectID   int
	resolution *ThemeResolution
	dataStore  db.DataStore
}

// walkHierarchy implements the template method pattern for hierarchy walking
func (r *HierarchyThemeResolver) walkHierarchy(ctx context.Context, walker *hierarchyWalker) (*ThemeResolution, error) {
	for walker.level <= LevelGlobal {
		step := r.resolveAtLevel(ctx, walker.level, walker.objectID)
		walker.resolution.Path = append(walker.resolution.Path, step)

		// If we found a theme at this level, try to load it
		if step.Found && step.ThemeName != "" {
			theme, err := r.loadTheme(step.ThemeName)
			if err == nil {
				// Successfully loaded theme
				return r.buildResolution(theme, walker, step), nil
			}
			// Theme loading failed, update step and continue
			step.Error = fmt.Sprintf("theme loading failed: %v", err)
			walker.resolution.Path[len(walker.resolution.Path)-1] = step
		}

		// Move up hierarchy
		parentID, parentLevel := r.getParent(ctx, walker.level, walker.objectID)
		walker.objectID = parentID
		walker.level = parentLevel

		if walker.level > LevelGlobal {
			break
		}
	}

	// No theme found in hierarchy, use default
	return r.ResolveDefault()
}

// walkHierarchyTrace walks hierarchy for tracing only (no theme loading)
func (r *HierarchyThemeResolver) walkHierarchyTrace(ctx context.Context, walker *hierarchyWalker) (*ThemeResolution, error) {
	for walker.level <= LevelGlobal {
		step := r.resolveAtLevel(ctx, walker.level, walker.objectID)
		walker.resolution.Path = append(walker.resolution.Path, step)

		// If we found a theme at this level, mark it as the effective source
		if step.Found && step.ThemeName != "" {
			walker.resolution.Source = walker.level
			walker.resolution.SourceName = step.Name
			walker.resolution.SourceID = step.ObjectID
			// Don't load theme, just return the path
			return walker.resolution, nil
		}

		// Move up hierarchy
		parentID, parentLevel := r.getParent(ctx, walker.level, walker.objectID)
		walker.objectID = parentID
		walker.level = parentLevel

		if walker.level > LevelGlobal {
			break
		}
	}

	// No theme found in hierarchy, mark default as source
	walker.resolution.Source = LevelGlobal
	walker.resolution.SourceName = "global default"
	walker.resolution.SourceID = 0
	walker.resolution.Path = append(walker.resolution.Path, ThemeStep{
		Level:     LevelGlobal,
		Name:      "global default",
		ThemeName: r.defaultTheme,
		Found:     true,
	})

	return walker.resolution, nil
}

// resolveAtLevel checks for a theme at the specified hierarchy level
func (r *HierarchyThemeResolver) resolveAtLevel(ctx context.Context, level HierarchyLevel, objectID int) ThemeStep {
	step := ThemeStep{
		Level:    level,
		ObjectID: objectID,
	}

	switch level {
	case LevelWorkspace:
		return r.resolveWorkspaceTheme(ctx, objectID, step)
	case LevelApp:
		return r.resolveAppTheme(ctx, objectID, step)
	case LevelDomain:
		return r.resolveDomainTheme(ctx, objectID, step)
	case LevelEcosystem:
		return r.resolveEcosystemTheme(ctx, objectID, step)
	case LevelGlobal:
		step.Name = "global default"
		step.ThemeName = r.defaultTheme
		step.Found = true
		return step
	default:
		step.Error = "unknown hierarchy level"
		return step
	}
}

// resolveWorkspaceTheme resolves theme from workspace NvimConfig
func (r *HierarchyThemeResolver) resolveWorkspaceTheme(ctx context.Context, workspaceID int, step ThemeStep) ThemeStep {
	workspace, err := r.dataStore.GetWorkspaceByID(workspaceID)
	if err != nil {
		step.Error = fmt.Sprintf("workspace not found: %v", err)
		return step
	}

	step.Name = workspace.Name

	// Parse workspace's NvimStructure to get theme configuration
	// In the current implementation, we need to check if there's theme info
	// This is a simplified approach - in a full implementation, we'd parse the NvimConfig YAML
	if workspace.NvimStructure.Valid && workspace.NvimStructure.String != "" {
		// TODO: Parse NvimConfig YAML to extract theme
		// For now, we'll assume no workspace-level theme override
		step.Found = false
		return step
	}

	step.Found = false
	return step
}

// resolveAppTheme resolves theme from app
func (r *HierarchyThemeResolver) resolveAppTheme(ctx context.Context, appID int, step ThemeStep) ThemeStep {
	app, err := r.dataStore.GetAppByID(appID)
	if err != nil {
		step.Error = fmt.Sprintf("app not found: %v", err)
		return step
	}

	step.Name = app.Name

	if app.Theme.Valid && app.Theme.String != "" {
		step.ThemeName = app.Theme.String
		step.Found = true
	} else {
		step.Found = false
	}

	return step
}

// resolveDomainTheme resolves theme from domain
func (r *HierarchyThemeResolver) resolveDomainTheme(ctx context.Context, domainID int, step ThemeStep) ThemeStep {
	domain, err := r.dataStore.GetDomainByID(domainID)
	if err != nil {
		step.Error = fmt.Sprintf("domain not found: %v", err)
		return step
	}

	step.Name = domain.Name

	if domain.Theme.Valid && domain.Theme.String != "" {
		step.ThemeName = domain.Theme.String
		step.Found = true
	} else {
		step.Found = false
	}

	return step
}

// resolveEcosystemTheme resolves theme from ecosystem
func (r *HierarchyThemeResolver) resolveEcosystemTheme(ctx context.Context, ecosystemID int, step ThemeStep) ThemeStep {
	ecosystem, err := r.dataStore.GetEcosystemByID(ecosystemID)
	if err != nil {
		step.Error = fmt.Sprintf("ecosystem not found: %v", err)
		return step
	}

	step.Name = ecosystem.Name

	if ecosystem.Theme.Valid && ecosystem.Theme.String != "" {
		step.ThemeName = ecosystem.Theme.String
		step.Found = true
	} else {
		step.Found = false
	}

	return step
}

// getParent returns the parent object ID and level for the given hierarchy level
func (r *HierarchyThemeResolver) getParent(ctx context.Context, level HierarchyLevel, objectID int) (int, HierarchyLevel) {
	switch level {
	case LevelWorkspace:
		// Get workspace's app ID
		if workspace, err := r.dataStore.GetWorkspaceByID(objectID); err == nil {
			return workspace.AppID, LevelApp
		}
		return 0, LevelGlobal
	case LevelApp:
		// Get app's domain ID
		if app, err := r.dataStore.GetAppByID(objectID); err == nil {
			return app.DomainID, LevelDomain
		}
		return 0, LevelGlobal
	case LevelDomain:
		// Get domain's ecosystem ID
		if domain, err := r.dataStore.GetDomainByID(objectID); err == nil {
			return domain.EcosystemID, LevelEcosystem
		}
		return 0, LevelGlobal
	case LevelEcosystem:
		return 0, LevelGlobal
	default:
		return 0, LevelGlobal
	}
}

// loadTheme loads a theme from the theme store or library
func (r *HierarchyThemeResolver) loadTheme(name string) (*theme.Theme, error) {
	// Try theme store first (custom themes)
	if r.themeStore != nil {
		theme, err := r.themeStore.Get(name)
		if err == nil {
			return theme, nil
		}
	}

	// Try library as fallback (built-in themes)
	theme, err := library.Get(name)
	if err != nil {
		return nil, fmt.Errorf("theme %q not found in store or library: %v", name, err)
	}

	return theme, nil
}

// buildResolution creates the final resolution result
func (r *HierarchyThemeResolver) buildResolution(theme *theme.Theme, walker *hierarchyWalker, step ThemeStep) *ThemeResolution {
	walker.resolution.Theme = theme
	walker.resolution.Source = walker.level
	walker.resolution.SourceName = step.Name
	walker.resolution.SourceID = step.ObjectID
	return walker.resolution
}
