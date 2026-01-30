package theme

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseYAML parses a YAML byte slice into a Theme.
func ParseYAML(data []byte) (*Theme, error) {
	var themeYAML ThemeYAML
	if err := yaml.Unmarshal(data, &themeYAML); err != nil {
		return nil, fmt.Errorf("failed to parse theme YAML: %w", err)
	}

	if themeYAML.Kind != "NvimTheme" {
		return nil, fmt.Errorf("invalid kind: expected NvimTheme, got %s", themeYAML.Kind)
	}

	if themeYAML.Metadata.Name == "" {
		return nil, fmt.Errorf("theme name is required")
	}

	if themeYAML.Spec.Plugin.Repo == "" {
		return nil, fmt.Errorf("theme plugin repo is required")
	}

	theme := &Theme{
		Name:        themeYAML.Metadata.Name,
		Description: themeYAML.Metadata.Description,
		Author:      themeYAML.Metadata.Author,
		Category:    themeYAML.Metadata.Category,
		Plugin:      themeYAML.Spec.Plugin,
		Style:       themeYAML.Spec.Style,
		Transparent: themeYAML.Spec.Transparent,
		Colors:      themeYAML.Spec.Colors,
		Options:     themeYAML.Spec.Options,
	}

	return theme, nil
}

// ToYAML converts a Theme to YAML bytes.
func (t *Theme) ToYAML() ([]byte, error) {
	themeYAML := ThemeYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimTheme",
		Metadata: ThemeMetadata{
			Name:        t.Name,
			Description: t.Description,
			Author:      t.Author,
			Category:    t.Category,
		},
		Spec: ThemeSpec{
			Plugin:      t.Plugin,
			Style:       t.Style,
			Transparent: t.Transparent,
			Colors:      t.Colors,
			Options:     t.Options,
		},
	}

	return yaml.Marshal(themeYAML)
}

// Validate checks if the theme configuration is valid.
func (t *Theme) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("theme name is required")
	}
	if t.Plugin.Repo == "" {
		return fmt.Errorf("theme plugin repo is required")
	}

	// Validate color format (should be hex)
	for key, color := range t.Colors {
		if !isValidHexColor(color) {
			return fmt.Errorf("invalid color format for %s: %s (expected hex like #RRGGBB)", key, color)
		}
	}

	return nil
}

// isValidHexColor checks if a string is a valid hex color.
func isValidHexColor(color string) bool {
	if !strings.HasPrefix(color, "#") {
		return false
	}
	color = strings.TrimPrefix(color, "#")
	if len(color) != 6 && len(color) != 3 {
		return false
	}
	for _, c := range color {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// HasCustomColors returns true if the theme has color overrides.
func (t *Theme) HasCustomColors() bool {
	return len(t.Colors) > 0
}

// HasOptions returns true if the theme has custom options.
func (t *Theme) HasOptions() bool {
	return len(t.Options) > 0
}

// GetColorschemeCommand returns the vim.cmd to set the colorscheme.
func (t *Theme) GetColorschemeCommand() string {
	setupName := GetSetupName(t.Plugin.Repo)
	if setupName == "" {
		// Extract name from repo (e.g., "user/theme.nvim" -> "theme")
		parts := strings.Split(t.Plugin.Repo, "/")
		if len(parts) == 2 {
			setupName = strings.TrimSuffix(parts[1], ".nvim")
			setupName = strings.TrimSuffix(setupName, "-nvim")
		}
	}

	// Apply style variant if specified
	if t.Style != "" {
		return fmt.Sprintf("%s-%s", setupName, t.Style)
	}
	return setupName
}
