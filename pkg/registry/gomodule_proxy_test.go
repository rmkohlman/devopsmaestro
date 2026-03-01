package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestGoModuleProxy_InterfaceCompliance(t *testing.T) {
	// Verify AthensManager implements GoModuleProxy interface
	var _ GoModuleProxy = (*AthensManager)(nil)
}

func TestMockGoModuleProxy_InterfaceCompliance(t *testing.T) {
	// Verify MockGoModuleProxy implements GoModuleProxy interface
	var _ GoModuleProxy = (*MockGoModuleProxy)(nil)
}

// =============================================================================
// Mock Implementation Tests
// =============================================================================

func TestMockGoModuleProxy_Start(t *testing.T) {
	mock := &MockGoModuleProxy{
		StartFunc: func(ctx context.Context) error {
			return nil
		},
	}

	err := mock.Start(context.Background())
	assert.NoError(t, err, "MockGoModuleProxy.Start should call StartFunc")
}

func TestMockGoModuleProxy_Stop(t *testing.T) {
	mock := &MockGoModuleProxy{
		StopFunc: func(ctx context.Context) error {
			return nil
		},
	}

	err := mock.Stop(context.Background())
	assert.NoError(t, err, "MockGoModuleProxy.Stop should call StopFunc")
}

func TestMockGoModuleProxy_Status(t *testing.T) {
	expectedStatus := &GoModuleProxyStatus{
		State: "running",
		PID:   12345,
	}

	mock := &MockGoModuleProxy{
		StatusFunc: func(ctx context.Context) (*GoModuleProxyStatus, error) {
			return expectedStatus, nil
		},
	}

	status, err := mock.Status(context.Background())
	assert.NoError(t, err, "MockGoModuleProxy.Status should call StatusFunc")
	assert.Equal(t, expectedStatus, status)
}

func TestMockGoModuleProxy_EnsureRunning(t *testing.T) {
	mock := &MockGoModuleProxy{
		EnsureRunningFunc: func(ctx context.Context) error {
			return nil
		},
	}

	err := mock.EnsureRunning(context.Background())
	assert.NoError(t, err, "MockGoModuleProxy.EnsureRunning should call EnsureRunningFunc")
}

func TestMockGoModuleProxy_IsRunning(t *testing.T) {
	mock := &MockGoModuleProxy{
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
	}

	running := mock.IsRunning(context.Background())
	assert.True(t, running, "MockGoModuleProxy.IsRunning should call IsRunningFunc")
}

func TestMockGoModuleProxy_GetEndpoint(t *testing.T) {
	mock := &MockGoModuleProxy{
		GetEndpointFunc: func() string {
			return "http://localhost:3000"
		},
	}

	endpoint := mock.GetEndpoint()
	assert.Equal(t, "http://localhost:3000", endpoint)
}

func TestMockGoModuleProxy_GetGoEnv(t *testing.T) {
	expectedEnv := map[string]string{
		"GOPROXY": "http://localhost:3000",
	}

	mock := &MockGoModuleProxy{
		GetGoEnvFunc: func() map[string]string {
			return expectedEnv
		},
	}

	env := mock.GetGoEnv()
	assert.Equal(t, expectedEnv, env)
}
