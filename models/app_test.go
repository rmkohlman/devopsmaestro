package models

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppGetLanguageConfig_ValidConfig(t *testing.T) {
	app := &App{
		Name:     "test-app",
		Path:     "/path/to/app",
		Language: sql.NullString{String: `{"name":"go","version":"1.22"}`, Valid: true},
	}

	cfg := app.GetLanguageConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "go", cfg.Name)
	assert.Equal(t, "1.22", cfg.Version)
}

func TestAppGetLanguageConfig_EmptyName(t *testing.T) {
	app := &App{
		Name:     "test-app",
		Path:     "/path/to/app",
		Language: sql.NullString{String: `{"name":"","version":"1.22"}`, Valid: true},
	}

	cfg := app.GetLanguageConfig()
	assert.Nil(t, cfg, "should return nil when language name is empty")
}

func TestAppGetLanguageConfig_InvalidJSON(t *testing.T) {
	app := &App{
		Name:     "test-app",
		Path:     "/path/to/app",
		Language: sql.NullString{String: `{invalid json}`, Valid: true},
	}

	cfg := app.GetLanguageConfig()
	assert.Nil(t, cfg, "should return nil for invalid JSON")
}

func TestAppGetLanguageConfig_NullField(t *testing.T) {
	app := &App{
		Name:     "test-app",
		Path:     "/path/to/app",
		Language: sql.NullString{String: "", Valid: false},
	}

	cfg := app.GetLanguageConfig()
	assert.Nil(t, cfg, "should return nil when field is null")
}

func TestAppGetLanguageConfig_EmptyString(t *testing.T) {
	app := &App{
		Name:     "test-app",
		Path:     "/path/to/app",
		Language: sql.NullString{String: "", Valid: true},
	}

	cfg := app.GetLanguageConfig()
	assert.Nil(t, cfg, "should return nil when string is empty")
}

func TestAppGetBuildConfig_ValidConfig(t *testing.T) {
	app := &App{
		Name:        "test-app",
		Path:        "/path/to/app",
		BuildConfig: sql.NullString{String: `{"dockerfile":"Dockerfile.dev","args":{"GO_VERSION":"1.22"}}`, Valid: true},
	}

	cfg := app.GetBuildConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "Dockerfile.dev", cfg.Dockerfile)
	assert.Equal(t, "1.22", cfg.Args["GO_VERSION"])
}

func TestAppGetBuildConfig_WithBuildpack(t *testing.T) {
	app := &App{
		Name:        "test-app",
		Path:        "/path/to/app",
		BuildConfig: sql.NullString{String: `{"buildpack":"go","target":"dev"}`, Valid: true},
	}

	cfg := app.GetBuildConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "go", cfg.Buildpack)
	assert.Equal(t, "dev", cfg.Target)
}

func TestAppGetBuildConfig_InvalidJSON(t *testing.T) {
	app := &App{
		Name:        "test-app",
		Path:        "/path/to/app",
		BuildConfig: sql.NullString{String: `{not valid json}`, Valid: true},
	}

	cfg := app.GetBuildConfig()
	assert.Nil(t, cfg, "should return nil for invalid JSON")
}

func TestAppGetBuildConfig_NullField(t *testing.T) {
	app := &App{
		Name:        "test-app",
		Path:        "/path/to/app",
		BuildConfig: sql.NullString{String: "", Valid: false},
	}

	cfg := app.GetBuildConfig()
	assert.Nil(t, cfg, "should return nil when field is null")
}

func TestAppGetBuildConfig_EmptyString(t *testing.T) {
	app := &App{
		Name:        "test-app",
		Path:        "/path/to/app",
		BuildConfig: sql.NullString{String: "", Valid: true},
	}

	cfg := app.GetBuildConfig()
	assert.Nil(t, cfg, "should return nil when string is empty")
}

func TestAppToYAML_WithLanguageAndBuild(t *testing.T) {
	app := &App{
		Name:        "myapp",
		Path:        "/path/to/myapp",
		Language:    sql.NullString{String: `{"name":"python","version":"3.11"}`, Valid: true},
		BuildConfig: sql.NullString{String: `{"dockerfile":"Dockerfile","args":{"DEBUG":"true"}}`, Valid: true},
	}

	yaml := app.ToYAML("test-domain", nil, "")

	assert.Equal(t, "devopsmaestro.io/v1", yaml.APIVersion)
	assert.Equal(t, "App", yaml.Kind)
	assert.Equal(t, "myapp", yaml.Metadata.Name)
	assert.Equal(t, "test-domain", yaml.Metadata.Domain)
	assert.Equal(t, "/path/to/myapp", yaml.Spec.Path)
	assert.Equal(t, "python", yaml.Spec.Language.Name)
	assert.Equal(t, "3.11", yaml.Spec.Language.Version)
	assert.Equal(t, "Dockerfile", yaml.Spec.Build.Dockerfile)
	assert.Equal(t, "true", yaml.Spec.Build.Args["DEBUG"])
}

func TestAppFromYAML_WithLanguageAndBuild(t *testing.T) {
	yaml := AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:   "testapp",
			Domain: "test-domain",
		},
		Spec: AppSpec{
			Path: "/code/testapp",
			Language: AppLanguageConfig{
				Name:    "node",
				Version: "20",
			},
			Build: AppBuildConfig{
				Dockerfile: "Dockerfile.dev",
				Args: map[string]string{
					"NODE_ENV": "development",
				},
			},
		},
	}

	app := &App{}
	app.FromYAML(yaml)

	assert.Equal(t, "testapp", app.Name)
	assert.Equal(t, "/code/testapp", app.Path)

	// Verify language was stored
	langCfg := app.GetLanguageConfig()
	require.NotNil(t, langCfg)
	assert.Equal(t, "node", langCfg.Name)
	assert.Equal(t, "20", langCfg.Version)

	// Verify build config was stored
	buildCfg := app.GetBuildConfig()
	require.NotNil(t, buildCfg)
	assert.Equal(t, "Dockerfile.dev", buildCfg.Dockerfile)
	assert.Equal(t, "development", buildCfg.Args["NODE_ENV"])
}

func TestAppRoundTrip_ToYAML_FromYAML(t *testing.T) {
	original := &App{
		Name:        "roundtrip-app",
		Path:        "/path/to/app",
		Description: sql.NullString{String: "A test application", Valid: true},
		Language:    sql.NullString{String: `{"name":"rust","version":"1.75"}`, Valid: true},
		BuildConfig: sql.NullString{String: `{"dockerfile":"Dockerfile","target":"release"}`, Valid: true},
	}

	// Convert to YAML
	yaml := original.ToYAML("my-domain", nil, "")

	// Convert back
	restored := &App{}
	restored.FromYAML(yaml)

	// Verify
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Path, restored.Path)
	assert.Equal(t, original.Description.String, restored.Description.String)

	// Compare language config
	origLang := original.GetLanguageConfig()
	restoredLang := restored.GetLanguageConfig()
	require.NotNil(t, origLang)
	require.NotNil(t, restoredLang)
	assert.Equal(t, origLang.Name, restoredLang.Name)
	assert.Equal(t, origLang.Version, restoredLang.Version)

	// Compare build config
	origBuild := original.GetBuildConfig()
	restoredBuild := restored.GetBuildConfig()
	require.NotNil(t, origBuild)
	require.NotNil(t, restoredBuild)
	assert.Equal(t, origBuild.Dockerfile, restoredBuild.Dockerfile)
	assert.Equal(t, origBuild.Target, restoredBuild.Target)
}

// =============================================================================
// v0.55.0 Phase 2 RED Tests: WI-2 — Fix App FromYAML Data Loss
//
// These tests expose the bug where App.FromYAML() silently drops build config
// when neither Dockerfile nor Buildpack is set (args-only, target-only, etc.).
//
// Additionally, tests for AppBuildConfig.IsEmpty() helper (WI-2 design decision
// 15) which replaces the growing condition list in FromYAML.
//
// RED status per test:
//   - TestApp_FromYAML_ArgsOnly_Persisted  → FAILS AT RUNTIME (bug: condition
//     only checks Dockerfile||Buildpack; args-only config is dropped).
//   - TestApp_FromYAML_TargetOnly_Persisted → FAILS AT RUNTIME (same bug).
//   - TestApp_BuildConfig_IsEmpty          → WILL NOT COMPILE (IsEmpty() method
//     does not exist on AppBuildConfig yet — WI-2).
//
// =============================================================================

// TestApp_FromYAML_ArgsOnly_Persisted verifies that when an AppYAML has only
// spec.build.args (no dockerfile or buildpack), FromYAML still persists the
// build config. This tests the WI-2 bug fix.
//
// RED: FAILS AT RUNTIME — current FromYAML condition:
//
//	if yaml.Spec.Build.Dockerfile != "" || yaml.Spec.Build.Buildpack != ""
//
// silently drops args-only build configs. Fix: expand condition or use IsEmpty().
func TestApp_FromYAML_ArgsOnly_Persisted(t *testing.T) {
	appYAML := AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:   "ml-api",
			Domain: "data-science",
		},
		Spec: AppSpec{
			Path: "/code/ml-api",
			Build: AppBuildConfig{
				// No Dockerfile, no Buildpack — args only
				Args: map[string]string{
					"CGO_ENABLED": "0",
					"GOOS":        "linux",
				},
			},
		},
	}

	app := &App{}
	app.FromYAML(appYAML)

	// FAILS TODAY: BuildConfig is nil because Dockerfile=="" && Buildpack==""
	buildCfg := app.GetBuildConfig()
	require.NotNil(t, buildCfg,
		"RED: App.GetBuildConfig() should return non-nil after FromYAML with args-only build config — FAILS until WI-2 fixes the FromYAML condition")

	assert.Equal(t, "0", buildCfg.Args["CGO_ENABLED"],
		"CGO_ENABLED build arg should be persisted via FromYAML")
	assert.Equal(t, "linux", buildCfg.Args["GOOS"],
		"GOOS build arg should be persisted via FromYAML")
}

// TestApp_FromYAML_TargetOnly_Persisted verifies that when an AppYAML has only
// spec.build.target (no dockerfile or buildpack), FromYAML still persists the
// build config. This tests the WI-2 bug fix.
//
// RED: FAILS AT RUNTIME — same root cause as TestApp_FromYAML_ArgsOnly_Persisted.
func TestApp_FromYAML_TargetOnly_Persisted(t *testing.T) {
	appYAML := AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:   "go-service",
			Domain: "backend",
		},
		Spec: AppSpec{
			Path: "/code/go-service",
			Build: AppBuildConfig{
				// No Dockerfile, no Buildpack — target only
				Target: "production",
			},
		},
	}

	app := &App{}
	app.FromYAML(appYAML)

	// FAILS TODAY: BuildConfig is nil because Dockerfile=="" && Buildpack==""
	buildCfg := app.GetBuildConfig()
	require.NotNil(t, buildCfg,
		"RED: App.GetBuildConfig() should return non-nil after FromYAML with target-only build config — FAILS until WI-2 fixes the FromYAML condition")

	assert.Equal(t, "production", buildCfg.Target,
		"Target should be persisted via FromYAML when no Dockerfile/Buildpack is set")
}

// TestApp_FromYAML_ContextOnly_Persisted verifies that when an AppYAML has only
// spec.build.context (no dockerfile, buildpack, args, or target), FromYAML
// still persists the build config. This tests the WI-2 bug fix.
//
// RED: FAILS AT RUNTIME — same root cause as the above two tests.
func TestApp_FromYAML_ContextOnly_Persisted(t *testing.T) {
	appYAML := AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:   "mono-service",
			Domain: "backend",
		},
		Spec: AppSpec{
			Path: "/code/mono",
			Build: AppBuildConfig{
				// Non-standard build context path
				Context: "./services/mono",
			},
		},
	}

	app := &App{}
	app.FromYAML(appYAML)

	// FAILS TODAY: BuildConfig is nil because Dockerfile=="" && Buildpack==""
	buildCfg := app.GetBuildConfig()
	require.NotNil(t, buildCfg,
		"RED: App.GetBuildConfig() should return non-nil after FromYAML with context-only build config — FAILS until WI-2")

	assert.Equal(t, "./services/mono", buildCfg.Context,
		"Context should be persisted via FromYAML when no Dockerfile/Buildpack is set")
}

// TestApp_BuildConfig_IsEmpty verifies the new AppBuildConfig.IsEmpty() helper
// method that returns true only when all fields are zero-valued/empty.
// This method is used by the WI-2 fix to replace the multi-field OR condition.
//
// RED: WILL NOT COMPILE — AppBuildConfig.IsEmpty() method does not exist yet (WI-2).
func TestApp_BuildConfig_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		cfg       AppBuildConfig
		wantEmpty bool
	}{
		{
			name:      "fully empty struct is empty",
			cfg:       AppBuildConfig{},
			wantEmpty: true,
		},
		{
			name: "args only is not empty",
			cfg: AppBuildConfig{
				Args: map[string]string{"CGO_ENABLED": "0"},
			},
			wantEmpty: false,
		},
		{
			name: "dockerfile only is not empty",
			cfg: AppBuildConfig{
				Dockerfile: "Dockerfile.dev",
			},
			wantEmpty: false,
		},
		{
			name: "target only is not empty",
			cfg: AppBuildConfig{
				Target: "production",
			},
			wantEmpty: false,
		},
		{
			name: "context only is not empty",
			cfg: AppBuildConfig{
				Context: "./services/api",
			},
			wantEmpty: false,
		},
		{
			name: "buildpack only is not empty",
			cfg: AppBuildConfig{
				Buildpack: "go",
			},
			wantEmpty: false,
		},
		{
			name: "all fields populated is not empty",
			cfg: AppBuildConfig{
				Dockerfile: "Dockerfile",
				Buildpack:  "go",
				Args:       map[string]string{"CGO_ENABLED": "0"},
				Target:     "production",
				Context:    "./",
			},
			wantEmpty: false,
		},
		{
			name: "empty args map is empty",
			cfg: AppBuildConfig{
				Args: map[string]string{}, // empty map — all fields still zero
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────
			// AppBuildConfig.IsEmpty() does not exist until WI-2 is implemented.
			got := tt.cfg.IsEmpty()
			// ─────────────────────────────────────────────────────────────────
			assert.Equal(t, tt.wantEmpty, got,
				"AppBuildConfig.IsEmpty() mismatch for case %q", tt.name)
		})
	}
}
