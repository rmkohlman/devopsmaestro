package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store defines the interface for theme storage.
type Store interface {
	// Get retrieves a theme by name.
	Get(name string) (*Theme, error)
	// List returns all stored themes.
	List() ([]*Theme, error)
	// Save stores a theme.
	Save(theme *Theme) error
	// Delete removes a theme by name.
	Delete(name string) error
	// GetActive returns the currently active theme.
	GetActive() (*Theme, error)
	// SetActive sets the active theme by name.
	SetActive(name string) error
	// Path returns the store's base path.
	Path() string
}

// FileStore implements Store using the filesystem.
type FileStore struct {
	basePath   string
	themesDir  string
	activeFile string
}

// NewFileStore creates a new file-based theme store.
func NewFileStore(basePath string) *FileStore {
	return &FileStore{
		basePath:   basePath,
		themesDir:  filepath.Join(basePath, "themes"),
		activeFile: filepath.Join(basePath, "active-theme"),
	}
}

// Init initializes the theme store directory structure.
func (s *FileStore) Init() error {
	if err := os.MkdirAll(s.themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}
	return nil
}

// Path returns the store's base path.
func (s *FileStore) Path() string {
	return s.basePath
}

// Get retrieves a theme by name.
func (s *FileStore) Get(name string) (*Theme, error) {
	filePath := filepath.Join(s.themesDir, name+".yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("theme %q not found", name)
		}
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	return ParseYAML(data)
}

// List returns all stored themes.
func (s *FileStore) List() ([]*Theme, error) {
	entries, err := os.ReadDir(s.themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Theme{}, nil
		}
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	var themes []*Theme
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".yaml")
		theme, err := s.Get(name)
		if err != nil {
			continue // Skip invalid themes
		}
		themes = append(themes, theme)
	}

	return themes, nil
}

// Save stores a theme.
func (s *FileStore) Save(theme *Theme) error {
	if err := s.Init(); err != nil {
		return err
	}

	if err := theme.Validate(); err != nil {
		return fmt.Errorf("invalid theme: %w", err)
	}

	data, err := theme.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	filePath := filepath.Join(s.themesDir, theme.Name+".yaml")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	return nil
}

// Delete removes a theme by name.
func (s *FileStore) Delete(name string) error {
	filePath := filepath.Join(s.themesDir, name+".yaml")
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("theme %q not found", name)
		}
		return fmt.Errorf("failed to delete theme: %w", err)
	}

	// If this was the active theme, clear the active file
	active, _ := s.GetActive()
	if active != nil && active.Name == name {
		os.Remove(s.activeFile)
	}

	return nil
}

// GetActive returns the currently active theme.
func (s *FileStore) GetActive() (*Theme, error) {
	data, err := os.ReadFile(s.activeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No active theme set
		}
		return nil, fmt.Errorf("failed to read active theme file: %w", err)
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return nil, nil
	}

	return s.Get(name)
}

// SetActive sets the active theme by name.
func (s *FileStore) SetActive(name string) error {
	// Verify theme exists
	if _, err := s.Get(name); err != nil {
		return err
	}

	if err := os.WriteFile(s.activeFile, []byte(name), 0644); err != nil {
		return fmt.Errorf("failed to write active theme file: %w", err)
	}

	return nil
}

// MemoryStore implements Store using in-memory storage (for testing).
type MemoryStore struct {
	themes map[string]*Theme
	active string
}

// NewMemoryStore creates a new in-memory theme store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		themes: make(map[string]*Theme),
	}
}

// Path returns an empty path for memory store.
func (s *MemoryStore) Path() string {
	return ""
}

// Get retrieves a theme by name.
func (s *MemoryStore) Get(name string) (*Theme, error) {
	theme, ok := s.themes[name]
	if !ok {
		return nil, fmt.Errorf("theme %q not found", name)
	}
	return theme, nil
}

// List returns all stored themes.
func (s *MemoryStore) List() ([]*Theme, error) {
	themes := make([]*Theme, 0, len(s.themes))
	for _, t := range s.themes {
		themes = append(themes, t)
	}
	return themes, nil
}

// Save stores a theme.
func (s *MemoryStore) Save(theme *Theme) error {
	if err := theme.Validate(); err != nil {
		return err
	}
	s.themes[theme.Name] = theme
	return nil
}

// Delete removes a theme by name.
func (s *MemoryStore) Delete(name string) error {
	if _, ok := s.themes[name]; !ok {
		return fmt.Errorf("theme %q not found", name)
	}
	delete(s.themes, name)
	if s.active == name {
		s.active = ""
	}
	return nil
}

// GetActive returns the currently active theme.
func (s *MemoryStore) GetActive() (*Theme, error) {
	if s.active == "" {
		return nil, nil
	}
	return s.Get(s.active)
}

// SetActive sets the active theme by name.
func (s *MemoryStore) SetActive(name string) error {
	if _, ok := s.themes[name]; !ok {
		return fmt.Errorf("theme %q not found", name)
	}
	s.active = name
	return nil
}
