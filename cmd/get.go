package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

var (
	getOutputFormat string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [resource]",
	Short: "Get resources (kubectl-style)",
	Long: `Get resources in various formats (colored, yaml, json, plain).

Resource aliases (kubectl-style):
  apps       → a, app
  workspaces → ws
  workspace  → ws
  context    → ctx
  platforms  → plat
  nvim plugins → np
  nvim themes  → nt

Examples:
  dvm get apps
  dvm get a                       # Same as 'get apps'
  dvm get workspaces
  dvm get ws                      # Same as 'get workspaces'
  dvm get workspace main
  dvm get ws main                 # Same as 'get workspace main'
  dvm get context
  dvm get ctx                     # Same as 'get context'
  dvm get np                      # Same as 'get nvim plugins'
  dvm get nt                      # Same as 'get nvim themes'
  dvm get workspace main -o yaml
  dvm get app my-api -o json
`,
}

// getWorkspacesCmd lists all workspaces in current app
var getWorkspacesCmd = &cobra.Command{
	Use:     "workspaces",
	Aliases: []string{"ws"},
	Short:   "List all workspaces in an app",
	Long: `List all workspaces in an app.

Examples:
  dvm get workspaces              # List workspaces in active app
  dvm get ws                      # Short form
  dvm get workspaces --app myapp  # List workspaces in specific app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspaces(cmd)
	},
}

// getWorkspaceCmd gets a specific workspace
var getWorkspaceCmd = &cobra.Command{
	Use:     "workspace [name]",
	Aliases: []string{"ws"},
	Short:   "Get a specific workspace",
	Long: `Get a specific workspace by name.

Examples:
  dvm get workspace main              # Get workspace from active app
  dvm get ws main                     # Short form
  dvm get workspace main --app myapp  # Get workspace from specific app
  dvm get workspace main -o yaml      # Output as YAML`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspace(cmd, args[0])
	},
}

// getPlatformsCmd lists all detected container platforms
var getPlatformsCmd = &cobra.Command{
	Use:     "platforms",
	Aliases: []string{"plat"},
	Short:   "List all detected container platforms",
	Long: `List all detected container platforms (OrbStack, Colima, Docker Desktop, Podman).

Examples:
  dvm get platforms
  dvm get plat          # Short form
  dvm get platforms -o yaml
  dvm get platforms -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPlatforms(cmd)
	},
}

// getContextCmd displays the current active context
var getContextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"ctx"},
	Short:   "Display the current context",
	Long: `Display the current active app and workspace context.

The context determines which app and workspace commands operate on by default.
Set context with 'dvm use app <name>' and 'dvm use workspace <name>'.

Context can also be set via environment variables:
  DVM_APP        - Override active app
  DVM_WORKSPACE  - Override active workspace

Examples:
  dvm get context       # Show current context
  dvm get ctx           # Short form
  dvm get context -o yaml
  dvm get context -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getContext(cmd)
	},
}

// getNvimPluginsShortCmd is a top-level shortcut for 'dvm get nvim plugins'
// Usage: dvm get np
var getNvimPluginsShortCmd = &cobra.Command{
	Use:   "np",
	Short: "List all nvim plugins (shortcut for 'nvim plugins')",
	Long: `List all nvim plugins stored in the database.

This is a shortcut for 'dvm get nvim plugins'.

Examples:
  dvm get np
  dvm get np -o yaml
  dvm get np -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPlugins(cmd)
	},
}

// getNvimThemesShortCmd is a top-level shortcut for 'dvm get nvim themes'
// Usage: dvm get nt
var getNvimThemesShortCmd = &cobra.Command{
	Use:   "nt",
	Short: "List all nvim themes (shortcut for 'nvim themes')",
	Long: `List all nvim themes stored in the database.

This is a shortcut for 'dvm get nvim themes'.

Examples:
  dvm get nt
  dvm get nt -o yaml
  dvm get nt -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getThemes(cmd)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getWorkspacesCmd)
	getCmd.AddCommand(getWorkspaceCmd)
	getCmd.AddCommand(getPlatformsCmd)
	getCmd.AddCommand(getContextCmd)

	// Add top-level shortcuts for nvim resources
	getCmd.AddCommand(getNvimPluginsShortCmd)
	getCmd.AddCommand(getNvimThemesShortCmd)

	// Add output format flag to all get commands
	// Maps to render package: json, yaml, plain, table, colored (default)
	getCmd.PersistentFlags().StringVarP(&getOutputFormat, "output", "o", "", "Output format (json, yaml, plain, table, colored)")

	// Add app flag for workspace commands
	getWorkspacesCmd.Flags().StringP("app", "a", "", "App name (defaults to active app)")
	getWorkspaceCmd.Flags().StringP("app", "a", "", "App name (defaults to active app)")
}

func getDataStore(cmd *cobra.Command) (db.DataStore, error) {
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return nil, fmt.Errorf("dataStore not initialized")
	}

	return *dataStore, nil
}

// ContextOutput represents context for output formatting
type ContextOutput struct {
	CurrentApp       string `yaml:"currentApp" json:"currentApp"`
	CurrentWorkspace string `yaml:"currentWorkspace" json:"currentWorkspace"`
}

func getContext(cmd *cobra.Command) error {
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	ctx, err := ctxMgr.LoadContext()
	if err != nil {
		return fmt.Errorf("failed to load context: %w", err)
	}

	// Build structured data
	data := ContextOutput{
		CurrentApp:       ctx.CurrentApp,
		CurrentWorkspace: ctx.CurrentWorkspace,
	}

	// Check if empty
	isEmpty := ctx.CurrentApp == ""

	// For structured output (JSON/YAML), always output the data structure
	// For human output, show nice key-value display
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	// Human-readable output
	if isEmpty {
		return render.Output(nil, render.Options{
			Empty:        true,
			EmptyMessage: "No active context",
			EmptyHints: []string{
				"dvm use app <name>",
				"dvm use workspace <name>",
			},
		})
	}

	workspace := ctx.CurrentWorkspace
	if workspace == "" {
		workspace = "(none)"
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "App", Value: ctx.CurrentApp},
		render.KeyValue{Key: "Workspace", Value: workspace},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Current Context",
	})
}

func getWorkspaces(cmd *cobra.Command) error {
	// Get app from flag or context
	appFlag, _ := cmd.Flags().GetString("app")

	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	var appName string
	if appFlag != "" {
		appName = appFlag
	} else {
		appName, err = ctxMgr.GetActiveApp()
		if err != nil {
			return fmt.Errorf("no app specified. Use --app <name> or 'dvm use app <name>' first")
		}
	}

	// Get active workspace (only relevant if viewing active app)
	var activeWorkspace string
	activeApp, _ := ctxMgr.GetActiveApp()
	if activeApp == appName {
		activeWorkspace, _ = ctxMgr.GetActiveWorkspace()
	}

	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get app to get its ID (search globally across all domains)
	app, err := sqlDS.GetAppByNameGlobal(appName)
	if err != nil {
		return fmt.Errorf("app '%s' not found: %w", appName, err)
	}

	// List workspaces for this app
	workspaces, err := sqlDS.ListWorkspacesByApp(app.ID)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: fmt.Sprintf("No workspaces found in app '%s'", appName),
			EmptyHints:   []string{"dvm create workspace <name>"},
		})
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		workspacesYAML := make([]models.WorkspaceYAML, len(workspaces))
		for i, ws := range workspaces {
			workspacesYAML[i] = ws.ToYAML(appName)
		}
		return render.OutputWith(getOutputFormat, workspacesYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "APP", "IMAGE", "STATUS", "CREATED"},
		Rows:    make([][]string, len(workspaces)),
	}

	for i, ws := range workspaces {
		name := ws.Name
		if ws.Name == activeWorkspace {
			name = "● " + name
		}
		tableData.Rows[i] = []string{
			name,
			appName,
			ws.ImageName,
			ws.Status,
			ws.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

func getWorkspace(cmd *cobra.Command, name string) error {
	// Get app from flag or context
	appFlag, _ := cmd.Flags().GetString("app")

	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	var appName string
	if appFlag != "" {
		appName = appFlag
	} else {
		appName, err = ctxMgr.GetActiveApp()
		if err != nil {
			return fmt.Errorf("no app specified. Use --app <name> or 'dvm use app <name>' first")
		}
	}

	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get app to get its ID (search globally across all domains)
	app, err := sqlDS.GetAppByNameGlobal(appName)
	if err != nil {
		return fmt.Errorf("app '%s' not found: %w", appName, err)
	}

	workspace, err := sqlDS.GetWorkspaceByName(app.ID, name)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, workspace.ToYAML(appName), render.Options{})
	}

	// For human output, show detail view
	var activeWorkspace string
	activeApp, _ := ctxMgr.GetActiveApp()
	if activeApp == appName {
		activeWorkspace, _ = ctxMgr.GetActiveWorkspace()
	}

	isActive := workspace.Name == activeWorkspace
	nameDisplay := workspace.Name
	if isActive {
		nameDisplay = "● " + nameDisplay + " (active)"
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: nameDisplay},
		render.KeyValue{Key: "App", Value: appName},
		render.KeyValue{Key: "Image", Value: workspace.ImageName},
		render.KeyValue{Key: "Status", Value: workspace.Status},
		render.KeyValue{Key: "Created", Value: workspace.CreatedAt.Format("2006-01-02 15:04:05")},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Workspace Details",
	})
}

func getPlugins(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindNvimPlugin)
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No plugins found",
			EmptyHints:   []string{"dvm apply -f plugin.yaml"},
		})
	}

	// Extract underlying plugins from resources
	plugins := make([]*plugin.Plugin, len(resources))
	for i, res := range resources {
		pr := res.(*handlers.NvimPluginResource)
		plugins[i] = pr.Plugin()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		pluginsYAML := make([]*plugin.PluginYAML, len(plugins))
		for i, p := range plugins {
			pluginsYAML[i] = p.ToYAML()
		}
		return render.OutputWith(getOutputFormat, pluginsYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "CATEGORY", "REPO", "VERSION"},
		Rows:    make([][]string, len(plugins)),
	}

	for i, p := range plugins {
		version := "latest"
		if p.Version != "" {
			version = p.Version
		} else if p.Branch != "" {
			version = "branch:" + p.Branch
		}

		enabledMark := "✓"
		if !p.Enabled {
			enabledMark = "✗"
		}

		tableData.Rows[i] = []string{
			p.Name + " " + enabledMark,
			p.Category,
			p.Repo,
			version,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

func getPlugin(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindNvimPlugin, name)
	if err != nil {
		return fmt.Errorf("failed to get plugin '%s': %w", name, err)
	}

	p := res.(*handlers.NvimPluginResource).Plugin()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, p.ToYAML(), render.Options{})
	}

	// For human output, show detail view
	version := "latest"
	if p.Version != "" {
		version = p.Version
	} else if p.Branch != "" {
		version = "branch:" + p.Branch
	}

	enabledStr := "yes"
	if !p.Enabled {
		enabledStr = "no"
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: p.Name},
		render.KeyValue{Key: "Repo", Value: p.Repo},
		render.KeyValue{Key: "Category", Value: p.Category},
		render.KeyValue{Key: "Version", Value: version},
		render.KeyValue{Key: "Enabled", Value: enabledStr},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Plugin Details",
	})
}

func getThemes(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindNvimTheme)
	if err != nil {
		return fmt.Errorf("failed to list themes: %w", err)
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No themes found",
			EmptyHints:   []string{"dvm apply -f theme.yaml"},
		})
	}

	// Extract underlying themes from resources
	themes := make([]*theme.Theme, len(resources))
	for i, res := range resources {
		tr := res.(*handlers.NvimThemeResource)
		themes[i] = tr.Theme()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		themesYAML := make([]*theme.ThemeYAML, len(themes))
		for i, t := range themes {
			themesYAML[i] = &theme.ThemeYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "NvimTheme",
				Metadata: theme.ThemeMetadata{
					Name:        t.Name,
					Description: t.Description,
					Author:      t.Author,
					Category:    t.Category,
				},
				Spec: theme.ThemeSpec{
					Plugin:      t.Plugin,
					Style:       t.Style,
					Transparent: t.Transparent,
					Colors:      t.Colors,
					Options:     t.Options,
				},
			}
		}
		return render.OutputWith(getOutputFormat, themesYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "CATEGORY", "PLUGIN", "STYLE"},
		Rows:    make([][]string, len(themes)),
	}

	for i, t := range themes {
		category := t.Category
		if category == "" {
			category = "-"
		}
		style := t.Style
		if style == "" {
			style = "default"
		}

		tableData.Rows[i] = []string{
			t.Name,
			category,
			t.Plugin.Repo,
			style,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

func getTheme(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindNvimTheme, name)
	if err != nil {
		return fmt.Errorf("failed to get theme '%s': %w", name, err)
	}

	t := res.(*handlers.NvimThemeResource).Theme()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		themeYAML := &theme.ThemeYAML{
			APIVersion: "devopsmaestro.io/v1",
			Kind:       "NvimTheme",
			Metadata: theme.ThemeMetadata{
				Name:        t.Name,
				Description: t.Description,
				Author:      t.Author,
				Category:    t.Category,
			},
			Spec: theme.ThemeSpec{
				Plugin:      t.Plugin,
				Style:       t.Style,
				Transparent: t.Transparent,
				Colors:      t.Colors,
				Options:     t.Options,
			},
		}
		return render.OutputWith(getOutputFormat, themeYAML, render.Options{})
	}

	// For human output, show detail view
	category := t.Category
	if category == "" {
		category = "-"
	}
	style := t.Style
	if style == "" {
		style = "default"
	}
	transparent := "no"
	if t.Transparent {
		transparent = "yes"
	}

	pairs := []render.KeyValue{
		{Key: "Name", Value: t.Name},
		{Key: "Plugin", Value: t.Plugin.Repo},
		{Key: "Category", Value: category},
		{Key: "Style", Value: style},
		{Key: "Transparent", Value: transparent},
	}

	if t.Description != "" {
		pairs = append(pairs, render.KeyValue{Key: "Description", Value: t.Description})
	}
	if t.Author != "" {
		pairs = append(pairs, render.KeyValue{Key: "Author", Value: t.Author})
	}

	kvData := render.NewOrderedKeyValueData(pairs...)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Theme Details",
	})
}

// PlatformOutput represents a platform for output
type PlatformOutput struct {
	Type         string `yaml:"type" json:"type"`
	Name         string `yaml:"name" json:"name"`
	SocketPath   string `yaml:"socketPath" json:"socketPath"`
	Profile      string `yaml:"profile,omitempty" json:"profile,omitempty"`
	IsContainerd bool   `yaml:"isContainerd" json:"isContainerd"`
	IsDocker     bool   `yaml:"isDockerCompatible" json:"isDockerCompatible"`
	Active       bool   `yaml:"active" json:"active"`
}

func getPlatforms(cmd *cobra.Command) error {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return fmt.Errorf("failed to create platform detector: %w", err)
	}

	platforms := detector.DetectAll()

	if len(platforms) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No container platforms detected",
			EmptyHints:   []string{"Install OrbStack, Colima, Docker Desktop, or Podman"},
		})
	}

	// Get active platform
	activePlatform, _ := detector.Detect()
	var activeName string
	if activePlatform != nil {
		activeName = string(activePlatform.Type)
	}

	// Build platform output data
	platformsOutput := make([]PlatformOutput, len(platforms))
	for i, p := range platforms {
		platformsOutput[i] = PlatformOutput{
			Type:         string(p.Type),
			Name:         p.Name,
			SocketPath:   p.SocketPath,
			Profile:      p.Profile,
			IsContainerd: p.IsContainerd(),
			IsDocker:     p.IsDockerCompatible(),
			Active:       string(p.Type) == activeName,
		}
	}

	// For JSON/YAML, output directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, platformsOutput, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"TYPE", "NAME", "SOCKET", "STATUS"},
		Rows:    make([][]string, len(platforms)),
	}

	for i, p := range platforms {
		status := ""
		if platformsOutput[i].Active {
			status = "● active"
		}

		socketDisplay := p.SocketPath
		if len(socketDisplay) > 45 {
			socketDisplay = "..." + socketDisplay[len(socketDisplay)-42:]
		}

		name := p.Name
		if p.IsContainerd() {
			name += " (containerd)"
		} else if p.IsDockerCompatible() {
			name += " (docker)"
		}

		tableData.Rows[i] = []string{
			string(p.Type),
			name,
			socketDisplay,
			status,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}
