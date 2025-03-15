package models

import "time"

// Context represents configuration settings for different entities in the system.
type Context struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Config      string    `db:"config" json:"config"` // Serialized configuration
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
