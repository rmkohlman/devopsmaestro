package main

import (
	"encoding/json"
	"fmt"

	"github.com/rmkohlman/MaestroSDK/render"
	theme "github.com/rmkohlman/MaestroTheme"
	themelibrary "github.com/rmkohlman/MaestroTheme/library"
	"gopkg.in/yaml.v3"
)

func getThemeStore() *theme.FileStore {
	dir := getConfigDir()
	return theme.NewFileStore(dir)
}

func outputThemes(themes []*theme.Theme, format string, activeName string) error {
	switch format {
	case "yaml":
		for i, t := range themes {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := t.ToYAML()
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		data, err := json.MarshalIndent(themes, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "CATEGORY", "PLUGIN", "ACTIVE", "DESCRIPTION")
		for _, t := range themes {
			active := ""
			if t.Name == activeName {
				active = "*"
			}
			tb.AddRow(t.Name, t.Category, t.Plugin.Repo, active, render.Truncate(t.Description, 35))
		}
		return render.OutputWith(format, tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputTheme(t *theme.Theme, format string) error {
	switch format {
	case "yaml", "":
		data, err := t.ToYAML()
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputThemeInfos(themes []themelibrary.ThemeInfo, format string) error {
	switch format {
	case "yaml":
		data, err := yaml.Marshal(themes)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(themes, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "CATEGORY", "PLUGIN", "DESCRIPTION")
		for _, t := range themes {
			tb.AddRow(t.Name, t.Category, t.Plugin, render.Truncate(t.Description, 40))
		}
		return render.OutputWith(format, tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}
