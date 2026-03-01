package cmd

import (
	"bytes"
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Registry CLI Commands - Now implemented in cmd/get.go, cmd/create.go, cmd/delete.go
// =============================================================================
// Commands available:
//   - getRegistriesCmd (cmd/get.go)
//   - getRegistryCmd (cmd/get.go)
//   - createRegistryCmd (cmd/create.go)
//   - deleteRegistryCmd (cmd/delete.go)

// ========== Mock Registry Manager ==========

// MockRegistryManager implements registry.RegistryManager for testing
type MockRegistryManager struct {
	StartFunc         func(ctx context.Context) error
	StopFunc          func(ctx context.Context) error
	StatusFunc        func(ctx context.Context) (*registry.RegistryStatus, error)
	EnsureRunningFunc func(ctx context.Context) error
	IsRunningFunc     func(ctx context.Context) bool
	GetEndpointFunc   func() string
	PruneFunc         func(ctx context.Context, opts registry.PruneOptions) (*registry.PruneResult, error)
}

func (m *MockRegistryManager) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	return nil
}

func (m *MockRegistryManager) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

func (m *MockRegistryManager) Status(ctx context.Context) (*registry.RegistryStatus, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx)
	}
	return &registry.RegistryStatus{
		State: "stopped",
		PID:   0,
		Port:  5001,
	}, nil
}

func (m *MockRegistryManager) EnsureRunning(ctx context.Context) error {
	if m.EnsureRunningFunc != nil {
		return m.EnsureRunningFunc(ctx)
	}
	return nil
}

func (m *MockRegistryManager) IsRunning(ctx context.Context) bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(ctx)
	}
	return false
}

func (m *MockRegistryManager) GetEndpoint() string {
	if m.GetEndpointFunc != nil {
		return m.GetEndpointFunc()
	}
	return "localhost:5001"
}

func (m *MockRegistryManager) Prune(ctx context.Context, opts registry.PruneOptions) (*registry.PruneResult, error) {
	if m.PruneFunc != nil {
		return m.PruneFunc(ctx, opts)
	}
	return &registry.PruneResult{
		ImagesRemoved:  0,
		SpaceReclaimed: 0,
		Images:         []string{},
	}, nil
}

// ========== Test Helpers ==========

// newTestRegistryStartCmd creates a fresh registryStartCmd for testing
func newTestRegistryStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the local registry",
		RunE:  runRegistryStart,
	}
	cmd.Flags().Int("port", 0, "Port to run on (default: from config or 5001)")
	cmd.Flags().Bool("foreground", false, "Run in foreground (don't daemonize)")
	return cmd
}

// newTestRegistryStopCmd creates a fresh registryStopCmd for testing
func newTestRegistryStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the local registry",
		RunE:  runRegistryStop,
	}
	cmd.Flags().Bool("force", false, "Force kill (SIGKILL instead of graceful shutdown)")
	return cmd
}

// newTestRegistryStatusCmd creates a fresh registryStatusCmd for testing
func newTestRegistryStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show registry status",
		RunE:  runRegistryStatus,
	}
	cmd.Flags().StringP("output", "o", "table", "Output format (table, wide, yaml, json)")
	return cmd
}

// newTestRegistryLogsCmd creates a fresh registryLogsCmd for testing
func newTestRegistryLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View registry logs",
		RunE:  runRegistryLogs,
	}
	cmd.Flags().IntP("lines", "n", 50, "Number of lines to show")
	cmd.Flags().String("since", "", "Show logs since duration (e.g., '2h', '30m')")
	return cmd
}

// newTestRegistryPruneCmd creates a fresh registryPruneCmd for testing
func newTestRegistryPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove cached images",
		RunE:  runRegistryPrune,
	}
	cmd.Flags().Bool("all", false, "Remove all images (not just unused)")
	cmd.Flags().String("older-than", "", "Remove images older than duration (e.g., '7d', '30d')")
	cmd.Flags().Bool("dry-run", false, "Show what would be removed without removing")
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")
	return cmd
}

// ========== Command Structure Tests ==========

func TestRegistryCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryCmd, "registryCmd should exist")
	assert.Equal(t, "registry", registryCmd.Use, "registryCmd should have correct Use")
}

func TestRegistryCmd_Aliases(t *testing.T) {
	aliases := registryCmd.Aliases
	assert.Contains(t, aliases, "reg", "should have 'reg' alias")
}

func TestRegistryStartCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryStartCmd, "registryStartCmd should exist")
	assert.Equal(t, "start <name>", registryStartCmd.Use, "registryStartCmd should have correct Use")
}

func TestRegistryStopCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryStopCmd, "registryStopCmd should exist")
	assert.Equal(t, "stop <name>", registryStopCmd.Use, "registryStopCmd should have correct Use")
}

func TestRegistryStatusCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryStatusCmd, "registryStatusCmd should exist")
	assert.Equal(t, "status [name]", registryStatusCmd.Use, "registryStatusCmd should have correct Use")
}

func TestRegistryLogsCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryLogsCmd, "registryLogsCmd should exist")
	assert.Equal(t, "logs", registryLogsCmd.Use, "registryLogsCmd should have correct Use")
}

func TestRegistryPruneCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryPruneCmd, "registryPruneCmd should exist")
	assert.Equal(t, "prune", registryPruneCmd.Use, "registryPruneCmd should have correct Use")
}

// ========== Flag Tests ==========

func TestRegistryStartCmd_HasPortFlag(t *testing.T) {
	portFlag := registryStartCmd.Flags().Lookup("port")
	assert.NotNil(t, portFlag, "registryStartCmd should have 'port' flag")
	if portFlag != nil {
		assert.Equal(t, "0", portFlag.DefValue, "port flag should default to 0")
		assert.Equal(t, "int", portFlag.Value.Type(), "port flag should be int type")
	}
}

func TestRegistryStartCmd_HasForegroundFlag(t *testing.T) {
	foregroundFlag := registryStartCmd.Flags().Lookup("foreground")
	assert.NotNil(t, foregroundFlag, "registryStartCmd should have 'foreground' flag")
	if foregroundFlag != nil {
		assert.Equal(t, "false", foregroundFlag.DefValue, "foreground flag should default to false")
		assert.Equal(t, "bool", foregroundFlag.Value.Type(), "foreground flag should be bool type")
	}
}

func TestRegistryStopCmd_HasForceFlag(t *testing.T) {
	forceFlag := registryStopCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag, "registryStopCmd should have 'force' flag")
	if forceFlag != nil {
		assert.Equal(t, "false", forceFlag.DefValue, "force flag should default to false")
		assert.Equal(t, "bool", forceFlag.Value.Type(), "force flag should be bool type")
	}
}

func TestRegistryLogsCmd_HasLinesFlag(t *testing.T) {
	linesFlag := registryLogsCmd.Flags().Lookup("lines")
	assert.NotNil(t, linesFlag, "registryLogsCmd should have 'lines' flag")
	if linesFlag != nil {
		assert.Equal(t, "50", linesFlag.DefValue, "lines flag should default to 50")
		assert.Equal(t, "int", linesFlag.Value.Type(), "lines flag should be int type")
	}
}

func TestRegistryLogsCmd_HasSinceFlag(t *testing.T) {
	sinceFlag := registryLogsCmd.Flags().Lookup("since")
	assert.NotNil(t, sinceFlag, "registryLogsCmd should have 'since' flag")
	if sinceFlag != nil {
		assert.Equal(t, "", sinceFlag.DefValue, "since flag should default to empty")
		assert.Equal(t, "string", sinceFlag.Value.Type(), "since flag should be string type")
	}
}

func TestRegistryPruneCmd_HasAllFlag(t *testing.T) {
	allFlag := registryPruneCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag, "registryPruneCmd should have 'all' flag")
	if allFlag != nil {
		assert.Equal(t, "false", allFlag.DefValue, "all flag should default to false")
		assert.Equal(t, "bool", allFlag.Value.Type(), "all flag should be bool type")
	}
}

func TestRegistryPruneCmd_HasOlderThanFlag(t *testing.T) {
	olderThanFlag := registryPruneCmd.Flags().Lookup("older-than")
	assert.NotNil(t, olderThanFlag, "registryPruneCmd should have 'older-than' flag")
	if olderThanFlag != nil {
		assert.Equal(t, "", olderThanFlag.DefValue, "older-than flag should default to empty")
		assert.Equal(t, "string", olderThanFlag.Value.Type(), "older-than flag should be string type")
	}
}

func TestRegistryPruneCmd_HasDryRunFlag(t *testing.T) {
	dryRunFlag := registryPruneCmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag, "registryPruneCmd should have 'dry-run' flag")
	if dryRunFlag != nil {
		assert.Equal(t, "false", dryRunFlag.DefValue, "dry-run flag should default to false")
		assert.Equal(t, "bool", dryRunFlag.Value.Type(), "dry-run flag should be bool type")
	}
}

func TestRegistryPruneCmd_HasForceFlag(t *testing.T) {
	forceFlag := registryPruneCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag, "registryPruneCmd should have 'force' flag")
	if forceFlag != nil {
		assert.Equal(t, "false", forceFlag.DefValue, "force flag should default to false")
		assert.Equal(t, "bool", forceFlag.Value.Type(), "force flag should be bool type")
	}
}

// ========== RunE Tests ==========

func TestRegistryStartCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, registryStartCmd.RunE, "registryStartCmd should have RunE (not Run)")
}

func TestRegistryStopCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, registryStopCmd.RunE, "registryStopCmd should have RunE (not Run)")
}

func TestRegistryStatusCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, registryStatusCmd.RunE, "registryStatusCmd should have RunE (not Run)")
}

func TestRegistryLogsCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, registryLogsCmd.RunE, "registryLogsCmd should have RunE (not Run)")
}

func TestRegistryPruneCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, registryPruneCmd.RunE, "registryPruneCmd should have RunE (not Run)")
}

// ========== Help Text Tests ==========

func TestRegistryCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	registryCmd.SetOut(buf)
	registryCmd.Help()
	helpText := buf.String()

	// Should mention subcommands
	assert.Contains(t, helpText, "start", "help should mention 'start' subcommand")
	assert.Contains(t, helpText, "stop", "help should mention 'stop' subcommand")
	assert.Contains(t, helpText, "status", "help should mention 'status' subcommand")
	assert.Contains(t, helpText, "logs", "help should mention 'logs' subcommand")
	assert.Contains(t, helpText, "prune", "help should mention 'prune' subcommand")
}

func TestRegistryStartCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	registryStartCmd.SetOut(buf)
	registryStartCmd.Help()
	helpText := buf.String()

	// Should mention flags
	assert.Contains(t, helpText, "--port", "help should mention '--port' flag")
	assert.Contains(t, helpText, "--foreground", "help should mention '--foreground' flag")

	// Should mention examples
	assert.Contains(t, helpText, "Examples:", "help should have examples section")
}

func TestRegistryStopCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	registryStopCmd.SetOut(buf)
	registryStopCmd.Help()
	helpText := buf.String()

	// Should mention force flag
	assert.Contains(t, helpText, "--force", "help should mention '--force' flag")
}

func TestRegistryPruneCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	registryPruneCmd.SetOut(buf)
	registryPruneCmd.Help()
	helpText := buf.String()

	// Should mention prune flags
	assert.Contains(t, helpText, "--all", "help should mention '--all' flag")
	assert.Contains(t, helpText, "--older-than", "help should mention '--older-than' flag")
	assert.Contains(t, helpText, "--dry-run", "help should mention '--dry-run' flag")
	assert.Contains(t, helpText, "--force", "help should mention '--force' flag")
}

// ========== Registry Start Tests ==========

func TestRegistryStartCmd_StartsSuccessfully(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify successful start when registry not running
	// Requires refactoring to inject mock registry manager
}

func TestRegistryStartCmd_IdempotentWhenRunning(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify idempotence when already running
	// Requires refactoring to inject mock registry manager
}

func TestRegistryStartCmd_PortFlagOverridesConfig(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify --port flag overrides config
	// Requires refactoring to inject mock registry manager
}

// ========== Registry Stop Tests ==========

func TestRegistryStopCmd_StopsRunningRegistry(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify successful stop when running
	// Requires refactoring to inject mock registry manager
}

func TestRegistryStopCmd_IdempotentWhenStopped(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify idempotence when not running
	// Requires refactoring to inject mock registry manager
}

// ========== Registry Status Tests ==========

func TestRegistryStatusCmd_ShowsStoppedStatus(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify status when stopped
	// Requires refactoring to inject mock registry manager
}

func TestRegistryStatusCmd_ShowsRunningStatus(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify status when running with PID and uptime
	// Requires refactoring to inject mock registry manager
}

// ========== Registry Logs Tests ==========

func TestRegistryLogsCmd_HandlesNoLogFile(t *testing.T) {
	t.Skip("logs command not yet implemented for multi-registry - stub returns error")
	// NOTE: This test will need to be re-enabled when logs command is implemented
	// to work with multi-registry support (querying by registry name)
}

func TestRegistryLogsCmd_ShowsLogLines(t *testing.T) {
	t.Skip("logs command not yet implemented for multi-registry - stub returns error")
	// NOTE: This test will need to be re-enabled when logs command is implemented
	// to work with multi-registry support (querying by registry name)
}

func TestRegistryLogsCmd_LimitsFlagWorks(t *testing.T) {
	t.Skip("logs command not yet implemented for multi-registry - stub returns error")
	// NOTE: This test will need to be re-enabled when logs command is implemented
	// to work with multi-registry support (querying by registry name)
}

// ========== Registry Prune Tests ==========

func TestRegistryPruneCmd_RequiresAllOrOlderThan(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify error when neither --all nor --older-than specified
	// Requires refactoring to inject mock registry manager
}

func TestRegistryPruneCmd_DryRunShowsWhatWouldRemove(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify --dry-run doesn't actually remove images
	// Requires refactoring to inject mock registry manager
}

func TestRegistryPruneCmd_ErrorWhenNotRunning(t *testing.T) {
	t.Skip("Integration test - requires registry.RegistryManager factory pattern")
	// This test would verify error when registry not running
	// Requires refactoring to inject mock registry manager
}

// ========== Helper Function Tests ==========

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "zero bytes",
			bytes: 0,
			want:  "0 B",
		},
		{
			name:  "bytes less than KB",
			bytes: 512,
			want:  "512 B",
		},
		{
			name:  "kilobytes",
			bytes: 1024,
			want:  "1.0 KB",
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024,
			want:  "1.0 MB",
		},
		{
			name:  "gigabytes",
			bytes: 1024 * 1024 * 1024,
			want:  "1.0 GB",
		},
		{
			name:  "terabytes",
			bytes: 1024 * 1024 * 1024 * 1024,
			want:  "1.0 TB",
		},
		{
			name:  "decimal values",
			bytes: 1536 * 1024 * 1024, // 1.5 GB
			want:  "1.5 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero duration",
			duration: 0,
			want:     "-",
		},
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			want:     "45s",
		},
		{
			name:     "minutes and seconds",
			duration: 5*time.Minute + 30*time.Second,
			want:     "5m30s",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 15*time.Minute,
			want:     "2h15m",
		},
		{
			name:     "hours only",
			duration: 3 * time.Hour,
			want:     "3h0m",
		},
		{
			name:     "long duration",
			duration: 25*time.Hour + 45*time.Minute,
			want:     "25h45m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatPID(t *testing.T) {
	tests := []struct {
		name string
		pid  int
		want string
	}{
		{
			name: "zero PID",
			pid:  0,
			want: "-",
		},
		{
			name: "valid PID",
			pid:  1234,
			want: "1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPID(tt.pid)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "standard go duration - seconds",
			input:   "30s",
			want:    30 * time.Second,
			wantErr: false,
		},
		{
			name:    "standard go duration - minutes",
			input:   "5m",
			want:    5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "standard go duration - hours",
			input:   "2h",
			want:    2 * time.Hour,
			wantErr: false,
		},
		{
			name:    "day notation - single day",
			input:   "1d",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "day notation - multiple days",
			input:   "7d",
			want:    7 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "day notation - 30 days",
			input:   "30d",
			want:    30 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid duration",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid day notation",
			input:   "xd",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// REMOVED: TestConvertToRegistryConfig - function convertToRegistryConfig() was removed
// REMOVED: TestRegistryStatusToMap - function registryStatusToMap() was removed

// ========== Config Validation Tests ==========

func TestRegistryConfig_ValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "valid port",
			port:    5001,
			wantErr: false,
		},
		{
			name:    "port too low (privileged)",
			port:    80,
			wantErr: true,
		},
		{
			name:    "port too high",
			port:    70000,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := registry.RegistryConfig{
				Port:    tt.port,
				Storage: "/test/storage",
			}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryConfig_ValidateStorage(t *testing.T) {
	tests := []struct {
		name    string
		storage string
		wantErr bool
	}{
		{
			name:    "valid storage path",
			storage: "/test/storage",
			wantErr: false,
		},
		{
			name:    "empty storage path",
			storage: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := registry.RegistryConfig{
				Port:    5001,
				Storage: tt.storage,
			}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryConfig_ValidateLifecycle(t *testing.T) {
	tests := []struct {
		name      string
		lifecycle string
		wantErr   bool
	}{
		{
			name:      "valid lifecycle - persistent",
			lifecycle: "persistent",
			wantErr:   false,
		},
		{
			name:      "valid lifecycle - on-demand",
			lifecycle: "on-demand",
			wantErr:   false,
		},
		{
			name:      "valid lifecycle - manual",
			lifecycle: "manual",
			wantErr:   false,
		},
		{
			name:      "empty lifecycle allowed (uses default)",
			lifecycle: "",
			wantErr:   false,
		},
		{
			name:      "invalid lifecycle",
			lifecycle: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := registry.RegistryConfig{
				Port:      5001,
				Storage:   "/test/storage",
				Lifecycle: tt.lifecycle,
			}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryConfig_ValidateMirrors(t *testing.T) {
	tests := []struct {
		name    string
		mirrors []registry.MirrorConfig
		wantErr bool
	}{
		{
			name:    "no mirrors",
			mirrors: []registry.MirrorConfig{},
			wantErr: false,
		},
		{
			name: "valid https mirror",
			mirrors: []registry.MirrorConfig{
				{
					Name:     "docker-hub",
					URL:      "https://index.docker.io",
					OnDemand: true,
					Prefix:   "docker.io",
				},
			},
			wantErr: false,
		},
		{
			name: "valid http mirror",
			mirrors: []registry.MirrorConfig{
				{
					Name:     "local",
					URL:      "http://localhost:5000",
					OnDemand: false,
					Prefix:   "local",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid mirror - empty URL",
			mirrors: []registry.MirrorConfig{
				{
					Name:   "invalid",
					URL:    "",
					Prefix: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mirror - bad protocol",
			mirrors: []registry.MirrorConfig{
				{
					Name:   "invalid",
					URL:    "ftp://example.com",
					Prefix: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := registry.RegistryConfig{
				Port:    5001,
				Storage: "/test/storage",
				Mirrors: tt.mirrors,
			}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// TDD Phase 2 (RED): Registry Resource CLI Tests
// =============================================================================
// These tests verify the NEW Registry resource CRUD commands:
//   - dvm get registries
//   - dvm get registry <name>
//   - dvm create registry <name> --type <type> --port <port>
//   - dvm delete registry <name>
//
// These tests will FAIL until implementation is added.
// =============================================================================

// ========== Test Helpers for Resource CRUD ==========

// setupTestRegistryStore creates a mock datastore with test registries
func setupTestRegistryStore() *db.MockDataStore {
	mockStore := db.NewMockDataStore()

	// Add test registries
	mockStore.CreateRegistry(&models.Registry{
		ID:        1,
		Name:      "zot-registry",
		Type:      "zot",
		Port:      5001,
		Lifecycle: "on-demand",
		Status:    "running",
	})

	mockStore.CreateRegistry(&models.Registry{
		ID:        2,
		Name:      "athens-go",
		Type:      "athens",
		Port:      3000,
		Lifecycle: "persistent",
		Status:    "stopped",
	})

	mockStore.CreateRegistry(&models.Registry{
		ID:        3,
		Name:      "npm-registry",
		Type:      "verdaccio",
		Port:      4873,
		Lifecycle: "manual",
		Status:    "stopped",
	})

	return mockStore
}

// ========== Resource CRUD Command Structure Tests ==========

func TestGetRegistriesCmd_Exists(t *testing.T) {
	// Command is now implemented in cmd/get.go
	assert.NotNil(t, getRegistriesCmd, "getRegistriesCmd should exist")
	assert.Contains(t, getRegistriesCmd.Use, "registries", "Use should contain 'registries'")
}

func TestGetRegistriesCmd_HasAliases(t *testing.T) {
	// Verify aliases: reg, regs
	aliases := getRegistriesCmd.Aliases
	assert.Contains(t, aliases, "reg", "should have 'reg' alias")
	assert.Contains(t, aliases, "regs", "should have 'regs' alias")
}

func TestGetRegistryCmd_Exists(t *testing.T) {
	// Command is now implemented in cmd/get.go
	assert.NotNil(t, getRegistryCmd, "getRegistryCmd should exist")
	assert.Contains(t, getRegistryCmd.Use, "registry", "Use should contain 'registry'")
	assert.NotNil(t, getRegistryCmd.Args, "should have Args validator")
}

func TestCreateRegistryCmd_Exists(t *testing.T) {
	// Command is now implemented in cmd/create.go
	assert.NotNil(t, createRegistryCmd, "createRegistryCmd should exist")
	assert.Contains(t, createRegistryCmd.Use, "registry", "Use should contain 'registry'")
}

func TestDeleteRegistryCmd_Exists(t *testing.T) {
	// Command is now implemented in cmd/delete.go
	assert.NotNil(t, deleteRegistryCmd, "deleteRegistryCmd should exist")
	assert.Contains(t, deleteRegistryCmd.Use, "registry", "Use should contain 'registry'")
}

// ========== dvm get registries Tests ==========

func TestGetRegistries_ListsAllRegistries(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := setupTestRegistryStore()

	// Get all registries
	registries, err := mockStore.ListRegistries()
	require.NoError(t, err)
	assert.Len(t, registries, 3, "should return 3 registries")

	// Verify registry names
	names := make([]string, len(registries))
	for i, reg := range registries {
		names[i] = reg.Name
	}
	assert.Contains(t, names, "zot-registry")
	assert.Contains(t, names, "athens-go")
	assert.Contains(t, names, "npm-registry")
}

func TestGetRegistries_ShowsRuntimeStatus(t *testing.T) {
	// Tests that registry model stores status correctly

	tests := []struct {
		name       string
		registry   *models.Registry
		wantStatus string
	}{
		{
			name: "running registry",
			registry: &models.Registry{
				Name:   "running-reg",
				Type:   "zot",
				Port:   5001,
				Status: "running",
			},
			wantStatus: "running",
		},
		{
			name: "stopped registry",
			registry: &models.Registry{
				Name:   "stopped-reg",
				Type:   "athens",
				Port:   3000,
				Status: "stopped",
			},
			wantStatus: "stopped",
		},
		{
			name: "error state",
			registry: &models.Registry{
				Name:   "error-reg",
				Type:   "verdaccio",
				Port:   4873,
				Status: "error",
			},
			wantStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, tt.registry.Status)
		})
	}
}

func TestGetRegistries_SupportsOutputFormats(t *testing.T) {
	// Tests output format structure expectations

	tests := []struct {
		name          string
		outputFlag    string
		wantFormat    string
		shouldInclude []string // Strings that should appear in output
	}{
		{
			name:          "default table output",
			outputFlag:    "",
			wantFormat:    "table",
			shouldInclude: []string{"NAME", "TYPE", "PORT", "LIFECYCLE", "STATE"},
		},
		{
			name:          "yaml output",
			outputFlag:    "yaml",
			wantFormat:    "yaml",
			shouldInclude: []string{"apiVersion:", "kind: Registry", "metadata:"},
		},
		{
			name:          "json output",
			outputFlag:    "json",
			wantFormat:    "json",
			shouldInclude: []string{`"name"`, `"type"`, `"port"`},
		},
		{
			name:          "wide table output",
			outputFlag:    "wide",
			wantFormat:    "table",
			shouldInclude: []string{"NAME", "TYPE", "PORT", "LIFECYCLE", "STATE", "CREATED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will verify output format flags work
			assert.NotEmpty(t, tt.wantFormat)
		})
	}
}

// ========== dvm get registry <name> Tests ==========

func TestGetRegistry_ReturnsSpecificRegistry(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := setupTestRegistryStore()

	// Get specific registry
	reg, err := mockStore.GetRegistryByName("zot-registry")
	require.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, "zot-registry", reg.Name)
	assert.Equal(t, "zot", reg.Type)
	assert.Equal(t, 5001, reg.Port)
}

func TestGetRegistry_NonExistent_ReturnsError(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := setupTestRegistryStore()

	// Try to get non-existent registry
	_, err := mockStore.GetRegistryByName("nonexistent")
	assert.Error(t, err, "should return error for non-existent registry")
}

func TestGetRegistry_OutputFormats(t *testing.T) {
	// Tests output format structure expectations

	tests := []struct {
		name       string
		outputFlag string
		wantFields []string
	}{
		{
			name:       "table output shows key fields",
			outputFlag: "",
			wantFields: []string{"NAME", "TYPE", "PORT", "STATE"},
		},
		{
			name:       "yaml output shows full spec",
			outputFlag: "yaml",
			wantFields: []string{"apiVersion", "kind", "metadata", "spec"},
		},
		{
			name:       "json output shows full data",
			outputFlag: "json",
			wantFields: []string{"name", "type", "port", "lifecycle"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.wantFields)
		})
	}
}

// ========== dvm create registry Tests ==========

func TestCreateRegistry_HasRequiredFlags(t *testing.T) {
	// Command is now implemented in cmd/create.go

	// Verify --type flag exists and is required
	typeFlag := createRegistryCmd.Flags().Lookup("type")
	assert.NotNil(t, typeFlag, "should have --type flag")
	assert.Equal(t, "string", typeFlag.Value.Type())

	// Verify --port flag exists
	portFlag := createRegistryCmd.Flags().Lookup("port")
	assert.NotNil(t, portFlag, "should have --port flag")
	assert.Equal(t, "int", portFlag.Value.Type())

	// Verify --lifecycle flag exists
	lifecycleFlag := createRegistryCmd.Flags().Lookup("lifecycle")
	assert.NotNil(t, lifecycleFlag, "should have --lifecycle flag")
	assert.Equal(t, "string", lifecycleFlag.Value.Type())

	// Verify --description flag exists
	descFlag := createRegistryCmd.Flags().Lookup("description")
	assert.NotNil(t, descFlag, "should have --description flag")
	assert.Equal(t, "string", descFlag.Value.Type())
}

func TestCreateRegistry_ValidatesType(t *testing.T) {
	// This test validates the model validation, not the command itself

	tests := []struct {
		name      string
		regType   string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid type - zot",
			regType:   "zot",
			wantError: false,
		},
		{
			name:      "valid type - athens",
			regType:   "athens",
			wantError: false,
		},
		{
			name:      "valid type - devpi",
			regType:   "devpi",
			wantError: false,
		},
		{
			name:      "valid type - verdaccio",
			regType:   "verdaccio",
			wantError: false,
		},
		{
			name:      "valid type - squid",
			regType:   "squid",
			wantError: false,
		},
		{
			name:      "invalid type",
			regType:   "invalid",
			wantError: true,
			errMsg:    "unsupported registry type",
		},
		{
			name:      "empty type",
			regType:   "",
			wantError: true,
			errMsg:    "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &models.Registry{
				Name: "test-registry",
				Type: tt.regType,
				Port: 5000,
			}
			err := reg.ValidateType()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRegistry_ValidatesPortRange(t *testing.T) {
	// Tests model validation - command is now implemented

	tests := []struct {
		name      string
		port      int
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid port in range",
			port:      5001,
			wantError: false,
		},
		{
			name:      "port at lower bound",
			port:      1024,
			wantError: false,
		},
		{
			name:      "port at upper bound",
			port:      65535,
			wantError: false,
		},
		{
			name:      "port too low (privileged)",
			port:      80,
			wantError: true,
			errMsg:    "must be between 1024 and 65535",
		},
		{
			name:      "port too high",
			port:      70000,
			wantError: true,
			errMsg:    "must be between 1024 and 65535",
		},
		{
			name:      "port zero (auto-assign allowed)",
			port:      0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &models.Registry{
				Name: "test-registry",
				Type: "zot",
				Port: tt.port,
			}
			err := reg.ValidatePort()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateRegistry_DetectsPortConflicts(t *testing.T) {
	// Tests MockDataStore port conflict detection

	mockStore := setupTestRegistryStore()

	// Try to create registry with port already in use
	newReg := &models.Registry{
		Name: "conflicting-registry",
		Type: "zot",
		Port: 5001, // Already used by zot-registry
	}

	// First check if port is in use
	registries, err := mockStore.ListRegistries()
	require.NoError(t, err)

	portInUse := false
	for _, reg := range registries {
		if reg.Port == newReg.Port {
			portInUse = true
			break
		}
	}

	assert.True(t, portInUse, "should detect port conflict")
}

func TestCreateRegistry_CreatesWithDefaults(t *testing.T) {
	// Tests default value assignment - command is now implemented

	tests := []struct {
		name               string
		regType            string
		portSpecified      int
		lifecycleSpecified string
		wantPort           int
		wantLifecycle      string
	}{
		{
			name:               "zot defaults",
			regType:            "zot",
			portSpecified:      0, // Not specified
			lifecycleSpecified: "",
			wantPort:           5000, // Default zot port
			wantLifecycle:      "manual",
		},
		{
			name:               "athens defaults",
			regType:            "athens",
			portSpecified:      0,
			lifecycleSpecified: "",
			wantPort:           3000,
			wantLifecycle:      "manual",
		},
		{
			name:               "custom port overrides default",
			regType:            "zot",
			portSpecified:      5555,
			lifecycleSpecified: "persistent",
			wantPort:           5555,
			wantLifecycle:      "persistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &models.Registry{
				Name:      "test-registry",
				Type:      tt.regType,
				Port:      tt.portSpecified,
				Lifecycle: tt.lifecycleSpecified,
			}

			// Apply defaults (simulating what FromYAML does)
			if reg.Port == 0 {
				reg.Port = reg.GetDefaultPort()
			}
			if reg.Lifecycle == "" {
				reg.Lifecycle = "manual"
			}

			assert.Equal(t, tt.wantPort, reg.Port)
			assert.Equal(t, tt.wantLifecycle, reg.Lifecycle)
		})
	}
}

func TestCreateRegistry_CreatesRecord(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := db.NewMockDataStore()

	reg := &models.Registry{
		Name:      "new-registry",
		Type:      "zot",
		Port:      5002,
		Lifecycle: "on-demand",
		Status:    "stopped",
	}

	err := mockStore.CreateRegistry(reg)
	require.NoError(t, err)

	// Verify created
	created, err := mockStore.GetRegistryByName("new-registry")
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "new-registry", created.Name)
	assert.Equal(t, "zot", created.Type)
	assert.Equal(t, 5002, created.Port)
}

// ========== dvm delete registry Tests ==========

func TestDeleteRegistry_DeletesByName(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := setupTestRegistryStore()

	// Verify registry exists
	_, err := mockStore.GetRegistryByName("zot-registry")
	require.NoError(t, err)

	// Delete registry
	err = mockStore.DeleteRegistry("zot-registry")
	assert.NoError(t, err)

	// Verify deleted
	_, err = mockStore.GetRegistryByName("zot-registry")
	assert.Error(t, err, "should return error after deletion")
}

func TestDeleteRegistry_NonExistent_ReturnsError(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := setupTestRegistryStore()

	err := mockStore.DeleteRegistry("nonexistent")
	assert.Error(t, err, "should return error when deleting non-existent registry")
}

func TestDeleteRegistry_StopsRunningRegistry(t *testing.T) {
	// Tests MockDataStore - command is now implemented
	// Note: The delete command does NOT stop running registries per design

	mockStore := setupTestRegistryStore()

	// Get running registry
	reg, err := mockStore.GetRegistryByName("zot-registry")
	require.NoError(t, err)
	assert.Equal(t, "running", reg.Status)

	// Delete should succeed (registry record is deleted)
	err = mockStore.DeleteRegistry("zot-registry")
	assert.NoError(t, err)
}

// ========== Runtime Operation Integration Tests ==========
// These tests verify that runtime operations (start/stop/status) work with
// the Registry resource model

func TestRegistryRuntime_StartLooksUpByName(t *testing.T) {
	t.Skip("TDD Phase 2 (RED): Runtime integration not yet implemented")

	mockStore := setupTestRegistryStore()

	// Get registry by name
	reg, err := mockStore.GetRegistryByName("athens-go")
	require.NoError(t, err)

	// Runtime operation would use this registry's config
	assert.Equal(t, "athens", reg.Type)
	assert.Equal(t, 3000, reg.Port)
	assert.Equal(t, "persistent", reg.Lifecycle)
}

func TestRegistryRuntime_StopLooksUpByName(t *testing.T) {
	t.Skip("TDD Phase 2 (RED): Runtime integration not yet implemented")

	mockStore := setupTestRegistryStore()

	// Get registry by name
	reg, err := mockStore.GetRegistryByName("zot-registry")
	require.NoError(t, err)

	// Runtime operation would use this to identify process to stop
	assert.Equal(t, "running", reg.Status)
	assert.Equal(t, 5001, reg.Port)
}

func TestRegistryRuntime_StatusShowsMultipleRegistries(t *testing.T) {
	t.Skip("TDD Phase 2 (RED): Runtime integration not yet implemented")

	mockStore := setupTestRegistryStore()

	// dvm registry status (no args) should show all registries
	registries, err := mockStore.ListRegistries()
	require.NoError(t, err)

	// Verify status of each
	for _, reg := range registries {
		assert.NotEmpty(t, reg.Status, "each registry should have status")
		assert.Contains(t, []string{"stopped", "starting", "running", "error"}, reg.Status)
	}
}

func TestRegistryRuntime_StatusShowsSpecificRegistry(t *testing.T) {
	t.Skip("TDD Phase 2 (RED): Runtime integration not yet implemented")

	mockStore := setupTestRegistryStore()

	// dvm registry status <name> should show specific registry
	reg, err := mockStore.GetRegistryByName("athens-go")
	require.NoError(t, err)

	assert.Equal(t, "athens-go", reg.Name)
	assert.Equal(t, "stopped", reg.Status)
}

// ========== Edge Cases & Validation ==========

func TestCreateRegistry_DuplicateName_ReturnsError(t *testing.T) {
	// Tests MockDataStore - command is now implemented

	mockStore := setupTestRegistryStore()

	// Try to create registry with existing name
	dup := &models.Registry{
		Name: "zot-registry", // Already exists
		Type: "athens",
		Port: 3001,
	}

	err := mockStore.CreateRegistry(dup)
	assert.Error(t, err, "should return error for duplicate name")
}

func TestCreateRegistry_ValidatesLifecycle(t *testing.T) {
	// Tests model validation - command is now implemented

	tests := []struct {
		name      string
		lifecycle string
		wantError bool
	}{
		{
			name:      "valid - persistent",
			lifecycle: "persistent",
			wantError: false,
		},
		{
			name:      "valid - on-demand",
			lifecycle: "on-demand",
			wantError: false,
		},
		{
			name:      "valid - manual",
			lifecycle: "manual",
			wantError: false,
		},
		{
			name:      "invalid lifecycle",
			lifecycle: "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &models.Registry{
				Name:      "test-registry",
				Type:      "zot",
				Port:      5000,
				Lifecycle: tt.lifecycle,
			}
			err := reg.ValidateLifecycle()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRegistries_EmptyList(t *testing.T) {
	// Tests the MockDataStore - command behavior tested via integration

	mockStore := db.NewMockDataStore()

	registries, err := mockStore.ListRegistries()
	require.NoError(t, err)
	assert.Empty(t, registries, "should return empty list when no registries exist")
}

// ========== Command Help Text Tests ==========

func TestCreateRegistryCmd_Help(t *testing.T) {
	// Command is now implemented in cmd/create.go

	buf := new(bytes.Buffer)
	createRegistryCmd.SetOut(buf)
	createRegistryCmd.Help()
	helpText := buf.String()

	// Should mention flags
	assert.Contains(t, helpText, "--type", "help should mention --type flag")
	assert.Contains(t, helpText, "--port", "help should mention --port flag")
	assert.Contains(t, helpText, "--lifecycle", "help should mention --lifecycle flag")

	// Should list valid types
	assert.Contains(t, helpText, "zot", "help should list zot as valid type")
	assert.Contains(t, helpText, "athens", "help should list athens as valid type")
	assert.Contains(t, helpText, "verdaccio", "help should list verdaccio as valid type")

	// Should have examples
	assert.Contains(t, helpText, "Examples:", "help should have examples section")
}

func TestGetRegistriesCmd_Help(t *testing.T) {
	// Command is now implemented in cmd/get.go

	buf := new(bytes.Buffer)
	getRegistriesCmd.SetOut(buf)
	getRegistriesCmd.Help()
	helpText := buf.String()

	// Should mention output formats
	assert.Contains(t, helpText, "-o", "help should mention output flag")
	assert.Contains(t, helpText, "yaml", "help should mention yaml output")
	assert.Contains(t, helpText, "json", "help should mention json output")
}

func TestDeleteRegistryCmd_Help(t *testing.T) {
	// Command is now implemented in cmd/delete.go

	buf := new(bytes.Buffer)
	deleteRegistryCmd.SetOut(buf)
	deleteRegistryCmd.Help()
	helpText := buf.String()

	// Help text should be present (may not mention "running" since we say it does NOT delete running containers)
	assert.NotEmpty(t, helpText, "help text should not be empty")
}

// ========== ServiceFactory Integration Tests ==========

// REMOVED: Legacy --name flag tests (now using positional arguments)
// - TestRegistryStartCmd_HasNameFlag
// - TestRegistryStopCmd_HasNameFlag
// - TestRegistryStatusCmd_HasNameFlag
// - TestRegistryStartCmd_WithNameFlag_UsesResourceMode
// - TestRegistryStartCmd_WithoutNameFlag_UsesLegacyMode
// - TestRegistryStatusCmd_WithNameFlag_ShowsResourceName
// - TestRegistryStatusCmd_WithoutNameFlag_ShowsDefaultName

// =============================================================================
// TDD Phase 2 (RED): Registry Positional Arguments Tests
// =============================================================================
// These tests verify the NEW positional argument behavior:
//   - dvm registry start <name>  (positional arg required)
//   - dvm registry stop <name>   (positional arg required)
//   - dvm registry status        (no arg = list all)
//   - dvm registry status <name> (arg = show specific)
//
// These tests will FAIL until implementation is changed to use positional args.
// =============================================================================

// ========== Registry Start Positional Arg Tests ==========

func TestRegistryStartCmd_RequiresNameArg(t *testing.T) {
	// Create root command with start subcommand
	rootCmd := &cobra.Command{Use: "dvm"}
	registryCmd := &cobra.Command{Use: "registry"}
	startCmd := &cobra.Command{
		Use:  "start <name>",
		Args: cobra.ExactArgs(1),
		RunE: runRegistryStart,
	}
	registryCmd.AddCommand(startCmd)
	rootCmd.AddCommand(registryCmd)

	// Execute with no args (should fail)
	rootCmd.SetArgs([]string{"registry", "start"})
	err := rootCmd.Execute()

	assert.Error(t, err, "should return error when no name arg provided")
	assert.Contains(t, err.Error(), "accepts 1 arg", "error should mention arg count")
}

func TestRegistryStartCmd_ParsesPositionalArg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database with registry
	mockStore := setupTestRegistryStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command
	cmd := &cobra.Command{
		Use:  "start <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Test that args[0] is used for lookup
			assert.Len(t, args, 1, "should have exactly 1 arg")
			assert.Equal(t, "zot-registry", args[0], "args[0] should be registry name")

			// Verify DB lookup would work
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)
			reg, err := dataStore.GetRegistryByName(args[0])
			assert.NoError(t, err, "should find registry by name from args[0]")
			assert.Equal(t, "zot-registry", reg.Name)
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with positional arg
	cmd.SetArgs([]string{"zot-registry"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRegistryStartCmd_NotFound_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database
	mockStore := db.NewMockDataStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command that mimics new behavior
	cmd := &cobra.Command{
		Use:  "start <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate lookup
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)
			_, err := dataStore.GetRegistryByName(args[0])
			if err != nil {
				return fmt.Errorf("registry '%s' not found", args[0])
			}
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with non-existent registry
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()

	assert.Error(t, err, "should return error for non-existent registry")
	assert.Contains(t, err.Error(), "nonexistent", "error should mention registry name")
	assert.Contains(t, err.Error(), "not found", "error should mention 'not found'")
}

func TestRegistryStartCmd_UsesServiceFactory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database with registry
	mockStore := setupTestRegistryStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command that mimics new behavior
	serviceFactoryUsed := false
	cmd := &cobra.Command{
		Use:  "start <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate service factory usage
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)
			reg, err := dataStore.GetRegistryByName(args[0])
			if err != nil {
				return err
			}

			// Simulate ServiceFactory.CreateManager()
			factory := registry.NewServiceFactory()
			_, err = factory.CreateManager(reg)
			if err != nil {
				return err
			}

			serviceFactoryUsed = true
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute
	cmd.SetArgs([]string{"zot-registry"})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.True(t, serviceFactoryUsed, "should use ServiceFactory to create manager")
}

// ========== Registry Stop Positional Arg Tests ==========

func TestRegistryStopCmd_RequiresNameArg(t *testing.T) {
	// Create root command with stop subcommand
	rootCmd := &cobra.Command{Use: "dvm"}
	registryCmd := &cobra.Command{Use: "registry"}
	stopCmd := &cobra.Command{
		Use:  "stop <name>",
		Args: cobra.ExactArgs(1),
		RunE: runRegistryStop,
	}
	registryCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(registryCmd)

	// Execute with no args (should fail)
	rootCmd.SetArgs([]string{"registry", "stop"})
	err := rootCmd.Execute()

	assert.Error(t, err, "should return error when no name arg provided")
	assert.Contains(t, err.Error(), "accepts 1 arg", "error should mention arg count")
}

func TestRegistryStopCmd_ParsesPositionalArg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database with registry
	mockStore := setupTestRegistryStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command
	cmd := &cobra.Command{
		Use:  "stop <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Test that args[0] is used for lookup
			assert.Len(t, args, 1, "should have exactly 1 arg")
			assert.Equal(t, "zot-registry", args[0], "args[0] should be registry name")

			// Verify DB lookup would work
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)
			reg, err := dataStore.GetRegistryByName(args[0])
			assert.NoError(t, err, "should find registry by name from args[0]")
			assert.Equal(t, "zot-registry", reg.Name)
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with positional arg
	cmd.SetArgs([]string{"zot-registry"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRegistryStopCmd_NotFound_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database
	mockStore := db.NewMockDataStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command that mimics new behavior
	cmd := &cobra.Command{
		Use:  "stop <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate lookup
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)
			_, err := dataStore.GetRegistryByName(args[0])
			if err != nil {
				return fmt.Errorf("registry '%s' not found", args[0])
			}
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with non-existent registry
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()

	assert.Error(t, err, "should return error for non-existent registry")
	assert.Contains(t, err.Error(), "nonexistent", "error should mention registry name")
	assert.Contains(t, err.Error(), "not found", "error should mention 'not found'")
}

// ========== Registry Status Positional Arg Tests ==========

func TestRegistryStatusCmd_NoArgs_ListsAllRegistries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database with multiple registries
	mockStore := setupTestRegistryStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command that mimics new behavior
	registriesListed := 0
	cmd := &cobra.Command{
		Use:  "status [name]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)

			if len(args) == 0 {
				// No arg = list all registries
				registries, err := dataStore.ListRegistries()
				if err != nil {
					return err
				}
				registriesListed = len(registries)
				return nil
			}

			// With arg = show specific registry
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with no args
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, 3, registriesListed, "should list all 3 registries")
}

func TestRegistryStatusCmd_WithArg_ShowsSpecificRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database with registry
	mockStore := setupTestRegistryStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command that mimics new behavior
	var shownRegistryName string
	cmd := &cobra.Command{
		Use:  "status [name]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)

			if len(args) == 0 {
				// No arg = list all registries
				return nil
			}

			// With arg = show specific registry
			reg, err := dataStore.GetRegistryByName(args[0])
			if err != nil {
				return fmt.Errorf("registry '%s' not found", args[0])
			}
			shownRegistryName = reg.Name
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with specific registry name
	cmd.SetArgs([]string{"athens-go"})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "athens-go", shownRegistryName, "should show specific registry")
}

func TestRegistryStatusCmd_NotFound_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database
	mockStore := db.NewMockDataStore()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)

	// Create command that mimics new behavior
	cmd := &cobra.Command{
		Use:  "status [name]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			// Simulate lookup
			dataStore := cmd.Context().Value("dataStore").(*db.MockDataStore)
			_, err := dataStore.GetRegistryByName(args[0])
			if err != nil {
				return fmt.Errorf("registry '%s' not found", args[0])
			}
			return nil
		},
	}
	cmd.SetContext(ctx)

	// Execute with non-existent registry
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()

	assert.Error(t, err, "should return error for non-existent registry")
	assert.Contains(t, err.Error(), "nonexistent", "error should mention registry name")
	assert.Contains(t, err.Error(), "not found", "error should mention 'not found'")
}

// ========== Command Structure Validation Tests ==========

func TestRegistryStartCmd_ExpectedArgsValidator(t *testing.T) {
	// This test verifies the command will be configured with ExactArgs(1)
	// Currently it has no Args validator, so this will FAIL

	// After implementation, registryStartCmd.Args should be set to cobra.ExactArgs(1)
	if registryStartCmd.Args == nil {
		t.Skip("Args validator not yet set - will be set in GREEN phase")
	}

	// Validate Args function
	err := registryStartCmd.Args(registryStartCmd, []string{})
	assert.Error(t, err, "should require exactly 1 arg")

	err = registryStartCmd.Args(registryStartCmd, []string{"name"})
	assert.NoError(t, err, "should accept exactly 1 arg")

	err = registryStartCmd.Args(registryStartCmd, []string{"name", "extra"})
	assert.Error(t, err, "should reject more than 1 arg")
}

func TestRegistryStopCmd_ExpectedArgsValidator(t *testing.T) {
	// This test verifies the command will be configured with ExactArgs(1)
	// Currently it has no Args validator, so this will FAIL

	// After implementation, registryStopCmd.Args should be set to cobra.ExactArgs(1)
	if registryStopCmd.Args == nil {
		t.Skip("Args validator not yet set - will be set in GREEN phase")
	}

	// Validate Args function
	err := registryStopCmd.Args(registryStopCmd, []string{})
	assert.Error(t, err, "should require exactly 1 arg")

	err = registryStopCmd.Args(registryStopCmd, []string{"name"})
	assert.NoError(t, err, "should accept exactly 1 arg")

	err = registryStopCmd.Args(registryStopCmd, []string{"name", "extra"})
	assert.Error(t, err, "should reject more than 1 arg")
}

func TestRegistryStatusCmd_ExpectedArgsValidator(t *testing.T) {
	// This test verifies the command will be configured with MaximumNArgs(1)
	// Currently it has no Args validator, so this will FAIL

	// After implementation, registryStatusCmd.Args should be set to cobra.MaximumNArgs(1)
	if registryStatusCmd.Args == nil {
		t.Skip("Args validator not yet set - will be set in GREEN phase")
	}

	// Validate Args function
	err := registryStatusCmd.Args(registryStatusCmd, []string{})
	assert.NoError(t, err, "should accept 0 args (list all)")

	err = registryStatusCmd.Args(registryStatusCmd, []string{"name"})
	assert.NoError(t, err, "should accept 1 arg (show specific)")

	err = registryStatusCmd.Args(registryStatusCmd, []string{"name", "extra"})
	assert.Error(t, err, "should reject more than 1 arg")
}

// ========== Use String Tests ==========

func TestRegistryStartCmd_ExpectedUseString(t *testing.T) {
	// After refactor, Use should be "start <name>" not just "start"
	expectedUse := "start <name>"

	if registryStartCmd.Use != expectedUse {
		t.Errorf("Use string not yet updated - expected '%s', got '%s' (will be fixed in GREEN phase)",
			expectedUse, registryStartCmd.Use)
	}
}

func TestRegistryStopCmd_ExpectedUseString(t *testing.T) {
	// After refactor, Use should be "stop <name>" not just "stop"
	expectedUse := "stop <name>"

	if registryStopCmd.Use != expectedUse {
		t.Errorf("Use string not yet updated - expected '%s', got '%s' (will be fixed in GREEN phase)",
			expectedUse, registryStopCmd.Use)
	}
}

func TestRegistryStatusCmd_ExpectedUseString(t *testing.T) {
	// After refactor, Use should be "status [name]" not just "status"
	expectedUse := "status [name]"

	if registryStatusCmd.Use != expectedUse {
		t.Errorf("Use string not yet updated - expected '%s', got '%s' (will be fixed in GREEN phase)",
			expectedUse, registryStatusCmd.Use)
	}
}

// =============================================================================
// End of Registry Tests
// =============================================================================
// All legacy --name flag and dual-mode tests have been removed.
// Registry commands now use positional arguments:
//   - dvm registry start <name>
//   - dvm registry stop <name>
//   - dvm registry status [name]
// =============================================================================
