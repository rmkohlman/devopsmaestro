package models

import (
	"database/sql"
	"time"
)

// RegistryHistory represents a registry configuration history entry
type RegistryHistory struct {
	ID               int64
	RegistryID       int
	Revision         int
	Config           string // JSON snapshot
	Enabled          bool
	Lifecycle        string
	Port             int
	Storage          string
	IdleTimeout      sql.NullInt64
	Action           string // 'start', 'restart', 'config_change', 'rollback'
	Status           string // 'success', 'failed', 'in_progress'
	User             sql.NullString
	ErrorMessage     sql.NullString
	PreviousRevision sql.NullInt64
	RegistryVersion  sql.NullString
	CreatedAt        time.Time
	CompletedAt      sql.NullTime
}
