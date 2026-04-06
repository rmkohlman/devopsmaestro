package handlers

// =============================================================================
// Registry Fields Round-Trip Tests — Issue #178
//
// These tests prove three data loss bugs in registry YAML round-trip:
//
//  1. Enabled field not exported — RegistrySpec lacks Enabled field.
//     After export → restore, all registries become disabled (Go bool zero
//     value = false).
//
//  2. Storage field not exported — custom storage paths are silently lost.
//
//  3. IdleTimeout field not exported — custom idle timeouts are silently lost.
//
// Additional edge case: FromYAML() should default Enabled=true when the field
// is omitted from YAML (matching DB default behavior).
//
// All tests MUST compile but MUST fail (red phase of TDD) until the fix lands.
// =============================================================================

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// TestRegistryYAMLExport_IncludesEnabledField
//
// Arrange: Create a Registry model with Enabled=true.
//
// Act:     Call RegistryHandler.ToYAML().
//
// Assert:  The resulting YAML contains "enabled: true".
//
// WHY IT FAILS: RegistrySpec has no Enabled field. Registry.ToYAML() never
//
//	copies r.Enabled into the spec. The YAML output will never contain an
//	enabled field — so after restore all registries silently become disabled.
//
// =============================================================================
func TestRegistryYAMLExport_IncludesEnabledField(t *testing.T) {
	h := NewRegistryHandler()

	reg := &models.Registry{
		ID:        1,
		Name:      "my-zot",
		Type:      "zot",
		Version:   "2.1.15",
		Enabled:   true,
		Port:      5001,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
		Status:    "stopped",
	}

	res := NewRegistryResource(reg)

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// MUST FAIL: RegistrySpec has no Enabled field so ToYAML() never writes it.
	// After fix: "enabled: true" must appear in the exported YAML so that
	// export → restore preserves the enabled state.
	if !strings.Contains(yamlStr, "enabled:") {
		t.Errorf("ToYAML() output is missing 'enabled' field — Enabled=true registry will become disabled after export/restore.\n"+
			"Root cause: RegistrySpec has no Enabled field; Registry.ToYAML() never copies r.Enabled.\n"+
			"Got YAML:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "enabled: true") {
		t.Errorf("ToYAML() output does not contain 'enabled: true' — registry was created with Enabled=true.\n"+
			"Got YAML:\n%s", yamlStr)
	}
}

// =============================================================================
// TestRegistryYAMLExport_IncludesStorageField
//
// Arrange: Create a Registry model with a custom Storage path.
//
// Act:     Call RegistryHandler.ToYAML().
//
// Assert:  The resulting YAML contains a "storage:" field with the custom path.
//
// WHY IT FAILS: RegistrySpec has no Storage field. Registry.ToYAML() never
//
//	copies r.Storage into the spec. Custom storage paths are silently dropped.
//
// =============================================================================
func TestRegistryYAMLExport_IncludesStorageField(t *testing.T) {
	h := NewRegistryHandler()

	reg := &models.Registry{
		ID:        2,
		Name:      "my-devpi",
		Type:      "devpi",
		Version:   "6.8.0",
		Enabled:   true,
		Port:      3141,
		Lifecycle: "persistent",
		Storage:   "/data/custom/devpi",
		Status:    "stopped",
	}

	res := NewRegistryResource(reg)

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// MUST FAIL: RegistrySpec has no Storage field so ToYAML() never writes it.
	// After fix: "storage:" must appear in the exported YAML with the custom path.
	if !strings.Contains(yamlStr, "storage:") {
		t.Errorf("ToYAML() output is missing 'storage' field — custom storage path '/data/custom/devpi' will be lost on export/restore.\n"+
			"Root cause: RegistrySpec has no Storage field; Registry.ToYAML() never copies r.Storage.\n"+
			"Got YAML:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "/data/custom/devpi") {
		t.Errorf("ToYAML() output does not contain the custom storage path '/data/custom/devpi'.\n"+
			"Got YAML:\n%s", yamlStr)
	}
}

// =============================================================================
// TestRegistryYAMLExport_IncludesIdleTimeoutField
//
// Arrange: Create an on-demand Registry with a custom IdleTimeout.
//
// Act:     Call RegistryHandler.ToYAML().
//
// Assert:  The resulting YAML contains an "idleTimeout:" field.
//
// WHY IT FAILS: RegistrySpec has no IdleTimeout field. Registry.ToYAML()
//
//	never copies r.IdleTimeout into the spec. Custom idle timeouts are lost.
//
// =============================================================================
func TestRegistryYAMLExport_IncludesIdleTimeoutField(t *testing.T) {
	h := NewRegistryHandler()

	reg := &models.Registry{
		ID:          3,
		Name:        "on-demand-athens",
		Type:        "athens",
		Version:     "0.13.1",
		Enabled:     true,
		Port:        3000,
		Lifecycle:   "on-demand",
		Storage:     "/var/lib/athens",
		IdleTimeout: 300,
		Status:      "stopped",
	}

	res := NewRegistryResource(reg)

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// MUST FAIL: RegistrySpec has no IdleTimeout field so ToYAML() never writes it.
	// After fix: "idleTimeout:" must appear in the exported YAML.
	if !strings.Contains(yamlStr, "idleTimeout:") {
		t.Errorf("ToYAML() output is missing 'idleTimeout' field — custom idle timeout of 300s will be lost on export/restore.\n"+
			"Root cause: RegistrySpec has no IdleTimeout field; Registry.ToYAML() never copies r.IdleTimeout.\n"+
			"Got YAML:\n%s", yamlStr)
	}
}

// =============================================================================
// TestRegistryApply_PreservesEnabled
//
// Arrange: Construct a RegistryYAML with Enabled=true (spec.enabled: true).
//
// Act:     Call RegistryHandler.Apply().
//
// Assert:  The resulting Registry DB record has Enabled=true.
//
// WHY IT FAILS: RegistrySpec has no Enabled field. yaml.Unmarshal silently
//
//	discards spec.enabled. Registry.FromYAML() never reads it. The resulting
//	Registry model has Enabled=false (Go zero value).
//
// =============================================================================
func TestRegistryApply_PreservesEnabled(t *testing.T) {
	h := NewRegistryHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: enabled-zot
spec:
  type: zot
  version: 2.1.15
  enabled: true
  port: 5001
  lifecycle: persistent
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() unexpected error = %v", err)
	}

	rr, ok := res.(*RegistryResource)
	if !ok {
		t.Fatalf("Apply() result is not *RegistryResource, got %T", res)
	}

	// MUST FAIL: RegistrySpec has no Enabled field, so yaml.Unmarshal discards
	// spec.enabled. Registry.FromYAML() never reads it. Enabled stays false.
	if !rr.Registry().Enabled {
		t.Errorf("Apply(): Registry.Enabled = false after applying YAML with 'enabled: true'.\n"+
			"Root cause: RegistrySpec has no Enabled field; spec.enabled is silently discarded by YAML parser.\n"+
			"Registry model: %+v", rr.Registry())
	}

	// Also verify the value persisted in the store
	stored, err := store.GetRegistryByName("enabled-zot")
	if err != nil {
		t.Fatalf("GetRegistryByName() after Apply() error = %v", err)
	}
	if !stored.Enabled {
		t.Errorf("stored Registry.Enabled = false — enabled state was not persisted to the DB.\n"+
			"Stored registry: %+v", stored)
	}
}

// =============================================================================
// TestRegistryApply_DefaultsEnabledTrue
//
// Arrange: Construct YAML that omits the enabled field entirely.
//
// Act:     Call RegistryHandler.Apply().
//
// Assert:  The resulting Registry has Enabled=true (matching DB default).
//
// WHY IT FAILS: When Enabled field is added to RegistrySpec, existing YAML
//
//	that omits it will decode to Enabled=false (Go zero value). FromYAML()
//	must explicitly default Enabled=true to match the DB default behavior.
//	Without this fix, restoring old YAML would silently disable all registries.
//
// =============================================================================
func TestRegistryApply_DefaultsEnabledTrue(t *testing.T) {
	h := NewRegistryHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// YAML with no "enabled:" field — simulates legacy/existing export files
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: legacy-verdaccio
spec:
  type: verdaccio
  version: 5.26.0
  port: 4873
  lifecycle: persistent
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() unexpected error = %v", err)
	}

	rr, ok := res.(*RegistryResource)
	if !ok {
		t.Fatalf("Apply() result is not *RegistryResource, got %T", res)
	}

	// MUST FAIL: Go bool zero value is false. Without explicit defaulting in
	// FromYAML(), a registry with no "enabled:" field becomes disabled.
	// The DB defaults to enabled=true, so FromYAML() must match.
	if !rr.Registry().Enabled {
		t.Errorf("Apply(): Registry.Enabled = false when 'enabled' field is omitted from YAML.\n"+
			"Expected: Enabled=true (matching DB default of enabled=true).\n"+
			"Fix: FromYAML() must default Enabled=true when the field is absent.\n"+
			"Registry model: %+v", rr.Registry())
	}
}

// =============================================================================
// TestRegistryRoundTrip_AllFieldsPreserved
//
// Full round-trip: create a registry with Enabled=true, custom Storage, and
// custom IdleTimeout → export to YAML → parse → apply → verify all 3 fields
// are preserved through the cycle.
//
// WHY IT FAILS: All three bugs contribute:
//
//  1. ToYAML() never writes Enabled/Storage/IdleTimeout (RegistrySpec lacks them)
//  2. FromYAML() never reads them back (same reason)
//  3. After apply, Enabled=false, Storage=default, IdleTimeout=0
//
// =============================================================================
func TestRegistryRoundTrip_AllFieldsPreserved(t *testing.T) {
	tests := []struct {
		name        string
		regName     string
		regType     string
		version     string
		port        int
		lifecycle   string
		storage     string
		idleTimeout int
		enabled     bool
	}{
		{
			name:        "on-demand registry with all custom fields",
			regName:     "rt-athens",
			regType:     "athens",
			version:     "0.13.1",
			port:        3000,
			lifecycle:   "on-demand",
			storage:     "/data/custom/athens",
			idleTimeout: 600,
			enabled:     true,
		},
		{
			name:        "persistent registry with custom storage",
			regName:     "rt-zot",
			regType:     "zot",
			version:     "2.1.15",
			port:        5001,
			lifecycle:   "persistent",
			storage:     "/mnt/fast-ssd/zot",
			idleTimeout: 0,
			enabled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewRegistryHandler()

			// Step 1: Build registry model with all fields set
			reg := &models.Registry{
				ID:          1,
				Name:        tt.regName,
				Type:        tt.regType,
				Version:     tt.version,
				Enabled:     tt.enabled,
				Port:        tt.port,
				Lifecycle:   tt.lifecycle,
				Storage:     tt.storage,
				IdleTimeout: tt.idleTimeout,
				Status:      "stopped",
			}

			// Step 2: Export to YAML
			res := NewRegistryResource(reg)
			yamlBytes, err := h.ToYAML(res)
			if err != nil {
				t.Fatalf("ToYAML() unexpected error = %v", err)
			}
			yamlStr := string(yamlBytes)

			// MUST FAIL: All three fields missing from export
			if !strings.Contains(yamlStr, "enabled:") {
				t.Errorf("[%s] ToYAML() missing 'enabled' field in export.\nYAML:\n%s", tt.name, yamlStr)
			}
			if !strings.Contains(yamlStr, "storage:") {
				t.Errorf("[%s] ToYAML() missing 'storage' field in export.\nYAML:\n%s", tt.name, yamlStr)
			}
			if tt.idleTimeout > 0 && !strings.Contains(yamlStr, "idleTimeout:") {
				t.Errorf("[%s] ToYAML() missing 'idleTimeout' field in export.\nYAML:\n%s", tt.name, yamlStr)
			}

			// Step 3: Parse the exported YAML to verify the struct fields are present
			var parsed models.RegistryYAML
			if err := yaml.Unmarshal(yamlBytes, &parsed); err != nil {
				t.Fatalf("[%s] yaml.Unmarshal of exported YAML error = %v", tt.name, err)
			}

			// Step 4: Apply the exported YAML to a fresh store (simulates restore)
			freshStore := db.NewMockDataStore()
			restoreCtx := resource.Context{DataStore: freshStore}

			applyRes, err := h.Apply(restoreCtx, yamlBytes)
			if err != nil {
				t.Fatalf("[%s] Apply() (restore) error = %v\nThis likely means a missing field caused validation to fail.\nExported YAML:\n%s",
					tt.name, err, yamlStr)
			}

			rr, ok := applyRes.(*RegistryResource)
			if !ok {
				t.Fatalf("[%s] Apply() result is not *RegistryResource, got %T", tt.name, applyRes)
			}

			restored := rr.Registry()

			// MUST FAIL: Enabled is dropped → becomes false
			if restored.Enabled != tt.enabled {
				t.Errorf("[%s] Registry.Enabled = %v after round-trip, want %v.\n"+
					"The enabled state was lost during export/restore.\n"+
					"Restored registry: %+v",
					tt.name, restored.Enabled, tt.enabled, restored)
			}

			// MUST FAIL: Storage is dropped → falls back to ApplyDefaults() value
			if tt.storage != "" && restored.Storage != tt.storage {
				t.Errorf("[%s] Registry.Storage = %q after round-trip, want %q.\n"+
					"The custom storage path was lost during export/restore.\n"+
					"Restored registry: %+v",
					tt.name, restored.Storage, tt.storage, restored)
			}

			// MUST FAIL: IdleTimeout is dropped → becomes 0 (or ApplyDefaults value)
			if tt.idleTimeout > 0 && restored.IdleTimeout != tt.idleTimeout {
				t.Errorf("[%s] Registry.IdleTimeout = %d after round-trip, want %d.\n"+
					"The custom idle timeout was lost during export/restore.\n"+
					"Restored registry: %+v",
					tt.name, restored.IdleTimeout, tt.idleTimeout, restored)
			}
		})
	}
}
