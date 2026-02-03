package terminalops

import (
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/pkg/terminalops/plugin"
	"devopsmaestro/pkg/terminalops/profile"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/shell"
)

func TestNewManager(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewWithOptions(Options{
		ConfigDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Verify directories were created
	dirs := []string{"prompts", "plugins", "shells", "profiles"}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to be created", dir)
		}
	}

	// Verify config dir is set correctly
	if mgr.ConfigDir() != tmpDir {
		t.Errorf("ConfigDir() = %s, want %s", mgr.ConfigDir(), tmpDir)
	}
}

func TestFilePromptStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFilePromptStore(tmpDir)

	// Create a test prompt
	p := &prompt.Prompt{
		Name:        "test-prompt",
		Type:        prompt.PromptTypeStarship,
		Category:    "test",
		Description: "A test prompt",
		Format:      "$directory$git_branch",
	}

	// Test Save
	if err := store.Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Exists
	if !store.Exists("test-prompt") {
		t.Error("Exists() returned false for saved prompt")
	}

	// Test Get
	got, err := store.Get("test-prompt")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("Get().Name = %s, want %s", got.Name, p.Name)
	}

	// Test List
	prompts, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("List() returned %d prompts, want 1", len(prompts))
	}

	// Test Delete
	if err := store.Delete("test-prompt"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if store.Exists("test-prompt") {
		t.Error("Exists() returned true after delete")
	}

	// Test Close
	if err := store.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestFilePluginStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFilePluginStore(tmpDir)

	// Create a test plugin
	p := &plugin.Plugin{
		Name:        "test-plugin",
		Source:      "zsh-users/zsh-autosuggestions",
		Category:    "completion",
		Description: "A test plugin",
	}

	// Test Save
	if err := store.Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	got, err := store.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("Get().Name = %s, want %s", got.Name, p.Name)
	}

	// Test List
	plugins, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("List() returned %d plugins, want 1", len(plugins))
	}

	// Test Delete
	if err := store.Delete("test-plugin"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	plugins, _ = store.List()
	if len(plugins) != 0 {
		t.Error("List() returned plugins after delete")
	}
}

func TestFileShellStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFileShellStore(tmpDir)

	// Create a test shell config
	s := &shell.Shell{
		Name:        "test-shell",
		Description: "A test shell config",
		ShellType:   shell.ShellTypeZsh,
		Aliases: []shell.Alias{
			{Name: "ll", Command: "ls -la"},
		},
	}

	// Test Save
	if err := store.Save(s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	got, err := store.Get("test-shell")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != s.Name {
		t.Errorf("Get().Name = %s, want %s", got.Name, s.Name)
	}

	// Test List
	shells, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(shells) != 1 {
		t.Errorf("List() returned %d shells, want 1", len(shells))
	}

	// Test Delete
	if err := store.Delete("test-shell"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	shells, _ = store.List()
	if len(shells) != 0 {
		t.Error("List() returned shells after delete")
	}
}

func TestFileProfileStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewFileProfileStore(tmpDir)

	// Create a test profile
	p := &profile.Profile{
		Name:        "test-profile",
		Description: "A test profile",
		Prompt: &prompt.Prompt{
			Name: "test-prompt",
			Type: prompt.PromptTypeStarship,
		},
	}

	// Test Save
	if err := store.Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	got, err := store.Get("test-profile")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("Get().Name = %s, want %s", got.Name, p.Name)
	}

	// Test List
	profiles, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("List() returned %d profiles, want 1", len(profiles))
	}

	// Test SetActive and GetActive
	if err := store.SetActive("test-profile"); err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	active, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}
	if active == nil {
		t.Fatal("GetActive returned nil")
	}
	if active.Name != "test-profile" {
		t.Errorf("GetActive().Name = %s, want test-profile", active.Name)
	}

	// Test Delete
	if err := store.Delete("test-profile"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	profiles, _ = store.List()
	if len(profiles) != 0 {
		t.Error("List() returned profiles after delete")
	}
}

func TestManagerOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewWithOptions(Options{
		ConfigDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Test Prompt Operations
	t.Run("PromptOperations", func(t *testing.T) {
		p := &prompt.Prompt{
			Name:     "mgr-test-prompt",
			Type:     prompt.PromptTypeStarship,
			Category: "test",
		}

		if err := mgr.SavePrompt(p); err != nil {
			t.Fatalf("SavePrompt failed: %v", err)
		}

		got, err := mgr.GetPrompt("mgr-test-prompt")
		if err != nil {
			t.Fatalf("GetPrompt failed: %v", err)
		}
		if got.Name != p.Name {
			t.Errorf("GetPrompt().Name = %s, want %s", got.Name, p.Name)
		}

		prompts, err := mgr.ListPrompts()
		if err != nil {
			t.Fatalf("ListPrompts failed: %v", err)
		}
		if len(prompts) != 1 {
			t.Errorf("ListPrompts() returned %d, want 1", len(prompts))
		}

		if err := mgr.DeletePrompt("mgr-test-prompt"); err != nil {
			t.Fatalf("DeletePrompt failed: %v", err)
		}
	})

	// Test Plugin Operations
	t.Run("PluginOperations", func(t *testing.T) {
		p := &plugin.Plugin{
			Name:     "mgr-test-plugin",
			Source:   "test/source",
			Category: "test",
		}

		if err := mgr.SavePlugin(p); err != nil {
			t.Fatalf("SavePlugin failed: %v", err)
		}

		got, err := mgr.GetPlugin("mgr-test-plugin")
		if err != nil {
			t.Fatalf("GetPlugin failed: %v", err)
		}
		if got.Name != p.Name {
			t.Errorf("GetPlugin().Name = %s, want %s", got.Name, p.Name)
		}

		plugins, err := mgr.ListPlugins()
		if err != nil {
			t.Fatalf("ListPlugins failed: %v", err)
		}
		if len(plugins) != 1 {
			t.Errorf("ListPlugins() returned %d, want 1", len(plugins))
		}

		if err := mgr.DeletePlugin("mgr-test-plugin"); err != nil {
			t.Fatalf("DeletePlugin failed: %v", err)
		}
	})

	// Test Shell Operations
	t.Run("ShellOperations", func(t *testing.T) {
		s := &shell.Shell{
			Name:        "mgr-test-shell",
			Description: "test shell",
			ShellType:   shell.ShellTypeZsh,
		}

		if err := mgr.SaveShell(s); err != nil {
			t.Fatalf("SaveShell failed: %v", err)
		}

		got, err := mgr.GetShell("mgr-test-shell")
		if err != nil {
			t.Fatalf("GetShell failed: %v", err)
		}
		if got.Name != s.Name {
			t.Errorf("GetShell().Name = %s, want %s", got.Name, s.Name)
		}

		shells, err := mgr.ListShells()
		if err != nil {
			t.Fatalf("ListShells failed: %v", err)
		}
		if len(shells) != 1 {
			t.Errorf("ListShells() returned %d, want 1", len(shells))
		}

		if err := mgr.DeleteShell("mgr-test-shell"); err != nil {
			t.Fatalf("DeleteShell failed: %v", err)
		}
	})

	// Test Profile Operations
	t.Run("ProfileOperations", func(t *testing.T) {
		p := &profile.Profile{
			Name:        "mgr-test-profile",
			Description: "test profile",
		}

		if err := mgr.SaveProfile(p); err != nil {
			t.Fatalf("SaveProfile failed: %v", err)
		}

		got, err := mgr.GetProfile("mgr-test-profile")
		if err != nil {
			t.Fatalf("GetProfile failed: %v", err)
		}
		if got.Name != p.Name {
			t.Errorf("GetProfile().Name = %s, want %s", got.Name, p.Name)
		}

		profiles, err := mgr.ListProfiles()
		if err != nil {
			t.Fatalf("ListProfiles failed: %v", err)
		}
		if len(profiles) != 1 {
			t.Errorf("ListProfiles() returned %d, want 1", len(profiles))
		}

		if err := mgr.DeleteProfile("mgr-test-profile"); err != nil {
			t.Fatalf("DeleteProfile failed: %v", err)
		}
	})
}

func TestInstallPreset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewWithOptions(Options{
		ConfigDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	presets := []string{"default", "minimal", "power-user"}
	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			if err := mgr.InstallPreset(preset); err != nil {
				t.Fatalf("InstallPreset(%s) failed: %v", preset, err)
			}

			// Verify it was saved
			got, err := mgr.GetProfile(preset)
			if err != nil {
				t.Fatalf("GetProfile(%s) failed: %v", preset, err)
			}
			if got.Name != preset {
				t.Errorf("GetProfile().Name = %s, want %s", got.Name, preset)
			}

			// Clean up
			mgr.DeleteProfile(preset)
		})
	}

	// Test unknown preset
	if err := mgr.InstallPreset("unknown"); err == nil {
		t.Error("InstallPreset(unknown) should return error")
	}
}

func TestGeneratePromptConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewWithOptions(Options{
		ConfigDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	p := &prompt.Prompt{
		Name:   "test",
		Type:   prompt.PromptTypeStarship,
		Format: "$directory",
	}

	output, err := mgr.GeneratePromptConfig(p)
	if err != nil {
		t.Fatalf("GeneratePromptConfig failed: %v", err)
	}

	if output == "" {
		t.Error("GeneratePromptConfig returned empty string")
	}
}

func TestAccessors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr, err := NewWithOptions(Options{
		ConfigDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Test accessors return non-nil
	if mgr.PromptStore() == nil {
		t.Error("PromptStore() returned nil")
	}
	if mgr.PluginStore() == nil {
		t.Error("PluginStore() returned nil")
	}
	if mgr.ShellStore() == nil {
		t.Error("ShellStore() returned nil")
	}
	if mgr.ProfileStore() == nil {
		t.Error("ProfileStore() returned nil")
	}
}

// =============================================================================
// MOCK STORES FOR TESTING
// =============================================================================

// MockPromptStore provides an in-memory implementation for testing.
type MockPromptStore struct {
	prompts map[string]*prompt.Prompt
}

func NewMockPromptStore() *MockPromptStore {
	return &MockPromptStore{prompts: make(map[string]*prompt.Prompt)}
}

func (m *MockPromptStore) Save(p *prompt.Prompt) error {
	m.prompts[p.Name] = p
	return nil
}

func (m *MockPromptStore) Get(name string) (*prompt.Prompt, error) {
	if p, ok := m.prompts[name]; ok {
		return p, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockPromptStore) List() ([]*prompt.Prompt, error) {
	var result []*prompt.Prompt
	for _, p := range m.prompts {
		result = append(result, p)
	}
	return result, nil
}

func (m *MockPromptStore) Delete(name string) error {
	delete(m.prompts, name)
	return nil
}

func (m *MockPromptStore) Exists(name string) bool {
	_, ok := m.prompts[name]
	return ok
}

func (m *MockPromptStore) Close() error { return nil }

// TestWithMockStore verifies custom stores can be injected.
func TestWithMockStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terminalops-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mockStore := NewMockPromptStore()

	mgr, err := NewWithOptions(Options{
		ConfigDir:   tmpDir,
		PromptStore: mockStore,
	})
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}
	defer mgr.Close()

	// Save a prompt
	p := &prompt.Prompt{Name: "mock-test", Type: prompt.PromptTypeStarship}
	if err := mgr.SavePrompt(p); err != nil {
		t.Fatalf("SavePrompt failed: %v", err)
	}

	// Verify it's in the mock store
	if !mockStore.Exists("mock-test") {
		t.Error("Prompt not found in mock store")
	}

	// Verify the manager accessor returns our mock
	if mgr.PromptStore() != mockStore {
		t.Error("PromptStore() did not return the injected mock")
	}
}
