// Package prompt provides types and utilities for terminal prompt management.
package prompt

// PromptStore defines the interface for prompt storage operations.
// Implementations can store prompts in files, databases, or memory.
type PromptStore interface {
	// Create adds a new prompt to the store.
	// Returns an error if a prompt with the same name already exists.
	Create(p *Prompt) error

	// Update modifies an existing prompt in the store.
	// Returns an error if the prompt doesn't exist.
	Update(p *Prompt) error

	// Upsert creates or updates a prompt (create if not exists, update if exists).
	Upsert(p *Prompt) error

	// Delete removes a prompt from the store by name.
	// Returns an error if the prompt doesn't exist.
	Delete(name string) error

	// Get retrieves a prompt by name.
	// Returns nil and an error if the prompt doesn't exist.
	Get(name string) (*Prompt, error)

	// List returns all prompts in the store.
	List() ([]*Prompt, error)

	// ListByType returns prompts of a specific type (starship, powerlevel10k, etc.).
	ListByType(promptType PromptType) ([]*Prompt, error)

	// ListByCategory returns prompts in a specific category.
	ListByCategory(category string) ([]*Prompt, error)

	// Exists checks if a prompt with the given name exists.
	Exists(name string) (bool, error)

	// Close releases any resources held by the store.
	Close() error
}

// ErrNotFound is returned when a prompt is not found.
type ErrNotFound struct {
	Name string
}

func (e *ErrNotFound) Error() string {
	return "prompt not found: " + e.Name
}

// ErrAlreadyExists is returned when trying to create a prompt that already exists.
type ErrAlreadyExists struct {
	Name string
}

func (e *ErrAlreadyExists) Error() string {
	return "prompt already exists: " + e.Name
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	_, ok := err.(*ErrNotFound)
	return ok
}

// IsAlreadyExists returns true if the error is an already exists error.
func IsAlreadyExists(err error) bool {
	_, ok := err.(*ErrAlreadyExists)
	return ok
}
