// Package handlers provides resource handlers for different resource types.
package handlers

import (
	"fmt"

	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/terminalops/prompt"

	"gopkg.in/yaml.v3"
)

// TerminalPromptHandler handles TerminalPrompt resources.
type TerminalPromptHandler struct{}

// NewTerminalPromptHandler creates a new TerminalPrompt handler.
func NewTerminalPromptHandler() *TerminalPromptHandler {
	return &TerminalPromptHandler{}
}

func (h *TerminalPromptHandler) Kind() string {
	return prompt.KindTerminalPrompt
}

// Apply creates or updates a prompt from YAML data.
func (h *TerminalPromptHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	p, err := prompt.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt YAML: %w", err)
	}

	// Get the appropriate store
	promptStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	// Upsert the prompt
	if err := promptStore.Upsert(p); err != nil {
		return nil, fmt.Errorf("failed to save prompt: %w", err)
	}

	return &TerminalPromptResource{prompt: p}, nil
}

// Get retrieves a prompt by name.
func (h *TerminalPromptHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	promptStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	p, err := promptStore.Get(name)
	if err != nil {
		return nil, err
	}

	return &TerminalPromptResource{prompt: p}, nil
}

// List returns all prompts.
func (h *TerminalPromptHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	promptStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	prompts, err := promptStore.List()
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(prompts))
	for i, p := range prompts {
		result[i] = &TerminalPromptResource{prompt: p}
	}
	return result, nil
}

// Delete removes a prompt by name.
func (h *TerminalPromptHandler) Delete(ctx resource.Context, name string) error {
	promptStore, err := h.getStore(ctx)
	if err != nil {
		return err
	}

	return promptStore.Delete(name)
}

// ToYAML serializes a prompt to YAML.
func (h *TerminalPromptHandler) ToYAML(res resource.Resource) ([]byte, error) {
	pr, ok := res.(*TerminalPromptResource)
	if !ok {
		return nil, fmt.Errorf("expected TerminalPromptResource, got %T", res)
	}

	yamlDoc := pr.prompt.ToYAML()
	return yaml.Marshal(yamlDoc)
}

// getStore returns the appropriate PromptStore based on context.
func (h *TerminalPromptHandler) getStore(ctx resource.Context) (prompt.PromptStore, error) {
	// If DataStore is provided, use SQLitePromptStore adapter
	if ctx.DataStore != nil {
		if ds, ok := ctx.DataStore.(prompt.PromptDataStore); ok {
			return prompt.NewSQLitePromptStore(ds), nil
		}
		return nil, fmt.Errorf("DataStore does not implement PromptDataStore: %T", ctx.DataStore)
	}

	// No store configured
	return nil, fmt.Errorf("no prompt store configured: DataStore is required")
}

// TerminalPromptResource wraps a prompt.Prompt to implement resource.Resource.
type TerminalPromptResource struct {
	prompt *prompt.Prompt
}

func (r *TerminalPromptResource) GetKind() string {
	return prompt.KindTerminalPrompt
}

func (r *TerminalPromptResource) GetName() string {
	return r.prompt.Name
}

func (r *TerminalPromptResource) Validate() error {
	if r.prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}
	if r.prompt.Type == "" {
		return fmt.Errorf("prompt type is required")
	}
	return nil
}

// Prompt returns the underlying prompt.Prompt.
func (r *TerminalPromptResource) Prompt() *prompt.Prompt {
	return r.prompt
}
