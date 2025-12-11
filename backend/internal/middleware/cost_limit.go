package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
)

// CostLimitMiddleware enforces cost-based limits.
func CostLimitMiddleware(costLimiter limit.CostLimiter, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.Next()
			return
		}

		// Check daily cost limit
		if apiKey.DailyCostLimit > 0 {
			allowed, current, err := costLimiter.CheckDailyLimit(c.Request.Context(), apiKey.ID, apiKey.DailyCostLimit)
			if err != nil {
				logger.Error("Failed to check daily cost limit", zap.Error(err))
			} else if !allowed {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": "Daily cost limit exceeded",
					"details": gin.H{
						"current": current,
						"limit":   apiKey.DailyCostLimit,
					},
				})
				c.Abort()
				return
			}
		}

		// Check weekly cost limit
		if apiKey.WeeklyCostLimit > 0 {
			allowed, current, err := costLimiter.CheckWeeklyLimit(c.Request.Context(), apiKey.ID, apiKey.WeeklyCostLimit)
			if err != nil {
				logger.Error("Failed to check weekly cost limit", zap.Error(err))
			} else if !allowed {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": "Weekly cost limit exceeded",
					"details": gin.H{
						"current": current,
						"limit":   apiKey.WeeklyCostLimit,
					},
				})
				c.Abort()
				return
			}
		}

		// Check monthly cost limit
		if apiKey.MonthlyCostLimit > 0 {
			allowed, current, err := costLimiter.CheckMonthlyLimit(c.Request.Context(), apiKey.ID, apiKey.MonthlyCostLimit)
			if err != nil {
				logger.Error("Failed to check monthly cost limit", zap.Error(err))
			} else if !allowed {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": "Monthly cost limit exceeded",
					"details": gin.H{
						"current": current,
						"limit":   apiKey.MonthlyCostLimit,
					},
				})
				c.Abort()
				return
			}
		}

		// Check total cost limit
		if apiKey.TotalCostLimit > 0 {
			allowed, current, err := costLimiter.CheckTotalLimit(c.Request.Context(), apiKey.ID, apiKey.TotalCostLimit)
			if err != nil {
				logger.Error("Failed to check total cost limit", zap.Error(err))
			} else if !allowed {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": "Total cost limit exceeded",
					"details": gin.H{
						"current": current,
						"limit":   apiKey.TotalCostLimit,
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
