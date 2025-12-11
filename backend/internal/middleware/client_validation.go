package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
)

// ClientValidationMiddleware enforces client restrictions.
func ClientValidationMiddleware(validator limit.ClientValidator, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.Next()
			return
		}

		if !apiKey.EnableClientRestriction {
			c.Next()
			return
		}

		userAgent := c.GetHeader("User-Agent")
		if !validator.IsClientAllowed(apiKey, userAgent) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Client not allowed",
				"details": gin.H{
					"user_agent": userAgent,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
