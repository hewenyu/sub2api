package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
)

// ConcurrencyLimitMiddleware enforces concurrent request limits.
func ConcurrencyLimitMiddleware(tracker limit.ConcurrencyTracker, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.Next()
			return
		}

		if apiKey.MaxConcurrentRequests <= 0 {
			c.Next()
			return
		}

		// Acquire concurrency slot atomically with limit enforcement
		requestID := uuid.New().String()
		acquired, err := tracker.Acquire(c.Request.Context(), apiKey.ID, requestID, 300, apiKey.MaxConcurrentRequests)
		if err != nil {
			logger.Error("Failed to acquire concurrency slot", zap.Error(err))
			c.Next()
			return
		}

		if !acquired {
			logger.Warn("Failed to acquire concurrency slot",
				zap.Int64("api_key_id", apiKey.ID),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Concurrency limit exceeded",
			})
			c.Abort()
			return
		}

		// Store request ID for cleanup
		c.Set("concurrency_request_id", requestID)

		// Release slot after request completes
		defer func() {
			if err := tracker.Release(c.Request.Context(), apiKey.ID, requestID); err != nil {
				logger.Error("Failed to release concurrency slot",
					zap.Int64("api_key_id", apiKey.ID),
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			}
		}()

		c.Next()
	}
}
