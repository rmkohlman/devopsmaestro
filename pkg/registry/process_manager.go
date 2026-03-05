package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// DefaultProcessManager implements ProcessManager using os/exec.
type DefaultProcessManager struct {
	config ProcessConfig
	cmd    *exec.Cmd
	pid    int
	mu     sync.RWMutex
}

// Start spawns a new process with the given binary and arguments.
func (p *DefaultProcessManager) Start(ctx context.Context, binary string, args []string, config ProcessConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check context first
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Check if already running
	if p.isRunningLocked() {
		return fmt.Errorf("%w", ErrProcessAlreadyRunning)
	}

	// Check for stale PID file and clean it up
	if config.PIDFile != "" {
		if err := p.cleanStalePIDFile(config.PIDFile); err != nil {
			return fmt.Errorf("failed to clean stale PID file: %w", err)
		}
	}

	// Update config
	p.config = config

	// Set default shutdown timeout if not specified
	if p.config.ShutdownTimeout == 0 {
		p.config.ShutdownTimeout = 10 * time.Second
	}

	// Create log file if specified
	var logFile *os.File
	var err error
	if config.LogFile != "" {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(config.LogFile), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		defer logFile.Close()
	}

	// Create command
	p.cmd = exec.CommandContext(ctx, binary, args...)

	// Set working directory if specified
	if config.WorkingDir != "" {
		p.cmd.Dir = config.WorkingDir
	}

	// Redirect output to log file
	if logFile != nil {
		p.cmd.Stdout = logFile
		p.cmd.Stderr = logFile
	}

	// Start the process
	if err := p.cmd.Start(); err != nil {
		// Check if binary not found
		if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("binary not found: %w", err)
		}
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Store PID
	p.pid = p.cmd.Process.Pid

	// Write PID file if specified
	if config.PIDFile != "" {
		if err := p.writePIDFile(config.PIDFile, p.pid); err != nil {
			// Kill the process we just started
			p.cmd.Process.Kill()
			return fmt.Errorf("failed to write PID file: %w", err)
		}
	}

	return nil
}

// Stop stops the process gracefully.
func (p *DefaultProcessManager) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If not running, this is idempotent
	if !p.isRunningLocked() {
		return nil
	}

	// Send SIGTERM first
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// Process might have already exited
		if !strings.Contains(err.Error(), "process already finished") {
			return fmt.Errorf("failed to send SIGTERM: %w", err)
		}
	}

	// Wait for process to exit gracefully or timeout
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	timeout := p.config.ShutdownTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	select {
	case <-done:
		// Process exited gracefully
	case <-time.After(timeout):
		// Force kill
		if err := p.cmd.Process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to send SIGKILL: %w", err)
		}
		<-done // Wait for process to actually die
	case <-ctx.Done():
		// Context cancelled - force kill immediately
		p.cmd.Process.Signal(syscall.SIGKILL)
		return ctx.Err()
	}

	// Remove PID file
	if p.config.PIDFile != "" {
		os.Remove(p.config.PIDFile)
	}

	// Reset state
	p.pid = 0
	p.cmd = nil

	return nil
}

// isRunningLocked checks if the process is running. Caller must hold at least p.mu.RLock.
func (p *DefaultProcessManager) isRunningLocked() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	// Check if process has already been waited on
	if p.cmd.ProcessState != nil {
		// Process has exited and Wait() was called
		return false
	}

	// Try to get process state without blocking
	// Send signal 0 to check if process exists
	err := p.cmd.Process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist or we can't signal it
		return false
	}

	// Signal succeeded, but process might be a zombie
	// Try a non-blocking wait to reap the zombie
	var status syscall.WaitStatus
	pid, err := syscall.Wait4(p.cmd.Process.Pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		// If we get ECHILD, the process doesn't exist or isn't our child
		if err == syscall.ECHILD {
			return false
		}
		// Other errors - assume process is still running
		return true
	}

	if pid == p.cmd.Process.Pid {
		// Process has exited (we just reaped it)
		return false
	}

	// pid == 0 means process is still running (WNOHANG returned immediately)
	return true
}

// IsRunning checks if the process is currently running.
func (p *DefaultProcessManager) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isRunningLocked()
}

// GetPID returns the process ID, or 0 if not running.
func (p *DefaultProcessManager) GetPID() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.pid
}

// writePIDFile writes the process ID to a file.
func (p *DefaultProcessManager) writePIDFile(path string, pid int) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write PID
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

// cleanStalePIDFile removes a PID file if the process is no longer running.
func (p *DefaultProcessManager) cleanStalePIDFile(path string) error {
	// Read PID file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file, nothing to clean
		}
		return err
	}

	// Parse PID
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		// Invalid PID file, remove it
		return os.Remove(path)
	}

	// Check if process is running
	proc, err := os.FindProcess(pid)
	if err != nil {
		// Process doesn't exist, remove stale file
		return os.Remove(path)
	}

	// Send signal 0 to check if process exists
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist, remove stale file
		return os.Remove(path)
	}

	// Process exists - this is a problem
	return fmt.Errorf("process %d is already running", pid)
}
