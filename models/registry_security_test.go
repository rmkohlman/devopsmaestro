package models

// =============================================================================
// Security Hardening Tests — Item 6 (TDD Phase 2 — RED)
//
// Minimum idle timeout validation for on-demand registries.
//
// BUG: Registry.Validate() does not enforce a minimum IdleTimeout for
// on-demand registries. A value of 1 second is silently accepted, meaning
// an on-demand registry could stop itself almost immediately after starting,
// causing flapping and confusing users.
//
// Fix: Validate() must reject on-demand registries with IdleTimeout between
// 1-59 seconds (inclusive). The sentinel value 0 means "use default" and must
// continue to be accepted. Values >= 60 seconds are valid.
//
// RED  today : Validate() has no IdleTimeout check → tests below FAIL.
// GREEN after: Validate() rejects 1-59, accepts 0 and >=60 → tests PASS.
// =============================================================================

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalValidRegistry returns a Registry that passes all current Validate()
// checks so we can isolate IdleTimeout as the single variable.
func minimalValidOnDemandRegistry() *Registry {
	return &Registry{
		Name:      "test-reg",
		Type:      "zot",
		Port:      5001,
		Lifecycle: "on-demand",
		Storage:   "/var/lib/zot",
		// IdleTimeout left at 0 (default sentinel) — will be set per test.
	}
}

// TestRegistry_Validate_RejectsSubMinimumIdleTimeout verifies that Validate()
// returns an error when an on-demand registry specifies an IdleTimeout between
// 1 and 59 seconds (too short to be useful and likely a misconfiguration).
//
// RED  today : Validate() ignores IdleTimeout → no error returned → FAILS.
// GREEN after: Validate() returns error for 1-59 → PASSES.
func TestRegistry_Validate_RejectsSubMinimumIdleTimeout(t *testing.T) {
	subMinimumValues := []int{1, 5, 10, 30, 59}

	for _, timeout := range subMinimumValues {
		t.Run("idle_timeout_"+itoa(timeout)+"s", func(t *testing.T) {
			reg := minimalValidOnDemandRegistry()
			reg.IdleTimeout = timeout

			err := reg.Validate()

			// BUG present : err == nil → assertion FAILS (RED).
			// BUG fixed   : err != nil with IdleTimeout mention → PASSES (GREEN).
			require.Error(t, err,
				"BUG ITEM 6: Validate() must reject on-demand registry with "+
					"IdleTimeout=%d seconds (below minimum of 60 seconds). "+
					"Short timeouts cause flapping and confusing behaviour.", timeout)
			assert.Contains(t, err.Error(), "IdleTimeout",
				"error must mention IdleTimeout so the cause is obvious; got: %v", err)
		})
	}
}

// TestRegistry_Validate_AcceptsZeroIdleTimeout verifies that IdleTimeout=0 is
// accepted as a sentinel value meaning "use ApplyDefaults() to set the default
// 30-minute timeout". This must NOT be rejected.
//
// RED/GREEN: This test is expected to PASS today (0 is not validated) and
// must continue to pass after the fix. It guards against over-eager validation.
func TestRegistry_Validate_AcceptsZeroIdleTimeout(t *testing.T) {
	reg := minimalValidOnDemandRegistry()
	reg.IdleTimeout = 0 // sentinel: use default

	err := reg.Validate()

	assert.NoError(t, err,
		"IdleTimeout=0 must be accepted as the 'use default' sentinel value; "+
			"ApplyDefaults() will set it to 1800 seconds.")
}

// TestRegistry_Validate_AcceptsMinimumIdleTimeout verifies that the boundary
// value of exactly 60 seconds is accepted.
//
// RED  today : 60s is accepted today (no validation), so this PASSES already.
//
//	After the fix it must still PASS.
//
// This test protects against off-by-one errors in the fix implementation.
func TestRegistry_Validate_AcceptsMinimumIdleTimeout(t *testing.T) {
	reg := minimalValidOnDemandRegistry()
	reg.IdleTimeout = 60 // minimum valid value

	err := reg.Validate()

	assert.NoError(t, err,
		"IdleTimeout=60 is the minimum valid value and must be accepted.")
}

// TestRegistry_Validate_AcceptsDefaultIdleTimeout verifies that the default
// idle timeout of 1800 seconds (30 minutes) continues to be accepted.
func TestRegistry_Validate_AcceptsDefaultIdleTimeout(t *testing.T) {
	reg := minimalValidOnDemandRegistry()
	reg.IdleTimeout = 1800 // default set by ApplyDefaults()

	err := reg.Validate()

	assert.NoError(t, err,
		"IdleTimeout=1800 (the default) must always be accepted.")
}

// TestRegistry_Validate_IdleTimeoutIgnoredForNonOnDemand verifies that the
// minimum idle timeout check ONLY applies to on-demand registries. A
// persistent or manual registry with a short IdleTimeout is nonsensical but
// not a security concern — Validate() must not break those cases.
func TestRegistry_Validate_IdleTimeoutIgnoredForNonOnDemand(t *testing.T) {
	lifecycles := []string{"persistent", "manual"}

	for _, lc := range lifecycles {
		t.Run(lc, func(t *testing.T) {
			reg := &Registry{
				Name:        "test-reg",
				Type:        "zot",
				Port:        5001,
				Lifecycle:   lc,
				Storage:     "/var/lib/zot",
				IdleTimeout: 1, // would be rejected for on-demand
			}

			err := reg.Validate()

			// IdleTimeout minimum must NOT apply to non-on-demand lifecycles.
			assert.NoError(t, err,
				"IdleTimeout minimum validation must only apply to on-demand "+
					"registries, not lifecycle=%q", lc)
		})
	}
}

// itoa converts an int to string without importing strconv in the test file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
