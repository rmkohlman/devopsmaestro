package main

import (
	"embed"
	"io/fs"
)

//go:embed migrations
var migrationsFS embed.FS

// GetEmbeddedMigrationsFS returns the embedded migrations filesystem
// This allows dvt to access migrations at build time even when installed via Homebrew
func GetEmbeddedMigrationsFS() (fs.FS, error) {
	return fs.Sub(migrationsFS, "migrations")
}
