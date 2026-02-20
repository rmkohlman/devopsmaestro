package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NvimPackageDB represents an nvim package collection stored in the database.
// This is separate from the YAML types in pkg/nvimops/package to maintain
// clear separation between database persistence and business logic.
type NvimPackageDB struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Category    sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	Labels      sql.NullString `db:"labels" json:"labels,omitempty" yaml:"labels,omitempty"`    // JSON object
	Plugins     string         `db:"plugins" json:"plugins" yaml:"plugins"`                     // JSON array - required
	Extends     sql.NullString `db:"extends" json:"extends,omitempty" yaml:"extends,omitempty"` // optional parent package
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// GetLabels returns the labels as a map, or empty map if no labels are stored.
func (p *NvimPackageDB) GetLabels() map[string]string {
	if !p.Labels.Valid {
		return make(map[string]string)
	}

	var labels map[string]string
	if err := json.Unmarshal([]byte(p.Labels.String), &labels); err != nil {
		return make(map[string]string)
	}

	return labels
}

// SetLabels stores the labels as a JSON string.
func (p *NvimPackageDB) SetLabels(labels map[string]string) error {
	if len(labels) == 0 {
		p.Labels = sql.NullString{Valid: false}
		return nil
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return err
	}

	p.Labels = sql.NullString{String: string(labelsJSON), Valid: true}
	return nil
}

// GetPlugins returns the plugins as a string slice.
func (p *NvimPackageDB) GetPlugins() []string {
	var plugins []string
	if err := json.Unmarshal([]byte(p.Plugins), &plugins); err != nil {
		return make([]string, 0)
	}

	return plugins
}

// SetPlugins stores the plugins as a JSON array string.
func (p *NvimPackageDB) SetPlugins(plugins []string) error {
	if plugins == nil {
		plugins = make([]string, 0)
	}

	pluginsJSON, err := json.Marshal(plugins)
	if err != nil {
		return err
	}

	p.Plugins = string(pluginsJSON)
	return nil
}
