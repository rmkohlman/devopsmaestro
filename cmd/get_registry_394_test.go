package cmd

// =============================================================================
// Issue #394: Tests for formatDuration helper in get_registry.go
// =============================================================================
// formatDuration converts a time.Duration to a human-readable string used in
// the UPTIME column of `dvm get registries`.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration_Seconds(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"zero seconds", 0, "0s"},
		{"one second", time.Second, "1s"},
		{"59 seconds", 59 * time.Second, "59s"},
		{"just under a minute", time.Minute - time.Second, "59s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatDuration(tt.input))
		})
	}
}

func TestFormatDuration_Minutes(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"one minute", time.Minute, "1m"},
		{"30 minutes", 30 * time.Minute, "30m"},
		{"59 minutes", 59 * time.Minute, "59m"},
		{"just under an hour", time.Hour - time.Minute, "59m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatDuration(tt.input))
		})
	}
}

func TestFormatDuration_Hours(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"one hour", time.Hour, "1h"},
		{"12 hours", 12 * time.Hour, "12h"},
		{"23 hours", 23 * time.Hour, "23h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatDuration(tt.input))
		})
	}
}

func TestFormatDuration_Days(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  string
	}{
		{"exactly one day", 24 * time.Hour, "1d"},
		{"two days", 48 * time.Hour, "2d"},
		{"one day two hours", 26 * time.Hour, "1d 2h"},
		{"three days two hours", 74 * time.Hour, "3d 2h"},
		{"large uptime", 100 * 24 * time.Hour, "100d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatDuration(tt.input))
		})
	}
}
