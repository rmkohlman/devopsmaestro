package cmd

import (
	"context"
	"devopsmaestro/operators"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

var systemDFCmd = &cobra.Command{
	Use:   "df",
	Short: "Show disk usage breakdown for dvm and container runtime",
	Long: `Show a Docker-style disk usage breakdown across dvm-managed
directories and the container runtime.

Categories:
  Build Cache      ~/.devopsmaestro/build-cache
  Build Staging    ~/.devopsmaestro/build-staging
  Registries       ~/.devopsmaestro/registries
  Workspace Images Docker/containerd images with dvm- prefix

Examples:
  dvm system df
  dvm system df -o json
  dvm system df -o yaml`,
	RunE: runSystemDF,
}

// SystemDFData holds the full df report for JSON/YAML output.
type SystemDFData struct {
	Categories []DFCategory `json:"categories" yaml:"categories"`
}

// DFCategory is one row of the disk usage table.
type DFCategory struct {
	Type         string `json:"type" yaml:"type"`
	Count        int    `json:"count" yaml:"count"`
	Active       int    `json:"active" yaml:"active"`
	Size         string `json:"size" yaml:"size"`
	SizeBytes    int64  `json:"size_bytes" yaml:"size_bytes"`
	Reclaimable  string `json:"reclaimable" yaml:"reclaimable"`
	ReclaimBytes int64  `json:"reclaimable_bytes" yaml:"reclaimable_bytes"`
}

func init() {
	systemMaintCmd.AddCommand(systemDFCmd)
	AddOutputFlag(systemDFCmd, "")
}

func runSystemDF(cmd *cobra.Command, args []string) error {
	outputFmt, _ := cmd.Flags().GetString("output")

	var categories []DFCategory

	// 1. DVM-managed directories
	pc, pathErr := paths.Default()
	if pathErr == nil {
		root := pc.Root()
		categories = append(categories, dvmDirCategory("Build Cache", filepath.Join(root, "build-cache")))
		categories = append(categories, dvmDirCategory("Build Staging", filepath.Join(root, "build-staging")))
		categories = append(categories, dvmDirCategory("Registries", filepath.Join(root, "registries")))
	}

	// 2. Runtime-reported categories (workspace images, build cache, etc.)
	runtimeCategories := getRuntimeDFCategories()
	categories = append(categories, runtimeCategories...)

	data := SystemDFData{Categories: categories}

	if outputFmt == "json" || outputFmt == "yaml" {
		return render.OutputWith(outputFmt, data, render.Options{})
	}

	// Table output
	renderDFTable(categories)
	return nil
}

// dvmDirCategory builds a DFCategory from a local directory path.
func dvmDirCategory(name, dirPath string) DFCategory {
	size := dirSize(dirPath)
	count := countFiles(dirPath)
	return DFCategory{
		Type:         name,
		Count:        count,
		Active:       0,
		Size:         formatBytes(size),
		SizeBytes:    size,
		Reclaimable:  formatBytes(size),
		ReclaimBytes: size,
	}
}

// countFiles returns the number of files in a directory tree.
func countFiles(dirPath string) int {
	count := 0
	_ = filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

// getRuntimeDFCategories queries the container runtime for disk usage and
// returns DFCategory entries for images/build-cache/volumes.
func getRuntimeDFCategories() []DFCategory {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return nil
	}
	platform, err := detector.Detect()
	if err != nil || platform == nil {
		return nil
	}
	if !platform.IsReachable() {
		return nil
	}

	cleaner := operators.NewSystemCleaner(platform)
	ctx := context.Background()
	usage, err := cleaner.GetRuntimeDiskUsage(ctx)
	if err != nil {
		return nil
	}

	var categories []DFCategory
	for _, u := range usage {
		categories = append(categories, DFCategory{
			Type:         u.Type,
			Count:        u.Count,
			Active:       u.Active,
			Size:         formatBytes(u.Size),
			SizeBytes:    u.Size,
			Reclaimable:  formatBytes(u.Reclaimable),
			ReclaimBytes: u.Reclaimable,
		})
	}
	return categories
}

func renderDFTable(categories []DFCategory) {
	render.Blank()
	// Render the entire table through a single tabwriter so headers and
	// data rows share the same writer (no info-log "ℹ" prefix) and
	// columns align across all rows.
	w := tabwriter.NewWriter(render.GetWriter(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TYPE\tCOUNT\tACTIVE\tSIZE\tRECLAIMABLE")
	fmt.Fprintln(w, "----\t-----\t------\t----\t-----------")
	for _, c := range categories {
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n",
			c.Type, c.Count, c.Active, c.Size, c.Reclaimable)
	}
	_ = w.Flush()
	render.Blank()
}
