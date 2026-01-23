package models

import (
	"database/sql"
	"time"
)

// Project represents a project entity in the system.
type Project struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Path        string         `db:"path" json:"path" yaml:"path"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// ProjectYAML represents the YAML serialization format for a project
type ProjectYAML struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   ProjectMetadata `yaml:"metadata"`
	Spec       ProjectSpec     `yaml:"spec"`
}

// ProjectMetadata contains project metadata
type ProjectMetadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// ProjectSpec contains project specification
type ProjectSpec struct {
	Path       string   `yaml:"path"`
	Workspaces []string `yaml:"workspaces,omitempty"`
}

// ToYAML converts a Project to YAML format
func (p *Project) ToYAML() ProjectYAML {
	description := ""
	if p.Description.Valid {
		description = p.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	return ProjectYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Project",
		Metadata: ProjectMetadata{
			Name:        p.Name,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: ProjectSpec{
			Path:       p.Path,
			Workspaces: []string{},
		},
	}
}

// FromYAML converts YAML format to a Project
func (p *Project) FromYAML(yaml ProjectYAML) {
	p.Name = yaml.Metadata.Name
	p.Path = yaml.Spec.Path

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		p.Description = sql.NullString{String: desc, Valid: true}
	}
}
