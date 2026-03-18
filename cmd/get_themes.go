package cmd

import (
	"fmt"

	"devopsmaestro/db"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"
	theme "github.com/rmkohlman/MaestroTheme"

	"github.com/spf13/cobra"
)

func getThemes(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindNvimTheme)
	if err != nil {
		return fmt.Errorf("failed to list themes: %w", err)
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No user themes found (34+ library themes available automatically)",
			EmptyHints:   []string{"dvm get nvim theme coolnight-ocean", "dvm apply -f theme.yaml"},
		})
	}

	// Extract underlying themes from resources
	themes := make([]*theme.Theme, len(resources))
	for i, res := range resources {
		tr := res.(*handlers.NvimThemeResource)
		themes[i] = tr.Theme()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		themesYAML := make([]*theme.ThemeYAML, len(themes))
		for i, t := range themes {
			themesYAML[i] = &theme.ThemeYAML{
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
		}
		return render.OutputWith(getOutputFormat, themesYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "CATEGORY", "PLUGIN", "STYLE"},
		Rows:    make([][]string, len(themes)),
	}

	for i, t := range themes {
		category := t.Category
		if category == "" {
			category = "-"
		}
		style := t.Style
		if style == "" {
			style = "default"
		}

		tableData.Rows[i] = []string{
			t.Name,
			category,
			t.Plugin.Repo,
			style,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
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
