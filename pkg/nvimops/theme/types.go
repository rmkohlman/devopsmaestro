// Package theme provides types and utilities for Neovim theme management.
package theme

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

// Common color keys used by themes
const (
	ColorBg           = "bg"
	ColorBgDark       = "bg_dark"
	ColorBgHighlight  = "bg_highlight"
	ColorBgSearch     = "bg_search"
	ColorBgVisual     = "bg_visual"
	ColorBgFloat      = "bg_float"
	ColorBgPopup      = "bg_popup"
	ColorBgSidebar    = "bg_sidebar"
	ColorBgStatusline = "bg_statusline"
	ColorFg           = "fg"
	ColorFgDark       = "fg_dark"
	ColorFgGutter     = "fg_gutter"
	ColorFgSidebar    = "fg_sidebar"
	ColorBorder       = "border"
	ColorComment      = "comment"
	ColorError        = "error"
	ColorWarning      = "warning"
	ColorInfo         = "info"
	ColorHint         = "hint"
)

// Theme categories
const (
	CategoryDark  = "dark"
	CategoryLight = "light"
	CategoryBoth  = "both"
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
