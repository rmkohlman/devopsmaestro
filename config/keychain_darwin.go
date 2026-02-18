//go:build darwin

package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetFromKeychain retrieves a password from macOS Keychain
func GetFromKeychain(service string) (string, error) {
	user := os.Getenv("USER")
	if user == "" {
		return "", fmt.Errorf("USER environment variable not set")
	}

	cmd := exec.Command("security", "find-generic-password",
		"-s", service,
		"-a", user,
		"-w")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 44 means item not found
			if exitErr.ExitCode() == 44 {
				return "", fmt.Errorf("credential %q not found in Keychain", service)
			}
			return "", fmt.Errorf("keychain lookup failed for %q: %s", service, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("keychain lookup failed for %q: %w", service, err)
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
