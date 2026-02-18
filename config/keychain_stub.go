//go:build !darwin

package config

import (
	"fmt"
	"runtime"
)

// GetFromKeychain retrieves a password from the system keychain
// On non-macOS platforms, this returns an error
func GetFromKeychain(service string) (string, error) {
	return "", fmt.Errorf("keychain integration not available on %s (only macOS supported)", runtime.GOOS)
}

// AddToKeychain adds or updates a password in the system keychain
// On non-macOS platforms, this returns an error
func AddToKeychain(service, value string) error {
	return fmt.Errorf("keychain integration not available on %s (only macOS supported)", runtime.GOOS)
}

// DeleteFromKeychain removes a password from the system keychain
// On non-macOS platforms, this returns an error
func DeleteFromKeychain(service string) error {
	return fmt.Errorf("keychain integration not available on %s (only macOS supported)", runtime.GOOS)
}

// KeychainAvailable returns true if the system keychain is available
func KeychainAvailable() bool {
	return false
}
