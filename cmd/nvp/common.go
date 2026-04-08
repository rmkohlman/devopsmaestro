package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rmkohlman/MaestroNvim/nvimops"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroNvim/nvimops/store"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// getConfigDir returns the nvp configuration directory.
func getConfigDir() string {
	if configDir != "" {
		return configDir
	}
	if dir := os.Getenv("NVP_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return paths.New(home).NVPRoot()
}

// getManager creates an nvimops Manager backed by the file store.
func getManager() (nvimops.Manager, error) {
	dir := getConfigDir()
	pluginsDir := filepath.Join(dir, "plugins")

	// Auto-create if doesn't exist
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	fileStore, err := store.NewFileStore(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return nvimops.NewWithOptions(nvimops.Options{
		Store: fileStore,
	})
}

// outputPlugins formats and prints a list of plugins.
func outputPlugins(plugins []*plugin.Plugin, format string) error {
	// Sort by name
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range plugins {
			if i > 0 {
				fmt.Println("---")
			}
			yml := p.ToYAML()
			data, err := yaml.Marshal(yml)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*plugin.PluginYAML
		for _, p := range plugins {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "CATEGORY", "ENABLED", "DESCRIPTION")
		for _, p := range plugins {
			enabled := "yes"
			if !p.Enabled {
				enabled = "no"
			}
			tb.AddRow(p.Name, p.Category, enabled, render.Truncate(p.Description, 40))
		}
		return render.OutputWith(format, tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// outputPlugin formats and prints a single plugin.
func outputPlugin(p *plugin.Plugin, format string) error {
	switch format {
	case "yaml", "":
		yml := p.ToYAML()
		data, err := yaml.Marshal(yml)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		yml := p.ToYAML()
		data, err := json.MarshalIndent(yml, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// hiddenAlias creates a hidden command that delegates to the target command.
// Used to keep deprecated verb names (list, show, install) working without
// showing them in --help output.
func hiddenAlias(name string, target *cobra.Command) *cobra.Command {
	alias := *target
	alias.Use = name
	alias.Aliases = nil
	alias.Hidden = true
	alias.Short = target.Short + " (deprecated: use " + target.Name() + ")"
	alias.Deprecated = "use '" + target.Name() + "' instead"
	return &alias
}
