package middleware

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggingMiddleware_TraceIDGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := createTestLogger(&logBuffer)

	router := gin.New()
	middleware := NewLoggingMiddleware(logger)
	router.Use(middleware.Handler())
	router.GET("/test", func(c *gin.Context) {
		traceID, exists := c.Get("trace_id")
		assert.True(t, exists)
		assert.NotEmpty(t, traceID)
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
}

func TestLoggingMiddleware_TraceIDPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := createTestLogger(&logBuffer)

	router := gin.New()
	middleware := NewLoggingMiddleware(logger)
	router.Use(middleware.Handler())
	router.GET("/test", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	expectedTraceID := "test-trace-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", expectedTraceID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, expectedTraceID, w.Header().Get("X-Trace-ID"))

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, expectedTraceID, response["trace_id"])
}

func TestLoggingMiddleware_SanitizeHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := createTestLogger(&logBuffer)

	router := gin.New()
	middleware := NewLoggingMiddleware(logger)
	router.Use(middleware.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("X-API-Key", "secret-key")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "secret-token")
	assert.NotContains(t, logOutput, "secret-key")
	assert.Contains(t, logOutput, "[REDACTED]")
	assert.Contains(t, logOutput, "application/json")
}

func TestLoggingMiddleware_SanitizeQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := createTestLogger(&logBuffer)

	router := gin.New()
	middleware := NewLoggingMiddleware(logger)
	router.Use(middleware.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test?password=secret123&token=abc&user=john", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "secret123")
	assert.NotContains(t, logOutput, "abc")
	assert.Contains(t, logOutput, "[REDACTED]")
	assert.Contains(t, logOutput, "john")
}

func TestLoggingMiddleware_RequestLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := createTestLogger(&logBuffer)

	router := gin.New()
	middleware := NewLoggingMiddleware(logger)
	router.Use(middleware.Handler())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/test?foo=bar", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "POST")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "trace_id")
	assert.Contains(t, logOutput, "status")
	assert.Contains(t, logOutput, "latency")
}

func TestLoggingMiddleware_ErrorLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer
	logger := createTestLogger(&logBuffer)

	router := gin.New()
	middleware := NewLoggingMiddleware(logger)
	router.Use(middleware.Handler())
	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "error")
	assert.Contains(t, logOutput, "500")
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
