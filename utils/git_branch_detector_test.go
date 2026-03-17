//go:build !integration

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseDefaultBranch verifies that ParseDefaultBranch correctly extracts
// the default branch name from raw `git ls-remote --symref <url> HEAD` output.
func TestParseDefaultBranch(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantBranch     string
		wantErrContain string // empty means no error expected
	}{
		{
			name:       "standard main branch",
			input:      "ref: refs/heads/main\tHEAD\nabc123\tHEAD\n",
			wantBranch: "main",
		},
		{
			name:       "master branch",
			input:      "ref: refs/heads/master\tHEAD\nabc123\tHEAD\n",
			wantBranch: "master",
		},
		{
			name:       "develop branch",
			input:      "ref: refs/heads/develop\tHEAD\nabc123\tHEAD\n",
			wantBranch: "develop",
		},
		{
			name:       "custom branch name",
			input:      "ref: refs/heads/trunk\tHEAD\nabc123\tHEAD\n",
			wantBranch: "trunk",
		},
		{
			name:       "branch with slashes",
			input:      "ref: refs/heads/release/v1\tHEAD\nabc123\tHEAD\n",
			wantBranch: "release/v1",
		},
		{
			name:           "empty output",
			input:          "",
			wantBranch:     "",
			wantErrContain: "could not detect",
		},
		{
			name:           "no symref line",
			input:          "abc123\tHEAD\n",
			wantBranch:     "",
			wantErrContain: "could not detect",
		},
		{
			name:           "malformed symref",
			input:          "ref: not-a-valid-ref\tHEAD\n",
			wantBranch:     "",
			wantErrContain: "could not detect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, err := ParseDefaultBranch(tt.input)

			if tt.wantErrContain != "" {
				assert.Error(t, err, "expected an error but got none")
				assert.Contains(t, err.Error(), tt.wantErrContain,
					"error message should contain %q", tt.wantErrContain)
				assert.Equal(t, "", branch, "branch should be empty on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.wantBranch, branch,
					"ParseDefaultBranch should return %q", tt.wantBranch)
			}
		})
	}
}
