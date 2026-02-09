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

	yaml := app.ToYAML("test-domain")

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
	yaml := original.ToYAML("my-domain")

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
