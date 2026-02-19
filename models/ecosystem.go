package models

import (
	"database/sql"
	"time"
)

// Ecosystem represents the top-level grouping in the object hierarchy.
// It serves as a platform or organizational boundary for domains.
//
// Hierarchy: Ecosystem -> Domain -> App -> Workspace
type Ecosystem struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Theme       sql.NullString `db:"theme" json:"theme,omitempty" yaml:"theme,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// EcosystemYAML represents the YAML serialization format for an ecosystem
type EcosystemYAML struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   EcosystemMetadata `yaml:"metadata"`
	Spec       EcosystemSpec     `yaml:"spec"`
}

// EcosystemMetadata contains ecosystem metadata
type EcosystemMetadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// EcosystemSpec contains ecosystem specification
type EcosystemSpec struct {
	Domains []string `yaml:"domains,omitempty"`
}

// ToYAML converts an Ecosystem to YAML format
func (e *Ecosystem) ToYAML() EcosystemYAML {
	description := ""
	if e.Description.Valid {
		description = e.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	return EcosystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Ecosystem",
		Metadata: EcosystemMetadata{
			Name:        e.Name,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: EcosystemSpec{
			Domains: []string{},
		},
	}
}

// FromYAML converts YAML format to an Ecosystem
func (e *Ecosystem) FromYAML(yaml EcosystemYAML) {
	e.Name = yaml.Metadata.Name

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		e.Description = sql.NullString{String: desc, Valid: true}
	}
}
