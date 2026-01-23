package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// TableRenderer handles beautiful table rendering with colored columns
type TableRenderer struct {
	headers []string
	rows    [][]string
	styles  *TableStyles
}

// TableStyles holds styling configuration for tables
type TableStyles struct {
	HeaderStyle       lipgloss.Style
	CellStyle         lipgloss.Style
	ActiveRowStyle    lipgloss.Style
	AlternateRowStyle lipgloss.Style
	BorderStyle       lipgloss.Style
	ColumnStyles      []lipgloss.Style // Per-column styles
}

// NewTableRenderer creates a new table renderer
func NewTableRenderer(headers []string) *TableRenderer {
	return &TableRenderer{
		headers: headers,
		rows:    [][]string{},
		styles:  DefaultTableStyles(),
	}
}

// DefaultTableStyles returns default table styling
func DefaultTableStyles() *TableStyles {
	return &TableStyles{
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Padding(0, 1),
		CellStyle: lipgloss.NewStyle().
			Foreground(TextColor).
			Padding(0, 1),
		ActiveRowStyle: lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true),
		AlternateRowStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")),
		BorderStyle: lipgloss.NewStyle().
			Foreground(MutedColor),
		ColumnStyles: []lipgloss.Style{},
	}
}

// AddRow adds a row to the table
func (tr *TableRenderer) AddRow(cells ...string) {
	tr.rows = append(tr.rows, cells)
}

// SetColumnStyle sets a custom style for a specific column
func (tr *TableRenderer) SetColumnStyle(columnIndex int, style lipgloss.Style) {
	// Ensure we have enough column styles
	for len(tr.styles.ColumnStyles) <= columnIndex {
		tr.styles.ColumnStyles = append(tr.styles.ColumnStyles, tr.styles.CellStyle)
	}
	tr.styles.ColumnStyles[columnIndex] = style
}

// SetColumnStyles sets styles for all columns
func (tr *TableRenderer) SetColumnStyles(styles []lipgloss.Style) {
	tr.styles.ColumnStyles = styles
}

// Render renders the table using lipgloss table
func (tr *TableRenderer) Render() string {
	if len(tr.rows) == 0 {
		return MutedStyle.Render("No data available")
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(tr.styles.BorderStyle).
		Headers(tr.headers...).
		Rows(tr.rows...)

	// Apply header style
	t = t.StyleFunc(func(row, col int) lipgloss.Style {
		// Header row
		if row == 0 {
			return tr.styles.HeaderStyle
		}

		// Apply column-specific styles if available
		if col < len(tr.styles.ColumnStyles) {
			return tr.styles.ColumnStyles[col]
		}

		// Alternate row colors
		if row%2 == 0 {
			return tr.styles.CellStyle
		}
		return tr.styles.AlternateRowStyle
	})

	return t.String()
}

// RenderSimple renders a simple table without borders (like tabwriter)
func (tr *TableRenderer) RenderSimple() string {
	if len(tr.rows) == 0 {
		return MutedStyle.Render("No data available")
	}

	// Calculate column widths
	colWidths := make([]int, len(tr.headers))
	for i, h := range tr.headers {
		colWidths[i] = len(h)
	}
	for _, row := range tr.rows {
		for i, cell := range row {
			// Remove ANSI codes for width calculation
			cleanCell := stripANSI(cell)
			if len(cleanCell) > colWidths[i] {
				colWidths[i] = len(cleanCell)
			}
		}
	}

	var output strings.Builder

	// Render headers
	for i, h := range tr.headers {
		styled := tr.styles.HeaderStyle.Render(h)
		output.WriteString(styled)
		if i < len(tr.headers)-1 {
			output.WriteString(strings.Repeat(" ", colWidths[i]-len(h)+3))
		}
	}
	output.WriteString("\n")

	// Render rows
	for _, row := range tr.rows {
		for i, cell := range row {
			// Apply column style if available
			var styled string
			if i < len(tr.styles.ColumnStyles) {
				styled = tr.styles.ColumnStyles[i].Render(cell)
			} else {
				styled = tr.styles.CellStyle.Render(cell)
			}

			output.WriteString(styled)
			if i < len(row)-1 {
				cleanCell := stripANSI(cell)
				output.WriteString(strings.Repeat(" ", colWidths[i]-len(cleanCell)+3))
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}

// stripANSI removes ANSI escape codes for accurate length calculation
func stripANSI(str string) string {
	// Simple ANSI stripping (lipgloss handles this better internally)
	result := ""
	inEscape := false
	for _, r := range str {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// ProjectsTableRenderer creates a styled table for projects
func ProjectsTableRenderer(activeProject string) *TableRenderer {
	tr := NewTableRenderer([]string{
		FormatResourceType("project") + " NAME",
		"PATH",
		IconTime + " CREATED",
	})

	// Column styles: Name (primary), Path (secondary), Date (muted)
	tr.SetColumnStyles([]lipgloss.Style{
		lipgloss.NewStyle().Foreground(PrimaryColor).Padding(0, 1),
		PathStyle.Copy().Padding(0, 1),
		DateStyle.Copy().Padding(0, 1),
	})

	return tr
}

// WorkspacesTableRenderer creates a styled table for workspaces
func WorkspacesTableRenderer(activeWorkspace string) *TableRenderer {
	tr := NewTableRenderer([]string{
		FormatResourceType("workspace") + " NAME",
		IconProject + " PROJECT",
		IconImage + " IMAGE",
		"STATUS",
		IconTime + " CREATED",
	})

	// Column styles
	tr.SetColumnStyles([]lipgloss.Style{
		lipgloss.NewStyle().Foreground(PrimaryColor).Padding(0, 1),   // Name
		lipgloss.NewStyle().Foreground(SecondaryColor).Padding(0, 1), // Project
		lipgloss.NewStyle().Foreground(InfoColor).Padding(0, 1),      // Image
		lipgloss.NewStyle().Foreground(SuccessColor).Padding(0, 1),   // Status
		DateStyle.Copy().Padding(0, 1),                               // Created
	})

	return tr
}

// PluginsTableRenderer creates a styled table for plugins
func PluginsTableRenderer(activePlugins []string) *TableRenderer {
	tr := NewTableRenderer([]string{
		FormatResourceType("plugin") + " NAME",
		IconCategory + " CATEGORY",
		IconGit + " REPO",
		IconTag + " VERSION",
	})

	// Column styles
	tr.SetColumnStyles([]lipgloss.Style{
		lipgloss.NewStyle().Foreground(PrimaryColor).Padding(0, 1), // Name
		CategoryStyle.Copy().Padding(0, 1),                         // Category
		PathStyle.Copy().Padding(0, 1),                             // Repo
		VersionStyle.Copy().Padding(0, 1),                          // Version
	})

	return tr
}

// Helper function to format active items
func FormatActiveItem(name string, isActive bool) string {
	if isActive {
		return RenderActive(name)
	}
	return name
}

// ProgressMessage renders a progress message with a bullet
func ProgressMessage(message string) string {
	return InfoStyle.Render(Arrow + message)
}

// StepMessage renders a step message with a check mark
func StepMessage(message string, success bool) string {
	if success {
		return SuccessStyle.Render(CheckMark + message)
	}
	return ErrorStyle.Render(CrossMark + message)
}

// Section header
func SectionHeader(title string) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(PrimaryColor).
		BorderBottom(true).
		Padding(0, 1)
	return style.Render(title)
}

// Info box
func InfoBox(message string) string {
	style := lipgloss.NewStyle().
		Foreground(InfoColor).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(InfoColor).
		Padding(0, 1).
		Margin(1, 0)
	return style.Render(message)
}

// Success box
func SuccessBox(message string) string {
	style := lipgloss.NewStyle().
		Foreground(SuccessColor).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(SuccessColor).
		Padding(0, 1).
		Margin(1, 0)
	return style.Render(CheckMark + message)
}

// Error box
func ErrorBox(message string) string {
	style := lipgloss.NewStyle().
		Foreground(ErrorColor).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ErrorColor).
		Padding(0, 1).
		Margin(1, 0)
	return style.Render(CrossMark + message)
}

// Simple list item
func ListItem(text string, level int) string {
	indent := strings.Repeat("  ", level)
	return fmt.Sprintf("%s%s %s", indent, Bullet, TextStyle.Render(text))
}
