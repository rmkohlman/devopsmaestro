package operators

import (
	"os"
	"path/filepath"

	"github.com/rmkohlman/MaestroSDK/paths"
)

// GitCredentialMount represents a host path to mount for git credential access.
type GitCredentialMount struct {
	Source      string
	Destination string
	ReadOnly    bool
}

// GetGitCredentialMounts returns bind-mount configs for git credentials that
// exist on the host. Mounts are always read-only to protect host files.
//
// Supported mounts:
//   - ~/.ssh       → /home/dev/.ssh:ro       (SSH keys & config)
//   - ~/.gitconfig → /home/dev/.gitconfig:ro  (git configuration)
//
// Paths that don't exist on the host are silently skipped.
func GetGitCredentialMounts() []GitCredentialMount {
	pc, err := paths.Default()
	if err != nil {
		return nil
	}
	return gitCredentialMountsFromHome(filepath.Dir(pc.Root()))
}

// gitCredentialMountsFromHome is the testable core of GetGitCredentialMounts.
func gitCredentialMountsFromHome(home string) []GitCredentialMount {
	candidates := []struct {
		rel  string
		dest string
	}{
		{".ssh", "/home/dev/.ssh"},
		{".gitconfig", "/home/dev/.gitconfig"},
	}

	var mounts []GitCredentialMount
	for _, c := range candidates {
		src := filepath.Join(home, c.rel)
		if _, err := os.Stat(src); err == nil {
			mounts = append(mounts, GitCredentialMount{
				Source:      src,
				Destination: c.dest,
				ReadOnly:    true,
			})
		}
	}
	return mounts
}
