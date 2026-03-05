package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateZotConfig generates a Zot registry configuration from a DVM config.
func GenerateZotConfig(cfg RegistryConfig) (map[string]interface{}, error) {
	// Validate config first
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Use default mirrors if none specified
	mirrors := cfg.Mirrors
	if len(mirrors) == 0 {
		mirrors = defaultMirrors()
	}

	// Build Zot config structure
	// Note: Zot v2.0+ expects address as just IP, port as string
	config := map[string]interface{}{
		"distSpecVersion": "1.1.0",
		"http": map[string]interface{}{
			"address": "127.0.0.1",
			"port":    fmt.Sprintf("%d", cfg.Port),
		},
		"storage": map[string]interface{}{
			"rootDirectory": cfg.Storage,
			"dedupe":        true,
			"gc":            true,
		},
		"log": map[string]interface{}{
			"level": "info",
		},
	}

	// Add extensions for mirroring/sync
	if len(mirrors) > 0 {
		registries := make([]interface{}, 0, len(mirrors))
		for _, mirror := range mirrors {
			reg := map[string]interface{}{
				"urls":     []string{mirror.URL}, // Zot v2.0+ expects "urls" array
				"onDemand": mirror.OnDemand,
			}

			// Add content config for prefix-based routing
			if mirror.Prefix != "" {
				reg["content"] = []map[string]interface{}{
					{
						"prefix": mirror.Prefix,
					},
				}
			}

			registries = append(registries, reg)
		}

		config["extensions"] = map[string]interface{}{
			"sync": map[string]interface{}{
				"enable":     true,
				"registries": registries,
			},
		}
	}

	return config, nil
}

// DefaultRegistryConfig returns a RegistryConfig with sensible defaults.
func DefaultRegistryConfig() RegistryConfig {
	homeDir, _ := os.UserHomeDir()
	return RegistryConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        5001,
		Storage:     filepath.Join(homeDir, ".devopsmaestro", "registry"),
		IdleTimeout: 30 * defaultMinute,
		Mirrors:     defaultMirrors(),
	}
}

// defaultMirrors returns the default set of registry mirrors.
func defaultMirrors() []MirrorConfig {
	return []MirrorConfig{
		{
			Name:     "docker-hub",
			URL:      "https://index.docker.io",
			OnDemand: true,
			Prefix:   "docker.io",
		},
		{
			Name:     "ghcr",
			URL:      "https://ghcr.io",
			OnDemand: true,
			Prefix:   "ghcr.io",
		},
	}
}

// Helper constant for time.Minute equivalent
const defaultMinute = 60000000000 // 60 seconds in nanoseconds
