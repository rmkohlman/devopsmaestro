// Package emulator provides types and utilities for terminal emulator management.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package emulator

import (
	"time"
)

// EmulatorType represents the type of terminal emulator
type EmulatorType string

const (
	// EmulatorTypeWezterm is the WezTerm terminal emulator
	EmulatorTypeWezterm EmulatorType = "wezterm"
	// EmulatorTypeAlacritty is the Alacritty terminal emulator
	EmulatorTypeAlacritty EmulatorType = "alacritty"
	// EmulatorTypeKitty is the Kitty terminal emulator
	EmulatorTypeKitty EmulatorType = "kitty"
	// EmulatorTypeITerm2 is the iTerm2 terminal emulator (macOS)
	EmulatorTypeITerm2 EmulatorType = "iterm2"
)

// Emulator represents a terminal emulator configuration.
// This is the canonical type used throughout terminal emulator management.
type Emulator struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Emulator type and configuration
	Type   EmulatorType   `json:"type" yaml:"type"`
	Config map[string]any `json:"config,omitempty" yaml:"config,omitempty"`

	// Theme and workspace association
	ThemeRef  string `json:"theme_ref,omitempty" yaml:"theme_ref,omitempty"`
	Workspace string `json:"workspace,omitempty" yaml:"workspace,omitempty"`

	// Metadata
	Category string            `json:"category,omitempty" yaml:"category,omitempty"`
	Labels   map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Enabled  bool              `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// NewEmulator creates a new Emulator with default values.
func NewEmulator(name string, emulatorType EmulatorType) *Emulator {
	return &Emulator{
		Name:    name,
		Type:    emulatorType,
		Config:  make(map[string]any),
		Labels:  make(map[string]string),
		Enabled: true,
	}
}

// IsEnabled returns whether the emulator configuration is enabled.
func (e *Emulator) IsEnabled() bool {
	return e.Enabled
}

// HasTheme returns whether the emulator has a theme reference.
func (e *Emulator) HasTheme() bool {
	return e.ThemeRef != ""
}

// GetConfigValue retrieves a configuration value by key.
func (e *Emulator) GetConfigValue(key string) (any, bool) {
	if e.Config == nil {
		return nil, false
	}
	value, exists := e.Config[key]
	return value, exists
}

// SetConfigValue sets a configuration value by key.
func (e *Emulator) SetConfigValue(key string, value any) {
	if e.Config == nil {
		e.Config = make(map[string]any)
	}
	e.Config[key] = value
}

// GetLabel retrieves a label value by key.
func (e *Emulator) GetLabel(key string) (string, bool) {
	if e.Labels == nil {
		return "", false
	}
	value, exists := e.Labels[key]
	return value, exists
}

// SetLabel sets a label value by key.
func (e *Emulator) SetLabel(key, value string) {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}
	e.Labels[key] = value
}
