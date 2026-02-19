package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"devopsmaestro/db"
	"devopsmaestro/models"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

var (
	appDescription string
	appDomain      string
	appPath        string
	appFromCwd     bool
)

// createAppCmd creates a new app
var createAppCmd = &cobra.Command{
	Use:     "app <name>",
	Aliases: []string{"application"},
	Short:   "Create a new app",
	Long: `Create a new app within a domain.

An app represents a codebase/application: Ecosystem -> Domain -> App -> Workspace.
Apps have a path to source code and can run in dev mode (Workspace) or live mode (Operator).

Examples:
  # Create an app from the current directory
  dvm create app my-api --from-cwd
  
  # Create an app with a specific path
  dvm create app my-api --path ~/code/my-api
  
  # Create an app in a specific domain
  dvm create app my-api --from-cwd --domain backend
  
  # Create with description
  dvm create app my-api --from-cwd --description "REST API service"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		// Determine app path
		var path string
		if appFromCwd {
			path, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		} else if appPath != "" {
			path, err = filepath.Abs(appPath)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}
		} else {
			render.Error("Must specify either --from-cwd or --path")
			return nil
		}

		// Verify path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}

		// Get domain - from flag or active context
		var domain *models.Domain
		if appDomain != "" {
			// Need active ecosystem to find the domain
			ecosystem, err := getActiveEcosystem(ds)
			if err != nil {
				render.Error("No active ecosystem set")
				render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
				return nil
			}
			domain, err = ds.GetDomainByName(ecosystem.ID, appDomain)
			if err != nil {
				return fmt.Errorf("domain '%s' not found in ecosystem '%s': %w", appDomain, ecosystem.Name, err)
			}
		} else {
			domain, err = getActiveDomain(ds)
			if err != nil {
				render.Error("No domain specified")
				render.Info("Hint: Use --domain <name> or 'dvm use domain <name>' to select a domain first")
				return nil
			}
		}

		// Get ecosystem name for display
		ecosystem, _ := ds.GetEcosystemByID(domain.EcosystemID)
		ecosystemName := ""
		if ecosystem != nil {
			ecosystemName = ecosystem.Name
		}

		render.Progress(fmt.Sprintf("Creating app '%s' in domain '%s' at %s...", appName, domain.Name, path))

		// Check if app already exists
		existing, _ := ds.GetAppByName(domain.ID, appName)
		if existing != nil {
			return fmt.Errorf("app '%s' already exists in domain '%s'", appName, domain.Name)
		}

		// Create app using handler helper
		app := handlers.NewAppFromModel(appName, domain.ID, path, appDescription)

		if err := ds.CreateApp(app); err != nil {
			return fmt.Errorf("failed to create app: %w", err)
		}

		// Get the created app to get its ID
		createdApp, err := ds.GetAppByName(domain.ID, appName)
		if err != nil {
			return fmt.Errorf("failed to retrieve created app: %w", err)
		}

		render.Success(fmt.Sprintf("App '%s' created successfully (ID: %d)", appName, createdApp.ID))
		render.Info(fmt.Sprintf("Ecosystem: %s", ecosystemName))
		render.Info(fmt.Sprintf("Domain: %s", domain.Name))
		render.Info(fmt.Sprintf("Path: %s", path))

		// Set app as active context
		if err := ds.SetActiveApp(&createdApp.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to set active app: %v", err))
		} else {
			render.Success(fmt.Sprintf("Set '%s' as active app", appName))
		}

		fmt.Println()
		render.Info("Next steps:")
		render.Info("  1. Create a workspace for this app:")
		render.Info("     dvm create workspace main")
		render.Info("  2. Build and attach:")
		render.Info("     dvm build && dvm attach")
		return nil
	},
}

// getAppsCmd lists all apps
var getAppsCmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"app", "application", "applications"},
	Short:   "List all apps",
	Long: `List all apps, optionally filtered by domain.

Examples:
  dvm get apps                          # List apps in active domain
  dvm get apps --domain backend
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
	Aliases: []string{"application"},
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
				return nil
			}
			domain, err = ds.GetDomainByName(ecosystem.ID, domainFlag)
			if err != nil {
				return fmt.Errorf("domain '%s' not found: %w", domainFlag, err)
			}
		} else {
			domain, err = getActiveDomain(ds)
			if err != nil {
				render.Error("No domain specified")
				render.Info("Hint: Use --domain <name>, --all, or 'dvm use domain <name>' first")
				return nil
			}
		}

		domainName = domain.Name
		apps, err = ds.ListAppsByDomain(domain.ID)
		if err != nil {
			return fmt.Errorf("failed to list apps: %w", err)
		}
	}

	// Get active app for highlighting
	ctx, _ := ds.GetContext()
	var activeAppID *int
	if ctx != nil {
		activeAppID = ctx.ActiveAppID
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

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		// Need to get domain names for YAML output
		appsYAML := make([]models.AppYAML, len(apps))
		for i, a := range apps {
			dom, _ := ds.GetDomainByID(a.DomainID)
			domName := ""
			if dom != nil {
				domName = dom.Name
			}
			appsYAML[i] = a.ToYAML(domName)
		}
		return render.OutputWith(getOutputFormat, appsYAML, render.Options{})
	}

	// For human output, build table data
	headers := []string{"NAME", "DOMAIN", "PATH", "CREATED"}
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
		dom, _ := ds.GetDomainByID(a.DomainID)
		domName := ""
		if dom != nil {
			domName = dom.Name
		}

		// Truncate path if too long
		pathDisplay := a.Path
		if len(pathDisplay) > 40 {
			pathDisplay = "..." + pathDisplay[len(pathDisplay)-37:]
		}

		row := []string{
			name,
			domName,
			pathDisplay,
			a.CreatedAt.Format("2006-01-02 15:04"),
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

	return render.OutputWith(getOutputFormat, tableData, render.Options{
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

	// Get domain from flag or active context
	var domain *models.Domain
	if domainFlag != "" {
		// Need active ecosystem to find the domain
		ecosystem, err := getActiveEcosystem(ds)
		if err != nil {
			render.Error("No active ecosystem set")
			render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
			return nil
		}
		domain, err = ds.GetDomainByName(ecosystem.ID, domainFlag)
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
			return nil
		}
	}

	// Get ecosystem for display
	ecosystem, _ := ds.GetEcosystemByID(domain.EcosystemID)
	ecosystemName := ""
	if ecosystem != nil {
		ecosystemName = ecosystem.Name
	}

	res, err := resource.Get(ctx, handlers.KindApp, name)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	app := res.(*handlers.AppResource).App()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, app.ToYAML(domain.Name), render.Options{})
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
	Aliases: []string{"application"},
	Short:   "Delete an app",
	Long: `Delete an app by name.

Examples:
  dvm delete app my-api
  dvm delete app my-api --domain backend`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

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

		// Get domain from flag or active context
		var domain *models.Domain
		if domainFlag != "" {
			// Need active ecosystem to find the domain
			ecosystem, err := getActiveEcosystem(ds)
			if err != nil {
				render.Error("No active ecosystem set")
				render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
				return nil
			}
			domain, err = ds.GetDomainByName(ecosystem.ID, domainFlag)
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
				return nil
			}
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
	useCmd.AddCommand(useAppCmd)

	// Check if deleteCmd exists before adding
	if deleteCmd != nil {
		deleteCmd.AddCommand(deleteAppCmd)
	}

	// App creation flags
	createAppCmd.Flags().StringVar(&appDescription, "description", "", "App description")
	createAppCmd.Flags().StringVar(&appDomain, "domain", "", "Domain name (defaults to active domain)")
	createAppCmd.Flags().StringVar(&appPath, "path", "", "Path to the app source code")
	createAppCmd.Flags().BoolVar(&appFromCwd, "from-cwd", false, "Use current working directory as app path")

	// App get/delete flags
	getAppsCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	getAppsCmd.Flags().BoolP("all", "A", false, "List apps from all domains")
	getAppsCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	getAppCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	getAppCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	deleteAppCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
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
