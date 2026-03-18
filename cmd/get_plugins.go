package cmd

import (
	"fmt"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

func getPlugins(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindNvimPlugin)
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No plugins found",
			EmptyHints:   []string{"dvm apply -f plugin.yaml"},
		})
	}

	// Extract underlying plugins from resources
	plugins := make([]*plugin.Plugin, len(resources))
	for i, res := range resources {
		pr := res.(*handlers.NvimPluginResource)
		plugins[i] = pr.Plugin()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		pluginsYAML := make([]*plugin.PluginYAML, len(plugins))
		for i, p := range plugins {
			pluginsYAML[i] = p.ToYAML()
		}
		return render.OutputWith(getOutputFormat, pluginsYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "CATEGORY", "REPO", "VERSION"},
		Rows:    make([][]string, len(plugins)),
	}

	for i, p := range plugins {
		version := "latest"
		if p.Version != "" {
			version = p.Version
		} else if p.Branch != "" {
			version = "branch:" + p.Branch
		}

		enabledMark := "✓"
		if !p.Enabled {
			enabledMark = "✗"
		}

		tableData.Rows[i] = []string{
			p.Name + " " + enabledMark,
			p.Category,
			p.Repo,
			version,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

func getPlugin(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindNvimPlugin, name)
	if err != nil {
		return fmt.Errorf("failed to get plugin '%s': %w", name, err)
	}

	p := res.(*handlers.NvimPluginResource).Plugin()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, p.ToYAML(), render.Options{})
	}

	// For human output, show detail view
	version := "latest"
	if p.Version != "" {
		version = p.Version
	} else if p.Branch != "" {
		version = "branch:" + p.Branch
	}

	enabledStr := "yes"
	if !p.Enabled {
		enabledStr = "no"
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: p.Name},
		render.KeyValue{Key: "Repo", Value: p.Repo},
		render.KeyValue{Key: "Category", Value: p.Category},
		render.KeyValue{Key: "Version", Value: version},
		render.KeyValue{Key: "Enabled", Value: enabledStr},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Plugin Details",
	})
}
