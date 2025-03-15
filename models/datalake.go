package models

import "time"

// DataLake represents a global data store available across the system.
type DataLake struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Data      string    `db:"data" json:"data"` // Serialized data
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
