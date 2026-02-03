// Package theme provides types and utilities for Neovim theme management.
package theme

import (
	"devopsmaestro/pkg/palette"
)

// Theme represents a Neovim colorscheme configuration.
type Theme struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string            `yaml:"author,omitempty" json:"author,omitempty"`
	Category    string            `yaml:"category,omitempty" json:"category,omitempty"`
	Plugin      ThemePlugin       `yaml:"plugin" json:"plugin"`
	Style       string            `yaml:"style,omitempty" json:"style,omitempty"`
	Transparent bool              `yaml:"transparent,omitempty" json:"transparent,omitempty"`
	Colors      map[string]string `yaml:"colors,omitempty" json:"colors,omitempty"`
	Options     map[string]any    `yaml:"options,omitempty" json:"options,omitempty"`
}

// ThemePlugin defines the colorscheme plugin to use.
type ThemePlugin struct {
	Repo   string `yaml:"repo" json:"repo"`
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
	Tag    string `yaml:"tag,omitempty" json:"tag,omitempty"`
}

// ThemeYAML represents the full YAML structure with metadata.
type ThemeYAML struct {
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
	Plugin      ThemePlugin       `yaml:"plugin"`
	Style       string            `yaml:"style,omitempty"`
	Transparent bool              `yaml:"transparent,omitempty"`
	Colors      map[string]string `yaml:"colors,omitempty"`
	Options     map[string]any    `yaml:"options,omitempty"`
}

// =============================================================================
// Color Constants (re-exported from palette package for convenience)
// =============================================================================
// These constants are re-exported from pkg/palette so existing code continues
// to work without changes. New code should import pkg/palette directly.

// Background colors
const (
	ColorBg           = palette.ColorBg
	ColorBgDark       = palette.ColorBgDark
	ColorBgHighlight  = palette.ColorBgHighlight
	ColorBgSearch     = palette.ColorBgSearch
	ColorBgVisual     = palette.ColorBgVisual
	ColorBgFloat      = palette.ColorBgFloat
	ColorBgPopup      = palette.ColorBgPopup
	ColorBgSidebar    = palette.ColorBgSidebar
	ColorBgStatusline = palette.ColorBgStatusline
)

// Foreground colors
const (
	ColorFg        = palette.ColorFg
	ColorFgDark    = palette.ColorFgDark
	ColorFgGutter  = palette.ColorFgGutter
	ColorFgSidebar = palette.ColorFgSidebar
)

// UI and diagnostic colors
const (
	ColorBorder  = palette.ColorBorder
	ColorComment = palette.ColorComment
	ColorError   = palette.ColorError
	ColorWarning = palette.ColorWarning
	ColorInfo    = palette.ColorInfo
	ColorHint    = palette.ColorHint
)

// Theme categories (re-exported from palette)
const (
	CategoryDark  = string(palette.CategoryDark)
	CategoryLight = string(palette.CategoryLight)
	CategoryBoth  = string(palette.CategoryBoth)
)

// Supported theme plugins with their setup function names
var SupportedThemePlugins = map[string]string{
	"folke/tokyonight.nvim":       "tokyonight",
	"catppuccin/nvim":             "catppuccin",
	"ellisonleao/gruvbox.nvim":    "gruvbox",
	"shaunsingh/nord.nvim":        "nord",
	"rebelot/kanagawa.nvim":       "kanagawa",
	"rose-pine/neovim":            "rose-pine",
	"EdenEast/nightfox.nvim":      "nightfox",
	"navarasu/onedark.nvim":       "onedark",
	"Mofiqul/dracula.nvim":        "dracula",
	"sainnhe/everforest":          "everforest",
	"sainnhe/sonokai":             "sonokai",
	"projekt0n/github-nvim-theme": "github-theme",
}

// GetSetupName returns the setup function name for a theme plugin.
func GetSetupName(repo string) string {
	if name, ok := SupportedThemePlugins[repo]; ok {
		return name
	}
	return ""
}

// =============================================================================
// Palette Integration
// =============================================================================

// ToPalette converts the theme to a generic palette that can be used
// by other tools like terminal configuration.
func (t *Theme) ToPalette() *palette.Palette {
	if t == nil {
		return nil
	}

	p := &palette.Palette{
		Name:        t.Name,
		Description: t.Description,
		Author:      t.Author,
	}

	// Map category
	switch t.Category {
	case CategoryDark:
		p.Category = palette.CategoryDark
	case CategoryLight:
		p.Category = palette.CategoryLight
	case CategoryBoth:
		p.Category = palette.CategoryBoth
	}

	// Copy colors
	if t.Colors != nil {
		p.Colors = make(map[string]string, len(t.Colors))
		for k, v := range t.Colors {
			p.Colors[k] = v
		}
	}

	return p
}

// ToTerminalColors extracts terminal-compatible ANSI colors from the theme.
// This is a convenience method that calls ToPalette().ToTerminalColors().
func (t *Theme) ToTerminalColors() map[string]string {
	p := t.ToPalette()
	if p == nil {
		return nil
	}
	return p.ToTerminalColors()
}
