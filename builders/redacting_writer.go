package builders

import (
	"bytes"
	"io"
	"sort"
)

// minSecretLen is the minimum length a buildArg value must have to be treated
// as a secret worth redacting. Shorter values (version numbers, flags, etc.)
// are ignored to avoid false positives.
const minSecretLen = 8

// redactionMarker is the replacement string for redacted secrets.
var redactionMarker = []byte("***")

// RedactingWriter wraps an io.Writer and replaces sensitive values in output.
type RedactingWriter struct {
	inner   io.Writer
	secrets [][]byte // sorted longest-first
	maxLen  int      // length of longest secret
	pending []byte   // buffered tail from previous Write
}

// NewRedactingWriter creates a RedactingWriter that replaces secret values from
// buildArgs with "***" in all output written to inner.
// If no secrets need redacting, returns inner directly for zero overhead.
func NewRedactingWriter(inner io.Writer, buildArgs map[string]string) io.Writer {
	// Collect qualifying secrets (>= minSecretLen)
	var secrets [][]byte
	for _, v := range buildArgs {
		if len(v) >= minSecretLen {
			secrets = append(secrets, []byte(v))
		}
	}

	// Fast path: no secrets to redact — zero overhead
	if len(secrets) == 0 {
		return inner
	}

	// Sort secrets longest-first to prevent partial matches when one secret
	// is a prefix/substring of another.
	sort.Slice(secrets, func(i, j int) bool {
		return len(secrets[i]) > len(secrets[j])
	})

	maxLen := len(secrets[0]) // secrets[0] is the longest after sort

	return &RedactingWriter{
		inner:   inner,
		secrets: secrets,
		maxLen:  maxLen,
	}
}

// Write implements io.Writer. It buffers up to maxLen-1 trailing bytes between
// calls so that secrets split across Write boundaries are still detected.
func (w *RedactingWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Prepend any pending bytes from the previous Write call
	buf := make([]byte, len(w.pending)+len(p))
	copy(buf, w.pending)
	copy(buf[len(w.pending):], p)
	w.pending = nil

	// Redact all secrets (longest first)
	for _, secret := range w.secrets {
		buf = bytes.ReplaceAll(buf, secret, redactionMarker)
	}

	// Determine how many trailing bytes to hold back. We retain up to
	// maxLen-1 bytes so that a secret split across two Write calls can be
	// detected when the next chunk arrives.
	retain := w.maxLen - 1
	if retain > len(buf) {
		retain = len(buf)
	}

	// Flush everything except the retained tail
	toWrite := buf[:len(buf)-retain]
	if len(toWrite) > 0 {
		if _, err := w.inner.Write(toWrite); err != nil {
			return 0, err
		}
	}

	// Store the tail as pending for next call
	w.pending = make([]byte, retain)
	copy(w.pending, buf[len(buf)-retain:])

	// io.Writer contract: return the original input length
	return len(p), nil
}

// Flush writes any remaining buffered bytes, performing a final redaction pass.
func (w *RedactingWriter) Flush() error {
	if len(w.pending) == 0 {
		return nil
	}

	// Final redaction pass on remaining pending bytes
	buf := w.pending
	w.pending = nil

	for _, secret := range w.secrets {
		buf = bytes.ReplaceAll(buf, secret, redactionMarker)
	}

	if len(buf) > 0 {
		_, err := w.inner.Write(buf)
		return err
	}
	return nil
}

// Close implements io.Closer by flushing remaining bytes.
func (w *RedactingWriter) Close() error {
	return w.Flush()
}
