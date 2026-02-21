package emulator

import "fmt"

// ErrNotFound indicates an emulator was not found.
type ErrNotFound struct {
	Name string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("emulator not found: %s", e.Name)
}

// ErrAlreadyExists indicates an emulator already exists.
type ErrAlreadyExists struct {
	Name string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("emulator already exists: %s", e.Name)
}
