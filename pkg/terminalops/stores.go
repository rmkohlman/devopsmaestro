package terminalops

import (
	"os"
	"path/filepath"
	"strings"

	"devopsmaestro/pkg/terminalops/plugin"
	"devopsmaestro/pkg/terminalops/profile"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/shell"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// FILE-BASED PROMPT STORE
// =============================================================================

// FilePromptStore implements PromptStore using file-based storage.
// Each prompt is stored as a YAML file in the specified directory.
type FilePromptStore struct {
	dir string
}

// NewFilePromptStore creates a new FilePromptStore.
func NewFilePromptStore(dir string) *FilePromptStore {
	return &FilePromptStore{dir: dir}
}

// Save stores a prompt to a YAML file.
func (s *FilePromptStore) Save(p *prompt.Prompt) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

// Get retrieves a prompt by name.
func (s *FilePromptStore) Get(name string) (*prompt.Prompt, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return prompt.Parse(data)
}

// List returns all stored prompts.
func (s *FilePromptStore) List() ([]*prompt.Prompt, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var prompts []*prompt.Prompt
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		p, err := s.Get(name)
		if err != nil {
			continue
		}
		prompts = append(prompts, p)
	}
	return prompts, nil
}

// Delete removes a prompt by name.
func (s *FilePromptStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

// Exists checks if a prompt exists.
func (s *FilePromptStore) Exists(name string) bool {
	_, err := os.Stat(filepath.Join(s.dir, name+".yaml"))
	return err == nil
}

// Close releases resources (no-op for file store).
func (s *FilePromptStore) Close() error { return nil }

// =============================================================================
// FILE-BASED PLUGIN STORE
// =============================================================================

// FilePluginStore implements PluginStore using file-based storage.
// Each plugin is stored as a YAML file in the specified directory.
type FilePluginStore struct {
	dir string
}

// NewFilePluginStore creates a new FilePluginStore.
func NewFilePluginStore(dir string) *FilePluginStore {
	return &FilePluginStore{dir: dir}
}

// Save stores a plugin to a YAML file.
func (s *FilePluginStore) Save(p *plugin.Plugin) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

// Get retrieves a plugin by name.
func (s *FilePluginStore) Get(name string) (*plugin.Plugin, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return plugin.Parse(data)
}

// List returns all stored plugins.
func (s *FilePluginStore) List() ([]*plugin.Plugin, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []*plugin.Plugin
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		p, err := s.Get(name)
		if err != nil {
			continue
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

// Delete removes a plugin by name.
func (s *FilePluginStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

// Close releases resources (no-op for file store).
func (s *FilePluginStore) Close() error { return nil }

// =============================================================================
// FILE-BASED SHELL STORE
// =============================================================================

// FileShellStore implements ShellStore using file-based storage.
// Each shell config is stored as a YAML file in the specified directory.
type FileShellStore struct {
	dir string
}

// NewFileShellStore creates a new FileShellStore.
func NewFileShellStore(dir string) *FileShellStore {
	return &FileShellStore{dir: dir}
}

// Save stores a shell config to a YAML file.
func (s *FileShellStore) Save(sh *shell.Shell) error {
	data, err := yaml.Marshal(sh.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, sh.Name+".yaml"), data, 0644)
}

// Get retrieves a shell config by name.
func (s *FileShellStore) Get(name string) (*shell.Shell, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return shell.Parse(data)
}

// List returns all stored shell configs.
func (s *FileShellStore) List() ([]*shell.Shell, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var shells []*shell.Shell
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		sh, err := s.Get(name)
		if err != nil {
			continue
		}
		shells = append(shells, sh)
	}
	return shells, nil
}

// Delete removes a shell config by name.
func (s *FileShellStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

// Close releases resources (no-op for file store).
func (s *FileShellStore) Close() error { return nil }

// =============================================================================
// FILE-BASED PROFILE STORE
// =============================================================================

// FileProfileStore implements ProfileStore using file-based storage.
// Each profile is stored as a YAML file in the specified directory.
// The active profile is tracked in a separate file.
type FileProfileStore struct {
	dir        string
	activePath string
}

// NewFileProfileStore creates a new FileProfileStore.
func NewFileProfileStore(dir string) *FileProfileStore {
	return &FileProfileStore{
		dir:        dir,
		activePath: filepath.Join(filepath.Dir(dir), ".active-profile"),
	}
}

// Save stores a profile to a YAML file.
func (s *FileProfileStore) Save(p *profile.Profile) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

// Get retrieves a profile by name.
func (s *FileProfileStore) Get(name string) (*profile.Profile, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return profile.Parse(data)
}

// List returns all stored profiles.
func (s *FileProfileStore) List() ([]*profile.Profile, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var profiles []*profile.Profile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		p, err := s.Get(name)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// Delete removes a profile by name.
func (s *FileProfileStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

// SetActive sets the active profile.
func (s *FileProfileStore) SetActive(name string) error {
	// Verify profile exists
	if _, err := s.Get(name); err != nil {
		return err
	}
	return os.WriteFile(s.activePath, []byte(name), 0644)
}

// GetActive returns the active profile, or nil if none is set.
func (s *FileProfileStore) GetActive() (*profile.Profile, error) {
	data, err := os.ReadFile(s.activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return nil, nil
	}
	return s.Get(name)
}

// Close releases resources (no-op for file store).
func (s *FileProfileStore) Close() error { return nil }

// =============================================================================
// INTERFACE COMPLIANCE
// =============================================================================

// Compile-time interface compliance checks
var (
	_ PromptStore  = (*FilePromptStore)(nil)
	_ PluginStore  = (*FilePluginStore)(nil)
	_ ShellStore   = (*FileShellStore)(nil)
	_ ProfileStore = (*FileProfileStore)(nil)
)
