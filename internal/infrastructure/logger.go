// Package infrastructure provides infrastructure layer implementations.
package infrastructure

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger for structured logging with context.
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new structured logger.
// If inProduction is true, logs are JSON formatted; otherwise human-readable.
func NewLogger(inProduction bool) *Logger {
	var handler slog.Handler

	if inProduction {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// Info logs an info-level message with attributes.
func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(nil, slog.LevelInfo, msg, attrs...)
}

// Warn logs a warn-level message with attributes.
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(nil, slog.LevelWarn, msg, attrs...)
}

// Error logs an error-level message with attributes.
func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(nil, slog.LevelError, msg, attrs...)
}

// Debug logs a debug-level message with attributes.
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(nil, slog.LevelDebug, msg, attrs...)
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
