package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test NewLogger with all format/level combinations
func TestNewLogger_JSONFormat_DebugLevel(t *testing.T) {
	logger := NewLogger("json", "debug")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_JSONFormat_InfoLevel(t *testing.T) {
	logger := NewLogger("json", "info")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_JSONFormat_WarnLevel(t *testing.T) {
	logger := NewLogger("json", "warn")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_JSONFormat_ErrorLevel(t *testing.T) {
	logger := NewLogger("json", "error")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_TextFormat_DebugLevel(t *testing.T) {
	logger := NewLogger("text", "debug")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_TextFormat_InfoLevel(t *testing.T) {
	logger := NewLogger("text", "info")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_TextFormat_WarnLevel(t *testing.T) {
	logger := NewLogger("text", "warn")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestNewLogger_TextFormat_ErrorLevel(t *testing.T) {
	logger := NewLogger("text", "error")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestLoggerInfo(_ *testing.T) {
	logger := NewLogger("text", "debug")
	// Should not panic
	logger.Info("test message", String("key", "value"))
}

func TestLoggerWarn(_ *testing.T) {
	logger := NewLogger("text", "debug")
	// Should not panic
	logger.Warn("test warning", String("key", "value"))
}

func TestLoggerError(_ *testing.T) {
	logger := NewLogger("text", "debug")
	// Should not panic
	logger.Error("test error", String("key", "value"))
}

func TestLoggerDebug(_ *testing.T) {
	logger := NewLogger("text", "debug")
	// Should not panic
	logger.Debug("test debug", String("key", "value"))
}

func TestLoggerWithContext(t *testing.T) {
	logger := NewLogger("text", "debug")
	contextLogger := logger.WithContext(String("request_id", "123"))
	assert.NotNil(t, contextLogger)
}

func TestAttributeHelpers(t *testing.T) {
	logger := NewLogger("text", "debug")

	// Test String
	attr := String("key", "value")
	assert.Equal(t, "key", attr.Key)

	// Test Int
	attr = Int("count", 42)
	assert.Equal(t, "count", attr.Key)

	// Test Int64
	attr = Int64("duration", 1000)
	assert.Equal(t, "duration", attr.Key)

	// Test Duration
	attr = Duration("elapsed", "5s")
	assert.Equal(t, "elapsed", attr.Key)

	// Test Any
	attr = Any("data", map[string]string{"foo": "bar"})
	assert.Equal(t, "data", attr.Key)

	logger.Info("test attributes", attr)
}

func TestMultipleAttributes(_ *testing.T) {
	logger := NewLogger("text", "debug")
	// Should handle multiple attributes without panic
	logger.Info("test message",
		String("user", "alice"),
		Int("age", 30),
		String("action", "login"),
	)
}

// Test GetLevel method
func TestGetLevel_Debug(t *testing.T) {
	logger := NewLogger("text", "debug")
	assert.Equal(t, "debug", logger.GetLevel())
}

func TestGetLevel_Info(t *testing.T) {
	logger := NewLogger("text", "info")
	assert.Equal(t, "info", logger.GetLevel())
}

func TestGetLevel_Warn(t *testing.T) {
	logger := NewLogger("text", "warn")
	assert.Equal(t, "warn", logger.GetLevel())
}

func TestGetLevel_Error(t *testing.T) {
	logger := NewLogger("text", "error")
	assert.Equal(t, "error", logger.GetLevel())
}

// Test IsDebug method
func TestIsDebug_True(t *testing.T) {
	logger := NewLogger("text", "debug")
	assert.True(t, logger.IsDebug())
}

func TestIsDebug_False_Info(t *testing.T) {
	logger := NewLogger("text", "info")
	assert.False(t, logger.IsDebug())
}

func TestIsDebug_False_Warn(t *testing.T) {
	logger := NewLogger("text", "warn")
	assert.False(t, logger.IsDebug())
}

func TestIsDebug_False_Error(t *testing.T) {
	logger := NewLogger("text", "error")
	assert.False(t, logger.IsDebug())
}
