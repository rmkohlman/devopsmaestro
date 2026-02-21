package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// TerminalEmulatorDB represents a terminal emulator configuration stored in the database.
type TerminalEmulatorDB struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Type        string         `db:"type" json:"type" yaml:"type"`                                    // wezterm, alacritty, kitty, etc.
	Config      string         `db:"config" json:"config" yaml:"config"`                              // JSON blob for emulator-specific config
	ThemeRef    sql.NullString `db:"theme_ref" json:"theme_ref,omitempty" yaml:"theme_ref,omitempty"` // Reference to theme name
	Category    sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`    // User-defined category
	Labels      string         `db:"labels" json:"labels" yaml:"labels"`                              // JSON object, always valid JSON
	Workspace   sql.NullString `db:"workspace" json:"workspace,omitempty" yaml:"workspace,omitempty"` // Associated workspace
	Enabled     bool           `db:"enabled" json:"enabled" yaml:"enabled"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// TerminalEmulatorYAML represents the YAML format for terminal emulator definition files
type TerminalEmulatorYAML struct {
	APIVersion string                   `yaml:"apiVersion"`
	Kind       string                   `yaml:"kind"`
	Metadata   TerminalEmulatorMetadata `yaml:"metadata"`
	Spec       TerminalEmulatorSpec     `yaml:"spec"`
}

// TerminalEmulatorMetadata contains terminal emulator metadata
type TerminalEmulatorMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// TerminalEmulatorSpec contains the terminal emulator specification
type TerminalEmulatorSpec struct {
	Type      string         `yaml:"type"`                // wezterm, alacritty, kitty, etc.
	Config    map[string]any `yaml:"config,omitempty"`    // Emulator-specific configuration
	ThemeRef  string         `yaml:"themeRef,omitempty"`  // Reference to a theme
	Workspace string         `yaml:"workspace,omitempty"` // Associated workspace
}

// ToYAML converts a database terminal emulator to YAML format
func (e *TerminalEmulatorDB) ToYAML() (TerminalEmulatorYAML, error) {
	yaml := TerminalEmulatorYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalEmulator",
		Metadata: TerminalEmulatorMetadata{
			Name: e.Name,
		},
		Spec: TerminalEmulatorSpec{
			Type: e.Type,
		},
	}

	if e.Description.Valid {
		yaml.Metadata.Description = e.Description.String
	}

	if e.Category.Valid {
		yaml.Metadata.Category = e.Category.String
	}

	// Parse labels JSON
	if e.Labels != "" && e.Labels != "{}" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(e.Labels), &labels); err == nil {
			yaml.Metadata.Labels = labels
		}
	}

	if e.ThemeRef.Valid {
		yaml.Spec.ThemeRef = e.ThemeRef.String
	}

	if e.Workspace.Valid {
		yaml.Spec.Workspace = e.Workspace.String
	}

	// Parse config JSON
	if e.Config != "" && e.Config != "{}" {
		var config map[string]any
		if err := json.Unmarshal([]byte(e.Config), &config); err == nil {
			yaml.Spec.Config = config
		}
	}

	return yaml, nil
}

// FromYAML converts YAML format to a database terminal emulator
func (e *TerminalEmulatorDB) FromYAML(yaml TerminalEmulatorYAML) error {
	e.Name = yaml.Metadata.Name
	e.Type = yaml.Spec.Type
	e.Enabled = true

	if yaml.Metadata.Description != "" {
		e.Description = sql.NullString{String: yaml.Metadata.Description, Valid: true}
	}

	if yaml.Metadata.Category != "" {
		e.Category = sql.NullString{String: yaml.Metadata.Category, Valid: true}
	}

	// Store labels as JSON
	if len(yaml.Metadata.Labels) > 0 {
		if labelsJSON, err := json.Marshal(yaml.Metadata.Labels); err == nil {
			e.Labels = string(labelsJSON)
		} else {
			e.Labels = "{}"
		}
	} else {
		e.Labels = "{}"
	}

	if yaml.Spec.ThemeRef != "" {
		e.ThemeRef = sql.NullString{String: yaml.Spec.ThemeRef, Valid: true}
	}

	if yaml.Spec.Workspace != "" {
		e.Workspace = sql.NullString{String: yaml.Spec.Workspace, Valid: true}
	}

	// Store config as JSON
	if len(yaml.Spec.Config) > 0 {
		if configJSON, err := json.Marshal(yaml.Spec.Config); err == nil {
			e.Config = string(configJSON)
		} else {
			e.Config = "{}"
		}
	} else {
		e.Config = "{}"
	}

	return nil
}

// GetConfig returns the configuration as a map
func (e *TerminalEmulatorDB) GetConfig() (map[string]any, error) {
	var config map[string]any
	if e.Config == "" || e.Config == "{}" {
		return map[string]any{}, nil
	}

	err := json.Unmarshal([]byte(e.Config), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetLabels returns the labels as a map
func (e *TerminalEmulatorDB) GetLabels() (map[string]string, error) {
	var labels map[string]string
	if e.Labels == "" || e.Labels == "{}" {
		return map[string]string{}, nil
	}

	err := json.Unmarshal([]byte(e.Labels), &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}
