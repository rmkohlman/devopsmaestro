package cmd

import (
	"fmt"

	"devopsmaestro/builders"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/nvimops"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/pkg/terminalops/shell"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

var (
	getOutputFormat    string
	getWorkspacesFlags HierarchyFlags
	getWorkspaceFlags  HierarchyFlags
	showTheme          bool // Flag to show theme resolution information
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
  dvm get nt                      # Same as 'get nvim themes' (34+ library themes)
  dvm get nvim theme coolnight-ocean    # Library theme (no install needed)
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

Flags:
  -A, --all         List all workspaces across all apps/domains/ecosystems
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name

Examples:
  dvm get workspaces              # List workspaces in active app
  dvm get ws                      # Short form
  dvm get workspaces -A           # List ALL workspaces across everything
  dvm get workspaces -a myapp     # List workspaces in specific app
  dvm get workspaces -e healthcare -a portal`,
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

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name (alternative to positional arg)

Examples:
  dvm get workspace main              # Get workspace from active app
  dvm get ws main                     # Short form
  dvm get workspace main -a myapp     # Get workspace from specific app
  dvm get workspace -a portal         # Get workspace if only one exists
  dvm get workspace main -o yaml      # Output as YAML`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return getWorkspace(cmd, name)
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
	Long: `List all nvim themes from user store and embedded library.

This is a shortcut for 'dvm get nvim themes'.
Shows 34+ library themes automatically available without installation.

Examples:
  dvm get nt
  dvm get nt -o yaml
  dvm get nt -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getThemes(cmd)
	},
}

// getDefaultsCmd displays default configuration values
var getDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Display default configuration values",
	Long: `Display default configuration values for containers and shells.

Shows the default values used when creating new workspaces if no explicit 
configuration is provided.

Examples:
  dvm get defaults
  dvm get defaults -o yaml
  dvm get defaults -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getDefaults(cmd)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getWorkspacesCmd)
	getCmd.AddCommand(getWorkspaceCmd)
	getCmd.AddCommand(getPlatformsCmd)
	getCmd.AddCommand(getContextCmd)
	getCmd.AddCommand(getDefaultsCmd)

	// Add top-level shortcuts for nvim resources
	getCmd.AddCommand(getNvimPluginsShortCmd)
	getCmd.AddCommand(getNvimThemesShortCmd)

	// Add output format flag to all get commands
	// Maps to render package: json, yaml, plain, table, colored (default)
	getCmd.PersistentFlags().StringVarP(&getOutputFormat, "output", "o", "", "Output format (json, yaml, plain, table, colored)")

	// Add hierarchy flags for workspace commands
	AddHierarchyFlags(getWorkspacesCmd, &getWorkspacesFlags)
	AddHierarchyFlags(getWorkspaceCmd, &getWorkspaceFlags)

	// Add --all flag to get workspaces (with -A shorthand for consistency)
	getWorkspacesCmd.Flags().BoolP("all", "A", false, "List all workspaces across all apps/domains/ecosystems")

	// Add --show-theme flag to hierarchy commands
	getWorkspacesCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	getWorkspaceCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
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
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	allFlag, _ := cmd.Flags().GetBool("all")

	// If --all/-A flag is set, list all workspaces across everything
	if allFlag {
		workspaces, err := sqlDS.ListAllWorkspaces()
		if err != nil {
			return fmt.Errorf("failed to list all workspaces: %w", err)
		}

		if len(workspaces) == 0 {
			return render.OutputWith(getOutputFormat, nil, render.Options{
				Empty:        true,
				EmptyMessage: "No workspaces found",
				EmptyHints:   []string{"dvm create workspace <name>"},
			})
		}

		// For JSON/YAML, output the model data directly
		if getOutputFormat == "json" || getOutputFormat == "yaml" {
			workspacesYAML := make([]models.WorkspaceYAML, len(workspaces))
			for i, ws := range workspaces {
				// Get app name for this workspace
				app, _ := sqlDS.GetAppByID(ws.AppID)
				appName := ""
				if app != nil {
					appName = app.Name
				}
				workspacesYAML[i] = ws.ToYAML(appName)
			}
			return render.OutputWith(getOutputFormat, workspacesYAML, render.Options{})
		}

		// For human output, build table data
		// We need to look up app names for display
		headers := []string{"NAME", "APP", "IMAGE", "STATUS"}
		if showTheme {
			headers = append(headers, "THEME", "THEME SOURCE")
		}

		tableData := render.TableData{
			Headers: headers,
			Rows:    make([][]string, len(workspaces)),
		}

		// Create theme resolver if needed
		var themeResolver themeresolver.ThemeResolver
		if showTheme {
			themeResolver, _ = themeresolver.NewThemeResolver(sqlDS, nil)
		}

		for i, ws := range workspaces {
			app, _ := sqlDS.GetAppByID(ws.AppID)
			appName := ""
			if app != nil {
				appName = app.Name
			}

			row := []string{
				ws.Name,
				appName,
				ws.ImageName,
				ws.Status,
			}

			// Add theme information if requested
			if showTheme && themeResolver != nil {
				themeName := themeresolver.DefaultTheme
				themeSource := "default"

				if resolution, err := themeResolver.GetResolutionPath(cmd.Context(), themeresolver.LevelWorkspace, ws.ID); err == nil {
					if resolution.Source != themeresolver.LevelGlobal {
						themeName = resolution.GetEffectiveThemeName()
						themeSource = resolution.Source.String()
					}
				}

				row = append(row, themeName, themeSource)
			}

			tableData.Rows[i] = row
		}

		return render.OutputWith(getOutputFormat, tableData, render.Options{
			Type: render.TypeTable,
		})
	}

	// Check if hierarchy flags were provided
	if getWorkspacesFlags.HasAnyFlag() {
		// Use resolver to find matching workspaces
		wsResolver := resolver.NewWorkspaceResolver(sqlDS)
		results, err := wsResolver.ResolveAll(getWorkspacesFlags.ToFilter())
		if err != nil {
			if resolver.IsNoWorkspaceFoundError(err) {
				return render.OutputWith(getOutputFormat, nil, render.Options{
					Empty:        true,
					EmptyMessage: "No workspaces found matching criteria",
					EmptyHints:   []string{"dvm create workspace <name>"},
				})
			}
			return fmt.Errorf("failed to resolve workspaces: %w", err)
		}

		if len(results) == 0 {
			return render.OutputWith(getOutputFormat, nil, render.Options{
				Empty:        true,
				EmptyMessage: "No workspaces found matching criteria",
				EmptyHints:   []string{"dvm create workspace <name>"},
			})
		}

		// For JSON/YAML, output the model data directly
		if getOutputFormat == "json" || getOutputFormat == "yaml" {
			workspacesYAML := make([]models.WorkspaceYAML, len(results))
			for i, wh := range results {
				workspacesYAML[i] = wh.Workspace.ToYAML(wh.App.Name)
			}
			return render.OutputWith(getOutputFormat, workspacesYAML, render.Options{})
		}

		// For human output, build table data with full path
		headers := []string{"NAME", "PATH", "IMAGE", "STATUS"}
		if showTheme {
			headers = append(headers, "THEME", "THEME SOURCE")
		}

		tableData := render.TableData{
			Headers: headers,
			Rows:    make([][]string, len(results)),
		}

		// Create theme resolver if needed
		var themeResolver themeresolver.ThemeResolver
		if showTheme {
			themeResolver, _ = themeresolver.NewThemeResolver(sqlDS, nil)
		}

		for i, wh := range results {
			row := []string{
				wh.Workspace.Name,
				wh.FullPath(),
				wh.Workspace.ImageName,
				wh.Workspace.Status,
			}

			// Add theme information if requested
			if showTheme && themeResolver != nil {
				themeName := themeresolver.DefaultTheme
				themeSource := "default"

				if resolution, err := themeResolver.GetResolutionPath(cmd.Context(), themeresolver.LevelWorkspace, wh.Workspace.ID); err == nil {
					if resolution.Source != themeresolver.LevelGlobal {
						themeName = resolution.GetEffectiveThemeName()
						themeSource = resolution.Source.String()
					}
				}

				row = append(row, themeName, themeSource)
			}

			tableData.Rows[i] = row
		}

		return render.OutputWith(getOutputFormat, tableData, render.Options{
			Type: render.TypeTable,
		})
	}

	// Fall back to existing context-based behavior
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	appName, err := ctxMgr.GetActiveApp()
	if err != nil {
		return fmt.Errorf("no app specified. Use -a <name> or 'dvm use app <name>' first")
	}

	// Get active workspace (only relevant if viewing active app)
	var activeWorkspace string
	activeApp, _ := ctxMgr.GetActiveApp()
	if activeApp == appName {
		activeWorkspace, _ = ctxMgr.GetActiveWorkspace()
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
	headers := []string{"NAME", "APP", "IMAGE", "STATUS", "CREATED"}
	if showTheme {
		headers = append(headers, "THEME", "THEME SOURCE")
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    make([][]string, len(workspaces)),
	}

	// Create theme resolver if needed
	var themeResolver themeresolver.ThemeResolver
	if showTheme {
		themeResolver, _ = themeresolver.NewThemeResolver(sqlDS, nil)
	}

	for i, ws := range workspaces {
		name := ws.Name
		if ws.Name == activeWorkspace {
			name = "● " + name
		}

		row := []string{
			name,
			appName,
			ws.ImageName,
			ws.Status,
			ws.CreatedAt.Format("2006-01-02 15:04"),
		}

		// Add theme information if requested
		if showTheme && themeResolver != nil {
			themeName := themeresolver.DefaultTheme
			themeSource := "default"

			if resolution, err := themeResolver.GetResolutionPath(cmd.Context(), themeresolver.LevelWorkspace, ws.ID); err == nil {
				if resolution.Source != themeresolver.LevelGlobal {
					themeName = resolution.GetEffectiveThemeName()
					themeSource = resolution.Source.String()
				}
			}

			row = append(row, themeName, themeSource)
		}

		tableData.Rows[i] = row
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

func getWorkspace(cmd *cobra.Command, name string) error {
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	var workspace *models.Workspace
	var app *models.App
	var appName string

	// If name is provided via positional arg, add it to the filter
	filter := getWorkspaceFlags.ToFilter()
	if name != "" {
		filter.WorkspaceName = name
	}

	// Check if any criteria were provided (flags or positional arg)
	if filter.EcosystemName != "" || filter.DomainName != "" || filter.AppName != "" || filter.WorkspaceName != "" {
		// Use resolver to find workspace
		wsResolver := resolver.NewWorkspaceResolver(sqlDS)
		result, err := wsResolver.Resolve(filter)
		if err != nil {
			// Check if ambiguous and provide helpful output
			if ambiguousErr, ok := resolver.IsAmbiguousError(err); ok {
				render.Warning("Multiple workspaces match your criteria")
				fmt.Println(ambiguousErr.FormatDisambiguation())
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Info("Hint: Use 'dvm get workspaces' to see available workspaces")
				return err
			}
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}

		workspace = result.Workspace
		app = result.App
		appName = app.Name

		// Update context to the resolved workspace
		if err := updateContextFromHierarchy(sqlDS, result); err != nil {
			// Continue anyway - this is not fatal
		}
	} else {
		// Fall back to existing context-based behavior
		ctxMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to create context manager: %w", err)
		}

		appName, err = ctxMgr.GetActiveApp()
		if err != nil {
			return fmt.Errorf("no app specified. Use -a <name> or 'dvm use app <name>' first")
		}

		// Get app to get its ID (search globally across all domains)
		app, err = sqlDS.GetAppByNameGlobal(appName)
		if err != nil {
			return fmt.Errorf("app '%s' not found: %w", appName, err)
		}

		// Need workspace name when using context-based lookup
		return fmt.Errorf("workspace name required. Use: dvm get workspace <name>")
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, workspace.ToYAML(appName), render.Options{})
	}

	// For human output, show detail view
	ctxMgr, _ := operators.NewContextManager()
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

	err = render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Workspace Details",
	})
	if err != nil {
		return err
	}

	// Show theme information if requested
	if showTheme {
		return showThemeResolution(cmd, sqlDS, themeresolver.LevelWorkspace, workspace.ID, workspace.Name)
	}

	return nil
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
			EmptyMessage: "No user themes found (34+ library themes available automatically)",
			EmptyHints:   []string{"dvm get nvim theme coolnight-ocean", "dvm apply -f theme.yaml"},
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

// showThemeResolution displays theme resolution information for a given hierarchy level and object ID
func showThemeResolution(cmd *cobra.Command, ds db.DataStore, level themeresolver.HierarchyLevel, objectID int, objectName string) error {
	// Create theme resolver
	themeResolver, err := themeresolver.NewThemeResolver(ds, nil) // nil theme store for now
	if err != nil {
		return fmt.Errorf("failed to create theme resolver: %w", err)
	}

	// Get theme resolution path
	resolution, err := themeResolver.GetResolutionPath(cmd.Context(), level, objectID)
	if err != nil {
		return fmt.Errorf("failed to resolve theme: %w", err)
	}

	fmt.Println()
	render.Info("Theme Resolution:")

	if resolution.Source != themeresolver.LevelGlobal {
		fmt.Printf("  Effective theme: %s\n", resolution.GetEffectiveThemeName())
		fmt.Printf("  Source: %s\n", resolution.GetSourceDescription())
	} else {
		fmt.Printf("  Effective theme: %s (default)\n", themeresolver.DefaultTheme)
		fmt.Printf("  Source: global default\n")
	}

	if len(resolution.Path) > 0 {
		fmt.Println("  Resolution path:")
		for _, step := range resolution.Path {
			status := "○" // Empty circle
			if step.Found && step.ThemeName != "" {
				status = "●" // Filled circle
			}

			fmt.Printf("    %s %s '%s'", status, step.Level.String(), step.Name)
			if step.ThemeName != "" {
				fmt.Printf(" → %s", step.ThemeName)
			}
			if step.Error != "" {
				fmt.Printf(" (error: %s)", step.Error)
			}
			fmt.Println()
		}
	}

	fmt.Println()
	render.Info("Legend: ● theme set, ○ no theme (inherits from parent)")

	return nil
}

// DefaultsOutput represents default configuration values for output
type DefaultsOutput struct {
	Theme     map[string]interface{} `yaml:"theme" json:"theme"`
	Shell     map[string]interface{} `yaml:"shell" json:"shell"`
	Nvim      map[string]interface{} `yaml:"nvim" json:"nvim"`
	Container map[string]interface{} `yaml:"container" json:"container"`
}

func getDefaults(cmd *cobra.Command) error {
	// Get defaults from all packages (hardcoded defaults)
	themeDefaults := themeresolver.GetDefaults()
	shellDefaults := shell.GetDefaults()
	nvimDefaults := nvimops.GetDefaults()
	containerDefaults := builders.GetContainerDefaults()

	// Override with user-set defaults from database
	ds, err := getDataStore(cmd)
	if err == nil {
		// Check for user-set nvim package
		if userPkg, err := ds.GetDefault("nvim-package"); err == nil && userPkg != "" {
			nvimDefaults["pluginPackage"] = userPkg
		}
		// Check for user-set terminal package
		if userTermPkg, err := ds.GetDefault("terminal-package"); err == nil && userTermPkg != "" {
			shellDefaults["terminalPackage"] = userTermPkg
		}
		// Check for user-set global theme
		if userTheme, err := ds.GetDefault("theme"); err == nil && userTheme != "" {
			themeDefaults["global"] = userTheme
		}
	}

	// Build structured data
	data := DefaultsOutput{
		Theme:     themeDefaults,
		Shell:     shellDefaults,
		Nvim:      nvimDefaults,
		Container: containerDefaults,
	}

	// For JSON/YAML, output the data structure directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	// For human-readable output, show organized key-value display
	fmt.Println()
	render.Info("Theme Defaults:")
	for key, value := range themeDefaults {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println()
	render.Info("Shell Defaults:")
	for key, value := range shellDefaults {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println()
	render.Info("Neovim Defaults:")
	for key, value := range nvimDefaults {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println()
	render.Info("Container Defaults:")
	for key, value := range containerDefaults {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return nil
}
