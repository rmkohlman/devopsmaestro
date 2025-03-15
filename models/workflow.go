package models

import "time"

// Workflow represents a sequence of tasks.
type Workflow struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Tasks       []Task    `db:"-" json:"tasks"` // Ignored by DB, used in JSON
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
