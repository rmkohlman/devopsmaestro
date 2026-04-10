package cmd

import (
	"fmt"

	"devopsmaestro/db"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"
	theme "github.com/rmkohlman/MaestroTheme"
	"github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

// themeEntry combines a theme with its source (library vs user) for unified listing.
type themeEntry struct {
	theme  *theme.Theme
	source string // "library" or "user"
}

func getThemes(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Collect all themes: library first, then user (user overrides library by name)
	entries, err := mergeLibraryAndUserThemes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list themes: %w", err)
	}

	if len(entries) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No themes found",
			EmptyHints:   []string{"dvm apply -f theme.yaml"},
		})
	}

	// For JSON/YAML, output the model data with source annotation
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return getThemesStructured(entries)
	}

	// For human output, build table with SOURCE column
	tableData := render.TableData{
		Headers: []string{"NAME", "SOURCE", "CATEGORY", "PLUGIN", "STYLE"},
		Rows:    make([][]string, len(entries)),
	}

	for i, e := range entries {
		category := e.theme.Category
		if category == "" {
			category = "-"
		}
		style := e.theme.Style
		if style == "" {
			style = "default"
		}

		tableData.Rows[i] = []string{
			e.theme.Name,
			e.source,
			category,
			e.theme.Plugin.Repo,
			style,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// mergeLibraryAndUserThemes combines library and user themes into a single list.
// User themes override library themes with the same name.
func mergeLibraryAndUserThemes(ctx resource.Context) ([]themeEntry, error) {
	// Start with library themes
	libraryInfos, err := library.List()
	if err != nil {
		// Non-fatal: continue with user themes only if library fails
		libraryInfos = nil
	}

	// Build a map for dedup: name → entry
	byName := make(map[string]themeEntry, len(libraryInfos))
	var order []string

	for _, info := range libraryInfos {
		t, err := library.Get(info.Name)
		if err != nil {
			continue
		}
		byName[info.Name] = themeEntry{theme: t, source: "library"}
		order = append(order, info.Name)
	}

	// Get user themes from DB store
	userResources, err := resource.List(ctx, handlers.KindNvimTheme)
	if err != nil {
		// If user store fails but we have library themes, return those
		if len(byName) > 0 {
			entries := make([]themeEntry, 0, len(order))
			for _, name := range order {
				entries = append(entries, byName[name])
			}
			return entries, nil
		}
		return nil, err
	}

	// Merge user themes — override library themes with same name
	for _, res := range userResources {
		tr := res.(*handlers.NvimThemeResource)
		t := tr.Theme()
		if _, exists := byName[t.Name]; !exists {
			order = append(order, t.Name)
		}
		byName[t.Name] = themeEntry{theme: t, source: "user"}
	}

	// Build final ordered list
	entries := make([]themeEntry, 0, len(order))
	for _, name := range order {
		entries = append(entries, byName[name])
	}
	return entries, nil
}

// getThemesStructured outputs themes in JSON/YAML with source annotation.
func getThemesStructured(entries []themeEntry) error {
	type themeWithSource struct {
		theme.ThemeYAML `yaml:",inline" json:",inline"`
		Source          string `yaml:"source" json:"source"`
	}

	output := make([]themeWithSource, len(entries))
	for i, e := range entries {
		output[i] = themeWithSource{
			ThemeYAML: theme.ThemeYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "NvimTheme",
				Metadata: theme.ThemeMetadata{
					Name:        e.theme.Name,
					Description: e.theme.Description,
					Author:      e.theme.Author,
					Category:    e.theme.Category,
				},
				Spec: theme.ThemeSpec{
					Plugin:      e.theme.Plugin,
					Style:       e.theme.Style,
					Transparent: e.theme.Transparent,
					Colors:      e.theme.Colors,
					Options:     e.theme.Options,
				},
			},
			Source: e.source,
		}
	}
	return render.OutputWith(getOutputFormat, output, render.Options{})
}

func getTheme(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindNvimTheme, name)
	if err != nil {
		return fmt.Errorf("failed to get theme '%s': %w", name, err)
	}

	t := res.(*handlers.NvimThemeResource).Theme()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		themeYAML := &theme.ThemeYAML{
			APIVersion: "devopsmaestro.io/v1",
			Kind:       "NvimTheme",
			Metadata: theme.ThemeMetadata{
				Name:        t.Name,
				Description: t.Description,
				Author:      t.Author,
				Category:    t.Category,
			},
			Spec: theme.ThemeSpec{
				Plugin:      t.Plugin,
				Style:       t.Style,
				Transparent: t.Transparent,
				Colors:      t.Colors,
				Options:     t.Options,
			},
		}
		return render.OutputWith(getOutputFormat, themeYAML, render.Options{})
	}

	// For human output, show detail view
	category := t.Category
	if category == "" {
		category = "-"
	}
	style := t.Style
	if style == "" {
		style = "default"
	}
	transparent := "no"
	if t.Transparent {
		transparent = "yes"
	}

	pairs := []render.KeyValue{
		{Key: "Name", Value: t.Name},
		{Key: "Plugin", Value: t.Plugin.Repo},
		{Key: "Category", Value: category},
		{Key: "Style", Value: style},
		{Key: "Transparent", Value: transparent},
	}

	// Show inherits field if set (from DB model)
	if inherits := getThemeInherits(ctx, t.Name); inherits != "" {
		pairs = append(pairs, render.KeyValue{Key: "Inherits", Value: inherits})
	}

	if t.Description != "" {
		pairs = append(pairs, render.KeyValue{Key: "Description", Value: t.Description})
	}
	if t.Author != "" {
		pairs = append(pairs, render.KeyValue{Key: "Author", Value: t.Author})
	}

	kvData := render.NewOrderedKeyValueData(pairs...)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Theme Details",
	})
}

// getEffectiveThemeDisplay shows the effective theme for the current context by resolving
// the hierarchy: workspace → app → domain → ecosystem → global default.
func getEffectiveThemeDisplay(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Determine the deepest active context level and object ID
	level, objectID, err := resolveActiveHierarchyLevel(ds)
	if err != nil {
		return err
	}

	// Create resolver and resolve the effective theme
	themeResolver := themeresolver.NewHierarchyThemeResolver(ds, nil)
	resolution, err := themeResolver.GetResolutionPath(cmd.Context(), level, objectID)
	if err != nil {
		return fmt.Errorf("failed to resolve effective theme: %w", err)
	}

	effectiveTheme := resolution.GetEffectiveThemeName()
	sourceDesc := resolution.GetSourceDescription()

	// For structured output (JSON/YAML)
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		type resolutionStep struct {
			Level    string `json:"level" yaml:"level"`
			Name     string `json:"name" yaml:"name"`
			Theme    string `json:"theme,omitempty" yaml:"theme,omitempty"`
			HasTheme bool   `json:"hasTheme" yaml:"hasTheme"`
		}

		steps := make([]resolutionStep, 0, len(resolution.Path))
		for _, step := range resolution.Path {
			steps = append(steps, resolutionStep{
				Level:    step.Level.String(),
				Name:     step.Name,
				Theme:    step.ThemeName,
				HasTheme: step.Found && step.ThemeName != "",
			})
		}

		data := struct {
			EffectiveTheme string           `json:"effectiveTheme" yaml:"effectiveTheme"`
			Source         string           `json:"source" yaml:"source"`
			SourceLevel    string           `json:"sourceLevel" yaml:"sourceLevel"`
			ResolutionPath []resolutionStep `json:"resolutionPath,omitempty" yaml:"resolutionPath,omitempty"`
		}{
			EffectiveTheme: effectiveTheme,
			Source:         sourceDesc,
			SourceLevel:    resolution.Source.String(),
			ResolutionPath: steps,
		}
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	// Build human-readable key-value display
	pairs := []render.KeyValue{
		{Key: "Effective Theme", Value: effectiveTheme},
		{Key: "Source", Value: sourceDesc},
	}

	// Show resolution path
	if len(resolution.Path) > 0 {
		for _, step := range resolution.Path {
			label := fmt.Sprintf("  %s '%s'", step.Level.String(), step.Name)
			if step.Found && step.ThemeName != "" {
				pairs = append(pairs, render.KeyValue{Key: label, Value: step.ThemeName})
			} else {
				pairs = append(pairs, render.KeyValue{Key: label, Value: "(inherits)"})
			}
		}
	}

	kvData := render.NewOrderedKeyValueData(pairs...)
	if err := render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Effective Theme",
	}); err != nil {
		return err
	}

	render.Blank()
	render.Info(fmt.Sprintf("Tip: Override with 'dvm set theme <name>' or 'dvm set theme <name> --workspace <ws>'"))

	return nil
}

// resolveActiveHierarchyLevel determines the deepest active context level from the DB.
// Returns the hierarchy level and object ID to start resolution from.
func resolveActiveHierarchyLevel(ds db.DataStore) (themeresolver.HierarchyLevel, int, error) {
	dbCtx, err := ds.GetContext()
	if err != nil || dbCtx == nil {
		// No context at all — resolve from global
		return themeresolver.LevelGlobal, 0, nil
	}

	// Try deepest first: workspace → app → ecosystem → global
	if dbCtx.ActiveWorkspaceID != nil {
		return themeresolver.LevelWorkspace, *dbCtx.ActiveWorkspaceID, nil
	}
	if dbCtx.ActiveAppID != nil {
		return themeresolver.LevelApp, *dbCtx.ActiveAppID, nil
	}
	if dbCtx.ActiveEcosystemID != nil {
		return themeresolver.LevelEcosystem, *dbCtx.ActiveEcosystemID, nil
	}

	return themeresolver.LevelGlobal, 0, nil
}

// showThemeResolution displays theme resolution information for a given hierarchy level and object ID
func showThemeResolution(cmd *cobra.Command, ds db.DataStore, level themeresolver.HierarchyLevel, objectID int, objectName string) error {
	// Create theme resolver
	themeResolver, err := themeresolver.NewThemeResolver(ds, nil) // nil theme store for now
	if err != nil {
		return fmt.Errorf("failed to create theme resolver: %w", err)
	}

	// Get theme resolution path
	resolution, err := themeResolver.GetResolutionPath(cmd.Context(), level, objectID)
	if err != nil {
		return fmt.Errorf("failed to resolve theme: %w", err)
	}

	render.Blank()
	render.Info("Theme Resolution:")

	if resolution.Source != themeresolver.LevelGlobal {
		render.Plainf("  Effective theme: %s", resolution.GetEffectiveThemeName())
		render.Plainf("  Source: %s", resolution.GetSourceDescription())
	} else {
		render.Plainf("  Effective theme: %s (default)", themeresolver.DefaultTheme)
		render.Plain("  Source: global default")
	}

	if len(resolution.Path) > 0 {
		render.Plain("  Resolution path:")
		for _, step := range resolution.Path {
			status := "○" // Empty circle
			if step.Found && step.ThemeName != "" {
				status = "●" // Filled circle
			}

			line := fmt.Sprintf("    %s %s '%s'", status, step.Level.String(), step.Name)
			if step.ThemeName != "" {
				line += fmt.Sprintf(" → %s", step.ThemeName)
			}
			if step.Error != "" {
				line += fmt.Sprintf(" (error: %s)", step.Error)
			}
			render.Plain(line)
		}
	}

	render.Blank()
	render.Info("Legend: ● theme set, ○ no theme (inherits from parent)")

	return nil
}

// getThemeInherits looks up the inherits field for a theme from the DB model.
// Returns empty string if not found or no inheritance.
func getThemeInherits(ctx resource.Context, themeName string) string {
	if ctx.DataStore == nil {
		return ""
	}
	ds, ok := ctx.DataStore.(db.DataStore)
	if !ok {
		return ""
	}
	dbTheme, err := ds.GetThemeByName(themeName)
	if err != nil {
		return ""
	}
	if dbTheme.Inherits.Valid {
		return dbTheme.Inherits.String
	}
	return ""
}
