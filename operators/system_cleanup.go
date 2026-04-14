package operators

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// DiskUsageCategory represents a category of disk usage.
type DiskUsageCategory struct {
	Type        string `json:"type"`
	Count       int    `json:"count"`
	Active      int    `json:"active"`
	Size        int64  `json:"size"`
	Reclaimable int64  `json:"reclaimable"`
}

// PruneResult holds the result of a prune operation.
type PruneResult struct {
	Type         string `json:"type"`
	BytesFreed   int64  `json:"bytes_freed"`
	ItemsRemoved int    `json:"items_removed"`
	DryRun       bool   `json:"dry_run"`
	Details      string `json:"details,omitempty"`
}

// SystemCleaner provides system cleanup operations against a container runtime.
// It wraps a Platform to determine which CLI tool (docker/nerdctl) to invoke.
type SystemCleaner struct {
	platform *Platform
}

// NewSystemCleaner creates a SystemCleaner for the given platform.
func NewSystemCleaner(platform *Platform) *SystemCleaner {
	return &SystemCleaner{platform: platform}
}

// cliTool returns the container CLI binary name for this platform.
func (sc *SystemCleaner) cliTool() string {
	if sc.platform.IsContainerd() {
		return "nerdctl"
	}
	return "docker"
}

// dockerDFEntry represents one row from `docker system df --format json`.
// Docker returns TotalCount and Active as strings, not integers.
type dockerDFEntry struct {
	Type        string `json:"Type"`
	TotalCount  string `json:"TotalCount"`
	Active      string `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"`
}

// GetRuntimeDiskUsage queries `docker system df` (or nerdctl) for runtime disk usage.
func (sc *SystemCleaner) GetRuntimeDiskUsage(ctx context.Context) ([]DiskUsageCategory, error) {
	tool := sc.cliTool()
	args := []string{"system", "df", "--format", "json"}
	if sc.platform.Type == PlatformColima && sc.platform.IsContainerd() {
		// nerdctl needs the address for containerd
		args = append([]string{"--address", sc.platform.SocketPath}, args...)
	}
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Env = sc.buildEnv()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run %s system df: %w", tool, err)
	}
	return sc.parseDockerDF(out)
}

// parseDockerDF parses the JSON lines output from `docker system df --format json`.
func (sc *SystemCleaner) parseDockerDF(data []byte) ([]DiskUsageCategory, error) {
	var categories []DiskUsageCategory
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry dockerDFEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // skip lines that don't parse
		}
		categories = append(categories, DiskUsageCategory{
			Type:        entry.Type,
			Count:       parseIntSafe(entry.TotalCount),
			Active:      parseIntSafe(entry.Active),
			Size:        parseDockerSize(entry.Size),
			Reclaimable: parseDockerSize(entry.Reclaimable),
		})
	}
	return categories, nil
}

// PruneBuildKit prunes the BuildKit cache. If dryRun is true, only reports what
// would be freed without actually pruning.
func (sc *SystemCleaner) PruneBuildKit(ctx context.Context, dryRun bool) (*PruneResult, error) {
	result := &PruneResult{Type: "BuildKit Cache", DryRun: dryRun}

	if dryRun {
		result.Details = fmt.Sprintf("Would run: %s buildx prune -f", sc.cliTool())
		return result, nil
	}

	tool := sc.cliTool()
	var cmd *exec.Cmd
	if tool == "docker" {
		cmd = exec.CommandContext(ctx, tool, "buildx", "prune", "-f")
	} else {
		cmd = exec.CommandContext(ctx, tool, "system", "prune", "--force")
		if sc.platform.IsContainerd() {
			cmd = exec.CommandContext(ctx, tool,
				"--address", sc.platform.SocketPath,
				"system", "prune", "--force")
		}
	}
	cmd.Env = sc.buildEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return result, fmt.Errorf("buildkit prune failed: %w (output: %s)", err, string(out))
	}
	result.Details = strings.TrimSpace(string(out))
	return result, nil
}

// ListDVMImages lists all container images with the "dvm-" prefix.
func (sc *SystemCleaner) ListDVMImages(ctx context.Context) ([]ImageInfo, error) {
	tool := sc.cliTool()
	args := []string{"images", "--format", "json", "--filter", "reference=dvm-*"}
	if sc.platform.IsContainerd() && tool == "nerdctl" {
		args = append([]string{"--address", sc.platform.SocketPath}, args...)
	}
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Env = sc.buildEnv()
	out, err := cmd.Output()
	if err != nil {
		// fallback: list all images and filter client-side
		return sc.listAllDVMImagesFallback(ctx)
	}
	return sc.parseDVMImages(out)
}

// ImageInfo represents a container image.
type ImageInfo struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       int64  `json:"size"`
	InUse      bool   `json:"in_use"`
}

// parseDVMImages parses JSON image output filtering for dvm- prefixed images.
func (sc *SystemCleaner) parseDVMImages(data []byte) ([]ImageInfo, error) {
	type dockerImage struct {
		ID         string `json:"ID"`
		Repository string `json:"Repository"`
		Tag        string `json:"Tag"`
		Size       string `json:"Size"`
	}
	var images []ImageInfo
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var img dockerImage
		if err := json.Unmarshal([]byte(line), &img); err != nil {
			continue
		}
		if strings.HasPrefix(img.Repository, "dvm-") {
			images = append(images, ImageInfo{
				ID:         img.ID,
				Repository: img.Repository,
				Tag:        img.Tag,
				Size:       parseDockerSize(img.Size),
			})
		}
	}
	return images, nil
}

// listAllDVMImagesFallback lists images using a broader query when filtering fails.
func (sc *SystemCleaner) listAllDVMImagesFallback(ctx context.Context) ([]ImageInfo, error) {
	tool := sc.cliTool()
	args := []string{"images", "--format", "json"}
	if sc.platform.IsContainerd() && tool == "nerdctl" {
		args = append([]string{"--address", sc.platform.SocketPath}, args...)
	}
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Env = sc.buildEnv()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	return sc.parseDVMImages(out)
}

// PruneImages removes unused dvm workspace images. activeContainerNames lists
// containers currently in use — images backing those containers are preserved.
func (sc *SystemCleaner) PruneImages(ctx context.Context, activeContainerNames []string, dryRun bool) (*PruneResult, error) {
	result := &PruneResult{Type: "DVM Workspace Images", DryRun: dryRun}

	images, err := sc.ListDVMImages(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to list dvm images: %w", err)
	}

	// Mark images that are in use
	sc.markInUseImages(images, activeContainerNames)

	var removable []ImageInfo
	for _, img := range images {
		if !img.InUse {
			removable = append(removable, img)
		}
	}

	result.ItemsRemoved = len(removable)
	for _, img := range removable {
		result.BytesFreed += img.Size
	}

	if dryRun {
		var names []string
		for _, img := range removable {
			names = append(names, fmt.Sprintf("%s:%s", img.Repository, img.Tag))
		}
		result.Details = fmt.Sprintf("Would remove %d image(s): %s",
			len(removable), strings.Join(names, ", "))
		return result, nil
	}

	// Actually remove images
	tool := sc.cliTool()
	for _, img := range removable {
		args := []string{"rmi", img.ID}
		if sc.platform.IsContainerd() && tool == "nerdctl" {
			args = append([]string{"--address", sc.platform.SocketPath}, args...)
		}
		rmCmd := exec.CommandContext(ctx, tool, args...)
		rmCmd.Env = sc.buildEnv()
		rmCmd.CombinedOutput() // best-effort
	}

	return result, nil
}

// markInUseImages marks images that back active containers as in-use.
func (sc *SystemCleaner) markInUseImages(images []ImageInfo, activeContainerNames []string) {
	activeSet := make(map[string]bool, len(activeContainerNames))
	for _, name := range activeContainerNames {
		activeSet[name] = true
	}
	for i := range images {
		// An image is "in use" if its repository matches an active container name
		if activeSet[images[i].Repository] {
			images[i].InUse = true
		}
	}
}

// buildEnv returns environment variables for the CLI tool subprocess.
func (sc *SystemCleaner) buildEnv() []string {
	env := os.Environ()
	if sc.platform.IsDockerCompatible() {
		env = append(env, fmt.Sprintf("DOCKER_HOST=unix://%s", sc.platform.SocketPath))
	}
	return env
}

// parseDockerSize parses a human-readable size string (e.g. "2.3GB", "512MB")
// into bytes. Returns 0 for unparseable strings.

// parseIntSafe converts a string to an int, returning 0 on failure.
func parseIntSafe(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}
func parseDockerSize(s string) int64 {
	// Strip parenthetical suffixes like "2.73GB (100%)"
	if idx := strings.Index(s, "("); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "0" || s == "0B" {
		return 0
	}
	s = strings.ToUpper(s)

	multipliers := []struct {
		suffix string
		mult   float64
	}{
		{"TB", 1e12},
		{"GB", 1e9},
		{"MB", 1e6},
		{"KB", 1e3},
		{"B", 1},
	}
	for _, m := range multipliers {
		if strings.HasSuffix(s, m.suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(s, m.suffix))
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0
			}
			return int64(val * m.mult)
		}
	}
	return 0
}
