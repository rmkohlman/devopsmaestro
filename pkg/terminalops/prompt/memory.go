// Package prompt provides types and utilities for terminal prompt management.
package prompt

import (
	"sync"
)

// MemoryStore implements PromptStore using in-memory storage.
// Useful for testing and standalone operation.
type MemoryStore struct {
	prompts map[string]*Prompt
	mu      sync.RWMutex
}

// NewMemoryStore creates a new in-memory prompt store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		prompts: make(map[string]*Prompt),
	}
}

// Create adds a new prompt to the store.
func (s *MemoryStore) Create(p *Prompt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.prompts[p.Name]; exists {
		return &ErrAlreadyExists{Name: p.Name}
	}
	s.prompts[p.Name] = p
	return nil
}

// Update modifies an existing prompt in the store.
func (s *MemoryStore) Update(p *Prompt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.prompts[p.Name]; !exists {
		return &ErrNotFound{Name: p.Name}
	}
	s.prompts[p.Name] = p
	return nil
}

// Upsert creates or updates a prompt.
func (s *MemoryStore) Upsert(p *Prompt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prompts[p.Name] = p
	return nil
}

// Delete removes a prompt from the store.
func (s *MemoryStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.prompts[name]; !exists {
		return &ErrNotFound{Name: name}
	}
	delete(s.prompts, name)
	return nil
}

// Get retrieves a prompt by name.
func (s *MemoryStore) Get(name string) (*Prompt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.prompts[name]
	if !exists {
		return nil, &ErrNotFound{Name: name}
	}
	return p, nil
}

// List returns all prompts in the store.
func (s *MemoryStore) List() ([]*Prompt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]*Prompt, 0, len(s.prompts))
	for _, p := range s.prompts {
		prompts = append(prompts, p)
	}
	return prompts, nil
}

// ListByType returns prompts of a specific type.
func (s *MemoryStore) ListByType(promptType PromptType) ([]*Prompt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var prompts []*Prompt
	for _, p := range s.prompts {
		if p.Type == promptType {
			prompts = append(prompts, p)
		}
	}
	return prompts, nil
}

// ListByCategory returns prompts in a specific category.
func (s *MemoryStore) ListByCategory(category string) ([]*Prompt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var prompts []*Prompt
	for _, p := range s.prompts {
		if p.Category == category {
			prompts = append(prompts, p)
		}
	}
	return prompts, nil
}

// Exists checks if a prompt with the given name exists.
func (s *MemoryStore) Exists(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.prompts[name]
	return exists, nil
}

// Close releases any resources held by the store.
func (s *MemoryStore) Close() error {
	return nil
}

// Ensure MemoryStore implements PromptStore.
var _ PromptStore = (*MemoryStore)(nil)
