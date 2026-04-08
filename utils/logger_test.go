package utils

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"", slog.LevelWarn},
		{"invalid", slog.LevelWarn},
		{"trace", slog.LevelWarn},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLogLevel(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInitLogger_TextFormat(t *testing.T) {
	// Capture output by initializing with text format
	InitLogger("debug", "text")

	// Verify the default logger was set (no panic, handler works)
	slog.Debug("test message", "key", "value")
}

func TestInitLogger_JSONFormat(t *testing.T) {
	InitLogger("info", "json")

	// Verify the default logger was set
	slog.Info("test json message", "key", "value")
}

func TestInitLogger_DefaultsToWarn(t *testing.T) {
	InitLogger("", "text")

	// The default level for an empty string should be warn
	lvl := ParseLogLevel("")
	assert.Equal(t, slog.LevelWarn, lvl)
}

func TestInitLogger_VerboseSetsDebug(t *testing.T) {
	// Simulates --verbose flag by passing "debug" level
	lvl := ParseLogLevel("debug")
	assert.Equal(t, slog.LevelDebug, lvl)

	InitLogger("debug", "text")
	// Debug messages should be visible at debug level
	slog.Debug("verbose debug message")
}

func TestLevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		logLevel  slog.Level
		shouldLog bool
	}{
		{"debug level allows debug", "debug", slog.LevelDebug, true},
		{"debug level allows info", "debug", slog.LevelInfo, true},
		{"info level blocks debug", "info", slog.LevelDebug, false},
		{"info level allows info", "info", slog.LevelInfo, true},
		{"warn level blocks info", "warn", slog.LevelInfo, false},
		{"warn level allows warn", "warn", slog.LevelWarn, true},
		{"error level blocks warn", "error", slog.LevelWarn, false},
		{"error level allows error", "error", slog.LevelError, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			lvl := ParseLogLevel(tt.level)
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: lvl})
			logger := slog.New(handler)

			logger.Log(nil, tt.logLevel, "test message")

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String(), "expected log output")
			} else {
				assert.Empty(t, buf.String(), "expected no log output")
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	valid := []string{"debug", "info", "warn", "warning", "error"}
	for _, v := range valid {
		assert.NoError(t, ValidateLogLevel(v))
	}
	assert.Error(t, ValidateLogLevel("invalid"))
	assert.Error(t, ValidateLogLevel(""))
	assert.Error(t, ValidateLogLevel("trace"))
}

func TestValidateLogFormat(t *testing.T) {
	assert.NoError(t, ValidateLogFormat("text"))
	assert.NoError(t, ValidateLogFormat("json"))
	assert.NoError(t, ValidateLogFormat("TEXT"))
	assert.NoError(t, ValidateLogFormat("JSON"))
	assert.Error(t, ValidateLogFormat("yaml"))
	assert.Error(t, ValidateLogFormat(""))
}

func TestValidLogLevels(t *testing.T) {
	levels := ValidLogLevels()
	require.Len(t, levels, 4)
	assert.Contains(t, levels, "debug")
	assert.Contains(t, levels, "info")
	assert.Contains(t, levels, "warn")
	assert.Contains(t, levels, "error")
}

func TestValidLogFormats(t *testing.T) {
	formats := ValidLogFormats()
	require.Len(t, formats, 2)
	assert.Contains(t, formats, "text")
	assert.Contains(t, formats, "json")
}
