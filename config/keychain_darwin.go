//go:build darwin

package config

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// keychainSubcommand returns the correct security(1) subcommand for the given type.
func keychainSubcommand(kt KeychainType) string {
	if kt == KeychainTypeInternet {
		return "find-internet-password"
	}
	return "find-generic-password"
}

// keychainExitError maps well-known macOS security(1) exit codes to descriptive errors.
func keychainExitError(exitCode int, service string) error {
	switch exitCode {
	case 36:
		return fmt.Errorf("keychain lookup cancelled by user for %q", service)
	case 44:
		return fmt.Errorf("credential %q not found in Keychain", service)
	case 51:
		return fmt.Errorf("keychain access denied for %q", service)
	default:
		return fmt.Errorf("keychain lookup failed for %q (exit code %d)", service, exitCode)
	}
}

// GetFromKeychain retrieves a password from macOS Keychain using label-based lookup.
// It uses the -l (label) flag instead of -s (service) and does NOT pass -a $USER,
// allowing it to find entries regardless of which account created them.
func GetFromKeychain(lookup KeychainLookup) (string, error) {
	if lookup.Label == "" {
		return "", fmt.Errorf("keychain label is required")
	}

	// Determine the security subcommand based on keychain type
	subcommand := keychainSubcommand(lookup.KeychainType)

	// Use -l (label) instead of -s (service), and do NOT pass -a $USER
	cmd := exec.Command("security", subcommand, "-l", lookup.Label, "-w")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", keychainExitError(exitErr.ExitCode(), lookup.Label)
		}
		return "", fmt.Errorf("keychain lookup failed for %q: %w", lookup.Label, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// AddToKeychain adds or updates a password in macOS Keychain
func AddToKeychain(service, value string) error {
	user := os.Getenv("USER")
	if user == "" {
		return fmt.Errorf("USER environment variable not set")
	}

	// Delete existing entry (ignore error if not found)
	deleteCmd := exec.Command("security", "delete-generic-password",
		"-s", service,
		"-a", user)
	deleteCmd.Run() // Ignore errors

	// Add new entry
	addCmd := exec.Command("security", "add-generic-password",
		"-s", service,
		"-a", user,
		"-w", value)

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add credential to Keychain: %w", err)
	}

	return nil
}

// DeleteFromKeychain removes a password from macOS Keychain
func DeleteFromKeychain(service string) error {
	user := os.Getenv("USER")
	if user == "" {
		return fmt.Errorf("USER environment variable not set")
	}

	cmd := exec.Command("security", "delete-generic-password",
		"-s", service,
		"-a", user)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete credential from Keychain: %w", err)
	}

	return nil
}

// KeychainAvailable returns true if macOS Keychain is available
func KeychainAvailable() bool {
	return true
}

// GetAccountFromKeychain retrieves the account (username) field from a macOS Keychain entry.
// Unlike GetFromKeychain which uses -w to get the password, this function parses
// the full entry output to extract the "acct" attribute.
// It uses the -l (label) flag and does NOT pass -a $USER.
func GetAccountFromKeychain(lookup KeychainLookup) (string, error) {
	if lookup.Label == "" {
		return "", fmt.Errorf("keychain label is required")
	}

	subcommand := keychainSubcommand(lookup.KeychainType)

	// Use -l (label), no -a $USER, no -w (need full output to parse acct)
	cmd := exec.Command("security", subcommand, "-l", lookup.Label)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", keychainExitError(exitErr.ExitCode(), lookup.Label)
		}
		return "", fmt.Errorf("keychain lookup failed for %q: %w", lookup.Label, err)
	}

	// Parse the "acct" attribute from security output
	// Format: "acct"<blob>="value"
	acctRegex := regexp.MustCompile(`"acct"<blob>="([^"]*)"`)
	matches := acctRegex.FindSubmatch(output)
	if matches == nil {
		return "", fmt.Errorf("account field not found in keychain entry for %q", lookup.Label)
	}

	account := string(matches[1])
	if account == "<NULL>" || account == "" {
		return "", fmt.Errorf("account field is empty in keychain entry for %q", lookup.Label)
	}

	return account, nil
}
