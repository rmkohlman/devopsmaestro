// Package theme provides types and utilities for Neovim theme management.
package theme

import (
	"fmt"
	"strings"
)

// TerminalEnvVars generates environment variables for terminal color configuration.
// These vars can be used by shell scripts, zsh-syntax-highlighting, etc.
// The prefix is "DVM_COLOR_" by default.
//
// Example output:
//
//	DVM_COLOR_BG=#1a1b26
//	DVM_COLOR_FG=#c0caf5
//	DVM_COLOR_RED=#f7768e
//	DVM_COLOR_GREEN=#9ece6a
//	... (all 16 ANSI colors + UI colors)
//	DVM_THEME=tokyonight-night
func (t *Theme) TerminalEnvVars() map[string]string {
	if t == nil {
		return make(map[string]string)
	}

	envVars := make(map[string]string)

	// Get terminal colors from the theme
	terminalColors := t.ToTerminalColors()

	// Convert terminal color keys to environment variable names
	for termKey, color := range terminalColors {
		envKey := terminalKeyToEnvKey(termKey)
		if envKey != "" {
			envVars[envKey] = color
		}
	}

	// Add theme metadata
	envVars["DVM_THEME"] = t.Name
	if t.Category != "" {
		envVars["DVM_THEME_CATEGORY"] = t.Category
	} else {
		envVars["DVM_THEME_CATEGORY"] = "unknown"
	}

	return envVars
}

// terminalKeyToEnvKey converts terminal color keys to environment variable names.
// Examples:
//
//	ansi_red -> DVM_COLOR_RED
//	ansi_bright_red -> DVM_COLOR_BRIGHT_RED
//	bg -> DVM_COLOR_BG
//	fg -> DVM_COLOR_FG
func terminalKeyToEnvKey(termKey string) string {
	switch termKey {
	// ANSI colors - normal
	case "ansi_black":
		return "DVM_COLOR_BLACK"
	case "ansi_red":
		return "DVM_COLOR_RED"
	case "ansi_green":
		return "DVM_COLOR_GREEN"
	case "ansi_yellow":
		return "DVM_COLOR_YELLOW"
	case "ansi_blue":
		return "DVM_COLOR_BLUE"
	case "ansi_magenta":
		return "DVM_COLOR_MAGENTA"
	case "ansi_cyan":
		return "DVM_COLOR_CYAN"
	case "ansi_white":
		return "DVM_COLOR_WHITE"

	// ANSI colors - bright
	case "ansi_bright_black":
		return "DVM_COLOR_BRIGHT_BLACK"
	case "ansi_bright_red":
		return "DVM_COLOR_BRIGHT_RED"
	case "ansi_bright_green":
		return "DVM_COLOR_BRIGHT_GREEN"
	case "ansi_bright_yellow":
		return "DVM_COLOR_BRIGHT_YELLOW"
	case "ansi_bright_blue":
		return "DVM_COLOR_BRIGHT_BLUE"
	case "ansi_bright_magenta":
		return "DVM_COLOR_BRIGHT_MAGENTA"
	case "ansi_bright_cyan":
		return "DVM_COLOR_BRIGHT_CYAN"
	case "ansi_bright_white":
		return "DVM_COLOR_BRIGHT_WHITE"

	// UI colors
	case "bg":
		return "DVM_COLOR_BG"
	case "fg":
		return "DVM_COLOR_FG"
	case "cursor":
		return "DVM_COLOR_CURSOR"
	case "cursor_text":
		return "DVM_COLOR_CURSOR_TEXT"
	case "selection":
		return "DVM_COLOR_SELECTION"
	case "selection_text":
		return "DVM_COLOR_SELECTION_TEXT"

	default:
		// For any other keys, try to convert them to uppercase env var format
		// Remove common prefixes and convert to uppercase
		envKey := strings.ToUpper(termKey)
		envKey = strings.ReplaceAll(envKey, "_", "_")
		if strings.HasPrefix(envKey, "ANSI_") {
			envKey = strings.TrimPrefix(envKey, "ANSI_")
		}
		return "DVM_COLOR_" + envKey
	}
}

// GetTerminalEnvVarsForTheme loads a theme by name and returns terminal env vars.
// It tries the provided Store first. If no store is provided, this function
// will return an error. Callers should use the library package directly if needed.
func GetTerminalEnvVarsForTheme(store Store, themeName string) (map[string]string, error) {
	if store == nil {
		return nil, fmt.Errorf("no theme store provided")
	}

	theme, err := store.Get(themeName)
	if err != nil {
		return nil, fmt.Errorf("failed to load theme %q: %w", themeName, err)
	}

	return theme.TerminalEnvVars(), nil
}
