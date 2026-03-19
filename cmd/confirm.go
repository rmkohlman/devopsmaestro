// Package cmd provides shared confirmation prompt logic for destructive operations.
// Used by cascade delete commands (ecosystem, domain, app, gitrepo) to prevent
// accidental data loss. Requires --force to bypass prompt, and explicitly requires
// --force when stdin is not a terminal (piped/scripted).
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rmkohlman/MaestroSDK/render"
	"golang.org/x/term"
)

// confirmDelete prompts the user to confirm a destructive operation.
// If force is true, the prompt is skipped.
// If stdin is not a terminal (piped/scripted), returns an error requiring --force.
// Returns true if the user confirms, false if they abort.
func confirmDelete(message string, force bool) (bool, error) {
	if force {
		return true, nil
	}

	// If stdin is not a terminal, require --force for safety
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, fmt.Errorf("stdin is not a terminal — use --force to confirm deletion in non-interactive mode")
	}

	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	if response != "y" && response != "Y" {
		render.Info("Aborted")
		return false, nil
	}
	return true, nil
}
