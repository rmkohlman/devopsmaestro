package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// currentTheme holds the active theme (default: auto)
var currentTheme = GetTheme(ThemeAuto)

// SetTheme changes the active color theme
func SetTheme(name ThemeName) {
	currentTheme = GetTheme(name)
	refreshStyles()
}

// GetCurrentTheme returns the currently active theme
func GetCurrentTheme() Theme {
	return currentTheme
}

// InitTheme initializes the theme from environment or config
func InitTheme() {
	// Priority: 1. Environment variable, 2. Config file, 3. Default (auto)
	themeName := ThemeAuto

	// Check for DVM_THEME environment variable
	if envTheme := os.Getenv("DVM_THEME"); envTheme != "" {
		themeName = ThemeName(envTheme)
	}
	// Note: Config file loading happens in main.go via config.LoadConfig()
	// We don't import config here to avoid circular dependencies

	SetTheme(themeName)
}

// Color palette (dynamic based on theme)
var (
	// Primary colors
	PrimaryColor   lipgloss.TerminalColor
	SecondaryColor lipgloss.TerminalColor
	SuccessColor   lipgloss.TerminalColor
	WarningColor   lipgloss.TerminalColor
	ErrorColor     lipgloss.TerminalColor
	InfoColor      lipgloss.TerminalColor

	// Text colors
	TextColor      lipgloss.TerminalColor
	MutedColor     lipgloss.TerminalColor
	HighlightColor lipgloss.TerminalColor

	// Background colors
	BackgroundColor lipgloss.TerminalColor
)

// refreshStyles updates all color variables and styles from the current theme
func refreshStyles() {
	// Update colors from theme
	PrimaryColor = currentTheme.Primary
	SecondaryColor = currentTheme.Secondary
	SuccessColor = currentTheme.Success
	WarningColor = currentTheme.Warning
	ErrorColor = currentTheme.Error
	InfoColor = currentTheme.Info
	MutedColor = currentTheme.Muted
	HighlightColor = currentTheme.Highlight
	BackgroundColor = currentTheme.Background

	// Text color adapts based on terminal background
	TextColor = lipgloss.AdaptiveColor{
		Light: "#1F2937", // Dark text for light backgrounds
		Dark:  "#E5E7EB", // Light text for dark backgrounds
	}

	// Rebuild all styles with new colors
	rebuildStyles()
}

func init() {
	// Initialize theme and colors on package load
	InitTheme()
}

// rebuildStyles recreates all style variables with current theme colors
func rebuildStyles() {
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		PaddingRight(1)

	ActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(SuccessColor)

	TextStyle = lipgloss.NewStyle().
		Foreground(TextColor)

	MutedStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessColor)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorColor)

	WarningStyle = lipgloss.NewStyle().
		Foreground(WarningColor)

	InfoStyle = lipgloss.NewStyle().
		Foreground(InfoColor)

	StatusActiveStyle = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	StatusInactiveStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	PathStyle = lipgloss.NewStyle().
		Foreground(SecondaryColor)

	DateStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	CategoryStyle = lipgloss.NewStyle().
		Foreground(InfoColor)

	VersionStyle = lipgloss.NewStyle().
		Foreground(WarningColor)
}

// Text styles (initialized by rebuildStyles)
var (
	// Header style (bold, primary color)
	HeaderStyle lipgloss.Style

	// Active item style (with asterisk indicator)
	ActiveStyle lipgloss.Style

	// Regular text
	TextStyle lipgloss.Style

	// Muted text (for secondary info)
	MutedStyle lipgloss.Style

	// Success text
	SuccessStyle lipgloss.Style

	// Error text
	ErrorStyle lipgloss.Style

	// Warning text
	WarningStyle lipgloss.Style

	// Info text
	InfoStyle lipgloss.Style

	// Status: active
	StatusActiveStyle lipgloss.Style

	// Status: inactive
	StatusInactiveStyle lipgloss.Style

	// Path style
	PathStyle lipgloss.Style

	// Date/time style
	DateStyle lipgloss.Style

	// Category/tag style
	CategoryStyle lipgloss.Style

	// Version style
	VersionStyle lipgloss.Style
)

// Symbols
const (
	ActiveIndicator   = "● "
	InactiveIndicator = "○ "
	CheckMark         = "✓ "
	CrossMark         = "✗ "
	Arrow             = "→ "
	Bullet            = "• "
)

// Resource Icons (emojis for different resource types)
const (
	// Core Resources
	IconProject   = "📁"   // Project/folder
	IconWorkspace = "💼"   // Workspace/briefcase
	IconPlugin    = "🔌"   // Plugin/electrical plug
	IconNvim      = "✏️ " // Neovim/editor
	IconContext   = "🎯"   // Context/target

	// Build & Deploy
	IconBuild     = "🔨" // Build/hammer
	IconDocker    = "🐳" // Docker/whale
	IconContainer = "📦" // Container/package
	IconImage     = "💿" // Image/disk

	// Code & Development
	IconCode       = "💻" // Code/laptop
	IconGit        = "🌳" // Git/tree (or branch)
	IconLanguage   = "📝" // Language/document
	IconDependency = "🔗" // Dependency/link

	// Data & Storage
	IconDatabase  = "🗄️ " // Database/filing cabinet
	IconDataLake  = "🌊"   // Data Lake/water
	IconDataStore = "💾"   // Data Store/floppy disk
	IconRecord    = "📄"   // Record/page
	IconStorage   = "🗃️ " // Storage/card file box

	// Pipeline & Orchestration
	IconPipeline      = "⚡" // Pipeline/zap (or 🔄 cycle)
	IconWorkflow      = "🔄" // Workflow/counterclockwise arrows
	IconOrchestration = "🎼" // Orchestration/musical score
	IconTask          = "✅" // Task/check box
	IconPrototype     = "🧪" // Prototype/test tube

	// Status & State
	IconRunning = "▶️ " // Running/play
	IconStopped = "⏸️ " // Stopped/pause
	IconSuccess = "✅"   // Success/check
	IconError   = "❌"   // Error/cross
	IconWarning = "⚠️ " // Warning/warning sign
	IconInfo    = "ℹ️ " // Info/information

	// UI Elements
	IconSettings = "⚙️ " // Settings/gear
	IconSearch   = "🔍"   // Search/magnifying glass
	IconFile     = "📄"   // File/page
	IconFolder   = "📂"   // Folder/open folder
	IconUser     = "👤"   // User/bust in silhouette
	IconTime     = "🕐"   // Time/clock
	IconTag      = "🏷️ " // Tag/label
	IconCategory = "📋"   // Category/clipboard
)

// Helper functions

// RenderActive returns a styled string for active items
func RenderActive(text string) string {
	return ActiveStyle.Render(ActiveIndicator + text)
}

// RenderInactive returns a styled string for inactive items
func RenderInactive(text string) string {
	return TextStyle.Render(text)
}

// RenderHeader returns a styled header
func RenderHeader(text string) string {
	return HeaderStyle.Render(text)
}

// RenderSuccess returns a styled success message
func RenderSuccess(text string) string {
	return SuccessStyle.Render(CheckMark + text)
}

// RenderError returns a styled error message
func RenderError(text string) string {
	return ErrorStyle.Render(CrossMark + text)
}

// RenderStatus returns a styled status based on value
func RenderStatus(status string) string {
	switch status {
	case "active", "running", "ready":
		return StatusActiveStyle.Render(status)
	case "inactive", "stopped", "pending":
		return StatusInactiveStyle.Render(status)
	default:
		return TextStyle.Render(status)
	}
}

// Resource formatters with icons

// FormatProject formats a project name with icon
func FormatProject(name string, isActive bool) string {
	if isActive {
		return ActiveStyle.Render(IconProject + " " + name)
	}
	return IconProject + " " + name
}

// FormatWorkspace formats a workspace name with icon
func FormatWorkspace(name string, isActive bool) string {
	if isActive {
		return ActiveStyle.Render(IconWorkspace + " " + name)
	}
	return IconWorkspace + " " + name
}

// FormatPlugin formats a plugin name with icon
func FormatPlugin(name string, enabled bool) string {
	icon := IconPlugin
	if enabled {
		return SuccessStyle.Render(icon + " " + name)
	}
	return MutedStyle.Render(icon + " " + name)
}

// FormatContext formats context info with icon
func FormatContext(appName, workspaceName string) string {
	appStyle := lipgloss.NewStyle().Foreground(PrimaryColor)
	workspaceStyle := lipgloss.NewStyle().Foreground(SecondaryColor)
	return fmt.Sprintf("%s %s/%s",
		IconContext,
		appStyle.Render(appName),
		workspaceStyle.Render(workspaceName))
}

// FormatBuildStep formats a build step message
func FormatBuildStep(step string, success bool) string {
	if success {
		return SuccessStyle.Render(IconSuccess + " " + step)
	}
	return InfoStyle.Render(IconBuild + " " + step)
}

// FormatResourceType formats resource type headers
func FormatResourceType(resourceType string) string {
	icon := ""
	switch resourceType {
	case "project", "projects":
		icon = IconProject
	case "workspace", "workspaces":
		icon = IconWorkspace
	case "plugin", "plugins":
		icon = IconPlugin
	case "pipeline", "pipelines":
		icon = IconPipeline
	case "workflow", "workflows":
		icon = IconWorkflow
	case "task", "tasks":
		icon = IconTask
	case "database", "databases":
		icon = IconDatabase
	case "datalake", "datalakes":
		icon = IconDataLake
	case "datastore", "datastores":
		icon = IconDataStore
	case "storage":
		icon = IconStorage
	case "dependency", "dependencies":
		icon = IconDependency
	case "prototype", "prototypes":
		icon = IconPrototype
	default:
		icon = IconFile
	}
	return HeaderStyle.Render(icon + " " + strings.ToUpper(resourceType))
}

// ColorizeYAML adds syntax highlighting to YAML output
// Keys are rendered in one color, values in another
func ColorizeYAML(yamlContent string) string {
	lines := strings.Split(yamlContent, "\n")
	var result strings.Builder

	keyStyle := lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true) // Cyan, bold
	valueStyle := lipgloss.NewStyle().Foreground(HighlightColor)          // Yellow
	commentStyle := lipgloss.NewStyle().Foreground(MutedColor)            // Gray
	separatorStyle := lipgloss.NewStyle().Foreground(MutedColor)          // Gray for ---

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			result.WriteString("\n")
			continue
		}

		// Handle document separators
		if strings.TrimSpace(line) == "---" {
			result.WriteString(separatorStyle.Render(line) + "\n")
			continue
		}

		// Handle comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			result.WriteString(commentStyle.Render(line) + "\n")
			continue
		}

		// Handle key: value pairs
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				// Get leading whitespace
				leadingSpace := ""
				for _, ch := range line {
					if ch == ' ' {
						leadingSpace += " "
					} else {
						break
					}
				}

				key := strings.TrimSpace(parts[0])
				value := parts[1]

				// Render key in cyan (bold) and value in yellow
				if strings.TrimSpace(value) == "" {
					// Key with no value on same line (like "spec:")
					result.WriteString(leadingSpace + keyStyle.Render(key+":") + "\n")
				} else {
					// Key with value (like "name: telescope")
					result.WriteString(leadingSpace + keyStyle.Render(key+":") + valueStyle.Render(value) + "\n")
				}
				continue
			}
		}

		// Handle list items (lines starting with -)
		if strings.Contains(line, "-") && strings.HasPrefix(strings.TrimSpace(line), "-") {
			// Color list items in yellow
			result.WriteString(valueStyle.Render(line) + "\n")
			continue
		}

		// Default: render as-is
		result.WriteString(line + "\n")
	}

	return result.String()
}
