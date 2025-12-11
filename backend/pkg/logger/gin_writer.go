package logger

import (
	"go.uber.org/zap"
)

// GinWriter is an io.Writer that writes to a zap logger.
type GinWriter struct {
	logger *zap.Logger
}

// NewGinWriter creates a new GinWriter that writes to the given zap logger.
func NewGinWriter(logger *zap.Logger) *GinWriter {
	return &GinWriter{logger: logger}
}

// Write implements io.Writer interface.
// It writes the data to the zap logger as an info message.
func (w *GinWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	// Log the message
	w.logger.Info(msg)

	return len(p), nil
}
