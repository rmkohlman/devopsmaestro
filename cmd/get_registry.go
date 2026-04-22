package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

// registryStatusTimeout bounds the time spent fetching live status for a
// single registry during listing. Status checks (PID file + signal probe)
// are fast, but we still bound them to avoid pathological hangs (issue #398).
const registryStatusTimeout = 2 * time.Second

// listVersion returns a fast, non-blocking version string for use in the
// `get registries` table. It avoids shelling out to npm/pipx/brew/athens
// (which can take several seconds each — issue #398). Use DetectVersion
// only on single-resource views or after explicit install/update.
func listVersion(factory *registry.ServiceFactory, r *models.Registry) string {
	if r.Version != "" {
		return r.Version
	}
	if s, err := factory.GetStrategy(r.Type); err == nil {
		if v := s.GetDefaultVersion(); v != "" {
			return v
		}
	}
	return ""
}

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

	// For JSON/YAML, produce a kind: List envelope (issue #154)
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		if len(resources) == 0 {
			return render.OutputWith(getOutputFormat, resource.NewResourceList(), render.Options{})
		}
		list := resource.NewResourceList()
		regs := make([]*models.Registry, len(resources))
		yamls := make([]models.RegistryYAML, len(resources))
		for i, res := range resources {
			regs[i] = res.(*handlers.RegistryResource).Registry()
			yamls[i] = regs[i].ToYAML()
		}
		statuses := make([]string, len(regs))
		var wg sync.WaitGroup
		for i, reg := range regs {
			wg.Add(1)
			go func(i int, reg *models.Registry) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), registryStatusTimeout)
				defer cancel()
				statuses[i] = registryLiveStatus(ctx, reg)
			}(i, reg)
		}
		wg.Wait()
		for i := range yamls {
			yamls[i].Status = &models.RegistryStatusYAML{
				State:    statuses[i],
				Endpoint: fmt.Sprintf("http://localhost:%d", regs[i].Port),
			}
			list.Items = append(list.Items, yamls[i])
		}
		return render.OutputWith(getOutputFormat, list, render.Options{})
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

	// Determine if wide format
	isWide := getOutputFormat == "wide"

	// For human output, build table data
	headers := getRegistriesTableHeaders(isWide)

	tableData := render.TableData{
		Headers: headers,
	}

	factory := registry.NewServiceFactory()

	// Fan out per-registry status / version / uptime collection. Version
	// detection for npm/pipx/brew/athens-managed registries shells out to
	// package managers and can take seconds each — running serially across
	// N registries was the cause of issue #398 (24s+ on first run). Each
	// goroutine has a bounded context so a single slow registry cannot
	// stall the whole listing.
	tableData.Rows = make([][]string, len(registries))
	var wg sync.WaitGroup
	for i, r := range registries {
		wg.Add(1)
		go func(i int, r *models.Registry) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), registryStatusTimeout)
			defer cancel()

			status := registryLiveStatus(ctx, r)

			version := listVersion(factory, r)

			uptime := "-"
			if status == "running" {
				if d := factory.GetUptime(r); d > 0 {
					uptime = formatDuration(d)
				}
			}

			row := []string{
				r.Name,
				r.Type,
				version,
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
		}(i, r)
	}
	wg.Wait()

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
		render.KeyValue{Key: "Version", Value: registry.Version},
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

// getRegistriesTableHeaders returns the table headers for registry list output.
// CC-3: VERSION column is positioned after TYPE.
func getRegistriesTableHeaders(wide bool) []string {
	if wide {
		return []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME", "CREATED"}
	}
	return []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME"}
}

// getRegistryDetailViewKeys returns the ordered keys for registry detail view.
// CC-5: Version key is positioned after Type.
func getRegistryDetailViewKeys() []string {
	return []string{"Name", "Type", "Version", "Port", "Lifecycle", "Status", "Description", "Created"}
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

// formatDuration formats a time.Duration into a human-readable string.
// Examples: "5s", "3m", "2h", "1d", "3d 2h".
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	remainHours := hours % 24
	if remainHours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, remainHours)
}
