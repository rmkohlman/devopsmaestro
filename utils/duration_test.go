package utils

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     time.Duration
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "happy path - days",
			input:   "90d",
			want:    90 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "happy path - standard hours",
			input:   "24h",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "happy path - hours and minutes",
			input:   "1h30m",
			want:    1*time.Hour + 30*time.Minute,
			wantErr: false,
		},
		{
			name:    "edge case - single day",
			input:   "1d",
			want:    24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "edge case - fractional days",
			input:   "0.5d",
			want:    12 * time.Hour,
			wantErr: false,
		},
		{
			name:    "error case - empty string",
			input:   "",
			want:    0,
			wantErr: true,
			errMsg:  "empty duration string",
		},
		{
			name:    "error case - invalid format",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "error case - zero days",
			input:   "0d",
			want:    0,
			wantErr: true,
			errMsg:  "duration must be positive",
		},
		{
			name:    "error case - negative hours",
			input:   "-1h",
			want:    0,
			wantErr: true,
			errMsg:  "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDuration(%q) expected error, got nil", tt.input)
					return
				}
				if tt.errMsg != "" && err.Error() != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseDuration(%q) error = %q, want containing %q", tt.input, err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}