package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
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
	Short: "List all workspaces in current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspaces(cmd)
	},
}

// getWorkspaceCmd gets a specific workspace
var getWorkspaceCmd = &cobra.Command{
	Use:   "workspace [name]",
	Short: "Get a specific workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspace(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getProjectsCmd)
	getCmd.AddCommand(getProjectCmd)
	getCmd.AddCommand(getWorkspacesCmd)
	getCmd.AddCommand(getWorkspaceCmd)

	// Add output format flag to all get commands
	getCmd.PersistentFlags().StringVarP(&getOutputFormat, "output", "o", "table", "Output format (table, yaml, json)")
}

func getDataStore(cmd *cobra.Command) (*db.SQLDataStore, error) {
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return nil, fmt.Errorf("dataStore not initialized")
	}

	sqlDS, ok := (*dataStore).(*db.SQLDataStore)
	if !ok {
		return nil, fmt.Errorf("expected SQLDataStore")
	}

	return sqlDS, nil
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

	switch getOutputFormat {
	case "yaml":
		return outputProjectsYAML(projects)
	case "json":
		return outputProjectsJSON(projects)
	default:
		return outputProjectsTable(projects)
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

	switch getOutputFormat {
	case "yaml":
		return outputProjectYAML(project)
	case "json":
		return outputProjectJSON(project)
	default:
		return outputProjectsTable([]*models.Project{project})
	}
}

func getWorkspaces(cmd *cobra.Command) error {
	// Get current context
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	projectName, err := ctxMgr.GetActiveProject()
	if err != nil {
		return fmt.Errorf("no active project set. Use 'dvm use project <name>' first")
	}

	// For now, just show the project name and a message
	fmt.Printf("Project: %s\n", projectName)
	fmt.Println("Listing workspaces not yet fully implemented")
	fmt.Println("TODO: Need to add ListWorkspacesByProjectID method to SQLDataStore")

	return nil
}

func getWorkspace(cmd *cobra.Command, name string) error {
	// Get current context
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	projectName, err := ctxMgr.GetActiveProject()
	if err != nil {
		return fmt.Errorf("no active project set. Use 'dvm use project <name>' first")
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
		return outputWorkspacesTable([]*models.Workspace{workspace}, project.Name)
	}
}

// Output functions

func outputProjectsTable(projects []*models.Project) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tPATH\tCREATED")
	for _, p := range projects {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			p.Name,
			p.Path,
			p.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	return nil
}

func outputProjectsYAML(projects []*models.Project) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	defer encoder.Close()

	for i, p := range projects {
		projectYAML := p.ToYAML()
		if err := encoder.Encode(&projectYAML); err != nil {
			return fmt.Errorf("failed to encode project %s: %w", p.Name, err)
		}
		// Add document separator between projects (but not after the last one)
		if i < len(projects)-1 {
			fmt.Println("---")
		}
	}
	return nil
}

func outputProjectYAML(project *models.Project) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	defer encoder.Close()

	projectYAML := project.ToYAML()
	return encoder.Encode(&projectYAML)
}

func outputWorkspacesTable(workspaces []*models.Workspace, projectName string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tPROJECT\tIMAGE\tSTATUS\tCREATED")
	for _, ws := range workspaces {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ws.Name,
			projectName,
			ws.ImageName,
			ws.Status,
			ws.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	return nil
}

func outputWorkspacesYAML(workspaces []*models.Workspace, projectName string) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	defer encoder.Close()

	for i, ws := range workspaces {
		workspaceYAML := ws.ToYAML(projectName)
		if err := encoder.Encode(&workspaceYAML); err != nil {
			return fmt.Errorf("failed to encode workspace %s: %w", ws.Name, err)
		}
		// Add document separator between workspaces (but not after the last one)
		if i < len(workspaces)-1 {
			fmt.Println("---")
		}
	}
	return nil
}

func outputWorkspaceYAML(workspace *models.Workspace, projectName string) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	defer encoder.Close()

	workspaceYAML := workspace.ToYAML(projectName)
	return encoder.Encode(&workspaceYAML)
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
