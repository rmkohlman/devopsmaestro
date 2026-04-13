package cmd

import (
	"bufio"
	"crypto/rand"
	"devopsmaestro/models"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rmkohlman/MaestroSDK/render"
	"golang.org/x/term"
)

// generateSandboxName creates a unique sandbox container name.
// Format: dvm-sandbox-<lang>-<shortid> (e.g., dvm-sandbox-python-a3f2)
func generateSandboxName(lang string) string {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a fixed suffix on error
		return fmt.Sprintf("dvm-sandbox-%s-0000", lang)
	}
	return fmt.Sprintf("dvm-sandbox-%s-%s", lang, hex.EncodeToString(b))
}

// resolveVersion determines the language version to use.
// Priority: --version flag > interactive picker (if TTY) > default version.
func resolveVersion(preset models.SandboxPreset) string {
	if isTTY() {
		return pickVersionInteractive(preset)
	}
	return preset.DefaultVersion
}

// isValidVersion checks if the given version is in the preset's version list.
func isValidVersion(preset models.SandboxPreset, version string) bool {
	for _, v := range preset.Versions {
		if v == version {
			return true
		}
	}
	return false
}

// isTTY returns true if stdin is a terminal.
func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// pickVersionInteractive shows a numbered list of versions and lets the user pick.
func pickVersionInteractive(preset models.SandboxPreset) string {
	fmt.Fprintf(os.Stderr, "Select %s version:\n", preset.Language)
	for i, v := range preset.Versions {
		marker := "  "
		suffix := ""
		if v == preset.DefaultVersion {
			marker = "▸ "
			suffix = " (default)"
		}
		fmt.Fprintf(os.Stderr, "  %s%d) %s%s\n", marker, i+1, v, suffix)
	}

	fmt.Fprintf(os.Stderr, "\nEnter number [1]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return preset.DefaultVersion
	}

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(preset.Versions) {
		render.Warningf("Invalid selection %q, using default %s", input, preset.DefaultVersion)
		return preset.DefaultVersion
	}

	return preset.Versions[num-1]
}

// buildSandboxLabels creates the container labels for a sandbox.
func buildSandboxLabels(lang, version, name string) map[string]string {
	return map[string]string{
		"dvm.sandbox":         "true",
		"dvm.sandbox.lang":    lang,
		"dvm.sandbox.version": version,
		"dvm.sandbox.name":    name,
	}
}
