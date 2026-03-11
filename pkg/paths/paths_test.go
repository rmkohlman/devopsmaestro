package paths

import (
	"path/filepath"
	"testing"
)

// fakeHome is the deterministic home directory used across all tests.
// filepath.Join ensures the path uses the correct separator on every platform.
var fakeHome = filepath.Join("/", "home", "testuser")

// ---------------------------------------------------------------------------
// TestNew
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	t.Run("valid homeDir returns non-nil PathConfig", func(t *testing.T) {
		pc := New(fakeHome)
		if pc == nil {
			t.Fatal("New() returned nil, want non-nil *PathConfig")
		}
	})

	t.Run("empty homeDir panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("New(\"\") did not panic, want panic")
			}
		}()
		New("") // should panic
	})
}

// ---------------------------------------------------------------------------
// TestDefault
// ---------------------------------------------------------------------------

func TestDefault(t *testing.T) {
	t.Run("returns non-nil PathConfig without error", func(t *testing.T) {
		pc, err := Default()
		if err != nil {
			t.Fatalf("Default() returned error: %v", err)
		}
		if pc == nil {
			t.Fatal("Default() returned nil PathConfig, want non-nil")
		}
	})
}

// ---------------------------------------------------------------------------
// TestDVMRootPaths
// ---------------------------------------------------------------------------

func TestDVMRootPaths(t *testing.T) {
	pc := New(fakeHome)
	root := filepath.Join(fakeHome, DVMDirName)
	templates := filepath.Join(root, "templates")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"Root",
			pc.Root(),
			root,
		},
		{
			"ConfigFile",
			pc.ConfigFile(),
			filepath.Join(root, "config.yaml"),
		},
		{
			"Database",
			pc.Database(),
			filepath.Join(root, DatabaseFile),
		},
		{
			"VersionFile",
			pc.VersionFile(),
			filepath.Join(root, ".version"),
		},
		{
			"ContextFile",
			pc.ContextFile(),
			filepath.Join(root, "context.yaml"),
		},
		{
			"NvimSyncStatus",
			pc.NvimSyncStatus(),
			filepath.Join(root, ".nvim-sync-status"),
		},
		{
			"LogsDir",
			pc.LogsDir(),
			filepath.Join(root, "logs"),
		},
		{
			"BackupsDir",
			pc.BackupsDir(),
			filepath.Join(root, "backups"),
		},
		{
			"TemplatesDir",
			pc.TemplatesDir(),
			templates,
		},
		{
			"NvimTemplatesDir",
			pc.NvimTemplatesDir(),
			filepath.Join(templates, "nvim"),
		},
		{
			"ShellTemplatesDir",
			pc.ShellTemplatesDir(),
			filepath.Join(templates, "shell"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestWorkspacePaths
// ---------------------------------------------------------------------------

func TestWorkspacePaths(t *testing.T) {
	pc := New(fakeHome)
	root := filepath.Join(fakeHome, DVMDirName)
	wsDir := filepath.Join(root, "workspaces")
	slug := "my-ws"
	wsPath := filepath.Join(wsDir, slug)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"WorkspacesDir",
			pc.WorkspacesDir(),
			wsDir,
		},
		{
			"WorkspacePath",
			pc.WorkspacePath(slug),
			wsPath,
		},
		{
			"WorkspaceRepoPath",
			pc.WorkspaceRepoPath(slug),
			filepath.Join(wsPath, "repo"),
		},
		{
			"WorkspaceVolumePath",
			pc.WorkspaceVolumePath(slug),
			filepath.Join(wsPath, "volume"),
		},
		{
			"WorkspaceConfigPath",
			pc.WorkspaceConfigPath(slug),
			filepath.Join(wsPath, ".dvm"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestWorkspacePathsSlugIsolation verifies that two different slugs produce
// distinct workspace paths (i.e., no slug leakage).
func TestWorkspacePathsSlugIsolation(t *testing.T) {
	pc := New(fakeHome)
	slugA := "workspace-alpha"
	slugB := "workspace-beta"

	tests := []struct {
		name  string
		pathA string
		pathB string
	}{
		{"WorkspacePath differs by slug", pc.WorkspacePath(slugA), pc.WorkspacePath(slugB)},
		{"WorkspaceRepoPath differs by slug", pc.WorkspaceRepoPath(slugA), pc.WorkspaceRepoPath(slugB)},
		{"WorkspaceVolumePath differs by slug", pc.WorkspaceVolumePath(slugA), pc.WorkspaceVolumePath(slugB)},
		{"WorkspaceConfigPath differs by slug", pc.WorkspaceConfigPath(slugA), pc.WorkspaceConfigPath(slugB)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pathA == tt.pathB {
				t.Errorf("paths are identical for different slugs: %q", tt.pathA)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGitAndBuildPaths
// ---------------------------------------------------------------------------

func TestGitAndBuildPaths(t *testing.T) {
	pc := New(fakeHome)
	root := filepath.Join(fakeHome, DVMDirName)
	appName := "my-app"

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"ReposDir",
			pc.ReposDir(),
			filepath.Join(root, "repos"),
		},
		{
			"BuildStagingDir",
			pc.BuildStagingDir(appName),
			filepath.Join(root, "build-staging", appName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRegistryPaths
// ---------------------------------------------------------------------------

func TestRegistryPaths(t *testing.T) {
	pc := New(fakeHome)
	root := filepath.Join(fakeHome, DVMDirName)
	registryName := "local-cache"

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"RegistryDir",
			pc.RegistryDir(registryName),
			filepath.Join(root, "registries", registryName),
		},
		{
			"RegistryStorage",
			pc.RegistryStorage(),
			filepath.Join(root, "registry"),
		},
		{
			"AthensStorage",
			pc.AthensStorage(),
			filepath.Join(root, "athens"),
		},
		{
			"VerdaccioStorage",
			pc.VerdaccioStorage(),
			filepath.Join(root, "verdaccio"),
		},
		{
			"DevpiStorage",
			pc.DevpiStorage(),
			filepath.Join(root, "devpi"),
		},
		{
			"SquidDir",
			pc.SquidDir(),
			filepath.Join(root, "squid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestNVPPaths
// ---------------------------------------------------------------------------

func TestNVPPaths(t *testing.T) {
	pc := New(fakeHome)
	nvpRoot := filepath.Join(fakeHome, NVPDirName)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"NVPRoot",
			pc.NVPRoot(),
			nvpRoot,
		},
		{
			"NVPPluginsDir",
			pc.NVPPluginsDir(),
			filepath.Join(nvpRoot, "plugins"),
		},
		{
			"NVPPackagesDir",
			pc.NVPPackagesDir(),
			filepath.Join(nvpRoot, "packages"),
		},
		{
			"NVPThemesDir",
			pc.NVPThemesDir(),
			filepath.Join(nvpRoot, "themes"),
		},
		{
			"NVPCoreConfig",
			pc.NVPCoreConfig(),
			filepath.Join(nvpRoot, "core.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDVTPaths
// ---------------------------------------------------------------------------

func TestDVTPaths(t *testing.T) {
	pc := New(fakeHome)
	dvtRoot := filepath.Join(fakeHome, DVTDirName)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"DVTRoot",
			pc.DVTRoot(),
			dvtRoot,
		},
		{
			"DVTPromptsDir",
			pc.DVTPromptsDir(),
			filepath.Join(dvtRoot, "prompts"),
		},
		{
			"DVTPluginsDir",
			pc.DVTPluginsDir(),
			filepath.Join(dvtRoot, "plugins"),
		},
		{
			"DVTShellsDir",
			pc.DVTShellsDir(),
			filepath.Join(dvtRoot, "shells"),
		},
		{
			"DVTProfilesDir",
			pc.DVTProfilesDir(),
			filepath.Join(dvtRoot, "profiles"),
		},
		{
			"DVTActiveProfile",
			pc.DVTActiveProfile(),
			filepath.Join(dvtRoot, ".active-profile"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestDatabasePathTilde
// ---------------------------------------------------------------------------

func TestDatabasePathTilde(t *testing.T) {
	pc := New(fakeHome)
	got := pc.DatabasePathTilde()
	want := filepath.Join("~", DVMDirName, DatabaseFile)

	if got != want {
		t.Errorf("DatabasePathTilde() = %q, want %q", got, want)
	}

	// Extra guard: the tilde path must NOT contain the fake home directory,
	// since it is not a real filesystem path.
	if got == filepath.Join(fakeHome, DVMDirName, DatabaseFile) {
		t.Errorf("DatabasePathTilde() returned real path %q; want tilde notation", got)
	}
}

// ---------------------------------------------------------------------------
// TestConstants
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"DVMDirName", DVMDirName, ".devopsmaestro"},
		{"NVPDirName", NVPDirName, ".nvp"},
		{"DVTDirName", DVTDirName, ".dvt"},
		{"DatabaseFile", DatabaseFile, "devopsmaestro.db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("constant %s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestImmutability
// ---------------------------------------------------------------------------

// TestImmutability proves that two PathConfigs with different home directories
// return fully independent paths and share no global state.
func TestImmutability(t *testing.T) {
	homeA := filepath.Join("/", "home", "alice")
	homeB := filepath.Join("/", "home", "bob")

	pcA := New(homeA)
	pcB := New(homeB)

	tests := []struct {
		name  string
		pathA string
		pathB string
	}{
		{"Root", pcA.Root(), pcB.Root()},
		{"ConfigFile", pcA.ConfigFile(), pcB.ConfigFile()},
		{"Database", pcA.Database(), pcB.Database()},
		{"WorkspacesDir", pcA.WorkspacesDir(), pcB.WorkspacesDir()},
		{"WorkspacePath", pcA.WorkspacePath("ws"), pcB.WorkspacePath("ws")},
		{"WorkspaceVolumePath", pcA.WorkspaceVolumePath("ws"), pcB.WorkspaceVolumePath("ws")},
		{"ReposDir", pcA.ReposDir(), pcB.ReposDir()},
		{"NVPRoot", pcA.NVPRoot(), pcB.NVPRoot()},
		{"DVTRoot", pcA.DVTRoot(), pcB.DVTRoot()},
	}

	for _, tt := range tests {
		t.Run(tt.name+" differs between configs", func(t *testing.T) {
			if tt.pathA == tt.pathB {
				t.Errorf(
					"paths are identical despite different home dirs:\n  homeA=%q\n  homeB=%q\n  path=%q",
					homeA, homeB, tt.pathA,
				)
			}
		})
	}

	// Also verify each path is actually rooted under its own home directory.
	t.Run("pathA rooted under homeA", func(t *testing.T) {
		got := pcA.Root()
		wantPrefix := homeA
		if len(got) < len(wantPrefix) || got[:len(wantPrefix)] != wantPrefix {
			t.Errorf("Root() = %q, expected prefix %q", got, wantPrefix)
		}
	})

	t.Run("pathB rooted under homeB", func(t *testing.T) {
		got := pcB.Root()
		wantPrefix := homeB
		if len(got) < len(wantPrefix) || got[:len(wantPrefix)] != wantPrefix {
			t.Errorf("Root() = %q, expected prefix %q", got, wantPrefix)
		}
	})
}

// ---------------------------------------------------------------------------
// TestRootToolsAreDistinct
// ---------------------------------------------------------------------------

// TestRootToolsAreDistinct verifies that the three tool roots (DVM, NVP, DVT)
// are distinct directories — they must never collide.
func TestRootToolsAreDistinct(t *testing.T) {
	pc := New(fakeHome)

	dvmRoot := pc.Root()
	nvpRoot := pc.NVPRoot()
	dvtRoot := pc.DVTRoot()

	if dvmRoot == nvpRoot {
		t.Errorf("DVM root and NVP root are the same: %q", dvmRoot)
	}
	if dvmRoot == dvtRoot {
		t.Errorf("DVM root and DVT root are the same: %q", dvmRoot)
	}
	if nvpRoot == dvtRoot {
		t.Errorf("NVP root and DVT root are the same: %q", nvpRoot)
	}
}
