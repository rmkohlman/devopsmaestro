package registry

import (
	"context"
	"sync"
	"time"
)

// BaseServiceManager provides common lifecycle management for all registry services.
// It handles process management, idle timers, and shared utility operations.
// This struct is designed to be embedded in specific manager implementations.
type BaseServiceManager struct {
	binaryManager  BinaryManager
	processManager ProcessManager

	mu             sync.RWMutex
	startTime      time.Time
	idleTimer      *time.Timer
	lastAccessTime time.Time
}

// NewBaseServiceManager creates a new BaseServiceManager with injected dependencies.
func NewBaseServiceManager(binary BinaryManager, process ProcessManager) BaseServiceManager {
	return BaseServiceManager{
		binaryManager:  binary,
		processManager: process,
	}
}

// RecordStart records the start time.
func (b *BaseServiceManager) RecordStart() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.RecordStartLocked()
}

// RecordStartLocked sets start time. Caller must hold mu.
func (b *BaseServiceManager) RecordStartLocked() {
	now := time.Now()
	b.startTime = now
	b.lastAccessTime = now
}

// GetUptime returns the duration since the service started.
func (b *BaseServiceManager) GetUptime() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.startTime.IsZero() {
		return 0
	}

	return time.Since(b.startTime)
}

// IsProcessRunning checks if the underlying process is running.
func (b *BaseServiceManager) IsProcessRunning() bool {
	return b.processManager.IsRunning()
}

// GetPID returns the process ID, or 0 if not running.
func (b *BaseServiceManager) GetPID() int {
	return b.processManager.GetPID()
}

// StopProcess stops the process and cleans up the idle timer.
func (b *BaseServiceManager) StopProcess(ctx context.Context) error {
	// Stop the idle timer first
	b.StopIdleTimer()

	// Stop the process
	return b.processManager.Stop(ctx)
}

// EnsureBinary ensures the service binary exists, downloading if necessary.
func (b *BaseServiceManager) EnsureBinary(ctx context.Context) (string, error) {
	return b.binaryManager.EnsureBinary(ctx)
}

// GetVersion returns the version of the installed binary.
func (b *BaseServiceManager) GetVersion(ctx context.Context) (string, error) {
	return b.binaryManager.GetVersion(ctx)
}

// SetupIdleTimer configures auto-shutdown for on-demand mode.
func (b *BaseServiceManager) SetupIdleTimer(lifecycle string, timeout time.Duration, stopFunc func()) {
	if lifecycle != "on-demand" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.ResetIdleTimerLocked(lifecycle, timeout, stopFunc)
}

// ResetIdleTimer resets the idle timer, extending the shutdown deadline.
func (b *BaseServiceManager) ResetIdleTimer(lifecycle string, timeout time.Duration, stopFunc func()) {
	if lifecycle != "on-demand" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.ResetIdleTimerLocked(lifecycle, timeout, stopFunc)
}

// ResetIdleTimerLocked resets idle timer. Caller must hold mu.
func (b *BaseServiceManager) ResetIdleTimerLocked(lifecycle string, timeout time.Duration, stopFunc func()) {
	if lifecycle != "on-demand" {
		return
	}
	b.lastAccessTime = time.Now()
	if b.idleTimer != nil {
		b.idleTimer.Stop()
	}
	b.idleTimer = time.AfterFunc(timeout, func() {
		go stopFunc() // Launch in goroutine to avoid blocking on mutex
	})
}

// StopIdleTimer stops the idle timer without triggering shutdown.
func (b *BaseServiceManager) StopIdleTimer() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.StopIdleTimerLocked()
}

// StopIdleTimerLocked stops idle timer. Caller must hold mu.
func (b *BaseServiceManager) StopIdleTimerLocked() {
	if b.idleTimer != nil {
		b.idleTimer.Stop()
		b.idleTimer = nil
	}
}
