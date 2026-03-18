package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	terminalpkglib "github.com/rmkohlman/MaestroTerminal/terminalops/package/library"
	terminalpluginlib "github.com/rmkohlman/MaestroTerminal/terminalops/plugin/library"
	terminalpromptlib "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/library"
	nvimpluginlib "github.com/rmkohlman/MaestroNvim/nvimops/library"
	nvimpkglib "github.com/rmkohlman/MaestroNvim/nvimops/package/library"
	nvimthemelib "github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

// runLibraryImport handles the 'library import' command
func runLibraryImport(cmd *cobra.Command, args []string) error {
	allFlag, _ := cmd.Flags().GetBool("all")

	if !allFlag && len(args) == 0 {
		return fmt.Errorf("specify at least one resource type or use --all")
	}

	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("failed to get data store: %w", err)
	}

	// Determine which types to import
	var types []string
	if allFlag {
		types = []string{
			"nvim-plugins",
			"nvim-themes",
			"nvim-packages",
			"terminal-prompts",
			"terminal-plugins",
			"terminal-packages",
		}
	} else {
		types = args
	}

	for _, t := range types {
		switch t {
		case "nvim-plugins":
			if err := importNvimPlugins(ds); err != nil {
				return fmt.Errorf("failed to import nvim-plugins: %w", err)
			}
		case "nvim-themes":
			if err := importNvimThemes(ds); err != nil {
				return fmt.Errorf("failed to import nvim-themes: %w", err)
			}
		case "nvim-packages":
			if err := importNvimPackages(ds); err != nil {
				return fmt.Errorf("failed to import nvim-packages: %w", err)
			}
		case "terminal-prompts":
			if err := importTerminalPrompts(ds); err != nil {
				return fmt.Errorf("failed to import terminal-prompts: %w", err)
			}
		case "terminal-plugins":
			if err := importTerminalPlugins(ds); err != nil {
				return fmt.Errorf("failed to import terminal-plugins: %w", err)
			}
		case "terminal-packages":
			if err := importTerminalPackages(ds); err != nil {
				return fmt.Errorf("failed to import terminal-packages: %w", err)
			}
		default:
			return fmt.Errorf("unknown resource type: %s (valid: nvim-plugins, nvim-themes, nvim-packages, terminal-prompts, terminal-plugins, terminal-packages)", t)
		}
	}

	return nil
}

// importNvimPlugins loads plugins from the library and upserts them into the DB.
func importNvimPlugins(ds db.PluginStore) error {
	lib, err := nvimpluginlib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load plugin library: %w", err)
	}

	for _, p := range lib.List() {
		pluginDB := &models.NvimPluginDB{
			Name:    p.Name,
			Repo:    p.Repo,
			Lazy:    p.Lazy,
			Enabled: p.Enabled,
		}

		if p.Description != "" {
			pluginDB.Description = sql.NullString{String: p.Description, Valid: true}
		}
		if p.Branch != "" {
			pluginDB.Branch = sql.NullString{String: p.Branch, Valid: true}
		}
		if p.Version != "" {
			pluginDB.Version = sql.NullString{String: p.Version, Valid: true}
		}
		if p.Category != "" {
			pluginDB.Category = sql.NullString{String: p.Category, Valid: true}
		}
		if p.Build != "" {
			pluginDB.Build = sql.NullString{String: p.Build, Valid: true}
		}
		if p.Config != "" {
			pluginDB.Config = sql.NullString{String: p.Config, Valid: true}
		}
		if p.Init != "" {
			pluginDB.Init = sql.NullString{String: p.Init, Valid: true}
		}
		if p.Priority != 0 {
			pluginDB.Priority = sql.NullInt64{Int64: int64(p.Priority), Valid: true}
		}
		if len(p.Event) > 0 {
			if eventJSON, err := json.Marshal(p.Event); err == nil {
				pluginDB.Event = sql.NullString{String: string(eventJSON), Valid: true}
			}
		}
		if len(p.Ft) > 0 {
			if ftJSON, err := json.Marshal(p.Ft); err == nil {
				pluginDB.Ft = sql.NullString{String: string(ftJSON), Valid: true}
			}
		}
		if len(p.Cmd) > 0 {
			if cmdJSON, err := json.Marshal(p.Cmd); err == nil {
				pluginDB.Cmd = sql.NullString{String: string(cmdJSON), Valid: true}
			}
		}
		if len(p.Keys) > 0 {
			if keysJSON, err := json.Marshal(p.Keys); err == nil {
				pluginDB.Keys = sql.NullString{String: string(keysJSON), Valid: true}
			}
		}
		if len(p.Dependencies) > 0 {
			if depsJSON, err := json.Marshal(p.Dependencies); err == nil {
				pluginDB.Dependencies = sql.NullString{String: string(depsJSON), Valid: true}
			}
		}
		if len(p.Keymaps) > 0 {
			if keymapsJSON, err := json.Marshal(p.Keymaps); err == nil {
				pluginDB.Keymaps = sql.NullString{String: string(keymapsJSON), Valid: true}
			}
		}
		if len(p.Tags) > 0 {
			if tagsJSON, err := json.Marshal(p.Tags); err == nil {
				pluginDB.Tags = sql.NullString{String: string(tagsJSON), Valid: true}
			}
		}
		if p.Opts != nil {
			if optsJSON, err := json.Marshal(p.Opts); err == nil {
				pluginDB.Opts = sql.NullString{String: string(optsJSON), Valid: true}
			}
		}

		if err := ds.UpsertPlugin(pluginDB); err != nil {
			return fmt.Errorf("failed to upsert plugin %s: %w", p.Name, err)
		}
	}

	return nil
}

// importNvimThemes loads themes from the library and creates them in the DB.
func importNvimThemes(ds db.ThemeStore) error {
	themeInfos, err := nvimthemelib.List()
	if err != nil {
		return fmt.Errorf("failed to load theme library: %w", err)
	}

	for _, info := range themeInfos {
		theme, err := nvimthemelib.Get(info.Name)
		if err != nil {
			return fmt.Errorf("failed to get theme %s: %w", info.Name, err)
		}

		themeDB := &models.NvimThemeDB{
			Name:       theme.Name,
			PluginRepo: theme.Plugin.Repo,
		}

		if theme.Description != "" {
			themeDB.Description = sql.NullString{String: theme.Description, Valid: true}
		}
		if theme.Author != "" {
			themeDB.Author = sql.NullString{String: theme.Author, Valid: true}
		}
		if theme.Category != "" {
			themeDB.Category = sql.NullString{String: theme.Category, Valid: true}
		}
		if theme.Style != "" {
			themeDB.Style = sql.NullString{String: theme.Style, Valid: true}
		}
		if theme.Plugin.Branch != "" {
			themeDB.PluginBranch = sql.NullString{String: theme.Plugin.Branch, Valid: true}
		}
		if theme.Plugin.Tag != "" {
			themeDB.PluginTag = sql.NullString{String: theme.Plugin.Tag, Valid: true}
		}
		themeDB.Transparent = theme.Transparent

		if len(theme.Colors) > 0 {
			if colorsJSON, err := json.Marshal(theme.Colors); err == nil {
				themeDB.Colors = sql.NullString{String: string(colorsJSON), Valid: true}
			}
		}
		if len(theme.Options) > 0 {
			if optionsJSON, err := json.Marshal(theme.Options); err == nil {
				themeDB.Options = sql.NullString{String: string(optionsJSON), Valid: true}
			}
		}

		if err := ds.CreateTheme(themeDB); err != nil {
			return fmt.Errorf("failed to create theme %s: %w", theme.Name, err)
		}
	}

	return nil
}

// importNvimPackages loads packages from the library and upserts them into the DB.
func importNvimPackages(ds db.NvimPackageStore) error {
	lib, err := nvimpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	for _, p := range lib.List() {
		pkgDB := &models.NvimPackageDB{
			Name: p.Name,
		}

		if p.Description != "" {
			pkgDB.Description = sql.NullString{String: p.Description, Valid: true}
		}
		if p.Category != "" {
			pkgDB.Category = sql.NullString{String: p.Category, Valid: true}
		}
		if p.Extends != "" {
			pkgDB.Extends = sql.NullString{String: p.Extends, Valid: true}
		}
		if err := pkgDB.SetPlugins(p.Plugins); err != nil {
			return fmt.Errorf("failed to set plugins for package %s: %w", p.Name, err)
		}

		if err := ds.UpsertPackage(pkgDB); err != nil {
			return fmt.Errorf("failed to upsert package %s: %w", p.Name, err)
		}
	}

	return nil
}

// importTerminalPrompts loads prompts from the library and upserts them into the DB.
func importTerminalPrompts(ds db.TerminalPromptStore) error {
	lib, err := terminalpromptlib.NewPromptLibrary()
	if err != nil {
		return fmt.Errorf("failed to load prompt library: %w", err)
	}

	for _, p := range lib.List() {
		promptDB := &models.TerminalPromptDB{
			Name:       p.Name,
			Type:       string(p.Type),
			AddNewline: p.AddNewline,
			Enabled:    p.Enabled,
		}

		if p.Description != "" {
			promptDB.Description = sql.NullString{String: p.Description, Valid: true}
		}
		if p.Palette != "" {
			promptDB.Palette = sql.NullString{String: p.Palette, Valid: true}
		}
		if p.Format != "" {
			promptDB.Format = sql.NullString{String: p.Format, Valid: true}
		}
		if p.PaletteRef != "" {
			promptDB.PaletteRef = sql.NullString{String: p.PaletteRef, Valid: true}
		}
		if p.RawConfig != "" {
			promptDB.RawConfig = sql.NullString{String: p.RawConfig, Valid: true}
		}
		if p.Category != "" {
			promptDB.Category = sql.NullString{String: p.Category, Valid: true}
		}
		if len(p.Modules) > 0 {
			if modulesJSON, err := json.Marshal(p.Modules); err == nil {
				promptDB.Modules = sql.NullString{String: string(modulesJSON), Valid: true}
			}
		}
		if p.Character != nil {
			if charJSON, err := json.Marshal(p.Character); err == nil {
				promptDB.Character = sql.NullString{String: string(charJSON), Valid: true}
			}
		}
		if len(p.Colors) > 0 {
			if colorsJSON, err := json.Marshal(p.Colors); err == nil {
				promptDB.Colors = sql.NullString{String: string(colorsJSON), Valid: true}
			}
		}
		if len(p.Tags) > 0 {
			if tagsJSON, err := json.Marshal(p.Tags); err == nil {
				promptDB.Tags = sql.NullString{String: string(tagsJSON), Valid: true}
			}
		}

		if err := ds.UpsertTerminalPrompt(promptDB); err != nil {
			return fmt.Errorf("failed to upsert terminal prompt %s: %w", p.Name, err)
		}
	}

	return nil
}

// importTerminalPlugins loads terminal plugins from the library and upserts them into the DB.
func importTerminalPlugins(ds db.TerminalPluginStore) error {
	lib, err := terminalpluginlib.NewPluginLibrary()
	if err != nil {
		return fmt.Errorf("failed to load terminal plugin library: %w", err)
	}

	for _, p := range lib.List() {
		pluginDB := &models.TerminalPluginDB{
			Name:    p.Name,
			Repo:    p.Repo,
			Shell:   "zsh",
			Manager: string(p.Manager),
			Enabled: p.Enabled,
		}

		if pluginDB.Manager == "" {
			pluginDB.Manager = "manual"
		}

		if p.Description != "" {
			pluginDB.Description = sql.NullString{String: p.Description, Valid: true}
		}
		if p.Category != "" {
			pluginDB.Category = sql.NullString{String: p.Category, Valid: true}
		}

		// Dependencies as JSON array
		if len(p.Dependencies) > 0 {
			if depsJSON, err := json.Marshal(p.Dependencies); err == nil {
				pluginDB.Dependencies = string(depsJSON)
			} else {
				pluginDB.Dependencies = "[]"
			}
		} else {
			pluginDB.Dependencies = "[]"
		}

		// Env as JSON object
		if len(p.Env) > 0 {
			if envJSON, err := json.Marshal(p.Env); err == nil {
				pluginDB.EnvVars = string(envJSON)
			} else {
				pluginDB.EnvVars = "{}"
			}
		} else {
			pluginDB.EnvVars = "{}"
		}

		// Labels as JSON object (empty by default)
		pluginDB.Labels = "{}"

		if err := ds.UpsertTerminalPlugin(pluginDB); err != nil {
			return fmt.Errorf("failed to upsert terminal plugin %s: %w", p.Name, err)
		}
	}

	return nil
}

// importTerminalPackages loads terminal packages from the library and upserts them into the DB.
func importTerminalPackages(ds db.TerminalPackageStore) error {
	lib, err := terminalpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load terminal package library: %w", err)
	}

	for _, p := range lib.List() {
		pkgDB := &models.TerminalPackageDB{
			Name: p.Name,
		}

		if p.Description != "" {
			pkgDB.Description = sql.NullString{String: p.Description, Valid: true}
		}
		if p.Category != "" {
			pkgDB.Category = sql.NullString{String: p.Category, Valid: true}
		}
		if p.Extends != "" {
			pkgDB.Extends = sql.NullString{String: p.Extends, Valid: true}
		}
		if err := pkgDB.SetPlugins(p.Plugins); err != nil {
			return fmt.Errorf("failed to set plugins for terminal package %s: %w", p.Name, err)
		}
		if err := pkgDB.SetPrompts(p.Prompts); err != nil {
			return fmt.Errorf("failed to set prompts for terminal package %s: %w", p.Name, err)
		}
		if err := pkgDB.SetProfiles(p.Profiles); err != nil {
			return fmt.Errorf("failed to set profiles for terminal package %s: %w", p.Name, err)
		}

		if p.WezTerm != nil {
			weztermMap := map[string]interface{}{
				"fontSize":    p.WezTerm.FontSize,
				"colorScheme": p.WezTerm.ColorScheme,
				"fontFamily":  p.WezTerm.FontFamily,
			}
			if err := pkgDB.SetWezTerm(weztermMap); err != nil {
				return fmt.Errorf("failed to set wezterm for terminal package %s: %w", p.Name, err)
			}
		}

		if err := ds.UpsertTerminalPackage(pkgDB); err != nil {
			return fmt.Errorf("failed to upsert terminal package %s: %w", p.Name, err)
		}
	}

	return nil
}
