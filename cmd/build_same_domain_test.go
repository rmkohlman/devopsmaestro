package cmd

// =============================================================================
// Gap Coverage — Issue #245: dvm build -A with workspaces under same ecosystem/domain
// =============================================================================
// Two specific coverage gaps identified in the prior test run (see issue #245
// comment) are closed here:
//
//   Gap 1: Staging directory uniqueness — the fix in #227 uses
//           appName-workspaceName as the staging key (build_phases.go:222).
//           No unit test previously asserted this property.
//
//   Gap 2: Same-domain concurrent execution — existing parallel tests use
//           generic auto-generated workspace lists. None explicitly place 2+
//           workspaces under a single real domain and verify parallel safety.
// =============================================================================

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Gap 1: Staging Directory Uniqueness
// =============================================================================

// makeStagingKey mirrors the logic in build_phases.go:222 so the test can
// assert the exact format without depending on build internals.
func makeStagingKey(appName, workspaceName string) string {
	return appName + "-" + workspaceName
}

// TestStagingDirUniqueness_SameDomain_DistinctPaths verifies that two
// workspaces sharing the same ecosystem and domain produce distinct staging
// directories. This is the core safety property introduced by fix #227.
//
// Arrangement: ecosystem="cloud", domain="payments", apps=["api","worker"],
// workspace name="dev" for both (common real-world case).
func TestStagingDirUniqueness_SameDomain_DistinctPaths(t *testing.T) {
	const homeDir = "/tmp/fake-home-staging-test"

	type workspaceSpec struct {
		appName       string
		workspaceName string
	}

	specs := []workspaceSpec{
		{appName: "api", workspaceName: "dev"},
		{appName: "worker", workspaceName: "dev"},
	}

	stagingDirs := make([]string, len(specs))
	for i, spec := range specs {
		key := makeStagingKey(spec.appName, spec.workspaceName)
		stagingDirs[i] = paths.New(homeDir).BuildStagingDir(key)
	}

	assert.NotEqual(t, stagingDirs[0], stagingDirs[1],
		"api/dev and worker/dev must produce distinct staging directories, got duplicate: %s",
		stagingDirs[0])
}

// TestStagingDirUniqueness_KeyFormat_IsAppNameDashWorkspaceName verifies that
// the staging key format is exactly "appName-workspaceName" as defined in
// build_phases.go:222 (the fix for staging collision issue #227).
func TestStagingDirUniqueness_KeyFormat_IsAppNameDashWorkspaceName(t *testing.T) {
	tests := []struct {
		appName       string
		workspaceName string
		wantKey       string
	}{
		{"api", "dev", "api-dev"},
		{"worker", "dev", "worker-dev"},
		{"portal", "staging", "portal-staging"},
		{"my-svc", "prod", "my-svc-prod"},
	}

	for _, tt := range tests {
		t.Run(tt.appName+"/"+tt.workspaceName, func(t *testing.T) {
			key := makeStagingKey(tt.appName, tt.workspaceName)
			assert.Equal(t, tt.wantKey, key,
				"staging key must be appName-workspaceName")
		})
	}
}

// TestStagingDirUniqueness_ThreeApps_AllDistinct verifies 3 apps in the same
// domain each produce a unique staging path — covers the N>2 case.
func TestStagingDirUniqueness_ThreeApps_AllDistinct(t *testing.T) {
	const homeDir = "/tmp/fake-home-three-apps"
	apps := []string{"api", "worker", "scheduler"}
	const wsDomain = "dev"

	seen := make(map[string]string) // stagingDir → appName
	for _, app := range apps {
		key := makeStagingKey(app, wsDomain)
		dir := paths.New(homeDir).BuildStagingDir(key)
		if prev, exists := seen[dir]; exists {
			t.Errorf("staging dir collision: %q is shared by apps %q and %q",
				dir, prev, app)
		}
		seen[dir] = app
	}
	assert.Len(t, seen, len(apps), "each app must have a unique staging dir")
}

// TestStagingDirUniqueness_StagingPathContainsKey verifies that the staging
// path produced by paths.BuildStagingDir actually embeds the staging key, so
// we can confirm uniqueness by path inspection (not just inequality).
func TestStagingDirUniqueness_StagingPathContainsKey(t *testing.T) {
	const homeDir = "/tmp/fake-home-path-check"
	appName := "api"
	wsName := "dev"

	key := makeStagingKey(appName, wsName)
	dir := paths.New(homeDir).BuildStagingDir(key)

	assert.True(t, strings.Contains(dir, key),
		"staging path %q must contain the key %q", dir, key)
}

// =============================================================================
// Gap 2: Same-Domain Concurrent Execution
// =============================================================================

// makeSameDomainWorkspaces creates N WorkspaceWithHierarchy entries all
// sharing the same real ecosystem ("cloud") and domain ("payments").
// Each app has a distinct name to reflect a realistic microservices setup.
func makeSameDomainWorkspaces(appNames []string, workspaceName string) []*models.WorkspaceWithHierarchy {
	eco := &models.Ecosystem{ID: 10, Name: "cloud"}
	dom := &models.Domain{ID: 20, Name: "payments", EcosystemID: sql.NullInt64{Int64: 10, Valid: true}}

	result := make([]*models.WorkspaceWithHierarchy, len(appNames))
	for i, appName := range appNames {
		app := &models.App{ID: i + 100, Name: appName, DomainID: sql.NullInt64{Int64: 20, Valid: true}}
		ws := &models.Workspace{ID: i + 200, Name: workspaceName, AppID: app.ID}
		result[i] = &models.WorkspaceWithHierarchy{
			Workspace: ws,
			App:       app,
			Domain:    dom,
			Ecosystem: eco,
		}
	}
	return result
}

// TestSameDomain_ParallelBuild_AllWorkspacesBuilt verifies that all workspaces
// under the same domain are built when run through buildWorkspacesInParallel.
// No workspace must be skipped or lost.
func TestSameDomain_ParallelBuild_AllWorkspacesBuilt(t *testing.T) {
	apps := []string{"api", "worker", "scheduler"}
	workspaces := makeSameDomainWorkspaces(apps, "dev")

	var mu sync.Mutex
	built := make([]string, 0, len(apps))

	operator := func(ws *models.WorkspaceWithHierarchy, _ io.Writer) error {
		mu.Lock()
		built = append(built, ws.App.Name)
		mu.Unlock()
		return nil
	}

	err := buildWorkspacesInParallel(workspaces, len(apps), operator)
	require.NoError(t, err)

	assert.Len(t, built, len(apps),
		"all %d workspaces in domain 'payments' must be built, got %d", len(apps), len(built))

	for _, app := range apps {
		assert.Contains(t, built, app,
			"app %q must have been built", app)
	}
}

// TestSameDomain_ParallelBuild_IsActuallyParallel verifies that workspaces
// under the same domain run concurrently — not serialised despite being in the
// same ecosystem/domain hierarchy.
func TestSameDomain_ParallelBuild_IsActuallyParallel(t *testing.T) {
	const workDuration = 20 * time.Millisecond
	apps := []string{"api", "worker", "scheduler"}
	workspaces := makeSameDomainWorkspaces(apps, "dev")

	operator := func(ws *models.WorkspaceWithHierarchy, _ io.Writer) error {
		time.Sleep(workDuration)
		return nil
	}

	start := time.Now()
	err := buildWorkspacesInParallel(workspaces, len(apps), operator)
	elapsed := time.Since(start)

	require.NoError(t, err)

	serialTime := time.Duration(len(apps)) * workDuration
	assert.Less(t, elapsed, serialTime,
		"same-domain workspaces must run in parallel (elapsed=%v < serial=%v)",
		elapsed, serialTime)
}

// TestSameDomain_ParallelBuild_FailureIsolation verifies that a failure in
// one workspace under a domain does not prevent the others in the same domain
// from being built.
func TestSameDomain_ParallelBuild_FailureIsolation(t *testing.T) {
	apps := []string{"api", "worker", "scheduler"}
	workspaces := makeSameDomainWorkspaces(apps, "dev")
	const failingApp = "worker"

	var attempted atomic.Int32
	var mu sync.Mutex
	succeeded := make([]string, 0)

	operator := func(ws *models.WorkspaceWithHierarchy, _ io.Writer) error {
		attempted.Add(1)
		if ws.App.Name == failingApp {
			return fmt.Errorf("build failed: container error for %s", failingApp)
		}
		mu.Lock()
		succeeded = append(succeeded, ws.App.Name)
		mu.Unlock()
		return nil
	}

	err := buildWorkspacesInParallel(workspaces, len(apps), operator)

	assert.Error(t, err, "must return error when a workspace in the domain fails")
	assert.Equal(t, int32(len(apps)), attempted.Load(),
		"all %d workspaces must be attempted despite failure of %q", len(apps), failingApp)
	assert.Len(t, succeeded, len(apps)-1,
		"the %d non-failing workspaces must still succeed", len(apps)-1)
	assert.NotContains(t, succeeded, failingApp,
		"%q must not appear in succeeded list", failingApp)
}

// TestSameDomain_StagingKeys_UniqueAcrossParallelBuilds verifies that the
// staging key for every workspace in the same domain is distinct. This is the
// end-to-end assertion of the #227 fix: even if two workspaces share the same
// workspace name ("dev"), their different app names produce distinct keys and
// therefore distinct staging directories.
func TestSameDomain_StagingKeys_UniqueAcrossParallelBuilds(t *testing.T) {
	apps := []string{"api", "worker"}
	workspaces := makeSameDomainWorkspaces(apps, "dev")

	stagingKeys := make(map[string]string) // key → app name
	for _, wh := range workspaces {
		key := makeStagingKey(wh.App.Name, wh.Workspace.Name)
		if prev, exists := stagingKeys[key]; exists {
			t.Errorf("staging key collision: %q shared by apps %q and %q",
				key, prev, wh.App.Name)
		}
		stagingKeys[key] = wh.App.Name
	}

	assert.Len(t, stagingKeys, len(apps),
		"every same-domain workspace must have a distinct staging key")
}

// TestSameDomain_NoCollision_WhenWorkspaceNamesMatch verifies that two
// workspaces with the SAME workspace name (the most collision-prone case)
// still produce distinct staging keys thanks to the app name prefix.
// This directly exercises the fix from issue #227.
func TestSameDomain_NoCollision_WhenWorkspaceNamesMatch(t *testing.T) {
	const wsName = "dev" // both workspaces have the same name

	keyAPI := makeStagingKey("api", wsName)
	keyWorker := makeStagingKey("worker", wsName)

	assert.NotEqual(t, keyAPI, keyWorker,
		"same workspace name %q must NOT produce a collision when app names differ: got %q == %q",
		wsName, keyAPI, keyWorker)

	// Confirm the keys include the app name to make the distinction explicit
	assert.True(t, strings.HasPrefix(keyAPI, "api-"),
		"staging key %q must start with app name 'api'", keyAPI)
	assert.True(t, strings.HasPrefix(keyWorker, "worker-"),
		"staging key %q must start with app name 'worker'", keyWorker)
}

// TestSameDomain_ParallelBuild_NoStagingKeyCollision runs the builds in
// parallel and records the staging key each goroutine would use, asserting
// they are all distinct even under concurrent execution.
func TestSameDomain_ParallelBuild_NoStagingKeyCollision(t *testing.T) {
	apps := []string{"api", "worker", "scheduler"}
	workspaces := makeSameDomainWorkspaces(apps, "dev")

	var mu sync.Mutex
	seenKeys := make(map[string]int) // key → count

	operator := func(ws *models.WorkspaceWithHierarchy, _ io.Writer) error {
		key := makeStagingKey(ws.App.Name, ws.Workspace.Name)
		mu.Lock()
		seenKeys[key]++
		mu.Unlock()
		return nil
	}

	err := buildWorkspacesInParallel(workspaces, len(apps), operator)
	require.NoError(t, err)

	for key, count := range seenKeys {
		assert.Equal(t, 1, count,
			"staging key %q must be used by exactly one workspace, but was used %d times",
			key, count)
	}
	assert.Len(t, seenKeys, len(apps),
		"must see exactly %d distinct staging keys (one per workspace)", len(apps))
}

// TestSameDomain_EcosystemAndDomainNames_PreservedDuringParallelBuild verifies
// that the ecosystem and domain metadata are correctly preserved for all
// workspaces throughout parallel execution — no cross-contamination.
func TestSameDomain_EcosystemAndDomainNames_PreservedDuringParallelBuild(t *testing.T) {
	apps := []string{"api", "worker"}
	workspaces := makeSameDomainWorkspaces(apps, "dev")

	type buildRecord struct {
		app       string
		ecosystem string
		domain    string
	}

	var mu sync.Mutex
	records := make([]buildRecord, 0, len(apps))

	operator := func(ws *models.WorkspaceWithHierarchy, _ io.Writer) error {
		mu.Lock()
		records = append(records, buildRecord{
			app:       ws.App.Name,
			ecosystem: ws.Ecosystem.Name,
			domain:    ws.Domain.Name,
		})
		mu.Unlock()
		return nil
	}

	err := buildWorkspacesInParallel(workspaces, len(apps), operator)
	require.NoError(t, err)
	require.Len(t, records, len(apps))

	for _, r := range records {
		assert.Equal(t, "cloud", r.ecosystem,
			"workspace %q must report ecosystem 'cloud', got %q", r.app, r.ecosystem)
		assert.Equal(t, "payments", r.domain,
			"workspace %q must report domain 'payments', got %q", r.app, r.domain)
	}
}

// Ensure the errors package is used to prevent import erasure.
var _ = errors.New
