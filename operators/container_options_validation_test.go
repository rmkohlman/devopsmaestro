package operators

import (
	"testing"
)

// =============================================================================
// Network Mode Validation Tests (#91)
// =============================================================================

func TestValidateNetworkMode_ValidModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{"empty (default)", ""},
		{"bridge", "bridge"},
		{"none", "none"},
		{"host", "host"},
		{"custom network", "my-network"},
		{"custom with dots", "my.network.v2"},
		{"custom with underscores", "my_network"},
		{"alphanumeric", "net123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateNetworkMode(tt.mode); err != nil {
				t.Errorf("ValidateNetworkMode(%q) = %v, want nil", tt.mode, err)
			}
		})
	}
}

func TestValidateNetworkMode_InvalidModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{"starts with hyphen", "-bad"},
		{"starts with dot", ".bad"},
		{"contains spaces", "my network"},
		{"contains slash", "my/network"},
		{"special chars", "net@work!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateNetworkMode(tt.mode); err == nil {
				t.Errorf("ValidateNetworkMode(%q) = nil, want error", tt.mode)
			}
		})
	}
}

func TestValidateNetworkMode_TooLong(t *testing.T) {
	longName := "a"
	for len(longName) <= 128 {
		longName += "a"
	}
	if err := ValidateNetworkMode(longName); err == nil {
		t.Error("ValidateNetworkMode with 129-char name should return error")
	}
}

// =============================================================================
// Memory Parsing Tests (#92)
// =============================================================================

func TestParseMemoryString_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"empty (no limit)", "", 0},
		{"megabytes lowercase", "512m", 512 * 1024 * 1024},
		{"megabytes uppercase", "512M", 512 * 1024 * 1024},
		{"gigabytes lowercase", "2g", 2 * 1024 * 1024 * 1024},
		{"gigabytes uppercase", "2G", 2 * 1024 * 1024 * 1024},
		{"kilobytes", "8192k", 8192 * 1024},
		{"plain bytes", "536870912", 536870912},
		{"4MB minimum", "4m", 4 * 1024 * 1024},
		{"1GB", "1g", 1024 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMemoryString(tt.input)
			if err != nil {
				t.Fatalf("ParseMemoryString(%q) error = %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseMemoryString(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseMemoryString_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"letters only", "abc"},
		{"negative", "-512m"},
		{"fractional", "1.5g"},
		{"below minimum", "1m"},
		{"below minimum bytes", "1024"},
		{"empty suffix with zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMemoryString(tt.input)
			if err == nil {
				t.Errorf("ParseMemoryString(%q) = nil error, want error", tt.input)
			}
		})
	}
}

// =============================================================================
// CPU Validation Tests (#92)
// =============================================================================

func TestValidateCPUs_Valid(t *testing.T) {
	tests := []struct {
		name string
		cpus float64
	}{
		{"zero (no limit)", 0},
		{"one core", 1.0},
		{"half core", 0.5},
		{"one and a half", 1.5},
		{"four cores", 4.0},
		{"max reasonable", 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCPUs(tt.cpus); err != nil {
				t.Errorf("ValidateCPUs(%f) = %v, want nil", tt.cpus, err)
			}
		})
	}
}

func TestValidateCPUs_Invalid(t *testing.T) {
	tests := []struct {
		name string
		cpus float64
	}{
		{"negative", -1.0},
		{"too high", 1025},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCPUs(tt.cpus); err == nil {
				t.Errorf("ValidateCPUs(%f) = nil, want error", tt.cpus)
			}
		})
	}
}

// =============================================================================
// StartOptions Integration Tests (#91, #92)
// =============================================================================

func TestStartOptions_NetworkAndResourceFields(t *testing.T) {
	opts := StartOptions{
		ImageName:     "test:latest",
		WorkspaceName: "test-ws",
		NetworkMode:   "none",
		CPUs:          2.0,
		Memory:        "4g",
	}

	// Verify the fields are accessible
	if opts.NetworkMode != "none" {
		t.Errorf("NetworkMode = %q, want %q", opts.NetworkMode, "none")
	}
	if opts.CPUs != 2.0 {
		t.Errorf("CPUs = %f, want 2.0", opts.CPUs)
	}
	if opts.Memory != "4g" {
		t.Errorf("Memory = %q, want %q", opts.Memory, "4g")
	}
}

func TestStartOptions_DefaultsAreNoLimits(t *testing.T) {
	opts := StartOptions{
		ImageName:     "test:latest",
		WorkspaceName: "test-ws",
	}

	if opts.NetworkMode != "" {
		t.Errorf("default NetworkMode = %q, want empty", opts.NetworkMode)
	}
	if opts.CPUs != 0 {
		t.Errorf("default CPUs = %f, want 0", opts.CPUs)
	}
	if opts.Memory != "" {
		t.Errorf("default Memory = %q, want empty", opts.Memory)
	}
}
