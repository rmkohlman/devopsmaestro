package db

import (
	"devopsmaestro/models"
)

// DataStore is the interface for higher-level application operations that interact with the database.
type DataStore interface {
	CreateProject(project *models.Project) error
	GetProjectByName(name string) (*models.Project, error)
	UpdateProject(project *models.Project) error
	ListProjects() ([]*models.Project, error)
	// Other methods like DeleteProject, ListDependencies, etc.
}
