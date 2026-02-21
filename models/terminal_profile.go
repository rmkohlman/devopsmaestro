package models

import (
	"database/sql"
	"time"
)

// TerminalProfileDB represents a terminal profile stored in the database.
type TerminalProfileDB struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Category    sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	PromptRef   sql.NullString `db:"prompt_ref" json:"prompt_ref,omitempty" yaml:"prompt_ref,omitempty"` // Reference to prompt by name
	PluginRefs  string         `db:"plugin_refs" json:"plugin_refs" yaml:"plugin_refs"`                  // JSON array of plugin names
	ShellRef    sql.NullString `db:"shell_ref" json:"shell_ref,omitempty" yaml:"shell_ref,omitempty"`    // Reference to shell config
	ThemeRef    sql.NullString `db:"theme_ref" json:"theme_ref,omitempty" yaml:"theme_ref,omitempty"`    // Reference to theme
	Tags        string         `db:"tags" json:"tags" yaml:"tags"`                                       // JSON array
	Labels      string         `db:"labels" json:"labels" yaml:"labels"`                                 // JSON object
	Enabled     bool           `db:"enabled" json:"enabled" yaml:"enabled"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}
