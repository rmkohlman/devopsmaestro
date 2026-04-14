package cmd

import (
	"devopsmaestro/operators"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

var systemInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system overview with runtime, platform, and disk usage",
	Long: `Show a comprehensive overview of the dvm system including:

- Platform type (OrbStack, Colima, Docker Desktop, Podman)
- DVM version
- Disk usage for ~/.devopsmaestro/ subdirectories
- Container runtime status

Examples:
  dvm system info
  dvm system info -o json
  dvm system info -o yaml`,
	RunE: runSystemInfo,
}

// SystemInfoData holds structured system info for JSON/YAML output.
type SystemInfoData struct {
	Version    string          `json:"version" yaml:"version"`
	Platform   string          `json:"platform" yaml:"platform"`
	Runtime    string          `json:"runtime" yaml:"runtime"`
	Status     string          `json:"status" yaml:"status"`
	GoVersion  string          `json:"go_version" yaml:"go_version"`
	Arch       string          `json:"arch" yaml:"arch"`
	DiskUsage  []DirUsageEntry `json:"disk_usage" yaml:"disk_usage"`
	TotalUsage string          `json:"total_usage" yaml:"total_usage"`
}

// DirUsageEntry holds disk usage for a single directory.
type DirUsageEntry struct {
	Name string `json:"name" yaml:"name"`
	Path string `json:"path" yaml:"path"`
	Size string `json:"size" yaml:"size"`
}

func init() {
	systemMaintCmd.AddCommand(systemInfoCmd)
	AddOutputFlag(systemInfoCmd, "")
}

func runSystemInfo(cmd *cobra.Command, args []string) error {
	outputFmt, _ := cmd.Flags().GetString("output")

	// Detect platform
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize platform detector: %w", err)
	}

	platformName := "unknown"
	runtimeStatus := "not found"
	runtimeType := "unknown"

	platform, detectErr := detector.Detect()
	if detectErr == nil && platform != nil {
		platformName = platform.Name
		if platform.IsReachable() {
			runtimeStatus = "running"
		} else {
			runtimeStatus = "stopped"
		}
		if platform.IsContainerd() {
			runtimeType = "containerd"
		} else {
			runtimeType = "docker"
		}
	}

	// Gather disk usage
	pc, pathErr := paths.Default()
	var dirEntries []DirUsageEntry
	var totalBytes int64
	if pathErr == nil {
		root := pc.Root()
		dirs := []struct{ name, sub string }{
			{"build-cache", "build-cache"},
			{"build-staging", "build-staging"},
			{"registries", "registries"},
			{"repos", "repos"},
			{"workspaces", "workspaces"},
		}
		for _, d := range dirs {
			p := filepath.Join(root, d.sub)
			size := dirSize(p)
			totalBytes += size
			dirEntries = append(dirEntries, DirUsageEntry{
				Name: d.name,
				Path: p,
				Size: formatBytes(size),
			})
		}
	}

	versionDisplay := Version
	if len(versionDisplay) > 0 && versionDisplay[0] != 'v' {
		versionDisplay = "v" + versionDisplay
	}

	info := SystemInfoData{
		Version:    versionDisplay,
		Platform:   platformName,
		Runtime:    runtimeType,
		Status:     runtimeStatus,
		GoVersion:  runtime.Version(),
		Arch:       runtime.GOOS + "/" + runtime.GOARCH,
		DiskUsage:  dirEntries,
		TotalUsage: formatBytes(totalBytes),
	}

	// JSON/YAML
	if outputFmt == "json" || outputFmt == "yaml" {
		return render.OutputWith(outputFmt, info, render.Options{})
	}

	// Human-readable output
	renderSystemInfo(info, platform)
	return nil
}

func renderSystemInfo(info SystemInfoData, platform *operators.Platform) {
	render.Blank()
	render.Info("Runtime Information")
	render.Info(fmt.Sprintf("  DVM Version:  %s", info.Version))
	render.Info(fmt.Sprintf("  Platform:     %s", info.Platform))
	render.Info(fmt.Sprintf("  Runtime:      %s", info.Runtime))
	if info.Status == "running" {
		render.Success(fmt.Sprintf("  Status:       %s", info.Status))
	} else {
		render.Warning(fmt.Sprintf("  Status:       %s", info.Status))
	}
	render.Info(fmt.Sprintf("  Go:           %s", info.GoVersion))
	render.Info(fmt.Sprintf("  Architecture: %s", info.Arch))

	render.Blank()
	render.Info("Disk Usage (~/.devopsmaestro)")
	for _, d := range info.DiskUsage {
		render.Info(fmt.Sprintf("  %-16s %s", d.Name, d.Size))
	}
	render.Info(fmt.Sprintf("  %-16s %s", "TOTAL", info.TotalUsage))

	render.Blank()

	// Provide quick actions based on status
	if platform != nil && info.Status == "stopped" {
		render.Warning("Quick Actions:")
		render.Info(fmt.Sprintf("  %s", platform.GetStartHint()))
	} else if info.Status == "running" {
		render.Info("Quick Actions:")
		render.Info("  dvm system df              # Detailed disk breakdown")
		render.Info("  dvm system prune --dry-run # Preview cleanup")
	}
	render.Blank()
}
