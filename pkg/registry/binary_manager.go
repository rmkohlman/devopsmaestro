package registry

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// maxBinarySize is the maximum allowed size for downloaded binaries (500 MB).
const maxBinarySize = 500 * 1024 * 1024

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

		// Check if installed version matches desired version
		needsUpdate, err := b.NeedsUpdate(ctx)
		if err != nil {
			// Can't determine installed version — re-download to ensure correct version
			if updateErr := b.Update(ctx); updateErr != nil {
				return "", fmt.Errorf("failed to update binary to version %s: %w", b.version, updateErr)
			}
			return binaryPath, nil
		}
		if needsUpdate {
			// Version mismatch — download the correct version
			if err := b.Update(ctx); err != nil {
				return "", fmt.Errorf("failed to download binary version %s: %w", b.version, err)
			}
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

	// Zot outputs version info as JSON to stderr via --version flag
	cmd := exec.CommandContext(ctx, binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	// Parse JSON to extract commit field: "v2.1.15-0-gace12e2"
	var versionInfo struct {
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		return "", fmt.Errorf("failed to parse version output: %w", err)
	}

	// Extract semver from commit: "v2.1.15-0-gace12e2" -> "2.1.15"
	commit := strings.TrimPrefix(versionInfo.Commit, "v")
	// Strip git describe suffix: "2.1.15-0-gace12e2" -> "2.1.15"
	if idx := strings.Index(commit, "-"); idx != -1 {
		commit = commit[:idx]
	}

	if commit == "" {
		return "", fmt.Errorf("could not extract version from commit field: %s", versionInfo.Commit)
	}

	return commit, nil
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
	// Apply defensive timeout if caller didn't set a deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
	}

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

	// Check Content-Length against maximum allowed size
	if resp.ContentLength > maxBinarySize {
		return "", fmt.Errorf("binary size %d exceeds maximum allowed %d", resp.ContentLength, maxBinarySize)
	}

	// Limit actual read to maxBinarySize
	limitedBody := io.LimitReader(resp.Body, maxBinarySize)

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
	if _, err := io.Copy(f, limitedBody); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}

	// Close file before verification and moving
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Verify integrity
	expectedSum, err := b.fetchChecksum(ctx, url)
	if err != nil {
		return "", fmt.Errorf("checksum fetch failed: %w", err)
	}
	if err := b.verifyChecksum(tempPath, expectedSum); err != nil {
		return "", fmt.Errorf("integrity verification failed: %w", err)
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
	if subtle.ConstantTimeCompare([]byte(actualSum), []byte(expectedSum)) != 1 {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSum, actualSum)
	}

	return nil
}

// fetchChecksum downloads the checksum file for the given binary URL and returns the expected SHA256 hex digest.
func (b *DefaultBinaryManager) fetchChecksum(ctx context.Context, binaryURL string) (string, error) {
	checksumURL := binaryURL + ".sha256"
	req, err := http.NewRequestWithContext(ctx, "GET", checksumURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create checksum request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download checksum: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum download failed with status %d", resp.StatusCode)
	}
	// Read at most 256 bytes (SHA256 hex is 64 chars + optional filename)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", fmt.Errorf("failed to read checksum: %w", err)
	}
	// Parse: could be just the hex digest, or "hexdigest  filename"
	parts := strings.Fields(strings.TrimSpace(string(body)))
	if len(parts) == 0 {
		return "", fmt.Errorf("empty checksum file")
	}
	return parts[0], nil
}
