package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/profile"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// OUTPUT HELPERS
// =============================================================================

func outputPrompts(prompts []*prompt.Prompt, format string) error {
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range prompts {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(p.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*prompt.PromptYAML
		for _, p := range prompts {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "TYPE", "CATEGORY", "DESCRIPTION")
		for _, p := range prompts {
			tb.AddRow(p.Name, string(p.Type), p.Category, render.Truncate(p.Description, 40))
		}
		return render.OutputWith("", tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputPrompt(p *prompt.Prompt, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(p.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(p.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputPlugins(plugins []*plugin.Plugin, format string) error {
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range plugins {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(p.ToYAML())
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
		tb := render.NewTableBuilder("NAME", "REPO", "MANAGER", "DESCRIPTION")
		for _, p := range plugins {
			tb.AddRow(p.Name, p.Repo, string(p.Manager), render.Truncate(p.Description, 35))
		}
		return render.OutputWith("", tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputPlugin(p *plugin.Plugin, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(p.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(p.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputShells(shells []*shell.Shell, format string) error {
	sort.Slice(shells, func(i, j int) bool {
		return shells[i].Name < shells[j].Name
	})

	switch format {
	case "yaml":
		for i, s := range shells {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(s.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*shell.ShellYAML
		for _, s := range shells {
			items = append(items, s.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "SHELL", "ALIASES", "ENV VARS", "FUNCTIONS")
		for _, s := range shells {
			tb.AddRow(
				s.Name, string(s.ShellType),
				fmt.Sprintf("%d", len(s.Aliases)),
				fmt.Sprintf("%d", len(s.Env)),
				fmt.Sprintf("%d", len(s.Functions)),
			)
		}
		return render.OutputWith("", tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputShell(s *shell.Shell, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(s.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(s.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputProfiles(profiles []*profile.Profile, format string) error {
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range profiles {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(p.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*profile.ProfileYAML
		for _, p := range profiles {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		tb := render.NewTableBuilder("NAME", "PROMPT", "PLUGINS", "DESCRIPTION")
		for _, p := range profiles {
			promptName := ""
			if p.Prompt != nil {
				promptName = p.Prompt.Name
			}
			tb.AddRow(p.Name, promptName, fmt.Sprintf("%d", len(p.Plugins)), render.Truncate(p.Description, 35))
		}
		return render.OutputWith("", tb.Build(), render.Options{Type: render.TypeTable})
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputProfile(p *profile.Profile, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(p.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(p.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}
