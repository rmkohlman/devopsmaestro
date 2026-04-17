package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
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

// AthensBinaryManager implements BinaryManager for Athens.
type AthensBinaryManager struct {
	binDir  string
	version string
}

// NewAthensBinaryManager creates a new AthensBinaryManager.
func NewAthensBinaryManager(binDir, version string) *AthensBinaryManager {
	return &AthensBinaryManager{
		binDir:  binDir,
		version: version,
	}
}

// EnsureBinary ensures the binary exists, downloading if necessary.
func (b *AthensBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	binaryPath := filepath.Join(b.binDir, "athens")

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
func (b *AthensBinaryManager) GetVersion(ctx context.Context) (string, error) {
	binaryPath := filepath.Join(b.binDir, "athens")

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("%w: %s", ErrBinaryNotFound, binaryPath)
	}

	// Run "athens --version" command
	cmd := exec.CommandContext(ctx, binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		// Athens might not support --version, return configured version
		return b.version, nil
	}

	// Parse version from output — Athens may return multi-line build details
	version := sanitizeVersion(string(output))
	if version == "" {
		return b.version, nil
	}

	return version, nil
}

// NeedsUpdate checks if the binary needs to be updated.
func (b *AthensBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
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
	if currentVer != b.version {
		return true, nil
	}

	return false, nil
}

// Update downloads and installs the latest version.
func (b *AthensBinaryManager) Update(ctx context.Context) error {
	binaryPath := filepath.Join(b.binDir, "athens")
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

	return nil
}

// downloadBinary downloads the Athens binary for the current platform.
func (b *AthensBinaryManager) downloadBinary(ctx context.Context, destPath string) (string, error) {
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
	// Format: https://github.com/gomods/athens/releases/download/v{version}/athens_{version}_{os}_{arch}.tar.gz
	url := fmt.Sprintf(
		"https://github.com/gomods/athens/releases/download/v%s/athens_%s_%s_%s.tar.gz",
		b.version,
		b.version,
		platform,
		arch,
	)

	// Download the archive
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

	// Extract from tar.gz
	if err := b.extractTarGz(resp.Body, destPath); err != nil {
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	// Ensure executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return "", fmt.Errorf("failed to set executable permissions: %w", err)
	}

	return destPath, nil
}

// extractTarGz extracts the athens binary from a tar.gz archive.
func (b *AthensBinaryManager) extractTarGz(r io.Reader, destPath string) error {
	// Create gzip reader
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Find and extract the athens binary
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the athens binary (might be in a subdirectory)
		if strings.HasSuffix(header.Name, "/athens") || header.Name == "athens" {
			// Create destination file
			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer f.Close()

			// Copy binary
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("athens binary not found in archive")
}

// verifyExecutable checks if the file is executable.
func (b *AthensBinaryManager) verifyExecutable(path string) error {
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
