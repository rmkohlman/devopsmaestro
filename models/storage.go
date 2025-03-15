package models

import "time"

// Storage represents a volume or storage entity in the system.
type Storage struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Type        string    `db:"type" json:"type"`     // e.g., "volume", "bucket"
	Config      string    `db:"config" json:"config"` // Serialized configuration
	WorkspaceID int       `db:"workspace_id" json:"workspace_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
