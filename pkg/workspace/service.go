package workspace

import (
	"database/sql"
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimbridge"
)

// HierarchyReader provides read access to the entity hierarchy needed for
// slug generation during workspace creation.
type HierarchyReader interface {
	GetAppByID(id int) (*models.App, error)
	GetDomainByID(id int) (*models.Domain, error)
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
		domain, err := hierarchy.GetDomainByID(app.DomainID)
		if err != nil {
			return fmt.Errorf("failed to get domain for slug generation: %w", err)
		}
		ecosystem, err := hierarchy.GetEcosystemByID(domain.EcosystemID)
		if err != nil {
			return fmt.Errorf("failed to get ecosystem for slug generation: %w", err)
		}
		workspace.Slug = GenerateSlug(ecosystem.Name, domain.Name, app.Name, workspace.Name)
	}

	return nil
}
