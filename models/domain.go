package models

import (
	"database/sql"
	"time"
)

// Domain represents a bounded context within an ecosystem.
// It serves as an organizational grouping for related apps.
//
// Hierarchy: Ecosystem -> Domain -> App -> Workspace
type Domain struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	EcosystemID int            `db:"ecosystem_id" json:"ecosystem_id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// DomainYAML represents the YAML serialization format for a domain
type DomainYAML struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   DomainMetadata `yaml:"metadata"`
	Spec       DomainSpec     `yaml:"spec"`
}

// DomainMetadata contains domain metadata
type DomainMetadata struct {
	Name        string            `yaml:"name"`
	Ecosystem   string            `yaml:"ecosystem"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// DomainSpec contains domain specification
type DomainSpec struct {
	Apps []string `yaml:"apps,omitempty"`
}

// ToYAML converts a Domain to YAML format
func (d *Domain) ToYAML(ecosystemName string) DomainYAML {
	description := ""
	if d.Description.Valid {
		description = d.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	return DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata: DomainMetadata{
			Name:        d.Name,
			Ecosystem:   ecosystemName,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: DomainSpec{
			Apps: []string{},
		},
	}
}

// FromYAML converts YAML format to a Domain
func (d *Domain) FromYAML(yaml DomainYAML) {
	d.Name = yaml.Metadata.Name

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		d.Description = sql.NullString{String: desc, Valid: true}
	}
}
