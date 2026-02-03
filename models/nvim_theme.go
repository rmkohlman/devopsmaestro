package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NvimThemeDB represents a Neovim theme stored in the database.
type NvimThemeDB struct {
	ID           int            `db:"id" json:"id" yaml:"-"`
	Name         string         `db:"name" json:"name" yaml:"name"`
	Description  sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Author       sql.NullString `db:"author" json:"author,omitempty" yaml:"author,omitempty"`
	Category     sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	PluginRepo   string         `db:"plugin_repo" json:"plugin_repo" yaml:"plugin_repo"`
	PluginBranch sql.NullString `db:"plugin_branch" json:"plugin_branch,omitempty" yaml:"plugin_branch,omitempty"`
	PluginTag    sql.NullString `db:"plugin_tag" json:"plugin_tag,omitempty" yaml:"plugin_tag,omitempty"`
	Style        sql.NullString `db:"style" json:"style,omitempty" yaml:"style,omitempty"`
	Transparent  bool           `db:"transparent" json:"transparent" yaml:"transparent"`
	Colors       sql.NullString `db:"colors" json:"colors,omitempty" yaml:"colors,omitempty"`    // JSON object
	Options      sql.NullString `db:"options" json:"options,omitempty" yaml:"options,omitempty"` // JSON object
	IsActive     bool           `db:"is_active" json:"is_active" yaml:"is_active"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// NvimThemeYAML represents the YAML format for theme definition files.
type NvimThemeYAML struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   ThemeMetadata `yaml:"metadata"`
	Spec       ThemeSpec     `yaml:"spec"`
}

// ThemeMetadata contains theme metadata.
type ThemeMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Author      string `yaml:"author,omitempty"`
	Category    string `yaml:"category,omitempty"`
}

// ThemeSpec contains the theme specification.
type ThemeSpec struct {
	Plugin      ThemePluginSpec        `yaml:"plugin"`
	Style       string                 `yaml:"style,omitempty"`
	Transparent bool                   `yaml:"transparent,omitempty"`
	Colors      map[string]string      `yaml:"colors,omitempty"`
	Options     map[string]interface{} `yaml:"options,omitempty"`
}

// ThemePluginSpec defines the colorscheme plugin to use.
type ThemePluginSpec struct {
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch,omitempty"`
	Tag    string `yaml:"tag,omitempty"`
}

// ToYAML converts a database theme to YAML format.
func (t *NvimThemeDB) ToYAML() (NvimThemeYAML, error) {
	yaml := NvimThemeYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimTheme",
		Metadata: ThemeMetadata{
			Name: t.Name,
		},
		Spec: ThemeSpec{
			Plugin: ThemePluginSpec{
				Repo: t.PluginRepo,
			},
			Transparent: t.Transparent,
		},
	}

	if t.Description.Valid {
		yaml.Metadata.Description = t.Description.String
	}

	if t.Author.Valid {
		yaml.Metadata.Author = t.Author.String
	}

	if t.Category.Valid {
		yaml.Metadata.Category = t.Category.String
	}

	if t.PluginBranch.Valid {
		yaml.Spec.Plugin.Branch = t.PluginBranch.String
	}

	if t.PluginTag.Valid {
		yaml.Spec.Plugin.Tag = t.PluginTag.String
	}

	if t.Style.Valid {
		yaml.Spec.Style = t.Style.String
	}

	if t.Colors.Valid {
		var colors map[string]string
		if err := json.Unmarshal([]byte(t.Colors.String), &colors); err == nil {
			yaml.Spec.Colors = colors
		}
	}

	if t.Options.Valid {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(t.Options.String), &options); err == nil {
			yaml.Spec.Options = options
		}
	}

	return yaml, nil
}

// FromYAML converts YAML format to a database theme.
func (t *NvimThemeDB) FromYAML(yaml NvimThemeYAML) error {
	t.Name = yaml.Metadata.Name
	t.PluginRepo = yaml.Spec.Plugin.Repo
	t.Transparent = yaml.Spec.Transparent

	if yaml.Metadata.Description != "" {
		t.Description = sql.NullString{String: yaml.Metadata.Description, Valid: true}
	}

	if yaml.Metadata.Author != "" {
		t.Author = sql.NullString{String: yaml.Metadata.Author, Valid: true}
	}

	if yaml.Metadata.Category != "" {
		t.Category = sql.NullString{String: yaml.Metadata.Category, Valid: true}
	}

	if yaml.Spec.Plugin.Branch != "" {
		t.PluginBranch = sql.NullString{String: yaml.Spec.Plugin.Branch, Valid: true}
	}

	if yaml.Spec.Plugin.Tag != "" {
		t.PluginTag = sql.NullString{String: yaml.Spec.Plugin.Tag, Valid: true}
	}

	if yaml.Spec.Style != "" {
		t.Style = sql.NullString{String: yaml.Spec.Style, Valid: true}
	}

	if len(yaml.Spec.Colors) > 0 {
		if colorsJSON, err := json.Marshal(yaml.Spec.Colors); err == nil {
			t.Colors = sql.NullString{String: string(colorsJSON), Valid: true}
		}
	}

	if len(yaml.Spec.Options) > 0 {
		if optionsJSON, err := json.Marshal(yaml.Spec.Options); err == nil {
			t.Options = sql.NullString{String: string(optionsJSON), Valid: true}
		}
	}

	return nil
}
