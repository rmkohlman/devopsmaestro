package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// System represents an organizational grouping within a domain.
// It provides an additional layer of hierarchy between Domain and App.
//
// Hierarchy: Ecosystem -> Domain -> System -> App -> Workspace
type System struct {
	ID              int            `db:"id" json:"id" yaml:"-"`
	EcosystemID     sql.NullInt64  `db:"ecosystem_id" json:"ecosystem_id,omitempty" yaml:"-"`
	DomainID        sql.NullInt64  `db:"domain_id" json:"domain_id,omitempty" yaml:"-"`
	Name            string         `db:"name" json:"name" yaml:"name"`
	Description     sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Theme           sql.NullString `db:"theme" json:"theme,omitempty" yaml:"theme,omitempty"`
	NvimPackage     sql.NullString `db:"nvim_package" json:"nvim_package,omitempty" yaml:"nvim_package,omitempty"`
	TerminalPackage sql.NullString `db:"terminal_package" json:"terminal_package,omitempty" yaml:"terminal_package,omitempty"`
	BuildArgs       sql.NullString `db:"build_args" json:"build_args,omitempty" yaml:"-"`
	CACerts         sql.NullString `db:"ca_certs" json:"ca_certs,omitempty" yaml:"-"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// SystemYAML represents the YAML serialization format for a system
type SystemYAML struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   SystemMetadata `yaml:"metadata"`
	Spec       SystemSpec     `yaml:"spec"`
}

// SystemMetadata contains system metadata
type SystemMetadata struct {
	Name        string            `yaml:"name"`
	Domain      string            `yaml:"domain,omitempty"`
	Ecosystem   string            `yaml:"ecosystem,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// SystemSpec contains system specification
type SystemSpec struct {
	Theme           string          `yaml:"theme,omitempty"`
	NvimPackage     string          `yaml:"nvimPackage,omitempty"`
	TerminalPackage string          `yaml:"terminalPackage,omitempty"`
	Apps            []string        `yaml:"apps,omitempty"`
	Build           BuildArgsConfig `yaml:"build,omitempty"`
	CACerts         []CACertConfig  `yaml:"caCerts,omitempty"`
}

// ToYAML converts a System to YAML format.
// domainName is the parent domain name (pass "" if none).
// appNames should contain the names of child apps (pass nil for empty).
func (s *System) ToYAML(domainName string, appNames []string) SystemYAML {
	description := ""
	if s.Description.Valid {
		description = s.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	theme := ""
	if s.Theme.Valid {
		theme = s.Theme.String
	}

	nvimPackage := ""
	if s.NvimPackage.Valid {
		nvimPackage = s.NvimPackage.String
	}

	terminalPackage := ""
	if s.TerminalPackage.Valid {
		terminalPackage = s.TerminalPackage.String
	}

	// Restore build args from DB JSON blob if present
	var buildConfig BuildArgsConfig
	if s.BuildArgs.Valid && s.BuildArgs.String != "" {
		var args map[string]string
		if err := json.Unmarshal([]byte(s.BuildArgs.String), &args); err == nil && len(args) > 0 {
			buildConfig.Args = args
		}
	}

	// Restore CA certs from DB JSON blob if present
	var caCerts []CACertConfig
	if s.CACerts.Valid && s.CACerts.String != "" {
		var certs []CACertConfig
		if err := json.Unmarshal([]byte(s.CACerts.String), &certs); err == nil && len(certs) > 0 {
			caCerts = certs
		}
	}

	return SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata: SystemMetadata{
			Name:        s.Name,
			Domain:      domainName,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: SystemSpec{
			Theme:           theme,
			NvimPackage:     nvimPackage,
			TerminalPackage: terminalPackage,
			Apps:            appNames,
			Build:           buildConfig,
			CACerts:         caCerts,
		},
	}
}

// FromYAML converts YAML format to a System
func (s *System) FromYAML(yaml SystemYAML) {
	s.Name = yaml.Metadata.Name

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		s.Description = sql.NullString{String: desc, Valid: true}
	}

	if yaml.Spec.Theme != "" {
		s.Theme = sql.NullString{String: yaml.Spec.Theme, Valid: true}
	}

	if yaml.Spec.NvimPackage != "" {
		s.NvimPackage = sql.NullString{String: yaml.Spec.NvimPackage, Valid: true}
	}

	if yaml.Spec.TerminalPackage != "" {
		s.TerminalPackage = sql.NullString{String: yaml.Spec.TerminalPackage, Valid: true}
	}

	// Persist build args as JSON
	if len(yaml.Spec.Build.Args) > 0 {
		if b, err := json.Marshal(yaml.Spec.Build.Args); err == nil {
			s.BuildArgs = sql.NullString{String: string(b), Valid: true}
		}
	}

	// Persist CA certs as JSON (separate column)
	if len(yaml.Spec.CACerts) > 0 {
		if b, err := json.Marshal(yaml.Spec.CACerts); err == nil {
			s.CACerts = sql.NullString{String: string(b), Valid: true}
		}
	}
}
