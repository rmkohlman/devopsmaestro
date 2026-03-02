package models

import (
	"database/sql"
	"time"
)

// CustomResourceDefinition represents a CRD in the database
type CustomResourceDefinition struct {
	ID         int            `db:"id" json:"id"`
	Kind       string         `db:"kind" json:"kind"`
	Group      string         `db:"group" json:"group"`
	Singular   string         `db:"singular" json:"singular"`
	Plural     string         `db:"plural" json:"plural"`
	ShortNames sql.NullString `db:"short_names" json:"short_names,omitempty"` // JSON array
	Scope      string         `db:"scope" json:"scope"`
	Versions   sql.NullString `db:"versions" json:"versions"` // JSON array
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

// CRDVersion represents a version of a CRD schema
type CRDVersion struct {
	Name    string                 `json:"name"`
	Served  bool                   `json:"served"`
	Storage bool                   `json:"storage"`
	Schema  map[string]interface{} `json:"schema"`
}

// CRDYAML represents the YAML format for CustomResourceDefinition
type CRDYAML struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   CRDMetadata `yaml:"metadata"`
	Spec       CRDSpec     `yaml:"spec"`
}

// CRDMetadata contains CRD metadata
type CRDMetadata struct {
	Name string `yaml:"name"`
}

// CRDSpec contains the CRD specification
type CRDSpec struct {
	Group    string       `yaml:"group"`
	Names    CRDNames     `yaml:"names"`
	Scope    string       `yaml:"scope"`
	Versions []CRDVersion `yaml:"versions"`
}

// CRDNames defines the naming options for a CRD
type CRDNames struct {
	Kind       string   `yaml:"kind"`
	Singular   string   `yaml:"singular"`
	Plural     string   `yaml:"plural"`
	ShortNames []string `yaml:"shortNames,omitempty"`
}

// GetKind returns "CustomResourceDefinition"
func (c *CustomResourceDefinition) GetKind() string {
	return "CustomResourceDefinition"
}

// GetName returns the Kind (which is the unique identifier)
func (c *CustomResourceDefinition) GetName() string {
	return c.Kind
}

// Validate checks if the CRD is valid
func (c *CustomResourceDefinition) Validate() error {
	// Validation will be implemented during GREEN phase
	return nil
}
