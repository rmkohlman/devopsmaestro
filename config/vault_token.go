package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// TokenCreator is a function type that creates a new vault token.
type TokenCreator func() (string, error)

// resolveExistingToken checks all existing sources for a vault token.
// Priority: env var -> viper config -> .vault_token file
// Returns empty string if no token found from any source.
// configDir must be a trusted path — not user-supplied input.
func resolveExistingToken(configDir string) string {
	// 1. Check MAV_TOKEN environment variable
	if token := os.Getenv("MAV_TOKEN"); token != "" {
		return token
	}

	// 2. Check viper config (vault.token)
	if token := viper.GetString("vault.token"); token != "" {
		return token
	}

	// 3. Read <configDir>/.vault_token file
	tokenFile := filepath.Join(configDir, ".vault_token")
	data, err := os.ReadFile(tokenFile)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			return token
		}
	}

	return ""
}

// persistToken writes a token to <configDir>/.vault_token with 0600 permissions.
func persistToken(configDir, token string) error {
	tokenFile := filepath.Join(configDir, ".vault_token")
	return os.WriteFile(tokenFile, []byte(token), 0600)
}

// mavTokenResponse represents the JSON response from `mav token create`.
type mavTokenResponse struct {
	Token string `json:"token"`
}

// defaultTokenCreator creates a vault token by shelling out to `mav token create`.
// Returns error if mav is not in PATH or token creation fails.
func defaultTokenCreator() (string, error) {
	mavPath, err := exec.LookPath("mav")
	if err != nil {
		return "", fmt.Errorf("mav not found in PATH: %w", err)
	}

	cmd := exec.Command(mavPath, "token", "create", "--name", "dvm-auto", "--scope", "read", "-o", "json")
	cmd.Stderr = nil
	// Security: only pass necessary environment variables (env allowlist)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mav token create failed: %w", err)
	}

	var resp mavTokenResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", fmt.Errorf("failed to parse mav token response: %w", err)
	}

	if resp.Token == "" {
		return "", fmt.Errorf("mav returned empty token")
	}

	return resp.Token, nil
}

// ResolveVaultTokenFromDir resolves a vault token using the full chain, with configDir injectable.
// Resolution: env -> viper -> file -> create (via TokenCreator)
// If no TokenCreator provided, uses defaultTokenCreator.
// Graceful degradation: if creation fails, returns ("", nil).
// configDir must be a trusted path — not user-supplied input.
func ResolveVaultTokenFromDir(configDir string, creator ...TokenCreator) (string, error) {
	// Check existing sources first
	if token := resolveExistingToken(configDir); token != "" {
		return token, nil
	}

	// Select creator
	var create TokenCreator
	if len(creator) > 0 && creator[0] != nil {
		create = creator[0]
	} else {
		create = defaultTokenCreator
	}

	// Attempt to create a new token
	token, err := create()
	if err != nil {
		slog.Warn("failed to auto-create vault token", "error", err)
		return "", nil // graceful degradation
	}

	if token == "" {
		return "", nil
	}

	// Persist the created token
	if err := persistToken(configDir, token); err != nil {
		slog.Warn("failed to persist vault token", "error", err)
		// Don't fail — we still have the token in memory
	}

	slog.Info("auto-created MaestroVault token")

	return token, nil
}

// ResolveVaultToken resolves a vault token using the production config directory.
// Convenience wrapper around ResolveVaultTokenFromDir.
func ResolveVaultToken(creator ...TokenCreator) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := filepath.Join(home, ".devopsmaestro")
	return ResolveVaultTokenFromDir(configDir, creator...)
}
