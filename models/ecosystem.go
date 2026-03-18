package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// BuildArgsConfig defines build arguments shared across hierarchy levels
// (Ecosystem and Domain). This is the YAML representation.
type BuildArgsConfig struct {
	Args map[string]string `yaml:"args,omitempty" json:"args,omitempty"`
}

// IsZero implements the yaml.v3 IsZero interface for omitempty support.
// Returns true when no build args are defined.
func (b BuildArgsConfig) IsZero() bool {
	return len(b.Args) == 0
}

// Ecosystem represents the top-level grouping in the object hierarchy.
// It serves as a platform or organizational boundary for domains.
//
// Hierarchy: Ecosystem -> Domain -> App -> Workspace
type Ecosystem struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Theme       sql.NullString `db:"theme" json:"theme,omitempty" yaml:"theme,omitempty"`
	BuildArgs   sql.NullString `db:"build_args" json:"build_args,omitempty" yaml:"-"`
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
	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Theme       string          `yaml:"theme,omitempty" json:"theme,omitempty"`
	Domains     []string        `yaml:"domains,omitempty" json:"domains,omitempty"`
	Build       BuildArgsConfig `yaml:"build,omitempty" json:"build,omitempty"`
}

// ToYAML converts an Ecosystem to YAML format.
// domainNames should contain the names of child domains (pass nil for empty).
func (e *Ecosystem) ToYAML(domainNames []string) EcosystemYAML {
	description := ""
	if e.Description.Valid {
		description = e.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	theme := ""
	if e.Theme.Valid {
		theme = e.Theme.String
	}

	// Restore build args from DB JSON blob if present
	var buildConfig BuildArgsConfig
	if e.BuildArgs.Valid && e.BuildArgs.String != "" {
		var args map[string]string
		if err := json.Unmarshal([]byte(e.BuildArgs.String), &args); err == nil && len(args) > 0 {
			buildConfig.Args = args
		}
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
			Description: description,
			Theme:       theme,
			Domains:     domainNames,
			Build:       buildConfig,
		},
	}
}

// FromYAML converts YAML format to an Ecosystem
func (e *Ecosystem) FromYAML(yaml EcosystemYAML) {
	e.Name = yaml.Metadata.Name

	// Prefer spec.description, fall back to annotations for backward compat
	if yaml.Spec.Description != "" {
		e.Description = sql.NullString{String: yaml.Spec.Description, Valid: true}
	} else if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		e.Description = sql.NullString{String: desc, Valid: true}
	}

	if yaml.Spec.Theme != "" {
		e.Theme = sql.NullString{String: yaml.Spec.Theme, Valid: true}
	}

	// Persist build args as JSON
	if len(yaml.Spec.Build.Args) > 0 {
		if b, err := json.Marshal(yaml.Spec.Build.Args); err == nil {
			e.BuildArgs = sql.NullString{String: string(b), Valid: true}
		}
	}
}
