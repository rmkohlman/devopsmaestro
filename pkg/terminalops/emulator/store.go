package emulator

// EmulatorStore defines the interface for terminal emulator storage operations.
// This allows different backends (database, file system, memory) to be used.
type EmulatorStore interface {
	// Create adds a new emulator configuration.
	// Returns an error if an emulator with the same name already exists.
	Create(emulator *Emulator) error

	// Get retrieves an emulator by name.
	// Returns nil and an error if the emulator doesn't exist.
	Get(name string) (*Emulator, error)

	// Update modifies an existing emulator.
	// Returns an error if the emulator doesn't exist.
	Update(emulator *Emulator) error

	// Upsert creates or updates an emulator (create if not exists, update if exists).
	Upsert(emulator *Emulator) error

	// Delete removes an emulator by name.
	// Returns an error if the emulator doesn't exist.
	Delete(name string) error

	// List returns all emulators in the store.
	List() ([]*Emulator, error)

	// ListByType returns emulators of a specific type (wezterm, alacritty, etc.).
	ListByType(emulatorType string) ([]*Emulator, error)

	// ListByWorkspace returns emulators for a specific workspace.
	ListByWorkspace(workspace string) ([]*Emulator, error)

	// Exists checks if an emulator with the given name exists.
	Exists(name string) (bool, error)

	// Close releases any resources held by the store.
	Close() error
}
