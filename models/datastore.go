package models

import "time"

// DataStore represents a data store specific to a project or workspace.
type DataStore struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Data        string    `db:"data" json:"data"` // Serialized data
	ProjectID   int       `db:"project_id" json:"project_id"`
	WorkspaceID int       `db:"workspace_id" json:"workspace_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
