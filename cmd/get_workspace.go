package cmd

import (
	"fmt"

	"devopsmaestro/models"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

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

		// For JSON/YAML, wrap in kind: List envelope for round-trip compatibility (issue #154)
		if getOutputFormat == "json" || getOutputFormat == "yaml" {
			handlers.RegisterAll()
			if len(workspaces) == 0 {
				return render.OutputWith(getOutputFormat, resource.NewResourceList(), render.Options{Type: render.TypeAuto})
			}
			wsResources := make([]resource.Resource, len(workspaces))
			for i, ws := range workspaces {
				app, _ := sqlDS.GetAppByID(ws.AppID)
				appName := ""
				domName := ""
				ecoName := ""
				if app != nil {
					appName = app.Name
					if app.DomainID.Valid {
						dom, _ := sqlDS.GetDomainByID(int(app.DomainID.Int64))
						if dom != nil {
							domName = dom.Name
							if dom.EcosystemID.Valid {
								eco, _ := sqlDS.GetEcosystemByID(int(dom.EcosystemID.Int64))
								if eco != nil {
									ecoName = eco.Name
								}
							}
						}
					}
				}
				gitRepoName := ""
				if ws.GitRepoID.Valid {
					gitRepo, gitErr := sqlDS.GetGitRepoByID(ws.GitRepoID.Int64)
					if gitErr == nil && gitRepo != nil {
						gitRepoName = gitRepo.Name
					}
				}
				wsResources[i] = handlers.NewWorkspaceResource(ws, appName, domName, gitRepoName, ecoName)
			}
			resCtx := resource.Context{DataStore: sqlDS}
			list, listErr := resource.BuildList(resCtx, wsResources)
			if listErr != nil {
				return fmt.Errorf("failed to build resource list: %w", listErr)
			}
			return render.OutputWith(getOutputFormat, list, render.Options{Type: render.TypeAuto})
		}

		if len(workspaces) == 0 {
			return render.OutputWith(getOutputFormat, nil, render.Options{
				Empty:        true,
				EmptyMessage: "No workspaces found",
				EmptyHints:   []string{"dvm create workspace <name>"},
			})
		}

		// Determine if wide format
		isWide := getOutputFormat == "wide"

		// For human output, build table data
		// We need to look up app names for display
		var headers []string
		if isWide {
			headers = []string{"NAME", "APP", "SYSTEM", "IMAGE", "STATUS", "CREATED", "CONTAINER-ID"}
		} else {
			headers = []string{"NAME", "APP", "SYSTEM", "IMAGE", "STATUS"}
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
			sysName := ""
			if app != nil {
				appName = app.Name
				if app.SystemID.Valid {
					if sys, sErr := sqlDS.GetSystemByID(int(app.SystemID.Int64)); sErr == nil && sys != nil {
						sysName = sys.Name
					}
				}
			}

			row := []string{
				ws.Name,
				appName,
				sysName,
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

		// For JSON/YAML, wrap in kind: List envelope for round-trip compatibility (issue #154)
		if getOutputFormat == "json" || getOutputFormat == "yaml" {
			handlers.RegisterAll()
			wsResources := make([]resource.Resource, len(results))
			for i, wh := range results {
				gitRepoName := ""
				if wh.Workspace.GitRepoID.Valid {
					gitRepo, gitErr := sqlDS.GetGitRepoByID(wh.Workspace.GitRepoID.Int64)
					if gitErr == nil && gitRepo != nil {
						gitRepoName = gitRepo.Name
					}
				}
				domName := ""
				if wh.Domain != nil {
					domName = wh.Domain.Name
				}
				ecoName := ""
				if wh.Ecosystem != nil {
					ecoName = wh.Ecosystem.Name
				}
				wsResources[i] = handlers.NewWorkspaceResource(wh.Workspace, wh.App.Name, domName, gitRepoName, ecoName)
			}
			resCtx := resource.Context{DataStore: sqlDS}
			list, listErr := resource.BuildList(resCtx, wsResources)
			if listErr != nil {
				return fmt.Errorf("failed to build resource list: %w", listErr)
			}
			return render.OutputWith(getOutputFormat, list, render.Options{Type: render.TypeAuto})
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

	// For JSON/YAML, wrap in kind: List envelope for round-trip compatibility (issue #154)
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		handlers.RegisterAll()
		if len(workspaces) == 0 {
			return render.OutputWith(getOutputFormat, resource.NewResourceList(), render.Options{Type: render.TypeAuto})
		}
		// Resolve domain/ecosystem names for context-free output
		domName := ""
		ecoName := ""
		if app.DomainID.Valid {
			dom, _ := sqlDS.GetDomainByID(int(app.DomainID.Int64))
			if dom != nil {
				domName = dom.Name
				if dom.EcosystemID.Valid {
					eco, _ := sqlDS.GetEcosystemByID(int(dom.EcosystemID.Int64))
					if eco != nil {
						ecoName = eco.Name
					}
				}
			}
		}
		wsResources := make([]resource.Resource, len(workspaces))
		for i, ws := range workspaces {
			gitRepoName := ""
			if ws.GitRepoID.Valid {
				gitRepo, gitErr := sqlDS.GetGitRepoByID(ws.GitRepoID.Int64)
				if gitErr == nil && gitRepo != nil {
					gitRepoName = gitRepo.Name
				}
			}
			wsResources[i] = handlers.NewWorkspaceResource(ws, appName, domName, gitRepoName, ecoName)
		}
		resCtx := resource.Context{DataStore: sqlDS}
		list, listErr := resource.BuildList(resCtx, wsResources)
		if listErr != nil {
			return fmt.Errorf("failed to build resource list: %w", listErr)
		}
		return render.OutputWith(getOutputFormat, list, render.Options{Type: render.TypeAuto})
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
				render.Plain(FormatSuggestions(SuggestAmbiguousWorkspace()...))
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Plain(FormatSuggestions(SuggestWorkspaceNotFound(filter.WorkspaceName)...))
				return err
			}
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}

		workspace = result.Workspace
		app = result.App
		appName = app.Name

	} else {
		// Fall back to existing context-based behavior (DB-backed)
		var err error
		appName, err = getActiveAppFromContext(sqlDS)
		if err != nil {
			return ErrorWithSuggestion(
				"no app specified",
				SuggestNoActiveApp()...,
			)
		}

		// Get app to get its ID (search globally across all domains)
		app, err = sqlDS.GetAppByNameGlobal(appName)
		if err != nil {
			return ErrorWithSuggestion(
				fmt.Sprintf("app %q not found", appName),
				SuggestAppNotFound(appName)...,
			)
		}

		// Need workspace name when using context-based lookup
		return ErrorWithSuggestion(
			"workspace name required",
			"Specify a name: dvm get workspace <name>",
			"List all workspaces: dvm get workspaces",
		)
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

	// Walk hierarchy: app -> system -> domain -> ecosystem
	systemName := ""
	domainName := ""
	ecosystemName := ""
	if app != nil {
		if app.SystemID.Valid {
			if sys, sErr := sqlDS.GetSystemByID(int(app.SystemID.Int64)); sErr == nil && sys != nil {
				systemName = sys.Name
			}
		}
		if app.DomainID.Valid {
			if dom, dErr := sqlDS.GetDomainByID(int(app.DomainID.Int64)); dErr == nil && dom != nil {
				domainName = dom.Name
				if dom.EcosystemID.Valid {
					if eco, eErr := sqlDS.GetEcosystemByID(int(dom.EcosystemID.Int64)); eErr == nil && eco != nil {
						ecosystemName = eco.Name
					}
				}
			}
		}
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: nameDisplay},
		render.KeyValue{Key: "App", Value: appName},
		render.KeyValue{Key: "System", Value: systemName},
		render.KeyValue{Key: "Domain", Value: domainName},
		render.KeyValue{Key: "Ecosystem", Value: ecosystemName},
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
