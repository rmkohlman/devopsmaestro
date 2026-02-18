package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"

	"github.com/spf13/cobra"
)

// completeResources is a generic helper that uses the Resource/Handler pattern
// to provide dynamic completions for resource names with descriptions.
func completeResources(cmd *cobra.Command, kind string) ([]string, cobra.ShellCompDirective) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	// Try to get datastore from context first (normal command context)
	ctx := cmd.Context()
	if dataStore := ctx.Value("dataStore"); dataStore != nil {
		if ds, ok := dataStore.(db.DataStore); ok {
			return getResourceCompletions(ds, kind)
		}
	}

	// For shell completion context, try to create a datastore with default config
	// This may fail if configuration isn't available, in which case we return empty
	createdDS, err := createDefaultDataStore()
	if err != nil {
		// Return empty on error, don't block completion
		return []string{}, cobra.ShellCompDirectiveDefault
	}
	defer func() {
		if closer, ok := createdDS.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	return getResourceCompletions(createdDS, kind)
}

// createDefaultDataStore creates a DataStore using default configuration
func createDefaultDataStore() (db.DataStore, error) {
	// Try the default factory first
	factory := db.NewDataStoreFactory()
	if ds, err := factory.Create(); err == nil {
		return ds, nil
	}

	// If that fails, try with explicit SQLite configuration
	cfg := db.DriverConfig{
		Type: db.DriverSQLite,
		Path: "~/.devopsmaestro/devopsmaestro.db", // Default path
	}

	driver, err := db.NewSQLiteDriver(cfg)
	if err != nil {
		return nil, err
	}

	if err := driver.Connect(); err != nil {
		return nil, err
	}

	return db.NewSQLDataStore(driver, nil), nil
}

// getResourceCompletions gets completions for a specific resource kind
func getResourceCompletions(ds db.DataStore, kind string) ([]string, cobra.ShellCompDirective) {
	// Get handler from registry
	handler := resource.GetHandler(kind)
	if handler == nil {
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	// Use Resource/Handler pattern to list resources
	resourceCtx := resource.Context{DataStore: ds}
	resources, err := handler.List(resourceCtx)
	if err != nil {
		// Return empty on error, don't block completion
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	// Convert to completion format with descriptions
	completions := make([]string, 0, len(resources))
	for _, res := range resources {
		name := res.GetName()
		desc := extractResourceDescription(res)
		if desc != "" {
			completions = append(completions, fmt.Sprintf("%s\t%s", name, desc))
		} else {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// extractResourceDescription extracts a description from different resource types.
// Returns empty string if no description is available.
func extractResourceDescription(res resource.Resource) string {
	switch r := res.(type) {
	case interface{ App() *models.App }:
		app := r.App()
		if app.Description.Valid {
			return app.Description.String
		}
	case interface{ Workspace() *models.Workspace }:
		workspace := r.Workspace()
		if workspace.Description.Valid {
			return workspace.Description.String
		}
	case interface{ Domain() *models.Domain }:
		domain := r.Domain()
		if domain.Description.Valid {
			return domain.Description.String
		}
	case interface{ Ecosystem() *models.Ecosystem }:
		ecosystem := r.Ecosystem()
		if ecosystem.Description.Valid {
			return ecosystem.Description.String
		}
	}
	return ""
}

// Specific completion functions for each resource type

func completeEcosystems(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Ecosystem")
}

func completeDomains(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Domain")
}

func completeApps(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "App")
}

func completeWorkspaces(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeResources(cmd, "Workspace")
}

// registerHierarchyFlagCompletions registers completion functions for hierarchy flags.
// Call this from command init() functions after AddHierarchyFlags().
func registerHierarchyFlagCompletions(cmd *cobra.Command) {
	cmd.RegisterFlagCompletionFunc("ecosystem", completeEcosystems)
	cmd.RegisterFlagCompletionFunc("domain", completeDomains)
	cmd.RegisterFlagCompletionFunc("app", completeApps)
	cmd.RegisterFlagCompletionFunc("workspace", completeWorkspaces)
}
