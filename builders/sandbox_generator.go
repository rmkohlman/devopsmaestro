package builders

import (
	"devopsmaestro/models"
	"fmt"
	"path/filepath"
	"strings"
)

// GenerateSandboxDockerfile creates a lightweight Dockerfile for an ephemeral sandbox.
// It produces a minimal image with the language runtime, essential tools, a non-root
// user, and optionally pre-installed dependencies.
//
// Parameters:
//   - preset: the language preset defining base image and install commands
//   - version: the language version to use
//   - depsFile: optional path to a dependency file (empty string to skip)
func GenerateSandboxDockerfile(preset models.SandboxPreset, version string, depsFile string) string {
	var b strings.Builder

	// Base image
	baseImage := preset.BaseImage(version)
	fmt.Fprintf(&b, "FROM %s\n\n", baseImage)

	// Essential packages — keep minimal
	b.WriteString("# Essential tools\n")
	b.WriteString("RUN apt-get update && apt-get install -y --no-install-recommends \\\n")
	b.WriteString("    git curl ca-certificates build-essential \\\n")
	b.WriteString("    && rm -rf /var/lib/apt/lists/*\n\n")

	// Create non-root dev user
	b.WriteString("# Create non-root dev user\n")
	b.WriteString("RUN groupadd -g 1000 dev && useradd -m -u 1000 -g dev -s /bin/bash dev\n\n")

	// Set up working directory
	b.WriteString("# Working directory\n")
	b.WriteString("WORKDIR /sandbox\n")
	b.WriteString("RUN chown dev:dev /sandbox\n\n")

	// Dependency file handling
	if depsFile != "" {
		baseName := filepath.Base(depsFile)
		fmt.Fprintf(&b, "# Install dependencies\n")
		fmt.Fprintf(&b, "COPY %s /sandbox/%s\n", baseName, baseName)

		installCmd := preset.DepsInstallCmd
		if strings.Contains(installCmd, "%s") {
			installCmd = fmt.Sprintf(installCmd, baseName)
		}
		fmt.Fprintf(&b, "RUN %s\n\n", installCmd)
	}

	// Switch to dev user
	b.WriteString("USER dev\n")
	b.WriteString("CMD [\"/bin/bash\"]\n")

	return b.String()
}
