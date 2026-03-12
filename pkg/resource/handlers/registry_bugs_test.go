package handlers

// =============================================================================
// BUG #3 (P1): Registry version never set on creation
//
// Root cause: RegistryHandler.Apply() calls reg.ApplyDefaults() and
// reg.Validate() but never sets reg.Version from the strategy's
// GetDefaultVersion(). A registry created via YAML with no `version:` field
// will have an empty version string forever.
//
// RC-1 Invariant: ApplyDefaults() must NOT set version — that is correct.
// The fix belongs in RegistryHandler.Apply(), after ApplyDefaults(), by
// calling the strategy to get the default version when reg.Version is "".
//
// Expected fix (inside RegistryHandler.Apply()):
//
//   if reg.Version == "" {
//       factory := registry.NewServiceFactory()
//       if strategy, err := factory.GetStrategy(reg.Type); err == nil {
//           if v := strategy.GetDefaultVersion(); v != "" {
//               reg.Version = v
//           }
//       }
//   }
//
// Test strategy:
//   Apply a Registry YAML with no `version:` field for type `zot`.
//   Assert resulting registry has Version == "2.1.15" (ZotStrategy's default).
//
//   RED  today : reg.Version == "" → test FAILS.
//   GREEN after: reg.Version == "2.1.15" → test PASSES.
// =============================================================================

import (
	"testing"

	"devopsmaestro/pkg/resource"
)

// TestRegistryHandler_Apply_SetsDefaultVersionForZot verifies that
// RegistryHandler.Apply() sets the default version from the strategy when
// the YAML omits the version field.
//
// BUG #3 (P1): Apply() currently never sets reg.Version, so a registry
// created without an explicit version field has an empty version string.
//
// RED  today : reg.Version == "" → t.Errorf fires (RED).
// GREEN after: reg.Version == "2.1.15" → assertion passes (GREEN).
func TestRegistryHandler_Apply_SetsDefaultVersionForZot(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{DataStore: ds}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: bug3-zot-no-version
spec:
  type: zot
  port: 5100
  lifecycle: persistent`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() returned unexpected error: %v", err)
	}

	regRes, ok := res.(*RegistryResource)
	if !ok {
		t.Fatalf("Apply() returned wrong resource type: %T", res)
	}

	reg := regRes.Registry()

	// BUG present : reg.Version == ""       → this check fires → RED.
	// BUG fixed   : reg.Version == "2.1.15" → check skipped   → GREEN.
	const wantVersion = "2.1.15"
	if reg.Version != wantVersion {
		t.Errorf("BUG #3: RegistryHandler.Apply() did not set default version for zot registry. "+
			"got Version=%q, want %q. "+
			"Fix: after ApplyDefaults(), call strategy.GetDefaultVersion() when reg.Version is empty.",
			reg.Version, wantVersion)
	}
}

// TestRegistryHandler_Apply_ExplicitVersionNotOverridden ensures that an
// explicit version in the YAML is preserved (don't regress once Bug #3 is fixed).
func TestRegistryHandler_Apply_ExplicitVersionNotOverridden(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{DataStore: ds}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: bug3-zot-explicit-version
spec:
  type: zot
  version: "1.0.0"
  port: 5100
  lifecycle: persistent`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() returned unexpected error: %v", err)
	}

	regRes, ok := res.(*RegistryResource)
	if !ok {
		t.Fatalf("Apply() returned wrong resource type: %T", res)
	}

	reg := regRes.Registry()

	const wantVersion = "1.0.0"
	if reg.Version != wantVersion {
		t.Errorf("Apply() must not override an explicit version: got %q, want %q",
			reg.Version, wantVersion)
	}
}
