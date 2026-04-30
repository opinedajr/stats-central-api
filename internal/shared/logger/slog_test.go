package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectPanic bool
	}{
		{"debug level", "debug", false},
		{"info level", "info", false},
		{"warn level", "warn", false},
		{"error level", "error", false},
		{"unknown level defaults to info", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			if logger == nil {
				t.Error("NewLogger() returned nil")
			}
		})
	}
}

func TestLoggerInterface(t *testing.T) {
	logger := NewLogger("info")

	// Check if implements Logger interface
	var _ Logger = logger
}

func TestLoggerLevels(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("debug")

	// Test all log levels - these should not panic
	t.Run("Debug", func(t *testing.T) {
		logger.Debug(ctx, "test debug message", "key", "value")
	})

	t.Run("Info", func(t *testing.T) {
		logger.Info(ctx, "test info message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn(ctx, "test warn message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		logger.Error(ctx, "test error message", "key", "value")
	})
}

func TestLoggerWith(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("info")

	// Test With method
	loggerWithField := logger.With("component", "test")
	if loggerWithField == nil {
		t.Error("With() returned nil")
	}

	// Should not panic
	loggerWithField.Info(ctx, "test with field")

	// Test chaining multiple With calls
	loggerWithFields := logger.With("key1", "value1").With("key2", "value2")
	loggerWithFields.Info(ctx, "test with multiple fields")
}

func TestLoggerWithGroup(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("info")

	// Test WithGroup method
	loggerWithGroup := logger.WithGroup("test_group")
	if loggerWithGroup == nil {
		t.Error("WithGroup() returned nil")
	}

	// Should not panic
	loggerWithGroup.Info(ctx, "test with group", "key", "value")
}

func TestLoggerLog(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("debug")

	// Test Log method with custom level
	logger.Log(ctx, 42, "test custom level", "key", "value")
}

func TestLoggerLogAttrs(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger("debug")

	// Test LogAttrs method
	logger.LogAttrs(ctx, 42, "test with attrs",
		slog.String("string_attr", "value"),
		slog.Int("int_attr", 123),
	)
}
