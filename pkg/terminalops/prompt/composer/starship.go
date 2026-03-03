// Package composer provides starship prompt composition.
package composer

import (
	"fmt"
	"regexp"

	"devopsmaestro/pkg/terminalops/prompt/extension"
	"devopsmaestro/pkg/terminalops/prompt/style"
)

// hexColorRegex matches hex color values like #ff0000, #FFF, #12345678
var hexColorRegex = regexp.MustCompile(`#[0-9A-Fa-f]{3,8}`)

// StarshipComposer composes starship prompts from styles and extensions.
type StarshipComposer struct {
	formatBuilder *FormatBuilder
}

// NewStarshipComposer creates a new StarshipComposer.
func NewStarshipComposer() *StarshipComposer {
	return &StarshipComposer{
		formatBuilder: NewFormatBuilder(),
	}
}

// Compose creates a ComposedPrompt from a style and extensions.
func (sc *StarshipComposer) Compose(promptStyle *style.PromptStyle, extensions []*extension.PromptExtension) (*ComposedPrompt, error) {
	// Step 1: Validate the style (check for hex colors)
	if err := sc.validateStyle(promptStyle); err != nil {
		return nil, fmt.Errorf("invalid style: %w", err)
	}

	// Step 2: Build a set of valid segment names from the style
	validSegments := make(map[string]bool)
	for _, seg := range promptStyle.Segments {
		validSegments[seg.Name] = true
	}

	// Step 3: Validate extensions (check for hex colors, unknown segments)
	for _, ext := range extensions {
		if err := sc.validateExtension(ext, validSegments); err != nil {
			return nil, err
		}
	}

	// Step 4: Build format string using FormatBuilder
	formatString := sc.formatBuilder.Build(promptStyle, extensions)

	// Step 5: Collect all module configs from extensions
	modules := make(map[string]ModuleConfig)
	for _, ext := range extensions {
		for moduleName, extModule := range ext.Modules {
			modules[moduleName] = ModuleConfig{
				Disabled: extModule.Disabled,
				Symbol:   extModule.Symbol,
				Format:   extModule.Format,
				Style:    extModule.Style,
				Options:  extModule.Options,
			}
		}
	}

	// Step 6: Return ComposedPrompt
	return &ComposedPrompt{
		Format:  formatString,
		Modules: modules,
		Palette: "", // Palette name will be set by the caller when rendering
	}, nil
}

// validateStyle checks if the style uses only palette keys (no hex colors).
func (sc *StarshipComposer) validateStyle(promptStyle *style.PromptStyle) error {
	for _, seg := range promptStyle.Segments {
		if seg.StartColor != "" && hexColorRegex.MatchString(seg.StartColor) {
			return fmt.Errorf("segment %q startColor %q must be palette key, not hex color", seg.Name, seg.StartColor)
		}
		if seg.EndColor != "" && hexColorRegex.MatchString(seg.EndColor) {
			return fmt.Errorf("segment %q endColor %q must be palette key, not hex color", seg.Name, seg.EndColor)
		}
	}
	return nil
}

// validateExtension checks if the extension is valid for composition.
func (sc *StarshipComposer) validateExtension(ext *extension.PromptExtension, validSegments map[string]bool) error {
	// Check if extension targets a valid segment
	if !validSegments[ext.Segment] {
		return fmt.Errorf("extension %q targets unknown segment %q", ext.Name, ext.Segment)
	}

	// Check for hex colors in module styles
	for moduleName, module := range ext.Modules {
		if module.Style != "" && hexColorRegex.MatchString(module.Style) {
			return fmt.Errorf("extension %q module %q contains hex color in style, must use palette keys only", ext.Name, moduleName)
		}
	}

	return nil
}
