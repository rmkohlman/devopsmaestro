package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// Flags for workspace/app filtering
var (
	nvimWorkspaceFlag string
	nvimAppFlag       string
)

// nvimGetCmd is the 'nvim' subcommand under 'get' for kubectl-style namespacing
// Usage: dvm get nvim plugins, dvm get nvim plugin <name>
var nvimGetCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Get nvim resources (plugins, themes)",
	Long: `Get nvim-related resources in kubectl-style namespaced format.

Use -w to filter by workspace (shows plugins configured for that workspace).
Without -w, shows all plugins in the global library.

Examples:
  dvm get nvim plugins              # List all global plugins
  dvm get nvim plugins -w dev       # List plugins for workspace 'dev'
  dvm get nvim plugin telescope     # Get specific plugin details
  dvm get nvim plugins -o yaml      # Output as YAML
`,
}

// nvimGetPluginsCmd lists nvim plugins (global or workspace-filtered)
// Usage: dvm get nvim plugins [-w workspace] [-a app]
var nvimGetPluginsCmd = &cobra.Command{
	Use:     "plugins",
	Aliases: []string{"np"},
	Short:   "List nvim plugins (global or per-workspace)",
	Long: `List nvim plugins from the global library or filtered by workspace.

Without flags: Lists all plugins in the global library (~/.nvp/plugins/).
With -w flag:  Lists only plugins configured for the specified workspace.

Examples:
  dvm get nvim plugins                  # List all global plugins
  dvm get nvim plugins -w dev           # List plugins for workspace 'dev'
  dvm get nvim plugins -a myapp -w dev  # Explicit app and workspace
  dvm get nvim plugins -o yaml          # Output as YAML`,
	RunE: runGetNvimPlugins,
}

// nvimGetPluginCmd gets a specific nvim plugin (namespaced version)
// Usage: dvm get nvim plugin <name>
var nvimGetPluginCmd = &cobra.Command{
	Use:   "plugin [name]",
	Short: "Get a specific nvim plugin",
	Long: `Get a specific nvim plugin by name.

Examples:
  dvm get nvim plugin telescope
  dvm get nvim plugin telescope -o yaml
  dvm get nvim plugin lspconfig -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Reuse existing getPlugin function from get.go
		return getPlugin(cmd, args[0])
	},
}

// nvimGetThemesCmd lists all nvim themes (namespaced version)
// Usage: dvm get nvim themes
var nvimGetThemesCmd = &cobra.Command{
	Use:     "themes",
	Aliases: []string{"nt"},
	Short:   "List all nvim themes (user + library)",
	Long: `List all nvim themes from user store and embedded library.

Shows combined list of user themes and 34+ library themes.
User themes override library themes with the same name.

Examples:
  dvm get nvim themes
  dvm get nvim themes -o yaml
  dvm get nvim themes -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getThemes(cmd)
	},
}

// nvimGetThemeCmd gets a specific nvim theme (namespaced version)
// Usage: dvm get nvim theme <name>
var nvimGetThemeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "Get a specific nvim theme",
	Long: `Get a specific nvim theme by name from user store or library.

Library themes are automatically available without installation.
User themes with same name override library themes.

Examples:
  dvm get nvim theme tokyonight-night     # Library theme (works instantly)
  dvm get nvim theme coolnight-ocean      # CoolNight library theme
  dvm get nvim theme my-custom -o yaml    # User theme`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getTheme(cmd, args[0])
	},
}

// nvimGetDefaultsCmd lists nvim defaults (namespaced version)
// Usage: dvm get nvim defaults
var nvimGetDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Show current nvim defaults",
	Long: `Show current nvim default configuration values.

Displays nvim-related defaults including:
- nvim-package: Default nvim package for workspaces
- theme: Default theme for workspaces
- plugins: Default plugins for workspaces

Examples:
  dvm get nvim defaults                   # Show table format
  dvm get nvim defaults -o yaml          # Show as YAML
  dvm get nvim defaults -o json          # Show as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getNvimDefaults(cmd)
	},
}

func init() {
	// Add nvim subcommand to get
	getCmd.AddCommand(nvimGetCmd)

	// Add resource types under nvim
	nvimGetCmd.AddCommand(nvimGetPluginsCmd)
	nvimGetCmd.AddCommand(nvimGetPluginCmd)
	nvimGetCmd.AddCommand(nvimGetThemesCmd)
	nvimGetCmd.AddCommand(nvimGetThemeCmd)
	nvimGetCmd.AddCommand(nvimGetDefaultsCmd)

	// Add workspace/app flags to plugins command
	nvimGetPluginsCmd.Flags().StringVarP(&nvimWorkspaceFlag, "workspace", "w", "", "Filter by workspace")
	nvimGetPluginsCmd.Flags().StringVarP(&nvimAppFlag, "app", "a", "", "App for workspace (defaults to active)")
}

// runGetNvimPlugins handles both global and workspace-scoped plugin listing
func runGetNvimPlugins(cmd *cobra.Command, args []string) error {
	// If no workspace flag, show global plugins
	if nvimWorkspaceFlag == "" {
		return getPlugins(cmd)
	}

	// Workspace-scoped: get workspace and its plugin list
	workspace, appName, err := getWorkspaceForPlugins(cmd, nvimAppFlag, nvimWorkspaceFlag)
	if err != nil {
		return err
	}

	// Get the plugin manager to read workspace plugins
	mgr, err := NewWorkspacePluginManager()
	if err != nil {
		return err
	}

	workspacePluginNames := mgr.ListPlugins(workspace)

	// If no plugins configured, show helpful message
	if len(workspacePluginNames) == 0 {
		return renderEmptyWorkspacePlugins(workspace.Name, appName)
	}

	// Get full plugin details from global library
	nvimMgr, err := getNvimManager(cmd)
	if err != nil {
		return err
	}
	defer nvimMgr.Close()

	allPlugins, err := nvimMgr.List()
	if err != nil {
		return fmt.Errorf("failed to list global plugins: %w", err)
	}

	// Filter to only workspace plugins
	plugins := filterPluginsByNames(allPlugins, workspacePluginNames)

	// Render output
	return renderWorkspacePlugins(workspace.Name, appName, plugins, workspacePluginNames)
}

// getWorkspaceForPlugins resolves the workspace from flags or context
func getWorkspaceForPlugins(cmd *cobra.Command, appFlag, workspaceFlag string) (*models.Workspace, string, error) {
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create context manager: %w", err)
	}

	// Resolve app name
	appName := appFlag
	if appName == "" {
		appName, err = ctxMgr.GetActiveApp()
		if err != nil {
			return nil, "", fmt.Errorf("no app specified. Use -a <app> or 'dvm use app <name>' first")
		}
	}

	// Get datastore
	ds, err := getDataStore(cmd)
	if err != nil {
		return nil, "", err
	}

	// Get app (v0.8.0+ uses App model with GetAppByNameGlobal)
	app, err := ds.GetAppByNameGlobal(appName)
	if err != nil {
		return nil, "", fmt.Errorf("app '%s' not found: %w", appName, err)
	}

	// Get workspace
	workspace, err := ds.GetWorkspaceByName(app.ID, workspaceFlag)
	if err != nil {
		return nil, "", fmt.Errorf("workspace '%s' not found in app '%s': %w", workspaceFlag, appName, err)
	}

	return workspace, appName, nil
}

// filterPluginsByNames filters plugins to only those in the names list
func filterPluginsByNames(plugins []*plugin.Plugin, names []string) []*plugin.Plugin {
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	var result []*plugin.Plugin
	for _, p := range plugins {
		if nameSet[p.Name] {
			result = append(result, p)
		}
	}
	return result
}

// renderEmptyWorkspacePlugins renders the empty state for workspace plugins
func renderEmptyWorkspacePlugins(workspaceName, appName string) error {
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		data := struct {
			Workspace string   `json:"workspace" yaml:"workspace"`
			App       string   `json:"app" yaml:"app"`
			Plugins   []string `json:"plugins" yaml:"plugins"`
			Message   string   `json:"message" yaml:"message"`
		}{
			Workspace: workspaceName,
			App:       appName,
			Plugins:   []string{},
			Message:   "No plugins configured. Build will use all global plugins.",
		}
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	render.Info(fmt.Sprintf("Workspace '%s' has no plugins configured", workspaceName))
	render.Info("Build will use all plugins from global library (~/.nvp/plugins/)")
	fmt.Println()
	render.Info("Configure workspace plugins with:")
	render.Info(fmt.Sprintf("  dvm set nvim plugin -w %s <plugin-names...>", workspaceName))
	render.Info(fmt.Sprintf("  dvm set nvim plugin -w %s --all", workspaceName))
	return nil
}

// renderWorkspacePlugins renders the workspace plugin list
func renderWorkspacePlugins(workspaceName, appName string, plugins []*plugin.Plugin, configuredNames []string) error {
	// For JSON/YAML output
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		pluginsYAML := make([]*plugin.PluginYAML, len(plugins))
		for i, p := range plugins {
			pluginsYAML[i] = p.ToYAML()
		}

		data := struct {
			Workspace string               `json:"workspace" yaml:"workspace"`
			App       string               `json:"app" yaml:"app"`
			Plugins   []*plugin.PluginYAML `json:"plugins" yaml:"plugins"`
		}{
			Workspace: workspaceName,
			App:       appName,
			Plugins:   pluginsYAML,
		}
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	// Human-readable output
	render.Success(fmt.Sprintf("Workspace '%s' has %d plugin(s) configured:", workspaceName, len(configuredNames)))
	fmt.Println()

	// Build table
	tableData := render.TableData{
		Headers: []string{"NAME", "CATEGORY", "REPO", "STATUS"},
		Rows:    make([][]string, 0, len(configuredNames)),
	}

	// Create a map for quick lookup
	pluginMap := make(map[string]*plugin.Plugin)
	for _, p := range plugins {
		pluginMap[p.Name] = p
	}

	// Show each configured plugin
	for _, name := range configuredNames {
		p, found := pluginMap[name]
		if found {
			status := "✓ installed"
			if !p.Enabled {
				status = "✗ disabled"
			}
			tableData.Rows = append(tableData.Rows, []string{
				p.Name,
				p.Category,
				p.Repo,
				status,
			})
		} else {
			// Plugin configured but not in global library
			tableData.Rows = append(tableData.Rows, []string{
				name,
				"-",
				"-",
				"⚠ not in library",
			})
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// WorkspacePluginsOutput represents workspace plugins for structured output
type WorkspacePluginsOutput struct {
	Workspace string   `json:"workspace" yaml:"workspace"`
	App       string   `json:"app" yaml:"app"`
	Plugins   []string `json:"plugins" yaml:"plugins"`
}

// getWorkspacePluginNames returns the plugin names for a workspace (helper for other commands)
func getWorkspacePluginNames(workspace *models.Workspace) []string {
	if !workspace.NvimPlugins.Valid || workspace.NvimPlugins.String == "" {
		return nil
	}
	return strings.Split(workspace.NvimPlugins.String, ",")
}

// getNvimDefaults shows current nvim defaults
func getNvimDefaults(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get all defaults from database
	allDefaults, err := ds.ListDefaults()
	if err != nil {
		return fmt.Errorf("failed to list defaults: %w", err)
	}

	// Filter to nvim-related keys
	nvimKeys := []string{"nvim-package", "theme", "plugins"}
	nvimDefaults := make(map[string]interface{})

	for _, key := range nvimKeys {
		if value, exists := allDefaults[key]; exists {
			// Handle plugins array specially - parse JSON if it's valid
			if key == "plugins" && value != "" {
				var plugins []string
				if err := json.Unmarshal([]byte(value), &plugins); err == nil {
					nvimDefaults[key] = plugins
				} else {
					// If JSON parsing fails, store as raw string
					nvimDefaults[key] = value
				}
			} else {
				nvimDefaults[key] = value
			}
		} else {
			// Show empty value for missing keys
			if key == "plugins" {
				nvimDefaults[key] = []string{}
			} else {
				nvimDefaults[key] = ""
			}
		}
	}

	// For structured output (JSON/YAML)
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, nvimDefaults, render.Options{})
	}

	// For table output
	tableData := render.TableData{
		Headers: []string{"KEY", "VALUE"},
		Rows:    make([][]string, 0, len(nvimKeys)),
	}

	for _, key := range nvimKeys {
		value := nvimDefaults[key]
		var displayValue string

		switch v := value.(type) {
		case []string:
			if len(v) == 0 {
				displayValue = "(none)"
			} else {
				displayValue = "[" + strings.Join(v, ", ") + "]"
			}
		case string:
			if v == "" {
				displayValue = "(none)"
			} else {
				displayValue = v
			}
		default:
			displayValue = fmt.Sprintf("%v", v)
		}

		tableData.Rows = append(tableData.Rows, []string{key, displayValue})
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}
