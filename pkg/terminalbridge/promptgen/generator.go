// Package promptgen extracts prompt configuration generation logic from the CLI
// layer into a reusable package. It handles the pipeline: prompt → YAML → renderer
// → generated config (e.g., starship.toml content).
//
// This keeps the CLI layer thin — it only orchestrates, while this package
// owns the generation logic.
package promptgen

import (
	"context"
	"fmt"

	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
)

// Generator produces shell prompt configuration files from prompt definitions.
// It encapsulates the full generation pipeline:
//   - Accept a Prompt (domain model)
//   - Resolve color palette from context
//   - Render to the target format (e.g., starship.toml)
type Generator struct {
	renderer prompt.PromptRenderer
}

// New creates a Generator with the default Starship renderer.
func New() *Generator {
	return &Generator{
		renderer: prompt.NewStarshipRenderer(),
	}
}

// NewWithRenderer creates a Generator with a custom PromptRenderer.
func NewWithRenderer(r prompt.PromptRenderer) *Generator {
	return &Generator{
		renderer: r,
	}
}

// Generate renders a prompt's configuration using the color palette from ctx.
// Returns the generated config string (e.g., starship.toml content).
func (g *Generator) Generate(ctx context.Context, p *prompt.Prompt) (string, error) {
	if p == nil {
		return "", fmt.Errorf("prompt cannot be nil")
	}

	promptYAML := p.ToYAML()

	// Resolve palette from context (falls back to defaults)
	provider := colors.FromContextOrDefault(ctx)
	pal := colors.ToPalette(provider)

	config, err := g.renderer.Render(promptYAML, pal)
	if err != nil {
		return "", fmt.Errorf("failed to render prompt %q: %w", p.Name, err)
	}

	return config, nil
}

// GenerateWithProvider renders using an explicit ColorProvider instead of context.
func (g *Generator) GenerateWithProvider(p *prompt.Prompt, provider colors.ColorProvider) (string, error) {
	if p == nil {
		return "", fmt.Errorf("prompt cannot be nil")
	}

	promptYAML := p.ToYAML()
	pal := colors.ToPalette(provider)

	config, err := g.renderer.Render(promptYAML, pal)
	if err != nil {
		return "", fmt.Errorf("failed to render prompt %q: %w", p.Name, err)
	}

	return config, nil
}

// Validate checks whether a prompt can be successfully rendered.
// Returns nil if the prompt would generate valid config.
func (g *Generator) Validate(ctx context.Context, p *prompt.Prompt) error {
	_, err := g.Generate(ctx, p)
	return err
}
