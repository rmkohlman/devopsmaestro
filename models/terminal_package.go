package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// TerminalPackageDB represents a terminal package collection stored in the database.
// This is separate from the YAML types in pkg/terminalops/package to maintain
// clear separation between database persistence and business logic.
type TerminalPackageDB struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Category    sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	Labels      sql.NullString `db:"labels" json:"labels,omitempty" yaml:"labels,omitempty"`    // JSON object
	Plugins     string         `db:"plugins" json:"plugins" yaml:"plugins"`                     // JSON array - required
	Prompts     string         `db:"prompts" json:"prompts" yaml:"prompts"`                     // JSON array - required
	Profiles    string         `db:"profiles" json:"profiles" yaml:"profiles"`                  // JSON array - required
	WezTerm     sql.NullString `db:"wezterm" json:"wezterm,omitempty" yaml:"wezterm,omitempty"` // JSON object - optional
	Extends     sql.NullString `db:"extends" json:"extends,omitempty" yaml:"extends,omitempty"` // optional parent package
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// GetLabels returns the labels as a map, or empty map if no labels are stored.
func (p *TerminalPackageDB) GetLabels() map[string]string {
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
func (p *TerminalPackageDB) SetLabels(labels map[string]string) error {
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
func (p *TerminalPackageDB) GetPlugins() []string {
	var plugins []string
	if err := json.Unmarshal([]byte(p.Plugins), &plugins); err != nil {
		return make([]string, 0)
	}

	return plugins
}

// SetPlugins stores the plugins as a JSON array string.
func (p *TerminalPackageDB) SetPlugins(plugins []string) error {
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

// GetPrompts returns the prompts as a string slice.
func (p *TerminalPackageDB) GetPrompts() []string {
	var prompts []string
	if err := json.Unmarshal([]byte(p.Prompts), &prompts); err != nil {
		return make([]string, 0)
	}

	return prompts
}

// SetPrompts stores the prompts as a JSON array string.
func (p *TerminalPackageDB) SetPrompts(prompts []string) error {
	if prompts == nil {
		prompts = make([]string, 0)
	}

	promptsJSON, err := json.Marshal(prompts)
	if err != nil {
		return err
	}

	p.Prompts = string(promptsJSON)
	return nil
}

// GetProfiles returns the profiles as a string slice.
func (p *TerminalPackageDB) GetProfiles() []string {
	var profiles []string
	if err := json.Unmarshal([]byte(p.Profiles), &profiles); err != nil {
		return make([]string, 0)
	}

	return profiles
}

// SetProfiles stores the profiles as a JSON array string.
func (p *TerminalPackageDB) SetProfiles(profiles []string) error {
	if profiles == nil {
		profiles = make([]string, 0)
	}

	profilesJSON, err := json.Marshal(profiles)
	if err != nil {
		return err
	}

	p.Profiles = string(profilesJSON)
	return nil
}

// GetWezTerm returns the WezTerm configuration as a struct, or nil if not set.
func (p *TerminalPackageDB) GetWezTerm() map[string]interface{} {
	if !p.WezTerm.Valid {
		return nil
	}

	var wezterm map[string]interface{}
	if err := json.Unmarshal([]byte(p.WezTerm.String), &wezterm); err != nil {
		return nil
	}

	return wezterm
}

// SetWezTerm stores the WezTerm configuration as a JSON string.
func (p *TerminalPackageDB) SetWezTerm(wezterm map[string]interface{}) error {
	if wezterm == nil {
		p.WezTerm = sql.NullString{Valid: false}
		return nil
	}

	weztermJSON, err := json.Marshal(wezterm)
	if err != nil {
		return err
	}

	p.WezTerm = sql.NullString{String: string(weztermJSON), Valid: true}
	return nil
}
