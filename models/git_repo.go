package models

import (
	"database/sql"
	"time"
)

// GitRepoDB represents a git repository configuration in the database.
type GitRepoDB struct {
	ID                  int            `db:"id"`
	Name                string         `db:"name"`
	URL                 string         `db:"url"`
	Slug                string         `db:"slug"`
	DefaultRef          string         `db:"default_ref"`
	AuthType            string         `db:"auth_type"`
	CredentialID        sql.NullInt64  `db:"credential_id"`
	AutoSync            bool           `db:"auto_sync"`
	SyncIntervalMinutes int            `db:"sync_interval_minutes"`
	LastSyncedAt        sql.NullTime   `db:"last_synced_at"`
	SyncStatus          string         `db:"sync_status"`
	SyncError           sql.NullString `db:"sync_error"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
}
