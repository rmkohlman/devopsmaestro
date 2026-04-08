package shellgen

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectShell parses the $SHELL environment variable and returns the detected
// shell name. Returns "zsh" as default if detection fails or the shell is
// not recognized.
//
// Supported return values: "bash", "zsh", "fish".
func DetectShell() string {
	return detectShellFromEnv(os.Getenv("SHELL"))
}

// detectShellFromEnv is the testable core of DetectShell.
func detectShellFromEnv(shellEnv string) string {
	if shellEnv == "" {
		return "zsh"
	}

	base := filepath.Base(shellEnv)
	base = strings.TrimSpace(base)

	switch base {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	default:
		return "zsh"
	}
}
