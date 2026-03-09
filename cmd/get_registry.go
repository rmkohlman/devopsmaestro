package cmd

import (
	"context"
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// Registry Resource Commands (dvm get registries, dvm get registry <name>)
// =============================================================================

// getRegistriesCmd lists all registries
var getRegistriesCmd = &cobra.Command{
	Use:     "registries",
	Aliases: []string{"reg", "regs"},
	Short:   "List all registries",
	Long: `List all registries (zot, athens, devpi, verdaccio, squid).

Examples:
  dvm get registries
  dvm get reg                     # Short form
  dvm get registries -o yaml
  dvm get registries -o json
  dvm get registries -o wide      # Show additional columns`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getRegistries(cmd)
	},
}

// getRegistryCmd gets a specific registry by name
var getRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Get a specific registry",
	Long: `Get a specific registry by name.

Examples:
  dvm get registry my-zot
  dvm get reg my-zot              # Short form
  dvm get registry my-zot -o yaml
  dvm get registry my-zot -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getRegistry(cmd, args[0])
	},
}

func getRegistries(cmd *cobra.Command) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	resources, err := resource.List(ctx, handlers.KindRegistry)
	if err != nil {
		return fmt.Errorf("failed to list registries: %w", err)
	}

	if len(resources) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No registries found",
			EmptyHints:   []string{"dvm create registry <name> --type <type>"},
		})
	}

	// Extract underlying registries from resources
	registries := make([]*models.Registry, len(resources))
	for i, res := range resources {
		rr := res.(*handlers.RegistryResource)
		registries[i] = rr.Registry()
	}

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		registriesYAML := make([]models.RegistryYAML, len(registries))
		for i, r := range registries {
			ry := r.ToYAML()
			status := registryLiveStatus(context.Background(), r)
			ry.Status = &models.RegistryStatusYAML{
				State:    status,
				Endpoint: fmt.Sprintf("http://localhost:%d", r.Port),
			}
			registriesYAML[i] = ry
		}
		return render.OutputWith(getOutputFormat, registriesYAML, render.Options{})
	}

	// Determine if wide format
	isWide := getOutputFormat == "wide"

	// For human output, build table data
	var headers []string
	if isWide {
		headers = []string{"NAME", "TYPE", "PORT", "LIFECYCLE", "STATE", "UPTIME", "CREATED"}
	} else {
		headers = []string{"NAME", "TYPE", "PORT", "LIFECYCLE", "STATE", "UPTIME"}
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    make([][]string, len(registries)),
	}

	for i, r := range registries {
		// Live status check via ServiceManager (reads PID file, not just DB)
		status := registryLiveStatus(context.Background(), r)

		// Uptime placeholder (would be from runtime status)
		uptime := "-"
		if status == "running" {
			uptime = "-" // Runtime status would provide actual uptime
		}

		row := []string{
			r.Name,
			r.Type,
			fmt.Sprintf("%d", r.Port),
			r.Lifecycle,
			status,
			uptime,
		}

		if isWide {
			// Add CREATED timestamp
			row = append(row, r.CreatedAt)
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

func getRegistry(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	res, err := resource.Get(ctx, handlers.KindRegistry, name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found: %w", name, err)
	}

	registry := res.(*handlers.RegistryResource).Registry()

	// For JSON/YAML, output the model data directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		ry := registry.ToYAML()
		status := registryLiveStatus(cmd.Context(), registry)
		ry.Status = &models.RegistryStatusYAML{
			State:    status,
			Endpoint: fmt.Sprintf("http://localhost:%d", registry.Port),
		}
		return render.OutputWith(getOutputFormat, ry, render.Options{})
	}

	// For human output, show detail view
	desc := ""
	if registry.Description.Valid {
		desc = registry.Description.String
	}

	status := registryLiveStatus(cmd.Context(), registry)

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: registry.Name},
		render.KeyValue{Key: "Type", Value: registry.Type},
		render.KeyValue{Key: "Port", Value: fmt.Sprintf("%d", registry.Port)},
		render.KeyValue{Key: "Lifecycle", Value: registry.Lifecycle},
		render.KeyValue{Key: "Status", Value: status},
		render.KeyValue{Key: "Description", Value: desc},
		render.KeyValue{Key: "Created", Value: registry.CreatedAt},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Registry Details",
	})
}

// registryLiveStatus checks whether a registry process is actually running
// by creating a ServiceManager and checking the PID file, rather than trusting
// the DB status field which may be stale across CLI invocations.
func registryLiveStatus(ctx context.Context, reg *models.Registry) string {
	factory := registry.NewServiceFactory()
	mgr, err := factory.CreateManager(reg)
	if err != nil {
		// Can't create manager — fall back to DB status
		if reg.Status != "" {
			return reg.Status
		}
		return "stopped"
	}
	if mgr.IsRunning(ctx) {
		return "running"
	}
	return "stopped"
}
