package workspace

import (
	"database/sql"
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimbridge"
)

// HierarchyReader provides read-only access to the entity hierarchy needed for
// workspace slug generation and default propagation during workspace creation.
// It is a narrow sub-interface of db.DataStore that covers only the
// lookup methods required by PrepareDefaults.
//
// Implemented by db.SQLDataStore (and db.MockDataStore in tests).
// Inject this interface in unit tests to avoid a full DataStore dependency.
type HierarchyReader interface {
	// GetAppByID retrieves an app by its primary key.
	// Returns an error if the app does not exist.
	GetAppByID(id int) (*models.App, error)

	// GetSystemByID retrieves a system by its primary key.
	// Returns an error if the system does not exist.
	GetSystemByID(id int) (*models.System, error)

	// GetDomainByID retrieves a domain by its primary key.
	// Returns an error if the domain does not exist.
	GetDomainByID(id int) (*models.Domain, error)

	// GetEcosystemByID retrieves an ecosystem by its primary key.
	// Returns an error if the ecosystem does not exist.
	GetEcosystemByID(id int) (*models.Ecosystem, error)
}

// PrepareDefaults applies business defaults to a workspace before persistence.
//   - Sets NvimStructure to the default nvim config if not already specified.
//   - Generates a hierarchical slug if one is not already set.
//
// This must be called before passing the workspace to the store's CreateWorkspace.
func PrepareDefaults(workspace *models.Workspace, hierarchy HierarchyReader) error {
	// Apply default nvim config if not specified
	if !workspace.NvimStructure.Valid || workspace.NvimStructure.String == "" {
		defaultConfig := nvimbridge.DefaultNvimConfig()
		workspace.NvimStructure = sql.NullString{
			String: defaultConfig.Structure,
			Valid:  true,
		}
	}

	// Generate slug if not provided
	if workspace.Slug == "" {
		app, err := hierarchy.GetAppByID(workspace.AppID)
		if err != nil {
			return fmt.Errorf("failed to get app for slug generation: %w", err)
		}

		// Look up system name if the app belongs to one
		systemName := ""
		if app.SystemID.Valid {
			system, err := hierarchy.GetSystemByID(int(app.SystemID.Int64))
			if err != nil {
				return fmt.Errorf("failed to get system for slug generation: %w", err)
			}
			systemName = system.Name
		}

		domain, err := hierarchy.GetDomainByID(app.DomainID)
		if err != nil {
			return fmt.Errorf("failed to get domain for slug generation: %w", err)
		}
		ecosystem, err := hierarchy.GetEcosystemByID(domain.EcosystemID)
		if err != nil {
			return fmt.Errorf("failed to get ecosystem for slug generation: %w", err)
		}
		workspace.Slug = GenerateSlug(ecosystem.Name, domain.Name, systemName, app.Name, workspace.Name)
	}

	return nil
}
