package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
)

// RateLimitMiddleware enforces rate limits per minute.
func RateLimitMiddleware(rateLimiter limit.RateLimiter, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.Next()
			return
		}

		// Check per-minute rate limit
		if apiKey.RateLimitPerMinute > 0 {
			window := limit.GetWindow(60)
			allowed, current, err := rateLimiter.CheckLimit(c.Request.Context(), apiKey.ID, window, int64(apiKey.RateLimitPerMinute))
			if err != nil {
				logger.Error("Failed to check rate limit", zap.Error(err))
				c.Next()
				return
			}

			if !allowed {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded (per minute)",
					"details": gin.H{
						"current": current,
						"limit":   apiKey.RateLimitPerMinute,
						"window":  "60s",
					},
				})
				c.Abort()
				return
			}

			// Increment counter
			if err := rateLimiter.IncrementCounter(c.Request.Context(), apiKey.ID, window, 60*time.Second); err != nil {
				logger.Error("Failed to increment rate limit counter", zap.Error(err))
			}
		}

		c.Next()
	}
}
