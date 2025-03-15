package models

import "time"

// Orchestration represents a higher-level orchestration of multiple pipelines.
type Orchestration struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Pipelines   []Pipeline `db:"-" json:"pipelines"` // Ignored by DB, used in JSON
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
