package models

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Registry Model Validation Tests
// =============================================================================

func TestRegistry_ValidateType(t *testing.T) {
	tests := []struct {
		name     string
		registry *Registry
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid type: zot",
			registry: &Registry{Name: "my-zot", Type: "zot", Port: 5000},
			wantErr:  false,
		},
		{
			name:     "valid type: athens",
			registry: &Registry{Name: "my-athens", Type: "athens", Port: 3000},
			wantErr:  false,
		},
		{
			name:     "valid type: devpi",
			registry: &Registry{Name: "my-devpi", Type: "devpi", Port: 3141},
			wantErr:  false,
		},
		{
			name:     "valid type: verdaccio",
			registry: &Registry{Name: "my-verdaccio", Type: "verdaccio", Port: 4873},
			wantErr:  false,
		},
		{
			name:     "valid type: squid",
			registry: &Registry{Name: "my-squid", Type: "squid", Port: 3128},
			wantErr:  false,
		},
		{
			name:     "invalid type: unknown",
			registry: &Registry{Name: "my-reg", Type: "unknown", Port: 5000},
			wantErr:  true,
			errMsg:   "unsupported registry type",
		},
		{
			name:     "empty type",
			registry: &Registry{Name: "my-reg", Type: "", Port: 5000},
			wantErr:  true,
			errMsg:   "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.registry.ValidateType()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistry_ValidatePort(t *testing.T) {
	tests := []struct {
		name     string
		registry *Registry
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid port: 1024",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: 1024},
			wantErr:  false,
		},
		{
			name:     "valid port: 5000",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: 5000},
			wantErr:  false,
		},
		{
			name:     "valid port: 65535",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: 65535},
			wantErr:  false,
		},
		{
			name:     "port below range: 1023",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: 1023},
			wantErr:  true,
			errMsg:   "port must be between 1024 and 65535",
		},
		{
			name:     "port above range: 65536",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: 65536},
			wantErr:  true,
			errMsg:   "port must be between 1024 and 65535",
		},
		{
			name:     "port zero (default allowed)",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: 0},
			wantErr:  false, // Port 0 means auto-assign
		},
		{
			name:     "negative port",
			registry: &Registry{Name: "my-reg", Type: "zot", Port: -1},
			wantErr:  true,
			errMsg:   "port must be between 1024 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.registry.ValidatePort()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistry_ValidateLifecycle(t *testing.T) {
	tests := []struct {
		name      string
		lifecycle string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid: persistent",
			lifecycle: "persistent",
			wantErr:   false,
		},
		{
			name:      "valid: on-demand",
			lifecycle: "on-demand",
			wantErr:   false,
		},
		{
			name:      "valid: manual",
			lifecycle: "manual",
			wantErr:   false,
		},
		{
			name:      "empty (should default to manual)",
			lifecycle: "",
			wantErr:   false,
		},
		{
			name:      "invalid: unknown",
			lifecycle: "unknown",
			wantErr:   true,
			errMsg:    "unsupported lifecycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:      "test-reg",
				Type:      "zot",
				Port:      5000,
				Lifecycle: tt.lifecycle,
			}
			err := reg.ValidateLifecycle()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistry_Validate(t *testing.T) {
	tests := []struct {
		name     string
		registry *Registry
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid registry",
			registry: &Registry{
				Name:      "test-reg",
				Type:      "zot",
				Port:      5000,
				Storage:   "/data/registry", // Required field
				Lifecycle: "persistent",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			registry: &Registry{
				Name:      "",
				Type:      "zot",
				Port:      5000,
				Lifecycle: "persistent",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "invalid type",
			registry: &Registry{
				Name:      "test-reg",
				Type:      "invalid",
				Port:      5000,
				Lifecycle: "persistent",
			},
			wantErr: true,
			errMsg:  "unsupported registry type",
		},
		{
			name: "invalid port",
			registry: &Registry{
				Name:      "test-reg",
				Type:      "zot",
				Port:      100,
				Lifecycle: "persistent",
			},
			wantErr: true,
			errMsg:  "port must be between 1024 and 65535",
		},
		{
			name: "invalid lifecycle",
			registry: &Registry{
				Name:      "test-reg",
				Type:      "zot",
				Port:      5000,
				Lifecycle: "always-on",
			},
			wantErr: true,
			errMsg:  "unsupported lifecycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.registry.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// Registry YAML Conversion Tests
// =============================================================================

func TestRegistry_ToYAML(t *testing.T) {
	reg := &Registry{
		Name:        "my-zot",
		Type:        "zot",
		Port:        5000,
		Lifecycle:   "persistent",
		Description: sql.NullString{String: "OCI registry for images", Valid: true},
		Config:      sql.NullString{String: `{"storage":"/data"}`, Valid: true},
	}

	yaml := reg.ToYAML()

	assert.Equal(t, "devopsmaestro.io/v1", yaml.APIVersion)
	assert.Equal(t, "Registry", yaml.Kind)
	assert.Equal(t, "my-zot", yaml.Metadata.Name)
	assert.Equal(t, "OCI registry for images", yaml.Metadata.Description)
	assert.Equal(t, "zot", yaml.Spec.Type)
	assert.Equal(t, 5000, yaml.Spec.Port)
	assert.Equal(t, "persistent", yaml.Spec.Lifecycle)
}

func TestRegistry_FromYAML(t *testing.T) {
	yaml := RegistryYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Registry",
		Metadata: RegistryMetadata{
			Name:        "my-athens",
			Description: "Go module proxy",
		},
		Spec: RegistrySpec{
			Type:      "athens",
			Port:      3000,
			Lifecycle: "on-demand",
			Config: map[string]interface{}{
				"storage": "disk",
			},
		},
	}

	reg := &Registry{}
	reg.FromYAML(yaml)

	assert.Equal(t, "my-athens", reg.Name)
	assert.Equal(t, "athens", reg.Type)
	assert.Equal(t, 3000, reg.Port)
	assert.Equal(t, "on-demand", reg.Lifecycle)
	assert.Equal(t, "Go module proxy", reg.Description.String)
	assert.True(t, reg.Description.Valid)
}

func TestRegistry_RoundTrip_ToYAML_FromYAML(t *testing.T) {
	original := &Registry{
		Name:        "roundtrip-reg",
		Type:        "verdaccio",
		Port:        4873,
		Lifecycle:   "manual",
		Description: sql.NullString{String: "npm registry", Valid: true},
	}

	// Convert to YAML
	yaml := original.ToYAML()

	// Convert back
	restored := &Registry{}
	restored.FromYAML(yaml)

	// Verify
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Port, restored.Port)
	assert.Equal(t, original.Lifecycle, restored.Lifecycle)
	assert.Equal(t, original.Description.String, restored.Description.String)
}

// =============================================================================
// Registry Type-Specific Tests
// =============================================================================

func TestRegistry_ZotDefaults(t *testing.T) {
	reg := &Registry{
		Name: "zot-reg",
		Type: "zot",
	}

	// Should apply defaults for zot
	yaml := reg.ToYAML()
	assert.Equal(t, "zot", yaml.Spec.Type)
	// Port should default to 5000 for zot if not specified
	// Lifecycle should default to "manual" if not specified
}

func TestRegistry_AthensDefaults(t *testing.T) {
	reg := &Registry{
		Name: "athens-reg",
		Type: "athens",
	}

	yaml := reg.ToYAML()
	assert.Equal(t, "athens", yaml.Spec.Type)
	// Port should default to 3000 for athens
}

func TestRegistry_DevpiDefaults(t *testing.T) {
	reg := &Registry{
		Name: "devpi-reg",
		Type: "devpi",
	}

	yaml := reg.ToYAML()
	assert.Equal(t, "devpi", yaml.Spec.Type)
	// Port should default to 3141 for devpi
}

func TestRegistry_VerdaccioDefaults(t *testing.T) {
	reg := &Registry{
		Name: "verdaccio-reg",
		Type: "verdaccio",
	}

	yaml := reg.ToYAML()
	assert.Equal(t, "verdaccio", yaml.Spec.Type)
	// Port should default to 4873 for verdaccio
}

func TestRegistry_SquidDefaults(t *testing.T) {
	reg := &Registry{
		Name: "squid-reg",
		Type: "squid",
	}

	yaml := reg.ToYAML()
	assert.Equal(t, "squid", yaml.Spec.Type)
	// Port should default to 3128 for squid
}

// =============================================================================
// Registry Config Tests
// =============================================================================

func TestRegistry_ConfigJSON(t *testing.T) {
	tests := []struct {
		name       string
		configJSON string
		wantValid  bool
	}{
		{
			name:       "valid JSON config",
			configJSON: `{"storage":"/data","auth":true}`,
			wantValid:  true,
		},
		{
			name:       "empty config",
			configJSON: "",
			wantValid:  true, // Empty is valid (no config)
		},
		{
			name:       "invalid JSON",
			configJSON: `{not valid json}`,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:    "test-reg",
				Type:    "zot",
				Port:    5000,
				Storage: "/data/registry", // Required field
				Config:  sql.NullString{String: tt.configJSON, Valid: tt.configJSON != ""},
			}

			// Validate should check JSON if present
			err := reg.Validate()
			if tt.wantValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid config JSON")
			}
		})
	}
}

// =============================================================================
// Registry New Fields Tests (TDD Phase 2 - RED)
// =============================================================================

func TestRegistry_NewFields(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		storage     string
		idleTimeout int
		wantEnabled bool
		wantStorage string
		wantIdle    int
	}{
		{
			name:        "all fields set to non-zero values",
			enabled:     true,
			storage:     "/custom/storage",
			idleTimeout: 3600,
			wantEnabled: true,
			wantStorage: "/custom/storage",
			wantIdle:    3600,
		},
		{
			name:        "enabled false explicitly",
			enabled:     false,
			storage:     "/data",
			idleTimeout: 1800,
			wantEnabled: false,
			wantStorage: "/data",
			wantIdle:    1800,
		},
		{
			name:        "zero values",
			enabled:     false,
			storage:     "",
			idleTimeout: 0,
			wantEnabled: false,
			wantStorage: "",
			wantIdle:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:        "test-reg",
				Type:        "zot",
				Port:        5000,
				Enabled:     tt.enabled,
				Storage:     tt.storage,
				IdleTimeout: tt.idleTimeout,
			}

			assert.Equal(t, tt.wantEnabled, reg.Enabled)
			assert.Equal(t, tt.wantStorage, reg.Storage)
			assert.Equal(t, tt.wantIdle, reg.IdleTimeout)
		})
	}
}

func TestRegistry_IsOnDemand(t *testing.T) {
	tests := []struct {
		name      string
		lifecycle string
		want      bool
	}{
		{
			name:      "on-demand lifecycle",
			lifecycle: "on-demand",
			want:      true,
		},
		{
			name:      "persistent lifecycle",
			lifecycle: "persistent",
			want:      false,
		},
		{
			name:      "manual lifecycle",
			lifecycle: "manual",
			want:      false,
		},
		{
			name:      "empty lifecycle",
			lifecycle: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:      "test-reg",
				Type:      "zot",
				Lifecycle: tt.lifecycle,
			}

			got := reg.IsOnDemand()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistry_ShouldStopAfterIdle(t *testing.T) {
	tests := []struct {
		name        string
		lifecycle   string
		idleTimeout int
		want        bool
	}{
		{
			name:        "on-demand with positive timeout",
			lifecycle:   "on-demand",
			idleTimeout: 1800,
			want:        true,
		},
		{
			name:        "on-demand with zero timeout",
			lifecycle:   "on-demand",
			idleTimeout: 0,
			want:        false,
		},
		{
			name:        "persistent with positive timeout",
			lifecycle:   "persistent",
			idleTimeout: 1800,
			want:        false,
		},
		{
			name:        "manual with positive timeout",
			lifecycle:   "manual",
			idleTimeout: 1800,
			want:        false,
		},
		{
			name:        "on-demand with negative timeout",
			lifecycle:   "on-demand",
			idleTimeout: -1,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:        "test-reg",
				Type:        "zot",
				Lifecycle:   tt.lifecycle,
				IdleTimeout: tt.idleTimeout,
			}

			got := reg.ShouldStopAfterIdle()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistry_GetIdleTimeoutDuration(t *testing.T) {
	tests := []struct {
		name        string
		idleTimeout int
		wantSeconds int
	}{
		{
			name:        "zero timeout",
			idleTimeout: 0,
			wantSeconds: 0,
		},
		{
			name:        "60 seconds",
			idleTimeout: 60,
			wantSeconds: 60,
		},
		{
			name:        "1800 seconds (30 minutes)",
			idleTimeout: 1800,
			wantSeconds: 1800,
		},
		{
			name:        "3600 seconds (1 hour)",
			idleTimeout: 3600,
			wantSeconds: 3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:        "test-reg",
				Type:        "zot",
				IdleTimeout: tt.idleTimeout,
			}

			got := reg.GetIdleTimeoutDuration()
			assert.Equal(t, tt.wantSeconds, int(got.Seconds()))
		})
	}
}

func TestRegistry_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name            string
		registry        Registry
		wantPort        int
		wantStorage     string
		wantIdleTimeout int
	}{
		{
			name: "zot with zero port",
			registry: Registry{
				Name: "zot-reg",
				Type: "zot",
				Port: 0,
			},
			wantPort:        5001,
			wantStorage:     "/var/lib/zot",
			wantIdleTimeout: 0, // Not on-demand, no default timeout
		},
		{
			name: "athens with explicit port",
			registry: Registry{
				Name: "athens-reg",
				Type: "athens",
				Port: 9000,
			},
			wantPort:        9000, // Should keep explicit port
			wantStorage:     "/var/lib/athens",
			wantIdleTimeout: 0,
		},
		{
			name: "devpi with empty storage",
			registry: Registry{
				Name:    "devpi-reg",
				Type:    "devpi",
				Port:    3141,
				Storage: "",
			},
			wantPort:        3141,
			wantStorage:     "/var/lib/devpi",
			wantIdleTimeout: 0,
		},
		{
			name: "verdaccio with explicit storage",
			registry: Registry{
				Name:    "verdaccio-reg",
				Type:    "verdaccio",
				Port:    4873,
				Storage: "/custom/storage",
			},
			wantPort:        4873,
			wantStorage:     "/custom/storage", // Should keep explicit storage
			wantIdleTimeout: 0,
		},
		{
			name: "on-demand zot with zero timeout",
			registry: Registry{
				Name:        "zot-ondemand",
				Type:        "zot",
				Port:        0,
				Lifecycle:   "on-demand",
				IdleTimeout: 0,
			},
			wantPort:        5001,
			wantStorage:     "/var/lib/zot",
			wantIdleTimeout: 1800, // Should set default timeout for on-demand
		},
		{
			name: "on-demand athens with explicit timeout",
			registry: Registry{
				Name:        "athens-ondemand",
				Type:        "athens",
				Port:        3000,
				Lifecycle:   "on-demand",
				IdleTimeout: 3600,
			},
			wantPort:        3000,
			wantStorage:     "/var/lib/athens",
			wantIdleTimeout: 3600, // Should keep explicit timeout
		},
		{
			name: "persistent registry should not set timeout",
			registry: Registry{
				Name:        "zot-persistent",
				Type:        "zot",
				Port:        5000,
				Lifecycle:   "persistent",
				IdleTimeout: 0,
			},
			wantPort:        5000,
			wantStorage:     "/var/lib/zot",
			wantIdleTimeout: 0, // Should NOT set timeout for persistent
		},
		{
			name: "squid with all defaults",
			registry: Registry{
				Name:      "squid-reg",
				Type:      "squid",
				Port:      0,
				Storage:   "",
				Lifecycle: "manual",
			},
			wantPort:        3128,
			wantStorage:     "/var/cache/squid",
			wantIdleTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := tt.registry
			reg.ApplyDefaults()

			assert.Equal(t, tt.wantPort, reg.Port, "Port mismatch")
			assert.Equal(t, tt.wantStorage, reg.Storage, "Storage mismatch")
			assert.Equal(t, tt.wantIdleTimeout, reg.IdleTimeout, "IdleTimeout mismatch")
		})
	}
}

// =============================================================================
// Zot Port Fix Tests (TDD Phase 2 - RED)
// Zot default port must be 5001 (not 5000) to avoid macOS AirPlay conflict.
// See: pkg/registry/strategy.go for the correct value.
// These tests FAIL until defaultPorts["zot"] is corrected to 5001.
// =============================================================================

func TestDefaultPorts_ZotIs5001(t *testing.T) {
	r := &Registry{Type: "zot"}
	assert.Equal(t, 5001, r.GetDefaultPort(), "Zot default port should be 5001 to avoid macOS AirPlay conflict on 5000")
}

func TestApplyDefaults_ZotPort5001(t *testing.T) {
	r := &Registry{
		Name:    "test-zot",
		Type:    "zot",
		Port:    0,
		Storage: "/tmp/test-zot",
	}
	r.ApplyDefaults()
	assert.Equal(t, 5001, r.Port, "Zot registry ApplyDefaults() should set port to 5001")
}

// =============================================================================
// TDD Phase 2 (RED): Bug 7 — RegistryStatusYAML in get registries JSON/YAML
// =============================================================================
// RegistryYAML currently has NO Status field. The `dvm get registries` output
// omits live process state (running/stopped, endpoint).
//
// These tests FAIL until:
//   1. RegistryStatusYAML struct is added to models/registry.go
//   2. Status *RegistryStatusYAML field is added to RegistryYAML (with omitempty)
// =============================================================================

// TestRegistryYAML_IncludesStatusSection verifies that when RegistryYAML.Status
// is populated, JSON serialization produces a "status" block with "state" and
// "endpoint" fields.
func TestRegistryYAML_IncludesStatusSection(t *testing.T) {
	regYAML := RegistryYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Registry",
		Metadata:   RegistryMetadata{Name: "my-zot"},
		Spec:       RegistrySpec{Type: "zot", Port: 5001},
		Status: &RegistryStatusYAML{
			State:    "running",
			Endpoint: "localhost:5001",
		},
	}

	data, err := json.Marshal(regYAML)
	assert.NoError(t, err, "JSON marshal should succeed")

	jsonStr := string(data)

	// The "status" key must be present
	assert.Contains(t, jsonStr, `"status"`, "JSON should include 'status' section (Bug 7: Status field missing)")

	// The "state" and "endpoint" fields must be present inside status
	assert.Contains(t, jsonStr, `"state"`, "status section should have 'state' field")
	assert.Contains(t, jsonStr, `"running"`, "state should be 'running'")
	assert.Contains(t, jsonStr, `"endpoint"`, "status section should have 'endpoint' field")
	assert.Contains(t, jsonStr, `"localhost:5001"`, "endpoint should be 'localhost:5001'")
}

// TestRegistryYAML_StatusOmittedWhenNil verifies that when Status is nil,
// the JSON output does NOT include a "status" key (omitempty behavior).
func TestRegistryYAML_StatusOmittedWhenNil(t *testing.T) {
	regYAML := RegistryYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Registry",
		Metadata:   RegistryMetadata{Name: "my-zot"},
		Spec:       RegistrySpec{Type: "zot", Port: 5001},
		// Status intentionally nil
	}

	data, err := json.Marshal(regYAML)
	assert.NoError(t, err, "JSON marshal should succeed")

	jsonStr := string(data)

	// When nil, "status" must NOT appear in JSON (omitempty)
	assert.NotContains(t, jsonStr, `"status"`, "JSON should NOT include 'status' when Status is nil (omitempty)")
}

// =============================================================================
// TDD Phase 2 (RED): Declarative Registry Version (v0.35.1)
//
// These tests WILL NOT COMPILE until Phase 3 adds:
//   - Registry.Version field (models/registry.go)
//   - RegistrySpec.Version field (models/registry.go)
//   - Registry.ValidateVersion() method (models/registry.go)
//
// WHY THEY FAIL:
//   - Registry struct has no Version field → compiler error
//   - RegistrySpec struct has no Version field → compiler error
//   - Registry has no ValidateVersion() method → compiler error
// =============================================================================

// TestRegistry_VersionField verifies that the Registry model has a Version field
// that can be set and read back.
// FAILS TO COMPILE: Registry struct has no Version field.
func TestRegistry_VersionField(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "empty version (use strategy default)",
			version: "",
			want:    "",
		},
		{
			name:    "explicit version 2.1.15",
			version: "2.1.15",
			want:    "2.1.15",
		},
		{
			name:    "semver 1.0.0",
			version: "1.0.0",
			want:    "1.0.0",
		},
		{
			name:    "pre-release version 2.0.0-rc1",
			version: "2.0.0-rc1",
			want:    "2.0.0-rc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:    "test-registry",
				Type:    "zot",
				Port:    5001,
				Version: tt.version, // COMPILE ERROR: Registry has no Version field
			}
			assert.Equal(t, tt.want, reg.Version,
				"Registry.Version should be settable and readable")
		})
	}
}

// TestRegistry_ToYAML_IncludesVersion verifies that ToYAML() propagates the
// Registry.Version field into RegistrySpec.Version in the output YAML.
// FAILS TO COMPILE: Registry.Version and RegistrySpec.Version don't exist.
func TestRegistry_ToYAML_IncludesVersion(t *testing.T) {
	reg := &Registry{
		Name:      "my-zot",
		Type:      "zot",
		Port:      5001,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
		Version:   "2.1.15", // COMPILE ERROR: Registry has no Version field
	}

	yamlOut := reg.ToYAML()

	// RegistrySpec must carry the version forward
	assert.Equal(t, "2.1.15", yamlOut.Spec.Version, // COMPILE ERROR: RegistrySpec has no Version field
		"ToYAML() must include Version in spec when set on Registry")
}

// TestRegistry_FromYAML_SetsVersion verifies that FromYAML() reads the version
// from RegistrySpec.Version and sets it on Registry.Version.
// FAILS TO COMPILE: Registry.Version and RegistrySpec.Version don't exist.
func TestRegistry_FromYAML_SetsVersion(t *testing.T) {
	yamlIn := RegistryYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Registry",
		Metadata:   RegistryMetadata{Name: "my-zot"},
		Spec: RegistrySpec{
			Type:      "zot",
			Port:      5001,
			Lifecycle: "persistent",
			Version:   "2.1.15", // COMPILE ERROR: RegistrySpec has no Version field
		},
	}

	reg := &Registry{}
	reg.FromYAML(yamlIn)

	assert.Equal(t, "2.1.15", reg.Version, // COMPILE ERROR: Registry has no Version field
		"FromYAML() must read Version from RegistrySpec.Version and set it on Registry")
}

// TestRegistry_RoundTrip_Version verifies that Version survives a full
// ToYAML → FromYAML round-trip without data loss or mutation.
// FAILS TO COMPILE: Registry.Version and RegistrySpec.Version don't exist.
func TestRegistry_RoundTrip_Version(t *testing.T) {
	original := &Registry{
		Name:      "my-zot",
		Type:      "zot",
		Port:      5001,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
		Version:   "2.1.15", // COMPILE ERROR: Registry has no Version field
	}

	// Convert to YAML and back
	intermediate := original.ToYAML()
	restored := &Registry{}
	restored.FromYAML(intermediate)

	// Version must survive the round-trip intact
	assert.Equal(t, original.Version, restored.Version, // COMPILE ERROR: Registry has no Version field
		"Version must survive ToYAML→FromYAML round-trip unchanged: got %q, want %q",
		restored.Version, original.Version)
}

// TestRegistry_ValidateVersion verifies the light semver validation on Registry.Version.
// Per RC-1: empty string is valid (means "use strategy default").
// Leading 'v' is NOT allowed (consistent with Go module conventions).
// FAILS TO COMPILE: Registry.ValidateVersion() method doesn't exist.
func TestRegistry_ValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
		errMsg  string
	}{
		// Valid versions
		{
			name:    "empty string is valid (use strategy default)",
			version: "",
			wantErr: false,
		},
		{
			name:    "canonical version 2.1.15",
			version: "2.1.15",
			wantErr: false,
		},
		{
			name:    "canonical version 1.0.0",
			version: "1.0.0",
			wantErr: false,
		},
		{
			name:    "patch 0.1.0",
			version: "0.1.0",
			wantErr: false,
		},
		{
			name:    "pre-release 2.0.0-rc1",
			version: "2.0.0-rc1",
			wantErr: false,
		},
		{
			name:    "pre-release with dot 1.2.3-beta.1",
			version: "1.2.3-beta.1",
			wantErr: false,
		},
		// Invalid versions
		{
			name:    "non-numeric string abc",
			version: "abc",
			wantErr: true,
			errMsg:  "invalid version",
		},
		{
			name:    "leading v prefix v2.1.15",
			version: "v2.1.15",
			wantErr: true,
			errMsg:  "invalid version",
		},
		{
			name:    "incomplete semver 2.1 (missing patch)",
			version: "2.1",
			wantErr: true,
			errMsg:  "invalid version",
		},
		{
			name:    "major only 2",
			version: "2",
			wantErr: true,
			errMsg:  "invalid version",
		},
		{
			name:    "leading dot .1.2.3",
			version: ".1.2.3",
			wantErr: true,
			errMsg:  "invalid version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &Registry{
				Name:    "test-registry",
				Type:    "zot",
				Port:    5001,
				Version: tt.version, // COMPILE ERROR: Registry has no Version field
			}
			err := reg.ValidateVersion() // COMPILE ERROR: ValidateVersion() doesn't exist
			if tt.wantErr {
				assert.Error(t, err,
					"ValidateVersion() should return error for version %q", tt.version)
				assert.Contains(t, err.Error(), tt.errMsg,
					"error message should contain %q for version %q", tt.errMsg, tt.version)
			} else {
				assert.NoError(t, err,
					"ValidateVersion() should accept version %q as valid", tt.version)
			}
		})
	}
}

// TestRegistry_ApplyDefaults_Version verifies that ApplyDefaults() does NOT set
// a Version on the Registry model.
// Per RC-1: version defaulting is the strategy layer's responsibility.
// ApplyDefaults() must NOT touch Version — that would violate the architecture.
// FAILS TO COMPILE: Registry.Version field doesn't exist.
func TestRegistry_ApplyDefaults_Version(t *testing.T) {
	reg := &Registry{
		Name:      "my-zot",
		Type:      "zot",
		Lifecycle: "persistent",
		Storage:   "",
		// Version intentionally left empty — ApplyDefaults must leave it alone
	}

	reg.ApplyDefaults()

	// ApplyDefaults must NOT set a version — that is the strategy layer's job (RC-1).
	assert.Equal(t, "", reg.Version, // COMPILE ERROR: Registry has no Version field
		"ApplyDefaults() must NOT set a default Version — version defaulting belongs to the strategy layer (RC-1)")
}
