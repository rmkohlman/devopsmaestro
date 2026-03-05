package models

import (
	"database/sql"
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
