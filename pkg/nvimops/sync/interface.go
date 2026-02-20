package sync

import "context"

// SourceHandler defines the interface for syncing plugins from external sources.
// Implementations handle specific sources like LazyVim, AstroNvim, NvChad, etc.
// This follows the Strategy pattern - each source has its own implementation.
type SourceHandler interface {
	// Name returns the unique identifier for this source (e.g., "lazyvim", "astronvim")
	Name() string

	// Description returns a human-readable description of the source
	Description() string

	// Sync imports plugins from the external source based on the provided options.
	// Returns a SyncResult with details about what was created, updated, or failed.
	Sync(ctx context.Context, options SyncOptions) (*SyncResult, error)

	// ListAvailable returns all plugins available from this source.
	// Used for discovery and filtering before sync.
	ListAvailable(ctx context.Context) ([]AvailablePlugin, error)

	// Validate checks if the source is accessible and properly configured.
	// Returns nil if the source is ready to use.
	Validate(ctx context.Context) error
}

// SourceHandlerFactory creates SourceHandler instances for different external sources.
// This allows the sync system to be extended with new sources without modifying core code.
type SourceHandlerFactory interface {
	// CreateHandler creates a handler for the specified source name.
	// Returns an error if the source is not supported.
	CreateHandler(sourceName string) (SourceHandler, error)

	// ListSources returns all available source names that can be created.
	ListSources() []string

	// IsSupported checks if a source name is supported by this factory.
	IsSupported(sourceName string) bool

	// GetHandlerInfo returns metadata about a supported source.
	GetHandlerInfo(sourceName string) (*SourceInfo, error)
}

// SourceInfo contains metadata about an external plugin source.
type SourceInfo struct {
	// Name is the unique identifier for the source
	Name string

	// Description is a human-readable description
	Description string

	// URL is the primary URL for the source (GitHub repo, website, etc.)
	URL string

	// Type indicates the type of source (github, local, remote, etc.)
	Type string

	// RequiresAuth indicates if the source requires authentication
	RequiresAuth bool

	// ConfigKeys lists configuration keys that can be set for this source
	ConfigKeys []string
}

// SourceType represents the type of external source.
type SourceType string

const (
	// SourceTypeGitHub represents sources hosted on GitHub
	SourceTypeGitHub SourceType = "github"

	// SourceTypeLocal represents local filesystem sources
	SourceTypeLocal SourceType = "local"

	// SourceTypeRemote represents remote HTTP sources
	SourceTypeRemote SourceType = "remote"

	// SourceTypeRegistry represents plugin registries
	SourceTypeRegistry SourceType = "registry"
)

// HandlerRegistration contains information needed to register a source handler.
type HandlerRegistration struct {
	// Name is the unique identifier for the source
	Name string

	// Info contains metadata about the source
	Info SourceInfo

	// CreateFunc is the factory function that creates handler instances
	CreateFunc func() SourceHandler
}
