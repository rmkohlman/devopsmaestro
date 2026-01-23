package models

import "time"

// Context represents the active project/workspace context (kubectl-style)
type Context struct {
	ID                int       `db:"id" json:"id"`
	ActiveProjectID   *int      `db:"active_project_id" json:"active_project_id"`
	ActiveWorkspaceID *int      `db:"active_workspace_id" json:"active_workspace_id"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
