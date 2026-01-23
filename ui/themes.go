package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// ThemeName represents available color themes
type ThemeName string

const (
	ThemeAuto            ThemeName = "auto"
	ThemeCatppuccinMocha ThemeName = "catppuccin-mocha"
	ThemeCatppuccinLatte ThemeName = "catppuccin-latte"
	ThemeTokyoNight      ThemeName = "tokyo-night"
	ThemeNord            ThemeName = "nord"
	ThemeDracula         ThemeName = "dracula"
	ThemeGruvboxDark     ThemeName = "gruvbox-dark"
	ThemeGruvboxLight    ThemeName = "gruvbox-light"
)

// Theme defines a color palette for the UI
type Theme struct {
	Name       ThemeName
	Primary    lipgloss.TerminalColor // Main brand color (headers, emphasis)
	Secondary  lipgloss.TerminalColor // Secondary accent (links, highlights)
	Success    lipgloss.TerminalColor // Success states (green)
	Warning    lipgloss.TerminalColor // Warning states (yellow)
	Error      lipgloss.TerminalColor // Error states (red)
	Info       lipgloss.TerminalColor // Info states (blue)
	Muted      lipgloss.TerminalColor // Dimmed text (gray)
	Highlight  lipgloss.TerminalColor // Special highlights
	Background lipgloss.TerminalColor // Background hints
}

// AvailableThemes returns a list of all available theme names
func AvailableThemes() []ThemeName {
	return []ThemeName{
		ThemeAuto,
		ThemeCatppuccinMocha,
		ThemeCatppuccinLatte,
		ThemeTokyoNight,
		ThemeNord,
		ThemeDracula,
		ThemeGruvboxDark,
		ThemeGruvboxLight,
	}
}

// GetTheme returns the specified theme, or auto theme if not found
func GetTheme(name ThemeName) Theme {
	switch name {
	case ThemeCatppuccinMocha:
		return catppuccinMochaTheme()
	case ThemeCatppuccinLatte:
		return catppuccinLatteTheme()
	case ThemeTokyoNight:
		return tokyoNightTheme()
	case ThemeNord:
		return nordTheme()
	case ThemeDracula:
		return draculaTheme()
	case ThemeGruvboxDark:
		return gruvboxDarkTheme()
	case ThemeGruvboxLight:
		return gruvboxLightTheme()
	case ThemeAuto:
		fallthrough
	default:
		return autoTheme()
	}
}

// autoTheme adapts to terminal light/dark background automatically
func autoTheme() Theme {
	return Theme{
		Name: ThemeAuto,
		Primary: lipgloss.AdaptiveColor{
			Light: "#5A4FCF", // Darker purple for light backgrounds
			Dark:  "#CBA6F7", // Lighter purple for dark backgrounds
		},
		Secondary: lipgloss.AdaptiveColor{
			Light: "#0284C7", // Darker cyan for light backgrounds
			Dark:  "#89DCEB", // Lighter cyan for dark backgrounds
		},
		Success: lipgloss.AdaptiveColor{
			Light: "#16A34A", // Darker green
			Dark:  "#A6E3A1", // Lighter green
		},
		Warning: lipgloss.AdaptiveColor{
			Light: "#CA8A04", // Darker yellow
			Dark:  "#F9E2AF", // Lighter yellow
		},
		Error: lipgloss.AdaptiveColor{
			Light: "#DC2626", // Darker red
			Dark:  "#F38BA8", // Lighter red
		},
		Info: lipgloss.AdaptiveColor{
			Light: "#2563EB", // Darker blue
			Dark:  "#89B4FA", // Lighter blue
		},
		Muted: lipgloss.AdaptiveColor{
			Light: "#6B7280", // Medium gray
			Dark:  "#6C7086", // Lighter gray
		},
		Highlight: lipgloss.AdaptiveColor{
			Light: "#D97706", // Darker orange
			Dark:  "#FAB387", // Lighter orange
		},
		Background: lipgloss.AdaptiveColor{
			Light: "#F3F4F6", // Light gray
			Dark:  "#1E1E2E", // Dark background
		},
	}
}

// catppuccinMochaTheme - Catppuccin Mocha (dark, soothing pastel)
// Reference: https://github.com/catppuccin/catppuccin
func catppuccinMochaTheme() Theme {
	return Theme{
		Name:       ThemeCatppuccinMocha,
		Primary:    lipgloss.Color("#CBA6F7"), // Mauve
		Secondary:  lipgloss.Color("#89DCEB"), // Sky
		Success:    lipgloss.Color("#A6E3A1"), // Green
		Warning:    lipgloss.Color("#F9E2AF"), // Yellow
		Error:      lipgloss.Color("#F38BA8"), // Red
		Info:       lipgloss.Color("#89B4FA"), // Blue
		Muted:      lipgloss.Color("#6C7086"), // Overlay0
		Highlight:  lipgloss.Color("#FAB387"), // Peach
		Background: lipgloss.Color("#1E1E2E"), // Base
	}
}

// catppuccinLatteTheme - Catppuccin Latte (light, warm pastel)
func catppuccinLatteTheme() Theme {
	return Theme{
		Name:       ThemeCatppuccinLatte,
		Primary:    lipgloss.Color("#8839EF"), // Mauve
		Secondary:  lipgloss.Color("#04A5E5"), // Sky
		Success:    lipgloss.Color("#40A02B"), // Green
		Warning:    lipgloss.Color("#DF8E1D"), // Yellow
		Error:      lipgloss.Color("#D20F39"), // Red
		Info:       lipgloss.Color("#1E66F5"), // Blue
		Muted:      lipgloss.Color("#9CA0B0"), // Overlay0
		Highlight:  lipgloss.Color("#FE640B"), // Peach
		Background: lipgloss.Color("#EFF1F5"), // Base
	}
}

// tokyoNightTheme - Tokyo Night (dark, vibrant blue-purple)
// Reference: https://github.com/folke/tokyonight.nvim
func tokyoNightTheme() Theme {
	return Theme{
		Name:       ThemeTokyoNight,
		Primary:    lipgloss.Color("#BB9AF7"), // Purple
		Secondary:  lipgloss.Color("#7DCFFF"), // Cyan
		Success:    lipgloss.Color("#9ECE6A"), // Green
		Warning:    lipgloss.Color("#E0AF68"), // Yellow
		Error:      lipgloss.Color("#F7768E"), // Red
		Info:       lipgloss.Color("#7AA2F7"), // Blue
		Muted:      lipgloss.Color("#565F89"), // Comment
		Highlight:  lipgloss.Color("#FF9E64"), // Orange
		Background: lipgloss.Color("#1A1B26"), // Background
	}
}

// nordTheme - Nord (cool, bluish, minimal)
// Reference: https://www.nordtheme.com
func nordTheme() Theme {
	return Theme{
		Name:       ThemeNord,
		Primary:    lipgloss.Color("#B48EAD"), // Nord15 (Purple)
		Secondary:  lipgloss.Color("#88C0D0"), // Nord8 (Cyan)
		Success:    lipgloss.Color("#A3BE8C"), // Nord14 (Green)
		Warning:    lipgloss.Color("#EBCB8B"), // Nord13 (Yellow)
		Error:      lipgloss.Color("#BF616A"), // Nord11 (Red)
		Info:       lipgloss.Color("#81A1C1"), // Nord9 (Blue)
		Muted:      lipgloss.Color("#4C566A"), // Nord3 (Gray)
		Highlight:  lipgloss.Color("#D08770"), // Nord12 (Orange)
		Background: lipgloss.Color("#2E3440"), // Nord0 (Background)
	}
}

// draculaTheme - Dracula (dark, vibrant purple-pink)
// Reference: https://draculatheme.com
func draculaTheme() Theme {
	return Theme{
		Name:       ThemeDracula,
		Primary:    lipgloss.Color("#BD93F9"), // Purple
		Secondary:  lipgloss.Color("#8BE9FD"), // Cyan
		Success:    lipgloss.Color("#50FA7B"), // Green
		Warning:    lipgloss.Color("#F1FA8C"), // Yellow
		Error:      lipgloss.Color("#FF5555"), // Red
		Info:       lipgloss.Color("#6272A4"), // Comment (blue-ish)
		Muted:      lipgloss.Color("#6272A4"), // Comment
		Highlight:  lipgloss.Color("#FFB86C"), // Orange
		Background: lipgloss.Color("#282A36"), // Background
	}
}

// gruvboxDarkTheme - Gruvbox Dark (warm, retro)
// Reference: https://github.com/morhetz/gruvbox
func gruvboxDarkTheme() Theme {
	return Theme{
		Name:       ThemeGruvboxDark,
		Primary:    lipgloss.Color("#D3869B"), // Purple
		Secondary:  lipgloss.Color("#8EC07C"), // Aqua
		Success:    lipgloss.Color("#B8BB26"), // Green
		Warning:    lipgloss.Color("#FABD2F"), // Yellow
		Error:      lipgloss.Color("#FB4934"), // Red
		Info:       lipgloss.Color("#83A598"), // Blue
		Muted:      lipgloss.Color("#928374"), // Gray
		Highlight:  lipgloss.Color("#FE8019"), // Orange
		Background: lipgloss.Color("#282828"), // Background
	}
}

// gruvboxLightTheme - Gruvbox Light (warm, retro, light)
func gruvboxLightTheme() Theme {
	return Theme{
		Name:       ThemeGruvboxLight,
		Primary:    lipgloss.Color("#8F3F71"), // Purple
		Secondary:  lipgloss.Color("#427B58"), // Aqua
		Success:    lipgloss.Color("#79740E"), // Green
		Warning:    lipgloss.Color("#B57614"), // Yellow
		Error:      lipgloss.Color("#9D0006"), // Red
		Info:       lipgloss.Color("#076678"), // Blue
		Muted:      lipgloss.Color("#7C6F64"), // Gray
		Highlight:  lipgloss.Color("#AF3A03"), // Orange
		Background: lipgloss.Color("#FBF1C7"), // Background
	}
}
