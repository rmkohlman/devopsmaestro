package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

var ecosystemDescription string

// createEcosystemCmd creates a new ecosystem
var createEcosystemCmd = &cobra.Command{
	Use:     "ecosystem <name>",
	Aliases: []string{"eco"},
	Short:   "Create a new ecosystem",
	Long: `Create a new ecosystem with the specified name.

An ecosystem is the top-level grouping in the hierarchy: Ecosystem -> Domain -> App -> Workspace.
It serves as a platform or organizational boundary for domains.

Examples:
  # Create an ecosystem
  dvm create ecosystem my-platform
  dvm create eco my-platform           # Short form
  
  # Create with description
  dvm create ecosystem my-platform --description "Main development platform"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ecosystemName := args[0]

		render.Progress(fmt.Sprintf("Creating ecosystem '%s'...", ecosystemName))

		// Build resource context
		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		// Check if ecosystem already exists
		existing, _ := ds.GetEcosystemByName(ecosystemName)
		if existing != nil {
			return fmt.Errorf("ecosystem '%s' already exists", ecosystemName)
		}

		// Create ecosystem using handler
		ecosystem := handlers.NewEcosystemFromModel(ecosystemName, ecosystemDescription)
		if err := ds.CreateEcosystem(ecosystem); err != nil {
			return fmt.Errorf("failed to create ecosystem: %w", err)
		}

		// Get the created ecosystem to get its ID
		createdEcosystem, err := ds.GetEcosystemByName(ecosystemName)
		if err != nil {
			return fmt.Errorf("failed to retrieve created ecosystem: %w", err)
		}

		render.Success(fmt.Sprintf("Ecosystem '%s' created successfully (ID: %d)", ecosystemName, createdEcosystem.ID))

		// Set ecosystem as active context
		if err := ds.SetActiveEcosystem(&createdEcosystem.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to set active ecosystem: %v", err))
		} else {
			render.Success(fmt.Sprintf("Set '%s' as active ecosystem", ecosystemName))
		}

		fmt.Println()
		render.Info("Next steps:")
		render.Info("  1. Create a domain in this ecosystem:")
		render.Info("     dvm create domain <name>")

		// For testing, we need to use ctx to avoid unused variable error
		_ = ctx
		return nil
	},
}

// getEcosystemsCmd lists all ecosystems
var getEcosystemsCmd = &cobra.Command{
	Use:     "ecosystems",
	Aliases: []string{"eco"},
	Short:   "List all ecosystems",
	Long: `List all ecosystems.

Examples:
  dvm get ecosystems
  dvm get eco                     # Short form
  dvm get ecosystems -o yaml
  dvm get ecosystems -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getEcosystems(cmd)
	},
}

// getEcosystemCmd gets a specific ecosystem
var getEcosystemCmd = &cobra.Command{
	Use:     "ecosystem <name>",
	Aliases: []string{"eco"},
	Short:   "Get a specific ecosystem",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getEcosystem(cmd, args[0])
	},
}

// useEcosystemCmd switches the active ecosystem
var useEcosystemCmd = &cobra.Command{
	Use:     "ecosystem <name>",
	Aliases: []string{"eco"},
	Short:   "Switch to an ecosystem",
	Long: `Set the specified ecosystem as the active context.

Use 'none' as the name to clear the ecosystem context (also clears domain and app).

Examples:
  dvm use ecosystem my-platform    # Set active ecosystem
  dvm use eco my-platform          # Short form
  dvm use ecosystem another        # Switch to another ecosystem
  dvm use ecosystem none           # Clear ecosystem context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ecosystemName := args[0]

		ds, err := getDataStore(cmd)
		if err != nil {
			return err
		}

		// Handle "none" to clear context
		if ecosystemName == "none" {
			if err := ds.SetActiveEcosystem(nil); err != nil {
				return fmt.Errorf("failed to clear ecosystem context: %w", err)
			}
			// Also clear downstream context (domain, app)
			ds.SetActiveDomain(nil)
			ds.SetActiveApp(nil)

			render.Success("Cleared ecosystem context (domain and app also cleared)")
			return nil
		}

		// Verify ecosystem exists
		ecosystem, err := ds.GetEcosystemByName(ecosystemName)
		if err != nil {
			render.Error(fmt.Sprintf("Ecosystem '%s' not found: %v", ecosystemName, err))
			render.Info("Hint: List available ecosystems with: dvm get ecosystems")
			return nil
		}

		// Set ecosystem as active
		if err := ds.SetActiveEcosystem(&ecosystem.ID); err != nil {
			return fmt.Errorf("failed to set active ecosystem: %w", err)
		}

		// Clear downstream context since we're switching ecosystems
		ds.SetActiveDomain(nil)
		ds.SetActiveApp(nil)

		render.Success(fmt.Sprintf("Switched to ecosystem '%s'", ecosystemName))
		fmt.Println()
		render.Info("Next: Select a domain with: dvm use domain <name>")
		return nil
	},
}

func getEcosystems(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindEcosystem)
	if err != nil {
		return fmt.Errorf("failed to list ecosystems: %w", err)
	}

	// Get active ecosystem for highlighting
	ds, _ := getDataStore(cmd)
	var activeEcosystemID *int
	if ds != nil {
		dbCtx, _ := ds.GetContext()
		if dbCtx != nil {
			activeEcosystemID = dbCtx.ActiveEcosystemID
		}
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No ecosystems found",
			EmptyHints:   []string{"dvm create ecosystem <name>"},
		})
	}

	// Extract underlying ecosystems from resources
	ecosystems := make([]*models.Ecosystem, len(resources))
	for i, res := range resources {
		er := res.(*handlers.EcosystemResource)
		ecosystems[i] = er.Ecosystem()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		ecosystemsYAML := make([]models.EcosystemYAML, len(ecosystems))
		for i, e := range ecosystems {
			ecosystemsYAML[i] = e.ToYAML()
		}
		return render.OutputWith(getOutputFormat, ecosystemsYAML, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"NAME", "DESCRIPTION", "CREATED"},
		Rows:    make([][]string, len(ecosystems)),
	}

	for i, e := range ecosystems {
		name := e.Name
		if activeEcosystemID != nil && e.ID == *activeEcosystemID {
			name = "● " + name // Active indicator
		}

		desc := ""
		if e.Description.Valid {
			desc = e.Description.String
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
		}

		tableData.Rows[i] = []string{
			name,
			desc,
			e.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

func getEcosystem(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindEcosystem, name)
	if err != nil {
		return fmt.Errorf("failed to get ecosystem: %w", err)
	}

	ecosystem := res.(*handlers.EcosystemResource).Ecosystem()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, ecosystem.ToYAML(), render.Options{})
	}

	// For human output, show detail view
	ds, _ := getDataStore(cmd)
	var activeEcosystemID *int
	if ds != nil {
		dbCtx, _ := ds.GetContext()
		if dbCtx != nil {
			activeEcosystemID = dbCtx.ActiveEcosystemID
		}
	}

	isActive := activeEcosystemID != nil && ecosystem.ID == *activeEcosystemID
	nameDisplay := ecosystem.Name
	if isActive {
		nameDisplay = "● " + nameDisplay + " (active)"
	}

	desc := ""
	if ecosystem.Description.Valid {
		desc = ecosystem.Description.String
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: nameDisplay},
		render.KeyValue{Key: "Description", Value: desc},
		render.KeyValue{Key: "Created", Value: ecosystem.CreatedAt.Format("2006-01-02 15:04:05")},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Ecosystem Details",
	})
}

// deleteEcosystemCmd deletes an ecosystem
var deleteEcosystemCmd = &cobra.Command{
	Use:     "ecosystem <name>",
	Aliases: []string{"eco"},
	Short:   "Delete an ecosystem",
	Long: `Delete an ecosystem by name.

WARNING: This will also delete all domains and apps within the ecosystem.

Examples:
  dvm delete ecosystem my-platform
  dvm delete eco my-platform       # Short form`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ecosystemName := args[0]

		// Build resource context and use unified handler
		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}

		render.Progress(fmt.Sprintf("Deleting ecosystem '%s'...", ecosystemName))

		if err := resource.Delete(ctx, handlers.KindEcosystem, ecosystemName); err != nil {
			return fmt.Errorf("failed to delete ecosystem: %w", err)
		}

		render.Success(fmt.Sprintf("Ecosystem '%s' deleted successfully", ecosystemName))
		return nil
	},
}

func init() {
	// Add ecosystem commands to parent commands
	createCmd.AddCommand(createEcosystemCmd)
	getCmd.AddCommand(getEcosystemsCmd)
	getCmd.AddCommand(getEcosystemCmd)
	useCmd.AddCommand(useEcosystemCmd)

	// Check if deleteCmd exists before adding
	if deleteCmd != nil {
		deleteCmd.AddCommand(deleteEcosystemCmd)
	}

	// Ecosystem creation flags
	createEcosystemCmd.Flags().StringVar(&ecosystemDescription, "description", "", "Ecosystem description")
}

// getActiveEcosystem returns the active ecosystem from the context
func getActiveEcosystem(ds db.DataStore) (*models.Ecosystem, error) {
	ctx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if ctx.ActiveEcosystemID == nil {
		return nil, fmt.Errorf("no active ecosystem set")
	}

	ecosystem, err := ds.GetEcosystemByID(*ctx.ActiveEcosystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ecosystem: %w", err)
	}

	return ecosystem, nil
}
