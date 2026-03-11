package theme

import (
	"fmt"
	"strings"
)

// generateSetup generates the theme-specific setup code.
func (g *Generator) generateSetup(sb *strings.Builder, t *Theme) {
	setupName := GetSetupName(t.Plugin.Repo)
	if setupName == "" {
		g.generateGenericSetup(sb, t)
		return
	}

	switch setupName {
	case "tokyonight":
		g.generateTokyonightSetup(sb, t)
	case "catppuccin":
		g.generateCatppuccinSetup(sb, t)
	case "gruvbox":
		g.generateGruvboxSetup(sb, t)
	case "nord":
		g.generateNordSetup(sb, t)
	case "kanagawa":
		g.generateKanagawaSetup(sb, t)
	case "rose-pine":
		g.generateRosePineSetup(sb, t)
	case "nightfox":
		g.generateNightfoxSetup(sb, t)
	case "onedark":
		g.generateOnedarkSetup(sb, t)
	case "dracula":
		g.generateDraculaSetup(sb, t)
	default:
		g.generateGenericSetup(sb, t)
	}
}

// generateTokyonightSetup generates setup for tokyonight.nvim
func (g *Generator) generateTokyonightSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"tokyonight\").setup({\n")

	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("      style = %q,\n", t.Style))
	}
	sb.WriteString(fmt.Sprintf("      transparent = %s,\n", boolToLua(t.Transparent)))

	// Handle styles option
	if styles, ok := t.Options["styles"].(map[string]any); ok {
		sb.WriteString("      styles = {\n")
		for k, v := range styles {
			if t.Transparent && (v == "dark" || v == "transparent") {
				sb.WriteString(fmt.Sprintf("        %s = palette.is_transparent() and \"transparent\" or \"dark\",\n", k))
			} else {
				sb.WriteString(fmt.Sprintf("        %s = %q,\n", k, v))
			}
		}
		sb.WriteString("      },\n")
	}

	// Custom color overrides using palette
	if t.HasCustomColors() {
		sb.WriteString("      on_colors = function(colors)\n")
		g.writeTokyonightColorOverrides(sb, t)
		sb.WriteString("      end,\n")
	}

	sb.WriteString("    })\n\n")
}

// writeTokyonightColorOverrides writes color overrides for tokyonight
func (g *Generator) writeTokyonightColorOverrides(sb *strings.Builder, t *Theme) {
	colorMap := map[string]string{
		"bg":            "colors.bg = palette.colors.bg",
		"bg_dark":       "colors.bg_dark = palette.is_transparent() and colors.none or palette.colors.bg_dark",
		"bg_float":      "colors.bg_float = palette.is_transparent() and colors.none or palette.colors.bg_dark",
		"bg_highlight":  "colors.bg_highlight = palette.colors.bg_highlight",
		"bg_popup":      "colors.bg_popup = palette.colors.bg_dark or palette.colors.bg_popup",
		"bg_search":     "colors.bg_search = palette.colors.bg_search",
		"bg_sidebar":    "colors.bg_sidebar = palette.is_transparent() and colors.none or palette.colors.bg_dark",
		"bg_statusline": "colors.bg_statusline = palette.is_transparent() and colors.none or palette.colors.bg_dark",
		"bg_visual":     "colors.bg_visual = palette.colors.bg_visual",
		"border":        "colors.border = palette.colors.border",
		"fg":            "colors.fg = palette.colors.fg",
		"fg_dark":       "colors.fg_dark = palette.colors.fg_dark",
		"fg_float":      "colors.fg_float = palette.colors.fg",
		"fg_gutter":     "colors.fg_gutter = palette.colors.fg_gutter",
		"fg_sidebar":    "colors.fg_sidebar = palette.colors.fg_dark",
	}

	for key := range t.Colors {
		if override, ok := colorMap[key]; ok {
			sb.WriteString(fmt.Sprintf("        %s\n", override))
		}
	}
}

// generateCatppuccinSetup generates setup for catppuccin/nvim
func (g *Generator) generateCatppuccinSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"catppuccin\").setup({\n")

	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("      flavour = %q,\n", t.Style))
	}
	sb.WriteString(fmt.Sprintf("      transparent_background = %s,\n", boolToLua(t.Transparent)))

	if t.HasCustomColors() {
		sb.WriteString("      color_overrides = {\n")
		sb.WriteString("        all = palette.colors,\n")
		sb.WriteString("      },\n")
	}

	sb.WriteString("    })\n\n")
}

// generateGruvboxSetup generates setup for gruvbox.nvim
func (g *Generator) generateGruvboxSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"gruvbox\").setup({\n")
	sb.WriteString(fmt.Sprintf("      transparent_mode = %s,\n", boolToLua(t.Transparent)))

	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("      contrast = %q,\n", t.Style))
	}

	if t.HasCustomColors() {
		sb.WriteString("      palette_overrides = palette.colors,\n")
	}

	sb.WriteString("    })\n\n")
}

// generateNordSetup generates setup for nord.nvim
func (g *Generator) generateNordSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString(fmt.Sprintf("    vim.g.nord_disable_background = %s\n", boolToLua(t.Transparent)))

	if t.HasOptions() {
		for key, value := range t.Options {
			sb.WriteString(fmt.Sprintf("    vim.g.nord_%s = %s\n", key, toLuaValue(value)))
		}
	}

	sb.WriteString("    require(\"nord\").set()\n\n")
}

// generateKanagawaSetup generates setup for kanagawa.nvim
func (g *Generator) generateKanagawaSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"kanagawa\").setup({\n")
	sb.WriteString(fmt.Sprintf("      transparent = %s,\n", boolToLua(t.Transparent)))

	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("      theme = %q,\n", t.Style))
	}

	if t.HasCustomColors() {
		sb.WriteString("      colors = { theme = { all = palette.colors } },\n")
	}

	sb.WriteString("    })\n\n")
}

// generateRosePineSetup generates setup for rose-pine/neovim
func (g *Generator) generateRosePineSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"rose-pine\").setup({\n")

	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("      variant = %q,\n", t.Style))
	}
	sb.WriteString(fmt.Sprintf("      disable_background = %s,\n", boolToLua(t.Transparent)))

	if t.HasCustomColors() {
		sb.WriteString("      palette = palette.colors,\n")
	}

	sb.WriteString("    })\n\n")
}

// generateNightfoxSetup generates setup for nightfox.nvim
func (g *Generator) generateNightfoxSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"nightfox\").setup({\n")
	sb.WriteString("      options = {\n")
	sb.WriteString(fmt.Sprintf("        transparent = %s,\n", boolToLua(t.Transparent)))
	sb.WriteString("      },\n")

	if t.HasCustomColors() {
		sb.WriteString("      palettes = { all = palette.colors },\n")
	}

	sb.WriteString("    })\n\n")
}

// generateOnedarkSetup generates setup for onedark.nvim
func (g *Generator) generateOnedarkSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString("    require(\"onedark\").setup({\n")

	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("      style = %q,\n", t.Style))
	}
	sb.WriteString(fmt.Sprintf("      transparent = %s,\n", boolToLua(t.Transparent)))

	if t.HasCustomColors() {
		sb.WriteString("      colors = palette.colors,\n")
	}

	sb.WriteString("    })\n")
	sb.WriteString("    require(\"onedark\").load()\n\n")
}

// generateDraculaSetup generates setup for dracula.nvim
func (g *Generator) generateDraculaSetup(sb *strings.Builder, t *Theme) {
	sb.WriteString(fmt.Sprintf("    vim.g.dracula_transparent_bg = %s\n", boolToLua(t.Transparent)))

	if t.HasCustomColors() {
		sb.WriteString("    vim.g.dracula_colors = palette.colors\n")
	}

	sb.WriteString("\n")
}

// generateGenericSetup generates a generic setup for unknown themes
func (g *Generator) generateGenericSetup(sb *strings.Builder, t *Theme) {
	parts := strings.Split(t.Plugin.Repo, "/")
	if len(parts) != 2 {
		return
	}

	moduleName := strings.TrimSuffix(parts[1], ".nvim")
	moduleName = strings.TrimSuffix(moduleName, "-nvim")
	moduleName = strings.ReplaceAll(moduleName, "-", "_")

	sb.WriteString(fmt.Sprintf("    local ok, theme = pcall(require, %q)\n", moduleName))
	sb.WriteString("    if ok and theme.setup then\n")
	sb.WriteString("      theme.setup({\n")
	sb.WriteString(fmt.Sprintf("        transparent = %s,\n", boolToLua(t.Transparent)))
	if t.Style != "" {
		sb.WriteString(fmt.Sprintf("        style = %q,\n", t.Style))
	}
	if t.HasCustomColors() {
		sb.WriteString("        colors = palette.colors,\n")
	}
	sb.WriteString("      })\n")
	sb.WriteString("    end\n\n")
}
