package cmd

import (
	"devopsmaestro/ui"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information (injected at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, build time, and commit hash of dvm.`,
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")

		// Add 'v' prefix only if not already present
		versionDisplay := Version
		if len(versionDisplay) > 0 && versionDisplay[0] != 'v' {
			versionDisplay = "v" + versionDisplay
		}

		if short {
			fmt.Println(versionDisplay)
			return
		}

		// Beautiful version output using UI styles
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
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "Print only version number")
	rootCmd.AddCommand(versionCmd)
}
