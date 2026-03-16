package cmd

import (
	"fmt"
	"strconv"

	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// AllResources represents all resources for JSON/YAML output.
type AllResources struct {
	Ecosystems  []AllResourceSummary `json:"ecosystems" yaml:"ecosystems"`
	Domains     []AllResourceSummary `json:"domains" yaml:"domains"`
	Apps        []AllResourceSummary `json:"apps" yaml:"apps"`
	Workspaces  []AllResourceSummary `json:"workspaces" yaml:"workspaces"`
	Credentials []AllResourceSummary `json:"credentials" yaml:"credentials"`
	Registries  []AllResourceSummary `json:"registries" yaml:"registries"`
	GitRepos    []AllResourceSummary `json:"gitRepos" yaml:"gitRepos"`
	NvimPlugins []AllResourceSummary `json:"nvimPlugins" yaml:"nvimPlugins"`
	NvimThemes  []AllResourceSummary `json:"nvimThemes" yaml:"nvimThemes"`
}

// AllResourceSummary represents a single resource in the "get all" output.
type AllResourceSummary struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
	Repo        string `json:"repo,omitempty" yaml:"repo,omitempty"`
	Category    string `json:"category,omitempty" yaml:"category,omitempty"`
}

// getAllCmd shows all resources across the system.
var getAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Show all resources",
	Long: `Show a summary of all resources across the system.

Displays all ecosystems, domains, apps, workspaces, credentials,
registries, git repos, nvim plugins, and nvim themes.

Examples:
  dvm get all              # Show all resources (human-readable)
  dvm get all -o json      # Output as JSON
  dvm get all -o yaml      # Output as YAML
  dvm get all -o table     # Output as plain table`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAll(cmd)
	},
}

func getAll(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("failed to get data store: %w", err)
	}

	// Collect all resources, treating errors as empty sections
	ecosystems, err := ds.ListEcosystems()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list ecosystems: %v", err))
		ecosystems = nil
	}

	domains, err := ds.ListAllDomains()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list domains: %v", err))
		domains = nil
	}

	apps, err := ds.ListAllApps()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list apps: %v", err))
		apps = nil
	}

	workspaces, err := ds.ListAllWorkspaces()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list workspaces: %v", err))
		workspaces = nil
	}

	credentials, err := ds.ListAllCredentials()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list credentials: %v", err))
		credentials = nil
	}

	registries, err := ds.ListRegistries()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list registries: %v", err))
		registries = nil
	}

	gitRepos, err := ds.ListGitRepos()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list git repos: %v", err))
		gitRepos = nil
	}

	plugins, err := ds.ListPlugins()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list nvim plugins: %v", err))
		plugins = nil
	}

	themes, err := ds.ListThemes()
	if err != nil {
		render.Warning(fmt.Sprintf("failed to list nvim themes: %v", err))
		themes = nil
	}

	// JSON/YAML: build a single composite struct and render once
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		all := AllResources{
			Ecosystems:  make([]AllResourceSummary, 0, len(ecosystems)),
			Domains:     make([]AllResourceSummary, 0, len(domains)),
			Apps:        make([]AllResourceSummary, 0, len(apps)),
			Workspaces:  make([]AllResourceSummary, 0, len(workspaces)),
			Credentials: make([]AllResourceSummary, 0, len(credentials)),
			Registries:  make([]AllResourceSummary, 0, len(registries)),
			GitRepos:    make([]AllResourceSummary, 0, len(gitRepos)),
			NvimPlugins: make([]AllResourceSummary, 0, len(plugins)),
			NvimThemes:  make([]AllResourceSummary, 0, len(themes)),
		}

		for _, e := range ecosystems {
			desc := ""
			if e.Description.Valid {
				desc = e.Description.String
			}
			all.Ecosystems = append(all.Ecosystems, AllResourceSummary{
				Name:        e.Name,
				Description: desc,
			})
		}

		for _, d := range domains {
			desc := ""
			if d.Description.Valid {
				desc = d.Description.String
			}
			all.Domains = append(all.Domains, AllResourceSummary{
				Name:        d.Name,
				Description: desc,
			})
		}

		for _, a := range apps {
			desc := ""
			if a.Description.Valid {
				desc = a.Description.String
			}
			lang := ""
			if a.Language.Valid {
				lang = a.Language.String
			}
			all.Apps = append(all.Apps, AllResourceSummary{
				Name:        a.Name,
				Description: desc,
				Type:        lang,
			})
		}

		for _, w := range workspaces {
			all.Workspaces = append(all.Workspaces, AllResourceSummary{
				Name:   w.Name,
				Status: w.Status,
				Type:   w.ImageName,
			})
		}

		for _, c := range credentials {
			desc := ""
			if c.Description != nil {
				desc = *c.Description
			}
			all.Credentials = append(all.Credentials, AllResourceSummary{
				Name:        c.Name,
				Description: desc,
				Type:        string(c.ScopeType),
			})
		}

		for _, r := range registries {
			desc := ""
			if r.Description.Valid {
				desc = r.Description.String
			}
			all.Registries = append(all.Registries, AllResourceSummary{
				Name:        r.Name,
				Description: desc,
				Type:        r.Type,
				Status:      r.Status,
			})
		}

		for _, g := range gitRepos {
			all.GitRepos = append(all.GitRepos, AllResourceSummary{
				Name: g.Name,
				URL:  g.URL,
				Type: g.AuthType,
			})
		}

		for _, p := range plugins {
			cat := ""
			if p.Category.Valid {
				cat = p.Category.String
			}
			all.NvimPlugins = append(all.NvimPlugins, AllResourceSummary{
				Name:     p.Name,
				Repo:     p.Repo,
				Category: cat,
			})
		}

		for _, t := range themes {
			active := ""
			if t.IsActive {
				active = "yes"
			}
			all.NvimThemes = append(all.NvimThemes, AllResourceSummary{
				Name:   t.Name,
				Repo:   t.PluginRepo,
				Status: active,
			})
		}

		return render.OutputWith(getOutputFormat, all, render.Options{Type: render.TypeAuto})
	}

	// Human-readable output: render each section separately

	// === Ecosystems ===
	render.Info("=== Ecosystems ===")
	if len(ecosystems) > 0 {
		rows := make([][]string, 0, len(ecosystems))
		for _, e := range ecosystems {
			desc := ""
			if e.Description.Valid {
				desc = e.Description.String
			}
			rows = append(rows, []string{e.Name, desc})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "DESCRIPTION"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Domains ===
	render.Info("=== Domains ===")
	if len(domains) > 0 {
		rows := make([][]string, 0, len(domains))
		for _, d := range domains {
			desc := ""
			if d.Description.Valid {
				desc = d.Description.String
			}
			rows = append(rows, []string{d.Name, desc})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "DESCRIPTION"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Apps ===
	render.Info("=== Apps ===")
	if len(apps) > 0 {
		rows := make([][]string, 0, len(apps))
		for _, a := range apps {
			lang := ""
			if a.Language.Valid {
				lang = a.Language.String
			}
			desc := ""
			if a.Description.Valid {
				desc = a.Description.String
			}
			rows = append(rows, []string{a.Name, lang, desc})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "LANGUAGE", "DESCRIPTION"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Workspaces ===
	render.Info("=== Workspaces ===")
	if len(workspaces) > 0 {
		rows := make([][]string, 0, len(workspaces))
		for _, w := range workspaces {
			rows = append(rows, []string{w.Name, w.ImageName, w.Status})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "IMAGE", "STATUS"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Credentials ===
	render.Info("=== Credentials ===")
	if len(credentials) > 0 {
		rows := make([][]string, 0, len(credentials))
		for _, c := range credentials {
			rows = append(rows, []string{c.Name, string(c.ScopeType), c.Source})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "SCOPE", "SOURCE"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Registries ===
	render.Info("=== Registries ===")
	if len(registries) > 0 {
		rows := make([][]string, 0, len(registries))
		for _, r := range registries {
			rows = append(rows, []string{r.Name, r.Type, r.Status, strconv.Itoa(r.Port)})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "TYPE", "STATUS", "PORT"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Git Repos ===
	render.Info("=== Git Repos ===")
	if len(gitRepos) > 0 {
		rows := make([][]string, 0, len(gitRepos))
		for _, g := range gitRepos {
			rows = append(rows, []string{g.Name, g.URL, g.AuthType})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "URL", "AUTH"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Nvim Plugins ===
	render.Info("=== Nvim Plugins ===")
	if len(plugins) > 0 {
		rows := make([][]string, 0, len(plugins))
		for _, p := range plugins {
			cat := ""
			if p.Category.Valid {
				cat = p.Category.String
			}
			rows = append(rows, []string{p.Name, p.Repo, cat})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "REPO", "CATEGORY"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}
	render.Blank()

	// === Nvim Themes ===
	render.Info("=== Nvim Themes ===")
	if len(themes) > 0 {
		rows := make([][]string, 0, len(themes))
		for _, t := range themes {
			active := "no"
			if t.IsActive {
				active = "yes"
			}
			rows = append(rows, []string{t.Name, t.PluginRepo, active})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "PLUGIN REPO", "ACTIVE"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
	} else {
		render.Plainf("  (none)")
	}

	return nil
}
