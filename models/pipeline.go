package models

import "time"

// Pipeline represents a pipeline consisting of multiple workflows.
type Pipeline struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Workflows   []Workflow `db:"-" json:"workflows"` // Ignored by DB, used in JSON
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
