package cmd

import (
	"devopsmaestro/ui"
	"fmt"
	"runtime"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// Version information (injected at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

// VersionInfo holds structured version data for JSON/YAML output.
type VersionInfo struct {
	Version   string `json:"version" yaml:"version"`
	Commit    string `json:"commit" yaml:"commit"`
	BuildTime string `json:"buildTime" yaml:"buildTime"`
	Go        string `json:"go" yaml:"go"`
	Platform  string `json:"platform" yaml:"platform"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Print the version, build time, and commit hash of dvm.

Examples:
  dvm version
  dvm version --short
  dvm version -o json
  dvm version -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		short, _ := cmd.Flags().GetBool("short")
		outputFormat, _ := cmd.Flags().GetString("output")

		// Add 'v' prefix only if not already present
		versionDisplay := Version
		if len(versionDisplay) > 0 && versionDisplay[0] != 'v' {
			versionDisplay = "v" + versionDisplay
		}

		if short {
			fmt.Println(versionDisplay)
			return nil
		}

		// Structured output for JSON/YAML
		if outputFormat == "json" || outputFormat == "yaml" {
			info := VersionInfo{
				Version:   versionDisplay,
				Commit:    Commit,
				BuildTime: BuildTime,
				Go:        runtime.Version(),
				Platform:  runtime.GOOS + "/" + runtime.GOARCH,
			}
			return render.OutputWith(outputFormat, info, render.Options{})
		}

		// Default: beautiful colored output using UI styles
		fmt.Println()
		fmt.Printf("%s %s\n",
			ui.HeaderStyle.Render("🚀 DevOpsMaestro (dvm)"),
			ui.VersionStyle.Render(versionDisplay))
		fmt.Println()
		fmt.Printf("  %s  %s\n",
			ui.MutedStyle.Render("Commit:    "),
			ui.PathStyle.Render(Commit))
		fmt.Printf("  %s  %s\n",
			ui.MutedStyle.Render("Built:     "),
			ui.DateStyle.Render(BuildTime))
		fmt.Printf("  %s  %s\n",
			ui.MutedStyle.Render("Go:        "),
			ui.InfoStyle.Render(runtime.Version()))
		fmt.Printf("  %s  %s\n",
			ui.MutedStyle.Render("Platform:  "),
			ui.TextStyle.Render(runtime.GOOS+"/"+runtime.GOARCH))
		fmt.Println()
		return nil
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "Print only version number")
	AddOutputFlag(versionCmd, "")
	rootCmd.AddCommand(versionCmd)
}
