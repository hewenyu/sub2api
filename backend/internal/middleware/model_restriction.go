package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Wei-Shaw/sub2api/backend/internal/service/limit"
)

// ModelRestrictionMiddleware enforces model restrictions.
func ModelRestrictionMiddleware(restrictor limit.ModelRestrictor, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKey(c)
		if apiKey == nil {
			c.Next()
			return
		}

		if !apiKey.EnableModelRestriction {
			c.Next()
			return
		}

		// Read request body to extract model
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			c.Next()
			return
		}

		// Restore body for downstream handlers so that subsequent middleware
		// and handlers (e.g. ShouldBindJSON) can read the original request body.
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Optionally expose the raw body to downstream handlers via context.
		c.Set("request_body", bodyBytes)

		// Parse request to get model
		var request struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(bodyBytes, &request); err != nil {
			logger.Error("Failed to parse request body", zap.Error(err))
			c.Next()
			return
		}

		// Check if model is allowed
		if request.Model != "" && !restrictor.IsModelAllowed(apiKey, request.Model) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Model not allowed",
				"details": gin.H{
					"model": request.Model,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
