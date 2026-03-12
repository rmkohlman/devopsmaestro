package cmd

// =============================================================================
// TDD Phase 2 (RED): Declarative Registry Version -- CLI Changes
// =============================================================================
// These tests define the specification for four CLI changes (CC-1, CC-3, CC-4,
// CC-5) related to the declarative registry version feature.
//
// Tests are written FIRST to drive the implementation (TDD RED phase).
// All tests WILL FAIL until the implementation changes are made.
//
// Changes Required:
//   CC-1: --version flag (long-form only, NO -v short form) on createRegistryCmd
//   CC-3: Table column order: NAME, TYPE, VERSION, PORT, LIFECYCLE, STATE, UPTIME
//         (VERSION moves from absent to 3rd column, after TYPE)
//   CC-4: MarkFlagRequired("type") on createRegistryCmd (replaces manual check)
//   CC-5: Detail view shows "Version" key-value; when desired != installed,
//         show both (e.g., "desired: 1.0.0, installed: 0.9.8")
//
// Helper functions required in implementation (will cause compile errors until added):
//   getRegistriesTableHeaders(wide bool) []string   -- cmd/get_registry.go
//   getRegistryDetailViewKeys() []string            -- cmd/get_registry.go
// =============================================================================

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CC-1: --version flag on dvm create registry
// =============================================================================

// TestCreateRegistryCmd_HasVersionFlag verifies that createRegistryCmd declares
// a --version flag for specifying the desired registry version.
//
// This test EXPECTS TO FAIL until --version is added to createRegistryCmd's
// flag set in cmd/create.go (init() function).
func TestCreateRegistryCmd_HasVersionFlag(t *testing.T) {
	flag := createRegistryCmd.Flags().Lookup("version")
	assert.NotNil(t, flag,
		"createRegistryCmd should have --version flag (CC-1: declarative registry version)")
}

// TestCreateRegistryCmd_VersionFlagIsStringType verifies the --version flag
// holds a string value (semantic version string like "2.1.0").
//
// This test EXPECTS TO FAIL until --version is implemented.
func TestCreateRegistryCmd_VersionFlagIsStringType(t *testing.T) {
	flag := createRegistryCmd.Flags().Lookup("version")
	require.NotNil(t, flag,
		"createRegistryCmd must have --version flag before type can be checked")

	assert.Equal(t, "string", flag.Value.Type(),
		"--version should be a string flag (semantic version like '2.1.0')")
}

// TestCreateRegistryCmd_VersionFlagDefaultsToEmpty verifies that when --version
// is not provided, the default value is an empty string (meaning "use latest").
//
// This test EXPECTS TO FAIL until --version is implemented.
func TestCreateRegistryCmd_VersionFlagDefaultsToEmpty(t *testing.T) {
	flag := createRegistryCmd.Flags().Lookup("version")
	require.NotNil(t, flag,
		"createRegistryCmd must have --version flag before default can be checked")

	assert.Equal(t, "", flag.DefValue,
		"--version should default to empty string (meaning 'use latest/default version')")
}

// TestCreateRegistryCmd_VersionFlagNoShortForm verifies that the --version flag
// has NO single-character shorthand. The short form -v is reserved for --verbose
// on the root command and must not be shadowed here.
//
// This test EXPECTS TO FAIL until --version is implemented (it will then pass
// only if correctly implemented WITHOUT a shorthand). If the developer mistakenly
// adds -v as the shorthand, this test catches that collision.
func TestCreateRegistryCmd_VersionFlagNoShortForm(t *testing.T) {
	flag := createRegistryCmd.Flags().Lookup("version")
	require.NotNil(t, flag,
		"createRegistryCmd must have --version flag before shorthand can be checked")

	assert.Empty(t, flag.Shorthand,
		"--version flag must NOT have a short form (-v collides with --verbose on root, CC-1)")
}

// =============================================================================
// CC-4: MarkFlagRequired("type") on createRegistryCmd
// =============================================================================

// TestCreateRegistryCmd_TypeFlagIsMarkedRequired verifies that the --type flag
// is annotated as required using cobra.MarkFlagRequired(), which provides
// automatic validation before RunE is called.
//
// Currently (pre-implementation), cmd/create.go line ~309 does a manual check:
//
//	if registryType == "" { return fmt.Errorf("--type is required ...") }
//
// After CC-4, MarkFlagRequired("type") replaces that manual check. Cobra stores
// the cobra.BashCompOneRequiredFlag annotation on the flag itself.
//
// This test EXPECTS TO FAIL until createRegistryCmd.MarkFlagRequired("type") is
// called in cmd/create.go's init() function.
func TestCreateRegistryCmd_TypeFlagIsMarkedRequired(t *testing.T) {
	flag := createRegistryCmd.Flags().Lookup("type")
	require.NotNil(t, flag,
		"createRegistryCmd must have --type flag")

	annotations := flag.Annotations
	require.NotNil(t, annotations,
		"--type flag should have annotations set via MarkFlagRequired (CC-4)")

	_, required := annotations[cobra.BashCompOneRequiredFlag]
	assert.True(t, required,
		"--type should be marked as required via MarkFlagRequired (CC-4: replaces manual empty-string check at line ~309 of create.go)")
}

// =============================================================================
// CC-3: Table column order -- VERSION column inserted after TYPE
// =============================================================================

// TestGetRegistries_TableHeaders_VersionAfterType verifies the expected column
// order for the default (non-wide) registry table output.
//
// Current headers (pre-implementation):
//
//	["NAME", "TYPE", "PORT", "LIFECYCLE", "STATE", "UPTIME"]
//
// Expected headers after CC-3:
//
//	["NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME"]
//
// This test EXPECTS TO FAIL until:
//  1. getRegistriesTableHeaders(wide bool) []string is added to cmd/get_registry.go
//  2. getRegistries() is updated to include "VERSION" as the 3rd column
func TestGetRegistries_TableHeaders_VersionAfterType(t *testing.T) {
	// Expected header order per CC-3
	expectedHeaders := []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME"}

	// Sanity-check the expected spec itself (these always pass)
	assert.Equal(t, "VERSION", expectedHeaders[2],
		"spec check: VERSION should be the 3rd column (index 2)")
	assert.Equal(t, "TYPE", expectedHeaders[1],
		"spec check: TYPE should be the 2nd column (index 1)")

	// FAILS until getRegistriesTableHeaders() is added to get_registry.go
	// and the headers slice in getRegistries() includes VERSION
	actualHeaders := getRegistriesTableHeaders(false /* not wide */)
	assert.Equal(t, expectedHeaders, actualHeaders,
		"Default registry table headers should match CC-3 spec: NAME, TYPE, VERSION, PORT, LIFECYCLE, STATE, UPTIME")
}

// TestGetRegistries_WideTableHeaders_VersionAfterType verifies the wide-format
// column order also includes VERSION after TYPE.
//
// Current wide headers (pre-implementation):
//
//	["NAME", "TYPE", "PORT", "LIFECYCLE", "STATE", "UPTIME", "CREATED"]
//
// Expected wide headers after CC-3:
//
//	["NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME", "CREATED"]
//
// This test EXPECTS TO FAIL until getRegistries() in cmd/get_registry.go is
// updated for the wide-format case as well.
func TestGetRegistries_WideTableHeaders_VersionAfterType(t *testing.T) {
	expectedWideHeaders := []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME", "CREATED"}

	// Sanity-check the expected spec itself (these always pass)
	assert.Equal(t, "CREATED", expectedWideHeaders[len(expectedWideHeaders)-1],
		"spec check: CREATED should remain the last column in wide format")
	assert.Equal(t, "VERSION", expectedWideHeaders[2],
		"spec check: VERSION should be the 3rd column (index 2) in wide format")

	// FAILS until getRegistriesTableHeaders() is added to get_registry.go
	actualWideHeaders := getRegistriesTableHeaders(true /* wide */)
	assert.Equal(t, expectedWideHeaders, actualWideHeaders,
		"Wide registry table headers should match CC-3 spec with VERSION after TYPE")
}

// TestGetRegistries_VersionColumnPosition verifies that VERSION is positioned
// at index 2 (between TYPE at index 1 and PORT at index 3) using positional
// assertions rather than full-slice equality. This makes failure messages more
// diagnostic.
//
// This test EXPECTS TO FAIL until CC-3 is implemented.
func TestGetRegistries_VersionColumnPosition(t *testing.T) {
	headers := getRegistriesTableHeaders(false)
	require.NotEmpty(t, headers, "getRegistriesTableHeaders should return non-empty headers")

	versionIdx := -1
	typeIdx := -1
	portIdx := -1
	for i, h := range headers {
		switch h {
		case "VERSION":
			versionIdx = i
		case "TYPE":
			typeIdx = i
		case "PORT":
			portIdx = i
		}
	}

	assert.NotEqual(t, -1, versionIdx,
		"VERSION column should exist in registry table headers (CC-3)")
	if versionIdx != -1 && typeIdx != -1 {
		assert.Equal(t, typeIdx+1, versionIdx,
			"VERSION column should immediately follow TYPE column (CC-3)")
	}
	if versionIdx != -1 && portIdx != -1 {
		assert.Equal(t, versionIdx+1, portIdx,
			"PORT column should immediately follow VERSION column (CC-3)")
	}
}

// =============================================================================
// CC-5: Detail view shows "Version" key-value pair
// =============================================================================

// TestGetRegistry_DetailView_HasVersionKey verifies that the registry detail view
// (dvm get registry <name>) includes a "Version" key in the key-value output.
//
// Current detail view keys (pre-implementation):
//
//	Name, Type, Port, Lifecycle, Status, Description, Created
//
// Expected detail view keys after CC-5:
//
//	Name, Type, Version, Port, Lifecycle, Status, Description, Created
//
// This test EXPECTS TO FAIL until getRegistryDetailViewKeys() is added to
// cmd/get_registry.go AND the KeyValue pairs in getRegistry() include "Version".
func TestGetRegistry_DetailView_HasVersionKey(t *testing.T) {
	expectedKeys := getRegistryDetailViewKeys()
	require.NotEmpty(t, expectedKeys,
		"getRegistryDetailViewKeys() must return non-empty list of keys")

	assert.Contains(t, expectedKeys, "Version",
		"Registry detail view should include 'Version' key (CC-5)")
}

// TestGetRegistry_DetailView_KeyOrder verifies the complete expected key order
// for the registry detail view, with "Version" positioned after "Type".
//
// This test EXPECTS TO FAIL until CC-5 is implemented in cmd/get_registry.go.
func TestGetRegistry_DetailView_KeyOrder(t *testing.T) {
	// Expected key order per CC-5: "Version" inserted after "Type", before "Port"
	expectedKeys := []string{"Name", "Type", "Version", "Port", "Lifecycle", "Status", "Description", "Created"}

	actualKeys := getRegistryDetailViewKeys()
	assert.Equal(t, expectedKeys, actualKeys,
		"Registry detail view keys should match CC-5 spec order: Name, Type, Version, Port, Lifecycle, Status, Description, Created")
}

// TestGetRegistry_DetailView_VersionAfterType verifies the positional requirement:
// the "Version" key must appear immediately after "Type" in the detail view.
//
// This test EXPECTS TO FAIL until CC-5 is implemented.
func TestGetRegistry_DetailView_VersionAfterType(t *testing.T) {
	keys := getRegistryDetailViewKeys()
	require.NotEmpty(t, keys, "detail view keys must not be empty")

	typeIdx := -1
	versionIdx := -1
	for i, k := range keys {
		switch k {
		case "Type":
			typeIdx = i
		case "Version":
			versionIdx = i
		}
	}

	assert.NotEqual(t, -1, typeIdx, "detail view must contain 'Type' key")
	assert.NotEqual(t, -1, versionIdx,
		"detail view must contain 'Version' key (CC-5)")
	if typeIdx != -1 && versionIdx != -1 {
		assert.Equal(t, typeIdx+1, versionIdx,
			"'Version' key must appear immediately after 'Type' in detail view (CC-5)")
	}
}

// =============================================================================
// Table-driven summary: all CC-related flags on createRegistryCmd
// =============================================================================

// TestCreateRegistryCmd_FlagSummary is a table-driven test that verifies the
// complete expected flag set for createRegistryCmd after all CC changes.
//
// Rows that WILL FAIL until implemented:
//   - "version" row: FAILS until CC-1 adds --version flag
//   - "type" wantRequired=true row: FAILS until CC-4 calls MarkFlagRequired
//
// Rows that already PASS (pre-existing flags unchanged):
//   - "port", "lifecycle", "description"
func TestCreateRegistryCmd_FlagSummary(t *testing.T) {
	tests := []struct {
		flagName     string
		wantExists   bool
		wantType     string
		wantShort    string // empty string means no shorthand expected
		wantRequired bool
		description  string
	}{
		{
			flagName:     "type",
			wantExists:   true,
			wantType:     "string",
			wantShort:    "t",
			wantRequired: true,
			description:  "CC-4: --type must be required via MarkFlagRequired",
		},
		{
			flagName:     "version",
			wantExists:   true,
			wantType:     "string",
			wantShort:    "", // CC-1: NO short form, -v is reserved for --verbose
			wantRequired: false,
			description:  "CC-1: --version flag with no short form",
		},
		{
			flagName:     "port",
			wantExists:   true,
			wantType:     "int",
			wantShort:    "p",
			wantRequired: false,
			description:  "pre-existing --port flag (unchanged)",
		},
		{
			flagName:     "lifecycle",
			wantExists:   true,
			wantType:     "string",
			wantShort:    "l",
			wantRequired: false,
			description:  "pre-existing --lifecycle flag (unchanged)",
		},
		{
			flagName:     "description",
			wantExists:   true,
			wantType:     "string",
			wantShort:    "d",
			wantRequired: false,
			description:  "pre-existing --description flag (unchanged)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := createRegistryCmd.Flags().Lookup(tt.flagName)

			if !tt.wantExists {
				assert.Nil(t, flag, "flag --%s should NOT exist: %s", tt.flagName, tt.description)
				return
			}

			require.NotNil(t, flag, "flag --%s must exist: %s", tt.flagName, tt.description)

			assert.Equal(t, tt.wantType, flag.Value.Type(),
				"--%s should be type %s: %s", tt.flagName, tt.wantType, tt.description)

			assert.Equal(t, tt.wantShort, flag.Shorthand,
				"--%s shorthand mismatch: %s", tt.flagName, tt.description)

			if tt.wantRequired {
				annotations := flag.Annotations
				require.NotNil(t, annotations,
					"--%s should have annotations (MarkFlagRequired): %s", tt.flagName, tt.description)
				_, required := annotations[cobra.BashCompOneRequiredFlag]
				assert.True(t, required,
					"--%s should be marked required: %s", tt.flagName, tt.description)
			}
		})
	}
}
