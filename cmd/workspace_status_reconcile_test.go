package cmd

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/operators"
)

// makeWS is a test helper that constructs a minimal *models.Workspace.
func makeWS(name, containerID, initialStatus string) *models.Workspace {
	ws := &models.Workspace{Name: name, Status: initialStatus}
	if containerID != "" {
		ws.ContainerID = sql.NullString{String: containerID, Valid: true}
	}
	return ws
}

// makeInfo is a test helper that constructs a minimal operators.WorkspaceInfo.
func makeInfo(id, name, status string) operators.WorkspaceInfo {
	return operators.WorkspaceInfo{ID: id, Name: name, Status: status}
}

func TestApplyWorkspaceStatusReconcile_MatchByFullID(t *testing.T) {
	const fullID = "abc123def456789012345678"
	ws := makeWS("myws", fullID, "stopped")
	infos := []operators.WorkspaceInfo{makeInfo(fullID, "myws", "Up 5 minutes")}

	applyWorkspaceStatusReconcile([]*models.Workspace{ws}, infos)

	if ws.Status != "running" {
		t.Errorf("expected running, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_MatchByShortID(t *testing.T) {
	// Runtime returns a long ID; DB only stores the 12-char short form.
	longID := "abc123def456789012345678"
	shortID := longID[:12]
	ws := makeWS("myws", shortID, "stopped")
	infos := []operators.WorkspaceInfo{makeInfo(longID, "myws", "Up 1 hour")}

	applyWorkspaceStatusReconcile([]*models.Workspace{ws}, infos)

	if ws.Status != "running" {
		t.Errorf("expected running, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_MatchByName(t *testing.T) {
	// No container ID stored — fall back to name matching.
	ws := makeWS("myws", "", "stopped")
	infos := []operators.WorkspaceInfo{makeInfo("someid123456", "myws", "running")}

	applyWorkspaceStatusReconcile([]*models.Workspace{ws}, infos)

	if ws.Status != "running" {
		t.Errorf("expected running, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_NoMatch_SetsStopped(t *testing.T) {
	ws := makeWS("ghost", "deadbeef123456789", "running")
	infos := []operators.WorkspaceInfo{makeInfo("aabbccdd1234", "other", "Up 2 hours")}

	applyWorkspaceStatusReconcile([]*models.Workspace{ws}, infos)

	if ws.Status != "stopped" {
		t.Errorf("expected stopped, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_NonRunningRuntimeEntry_NotCounted(t *testing.T) {
	ws := makeWS("myws", "aabbccdd1234", "running")
	// Runtime says "Exited" — should not count as running.
	infos := []operators.WorkspaceInfo{makeInfo("aabbccdd1234", "myws", "Exited (0)")}

	applyWorkspaceStatusReconcile([]*models.Workspace{ws}, infos)

	if ws.Status != "stopped" {
		t.Errorf("expected stopped, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_NilEntries_Skipped(t *testing.T) {
	ws := makeWS("ok", "aabbccdd123456", "stopped")
	infos := []operators.WorkspaceInfo{makeInfo("aabbccdd123456", "ok", "running")}

	// nil entries must not panic.
	applyWorkspaceStatusReconcile([]*models.Workspace{nil, ws, nil}, infos)

	if ws.Status != "running" {
		t.Errorf("expected running, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_EmptyInfos_AllStopped(t *testing.T) {
	ws := makeWS("myws", "aabbccdd1234", "running")

	applyWorkspaceStatusReconcile([]*models.Workspace{ws}, nil)

	if ws.Status != "stopped" {
		t.Errorf("expected stopped, got %q", ws.Status)
	}
}

func TestApplyWorkspaceStatusReconcile_MultipleWorkspaces(t *testing.T) {
	tests := []struct {
		name        string
		wsName      string
		containerID string
		initial     string
		runtimeID   string
		runtimeName string
		runtimeStat string
		wantStatus  string
	}{
		{"running by id", "ws1", "aaaa1234bbbb", "stopped", "aaaa1234bbbb", "ws1", "Up", "running"},
		{"stopped no match", "ws2", "cccc5678dddd", "running", "eeee9999ffff", "other", "Up", "stopped"},
		{"running by name fallback", "ws3", "", "stopped", "ffff0000aaaa", "ws3", "running", "running"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ws := makeWS(tc.wsName, tc.containerID, tc.initial)
			infos := []operators.WorkspaceInfo{makeInfo(tc.runtimeID, tc.runtimeName, tc.runtimeStat)}
			applyWorkspaceStatusReconcile([]*models.Workspace{ws}, infos)
			if ws.Status != tc.wantStatus {
				t.Errorf("got %q, want %q", ws.Status, tc.wantStatus)
			}
		})
	}
}

func TestReconcileWorkspaceHierarchyStatuses_NilEntries(t *testing.T) {
	// Should not panic on nil or nil-workspace entries.
	results := []*models.WorkspaceWithHierarchy{
		nil,
		{Workspace: nil},
	}
	// We can't inject a runtime, so this just exercises the nil guard path.
	// On a machine without a container runtime, the function returns early and
	// DB values are preserved — acceptable offline behavior.
	reconcileWorkspaceHierarchyStatuses(results) // must not panic
}
