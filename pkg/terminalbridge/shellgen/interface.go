package shellgen

import "fmt"

// ShellGenerator defines the interface for generating shell configuration files.
// Each implementation produces syntactically correct output for its target shell.
type ShellGenerator interface {
	// Generate produces the full shell config file content from the given config.
	Generate(config ShellConfig) (string, error)

	// ShellName returns the name of the shell this generator targets (e.g., "zsh", "bash", "fish").
	ShellName() string
}

// NewShellGenerator returns the appropriate ShellGenerator for the given shell name.
// Supported values: "zsh", "bash", "fish".
func NewShellGenerator(shell string) (ShellGenerator, error) {
	switch shell {
	case "zsh":
		return NewZshGenerator(), nil
	case "bash":
		return NewBashGenerator(), nil
	case "fish":
		return NewFishGenerator(), nil
	default:
		return nil, fmt.Errorf("unsupported shell: %q (supported: zsh, bash, fish)", shell)
	}
}
