package models

import "time"

// Workspace represents a workspace entity associated with a project.
type Workspace struct {
	ID          int       `db:"id" json:"id"`
	ProjectID   int       `db:"project_id" json:"project_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
