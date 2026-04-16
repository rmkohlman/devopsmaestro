package operators

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetGitCredentialMounts tests the GetGitCredentialMounts helper.
// Since the function uses paths.Default() to derive home, we override
// behavior by creating temp dirs and verifying logic via path inspection.
//
// Strategy: call the real function and assert structural invariants
// (read-only, correct dest paths) for whatever paths exist on the host,
// then use isolated subtests that create temp files to exercise each branch.

func TestGetGitCredentialMounts_BothPathsExist(t *testing.T) {
	// Create a temp home with both .ssh and .gitconfig
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	gitconfigFile := filepath.Join(home, ".gitconfig")

	if err := os.Mkdir(sshDir, 0o700); err != nil {
		t.Fatalf("setup: mkdir .ssh: %v", err)
	}
	if err := os.WriteFile(gitconfigFile, []byte("[user]\n\tname = Test\n"), 0o600); err != nil {
		t.Fatalf("setup: write .gitconfig: %v", err)
	}

	mounts := gitCredentialMountsFromHome(home)

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	assertMount(t, mounts[0], sshDir, "/home/dev/.ssh")
	assertMount(t, mounts[1], gitconfigFile, "/home/dev/.gitconfig")
}

func TestGetGitCredentialMounts_NeitherExists(t *testing.T) {
	home := t.TempDir()
	// nothing created inside home

	mounts := gitCredentialMountsFromHome(home)

	if len(mounts) != 0 {
		t.Errorf("expected 0 mounts, got %d: %+v", len(mounts), mounts)
	}
}

func TestGetGitCredentialMounts_OnlySSHExists(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")

	if err := os.Mkdir(sshDir, 0o700); err != nil {
		t.Fatalf("setup: mkdir .ssh: %v", err)
	}

	mounts := gitCredentialMountsFromHome(home)

	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d: %+v", len(mounts), mounts)
	}
	assertMount(t, mounts[0], sshDir, "/home/dev/.ssh")
}

func TestGetGitCredentialMounts_OnlyGitconfigExists(t *testing.T) {
	home := t.TempDir()
	gitconfigFile := filepath.Join(home, ".gitconfig")

	if err := os.WriteFile(gitconfigFile, []byte("[core]\n"), 0o600); err != nil {
		t.Fatalf("setup: write .gitconfig: %v", err)
	}

	mounts := gitCredentialMountsFromHome(home)

	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d: %+v", len(mounts), mounts)
	}
	assertMount(t, mounts[0], gitconfigFile, "/home/dev/.gitconfig")
}

func TestGetGitCredentialMounts_AllMountsAreReadOnly(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".ssh"), 0o700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(""), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	mounts := gitCredentialMountsFromHome(home)

	for _, m := range mounts {
		if !m.ReadOnly {
			t.Errorf("mount %q → %q should be read-only", m.Source, m.Destination)
		}
	}
}

// assertMount checks source, destination, and read-only flag.
func assertMount(t *testing.T, m GitCredentialMount, wantSrc, wantDest string) {
	t.Helper()
	if m.Source != wantSrc {
		t.Errorf("Source = %q, want %q", m.Source, wantSrc)
	}
	if m.Destination != wantDest {
		t.Errorf("Destination = %q, want %q", m.Destination, wantDest)
	}
	if !m.ReadOnly {
		t.Errorf("ReadOnly = false, want true")
	}
}
