// Package shell provides types and utilities for shell configuration management.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package shell

import (
	"time"
)

// ShellType identifies the shell to configure.
type ShellType string

const (
	// ShellTypeZsh is the Z shell
	ShellTypeZsh ShellType = "zsh"
	// ShellTypeBash is the Bourne Again shell
	ShellTypeBash ShellType = "bash"
	// ShellTypeFish is the Friendly Interactive shell
	ShellTypeFish ShellType = "fish"
)

// Shell represents shell configuration (aliases, env vars, options).
// This is the canonical type used throughout terminal-manager.
type Shell struct {
	// Core identification
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	ShellType   ShellType `json:"shell_type" yaml:"shell_type"`

	// Environment variables
	Env []EnvVar `json:"env,omitempty" yaml:"env,omitempty"`

	// Aliases
	Aliases []Alias `json:"aliases,omitempty" yaml:"aliases,omitempty"`

	// Functions (shell functions)
	Functions []Function `json:"functions,omitempty" yaml:"functions,omitempty"`

	// Path additions
	PathPrepend []string `json:"path_prepend,omitempty" yaml:"path_prepend,omitempty"`
	PathAppend  []string `json:"path_append,omitempty" yaml:"path_append,omitempty"`

	// Shell options (setopt for zsh, shopt for bash)
	Options []string `json:"options,omitempty" yaml:"options,omitempty"`

	// History configuration
	History *HistoryConfig `json:"history,omitempty" yaml:"history,omitempty"`

	// Keybindings
	Keybindings []Keybinding `json:"keybindings,omitempty" yaml:"keybindings,omitempty"`

	// Raw shell code to include (escape hatch)
	RawConfig string `json:"raw_config,omitempty" yaml:"raw_config,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Name        string `json:"name" yaml:"name"`
	Value       string `json:"value,omitempty" yaml:"value,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Expand controls whether to expand variables in the value
	Expand bool `json:"expand,omitempty" yaml:"expand,omitempty"`
}

// Alias represents a shell alias.
type Alias struct {
	Name        string `json:"name" yaml:"name"`
	Command     string `json:"command" yaml:"command"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Global makes this a global alias (zsh only, expands anywhere in command)
	Global bool `json:"global,omitempty" yaml:"global,omitempty"`
}

// Function represents a shell function.
type Function struct {
	Name        string `json:"name" yaml:"name"`
	Body        string `json:"body" yaml:"body"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// HistoryConfig configures shell history behavior.
type HistoryConfig struct {
	Size           int    `json:"size,omitempty" yaml:"size,omitempty"`
	File           string `json:"file,omitempty" yaml:"file,omitempty"`
	IgnoreDups     bool   `json:"ignore_dups,omitempty" yaml:"ignore_dups,omitempty"`
	IgnoreSpace    bool   `json:"ignore_space,omitempty" yaml:"ignore_space,omitempty"`
	ShareHistory   bool   `json:"share_history,omitempty" yaml:"share_history,omitempty"`
	ExtendedFormat bool   `json:"extended_format,omitempty" yaml:"extended_format,omitempty"` // Include timestamps
}

// Keybinding represents a shell keybinding.
type Keybinding struct {
	Key         string `json:"key" yaml:"key"`
	Widget      string `json:"widget,omitempty" yaml:"widget,omitempty"`   // zsh widget name
	Command     string `json:"command,omitempty" yaml:"command,omitempty"` // command to run
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// ShellYAML represents the full YAML document format (kubectl-style).
type ShellYAML struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   ShellMetadata `yaml:"metadata"`
	Spec       ShellSpec     `yaml:"spec"`
}

// ShellMetadata contains shell metadata in the YAML format.
type ShellMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// ShellSpec contains the shell specification in the YAML format.
type ShellSpec struct {
	ShellType   ShellType      `yaml:"shellType"`
	Env         []EnvVar       `yaml:"env,omitempty"`
	Aliases     []Alias        `yaml:"aliases,omitempty"`
	Functions   []Function     `yaml:"functions,omitempty"`
	PathPrepend []string       `yaml:"pathPrepend,omitempty"`
	PathAppend  []string       `yaml:"pathAppend,omitempty"`
	Options     []string       `yaml:"options,omitempty"`
	History     *HistoryConfig `yaml:"history,omitempty"`
	Keybindings []Keybinding   `yaml:"keybindings,omitempty"`
	RawConfig   string         `yaml:"rawConfig,omitempty"`
	Enabled     *bool          `yaml:"enabled,omitempty"`
}

// NewShell creates a new Shell with default values.
func NewShell(name string, shellType ShellType) *Shell {
	return &Shell{
		Name:      name,
		ShellType: shellType,
		Enabled:   true,
	}
}

// NewShellYAML creates a new ShellYAML with proper defaults.
func NewShellYAML(name string, shellType ShellType) *ShellYAML {
	return &ShellYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalShell",
		Metadata: ShellMetadata{
			Name: name,
		},
		Spec: ShellSpec{
			ShellType: shellType,
		},
	}
}

// ToShell converts ShellYAML to the canonical Shell type.
func (sy *ShellYAML) ToShell() *Shell {
	enabled := true
	if sy.Spec.Enabled != nil {
		enabled = *sy.Spec.Enabled
	}

	return &Shell{
		Name:        sy.Metadata.Name,
		Description: sy.Metadata.Description,
		Category:    sy.Metadata.Category,
		Tags:        sy.Metadata.Tags,
		ShellType:   sy.Spec.ShellType,
		Env:         sy.Spec.Env,
		Aliases:     sy.Spec.Aliases,
		Functions:   sy.Spec.Functions,
		PathPrepend: sy.Spec.PathPrepend,
		PathAppend:  sy.Spec.PathAppend,
		Options:     sy.Spec.Options,
		History:     sy.Spec.History,
		Keybindings: sy.Spec.Keybindings,
		RawConfig:   sy.Spec.RawConfig,
		Enabled:     enabled,
	}
}

// ToYAML converts a Shell to the ShellYAML format.
func (s *Shell) ToYAML() *ShellYAML {
	var enabledPtr *bool
	if !s.Enabled {
		enabledPtr = &s.Enabled
	}

	return &ShellYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalShell",
		Metadata: ShellMetadata{
			Name:        s.Name,
			Description: s.Description,
			Category:    s.Category,
			Tags:        s.Tags,
		},
		Spec: ShellSpec{
			ShellType:   s.ShellType,
			Env:         s.Env,
			Aliases:     s.Aliases,
			Functions:   s.Functions,
			PathPrepend: s.PathPrepend,
			PathAppend:  s.PathAppend,
			Options:     s.Options,
			History:     s.History,
			Keybindings: s.Keybindings,
			RawConfig:   s.RawConfig,
			Enabled:     enabledPtr,
		},
	}
}

// DefaultZshOptions returns commonly used zsh options.
func DefaultZshOptions() []string {
	return []string{
		"AUTO_CD",              // cd by typing directory name
		"AUTO_PUSHD",           // push directory onto stack on cd
		"PUSHD_IGNORE_DUPS",    // don't push duplicates
		"CORRECT",              // command correction
		"EXTENDED_GLOB",        // extended globbing
		"NO_BEEP",              // disable beep
		"INTERACTIVE_COMMENTS", // allow comments in interactive shell
	}
}

// DefaultHistoryConfig returns reasonable history defaults.
func DefaultHistoryConfig() *HistoryConfig {
	return &HistoryConfig{
		Size:           10000,
		File:           "${HOME}/.zsh_history",
		IgnoreDups:     true,
		IgnoreSpace:    true,
		ShareHistory:   true,
		ExtendedFormat: true,
	}
}

// CommonAliases returns commonly used shell aliases.
func CommonAliases() []Alias {
	return []Alias{
		{Name: "ll", Command: "ls -la", Description: "Long listing with hidden files"},
		{Name: "la", Command: "ls -A", Description: "List all except . and .."},
		{Name: "l", Command: "ls -CF", Description: "List in columns"},
		{Name: "..", Command: "cd ..", Description: "Go up one directory"},
		{Name: "...", Command: "cd ../..", Description: "Go up two directories"},
		{Name: "md", Command: "mkdir -p", Description: "Create directory with parents"},
		{Name: "rd", Command: "rmdir", Description: "Remove directory"},
	}
}

// DevAliases returns development-related aliases.
func DevAliases() []Alias {
	return []Alias{
		{Name: "g", Command: "git", Description: "Git shortcut"},
		{Name: "ga", Command: "git add", Description: "Git add"},
		{Name: "gc", Command: "git commit", Description: "Git commit"},
		{Name: "gco", Command: "git checkout", Description: "Git checkout"},
		{Name: "gd", Command: "git diff", Description: "Git diff"},
		{Name: "gl", Command: "git log --oneline", Description: "Git log oneline"},
		{Name: "gp", Command: "git push", Description: "Git push"},
		{Name: "gs", Command: "git status", Description: "Git status"},
		{Name: "k", Command: "kubectl", Description: "Kubectl shortcut"},
		{Name: "d", Command: "docker", Description: "Docker shortcut"},
		{Name: "dc", Command: "docker compose", Description: "Docker compose"},
	}
}
