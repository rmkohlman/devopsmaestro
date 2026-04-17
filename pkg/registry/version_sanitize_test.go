package registry

import "testing"

func TestSanitizeVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple version string",
			input: "7.5.0",
			want:  "7.5.0",
		},
		{
			name:  "multi-line athens output",
			input: "Build Details:\n    Version:    v0.14.1\n    Date: 2024-06-02-23:35:29-UTC",
			want:  "0.14.1",
		},
		{
			name:  "version with v prefix",
			input: "v2.1.15",
			want:  "2.1.15",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "already clean semver",
			input: "1.2.3",
			want:  "1.2.3",
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  "",
		},
		{
			name:  "version key lowercase",
			input: "some tool\nversion: v3.0.0\nbuild date: today",
			want:  "3.0.0",
		},
		{
			name:  "semver with prerelease",
			input: "v1.0.0-rc1",
			want:  "1.0.0-rc1",
		},
		{
			name:  "multi-line fallback to first semver",
			input: "tool output\nrunning version 2.5.1 of something",
			want:  "2.5.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeVersion(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
