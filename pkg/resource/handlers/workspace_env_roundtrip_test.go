package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #185 — dvm apply fails with NOT NULL constraint when
// env field is missing from workspace YAML
//
// Root cause:
//  1. models/workspace.go WorkspaceSpec.Env tagged `yaml:"env,omitempty"` —
//     empty maps are excluded from YAML export
//  2. workspace.go ToYAML() explicitly sets envMap = nil when empty, ensuring
//     omitempty removes the field: lines 256-260
//  3. models/workspace.go FromYAML() only calls SetEnv when len > 0 (line 352)
//     — if env field is absent/nil in YAML, w.Env stays sql.NullString{Valid:false}
//  4. SQLite workspaces.env column has NOT NULL constraint — storing a workspace
//     with Env.Valid=false causes: "NOT NULL constraint failed: workspaces.env"
//
// Fix required (per issue #185 acceptance criteria):
//  1. Remove omitempty from WorkspaceSpec.Env yaml tag (or remove the nil-forcing
//     logic in ToYAML) so env field always appears in exported YAML
//  2. In FromYAML or Apply(), default nil Env to SetEnv(map[string]string{})
//     so the NOT NULL constraint is always satisfied
//
// ALL tests in this file MUST FAIL until the fix is implemented.
// =============================================================================

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Test 1: ToYAML must always include env field, even when empty
//
// RED: workspace.go ToYAML() explicitly forces envMap = nil when len == 0
// (lines 256-260), so the `yaml:"env,omitempty"` tag will suppress the field.
// The exported YAML will never contain "env:" for a workspace with no env vars.
// =============================================================================

func TestWorkspaceEnvYAMLExport_AlwaysIncludesEnvField(t *testing.T) {
	h := NewWorkspaceHandler()

	ws := &models.Workspace{
		ID:        1,
		AppID:     1,
		Name:      "env-export-ws",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	// Explicitly set empty env map — simulates a workspace that has no env vars
	ws.SetEnv(map[string]string{})

	res := NewWorkspaceResource(ws, "test-app", "test-domain", "")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// RED: The `omitempty` tag combined with the nil-forcing logic in ToYAML
	// means "env:" will be absent from the output for empty env maps.
	// After fix: env field must always be present so round-trip apply succeeds.
	if !strings.Contains(yamlStr, "env:") {
		t.Errorf(
			"ToYAML() output is missing 'env:' field — the env field is omitted when empty.\n"+
				"This breaks dvm apply when re-importing exported YAML.\n"+
				"Bug: WorkspaceSpec.Env has yaml:\"env,omitempty\" and ToYAML() forces envMap=nil "+
				"when empty (models/workspace.go lines 256-260).\n"+
				"Got YAML:\n%s",
			yamlStr,
		)
	}
}

// =============================================================================
// Test 2: Apply with nil Spec.Env (env field absent from YAML) must not fail
//
// RED: FromYAML() only calls SetEnv when len(yaml.Spec.Env) > 0.
// If YAML has no env field, Spec.Env is nil → SetEnv never called →
// workspace.Env stays sql.NullString{Valid:false} → SQLite NOT NULL fails.
//
// The mock store doesn't enforce NOT NULL, so this test detects the bug by
// checking that the workspace's Env column is valid (non-NULL) after Apply.
// =============================================================================

func TestWorkspaceEnvApply_NilEnvDefaultsToEmptyMap(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	// YAML that omits the env field entirely — as produced by current `dvm get all -o yaml`
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: nil-env-ws
  app: ws-app
  domain: ws-domain
spec:
  image:
    name: ubuntu:22.04
  container:
    command: ["/bin/zsh", "-l"]
    uid: 1000
    gid: 1000
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf(
			"Apply() with missing env field returned error: %v\n"+
				"Bug: NOT NULL constraint triggered because Env column is left invalid.\n"+
				"Fix: default nil Spec.Env to empty map in FromYAML or Apply()",
			err,
		)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("Apply() result is not *WorkspaceResource, got %T", res)
	}

	ws := wr.Workspace()

	// RED: workspace.Env.Valid will be false because SetEnv was never called.
	// The fix must ensure Env is always set to a valid (non-NULL) value.
	if !ws.Env.Valid {
		t.Errorf(
			"workspace.Env.Valid = false after Apply() with no env field in YAML.\n" +
				"This would cause 'NOT NULL constraint failed: workspaces.env' in SQLite.\n" +
				"Bug: FromYAML() only calls SetEnv when len(yaml.Spec.Env) > 0 " +
				"(models/workspace.go line 352-354).\n" +
				"Fix: always call SetEnv, even with empty map, before DB write.",
		)
	}
}

// =============================================================================
// Test 3: Apply with explicit empty env map must succeed (NOT NULL satisfied)
//
// ORANGE: This tests the explicit `env: {}` case. The YAML parser may give us
// an initialized (but empty) map, which still fails the `len > 0` guard in
// FromYAML. This documents that `env: {}` is also broken, not just absent env.
// =============================================================================

func TestWorkspaceEnvApply_EmptyEnvMapSucceeds(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	// YAML with explicit empty env map — the "correct" workaround users shouldn't need
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: empty-env-ws
  app: ws-app
  domain: ws-domain
spec:
  image:
    name: ubuntu:22.04
  env: {}
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf(
			"Apply() with 'env: {}' returned error: %v\n"+
				"Bug: empty map still doesn't satisfy len > 0 guard in FromYAML.\n"+
				"Fix: always call SetEnv(map[string]string{}) when env is nil or empty.",
			err,
		)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("Apply() result is not *WorkspaceResource, got %T", res)
	}

	ws := wr.Workspace()

	// RED: workspace.Env.Valid will be false — empty map triggers same bug as nil.
	if !ws.Env.Valid {
		t.Errorf(
			"workspace.Env.Valid = false after Apply() with 'env: {}'.\n" +
				"This would cause 'NOT NULL constraint failed: workspaces.env' in SQLite.\n" +
				"Fix: SetEnv must be called even when env map is empty.",
		)
	}

	// Also verify the stored env value is a valid JSON representation
	if ws.Env.Valid && ws.Env.String == "" {
		t.Errorf(
			"workspace.Env.String is empty string — should be '{}' for empty env map.\n" +
				"SQLite NOT NULL is satisfied but the value is semantically wrong.",
		)
	}
}

// =============================================================================
// Test 4: Full round-trip — export YAML → parse → verify env field present
//
// RED: Export will omit env field (omitempty), so the parsed-back YAML will
// have Spec.Env == nil. This test verifies the full export path is broken.
// =============================================================================

func TestWorkspaceEnvRoundTrip_ExportParseApply(t *testing.T) {
	// ── Step 1: Build a workspace model with empty env ────────────────────
	ws := &models.Workspace{
		ID:        1,
		AppID:     1,
		Name:      "roundtrip-ws",
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	// SetEnv with empty map — simulates a workspace that was stored with Env="{}"
	ws.SetEnv(map[string]string{})

	// ── Step 2: Export to YAML ─────────────────────────────────────────────
	h := NewWorkspaceHandler()
	res := NewWorkspaceResource(ws, "test-app", "test-domain", "")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// RED: The exported YAML won't have "env:" due to omitempty + nil-forcing.
	// Assert export includes env field for completeness.
	if !strings.Contains(yamlStr, "env:") {
		t.Errorf(
			"Exported YAML is missing 'env:' field.\n"+
				"This means apply will fail with NOT NULL when re-importing.\n"+
				"Got YAML:\n%s",
			yamlStr,
		)
	}

	// ── Step 3: Parse the exported YAML back ──────────────────────────────
	var wsYAML models.WorkspaceYAML
	if err := yaml.Unmarshal(yamlBytes, &wsYAML); err != nil {
		t.Fatalf("yaml.Unmarshal of exported YAML: %v", err)
	}

	// RED: wsYAML.Spec.Env will be nil because env field was omitted in export.
	// After fix: env field present in YAML → Spec.Env is a non-nil empty map.
	if wsYAML.Spec.Env == nil {
		t.Errorf(
			"After parsing exported YAML, WorkspaceSpec.Env == nil.\n" +
				"The env field was not included in the export, so it cannot be parsed.\n" +
				"This confirms the round-trip is broken: export omits env → parse gets nil " +
				"→ apply fails with NOT NULL constraint.",
		)
	}

	// ── Step 4: Apply the parsed YAML to a fresh store ────────────────────
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "rt-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}
	dom := &models.Domain{Name: "test-domain", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	if err := store.CreateDomain(dom); err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}
	app := &models.App{Name: "test-app", DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true}, Path: "/test/app"}
	if err := store.CreateApp(app); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}

	ctx := resource.Context{DataStore: store}
	applyRes, err := h.Apply(ctx, yamlBytes)
	if err != nil {
		t.Fatalf(
			"Apply() of exported YAML failed: %v\n"+
				"Bug: env field missing from export → nil Spec.Env → NOT NULL constraint failure.\n"+
				"This is the exact bug reported in issue #185.",
			err,
		)
	}

	wr, ok := applyRes.(*WorkspaceResource)
	if !ok {
		t.Fatalf("Apply() result is not *WorkspaceResource")
	}

	appliedWS := wr.Workspace()

	// RED: applied workspace Env will be invalid (NOT NULL would fail in SQLite)
	if !appliedWS.Env.Valid {
		t.Errorf(
			"After full round-trip (export → apply), workspace.Env.Valid = false.\n" +
				"This confirms the backup/restore workflow is broken:\n" +
				"  1. dvm get all -o yaml → env field omitted\n" +
				"  2. dvm apply -f backup.yaml → NOT NULL constraint failed: workspaces.env\n" +
				"Fix: export must include env field AND apply must default nil env to empty map.",
		)
	}
}
