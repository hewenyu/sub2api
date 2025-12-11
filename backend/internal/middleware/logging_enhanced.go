package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	sensitiveHeaders = map[string]bool{
		"authorization": true,
		"x-api-key":     true,
		"cookie":        true,
		"set-cookie":    true,
	}

	sensitiveQueryParams = map[string]bool{
		"password": true,
		"token":    true,
		"api_key":  true,
		"secret":   true,
	}
)

type LoggingMiddleware struct {
	logger *zap.Logger
}

func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		sanitizedHeaders := m.sanitizeHeaders(c.Request.Header)
		sanitizedQuery := m.sanitizeQuery(c.Request.URL.Query())

		m.logger.Info("Request received",
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Any("query", sanitizedQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.Any("headers", sanitizedHeaders),
			zap.Int64("content_length", c.Request.ContentLength),
		)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int("response_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		if statusCode >= 500 {
			m.logger.Error("Request completed with server error", fields...)
		} else if statusCode >= 400 {
			m.logger.Warn("Request completed with client error", fields...)
		} else {
			m.logger.Info("Request completed", fields...)
		}
	}
}

func (m *LoggingMiddleware) sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = strings.Join(values, ", ")
		}
	}
	return sanitized
}

func (m *LoggingMiddleware) sanitizeQuery(query url.Values) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range query {
		lowerKey := strings.ToLower(key)
		if sensitiveQueryParams[lowerKey] {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = strings.Join(values, ", ")
		}
	}
	return sanitized
}
