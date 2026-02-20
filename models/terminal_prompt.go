package models

import (
	"database/sql"
	"time"
)

// TerminalPromptDB represents a terminal prompt stored in the database.
type TerminalPromptDB struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	Type        string         `db:"type" json:"type" yaml:"type"` // starship, powerlevel10k, oh-my-posh
	AddNewline  bool           `db:"add_newline" json:"add_newline" yaml:"add_newline"`
	Palette     sql.NullString `db:"palette" json:"palette,omitempty" yaml:"palette,omitempty"`
	Format      sql.NullString `db:"format" json:"format,omitempty" yaml:"format,omitempty"`
	Modules     sql.NullString `db:"modules" json:"modules,omitempty" yaml:"modules,omitempty"`       // JSON: map[string]ModuleConfig
	Character   sql.NullString `db:"character" json:"character,omitempty" yaml:"character,omitempty"` // JSON: *CharacterConfig
	PaletteRef  sql.NullString `db:"palette_ref" json:"palette_ref,omitempty" yaml:"palette_ref,omitempty"`
	Colors      sql.NullString `db:"colors" json:"colors,omitempty" yaml:"colors,omitempty"` // JSON: map[string]string
	RawConfig   sql.NullString `db:"raw_config" json:"raw_config,omitempty" yaml:"raw_config,omitempty"`
	Category    sql.NullString `db:"category" json:"category,omitempty" yaml:"category,omitempty"`
	Tags        sql.NullString `db:"tags" json:"tags,omitempty" yaml:"tags,omitempty"` // JSON: []string
	Enabled     bool           `db:"enabled" json:"enabled" yaml:"enabled"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}
