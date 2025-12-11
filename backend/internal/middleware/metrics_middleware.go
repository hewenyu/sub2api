package middleware

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/backend/internal/metrics"
	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// MetricsMiddleware records request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		apiKeyID := "unknown"
		if key, exists := c.Get("api_key"); exists {
			if apiKey, ok := key.(*model.APIKey); ok {
				apiKeyID = strconv.FormatInt(apiKey.ID, 10)
			}
		}

		accountIDStr := "unknown"
		if id, exists := c.Get("account_id"); exists {
			if accountID, ok := id.(int64); ok {
				accountIDStr = strconv.FormatInt(accountID, 10)
			}
		}

		modelStr := "unknown"
		if m, exists := c.Get("model"); exists {
			if model, ok := m.(string); ok {
				modelStr = model
			}
		}

		status := strconv.Itoa(c.Writer.Status())

		metrics.RequestsTotal.WithLabelValues(apiKeyID, accountIDStr, modelStr, status).Inc()
		metrics.RequestDuration.WithLabelValues(apiKeyID, accountIDStr, modelStr).Observe(duration)

		if usage, exists := c.Get("usage"); exists {
			if u, ok := usage.(map[string]interface{}); ok {
				if inputTokens, ok := u["input_tokens"].(int); ok {
					metrics.TokensTotal.WithLabelValues(apiKeyID, accountIDStr, modelStr, "input").Add(float64(inputTokens))
				}
				if outputTokens, ok := u["output_tokens"].(int); ok {
					metrics.TokensTotal.WithLabelValues(apiKeyID, accountIDStr, modelStr, "output").Add(float64(outputTokens))
				}
				if cacheReadTokens, ok := u["cache_read_tokens"].(int); ok && cacheReadTokens > 0 {
					metrics.TokensTotal.WithLabelValues(apiKeyID, accountIDStr, modelStr, "cache_read").Add(float64(cacheReadTokens))
				}
				if cacheCreationTokens, ok := u["cache_creation_tokens"].(int); ok && cacheCreationTokens > 0 {
					metrics.TokensTotal.WithLabelValues(apiKeyID, accountIDStr, modelStr, "cache_write").Add(float64(cacheCreationTokens))
				}
			}
		}

		if cost, exists := c.Get("cost"); exists {
			if c, ok := cost.(float64); ok {
				metrics.CostTotal.WithLabelValues(apiKeyID, accountIDStr, modelStr).Add(c)
			}
		}
	}
}
