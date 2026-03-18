package cmd

import (
	"fmt"

	"devopsmaestro/models"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/resolver"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

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

				// Resolve GitRepo name if GitRepoID is set
				gitRepoName := ""
				if ws.GitRepoID.Valid {
					gitRepo, err := sqlDS.GetGitRepoByID(ws.GitRepoID.Int64)
					if err == nil && gitRepo != nil {
						gitRepoName = gitRepo.Name
					}
				}

				workspacesYAML[i] = ws.ToYAML(appName, gitRepoName)
			}
			return render.OutputWith(getOutputFormat, workspacesYAML, render.Options{})
		}

		// Determine if wide format
		isWide := getOutputFormat == "wide"

		// For human output, build table data
		// We need to look up app names for display
		var headers []string
		if isWide {
			headers = []string{"NAME", "APP", "IMAGE", "STATUS", "CREATED", "CONTAINER-ID"}
		} else {
			headers = []string{"NAME", "APP", "IMAGE", "STATUS"}
		}
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

			if isWide {
				// Add CREATED timestamp
				row = append(row, ws.CreatedAt.Format("2006-01-02 15:04"))
				// Add CONTAINER-ID (truncated to 12 chars like Docker)
				containerID := "<none>"
				if ws.ContainerID.Valid && ws.ContainerID.String != "" {
					containerID = ws.ContainerID.String
					if len(containerID) > 12 {
						containerID = containerID[:12]
					}
				}
				row = append(row, containerID)
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

		// For rendering, treat "wide" as table format
		renderFormat := getOutputFormat
		if isWide {
			renderFormat = "table"
		}

		return render.OutputWith(renderFormat, tableData, render.Options{
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
				// Resolve GitRepo name if GitRepoID is set
				gitRepoName := ""
				if wh.Workspace.GitRepoID.Valid {
					gitRepo, err := sqlDS.GetGitRepoByID(wh.Workspace.GitRepoID.Int64)
					if err == nil && gitRepo != nil {
						gitRepoName = gitRepo.Name
					}
				}
				workspacesYAML[i] = wh.Workspace.ToYAML(wh.App.Name, gitRepoName)
			}
			return render.OutputWith(getOutputFormat, workspacesYAML, render.Options{})
		}

		// Determine if wide format
		isWide := getOutputFormat == "wide"

		// For human output, build table data with full path
		var headers []string
		if isWide {
			headers = []string{"NAME", "PATH", "IMAGE", "STATUS", "CREATED", "CONTAINER-ID"}
		} else {
			headers = []string{"NAME", "PATH", "IMAGE", "STATUS"}
		}
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

			if isWide {
				// Add CREATED timestamp
				row = append(row, wh.Workspace.CreatedAt.Format("2006-01-02 15:04"))
				// Add CONTAINER-ID (truncated to 12 chars like Docker)
				containerID := "<none>"
				if wh.Workspace.ContainerID.Valid && wh.Workspace.ContainerID.String != "" {
					containerID = wh.Workspace.ContainerID.String
					if len(containerID) > 12 {
						containerID = containerID[:12]
					}
				}
				row = append(row, containerID)
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

		// For rendering, treat "wide" as table format
		renderFormat := getOutputFormat
		if isWide {
			renderFormat = "table"
		}

		return render.OutputWith(renderFormat, tableData, render.Options{
			Type: render.TypeTable,
		})
	}

	// Fall back to existing context-based behavior (DB-backed)
	appName, err := getActiveAppFromContext(sqlDS)
	if err != nil {
		return fmt.Errorf("no app specified. Use -a <name> or 'dvm use app <name>' first")
	}

	// Get active workspace (only relevant if viewing active app - for marking with ●)
	activeWorkspace, _ := getActiveWorkspaceFromContext(sqlDS)

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
			// Resolve GitRepo name if GitRepoID is set
			gitRepoName := ""
			if ws.GitRepoID.Valid {
				gitRepo, err := sqlDS.GetGitRepoByID(ws.GitRepoID.Int64)
				if err == nil && gitRepo != nil {
					gitRepoName = gitRepo.Name
				}
			}
			workspacesYAML[i] = ws.ToYAML(appName, gitRepoName)
		}
		return render.OutputWith(getOutputFormat, workspacesYAML, render.Options{})
	}

	// Determine if wide format
	isWide := getOutputFormat == "wide"

	// For human output, build table data
	var headers []string
	if isWide {
		headers = []string{"NAME", "APP", "IMAGE", "STATUS", "CREATED", "CONTAINER-ID"}
	} else {
		headers = []string{"NAME", "APP", "IMAGE", "STATUS"}
	}
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
		}

		if isWide {
			// Add CREATED timestamp
			row = append(row, ws.CreatedAt.Format("2006-01-02 15:04"))
			// Add CONTAINER-ID (truncated to 12 chars like Docker)
			containerID := "<none>"
			if ws.ContainerID.Valid && ws.ContainerID.String != "" {
				containerID = ws.ContainerID.String
				if len(containerID) > 12 {
					containerID = containerID[:12]
				}
			}
			row = append(row, containerID)
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

	// For rendering, treat "wide" as table format
	renderFormat := getOutputFormat
	if isWide {
		renderFormat = "table"
	}

	return render.OutputWith(renderFormat, tableData, render.Options{
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
				render.Plain(ambiguousErr.FormatDisambiguation())
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
		// Fall back to existing context-based behavior (DB-backed)
		var err error
		appName, err = getActiveAppFromContext(sqlDS)
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
		// Resolve GitRepo name if GitRepoID is set
		gitRepoName := ""
		if workspace.GitRepoID.Valid {
			gitRepo, err := sqlDS.GetGitRepoByID(workspace.GitRepoID.Int64)
			if err == nil && gitRepo != nil {
				gitRepoName = gitRepo.Name
			}
		}
		return render.OutputWith(getOutputFormat, workspace.ToYAML(appName, gitRepoName), render.Options{})
	}

	// For human output, show detail view
	// Check if this workspace is the active one (DB-backed)
	activeWorkspace, _ := getActiveWorkspaceFromContext(sqlDS)

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
