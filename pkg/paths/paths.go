// Package paths provides a centralized, testable path configuration for all
// DevOpsMaestro tools (dvm, nvp, dvt). It replaces scattered hardcoded path
// constructions with a single PathConfig struct whose methods return
// deterministic paths derived from a home directory.
//
// Usage:
//
//	// In production — uses os.UserHomeDir()
//	pc, err := paths.Default()
//
//	// In tests — fully deterministic, no OS dependency
//	pc := paths.New("/tmp/fakehome")
//
//	dbPath := pc.Database()            // "/tmp/fakehome/.devopsmaestro/devopsmaestro.db"
//	ws     := pc.WorkspacePath("myws") // "/tmp/fakehome/.devopsmaestro/workspaces/myws"
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// Directory and file name constants used across the DevOpsMaestro ecosystem.
// Exported so callers (e.g. skip-dir logic in cmd/build_helpers.go) can
// reference them without importing a concrete path.
const (
	// DVMDirName is the hidden directory under $HOME for dvm state.
	DVMDirName = ".devopsmaestro"

	// NVPDirName is the hidden directory under $HOME for nvp state.
	NVPDirName = ".nvp"

	// DVTDirName is the hidden directory under $HOME for dvt state.
	DVTDirName = ".dvt"

	// DatabaseFile is the SQLite database filename inside the dvm root.
	DatabaseFile = "devopsmaestro.db"
)

// PathConfig holds the resolved home directory and exposes methods that return
// fully-qualified paths for every well-known location in the DevOpsMaestro
// filesystem layout. The struct is immutable — homeDir is set once at
// construction and never changes.
type PathConfig struct {
	homeDir string
}

// New creates a PathConfig rooted at the given home directory. This
// constructor has no OS dependencies, making it ideal for tests.
//
// It panics if homeDir is empty because that indicates a programming error
// in the caller — every code path must supply a valid home directory.
func New(homeDir string) *PathConfig {
	if homeDir == "" {
		panic("paths.New: homeDir must not be empty")
	}
	return &PathConfig{homeDir: homeDir}
}

// Default creates a PathConfig using the current user's home directory
// returned by os.UserHomeDir(). It is the standard constructor for
// production code.
func Default() (*PathConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("paths.Default: unable to determine home directory: %w", err)
	}
	return New(home), nil
}

// ---------------------------------------------------------------------------
// DVM root methods
// ---------------------------------------------------------------------------

// Root returns the top-level dvm data directory ({home}/.devopsmaestro).
func (p *PathConfig) Root() string {
	return filepath.Join(p.homeDir, DVMDirName)
}

// ConfigFile returns the path to the dvm configuration file
// ({root}/config.yaml).
func (p *PathConfig) ConfigFile() string {
	return filepath.Join(p.Root(), "config.yaml")
}

// Database returns the path to the SQLite database
// ({root}/devopsmaestro.db).
func (p *PathConfig) Database() string {
	return filepath.Join(p.Root(), DatabaseFile)
}

// VersionFile returns the path to the version tracking file
// ({root}/.version).
func (p *PathConfig) VersionFile() string {
	return filepath.Join(p.Root(), ".version")
}

// ContextFile returns the path to the active context file
// ({root}/context.yaml).
func (p *PathConfig) ContextFile() string {
	return filepath.Join(p.Root(), "context.yaml")
}

// NvimSyncStatus returns the path to the nvim sync-status tracking file
// ({root}/.nvim-sync-status).
func (p *PathConfig) NvimSyncStatus() string {
	return filepath.Join(p.Root(), ".nvim-sync-status")
}

// LogsDir returns the directory for log files ({root}/logs).
func (p *PathConfig) LogsDir() string {
	return filepath.Join(p.Root(), "logs")
}

// BackupsDir returns the directory for backups ({root}/backups).
func (p *PathConfig) BackupsDir() string {
	return filepath.Join(p.Root(), "backups")
}

// TemplatesDir returns the top-level templates directory
// ({root}/templates).
func (p *PathConfig) TemplatesDir() string {
	return filepath.Join(p.Root(), "templates")
}

// NvimTemplatesDir returns the nvim templates directory
// ({root}/templates/nvim).
func (p *PathConfig) NvimTemplatesDir() string {
	return filepath.Join(p.TemplatesDir(), "nvim")
}

// ShellTemplatesDir returns the shell templates directory
// ({root}/templates/shell).
func (p *PathConfig) ShellTemplatesDir() string {
	return filepath.Join(p.TemplatesDir(), "shell")
}

// ---------------------------------------------------------------------------
// Workspace methods
// ---------------------------------------------------------------------------

// WorkspacesDir returns the directory that holds all workspaces
// ({root}/workspaces).
func (p *PathConfig) WorkspacesDir() string {
	return filepath.Join(p.Root(), "workspaces")
}

// WorkspacePath returns the root directory for a single workspace identified
// by slug ({root}/workspaces/{slug}).
func (p *PathConfig) WorkspacePath(slug string) string {
	return filepath.Join(p.WorkspacesDir(), slug)
}

// WorkspaceRepoPath returns the repo sub-directory for a workspace
// ({root}/workspaces/{slug}/repo).
func (p *PathConfig) WorkspaceRepoPath(slug string) string {
	return filepath.Join(p.WorkspacePath(slug), "repo")
}

// WorkspaceVolumePath returns the volume sub-directory for a workspace
// ({root}/workspaces/{slug}/volume).
func (p *PathConfig) WorkspaceVolumePath(slug string) string {
	return filepath.Join(p.WorkspacePath(slug), "volume")
}

// WorkspaceConfigPath returns the dvm config sub-directory for a workspace
// ({root}/workspaces/{slug}/.dvm).
func (p *PathConfig) WorkspaceConfigPath(slug string) string {
	return filepath.Join(p.WorkspacePath(slug), ".dvm")
}

// ---------------------------------------------------------------------------
// Git & build methods
// ---------------------------------------------------------------------------

// ReposDir returns the directory for cloned repositories ({root}/repos).
func (p *PathConfig) ReposDir() string {
	return filepath.Join(p.Root(), "repos")
}

// BuildStagingDir returns the build-staging directory for a given app
// ({root}/build-staging/{appName}).
func (p *PathConfig) BuildStagingDir(appName string) string {
	return filepath.Join(p.Root(), "build-staging", appName)
}

// ---------------------------------------------------------------------------
// Registry methods
// ---------------------------------------------------------------------------

// RegistryDir returns the configuration directory for a named registry
// ({root}/registries/{name}).
func (p *PathConfig) RegistryDir(name string) string {
	return filepath.Join(p.Root(), "registries", name)
}

// RegistryStorage returns the generic registry storage directory
// ({root}/registry).
func (p *PathConfig) RegistryStorage() string {
	return filepath.Join(p.Root(), "registry")
}

// AthensStorage returns the Athens Go module proxy storage directory
// ({root}/athens).
func (p *PathConfig) AthensStorage() string {
	return filepath.Join(p.Root(), "athens")
}

// VerdaccioStorage returns the Verdaccio npm registry storage directory
// ({root}/verdaccio).
func (p *PathConfig) VerdaccioStorage() string {
	return filepath.Join(p.Root(), "verdaccio")
}

// DevpiStorage returns the devpi PyPI proxy storage directory
// ({root}/devpi).
func (p *PathConfig) DevpiStorage() string {
	return filepath.Join(p.Root(), "devpi")
}

// SquidDir returns the Squid HTTP cache directory ({root}/squid).
func (p *PathConfig) SquidDir() string {
	return filepath.Join(p.Root(), "squid")
}

// ---------------------------------------------------------------------------
// NVP methods
// ---------------------------------------------------------------------------

// NVPRoot returns the top-level nvp data directory ({home}/.nvp).
func (p *PathConfig) NVPRoot() string {
	return filepath.Join(p.homeDir, NVPDirName)
}

// NVPPluginsDir returns the nvp plugins directory ({nvpRoot}/plugins).
func (p *PathConfig) NVPPluginsDir() string {
	return filepath.Join(p.NVPRoot(), "plugins")
}

// NVPPackagesDir returns the nvp packages directory ({nvpRoot}/packages).
func (p *PathConfig) NVPPackagesDir() string {
	return filepath.Join(p.NVPRoot(), "packages")
}

// NVPThemesDir returns the nvp themes directory ({nvpRoot}/themes).
func (p *PathConfig) NVPThemesDir() string {
	return filepath.Join(p.NVPRoot(), "themes")
}

// NVPCoreConfig returns the path to the nvp core configuration file
// ({nvpRoot}/core.yaml).
func (p *PathConfig) NVPCoreConfig() string {
	return filepath.Join(p.NVPRoot(), "core.yaml")
}

// ---------------------------------------------------------------------------
// DVT methods
// ---------------------------------------------------------------------------

// DVTRoot returns the top-level dvt data directory ({home}/.dvt).
func (p *PathConfig) DVTRoot() string {
	return filepath.Join(p.homeDir, DVTDirName)
}

// DVTPromptsDir returns the dvt prompts directory ({dvtRoot}/prompts).
func (p *PathConfig) DVTPromptsDir() string {
	return filepath.Join(p.DVTRoot(), "prompts")
}

// DVTPluginsDir returns the dvt plugins directory ({dvtRoot}/plugins).
func (p *PathConfig) DVTPluginsDir() string {
	return filepath.Join(p.DVTRoot(), "plugins")
}

// DVTShellsDir returns the dvt shells directory ({dvtRoot}/shells).
func (p *PathConfig) DVTShellsDir() string {
	return filepath.Join(p.DVTRoot(), "shells")
}

// DVTProfilesDir returns the dvt profiles directory ({dvtRoot}/profiles).
func (p *PathConfig) DVTProfilesDir() string {
	return filepath.Join(p.DVTRoot(), "profiles")
}

// DVTActiveProfile returns the path to the dvt active-profile marker file
// ({dvtRoot}/.active-profile).
func (p *PathConfig) DVTActiveProfile() string {
	return filepath.Join(p.DVTRoot(), ".active-profile")
}

// ---------------------------------------------------------------------------
// Helper methods
// ---------------------------------------------------------------------------

// DatabasePathTilde returns the tilde-notation string
// "~/.devopsmaestro/devopsmaestro.db". This is NOT a real filesystem path —
// it is used as a default config value in viper configurations for the nvp
// and dvt binaries, which expand tilde at runtime.
func (p *PathConfig) DatabasePathTilde() string {
	return filepath.Join("~", DVMDirName, DatabaseFile)
}
