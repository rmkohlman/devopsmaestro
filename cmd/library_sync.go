package cmd

import (
	"crypto/sha256"
	"devopsmaestro/db"
	"encoding/hex"
	"fmt"
	"log/slog"
)

// defaultKeyLibraryFingerprint is the defaults table key used to store
// the fingerprint of the currently-imported embedded libraries.
const defaultKeyLibraryFingerprint = "library.fingerprint"

// EnsureLibrarySynced compares the embedded library fingerprint with the
// value stored in the database. If they differ (or no fingerprint exists),
// it re-imports all library resource types and updates the stored fingerprint.
//
// This is called at build time so the DB always reflects the latest embedded
// library data, even when the user hasn't manually run `dvm library import`.
func EnsureLibrarySynced(ds db.DataStore) error {
	fingerprint, err := ComputeLibraryFingerprint()
	if err != nil {
		slog.Warn("failed to compute library fingerprint, skipping auto-sync", "error", err)
		return nil // Non-fatal: fall back to existing DB data
	}

	stored, err := ds.GetDefault(defaultKeyLibraryFingerprint)
	if err != nil {
		slog.Warn("failed to read stored library fingerprint", "error", err)
		return nil // Non-fatal
	}

	if stored == fingerprint {
		slog.Debug("library fingerprint matches, skipping auto-sync", "fingerprint", fingerprint[:12])
		return nil
	}

	slog.Info("embedded library changed, auto-syncing to database",
		"stored", truncateHash(stored),
		"current", truncateHash(fingerprint),
	)

	if err := importAllLibraries(ds); err != nil {
		return fmt.Errorf("library auto-sync failed: %w", err)
	}

	if err := ds.SetDefault(defaultKeyLibraryFingerprint, fingerprint); err != nil {
		slog.Warn("failed to store library fingerprint after sync", "error", err)
		// Import succeeded, so don't fail the build
	}

	slog.Info("library auto-sync completed", "fingerprint", fingerprint[:12])
	return nil
}

// importAllLibraries imports all embedded library resource types to the DB.
// This is the same set of imports as `dvm library import --all`.
func importAllLibraries(ds db.DataStore) error {
	importers := []struct {
		name string
		fn   func() error
	}{
		{"nvim-plugins", func() error { return importNvimPlugins(ds) }},
		{"nvim-themes", func() error { return importNvimThemes(ds) }},
		{"nvim-packages", func() error { return importNvimPackages(ds) }},
		{"terminal-prompts", func() error { return importTerminalPrompts(ds) }},
		{"terminal-plugins", func() error { return importTerminalPlugins(ds) }},
		{"terminal-packages", func() error { return importTerminalPackages(ds) }},
		{"terminal-emulators", func() error { return importTerminalEmulators(ds) }},
	}

	for _, imp := range importers {
		if err := imp.fn(); err != nil {
			return fmt.Errorf("failed to import %s: %w", imp.name, err)
		}
		slog.Debug("auto-synced library", "type", imp.name)
	}

	return nil
}

// ComputeLibraryFingerprint produces a deterministic SHA-256 hash over the
// content of every embedded library. If the embedded YAML changes (plugin
// added, field updated, package modified, etc.) the fingerprint changes,
// triggering a re-import on the next build.
func ComputeLibraryFingerprint() (string, error) {
	h := sha256.New()

	// 1. Nvim plugins
	if err := hashNvimPlugins(h); err != nil {
		return "", fmt.Errorf("nvim plugins: %w", err)
	}

	// 2. Nvim themes
	if err := hashNvimThemes(h); err != nil {
		return "", fmt.Errorf("nvim themes: %w", err)
	}

	// 3. Nvim packages
	if err := hashNvimPackages(h); err != nil {
		return "", fmt.Errorf("nvim packages: %w", err)
	}

	// 4. Terminal prompts
	if err := hashTerminalPrompts(h); err != nil {
		return "", fmt.Errorf("terminal prompts: %w", err)
	}

	// 5. Terminal plugins
	if err := hashTerminalPlugins(h); err != nil {
		return "", fmt.Errorf("terminal plugins: %w", err)
	}

	// 6. Terminal packages
	if err := hashTerminalPackages(h); err != nil {
		return "", fmt.Errorf("terminal packages: %w", err)
	}

	// 7. Terminal emulators
	if err := hashTerminalEmulators(h); err != nil {
		return "", fmt.Errorf("terminal emulators: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// truncateHash returns a short prefix of a hash string for logging.
func truncateHash(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	if s == "" {
		return "(none)"
	}
	return s
}
