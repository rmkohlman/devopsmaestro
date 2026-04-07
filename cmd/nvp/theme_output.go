package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tPLUGIN\tACTIVE\tDESCRIPTION")
		for _, t := range themes {
			active := ""
			if t.Name == activeName {
				active = "*"
			}
			desc := t.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Name, t.Category, t.Plugin.Repo, active, desc)
		}
		w.Flush()
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tPLUGIN\tDESCRIPTION")
		for _, t := range themes {
			desc := t.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Name, t.Category, t.Plugin, desc)
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}
