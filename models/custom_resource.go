package models

import (
	"database/sql"
	"time"
)

// CustomResource represents an instance of a custom resource
type CustomResource struct {
	ID        int            `db:"id" json:"id"`
	Kind      string         `db:"kind" json:"kind"`
	Name      string         `db:"name" json:"name"`
	Namespace sql.NullString `db:"namespace" json:"namespace,omitempty"`
	Spec      sql.NullString `db:"spec" json:"spec"`     // JSON
	Status    sql.NullString `db:"status" json:"status"` // JSON
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

// CustomResourceYAML represents the YAML format for custom resources
type CustomResourceYAML struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
	Status     map[string]interface{} `yaml:"status,omitempty"`
}

// GetKind returns the custom resource kind
func (c *CustomResource) GetKind() string {
	return c.Kind
}

// GetName returns the custom resource name
func (c *CustomResource) GetName() string {
	return c.Name
}

// Validate checks if the custom resource is valid
func (c *CustomResource) Validate() error {
	// Validation will be implemented during GREEN phase
	return nil
}
