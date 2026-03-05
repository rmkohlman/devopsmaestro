package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/models"
	nvimpluginlib "devopsmaestro/pkg/nvimops/library"
	nvimpkglib "devopsmaestro/pkg/nvimops/package/library"
	nvimthemelib "devopsmaestro/pkg/nvimops/theme/library"
	terminalpkglib "devopsmaestro/pkg/terminalops/package/library"
	terminalpluginlib "devopsmaestro/pkg/terminalops/plugin/library"
	terminalpromptlib "devopsmaestro/pkg/terminalops/prompt/library"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// libraryCmd is the root 'library' command
var libraryCmd = &cobra.Command{
	Use:     "library",
	Aliases: []string{"lib"},
	Short:   "Browse plugin and theme libraries",
	Long: `Browse plugin and theme libraries.

View available plugins, themes, prompts, and packages from the embedded libraries.

Examples:
  dvm library list plugins               # List nvim plugins
  dvm library list themes                # List nvim themes
  dvm library list terminal prompts      # List terminal prompts
  dvm lib ls np                          # Short form: nvim plugins
  dvm library show plugin telescope      # Show plugin details
  dvm library show theme coolnight-ocean # Show theme details`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// libraryListCmd is the 'list' subcommand
var libraryListCmd = &cobra.Command{
	Use:     "list [resource-type]",
	Aliases: []string{"ls"},
	Short:   "List library resources",
	Long: `List library resources (plugins, themes, prompts, packages).

Resource types:
  plugins               - List nvim plugins (alias: np)
  themes                - List nvim themes (alias: nt)
  nvim packages         - List nvim plugin bundles
  terminal prompts      - List terminal prompts (alias: tp)
  terminal plugins      - List shell plugins (alias: tpl)
  terminal packages     - List terminal bundles

Examples:
  dvm library list plugins
  dvm library list themes
  dvm library list terminal prompts
  dvm lib ls np                    # Short form
  dvm library list plugins -o yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLibraryList,
}

// libraryShowCmd is the 'show' subcommand
var libraryShowCmd = &cobra.Command{
	Use:   "show [resource-type] [name]",
	Short: "Show library resource details",
	Long: `Show details of a specific library resource.

Resource types:
  plugin [name]               - Show nvim plugin details
  theme [name]                - Show nvim theme details
  nvim-package [name]         - Show nvim package details
  terminal prompt [name]      - Show terminal prompt details
  terminal plugin [name]      - Show terminal plugin details
  terminal-package [name]     - Show terminal package details

Examples:
  dvm library show plugin telescope
  dvm library show theme coolnight-ocean
  dvm library show nvim-package core
  dvm library show terminal prompt starship-default
  dvm library show terminal plugin zsh-autosuggestions
  dvm library show terminal-package core`,
	Args: cobra.MinimumNArgs(2),
	RunE: runLibraryShow,
}

// libraryImportCmd is the 'import' subcommand
var libraryImportCmd = &cobra.Command{
	Use:   "import [resource-type...]",
	Short: "Import library resources into the database",
	Long: `Import library resources into the database.

Imports embedded library resources into the local database for use with workspaces.

Resource types:
  nvim-plugins          - Import nvim plugins
  nvim-themes           - Import nvim themes
  nvim-packages         - Import nvim packages
  terminal-prompts      - Import terminal prompts
  terminal-plugins      - Import terminal plugins
  terminal-packages     - Import terminal packages

Flags:
  --all                 - Import all resource types

Examples:
  dvm library import nvim-plugins          # Import just nvim plugins
  dvm library import nvim-themes           # Import just nvim themes
  dvm library import --all                 # Import all resource types`,
	RunE: runLibraryImport,
}

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
func importNvimPlugins(ds db.DataStore) error {
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
func importNvimThemes(ds db.DataStore) error {
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
func importNvimPackages(ds db.DataStore) error {
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
func importTerminalPrompts(ds db.DataStore) error {
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
func importTerminalPlugins(ds db.DataStore) error {
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
func importTerminalPackages(ds db.DataStore) error {
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

// runLibraryList handles the 'library list' command
func runLibraryList(cmd *cobra.Command, args []string) error {
	outputFormat, _ := cmd.Flags().GetString("output")

	// Normalize resource type and handle aliases
	resourceType := normalizeResourceType(args)

	switch resourceType {
	case "plugins", "np":
		return listNvimPlugins(cmd, outputFormat)
	case "themes", "nt":
		return listNvimThemes(cmd, outputFormat)
	case "nvim-packages":
		return listNvimPackages(cmd, outputFormat)
	case "terminal-prompts", "tp":
		return listTerminalPrompts(cmd, outputFormat)
	case "terminal-plugins", "tpl":
		return listTerminalPlugins(cmd, outputFormat)
	case "terminal-packages":
		return listTerminalPackages(cmd, outputFormat)
	default:
		return fmt.Errorf("unknown resource type: %s (valid: plugins, themes, nvim packages, terminal prompts, terminal plugins, terminal packages)", strings.Join(args, " "))
	}
}

// runLibraryShow handles the 'library show' command
func runLibraryShow(cmd *cobra.Command, args []string) error {
	outputFormat, _ := cmd.Flags().GetString("output")

	// Normalize resource type and handle aliases
	resourceType := normalizeResourceType(args[:len(args)-1])
	name := args[len(args)-1]

	switch resourceType {
	case "plugin":
		return showNvimPlugin(cmd, name, outputFormat)
	case "theme":
		return showNvimTheme(cmd, name, outputFormat)
	case "nvim-package":
		return showNvimPackage(cmd, name, outputFormat)
	case "terminal-prompt":
		return showTerminalPrompt(cmd, name, outputFormat)
	case "terminal-plugin":
		return showTerminalPlugin(cmd, name, outputFormat)
	case "terminal-package":
		return showTerminalPackage(cmd, name, outputFormat)
	default:
		return fmt.Errorf("unknown resource type: %s (valid: plugin, theme, nvim-package, terminal prompt, terminal plugin, terminal-package)", strings.Join(args[:len(args)-1], " "))
	}
}

// normalizeResourceType converts command args to a normalized resource type
func normalizeResourceType(args []string) string {
	if len(args) == 0 {
		return ""
	}

	// Pre-pass: if a single arg contains a space (e.g., Cobra passed "nvim packages"
	// as one quoted arg), split it into separate words so it flows through multi-word logic.
	if len(args) == 1 && strings.Contains(args[0], " ") {
		args = strings.Fields(args[0])
	}

	// Single word types
	switch args[0] {
	case "plugins", "np":
		return "plugins"
	case "themes", "nt":
		return "themes"
	case "plugin":
		return "plugin"
	case "theme":
		return "theme"
	case "tp":
		return "terminal-prompts"
	case "tpl":
		return "terminal-plugins"
	case "nvim-package":
		return "nvim-package"
	case "nvim-packages":
		return "nvim-packages"
	case "terminal-package":
		return "terminal-package"
	case "terminal-packages":
		return "terminal-packages"
	}

	// Multi-word types
	if len(args) >= 2 {
		combined := strings.Join(args, " ")
		switch combined {
		case "nvim packages":
			return "nvim-packages"
		case "nvim package":
			return "nvim-package"
		case "terminal prompts":
			return "terminal-prompts"
		case "terminal plugins":
			return "terminal-plugins"
		case "terminal packages":
			return "terminal-packages"
		case "terminal prompt":
			return "terminal-prompt"
		case "terminal plugin":
			return "terminal-plugin"
		case "terminal package":
			return "terminal-package"
		}
	}

	return strings.Join(args, "-")
}

// listNvimPlugins lists nvim plugins from the library
func listNvimPlugins(cmd *cobra.Command, outputFormat string) error {
	lib, err := nvimpluginlib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load plugin library: %w", err)
	}

	plugins := lib.List()

	// Convert to output format
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, plugins, render.Options{})
	}

	// Table format
	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(plugins))
	for _, p := range plugins {
		rows = append(rows, []string{p.Name, p.Description})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{})
}

// listNvimThemes lists nvim themes from the library
func listNvimThemes(cmd *cobra.Command, outputFormat string) error {
	themes, err := nvimthemelib.List()
	if err != nil {
		return fmt.Errorf("failed to load theme library: %w", err)
	}

	// Convert to output format
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, themes, render.Options{})
	}

	// Table format
	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(themes))
	for _, t := range themes {
		rows = append(rows, []string{t.Name, t.Description})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{})
}

// listNvimPackages lists nvim packages from the library
func listNvimPackages(cmd *cobra.Command, outputFormat string) error {
	lib, err := nvimpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	packages := lib.List()

	// Convert to output format
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, packages, render.Options{})
	}

	// Table format
	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(packages))
	for _, p := range packages {
		rows = append(rows, []string{p.Name, p.Description})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{})
}

// listTerminalPrompts lists terminal prompts from the library
func listTerminalPrompts(cmd *cobra.Command, outputFormat string) error {
	lib, err := terminalpromptlib.NewPromptLibrary()
	if err != nil {
		return fmt.Errorf("failed to load prompt library: %w", err)
	}

	prompts := lib.List()

	// Convert to output format
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, prompts, render.Options{})
	}

	// Table format
	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(prompts))
	for _, p := range prompts {
		rows = append(rows, []string{p.Name, p.Description})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{})
}

// listTerminalPlugins lists terminal plugins from the library
func listTerminalPlugins(cmd *cobra.Command, outputFormat string) error {
	lib, err := terminalpluginlib.NewPluginLibrary()
	if err != nil {
		return fmt.Errorf("failed to load plugin library: %w", err)
	}

	plugins := lib.List()

	// Convert to output format
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, plugins, render.Options{})
	}

	// Table format
	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(plugins))
	for _, p := range plugins {
		rows = append(rows, []string{p.Name, p.Description})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{})
}

// listTerminalPackages lists terminal packages from the library
func listTerminalPackages(cmd *cobra.Command, outputFormat string) error {
	lib, err := terminalpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	packages := lib.List()

	// Convert to output format
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, packages, render.Options{})
	}

	// Table format
	headers := []string{"NAME", "DESCRIPTION"}
	rows := make([][]string, 0, len(packages))
	for _, p := range packages {
		rows = append(rows, []string{p.Name, p.Description})
	}

	return render.OutputWith(outputFormat, render.TableData{
		Headers: headers,
		Rows:    rows,
	}, render.Options{})
}

// showNvimPlugin shows details of a specific nvim plugin
func showNvimPlugin(cmd *cobra.Command, name string, outputFormat string) error {
	lib, err := nvimpluginlib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load plugin library: %w", err)
	}

	plugin, ok := lib.Get(name)
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// For structured formats (yaml/json), pass the raw struct
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, plugin, render.Options{})
	}

	// For table format, convert to key-value display
	kvData := map[string]string{
		"Name":        plugin.Name,
		"Description": plugin.Description,
		"Repository":  plugin.Repo,
		"Branch":      plugin.Branch,
		"Version":     plugin.Version,
		"Lazy":        fmt.Sprintf("%t", plugin.Lazy),
	}

	return render.OutputWith(outputFormat, kvData, render.Options{})
}

// showNvimTheme shows details of a specific nvim theme
func showNvimTheme(cmd *cobra.Command, name string, outputFormat string) error {
	theme, err := nvimthemelib.Get(name)
	if err != nil {
		return fmt.Errorf("theme not found: %s", name)
	}

	// For structured formats (yaml/json), pass the raw struct
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, theme, render.Options{})
	}

	// For table format, convert to key-value display
	kvData := map[string]string{
		"Name":        theme.Name,
		"Description": theme.Description,
		"Author":      theme.Author,
		"Category":    theme.Category,
		"Style":       theme.Style,
		"Repository":  theme.Plugin.Repo,
	}

	return render.OutputWith(outputFormat, kvData, render.Options{})
}

// showTerminalPrompt shows details of a specific terminal prompt
func showTerminalPrompt(cmd *cobra.Command, name string, outputFormat string) error {
	lib, err := terminalpromptlib.NewPromptLibrary()
	if err != nil {
		return fmt.Errorf("failed to load prompt library: %w", err)
	}

	prompt, err := lib.Get(name)
	if err != nil {
		return fmt.Errorf("prompt not found: %s", name)
	}

	// For structured formats (yaml/json), pass the raw struct
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, prompt, render.Options{})
	}

	// For table format, convert to key-value display
	kvData := map[string]string{
		"Name":        prompt.Name,
		"Description": prompt.Description,
		"Type":        string(prompt.Type),
		"Palette":     prompt.Palette,
	}

	return render.OutputWith(outputFormat, kvData, render.Options{})
}

// showTerminalPlugin shows details of a specific terminal plugin
func showTerminalPlugin(cmd *cobra.Command, name string, outputFormat string) error {
	lib, err := terminalpluginlib.NewPluginLibrary()
	if err != nil {
		return fmt.Errorf("failed to load plugin library: %w", err)
	}

	plugin, err := lib.Get(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// For structured formats (yaml/json), pass the raw struct
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputWith(outputFormat, plugin, render.Options{})
	}

	// For table format, convert to key-value display
	kvData := map[string]string{
		"Name":        plugin.Name,
		"Description": plugin.Description,
		"Repository":  plugin.Repo,
		"Manager":     string(plugin.Manager),
	}

	return render.OutputWith(outputFormat, kvData, render.Options{})
}

// showNvimPackage shows details of a specific nvim package
func showNvimPackage(cmd *cobra.Command, name string, outputFormat string) error {
	lib, err := nvimpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	p, ok := lib.Get(name)
	if !ok {
		return fmt.Errorf("package not found: %s", name)
	}

	w := cmd.OutOrStdout()

	// For structured formats (yaml/json), pass the raw struct
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputTo(w, outputFormat, p, render.Options{})
	}

	// For table format, convert to key-value display
	kvData := map[string]string{
		"Name":        p.Name,
		"Description": p.Description,
		"Category":    p.Category,
		"Extends":     p.Extends,
		"Plugins":     strings.Join(p.Plugins, ", "),
		"Tags":        strings.Join(p.Tags, ", "),
		"Enabled":     fmt.Sprintf("%t", p.Enabled),
	}

	return render.OutputTo(w, outputFormat, kvData, render.Options{})
}

// showTerminalPackage shows details of a specific terminal package
func showTerminalPackage(cmd *cobra.Command, name string, outputFormat string) error {
	lib, err := terminalpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	p, ok := lib.Get(name)
	if !ok {
		return fmt.Errorf("package not found: %s", name)
	}

	w := cmd.OutOrStdout()

	// For structured formats (yaml/json), pass the raw struct
	if outputFormat == "yaml" || outputFormat == "json" {
		return render.OutputTo(w, outputFormat, p, render.Options{})
	}

	// For table format, convert to key-value display
	kvData := map[string]string{
		"Name":        p.Name,
		"Description": p.Description,
		"Category":    p.Category,
		"Extends":     p.Extends,
		"Plugins":     strings.Join(p.Plugins, ", "),
		"Prompts":     strings.Join(p.Prompts, ", "),
		"Tags":        strings.Join(p.Tags, ", "),
		"Enabled":     fmt.Sprintf("%t", p.Enabled),
	}

	return render.OutputTo(w, outputFormat, kvData, render.Options{})
}

func init() {
	// Add output format flag to list and show commands
	libraryListCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")
	libraryShowCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	// Add flags to import command
	libraryImportCmd.Flags().Bool("all", false, "Import all resource types")
	libraryImportCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	// Add subcommands
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryShowCmd)
	libraryCmd.AddCommand(libraryImportCmd)

	// Add library command to root
	rootCmd.AddCommand(libraryCmd)
}
