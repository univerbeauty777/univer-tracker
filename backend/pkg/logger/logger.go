// Package logger provides a structured logger built on log/slog.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New creates a new structured logger configured for the given environment.
// In development: human-readable text output with debug level.
// In production: JSON output with info level.
func New(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}

	if strings.ToLower(env) == "development" {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
