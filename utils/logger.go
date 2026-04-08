package utils

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// InitLogger configures the global slog logger with the given level and format.
// All log output is written to stderr so it never interferes with stdout
// (which is reserved for user-facing output via render.*).
//
// level: "debug", "info", "warn", "error" (case-insensitive, default "warn")
// format: "text" or "json" (case-insensitive, default "text")
func InitLogger(level, format string) {
	lvl := ParseLogLevel(level)

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// ParseLogLevel converts a string log level to a slog.Level.
// Returns slog.LevelWarn for unrecognised values.
func ParseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

// ValidLogLevels returns the accepted log level strings.
func ValidLogLevels() []string {
	return []string{"debug", "info", "warn", "error"}
}

// ValidLogFormats returns the accepted log format strings.
func ValidLogFormats() []string {
	return []string{"text", "json"}
}

// ValidateLogLevel returns an error if the level string is not recognised.
func ValidateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "warning", "error":
		return nil
	default:
		return fmt.Errorf("invalid log level %q: must be one of debug, info, warn, error", level)
	}
}

// ValidateLogFormat returns an error if the format string is not recognised.
func ValidateLogFormat(format string) error {
	switch strings.ToLower(format) {
	case "text", "json":
		return nil
	default:
		return fmt.Errorf("invalid log format %q: must be one of text, json", format)
	}
}
