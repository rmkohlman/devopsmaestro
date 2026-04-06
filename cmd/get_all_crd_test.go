package cmd

// ---------------------------------------------------------------------------
// CRD instance export tests — Issue #180
//
// These are RED-phase TDD tests. They verify that `dvm get all -o yaml`
// exports custom resource *instances* (not just CRD definitions). All tests
// in this file are expected to FAIL until the fix is implemented in get_all.go.
// ---------------------------------------------------------------------------

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// seedCRDAndInstance inserts a CRD definition and one custom resource instance
// into the given dataStore. It returns the inserted CRD kind and instance name.
func seedCRDAndInstance(t *testing.T, ds interface {
	CreateCRD(crd *models.CustomResourceDefinition) error
	CreateCustomResource(resource *models.CustomResource) error
}) (kind, instanceName string) {
	t.Helper()

	kind = "AppConfig"
	instanceName = "my-app"

	crd := &models.CustomResourceDefinition{
		Kind:     kind,
		Group:    "devopsmaestro.io",
		Singular: "appconfig",
		Plural:   "appconfigs",
		Scope:    "Global",
		Versions: sql.NullString{
			String: `[{"name":"v1alpha1","served":true,"storage":true}]`,
			Valid:  true,
		},
	}
	require.NoError(t, ds.CreateCRD(crd), "failed to seed CRD definition")

	instance := &models.CustomResource{
		Kind:      kind,
		Name:      instanceName,
		Namespace: sql.NullString{String: "default", Valid: true},
		Spec:      sql.NullString{String: `{"replicas":3}`, Valid: true},
	}
	require.NoError(t, ds.CreateCustomResource(instance), "failed to seed custom resource instance")

	return kind, instanceName
}

// ---------------------------------------------------------------------------
// TestCRDInstancesExported_InGetAll
//
// Verifies that `get all -o yaml` includes custom resource *instances* in its
// output — not just the CRD definition. This MUST FAIL until get_all.go is
// updated to iterate registered CRD kinds and append their instances.
// ---------------------------------------------------------------------------

func TestCRDInstancesExported_InGetAll(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	_, instanceName := seedCRDAndInstance(t, ds)

	cmd := newGetAllTestCmd(t, ds)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll with YAML output should not error")

	var result map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "output should be valid YAML; got: %s", buf.String())

	// The YAML must be a List document
	assert.Equal(t, "List", result["kind"], "output should be kind: List")

	// items must include the custom resource instance
	itemsRaw, exists := result["items"]
	require.True(t, exists, "YAML output should have items field")
	require.NotNil(t, itemsRaw, "items should not be nil")

	items, ok := itemsRaw.([]interface{})
	require.True(t, ok, "items should be a YAML sequence")

	// Collect item names from metadata to confirm our instance is present
	var foundKind, foundName bool
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		if kind, _ := item["kind"].(string); kind == "AppConfig" {
			foundKind = true
			if meta, ok := item["metadata"].(map[string]interface{}); ok {
				if name, _ := meta["name"].(string); name == instanceName {
					foundName = true
				}
			}
		}
	}

	// These assertions MUST FAIL (RED phase) — instances are never exported today
	assert.True(t, foundKind,
		"items should contain an item with kind=AppConfig (custom resource instance kind)")
	assert.True(t, foundName,
		"items should contain the custom resource instance named %q", instanceName)
}

// ---------------------------------------------------------------------------
// TestCRDDefinitionsBeforeInstances_InGetAll
//
// Verifies that CRD definitions appear BEFORE their custom resource instances
// in the exported YAML items list. A consumer applying the YAML must create
// the CRD schema before applying instances against it.
//
// This MUST FAIL until both:
//   (a) instances are exported (TestCRDInstancesExported_InGetAll fix), AND
//   (b) CRD definitions are ordered before instances in allResources.
// ---------------------------------------------------------------------------

func TestCRDDefinitionsBeforeInstances_InGetAll(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	seedCRDAndInstance(t, ds)

	cmd := newGetAllTestCmd(t, ds)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll with YAML output should not error")

	output := buf.String()

	// Find the byte positions of CRD definition and instance in the output.
	//
	// The CRD *definition* is identified by kind: CustomResourceDefinition.
	// The custom resource *instance* is identified by "name: my-app" appearing in
	// the metadata of a top-level item — this key does NOT appear in the CRD
	// definition's YAML (which uses "name: appconfigs.devopsmaestro.io").
	defPos := strings.Index(output, "kind: CustomResourceDefinition")
	instancePos := strings.Index(output, "name: my-app")

	// CRD definition must be present
	require.NotEqual(t, -1, defPos,
		"output should contain 'kind: CustomResourceDefinition'")

	// This MUST FAIL until instances are exported — the instance metadata name
	// "my-app" should appear in the output only when instances are included.
	require.NotEqual(t, -1, instancePos,
		"output should contain 'name: my-app' (instance metadata) — FAILS until instances are exported (Issue #180)")

	assert.Less(t, defPos, instancePos,
		"CRD definition (kind: CustomResourceDefinition) should appear BEFORE instance metadata (name: my-app) in YAML output")
}

// ---------------------------------------------------------------------------
// TestCRDInstanceRoundTrip
//
// Verifies the full backup→restore round-trip: seed a CRD definition + instance,
// export via get-all YAML, confirm the instance is present in the exported YAML.
//
// This MUST FAIL until get_all.go exports instances.
// ---------------------------------------------------------------------------

func TestCRDInstanceRoundTrip(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	kind, instanceName := seedCRDAndInstance(t, ds)

	cmd := newGetAllTestCmd(t, ds)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err, "getAll should not error during round-trip export")

	exportedYAML := buf.String()
	require.NotEmpty(t, exportedYAML, "exported YAML should not be empty")

	// Verify the YAML is a List
	var result map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(exportedYAML), &result),
		"exported YAML must be valid")
	assert.Equal(t, "List", result["kind"], "exported YAML should be kind: List")

	// CRITICAL: The instance must be present in the exported YAML
	// Without this, backup→wipe→restore silently loses all custom resource data.
	assert.Contains(t, exportedYAML, instanceName,
		"exported YAML must contain the custom resource instance name %q — FAILS because instances are never exported (Issue #180)", instanceName)
	assert.Contains(t, exportedYAML, kind,
		"exported YAML must contain the custom resource kind %q", kind)

	// Confirm the spec data is preserved
	assert.Contains(t, exportedYAML, "replicas",
		"exported YAML must contain spec fields of the custom resource instance")
}
