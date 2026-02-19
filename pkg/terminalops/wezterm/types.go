// Package wezterm provides types and utilities for WezTerm terminal configuration management.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package wezterm

import (
	"time"
)

// WeztermYAML is the kubectl-style YAML wrapper for WezTerm configurations.
type WeztermYAML struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"` // "WeztermConfig"
	Metadata   WeztermMetadata `yaml:"metadata" json:"metadata"`
	Spec       WeztermSpec     `yaml:"spec" json:"spec"`
}

// WeztermMetadata contains WezTerm configuration metadata in the YAML format.
type WeztermMetadata struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Category    string            `yaml:"category,omitempty" json:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// WeztermSpec contains the WezTerm specification in the YAML format.
type WeztermSpec struct {
	Font       FontConfig              `yaml:"font" json:"font"`
	Window     WindowConfig            `yaml:"window" json:"window"`
	Colors     *ColorConfig            `yaml:"colors,omitempty" json:"colors,omitempty"`
	ThemeRef   string                  `yaml:"themeRef,omitempty" json:"themeRef,omitempty"`
	Leader     *LeaderKey              `yaml:"leader,omitempty" json:"leader,omitempty"`
	Keys       []Keybinding            `yaml:"keys,omitempty" json:"keys,omitempty"`
	KeyTables  map[string][]Keybinding `yaml:"keyTables,omitempty" json:"keyTables,omitempty"`
	TabBar     *TabBarConfig           `yaml:"tabBar,omitempty" json:"tabBar,omitempty"`
	Pane       *PaneConfig             `yaml:"pane,omitempty" json:"pane,omitempty"`
	Plugins    []PluginConfig          `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	Scrollback int                     `yaml:"scrollback,omitempty" json:"scrollback,omitempty"`
	Workspace  string                  `yaml:"workspace,omitempty" json:"workspace,omitempty"`
	Enabled    *bool                   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// FontConfig represents font configuration.
type FontConfig struct {
	Family string  `yaml:"family" json:"family"`
	Size   float64 `yaml:"size" json:"size"`
}

// WindowConfig represents window configuration.
type WindowConfig struct {
	Opacity       float64 `yaml:"opacity" json:"opacity"`
	Blur          int     `yaml:"blur,omitempty" json:"blur,omitempty"`
	Decorations   string  `yaml:"decorations,omitempty" json:"decorations,omitempty"`
	InitialRows   int     `yaml:"initialRows,omitempty" json:"initialRows,omitempty"`
	InitialCols   int     `yaml:"initialCols,omitempty" json:"initialCols,omitempty"`
	CloseOnExit   string  `yaml:"closeOnExit,omitempty" json:"closeOnExit,omitempty"`
	PaddingLeft   int     `yaml:"paddingLeft,omitempty" json:"paddingLeft,omitempty"`
	PaddingRight  int     `yaml:"paddingRight,omitempty" json:"paddingRight,omitempty"`
	PaddingTop    int     `yaml:"paddingTop,omitempty" json:"paddingTop,omitempty"`
	PaddingBottom int     `yaml:"paddingBottom,omitempty" json:"paddingBottom,omitempty"`
}

// ColorConfig represents color scheme configuration.
type ColorConfig struct {
	Foreground   string   `yaml:"foreground" json:"foreground"`
	Background   string   `yaml:"background" json:"background"`
	CursorBg     string   `yaml:"cursor_bg" json:"cursor_bg"`
	CursorFg     string   `yaml:"cursor_fg" json:"cursor_fg"`
	CursorBorder string   `yaml:"cursor_border,omitempty" json:"cursor_border,omitempty"`
	SelectionBg  string   `yaml:"selection_bg" json:"selection_bg"`
	SelectionFg  string   `yaml:"selection_fg" json:"selection_fg"`
	ANSI         []string `yaml:"ansi" json:"ansi"`       // 8 colors
	Brights      []string `yaml:"brights" json:"brights"` // 8 colors
}

// LeaderKey represents the leader key configuration.
type LeaderKey struct {
	Key     string `yaml:"key" json:"key"`
	Mods    string `yaml:"mods" json:"mods"`
	Timeout int    `yaml:"timeout" json:"timeout"` // milliseconds
}

// Keybinding represents a key binding configuration.
type Keybinding struct {
	Key    string `yaml:"key" json:"key"`
	Mods   string `yaml:"mods,omitempty" json:"mods,omitempty"`
	Action string `yaml:"action" json:"action"`
	Args   any    `yaml:"args,omitempty" json:"args,omitempty"`
}

// TabBarConfig represents tab bar configuration.
type TabBarConfig struct {
	Enabled          bool   `yaml:"enabled" json:"enabled"`
	Position         string `yaml:"position,omitempty" json:"position,omitempty"` // "Top", "Bottom"
	MaxWidth         int    `yaml:"maxWidth,omitempty" json:"maxWidth,omitempty"`
	ShowNewTab       bool   `yaml:"showNewTab,omitempty" json:"showNewTab,omitempty"`
	FancyTabBar      bool   `yaml:"fancyTabBar,omitempty" json:"fancyTabBar,omitempty"`
	HideTabBarIfOnly bool   `yaml:"hideTabBarIfOnly,omitempty" json:"hideTabBarIfOnly,omitempty"`
}

// PaneConfig represents pane configuration.
type PaneConfig struct {
	InactiveSaturation float64 `yaml:"inactiveSaturation,omitempty" json:"inactiveSaturation,omitempty"`
	InactiveBrightness float64 `yaml:"inactiveBrightness,omitempty" json:"inactiveBrightness,omitempty"`
}

// PluginConfig represents plugin configuration.
type PluginConfig struct {
	Name   string         `yaml:"name" json:"name"`
	Source string         `yaml:"source" json:"source"`
	Config map[string]any `yaml:"config,omitempty" json:"config,omitempty"`
}

// WezTerm represents a WezTerm configuration.
// This is the canonical type used throughout terminal-manager.
type WezTerm struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Configuration
	Font       FontConfig              `json:"font" yaml:"font"`
	Window     WindowConfig            `json:"window" yaml:"window"`
	Colors     *ColorConfig            `json:"colors,omitempty" yaml:"colors,omitempty"`
	ThemeRef   string                  `json:"theme_ref,omitempty" yaml:"theme_ref,omitempty"`
	Leader     *LeaderKey              `json:"leader,omitempty" yaml:"leader,omitempty"`
	Keys       []Keybinding            `json:"keys,omitempty" yaml:"keys,omitempty"`
	KeyTables  map[string][]Keybinding `json:"key_tables,omitempty" yaml:"key_tables,omitempty"`
	TabBar     *TabBarConfig           `json:"tab_bar,omitempty" yaml:"tab_bar,omitempty"`
	Pane       *PaneConfig             `json:"pane,omitempty" yaml:"pane,omitempty"`
	Plugins    []PluginConfig          `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	Scrollback int                     `json:"scrollback,omitempty" yaml:"scrollback,omitempty"`
	Workspace  string                  `json:"workspace,omitempty" yaml:"workspace,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// NewWezTerm creates a new WezTerm configuration with default values.
func NewWezTerm(name string) *WezTerm {
	return &WezTerm{
		Name:    name,
		Enabled: true,
		Font: FontConfig{
			Family: "MesloLGS Nerd Font Mono",
			Size:   14,
		},
		Window: WindowConfig{
			Opacity: 1.0,
		},
	}
}

// NewWeztermYAML creates a new WeztermYAML with proper defaults.
func NewWeztermYAML(name string) *WeztermYAML {
	return &WeztermYAML{
		APIVersion: "devopsmaestro.dev/v1alpha1",
		Kind:       "WeztermConfig",
		Metadata: WeztermMetadata{
			Name: name,
		},
		Spec: WeztermSpec{
			Font: FontConfig{
				Family: "MesloLGS Nerd Font Mono",
				Size:   14,
			},
			Window: WindowConfig{
				Opacity: 1.0,
			},
		},
	}
}

// ToWezTerm converts WeztermYAML to the canonical WezTerm type.
func (wy *WeztermYAML) ToWezTerm() *WezTerm {
	// Default to enabled unless explicitly set to false
	enabled := true
	if wy.Spec.Enabled != nil {
		enabled = *wy.Spec.Enabled
	}

	w := &WezTerm{
		Name:        wy.Metadata.Name,
		Description: wy.Metadata.Description,
		Category:    wy.Metadata.Category,
		Tags:        wy.Metadata.Tags,
		Font:        wy.Spec.Font,
		Window:      wy.Spec.Window,
		Colors:      wy.Spec.Colors,
		ThemeRef:    wy.Spec.ThemeRef,
		Leader:      wy.Spec.Leader,
		Keys:        wy.Spec.Keys,
		KeyTables:   wy.Spec.KeyTables,
		TabBar:      wy.Spec.TabBar,
		Pane:        wy.Spec.Pane,
		Plugins:     wy.Spec.Plugins,
		Scrollback:  wy.Spec.Scrollback,
		Workspace:   wy.Spec.Workspace,
		Enabled:     enabled,
	}

	return w
}

// ToYAML converts a WezTerm to the WeztermYAML format.
func (w *WezTerm) ToYAML() *WeztermYAML {
	// Only include enabled field if disabled (to avoid cluttering YAML)
	var enabledPtr *bool
	if !w.Enabled {
		enabledPtr = &w.Enabled
	}

	wy := &WeztermYAML{
		APIVersion: "devopsmaestro.dev/v1alpha1",
		Kind:       "WeztermConfig",
		Metadata: WeztermMetadata{
			Name:        w.Name,
			Description: w.Description,
			Category:    w.Category,
			Tags:        w.Tags,
		},
		Spec: WeztermSpec{
			Font:       w.Font,
			Window:     w.Window,
			Colors:     w.Colors,
			ThemeRef:   w.ThemeRef,
			Leader:     w.Leader,
			Keys:       w.Keys,
			KeyTables:  w.KeyTables,
			TabBar:     w.TabBar,
			Pane:       w.Pane,
			Plugins:    w.Plugins,
			Scrollback: w.Scrollback,
			Workspace:  w.Workspace,
			Enabled:    enabledPtr,
		},
	}

	return wy
}

// HasColors returns true if the config has colors defined.
func (w *WezTerm) HasColors() bool {
	return w.Colors != nil
}

// HasThemeRef returns true if the config references a theme.
func (w *WezTerm) HasThemeRef() bool {
	return w.ThemeRef != ""
}

// HasLeader returns true if the config has a leader key defined.
func (w *WezTerm) HasLeader() bool {
	return w.Leader != nil
}

// HasKeybindings returns true if the config has custom keybindings.
func (w *WezTerm) HasKeybindings() bool {
	return len(w.Keys) > 0
}

// HasKeyTables returns true if the config has key tables defined.
func (w *WezTerm) HasKeyTables() bool {
	return len(w.KeyTables) > 0
}

// HasPlugins returns true if the config has plugins defined.
func (w *WezTerm) HasPlugins() bool {
	return len(w.Plugins) > 0
}
