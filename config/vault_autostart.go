package config

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// DefaultVaultSocketPath returns the default MaestroVault Unix socket path.
// Returns an error if the home directory cannot be determined.
func DefaultVaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory for vault socket: %w", err)
	}
	return filepath.Join(home, ".maestrovault", "maestrovault.sock"), nil
}

// EnsureVaultDaemon checks if the MaestroVault daemon is running and starts it if not.
// It connects to the Unix socket at ~/.maestrovault/maestrovault.sock.
// If the socket is missing or not responding, it starts `mav serve --no-touchid` in the background.
// Returns nil if the daemon is running (or was successfully started), error otherwise.
func EnsureVaultDaemon() error {
	socketPath, err := DefaultVaultSocketPath()
	if err != nil {
		return err
	}

	// Check if socket exists and is connectable
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err == nil {
		conn.Close()
		return nil // daemon is already running
	}

	// Socket not available — try to start the daemon
	mavPath, err := exec.LookPath("mav")
	if err != nil {
		return fmt.Errorf("MaestroVault (mav) not found in PATH: install with 'brew install rmkohlman/tap/maestrovault'")
	}

	cmd := exec.Command(mavPath, "serve", "--no-touchid")
	cmd.Stdout = nil
	cmd.Stderr = nil
	// Detach from parent process
	// Only pass necessary environment variables to the daemon (security: M-1)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}
	if mavToken := os.Getenv("MAV_TOKEN"); mavToken != "" {
		cmd.Env = append(cmd.Env, "MAV_TOKEN="+mavToken)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MaestroVault daemon: %w", err)
	}

	// Release the process so it runs independently
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("failed to detach MaestroVault daemon: %w", err)
	}

	// Wait briefly for the daemon to initialize
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
	}

	return fmt.Errorf("MaestroVault daemon started but not responding after 5 seconds")
}
