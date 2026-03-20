package operators

import (
	"strings"
	"testing"
)

// TestShellEscape tests the shellEscape() function which is security-critical —
// it prevents command injection by ensuring user-supplied values are safely
// wrapped in single quotes with internal single quotes properly escaped.
//
// The escaping strategy:
//   - Wraps input in single quotes: 'input'
//   - Replaces each ' with '\” (end quote, escaped quote, start quote)
//
// See: operators/containerd_runtime_v2_start.go
func TestShellEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// ── Basic cases ─────────────────────────────────────────────────────────
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "plain alphanumeric string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "string with only numbers",
			input:    "12345",
			expected: "'12345'",
		},
		{
			name:     "string with hyphens and underscores",
			input:    "my-workspace_name",
			expected: "'my-workspace_name'",
		},

		// ── Whitespace ───────────────────────────────────────────────────────────
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "string with leading and trailing spaces",
			input:    "  spaced  ",
			expected: "'  spaced  '",
		},
		{
			name:     "string with tab character",
			input:    "hello\tworld",
			expected: "'hello\tworld'",
		},
		{
			name:     "string with newline character",
			input:    "line1\nline2",
			expected: "'line1\nline2'",
		},
		{
			name:     "string with carriage return",
			input:    "line1\rline2",
			expected: "'line1\rline2'",
		},
		{
			name:     "string with multiple newlines",
			input:    "a\nb\nc",
			expected: "'a\nb\nc'",
		},

		// ── Single quotes (the critical escaping case) ───────────────────────────
		{
			name:     "single quote only",
			input:    "'",
			expected: "''\\'''",
		},
		{
			name:     "string with one single quote",
			input:    "it's",
			expected: "'it'\\''s'",
		},
		{
			name:     "string with multiple single quotes",
			input:    "it's a test's value",
			expected: "'it'\\''s a test'\\''s value'",
		},
		{
			name:     "string that is only single quotes",
			input:    "'''",
			expected: "''\\'''\\'''\\'''",
		},
		{
			name:     "string starting with single quote",
			input:    "'leading",
			expected: "''\\''leading'",
		},
		{
			name:     "string ending with single quote",
			input:    "trailing'",
			expected: "'trailing'\\'''",
		},

		// ── Double quotes ────────────────────────────────────────────────────────
		{
			name:     "string with double quotes",
			input:    `say "hello"`,
			expected: `'say "hello"'`,
		},
		{
			name:     "string that is only double quotes",
			input:    `""`,
			expected: `'""'`,
		},
		{
			name:     "string with mixed single and double quotes",
			input:    `it's a "test"`,
			expected: `'it'\''s a "test"'`,
		},

		// ── Shell metacharacters ─────────────────────────────────────────────────
		{
			name:     "dollar sign variable expansion",
			input:    "$HOME",
			expected: "'$HOME'",
		},
		{
			name:     "dollar sign with braces",
			input:    "${PATH}",
			expected: "'${PATH}'",
		},
		{
			name:     "backtick command substitution",
			input:    "`whoami`",
			expected: "'`whoami`'",
		},
		{
			name:     "pipe character",
			input:    "foo|bar",
			expected: "'foo|bar'",
		},
		{
			name:     "semicolon",
			input:    "foo;bar",
			expected: "'foo;bar'",
		},
		{
			name:     "ampersand",
			input:    "foo&bar",
			expected: "'foo&bar'",
		},
		{
			name:     "double ampersand (and operator)",
			input:    "foo&&bar",
			expected: "'foo&&bar'",
		},
		{
			name:     "double pipe (or operator)",
			input:    "foo||bar",
			expected: "'foo||bar'",
		},
		{
			name:     "redirection operator greater-than",
			input:    "foo>bar",
			expected: "'foo>bar'",
		},
		{
			name:     "redirection operator less-than",
			input:    "foo<bar",
			expected: "'foo<bar'",
		},
		{
			name:     "exclamation mark (history expansion)",
			input:    "foo!bar",
			expected: "'foo!bar'",
		},
		{
			name:     "asterisk glob",
			input:    "*.go",
			expected: "'*.go'",
		},
		{
			name:     "question mark glob",
			input:    "file?.txt",
			expected: "'file?.txt'",
		},
		{
			name:     "tilde home expansion",
			input:    "~/Documents",
			expected: "'~/Documents'",
		},
		{
			name:     "hash comment character",
			input:    "foo#comment",
			expected: "'foo#comment'",
		},
		{
			name:     "backslash",
			input:    `foo\bar`,
			expected: `'foo\bar'`,
		},
		{
			name:     "parentheses subshell",
			input:    "(subshell)",
			expected: "'(subshell)'",
		},
		{
			name:     "curly braces brace expansion",
			input:    "{a,b,c}",
			expected: "'{a,b,c}'",
		},

		// ── Command injection attempts ───────────────────────────────────────────
		{
			name:     "injection: rm -rf /",
			input:    "; rm -rf /",
			expected: "'; rm -rf /'",
		},
		{
			name:     "injection: command substitution $(...)",
			input:    "$(malicious command)",
			expected: "'$(malicious command)'",
		},
		{
			name:     "injection: backtick substitution",
			input:    "`rm -rf /tmp`",
			expected: "'`rm -rf /tmp`'",
		},
		{
			name:     "injection: chained commands via semicolons",
			input:    "foo; cat /etc/passwd; echo done",
			expected: "'foo; cat /etc/passwd; echo done'",
		},
		{
			name:     "injection: pipe to shell",
			input:    "foo | sh",
			expected: "'foo | sh'",
		},
		{
			name:     "injection: newline-based injection",
			input:    "foo\nbad command",
			expected: "'foo\nbad command'",
		},
		{
			name:     "injection: env variable exfiltration",
			input:    "$SECRET_KEY",
			expected: "'$SECRET_KEY'",
		},
		{
			name:  "injection: nested single quotes with injection",
			input: "'; rm -rf /; echo '",
			// Input starts and ends with a single quote; each gets escaped as '\''
			expected: "''\\''; rm -rf /; echo '\\'''",
		},
		{
			name:     "injection: double ampersand chaining",
			input:    "valid && malicious",
			expected: "'valid && malicious'",
		},
		{
			name:     "injection: here-doc",
			input:    "foo <<EOF\nbad\nEOF",
			expected: "'foo <<EOF\nbad\nEOF'",
		},

		// ── Path-like strings ────────────────────────────────────────────────────
		{
			name:     "unix absolute path",
			input:    "/usr/local/bin/dvm",
			expected: "'/usr/local/bin/dvm'",
		},
		{
			name:     "path with spaces",
			input:    "/home/user/my documents/file.txt",
			expected: "'/home/user/my documents/file.txt'",
		},
		{
			name:     "path with single quote in directory name",
			input:    "/home/user/it's here/file.txt",
			expected: "'/home/user/it'\\''s here/file.txt'",
		},
		{
			name:     "relative path",
			input:    "../../../etc/passwd",
			expected: "'../../../etc/passwd'",
		},
		{
			name:     "path with glob characters",
			input:    "/var/log/*.log",
			expected: "'/var/log/*.log'",
		},
		{
			name:     "docker image reference",
			input:    "registry.example.com/org/image:latest",
			expected: "'registry.example.com/org/image:latest'",
		},
		{
			name:     "container name with dvm prefix",
			input:    "dvm-production-backend-userservice-dev",
			expected: "'dvm-production-backend-userservice-dev'",
		},

		// ── Already-escaped strings (no double-escaping) ─────────────────────────
		{
			name:     "already single-quoted string",
			input:    "'already quoted'",
			expected: "''\\''already quoted'\\'''",
		},
		{
			name:  "string with escaped single quote sequence",
			input: "it'\\''s",
			// The input contains 3 single quotes (at positions: after 't', after '\', after ''); each gets escaped
			expected: "'it'\\''\\'\\'''\\''s'",
		},

		// ── Unicode and multibyte characters ─────────────────────────────────────
		{
			name:     "unicode emoji",
			input:    "hello 🌍",
			expected: "'hello 🌍'",
		},
		{
			name:     "japanese characters",
			input:    "テスト",
			expected: "'テスト'",
		},
		{
			name:     "arabic text",
			input:    "مرحبا",
			expected: "'مرحبا'",
		},
		{
			name:     "chinese characters",
			input:    "你好世界",
			expected: "'你好世界'",
		},
		{
			name:     "mixed ascii and unicode",
			input:    "workspace-数据-dev",
			expected: "'workspace-数据-dev'",
		},
		{
			name:     "unicode with single quote",
			input:    "café's",
			expected: "'café'\\''s'",
		},

		// ── Null bytes ───────────────────────────────────────────────────────────
		{
			name:     "null byte in string",
			input:    "foo\x00bar",
			expected: "'foo\x00bar'",
		},

		// ── Very long strings ────────────────────────────────────────────────────
		{
			name:     "very long string without special chars",
			input:    strings.Repeat("a", 10000),
			expected: "'" + strings.Repeat("a", 10000) + "'",
		},
		{
			name:     "very long string with single quotes",
			input:    strings.Repeat("a'b", 1000),
			expected: "'" + strings.Repeat("a'\\''b", 1000) + "'",
		},

		// ── Namespace and container name patterns (real-world usage) ─────────────
		{
			name:     "k8s namespace style",
			input:    "devopsmaestro",
			expected: "'devopsmaestro'",
		},
		{
			name:     "nerdctl namespace with special chars",
			input:    "my-namespace.v2",
			expected: "'my-namespace.v2'",
		},
		{
			name:     "workspace name with version tag",
			input:    "workspace:v1.2.3",
			expected: "'workspace:v1.2.3'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shellEscape(tt.input)
			if got != tt.expected {
				t.Errorf("shellEscape(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestShellEscape_OutputAlwaysWrappedInSingleQuotes verifies the invariant that
// the output of shellEscape always begins and ends with a single quote.
func TestShellEscape_OutputAlwaysWrappedInSingleQuotes(t *testing.T) {
	inputs := []string{
		"",
		"simple",
		"'",
		"it's",
		"$HOME",
		"`whoami`",
		"; rm -rf /",
		"\n",
		"\t",
		"🌍",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got := shellEscape(input)
			if len(got) < 2 {
				t.Errorf("shellEscape(%q) = %q: output too short to be properly quoted", input, got)
				return
			}
			if got[0] != '\'' {
				t.Errorf("shellEscape(%q) = %q: does not start with single quote", input, got)
			}
			if got[len(got)-1] != '\'' {
				t.Errorf("shellEscape(%q) = %q: does not end with single quote", input, got)
			}
		})
	}
}

// TestShellEscape_NoRawSingleQuotesInsideOuterQuotes verifies that the output
// never contains a bare single-quote character within the outermost quoting
// context — every internal single quote must be followed by \\” (the escape
// sequence) or preceded by one, ensuring the shell sees no unmatched quotes.
//
// Validation approach: strip the outer wrapping quotes and verify that any
// remaining single quote is always part of the escape sequence '\”.
func TestShellEscape_EscapeSequenceIsCorrect(t *testing.T) {
	// A string with N single quotes should produce exactly N occurrences of
	// the escape sequence '\''.
	tests := []struct {
		name            string
		input           string
		expectedEscapes int
	}{
		{"zero single quotes", "hello", 0},
		{"one single quote", "it's", 1},
		{"two single quotes", "it's a test's", 2},
		{"three single quotes", "a'b'c'd", 3},
		{"five single quotes", strings.Repeat("'", 5), 5},
	}

	escapeSeq := "'\\''"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shellEscape(tt.input)
			count := strings.Count(got, escapeSeq)
			if count != tt.expectedEscapes {
				t.Errorf("shellEscape(%q) = %q: expected %d escape sequences %q, got %d",
					tt.input, got, tt.expectedEscapes, escapeSeq, count)
			}
		})
	}
}

// TestShellEscape_MetacharactersNotInterpreted verifies that known shell
// metacharacters are rendered inert — they appear verbatim in the output
// and are not converted or removed.
func TestShellEscape_MetacharactersNotInterpreted(t *testing.T) {
	metacharacters := []struct {
		name string
		char string
	}{
		{"dollar sign", "$"},
		{"backtick", "`"},
		{"pipe", "|"},
		{"semicolon", ";"},
		{"ampersand", "&"},
		{"greater-than", ">"},
		{"less-than", "<"},
		{"exclamation", "!"},
		{"asterisk", "*"},
		{"question mark", "?"},
		{"open paren", "("},
		{"close paren", ")"},
		{"backslash", `\`},
	}

	for _, mc := range metacharacters {
		t.Run(mc.name, func(t *testing.T) {
			input := "prefix" + mc.char + "suffix"
			got := shellEscape(input)
			// The metacharacter must still appear in the output
			if !strings.Contains(got, mc.char) {
				t.Errorf("shellEscape(%q) = %q: metacharacter %q was removed or transformed",
					input, got, mc.char)
			}
		})
	}
}
