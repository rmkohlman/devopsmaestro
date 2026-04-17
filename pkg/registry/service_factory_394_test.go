package registry

// =============================================================================
// Issue #394: Tests for DetectVersion and GetUptime on ServiceFactory
// =============================================================================
// These tests verify the fix for "dvm get registries doesn't show VERSION or UPTIME".
// DetectVersion and GetUptime were added to service_factory.go to populate the
// VERSION and UPTIME columns in `dvm get registries`.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DetectVersion tests
// =============================================================================

func TestServiceFactory_DetectVersion_ReturnsModelVersion(t *testing.T) {
	// When the registry model already has a version set, it should be returned as-is.
	factory := NewServiceFactory()
	reg := &models.Registry{
		Name:    "my-zot",
		Type:    "zot",
		Port:    5001,
		Version: "2.1.15",
	}

	got := factory.DetectVersion(context.Background(), reg)
	assert.Equal(t, "2.1.15", got, "should return the model's pre-set version")
}

func TestServiceFactory_DetectVersion_FallsBackToStrategyDefault(t *testing.T) {
	// When the registry model has no version set, fall back to the strategy default.
	// Zot's default version is non-empty (defined in NewZotStrategy).
	factory := NewServiceFactory()
	reg := &models.Registry{
		Name:    "my-zot",
		Type:    "zot",
		Port:    5001,
		Version: "", // not set
	}

	strategy, err := factory.GetStrategy("zot")
	require.NoError(t, err)
	defaultVersion := strategy.GetDefaultVersion()

	if defaultVersion == "" {
		t.Skip("zot strategy has no default version configured; skipping fallback test")
	}

	got := factory.DetectVersion(context.Background(), reg)
	assert.Equal(t, defaultVersion, got, "should fall back to the strategy's default version")
}

func TestServiceFactory_DetectVersion_ReturnsEmpty_ForUnknownType(t *testing.T) {
	// Unsupported registry type should return empty string without panic.
	factory := NewServiceFactory()
	reg := &models.Registry{
		Name:    "unknown-reg",
		Type:    "unsupported-type",
		Port:    9999,
		Version: "",
	}

	got := factory.DetectVersion(context.Background(), reg)
	assert.Equal(t, "", got, "unsupported registry type should return empty version string")
}

func TestServiceFactory_DetectVersion_AllSupportedTypes_NoPanic(t *testing.T) {
	// Smoke test: DetectVersion must not panic for any supported registry type.
	factory := NewServiceFactory()

	tests := []struct {
		regType string
		port    int
	}{
		{"zot", 5001},
		{"athens", 3000},
		{"devpi", 3141},
		{"verdaccio", 4873},
		{"squid", 3128},
	}

	for _, tt := range tests {
		t.Run(tt.regType, func(t *testing.T) {
			reg := &models.Registry{
				Name: "test-" + tt.regType,
				Type: tt.regType,
				Port: tt.port,
			}
			assert.NotPanics(t, func() {
				factory.DetectVersion(context.Background(), reg)
			}, "DetectVersion must not panic for type %s", tt.regType)
		})
	}
}

// =============================================================================
// GetUptime tests
// =============================================================================

func TestServiceFactory_GetUptime_ReturnsZero_ForUnknownType(t *testing.T) {
	// An unrecognised registry type has no PID file mapping; expect 0.
	factory := NewServiceFactory()
	reg := &models.Registry{
		Name: "ghost",
		Type: "unsupported-type",
		Port: 9999,
	}

	got := factory.GetUptime(reg)
	assert.Equal(t, time.Duration(0), got, "unknown registry type should return 0 uptime")
}

func TestServiceFactory_GetUptime_ReturnsZero_WhenPidFileMissing(t *testing.T) {
	// Supported type but no PID file on disk → expect 0 (not an error).
	factory := NewServiceFactory()
	reg := &models.Registry{
		Name: "test-zot",
		Type: "zot",
		Port: 5001,
		// Storage not set → resolveStoragePath will fall back to default path
		// which doesn't exist in CI; either way, PID file won't be present.
	}

	got := factory.GetUptime(reg)
	assert.Equal(t, time.Duration(0), got, "missing PID file should yield 0 uptime")
}

func TestServiceFactory_GetUptime_ReturnsDuration_WhenPidFileExists(t *testing.T) {
	// Write a PID file inside ~/.devopsmaestro/registries/<name> so it passes
	// the validateStoragePath security check.
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	storageDir := filepath.Join(homeDir, ".devopsmaestro", "registries", "uptime-test-zot")
	require.NoError(t, os.MkdirAll(storageDir, 0o755))
	t.Cleanup(func() { os.RemoveAll(storageDir) })

	pidFile := filepath.Join(storageDir, "zot.pid")
	require.NoError(t, os.WriteFile(pidFile, []byte("12345"), 0o644))

	// Sleep briefly so the file mod-time is measurably in the past.
	time.Sleep(10 * time.Millisecond)

	factory := NewServiceFactory()
	reg := &models.Registry{
		Name: "uptime-test-zot",
		Type: "zot",
		Port: 5001,
		Config: sql.NullString{
			String: fmt.Sprintf(`{"storage":%q}`, storageDir),
			Valid:  true,
		},
	}

	got := factory.GetUptime(reg)
	assert.Greater(t, got, time.Duration(0), "GetUptime should return positive duration when PID file is present")
}

func TestServiceFactory_GetUptime_AllSupportedTypes_NoPanic(t *testing.T) {
	// Smoke test: GetUptime must not panic for any supported registry type.
	factory := NewServiceFactory()

	tests := []struct {
		regType string
		port    int
	}{
		{"zot", 5001},
		{"athens", 3000},
		{"devpi", 3141},
		{"verdaccio", 4873},
		{"squid", 3128},
	}

	for _, tt := range tests {
		t.Run(tt.regType, func(t *testing.T) {
			reg := &models.Registry{
				Name: "test-" + tt.regType,
				Type: tt.regType,
				Port: tt.port,
			}
			assert.NotPanics(t, func() {
				factory.GetUptime(reg)
			}, "GetUptime must not panic for type %s", tt.regType)
		})
	}
}
