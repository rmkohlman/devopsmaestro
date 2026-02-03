package prompt

import (
	"testing"
)

func TestNewPrompt(t *testing.T) {
	p := NewPrompt("test-prompt", PromptTypeStarship)

	if p.Name != "test-prompt" {
		t.Errorf("expected name 'test-prompt', got %s", p.Name)
	}
	if p.Type != PromptTypeStarship {
		t.Errorf("expected type starship, got %s", p.Type)
	}
	if !p.Enabled {
		t.Error("expected enabled to be true by default")
	}
}

func TestPromptYAML_ToPrompt(t *testing.T) {
	py := &PromptYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPrompt",
		Metadata: PromptMetadata{
			Name:        "my-prompt",
			Description: "A test prompt",
			Category:    "minimal",
			Tags:        []string{"fast", "simple"},
		},
		Spec: PromptSpec{
			Type:   PromptTypeStarship,
			Format: "$directory$character",
			Character: &CharacterConfig{
				SuccessSymbol: "[❯](green)",
				ErrorSymbol:   "[❯](red)",
			},
		},
	}

	p := py.ToPrompt()

	if p.Name != "my-prompt" {
		t.Errorf("expected name 'my-prompt', got %s", p.Name)
	}
	if p.Description != "A test prompt" {
		t.Errorf("expected description 'A test prompt', got %s", p.Description)
	}
	if p.Type != PromptTypeStarship {
		t.Errorf("expected type starship, got %s", p.Type)
	}
	if p.Format != "$directory$character" {
		t.Errorf("expected format '$directory$character', got %s", p.Format)
	}
	if p.Character == nil {
		t.Fatal("expected character config to be set")
	}
	if p.Character.SuccessSymbol != "[❯](green)" {
		t.Errorf("expected success symbol '[❯](green)', got %s", p.Character.SuccessSymbol)
	}
	if !p.Enabled {
		t.Error("expected enabled to be true by default")
	}
}

func TestPrompt_ToYAML(t *testing.T) {
	p := &Prompt{
		Name:        "my-prompt",
		Description: "A test prompt",
		Type:        PromptTypeStarship,
		Format:      "$directory$character",
		Category:    "minimal",
		Enabled:     true,
	}

	py := p.ToYAML()

	if py.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("expected apiVersion 'devopsmaestro.io/v1', got %s", py.APIVersion)
	}
	if py.Kind != "TerminalPrompt" {
		t.Errorf("expected kind 'TerminalPrompt', got %s", py.Kind)
	}
	if py.Metadata.Name != "my-prompt" {
		t.Errorf("expected name 'my-prompt', got %s", py.Metadata.Name)
	}
	if py.Spec.Type != PromptTypeStarship {
		t.Errorf("expected type starship, got %s", py.Spec.Type)
	}
	// Enabled should be nil when true (to avoid cluttering YAML)
	if py.Spec.Enabled != nil {
		t.Error("expected enabled to be nil when true")
	}
}

func TestPrompt_ToYAML_Disabled(t *testing.T) {
	p := &Prompt{
		Name:    "my-prompt",
		Type:    PromptTypeStarship,
		Enabled: false,
	}

	py := p.ToYAML()

	if py.Spec.Enabled == nil {
		t.Fatal("expected enabled to be set when false")
	}
	if *py.Spec.Enabled != false {
		t.Error("expected enabled to be false")
	}
}

func TestPrompt_TypeChecks(t *testing.T) {
	tests := []struct {
		promptType      PromptType
		isStarship      bool
		isPowerlevel10k bool
		isOhMyPosh      bool
	}{
		{PromptTypeStarship, true, false, false},
		{PromptTypePowerlevel10k, false, true, false},
		{PromptTypeOhMyPosh, false, false, true},
	}

	for _, tt := range tests {
		p := &Prompt{Type: tt.promptType}

		if p.IsStarship() != tt.isStarship {
			t.Errorf("IsStarship() for %s: expected %v, got %v", tt.promptType, tt.isStarship, p.IsStarship())
		}
		if p.IsPowerlevel10k() != tt.isPowerlevel10k {
			t.Errorf("IsPowerlevel10k() for %s: expected %v, got %v", tt.promptType, tt.isPowerlevel10k, p.IsPowerlevel10k())
		}
		if p.IsOhMyPosh() != tt.isOhMyPosh {
			t.Errorf("IsOhMyPosh() for %s: expected %v, got %v", tt.promptType, tt.isOhMyPosh, p.IsOhMyPosh())
		}
	}
}

func TestParse(t *testing.T) {
	yaml := `apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: test-prompt
  description: A test prompt
spec:
  type: starship
  format: "$directory$git_branch$character"
  character:
    success_symbol: "[❯](bold green)"
    error_symbol: "[❯](bold red)"
  modules:
    directory:
      options:
        truncation_length: 3
`

	p, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if p.Name != "test-prompt" {
		t.Errorf("expected name 'test-prompt', got %s", p.Name)
	}
	if p.Type != PromptTypeStarship {
		t.Errorf("expected type starship, got %s", p.Type)
	}
	if p.Format != "$directory$git_branch$character" {
		t.Errorf("expected format '$directory$git_branch$character', got %s", p.Format)
	}
	if p.Character == nil {
		t.Fatal("expected character config")
	}
	if p.Character.SuccessSymbol != "[❯](bold green)" {
		t.Errorf("expected success symbol '[❯](bold green)', got %s", p.Character.SuccessSymbol)
	}
}

func TestParse_InvalidKind(t *testing.T) {
	yaml := `apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test
spec:
  type: starship
`

	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid kind")
	}
}

func TestParse_MissingType(t *testing.T) {
	yaml := `apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: test
spec: {}
`

	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for missing type")
	}
}

func TestParse_InvalidType(t *testing.T) {
	yaml := `apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: test
spec:
  type: invalid-type
`

	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid type")
	}
}
