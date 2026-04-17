package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/mirror"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/utils"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

var (
	appDescription string
	appDomain      string
	appSystem      string
	appPath        string
	appFromCwd     bool
	appRepo        string
)

// Dry-run flags for app commands
var (
	createAppDryRun bool
	deleteAppDryRun bool
)

// createAppCmd creates a new app
var createAppCmd = &cobra.Command{
	Use:     "app <name>",
	Aliases: []string{"application", "a"},
	Short:   "Create a new app",
	Long: `Create a new app within a domain.

An app represents a codebase/application: Ecosystem -> Domain -> App -> Workspace.

Source Location (choose one):
  --from-cwd          Use current directory as source
  --path <path>       Use local filesystem path
  --repo <url|name>   Use git repository (URL or GitRepo name)

Examples:
  # Create an app from the current directory
  dvm create app my-api --from-cwd
  
  # Create an app with a specific path
  dvm create app my-api --path ~/code/my-api
  
  # Create an app in a specific domain
  dvm create app my-api --from-cwd --domain backend
  
  # Create an app in a specific system
  dvm create app my-api --from-cwd --system auth-system
  
  # Create from git URL (auto-creates GitRepo)
  dvm create app golang-app --repo https://github.com/rmkohlman/dvm-test-golang.git
  
  # Create using existing GitRepo
  dvm create app golang-app --repo my-gitrepo
  
  # Create with description
  dvm create app my-api --from-cwd --description "REST API service"

Next Steps:
  1. Create a workspace for this app:
     dvm create workspace main
  2. Build and attach:
     dvm build && dvm attach`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		// Validate name is not empty
		if err := ValidateResourceName(appName, "app"); err != nil {
			return err
		}

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		// Count how many source flags are set
		flagsSet := 0
		if appFromCwd {
			flagsSet++
		}
		if appPath != "" {
			flagsSet++
		}
		if appRepo != "" {
			flagsSet++
		}

		// Validate mutually exclusive flags
		if flagsSet == 0 {
			return fmt.Errorf("must specify one of: --from-cwd, --path, or --repo")
		}
		if flagsSet > 1 {
			return fmt.Errorf("flags --from-cwd, --path, and --repo are mutually exclusive")
		}

		// Variables to track GitRepo if using --repo
		var gitRepoID *int
		var gitRepoName string

		// Determine app path
		var path string
		if appFromCwd {
			path, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			// Verify path exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", path)
			}
		} else if appPath != "" {
			path, err = filepath.Abs(appPath)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}
			// Verify path exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", path)
			}
		} else if appRepo != "" {
			// Handle --repo flag: either URL or existing GitRepo name
			gitRepoID, gitRepoName, path, err = resolveOrCreateGitRepo(ds, appRepo)
			if err != nil {
				return err
			}
		}

		// Get domain - from flag or active context
		var domain *models.Domain
		if appDomain != "" {
			// Need active ecosystem to find the domain
			ecosystem, err := getActiveEcosystem(ds)
			if err != nil {
				render.Error("No active ecosystem set")
				render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
				return errSilent
			}
			domain, err = ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, appDomain)
			if err != nil {
				return fmt.Errorf("domain '%s' not found in ecosystem '%s': %w", appDomain, ecosystem.Name, err)
			}
		} else {
			domain, err = getActiveDomain(ds)
			if err != nil {
				render.Error("No domain specified")
				render.Info("Hint: Use --domain <name> or 'dvm use domain <name>' to select a domain first")
				return errSilent
			}
		}

		// Resolve system from flag or active context (optional)
		var system *models.System
		if appSystem != "" {
			// Global system lookup: find system across all domains (#287)
			resolution, sErr := resolveSystemGlobally(ds, appSystem, domain)
			if sErr != nil {
				return sErr
			}
			system = resolution.System
			// Use the system's domain if it differs from active context
			if resolution.Domain.ID != domain.ID {
				render.Info(fmt.Sprintf("System '%s' is in domain '%s' (active domain: '%s') — using system's domain",
					appSystem, resolution.Domain.Name, domain.Name))
				domain = resolution.Domain
			}
		} else {
			// Try active system context (optional — apps don't require a system)
			system, _ = getActiveSystem(ds)
		}

		// Get ecosystem name for display
		ecosystemName := ""
		if domain.EcosystemID.Valid {
			ecosystem, _ := ds.GetEcosystemByID(int(domain.EcosystemID.Int64))
			if ecosystem != nil {
				ecosystemName = ecosystem.Name
			}
		}

		render.Progress(fmt.Sprintf("Creating app '%s' in domain '%s'...", appName, domain.Name))

		// Dry-run: preview what would be created
		if createAppDryRun {
			render.Plain(fmt.Sprintf("Would create app %q in domain %q (ecosystem %q)", appName, domain.Name, ecosystemName))
			render.Plain(fmt.Sprintf("  path: %s", path))
			if gitRepoName != "" {
				render.Plain(fmt.Sprintf("  gitrepo: %s", gitRepoName))
			}
			return nil
		}

		// Check if app already exists
		existing, _ := ds.GetAppByName(sql.NullInt64{Int64: int64(domain.ID), Valid: true}, appName)
		if existing != nil {
			return fmt.Errorf("app '%s' already exists in domain '%s'", appName, domain.Name)
		}

		// Create app using handler helper
		app := handlers.NewAppFromModel(appName, domain.ID, path, appDescription)

		// Link System if resolved
		if system != nil {
			app.SystemID = sql.NullInt64{Int64: int64(system.ID), Valid: true}
		}

		// Link GitRepo if using --repo
		if gitRepoID != nil {
			app.GitRepoID = sql.NullInt64{Int64: int64(*gitRepoID), Valid: true}
		}

		if err := ds.CreateApp(app); err != nil {
			return fmt.Errorf("failed to create app: %w", err)
		}

		// Get the created app to get its ID
		createdApp, err := ds.GetAppByName(sql.NullInt64{Int64: int64(domain.ID), Valid: true}, appName)
		if err != nil {
			return fmt.Errorf("failed to retrieve created app: %w", err)
		}

		render.Success(fmt.Sprintf("App '%s' created successfully (ID: %d)", appName, createdApp.ID))
		render.Info(fmt.Sprintf("Ecosystem: %s", ecosystemName))
		render.Info(fmt.Sprintf("Domain: %s", domain.Name))
		if system != nil {
			render.Info(fmt.Sprintf("System: %s", system.Name))
		}
		if gitRepoName != "" {
			render.Info(fmt.Sprintf("GitRepo: %s", gitRepoName))
		} else {
			render.Info(fmt.Sprintf("Path: %s", path))
		}

		// Set app as active context
		if err := ds.SetActiveApp(&createdApp.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to set active app: %v", err))
		} else {
			render.Success(fmt.Sprintf("Set '%s' as active app", appName))
		}

		render.Blank()
		render.Info("Next steps:")
		render.Info("  1. Create a workspace for this app:")
		if gitRepoName != "" {
			render.Info(fmt.Sprintf("     dvm create workspace main --repo %s", gitRepoName))
		} else {
			render.Info("     dvm create workspace main")
		}
		render.Info("  2. Build and attach:")
		render.Info("     dvm build && dvm attach")
		return nil
	},
}

// getAppsCmd lists all apps
var getAppsCmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"application", "applications", "a"},
	Short:   "List all apps",
	Long: `List all apps, optionally filtered by domain or system.

Examples:
  dvm get apps                          # List apps in active domain
  dvm get apps --domain backend
  dvm get apps --system auth-system     # Filter by system
  dvm get apps -A                       # List all apps across all domains
  dvm get apps --all                    # Same as -A
  dvm get apps -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getApps(cmd)
	},
}

// getAppCmd gets a specific app
var getAppCmd = &cobra.Command{
	Use:     "app <name>",
	Aliases: []string{"application", "a"},
	Short:   "Get a specific app",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getApp(cmd, args[0])
	},
}

// NOTE: useAppCmd is defined in use.go to maintain consistency with the file-based
// ContextManager approach used by the other use commands.

func getApps(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	domainFlag, _ := cmd.Flags().GetString("domain")
	systemFlag, _ := cmd.Flags().GetString("system")

	var apps []*models.App
	var domainName string

	if allFlag {
		// List all apps across all domains
		apps, err = ds.ListAllApps()
		if err != nil {
			return fmt.Errorf("failed to list apps: %w", err)
		}
		domainName = "(all)"
	} else {
		// Get domain from flag or active context
		var domain *models.Domain
		if domainFlag != "" {
			// Need active ecosystem to find the domain
			ecosystem, err := getActiveEcosystem(ds)
			if err != nil {
				render.Error("No active ecosystem set")
				render.Info("Hint: Use --all, or set active ecosystem first with: dvm use ecosystem <name>")
				return errSilent
			}
			domain, err = ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, domainFlag)
			if err != nil {
				return fmt.Errorf("domain '%s' not found: %w", domainFlag, err)
			}
		} else {
			domain, err = getActiveDomain(ds)
			if err != nil {
				render.Error("No domain specified")
				render.Info("Hint: Use --domain <name>, --all, or 'dvm use domain <name>' first")
				return errSilent
			}
		}

		domainName = domain.Name
		apps, err = ds.ListAppsByDomain(domain.ID)
		if err != nil {
			return fmt.Errorf("failed to list apps: %w", err)
		}
	}

	// Filter by system if --system flag is provided
	if systemFlag != "" {
		// Global system lookup: find system across all domains (#287)
		var activeDomain *models.Domain
		if domainName != "" && domainName != "(all)" {
			ecosystem, ecoErr := getActiveEcosystem(ds)
			if ecoErr == nil {
				activeDomain, _ = ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, domainName)
			}
		}
		resolution, sErr := resolveSystemGlobally(ds, systemFlag, activeDomain)
		if sErr != nil {
			return sErr
		}
		targetSystem := resolution.System
		filtered := make([]*models.App, 0, len(apps))
		for _, a := range apps {
			if a.SystemID.Valid && a.SystemID.Int64 == int64(targetSystem.ID) {
				filtered = append(filtered, a)
			}
		}
		apps = filtered
	}

	// Get active app for highlighting
	ctx, _ := ds.GetContext()
	var activeAppID *int
	if ctx != nil {
		activeAppID = ctx.ActiveAppID
	}

	// For JSON/YAML, wrap in kind: List envelope for round-trip compatibility (issue #154)
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		handlers.RegisterAll()
		if len(apps) == 0 {
			return render.OutputWith(getOutputFormat, resource.NewResourceList(), render.Options{Type: render.TypeAuto})
		}
		// Convert app models to Resource objects for BuildList
		appResources := make([]resource.Resource, len(apps))
		for i, a := range apps {
			domName := ""
			ecoName := ""
			if a.DomainID.Valid {
				dom, _ := ds.GetDomainByID(int(a.DomainID.Int64))
				if dom != nil {
					domName = dom.Name
					if dom.EcosystemID.Valid {
						eco, _ := ds.GetEcosystemByID(int(dom.EcosystemID.Int64))
						if eco != nil {
							ecoName = eco.Name
						}
					}
				}
			}
			// Resolve git repo name if associated
			gitRepoName := ""
			if a.GitRepoID.Valid {
				if gr, grErr := ds.GetGitRepoByID(a.GitRepoID.Int64); grErr == nil && gr != nil {
					gitRepoName = gr.Name
				}
			}
			// Resolve system name if associated
			sysName := ""
			if a.SystemID.Valid {
				if sys, sErr := ds.GetSystemByID(int(a.SystemID.Int64)); sErr == nil && sys != nil {
					sysName = sys.Name
				}
			}
			appResources[i] = handlers.NewAppResource(a, domName, ecoName, gitRepoName, sysName)
		}
		resCtx := resource.Context{DataStore: ds}
		list, err := resource.BuildList(resCtx, appResources)
		if err != nil {
			return fmt.Errorf("failed to build resource list: %w", err)
		}
		return render.OutputWith(getOutputFormat, list, render.Options{Type: render.TypeAuto})
	}

	if len(apps) == 0 {
		msg := fmt.Sprintf("No apps found in domain '%s'", domainName)
		if allFlag {
			msg = "No apps found"
		}
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: msg,
			EmptyHints:   []string{"dvm create app <name> --path <path>"},
		})
	}

	// Determine if wide format
	isWide := getOutputFormat == "wide"

	// For human output, build table data
	var headers []string
	if isWide {
		headers = []string{"NAME", "DOMAIN", "SYSTEM", "PATH", "CREATED", "ID", "GITREPO"}
	} else {
		headers = []string{"NAME", "DOMAIN", "SYSTEM", "PATH", "CREATED"}
	}
	if showTheme {
		headers = append(headers, "THEME", "THEME SOURCE")
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    make([][]string, len(apps)),
	}

	// Create theme resolver if needed
	var themeResolver themeresolver.ThemeResolver
	if showTheme {
		themeResolver, _ = themeresolver.NewThemeResolver(ds, nil)
	}

	for i, a := range apps {
		name := a.Name
		if activeAppID != nil && a.ID == *activeAppID {
			name = "● " + name // Active indicator
		}

		// Get domain name for display
		domName := ""
		if a.DomainID.Valid {
			dom, _ := ds.GetDomainByID(int(a.DomainID.Int64))
			if dom != nil {
				domName = dom.Name
			}
		}

		// Get system name for display
		sysName := ""
		if a.SystemID.Valid {
			sys, _ := ds.GetSystemByID(int(a.SystemID.Int64))
			if sys != nil {
				sysName = sys.Name
			}
		}

		// Truncate path if too long
		pathDisplay := a.Path
		if len(pathDisplay) > 40 {
			pathDisplay = "..." + pathDisplay[len(pathDisplay)-37:]
		}

		row := []string{
			name,
			domName,
			sysName,
			pathDisplay,
			a.CreatedAt.Format("2006-01-02 15:04"),
		}

		if isWide {
			// Add ID
			row = append(row, fmt.Sprintf("%d", a.ID))
			// Add GITREPO - currently apps don't have a direct GitRepo association
			row = append(row, "<none>")
		}

		// Add theme information if requested
		if showTheme && themeResolver != nil {
			themeName := themeresolver.DefaultTheme
			themeSource := "default"

			if resolution, err := themeResolver.GetResolutionPath(cmd.Context(), themeresolver.LevelApp, a.ID); err == nil {
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

func getApp(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	domainFlag, _ := cmd.Flags().GetString("domain")
	systemFlag, _ := cmd.Flags().GetString("system")

	// Get domain from flag or active context
	var domain *models.Domain
	if domainFlag != "" {
		// Need active ecosystem to find the domain
		ecosystem, err := getActiveEcosystem(ds)
		if err != nil {
			render.Error("No active ecosystem set")
			render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
			return errSilent
		}
		domain, err = ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, domainFlag)
		if err != nil {
			return fmt.Errorf("domain '%s' not found: %w", domainFlag, err)
		}
		// Temporarily set active domain for handler
		ds.SetActiveDomain(&domain.ID)
	} else {
		domain, err = getActiveDomain(ds)
		if err != nil {
			render.Error("No domain specified")
			render.Info("Hint: Use --domain <name> or 'dvm use domain <name>' first")
			return errSilent
		}
	}

	// Resolve system from flag if provided (for context)
	if systemFlag != "" {
		// Global system lookup: find system across all domains (#287)
		resolution, sErr := resolveSystemGlobally(ds, systemFlag, domain)
		if sErr != nil {
			return sErr
		}
		ds.SetActiveSystem(&resolution.System.ID)
		// Use the system's domain if it differs from active context
		if resolution.Domain.ID != domain.ID {
			render.Info(fmt.Sprintf("System '%s' is in domain '%s' (active domain: '%s') — using system's domain",
				systemFlag, resolution.Domain.Name, domain.Name))
			domain = resolution.Domain
			ds.SetActiveDomain(&domain.ID)
		}
	}

	// Get ecosystem for display
	ecosystemName := ""
	if domain.EcosystemID.Valid {
		ecosystem, _ := ds.GetEcosystemByID(int(domain.EcosystemID.Int64))
		if ecosystem != nil {
			ecosystemName = ecosystem.Name
		}
	}

	res, err := resource.Get(ctx, handlers.KindApp, name)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	app := res.(*handlers.AppResource).App()

	// Resolve system name if associated
	systemName := ""
	if app.SystemID.Valid {
		if sys, sErr := ds.GetSystemByID(int(app.SystemID.Int64)); sErr == nil && sys != nil {
			systemName = sys.Name
		}
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		workspaces, _ := ds.ListWorkspacesByApp(app.ID)
		wsNames := make([]string, len(workspaces))
		for j, w := range workspaces {
			wsNames[j] = w.Name
		}
		// Resolve git repo name if associated
		gitRepoName := ""
		if app.GitRepoID.Valid {
			if gr, grErr := ds.GetGitRepoByID(app.GitRepoID.Int64); grErr == nil && gr != nil {
				gitRepoName = gr.Name
			}
		}
		yamlDoc := app.ToYAML(domain.Name, wsNames, gitRepoName, systemName)
		yamlDoc.Metadata.Ecosystem = ecosystemName
		return render.OutputWith(getOutputFormat, yamlDoc, render.Options{})
	}

	// For human output, show detail view
	dbCtx, _ := ds.GetContext()
	var activeAppID *int
	if dbCtx != nil {
		activeAppID = dbCtx.ActiveAppID
	}

	isActive := activeAppID != nil && app.ID == *activeAppID
	nameDisplay := app.Name
	if isActive {
		nameDisplay = "● " + nameDisplay + " (active)"
	}

	desc := ""
	if app.Description.Valid {
		desc = app.Description.String
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: nameDisplay},
		render.KeyValue{Key: "Ecosystem", Value: ecosystemName},
		render.KeyValue{Key: "Domain", Value: domain.Name},
		render.KeyValue{Key: "System", Value: systemName},
		render.KeyValue{Key: "Path", Value: app.Path},
		render.KeyValue{Key: "Description", Value: desc},
		render.KeyValue{Key: "Created", Value: app.CreatedAt.Format("2006-01-02 15:04:05")},
	)

	err = render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "App Details",
	})
	if err != nil {
		return err
	}

	// Show theme information if requested
	if showTheme {
		return showThemeResolution(cmd, ds, themeresolver.LevelApp, app.ID, app.Name)
	}

	return nil
}

// deleteAppCmd deletes an app
var deleteAppCmd = &cobra.Command{
	Use:     "app <name>",
	Aliases: []string{"application", "a"},
	Short:   "Delete an app",
	Long: `Delete an app by name.

WARNING: This will cascade-delete all workspaces within the app.
By default, you will be prompted for confirmation. Use --force to skip.

Examples:
  dvm delete app my-api
  dvm delete app my-api --domain backend
  dvm delete app my-api --system auth-system
  dvm delete app my-api --force              # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		domainFlag, _ := cmd.Flags().GetString("domain")
		systemFlag, _ := cmd.Flags().GetString("system")

		// Get domain from flag or active context
		var domain *models.Domain
		if domainFlag != "" {
			// Need active ecosystem to find the domain
			ecosystem, err := getActiveEcosystem(ds)
			if err != nil {
				render.Error("No active ecosystem set")
				render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
				return errSilent
			}
			domain, err = ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, domainFlag)
			if err != nil {
				return fmt.Errorf("domain '%s' not found: %w", domainFlag, err)
			}
			// Temporarily set active domain for handler
			ds.SetActiveDomain(&domain.ID)
		} else {
			domain, err = getActiveDomain(ds)
			if err != nil {
				render.Error("No domain specified")
				render.Info("Hint: Use --domain <name> or 'dvm use domain <name>' first")
				return errSilent
			}
		}

		// Resolve system from flag if provided (for context)
		if systemFlag != "" {
			// Global system lookup: find system across all domains (#287)
			resolution, sErr := resolveSystemGlobally(ds, systemFlag, domain)
			if sErr != nil {
				return sErr
			}
			ds.SetActiveSystem(&resolution.System.ID)
			// Use the system's domain if it differs from active context
			if resolution.Domain.ID != domain.ID {
				render.Info(fmt.Sprintf("System '%s' is in domain '%s' (active domain: '%s') — using system's domain",
					systemFlag, resolution.Domain.Name, domain.Name))
				domain = resolution.Domain
				ds.SetActiveDomain(&domain.ID)
			}
		}

		// Look up app to show cascade info
		app, err := ds.GetAppByName(sql.NullInt64{Int64: int64(domain.ID), Valid: true}, appName)
		if err != nil {
			return fmt.Errorf("app '%s' not found in domain '%s'", appName, domain.Name)
		}

		// Count cascade children for the confirmation message
		workspaces, _ := ds.ListWorkspacesByApp(app.ID)

		// Build confirmation message showing cascade scope
		msg := fmt.Sprintf("Delete app '%s' from domain '%s'", appName, domain.Name)
		if len(workspaces) > 0 {
			msg += fmt.Sprintf(" and all its workspaces (%d workspace(s))?", len(workspaces))
		} else {
			msg += "?"
		}

		// Dry-run: preview what would be deleted
		if deleteAppDryRun {
			render.Plain(fmt.Sprintf("Would delete app %q from domain %q (%d workspace(s))",
				appName, domain.Name, len(workspaces)))
			return nil
		}

		force, _ := cmd.Flags().GetBool("force")
		confirmed, err := confirmDelete(msg, force)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		// Build resource context and use unified handler
		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}

		render.Progress(fmt.Sprintf("Deleting app '%s' from domain '%s'...", appName, domain.Name))

		if err := resource.Delete(ctx, handlers.KindApp, appName); err != nil {
			return fmt.Errorf("failed to delete app: %w", err)
		}

		render.Success(fmt.Sprintf("App '%s' deleted successfully", appName))
		return nil
	},
}

func init() {
	// Add app commands to parent commands
	createCmd.AddCommand(createAppCmd)
	getCmd.AddCommand(getAppsCmd)
	getCmd.AddCommand(getAppCmd)
	// Note: useAppCmd is registered in cmd/use.go to avoid duplicate registration

	// Check if deleteCmd exists before adding
	if deleteCmd != nil {
		deleteCmd.AddCommand(deleteAppCmd)
	}

	// App creation flags
	createAppCmd.Flags().StringVar(&appDescription, "description", "", "App description")
	createAppCmd.Flags().StringVar(&appDomain, "domain", "", "Domain name (defaults to active domain)")
	createAppCmd.Flags().StringVarP(&appSystem, "system", "s", "", "System name (defaults to active system)")
	createAppCmd.Flags().StringVar(&appPath, "path", "", "Path to the app source code")
	createAppCmd.Flags().BoolVar(&appFromCwd, "from-cwd", false, "Use current working directory as app path")
	createAppCmd.Flags().StringVar(&appRepo, "repo", "", "Git repository (URL or existing GitRepo name)")
	AddDryRunFlag(createAppCmd, &createAppDryRun)

	// App get/delete flags
	getAppsCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	getAppsCmd.Flags().StringP("system", "s", "", "System name (filter apps by system)")
	AddAllFlag(getAppsCmd, "List apps from all domains")
	getAppsCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	getAppCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	getAppCmd.Flags().StringP("system", "s", "", "System name (resolve system context)")
	getAppCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	deleteAppCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	deleteAppCmd.Flags().StringP("system", "s", "", "System name (resolve system context)")
	AddForceConfirmFlag(deleteAppCmd)
	AddDryRunFlag(deleteAppCmd, &deleteAppDryRun)
}

// getActiveApp returns the active app from the context
func getActiveApp(ds db.DataStore) (*models.App, error) {
	ctx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if ctx.ActiveAppID == nil {
		return nil, fmt.Errorf("no active app set")
	}

	app, err := ds.GetAppByID(*ctx.ActiveAppID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app: %w", err)
	}

	return app, nil
}

// resolveOrCreateGitRepo handles the --repo flag logic:
// - If repo is a URL, auto-create a GitRepo (or reuse existing by URL)
// - If repo is a name, look up existing GitRepo
// Returns: gitRepoID, gitRepoName, path (for app), error
func resolveOrCreateGitRepo(ds db.DataStore, repo string) (*int, string, string, error) {
	// Check if repo looks like a URL
	isURL := strings.Contains(repo, "://") || strings.HasPrefix(repo, "git@")

	if isURL {
		// Validate the URL
		if err := mirror.ValidateGitURL(repo); err != nil {
			return nil, "", "", fmt.Errorf("invalid git URL: %w", err)
		}

		// Generate slug from URL
		slug, err := mirror.GenerateSlug(repo)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate slug from URL: %w", err)
		}

		// Check if a GitRepo with this URL already exists
		existingRepos, err := ds.ListGitRepos()
		if err == nil {
			for _, r := range existingRepos {
				if r.URL == repo {
					// Reuse existing GitRepo
					id := r.ID
					path := getGitRepoPath(r.Slug)
					return &id, r.Name, path, nil
				}
			}
		}

		// Check if a GitRepo with this slug name exists
		existingBySlug, err := ds.GetGitRepoByName(slug)
		if err != nil && !db.IsNotFound(err) {
			return nil, "", "", fmt.Errorf("failed to check existing GitRepo: %w", err)
		}
		if existingBySlug != nil {
			// Use existing GitRepo if URL matches
			if existingBySlug.URL == repo {
				id := existingBySlug.ID
				path := getGitRepoPath(existingBySlug.Slug)
				return &id, existingBySlug.Name, path, nil
			}
			// Different URL - require user intervention instead of auto-generating name
			return nil, "", "", fmt.Errorf(
				"GitRepo name '%s' already exists with different URL\n\n"+
					"  Existing URL: %s\n"+
					"  Provided URL: %s\n\n"+
					"To create with custom name:\n"+
					"  dvm create gitrepo <custom-name> --url %s\n"+
					"  dvm create app <app-name> --repo <custom-name>",
				slug, existingBySlug.URL, repo, repo)
		}

		// Create new GitRepo
		render.Progress(fmt.Sprintf("Creating GitRepo '%s' from URL...", slug))

		// Auto-detect default branch for the new GitRepo
		detectedRef := utils.DetectDefaultBranch(repo)

		gitRepo := &models.GitRepoDB{
			Name:                slug,
			URL:                 repo,
			Slug:                slug,
			DefaultRef:          detectedRef,
			AuthType:            "none",
			AutoSync:            true,
			SyncIntervalMinutes: 60,
			SyncStatus:          "pending",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		if err := ds.CreateGitRepo(gitRepo); err != nil {
			return nil, "", "", fmt.Errorf("failed to create GitRepo: %w", err)
		}

		// Clone the repository
		baseDir := getGitRepoBaseDir()
		mirrorMgr := mirror.NewGitMirrorManager(baseDir)
		if _, err := mirrorMgr.Clone(repo, slug); err != nil {
			// Update sync status to failed but continue
			gitRepo.SyncStatus = "failed"
			gitRepo.SyncError = sql.NullString{String: err.Error(), Valid: true}
			ds.UpdateGitRepo(gitRepo)
			render.Warning(fmt.Sprintf("Created GitRepo '%s' but initial sync failed: %v", slug, err))
		} else {
			// Update sync status to synced
			gitRepo.LastSyncedAt = sql.NullTime{Time: time.Now(), Valid: true}
			gitRepo.SyncStatus = "synced"
			ds.UpdateGitRepo(gitRepo)
		}

		// Get the created repo to get its ID
		createdRepo, err := ds.GetGitRepoByName(slug)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to retrieve created GitRepo: %w", err)
		}

		render.Success(fmt.Sprintf("Created GitRepo '%s'", slug))
		id := createdRepo.ID
		path := getGitRepoPath(slug)
		return &id, slug, path, nil
	}

	// Not a URL - look up existing GitRepo by name
	existingRepo, err := ds.GetGitRepoByName(repo)
	if err != nil || existingRepo == nil {
		return nil, "", "", fmt.Errorf(
			"GitRepo '%s' not found\n\n"+
				"To auto-create from URL:\n"+
				"  dvm create app <app-name> --repo https://github.com/user/repo.git\n\n"+
				"Or create GitRepo first:\n"+
				"  dvm create gitrepo %s --url <url>\n"+
				"  dvm create app <app-name> --repo %s",
			repo, repo, repo)
	}

	id := existingRepo.ID
	path := getGitRepoPath(existingRepo.Slug)
	return &id, existingRepo.Name, path, nil
}

// getGitRepoPath returns the local path for a GitRepo's mirror
func getGitRepoPath(slug string) string {
	baseDir := getGitRepoBaseDir()
	return filepath.Join(baseDir, slug)
}
