package registry

// Tests for bug fixes in issue #386:
//   - EndpointForVM() rewrites localhost/127.0.0.1 → host.docker.internal
//   - SquidManager.IsRunning() falls back to port check when PID file is absent
//   - VerdaccioManager.IsRunning() falls back to port check when PID file is absent

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// EndpointForVM tests (buildkit_config.go)
// =============================================================================

func TestEndpointForVM_RewritesLocalhost(t *testing.T) {
	got := EndpointForVM("localhost:5001")
	assert.Equal(t, "host.docker.internal:5001", got)
}

func TestEndpointForVM_Rewrites127001(t *testing.T) {
	got := EndpointForVM("127.0.0.1:5001")
	assert.Equal(t, "host.docker.internal:5001", got)
}

func TestEndpointForVM_LeavesNonLocalhostUnchanged(t *testing.T) {
	got := EndpointForVM("192.168.1.10:5001")
	assert.Equal(t, "192.168.1.10:5001", got)
}

func TestEndpointForVM_LeavesHostDockerInternalUnchanged(t *testing.T) {
	got := EndpointForVM("host.docker.internal:5001")
	assert.Equal(t, "host.docker.internal:5001", got)
}

func TestEndpointForVM_HandlesFullURL(t *testing.T) {
	got := EndpointForVM("http://localhost:5001")
	assert.Equal(t, "http://host.docker.internal:5001", got)
}

// =============================================================================
// Stub ProcessManager for port-fallback tests
// =============================================================================

// stubProcessManager is a ProcessManager that always reports IsRunning=false,
// simulating a missing/stale PID file.
type stubProcessManager struct{}

func (s *stubProcessManager) Start(_ context.Context, _ string, _ []string, _ ProcessConfig) error {
	return nil
}
func (s *stubProcessManager) Stop(_ context.Context) error { return nil }
func (s *stubProcessManager) IsRunning() bool              { return false }
func (s *stubProcessManager) GetPID() int                  { return 0 }

// grabPort listens on an ephemeral port on all interfaces and returns the port + closer.
// We bind to 0.0.0.0 so IsPortAvailable (which also binds to 0.0.0.0) will see the conflict.
func grabPort(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

// =============================================================================
// SquidManager port-fallback IsRunning tests (#386)
// =============================================================================

func TestSquidManager_IsRunning_PortFallback_ReturnsTrueWhenPortOccupied(t *testing.T) {
	port, closePort := grabPort(t)
	defer closePort()

	cfg := HttpProxyConfig{Port: port, LogDir: t.TempDir(), CacheDir: t.TempDir(), PidFile: t.TempDir() + "/squid.pid"}
	cfg.ApplyDefaults()
	mgr := NewSquidManagerWithDeps(cfg, NewMockBrewBinaryManager(t.TempDir(), "6.0"), &stubProcessManager{})

	// PID manager says not running, but port is occupied → should report running.
	assert.True(t, mgr.IsRunning(context.Background()), "expected IsRunning=true when port is occupied")
}

func TestSquidManager_IsRunning_PortFallback_ReturnsFalseWhenPortFree(t *testing.T) {
	port, closePort := grabPort(t)
	closePort() // release immediately so port is free

	cfg := HttpProxyConfig{Port: port, LogDir: t.TempDir(), CacheDir: t.TempDir(), PidFile: t.TempDir() + "/squid.pid"}
	cfg.ApplyDefaults()
	mgr := NewSquidManagerWithDeps(cfg, NewMockBrewBinaryManager(t.TempDir(), "6.0"), &stubProcessManager{})

	assert.False(t, mgr.IsRunning(context.Background()), "expected IsRunning=false when port is free")
}

// =============================================================================
// VerdaccioManager port-fallback IsRunning tests (#386)
// =============================================================================

func makeVerdaccioManager(t *testing.T, port int) *VerdaccioManager {
	t.Helper()
	cfg := NpmProxyConfig{
		Port:      port,
		Storage:   t.TempDir(),
		Lifecycle: "persistent",
	}
	bin := NewMockBrewBinaryManager(t.TempDir(), "5.0")
	proc := &stubProcessManager{}
	mgr, err := NewVerdaccioManager(cfg, bin, proc)
	require.NoError(t, err)
	return mgr
}

func TestVerdaccioManager_IsRunning_PortFallback_ReturnsTrueWhenPortOccupied(t *testing.T) {
	port, closePort := grabPort(t)
	defer closePort()

	mgr := makeVerdaccioManager(t, port)
	assert.True(t, mgr.IsRunning(context.Background()), "expected IsRunning=true when port is occupied")
}

func TestVerdaccioManager_IsRunning_PortFallback_ReturnsFalseWhenPortFree(t *testing.T) {
	port, closePort := grabPort(t)
	closePort()

	mgr := makeVerdaccioManager(t, port)
	assert.False(t, mgr.IsRunning(context.Background()), "expected IsRunning=false when port is free")
}
