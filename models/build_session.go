package models

import (
	"database/sql"
	"time"
)

// BuildSession represents a build session that tracks a batch of workspace builds.
type BuildSession struct {
	ID              string
	StartedAt       time.Time
	CompletedAt     sql.NullTime
	Status          string // running, completed, failed, cancelled
	TotalWorkspaces int
	Succeeded       int
	Failed          int
	CreatedAt       time.Time
}

// BuildSessionWorkspace represents a single workspace within a build session.
type BuildSessionWorkspace struct {
	ID              int
	SessionID       string
	WorkspaceID     int
	Status          string // queued, building, succeeded, failed
	StartedAt       sql.NullTime
	CompletedAt     sql.NullTime
	DurationSeconds sql.NullInt64
	ImageTag        sql.NullString
	ErrorMessage    sql.NullString
}
