package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops"
	nvimconfig "devopsmaestro/pkg/nvimops/config"
	"devopsmaestro/pkg/nvimops/library"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/paths"
	"devopsmaestro/render"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	terminalplugin "devopsmaestro/pkg/terminalops/plugin"
)

// generateNvimConfig generates nvim configuration and copies to staging directory.
// It filters plugins based on the workspace's configured plugin list.
// Reads plugin data from the database (source of truth).
// Returns a PluginManifest for use by Dockerfile generator.
func generateNvimConfig(workspacePlugins []string, stagingDir, homeDir string, ds db.DataStore, app *models.App, workspace *models.Workspace, appName, workspaceName, language string) (*plugin.PluginManifest, error) {
	render.Progress("Generating Neovim configuration...")

	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create nvim config directory: %w", err)
	}

	// Load core config from ~/.nvp/core.yaml or use defaults
	pc := paths.New(homeDir)
	coreConfigPath := pc.NVPCoreConfig()

	var cfg *nvimconfig.CoreConfig
	var err error

	if _, statErr := os.Stat(coreConfigPath); statErr == nil {
		cfg, err = nvimconfig.ParseYAMLFile(coreConfigPath)
		if err != nil {
			slog.Warn("failed to parse core.yaml, using defaults", "error", err)
			cfg = nvimconfig.DefaultCoreConfig()
		}
	} else {
		slog.Debug("no core.yaml found, using defaults")
		cfg = nvimconfig.DefaultCoreConfig()
	}

	// Load plugins from database (source of truth)
	dbAdapter := store.NewDBStoreAdapter(ds)
	allPlugins, err := dbAdapter.List()
	if err != nil {
		slog.Warn("failed to list plugins from database", "error", err)
		allPlugins = []*plugin.Plugin{}
	}

	// Initialize plugin library for fallback
	pluginLibrary, err := library.NewLibrary()
	if err != nil {
		slog.Warn("failed to initialize plugin library", "error", err)
		pluginLibrary = nil
	}

	// Build a map of plugin names for quick lookup
	pluginMap := make(map[string]*plugin.Plugin)
	for _, p := range allPlugins {
		if p.Enabled {
			pluginMap[p.Name] = p
		}
	}

	// Filter plugins based on workspace configuration
	var enabledPlugins []*plugin.Plugin
	if len(workspacePlugins) > 0 {
		// Workspace has a specific plugin list - use only those
		for _, name := range workspacePlugins {
			if p, ok := pluginMap[name]; ok {
				enabledPlugins = append(enabledPlugins, p)
			} else {
				// Try loading from library as fallback
				if pluginLibrary != nil {
					if libPlugin, found := pluginLibrary.Get(name); found {
						enabledPlugins = append(enabledPlugins, libPlugin)
						slog.Debug("loaded workspace plugin from library", "plugin", name)
					} else {
						slog.Warn("workspace references unknown plugin", "plugin", name)
						render.Warning(fmt.Sprintf("Plugin '%s' not found in database or library (skipping)", name))
					}
				} else {
					slog.Warn("workspace references unknown plugin", "plugin", name)
					render.Warning(fmt.Sprintf("Plugin '%s' not found in database (skipping)", name))
				}
			}
		}
		slog.Debug("using workspace-specific plugins", "count", len(enabledPlugins), "requested", len(workspacePlugins))
	} else {
		// No plugins configured for workspace - check for default package
		defaultPkg, err := ds.GetDefault("nvim-package")
		if err == nil && defaultPkg != "" {
			// Try to resolve the default package
			packagePlugins, err := resolveDefaultPackagePlugins(defaultPkg, ds)
			if err != nil {
				slog.Warn("failed to resolve default package, falling back to all enabled plugins", "package", defaultPkg, "error", err)
				render.Warning(fmt.Sprintf("Failed to resolve default package '%s', using all enabled plugins", defaultPkg))
			} else {
				// Use plugins from the resolved package
				for _, pluginName := range packagePlugins {
					if p, ok := pluginMap[pluginName]; ok {
						enabledPlugins = append(enabledPlugins, p)
					} else {
						// Try loading from library as fallback
						if pluginLibrary != nil {
							if libPlugin, found := pluginLibrary.Get(pluginName); found {
								enabledPlugins = append(enabledPlugins, libPlugin)
								slog.Debug("loaded package plugin from library", "plugin", pluginName, "package", defaultPkg)
							} else {
								slog.Warn("default package references unknown plugin", "plugin", pluginName, "package", defaultPkg)
								render.Warning(fmt.Sprintf("Plugin '%s' from package '%s' not found in database or library (skipping)", pluginName, defaultPkg))
							}
						} else {
							slog.Warn("default package references unknown plugin", "plugin", pluginName, "package", defaultPkg)
							render.Warning(fmt.Sprintf("Plugin '%s' from package '%s' not found in database (skipping)", pluginName, defaultPkg))
						}
					}
				}
				slog.Debug("using plugins from default package", "package", defaultPkg, "count", len(enabledPlugins), "resolved_plugins", len(packagePlugins))
			}
		}

		// If no default package, try language-aware package selection
		if len(enabledPlugins) == 0 && language != "" && language != "unknown" {
			langPkg := nvimops.GetLanguagePackage(language)
			if langPkg != "" {
				langPlugins, err := resolveDefaultPackagePlugins(langPkg, ds)
				if err == nil {
					for _, pluginName := range langPlugins {
						if p, ok := pluginMap[pluginName]; ok {
							enabledPlugins = append(enabledPlugins, p)
						} else if pluginLibrary != nil {
							if libPlugin, found := pluginLibrary.Get(pluginName); found {
								enabledPlugins = append(enabledPlugins, libPlugin)
								slog.Debug("loaded language package plugin from library", "plugin", pluginName, "package", langPkg)
							} else {
								slog.Warn("language package references unknown plugin", "plugin", pluginName, "package", langPkg)
							}
						}
					}
					if len(enabledPlugins) > 0 {
						slog.Info("auto-selected language package", "package", langPkg, "language", language, "plugins", len(enabledPlugins))
						render.Info(fmt.Sprintf("Auto-selected '%s' package for %s workspace", langPkg, language))
					}
				} else {
					slog.Debug("failed to resolve language package", "package", langPkg, "error", err)
				}
			}
		}

		// If no default package or language package resolved, fall back to all enabled plugins
		if len(enabledPlugins) == 0 {
			for _, p := range allPlugins {
				if p.Enabled {
					enabledPlugins = append(enabledPlugins, p)
				}
			}
			slog.Debug("no package resolved, using all enabled plugins", "count", len(enabledPlugins))
		}

		// Final fallback: if still no plugins, load "core" package from embedded library
		// This ensures essential plugins (treesitter, mason/lspconfig, telescope) are always available
		if len(enabledPlugins) == 0 && pluginLibrary != nil {
			corePluginNames := []string{"treesitter", "telescope", "which-key", "lspconfig", "nvim-cmp", "gitsigns"}
			for _, pluginName := range corePluginNames {
				if libPlugin, found := pluginLibrary.Get(pluginName); found {
					enabledPlugins = append(enabledPlugins, libPlugin)
				}
			}
			slog.Info("no plugins configured, using embedded core package", "count", len(enabledPlugins))
			render.Info("No plugins configured - using default core package (treesitter, telescope, lsp, etc.)")
		}
	}

	slog.Debug("loaded nvp config", "plugins", len(enabledPlugins), "core_config", coreConfigPath)

	// Generate the full nvim config structure
	gen := nvimconfig.NewGenerator()
	if err := gen.WriteToDirectory(cfg, enabledPlugins, nvimConfigPath); err != nil {
		return nil, fmt.Errorf("failed to generate nvim config: %w", err)
	}

	// Create plugin manifest for Dockerfile generator
	manifest := plugin.ResolveManifest(enabledPlugins)

	// Generate theme from hierarchy (not global ~/.nvp/active-theme)
	themeStore := theme.NewFileStore(pc.NVPRoot())
	themeCtx := context.Background()
	resolvedTheme, themeErr := resolveWorkspaceTheme(themeCtx, ds, themeStore, workspace)
	if themeErr != nil {
		slog.Debug("no theme resolved for nvim config", "error", themeErr)
	}

	if resolvedTheme != nil {
		themeGen := theme.NewGenerator()
		generated, err := themeGen.Generate(resolvedTheme)
		if err == nil {
			ns := cfg.Namespace
			if ns == "" {
				ns = "workspace"
			}

			// Write theme files
			themeFiles := map[string]string{
				filepath.Join(nvimConfigPath, "lua", "theme", "palette.lua"):           generated.PaletteLua,
				filepath.Join(nvimConfigPath, "lua", "theme", "init.lua"):              generated.InitLua,
				filepath.Join(nvimConfigPath, "lua", ns, "plugins", "colorscheme.lua"): generated.PluginLua,
			}

			// Add standalone colorscheme implementation for standalone themes
			// This file contains the actual vim.api.nvim_set_hl() calls that apply the colors
			if generated.ColorschemeLua != "" {
				themeFiles[filepath.Join(nvimConfigPath, "lua", "theme", "colorscheme.lua")] = generated.ColorschemeLua
			}

			for path, content := range themeFiles {
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					slog.Warn("failed to create theme dir", "path", dir, "error", err)
					continue
				}
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					slog.Warn("failed to write theme file", "path", path, "error", err)
					continue
				}
			}
			slog.Debug("generated theme from hierarchy", "theme", resolvedTheme.Name, "workspace", workspace.Name)
		} else {
			slog.Warn("failed to generate theme", "error", err)
		}
	}

	render.Success(fmt.Sprintf("Neovim configuration generated (%d plugins)", len(enabledPlugins)))

	return manifest, nil
}

// appendPluginLoading appends terminal plugin loading configuration to the .zshrc file.
func appendPluginLoading(zshrcPath string, ds db.DataStore) error {
	// Get enabled terminal plugins from database
	plugins, err := ds.ListTerminalPlugins()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	// Filter to only enabled plugins and convert to plugin.Plugin
	var enabledPlugins []*terminalplugin.Plugin
	for _, dbPlugin := range plugins {
		if dbPlugin.Enabled {
			// Convert to plugin.Plugin using conversion pattern
			p := dbModelToPlugin(dbPlugin)
			enabledPlugins = append(enabledPlugins, p)
		}
	}

	if len(enabledPlugins) == 0 {
		return nil // No plugins to load - non-fatal
	}

	// Use pkg/terminalops/plugin/generator.go to generate loading script
	generator := terminalplugin.NewZshGenerator("$HOME/.local/share/zsh/plugins")
	pluginScript, err := generator.Generate(enabledPlugins)
	if err != nil {
		return fmt.Errorf("failed to generate plugin script: %w", err)
	}

	// Append to existing .zshrc file
	file, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open zshrc for appending: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("\n" + pluginScript); err != nil {
		return fmt.Errorf("failed to append plugin script: %w", err)
	}

	return nil
}

// dbModelToPlugin converts a models.TerminalPluginDB to terminalplugin.Plugin.
// This is adapted from pkg/terminalops/store/db_adapter.go
func dbModelToPlugin(db *models.TerminalPluginDB) *terminalplugin.Plugin {
	p := &terminalplugin.Plugin{
		Name:    db.Name,
		Repo:    db.Repo,
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}

	// Plugin manager
	p.Manager = terminalplugin.PluginManager(db.Manager)

	// Load command and config
	if db.LoadCommand.Valid {
		p.Config = db.LoadCommand.String

		// If it looks like an oh-my-zsh plugin, extract the plugin name
		if strings.HasPrefix(db.LoadCommand.String, "plugins+=") {
			p.OhMyZshPlugin = strings.TrimPrefix(db.LoadCommand.String, "plugins+=")
		}
	}

	// Source file
	if db.SourceFile.Valid {
		p.SourceFiles = []string{db.SourceFile.String}
	}

	// Parse dependencies JSON
	if db.Dependencies != "" && db.Dependencies != "[]" {
		var deps []string
		if err := json.Unmarshal([]byte(db.Dependencies), &deps); err == nil {
			p.Dependencies = deps
		}
	}

	// Parse env vars JSON
	if db.EnvVars != "" && db.EnvVars != "{}" {
		var envVars map[string]string
		if err := json.Unmarshal([]byte(db.EnvVars), &envVars); err == nil {
			p.Env = envVars
		}
	}

	// Parse labels JSON and extract metadata
	if db.Labels != "" && db.Labels != "{}" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(db.Labels), &labels); err == nil {
			// Extract tags
			var tags []string
			for key, value := range labels {
				if strings.HasPrefix(key, "tag:") && value == "true" {
					tags = append(tags, strings.TrimPrefix(key, "tag:"))
				}
			}
			p.Tags = tags

			// Extract other metadata
			if loadMode, ok := labels["load_mode"]; ok {
				p.LoadMode = terminalplugin.LoadMode(loadMode)
			}
			if branch, ok := labels["branch"]; ok {
				p.Branch = branch
			}
			if tag, ok := labels["tag"]; ok {
				p.Tag = tag
			}
			if priorityStr, ok := labels["priority"]; ok {
				var priority int
				fmt.Sscanf(priorityStr, "%d", &priority)
				p.Priority = priority
			}
		}
	}

	// Timestamps
	if !db.CreatedAt.IsZero() {
		p.CreatedAt = &db.CreatedAt
	}
	if !db.UpdatedAt.IsZero() {
		p.UpdatedAt = &db.UpdatedAt
	}

	return p
}
