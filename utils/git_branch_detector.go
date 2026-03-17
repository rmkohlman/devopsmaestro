package utils

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ParseDefaultBranch parses the output of `git ls-remote --symref <url> HEAD`
// and returns the default branch name.
//
// Expected output format:
//
//	ref: refs/heads/main	HEAD
//	abc123...	HEAD
//
// Returns the branch name (e.g., "main", "master", "develop").
func ParseDefaultBranch(output string) (string, error) {
	lines := strings.Split(output, "\n")
	prefix := "ref: refs/heads/"

	for _, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		// Line format: "ref: refs/heads/<branch>\tHEAD"
		// Extract everything between "refs/heads/" and the tab character
		rest := line[len(prefix):]
		tabIdx := strings.Index(rest, "\t")
		if tabIdx <= 0 {
			continue
		}

		branch := rest[:tabIdx]
		if branch != "" {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not detect default branch from git ls-remote output")
}

// DetectDefaultBranch runs `git ls-remote --symref <url> HEAD` and parses
// the output to determine the remote repository's default branch.
// Falls back to "main" if detection fails.
func DetectDefaultBranch(repoURL string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--symref", repoURL, "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "main"
	}

	branch, err := ParseDefaultBranch(string(out))
	if err != nil {
		return "main"
	}

	return branch
}
