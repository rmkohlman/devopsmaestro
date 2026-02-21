// Package profile provides types and utilities for terminal profile management.
package profile

// ProfileStore defines the interface for profile storage operations.
// Implementations can store profiles in files, databases, or memory.
type ProfileStore interface {
	// Create adds a new profile to the store.
	// Returns an error if a profile with the same name already exists.
	Create(p *Profile) error

	// Update modifies an existing profile in the store.
	// Returns an error if the profile doesn't exist.
	Update(p *Profile) error

	// Upsert creates or updates a profile (create if not exists, update if exists).
	Upsert(p *Profile) error

	// Delete removes a profile from the store by name.
	// Returns an error if the profile doesn't exist.
	Delete(name string) error

	// Get retrieves a profile by name.
	// Returns nil and an error if the profile doesn't exist.
	Get(name string) (*Profile, error)

	// List returns all profiles in the store.
	List() ([]*Profile, error)

	// ListByCategory returns profiles in a specific category.
	ListByCategory(category string) ([]*Profile, error)

	// Exists checks if a profile with the given name exists.
	Exists(name string) (bool, error)

	// Close releases any resources held by the store.
	Close() error
}

// ErrNotFound is returned when a profile is not found.
type ErrNotFound struct {
	Name string
}

func (e *ErrNotFound) Error() string {
	return "profile not found: " + e.Name
}

// ErrAlreadyExists is returned when trying to create a profile that already exists.
type ErrAlreadyExists struct {
	Name string
}

func (e *ErrAlreadyExists) Error() string {
	return "profile already exists: " + e.Name
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
