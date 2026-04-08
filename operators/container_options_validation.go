package operators

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidNetworkModes are the well-known Docker/containerd network modes.
var ValidNetworkModes = []string{"bridge", "none", "host"}

// networkNameRegex validates custom network names (alphanumeric, hyphens, underscores).
var networkNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// ValidateNetworkMode validates the network mode string.
// Accepted values: "bridge", "none", "host", or a custom network name
// matching the networkNameRegex pattern.
// An empty string is treated as the default ("bridge") and is valid.
func ValidateNetworkMode(mode string) error {
	if mode == "" {
		return nil // empty means default (bridge)
	}

	// Check well-known modes
	for _, valid := range ValidNetworkModes {
		if mode == valid {
			return nil
		}
	}

	// Check custom network name format
	if !networkNameRegex.MatchString(mode) {
		return fmt.Errorf("invalid network mode %q: must be one of %v or a valid network name", mode, ValidNetworkModes)
	}

	if len(mode) > 128 {
		return fmt.Errorf("network name %q exceeds maximum length of 128 characters", mode)
	}

	return nil
}

// memoryRegex matches memory strings like "512m", "2g", "1024k", "1073741824".
var memoryRegex = regexp.MustCompile(`^(\d+)([kmgKMG])?[bB]?$`)

// ParseMemoryString parses a human-readable memory string (e.g., "512m", "2g")
// into bytes. Supported suffixes: k/K (kilobytes), m/M (megabytes), g/G (gigabytes).
// A plain number is interpreted as bytes.
// Returns 0 for an empty string (no limit).
func ParseMemoryString(memory string) (int64, error) {
	if memory == "" {
		return 0, nil
	}

	memory = strings.TrimSpace(memory)

	match := memoryRegex.FindStringSubmatch(memory)
	if match == nil {
		return 0, fmt.Errorf("invalid memory format %q: use a number with optional suffix (k, m, g), e.g. '512m', '2g'", memory)
	}

	value, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory value %q: %w", memory, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("memory value must be positive, got %d", value)
	}

	suffix := strings.ToLower(match[2])
	switch suffix {
	case "k":
		value *= 1024
	case "m":
		value *= 1024 * 1024
	case "g":
		value *= 1024 * 1024 * 1024
	case "":
		// Plain bytes, no conversion
	}

	// Minimum 4MB (Docker's minimum)
	const minMemory = 4 * 1024 * 1024
	if value < minMemory {
		return 0, fmt.Errorf("memory limit %q is below minimum of 4m (4 megabytes)", memory)
	}

	return value, nil
}

// ValidateCPUs validates the CPU limit value.
// Must be positive. A value of 0 means no limit.
func ValidateCPUs(cpus float64) error {
	if cpus < 0 {
		return fmt.Errorf("CPU limit must be non-negative, got %f", cpus)
	}
	if cpus > 1024 {
		return fmt.Errorf("CPU limit %f is unreasonably high (max 1024)", cpus)
	}
	return nil
}
