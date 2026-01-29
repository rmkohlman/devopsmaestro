package output

import "github.com/charmbracelet/lipgloss"

// ThemeName identifies a color theme
type ThemeName string

const (
	ThemeAuto       ThemeName = "auto"       // Adapts to terminal background
	ThemePlain      ThemeName = "plain"      // No colors
	ThemeCatppuccin ThemeName = "catppuccin" // Catppuccin Mocha (default)
	ThemeTokyoNight ThemeName = "tokyo-night"
	ThemeNord       ThemeName = "nord"
	ThemeDracula    ThemeName = "dracula"
	ThemeGruvbox    ThemeName = "gruvbox"
	ThemeMonokai    ThemeName = "monokai"
	ThemeSolarized  ThemeName = "solarized"
)

// Theme defines a color palette for output formatting.
// Themes are swappable and allow complete visual customization
// without changing the formatter logic.
type Theme interface {
	// Name returns the theme identifier
	Name() ThemeName

	// Colors
	Primary() lipgloss.TerminalColor    // Main brand/emphasis color
	Secondary() lipgloss.TerminalColor  // Secondary accent color
	Success() lipgloss.TerminalColor    // Success states (typically green)
	Warning() lipgloss.TerminalColor    // Warning states (typically yellow)
	Error() lipgloss.TerminalColor      // Error states (typically red)
	Info() lipgloss.TerminalColor       // Info states (typically blue)
	Muted() lipgloss.TerminalColor      // Dimmed/subtle text
	Highlight() lipgloss.TerminalColor  // Special highlights
	Text() lipgloss.TerminalColor       // Default text color
	Background() lipgloss.TerminalColor // Background color hints

	// Icons (can return empty strings for plain themes)
	IconSuccess() string
	IconWarning() string
	IconError() string
	IconInfo() string
	IconProgress() string
	IconBullet() string
	IconArrow() string
	IconCheck() string
	IconCross() string
	IconStar() string

	// Styles (pre-configured lipgloss styles)
	HeaderStyle() lipgloss.Style
	TitleStyle() lipgloss.Style
	SubtitleStyle() lipgloss.Style
	BodyStyle() lipgloss.Style
	CodeStyle() lipgloss.Style
	PathStyle() lipgloss.Style
	KeyStyle() lipgloss.Style
	ValueStyle() lipgloss.Style
	BorderStyle() lipgloss.Style
}

// Icons holds Unicode icons for rich output
type Icons struct {
	Success  string
	Warning  string
	Error    string
	Info     string
	Progress string
	Bullet   string
	Arrow    string
	Check    string
	Cross    string
	Star     string
	Folder   string
	File     string
	Git      string
	Docker   string
	Time     string
	Tag      string
	User     string
	Lock     string
	Unlock   string
	Link     string
	Plus     string
	Minus    string
	Edit     string
	Delete   string
	Refresh  string
	Search   string
	Filter   string
	Sort     string
	Settings string
	Help     string
	Debug    string
}

// DefaultIcons returns the default set of Unicode icons
func DefaultIcons() Icons {
	return Icons{
		Success:  "✓",
		Warning:  "⚠",
		Error:    "✗",
		Info:     "ℹ",
		Progress: "→",
		Bullet:   "•",
		Arrow:    "→",
		Check:    "✓",
		Cross:    "✗",
		Star:     "★",
		Folder:   "📁",
		File:     "📄",
		Git:      "",
		Docker:   "🐳",
		Time:     "⏱",
		Tag:      "🏷",
		User:     "👤",
		Lock:     "🔒",
		Unlock:   "🔓",
		Link:     "🔗",
		Plus:     "+",
		Minus:    "-",
		Edit:     "✏",
		Delete:   "🗑",
		Refresh:  "🔄",
		Search:   "🔍",
		Filter:   "🔽",
		Sort:     "↕",
		Settings: "⚙",
		Help:     "?",
		Debug:    "🐛",
	}
}

// NerdFontIcons returns icons using Nerd Font glyphs
// These require a Nerd Font to be installed
// See: https://www.nerdfonts.com/cheat-sheet
func NerdFontIcons() Icons {
	return Icons{
		Success:  "\uf00c", // nf-fa-check
		Warning:  "\uf071", // nf-fa-exclamation_triangle
		Error:    "\uf00d", // nf-fa-times
		Info:     "\uf05a", // nf-fa-info_circle
		Progress: "\uf061", // nf-fa-arrow_right
		Bullet:   "\uf111", // nf-fa-circle
		Arrow:    "\uf054", // nf-fa-chevron_right
		Check:    "\uf00c", // nf-fa-check
		Cross:    "\uf00d", // nf-fa-times
		Star:     "\uf005", // nf-fa-star
		Folder:   "\uf07b", // nf-fa-folder
		File:     "\uf15b", // nf-fa-file
		Git:      "\ue702", // nf-dev-git
		Docker:   "\uf308", // nf-linux-docker
		Time:     "\uf017", // nf-fa-clock_o
		Tag:      "\uf02b", // nf-fa-tag
		User:     "\uf007", // nf-fa-user
		Lock:     "\uf023", // nf-fa-lock
		Unlock:   "\uf09c", // nf-fa-unlock
		Link:     "\uf0c1", // nf-fa-link
		Plus:     "\uf067", // nf-fa-plus
		Minus:    "\uf068", // nf-fa-minus
		Edit:     "\uf044", // nf-fa-pencil_square_o
		Delete:   "\uf1f8", // nf-fa-trash
		Refresh:  "\uf021", // nf-fa-refresh
		Search:   "\uf002", // nf-fa-search
		Filter:   "\uf0b0", // nf-fa-filter
		Sort:     "\uf0dc", // nf-fa-sort
		Settings: "\uf013", // nf-fa-cog
		Help:     "\uf059", // nf-fa-question_circle
		Debug:    "\uf188", // nf-fa-bug
	}
}

// PlainIcons returns ASCII-only icons for maximum compatibility
func PlainIcons() Icons {
	return Icons{
		Success:  "[OK]",
		Warning:  "[!]",
		Error:    "[X]",
		Info:     "[i]",
		Progress: "->",
		Bullet:   "*",
		Arrow:    "->",
		Check:    "[x]",
		Cross:    "[ ]",
		Star:     "*",
		Folder:   "[D]",
		File:     "[F]",
		Git:      "[G]",
		Docker:   "[D]",
		Time:     "[T]",
		Tag:      "[#]",
		User:     "[@]",
		Lock:     "[L]",
		Unlock:   "[U]",
		Link:     "[>]",
		Plus:     "[+]",
		Minus:    "[-]",
		Edit:     "[E]",
		Delete:   "[D]",
		Refresh:  "[R]",
		Search:   "[?]",
		Filter:   "[F]",
		Sort:     "[S]",
		Settings: "[=]",
		Help:     "[?]",
		Debug:    "[D]",
	}
}
