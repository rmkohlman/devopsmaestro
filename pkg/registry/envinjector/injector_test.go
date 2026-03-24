package envinjector

import (
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
)

// TestEnvironmentInjector_InjectForAttach_PyPI tests PyPI env var injection
func TestEnvironmentInjector_InjectForAttach_PyPI(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-devpi",
		Type: "devpi",
		Port: 3141,
	}

	envVars := injector.InjectForAttach(registry)

	// Should inject PIP_INDEX_URL
	assert.Contains(t, envVars, "PIP_INDEX_URL")
	assert.Equal(t, "http://localhost:3141/root/pypi/+simple/", envVars["PIP_INDEX_URL"])

	// Should inject PIP_TRUSTED_HOST for localhost
	assert.Contains(t, envVars, "PIP_TRUSTED_HOST")
	assert.Equal(t, "localhost", envVars["PIP_TRUSTED_HOST"])
}

// TestEnvironmentInjector_InjectForAttach_NPM tests NPM env var injection
func TestEnvironmentInjector_InjectForAttach_NPM(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-verdaccio",
		Type: "verdaccio",
		Port: 4873,
	}

	envVars := injector.InjectForAttach(registry)

	// Should inject NPM_CONFIG_REGISTRY
	assert.Contains(t, envVars, "NPM_CONFIG_REGISTRY")
	assert.Equal(t, "http://localhost:4873/", envVars["NPM_CONFIG_REGISTRY"])

	// Also check uppercase variant
	assert.Contains(t, envVars, "npm_config_registry")
	assert.Equal(t, "http://localhost:4873/", envVars["npm_config_registry"])
}

// TestEnvironmentInjector_InjectForAttach_Go tests Go env var injection
func TestEnvironmentInjector_InjectForAttach_Go(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-athens",
		Type: "athens",
		Port: 3000,
	}

	envVars := injector.InjectForAttach(registry)

	// Should inject GOPROXY
	assert.Contains(t, envVars, "GOPROXY")
	assert.Equal(t, "http://localhost:3000", envVars["GOPROXY"])

	// Should inject GONOSUMDB for privacy
	assert.Contains(t, envVars, "GONOSUMDB")
	assert.Equal(t, "*", envVars["GONOSUMDB"])

	// Should inject GOPRIVATE
	assert.Contains(t, envVars, "GOPRIVATE")
	assert.Equal(t, "*", envVars["GOPRIVATE"])
}

// TestEnvironmentInjector_InjectForAttach_HTTP tests HTTP proxy env var injection
func TestEnvironmentInjector_InjectForAttach_HTTP(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-squid",
		Type: "squid",
		Port: 3128,
	}

	envVars := injector.InjectForAttach(registry)

	// Should inject HTTP_PROXY
	assert.Contains(t, envVars, "HTTP_PROXY")
	assert.Equal(t, "http://localhost:3128", envVars["HTTP_PROXY"])

	// Should inject HTTPS_PROXY
	assert.Contains(t, envVars, "HTTPS_PROXY")
	assert.Equal(t, "http://localhost:3128", envVars["HTTPS_PROXY"])

	// Should inject NO_PROXY
	assert.Contains(t, envVars, "NO_PROXY")
	assert.Contains(t, envVars["NO_PROXY"], "localhost")
	assert.Contains(t, envVars["NO_PROXY"], "127.0.0.1")
}

// TestEnvironmentInjector_InjectForBuild tests build-time env vars
func TestEnvironmentInjector_InjectForBuild(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-devpi",
		Type: "devpi",
		Port: 3141,
	}

	// Build-time env vars should use host.docker.internal for Docker builds
	envVars := injector.InjectForBuild(registry)

	assert.Contains(t, envVars, "PIP_INDEX_URL")
	assert.Contains(t, envVars["PIP_INDEX_URL"], "host.docker.internal")
	assert.Equal(t, "http://host.docker.internal:3141/root/pypi/+simple/", envVars["PIP_INDEX_URL"])
}

// TestEnvironmentInjector_PIPTrustedHostOnlyLocal tests PIP_TRUSTED_HOST only for localhost
func TestEnvironmentInjector_PIPTrustedHostOnlyLocal(t *testing.T) {
	tests := []struct {
		name            string
		registryHost    string
		shouldHaveTrust bool
	}{
		{
			name:            "localhost should have PIP_TRUSTED_HOST",
			registryHost:    "localhost",
			shouldHaveTrust: true,
		},
		{
			name:            "127.0.0.1 should have PIP_TRUSTED_HOST",
			registryHost:    "127.0.0.1",
			shouldHaveTrust: true,
		},
		{
			name:            "external host should NOT have PIP_TRUSTED_HOST",
			registryHost:    "registry.example.com",
			shouldHaveTrust: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := NewEnvironmentInjector()

			// Mock registry with custom host
			// This would require the injector to accept a host parameter
			// For now, we test the IsLocalHost logic
			envVars := injector.InjectForAttachWithHost(&models.Registry{
				Name: "test-devpi",
				Type: "devpi",
				Port: 3141,
			}, tt.registryHost)

			if tt.shouldHaveTrust {
				assert.Contains(t, envVars, "PIP_TRUSTED_HOST")
			} else {
				assert.NotContains(t, envVars, "PIP_TRUSTED_HOST", "Should not set PIP_TRUSTED_HOST for external hosts")
			}
		})
	}
}

// TestEnvironmentInjector_MultipleRegistries tests combining env vars from multiple registries
func TestEnvironmentInjector_MultipleRegistries(t *testing.T) {
	injector := NewEnvironmentInjector()

	registries := []*models.Registry{
		{Name: "my-devpi", Type: "devpi", Port: 3141},
		{Name: "my-verdaccio", Type: "verdaccio", Port: 4873},
		{Name: "my-athens", Type: "athens", Port: 3000},
	}

	envVars := injector.InjectForAttachMultiple(registries)

	// Should have env vars for all registries
	assert.Contains(t, envVars, "PIP_INDEX_URL")
	assert.Contains(t, envVars, "NPM_CONFIG_REGISTRY")
	assert.Contains(t, envVars, "GOPROXY")

	// Verify correct values
	assert.Equal(t, "http://localhost:3141/root/pypi/+simple/", envVars["PIP_INDEX_URL"])
	assert.Equal(t, "http://localhost:4873/", envVars["NPM_CONFIG_REGISTRY"])
	assert.Equal(t, "http://localhost:3000", envVars["GOPROXY"])
}

// TestEnvironmentInjector_OCI tests OCI registry env vars (if any)
func TestEnvironmentInjector_OCI(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-zot",
		Type: "zot",
		Port: 5000,
	}

	envVars := injector.InjectForAttach(registry)

	// OCI registries typically don't need env vars (use docker config instead)
	// But we might inject DOCKER_REGISTRY or similar for convenience
	// For now, this test documents the expected behavior

	// Could inject registry endpoint for scripts
	if val, ok := envVars["DVM_OCI_REGISTRY"]; ok {
		assert.Equal(t, "localhost:5000", val)
	}
}

// TestEnvironmentInjector_EmptyRegistry tests handling of nil/empty registry
func TestEnvironmentInjector_EmptyRegistry(t *testing.T) {
	injector := NewEnvironmentInjector()

	// Should not panic with nil registry
	envVars := injector.InjectForAttach(nil)
	assert.NotNil(t, envVars)
	assert.Empty(t, envVars, "Nil registry should return empty map")
}

// TestEnvironmentInjector_CustomPort tests registries with non-default ports
func TestEnvironmentInjector_CustomPort(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-devpi",
		Type: "devpi",
		Port: 8080, // Custom port
	}

	envVars := injector.InjectForAttach(registry)

	assert.Contains(t, envVars, "PIP_INDEX_URL")
	assert.Contains(t, envVars["PIP_INDEX_URL"], ":8080")
}

// TestEnvironmentInjector_InjectForDocker tests Docker-specific injection
func TestEnvironmentInjector_InjectForDocker(t *testing.T) {
	injector := NewEnvironmentInjector()

	registries := []*models.Registry{
		{Name: "my-zot", Type: "zot", Port: 5000},
	}

	// For Docker, should use host.docker.internal
	dockerEnv := injector.InjectForDocker(registries)

	// Verify Docker-specific variables
	assert.NotEmpty(t, dockerEnv)
}

// TestEnvironmentInjector_CaseSensitivity tests case sensitivity of env vars
func TestEnvironmentInjector_CaseSensitivity(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-squid",
		Type: "squid",
		Port: 3128,
	}

	envVars := injector.InjectForAttach(registry)

	// HTTP_PROXY should be uppercase
	assert.Contains(t, envVars, "HTTP_PROXY")
	assert.Contains(t, envVars, "HTTPS_PROXY")
	assert.Contains(t, envVars, "NO_PROXY")

	// Some tools expect lowercase variants
	assert.Contains(t, envVars, "http_proxy")
	assert.Contains(t, envVars, "https_proxy")
	assert.Contains(t, envVars, "no_proxy")
}

// TestEnvironmentInjector_PathFormat tests correct path formatting
func TestEnvironmentInjector_PathFormat(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-devpi",
		Type: "devpi",
		Port: 3141,
	}

	envVars := injector.InjectForAttach(registry)

	pipURL := envVars["PIP_INDEX_URL"]

	// Should have trailing slash for PyPI
	assert.True(t, pipURL[len(pipURL)-1] == '/', "PIP_INDEX_URL should have trailing slash")

	// Should have correct path structure
	assert.Contains(t, pipURL, "/root/pypi/+simple/")
}

// TestEnvironmentInjector_NoProxyList tests NO_PROXY list construction
func TestEnvironmentInjector_NoProxyList(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-squid",
		Type: "squid",
		Port: 3128,
	}

	envVars := injector.InjectForAttach(registry)

	noProxy := envVars["NO_PROXY"]

	// Should include common local addresses
	assert.Contains(t, noProxy, "localhost")
	assert.Contains(t, noProxy, "127.0.0.1")
	assert.Contains(t, noProxy, "::1")

	// Should be comma-separated
	assert.Contains(t, noProxy, ",")
}

// TestEnvironmentInjector_UnknownRegistryType tests handling of unknown types
func TestEnvironmentInjector_UnknownRegistryType(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-custom",
		Type: "custom-type",
		Port: 9000,
	}

	envVars := injector.InjectForAttach(registry)

	// Should return empty or minimal env vars for unknown types
	// Implementation should gracefully handle this
	assert.NotNil(t, envVars)
}

// ---------------------------------------------------------------------------
// Tests for issue #148: host.docker.internal not treated as local host,
// causing PIP_TRUSTED_HOST to be omitted for Docker build-time devpi injection.
// These tests FAIL until isLocalHost() includes "host.docker.internal".
// ---------------------------------------------------------------------------

// TestIsLocalHost_HostDockerInternal verifies that host.docker.internal is
// recognised as a local host so PIP_TRUSTED_HOST is set during Docker builds.
// FAILS until isLocalHost() is updated (issue #148).
func TestIsLocalHost_HostDockerInternal(t *testing.T) {
	// isLocalHost is package-private; exercise it via InjectForAttachWithHost.
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "devpi",
		Type: "devpi",
		Port: 3141,
	}

	envVars := injector.InjectForAttachWithHost(registry, "host.docker.internal")

	// host.docker.internal is the Docker-internal loopback — it must be trusted.
	assert.Contains(t, envVars, "PIP_TRUSTED_HOST",
		"isLocalHost(\"host.docker.internal\") must return true so PIP_TRUSTED_HOST is set")
	assert.Equal(t, "host.docker.internal", envVars["PIP_TRUSTED_HOST"],
		"PIP_TRUSTED_HOST must equal \"host.docker.internal\"")
}

// TestInjectForBuild_DevPI_SetsPIPTrustedHost verifies that InjectForBuild()
// always includes PIP_TRUSTED_HOST=host.docker.internal for devpi registries.
// FAILS until the fix in issue #148 is applied.
func TestInjectForBuild_DevPI_SetsPIPTrustedHost(t *testing.T) {
	injector := NewEnvironmentInjector()

	registry := &models.Registry{
		Name: "my-devpi",
		Type: "devpi",
		Port: 3141,
	}

	envVars := injector.InjectForBuild(registry)

	// PIP_INDEX_URL must point at host.docker.internal (existing behaviour).
	assert.Contains(t, envVars, "PIP_INDEX_URL")
	assert.Equal(t, "http://host.docker.internal:3141/root/pypi/+simple/", envVars["PIP_INDEX_URL"])

	// PIP_TRUSTED_HOST must be set — pip rejects HTTP URLs from untrusted hosts.
	assert.Contains(t, envVars, "PIP_TRUSTED_HOST",
		"InjectForBuild must set PIP_TRUSTED_HOST so pip accepts the HTTP devpi URL")
	assert.Equal(t, "host.docker.internal", envVars["PIP_TRUSTED_HOST"],
		"PIP_TRUSTED_HOST must equal \"host.docker.internal\" for Docker build-time injection")
}

// TestPIPTrustedHostOnlyLocal_IncludesHostDockerInternal extends the existing
// table-driven test to assert host.docker.internal is treated as local.
// FAILS until isLocalHost() is updated (issue #148).
func TestPIPTrustedHostOnlyLocal_IncludesHostDockerInternal(t *testing.T) {
	tests := []struct {
		name            string
		registryHost    string
		shouldHaveTrust bool
	}{
		{
			name:            "localhost should have PIP_TRUSTED_HOST",
			registryHost:    "localhost",
			shouldHaveTrust: true,
		},
		{
			name:            "127.0.0.1 should have PIP_TRUSTED_HOST",
			registryHost:    "127.0.0.1",
			shouldHaveTrust: true,
		},
		{
			name:            "::1 should have PIP_TRUSTED_HOST",
			registryHost:    "::1",
			shouldHaveTrust: true,
		},
		{
			// This case currently FAILS — the bug in issue #148.
			name:            "host.docker.internal should have PIP_TRUSTED_HOST",
			registryHost:    "host.docker.internal",
			shouldHaveTrust: true,
		},
		{
			name:            "external host should NOT have PIP_TRUSTED_HOST",
			registryHost:    "registry.example.com",
			shouldHaveTrust: false,
		},
	}

	injector := NewEnvironmentInjector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := injector.InjectForAttachWithHost(&models.Registry{
				Name: "test-devpi",
				Type: "devpi",
				Port: 3141,
			}, tt.registryHost)

			if tt.shouldHaveTrust {
				assert.Contains(t, envVars, "PIP_TRUSTED_HOST",
					"host %q should be trusted but PIP_TRUSTED_HOST was not set", tt.registryHost)
				assert.Equal(t, tt.registryHost, envVars["PIP_TRUSTED_HOST"])
			} else {
				assert.NotContains(t, envVars, "PIP_TRUSTED_HOST",
					"host %q should NOT be trusted but PIP_TRUSTED_HOST was set", tt.registryHost)
			}
		})
	}
}
