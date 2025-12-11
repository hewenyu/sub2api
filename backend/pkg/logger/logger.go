package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig holds the configuration for the logger.
type LoggingConfig struct {
	Level           string
	Format          string
	OutputPath      string
	ErrorOutputPath string
}

// NewLogger creates a new zap logger instance based on the provided configuration.
func NewLogger(config LoggingConfig) (*zap.Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if config.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Build configuration
	// Only output to file (not stdout/stderr) to keep console clean
	var outputPaths []string
	var errorOutputPaths []string

	if config.OutputPath != "" {
		// Create directory for output path if it doesn't exist
		if err := ensureLogDir(config.OutputPath); err != nil {
			return nil, err
		}
		outputPaths = append(outputPaths, config.OutputPath)
	} else {
		// Fallback to stdout if no file path specified
		outputPaths = append(outputPaths, "stdout")
	}

	if config.ErrorOutputPath != "" {
		// Create directory for error output path if it doesn't exist
		if err := ensureLogDir(config.ErrorOutputPath); err != nil {
			return nil, err
		}
		errorOutputPaths = append(errorOutputPaths, config.ErrorOutputPath)
	} else {
		// Fallback to stderr if no file path specified
		errorOutputPaths = append(errorOutputPaths, "stderr")
	}

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      config.Format != "json",
		Encoding:         getEncoding(config.Format),
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	}

	return zapConfig.Build()
}

// getEncoding returns the appropriate encoding based on the format.
func getEncoding(format string) string {
	if format == "json" {
		return "json"
	}
	return "console"
}

// ensureLogDir creates the directory for the log file if it doesn't exist.
func ensureLogDir(logPath string) error {
	// Skip if it's stdout or stderr
	if logPath == "stdout" || logPath == "stderr" {
		return nil
	}

	// Get directory from file path
	dir := filepath.Dir(logPath)
	if dir == "" || dir == "." {
		return nil
	}

	// Create directory with appropriate permissions
	return os.MkdirAll(dir, 0755)
}
