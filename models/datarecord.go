package models

import "time"

// DataRecord represents a record of data stored within a workspace.
type DataRecord struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Data        string    `db:"data" json:"data"` // Serialized data
	WorkspaceID int       `db:"workspace_id" json:"workspace_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
