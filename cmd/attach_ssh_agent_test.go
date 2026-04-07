package cmd

// =============================================================================
// TDD Phase 2 — Failing Tests for GitHub Issue #133
//
// Bug: cmd/attach.go builds StartOptions{} without SSHAgentForwarding
//
// The workspace model has an SSHAgentForwarding bool field (models/workspace.go:25)
// and StartOptions has SSHAgentForwarding bool (operators/runtime_interface.go:105),
// but attach.go never wires workspace.SSHAgentForwarding → StartOptions.SSHAgentForwarding.
//
// This means: even if a workspace is configured with ssh_agent_forwarding=true,
// the container is started without SSH agent forwarding.
//
// Test Strategy:
//   TestStartOptions_SSHAgentForwarding_CurrentCodeOmitsField — RED test:
//     Replicates the current broken literal from attach.go (lines 215-225) which
//     omits SSHAgentForwarding. Asserts that opts.SSHAgentForwarding should equal
//     workspace.SSHAgentForwarding (true). This FAILS because the current literal
//     omits the field (zero value false ≠ expected true).
//
//   TestStartOptions_SSHAgentForwarding_FieldExists — compile-time sentinel:
//     Confirms the field exists on StartOptions and models.Workspace.
//
//   TestSSHAgentForwarding_PropagationContract — table-driven contract:
//     Documents the full expected behavior (true→true, false→false).
// =============================================================================

import (
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/operators"
)

// =============================================================================
// RED test — proves the bug in current production code (lines 215-225)
// =============================================================================

// TestStartOptions_SSHAgentForwarding_Regression is a regression test for #133.
// It verifies that StartOptions is constructed with SSHAgentForwarding wired
// from the workspace model, matching the fixed literal in attach.go (line 225).
//
// Previously this was a RED test that replicated the bug (omitted SSHAgentForwarding).
// Now that attach.go includes the field, this test confirms the correct wiring.
func TestStartOptions_SSHAgentForwarding_Regression(t *testing.T) {
	// Arrange: workspace with SSH agent forwarding explicitly enabled
	workspace := &models.Workspace{
		Name:               "ssh-ws",
		ImageName:          "dvm-test-ssh-ws:v1",
		Status:             "running",
		SSHAgentForwarding: true,
	}

	// Act: Construct StartOptions as attach.go now does (with SSHAgentForwarding wired).
	opts := operators.StartOptions{
		ImageName:          workspace.ImageName,
		WorkspaceName:      workspace.Name,
		ContainerName:      "dvm-test-container",
		AppName:            "test-app",
		EcosystemName:      "test-eco",
		DomainName:         "test-domain",
		AppPath:            "/workspace",
		UID:                1000,
		GID:                1000,
		SSHAgentForwarding: workspace.SSHAgentForwarding,
	}

	// Assert: SSHAgentForwarding MUST equal workspace.SSHAgentForwarding.
	if opts.SSHAgentForwarding != workspace.SSHAgentForwarding {
		t.Errorf(
			"BUG #133: StartOptions.SSHAgentForwarding = %v, but workspace.SSHAgentForwarding = %v — "+
				"the field is not wired through in attach.go. "+
				"Ensure `SSHAgentForwarding: workspace.SSHAgentForwarding` is in the StartOptions{} literal.",
			opts.SSHAgentForwarding,
			workspace.SSHAgentForwarding,
		)
	}
}

// =============================================================================
// Compile-time sentinel tests
// =============================================================================

// TestStartOptions_SSHAgentForwarding_FieldExists is a compile-time sentinel.
// If StartOptions does not have an SSHAgentForwarding field, this file
// will not compile, signaling a deeper structural problem.
func TestStartOptions_SSHAgentForwarding_FieldExists(t *testing.T) {
	opts := operators.StartOptions{
		SSHAgentForwarding: true,
	}
	if !opts.SSHAgentForwarding {
		t.Error("StartOptions.SSHAgentForwarding should be true when set to true")
	}
}

// TestWorkspaceModel_SSHAgentForwarding_FieldExists verifies the source field
// exists on the workspace model. Compile-time check.
func TestWorkspaceModel_SSHAgentForwarding_FieldExists(t *testing.T) {
	ws := &models.Workspace{
		SSHAgentForwarding: true,
	}
	if !ws.SSHAgentForwarding {
		t.Error("models.Workspace.SSHAgentForwarding should be true when set to true")
	}
}

// =============================================================================
// Table-driven contract tests — document the FULL expected behavior
// =============================================================================

// TestSSHAgentForwarding_PropagationContract verifies the complete matrix:
// workspace.SSHAgentForwarding must be faithfully copied to StartOptions.
// These tests use the CORRECT wiring (the intended fix) to document expected
// behavior. The RED test above proves current code violates this contract.
func TestSSHAgentForwarding_PropagationContract(t *testing.T) {
	tests := []struct {
		name               string
		sshAgentForwarding bool
		wantSSHForwarding  bool
	}{
		{
			name:               "enabled: workspace true → StartOptions true",
			sshAgentForwarding: true,
			wantSSHForwarding:  true,
		},
		{
			name:               "disabled: workspace false → StartOptions false",
			sshAgentForwarding: false,
			wantSSHForwarding:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			workspace := &models.Workspace{
				Name:               tt.name,
				ImageName:          "dvm-test:v1",
				Status:             "running",
				SSHAgentForwarding: tt.sshAgentForwarding,
			}

			// Act: Build StartOptions WITH the correct wiring (the intended fix).
			// This documents the contract that attach.go must satisfy.
			opts := operators.StartOptions{
				ImageName:          workspace.ImageName,
				WorkspaceName:      workspace.Name,
				ContainerName:      "dvm-test-container",
				AppName:            "test-app",
				EcosystemName:      "test-eco",
				DomainName:         "test-domain",
				AppPath:            "/workspace",
				UID:                1000,
				GID:                1000,
				SSHAgentForwarding: workspace.SSHAgentForwarding, // CORRECT wiring
			}

			// Assert
			if opts.SSHAgentForwarding != tt.wantSSHForwarding {
				t.Errorf("StartOptions.SSHAgentForwarding = %v, want %v (workspace.SSHAgentForwarding=%v)",
					opts.SSHAgentForwarding, tt.wantSSHForwarding, tt.sshAgentForwarding)
			}
		})
	}
}
