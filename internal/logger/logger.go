package logger

import (
	"io"
	"log/slog"
	"os"
)

// Config configures the logger
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, file
}

// Logger wraps slog.Logger
type Logger struct {
	*slog.Logger
}

// New creates a new logger
func New(cfg Config) *Logger {
	// Parse level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Setup output
	var output io.Writer
	switch cfg.Output {
	case "stderr":
		output = os.Stderr
	case "stdout":
		output = os.Stdout
	default:
		output = os.Stdout
	}

	// Setup handler
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}
