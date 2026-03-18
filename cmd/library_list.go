package cmd

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroSDK/render"

	terminalpkglib "devopsmaestro/pkg/terminalops/package/library"
	terminalpluginlib "devopsmaestro/pkg/terminalops/plugin/library"
	terminalpromptlib "devopsmaestro/pkg/terminalops/prompt/library"
	nvimpluginlib "github.com/rmkohlman/MaestroNvim/nvimops/library"
	nvimpkglib "github.com/rmkohlman/MaestroNvim/nvimops/package/library"
	nvimthemelib "github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

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
