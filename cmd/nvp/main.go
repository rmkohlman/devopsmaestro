package main

import (
	"os"

	"devopsmaestro/render"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	if err := Execute(); err != nil {
		render.ErrorToStderr(err.Error())
		os.Exit(1)
	}
}
