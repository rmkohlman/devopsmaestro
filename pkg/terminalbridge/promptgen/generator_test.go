package promptgen

import (
	"context"
	"testing"

	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.renderer == nil {
		t.Fatal("renderer should not be nil")
	}
}

func TestNewWithRenderer(t *testing.T) {
	r := prompt.NewStarshipRenderer()
	g := NewWithRenderer(r)
	if g == nil {
		t.Fatal("NewWithRenderer() returned nil")
	}
	if g.renderer != r {
		t.Fatal("renderer should be the one we passed in")
	}
}

func TestGenerate_NilPrompt(t *testing.T) {
	g := New()
	_, err := g.Generate(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil prompt")
	}
	if err.Error() != "prompt cannot be nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerate_BasicPrompt(t *testing.T) {
	g := New()
	p := &prompt.Prompt{
		Name:    "test-basic",
		Type:    prompt.PromptTypeStarship,
		Enabled: true,
	}

	config, err := g.Generate(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == "" {
		t.Fatal("expected non-empty config")
	}
}

func TestGenerate_WithColorProvider(t *testing.T) {
	g := New()
	p := &prompt.Prompt{
		Name:    "test-colors",
		Type:    prompt.PromptTypeStarship,
		Enabled: true,
		Colors: map[string]string{
			"primary": "${theme.blue}",
		},
	}

	// Inject a color provider into context
	provider := colors.NewDefaultColorProvider()
	ctx := colors.WithProvider(context.Background(), provider)

	config, err := g.Generate(ctx, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == "" {
		t.Fatal("expected non-empty config")
	}
}

func TestGenerateWithProvider(t *testing.T) {
	g := New()
	p := &prompt.Prompt{
		Name:    "test-provider",
		Type:    prompt.PromptTypeStarship,
		Enabled: true,
	}

	provider := colors.NewDefaultColorProvider()
	config, err := g.GenerateWithProvider(p, provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == "" {
		t.Fatal("expected non-empty config")
	}
}

func TestGenerateWithProvider_NilPrompt(t *testing.T) {
	g := New()
	provider := colors.NewDefaultColorProvider()
	_, err := g.GenerateWithProvider(nil, provider)
	if err == nil {
		t.Fatal("expected error for nil prompt")
	}
}

func TestValidate_ValidPrompt(t *testing.T) {
	g := New()
	p := &prompt.Prompt{
		Name:    "test-validate",
		Type:    prompt.PromptTypeStarship,
		Enabled: true,
	}

	err := g.Validate(context.Background(), p)
	if err != nil {
		t.Fatalf("expected no error for valid prompt, got: %v", err)
	}
}

func TestValidate_NilPrompt(t *testing.T) {
	g := New()
	err := g.Validate(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil prompt")
	}
}
