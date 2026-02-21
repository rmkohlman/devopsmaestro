package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// TerminalPluginDB represents a terminal plugin stored in the database.
type TerminalPluginDB struct {
	ID           int            `db:"id" json:"id" yaml:"-"`
	Name         string         `db:"name" json:"name" yaml:"name"`
	Description  sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Repo         string         `db:"repo" json:"repo" yaml:"repo"`
	Category     sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	Shell        string         `db:"shell" json:"shell" yaml:"shell"`                                          // zsh, bash, fish, etc.
	Manager      string         `db:"manager" json:"manager" yaml:"manager"`                                    // oh-my-zsh, prezto, manual, etc.
	LoadCommand  sql.NullString `db:"load_command" json:"load_command,omitempty" yaml:"load_command,omitempty"` // Command to load the plugin
	SourceFile   sql.NullString `db:"source_file" json:"source_file,omitempty" yaml:"source_file,omitempty"`    // File to source
	Dependencies string         `db:"dependencies" json:"dependencies" yaml:"dependencies"`                     // JSON array, always valid JSON
	EnvVars      string         `db:"env_vars" json:"env_vars" yaml:"env_vars"`                                 // JSON object, always valid JSON
	Labels       string         `db:"labels" json:"labels" yaml:"labels"`                                       // JSON object, always valid JSON
	Enabled      bool           `db:"enabled" json:"enabled" yaml:"enabled"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// TerminalPluginYAML represents the YAML format for terminal plugin definition files
type TerminalPluginYAML struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   TerminalPluginMetadata `yaml:"metadata"`
	Spec       TerminalPluginSpec     `yaml:"spec"`
}

// TerminalPluginMetadata contains terminal plugin metadata
type TerminalPluginMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// TerminalPluginSpec contains the terminal plugin specification
type TerminalPluginSpec struct {
	Repo         string            `yaml:"repo"`
	Shell        string            `yaml:"shell,omitempty"`   // defaults to 'zsh'
	Manager      string            `yaml:"manager,omitempty"` // defaults to 'manual'
	LoadCommand  string            `yaml:"loadCommand,omitempty"`
	SourceFile   string            `yaml:"sourceFile,omitempty"`
	Dependencies []string          `yaml:"dependencies,omitempty"`
	EnvVars      map[string]string `yaml:"envVars,omitempty"`
}

// ToYAML converts a database terminal plugin to YAML format
func (p *TerminalPluginDB) ToYAML() (TerminalPluginYAML, error) {
	yaml := TerminalPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPlugin",
		Metadata: TerminalPluginMetadata{
			Name: p.Name,
		},
		Spec: TerminalPluginSpec{
			Repo:    p.Repo,
			Shell:   p.Shell,
			Manager: p.Manager,
		},
	}

	if p.Description.Valid {
		yaml.Metadata.Description = p.Description.String
	}

	if p.Category.Valid {
		yaml.Metadata.Category = p.Category.String
	}

	// Parse labels JSON
	if p.Labels != "" && p.Labels != "{}" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(p.Labels), &labels); err == nil {
			yaml.Metadata.Labels = labels
		}
	}

	if p.LoadCommand.Valid {
		yaml.Spec.LoadCommand = p.LoadCommand.String
	}

	if p.SourceFile.Valid {
		yaml.Spec.SourceFile = p.SourceFile.String
	}

	// Parse dependencies JSON
	if p.Dependencies != "" && p.Dependencies != "[]" {
		var deps []string
		if err := json.Unmarshal([]byte(p.Dependencies), &deps); err == nil {
			yaml.Spec.Dependencies = deps
		}
	}

	// Parse env vars JSON
	if p.EnvVars != "" && p.EnvVars != "{}" {
		var envVars map[string]string
		if err := json.Unmarshal([]byte(p.EnvVars), &envVars); err == nil {
			yaml.Spec.EnvVars = envVars
		}
	}

	return yaml, nil
}

// FromYAML converts YAML format to a database terminal plugin
func (p *TerminalPluginDB) FromYAML(yaml TerminalPluginYAML) error {
	p.Name = yaml.Metadata.Name
	p.Repo = yaml.Spec.Repo
	p.Shell = yaml.Spec.Shell
	p.Manager = yaml.Spec.Manager
	p.Enabled = true

	// Set defaults
	if p.Shell == "" {
		p.Shell = "zsh"
	}
	if p.Manager == "" {
		p.Manager = "manual"
	}

	if yaml.Metadata.Description != "" {
		p.Description = sql.NullString{String: yaml.Metadata.Description, Valid: true}
	}

	if yaml.Metadata.Category != "" {
		p.Category = sql.NullString{String: yaml.Metadata.Category, Valid: true}
	}

	// Store labels as JSON
	if len(yaml.Metadata.Labels) > 0 {
		if labelsJSON, err := json.Marshal(yaml.Metadata.Labels); err == nil {
			p.Labels = string(labelsJSON)
		} else {
			p.Labels = "{}"
		}
	} else {
		p.Labels = "{}"
	}

	if yaml.Spec.LoadCommand != "" {
		p.LoadCommand = sql.NullString{String: yaml.Spec.LoadCommand, Valid: true}
	}

	if yaml.Spec.SourceFile != "" {
		p.SourceFile = sql.NullString{String: yaml.Spec.SourceFile, Valid: true}
	}

	// Store dependencies as JSON
	if len(yaml.Spec.Dependencies) > 0 {
		if depsJSON, err := json.Marshal(yaml.Spec.Dependencies); err == nil {
			p.Dependencies = string(depsJSON)
		} else {
			p.Dependencies = "[]"
		}
	} else {
		p.Dependencies = "[]"
	}

	// Store env vars as JSON
	if len(yaml.Spec.EnvVars) > 0 {
		if envVarsJSON, err := json.Marshal(yaml.Spec.EnvVars); err == nil {
			p.EnvVars = string(envVarsJSON)
		} else {
			p.EnvVars = "{}"
		}
	} else {
		p.EnvVars = "{}"
	}

	return nil
}

// GetDependencies returns the dependencies as a string slice
func (p *TerminalPluginDB) GetDependencies() ([]string, error) {
	var deps []string
	if p.Dependencies == "" || p.Dependencies == "[]" {
		return deps, nil
	}

	err := json.Unmarshal([]byte(p.Dependencies), &deps)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

// GetEnvVars returns the environment variables as a map
func (p *TerminalPluginDB) GetEnvVars() (map[string]string, error) {
	var envVars map[string]string
	if p.EnvVars == "" || p.EnvVars == "{}" {
		return map[string]string{}, nil
	}

	err := json.Unmarshal([]byte(p.EnvVars), &envVars)
	if err != nil {
		return nil, err
	}
	return envVars, nil
}

// GetLabels returns the labels as a map
func (p *TerminalPluginDB) GetLabels() (map[string]string, error) {
	var labels map[string]string
	if p.Labels == "" || p.Labels == "{}" {
		return map[string]string{}, nil
	}

	err := json.Unmarshal([]byte(p.Labels), &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}
