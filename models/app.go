package models

import (
	"database/sql"
	"time"
)

// App represents a codebase/application within a domain.
// The App is "the thing you build and run". It has a path to source code
// and can run in dev mode (Workspace) or live mode (managed by Operator).
//
// Hierarchy: Ecosystem -> Domain -> App -> Workspace
type App struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	DomainID    int            `db:"domain_id" json:"domain_id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Path        string         `db:"path" json:"path" yaml:"path"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// AppYAML represents the YAML serialization format for an app
type AppYAML struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   AppMetadata `yaml:"metadata"`
	Spec       AppSpec     `yaml:"spec"`
}

// AppMetadata contains app metadata
type AppMetadata struct {
	Name        string            `yaml:"name"`
	Domain      string            `yaml:"domain"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// AppSpec contains app specification
type AppSpec struct {
	Path       string   `yaml:"path"`
	Workspaces []string `yaml:"workspaces,omitempty"`
}

// ToYAML converts an App to YAML format
func (a *App) ToYAML(domainName string) AppYAML {
	description := ""
	if a.Description.Valid {
		description = a.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	return AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:        a.Name,
			Domain:      domainName,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: AppSpec{
			Path:       a.Path,
			Workspaces: []string{},
		},
	}
}

// FromYAML converts YAML format to an App
func (a *App) FromYAML(yaml AppYAML) {
	a.Name = yaml.Metadata.Name
	a.Path = yaml.Spec.Path

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		a.Description = sql.NullString{String: desc, Valid: true}
	}
}
