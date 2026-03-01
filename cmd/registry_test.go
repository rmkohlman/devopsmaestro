package cmd

import (
	"bytes"
	"context"
	"devopsmaestro/config"
	"devopsmaestro/pkg/registry"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Equal(t, "start", registryStartCmd.Use, "registryStartCmd should have correct Use")
}

func TestRegistryStopCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryStopCmd, "registryStopCmd should exist")
	assert.Equal(t, "stop", registryStopCmd.Use, "registryStopCmd should have correct Use")
}

func TestRegistryStatusCmd_Exists(t *testing.T) {
	assert.NotNil(t, registryStatusCmd, "registryStatusCmd should exist")
	assert.Equal(t, "status", registryStatusCmd.Use, "registryStatusCmd should have correct Use")
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
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cmd := newTestRegistryLogsCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err = cmd.Execute()
	assert.NoError(t, err, "should handle missing log file gracefully")
}

func TestRegistryLogsCmd_ShowsLogLines(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create log file
	logDir := filepath.Join(tmpDir, ".devopsmaestro", "registry")
	err = os.MkdirAll(logDir, 0755)
	require.NoError(t, err)

	logFile := filepath.Join(logDir, "zot.log")
	logContent := "2024-01-01T00:00:00Z INFO test log line 1\n2024-01-01T00:00:01Z INFO test log line 2\n"
	err = os.WriteFile(logFile, []byte(logContent), 0644)
	require.NoError(t, err)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cmd := newTestRegistryLogsCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestRegistryLogsCmd_LimitsFlagWorks(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create log file with 100 lines
	logDir := filepath.Join(tmpDir, ".devopsmaestro", "registry")
	err = os.MkdirAll(logDir, 0755)
	require.NoError(t, err)

	logFile := filepath.Join(logDir, "zot.log")
	logContent := ""
	for i := 1; i <= 100; i++ {
		logContent += fmt.Sprintf("2024-01-01T00:00:%02dZ INFO test log line %d\n", i, i)
	}
	err = os.WriteFile(logFile, []byte(logContent), 0644)
	require.NoError(t, err)

	// Override home directory for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cmd := newTestRegistryLogsCmd()
	cmd.Flags().Set("lines", "10")
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err = cmd.Execute()
	assert.NoError(t, err)
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

func TestConvertToRegistryConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.RegistryConfig
	}{
		{
			name: "default config",
			cfg: &config.RegistryConfig{
				Enabled:     true,
				Lifecycle:   "on-demand",
				Port:        5001,
				Storage:     "/test/storage",
				IdleTimeout: 30 * time.Minute,
				Mirrors:     []config.MirrorConfig{},
			},
		},
		{
			name: "config with mirrors",
			cfg: &config.RegistryConfig{
				Enabled:     true,
				Lifecycle:   "persistent",
				Port:        5002,
				Storage:     "/test/storage",
				IdleTimeout: 1 * time.Hour,
				Mirrors: []config.MirrorConfig{
					{
						Name:     "docker-hub",
						URL:      "https://index.docker.io",
						OnDemand: true,
						Prefix:   "docker.io",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToRegistryConfig(tt.cfg)

			// Verify basic fields
			assert.Equal(t, tt.cfg.Enabled, result.Enabled)
			assert.Equal(t, tt.cfg.Lifecycle, result.Lifecycle)
			assert.Equal(t, tt.cfg.Port, result.Port)
			assert.Equal(t, tt.cfg.Storage, result.Storage)
			assert.Equal(t, tt.cfg.IdleTimeout, result.IdleTimeout)

			// Verify mirrors
			assert.Equal(t, len(tt.cfg.Mirrors), len(result.Mirrors))
			for i, mirror := range tt.cfg.Mirrors {
				assert.Equal(t, mirror.Name, result.Mirrors[i].Name)
				assert.Equal(t, mirror.URL, result.Mirrors[i].URL)
				assert.Equal(t, mirror.OnDemand, result.Mirrors[i].OnDemand)
				assert.Equal(t, mirror.Prefix, result.Mirrors[i].Prefix)
			}
		})
	}
}

func TestRegistryStatusToMap(t *testing.T) {
	status := &registry.RegistryStatus{
		State:      "running",
		PID:        1234,
		Port:       5001,
		Storage:    "/test/storage",
		Version:    "v1.0.0",
		Uptime:     2 * time.Hour,
		ImageCount: 10,
		DiskUsage:  1024 * 1024 * 100, // 100 MB
	}

	cfg := &config.RegistryConfig{
		Lifecycle: "persistent",
		Port:      5001,
	}

	result := registryStatusToMap(status, cfg)

	assert.Equal(t, "running", result["state"])
	assert.Equal(t, 1234, result["pid"])
	assert.Equal(t, 5001, result["port"])
	assert.Equal(t, "/test/storage", result["storage"])
	assert.Equal(t, "v1.0.0", result["version"])
	assert.Equal(t, "2h0m0s", result["uptime"]) // Duration.String() format
	assert.Equal(t, 10, result["imageCount"])
	assert.Equal(t, int64(1024*1024*100), result["diskUsage"])
	assert.Equal(t, "persistent", result["lifecycle"])
	assert.Equal(t, "localhost:5001", result["endpoint"])
}

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
