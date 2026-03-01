package cmd

import (
	"fmt"
	"strings"

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
  terminal prompt [name]      - Show terminal prompt details
  terminal plugin [name]      - Show terminal plugin details

Examples:
  dvm library show plugin telescope
  dvm library show theme coolnight-ocean
  dvm library show terminal prompt starship-default
  dvm library show terminal plugin zsh-autosuggestions`,
	Args: cobra.MinimumNArgs(2),
	RunE: runLibraryShow,
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
	case "terminal-prompt":
		return showTerminalPrompt(cmd, name, outputFormat)
	case "terminal-plugin":
		return showTerminalPlugin(cmd, name, outputFormat)
	default:
		return fmt.Errorf("unknown resource type: %s (valid: plugin, theme, terminal prompt, terminal plugin)", strings.Join(args[:len(args)-1], " "))
	}
}

// normalizeResourceType converts command args to a normalized resource type
func normalizeResourceType(args []string) string {
	if len(args) == 0 {
		return ""
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
	}

	// Multi-word types
	if len(args) >= 2 {
		combined := strings.Join(args, " ")
		switch combined {
		case "nvim packages":
			return "nvim-packages"
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

	return render.OutputWith(outputFormat, plugin, render.Options{})
}

// showNvimTheme shows details of a specific nvim theme
func showNvimTheme(cmd *cobra.Command, name string, outputFormat string) error {
	theme, err := nvimthemelib.Get(name)
	if err != nil {
		return fmt.Errorf("theme not found: %s", name)
	}

	return render.OutputWith(outputFormat, theme, render.Options{})
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

	return render.OutputWith(outputFormat, prompt, render.Options{})
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

	return render.OutputWith(outputFormat, plugin, render.Options{})
}

func init() {
	// Add output format flag to list and show commands
	libraryListCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")
	libraryShowCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	// Add subcommands
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryShowCmd)

	// Add library command to root
	rootCmd.AddCommand(libraryCmd)
}
