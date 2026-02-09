package infrastructure

import (
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(false)
	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}
	if logger.logger == nil {
		t.Fatal("Expected non-nil slog.Logger")
	}
}

func TestNewLoggerProduction(t *testing.T) {
	logger := NewLogger(true)
	if logger == nil {
		t.Fatal("Expected non-nil logger for production mode")
	}
	if logger.logger == nil {
		t.Fatal("Expected non-nil slog.Logger for production mode")
	}
}

func TestLoggerInfo(t *testing.T) {
	logger := NewLogger(false)
	// Should not panic
	logger.Info("test message", String("key", "value"))
}

func TestLoggerWarn(t *testing.T) {
	logger := NewLogger(false)
	// Should not panic
	logger.Warn("test warning", String("key", "value"))
}

func TestLoggerError(t *testing.T) {
	logger := NewLogger(false)
	// Should not panic
	logger.Error("test error", String("key", "value"))
}

func TestLoggerDebug(t *testing.T) {
	logger := NewLogger(false)
	// Should not panic
	logger.Debug("test debug", String("key", "value"))
}

func TestLoggerWithContext(t *testing.T) {
	logger := NewLogger(false)
	contextLogger := logger.WithContext(String("request_id", "123"))
	if contextLogger == nil {
		t.Fatal("Expected non-nil context logger")
	}
}

func TestAttributeHelpers(t *testing.T) {
	logger := NewLogger(false)

	// Test String
	attr := String("key", "value")
	if attr.Key != "key" {
		t.Errorf("Expected key 'key', got %s", attr.Key)
	}

	// Test Int
	attr = Int("count", 42)
	if attr.Key != "count" {
		t.Errorf("Expected key 'count', got %s", attr.Key)
	}

	// Test Int64
	attr = Int64("duration", 1000)
	if attr.Key != "duration" {
		t.Errorf("Expected key 'duration', got %s", attr.Key)
	}

	// Test Duration
	attr = Duration("elapsed", "5s")
	if attr.Key != "elapsed" {
		t.Errorf("Expected key 'elapsed', got %s", attr.Key)
	}

	// Test Any
	attr = Any("data", map[string]string{"foo": "bar"})
	if attr.Key != "data" {
		t.Errorf("Expected key 'data', got %s", attr.Key)
	}

	logger.Info("test attributes", attr)
}

func TestMultipleAttributes(t *testing.T) {
	logger := NewLogger(false)
	// Should handle multiple attributes without panic
	logger.Info("test message",
		String("user", "alice"),
		Int("age", 30),
		String("action", "login"),
	)
}
