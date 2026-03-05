package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultBinaryManager implements BinaryManager for Zot.
type DefaultBinaryManager struct {
	binDir  string
	version string
}

// EnsureBinary ensures the binary exists, downloading if necessary.
func (b *DefaultBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	binaryPath := filepath.Join(b.binDir, "zot")

	// Check if binary already exists
	if _, err := os.Stat(binaryPath); err == nil {
		// Binary exists, verify it's executable
		if err := b.verifyExecutable(binaryPath); err != nil {
			// Re-download if not executable
			return b.downloadBinary(ctx, binaryPath)
		}
		return binaryPath, nil
	}

	// Binary doesn't exist, download it
	return b.downloadBinary(ctx, binaryPath)
}

// GetVersion returns the version of the currently installed binary.
func (b *DefaultBinaryManager) GetVersion(ctx context.Context) (string, error) {
	binaryPath := filepath.Join(b.binDir, "zot")

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("%w: %s", ErrBinaryNotFound, binaryPath)
	}

	// Run "zot version" command
	cmd := exec.CommandContext(ctx, binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	// Parse version from output
	// Expected format: "zot v1.4.3" or similar
	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "zot")
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")

	return version, nil
}

// NeedsUpdate checks if the binary needs to be updated.
func (b *DefaultBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	// Get current version
	currentVer, err := b.GetVersion(ctx)
	if err != nil {
		// If binary doesn't exist, we need to "update" (download)
		if errors.Is(err, ErrBinaryNotFound) {
			return true, nil
		}
		return false, err
	}

	// Compare versions (simple string comparison for now)
	// TODO: Use proper semver comparison
	if currentVer != b.version {
		return true, nil
	}

	return false, nil
}

// Update downloads and installs the latest version.
func (b *DefaultBinaryManager) Update(ctx context.Context) error {
	binaryPath := filepath.Join(b.binDir, "zot")
	backupPath := binaryPath + ".backup"

	// Backup existing binary if it exists
	if _, err := os.Stat(binaryPath); err == nil {
		if err := os.Rename(binaryPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing binary: %w", err)
		}
	}

	// Download new binary
	_, err := b.downloadBinary(ctx, binaryPath)
	if err != nil {
		// Rollback on failure
		if _, statErr := os.Stat(backupPath); statErr == nil {
			os.Rename(backupPath, binaryPath)
		}
		return fmt.Errorf("failed to update binary: %w", err)
	}

	// Keep backup for manual rollback - user can remove if desired

	return nil
}

// downloadBinary downloads the Zot binary for the current platform.
func (b *DefaultBinaryManager) downloadBinary(ctx context.Context, destPath string) (string, error) {
	// Ensure directory exists
	if err := os.MkdirAll(b.binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create binary directory: %w", err)
	}

	// Determine platform
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Construct download URL
	// Format: https://github.com/project-zot/zot/releases/download/v{version}/zot-{platform}-{arch}
	url := fmt.Sprintf(
		"https://github.com/project-zot/zot/releases/download/v%s/zot-%s-%s",
		b.version,
		platform,
		arch,
	)

	// Download the binary
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrDownloadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d from %s", ErrDownloadFailed, resp.StatusCode, url)
	}

	// Create temporary file
	tempPath := destPath + ".tmp"
	f, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(tempPath)
	}()

	// Download to temp file
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	// Close file before moving
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempPath, destPath); err != nil {
		return "", fmt.Errorf("failed to move binary to final location: %w", err)
	}

	// Ensure executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return "", fmt.Errorf("failed to set executable permissions: %w", err)
	}

	return destPath, nil
}

// verifyExecutable checks if the file is executable.
func (b *DefaultBinaryManager) verifyExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Check if executable bit is set
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("file is not executable")
	}

	return nil
}

// verifyChecksum verifies the SHA256 checksum of a file.
func (b *DefaultBinaryManager) verifyChecksum(path string, expectedSum string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}

	actualSum := hex.EncodeToString(hash.Sum(nil))
	if actualSum != expectedSum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSum, actualSum)
	}

	return nil
}
