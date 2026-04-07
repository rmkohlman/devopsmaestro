package main

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroSDK/render"
	theme "github.com/rmkohlman/MaestroTheme"
	themelibrary "github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

var themePreviewCmd = &cobra.Command{
	Use:   "preview <name>",
	Short: "Preview a theme's colors in the terminal",
	Long: `Preview a theme's color palette directly in the terminal.

This shows how the theme's colors will look, including:
  - Background and foreground colors
  - Syntax highlighting colors (keywords, strings, comments, etc.)
  - UI elements (errors, warnings, selections, etc.)
  - Sample code snippet with the theme applied

The preview works with both installed themes and library themes.

Examples:
  nvp theme preview tokyonight-night
  nvp theme preview catppuccin-mocha
  nvp theme preview --all              # Preview all library themes`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if all {
			// Preview all library themes
			themes, err := themelibrary.List()
			if err != nil {
				return err
			}
			for i, info := range themes {
				if i > 0 {
					fmt.Println()
					fmt.Println(strings.Repeat("─", 60))
					fmt.Println()
				}
				t, err := themelibrary.Get(info.Name)
				if err != nil {
					render.WarningfToStderr("could not load %s: %v", info.Name, err)
					continue
				}
				printThemePreview(t)
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("theme name required (or use --all)")
		}

		name := args[0]

		// Try installed themes first
		themeStore := getThemeStore()
		t, err := themeStore.Get(name)
		if err != nil {
			// Try library
			t, err = themelibrary.Get(name)
			if err != nil {
				return fmt.Errorf("theme not found: %s (not installed or in library)", name)
			}
		}

		printThemePreview(t)
		return nil
	},
}

// printThemePreview renders a colorized preview of a theme
func printThemePreview(t *theme.Theme) {
	// Get colors with fallbacks
	getColor := func(key string, fallback string) string {
		if c, ok := t.Colors[key]; ok && c != "" {
			return c
		}
		return fallback
	}

	bg := getColor("bg", "#1a1b26")
	fg := getColor("fg", "#c0caf5")
	comment := getColor("comment", "#565f89")
	red := getColor("red", "#f7768e")
	green := getColor("green", "#9ece6a")
	yellow := getColor("yellow", "#e0af68")
	blue := getColor("blue", "#7aa2f7")
	magenta := getColor("magenta", "#bb9af7")
	cyan := getColor("cyan", "#7dcfff")
	orange := getColor("orange", "#ff9e64")

	// Header
	fmt.Printf("╭─────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│ %s%-55s%s │\n", colorFgRGB(fg), centerText(t.Name, 55), "\033[0m")
	if t.Description != "" {
		desc := t.Description
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}
		fmt.Printf("│ %s%-55s%s │\n", colorFgRGB(comment), centerText(desc, 55), "\033[0m")
	}
	fmt.Printf("╰─────────────────────────────────────────────────────────╯\n\n")

	// Color swatches
	fmt.Printf("  %s████%s %s████%s %s████%s %s████%s %s████%s %s████%s %s████%s %s████%s\n",
		colorBgRGB(bg), "\033[0m",
		colorBgRGB(red), "\033[0m",
		colorBgRGB(green), "\033[0m",
		colorBgRGB(yellow), "\033[0m",
		colorBgRGB(blue), "\033[0m",
		colorBgRGB(magenta), "\033[0m",
		colorBgRGB(cyan), "\033[0m",
		colorBgRGB(fg), "\033[0m",
	)
	fmt.Printf("  %-4s %-4s %-4s %-4s %-4s %-4s %-4s %-4s\n",
		"bg", "red", "grn", "yel", "blu", "mag", "cyn", "fg")
	fmt.Println()

	// Sample code preview with syntax highlighting
	fmt.Printf("  %s// Sample code preview%s\n", colorFgRGB(comment), "\033[0m")
	fmt.Printf("  %sfunc%s %smain%s() {\n", colorFgRGB(magenta), "\033[0m", colorFgRGB(blue), "\033[0m")
	fmt.Printf("      %smessage%s %s:=%s %s\"Hello, World!\"%s\n",
		colorFgRGB(fg), "\033[0m",
		colorFgRGB(cyan), "\033[0m",
		colorFgRGB(green), "\033[0m")
	fmt.Printf("      %scount%s %s:=%s %s42%s\n",
		colorFgRGB(fg), "\033[0m",
		colorFgRGB(cyan), "\033[0m",
		colorFgRGB(orange), "\033[0m")
	fmt.Printf("      %sfmt%s.%sPrintln%s(%smessage%s, %scount%s)\n",
		colorFgRGB(cyan), "\033[0m",
		colorFgRGB(blue), "\033[0m",
		colorFgRGB(fg), "\033[0m",
		colorFgRGB(fg), "\033[0m")
	fmt.Printf("      %s// TODO: add more features%s\n", colorFgRGB(comment), "\033[0m")
	fmt.Printf("  }\n\n")

	// Diagnostics preview
	fmt.Printf("  %s✗ Error: something went wrong%s\n", colorFgRGB(red), "\033[0m")
	fmt.Printf("  %s⚠ Warning: deprecated function%s\n", colorFgRGB(yellow), "\033[0m")
	fmt.Printf("  %sℹ Info: build completed%s\n", colorFgRGB(blue), "\033[0m")
	fmt.Printf("  %s✓ Success: all tests passed%s\n", colorFgRGB(green), "\033[0m")
}

// colorFgRGB returns ANSI escape code for true color foreground
func colorFgRGB(hex string) string {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// colorBgRGB returns ANSI escape code for true color background
func colorBgRGB(hex string) string {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// parseHexColor converts hex color to RGB values
func parseHexColor(hex string) (r, g, b int, err error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}
	_, err = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return
}

// centerText centers text within a given width
func centerText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	padding := (width - len(s)) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-len(s)-padding)
}
