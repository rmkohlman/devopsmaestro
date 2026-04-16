package registry

import (
	"context"
	"errors"
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
	config  ProcessConfig
	cmd     *exec.Cmd
	pid     int
	logFile *os.File
	mu      sync.RWMutex
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
		// cleanStalePIDFile may have adopted a running process —
		// re-check before spawning a duplicate. However, verify the
		// adopted process is actually alive with a fresh signal check
		// to avoid false positives from PID reuse (#385).
		if p.pid > 0 {
			if isProcessAlive(p.pid) {
				return fmt.Errorf("%w", ErrProcessAlreadyRunning)
			}
			// Process died between adoption and check — clean up and continue
			p.pid = 0
			os.Remove(config.PIDFile)
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
		if err := os.MkdirAll(filepath.Dir(config.LogFile), 0700); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFileMode)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
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
		// Close log file on start failure
		if logFile != nil {
			logFile.Close()
		}
		// Check if binary not found
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("binary not found: %w", err)
		}
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Store the log file so it remains open for the child process
	p.logFile = logFile

	// Store PID
	p.pid = p.cmd.Process.Pid

	// Write PID file if specified
	if config.PIDFile != "" {
		if err := p.writePIDFile(config.PIDFile, p.pid); err != nil {
			// Kill the process we just started
			p.cmd.Process.Kill()
			// Close the log file since we're aborting
			if p.logFile != nil {
				p.logFile.Close()
				p.logFile = nil
			}
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

	// If we have the in-memory cmd, use it directly
	if p.cmd != nil && p.cmd.Process != nil {
		return p.stopInMemoryProcess(ctx)
	}

	// Fallback: stop process discovered via PID file
	return p.stopFromPIDFile(ctx)
}

// stopInMemoryProcess stops a process we spawned in this CLI invocation.
func (p *DefaultProcessManager) stopInMemoryProcess(ctx context.Context) error {
	// Send SIGTERM first
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		if !errors.Is(err, os.ErrProcessDone) {
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
		<-done
	case <-ctx.Done():
		p.cmd.Process.Signal(syscall.SIGKILL)
		return ctx.Err()
	}

	// Close the log file now that the process has exited
	if p.logFile != nil {
		p.logFile.Close()
		p.logFile = nil
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

// stopFromPIDFile stops a process discovered via PID file (from a previous CLI invocation).
func (p *DefaultProcessManager) stopFromPIDFile(ctx context.Context) error {
	pid := p.readPIDFileLocked()
	if pid <= 0 {
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Send SIGTERM first
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if !errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
		}
		// Process already gone, just clean up
		p.removePIDFile()
		return nil
	}

	// Wait for process to exit or timeout
	timeout := p.config.ShutdownTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			// Force kill
			if err := proc.Signal(syscall.SIGKILL); err == nil {
				// Wait briefly for SIGKILL to take effect
				time.Sleep(200 * time.Millisecond)
			}
			p.removePIDFile()
			return nil
		case <-ctx.Done():
			proc.Signal(syscall.SIGKILL)
			p.removePIDFile()
			return ctx.Err()
		case <-ticker.C:
			if !isProcessAlive(pid) {
				p.removePIDFile()
				return nil
			}
		}
	}
}

// removePIDFile removes the configured PID file and closes the log file if open.
func (p *DefaultProcessManager) removePIDFile() {
	// Close the log file if it's still open
	if p.logFile != nil {
		p.logFile.Close()
		p.logFile = nil
	}
	if p.config.PIDFile != "" {
		os.Remove(p.config.PIDFile)
	}
	p.pid = 0
}

// isRunningLocked checks if the process is running. Caller must hold at least p.mu.RLock.
func (p *DefaultProcessManager) isRunningLocked() bool {
	// If we have the in-memory cmd, check it directly
	if p.cmd != nil && p.cmd.Process != nil {
		// Check if process has already been waited on
		if p.cmd.ProcessState != nil {
			return false
		}

		err := p.cmd.Process.Signal(syscall.Signal(0))
		if err != nil {
			return false
		}

		// Signal succeeded, but process might be a zombie
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(p.cmd.Process.Pid, &status, syscall.WNOHANG, nil)
		if err != nil {
			if err == syscall.ECHILD {
				return false
			}
			return true
		}
		if pid == p.cmd.Process.Pid {
			return false
		}
		return true
	}

	// Fallback: check PID file for processes started by a previous CLI invocation
	return p.isRunningFromPIDFile()
}

// isRunningFromPIDFile checks if a process is alive by reading the PID file.
func (p *DefaultProcessManager) isRunningFromPIDFile() bool {
	pid := p.readPIDFileLocked()
	if pid <= 0 {
		return false
	}
	return isProcessAlive(pid)
}

// readPIDFileLocked reads the PID from the configured PID file. Returns 0 if not found/invalid.
// Caller must hold at least p.mu.RLock.
func (p *DefaultProcessManager) readPIDFileLocked() int {
	if p.config.PIDFile == "" {
		return 0
	}
	data, err := os.ReadFile(p.config.PIDFile)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return pid
}

// isProcessAlive checks if a process with the given PID is running.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
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
	if p.cmd != nil && p.cmd.Process != nil {
		return p.pid
	}
	// Fallback: read from PID file
	return p.readPIDFileLocked()
}

// writePIDFile writes the process ID to a file.
func (p *DefaultProcessManager) writePIDFile(path string, pid int) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Write PID — 0600: PID files should be owner-only
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0600)
}

// cleanStalePIDFile removes a PID file if the process is no longer running.
// If the process IS running, it adopts it (stores the PID and config) rather
// than returning an error. The caller (Zot/Athens/etc. manager) is responsible
// for deciding whether the running process is the correct service via port
// availability and health probes.
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
		// Invalid PID file, remove it.
		// Tolerate ENOENT — concurrent goroutine may have already removed it (#364).
		if rmErr := os.Remove(path); rmErr != nil && !os.IsNotExist(rmErr) {
			return rmErr
		}
		return nil
	}

	// Check if process is running
	if !isProcessAlive(pid) {
		// Process doesn't exist, remove stale file.
		// Tolerate ENOENT — a concurrent goroutine may have already removed
		// the same PID file (race condition under high parallelism, see #364).
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	// Process exists — adopt it by recording the PID so that IsRunning()
	// returns true and the manager layer can short-circuit before trying
	// to spawn a duplicate.
	p.pid = pid
	p.config.PIDFile = path
	return nil
}
