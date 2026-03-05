// Package registry - TDD RED phase tests for security bugs.
//
// These tests are intentionally FAILING. They document the correct behavior
// that must be implemented. Do NOT fix the production code until Phase 3.
//
// Bug list:
//   B3  - ZotManager.GetEndpoint() missing http:// prefix
//   B4  - DefaultProcessManager has no mutex (race detector)
//   B8  - All config generators bind to 0.0.0.0 instead of 127.0.0.1
//   B10 - waitForReady uses http.Get (no timeout, not context-aware)

package registry

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// =============================================================================
// B3: ZotManager.GetEndpoint() must return http://localhost:{port}
//
// Current behaviour:  "localhost:5001"   (missing scheme)
// Required behaviour: "http://localhost:5001"
//
// Without the scheme the endpoint cannot be used directly as a URL in Docker
// client calls or health-check HTTP requests.
// =============================================================================

func TestZotManager_GetEndpoint_HasHTTPPrefix(t *testing.T) {
	zm := NewZotManagerWithDeps(
		RegistryConfig{Port: 5001, Storage: t.TempDir(), Enabled: true},
		&MockBinaryManager{},
		NewProcessManager(ProcessConfig{}),
	)
	ep := zm.GetEndpoint()
	if !strings.HasPrefix(ep, "http://") {
		t.Errorf("GetEndpoint() = %q, want http:// prefix", ep)
	}
}

func TestZotManager_GetEndpoint_CorrectFullURL(t *testing.T) {
	tests := []struct {
		name string
		port int
		want string
	}{
		{"default port", 5001, "http://localhost:5001"},
		{"custom port", 5555, "http://localhost:5555"},
		{"high port", 49152, "http://localhost:49152"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zm := NewZotManagerWithDeps(
				RegistryConfig{Port: tt.port, Storage: t.TempDir(), Enabled: true},
				&MockBinaryManager{},
				NewProcessManager(ProcessConfig{}),
			)
			got := zm.GetEndpoint()
			if got != tt.want {
				t.Errorf("GetEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// B4: DefaultProcessManager concurrent access — data race without mutex
//
// The struct has `cmd *exec.Cmd` and `pid int` fields that are read/written
// from multiple goroutines without synchronisation.  This test passes normally
// but triggers the race detector with `go test -race`.
// =============================================================================

func TestDefaultProcessManager_ConcurrentAccess(t *testing.T) {
	pm := &DefaultProcessManager{}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pm.IsRunning()
			_ = pm.GetPID()
		}()
	}
	wg.Wait()
}

func TestDefaultProcessManager_ConcurrentReadWrite(t *testing.T) {
	// Verify that concurrent reads and Stop() calls do not race.
	// Before the B4 fix, DefaultProcessManager had no mutex and this test
	// would trigger the race detector.  Now that the mutex is in place,
	// all access goes through the public API which holds the lock.
	pm := &DefaultProcessManager{}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = pm.IsRunning()
		}()
		go func() {
			defer wg.Done()
			_ = pm.GetPID()
		}()
		go func() {
			defer wg.Done()
			// Use the public Stop() API which now holds the mutex.
			// On a fresh/stopped manager this is a no-op (idempotent).
			_ = pm.Stop(context.Background())
		}()
	}
	wg.Wait()
}

// =============================================================================
// B8: Config generators must bind to 127.0.0.1, not 0.0.0.0
//
// Binding to 0.0.0.0 exposes local registry services on all network interfaces
// (including bridged/host-network containers and LAN). All generated configs
// must restrict listening to the loopback address only.
//
// Five sub-checks:
//   B8a - GenerateZotConfig http.address == "127.0.0.1"
//   B8b - GenerateAthensConfig Port field == "127.0.0.1:{port}"
//   B8c - GenerateVerdaccioConfig listen field == "127.0.0.1:{port}"
//   B8d - GenerateSquidConfig http_port directive == "127.0.0.1:{port}"
//   B8e - Generated configs do NOT contain "0.0.0.0"
// =============================================================================

// B8a — Zot config.go: http.address must be "127.0.0.1"
func TestGenerateZotConfig_BindsToLoopback(t *testing.T) {
	cfg := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
	}

	zotConfig, err := GenerateZotConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateZotConfig() error = %v", err)
	}

	httpSection, ok := zotConfig["http"].(map[string]interface{})
	if !ok {
		t.Fatal("GenerateZotConfig() missing http section")
	}

	addr, ok := httpSection["address"].(string)
	if !ok {
		t.Fatal("GenerateZotConfig() http section missing address field")
	}

	if addr != "127.0.0.1" {
		t.Errorf("http.address = %q, want %q (security: must not bind to 0.0.0.0)", addr, "127.0.0.1")
	}
}

func TestGenerateZotConfig_DoesNotContain_0_0_0_0(t *testing.T) {
	cfg := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
	}

	zotConfig, err := GenerateZotConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateZotConfig() error = %v", err)
	}

	httpSection, ok := zotConfig["http"].(map[string]interface{})
	if !ok {
		t.Fatal("GenerateZotConfig() missing http section")
	}

	addr, _ := httpSection["address"].(string)
	if addr == "0.0.0.0" {
		t.Errorf("http.address = %q: binding to 0.0.0.0 is a security vulnerability; use 127.0.0.1", addr)
	}
}

// B8b — Athens config_athens.go: Port line must include 127.0.0.1
// Currently generates: `Port = ":3000"` (binds all interfaces)
// Required:            `Port = "127.0.0.1:3000"`
func TestGenerateAthensConfig_BindsToLoopback(t *testing.T) {
	cfg := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   "/tmp/athens",
	}

	athensConfig, err := GenerateAthensConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateAthensConfig() error = %v", err)
	}

	// The Port TOML value must include the loopback address.
	// Acceptable forms: `Port = "127.0.0.1:3000"`
	if !strings.Contains(athensConfig, "127.0.0.1") {
		t.Errorf("GenerateAthensConfig() config does not contain 127.0.0.1:\n%s", athensConfig)
	}
}

func TestGenerateAthensConfig_DoesNotBindAllInterfaces(t *testing.T) {
	cfg := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   "/tmp/athens",
	}

	athensConfig, err := GenerateAthensConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateAthensConfig() error = %v", err)
	}

	// The current buggy form is `Port = ":3000"` which binds 0.0.0.0.
	// After the fix the colon-only prefix must not appear as the port value.
	if strings.Contains(athensConfig, `Port = ":`) {
		t.Errorf(
			"GenerateAthensConfig() Port value uses bare \":port\" syntax which binds 0.0.0.0; "+
				"use \"127.0.0.1:port\" instead.\nConfig:\n%s",
			athensConfig,
		)
	}
}

// B8c — Verdaccio config_verdaccio.go: listen field must be "127.0.0.1:{port}"
// Currently generates: `listen: 0.0.0.0:4873`
// Required:            `listen: 127.0.0.1:4873`
func TestGenerateVerdaccioConfig_BindsToLoopback(t *testing.T) {
	cfg := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "/tmp/verdaccio",
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateVerdaccioConfig() error = %v", err)
	}

	if !strings.Contains(verdaccioConfig, "127.0.0.1") {
		t.Errorf(
			"GenerateVerdaccioConfig() YAML does not contain 127.0.0.1 "+
				"(security: listen field must not bind to 0.0.0.0).\nConfig:\n%s",
			verdaccioConfig,
		)
	}
}

func TestGenerateVerdaccioConfig_ListenField_IsLoopback(t *testing.T) {
	cfg := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "/tmp/verdaccio",
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateVerdaccioConfig() error = %v", err)
	}

	// The listen value "0.0.0.0:4873" must NOT appear.
	if strings.Contains(verdaccioConfig, "0.0.0.0") {
		t.Errorf(
			"GenerateVerdaccioConfig() listen field contains 0.0.0.0; "+
				"must use 127.0.0.1 to restrict to loopback.\nConfig:\n%s",
			verdaccioConfig,
		)
	}

	// And the correct loopback form must be present.
	want := "127.0.0.1:4873"
	if !strings.Contains(verdaccioConfig, want) {
		t.Errorf("GenerateVerdaccioConfig() listen field does not contain %q.\nConfig:\n%s", want, verdaccioConfig)
	}
}

// B8d — Squid squid_manager.go: http_port directive must specify 127.0.0.1
// Currently generates: `http_port 3128`     (binds 0.0.0.0)
// Required:            `http_port 127.0.0.1:3128`
func TestGenerateSquidConfig_BindsToLoopback(t *testing.T) {
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	squidConfig, err := GenerateSquidConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateSquidConfig() error = %v", err)
	}

	if !strings.Contains(squidConfig, "127.0.0.1") {
		t.Errorf(
			"GenerateSquidConfig() squid.conf does not contain 127.0.0.1 "+
				"(security: http_port must bind to loopback only).\nConfig:\n%s",
			squidConfig,
		)
	}
}

func TestGenerateSquidConfig_HTTPPort_IsLoopback(t *testing.T) {
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	squidConfig, err := GenerateSquidConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateSquidConfig() error = %v", err)
	}

	// The correct directive form after the fix.
	want := "http_port 127.0.0.1:3128"
	if !strings.Contains(squidConfig, want) {
		t.Errorf(
			"GenerateSquidConfig() http_port directive = (see config), want %q.\nConfig:\n%s",
			want, squidConfig,
		)
	}
}

func TestGenerateSquidConfig_DoesNotBindAllInterfaces(t *testing.T) {
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	squidConfig, err := GenerateSquidConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateSquidConfig() error = %v", err)
	}

	// The bare `http_port 3128` form implicitly binds 0.0.0.0.
	// After the fix this exact pattern must not appear.
	barePort := "http_port 3128\n"
	if strings.Contains(squidConfig, barePort) {
		t.Errorf(
			"GenerateSquidConfig() contains bare %q which binds 0.0.0.0; "+
				"use \"http_port 127.0.0.1:3128\" instead.\nConfig:\n%s",
			strings.TrimRight(barePort, "\n"), squidConfig,
		)
	}
}
