package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWithTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(&buf)

	loggerWithTrace := WithTraceID("trace-123", logger)
	loggerWithTrace.Info("test message")

	var logEntry map[string]any
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.Equal(t, "trace-123", logEntry["trace_id"])
	assert.Equal(t, "test message", logEntry["msg"])
}

func TestWithContext_TraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(&buf)

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Set("trace_id", "context-trace-456")

	loggerWithContext := WithContext(c, logger)
	loggerWithContext.Info("test message")

	var logEntry map[string]any
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.Equal(t, "context-trace-456", logEntry["trace_id"])
}

func TestWithContext_NoTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(&buf)

	ctx := context.Background()
	loggerWithContext := WithContext(ctx, logger)
	loggerWithContext.Info("test message")

	var logEntry map[string]any
	json.Unmarshal(buf.Bytes(), &logEntry)

	_, exists := logEntry["trace_id"]
	assert.False(t, exists)
}

func TestWithUser(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(&buf)

	loggerWithUser := WithUser(42, logger)
	loggerWithUser.Info("test message")

	var logEntry map[string]any
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.Equal(t, float64(42), logEntry["user_id"])
}

func TestWithAccount(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(&buf)

	loggerWithAccount := WithAccount(99, logger)
	loggerWithAccount.Info("test message")

	var logEntry map[string]any
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.Equal(t, float64(99), logEntry["account_id"])
}

func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	logger := createTestLogger(&buf)

	err := errors.New("test error")
	LogError(logger, err, "operation failed", zap.String("operation", "test"))

	logOutput := buf.String()
	assert.Contains(t, logOutput, "operation failed")
	assert.Contains(t, logOutput, "test error")
	assert.Contains(t, logOutput, "error_type")
	assert.Contains(t, logOutput, "stack_trace")
	assert.Contains(t, logOutput, "operation")
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"invalid", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := GetLogLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func createTestLogger(buf *bytes.Buffer) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(buf),
		zapcore.DebugLevel,
	)

	return zap.New(core)
}
