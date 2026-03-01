package registry

import "errors"

var (
	// ErrNotRunning indicates the registry is not running
	ErrNotRunning = errors.New("registry not running")

	// ErrAlreadyRunning indicates the registry is already running
	ErrAlreadyRunning = errors.New("registry already running")

	// ErrPortInUse indicates the configured port is already in use
	ErrPortInUse = errors.New("port already in use")

	// ErrBinaryNotFound indicates the Zot binary was not found
	ErrBinaryNotFound = errors.New("binary not found")

	// ErrDownloadFailed indicates binary download failed
	ErrDownloadFailed = errors.New("download failed")

	// ErrInvalidConfig indicates configuration validation failed
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrProcessAlreadyRunning indicates the process is already running
	ErrProcessAlreadyRunning = errors.New("process already running")
)
