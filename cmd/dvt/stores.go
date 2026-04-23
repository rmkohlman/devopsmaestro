package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/pkg/terminalbridge"
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/profile"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// FILE STORES
// =============================================================================

// PromptFileStore implements prompt.PromptStore using file-based storage.
// In addition to CRUD on prompt YAML files in `dir`, it tracks the currently
// active prompt name in a small marker file (`activePath`), mirroring the
// pattern used by ProfileFileStore.
type PromptFileStore struct {
	dir        string
	activePath string
}

func getPromptStore() *PromptFileStore {
	dir := filepath.Join(getConfigDir(), "prompts")
	os.MkdirAll(dir, 0755)
	return &PromptFileStore{
		dir:        dir,
		activePath: filepath.Join(getConfigDir(), ".active-prompt"),
	}
}

// SetActive marks the named prompt as the currently active one.
// The prompt must already exist in the local store.
func (s *PromptFileStore) SetActive(name string) error {
	if !s.Exists(name) {
		return fmt.Errorf("prompt not found in local store: %s", name)
	}
	return os.WriteFile(s.activePath, []byte(name), 0644)
}

// GetActive returns the currently active prompt name, or "" if none is set.
// Returns an empty string (and nil error) when no active-prompt marker exists.
func (s *PromptFileStore) GetActive() (string, error) {
	data, err := os.ReadFile(s.activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// ClearActive removes the active-prompt marker.
func (s *PromptFileStore) ClearActive() error {
	if err := os.Remove(s.activePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *PromptFileStore) Save(p *prompt.Prompt) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

func (s *PromptFileStore) Get(name string) (*prompt.Prompt, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return prompt.Parse(data)
}

func (s *PromptFileStore) List() ([]*prompt.Prompt, error) {
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

func (s *PromptFileStore) Delete(name string) error {
	if err := os.Remove(filepath.Join(s.dir, name+".yaml")); err != nil {
		return err
	}
	// If the deleted prompt was the active one, clear the active marker so we
	// don't leave a dangling reference.
	if active, _ := s.GetActive(); active == name {
		_ = s.ClearActive()
	}
	return nil
}

func (s *PromptFileStore) Exists(name string) bool {
	_, err := os.Stat(filepath.Join(s.dir, name+".yaml"))
	return err == nil
}

func (s *PromptFileStore) Close() error { return nil }

// PluginFileStore implements plugin storage
type PluginFileStore struct {
	dir string
}

// getPluginStore extracts DataStore from command context and returns database-backed plugin store
func getPluginStore(cmd *cobra.Command) (plugin.PluginStore, error) {
	// Extract DataStore from context (following established dvt pattern)
	dataStoreInterface := cmd.Context().Value("dataStore")
	if dataStoreInterface == nil {
		return nil, fmt.Errorf("database not initialized - run 'dvt init' or check configuration")
	}

	dataStore := dataStoreInterface.(*db.DataStore)

	// Return database-backed plugin store via factory
	return terminalbridge.NewDBPluginStore(*dataStore), nil
}

func (s *PluginFileStore) Save(p *plugin.Plugin) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

func (s *PluginFileStore) Get(name string) (*plugin.Plugin, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return plugin.Parse(data)
}

func (s *PluginFileStore) List() ([]*plugin.Plugin, error) {
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

func (s *PluginFileStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

// ShellFileStore implements shell config storage
type ShellFileStore struct {
	dir string
}

func getShellStore() *ShellFileStore {
	dir := filepath.Join(getConfigDir(), "shells")
	os.MkdirAll(dir, 0755)
	return &ShellFileStore{dir: dir}
}

func (s *ShellFileStore) Save(sh *shell.Shell) error {
	data, err := yaml.Marshal(sh.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, sh.Name+".yaml"), data, 0644)
}

func (s *ShellFileStore) Get(name string) (*shell.Shell, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return shell.Parse(data)
}

func (s *ShellFileStore) List() ([]*shell.Shell, error) {
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

// ProfileFileStore implements profile storage
type ProfileFileStore struct {
	dir        string
	activePath string
}

func getProfileStore() *ProfileFileStore {
	dir := filepath.Join(getConfigDir(), "profiles")
	os.MkdirAll(dir, 0755)
	return &ProfileFileStore{
		dir:        dir,
		activePath: filepath.Join(getConfigDir(), ".active-profile"),
	}
}

func (s *ProfileFileStore) Save(p *profile.Profile) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

func (s *ProfileFileStore) Get(name string) (*profile.Profile, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return profile.Parse(data)
}

func (s *ProfileFileStore) List() ([]*profile.Profile, error) {
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

func (s *ProfileFileStore) SetActive(name string) error {
	// Verify profile exists
	if _, err := s.Get(name); err != nil {
		return fmt.Errorf("profile not found: %s", name)
	}
	return os.WriteFile(s.activePath, []byte(name), 0644)
}

func (s *ProfileFileStore) GetActive() (*profile.Profile, error) {
	data, err := os.ReadFile(s.activePath)
	if err != nil {
		return nil, nil
	}
	name := strings.TrimSpace(string(data))
	return s.Get(name)
}
