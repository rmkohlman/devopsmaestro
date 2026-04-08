// Package utils provides shared utility functions for devopsmaestro.
package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses a duration string that supports Go's standard duration
// format plus a "d" suffix for days (e.g., "90d", "7d", "24h", "1h30m").
// Days are converted to hours (1d = 24h).
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Check for day suffix
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.ParseFloat(daysStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid day duration %q: %w", s, err)
		}
		if days <= 0 {
			return 0, fmt.Errorf("duration must be positive, got %q", s)
		}
		return time.Duration(days * float64(24*time.Hour)), nil
	}

	// Fall back to standard Go duration parsing
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive, got %q", s)
	}
	return d, nil
}
