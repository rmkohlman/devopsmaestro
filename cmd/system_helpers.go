package cmd

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

func createSystem(cmd *cobra.Command, systemName string) error {
	if err := ValidateResourceName(systemName, "system"); err != nil {
		return err
	}

	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Resolve domain from flag or active context (optional)
	var domain *models.Domain
	var domainID sql.NullInt64
	var ecosystemID sql.NullInt64
	var domainName, ecosystemName string

	if systemDomain != "" {
		// Domain specified — also need ecosystem
		var ecosystem *models.Ecosystem
		if systemEcosystem != "" {
			ecosystem, err = ds.GetEcosystemByName(systemEcosystem)
			if err != nil {
				return fmt.Errorf("ecosystem '%s' not found: %w", systemEcosystem, err)
			}
		} else {
			ecosystem, err = getActiveEcosystem(ds)
			if err != nil {
				render.Error("No ecosystem specified")
				render.Info("Hint: Use --ecosystem <name> or 'dvm use ecosystem <name>' first")
				return errSilent
			}
		}
		ecosystemName = ecosystem.Name
		ecosystemID = sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}

		domain, err = ds.GetDomainByName(ecosystem.ID, systemDomain)
		if err != nil {
			return fmt.Errorf("domain '%s' not found in ecosystem '%s': %w", systemDomain, ecosystem.Name, err)
		}
		domainID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
		domainName = domain.Name
	} else {
		// Try active domain
		domain, err = getActiveDomain(ds)
		if err == nil && domain != nil {
			domainID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
			domainName = domain.Name
			if eco, e := ds.GetEcosystemByID(domain.EcosystemID); e == nil {
				ecosystemName = eco.Name
				ecosystemID = sql.NullInt64{Int64: int64(eco.ID), Valid: true}
			}
		}
		// It's OK to have no domain — systems can be standalone
	}

	contextMsg := "globally"
	if domainName != "" {
		contextMsg = fmt.Sprintf("in domain '%s'", domainName)
	}
	render.Progress(fmt.Sprintf("Creating system '%s' %s...", systemName, contextMsg))

	// Dry-run: preview what would be created
	if createSystemDryRun {
		render.Plain(fmt.Sprintf("Would create system %q %s", systemName, contextMsg))
		if systemDescription != "" {
			render.Plain(fmt.Sprintf("  description: %s", systemDescription))
		}
		return nil
	}

	// Check if system already exists
	existing, _ := ds.GetSystemByName(domainID, systemName)
	if existing != nil {
		return fmt.Errorf("system '%s' already exists %s", systemName, contextMsg)
	}

	system := handlers.NewSystemFromModel(systemName, domainID, ecosystemID, systemDescription)

	if err := ds.CreateSystem(system); err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	// Get the created system to get its ID
	createdSystem, err := ds.GetSystemByName(domainID, systemName)
	if err != nil {
		return fmt.Errorf("failed to retrieve created system: %w", err)
	}

	render.Success(fmt.Sprintf("System '%s' created successfully (ID: %d)", systemName, createdSystem.ID))
	if domainName != "" {
		render.Info(fmt.Sprintf("Domain: %s", domainName))
	}
	if ecosystemName != "" {
		render.Info(fmt.Sprintf("Ecosystem: %s", ecosystemName))
	}

	// Set system as active context
	if err := ds.SetActiveSystem(&createdSystem.ID); err != nil {
		render.Warning(fmt.Sprintf("Failed to set active system: %v", err))
	} else {
		render.Success(fmt.Sprintf("Set '%s' as active system", systemName))
	}

	render.Blank()
	render.Info("Next steps:")
	render.Info("  1. Create an app in this system:")
	render.Info("     dvm create app <name> --path <path>")
	return nil
}

func getSystems(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	domainFlag, _ := cmd.Flags().GetString("domain")

	var systems []*models.System
	var contextLabel string

	if allFlag {
		systems, err = ds.ListSystems()
		if err != nil {
			return fmt.Errorf("failed to list systems: %w", err)
		}
		contextLabel = "(all)"
	} else {
		var domain *models.Domain
		if domainFlag != "" {
			ecosystem, ecoErr := resolveEcosystemForDomain(ds, cmd)
			if ecoErr != nil {
				return ecoErr
			}
			domain, err = ds.GetDomainByName(ecosystem.ID, domainFlag)
			if err != nil {
				return fmt.Errorf("domain '%s' not found: %w", domainFlag, err)
			}
		} else {
			domain, err = getActiveDomain(ds)
			if err != nil {
				// No active domain — show all
				systems, err = ds.ListSystems()
				if err != nil {
					return fmt.Errorf("failed to list systems: %w", err)
				}
				contextLabel = "(all)"
			}
		}

		if domain != nil {
			contextLabel = domain.Name
			systems, err = ds.ListSystemsByDomain(domain.ID)
			if err != nil {
				return fmt.Errorf("failed to list systems: %w", err)
			}
		}
	}

	// Get active system for highlighting
	ctx, _ := ds.GetContext()
	var activeSystemID *int
	if ctx != nil {
		activeSystemID = ctx.ActiveSystemID
	}

	// JSON/YAML output
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		handlers.RegisterAll()
		if len(systems) == 0 {
			return render.OutputWith(getOutputFormat, resource.NewResourceList(), render.Options{Type: render.TypeAuto})
		}
		systemResources := make([]resource.Resource, len(systems))
		for i, s := range systems {
			domName, ecoName := resolveSystemParents(ds, s)
			systemResources[i] = handlers.NewSystemResource(s, domName, ecoName)
		}
		resCtx := resource.Context{DataStore: ds}
		list, err := resource.BuildList(resCtx, systemResources)
		if err != nil {
			return fmt.Errorf("failed to build resource list: %w", err)
		}
		return render.OutputWith(getOutputFormat, list, render.Options{Type: render.TypeAuto})
	}

	if len(systems) == 0 {
		msg := fmt.Sprintf("No systems found in domain '%s'", contextLabel)
		if allFlag || contextLabel == "(all)" {
			msg = "No systems found"
		}
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: msg,
			EmptyHints:   []string{"dvm create system <name>"},
		})
	}

	// Build table
	wide := getOutputFormat == "wide"
	b := &systemTableBuilder{DataStore: ds, ActiveID: activeSystemID}
	td := BuildTable(b, systems, wide)

	renderFormat := getOutputFormat
	if wide {
		renderFormat = "table"
	}
	return render.OutputWith(renderFormat, td, render.Options{
		Type: render.TypeTable,
	})
}

func getSystem(cmd *cobra.Command, name string) error {
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Resolve domain context if flag provided
	domainFlag, _ := cmd.Flags().GetString("domain")
	if domainFlag != "" {
		ecosystem, ecoErr := resolveEcosystemForDomain(ds, cmd)
		if ecoErr != nil {
			return ecoErr
		}
		domain, domErr := ds.GetDomainByName(ecosystem.ID, domainFlag)
		if domErr != nil {
			return fmt.Errorf("domain '%s' not found: %w", domainFlag, domErr)
		}
		ds.SetActiveDomain(&domain.ID)
	}

	res, err := resource.Get(ctx, handlers.KindSystem, name)
	if err != nil {
		return fmt.Errorf("failed to get system: %w", err)
	}

	system := res.(*handlers.SystemResource).System()
	domName := res.(*handlers.SystemResource).DomainName()
	ecoName := res.(*handlers.SystemResource).EcosystemName()

	// JSON/YAML
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, system.ToYAML(domName, nil), render.Options{})
	}

	// Human output — detail view
	dbCtx, _ := ds.GetContext()
	var activeSystemID *int
	if dbCtx != nil {
		activeSystemID = dbCtx.ActiveSystemID
	}

	isActive := activeSystemID != nil && system.ID == *activeSystemID
	nameDisplay := system.Name
	if isActive {
		nameDisplay = "● " + nameDisplay + " (active)"
	}

	desc := ""
	if system.Description.Valid {
		desc = system.Description.String
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: nameDisplay},
		render.KeyValue{Key: "Domain", Value: domName},
		render.KeyValue{Key: "Ecosystem", Value: ecoName},
		render.KeyValue{Key: "Description", Value: desc},
		render.KeyValue{Key: "Created", Value: system.CreatedAt.Format("2006-01-02 15:04:05")},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "System Details",
	})
}

func useSystem(cmd *cobra.Command, systemName string) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Handle "none" to clear context
	if systemName == "none" {
		if err := ds.SetActiveSystem(nil); err != nil {
			return fmt.Errorf("failed to clear system context: %w", err)
		}
		// Also clear downstream context (app, workspace)
		ds.SetActiveApp(nil)
		ds.SetActiveWorkspace(nil)

		render.Success("Cleared system context (app and workspace also cleared)")
		return nil
	}

	// Find the system — use FindSystemsByName for disambiguation
	matches, err := ds.FindSystemsByName(systemName)
	if err != nil {
		return fmt.Errorf("failed to find system: %w", err)
	}

	if len(matches) == 0 {
		render.Error(fmt.Sprintf("System '%s' not found", systemName))
		render.Info("Hint: List available systems with: dvm get systems")
		return errSilent
	}

	// If multiple matches, try to disambiguate via active domain
	var system *models.System
	if len(matches) == 1 {
		system = matches[0].System
	} else {
		// Try to narrow by active domain
		activeDomain, domErr := getActiveDomain(ds)
		if domErr == nil && activeDomain != nil {
			for _, m := range matches {
				if m.Domain != nil && m.Domain.ID == activeDomain.ID {
					system = m.System
					break
				}
			}
		}
		if system == nil {
			render.Error(fmt.Sprintf("Ambiguous system name '%s' (%d matches)", systemName, len(matches)))
			render.Info("Hint: Use 'dvm use domain <name>' to set context, or use --domain flag")
			return errSilent
		}
	}

	// Handle --export flag
	exportFlag, _ := cmd.Flags().GetBool("export")
	if exportFlag {
		fmt.Fprintf(cmd.OutOrStdout(), "export DVM_SYSTEM=%s\n", systemName)
		return nil
	}

	// Dry-run
	if useSystemDryRun {
		render.Plain(fmt.Sprintf("Would switch active system to %q", systemName))
		return nil
	}

	// Save current context before switching
	if err := saveCurrentContext(ds); err != nil {
		return fmt.Errorf("failed to save previous context: %w", err)
	}

	// Set system as active
	if err := ds.SetActiveSystem(&system.ID); err != nil {
		return fmt.Errorf("failed to set active system: %w", err)
	}

	// Clear downstream context
	ds.SetActiveApp(nil)
	ds.SetActiveWorkspace(nil)

	render.Success(fmt.Sprintf("Switched to system '%s'", systemName))
	render.Blank()
	render.Info("Next: Select an app with: dvm use app <name>")
	return nil
}

func deleteSystem(cmd *cobra.Command, systemName string) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	domainFlag, _ := cmd.Flags().GetString("domain")

	// Resolve domain context
	var domainID sql.NullInt64
	if domainFlag != "" {
		ecosystem, ecoErr := resolveEcosystemForDomain(ds, cmd)
		if ecoErr != nil {
			return ecoErr
		}
		domain, domErr := ds.GetDomainByName(ecosystem.ID, domainFlag)
		if domErr != nil {
			return fmt.Errorf("domain '%s' not found: %w", domainFlag, domErr)
		}
		domainID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
		ds.SetActiveDomain(&domain.ID)
	} else {
		domain, domErr := getActiveDomain(ds)
		if domErr == nil && domain != nil {
			domainID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
		}
	}

	// Look up system
	system, err := ds.GetSystemByName(domainID, systemName)
	if err != nil {
		contextMsg := ""
		if domainFlag != "" {
			contextMsg = fmt.Sprintf(" in domain '%s'", domainFlag)
		}
		return fmt.Errorf("system '%s' not found%s", systemName, contextMsg)
	}

	// Build confirmation message
	msg := fmt.Sprintf("Delete system '%s'", systemName)

	// Dry-run
	if deleteSystemDryRun {
		render.Plain(fmt.Sprintf("Would delete system %q", systemName))
		return nil
	}

	force, _ := cmd.Flags().GetBool("force")
	confirmed, err := confirmDelete(msg+"?", force)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	// Use resource handler for deletion
	resCtx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	_ = system // Used for lookup above; deletion is by name via handler

	render.Progress(fmt.Sprintf("Deleting system '%s'...", systemName))

	if err := resource.Delete(resCtx, handlers.KindSystem, systemName); err != nil {
		return fmt.Errorf("failed to delete system: %w", err)
	}

	render.Success(fmt.Sprintf("System '%s' deleted successfully", systemName))
	return nil
}

// resolveEcosystemForDomain resolves the ecosystem for domain resolution.
// Uses --ecosystem flag if available, otherwise falls back to active ecosystem.
func resolveEcosystemForDomain(ds db.DataStore, cmd *cobra.Command) (*models.Ecosystem, error) {
	ecoFlag, _ := cmd.Flags().GetString("ecosystem")
	if ecoFlag != "" {
		eco, err := ds.GetEcosystemByName(ecoFlag)
		if err != nil {
			return nil, fmt.Errorf("ecosystem '%s' not found: %w", ecoFlag, err)
		}
		return eco, nil
	}
	eco, err := getActiveEcosystem(ds)
	if err != nil {
		render.Error("No ecosystem specified")
		render.Info("Hint: Use --ecosystem <name> or 'dvm use ecosystem <name>' first")
		return nil, errSilent
	}
	return eco, nil
}

// resolveSystemParents resolves domain and ecosystem names for a system.
func resolveSystemParents(ds db.DataStore, system *models.System) (domainName, ecosystemName string) {
	if system.DomainID.Valid {
		if domain, err := ds.GetDomainByID(int(system.DomainID.Int64)); err == nil {
			domainName = domain.Name
			if eco, err := ds.GetEcosystemByID(domain.EcosystemID); err == nil {
				ecosystemName = eco.Name
			}
		}
	}
	return
}

// getActiveSystem returns the active system from the context
func getActiveSystem(ds db.DataStore) (*models.System, error) {
	ctx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if ctx.ActiveSystemID == nil {
		return nil, fmt.Errorf("no active system set")
	}

	system, err := ds.GetSystemByID(*ctx.ActiveSystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}

	return system, nil
}
