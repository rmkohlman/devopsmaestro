package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
)

// makeTestPromptStore creates a PromptFileStore backed by a temp directory.
func makeTestPromptStore(t *testing.T) *PromptFileStore {
	t.Helper()
	dir := t.TempDir()
	promptsDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("mkdir prompts: %v", err)
	}
	return &PromptFileStore{
		dir:        promptsDir,
		activePath: filepath.Join(dir, ".active-prompt"),
	}
}

// saveTestPrompt writes a minimal prompt YAML to the store.
func saveTestPrompt(t *testing.T, s *PromptFileStore, name string) {
	t.Helper()
	p := &prompt.Prompt{Name: name, Type: prompt.PromptTypeStarship}
	if err := s.Save(p); err != nil {
		t.Fatalf("save prompt %s: %v", name, err)
	}
}

// ---------------------------------------------------------------------------
// PromptFileStore.SetActive / GetActive / ClearActive
// ---------------------------------------------------------------------------

func TestPromptFileStore_SetActive_GetActive_RoundTrip(t *testing.T) {
	s := makeTestPromptStore(t)
	saveTestPrompt(t, s, "coolnight")

	if err := s.SetActive("coolnight"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	got, err := s.GetActive()
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if got != "coolnight" {
		t.Errorf("GetActive = %q, want %q", got, "coolnight")
	}
}

func TestPromptFileStore_GetActive_ReturnsEmpty_WhenNoMarker(t *testing.T) {
	s := makeTestPromptStore(t)

	got, err := s.GetActive()
	if err != nil {
		t.Fatalf("GetActive with no marker: %v", err)
	}
	if got != "" {
		t.Errorf("GetActive = %q, want empty string", got)
	}
}

func TestPromptFileStore_SetActive_Fails_WhenPromptNotInStore(t *testing.T) {
	s := makeTestPromptStore(t)

	err := s.SetActive("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent prompt, got nil")
	}
}

func TestPromptFileStore_ClearActive_RemovesMarker(t *testing.T) {
	s := makeTestPromptStore(t)
	saveTestPrompt(t, s, "minimal")
	_ = s.SetActive("minimal")

	if err := s.ClearActive(); err != nil {
		t.Fatalf("ClearActive: %v", err)
	}

	got, err := s.GetActive()
	if err != nil {
		t.Fatalf("GetActive after clear: %v", err)
	}
	if got != "" {
		t.Errorf("GetActive after clear = %q, want empty", got)
	}
}

func TestPromptFileStore_ClearActive_IsIdempotent_WhenNoMarker(t *testing.T) {
	s := makeTestPromptStore(t)

	if err := s.ClearActive(); err != nil {
		t.Fatalf("ClearActive with no marker should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Delete cascade
// ---------------------------------------------------------------------------

func TestPromptFileStore_Delete_ClearsActiveMarker(t *testing.T) {
	s := makeTestPromptStore(t)
	saveTestPrompt(t, s, "cascade")
	_ = s.SetActive("cascade")

	if err := s.Delete("cascade"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := s.GetActive()
	if err != nil {
		t.Fatalf("GetActive after delete: %v", err)
	}
	if got != "" {
		t.Errorf("active marker should be cleared after deleting active prompt, got %q", got)
	}
}

func TestPromptFileStore_Delete_DoesNotClearMarker_WhenDifferentPromptActive(t *testing.T) {
	s := makeTestPromptStore(t)
	saveTestPrompt(t, s, "alpha")
	saveTestPrompt(t, s, "beta")
	_ = s.SetActive("beta")

	if err := s.Delete("alpha"); err != nil {
		t.Fatalf("Delete alpha: %v", err)
	}

	got, err := s.GetActive()
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if got != "beta" {
		t.Errorf("active marker should still be %q after deleting a different prompt, got %q", "beta", got)
	}
}

// ---------------------------------------------------------------------------
// SetActive overwrites previous value
// ---------------------------------------------------------------------------

func TestPromptFileStore_SetActive_Overwrites(t *testing.T) {
	s := makeTestPromptStore(t)
	saveTestPrompt(t, s, "first")
	saveTestPrompt(t, s, "second")

	_ = s.SetActive("first")
	_ = s.SetActive("second")

	got, err := s.GetActive()
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if got != "second" {
		t.Errorf("GetActive = %q, want %q", got, "second")
	}
}
