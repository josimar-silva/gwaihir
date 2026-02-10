// Package infrastructure provides infrastructure layer implementations.
package infrastructure

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger for structured logging with context.
type Logger struct {
	logger *slog.Logger
	level  string
}

// NewLogger creates a new structured logger.
// format: "json" or "text"
// level: "debug", "info", "warn", or "error"
func NewLogger(format, level string) *Logger {
	var handler slog.Handler
	var slogLevel slog.Level

	// Convert string level to slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create appropriate handler based on format
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	}

	return &Logger{
		logger: slog.New(handler),
		level:  level,
	}
}

// Info logs an info-level message with attributes.
func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

// Warn logs a warn-level message with attributes.
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
}

// Error logs an error-level message with attributes.
func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

// Debug logs a debug-level message with attributes.
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
}

// WithContext returns a new Logger with additional context attributes.
func (l *Logger) WithContext(attrs ...slog.Attr) *Logger {
	attrSlice := make([]any, len(attrs))
	for i, attr := range attrs {
		attrSlice[i] = attr
	}
	return &Logger{
		logger: l.logger.With(attrSlice...),
	}
}

// String converts an attribute key-value pair.
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int converts an attribute key-value pair.
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Int64 converts an attribute key-value pair.
func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

// Duration converts an attribute key-value pair.
func Duration(key string, value interface{}) slog.Attr {
	return slog.Any(key, value)
}

// Any converts an attribute key-value pair for any type.
func Any(key string, value interface{}) slog.Attr {
	return slog.Any(key, value)
}

// GetLevel returns the configured log level.
func (l *Logger) GetLevel() string {
	return l.level
}

// IsDebug returns true if the log level is "debug".
func (l *Logger) IsDebug() bool {
	return l.level == "debug"
}
