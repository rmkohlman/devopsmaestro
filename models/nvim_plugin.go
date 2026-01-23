package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NvimPlugin represents a reusable nvim plugin definition stored in the database
type NvimPluginDB struct {
	ID           int            `db:"id" json:"id" yaml:"-"`
	Name         string         `db:"name" json:"name" yaml:"name"`
	Description  sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Repo         string         `db:"repo" json:"repo" yaml:"repo"`
	Branch       sql.NullString `db:"branch" json:"branch,omitempty" yaml:"branch,omitempty"`
	Version      sql.NullString `db:"version" json:"version,omitempty" yaml:"version,omitempty"`
	Priority     sql.NullInt64  `db:"priority" json:"priority,omitempty" yaml:"priority,omitempty"`
	Lazy         bool           `db:"lazy" json:"lazy" yaml:"lazy"`
	Event        sql.NullString `db:"event" json:"event,omitempty" yaml:"event,omitempty"`                      // JSON array
	Ft           sql.NullString `db:"ft" json:"ft,omitempty" yaml:"ft,omitempty"`                               // JSON array
	Keys         sql.NullString `db:"keys" json:"keys,omitempty" yaml:"keys,omitempty"`                         // JSON array
	Cmd          sql.NullString `db:"cmd" json:"cmd,omitempty" yaml:"cmd,omitempty"`                            // JSON array
	Dependencies sql.NullString `db:"dependencies" json:"dependencies,omitempty" yaml:"dependencies,omitempty"` // JSON array
	Build        sql.NullString `db:"build" json:"build,omitempty" yaml:"build,omitempty"`
	Config       sql.NullString `db:"config" json:"config,omitempty" yaml:"config,omitempty"`    // Lua code
	Init         sql.NullString `db:"init" json:"init,omitempty" yaml:"init,omitempty"`          // Lua code
	Opts         sql.NullString `db:"opts" json:"opts,omitempty" yaml:"opts,omitempty"`          // JSON object
	Keymaps      sql.NullString `db:"keymaps" json:"keymaps,omitempty" yaml:"keymaps,omitempty"` // JSON array
	Category     sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	Tags         sql.NullString `db:"tags" json:"tags,omitempty" yaml:"tags,omitempty"` // JSON array
	Enabled      bool           `db:"enabled" json:"enabled" yaml:"enabled"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// NvimPluginYAML represents the YAML format for plugin definition files
type NvimPluginYAML struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   PluginMetadata `yaml:"metadata"`
	Spec       PluginSpec     `yaml:"spec"`
}

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// PluginSpec contains the plugin specification
type PluginSpec struct {
	Repo         string                 `yaml:"repo"`
	Branch       string                 `yaml:"branch,omitempty"`
	Version      string                 `yaml:"version,omitempty"`
	Priority     int                    `yaml:"priority,omitempty"`
	Lazy         bool                   `yaml:"lazy,omitempty"`
	Event        interface{}            `yaml:"event,omitempty"` // string or []string
	Ft           interface{}            `yaml:"ft,omitempty"`    // string or []string
	Keys         []PluginKeymap         `yaml:"keys,omitempty"`
	Cmd          interface{}            `yaml:"cmd,omitempty"`          // string or []string
	Dependencies []interface{}          `yaml:"dependencies,omitempty"` // string or PluginDependency
	Build        string                 `yaml:"build,omitempty"`
	Config       string                 `yaml:"config,omitempty"`
	Init         string                 `yaml:"init,omitempty"`
	Opts         map[string]interface{} `yaml:"opts,omitempty"`
	Keymaps      []PluginKeymap         `yaml:"keymaps,omitempty"`
}

// PluginDependency represents a plugin dependency
type PluginDependency struct {
	Repo    string `yaml:"repo"`
	Build   string `yaml:"build,omitempty"`
	Version string `yaml:"version,omitempty"`
	Branch  string `yaml:"branch,omitempty"`
}

// PluginKeymap represents a key mapping
type PluginKeymap struct {
	Key    string      `yaml:"key"`
	Mode   interface{} `yaml:"mode,omitempty"`   // string or []string
	Action string      `yaml:"action,omitempty"` // Lua code or command
	Desc   string      `yaml:"desc,omitempty"`
}

// ToYAML converts a database plugin to YAML format
func (p *NvimPluginDB) ToYAML() (NvimPluginYAML, error) {
	yaml := NvimPluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: p.Name,
		},
		Spec: PluginSpec{
			Repo: p.Repo,
			Lazy: p.Lazy,
		},
	}

	if p.Description.Valid {
		yaml.Metadata.Description = p.Description.String
	}

	if p.Category.Valid {
		yaml.Metadata.Category = p.Category.String
	}

	if p.Tags.Valid {
		var tags []string
		if err := json.Unmarshal([]byte(p.Tags.String), &tags); err == nil {
			yaml.Metadata.Tags = tags
		}
	}

	if p.Branch.Valid {
		yaml.Spec.Branch = p.Branch.String
	}

	if p.Version.Valid {
		yaml.Spec.Version = p.Version.String
	}

	if p.Priority.Valid {
		yaml.Spec.Priority = int(p.Priority.Int64)
	}

	if p.Event.Valid {
		var event interface{}
		if err := json.Unmarshal([]byte(p.Event.String), &event); err == nil {
			yaml.Spec.Event = event
		}
	}

	if p.Ft.Valid {
		var ft interface{}
		if err := json.Unmarshal([]byte(p.Ft.String), &ft); err == nil {
			yaml.Spec.Ft = ft
		}
	}

	if p.Keys.Valid {
		var keys []PluginKeymap
		if err := json.Unmarshal([]byte(p.Keys.String), &keys); err == nil {
			yaml.Spec.Keys = keys
		}
	}

	if p.Cmd.Valid {
		var cmd interface{}
		if err := json.Unmarshal([]byte(p.Cmd.String), &cmd); err == nil {
			yaml.Spec.Cmd = cmd
		}
	}

	if p.Dependencies.Valid {
		var deps []interface{}
		if err := json.Unmarshal([]byte(p.Dependencies.String), &deps); err == nil {
			yaml.Spec.Dependencies = deps
		}
	}

	if p.Build.Valid {
		yaml.Spec.Build = p.Build.String
	}

	if p.Config.Valid {
		yaml.Spec.Config = p.Config.String
	}

	if p.Init.Valid {
		yaml.Spec.Init = p.Init.String
	}

	if p.Opts.Valid {
		var opts map[string]interface{}
		if err := json.Unmarshal([]byte(p.Opts.String), &opts); err == nil {
			yaml.Spec.Opts = opts
		}
	}

	if p.Keymaps.Valid {
		var keymaps []PluginKeymap
		if err := json.Unmarshal([]byte(p.Keymaps.String), &keymaps); err == nil {
			yaml.Spec.Keymaps = keymaps
		}
	}

	return yaml, nil
}

// FromYAML converts YAML format to a database plugin
func (p *NvimPluginDB) FromYAML(yaml NvimPluginYAML) error {
	p.Name = yaml.Metadata.Name
	p.Repo = yaml.Spec.Repo
	p.Lazy = yaml.Spec.Lazy
	p.Enabled = true

	if yaml.Metadata.Description != "" {
		p.Description = sql.NullString{String: yaml.Metadata.Description, Valid: true}
	}

	if yaml.Metadata.Category != "" {
		p.Category = sql.NullString{String: yaml.Metadata.Category, Valid: true}
	}

	if len(yaml.Metadata.Tags) > 0 {
		if tagsJSON, err := json.Marshal(yaml.Metadata.Tags); err == nil {
			p.Tags = sql.NullString{String: string(tagsJSON), Valid: true}
		}
	}

	if yaml.Spec.Branch != "" {
		p.Branch = sql.NullString{String: yaml.Spec.Branch, Valid: true}
	}

	if yaml.Spec.Version != "" {
		p.Version = sql.NullString{String: yaml.Spec.Version, Valid: true}
	}

	if yaml.Spec.Priority != 0 {
		p.Priority = sql.NullInt64{Int64: int64(yaml.Spec.Priority), Valid: true}
	}

	if yaml.Spec.Event != nil {
		if eventJSON, err := json.Marshal(yaml.Spec.Event); err == nil {
			p.Event = sql.NullString{String: string(eventJSON), Valid: true}
		}
	}

	if yaml.Spec.Ft != nil {
		if ftJSON, err := json.Marshal(yaml.Spec.Ft); err == nil {
			p.Ft = sql.NullString{String: string(ftJSON), Valid: true}
		}
	}

	if len(yaml.Spec.Keys) > 0 {
		if keysJSON, err := json.Marshal(yaml.Spec.Keys); err == nil {
			p.Keys = sql.NullString{String: string(keysJSON), Valid: true}
		}
	}

	if yaml.Spec.Cmd != nil {
		if cmdJSON, err := json.Marshal(yaml.Spec.Cmd); err == nil {
			p.Cmd = sql.NullString{String: string(cmdJSON), Valid: true}
		}
	}

	if len(yaml.Spec.Dependencies) > 0 {
		if depsJSON, err := json.Marshal(yaml.Spec.Dependencies); err == nil {
			p.Dependencies = sql.NullString{String: string(depsJSON), Valid: true}
		}
	}

	if yaml.Spec.Build != "" {
		p.Build = sql.NullString{String: yaml.Spec.Build, Valid: true}
	}

	if yaml.Spec.Config != "" {
		p.Config = sql.NullString{String: yaml.Spec.Config, Valid: true}
	}

	if yaml.Spec.Init != "" {
		p.Init = sql.NullString{String: yaml.Spec.Init, Valid: true}
	}

	if len(yaml.Spec.Opts) > 0 {
		if optsJSON, err := json.Marshal(yaml.Spec.Opts); err == nil {
			p.Opts = sql.NullString{String: string(optsJSON), Valid: true}
		}
	}

	if len(yaml.Spec.Keymaps) > 0 {
		if keymapsJSON, err := json.Marshal(yaml.Spec.Keymaps); err == nil {
			p.Keymaps = sql.NullString{String: string(keymapsJSON), Valid: true}
		}
	}

	return nil
}
