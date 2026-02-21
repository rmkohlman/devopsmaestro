// Package library provides access to curated terminal emulator configurations.
package library

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/pkg/terminalops/emulator"
)

//go:embed emulators/*.yaml
var embeddedEmulators embed.FS

// Library provides access to curated terminal emulator configurations.
type Library interface {
	// List returns all available emulator names
	List() []string

	// Get returns an emulator by name
	Get(name string) (*emulator.Emulator, error)

	// ListByType returns emulators of a specific type
	ListByType(emulatorType string) []*emulator.Emulator

	// Categories returns all unique categories
	Categories() []string

	// Count returns the total number of emulators
	Count() int

	// Has checks if an emulator exists
	Has(name string) bool

	// All returns all emulators
	All() []*emulator.Emulator
}

// EmulatorLibrary implements Library interface.
type EmulatorLibrary struct {
	emulators map[string]*emulator.Emulator
}

// NewEmulatorLibrary creates a new EmulatorLibrary by loading all embedded emulators.
func NewEmulatorLibrary() (*EmulatorLibrary, error) {
	lib := &EmulatorLibrary{
		emulators: make(map[string]*emulator.Emulator),
	}

	// Load embedded emulators
	err := fs.WalkDir(embeddedEmulators, "emulators", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		data, err := embeddedEmulators.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		e, err := emulator.Parse(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		lib.emulators[e.Name] = e
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load library: %w", err)
	}

	return lib, nil
}

// NewLibraryFromDir creates a Library from a directory of YAML files.
func NewLibraryFromDir(dir string) (*EmulatorLibrary, error) {
	lib := &EmulatorLibrary{
		emulators: make(map[string]*emulator.Emulator),
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		e, err := emulator.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		lib.emulators[e.Name] = e
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load library from %s: %w", dir, err)
	}

	return lib, nil
}

// List returns all available emulator names.
func (lib *EmulatorLibrary) List() []string {
	names := make([]string, 0, len(lib.emulators))
	for name := range lib.emulators {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Get returns an emulator by name.
func (lib *EmulatorLibrary) Get(name string) (*emulator.Emulator, error) {
	e, ok := lib.emulators[name]
	if !ok {
		return nil, fmt.Errorf("emulator not found: %s", name)
	}
	return e, nil
}

// ListByType returns emulators of a specific type.
func (lib *EmulatorLibrary) ListByType(emulatorType string) []*emulator.Emulator {
	var emulators []*emulator.Emulator
	for _, e := range lib.emulators {
		if string(e.Type) == emulatorType {
			emulators = append(emulators, e)
		}
	}
	sort.Slice(emulators, func(i, j int) bool {
		return emulators[i].Name < emulators[j].Name
	})
	return emulators
}

// Categories returns all unique categories.
func (lib *EmulatorLibrary) Categories() []string {
	categorySet := make(map[string]struct{})
	for _, e := range lib.emulators {
		if e.Category != "" {
			categorySet[e.Category] = struct{}{}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

// Count returns the total number of emulators.
func (lib *EmulatorLibrary) Count() int {
	return len(lib.emulators)
}

// Has checks if an emulator exists.
func (lib *EmulatorLibrary) Has(name string) bool {
	_, ok := lib.emulators[name]
	return ok
}

// All returns all emulators.
func (lib *EmulatorLibrary) All() []*emulator.Emulator {
	emulators := make([]*emulator.Emulator, 0, len(lib.emulators))
	for _, e := range lib.emulators {
		emulators = append(emulators, e)
	}
	sort.Slice(emulators, func(i, j int) bool {
		return emulators[i].Name < emulators[j].Name
	})
	return emulators
}

// ListByCategory returns emulators in a specific category.
func (lib *EmulatorLibrary) ListByCategory(category string) []*emulator.Emulator {
	var emulators []*emulator.Emulator
	for _, e := range lib.emulators {
		if e.Category == category {
			emulators = append(emulators, e)
		}
	}
	sort.Slice(emulators, func(i, j int) bool {
		return emulators[i].Name < emulators[j].Name
	})
	return emulators
}

// DefaultLibrary is the singleton instance of the library.
var DefaultLibrary Library

// init loads the default library.
func init() {
	var err error
	DefaultLibrary, err = NewEmulatorLibrary()
	if err != nil {
		// This should not happen in normal execution since files are embedded
		panic(fmt.Errorf("failed to load default emulator library: %w", err))
	}
}

// Default returns the default emulator library.
func Default() Library {
	return DefaultLibrary
}
