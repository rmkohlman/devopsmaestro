package builders

// =============================================================================
// Issue #124 — Integration-level tests: opencode in all 4 user configurations
//
// Tests the end-to-end behavior of the opencode feature across all combinations
// of tools.opencode (CLI tool) and opencode.nvim (nvim plugin):
//
//   Scenario 1: CLI Only  — tools.opencode: true, no opencode nvim plugin
//   Scenario 2: Plugin Only — no tools.opencode, opencode plugin in NvimConfig
//   Scenario 3: Both — tools.opencode: true AND opencode plugin in NvimConfig
//   Scenario 4: Neither — default, no tools and no opencode plugin
//
// These tests focus on the DOCKERFILE boundary: does the opencode builder stage
// appear exactly when and only when tools.opencode is true?
//
// Additionally, the Scenario 2 plugin side is verified: the opencode.nvim plugin
// must exist in the embedded plugin library with the correct repo and metadata.
// =============================================================================

import (
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroNvim/nvimops/library"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// TestOpencode_Scenario1_CLIOnly verifies that when tools.opencode=true but no
// opencode plugin is listed in NvimConfig, the Dockerfile includes the opencode
// builder stage and its COPY directive, with no nvim-plugin-related duplication.
//
// Scenario: workspace YAML has `tools: opencode: true`, NvimConfig.Plugins is empty.
func TestOpencode_Scenario1_CLIOnly(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "cli-only-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{
			Opencode: true,
		},
		// No opencode plugin in Nvim.Plugins
		Nvim: models.NvimConfig{
			Structure:     "lazyvim",
			PluginPackage: "core",
			Plugins:       nil,
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// MUST have builder stage — tools.opencode=true drives this
	wantPatterns := []string{
		"# --- Parallel builder: opencode ---",
		"AS opencode-builder",
		"COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("[Scenario 1 - CLI Only] Generate() missing expected pattern: %q\n"+
				"tools.opencode=true must produce an opencode builder stage.", want)
		}
	}

	// No duplicated builder stage (exactly one occurrence)
	stageCount := strings.Count(dockerfile, "AS opencode-builder")
	if stageCount != 1 {
		t.Errorf("[Scenario 1 - CLI Only] 'AS opencode-builder' appears %d times, want exactly 1", stageCount)
	}
}

// TestOpencode_Scenario2_PluginOnly verifies that when NvimConfig.Plugins
// includes "opencode" but tools.opencode=false, the Dockerfile does NOT
// include the opencode builder stage.
//
// The nvim plugin name in NvimConfig must never trigger a Dockerfile builder
// stage — those are controlled exclusively by the ToolsConfig.
//
// Scenario: workspace YAML has no tools section, but Nvim.Plugins = ["opencode"].
func TestOpencode_Scenario2_PluginOnly(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "python/debian — plugin only", language: "python", version: "3.11"},
		{name: "golang/alpine — plugin only", language: "golang", version: "1.22"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "plugin-only-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{
				// tools.opencode is NOT set (zero value = false)
				Nvim: models.NvimConfig{
					Structure:     "lazyvim",
					PluginPackage: "core",
					Plugins:       []string{"opencode"}, // plugin reference, not CLI tool
				},
			}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// MUST NOT have builder stage — NvimConfig plugins do NOT drive Dockerfile stages
			badPatterns := []string{
				"AS opencode-builder",
				"# --- Parallel builder: opencode ---",
				"COPY --from=opencode-builder",
			}
			for _, bad := range badPatterns {
				if strings.Contains(dockerfile, bad) {
					t.Errorf("[Scenario 2 - Plugin Only] [%s] Dockerfile should NOT contain %q.\n"+
						"NvimConfig.Plugins references are for nvim config, NOT for installing binary tools.\n"+
						"tools.opencode must be explicitly true to trigger the builder stage.", bad, tt.name)
				}
			}
		})
	}
}

// TestOpencode_Scenario3_Both verifies that when BOTH tools.opencode=true AND
// the opencode plugin appears in NvimConfig.Plugins, the Dockerfile includes
// the builder stage exactly once with no conflicts or duplicate stages.
//
// Scenario: workspace YAML has `tools: opencode: true` AND Nvim.Plugins = ["opencode"].
func TestOpencode_Scenario3_Both(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "both-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{
			Opencode: true,
		},
		Nvim: models.NvimConfig{
			Structure:     "lazyvim",
			PluginPackage: "core",
			Plugins:       []string{"opencode"}, // ALSO has the nvim plugin
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// MUST have builder stage — tools.opencode=true
	wantPatterns := []string{
		"# --- Parallel builder: opencode ---",
		"AS opencode-builder",
		"COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("[Scenario 3 - Both] Generate() missing expected pattern: %q", want)
		}
	}

	// MUST NOT have duplicate builder stages — NvimConfig.Plugins must not add a second one
	stageCount := strings.Count(dockerfile, "AS opencode-builder")
	if stageCount != 1 {
		t.Errorf("[Scenario 3 - Both] 'AS opencode-builder' appears %d times, want exactly 1.\n"+
			"Having opencode in both tools AND NvimConfig.Plugins must not produce duplicate stages.",
			stageCount)
	}

	copyCount := strings.Count(dockerfile, "COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode")
	if copyCount != 1 {
		t.Errorf("[Scenario 3 - Both] COPY --from=opencode-builder appears %d times, want exactly 1.\n"+
			"No duplicate COPY directives should appear.", copyCount)
	}
}

// TestOpencode_Scenario4_Neither verifies that with no tools config and no
// opencode plugin, the Dockerfile contains no opencode-related content at all.
// This is the default "clean" workspace — no regression check.
//
// Scenario: zero-value WorkspaceSpec, no tools, no plugins.
func TestOpencode_Scenario4_Neither(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "python/debian — neither", language: "python", version: "3.11"},
		{name: "golang/alpine — neither", language: "golang", version: "1.22"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "neither-ws",
				ImageName: "test:latest",
			}
			// Zero-value WorkspaceSpec: no tools, no opencode plugin
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// MUST NOT have any opencode content
			badPatterns := []string{
				"opencode-builder",
				"opencode",
			}
			for _, bad := range badPatterns {
				if strings.Contains(dockerfile, bad) {
					t.Errorf("[Scenario 4 - Neither] [%s] Dockerfile should NOT contain %q.\n"+
						"Zero-value workspace must produce no opencode content.", bad, tt.name)
				}
			}
		})
	}
}

// TestOpencode_Scenario1vs4_IsolationInvariant verifies the fundamental invariant:
// the ONLY signal that controls opencode builder stage inclusion is
// WorkspaceSpec.Tools.Opencode. NvimConfig is irrelevant to this decision.
//
// This table-driven test captures all four combinations in one place as a
// regression guard.
func TestOpencode_Scenario1vs4_IsolationInvariant(t *testing.T) {
	tests := []struct {
		name             string
		toolsOpencode    bool
		nvimPlugins      []string
		wantBuilderStage bool
	}{
		{
			name:             "scenario 1: CLI only — builder present",
			toolsOpencode:    true,
			nvimPlugins:      nil,
			wantBuilderStage: true,
		},
		{
			name:             "scenario 2: plugin only — builder absent",
			toolsOpencode:    false,
			nvimPlugins:      []string{"opencode"},
			wantBuilderStage: false,
		},
		{
			name:             "scenario 3: both — builder present (exactly once)",
			toolsOpencode:    true,
			nvimPlugins:      []string{"opencode"},
			wantBuilderStage: true,
		},
		{
			name:             "scenario 4: neither — builder absent",
			toolsOpencode:    false,
			nvimPlugins:      nil,
			wantBuilderStage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "invariant-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{
				Tools: models.ToolsConfig{
					Opencode: tt.toolsOpencode,
				},
				Nvim: models.NvimConfig{
					Structure:     "lazyvim",
					PluginPackage: "core",
					Plugins:       tt.nvimPlugins,
				},
			}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      "python",
				Version:       "3.11",
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			hasBuilderStage := strings.Contains(dockerfile, "AS opencode-builder")

			if tt.wantBuilderStage && !hasBuilderStage {
				t.Errorf("[%s] expected 'AS opencode-builder' in Dockerfile but it was absent.\n"+
					"tools.opencode=%v nvimPlugins=%v", tt.name, tt.toolsOpencode, tt.nvimPlugins)
			}
			if !tt.wantBuilderStage && hasBuilderStage {
				t.Errorf("[%s] expected NO 'AS opencode-builder' in Dockerfile but it was present.\n"+
					"tools.opencode=%v nvimPlugins=%v", tt.name, tt.toolsOpencode, tt.nvimPlugins)
			}

			// Extra guard for scenarios 3: no duplicate stages
			if tt.wantBuilderStage {
				stageCount := strings.Count(dockerfile, "AS opencode-builder")
				if stageCount != 1 {
					t.Errorf("[%s] 'AS opencode-builder' appears %d times, want exactly 1", tt.name, stageCount)
				}
			}
		})
	}
}

// TestOpencode_Scenario2_PluginLibraryEntry verifies the plugin-side of Scenario 2:
// the opencode.nvim plugin is present in the embedded plugin library with the
// correct repo and metadata. This confirms that a user who adds "opencode" to
// NvimConfig.Plugins will get a valid, well-formed plugin definition resolved.
//
// This test directly verifies the MaestroNvim plugin library entry for opencode.
func TestOpencode_Scenario2_PluginLibraryEntry(t *testing.T) {
	lib, err := library.NewLibrary()
	if err != nil {
		t.Fatalf("library.NewLibrary() failed: %v", err)
	}

	plugin, ok := lib.Get("opencode")
	if !ok {
		t.Fatal("opencode plugin not found in embedded library — " +
			"expected 53-opencode.yaml in MaestroNvim plugin library")
	}

	// Name must be correct
	if plugin.Name != "opencode" {
		t.Errorf("plugin.Name = %q, want %q", plugin.Name, "opencode")
	}

	// Must have a repo — required for lazy.nvim to install
	if plugin.Repo == "" {
		t.Errorf("[Scenario 2 - Plugin Only] opencode plugin has empty Repo.\n" +
			"The repo field is required for nvim plugin installation.")
	}
	wantRepo := "nickjvandyke/opencode.nvim"
	if plugin.Repo != wantRepo {
		t.Errorf("opencode plugin Repo = %q, want %q", plugin.Repo, wantRepo)
	}

	// Must have a description
	if plugin.Description == "" {
		t.Errorf("[Scenario 2 - Plugin Only] opencode plugin has empty Description.")
	}

	// Must be in the "ai" category to be discoverable by category
	if plugin.Category != "ai" {
		t.Errorf("opencode plugin Category = %q, want %q", plugin.Category, "ai")
	}

	// Must have the "opencode" tag for discoverability
	hasOpencodeTag := false
	for _, tag := range plugin.Tags {
		if tag == "opencode" {
			hasOpencodeTag = true
			break
		}
	}
	if !hasOpencodeTag {
		t.Errorf("opencode plugin Tags = %v, want to contain %q", plugin.Tags, "opencode")
	}

	// Must have config content — this is what gets generated into Lua
	if plugin.Config == "" {
		t.Errorf("[Scenario 2 - Plugin Only] opencode plugin has empty Config.\n" +
			"Config is required to generate the Lua setup block.")
	}
}
