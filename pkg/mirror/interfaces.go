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

// MirrorInspector provides read-only inspection of bare git mirrors.
// GitMirrorManager implements both MirrorManager and MirrorInspector.
// Commands that only need inspection can accept the narrower interface.
type MirrorInspector interface {
	// ListBranches returns branch refs from a bare mirror.
	ListBranches(slug string) ([]RefInfo, error)

	// ListTags returns tag refs from a bare mirror.
	ListTags(slug string) ([]RefInfo, error)

	// DiskUsage returns the total size in bytes of a mirror on disk.
	DiskUsage(slug string) (int64, error)

	// Verify runs git fsck on a mirror and returns nil if healthy.
	Verify(slug string) error
}

// RefInfo represents a git reference (branch or tag) with metadata.
type RefInfo struct {
	// Name is the short ref name (e.g., "main", "v1.0.0").
	Name string

	// Hash is the abbreviated commit hash.
	Hash string

	// Date is the creator date in ISO 8601 format.
	Date string
}
