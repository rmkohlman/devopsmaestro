package models

import "time"

// Dependency represents a dependency entity that can be associated with a workspace or a project.
type Dependency struct {
	ID          int       `db:"id" json:"id"`
	WorkspaceID *int      `db:"workspace_id,omitempty" json:"workspace_id,omitempty"` // Nullable, used if specific to a workspace
	ProjectID   int       `db:"project_id" json:"project_id"`                         // Always required, links to the project
	Name        string    `db:"name" json:"name"`
	Type        string    `db:"type" json:"type"`     // e.g., database, service, etc.
	Config      string    `db:"config" json:"config"` // Serialized configuration (e.g., JSON or YAML)
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
