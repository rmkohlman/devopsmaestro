package models

import "time"

// Context represents the active selection state for the object hierarchy (kubectl-style).
// The hierarchy is: Ecosystem -> Domain -> App -> Workspace
//
// Note: ActiveProjectID is deprecated and will be removed in v0.9.0 when
// Workspace migrates from Project to App.
type Context struct {
	ID                int       `db:"id" json:"id"`
	ActiveEcosystemID *int      `db:"active_ecosystem_id" json:"active_ecosystem_id,omitempty"`
	ActiveDomainID    *int      `db:"active_domain_id" json:"active_domain_id,omitempty"`
	ActiveAppID       *int      `db:"active_app_id" json:"active_app_id,omitempty"`
	ActiveWorkspaceID *int      `db:"active_workspace_id" json:"active_workspace_id,omitempty"`
	ActiveProjectID   *int      `db:"active_project_id" json:"active_project_id,omitempty"` // DEPRECATED: use ActiveAppID
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
