package mirror

// MirrorManager handles bare git repository mirrors.
type MirrorManager interface {
	// Clone creates a new bare mirror from a remote URL.
	Clone(url string, slug string) (string, error)

	// Sync updates an existing mirror from its remote.
	Sync(slug string) error

	// Delete removes a mirror from disk.
	Delete(slug string) error

	// Exists checks if a mirror exists locally.
	Exists(slug string) bool

	// GetPath returns the filesystem path for a mirror.
	GetPath(slug string) string

	// CloneToWorkspace clones from a mirror to a workspace path.
	CloneToWorkspace(mirrorSlug string, destPath string, ref string) error
}
