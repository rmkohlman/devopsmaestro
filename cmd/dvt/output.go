package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tCATEGORY\tDESCRIPTION")
		for _, p := range prompts {
			desc := p.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Type, p.Category, desc)
		}
		w.Flush()
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tREPO\tMANAGER\tDESCRIPTION")
		for _, p := range plugins {
			desc := p.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Repo, p.Manager, desc)
		}
		w.Flush()
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSHELL\tALIASES\tENV VARS\tFUNCTIONS")
		for _, s := range shells {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\n",
				s.Name, s.ShellType, len(s.Aliases), len(s.Env), len(s.Functions))
		}
		w.Flush()
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPROMPT\tPLUGINS\tDESCRIPTION")
		for _, p := range profiles {
			desc := p.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			promptName := ""
			if p.Prompt != nil {
				promptName = p.Prompt.Name
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", p.Name, promptName, len(p.Plugins), desc)
		}
		w.Flush()
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
