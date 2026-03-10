package render

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlank(t *testing.T) {
	var buf bytes.Buffer
	SetWriter(&buf)
	defer SetWriter(os.Stdout)

	err := Blank()
	require.NoError(t, err)
	assert.Equal(t, "\n", buf.String())
}

func TestFormattedMessageFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetWriter(&buf)
	defer SetWriter(os.Stdout)

	// Use plain renderer for predictable output
	SetDefault(RendererPlain)
	defer SetDefault(RendererColored)

	tests := []struct {
		name           string
		fn             func(string, ...any) error
		format         string
		args           []any
		expectedPrefix string
		expectedBody   string
	}{
		{
			name:           "Infof",
			fn:             Infof,
			format:         "hello %s",
			args:           []any{"world"},
			expectedPrefix: "[INFO]",
			expectedBody:   "hello world",
		},
		{
			name:           "Successf",
			fn:             Successf,
			format:         "created %d items",
			args:           []any{5},
			expectedPrefix: "[OK]",
			expectedBody:   "created 5 items",
		},
		{
			name:           "Warningf",
			fn:             Warningf,
			format:         "disk usage at %d%%",
			args:           []any{90},
			expectedPrefix: "[WARN]",
			expectedBody:   "disk usage at 90%",
		},
		{
			name:           "Errorf",
			fn:             Errorf,
			format:         "failed to connect to %s:%d",
			args:           []any{"localhost", 8080},
			expectedPrefix: "[ERROR]",
			expectedBody:   "failed to connect to localhost:8080",
		},
		{
			name:           "Progressf",
			fn:             Progressf,
			format:         "building image %s",
			args:           []any{"myapp:latest"},
			expectedPrefix: "->",
			expectedBody:   "building image myapp:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := tt.fn(tt.format, tt.args...)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tt.expectedPrefix)
			assert.Contains(t, buf.String(), tt.expectedBody)
		})
	}
}

func TestStderrFunctions(t *testing.T) {
	// Save original stderr and replace with a buffer via a pipe
	// We use MsgTo with an explicit writer, so we can test by capturing
	// the writer argument. Since the Stderr functions hard-code os.Stderr,
	// we redirect os.Stderr to a pipe for testing.

	// Use plain renderer for predictable output
	SetDefault(RendererPlain)
	defer SetDefault(RendererColored)

	t.Run("WarningToStderr", func(t *testing.T) {
		// Create a pipe to capture stderr
		r, w, err := os.Pipe()
		require.NoError(t, err)

		origStderr := os.Stderr
		os.Stderr = w

		err = WarningToStderr("disk space low")
		require.NoError(t, err)

		// Close write end so read completes
		w.Close()
		os.Stderr = origStderr

		var buf bytes.Buffer
		_, err = buf.ReadFrom(r)
		require.NoError(t, err)
		r.Close()

		assert.Contains(t, buf.String(), "[WARN]")
		assert.Contains(t, buf.String(), "disk space low")
	})

	t.Run("ErrorToStderr", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)

		origStderr := os.Stderr
		os.Stderr = w

		err = ErrorToStderr("connection failed")
		require.NoError(t, err)

		w.Close()
		os.Stderr = origStderr

		var buf bytes.Buffer
		_, err = buf.ReadFrom(r)
		require.NoError(t, err)
		r.Close()

		assert.Contains(t, buf.String(), "[ERROR]")
		assert.Contains(t, buf.String(), "connection failed")
	})

	t.Run("InfoToStderr", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)

		origStderr := os.Stderr
		os.Stderr = w

		err = InfoToStderr("starting process")
		require.NoError(t, err)

		w.Close()
		os.Stderr = origStderr

		var buf bytes.Buffer
		_, err = buf.ReadFrom(r)
		require.NoError(t, err)
		r.Close()

		assert.Contains(t, buf.String(), "[INFO]")
		assert.Contains(t, buf.String(), "starting process")
	})
}
