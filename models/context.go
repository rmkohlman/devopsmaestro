package models

import "time"

// Context represents the active selection state for the object hierarchy (kubectl-style).
// The hierarchy is: Ecosystem -> Domain -> App -> Workspace
type Context struct {
	ID                int       `db:"id" json:"id"`
	ActiveEcosystemID *int      `db:"active_ecosystem_id" json:"active_ecosystem_id,omitempty"`
	ActiveDomainID    *int      `db:"active_domain_id" json:"active_domain_id,omitempty"`
	ActiveAppID       *int      `db:"active_app_id" json:"active_app_id,omitempty"`
	ActiveWorkspaceID *int      `db:"active_workspace_id" json:"active_workspace_id,omitempty"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
