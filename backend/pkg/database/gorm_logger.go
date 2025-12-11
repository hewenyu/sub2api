package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger is a custom logger for GORM that uses zap.
type GormLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// NewGormLogger creates a new GORM logger that uses zap.
func NewGormLogger(zapLogger *zap.Logger, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  logger.Info,
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: true,
	}
}

// LogMode sets the log level for the logger.
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs info messages.
func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn logs warning messages.
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error logs error messages.
func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace logs SQL queries with execution time.
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	}

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		fields = append(fields, zap.Error(err))
		l.ZapLogger.Error("Database error", fields...)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		fields = append(fields, zap.Duration("slow_threshold", l.SlowThreshold))
		l.ZapLogger.Warn("Slow SQL query", fields...)
	case l.LogLevel >= logger.Info:
		l.ZapLogger.Debug("SQL query", fields...)
	}
}

// Printf implements the logger.Writer interface (for compatibility).
func (l *GormLogger) Printf(format string, args ...any) {
	l.ZapLogger.Info(fmt.Sprintf(format, args...))
}
