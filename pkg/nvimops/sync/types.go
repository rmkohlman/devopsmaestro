// Package sync provides a strategy pattern for syncing plugins from external sources.
// This allows DevOpsMaestro to import plugins from LazyVim, AstroNvim, NvChad, and other
// Neovim configuration frameworks.
package sync

import (
	"fmt"
)

// PackageCreator defines an interface for creating packages during sync operations.
// This allows different implementations (file-based, database-based, etc.).
type PackageCreator interface {
	// CreatePackage creates or updates a package with the given plugins.
	// sourceName is the name of the sync source (e.g., "lazyvim").
	// plugins is a list of plugin names that should be included in the package.
	CreatePackage(sourceName string, plugins []string) error
}

// SyncOptions contains configuration for syncing plugins from external sources.
// Use the Builder pattern via NewSyncOptions() to create instances.
type SyncOptions struct {
	// DryRun indicates whether to simulate the sync without writing files
	DryRun bool

	// Filters contains key-value pairs to filter plugins during sync.
	// Common filters: category=lang, tag=lsp, enabled=true
	Filters map[string]string

	// TargetDir is the directory where YAML plugin definitions will be written.
	// Defaults to ~/.config/nvp/plugins/
	TargetDir string

	// Overwrite indicates whether to replace existing plugin definitions
	Overwrite bool

	// PackageCreator is an optional interface for creating packages during sync.
	// If nil, no packages will be created.
	PackageCreator PackageCreator
}

// SyncResult contains the results of a sync operation from an external source.
type SyncResult struct {
	// PluginsCreated contains names of plugin YAML files that were created
	PluginsCreated []string

	// PluginsUpdated contains names of plugin YAML files that were updated
	PluginsUpdated []string

	// PackagesCreated contains names of package YAML files that were created
	// (packages group related plugins together)
	PackagesCreated []string

	// PackagesUpdated contains names of package YAML files that were updated
	PackagesUpdated []string

	// Errors contains any errors that occurred during sync
	Errors []error

	// SourceName is the name of the source that was synced
	SourceName string

	// TotalAvailable is the total number of plugins available from the source
	TotalAvailable int

	// TotalSynced is the number of plugins that were successfully synced
	TotalSynced int
}

// AvailablePlugin represents a plugin that can be imported from an external source.
type AvailablePlugin struct {
	// Name is the plugin identifier (used for YAML filename)
	Name string

	// Description is a human-readable description of the plugin
	Description string

	// Category is the plugin category (lang, navigation, ui, etc.)
	Category string

	// Repo is the GitHub repository (e.g., "nvim-telescope/telescope.nvim")
	Repo string

	// Labels contains additional metadata about the plugin.
	// Common labels: enabled, lazy, priority, framework
	Labels map[string]string

	// Config contains the Lua configuration code (if available)
	Config string

	// Dependencies lists other plugins this plugin depends on
	Dependencies []string

	// SourceName is the name of the source this plugin came from
	SourceName string
}

// SyncOptionsBuilder provides a fluent interface for building SyncOptions.
// Use NewSyncOptions() to create a builder instance.
type SyncOptionsBuilder struct {
	options SyncOptions
}

// NewSyncOptions creates a new SyncOptionsBuilder with default values.
func NewSyncOptions() *SyncOptionsBuilder {
	return &SyncOptionsBuilder{
		options: SyncOptions{
			DryRun:    false,
			Filters:   make(map[string]string),
			TargetDir: "", // Will use default if empty
			Overwrite: false,
		},
	}
}

// DryRun sets whether to simulate the sync without writing files.
func (b *SyncOptionsBuilder) DryRun(dryRun bool) *SyncOptionsBuilder {
	b.options.DryRun = dryRun
	return b
}

// WithFilter adds a filter to apply during sync.
// Common filters: category=lang, tag=lsp, enabled=true
func (b *SyncOptionsBuilder) WithFilter(key, value string) *SyncOptionsBuilder {
	if b.options.Filters == nil {
		b.options.Filters = make(map[string]string)
	}
	b.options.Filters[key] = value
	return b
}

// WithFilters sets multiple filters at once.
func (b *SyncOptionsBuilder) WithFilters(filters map[string]string) *SyncOptionsBuilder {
	if b.options.Filters == nil {
		b.options.Filters = make(map[string]string)
	}
	for k, v := range filters {
		b.options.Filters[k] = v
	}
	return b
}

// WithTargetDir sets the directory where YAML files will be written.
func (b *SyncOptionsBuilder) WithTargetDir(dir string) *SyncOptionsBuilder {
	b.options.TargetDir = dir
	return b
}

// Overwrite sets whether to replace existing plugin definitions.
func (b *SyncOptionsBuilder) Overwrite(overwrite bool) *SyncOptionsBuilder {
	b.options.Overwrite = overwrite
	return b
}

// WithDataStore sets the DataStore for package operations.
// Optional - if not set, package creation will be skipped.
func (b *SyncOptionsBuilder) WithPackageCreator(creator PackageCreator) *SyncOptionsBuilder {
	b.options.PackageCreator = creator
	return b
}

// Build returns the configured SyncOptions.
func (b *SyncOptionsBuilder) Build() SyncOptions {
	return b.options
}

// HasFilter checks if a specific filter is set.
func (opts SyncOptions) HasFilter(key string) bool {
	_, exists := opts.Filters[key]
	return exists
}

// GetFilter returns the value of a filter, or empty string if not set.
func (opts SyncOptions) GetFilter(key string) string {
	return opts.Filters[key]
}

// MatchesFilter checks if a plugin matches the specified filter.
// This is a utility method for source implementations.
func (opts SyncOptions) MatchesFilter(key, value string) bool {
	if !opts.HasFilter(key) {
		return true // No filter means all match
	}
	return opts.GetFilter(key) == value
}

// MatchesAvailablePlugin checks if an AvailablePlugin matches all filters.
func (opts SyncOptions) MatchesAvailablePlugin(plugin AvailablePlugin) bool {
	for key, expectedValue := range opts.Filters {
		var actualValue string

		switch key {
		case "category":
			actualValue = plugin.Category
		case "name":
			actualValue = plugin.Name
		case "source":
			actualValue = plugin.SourceName
		default:
			// Check in labels
			actualValue = plugin.Labels[key]
		}

		if actualValue != expectedValue {
			return false
		}
	}
	return true
}

// AddError is a convenience method to add an error to SyncResult.
func (r *SyncResult) AddError(err error) {
	if r.Errors == nil {
		r.Errors = make([]error, 0)
	}
	r.Errors = append(r.Errors, err)
}

// HasErrors returns true if any errors occurred during sync.
func (r *SyncResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AddPluginCreated adds a plugin to the created list.
func (r *SyncResult) AddPluginCreated(name string) {
	if r.PluginsCreated == nil {
		r.PluginsCreated = make([]string, 0)
	}
	r.PluginsCreated = append(r.PluginsCreated, name)
	r.TotalSynced++
}

// AddPluginUpdated adds a plugin to the updated list.
func (r *SyncResult) AddPluginUpdated(name string) {
	if r.PluginsUpdated == nil {
		r.PluginsUpdated = make([]string, 0)
	}
	r.PluginsUpdated = append(r.PluginsUpdated, name)
	r.TotalSynced++
}

// AddPackageCreated adds a package to the created list.
func (r *SyncResult) AddPackageCreated(name string) {
	if r.PackagesCreated == nil {
		r.PackagesCreated = make([]string, 0)
	}
	r.PackagesCreated = append(r.PackagesCreated, name)
}

// AddPackageUpdated adds a package to the updated list.
func (r *SyncResult) AddPackageUpdated(name string) {
	if r.PackagesUpdated == nil {
		r.PackagesUpdated = make([]string, 0)
	}
	r.PackagesUpdated = append(r.PackagesUpdated, name)
}

// Summary returns a human-readable summary of the sync result.
func (r *SyncResult) Summary() string {
	if r.TotalSynced == 0 {
		return "No plugins synced"
	}

	created := len(r.PluginsCreated)
	updated := len(r.PluginsUpdated)

	summary := ""
	if created > 0 {
		summary += fmt.Sprintf("%d created", created)
	}
	if updated > 0 {
		if summary != "" {
			summary += ", "
		}
		summary += fmt.Sprintf("%d updated", updated)
	}

	return summary
}
