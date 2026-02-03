package cmd

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/output"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	getOutputFormat string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [resource]",
	Short: "Get resources (kubectl-style)",
	Long: `Get resources in various formats (table, yaml, json).

Examples:
  dvm get projects
  dvm get workspaces
  dvm get project my-api
  dvm get workspace main
  dvm get workspace main -o yaml
  dvm get project my-api -o yaml
`,
}

// getProjectsCmd lists all projects
var getProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getProjects(cmd)
	},
}

// getProjectCmd gets a specific project
var getProjectCmd = &cobra.Command{
	Use:   "project [name]",
	Short: "Get a specific project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getProject(cmd, args[0])
	},
}

// getWorkspacesCmd lists all workspaces in current project
var getWorkspacesCmd = &cobra.Command{
	Use:   "workspaces",
	Short: "List all workspaces in a project",
	Long: `List all workspaces in a project.

Examples:
  dvm get workspaces              # List workspaces in active project
  dvm get workspaces -p myproject # List workspaces in specific project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspaces(cmd)
	},
}

// getWorkspaceCmd gets a specific workspace
var getWorkspaceCmd = &cobra.Command{
	Use:   "workspace [name]",
	Short: "Get a specific workspace",
	Long: `Get a specific workspace by name.

Examples:
  dvm get workspace main              # Get workspace from active project
  dvm get workspace main -p myproject # Get workspace from specific project
  dvm get workspace main -o yaml      # Output as YAML`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspace(cmd, args[0])
	},
}

// getPlatformsCmd lists all detected container platforms
var getPlatformsCmd = &cobra.Command{
	Use:   "platforms",
	Short: "List all detected container platforms",
	Long: `List all detected container platforms (OrbStack, Colima, Docker Desktop, Podman).

Examples:
  dvm get platforms
  dvm get platforms -o yaml
  dvm get platforms -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPlatforms(cmd)
	},
}

// getContextCmd displays the current active context
var getContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show current active context",
	Long: `Display the currently active project and workspace context.

The context determines the default project and workspace used by commands
when -p (project) or -w (workspace) flags are not specified.

Context can be set with:
  dvm use project <name>      # Set active project
  dvm use workspace <name>    # Set active workspace

Context can also be overridden with environment variables:
  DVM_PROJECT=myproject dvm get workspaces
  DVM_WORKSPACE=dev dvm attach

Examples:
  dvm get context
  dvm get context -o yaml
  dvm get context -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getContext(cmd)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getProjectsCmd)
	getCmd.AddCommand(getProjectCmd)
	getCmd.AddCommand(getWorkspacesCmd)
	getCmd.AddCommand(getWorkspaceCmd)
	getCmd.AddCommand(getPlatformsCmd)
	getCmd.AddCommand(getContextCmd)

	// Add output format flag to all get commands
	// Default to empty string which maps to "colored" formatter for rich terminal output
	getCmd.PersistentFlags().StringVarP(&getOutputFormat, "output", "o", "", "Output format (table, yaml, json, plain)")

	// Add project flag for workspace commands
	getWorkspacesCmd.Flags().StringP("project", "p", "", "Project name (defaults to active project)")
	getWorkspaceCmd.Flags().StringP("project", "p", "", "Project name (defaults to active project)")
}

func getDataStore(cmd *cobra.Command) (db.DataStore, error) {
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return nil, fmt.Errorf("dataStore not initialized")
	}

	return *dataStore, nil
}

func getProjects(cmd *cobra.Command) error {
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	projects, err := sqlDS.ListProjects()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	// Get active project
	ctxMgr, err := operators.NewContextManager()
	var activeProject string
	if err == nil {
		activeProject, _ = ctxMgr.GetActiveProject()
	}

	switch getOutputFormat {
	case "yaml":
		return outputProjectsYAML(projects)
	case "json":
		return outputProjectsJSON(projects)
	default:
		return outputProjectsTable(projects, activeProject)
	}
}

func getProject(cmd *cobra.Command, name string) error {
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	project, err := sqlDS.GetProjectByName(name)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get active project
	ctxMgr, err := operators.NewContextManager()
	var activeProject string
	if err == nil {
		activeProject, _ = ctxMgr.GetActiveProject()
	}

	switch getOutputFormat {
	case "yaml":
		return outputProjectYAML(project)
	case "json":
		return outputProjectJSON(project)
	default:
		return outputProjectsTable([]*models.Project{project}, activeProject)
	}
}

func getWorkspaces(cmd *cobra.Command) error {
	// Get project from flag or context
	projectFlag, _ := cmd.Flags().GetString("project")

	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	var projectName string
	if projectFlag != "" {
		projectName = projectFlag
	} else {
		projectName, err = ctxMgr.GetActiveProject()
		if err != nil {
			return fmt.Errorf("no project specified. Use -p <project> or 'dvm use project <name>' first")
		}
	}

	// Get active workspace (only relevant if viewing active project)
	var activeWorkspace string
	activeProject, _ := ctxMgr.GetActiveProject()
	if activeProject == projectName {
		activeWorkspace, _ = ctxMgr.GetActiveWorkspace()
	}

	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get project to get its ID
	project, err := sqlDS.GetProjectByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// List workspaces for this project
	workspaces, err := sqlDS.ListWorkspacesByProject(project.ID)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		fmt.Printf("No workspaces found in project '%s'\n", projectName)
		return nil
	}

	switch getOutputFormat {
	case "yaml":
		return outputWorkspacesYAML(workspaces, projectName)
	case "json":
		return outputWorkspacesJSON(workspaces, projectName)
	default:
		return outputWorkspacesTable(workspaces, projectName, activeWorkspace)
	}
}

func getWorkspace(cmd *cobra.Command, name string) error {
	// Get project from flag or context
	projectFlag, _ := cmd.Flags().GetString("project")

	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	var projectName string
	if projectFlag != "" {
		projectName = projectFlag
	} else {
		projectName, err = ctxMgr.GetActiveProject()
		if err != nil {
			return fmt.Errorf("no project specified. Use -p <project> or 'dvm use project <name>' first")
		}
	}

	// Get active workspace (only relevant if viewing active project)
	var activeWorkspace string
	activeProject, _ := ctxMgr.GetActiveProject()
	if activeProject == projectName {
		activeWorkspace, _ = ctxMgr.GetActiveWorkspace()
	}

	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get project to get its ID
	project, err := sqlDS.GetProjectByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	workspace, err := sqlDS.GetWorkspaceByName(project.ID, name)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	switch getOutputFormat {
	case "yaml":
		return outputWorkspaceYAML(workspace, project.Name)
	case "json":
		return outputWorkspaceJSON(workspace, project.Name)
	default:
		return outputWorkspacesTable([]*models.Workspace{workspace}, project.Name, activeWorkspace)
	}
}

func getPlugins(cmd *cobra.Command) error {
	// Use nvimops.Manager for unified storage with nvp CLI
	mgr, err := getNvimManager(cmd)
	if err != nil {
		return err
	}
	defer mgr.Close()

	plugins, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	if len(plugins) == 0 {
		fmt.Println(ui.MutedStyle.Render("No plugins found"))
		return nil
	}

	switch getOutputFormat {
	case "yaml":
		return outputNvimPluginsYAML(plugins)
	case "json":
		return outputNvimPluginsJSON(plugins)
	default:
		return outputNvimPluginsTable(plugins)
	}
}

func getPlugin(cmd *cobra.Command, name string) error {
	// Use nvimops.Manager for unified storage with nvp CLI
	mgr, err := getNvimManager(cmd)
	if err != nil {
		return err
	}
	defer mgr.Close()

	p, err := mgr.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get plugin '%s': %w", name, err)
	}

	switch getOutputFormat {
	case "yaml":
		return outputNvimPluginYAML(p)
	case "json":
		return outputNvimPluginJSON(p)
	default:
		return outputNvimPluginsTable([]*plugin.Plugin{p})
	}
}

// Output functions

func outputProjectsTable(projects []*models.Project, activeProject string) error {
	// Create table renderer with colored columns
	tr := ui.ProjectsTableRenderer(activeProject)

	for _, p := range projects {
		// Format name with active indicator and icon
		name := p.Name
		if activeProject != "" && p.Name == activeProject {
			name = ui.FormatActiveItem(name, true)
		}

		tr.AddRow(
			name,
			ui.PathStyle.Render(p.Path),
			ui.DateStyle.Render(p.CreatedAt.Format("2006-01-02 15:04")),
		)
	}

	fmt.Println(tr.RenderSimple())
	return nil
}

func outputProjectsYAML(projects []*models.Project) error {
	var yamlOutput string

	for i, p := range projects {
		projectYAML := p.ToYAML()
		data, err := yaml.Marshal(&projectYAML)
		if err != nil {
			return fmt.Errorf("failed to encode project %s: %w", p.Name, err)
		}

		yamlOutput += string(data)

		// Add document separator between projects (but not after the last one)
		if i < len(projects)-1 {
			yamlOutput += "---\n"
		}
	}

	// Colorize and print the entire YAML output
	fmt.Print(ui.ColorizeYAML(yamlOutput))
	return nil
}

func outputProjectYAML(project *models.Project) error {
	projectYAML := project.ToYAML()
	data, err := yaml.Marshal(&projectYAML)
	if err != nil {
		return fmt.Errorf("failed to encode project: %w", err)
	}

	// Colorize and print the YAML output
	fmt.Print(ui.ColorizeYAML(string(data)))
	return nil
}

func outputWorkspacesTable(workspaces []*models.Workspace, projectName string, activeWorkspace string) error {
	// Create table renderer with colored columns
	tr := ui.WorkspacesTableRenderer(activeWorkspace)

	for _, ws := range workspaces {
		// Format name with active indicator
		name := ws.Name
		if activeWorkspace != "" && ws.Name == activeWorkspace {
			name = ui.FormatActiveItem(name, true)
		}

		// Style status
		status := ui.RenderStatus(ws.Status)

		tr.AddRow(
			name,
			ui.TextStyle.Render(projectName),
			ui.InfoStyle.Render(ws.ImageName),
			status,
			ui.DateStyle.Render(ws.CreatedAt.Format("2006-01-02 15:04")),
		)
	}

	fmt.Println(tr.RenderSimple())
	return nil
}

func outputWorkspacesYAML(workspaces []*models.Workspace, projectName string) error {
	var yamlOutput string

	for i, ws := range workspaces {
		workspaceYAML := ws.ToYAML(projectName)
		data, err := yaml.Marshal(&workspaceYAML)
		if err != nil {
			return fmt.Errorf("failed to encode workspace %s: %w", ws.Name, err)
		}

		yamlOutput += string(data)

		// Add document separator between workspaces (but not after the last one)
		if i < len(workspaces)-1 {
			yamlOutput += "---\n"
		}
	}

	// Colorize and print the entire YAML output
	fmt.Print(ui.ColorizeYAML(yamlOutput))
	return nil
}

func outputWorkspaceYAML(workspace *models.Workspace, projectName string) error {
	workspaceYAML := workspace.ToYAML(projectName)
	data, err := yaml.Marshal(&workspaceYAML)
	if err != nil {
		return fmt.Errorf("failed to encode workspace: %w", err)
	}

	// Colorize and print the YAML output
	fmt.Print(ui.ColorizeYAML(string(data)))
	return nil
}

// JSON output functions

func outputProjectsJSON(projects []*models.Project) error {
	projectsYAML := make([]models.ProjectYAML, len(projects))
	for i, p := range projects {
		projectsYAML[i] = p.ToYAML()
	}

	data, err := json.MarshalIndent(projectsYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputProjectJSON(project *models.Project) error {
	projectYAML := project.ToYAML()
	data, err := json.MarshalIndent(projectYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputWorkspacesJSON(workspaces []*models.Workspace, projectName string) error {
	workspacesYAML := make([]models.WorkspaceYAML, len(workspaces))
	for i, ws := range workspaces {
		workspacesYAML[i] = ws.ToYAML(projectName)
	}

	data, err := json.MarshalIndent(workspacesYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputWorkspaceJSON(workspace *models.Workspace, projectName string) error {
	workspaceYAML := workspace.ToYAML(projectName)
	data, err := json.MarshalIndent(workspaceYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Plugin output functions

func outputPluginsTable(plugins []*models.NvimPluginDB) error {
	// Create table renderer with colored columns
	tr := ui.PluginsTableRenderer([]string{})

	for _, p := range plugins {
		category := ""
		if p.Category.Valid {
			category = p.Category.String
		}

		// Get version or default
		version := "latest"
		if p.Version.Valid && p.Version.String != "" {
			version = p.Version.String
		} else if p.Branch.Valid && p.Branch.String != "" {
			version = "branch:" + p.Branch.String
		}

		// Format enabled status
		enabledStr := ui.CheckMark
		if !p.Enabled {
			enabledStr = ui.CrossMark
		}

		tr.AddRow(
			p.Name+" "+ui.MutedStyle.Render(enabledStr),
			ui.CategoryStyle.Render(category),
			ui.PathStyle.Render(p.Repo),
			ui.VersionStyle.Render(version),
		)
	}

	fmt.Println(tr.RenderSimple())
	fmt.Println()
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("Total: %d plugins", len(plugins))))
	return nil
}

func outputPluginsYAML(plugins []*models.NvimPluginDB) error {
	var yamlOutput string

	for i, p := range plugins {
		pluginYAML, err := p.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to convert plugin %s to YAML: %w", p.Name, err)
		}

		data, err := yaml.Marshal(&pluginYAML)
		if err != nil {
			return fmt.Errorf("failed to encode plugin %s: %w", p.Name, err)
		}

		yamlOutput += string(data)

		// Add document separator between plugins (but not after the last one)
		if i < len(plugins)-1 {
			yamlOutput += "---\n"
		}
	}

	// Colorize and print the entire YAML output
	fmt.Print(ui.ColorizeYAML(yamlOutput))
	return nil
}

func outputPluginYAML(plugin *models.NvimPluginDB) error {
	pluginYAML, err := plugin.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to convert plugin to YAML: %w", err)
	}

	data, err := yaml.Marshal(&pluginYAML)
	if err != nil {
		return fmt.Errorf("failed to encode plugin: %w", err)
	}

	// Colorize and print the YAML output
	fmt.Print(ui.ColorizeYAML(string(data)))
	return nil
}

func outputPluginsJSON(plugins []*models.NvimPluginDB) error {
	pluginsYAML := make([]models.NvimPluginYAML, 0, len(plugins))
	for _, p := range plugins {
		pluginYAML, err := p.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to convert plugin %s: %w", p.Name, err)
		}
		pluginsYAML = append(pluginsYAML, pluginYAML)
	}

	data, err := json.MarshalIndent(pluginsYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputPluginJSON(plugin *models.NvimPluginDB) error {
	pluginYAML, err := plugin.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to convert plugin to YAML: %w", err)
	}

	data, err := json.MarshalIndent(pluginYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Output functions for plugin.Plugin type (used with nvimops.Manager)

func outputNvimPluginsTable(plugins []*plugin.Plugin) error {
	// Create table renderer with colored columns
	tr := ui.PluginsTableRenderer([]string{})

	for _, p := range plugins {
		// Get version or default
		version := "latest"
		if p.Version != "" {
			version = p.Version
		} else if p.Branch != "" {
			version = "branch:" + p.Branch
		}

		// Format enabled status
		enabledStr := ui.CheckMark
		if !p.Enabled {
			enabledStr = ui.CrossMark
		}

		tr.AddRow(
			p.Name+" "+ui.MutedStyle.Render(enabledStr),
			ui.CategoryStyle.Render(p.Category),
			ui.PathStyle.Render(p.Repo),
			ui.VersionStyle.Render(version),
		)
	}

	fmt.Println(tr.RenderSimple())
	fmt.Println()
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("Total: %d plugins", len(plugins))))
	return nil
}

func outputNvimPluginsYAML(plugins []*plugin.Plugin) error {
	var yamlOutput string

	for i, p := range plugins {
		pluginYAML := p.ToYAML()
		data, err := yaml.Marshal(&pluginYAML)
		if err != nil {
			return fmt.Errorf("failed to encode plugin %s: %w", p.Name, err)
		}

		yamlOutput += string(data)

		// Add document separator between plugins (but not after the last one)
		if i < len(plugins)-1 {
			yamlOutput += "---\n"
		}
	}

	// Colorize and print the entire YAML output
	fmt.Print(ui.ColorizeYAML(yamlOutput))
	return nil
}

func outputNvimPluginYAML(p *plugin.Plugin) error {
	pluginYAML := p.ToYAML()
	data, err := yaml.Marshal(&pluginYAML)
	if err != nil {
		return fmt.Errorf("failed to encode plugin: %w", err)
	}

	// Colorize and print the YAML output
	fmt.Print(ui.ColorizeYAML(string(data)))
	return nil
}

func outputNvimPluginsJSON(plugins []*plugin.Plugin) error {
	pluginsYAML := make([]*plugin.PluginYAML, 0, len(plugins))
	for _, p := range plugins {
		pluginsYAML = append(pluginsYAML, p.ToYAML())
	}

	data, err := json.MarshalIndent(pluginsYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputNvimPluginJSON(p *plugin.Plugin) error {
	pluginYAML := p.ToYAML()

	data, err := json.MarshalIndent(pluginYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Platform output functions

func getContext(cmd *cobra.Command) error {
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	ctx, err := ctxMgr.LoadContext()
	if err != nil {
		return fmt.Errorf("failed to load context: %w", err)
	}

	// Use the decoupled formatter based on output format
	formatter := output.ForOutput(getOutputFormat)

	// Prepare context data for structured output
	contextData := ContextOutput{
		CurrentProject:   ctx.CurrentProject,
		CurrentWorkspace: ctx.CurrentWorkspace,
	}

	return outputContext(formatter, contextData)
}

// ContextOutput represents context for output formatting
type ContextOutput struct {
	CurrentProject   string `yaml:"currentProject" json:"currentProject"`
	CurrentWorkspace string `yaml:"currentWorkspace" json:"currentWorkspace"`
}

func outputContext(f output.Formatter, ctx ContextOutput) error {
	switch f.GetStyle() {
	case output.StyleJSON, output.StyleYAML:
		// Structured output - use the formatter's Object method
		return f.Object(ctx)
	default:
		// Table/colored output - use KeyValue for nice display
		if ctx.CurrentProject == "" {
			f.Info("No active context")
			f.NewLine()
			f.Println("Set context with:")
			f.List([]string{
				"dvm use project <name>",
				"dvm use workspace <name>",
			})
			return nil
		}

		f.Section("Current Context")

		workspace := ctx.CurrentWorkspace
		if workspace == "" {
			workspace = "(none)"
		}

		f.KeyValue(map[string]string{
			"Project":   ctx.CurrentProject,
			"Workspace": workspace,
		})

		return nil
	}
}

func getPlatforms(cmd *cobra.Command) error {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return fmt.Errorf("failed to create platform detector: %w", err)
	}

	platforms := detector.DetectAll()

	if len(platforms) == 0 {
		fmt.Println(ui.MutedStyle.Render("No container platforms detected"))
		fmt.Println(ui.MutedStyle.Render("Install OrbStack, Colima, Docker Desktop, or Podman"))
		return nil
	}

	// Get active platform
	activePlatform, _ := detector.Detect()
	var activeName string
	if activePlatform != nil {
		activeName = string(activePlatform.Type)
	}

	switch getOutputFormat {
	case "yaml":
		return outputPlatformsYAML(platforms, activeName)
	case "json":
		return outputPlatformsJSON(platforms, activeName)
	default:
		return outputPlatformsTable(platforms, activeName)
	}
}

// PlatformYAML represents a platform in YAML/JSON format
type PlatformYAML struct {
	Type         string `yaml:"type" json:"type"`
	Name         string `yaml:"name" json:"name"`
	SocketPath   string `yaml:"socketPath" json:"socketPath"`
	Profile      string `yaml:"profile,omitempty" json:"profile,omitempty"`
	IsContainerd bool   `yaml:"isContainerd" json:"isContainerd"`
	IsDocker     bool   `yaml:"isDockerCompatible" json:"isDockerCompatible"`
	Active       bool   `yaml:"active" json:"active"`
}

func outputPlatformsTable(platforms []*operators.Platform, activeName string) error {
	// Table header
	fmt.Printf("%-12s %-30s %-50s %-12s\n",
		"TYPE", "NAME", "SOCKET", "STATUS")
	fmt.Printf("%-12s %-30s %-50s %-12s\n",
		"----", "----", "------", "------")

	for _, p := range platforms {
		isActive := string(p.Type) == activeName
		status := ""
		if isActive {
			status = ui.SuccessStyle.Render("* active")
		}

		socketDisplay := p.SocketPath
		if len(socketDisplay) > 48 {
			socketDisplay = "..." + socketDisplay[len(socketDisplay)-45:]
		}

		name := p.Name
		if p.IsContainerd() {
			name += " (containerd)"
		} else if p.IsDockerCompatible() {
			name += " (docker)"
		}

		fmt.Printf("%-12s %-30s %-50s %-12s\n",
			string(p.Type),
			name,
			socketDisplay,
			status)
	}

	fmt.Println()
	fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("Total: %d platforms detected", len(platforms))))
	fmt.Println(ui.MutedStyle.Render("Use DVM_PLATFORM=<type> to select a specific platform"))
	return nil
}

func outputPlatformsYAML(platforms []*operators.Platform, activeName string) error {
	var yamlOutput string

	for i, p := range platforms {
		platformYAML := PlatformYAML{
			Type:         string(p.Type),
			Name:         p.Name,
			SocketPath:   p.SocketPath,
			Profile:      p.Profile,
			IsContainerd: p.IsContainerd(),
			IsDocker:     p.IsDockerCompatible(),
			Active:       string(p.Type) == activeName,
		}

		data, err := yaml.Marshal(&platformYAML)
		if err != nil {
			return fmt.Errorf("failed to encode platform %s: %w", p.Name, err)
		}

		yamlOutput += string(data)

		if i < len(platforms)-1 {
			yamlOutput += "---\n"
		}
	}

	fmt.Print(ui.ColorizeYAML(yamlOutput))
	return nil
}

func outputPlatformsJSON(platforms []*operators.Platform, activeName string) error {
	platformsYAML := make([]PlatformYAML, 0, len(platforms))
	for _, p := range platforms {
		platformsYAML = append(platformsYAML, PlatformYAML{
			Type:         string(p.Type),
			Name:         p.Name,
			SocketPath:   p.SocketPath,
			Profile:      p.Profile,
			IsContainerd: p.IsContainerd(),
			IsDocker:     p.IsDockerCompatible(),
			Active:       string(p.Type) == activeName,
		})
	}

	data, err := json.MarshalIndent(platformsYAML, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
