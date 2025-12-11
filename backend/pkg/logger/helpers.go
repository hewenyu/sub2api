package logger

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func WithContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if traceID, exists := ginCtx.Get("trace_id"); exists {
			if traceIDStr, ok := traceID.(string); ok {
				return logger.With(zap.String("trace_id", traceIDStr))
			}
		}
	}
	return logger
}

func WithTraceID(traceID string, logger *zap.Logger) *zap.Logger {
	return logger.With(zap.String("trace_id", traceID))
}

func WithUser(userID uint, logger *zap.Logger) *zap.Logger {
	return logger.With(zap.Uint("user_id", userID))
}

func WithAccount(accountID uint, logger *zap.Logger) *zap.Logger {
	return logger.With(zap.Uint("account_id", accountID))
}

func LogError(logger *zap.Logger, err error, msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.Error(err),
		zap.String("error_type", fmt.Sprintf("%T", err)),
		zap.Stack("stack_trace"))
	logger.Error(msg, fields...)
}

func GetLogLevel(level string) zapcore.Level {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return zapcore.InfoLevel
	}
	return parsedLevel
}
