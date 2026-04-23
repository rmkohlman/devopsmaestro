package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
)

// overridePromptStore temporarily replaces the config dir so getPromptStore()
// returns a store backed by a temp directory.
func overridePromptStore(t *testing.T) (store *PromptFileStore, restore func()) {
	t.Helper()
	dir := t.TempDir()
	promptsDir := filepath.Join(dir, "prompts")
	_ = os.MkdirAll(promptsDir, 0755)

	orig := os.Getenv("DVT_CONFIG_DIR")
	os.Setenv("DVT_CONFIG_DIR", dir)

	store = &PromptFileStore{
		dir:        promptsDir,
		activePath: filepath.Join(dir, ".active-prompt"),
	}
	restore = func() {
		if orig == "" {
			os.Unsetenv("DVT_CONFIG_DIR")
		} else {
			os.Setenv("DVT_CONFIG_DIR", orig)
		}
	}
	return store, restore
}

// runPromptShowCmd executes promptShowCmd and captures output.
func runPromptShowCmd(t *testing.T) (stdout string, err error) {
	t.Helper()
	buf := &bytes.Buffer{}
	promptShowCmd.SetOut(buf)
	promptShowCmd.SetErr(buf)
	err = promptShowCmd.RunE(promptShowCmd, []string{})
	return buf.String(), err
}

// ---------------------------------------------------------------------------
// dvt prompt show
// ---------------------------------------------------------------------------

func TestPromptShowCmd_ReturnsError_WhenNoActivePrompt(t *testing.T) {
	store, restore := overridePromptStore(t)
	defer restore()
	_ = store // ensure temp dir is set up

	err := promptShowCmd.RunE(promptShowCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no active prompt is set, got nil")
	}
}

func TestPromptShowCmd_PrintsActiveName_WhenSet(t *testing.T) {
	store, restore := overridePromptStore(t)
	defer restore()

	// Save a prompt and mark it active in the temp store
	p := &prompt.Prompt{Name: "testprompt", Type: prompt.PromptTypeStarship}
	if err := store.Save(p); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.SetActive("testprompt"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	// Write the marker file to the path getPromptStore() will use
	configDir := os.Getenv("DVT_CONFIG_DIR")
	markerPath := filepath.Join(configDir, ".active-prompt")
	if err := os.WriteFile(markerPath, []byte("testprompt"), 0644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	// Also ensure the prompt YAML exists where getPromptStore() looks
	promptsDir := filepath.Join(configDir, "prompts")
	_ = os.MkdirAll(promptsDir, 0755)
	if err := store.Save(p); err != nil {
		t.Fatalf("save to prompts dir: %v", err)
	}

	// Run the command — it calls getPromptStore() which reads DVM_CONFIG_DIR
	err := promptShowCmd.RunE(promptShowCmd, []string{})
	if err != nil {
		t.Fatalf("promptShowCmd unexpectedly returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// dvt prompt show aliases
// ---------------------------------------------------------------------------

func TestPromptShowCmd_HasAliases(t *testing.T) {
	aliases := promptShowCmd.Aliases
	wantAliases := []string{"current", "active"}
	for _, want := range wantAliases {
		found := false
		for _, a := range aliases {
			if a == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("promptShowCmd missing alias %q; aliases = %v", want, aliases)
		}
	}
}

// ---------------------------------------------------------------------------
// dvt prompt show — error message content
// ---------------------------------------------------------------------------

func TestPromptShowCmd_ErrorMessage_ContainsHelpText(t *testing.T) {
	store, restore := overridePromptStore(t)
	defer restore()
	_ = store

	err := promptShowCmd.RunE(promptShowCmd, []string{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active prompt") {
		t.Errorf("error message %q should mention 'no active prompt'", err.Error())
	}
}
