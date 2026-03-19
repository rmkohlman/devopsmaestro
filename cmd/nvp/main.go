package main

import (
	"os"

	"github.com/rmkohlman/MaestroSDK/render"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	if err := Execute(); err != nil {
		// errSilent means the command already displayed the error via render
		if err.Error() != "" {
			render.ErrorToStderr(err.Error())
		}
		os.Exit(1)
	}
}
