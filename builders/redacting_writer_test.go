package builders

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// =============================================================================
// SM-7: RedactingWriter Tests
// RED: These tests FAIL until RedactingWriter is fully implemented.
// =============================================================================

// TestRedactingWriter_BasicRedaction verifies that a single secret value in
// the output stream is replaced with "***".
func TestRedactingWriter_BasicRedaction(t *testing.T) {
	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"GITHUB_PAT": "ghp_abc123xyz",
	})

	input := "Downloading https://ghp_abc123xyz@github.com/repo"
	_, err := w.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Flush if the writer supports it
	if fw, ok := w.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	}

	got := buf.String()
	if strings.Contains(got, "ghp_abc123xyz") {
		t.Errorf("BasicRedaction: output still contains secret; got %q", got)
	}
	want := "Downloading https://***@github.com/repo"
	if got != want {
		t.Errorf("BasicRedaction: got %q, want %q", got, want)
	}
}

// TestRedactingWriter_MultipleSecrets verifies that multiple different secrets
// are all redacted from the output.
func TestRedactingWriter_MultipleSecrets(t *testing.T) {
	tests := []struct {
		name      string
		buildArgs map[string]string
		input     string
		wantIn    string   // substring that MUST appear
		notWantIn []string // substrings that must NOT appear
	}{
		{
			name: "two secrets both redacted",
			buildArgs: map[string]string{
				"GITHUB_PAT": "ghp_secret1_longvalue",
				"NPM_TOKEN":  "npm_secret2_longvalue",
			},
			input:     "pat=ghp_secret1_longvalue token=npm_secret2_longvalue",
			wantIn:    "***",
			notWantIn: []string{"ghp_secret1_longvalue", "npm_secret2_longvalue"},
		},
		{
			name: "three secrets all redacted",
			buildArgs: map[string]string{
				"SECRET_A": "secretvalueAAA",
				"SECRET_B": "secretvalueBBB",
				"SECRET_C": "secretvalueCCC",
			},
			input:     "a=secretvalueAAA b=secretvalueBBB c=secretvalueCCC",
			wantIn:    "***",
			notWantIn: []string{"secretvalueAAA", "secretvalueBBB", "secretvalueCCC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewRedactingWriter(&buf, tt.buildArgs)

			_, err := w.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if fw, ok := w.(interface{ Flush() error }); ok {
				if err := fw.Flush(); err != nil {
					t.Fatalf("Flush() error = %v", err)
				}
			}

			got := buf.String()
			if !strings.Contains(got, tt.wantIn) {
				t.Errorf("%s: output missing redaction marker %q; got %q", tt.name, tt.wantIn, got)
			}
			for _, bad := range tt.notWantIn {
				if strings.Contains(got, bad) {
					t.Errorf("%s: output still contains secret %q; got %q", tt.name, bad, got)
				}
			}
		})
	}
}

// TestRedactingWriter_SecretAppearsMultipleTimes verifies that every occurrence
// of a secret within a single Write call is replaced.
func TestRedactingWriter_SecretAppearsMultipleTimes(t *testing.T) {
	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"TOKEN": "ghp_abc12345",
	})

	input := "token=ghp_abc12345 and again token=ghp_abc12345 done"
	_, err := w.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if fw, ok := w.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	}

	got := buf.String()
	if strings.Contains(got, "ghp_abc12345") {
		t.Errorf("SecretAppearsMultipleTimes: secret still present in output; got %q", got)
	}

	// Should have two redaction markers
	count := strings.Count(got, "***")
	if count < 2 {
		t.Errorf("SecretAppearsMultipleTimes: expected at least 2 redactions, got %d in %q", count, got)
	}
}

// TestRedactingWriter_NoSecrets verifies that when buildArgs is empty (or
// contains no long-enough values), the writer uses the zero-overhead fast path
// and returns the inner writer directly.
func TestRedactingWriter_NoSecrets(t *testing.T) {
	var buf bytes.Buffer
	inner := &buf

	// Empty buildArgs — no secrets to redact
	w := NewRedactingWriter(inner, map[string]string{})

	// Fast path: the returned writer should BE the inner writer
	if w != io.Writer(inner) {
		t.Error("NoSecrets: expected NewRedactingWriter to return inner writer directly (zero-overhead fast path)")
	}

	input := "normal output with no secrets"
	_, err := w.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if buf.String() != input {
		t.Errorf("NoSecrets: got %q, want %q", buf.String(), input)
	}
}

// TestRedactingWriter_NilBuildArgs verifies that a nil buildArgs map also uses
// the zero-overhead fast path.
func TestRedactingWriter_NilBuildArgs(t *testing.T) {
	var buf bytes.Buffer
	inner := &buf

	w := NewRedactingWriter(inner, nil)

	// Fast path: the returned writer should BE the inner writer
	if w != io.Writer(inner) {
		t.Error("NilBuildArgs: expected NewRedactingWriter to return inner writer directly for nil map")
	}
}

// TestRedactingWriter_ShortValuesIgnored verifies that buildArg values shorter
// than 8 characters are NOT treated as secrets (to avoid redacting common
// version numbers, flags, etc.).
func TestRedactingWriter_ShortValuesIgnored(t *testing.T) {
	tests := []struct {
		name      string
		buildArgs map[string]string
		input     string
		wantExact string
	}{
		{
			name:      "3-char value not redacted",
			buildArgs: map[string]string{"SHORT": "abc"},
			input:     "value=abc output",
			wantExact: "value=abc output",
		},
		{
			name:      "7-char value not redacted",
			buildArgs: map[string]string{"VER": "1.2.3.4"},
			input:     "version=1.2.3.4 built",
			wantExact: "version=1.2.3.4 built",
		},
		{
			name:      "empty value not redacted",
			buildArgs: map[string]string{"EMPTY": ""},
			input:     "nothing here",
			wantExact: "nothing here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewRedactingWriter(&buf, tt.buildArgs)

			// For short values, fast path should still be used
			if _, ok := w.(*RedactingWriter); ok {
				t.Errorf("%s: expected fast path (inner returned directly) for short values", tt.name)
			}

			_, err := w.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			got := buf.String()
			if got != tt.wantExact {
				t.Errorf("%s: got %q, want %q", tt.name, got, tt.wantExact)
			}
		})
	}
}

// TestRedactingWriter_MinLengthBoundary verifies that values of exactly 8
// characters ARE treated as secrets and get redacted.
func TestRedactingWriter_MinLengthBoundary(t *testing.T) {
	tests := []struct {
		name      string
		buildArgs map[string]string
		input     string
		secret    string
	}{
		{
			name:      "exactly 8 chars is redacted",
			buildArgs: map[string]string{"KEY": "12345678"},
			input:     "token=12345678 done",
			secret:    "12345678",
		},
		{
			name:      "exactly 9 chars is redacted",
			buildArgs: map[string]string{"KEY": "123456789"},
			input:     "token=123456789 done",
			secret:    "123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewRedactingWriter(&buf, tt.buildArgs)

			_, err := w.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if fw, ok := w.(interface{ Flush() error }); ok {
				if err := fw.Flush(); err != nil {
					t.Fatalf("Flush() error = %v", err)
				}
			}

			got := buf.String()
			if strings.Contains(got, tt.secret) {
				t.Errorf("%s: secret %q still present in output; got %q", tt.name, tt.secret, got)
			}
			if !strings.Contains(got, "***") {
				t.Errorf("%s: redaction marker *** missing from output; got %q", tt.name, got)
			}
		})
	}
}

// TestRedactingWriter_SplitAcrossWrites verifies that a secret split across
// two consecutive Write() calls is still fully redacted after Flush().
func TestRedactingWriter_SplitAcrossWrites(t *testing.T) {
	secret := "ghp_abc123xyz"

	// Split the secret: "ghp_abc1" in write 1, "23xyz" in write 2
	// (boundary is mid-secret)
	splitAt := 8
	part1 := "https://" + secret[:splitAt]
	part2 := secret[splitAt:] + "@github.com"

	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"PAT": secret,
	})

	_, err := w.Write([]byte(part1))
	if err != nil {
		t.Fatalf("Write(part1) error = %v", err)
	}

	_, err = w.Write([]byte(part2))
	if err != nil {
		t.Fatalf("Write(part2) error = %v", err)
	}

	if fw, ok := w.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	}

	got := buf.String()
	if strings.Contains(got, secret) {
		t.Errorf("SplitAcrossWrites: secret still present after flush; got %q", got)
	}
	if !strings.Contains(got, "***") {
		t.Errorf("SplitAcrossWrites: redaction marker *** missing; got %q", got)
	}
}

// TestRedactingWriter_Flush verifies that calling Flush after a Write causes
// all pending/buffered bytes to be written to the inner writer.
func TestRedactingWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"TOKEN": "my_secret_token_value",
	})

	rw, ok := w.(*RedactingWriter)
	if !ok {
		// FAIL: stub returns inner directly; implementation must return *RedactingWriter for secrets
		t.Fatal("NewRedactingWriter with secrets must return *RedactingWriter, got inner writer (stub not implemented)")
	}

	_, err := w.Write([]byte("output: my_secret_token_value end"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := rw.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	got := buf.String()
	// After flush, all data must be written
	if got == "" {
		t.Error("Flush: inner writer is empty after Flush(), expected data to be written")
	}
	// The secret must be redacted
	if strings.Contains(got, "my_secret_token_value") {
		t.Errorf("Flush: secret still present after Flush(); got %q", got)
	}
}

// TestRedactingWriter_LongestFirstPreventsPartialMatch verifies that when two
// secrets overlap (one is a prefix of the other), the longer secret is matched
// first to prevent a partial redaction leaving a fragment behind.
func TestRedactingWriter_LongestFirstPreventsPartialMatch(t *testing.T) {
	shortSecret := "ghp_abc123456"
	longSecret := "ghp_abc123456_extended"

	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"SHORT": shortSecret,
		"LONG":  longSecret,
	})

	// Input contains the longer secret — must produce exactly one ***
	input := "token=" + longSecret + " done"
	_, err := w.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if fw, ok := w.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	}

	got := buf.String()

	// Neither secret should be visible
	if strings.Contains(got, shortSecret) {
		t.Errorf("LongestFirst: short secret still visible; got %q", got)
	}
	if strings.Contains(got, longSecret) {
		t.Errorf("LongestFirst: long secret still visible; got %q", got)
	}

	// Should have exactly one redaction marker (not two from double-replacing)
	count := strings.Count(got, "***")
	if count != 1 {
		t.Errorf("LongestFirst: expected exactly 1 redaction marker, got %d in %q", count, got)
	}
}

// TestRedactingWriter_EmptyWrite verifies that Write([]byte{}) does not panic
// or return an error.
func TestRedactingWriter_EmptyWrite(t *testing.T) {
	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"TOKEN": "super_secret_value",
	})

	n, err := w.Write([]byte{})
	if err != nil {
		t.Errorf("EmptyWrite: Write([]byte{}) error = %v, want nil", err)
	}
	if n != 0 {
		t.Errorf("EmptyWrite: Write([]byte{}) returned n=%d, want 0", n)
	}
}

// TestRedactingWriter_BinaryOutput verifies that non-UTF8 / binary bytes pass
// through unchanged as long as they don't match a secret sequence.
func TestRedactingWriter_BinaryOutput(t *testing.T) {
	secret := "mysecretvalue123" // 16 chars, definitely a secret

	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"SECRET": secret,
	})

	// Binary data that does NOT contain the secret
	binaryData := []byte{0x00, 0x01, 0xFE, 0xFF, 0x80, 0x7F, 0x1B, 0x5B}
	_, err := w.Write(binaryData)
	if err != nil {
		t.Fatalf("BinaryOutput: Write() error = %v", err)
	}

	if fw, ok := w.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			t.Fatalf("Flush() error = %v", err)
		}
	}

	got := buf.Bytes()
	if !bytes.Equal(got, binaryData) {
		t.Errorf("BinaryOutput: binary data modified unexpectedly;\ngot  %v\nwant %v", got, binaryData)
	}
}

// TestRedactingWriter_ReturnsOriginalLength verifies that Write() returns
// len(p) (the length of the input), not the length of the redacted output.
// This is required by the io.Writer contract.
func TestRedactingWriter_ReturnsOriginalLength(t *testing.T) {
	tests := []struct {
		name      string
		buildArgs map[string]string
		input     string
	}{
		{
			name:      "secret shorter than input returns full length",
			buildArgs: map[string]string{"TOKEN": "supersecretvalue"},
			input:     "prefix supersecretvalue suffix",
		},
		{
			name:      "multiple secrets returns full length",
			buildArgs: map[string]string{"A": "secretvalue1234", "B": "anothervalue567"},
			input:     "a=secretvalue1234 b=anothervalue567",
		},
		{
			name:      "no match returns full length",
			buildArgs: map[string]string{"TOKEN": "notpresentvalue1"},
			input:     "completely different output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewRedactingWriter(&buf, tt.buildArgs)

			p := []byte(tt.input)
			n, err := w.Write(p)
			if err != nil {
				t.Fatalf("%s: Write() error = %v", tt.name, err)
			}
			if n != len(p) {
				t.Errorf("%s: Write() returned n=%d, want %d (io.Writer contract)", tt.name, n, len(p))
			}
		})
	}
}

// TestRedactingWriter_PipOutputSimulation simulates a real pip install line
// containing a GitHub PAT embedded in a private package URL, verifying that
// the PAT is fully redacted from the streamed output.
func TestRedactingWriter_PipOutputSimulation(t *testing.T) {
	realPAT := "ghp_WJ0M3TrealPATvalue"

	pipLines := strings.Join([]string{
		"Collecting mypackage @ git+https://" + realPAT + "@github.com/org/repo.git",
		"  Downloading https://" + realPAT + "@github.com/org/repo/archive/main.tar.gz",
		"Building wheels for collected packages: mypackage",
		"  Building wheel for mypackage (setup.py) ... done",
		"Successfully installed mypackage-1.0.0",
	}, "\n")

	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"PIP_EXTRA_INDEX_URL": "https://" + realPAT + "@github.com/org/repo",
		"GITHUB_PAT":          realPAT,
	})

	_, err := w.Write([]byte(pipLines))
	if err != nil {
		t.Fatalf("PipOutputSimulation: Write() error = %v", err)
	}

	if fw, ok := w.(interface{ Flush() error }); ok {
		if err := fw.Flush(); err != nil {
			t.Fatalf("PipOutputSimulation: Flush() error = %v", err)
		}
	}

	got := buf.String()

	// The PAT must not appear anywhere in the output
	if strings.Contains(got, realPAT) {
		t.Errorf("PipOutputSimulation: PAT still visible in output;\ngot:\n%s", got)
	}

	// Safe content must still be present
	if !strings.Contains(got, "Successfully installed") {
		t.Errorf("PipOutputSimulation: expected 'Successfully installed' to be preserved; got:\n%s", got)
	}
	if !strings.Contains(got, "***") {
		t.Errorf("PipOutputSimulation: expected *** redaction markers; got:\n%s", got)
	}
}

// TestRedactingWriter_Close verifies that Close() flushes remaining bytes
// and does not return an error on a fresh writer.
func TestRedactingWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	w := NewRedactingWriter(&buf, map[string]string{
		"TOKEN": "my_close_secret_val",
	})

	rw, ok := w.(*RedactingWriter)
	if !ok {
		// FAIL: stub returns inner directly; implementation must return *RedactingWriter for secrets
		t.Fatal("NewRedactingWriter with secrets must return *RedactingWriter, got inner writer (stub not implemented)")
	}

	_, err := w.Write([]byte("output my_close_secret_val end"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := rw.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	got := buf.String()
	if got == "" {
		t.Error("Close: nothing written to inner writer after Close()")
	}
	if strings.Contains(got, "my_close_secret_val") {
		t.Errorf("Close: secret not redacted after Close(); got %q", got)
	}
}
