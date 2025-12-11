package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/metrics"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
)

// RateLimitMiddleware enforces rate limits with multiple time windows.
func RateLimitMiddleware(rateLimiter limit.RateLimiter, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.Next()
			return
		}

		windows := []struct {
			name    string
			seconds int
			limit   int
		}{
			{"Minute", 60, apiKey.RateLimitPerMinute},
			{"Hour", 3600, apiKey.RateLimitPerHour},
			{"Day", 86400, apiKey.RateLimitPerDay},
		}

		for _, w := range windows {
			if w.limit <= 0 {
				continue
			}

			key := fmt.Sprintf("ratelimit:apikey:%d:%d", apiKey.ID, w.seconds)
			allowed, remaining, resetAt, err := rateLimiter.CheckWindow(c.Request.Context(), key, int64(w.limit), w.seconds)
			if err != nil {
				logger.Error("Failed to check rate limit",
					zap.Error(err),
					zap.String("window", w.name),
				)
				c.Next()
				return
			}

			apiKeyIDStr := strconv.FormatInt(apiKey.ID, 10)
			windowName := strings.ToLower(w.name)
			metrics.RateLimitCurrent.WithLabelValues(apiKeyIDStr, windowName).Set(float64(int64(w.limit) - remaining))

			c.Header(fmt.Sprintf("X-RateLimit-Limit-%s", w.name), fmt.Sprintf("%d", w.limit))
			c.Header(fmt.Sprintf("X-RateLimit-Remaining-%s", w.name), fmt.Sprintf("%d", remaining))
			c.Header(fmt.Sprintf("X-RateLimit-Reset-%s", w.name), fmt.Sprintf("%d", resetAt.Unix()))

			if !allowed {
				metrics.RateLimitHitsTotal.WithLabelValues(apiKeyIDStr, windowName).Inc()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": fmt.Sprintf("Rate limit exceeded (per %s)", w.name),
					"details": gin.H{
						"limit":    w.limit,
						"window":   w.name,
						"reset_at": resetAt.Unix(),
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
