package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// =============================================================================
// Prepare() — ErrBinaryNotInstalled distinction tests
// =============================================================================

// TestPrepare_BinaryNotInstalled_UsesActionableWarning verifies that when
// manager.Start() returns an error wrapping ErrBinaryNotInstalled, Prepare()
// records a warning message containing the install hint (not a generic error).
func TestPrepare_BinaryNotInstalled_UsesActionableWarning(t *testing.T) {
	store := db.NewMockDataStore()
	squidReg := &models.Registry{
		ID:        1,
		Name:      "squid-proxy",
		Type:      "squid",
		Enabled:   true,
		Lifecycle: "on-demand",
		Port:      3128,
	}
	store.Registries[squidReg.Name] = squidReg

	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					return fmt.Errorf("%w: squid (install with: brew install squid)",
						ErrBinaryNotInstalled)
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err, "ErrBinaryNotInstalled must be non-fatal")
	require.NotNil(t, result)
	require.Len(t, result.Warnings, 1, "exactly one warning should be recorded")

	w := result.Warnings[0]
	assert.Contains(t, w, "brew install", "warning must include the install command")
	assert.Contains(t, w, "squid-proxy", "warning must name the registry")
	assert.NotContains(t, w, "failed to start:", "should not use the generic 'failed to start' message")
}

// TestPrepare_BinaryInstalledButFailed_UsesGenericWarning verifies that when
// manager.Start() returns an error NOT wrapping ErrBinaryNotInstalled (binary
// is present but crashed), Prepare() records the generic "installed but failed
// to start" warning — not the install-hint message.
func TestPrepare_BinaryInstalledButFailed_UsesGenericWarning(t *testing.T) {
	store := db.NewMockDataStore()
	squidReg := &models.Registry{
		ID:        1,
		Name:      "squid-proxy",
		Type:      "squid",
		Enabled:   true,
		Lifecycle: "on-demand",
		Port:      3128,
	}
	store.Registries[squidReg.Name] = squidReg

	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					// Not ErrBinaryNotInstalled — binary exists but config is broken
					return fmt.Errorf("config file invalid: /etc/squid/squid.conf")
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err, "a start failure must be non-fatal")
	require.NotNil(t, result)
	require.Len(t, result.Warnings, 1)

	w := result.Warnings[0]
	assert.Contains(t, w, "installed but failed to start", "generic message for non-install errors")
	assert.NotContains(t, w, "brew install", "install hint must not appear for runtime failures")
}

// TestPrepare_BinaryNotInstalled_MarksUnhealthy ensures the registry is added
// to CacheReadiness.Unhealthy regardless of the error type.
func TestPrepare_BinaryNotInstalled_MarksUnhealthy(t *testing.T) {
	store := db.NewMockDataStore()
	store.Registries["squid-proxy"] = &models.Registry{
		ID: 1, Name: "squid-proxy", Type: "squid",
		Enabled: true, Lifecycle: "on-demand", Port: 3128,
	}

	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					return fmt.Errorf("%w: squid", ErrBinaryNotInstalled)
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	assert.Contains(t, result.CacheReadiness.Unhealthy, "squid-proxy",
		"missing binary must mark registry as unhealthy")
	assert.False(t, result.CacheReadiness.AllHealthy)
	assert.Equal(t, 0, result.CacheReadiness.HealthyCount)
}

// TestPrepare_MixedErrors_BothMessageTypesPresent verifies that when one registry
// fails with ErrBinaryNotInstalled and another fails with a generic error, both
// correct message types appear in the Warnings slice.
func TestPrepare_MixedErrors_BothMessageTypesPresent(t *testing.T) {
	store := db.NewMockDataStore()
	store.Registries["squid-proxy"] = &models.Registry{
		ID: 1, Name: "squid-proxy", Type: "squid",
		Enabled: true, Lifecycle: "on-demand", Port: 3128,
	}
	store.Registries["pip-proxy"] = &models.Registry{
		ID: 2, Name: "pip-proxy", Type: "devpi",
		Enabled: true, Lifecycle: "on-demand", Port: 3141,
	}

	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					if reg.Name == "squid-proxy" {
						return fmt.Errorf("%w: squid", ErrBinaryNotInstalled)
					}
					return fmt.Errorf("devpi process crashed")
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.Len(t, result.Warnings, 2)

	warnText := result.Warnings[0] + " " + result.Warnings[1]
	assert.Contains(t, warnText, "brew install", "actionable install hint must appear")
	assert.Contains(t, warnText, "installed but failed to start", "generic message must appear")
}
