package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

var (
	domainDescription string
	domainEcosystem   string
)

// createDomainCmd creates a new domain
var createDomainCmd = &cobra.Command{
	Use:     "domain <name>",
	Aliases: []string{"dom"},
	Short:   "Create a new domain",
	Long: `Create a new domain within an ecosystem.

A domain represents a bounded context within an ecosystem: Ecosystem -> Domain -> App -> Workspace.
Domains group related applications together.

Examples:
  # Create a domain in the active ecosystem
  dvm create domain backend
  dvm create dom backend             # Short form
  
  # Create a domain in a specific ecosystem
  dvm create domain backend --ecosystem my-platform
  
  # Create with description
  dvm create domain backend --description "Backend services"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domainName := args[0]

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		// Get ecosystem - from flag or active context
		var ecosystem *models.Ecosystem
		if domainEcosystem != "" {
			ecosystem, err = ds.GetEcosystemByName(domainEcosystem)
			if err != nil {
				return fmt.Errorf("ecosystem '%s' not found: %w", domainEcosystem, err)
			}
		} else {
			ecosystem, err = getActiveEcosystem(ds)
			if err != nil {
				render.Error("No ecosystem specified")
				render.Info("Hint: Use --ecosystem <name> or 'dvm use ecosystem <name>' to select an ecosystem first")
				return nil
			}
		}

		render.Progress(fmt.Sprintf("Creating domain '%s' in ecosystem '%s'...", domainName, ecosystem.Name))

		// Check if domain already exists
		existing, _ := ds.GetDomainByName(ecosystem.ID, domainName)
		if existing != nil {
			return fmt.Errorf("domain '%s' already exists in ecosystem '%s'", domainName, ecosystem.Name)
		}

		// Create domain using handler helper
		domain := handlers.NewDomainFromModel(domainName, ecosystem.ID, domainDescription)

		if err := ds.CreateDomain(domain); err != nil {
			return fmt.Errorf("failed to create domain: %w", err)
		}

		// Get the created domain to get its ID
		createdDomain, err := ds.GetDomainByName(ecosystem.ID, domainName)
		if err != nil {
			return fmt.Errorf("failed to retrieve created domain: %w", err)
		}

		render.Success(fmt.Sprintf("Domain '%s' created successfully (ID: %d)", domainName, createdDomain.ID))
		render.Info(fmt.Sprintf("Ecosystem: %s", ecosystem.Name))

		// Set domain as active context
		if err := ds.SetActiveDomain(&createdDomain.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to set active domain: %v", err))
		} else {
			render.Success(fmt.Sprintf("Set '%s' as active domain", domainName))
		}

		fmt.Println()
		render.Info("Next steps:")
		render.Info("  1. Create an app in this domain:")
		render.Info("     dvm create app <name> --path <path>")
		return nil
	},
}

// getDomainsCmd lists all domains
var getDomainsCmd = &cobra.Command{
	Use:     "domains",
	Aliases: []string{"dom"},
	Short:   "List all domains",
	Long: `List all domains, optionally filtered by ecosystem.

Examples:
  dvm get domains                       # List domains in active ecosystem
  dvm get dom                           # Short form
  dvm get domains --ecosystem my-platform
  dvm get domains -A                    # List all domains across all ecosystems
  dvm get domains --all                 # Same as -A
  dvm get domains -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getDomains(cmd)
	},
}

// getDomainCmd gets a specific domain
var getDomainCmd = &cobra.Command{
	Use:     "domain <name>",
	Aliases: []string{"dom"},
	Short:   "Get a specific domain",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getDomain(cmd, args[0])
	},
}

// useDomainCmd switches the active domain
var useDomainCmd = &cobra.Command{
	Use:     "domain <name>",
	Aliases: []string{"dom"},
	Short:   "Switch to a domain",
	Long: `Set the specified domain as the active context.

Requires an active ecosystem to be set first (unless clearing with 'none').
Use 'none' as the name to clear the domain context (also clears app).

Examples:
  dvm use domain backend        # Set active domain
  dvm use dom backend           # Short form
  dvm use domain frontend       # Switch to another domain
  dvm use domain none           # Clear domain context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domainName := args[0]

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		// Handle "none" to clear context
		if domainName == "none" {
			if err := ds.SetActiveDomain(nil); err != nil {
				return fmt.Errorf("failed to clear domain context: %w", err)
			}
			// Also clear downstream context (app)
			ds.SetActiveApp(nil)

			render.Success("Cleared domain context (app also cleared)")
			return nil
		}

		// Get active ecosystem
		ecosystem, err := getActiveEcosystem(ds)
		if err != nil {
			render.Error("No active ecosystem set")
			render.Info("Hint: Set active ecosystem first with: dvm use ecosystem <name>")
			return nil
		}

		// Verify domain exists
		domain, err := ds.GetDomainByName(ecosystem.ID, domainName)
		if err != nil {
			render.Error(fmt.Sprintf("Domain '%s' not found in ecosystem '%s': %v", domainName, ecosystem.Name, err))
			render.Info("Hint: List available domains with: dvm get domains")
			return nil
		}

		// Set domain as active
		if err := ds.SetActiveDomain(&domain.ID); err != nil {
			return fmt.Errorf("failed to set active domain: %w", err)
		}

		// Clear downstream context since we're switching domains
		ds.SetActiveApp(nil)

		render.Success(fmt.Sprintf("Switched to domain '%s' in ecosystem '%s'", domainName, ecosystem.Name))
		fmt.Println()
		render.Info("Next: Select an app with: dvm use app <name>")
		return nil
	},
}

func getDomains(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	ecosystemFlag, _ := cmd.Flags().GetString("ecosystem")

	var domains []*models.Domain
	var ecosystemName string

	if allFlag {
		// List all domains across all ecosystems
		domains, err = ds.ListAllDomains()
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}
		ecosystemName = "(all)"
	} else {
		// Get ecosystem from flag or active context
		var ecosystem *models.Ecosystem
		if ecosystemFlag != "" {
			ecosystem, err = ds.GetEcosystemByName(ecosystemFlag)
			if err != nil {
				return fmt.Errorf("ecosystem '%s' not found: %w", ecosystemFlag, err)
			}
		} else {
			ecosystem, err = getActiveEcosystem(ds)
			if err != nil {
				render.Error("No ecosystem specified")
				render.Info("Hint: Use --ecosystem <name>, --all, or 'dvm use ecosystem <name>' first")
				return nil
			}
		}

		ecosystemName = ecosystem.Name
		domains, err = ds.ListDomainsByEcosystem(ecosystem.ID)
		if err != nil {
			return fmt.Errorf("failed to list domains: %w", err)
		}
	}

	// Get active domain for highlighting
	ctx, _ := ds.GetContext()
	var activeDomainID *int
	if ctx != nil {
		activeDomainID = ctx.ActiveDomainID
	}

	if len(domains) == 0 {
		msg := fmt.Sprintf("No domains found in ecosystem '%s'", ecosystemName)
		if allFlag {
			msg = "No domains found"
		}
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: msg,
			EmptyHints:   []string{"dvm create domain <name>"},
		})
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		// Need to get ecosystem names for YAML output
		domainsYAML := make([]models.DomainYAML, len(domains))
		for i, d := range domains {
			eco, _ := ds.GetEcosystemByID(d.EcosystemID)
			ecoName := ""
			if eco != nil {
				ecoName = eco.Name
			}
			domainsYAML[i] = d.ToYAML(ecoName)
		}
		return render.OutputWith(getOutputFormat, domainsYAML, render.Options{})
	}

	// For human output, build table data
	headers := []string{"NAME", "ECOSYSTEM", "DESCRIPTION", "CREATED"}
	if showTheme {
		headers = append(headers, "THEME", "THEME SOURCE")
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    make([][]string, len(domains)),
	}

	// Create theme resolver if needed
	var themeResolver themeresolver.ThemeResolver
	if showTheme {
		themeResolver, _ = themeresolver.NewThemeResolver(ds, nil)
	}

	for i, d := range domains {
		name := d.Name
		if activeDomainID != nil && d.ID == *activeDomainID {
			name = "● " + name // Active indicator
		}

		// Get ecosystem name for display
		eco, _ := ds.GetEcosystemByID(d.EcosystemID)
		ecoName := ""
		if eco != nil {
			ecoName = eco.Name
		}

		desc := ""
		if d.Description.Valid {
			desc = d.Description.String
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
		}

		row := []string{
			name,
			ecoName,
			desc,
			d.CreatedAt.Format("2006-01-02 15:04"),
		}

		// Add theme information if requested
		if showTheme && themeResolver != nil {
			themeName := themeresolver.DefaultTheme
			themeSource := "default"

			if resolution, err := themeResolver.GetResolutionPath(cmd.Context(), themeresolver.LevelDomain, d.ID); err == nil {
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

func getDomain(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	ecosystemFlag, _ := cmd.Flags().GetString("ecosystem")

	// Get ecosystem from flag or active context
	var ecosystem *models.Ecosystem
	if ecosystemFlag != "" {
		ecosystem, err = ds.GetEcosystemByName(ecosystemFlag)
		if err != nil {
			return fmt.Errorf("ecosystem '%s' not found: %w", ecosystemFlag, err)
		}
		// Temporarily set active ecosystem for handler
		ds.SetActiveEcosystem(&ecosystem.ID)
	} else {
		ecosystem, err = getActiveEcosystem(ds)
		if err != nil {
			render.Error("No ecosystem specified")
			render.Info("Hint: Use --ecosystem <name> or 'dvm use ecosystem <name>' first")
			return nil
		}
	}

	res, err := resource.Get(ctx, handlers.KindDomain, name)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	domain := res.(*handlers.DomainResource).Domain()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, domain.ToYAML(ecosystem.Name), render.Options{})
	}

	// For human output, show detail view
	dbCtx, _ := ds.GetContext()
	var activeDomainID *int
	if dbCtx != nil {
		activeDomainID = dbCtx.ActiveDomainID
	}

	isActive := activeDomainID != nil && domain.ID == *activeDomainID
	nameDisplay := domain.Name
	if isActive {
		nameDisplay = "● " + nameDisplay + " (active)"
	}

	desc := ""
	if domain.Description.Valid {
		desc = domain.Description.String
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: nameDisplay},
		render.KeyValue{Key: "Ecosystem", Value: ecosystem.Name},
		render.KeyValue{Key: "Description", Value: desc},
		render.KeyValue{Key: "Created", Value: domain.CreatedAt.Format("2006-01-02 15:04:05")},
	)

	err = render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Domain Details",
	})
	if err != nil {
		return err
	}

	// Show theme information if requested
	if showTheme {
		return showThemeResolution(cmd, ds, themeresolver.LevelDomain, domain.ID, domain.Name)
	}

	return nil
}

// deleteDomainCmd deletes a domain
var deleteDomainCmd = &cobra.Command{
	Use:     "domain <name>",
	Aliases: []string{"dom"},
	Short:   "Delete a domain",
	Long: `Delete a domain by name.

WARNING: This will also delete all apps within the domain.

Examples:
  dvm delete domain backend
  dvm delete dom backend         # Short form
  dvm delete domain backend --ecosystem my-platform`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domainName := args[0]

		// Build resource context and use unified handler
		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		ecosystemFlag, _ := cmd.Flags().GetString("ecosystem")

		// Get ecosystem from flag or active context
		var ecosystem *models.Ecosystem
		if ecosystemFlag != "" {
			ecosystem, err = ds.GetEcosystemByName(ecosystemFlag)
			if err != nil {
				return fmt.Errorf("ecosystem '%s' not found: %w", ecosystemFlag, err)
			}
			// Temporarily set active ecosystem for handler
			ds.SetActiveEcosystem(&ecosystem.ID)
		} else {
			ecosystem, err = getActiveEcosystem(ds)
			if err != nil {
				render.Error("No ecosystem specified")
				render.Info("Hint: Use --ecosystem <name> or 'dvm use ecosystem <name>' first")
				return nil
			}
		}

		render.Progress(fmt.Sprintf("Deleting domain '%s' from ecosystem '%s'...", domainName, ecosystem.Name))

		if err := resource.Delete(ctx, handlers.KindDomain, domainName); err != nil {
			return fmt.Errorf("failed to delete domain: %w", err)
		}

		render.Success(fmt.Sprintf("Domain '%s' deleted successfully", domainName))
		return nil
	},
}

func init() {
	// Add domain commands to parent commands
	createCmd.AddCommand(createDomainCmd)
	getCmd.AddCommand(getDomainsCmd)
	getCmd.AddCommand(getDomainCmd)
	useCmd.AddCommand(useDomainCmd)

	// Check if deleteCmd exists before adding
	if deleteCmd != nil {
		deleteCmd.AddCommand(deleteDomainCmd)
	}

	// Domain creation flags
	createDomainCmd.Flags().StringVar(&domainDescription, "description", "", "Domain description")
	createDomainCmd.Flags().StringVar(&domainEcosystem, "ecosystem", "", "Ecosystem name (defaults to active ecosystem)")

	// Domain get/delete flags
	getDomainsCmd.Flags().StringP("ecosystem", "e", "", "Ecosystem name (defaults to active ecosystem)")
	getDomainsCmd.Flags().BoolP("all", "A", false, "List domains from all ecosystems")
	getDomainsCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	getDomainCmd.Flags().StringP("ecosystem", "e", "", "Ecosystem name (defaults to active ecosystem)")
	getDomainCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	deleteDomainCmd.Flags().StringP("ecosystem", "e", "", "Ecosystem name (defaults to active ecosystem)")
}

// getActiveDomain returns the active domain from the context
func getActiveDomain(ds db.DataStore) (*models.Domain, error) {
	ctx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if ctx.ActiveDomainID == nil {
		return nil, fmt.Errorf("no active domain set")
	}

	domain, err := ds.GetDomainByID(*ctx.ActiveDomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return domain, nil
}
